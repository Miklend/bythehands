package handlers

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"time"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"

	"taalkbout/internal/bot/client"
)

func (a *App) handleStart(ctx context.Context, sess *Session, chatID int64, startArg string) {
	onboarding := `TalkaBot — инструмент для пары.

Он помогает:
• фиксировать важные темы
• отмечать повторения ситуаций
• готовиться к разговору
• завершать разговор итогом

Важно:
Бот не решает, кто прав.
Бот хранит факты и структуру.
Использование зависит только от вас.`

	startArg = strings.TrimSpace(startArg)
	if startArg != "" {
		sess.PendingInviteToken = startArg
		text := "Ты присоединяешься к паре.\nПодтвердить?"
		a.sendText(chatID, text, inviteConfirmKeyboard(startArg))
		return
	}

	sess.State = StateIdle
	_ = onboarding // onboarding оставим в /howto; по /start сразу показываем меню
	a.showMenu(ctx, sess, chatID, 0)
}

func (a *App) showHowTo(chatID int64, messageID int) {
	text := "Как пользоваться:\n\n1) /start — начало\n2) Создай пару и отправь ссылку партнеру\n3) Открой меню (кнопка «Главное меню» или напиши «menu»)\n4) Создавайте темы, отмечайте повторения, начинайте фокус\n\nЕсли у тебя есть invite-ссылка, просто открой её."
	kb := tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(tgbotapi.NewInlineKeyboardButtonData("Назад", cbNavSettings)),
		tgbotapi.NewInlineKeyboardRow(tgbotapi.NewInlineKeyboardButtonData("В меню", cbMenu)),
	)
	a.editOrSend(chatID, messageID, text, &kb)
}

func (a *App) showMenu(ctx context.Context, sess *Session, chatID int64, messageID int) {
	if err := a.refreshCurrentPair(ctx, sess); err != nil {
		a.editOrSend(chatID, messageID, "Сервис временно недоступен. Попробуй позже.", nil)
		return
	}

	if sess.CurrentPairID == nil {
		a.editOrSend(chatID, messageID, "Сначала создай пару, войди по ссылке или включи тестовый режим.", noPairKeyboard())
		return
	}

	header := "Главное меню"
	if sess.CurrentPairIsTest {
		header = "Главное меню\n\n(Тестовый режим активен)"
	}
	a.editOrSend(chatID, messageID, header, menuKeyboard())
}

func (a *App) createPair(ctx context.Context, sess *Session, chatID int64, messageID int) {
	res, err := a.api.CreatePair(ctx, sess.APIUserID)
	if err != nil {
		a.editOrSend(chatID, messageID, "Не получилось создать пару. Попробуй позже.", backMenuKeyboard())
		return
	}
	_, _ = a.api.SetPreferences(ctx, sess.APIUserID, &res.Pair.ID)
	_ = a.refreshCurrentPair(ctx, sess)

	link := fmt.Sprintf("https://t.me/%s?start=%s", a.botUser, res.InviteToken)
	text := fmt.Sprintf("Пара создана.\nОтправь эту ссылку партнеру.\nПосле подключения у вас появится общее пространство.\n\n%s", link)
	a.editOrSend(chatID, messageID, text, pairCreatedKeyboard())
}

func (a *App) checkPair(ctx context.Context, sess *Session, chatID int64, messageID int) {
	if err := a.refreshCurrentPair(ctx, sess); err != nil || sess.CurrentPairID == nil {
		a.editOrSend(chatID, messageID, "Пары пока нет. Создай пару или включи тестовый режим.", noPairKeyboard())
		return
	}
	_, members, err := a.api.GetPair(ctx, *sess.CurrentPairID)
	if err != nil {
		a.editOrSend(chatID, messageID, "Не получилось проверить пару. Попробуй позже.", backMenuKeyboard())
		return
	}
	if len(members) >= 2 {
		a.showMenu(ctx, sess, chatID, messageID)
		return
	}
	a.editOrSend(chatID, messageID, "Пока в паре только ты. Отправь invite-ссылку партнеру и нажми «Проверить подключение».", pairCreatedKeyboard())
}

func (a *App) confirmJoinInvite(ctx context.Context, sess *Session, chatID int64, messageID int, token string) {
	p, _, err := a.api.JoinInvite(ctx, token, sess.APIUserID)
	if err != nil {
		a.editOrSend(chatID, messageID, "Не получилось присоединиться по ссылке. Проверь ссылку или попробуй позже.", backMenuKeyboard())
		return
	}
	_, _ = a.api.SetPreferences(ctx, sess.APIUserID, &p.ID)
	_ = a.refreshCurrentPair(ctx, sess)
	a.editOrSend(chatID, messageID, "Готово. Ты присоединился к паре.", backToMenuKeyboard())
	if p.WelcomeMessage != nil {
		if t := strings.TrimSpace(*p.WelcomeMessage); t != "" {
			a.sendText(chatID, t, backToMenuKeyboard())
		}
	}
	a.showMenu(ctx, sess, chatID, 0)
}

func (a *App) showTopics(ctx context.Context, sess *Session, chatID int64, messageID int) {
	if err := a.refreshCurrentPair(ctx, sess); err != nil || sess.CurrentPairID == nil {
		a.editOrSend(chatID, messageID, "Сначала создай пару или включи тестовый режим.", noPairKeyboard())
		return
	}

	active, err := a.api.ListIssues(ctx, *sess.CurrentPairID, "active")
	if err != nil {
		a.editOrSend(chatID, messageID, "Не получилось загрузить темы. Попробуй позже.", backMenuKeyboard())
		return
	}
	resolved, err := a.api.ListIssues(ctx, *sess.CurrentPairID, "resolved")
	if err != nil {
		a.editOrSend(chatID, messageID, "Не получилось загрузить темы. Попробуй позже.", backMenuKeyboard())
		return
	}
	repeats := 0
	for _, it := range active {
		if it.RepeatCount > 0 {
			repeats++
		}
	}

	kb := tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData(fmt.Sprintf("Активные (%d)", len(active)), cbTopicsA),
			tgbotapi.NewInlineKeyboardButtonData(fmt.Sprintf("На повторе (%d)", repeats), cbTopicsRep),
			tgbotapi.NewInlineKeyboardButtonData(fmt.Sprintf("Решённые (%d)", len(resolved)), cbTopicsR),
		),
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("➕ Новая тема", cbIssueNew),
		),
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("Назад", cbMenu),
		),
	)
	a.editOrSend(chatID, messageID, "Темы: выбери список.", &kb)
}

func (a *App) showTopicsList(ctx context.Context, sess *Session, chatID int64, messageID int, status string) {
	if err := a.refreshCurrentPair(ctx, sess); err != nil || sess.CurrentPairID == nil {
		a.editOrSend(chatID, messageID, "Сначала создай пару или включи тестовый режим.", noPairKeyboard())
		return
	}

	items, err := a.api.ListIssues(ctx, *sess.CurrentPairID, status)
	if err != nil {
		a.editOrSend(chatID, messageID, "Не получилось загрузить темы. Попробуй позже.", backMenuKeyboard())
		return
	}

	header := "Темы"
	switch status {
	case "active":
		header = "Активные темы"
	case "resolved":
		header = "Решённые темы"
	}

	if len(items) == 0 {
		var rows [][]tgbotapi.InlineKeyboardButton
		if status == "active" {
			rows = append(rows, tgbotapi.NewInlineKeyboardRow(tgbotapi.NewInlineKeyboardButtonData("➕ Новая тема", cbIssueNew)))
		}
		rows = append(rows, tgbotapi.NewInlineKeyboardRow(tgbotapi.NewInlineKeyboardButtonData("Назад", cbTopics)))
		kb := tgbotapi.NewInlineKeyboardMarkup(rows...)
		a.editOrSend(chatID, messageID, header+": пока пусто.", &kb)
		return
	}

	var b strings.Builder
	b.WriteString(header + ":\n\n")
	for i, it := range items {
		fmt.Fprintf(&b, "%d. %s\nВажность: %s\nПовторений: %d\n\n", i+1, it.Title, issuePriorityLabel(it.Priority), it.RepeatCount)
	}

	var rows [][]tgbotapi.InlineKeyboardButton
	for _, it := range items {
		rows = append(rows, tgbotapi.NewInlineKeyboardRow(tgbotapi.NewInlineKeyboardButtonData(it.Title, cbIssueOpenPrefix+it.ID)))
	}
	rows = append(rows, tgbotapi.NewInlineKeyboardRow(tgbotapi.NewInlineKeyboardButtonData("Назад", cbTopics)))
	kb := tgbotapi.NewInlineKeyboardMarkup(rows...)
	a.editOrSend(chatID, messageID, b.String(), &kb)
}

func (a *App) showTopicsRepeats(ctx context.Context, sess *Session, chatID int64, messageID int) {
	if err := a.refreshCurrentPair(ctx, sess); err != nil || sess.CurrentPairID == nil {
		a.editOrSend(chatID, messageID, "Сначала создай пару или включи тестовый режим.", noPairKeyboard())
		return
	}

	items, err := a.api.ListIssues(ctx, *sess.CurrentPairID, "active")
	if err != nil {
		a.editOrSend(chatID, messageID, "Не получилось загрузить темы. Попробуй позже.", backMenuKeyboard())
		return
	}
	var filtered []client.Issue
	for _, it := range items {
		if it.RepeatCount > 0 {
			filtered = append(filtered, it)
		}
	}
	if len(filtered) == 0 {
		kb := tgbotapi.NewInlineKeyboardMarkup(
			tgbotapi.NewInlineKeyboardRow(tgbotapi.NewInlineKeyboardButtonData("Назад", cbTopics)),
		)
		a.editOrSend(chatID, messageID, "Темы на повторе: пока пусто.", &kb)
		return
	}

	var b strings.Builder
	b.WriteString("Темы на повторе:\n\n")
	for i, it := range filtered {
		fmt.Fprintf(&b, "%d. %s\nВажность: %s\nПовторений: %d\n\n", i+1, it.Title, issuePriorityLabel(it.Priority), it.RepeatCount)
	}

	var rows [][]tgbotapi.InlineKeyboardButton
	for _, it := range filtered {
		rows = append(rows, tgbotapi.NewInlineKeyboardRow(tgbotapi.NewInlineKeyboardButtonData(it.Title, cbIssueOpenPrefix+it.ID)))
	}
	rows = append(rows, tgbotapi.NewInlineKeyboardRow(tgbotapi.NewInlineKeyboardButtonData("Назад", cbTopics)))
	kb := tgbotapi.NewInlineKeyboardMarkup(rows...)
	a.editOrSend(chatID, messageID, b.String(), &kb)
}

func (a *App) startAddIssue(ctx context.Context, sess *Session, chatID int64, messageID int) {
	if err := a.refreshCurrentPair(ctx, sess); err != nil || sess.CurrentPairID == nil {
		a.editOrSend(chatID, messageID, "Сначала создай пару или включи тестовый режим.", noPairKeyboard())
		return
	}

	sess.State = StateAddIssueTitle
	sess.AddIssueTitle = ""
	sess.AddIssueDescription = ""
	sess.AddIssuePriority = ""
	sess.AddIssueVisibility = ""
	sess.AddIssueThreshold = 0
	a.editOrSend(chatID, messageID, "Название темы?", cancelKeyboard())
}

func (a *App) finishAddIssue(ctx context.Context, sess *Session, chatID int64, messageID int) {
	if sess.CurrentPairID == nil {
		a.editOrSend(chatID, messageID, "Сначала выбери пространство (/menu).", noPairKeyboard())
		return
	}

	vis := sess.AddIssueVisibility
	thr := 0
	if vis == "hidden_until_repeats" {
		thr = sess.AddIssueThreshold
	}
	iss, err := a.api.CreateIssue(ctx, *sess.CurrentPairID, client.CreateIssueRequest{
		CreatedByUserID: sess.APIUserID,
		Title:           sess.AddIssueTitle,
		Description:     sess.AddIssueDescription,
		Priority:        sess.AddIssuePriority,
		Visibility:      vis,
		RepeatThreshold: thr,
	})
	if err != nil {
		a.editOrSend(chatID, messageID, "Не получилось создать тему. Попробуй позже.", backMenuKeyboard())
		return
	}

	sess.State = StateIdle
	card := a.formatIssueCard(iss, nil)
	a.editOrSend(chatID, messageID, "Тема создана.\n\n"+card, issueCreatedKeyboard(iss.ID))
}

func (a *App) showIssue(ctx context.Context, sess *Session, chatID int64, messageID int, issueID string) {
	it, err := a.api.GetIssue(ctx, issueID)
	if err != nil {
		a.editOrSend(chatID, messageID, "Не получилось открыть тему. Попробуй позже.", backMenuKeyboard())
		return
	}
	var last *time.Time
	reps, _ := a.api.ListRepeats(ctx, issueID, 1)
	if len(reps) > 0 {
		t := reps[0].CreatedAt
		last = &t
	}
	text := a.formatIssueCard(it, last)
	a.editOrSend(chatID, messageID, text, issueCardKeyboard(issueID, it.Status))
}

func (a *App) showIssueSettings(ctx context.Context, sess *Session, chatID int64, messageID int, issueID string) {
	it, err := a.api.GetIssue(ctx, issueID)
	if err != nil {
		a.editOrSend(chatID, messageID, "Не получилось открыть настройки темы.", backMenuKeyboard())
		return
	}
	text := fmt.Sprintf("Настройки темы\n\nТема: %s\nЛимит повторений: %d", it.Title, it.RepeatThreshold)
	kb := tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(tgbotapi.NewInlineKeyboardButtonData("✏️ Переименовать", cbIssueRenamePrefix+issueID)),
		tgbotapi.NewInlineKeyboardRow(tgbotapi.NewInlineKeyboardButtonData("🔁 Лимит повторений", cbIssueRepeatLimitPrefix+issueID)),
		tgbotapi.NewInlineKeyboardRow(tgbotapi.NewInlineKeyboardButtonData("Назад", cbIssueOpenPrefix+issueID)),
		tgbotapi.NewInlineKeyboardRow(tgbotapi.NewInlineKeyboardButtonData("В меню", cbMenu)),
	)
	a.editOrSend(chatID, messageID, text, &kb)
}

func (a *App) showRepeats(ctx context.Context, sess *Session, chatID int64, messageID int, issueID string) {
	reps, err := a.api.ListRepeats(ctx, issueID, 10)
	if err != nil {
		a.editOrSend(chatID, messageID, "Не получилось загрузить повторения.", backMenuKeyboard())
		return
	}
	backBtn := tgbotapi.NewInlineKeyboardButtonData("Назад к теме", cbIssueOpenPrefix+issueID)
	if sess.State == StateFocusPlan && sess.FocusIssueID == issueID {
		backBtn = tgbotapi.NewInlineKeyboardButtonData("Назад к фокусу", cbFocusPlan)
	}
	if len(reps) == 0 {
		kb := tgbotapi.NewInlineKeyboardMarkup(
			tgbotapi.NewInlineKeyboardRow(tgbotapi.NewInlineKeyboardButtonData("🔁 Повторилось", cbIssueRepeatPrefix+issueID)),
			tgbotapi.NewInlineKeyboardRow(backBtn),
			tgbotapi.NewInlineKeyboardRow(tgbotapi.NewInlineKeyboardButtonData("В меню", cbMenu)),
		)
		a.editOrSend(chatID, messageID, "Повторений пока нет.", &kb)
		return
	}
	var b strings.Builder
	b.WriteString("Повторения (последние):\n\n")
	for i, r := range reps {
		n := ""
		if r.Note != nil {
			n = *r.Note
		}
		fmt.Fprintf(&b, "%d) %s\n%s\n\n", i+1, r.CreatedAt.Format("2006-01-02 15:04"), n)
	}
	var rows [][]tgbotapi.InlineKeyboardButton
	for _, r := range reps {
		rows = append(rows, tgbotapi.NewInlineKeyboardRow(tgbotapi.NewInlineKeyboardButtonData(r.CreatedAt.Format("2006-01-02 15:04"), cbRepeatOpenPrefix+r.ID)))
	}
	rows = append(rows, tgbotapi.NewInlineKeyboardRow(backBtn))
	rows = append(rows, tgbotapi.NewInlineKeyboardRow(tgbotapi.NewInlineKeyboardButtonData("В меню", cbMenu)))
	kb := tgbotapi.NewInlineKeyboardMarkup(rows...)
	a.editOrSend(chatID, messageID, b.String(), &kb)
}

func (a *App) showRepeat(ctx context.Context, sess *Session, chatID int64, messageID int, repeatID string) {
	rep, err := a.api.GetRepeat(ctx, repeatID)
	if err != nil {
		a.editOrSend(chatID, messageID, "Не получилось открыть повторение.", backMenuKeyboard())
		return
	}
	note := ""
	if rep.Note != nil {
		note = *rep.Note
	}
	text := fmt.Sprintf("Повторение\n\nДата: %s\n\nЗаметка:\n%s", rep.CreatedAt.Format("2006-01-02 15:04"), note)
	kb := tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(tgbotapi.NewInlineKeyboardButtonData("Не согласен", cbRepeatDisagreePrefix+repeatID)),
		tgbotapi.NewInlineKeyboardRow(tgbotapi.NewInlineKeyboardButtonData("🗑 Удалить", cbRepeatDeletePrefix+repeatID)),
		tgbotapi.NewInlineKeyboardRow(tgbotapi.NewInlineKeyboardButtonData("Назад", cbIssueRepeatsPrefix+rep.IssueID)),
		tgbotapi.NewInlineKeyboardRow(tgbotapi.NewInlineKeyboardButtonData("В меню", cbMenu)),
	)
	a.editOrSend(chatID, messageID, text, &kb)
}

func (a *App) confirmCloseIssue(ctx context.Context, sess *Session, chatID int64, messageID int, issueID string) {
	a.editOrSend(chatID, messageID, "Закрыть тему? Это действие лучше делать осознанно.", issueCloseConfirmKeyboard(issueID))
}

func (a *App) updateIssueStatus(ctx context.Context, sess *Session, chatID int64, messageID int, issueID string, status string) {
	it, err := a.api.UpdateIssueStatus(ctx, issueID, status)
	if err != nil {
		a.editOrSend(chatID, messageID, "Не получилось изменить статус. Попробуй позже.", backMenuKeyboard())
		return
	}
	a.editOrSend(chatID, messageID, "Статус обновлен.\n\n"+a.formatIssueCard(it, nil), issueCardKeyboard(issueID, it.Status))
}

func (a *App) startFocusFromCommand(ctx context.Context, sess *Session, chatID int64, messageID int) {
	if sess.ConversationID != "" {
		a.showActiveConversation(ctx, sess, chatID, messageID)
		return
	}
	if err := a.refreshCurrentPair(ctx, sess); err != nil || sess.CurrentPairID == nil {
		a.editOrSend(chatID, messageID, "Сначала создай пару или включи тестовый режим.", noPairKeyboard())
		return
	}
	items, err := a.api.ListIssues(ctx, *sess.CurrentPairID, "active")
	if err != nil {
		a.editOrSend(chatID, messageID, "Не получилось загрузить темы.", backMenuKeyboard())
		return
	}
	if len(items) == 0 {
		kb := tgbotapi.NewInlineKeyboardMarkup(
			tgbotapi.NewInlineKeyboardRow(tgbotapi.NewInlineKeyboardButtonData("➕ Новая тема", cbIssueNew)),
		)
		a.editOrSend(chatID, messageID, "Нет активных тем. Сначала создай тему.", &kb)
		return
	}
	sess.State = StateFocusSelectIssue
	var rows [][]tgbotapi.InlineKeyboardButton
	for _, it := range items {
		rows = append(rows, tgbotapi.NewInlineKeyboardRow(tgbotapi.NewInlineKeyboardButtonData(it.Title, cbFocusStartPrefix+it.ID)))
	}
	rows = append(rows, tgbotapi.NewInlineKeyboardRow(tgbotapi.NewInlineKeyboardButtonData("В меню", cbMenu)))
	kb := tgbotapi.NewInlineKeyboardMarkup(rows...)
	a.editOrSend(chatID, messageID, "Выбери тему для фокуса.", &kb)
}

func (a *App) startFocusForIssue(ctx context.Context, sess *Session, chatID int64, messageID int, issueID string) {
	sess.FocusIssueID = issueID
	sess.FocusGoal = ""
	sess.FocusQuestions = ""
	sess.FocusStartStateSelf = ""
	sess.FocusStartStatePartner = ""
	sess.State = StateFocusGoal
	a.askFocusGoal(chatID, messageID)
}

func (a *App) askFocusGoal(chatID int64, messageID int) {
	a.editOrSend(chatID, messageID, "Цель разговора: что хотите получить?\n\nНапиши одним сообщением.", backKeyboard(cbFocusBackToSelect))
}

func (a *App) showFocusPlan(ctx context.Context, sess *Session, chatID int64, messageID int) {
	issueItem, err := a.api.GetIssue(ctx, sess.FocusIssueID)
	if err != nil {
		a.editOrSend(chatID, messageID, "Не получилось открыть тему.", backMenuKeyboard())
		return
	}
	startState := "—"
	if strings.TrimSpace(sess.FocusStartStateSelf) != "" || strings.TrimSpace(sess.FocusStartStatePartner) != "" {
		startState = fmt.Sprintf("Я: %s; Партнер: %s", strings.TrimSpace(sess.FocusStartStateSelf), strings.TrimSpace(sess.FocusStartStatePartner))
	}
	text := fmt.Sprintf("План разговора.\n\nТема: %s\nЦель: %s\nВопросы: %s\nСостояние на старте: %s",
		issueItem.Title,
		sess.FocusGoal,
		sess.FocusQuestions,
		startState,
	)
	kb := tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(tgbotapi.NewInlineKeyboardButtonData("▶️ Начать разговор", cbFocusBegin)),
		tgbotapi.NewInlineKeyboardRow(tgbotapi.NewInlineKeyboardButtonData("📝 Смотреть повторения", cbIssueRepeatsPrefix+sess.FocusIssueID)),
		tgbotapi.NewInlineKeyboardRow(tgbotapi.NewInlineKeyboardButtonData("Назад", cbFocusBackToPartnerState)),
	)
	a.editOrSend(chatID, messageID, text, &kb)
}

func (a *App) beginConversation(ctx context.Context, sess *Session, chatID int64, messageID int) {
	if err := a.refreshCurrentPair(ctx, sess); err != nil || sess.CurrentPairID == nil {
		a.editOrSend(chatID, messageID, "Сначала выбери пространство.", noPairKeyboard())
		return
	}
	goal := sess.FocusGoal
	questions := sess.FocusQuestions
	startState := fmt.Sprintf("Я: %s; Партнер: %s", strings.TrimSpace(sess.FocusStartStateSelf), strings.TrimSpace(sess.FocusStartStatePartner))
	cs, err := a.api.StartConversation(ctx, sess.FocusIssueID, *sess.CurrentPairID, &goal, &questions, &startState)
	if err != nil {
		a.editOrSend(chatID, messageID, "Не получилось начать разговор. Попробуй позже.", backMenuKeyboard())
		return
	}
	sess.ConversationID = cs.ID
	sess.ConversationIssueID = sess.FocusIssueID
	sess.ConversationStartedAt = time.Now()
	sess.State = StateIdle
	a.showActiveConversation(ctx, sess, chatID, messageID)
}

func (a *App) showActiveConversation(ctx context.Context, sess *Session, chatID int64, messageID int) {
	if sess.ConversationID == "" {
		a.editOrSend(chatID, messageID, "Нет активного разговора. Нажми /focus.", nil)
		return
	}
	issueItem, _ := a.api.GetIssue(ctx, sess.ConversationIssueID)
	cs, err := a.api.GetConversation(ctx, sess.ConversationID)
	if err != nil {
		a.editOrSend(chatID, messageID, "Не получилось открыть разговор.", backMenuKeyboard())
		return
	}
	paused := cs.Status == "paused"
	dur := formatDuration(time.Since(sess.ConversationStartedAt))
	text := fmt.Sprintf("Идет разговор по теме: %s\nТаймер: %s", issueItem.Title, dur)
	if paused {
		text = fmt.Sprintf("Разговор на паузе по теме: %s\nТаймер: %s", issueItem.Title, dur)
	}
	a.editOrSend(chatID, messageID, text, conversationKeyboard(sess.ConversationID, paused))
}

func (a *App) pauseConversation(ctx context.Context, sess *Session, chatID int64, messageID int, conversationID string) {
	_, err := a.api.PauseConversation(ctx, conversationID)
	if err != nil {
		a.editOrSend(chatID, messageID, "Не получилось поставить на паузу.", backMenuKeyboard())
		return
	}
	a.showActiveConversation(ctx, sess, chatID, messageID)
}

func (a *App) resumeConversation(ctx context.Context, sess *Session, chatID int64, messageID int, conversationID string) {
	_, err := a.api.ResumeConversation(ctx, conversationID)
	if err != nil {
		a.editOrSend(chatID, messageID, "Не получилось продолжить разговор.", backMenuKeyboard())
		return
	}
	a.showActiveConversation(ctx, sess, chatID, messageID)
}

func (a *App) finishConversationEarly(ctx context.Context, sess *Session, chatID int64, messageID int) {
	if sess.ConversationID == "" {
		a.editOrSend(chatID, messageID, "Нет активного разговора.", backToMenuKeyboard())
		return
	}
	reasonLabel := earlyEndReasonLabel(sess.EarlyEndReason)
	reason := reasonLabel
	cs, err := a.api.FinishConversation(ctx, sess.ConversationID, client.FinishConversationRequest{
		ResultStatus:  "postponed",
		EndedEarly:    true,
		EndedByUserID: &sess.APIUserID,
		EndReason:     &reason,
	})
	if err != nil {
		a.editOrSend(chatID, messageID, "Не получилось завершить разговор досрочно.", backMenuKeyboard())
		return
	}
	_, _ = a.api.UpdateIssueStatus(ctx, cs.IssueID, "postponed")
	sess.State = StateIdle
	sess.EarlyEndReason = ""
	sess.ConversationID = ""
	kb := tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(tgbotapi.NewInlineKeyboardButtonData("Открыть тему", cbIssueOpenPrefix+cs.IssueID)),
		tgbotapi.NewInlineKeyboardRow(tgbotapi.NewInlineKeyboardButtonData("История", cbNavHistory)),
	)
	a.editOrSend(chatID, messageID, "Разговор завершен досрочно.\nПричина: "+reasonLabel, &kb)
}

func (a *App) showHistory(ctx context.Context, sess *Session, chatID int64, messageID int) {
	if err := a.refreshCurrentPair(ctx, sess); err != nil || sess.CurrentPairID == nil {
		a.editOrSend(chatID, messageID, "Сначала создай пару или включи тестовый режим.", noPairKeyboard())
		return
	}
	items, err := a.api.ListConversations(ctx, *sess.CurrentPairID, "finished")
	if err != nil {
		a.editOrSend(chatID, messageID, "Не получилось загрузить историю.", backMenuKeyboard())
		return
	}
	if len(items) == 0 {
		a.editOrSend(chatID, messageID, "История пока пустая.", backMenuKeyboard())
		return
	}
	var b strings.Builder
	b.WriteString("История разговоров:\n\n")
	var rows [][]tgbotapi.InlineKeyboardButton
	for _, it := range items {
		iss, _ := a.api.GetIssue(ctx, it.IssueID)
		status := ""
		if it.ResultStatus != nil {
			status = resultStatusLabel(*it.ResultStatus)
		}
		line := fmt.Sprintf("%s — %s (%s)", it.CreatedAt.Format("2006-01-02"), iss.Title, status)
		b.WriteString(line + "\n")
		rows = append(rows, tgbotapi.NewInlineKeyboardRow(tgbotapi.NewInlineKeyboardButtonData(line, cbHistoryOpenPrefix+it.ID)))
	}
	rows = append(rows, tgbotapi.NewInlineKeyboardRow(tgbotapi.NewInlineKeyboardButtonData("Назад", cbMenu)))
	kb := tgbotapi.NewInlineKeyboardMarkup(rows...)
	a.editOrSend(chatID, messageID, b.String(), &kb)
}

func (a *App) showNotes(ctx context.Context, sess *Session, chatID int64, messageID int) {
	if err := a.refreshCurrentPair(ctx, sess); err != nil || sess.CurrentPairID == nil {
		a.editOrSend(chatID, messageID, "Сначала создай пару или включи тестовый режим.", noPairKeyboard())
		return
	}
	items, err := a.api.ListPairNotes(ctx, *sess.CurrentPairID, 20)
	if err != nil {
		a.editOrSend(chatID, messageID, "Не получилось загрузить заметки.", nil)
		return
	}
	if len(items) == 0 {
		a.editOrSend(chatID, messageID, "Мысли и заметки пока пустые.\n\nВо время фокуса можно нажимать «📝 Записать мысль».", backToMenuKeyboard())
		return
	}

	var b strings.Builder
	b.WriteString("Мысли и заметки (последние):\n\n")
	var rows [][]tgbotapi.InlineKeyboardButton
	for _, it := range items {
		line := fmt.Sprintf("%s — %s\n%s", it.CreatedAt.Format("2006-01-02 15:04"), it.IssueTitle, it.Text)
		b.WriteString(line + "\n\n")
		btn := fmt.Sprintf("%s — %s", it.CreatedAt.Format("2006-01-02"), it.IssueTitle)
		rows = append(rows, tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData(btn, cbHistoryOpenPrefix+it.ConversationID),
			tgbotapi.NewInlineKeyboardButtonData("🗑", cbNoteDeletePrefix+it.ID),
		))
	}
	rows = append(rows, tgbotapi.NewInlineKeyboardRow(tgbotapi.NewInlineKeyboardButtonData("Назад", cbMenu)))
	kb := tgbotapi.NewInlineKeyboardMarkup(rows...)
	a.editOrSend(chatID, messageID, b.String(), &kb)
}

func (a *App) openHistoryConversation(ctx context.Context, sess *Session, chatID int64, messageID int, conversationID string) {
	cs, err := a.api.GetConversation(ctx, conversationID)
	if err != nil {
		a.editOrSend(chatID, messageID, "Не получилось открыть запись.", backMenuKeyboard())
		return
	}
	iss, _ := a.api.GetIssue(ctx, cs.IssueID)
	notes, _ := a.api.ListConversationNotes(ctx, conversationID, 20)
	side, _ := a.api.ListConversationSideIssues(ctx, conversationID)
	rs := ""
	if cs.ResultStatus != nil {
		rs = resultStatusLabel(*cs.ResultStatus)
	}
	rt := ""
	if cs.ResultText != nil {
		rt = *cs.ResultText
	}
	goal := "—"
	if cs.Goal != nil && strings.TrimSpace(*cs.Goal) != "" {
		goal = *cs.Goal
	}
	qs := "—"
	if cs.Questions != nil && strings.TrimSpace(*cs.Questions) != "" {
		qs = *cs.Questions
	}
	ss := "—"
	if cs.StartState != nil && strings.TrimSpace(*cs.StartState) != "" {
		ss = *cs.StartState
	}
	es := "—"
	if cs.EndState != nil && strings.TrimSpace(*cs.EndState) != "" {
		es = *cs.EndState
	}
	early := ""
	if cs.EndedEarly {
		who := "—"
		if cs.EndedByUserID != nil && strings.TrimSpace(*cs.EndedByUserID) != "" {
			if strings.TrimSpace(*cs.EndedByUserID) == strings.TrimSpace(sess.APIUserID) {
				who = "Я"
			} else {
				who = "Партнер"
			}
		}
		reason := "—"
		if cs.EndReason != nil && strings.TrimSpace(*cs.EndReason) != "" {
			reason = strings.TrimSpace(*cs.EndReason)
		}
		early = fmt.Sprintf("\nДосрочно завершено: да\nКто: %s\nПочему: %s", who, reason)
	}

	var b strings.Builder
	fmt.Fprintf(&b, "Разговор\n\nДата: %s\nТема: %s\nЦель: %s\nВопросы: %s\nСостояние на старте: %s\nСостояние в конце: %s\nИтог: %s%s\n\n%s",
		cs.CreatedAt.Format("2006-01-02 15:04"),
		iss.Title,
		goal,
		qs,
		ss,
		es,
		rs,
		early,
		rt,
	)
	if len(notes) > 0 {
		b.WriteString("\n\nЗаметки:\n")
		for _, n := range notes {
			fmt.Fprintf(&b, "• %s — %s\n", n.CreatedAt.Format("2006-01-02 15:04"), n.Text)
		}
	}
	if len(side) > 0 {
		b.WriteString("\n\nПобочные темы:\n")
		for _, it := range side {
			fmt.Fprintf(&b, "• %s\n", it.Title)
		}
	}
	kb := tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(tgbotapi.NewInlineKeyboardButtonData("Назад", cbNavHistory)),
		tgbotapi.NewInlineKeyboardRow(tgbotapi.NewInlineKeyboardButtonData("В меню", cbMenu)),
	)
	a.editOrSend(chatID, messageID, b.String(), &kb)
}

func (a *App) showIssueHistory(ctx context.Context, sess *Session, chatID int64, messageID int, issueID string) {
	if err := a.refreshCurrentPair(ctx, sess); err != nil || sess.CurrentPairID == nil {
		a.editOrSend(chatID, messageID, "Сначала создай пару или включи тестовый режим.", noPairKeyboard())
		return
	}
	items, err := a.api.ListConversations(ctx, *sess.CurrentPairID, "finished")
	if err != nil {
		a.editOrSend(chatID, messageID, "Не получилось загрузить историю.", backKeyboard(cbIssueOpenPrefix+issueID))
		return
	}
	var filtered []client.ConversationSession
	for _, it := range items {
		if it.IssueID == issueID {
			filtered = append(filtered, it)
		}
	}
	iss, _ := a.api.GetIssue(ctx, issueID)
	if len(filtered) == 0 {
		kb := tgbotapi.NewInlineKeyboardMarkup(
			tgbotapi.NewInlineKeyboardRow(tgbotapi.NewInlineKeyboardButtonData("Назад к теме", cbIssueOpenPrefix+issueID)),
		)
		a.editOrSend(chatID, messageID, "История по теме пуста.\n\nТема: "+iss.Title, &kb)
		return
	}

	var b strings.Builder
	b.WriteString("История по теме:\n\n")
	b.WriteString("Тема: " + iss.Title + "\n\n")
	var rows [][]tgbotapi.InlineKeyboardButton
	for _, it := range filtered {
		status := ""
		if it.ResultStatus != nil {
			status = resultStatusLabel(*it.ResultStatus)
		}
		line := fmt.Sprintf("%s — %s", it.CreatedAt.Format("2006-01-02"), status)
		b.WriteString(line + "\n")
		rows = append(rows, tgbotapi.NewInlineKeyboardRow(tgbotapi.NewInlineKeyboardButtonData(line, cbHistoryOpenPrefix+it.ID)))
	}
	rows = append(rows, tgbotapi.NewInlineKeyboardRow(tgbotapi.NewInlineKeyboardButtonData("Назад к теме", cbIssueOpenPrefix+issueID)))
	rows = append(rows, tgbotapi.NewInlineKeyboardRow(tgbotapi.NewInlineKeyboardButtonData("В меню", cbMenu)))
	kb := tgbotapi.NewInlineKeyboardMarkup(rows...)
	a.editOrSend(chatID, messageID, b.String(), &kb)
}

func (a *App) showSettings(ctx context.Context, sess *Session, chatID int64, messageID int) {
	kb := tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(tgbotapi.NewInlineKeyboardButtonData("Инструкция", cbHowTo)),
		tgbotapi.NewInlineKeyboardRow(tgbotapi.NewInlineKeyboardButtonData("Правила использования", "settings:rules")),
		tgbotapi.NewInlineKeyboardRow(tgbotapi.NewInlineKeyboardButtonData("Моя пара", "settings:pair")),
		tgbotapi.NewInlineKeyboardRow(tgbotapi.NewInlineKeyboardButtonData("Поделиться ссылкой", cbSettingsShare)),
		tgbotapi.NewInlineKeyboardRow(tgbotapi.NewInlineKeyboardButtonData("Назад", cbMenu)),
	)
	a.editOrSend(chatID, messageID, "Настройки", &kb)
}

func (a *App) showPairSpaces(ctx context.Context, sess *Session, chatID int64, messageID int) {
	pairs, err := a.api.ListPairs(ctx, sess.APIUserID)
	if err != nil {
		a.editOrSend(chatID, messageID, "Не получилось загрузить пары.", backMenuKeyboard())
		return
	}
	if len(pairs) == 0 {
		a.editOrSend(chatID, messageID, "Пока нет пар. Создай пару или включи тестовый режим.", noPairKeyboard())
		return
	}
	if err := a.refreshCurrentPair(ctx, sess); err != nil {
		a.editOrSend(chatID, messageID, "Сервис временно недоступен. Попробуй позже.", backMenuKeyboard())
		return
	}

	var b strings.Builder
	b.WriteString("Моя пара / пространства:\n\n")
	var rows [][]tgbotapi.InlineKeyboardButton
	seenTest := false
	for _, p := range pairs {
		if p.Status != "active" {
			continue
		}
		if p.IsTest {
			if seenTest {
				continue
			}
			seenTest = true
		}
		label := p.ID
		if p.IsTest {
			label = "Тестовое пространство"
		}
		active := ""
		if sess.CurrentPairID != nil && *sess.CurrentPairID == p.ID {
			active = " (текущее)"
		}
		btn := tgbotapi.NewInlineKeyboardButtonData(label+active, cbPrefSwitchPrefix+p.ID)
		rows = append(rows, tgbotapi.NewInlineKeyboardRow(btn))
	}
	if len(rows) == 0 {
		kb := tgbotapi.NewInlineKeyboardMarkup(
			tgbotapi.NewInlineKeyboardRow(tgbotapi.NewInlineKeyboardButtonData("Назад", cbNavSettings)),
		)
		a.editOrSend(chatID, messageID, "Пока нет активных пространств.", &kb)
		return
	}

	// Test mode toggle is here (not in settings).
	if sess.CurrentPairID != nil && sess.CurrentPairIsTest {
		rows = append(rows, tgbotapi.NewInlineKeyboardRow(tgbotapi.NewInlineKeyboardButtonData("Выключить тестовый режим", cbTestStop)))
	} else {
		rows = append(rows, tgbotapi.NewInlineKeyboardRow(tgbotapi.NewInlineKeyboardButtonData("Включить тестовый режим", cbTestStart)))
	}
	rows = append(rows, tgbotapi.NewInlineKeyboardRow(tgbotapi.NewInlineKeyboardButtonData("Назад", cbNavSettings)))
	kb := tgbotapi.NewInlineKeyboardMarkup(rows...)
	a.editOrSend(chatID, messageID, b.String(), &kb)
}

func (a *App) sharePairLink(ctx context.Context, sess *Session, chatID int64, messageID int) {
	if err := a.refreshCurrentPair(ctx, sess); err != nil || sess.CurrentPairID == nil {
		a.editOrSend(chatID, messageID, "Сначала создай пару или включи тестовый режим.", noPairKeyboard())
		return
	}
	token, err := a.api.CreatePairInvite(ctx, *sess.CurrentPairID, sess.APIUserID)
	if err != nil {
		a.editOrSend(chatID, messageID, "Не получилось создать ссылку. Попробуй позже.", backToMenuKeyboard())
		return
	}
	link := fmt.Sprintf("https://t.me/%s?start=%s", a.botUser, token)
	p, _, _ := a.api.GetPair(ctx, *sess.CurrentPairID)
	welcome := "—"
	if p.WelcomeMessage != nil && strings.TrimSpace(*p.WelcomeMessage) != "" {
		welcome = strings.TrimSpace(*p.WelcomeMessage)
	}

	text := "Поделиться ссылкой:\n\nСсылка для партнера:\n" + link + "\n\nПриветствие партнеру (покажется после подключения):\n" + welcome
	kb := tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(tgbotapi.NewInlineKeyboardButtonData("✏️ Приветствие", cbPairWelcome)),
		tgbotapi.NewInlineKeyboardRow(tgbotapi.NewInlineKeyboardButtonData("🔄 Обновить ссылку", cbSettingsShare)),
		tgbotapi.NewInlineKeyboardRow(tgbotapi.NewInlineKeyboardButtonData("Назад", cbNavSettings)),
		tgbotapi.NewInlineKeyboardRow(tgbotapi.NewInlineKeyboardButtonData("В меню", cbMenu)),
	)
	a.editOrSend(chatID, messageID, text, &kb)
}

func (a *App) showTestMode(ctx context.Context, sess *Session, chatID int64, messageID int) {
	if err := a.refreshCurrentPair(ctx, sess); err != nil {
		a.editOrSend(chatID, messageID, "Сервис временно недоступен. Попробуй позже.", nil)
		return
	}
	if sess.CurrentPairID != nil && sess.CurrentPairIsTest {
		kb := tgbotapi.NewInlineKeyboardMarkup(
			tgbotapi.NewInlineKeyboardRow(tgbotapi.NewInlineKeyboardButtonData("Выключить тестовый режим", cbTestStop)),
			tgbotapi.NewInlineKeyboardRow(tgbotapi.NewInlineKeyboardButtonData("Назад", cbNavSettings)),
		)
		a.editOrSend(chatID, messageID, "Тестовый режим активен.", &kb)
		return
	}
	text := "Тестовый режим позволяет проверить весь функционал без второго партнера.\n\nБудет создана тестовая пара с виртуальным партнером.\nЭто нужно только для проверки сценариев."
	kb := tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(tgbotapi.NewInlineKeyboardButtonData("Включить тестовый режим", cbTestStart)),
		tgbotapi.NewInlineKeyboardRow(tgbotapi.NewInlineKeyboardButtonData("Назад", cbNavSettings)),
	)
	a.editOrSend(chatID, messageID, text, &kb)
}

func (a *App) startTestMode(ctx context.Context, sess *Session, chatID int64, messageID int) {
	_, err := a.api.StartTestMode(ctx, sess.APIUserID)
	if err != nil {
		a.editOrSend(chatID, messageID, "Не получилось включить тестовый режим. Попробуй позже.", backMenuKeyboard())
		return
	}
	_ = a.refreshCurrentPair(ctx, sess)
	a.showPairSpaces(ctx, sess, chatID, messageID)
}

func (a *App) stopTestMode(ctx context.Context, sess *Session, chatID int64, messageID int) {
	_, err := a.api.StopTestMode(ctx, sess.APIUserID)
	if err != nil {
		a.editOrSend(chatID, messageID, "Не получилось выключить тестовый режим.", backMenuKeyboard())
		return
	}
	_ = a.refreshCurrentPair(ctx, sess)
	a.showPairSpaces(ctx, sess, chatID, messageID)
}

func (a *App) switchCurrentPair(ctx context.Context, sess *Session, chatID int64, messageID int, pairID string) {
	_, err := a.api.SetPreferences(ctx, sess.APIUserID, &pairID)
	if err != nil {
		a.editOrSend(chatID, messageID, "Не получилось переключить пространство.", backMenuKeyboard())
		return
	}
	_ = a.refreshCurrentPair(ctx, sess)
	a.showMenu(ctx, sess, chatID, messageID)
}

func (a *App) cancelFlow(ctx context.Context, sess *Session, chatID int64, messageID int) {
	sess.State = StateIdle
	sess.CurrentIssueID = ""
	sess.CurrentRepeatID = ""
	sess.PendingInviteToken = ""
	sess.SideIssueTitle = ""
	sess.EarlyEndReason = ""
	sess.AddIssueTitle = ""
	sess.AddIssueDescription = ""
	sess.AddIssuePriority = ""
	sess.AddIssueVisibility = ""
	sess.AddIssueThreshold = 0
	sess.FocusIssueID = ""
	sess.FocusGoal = ""
	sess.FocusQuestions = ""
	sess.FocusStartStateSelf = ""
	sess.FocusStartStatePartner = ""
	sess.FinishResultStatus = ""
	sess.FinishEndStateSelf = ""
	sess.FinishEndStatePartner = ""
	sess.QuickIssueTitle = ""
	sess.PendingForwardText = ""
	a.showMenu(ctx, sess, chatID, messageID)
}

func (a *App) handleFSMText(ctx context.Context, sess *Session, msg *tgbotapi.Message) {
	text := strings.TrimSpace(msg.Text)
	if text == "" {
		if msg.Caption != "" {
			text = strings.TrimSpace(msg.Caption)
		}
	}
	if isForwardedMessage(msg) && (sess.State == StateAddRepeatNote || sess.State == StateConvNote) {
		sess.PendingForwardText = formatForwardedMessage(msg)
		if sess.State == StateAddRepeatNote {
			sess.State = StateAddRepeatForwardComment
			a.sendText(msg.Chat.ID, "Добавь комментарий к пересланному сообщению одним сообщением.", nil)
			return
		}
		sess.State = StateConvForwardComment
		a.sendText(msg.Chat.ID, "Добавь комментарий к пересланному сообщению одним сообщением.", backKeyboard(cbConvBackToActive))
		return
	}

	switch sess.State {
	case StatePairWelcome:
		if strings.TrimSpace(text) == "" {
			a.sendText(msg.Chat.ID, "Напиши текст приветствия.", backKeyboard(cbSettingsShare))
			return
		}
		if err := a.refreshCurrentPair(ctx, sess); err != nil || sess.CurrentPairID == nil || sess.CurrentPairIsTest {
			sess.State = StateIdle
			a.sendText(msg.Chat.ID, "Сначала выбери основное пространство.", nil)
			return
		}
		if _, err := a.api.SetPairWelcomeMessage(ctx, *sess.CurrentPairID, sess.APIUserID, &text); err != nil {
			sess.State = StateIdle
			a.sendText(msg.Chat.ID, "Не получилось сохранить приветствие.", nil)
			return
		}
		sess.State = StateIdle
		a.sharePairLink(ctx, sess, msg.Chat.ID, 0)
	case StateIssueRename:
		if strings.TrimSpace(text) == "" {
			a.sendText(msg.Chat.ID, "Напиши название текстом.", backKeyboard(cbIssueSettingsPrefix+sess.CurrentIssueID))
			return
		}
		it, err := a.api.UpdateIssue(ctx, sess.CurrentIssueID, client.UpdateIssueRequest{Title: &text})
		if err != nil {
			sess.State = StateIdle
			a.sendText(msg.Chat.ID, "Не получилось переименовать тему.", backMenuKeyboard())
			return
		}
		sess.State = StateIdle
		a.showIssueSettings(ctx, sess, msg.Chat.ID, 0, it.ID)
	case StateIssueRepeatLimit:
		val := strings.TrimSpace(text)
		if val == "" {
			a.sendText(msg.Chat.ID, "Напиши число.", backKeyboard(cbIssueSettingsPrefix+sess.CurrentIssueID))
			return
		}
		limit, ok := parsePositiveInt(val)
		if !ok {
			a.sendText(msg.Chat.ID, "Нужен формат числа (например 0, 1, 2...).", backKeyboard(cbIssueSettingsPrefix+sess.CurrentIssueID))
			return
		}
		it, err := a.api.UpdateIssue(ctx, sess.CurrentIssueID, client.UpdateIssueRequest{RepeatThreshold: &limit})
		if err != nil {
			sess.State = StateIdle
			a.sendText(msg.Chat.ID, "Не получилось обновить лимит повторений.", backMenuKeyboard())
			return
		}
		sess.State = StateIdle
		a.showIssueSettings(ctx, sess, msg.Chat.ID, 0, it.ID)
	case StateAddIssueTitle:
		if text == "" {
			a.sendText(msg.Chat.ID, "Напиши название текстом.", cancelKeyboard())
			return
		}
		sess.AddIssueTitle = text
		sess.State = StateAddIssueDescription
		a.sendText(msg.Chat.ID, "Опиши ситуацию подробнее.", backCancelKeyboard("issue:back_title"))
	case StateAddIssueDescription:
		if text == "" {
			a.sendText(msg.Chat.ID, "Напиши описание текстом.", cancelKeyboard())
			return
		}
		sess.AddIssueDescription = text
		sess.State = StateAddIssuePriority
		a.sendText(msg.Chat.ID, "Выбери важность.", priorityKeyboard())
	case StateAddRepeatNote:
		if text == "" {
			a.sendText(msg.Chat.ID, "Нужен текст. Напиши короткую заметку.", cancelKeyboard())
			return
		}
		it, err := a.api.RepeatIssue(ctx, sess.CurrentIssueID, sess.APIUserID, text)
		if err != nil {
			a.sendText(msg.Chat.ID, "Не получилось зафиксировать повторение. Попробуй позже.", backMenuKeyboard())
			sess.State = StateIdle
			return
		}
		sess.State = StateIdle
		kb := tgbotapi.NewInlineKeyboardMarkup(
			tgbotapi.NewInlineKeyboardRow(tgbotapi.NewInlineKeyboardButtonData("Открыть тему", cbIssueOpenPrefix+it.ID)),
			tgbotapi.NewInlineKeyboardRow(tgbotapi.NewInlineKeyboardButtonData("Начать фокус", cbFocusStartPrefix+it.ID)),
		)
		a.sendText(msg.Chat.ID, fmt.Sprintf("Повторение зафиксировано.\nВсего повторений: %d.", it.RepeatCount), &kb)
	case StateAddRepeatForwardComment:
		if strings.TrimSpace(text) == "" {
			a.sendText(msg.Chat.ID, "Нужен комментарий текстом.", nil)
			return
		}
		note := strings.TrimSpace(text)
		if strings.TrimSpace(sess.PendingForwardText) != "" {
			note = "Переслано:\n" + strings.TrimSpace(sess.PendingForwardText) + "\n\nКомментарий:\n" + note
		}
		sess.PendingForwardText = ""
		it, err := a.api.RepeatIssue(ctx, sess.CurrentIssueID, sess.APIUserID, note)
		if err != nil {
			a.sendText(msg.Chat.ID, "Не получилось зафиксировать повторение. Попробуй позже.", backMenuKeyboard())
			sess.State = StateIdle
			return
		}
		sess.State = StateIdle
		kb := tgbotapi.NewInlineKeyboardMarkup(
			tgbotapi.NewInlineKeyboardRow(tgbotapi.NewInlineKeyboardButtonData("Открыть тему", cbIssueOpenPrefix+it.ID)),
			tgbotapi.NewInlineKeyboardRow(tgbotapi.NewInlineKeyboardButtonData("Начать фокус", cbFocusStartPrefix+it.ID)),
		)
		a.sendText(msg.Chat.ID, fmt.Sprintf("Повторение зафиксировано.\nВсего повторений: %d.", it.RepeatCount), &kb)
	case StateAddRepeatDisagreeNote:
		if text == "" {
			a.sendText(msg.Chat.ID, "Нужен текст. Напиши заметку.", cancelKeyboard())
			return
		}
		_, err := a.api.AddRepeatDisagreement(ctx, sess.CurrentRepeatID, sess.APIUserID, text)
		if err != nil {
			a.sendText(msg.Chat.ID, "Не получилось сохранить заметку несогласия. Возможно, она уже есть.", backMenuKeyboard())
			sess.State = StateIdle
			return
		}
		sess.State = StateIdle
		a.sendText(msg.Chat.ID, "Позиция сохранена.", backMenuKeyboard())
	case StateFocusGoal:
		if text == "" {
			a.sendText(msg.Chat.ID, "Напиши цель текстом.", cancelKeyboard())
			return
		}
		sess.FocusGoal = text
		sess.State = StateFocusQuestions
		a.sendText(msg.Chat.ID, "Главные вопросы разговора.", backKeyboard(cbFocusBackToGoal))
	case StateFocusQuestions:
		if text == "" {
			a.sendText(msg.Chat.ID, "Напиши вопросы текстом.", cancelKeyboard())
			return
		}
		sess.FocusQuestions = text
		sess.State = StateFocusStartStateSelf
		a.sendText(msg.Chat.ID, "Состояние перед разговором.\n\nТвое состояние:", stateKeyboard(cbFocusStartStateSelfPrefix, "focus:back_questions"))
	case StateConvNote:
		if text == "" {
			a.sendText(msg.Chat.ID, "Напиши мысль текстом.", backKeyboard(cbConvBackToActive))
			return
		}
		if err := a.api.AddConversationNote(ctx, sess.ConversationID, sess.APIUserID, text); err != nil {
			a.sendText(msg.Chat.ID, "Не получилось сохранить заметку.", backMenuKeyboard())
			sess.State = StateIdle
			return
		}
		sess.State = StateIdle
		a.sendText(msg.Chat.ID, "Сохранено.", conversationKeyboard(sess.ConversationID, false))
	case StateConvForwardComment:
		if strings.TrimSpace(text) == "" {
			a.sendText(msg.Chat.ID, "Нужен комментарий текстом.", backKeyboard(cbConvBackToActive))
			return
		}
		note := strings.TrimSpace(text)
		if strings.TrimSpace(sess.PendingForwardText) != "" {
			note = "Переслано:\n" + strings.TrimSpace(sess.PendingForwardText) + "\n\nКомментарий:\n" + note
		}
		sess.PendingForwardText = ""
		if err := a.api.AddConversationNote(ctx, sess.ConversationID, sess.APIUserID, note); err != nil {
			a.sendText(msg.Chat.ID, "Не получилось сохранить заметку.", backMenuKeyboard())
			sess.State = StateIdle
			return
		}
		sess.State = StateIdle
		a.sendText(msg.Chat.ID, "Сохранено.", conversationKeyboard(sess.ConversationID, false))
	case StateConvSideTitle:
		if text == "" {
			a.sendText(msg.Chat.ID, "Напиши название текстом.", backKeyboard(cbConvBackToActive))
			return
		}
		sess.SideIssueTitle = text
		sess.State = StateConvSideDesc
		a.sendText(msg.Chat.ID, "Коротко опиши побочную тему.", backKeyboard(cbConvBackToActive))
	case StateConvSideDesc:
		if text == "" {
			a.sendText(msg.Chat.ID, "Напиши описание текстом.", backKeyboard(cbConvBackToActive))
			return
		}
		_, err := a.api.AddSideIssue(ctx, sess.ConversationID, client.AddSideIssueRequest{
			CreatedByUserID: sess.APIUserID,
			Title:           sess.SideIssueTitle,
			Description:     text,
		})
		if err != nil {
			a.sendText(msg.Chat.ID, "Не получилось сохранить побочную тему.", backMenuKeyboard())
			sess.State = StateIdle
			return
		}
		sess.State = StateIdle
		a.sendText(msg.Chat.ID, "Побочная тема сохранена. Можно вернуться к разговору.", conversationKeyboard(sess.ConversationID, false))
	case StateConvEarlyOther:
		if strings.TrimSpace(text) == "" {
			a.sendText(msg.Chat.ID, "Нужен текст.", backKeyboard(cbConvBackToActive))
			return
		}
		sess.State = StateIdle
		sess.EarlyEndReason = "other:" + strings.TrimSpace(text)
		a.finishConversationEarly(ctx, sess, msg.Chat.ID, 0)
	case StateFinishText:
		if text == "" {
			a.sendText(msg.Chat.ID, "Напиши итог текстом.", backKeyboard(cbFinishBackToStatus))
			return
		}
		rt := text
		endState := fmt.Sprintf("Я: %s; Партнер: %s", strings.TrimSpace(sess.FinishEndStateSelf), strings.TrimSpace(sess.FinishEndStatePartner))
		cs, err := a.api.FinishConversation(ctx, sess.ConversationID, client.FinishConversationRequest{
			ResultStatus: sess.FinishResultStatus,
			ResultText:   &rt,
			EndState:     &endState,
		})
		if err != nil {
			a.sendText(msg.Chat.ID, "Не получилось сохранить итог.", backMenuKeyboard())
			sess.State = StateIdle
			return
		}
		// Update issue status depending on result.
		switch sess.FinishResultStatus {
		case "resolved":
			_, _ = a.api.UpdateIssueStatus(ctx, cs.IssueID, "resolved")
		case "postponed":
			_, _ = a.api.UpdateIssueStatus(ctx, cs.IssueID, "postponed")
		}
		sess.State = StateIdle
		sess.ConversationID = ""
		kb := tgbotapi.NewInlineKeyboardMarkup(
			tgbotapi.NewInlineKeyboardRow(tgbotapi.NewInlineKeyboardButtonData("Открыть тему", cbIssueOpenPrefix+cs.IssueID)),
			tgbotapi.NewInlineKeyboardRow(tgbotapi.NewInlineKeyboardButtonData("История", cbNavHistory)),
		)
		a.sendText(msg.Chat.ID, "Итог сохранен.", &kb)
	default:
		if sess.State == StateIdle && text != "" {
			sess.QuickIssueTitle = text
			a.sendText(msg.Chat.ID, fmt.Sprintf("Похоже, это новая тема:\n\n«%s»\n\nСоздать?", text), quickIssueConfirmKeyboard())
			return
		}
		a.sendText(msg.Chat.ID, "Нажми /menu.", nil)
	}
}

func (a *App) formatIssueCard(it client.Issue, lastRepeat *time.Time) string {
	last := "—"
	if lastRepeat != nil {
		lt := lastRepeat.In(time.Local)
		last = lt.Format("2006-01-02 15:04")
	}
	return fmt.Sprintf("%s\n\nСтатус: %s\nВажность: %s\nПовторений: %d\nЛимит повторений: %d\nПоследний случай: %s",
		it.Title,
		issueStatusLabel(it.Status),
		issuePriorityLabel(it.Priority),
		it.RepeatCount,
		it.RepeatThreshold,
		last,
	)
}

func stateLabel(code string) string {
	switch strings.TrimSpace(code) {
	case "calm":
		return "Спокойно"
	case "tense":
		return "Напряженно"
	case "irritated":
		return "Раздраженно"
	case "tired":
		return "Устал(а)"
	case "not_ready":
		return "Не готов(а)"
	default:
		return code
	}
}

func issueStatusLabel(status string) string {
	switch strings.TrimSpace(status) {
	case "active":
		return "Активная"
	case "resolved":
		return "Решена"
	case "postponed":
		return "Отложена"
	case "archived":
		return "Архив"
	default:
		return status
	}
}

func parsePositiveInt(s string) (int, bool) {
	s = strings.TrimSpace(s)
	if s == "" {
		return 0, false
	}
	n, err := strconv.Atoi(s)
	if err != nil {
		return 0, false
	}
	if n < 0 {
		return 0, false
	}
	return n, true
}

func earlyEndReasonLabel(code string) string {
	code = strings.TrimSpace(code)
	switch code {
	case "no_time":
		return "Нет времени"
	case "need_break":
		return "Нужна пауза"
	case "too_emotional":
		return "Слишком эмоционально"
	}
	if strings.HasPrefix(code, "other:") {
		v := strings.TrimSpace(strings.TrimPrefix(code, "other:"))
		if v != "" {
			return v
		}
		return "Другое"
	}
	if code != "" {
		return code
	}
	return "—"
}

func issuePriorityLabel(priority string) string {
	switch strings.TrimSpace(priority) {
	case "low":
		return "Низкая"
	case "medium":
		return "Средняя"
	case "high":
		return "Высокая"
	default:
		return priority
	}
}

func resultStatusLabel(status string) string {
	switch strings.TrimSpace(status) {
	case "resolved":
		return "Решено"
	case "partially_resolved":
		return "Частично"
	case "postponed":
		return "Вернуться позже"
	case "unresolved":
		return "Не договорились"
	default:
		return status
	}
}

func isForwardedMessage(msg *tgbotapi.Message) bool {
	if msg == nil {
		return false
	}
	if msg.ForwardFrom != nil || msg.ForwardFromChat != nil {
		return true
	}
	if msg.ForwardDate != 0 || msg.ForwardFromMessageID != 0 {
		return true
	}
	if msg.ForwardSenderName != "" {
		return true
	}
	if msg.IsAutomaticForward {
		return true
	}
	return false
}

func formatForwardedMessage(msg *tgbotapi.Message) string {
	if msg == nil {
		return ""
	}
	var b strings.Builder
	from := ""
	if msg.ForwardFrom != nil {
		from = strings.TrimSpace(strings.Join([]string{msg.ForwardFrom.FirstName, msg.ForwardFrom.LastName}, " "))
		if from == "" {
			from = msg.ForwardFrom.UserName
		}
	}
	if from == "" && msg.ForwardFromChat != nil {
		from = msg.ForwardFromChat.Title
	}
	if from == "" && msg.ForwardSenderName != "" {
		from = msg.ForwardSenderName
	}
	if from != "" {
		b.WriteString("От: " + from + "\n")
	}
	body := strings.TrimSpace(msg.Text)
	if body == "" && msg.Caption != "" {
		body = strings.TrimSpace(msg.Caption)
	}
	if body == "" {
		body = "[без текста]"
	}
	const max = 800
	if len(body) > max {
		body = body[:max] + "…"
	}
	b.WriteString(body)
	return b.String()
}
