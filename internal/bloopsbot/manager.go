package bloopsbot

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"runtime"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/bloops-games/bloops/internal/bloopsbot/builder"
	"github.com/bloops-games/bloops/internal/bloopsbot/match"
	"github.com/bloops-games/bloops/internal/bloopsbot/resource"
	"github.com/bloops-games/bloops/internal/bloopsbot/util"
	stateDB "github.com/bloops-games/bloops/internal/database/matchstate/database"
	matchstateModel "github.com/bloops-games/bloops/internal/database/matchstate/model"
	statDb "github.com/bloops-games/bloops/internal/database/stat/database"
	statModel "github.com/bloops-games/bloops/internal/database/stat/model"
	userDb "github.com/bloops-games/bloops/internal/database/user/database"
	userModel "github.com/bloops-games/bloops/internal/database/user/model"
	"github.com/bloops-games/bloops/internal/logging"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
)

type (
	commandCbHandlerFunc  = func(string) error
	commandHandlerFunc    = func(userModel.User, int64) error
	commandMiddlewareFunc = func(userModel.User, int64) (bool, error)
)

var ErrTelegramResponseTypeNotFound = fmt.Errorf("telegram response not found")

func NewManager(
	tg *tgbotapi.BotAPI,
	config *Config,
	userDB *userDb.DB,
	statDB *statDb.DB,
	stateDB *stateDB.DB,
) *manager {
	return &manager{
		tg:                   tg,
		config:               config,
		userBuildingSessions: map[int64]*builder.Session{},
		userMatchSessions:    map[int64]*match.Session{},
		matchSessions:        map[int64]*match.Session{},
		commandCbHandlers:    map[int64]commandCbHandlerFunc{},
		commandHandlers:      map[string]commandHandler{},
		userDB:               userDB,
		statDB:               statDB,
		stateDB:              stateDB,
	}
}

type commandHandler struct {
	commandFn    commandHandlerFunc
	middlewareFn []commandMiddlewareFunc
}

func (t commandHandler) execute(u userModel.User, chatID int64) error {
	for _, f := range t.middlewareFn {
		ok, err := f(u, chatID)
		if err != nil {
			return fmt.Errorf("command handler execute: %w", err)
		}

		if !ok {
			return nil
		}
	}

	return t.commandFn(u, chatID)
}

type manager struct {
	tg     *tgbotapi.BotAPI
	config *Config

	mtx sync.RWMutex
	// key: UserID active building session
	userBuildingSessions map[int64]*builder.Session
	// key: UserID active playing session
	userMatchSessions map[int64]*match.Session
	// key: generated int64 code
	matchSessions map[int64]*match.Session
	// command callbacks
	commandCbHandlers map[int64]commandCbHandlerFunc
	// command handlers
	commandHandlers map[string]commandHandler

	userDB     *userDb.DB
	statDB     *statDb.DB
	stateDB    *stateDB.DB
	cancel     func()
	ctxSess    context.Context
	cancelSess func()
}

func (m *manager) Stop() {
	m.cancel()
}

func (m *manager) Run(ctx context.Context) error {
	var updates tgbotapi.UpdatesChannel
	ctx, cancel := context.WithCancel(ctx)
	logger := logging.FromContext(ctx)
	m.cancel = cancel
	m.ctxSess, m.cancelSess = context.WithCancel(context.Background())

	if m.config.BotWebhookHookURL != "" {
		_, err := m.tg.SetWebhook(tgbotapi.NewWebhook(m.config.BotWebhookHookURL + m.config.BotToken))
		if err != nil {
			return fmt.Errorf("tg bot set webhook: %w", err)
		}

		info, err := m.tg.GetWebhookInfo()
		if err != nil {
			return fmt.Errorf("get webhook info: %w", err)
		}

		if info.LastErrorDate != 0 {
			logger.Errorf("Telegram callback failed: %s", info.LastErrorMessage)
		}

		updates = m.tg.ListenForWebhook("/" + m.config.BotToken)
		go func() {
			if err := http.ListenAndServe(m.config.BotWebhookAddr, nil); err != nil {
				logger.Fatalf("listen and serve http stopped: %v", err)
				cancel()
			}
		}()
	} else {
		resp, err := m.tg.RemoveWebhook()
		if err != nil {
			return fmt.Errorf("remove webhook: %w", err)
		}

		if !resp.Ok {
			if resp.ErrorCode > 0 {
				return fmt.Errorf("remove webhook with error code %d and description %s", resp.ErrorCode, resp.Description)
			}
			return fmt.Errorf("remove webhook response not ok=)")
		}

		upd := tgbotapi.NewUpdate(0)
		upd.Timeout = int(m.config.TgBotPollTimeout.Seconds())
		up, err := m.tg.GetUpdatesChan(upd)
		if err != nil {
			return fmt.Errorf("tg get updates chan: %w", err)
		}
		updates = up
	}

	userMiddleware := []commandMiddlewareFunc{m.isActive}
	adminMiddleware := []commandMiddlewareFunc{m.isAdmin}
	// register text command handlers
	m.registerCommandHandler(
		resource.CmdStart,
		commandHandler{commandFn: m.handleStartCommand, middlewareFn: userMiddleware},
	)
	m.registerCommandHandler(
		resource.CmdFeedback,
		commandHandler{commandFn: m.handleFeedbackCommand, middlewareFn: userMiddleware},
	)
	m.registerCommandHandler(
		resource.CmdRules,
		commandHandler{commandFn: m.handleRulesButton, middlewareFn: userMiddleware},
	)
	m.registerCommandHandler(
		resource.CmdProfile,
		commandHandler{commandFn: m.handleProfileCmd, middlewareFn: userMiddleware},
	)
	m.registerCommandHandler(
		resource.ProfileButtonText,
		commandHandler{commandFn: m.handleProfileButton, middlewareFn: userMiddleware},
	)
	m.registerCommandHandler(
		resource.CreateButtonText,
		commandHandler{commandFn: m.handleCreateButton, middlewareFn: userMiddleware},
	)
	m.registerCommandHandler(
		resource.JoinButtonText,
		commandHandler{commandFn: m.handleJoinButton, middlewareFn: userMiddleware},
	)
	m.registerCommandHandler(
		resource.LeaveButtonText,
		commandHandler{commandFn: m.handleButtonExit, middlewareFn: userMiddleware},
	)
	m.registerCommandHandler(
		resource.RuleButtonText,
		commandHandler{commandFn: m.handleRulesButton, middlewareFn: userMiddleware},
	)
	m.registerCommandHandler(
		resource.CmdAddPlayer,
		commandHandler{commandFn: m.handleRegisterOfflinePlayerCmd, middlewareFn: userMiddleware},
	)
	m.registerCommandHandler(
		resource.CmdBan,
		commandHandler{commandFn: m.handleBanCommand, middlewareFn: adminMiddleware},
	)

	// restoreInterruptedGames not completed sessions
	if err := m.restoreInterruptedGames(); err != nil {
		return fmt.Errorf("restoreInterruptedGames: %w", err)
	}

	wg := &sync.WaitGroup{}
	poolWorkerNum := runtime.NumCPU()
	wg.Add(poolWorkerNum)

	for i := 0; i < poolWorkerNum; i++ {
		go m.pool(ctx, wg, updates)
	}

	wg.Wait()
	m.shutdown()
	return nil
}

func (m *manager) pool(ctx context.Context, wg *sync.WaitGroup, updCh tgbotapi.UpdatesChannel) {
	defer wg.Done()
	logger := logging.FromContext(ctx).Named("manager.pool")
	for {
		select {
		case update := <-updCh:
			u, err := m.recvUser(update)
			if err != nil {
				logger.Errorf("recv user: %v", err)
				continue
			}

			if update.Message != nil {
				if update.Message.Chat.IsGroup() || update.Message.Chat.IsSuperGroup() {
					msg := tgbotapi.NewMessage(update.Message.Chat.ID, resource.TextChatNotAllowed)
					msg.ParseMode = tgbotapi.ModeMarkdown
					if _, err := m.tg.Send(msg); err != nil {
						logger.Errorf("send msg: %v", err)
					}
					continue
				}

				if err := m.route(ctx, u, update); err != nil {
					if !errors.Is(err, match.ErrValidation) {
						logger.Errorf("handle command query: %v", err)
					}
				}
			}

			if update.CallbackQuery != nil {
				if err := m.handleCallbackQuery(ctx, u, update); err != nil {
					logger.Errorf("handle commandCbHandler query: %v", err)
				}
			}
		case <-ctx.Done():
			return
		}
	}
}

func (m *manager) route(ctx context.Context, u userModel.User, upd tgbotapi.Update) error {
	logger := logging.FromContext(ctx).Named("bloopsbot.manager.route")
	logger.Infof("Command received from user %s, command %s", u.FirstName, upd.Message.Text)

	if handler, ok := m.commandHandler(upd.Message.Text); ok {
		if err := handler.execute(u, upd.Message.Chat.ID); err != nil {
			return fmt.Errorf("execute command text handler: %w", err)
		}

		return nil
	}

	if cb, ok := m.commandCbHandler(u.ID); ok {
		if err := cb(upd.Message.Text); err != nil {
			return fmt.Errorf("execute cb: %w", err)
		}

		return nil
	}

	if session, ok := m.userBuildingSession(u.ID); ok {
		if err := session.Execute(upd); err != nil {
			return fmt.Errorf("execute building session: %w", err)
		}

		return nil
	}

	if session, ok := m.userMatchSession(u.ID); ok {
		if err := session.Execute(u.ID, upd); err != nil {
			return fmt.Errorf("execute playing session: %w", err)
		}

		return nil
	}

	return nil
}

func (m *manager) handleCallbackQuery(ctx context.Context, u userModel.User, upd tgbotapi.Update) error {
	logger := logging.FromContext(ctx).Named("bloopsbot.manager.handlerCallbackQuery")
	logger.Infof(
		"Command received from user %s, command %s, data %s",
		u.FirstName,
		upd.CallbackQuery.Message.Text,
		upd.CallbackQuery.Data,
	)

	if session, ok := m.userBuildingSession(u.ID); ok {
		if err := session.Execute(upd); err != nil {
			return fmt.Errorf("execute building cb: %w", err)
		}
	}

	if session, ok := m.userMatchSession(u.ID); ok {
		if err := session.Execute(u.ID, upd); err != nil {
			return fmt.Errorf("execute playing cb: %w", err)
		}
	}

	return nil
}

func (m *manager) buildGameConfig(session *builder.Session, code int64) match.Config {
	config := match.Config{
		Timeout:    m.config.PlayingTimeout,
		Code:       code,
		Tg:         m.tg,
		DoneFn:     m.matchDoneFn,
		WarnFn:     m.matchWarnFn,
		AuthorID:   session.AuthorID,
		AuthorName: session.AuthorName,
		RoundsNum:  session.RoundsNum,
		RoundTime:  session.RoundTime,
		Bloopses:   []resource.Bloops{},
		Categories: []string{},
		Letters:    []string{},
		Vote:       session.Vote,
	}

	for _, category := range session.Categories {
		if category.Status {
			config.Categories = append(config.Categories, category.Text)
		}
	}

	for _, letter := range session.Letters {
		if letter.Status {
			config.Letters = append(config.Letters, letter.Text)
		}
	}

	if session.Bloops {
		config.Bloopses = make([]resource.Bloops, len(resource.Bloopses))
		copy(config.Bloopses, resource.Bloopses)
	}

	return config
}

func (m *manager) builderWarnFn(session *builder.Session) error {
	m.mtx.Lock()
	defer m.mtx.Unlock()
	delete(m.userBuildingSessions, session.AuthorID)

	return nil
}

func (m *manager) builderDoneFn(session *builder.Session) error {
	defer func() {
		m.mtx.Lock()
		defer m.mtx.Unlock()
		delete(m.userBuildingSessions, session.AuthorID)
	}()

	code, err := util.GenerateCodeHash()
	if err != nil {
		return fmt.Errorf("hash: %w", err)
	}

	for {
		if _, ok := m.matchSession(code); !ok {
			session := match.NewSession(m.buildGameConfig(session, code))
			session.Run(m.ctxSess)
			m.mtx.Lock()
			m.matchSessions[code] = session
			m.mtx.Unlock()
			break
		}
	}

	msg := tgbotapi.NewMessage(session.ChatID, resource.TextCreationGameCompletedSuccessfulMsg)
	msg.ParseMode = tgbotapi.ModeMarkdown
	if _, err := m.tg.Send(msg); err != nil {
		return fmt.Errorf("send msg: %w", err)
	}

	if _, err := m.tg.Send(tgbotapi.NewStickerShare(session.ChatID, resource.GenerateSticker(true))); err != nil {
		return fmt.Errorf("send msg: %w", err)
	}

	msg = tgbotapi.NewMessage(session.ChatID, strconv.Itoa(int(code)))
	msg.ParseMode = tgbotapi.ModeMarkdown
	msg.ReplyMarkup = resource.CommonButtons
	if _, err := m.tg.Send(msg); err != nil {
		return fmt.Errorf("send msg: %w", err)
	}

	return nil
}

func (m *manager) matchWarnFn(session *match.Session) error {
	m.mtx.Lock()
	defer m.mtx.Unlock()

	if err := m.serializeGames(session); err != nil {
		return fmt.Errorf("serializeGames match session: %w", err)
	}

	for _, player := range session.Players {
		delete(m.userMatchSessions, player.UserID)
	}

	delete(m.matchSessions, session.Code)

	return nil
}

func (m *manager) matchDoneFn(session *match.Session) error {
	m.mtx.Lock()
	defer m.mtx.Unlock()
	if err := m.appendStat(session); err != nil {
		return fmt.Errorf("append stat: %w", err)
	}
	for _, player := range session.Players {
		delete(m.userMatchSessions, player.UserID)
	}

	delete(m.matchSessions, session.Code)

	return nil
}

func (m *manager) registerCommandHandler(cmd string, handler commandHandler) {
	m.mtx.Lock()
	defer m.mtx.Unlock()
	m.commandHandlers[cmd] = handler
}

func (m *manager) commandHandler(cmd string) (commandHandler, bool) {
	m.mtx.RLock()
	defer m.mtx.RUnlock()
	handler, ok := m.commandHandlers[cmd]
	return handler, ok
}

func (m *manager) registerCommandCbHandler(userID int64, fn commandCbHandlerFunc) {
	m.mtx.Lock()
	defer m.mtx.Unlock()
	m.commandCbHandlers[userID] = fn
}

func (m *manager) commandCbHandler(userID int64) (func(msg string) error, bool) {
	m.mtx.RLock()
	defer m.mtx.RUnlock()
	cb, ok := m.commandCbHandlers[userID]
	return cb, ok
}

func (m *manager) resetUserSessions(userID int64) {
	m.mtx.Lock()
	defer m.mtx.Unlock()
	delete(m.userBuildingSessions, userID)
	delete(m.userMatchSessions, userID)
	delete(m.commandCbHandlers, userID)
}

func (m *manager) userBuildingSession(userID int64) (*builder.Session, bool) {
	m.mtx.RLock()
	defer m.mtx.RUnlock()
	session, ok := m.userBuildingSessions[userID]

	return session, ok
}

func (m *manager) userMatchSession(userID int64) (*match.Session, bool) {
	m.mtx.RLock()
	defer m.mtx.RUnlock()
	session, ok := m.userMatchSessions[userID]

	return session, ok
}

func (m *manager) matchSession(code int64) (*match.Session, bool) {
	m.mtx.RLock()
	defer m.mtx.RUnlock()
	session, ok := m.matchSessions[code]

	return session, ok
}

func (m *manager) shutdown() {
	var ss int
	m.cancelSess()
	m.mtx.RLock()
	bn, ms := len(m.userBuildingSessions), len(m.matchSessions)
	m.mtx.RUnlock()
	ss = bn + ms
	ticker := time.NewTicker(200 * time.Millisecond)
	for ss > 0 {
		select {
		case <-ticker.C:
			m.mtx.RLock()
			bn, ms := len(m.userBuildingSessions), len(m.matchSessions)
			m.mtx.RUnlock()
			ss = bn + ms
		default:
		}
	}
}

func (m *manager) recvUser(upd tgbotapi.Update) (userModel.User, error) {
	var tgUser *tgbotapi.User
	var u userModel.User
	switch {
	case upd.CallbackQuery != nil:
		tgUser = upd.CallbackQuery.From
	case upd.Message != nil:
		tgUser = upd.Message.From
	default:
		return u, ErrTelegramResponseTypeNotFound
	}

	u, err := m.userDB.Fetch(int64(tgUser.ID))
	if err != nil {
		if errors.Is(err, userDb.ErrNotFound) {
			username := strings.TrimPrefix(tgUser.UserName, "@")
			adminUsername := m.config.Admin

			newUser := userModel.User{
				ID:           int64(tgUser.ID),
				FirstName:    tgUser.FirstName,
				LastName:     tgUser.LastName,
				LanguageCode: tgUser.LanguageCode,
				Username:     tgUser.UserName,
				Admin:        username == adminUsername,
				Status:       userModel.StatusActive,
				CreatedAt:    time.Now(),
			}

			if err := m.userDB.Store(newUser); err != nil {
				return u, fmt.Errorf("userdb store: %w", err)
			}
			u = newUser
		}
	}

	stat, err := m.statDB.FetchRateStat(u.ID)
	if err != nil {
		if errors.Is(err, statDb.ErrNotFound) {
			return u, nil
		}
		return u, fmt.Errorf("fetch profile stat: %w", err)
	}

	u.Stars = stat.Stars
	u.Bloops = stat.Bloops

	return u, nil
}

func NewMatchSessionFromSerialized(
	ser matchstateModel.State,
	tg *tgbotapi.BotAPI,
	doneFn func(session *match.Session) error,
	warnFn func(session *match.Session) error,
) *match.Session {
	c := match.Config{
		AuthorID:   ser.AuthorID,
		AuthorName: ser.AuthorName,
		RoundsNum:  ser.RoundsNum,
		RoundTime:  ser.RoundTime,
		Categories: make([]string, len(ser.Categories)),
		Letters:    make([]string, len(ser.Letters)),
		Bloopses:   make([]resource.Bloops, len(ser.Bloopses)),
		Vote:       ser.Vote,
		Code:       ser.Code,
		Timeout:    ser.Timeout,
		Tg:         tg,
		DoneFn:     doneFn,
		WarnFn:     warnFn,
	}

	copy(c.Categories, ser.Categories)
	copy(c.Letters, ser.Letters)
	copy(c.Bloopses, ser.Bloopses)

	s := match.NewSession(c)
	s.State = ser.State
	s.CurrRoundIdx = ser.CurrRoundIdx
	s.Players = make([]*matchstateModel.Player, len(ser.Players))
	copy(s.Players, ser.Players)
	return s
}

func (m *manager) serializeGames(session *match.Session) error {
	s := matchstateModel.State{
		Timeout:      session.Config.Timeout,
		AuthorID:     session.Config.AuthorID,
		AuthorName:   session.Config.AuthorName,
		RoundsNum:    session.Config.RoundsNum,
		RoundTime:    session.Config.RoundTime,
		Vote:         session.Config.Vote,
		Code:         session.Config.Code,
		State:        session.State,
		CurrRoundIdx: session.CurrRoundIdx,
		CreatedAt:    session.CreatedAt,
		Categories:   make([]string, len(session.Config.Categories)),
		Letters:      make([]string, len(session.Config.Letters)),
		Bloopses:     make([]resource.Bloops, len(session.Config.Bloopses)),
		Players:      make([]*matchstateModel.Player, len(session.Players)),
	}

	copy(s.Categories, session.Config.Categories)
	copy(s.Letters, session.Config.Letters)
	copy(s.Bloopses, session.Config.Bloopses)
	copy(s.Players, session.Players)

	if err := m.stateDB.Add(s); err != nil {
		return fmt.Errorf("state db add: %w", err)
	}

	return nil
}

func (m *manager) restoreInterruptedGames() error {
	states, err := m.stateDB.FetchAll()
	if err != nil && !errors.Is(err, stateDB.ErrEntryNotFound) {
		return fmt.Errorf("stat db fetch all: %w", err)
	}

	m.mtx.Lock()
	for _, state := range states {
		session := NewMatchSessionFromSerialized(state, m.tg, m.matchDoneFn, m.matchWarnFn)
		session.Run(m.ctxSess)
		m.matchSessions[session.Config.Code] = session
		for _, player := range session.Players {
			if !player.Offline {
				m.userMatchSessions[player.UserID] = session
			}
		}
	}

	for _, session := range m.matchSessions {
		session.MoveState(session.State)
	}

	m.mtx.Unlock()

	if len(states) > 0 {
		if err := m.stateDB.Clean(); err != nil {
			if !errors.Is(err, stateDB.ErrBucketNotFound) {
				return fmt.Errorf("state db clean: %w", err)
			}
		}
	}

	return nil
}

func (m *manager) appendStat(session *match.Session) error {
	favorites := session.Favorites()
	stats := make([]statModel.Stat, 0)

	for _, player := range session.Players {
		stat := statModel.NewStat(player.UserID)
		if player.Offline {
			continue
		}

		for _, score := range favorites {
			if player.UserID == score.Player.UserID {
				stat.Conclusion = statModel.StatusFavorite
			}
		}

		stat.Categories = make([]string, len(session.Config.Categories))
		copy(stat.Categories, session.Config.Categories)

		stat.RoundsNum = session.Config.RoundsNum
		stat.PlayersNum = len(session.Players)

		var (
			bestDuration, worstDuration time.Duration = 2 << 31, 0
			sumDuration, durationNum    time.Duration = 0, 0
			bestPoints, worstPoints                   = 0, 2 << 28
			sumPoints, pointsNum                      = 0, 0
		)

		for _, rate := range player.Rates {
			if !rate.Bloops {
				durationNum += 1
				if rate.Duration < bestDuration {
					bestDuration = rate.Duration
				}
				if rate.Duration > worstDuration {
					worstDuration = rate.Duration
				}
				sumDuration += rate.Duration
			} else {
				stat.Bloops = append(stat.Bloops, rate.BloopsName)
			}

			pointsNum += 1
			if rate.Points > bestPoints {
				bestPoints = rate.Points
			}
			if rate.Points < worstPoints {
				worstPoints = rate.Points
			}
			sumPoints += rate.Points
		}

		stat.SumPoints = sumPoints
		stat.BestPoints = bestPoints
		stat.WorstPoints = worstPoints

		if pointsNum > 0 {
			stat.AveragePoints = sumPoints / pointsNum
		}

		stat.BestDuration = bestDuration
		stat.WorstDuration = worstDuration

		if durationNum > 0 {
			stat.AverageDuration = sumDuration / durationNum
		}

		stat.SumDuration = sumDuration
		stats = append(stats, stat)
	}

	for _, stat := range stats {
		if err := m.statDB.Add(stat); err != nil {
			return fmt.Errorf("stat db add: %w", err)
		}
	}

	return nil
}
