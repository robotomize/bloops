package resource

import (
	"fmt"

	"github.com/bloops-games/bloops/internal/hashutil"
	"github.com/enescakir/emoji"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
)

var (
	// common menu button text
	CreateButtonText      = emoji.Fire.String() + " Создать игру"
	LeaveButtonText       = emoji.ChequeredFlag.String() + " Выйти"
	StartButtonText       = emoji.Rocket.String() + " Начать"
	JoinButtonText        = emoji.VideoGame.String() + " Присоединиться к игре"
	RatingButtonText      = emoji.Star.String() + " Таблица лидеров"
	RuleButtonText        = "Правила"
	GameSettingButtonText = "Параметры игы"
	ProfileButtonText     = emoji.Alien.String() + " Профиль"

	// builder inline button text
	BuilderInlineNextText = "Далее"
	BuilderInlineNextData = fmt.Sprintf("%s:%s", BuilderInlineNextText, hashutil.SerializedSha1FromTime())
	BuilderInlinePrevText = "Назад"
	BuilderInlinePrevData = fmt.Sprintf("%s:%s", BuilderInlinePrevText, hashutil.SerializedSha1FromTime())
	BuilderInlineDoneText = emoji.ChequeredFlag.String() + " Завершить"
	BuilderInlineDoneData = fmt.Sprintf("%s:%s", BuilderInlineDoneText, hashutil.SerializedSha1FromTime())
)

var (
	// keyboard buttons
	CreateButton      = tgbotapi.NewKeyboardButton(CreateButtonText)
	JoinButton        = tgbotapi.NewKeyboardButton(JoinButtonText)
	LeaveButton       = tgbotapi.NewKeyboardButton(LeaveButtonText)
	StartButton       = tgbotapi.NewKeyboardButton(StartButtonText)
	RatingButton      = tgbotapi.NewKeyboardButton(RatingButtonText)
	RulesButton       = tgbotapi.NewKeyboardButton(RuleButtonText)
	ProfileButton     = tgbotapi.NewKeyboardButton(ProfileButtonText)
	GameSettingButton = tgbotapi.NewKeyboardButton(GameSettingButtonText)

	CommonButtons = tgbotapi.NewReplyKeyboard(
		tgbotapi.NewKeyboardButtonRow(CreateButton),
		tgbotapi.NewKeyboardButtonRow(JoinButton),
		tgbotapi.NewKeyboardButtonRow(RulesButton, ProfileButton),
	)
	LeaveMenuButton = tgbotapi.NewKeyboardButton(LeaveButtonText)
)
