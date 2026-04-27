package handlers

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"strings"
	"time"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"

	"taalkbout/internal/bot/client"
)

type App struct {
	log     *slog.Logger
	bot     *tgbotapi.BotAPI
	api     *client.Client
	store   *MemoryStore
	botUser string
}

func NewApp(log *slog.Logger, bot *tgbotapi.BotAPI, api *client.Client) *App {
	return &App{
		log:     log,
		bot:     bot,
		api:     api,
		store:   NewMemoryStore(),
		botUser: bot.Self.UserName,
	}
}

func (a *App) HandleUpdate(ctx context.Context, upd *tgbotapi.Update) {
	if upd == nil {
		return
	}

	switch {
	case upd.CallbackQuery != nil:
		a.handleCallback(ctx, upd.CallbackQuery)
	case upd.Message != nil:
		a.handleMessage(ctx, upd.Message)
	}
}

func (a *App) handleMessage(ctx context.Context, msg *tgbotapi.Message) {
	if msg == nil || msg.From == nil {
		return
	}

	sess := a.store.GetOrCreate(msg.From.ID, msg.Chat.ID)
	if err := a.ensureAPIUser(ctx, sess, msg.From); err != nil {
		a.sendText(msg.Chat.ID, "Сервис временно недоступен. Попробуй позже.", nil)
		return
	}

	if !msg.IsCommand() {
		txt := strings.TrimSpace(msg.Text)
		if txt == "" && msg.Caption != "" {
			txt = strings.TrimSpace(msg.Caption)
		}
		if strings.EqualFold(txt, "menu") || strings.EqualFold(txt, "меню") {
			a.cancelFlow(ctx, sess, msg.Chat.ID, 0)
			return
		}
	}

	if msg.IsCommand() {
		cmd := strings.TrimPrefix(msg.Command(), "/")
		arg := strings.TrimSpace(msg.CommandArguments())
		switch cmd {
		case "start":
			a.handleStart(ctx, sess, msg.Chat.ID, arg)
		case "menu":
			a.cancelFlow(ctx, sess, msg.Chat.ID, 0)
		case "topics":
			a.showTopics(ctx, sess, msg.Chat.ID, 0)
		case "focus":
			a.startFocusFromCommand(ctx, sess, msg.Chat.ID, 0)
		case "history":
			a.showHistory(ctx, sess, msg.Chat.ID, 0)
		case "notes":
			a.showNotes(ctx, sess, msg.Chat.ID, 0)
		case "settings":
			a.showSettings(ctx, sess, msg.Chat.ID, 0)
		case "testmode":
			a.showTestMode(ctx, sess, msg.Chat.ID, 0)
		default:
			a.sendText(msg.Chat.ID, "Не знаю эту команду. Нажми /menu.", nil)
		}
		return
	}

	// Text inside FSM
	a.handleFSMText(ctx, sess, msg)
}

func (a *App) handleCallback(ctx context.Context, q *tgbotapi.CallbackQuery) {
	if q == nil || q.From == nil {
		return
	}
	_ = a.answerCallback(q.ID)

	chatID := int64(0)
	messageID := 0
	if q.Message != nil {
		chatID = q.Message.Chat.ID
		messageID = q.Message.MessageID
	}
	if chatID == 0 {
		return
	}

	sess := a.store.GetOrCreate(q.From.ID, chatID)
	if err := a.ensureAPIUser(ctx, sess, q.From); err != nil {
		a.editOrSend(chatID, messageID, "Сервис временно недоступен. Попробуй позже.", nil)
		return
	}

	data := strings.TrimSpace(q.Data)
	switch {
	case data == "settings:rules":
		text := "Правила использования TalkaBot:\n\n1. Бот не определяет, кто прав.\n2. Бот не заменяет разговор.\n3. Бот не нужен для наказания партнера.\n4. Повторения — это фиксация восприятия и фактов, а не приговор.\n5. Несогласие обсуждается в фокусе разговора.\n6. В боте не устраиваем спор."
		kb := tgbotapi.NewInlineKeyboardMarkup(
			tgbotapi.NewInlineKeyboardRow(tgbotapi.NewInlineKeyboardButtonData("Назад", cbNavSettings)),
			tgbotapi.NewInlineKeyboardRow(tgbotapi.NewInlineKeyboardButtonData("В меню", cbMenu)),
		)
		a.editOrSend(chatID, messageID, text, &kb)
	case data == "settings:pair":
		a.showPairSpaces(ctx, sess, chatID, messageID)
	case data == cbSettingsShare:
		a.sharePairLink(ctx, sess, chatID, messageID)
	case data == cbPairWelcome:
		sess.State = StatePairWelcome
		a.editOrSend(chatID, messageID, "Текст приветствия для партнера (появится после подключения):", backKeyboard(cbSettingsShare))
	case data == cbPairMyName:
		sess.State = StatePairMyName
		a.editOrSend(chatID, messageID, "Твое имя (будет использоваться в фокусе и истории):", backKeyboard(cbSettingsShare))
	case strings.HasPrefix(data, "issue:prio:"):
		val := strings.TrimPrefix(data, "issue:prio:")
		sess.AddIssuePriority = val
		sess.State = StateAddIssueVisibility
		a.editOrSend(chatID, messageID, "Выбери видимость темы.", visibilityKeyboard())
	case strings.HasPrefix(data, "issue:vis:"):
		val := strings.TrimPrefix(data, "issue:vis:")
		sess.AddIssueVisibility = val
		if val == "hidden_until_repeats" {
			sess.State = StateAddIssueThreshold
			a.editOrSend(chatID, messageID, "После скольких повторений открыть тему?", thresholdKeyboard())
			break
		}
		sess.AddIssueThreshold = 0
		a.finishAddIssue(ctx, sess, chatID, messageID)
	case strings.HasPrefix(data, "issue:thr:"):
		val := strings.TrimPrefix(data, "issue:thr:")
		switch val {
		case "1":
			sess.AddIssueThreshold = 1
		case "2":
			sess.AddIssueThreshold = 2
		case "3":
			sess.AddIssueThreshold = 3
		case "5":
			sess.AddIssueThreshold = 5
		default:
			sess.AddIssueThreshold = 1
		}
		a.finishAddIssue(ctx, sess, chatID, messageID)
	case data == "issue:back_desc":
		sess.State = StateAddIssueDescription
		a.editOrSend(chatID, messageID, "Опиши ситуацию подробнее.", backCancelKeyboard("issue:back_title"))
	case data == "issue:back_title":
		sess.State = StateAddIssueTitle
		a.editOrSend(chatID, messageID, "Название темы?", cancelKeyboard())
	case data == "issue:back_prio":
		sess.State = StateAddIssuePriority
		a.editOrSend(chatID, messageID, "Выбери важность.", priorityKeyboard())
	case data == "issue:back_vis":
		sess.State = StateAddIssueVisibility
		a.editOrSend(chatID, messageID, "Выбери видимость темы.", visibilityKeyboard())
	case data == "focus:back_questions":
		sess.State = StateFocusQuestions
		a.editOrSend(chatID, messageID, "Главные вопросы разговора.", backKeyboard(cbFocusBackToGoal))
	case data == cbMenu:
		a.showMenu(ctx, sess, chatID, messageID)
	case data == cbNavFocus:
		a.startFocusFromCommand(ctx, sess, chatID, messageID)
	case data == cbNavNotes:
		a.showNotes(ctx, sess, chatID, messageID)
	case data == cbNavHistory:
		a.showHistory(ctx, sess, chatID, messageID)
	case data == cbNavSettings:
		a.showSettings(ctx, sess, chatID, messageID)
	case data == cbHowTo:
		a.showHowTo(chatID, messageID)
	case data == cbPairCreate:
		a.createPair(ctx, sess, chatID, messageID)
	case data == cbPairCheck:
		a.checkPair(ctx, sess, chatID, messageID)
	case strings.HasPrefix(data, cbInviteConfirmPrefix):
		token := strings.TrimPrefix(data, cbInviteConfirmPrefix)
		a.confirmJoinInvite(ctx, sess, chatID, messageID, token)
	case data == cbInviteCancel:
		sess.State = StateIdle
		sess.PendingInviteToken = ""
		a.showMenu(ctx, sess, chatID, messageID)
	case data == cbTestStart:
		a.startTestMode(ctx, sess, chatID, messageID)
	case data == cbTestStop:
		a.stopTestMode(ctx, sess, chatID, messageID)
	case data == cbTopics:
		a.showTopics(ctx, sess, chatID, messageID)
	case data == cbTopicsA:
		a.showTopicsList(ctx, sess, chatID, messageID, "active")
	case data == cbTopicsRep:
		a.showTopicsRepeats(ctx, sess, chatID, messageID)
	case data == cbTopicsR:
		a.showTopicsList(ctx, sess, chatID, messageID, "resolved")
	case data == cbTopicsSearch:
		sess.State = StateTopicsSearch
		a.editOrSend(chatID, messageID, "Поиск по темам: напиши запрос.", backKeyboard(cbTopics))
	case data == cbIssueNew:
		a.startAddIssue(ctx, sess, chatID, messageID)
	case data == cbIssueQuickYes:
		if err := a.refreshCurrentPair(ctx, sess); err != nil || sess.CurrentPairID == nil {
			a.editOrSend(chatID, messageID, "Чтобы создать тему, сначала создай пару или включи тестовый режим.", noPairKeyboard())
			return
		}
		title := strings.TrimSpace(sess.QuickIssueTitle)
		if title == "" {
			a.editOrSend(chatID, messageID, "Напиши название темы текстом.", nil)
			return
		}
		sess.State = StateAddIssueDescription
		sess.AddIssueTitle = title
		sess.AddIssueDescription = ""
		sess.AddIssuePriority = ""
		sess.AddIssueVisibility = ""
		sess.AddIssueThreshold = 0
		sess.QuickIssueTitle = ""
		a.editOrSend(chatID, messageID, "Опиши ситуацию подробнее.\n\nТема: "+title, backCancelKeyboard("issue:back_title"))
	case data == cbIssueQuickNo:
		sess.QuickIssueTitle = ""
		a.showMenu(ctx, sess, chatID, messageID)
	case strings.HasPrefix(data, cbIssueOpenPrefix):
		issueID := strings.TrimPrefix(data, cbIssueOpenPrefix)
		a.showIssue(ctx, sess, chatID, messageID, issueID)
	case strings.HasPrefix(data, cbIssueRepeatsPrefix):
		issueID := strings.TrimPrefix(data, cbIssueRepeatsPrefix)
		a.showRepeats(ctx, sess, chatID, messageID, issueID)
	case strings.HasPrefix(data, cbRepeatOpenPrefix):
		repeatID := strings.TrimPrefix(data, cbRepeatOpenPrefix)
		a.showRepeat(ctx, sess, chatID, messageID, repeatID)
	case strings.HasPrefix(data, cbRepeatDisagreePrefix):
		repeatID := strings.TrimPrefix(data, cbRepeatDisagreePrefix)
		sess.State = StateAddRepeatDisagreeNote
		sess.CurrentRepeatID = repeatID
		a.editOrSend(chatID, messageID, "Оставь заметку с позицией.", cancelKeyboard())
	case strings.HasPrefix(data, cbRepeatDeletePrefix):
		repeatID := strings.TrimPrefix(data, cbRepeatDeletePrefix)
		a.editOrSend(chatID, messageID, "Удалить повторение?", repeatDeleteConfirmKeyboard(repeatID))
	case strings.HasPrefix(data, cbRepeatDeleteYesPrefix):
		repeatID := strings.TrimPrefix(data, cbRepeatDeleteYesPrefix)
		rep, err := a.api.GetRepeat(ctx, repeatID)
		if err != nil {
			a.editOrSend(chatID, messageID, "Не получилось удалить повторение.", nil)
			return
		}
		if err := a.api.DeleteRepeat(ctx, repeatID); err != nil {
			a.editOrSend(chatID, messageID, "Не получилось удалить повторение.", nil)
			return
		}
		a.showRepeats(ctx, sess, chatID, messageID, rep.IssueID)
	case strings.HasPrefix(data, cbIssueRepeatPrefix):
		issueID := strings.TrimPrefix(data, cbIssueRepeatPrefix)
		sess.State = StateAddRepeatNote
		sess.CurrentIssueID = issueID
		a.editOrSend(chatID, messageID, "Короткая заметка о повторении (или перешли сообщение).", cancelKeyboard())
	case strings.HasPrefix(data, cbIssueClosePrefix):
		issueID := strings.TrimPrefix(data, cbIssueClosePrefix)
		a.confirmCloseIssue(ctx, sess, chatID, messageID, issueID)
	case strings.HasPrefix(data, cbIssueResolveConfirmPrefix):
		issueID := strings.TrimPrefix(data, cbIssueResolveConfirmPrefix)
		a.updateIssueStatus(ctx, sess, chatID, messageID, issueID, "resolved")
	case strings.HasPrefix(data, cbIssueRestorePrefix):
		issueID := strings.TrimPrefix(data, cbIssueRestorePrefix)
		a.updateIssueStatus(ctx, sess, chatID, messageID, issueID, "active")
	case strings.HasPrefix(data, cbIssueHistoryPrefix):
		issueID := strings.TrimPrefix(data, cbIssueHistoryPrefix)
		a.showIssueHistory(ctx, sess, chatID, messageID, issueID)
	case strings.HasPrefix(data, cbIssueDeletePrefix):
		issueID := strings.TrimPrefix(data, cbIssueDeletePrefix)
		a.editOrSend(chatID, messageID, "Удалить тему? Это удалит тему и все связанные данные.", issueDeleteConfirmKeyboard(issueID))
	case strings.HasPrefix(data, cbIssueDeleteYesPrefix):
		issueID := strings.TrimPrefix(data, cbIssueDeleteYesPrefix)
		if err := a.api.DeleteIssue(ctx, issueID); err != nil {
			a.editOrSend(chatID, messageID, "Не получилось удалить тему. Попробуй позже.", nil)
			return
		}
		a.editOrSend(chatID, messageID, "Тема удалена.", backToMenuKeyboard())
	case strings.HasPrefix(data, cbIssueSettingsPrefix):
		issueID := strings.TrimPrefix(data, cbIssueSettingsPrefix)
		a.showIssueSettings(ctx, sess, chatID, messageID, issueID)
	case strings.HasPrefix(data, cbIssueRenamePrefix):
		issueID := strings.TrimPrefix(data, cbIssueRenamePrefix)
		sess.State = StateIssueRename
		sess.CurrentIssueID = issueID
		a.editOrSend(chatID, messageID, "Новое название темы:", backKeyboard(cbIssueSettingsPrefix+issueID))
	case strings.HasPrefix(data, cbIssueRepeatLimitPrefix):
		issueID := strings.TrimPrefix(data, cbIssueRepeatLimitPrefix)
		sess.State = StateIssueRepeatLimit
		sess.CurrentIssueID = issueID
		a.editOrSend(chatID, messageID, "Новый лимит повторений (число):", backKeyboard(cbIssueSettingsPrefix+issueID))
	case strings.HasPrefix(data, cbIssueShowLimitPrefix):
		issueID := strings.TrimPrefix(data, cbIssueShowLimitPrefix)
		sess.State = StateIssueShowLimit
		sess.CurrentIssueID = issueID
		a.editOrSend(chatID, messageID, "Новый лимит показа (после повторений, число):", backKeyboard(cbIssueSettingsPrefix+issueID))
	case strings.HasPrefix(data, cbFocusStartPrefix):
		issueID := strings.TrimPrefix(data, cbFocusStartPrefix)
		a.startFocusForIssue(ctx, sess, chatID, messageID, issueID)
	case strings.HasPrefix(data, cbFocusStartStateSelfPrefix):
		code := strings.TrimPrefix(data, cbFocusStartStateSelfPrefix)
		sess.FocusStartStateSelf = stateLabel(code)
		sess.State = StateFocusStartStatePartner
		_, partnerName := a.pairNames(ctx, sess)
		a.editOrSend(chatID, messageID, "Состояние перед разговором.\n\nСостояние "+partnerName+":", stateKeyboard(cbFocusStartStatePartnerPrefix, cbFocusBackToSelfState))
	case strings.HasPrefix(data, cbFocusStartStatePartnerPrefix):
		code := strings.TrimPrefix(data, cbFocusStartStatePartnerPrefix)
		sess.FocusStartStatePartner = stateLabel(code)
		sess.State = StateFocusRuleLimit
		a.editOrSend(chatID, messageID, "Лимит нарушений правил (0 = без лимита):", focusRuleLimitKeyboard())
	case strings.HasPrefix(data, cbFocusRuleLimitPrefix):
		val := strings.TrimPrefix(data, cbFocusRuleLimitPrefix)
		limit, ok := parsePositiveInt(val)
		if !ok {
			a.editOrSend(chatID, messageID, "Нужен формат числа.", focusRuleLimitKeyboard())
			return
		}
		sess.FocusRuleLimit = limit
		sess.State = StateFocusPlan
		a.showFocusPlan(ctx, sess, chatID, messageID)
	case data == cbFocusBegin:
		a.beginConversation(ctx, sess, chatID, messageID)
	case data == cbFocusPlan:
		sess.State = StateFocusPlan
		a.showFocusPlan(ctx, sess, chatID, messageID)
	case data == cbFocusBackToSelfState:
		sess.State = StateFocusStartStateSelf
		a.editOrSend(chatID, messageID, "Состояние перед разговором.\n\nТвое состояние:", stateKeyboard(cbFocusStartStateSelfPrefix, "focus:back_questions"))
	case data == cbFocusBackToSelect:
		sess.FocusIssueID = ""
		sess.FocusGoal = ""
		sess.FocusQuestions = ""
		sess.FocusStartStateSelf = ""
		sess.FocusStartStatePartner = ""
		a.startFocusFromCommand(ctx, sess, chatID, messageID)
	case data == cbFocusBackToPartnerState:
		sess.State = StateFocusStartStatePartner
		_, partnerName := a.pairNames(ctx, sess)
		a.editOrSend(chatID, messageID, "Состояние перед разговором.\n\nСостояние "+partnerName+":", stateKeyboard(cbFocusStartStatePartnerPrefix, cbFocusBackToSelfState))
	case strings.HasPrefix(data, cbConvPausePrefix):
		id := strings.TrimPrefix(data, cbConvPausePrefix)
		a.pauseConversation(ctx, sess, chatID, messageID, id)
	case strings.HasPrefix(data, cbConvResumePrefix):
		id := strings.TrimPrefix(data, cbConvResumePrefix)
		a.resumeConversation(ctx, sess, chatID, messageID, id)
	case strings.HasPrefix(data, cbConvNotePrefix):
		id := strings.TrimPrefix(data, cbConvNotePrefix)
		sess.State = StateConvNote
		sess.ConversationID = id
		a.editOrSend(chatID, messageID, "Запиши мысль одним сообщением.", backKeyboard(cbConvBackToActive))
	case strings.HasPrefix(data, cbConvSidePrefix):
		id := strings.TrimPrefix(data, cbConvSidePrefix)
		sess.State = StateConvSideTitle
		sess.ConversationID = id
		a.editOrSend(chatID, messageID, "Побочная тема: напиши короткое название.", backKeyboard(cbConvBackToActive))
	case strings.HasPrefix(data, cbConvRepeatsPrefix):
		id := strings.TrimPrefix(data, cbConvRepeatsPrefix)
		a.showConversationRepeats(ctx, sess, chatID, messageID, id)
	case strings.HasPrefix(data, cbConvRulesPrefix):
		id := strings.TrimPrefix(data, cbConvRulesPrefix)
		sess.ConversationID = id
		text := "Правила разговора:\n\n" + focusRulesText()
		kb := tgbotapi.NewInlineKeyboardMarkup(
			tgbotapi.NewInlineKeyboardRow(tgbotapi.NewInlineKeyboardButtonData("Назад", cbConvBackToActive)),
		)
		a.editOrSend(chatID, messageID, text, &kb)
	case strings.HasPrefix(data, cbConvViolationPrefix):
		id := strings.TrimPrefix(data, cbConvViolationPrefix)
		sess.ConversationID = id
		sess.PendingRuleCode = ""
		a.editOrSend(chatID, messageID, "Какое правило нарушено?", convViolationRulesKeyboard())
	case strings.HasPrefix(data, cbConvViolationPickPrefix):
		code := strings.TrimPrefix(data, cbConvViolationPickPrefix)
		sess.PendingRuleCode = code
		sess.State = StateConvViolationNote
		a.editOrSend(chatID, messageID, "Пояснение (одним сообщением):", backKeyboard(cbConvBackToActive))
	case strings.HasPrefix(data, cbConvFinishPrefix):
		id := strings.TrimPrefix(data, cbConvFinishPrefix)
		sess.State = StateFinishStatus
		sess.ConversationID = id
		a.editOrSend(chatID, messageID, "Какой итог?", finishStatusKeyboard())
	case strings.HasPrefix(data, cbConvEarlyPrefix):
		id := strings.TrimPrefix(data, cbConvEarlyPrefix)
		sess.State = StateIdle
		sess.ConversationID = id
		sess.EarlyEndReason = ""
		sess.EarlyEndInitiative = ""
		kb := tgbotapi.NewInlineKeyboardMarkup(
			tgbotapi.NewInlineKeyboardRow(
				tgbotapi.NewInlineKeyboardButtonData("Нет времени", cbConvEarlyReason+"no_time"),
				tgbotapi.NewInlineKeyboardButtonData("Нужна пауза", cbConvEarlyReason+"need_break"),
			),
			tgbotapi.NewInlineKeyboardRow(
				tgbotapi.NewInlineKeyboardButtonData("Слишком эмоционально", cbConvEarlyReason+"too_emotional"),
			),
			tgbotapi.NewInlineKeyboardRow(
				tgbotapi.NewInlineKeyboardButtonData("Другое…", cbConvEarlyReason+"other"),
			),
			tgbotapi.NewInlineKeyboardRow(
				tgbotapi.NewInlineKeyboardButtonData("Назад", cbConvBackToActive),
			),
		)
		a.editOrSend(chatID, messageID, "Досрочное завершение.\n\nПочему завершаем?", &kb)
	case strings.HasPrefix(data, cbConvEarlyReason):
		code := strings.TrimPrefix(data, cbConvEarlyReason)
		switch code {
		case "other":
			sess.State = StateConvEarlyOther
			a.editOrSend(chatID, messageID, "Почему? Напиши одним сообщением.", backKeyboard(cbConvBackToActive))
		default:
			sess.EarlyEndReason = code
			a.editOrSend(chatID, messageID, "По чьей инициативе завершаем?", earlyInitiativeKeyboard())
		}
	case strings.HasPrefix(data, cbConvEarlyInit):
		code := strings.TrimPrefix(data, cbConvEarlyInit)
		sess.EarlyEndInitiative = strings.TrimSpace(code)
		a.finishConversationEarly(ctx, sess, chatID, messageID)
	case strings.HasPrefix(data, cbFinishStatusPrefix):
		val := strings.TrimPrefix(data, cbFinishStatusPrefix)
		sess.FinishResultStatus = val
		sess.FinishEndStateSelf = ""
		sess.FinishEndStatePartner = ""
		sess.State = StateFinishEndStateSelf
		a.editOrSend(chatID, messageID, "Состояние в конце разговора.\n\nТвое состояние:", stateKeyboard(cbFinishEndStateSelfPrefix, cbFinishBackToStatus))
	case strings.HasPrefix(data, cbFinishEndStateSelfPrefix):
		code := strings.TrimPrefix(data, cbFinishEndStateSelfPrefix)
		sess.FinishEndStateSelf = stateLabel(code)
		sess.State = StateFinishEndStatePartner
		_, partnerName := a.pairNames(ctx, sess)
		a.editOrSend(chatID, messageID, "Состояние в конце разговора.\n\nСостояние "+partnerName+":", stateKeyboard(cbFinishEndStatePartnerPrefix, cbFinishBackToEndSelf))
	case strings.HasPrefix(data, cbFinishEndStatePartnerPrefix):
		code := strings.TrimPrefix(data, cbFinishEndStatePartnerPrefix)
		sess.FinishEndStatePartner = stateLabel(code)
		sess.State = StateFinishText
		a.editOrSend(chatID, messageID, "Что зафиксировать как итог?", backKeyboard(cbFinishBackToStatus))
	case strings.HasPrefix(data, cbHistoryOpenPrefix):
		id := strings.TrimPrefix(data, cbHistoryOpenPrefix)
		a.openHistoryConversation(ctx, sess, chatID, messageID, id)
	case strings.HasPrefix(data, cbNoteDeletePrefix):
		noteID := strings.TrimPrefix(data, cbNoteDeletePrefix)
		a.editOrSend(chatID, messageID, "Удалить заметку?", noteDeleteConfirmKeyboard(noteID))
	case strings.HasPrefix(data, cbNoteDeleteYesPrefix):
		noteID := strings.TrimPrefix(data, cbNoteDeleteYesPrefix)
		if err := a.api.DeleteConversationNote(ctx, noteID); err != nil {
			a.editOrSend(chatID, messageID, "Не получилось удалить заметку.", nil)
			return
		}
		a.showNotes(ctx, sess, chatID, messageID)
	case strings.HasPrefix(data, cbPrefSwitchPrefix):
		pairID := strings.TrimPrefix(data, cbPrefSwitchPrefix)
		a.switchCurrentPair(ctx, sess, chatID, messageID, pairID)
	case data == cbBackToMenu:
		a.showMenu(ctx, sess, chatID, messageID)
	case data == cbCancel:
		a.cancelFlow(ctx, sess, chatID, messageID)
	case data == cbFocusBackToGoal:
		sess.State = StateFocusGoal
		a.askFocusGoal(chatID, messageID)
	case data == cbConvBackToActive:
		a.showActiveConversation(ctx, sess, chatID, messageID)
	case data == cbFinishBackToStatus:
		sess.State = StateFinishStatus
		a.editOrSend(chatID, messageID, "Какой итог?", finishStatusKeyboard())
	case data == cbFinishBackToEndSelf:
		sess.State = StateFinishEndStateSelf
		a.editOrSend(chatID, messageID, "Состояние в конце разговора.\n\nТвое состояние:", stateKeyboard(cbFinishEndStateSelfPrefix, cbFinishBackToStatus))
	default:
		a.editOrSend(chatID, messageID, "Не понял действие. Нажми /menu.", nil)
	}
}

func (a *App) ensureAPIUser(ctx context.Context, sess *Session, tg *tgbotapi.User) error {
	if sess.APIUserID != "" {
		return nil
	}
	display := strings.TrimSpace(strings.Join([]string{tg.FirstName, tg.LastName}, " "))
	if display == "" {
		display = tg.UserName
	}

	var username *string
	if strings.TrimSpace(tg.UserName) != "" {
		v := strings.TrimSpace(tg.UserName)
		username = &v
	}
	var displayName *string
	if strings.TrimSpace(display) != "" {
		v := display
		displayName = &v
	}

	u, err := a.api.UpsertTelegramUser(ctx, client.UpsertTelegramUserRequest{
		TelegramID:  tg.ID,
		Username:    username,
		DisplayName: displayName,
	})
	if err != nil {
		a.log.Error("api upsert telegram user failed", slog.Any("err", err))
		return err
	}
	sess.APIUserID = u.ID
	return nil
}

func (a *App) refreshCurrentPair(ctx context.Context, sess *Session) error {
	pref, err := a.api.GetPreferences(ctx, sess.APIUserID)
	if err != nil {
		return err
	}
	sess.CurrentPairID = pref.CurrentPairID
	sess.CurrentPairIsTest = false
	if pref.CurrentPairID != nil {
		p, _, err := a.api.GetPair(ctx, *pref.CurrentPairID)
		if err != nil {
			return err
		}
		sess.CurrentPairIsTest = p.IsTest
	}
	return nil
}

func (a *App) sendText(chatID int64, text string, kb *tgbotapi.InlineKeyboardMarkup) {
	msg := tgbotapi.NewMessage(chatID, text)
	if kb != nil {
		msg.ReplyMarkup = kb
	}
	_, _ = a.bot.Send(msg)
}

func (a *App) editOrSend(chatID int64, messageID int, text string, kb *tgbotapi.InlineKeyboardMarkup) {
	if messageID > 0 {
		edit := tgbotapi.NewEditMessageText(chatID, messageID, text)
		if kb != nil {
			edit.ReplyMarkup = kb
		}
		_, err := a.bot.Send(edit)
		if err == nil {
			return
		}
	}
	a.sendText(chatID, text, kb)
}

func (a *App) answerCallback(id string) error {
	_, err := a.bot.Request(tgbotapi.NewCallback(id, ""))
	if err != nil && !errors.Is(err, context.Canceled) {
		return err
	}
	return nil
}

func formatDuration(d time.Duration) string {
	if d < 0 {
		d = 0
	}
	m := int(d.Minutes())
	s := int(d.Seconds()) % 60
	return fmt.Sprintf("%02d:%02d", m, s)
}
