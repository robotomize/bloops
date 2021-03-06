package resource

import "github.com/enescakir/emoji"

// manage text messages
var (
	TextAuthorGreetingMsg = "\n\nТы - ведущий игрок " + emoji.FlexedBiceps.String() + "\n\n" +
		"Когда все игроки присоединятся тебе нужно нажать\n" + emoji.Rocket.String() + " *Начать* " + " для старта"
	TextJoinedGameMsg                      = "Ты присоединился к игре! "
	TextFeedbackMsg                        = "Ты можешь отправить анонимный отзыв"
	TextBanMsg                             = "Отправь username пользователя"
	TextGameRoomNotFoundMsg                = "Игровая комната не найдена"
	TextSendJoinedCodeMsg                  = "Отправь код подключения к игре"
	TextLeavingSessionsMsg                 = "Ты покинул все игровые сеансы"
	TextSendOfflinePlayerUsernameMsg       = "Отправь имя оффлайн пользователя"
	TextSendProfileMsg                     = "Отправь @username пользователя"
	TextBuilderWarnMsg                     = emoji.BrokenHeart.String() + " К сожалению " + emoji.Robot.String() + " бот обновляется, необходимо попробовать заново через несколько минут"
	TextMatchWarnMsg                       = emoji.BrokenHeart.String() + " К сожалению " + emoji.Robot.String() + " бот обновляется, этот раунд начнется заново через несколько секунд!"
	TextProfileCmdUserNotFound             = "Пользователь не найден"
	TextGameRoomNotFound                   = "Тебе нужно присоединиться к игре, чтобы добавлять оффлайн игроков"
	TextOfflinePlayerAdded                 = "Оффлайн игрок добавлен. Все сообщения будут приходить тебе"
	TextCreationGameCompletedSuccessfulMsg = emoji.Unicorn.String() + " Игровая комната создана.\n\nДля входа нужно " +
		"нажать кнопку " + emoji.VideoGame.String() + " *Присоединится к игре* и ввести этот код.\n\n" +
		emoji.PartyingFace.String() + " Отправь код тем, с кем собираешься играть"

	TextSettingsMsg = emoji.Gear.String() + " Настраиваем параметры игры"

	TextGreetingMsg = emoji.ChristmasTree.String() + emoji.ChristmasTree.String() + emoji.ChristmasTree.String() + "Привет, %s\n\n" +
		"Это " + `@blooops\_bot` + emoji.Robot.String() + " - бот, для игры в небольшие викторины, где участники должны за " + emoji.Stopwatch.String() + " 30 сек " +
		"назвать по одному слову из нескольких категорий, начинающихся на выпавшую букву\n\n" +
		"Бот" + emoji.Robot.String() + " предназначен для ведения игр в оффлайн" +
		" Он подсчитывает очки, генерирует буквы, создает лидерборды, и задает правила, а вы играете!" + emoji.Unicorn.String() + "\n\n" +
		"*Правила:* " + CmdRules + "\n\n" +
		"*Обратная связь:* @robotomize\n" +
		"*Проект на github:* [bloops_bot](https://github.com/robotomize/bloopsbot)"

	TextRulesMsg = emoji.Bookmark.String() + " *Правила игры*\n\n" +
		"Участники должны за " + emoji.Stopwatch.String() + " 30 сек " +
		"назвать по одному слову из нескольких категорий, начинающихся на выпавшую букву\n" +
		"По итогам нескольких раундов побеждают игроки с наибольшим количеством очков" + emoji.Trophy.String() + "\n\n" +
		emoji.CrossMark.String() + " *Ограничения* - от 2х человек, " + `@bloopsbot\_bot ` + emoji.Robot.String() + " предназначен для ведения игр в оффлайн\n\n" +
		emoji.Joystick.String() + " *Что делать?* - \nдля начала ведущий игрок должен " + emoji.Fire.String() + " *Создать игру* и выполнить действия по настройке." +
		" Ему будет выслан код, который он сообщает участникам. Затем игроки " +
		"присоединяются к игре и ведуший нажимает кнопку \n" + emoji.Rocket.String() + " *Начать*\n\n" +
		emoji.Loudspeaker.String() + " *Голосование* - после каждого раунда игроки определяют справился ли участник с заданием, если решили, что нет, то игрок не получает заработанные в раунде очки\n\n" +
		emoji.GemStone.String() + " *Блюпсы* - это дополнительные задания, " +
		"которые нужно выполнять параллельно с основным процессом игры, они выпадают игроку с некоторым шансом \n\n" +
		"*Список команд:* \n" +
		"/start - устанавливает бот и отправляет краткую справку по проекту\n" +
		"/rules - отправляет набор правил игры\n" +
		"/feedback - отправить анонимный отзыв\n" +
		"/profile - позволяет посмотреть профиль другого игрока\n" +
		"/add - если ты зашел в игровую команту, то можешь добавить игроков у которых нет телеграмма, так называемых виртуальных игроков, их задания будут приходить тебе. Ты можешь дать им свой смартфон, когда подойдет их очередь играть\n\n" +
		"*Обратная связь:* @robotomize\n" +
		"*Проект на github:* [bloops_bot](https://github.com/robotomize/bloopsbot)"
	TextChatNotAllowed = emoji.WomanGesturingNo.String() + " Бот не работает с групповыми чатами =("
)

// builder text messages
var (
	TextChooseCategories            = "Выбери категории или напиши свою"
	TextChooseRoundsNum             = "Выбери количество раундов(по умолчанию 1)"
	TextDeleteComplexLetters        = "Убери сложные буквы"
	TextVoteAllowed                 = emoji.Loudspeaker.String() + " Добавить голосование?\n\nПодробнее: /rules"
	TextBloopsAllowed               = emoji.GemStone.String() + " Добавить блюпсы?\n\nПодробнее: /rules"
	TextConfigurationDone           = "Завершить процесс создания игры?"
	TextAddLeastCategoryToComplete  = "Необходимо добавить больше категорий"
	TextAddLeastOneLetterToComplete = "Добавьте хотя бы одну букву для завершения"
	TextAddedCategory               = "Добавлена категория %s"
	TextDeletedCategory             = "Удалена категория %s"
	TextRoundsNumAnswer             = "Количество раундов - %d"
	TextAddedLetter                 = "Добавлена буква %s"
	TextDeletedLetter               = "Удалена буква %s"
	TextVoteYes                     = emoji.ThumbsUp.String() + " Да"
	TextVoteNo                      = emoji.ThumbsDown.String() + " Нет"
)

// match text messages
var (
	TextThumbUp                            = emoji.ThumbsUp.String()
	TextThumbDown                          = emoji.ThumbsDown.String()
	TextLeaderboardHeader                  = "*Результаты игры*\n\n"
	TextRoundFavoriteMsg                   = emoji.ChequeredFlag.String() + " Раунд %d завершен"
	TextClickStartBtnMsg                   = emoji.ChequeredFlag.String() + " Нажми кнопку, когда будешь готов"
	TextStartBtnData                       = "Я готов!"
	TextStopBtnData                        = "Стоп"
	TextStartBtnDataAnswer                 = "Старт!"
	TextChallengeBtnDataAnswer             = "Понятно"
	TextStopBtnDataAnswer                  = "Стоп!"
	TextTimerBtnData                       = "Таймер"
	TextStartLetterMsg                     = "Слова на букву - "
	TextNextPlayerMsg                      = "*%s* - твоя очередь"
	TextPlayerLeftGameMsg                  = "Игрок %s покинул игру"
	TextPlayerJoinedGameMsg                = "Игрок %s присоединился к игре"
	TextStopPlayerRoundMsg                 = "Завершено! Ты набрал %d очков!"
	TextGameStarted                        = "Игра началась!"
	TextValidationRequiresMoreOnePlayerMsg = "Чтобы начать игру необходимо как минимум %d игрока. Ты можешь добавить виртуального игрока командой /add \n\nПодробнее для чего нужна команда /add можно посмотреть в /rules"
	TextVoteMsg                            = "Голосование, игрок всё правильно назвал?"
	TextBroadcastCrashMsg                  = "Из-за ошибки в работе сервиса игра была аварийно завершена, попробуйте создать игру заново"
	TextStopButton                         = "Нажми Стоп, когда закончишь"
)
