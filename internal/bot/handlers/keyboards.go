package handlers

import (
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

func menuKeyboard() *tgbotapi.InlineKeyboardMarkup {
	kb := tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("➕ Новая тема", cbIssueNew),
		),
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("📋 Темы", cbTopics),
			tgbotapi.NewInlineKeyboardButtonData("🎯 Фокус", cbNavFocus),
		),
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("📝 Мысли", cbNavNotes),
			tgbotapi.NewInlineKeyboardButtonData("📚 История", cbNavHistory),
		),
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("⚙️ Настройки", cbNavSettings),
		),
	)
	return &kb
}

func onboardingKeyboard() *tgbotapi.InlineKeyboardMarkup {
	kb := tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("Создать пару", cbPairCreate),
		),
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("Войти по ссылке", cbHowTo),
		),
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("Тестовый режим", cbTestStart),
			tgbotapi.NewInlineKeyboardButtonData("Как пользоваться", cbHowTo),
		),
	)
	return &kb
}

func noPairKeyboard() *tgbotapi.InlineKeyboardMarkup {
	kb := tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("Создать пару", cbPairCreate),
		),
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("Тестовый режим", cbTestStart),
		),
	)
	return &kb
}

func backMenuKeyboard() *tgbotapi.InlineKeyboardMarkup {
	return nil
}

func cancelKeyboard() *tgbotapi.InlineKeyboardMarkup {
	kb := tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("Отмена", cbCancel),
		),
	)
	return &kb
}

func backCancelKeyboard(back string) *tgbotapi.InlineKeyboardMarkup {
	kb := tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("Назад", back),
			tgbotapi.NewInlineKeyboardButtonData("Отмена", cbCancel),
		),
	)
	return &kb
}

func backKeyboard(back string) *tgbotapi.InlineKeyboardMarkup {
	if back == "" {
		return nil
	}
	kb := tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("Назад", back),
		),
	)
	return &kb
}

func backToMenuKeyboard() *tgbotapi.InlineKeyboardMarkup {
	kb := tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("Назад", cbMenu),
		),
	)
	return &kb
}

func quickIssueConfirmKeyboard() *tgbotapi.InlineKeyboardMarkup {
	kb := tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("✅ Создать", cbIssueQuickYes),
			tgbotapi.NewInlineKeyboardButtonData("❌ Не сейчас", cbIssueQuickNo),
		),
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("Меню", cbMenu),
		),
	)
	return &kb
}

func stateKeyboard(prefix string, back string) *tgbotapi.InlineKeyboardMarkup {
	var rows [][]tgbotapi.InlineKeyboardButton
	rows = append(rows, tgbotapi.NewInlineKeyboardRow(
		tgbotapi.NewInlineKeyboardButtonData("Спокойно", prefix+"calm"),
		tgbotapi.NewInlineKeyboardButtonData("Напряженно", prefix+"tense"),
	))
	rows = append(rows, tgbotapi.NewInlineKeyboardRow(
		tgbotapi.NewInlineKeyboardButtonData("Раздраженно", prefix+"irritated"),
		tgbotapi.NewInlineKeyboardButtonData("Устал(а)", prefix+"tired"),
	))
	rows = append(rows, tgbotapi.NewInlineKeyboardRow(
		tgbotapi.NewInlineKeyboardButtonData("Не готов(а)", prefix+"not_ready"),
	))
	if back != "" {
		rows = append(rows, tgbotapi.NewInlineKeyboardRow(tgbotapi.NewInlineKeyboardButtonData("Назад", back)))
	}
	kb := tgbotapi.NewInlineKeyboardMarkup(rows...)
	return &kb
}

func pairCreatedKeyboard() *tgbotapi.InlineKeyboardMarkup {
	kb := tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("Проверить подключение", cbPairCheck),
		),
	)
	return &kb
}

func inviteConfirmKeyboard(token string) *tgbotapi.InlineKeyboardMarkup {
	kb := tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("Подтвердить", cbInviteConfirmPrefix+token),
			tgbotapi.NewInlineKeyboardButtonData("Отмена", cbInviteCancel),
		),
	)
	return &kb
}

func priorityKeyboard() *tgbotapi.InlineKeyboardMarkup {
	kb := tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("Низкая", "issue:prio:low"),
			tgbotapi.NewInlineKeyboardButtonData("Средняя", "issue:prio:medium"),
			tgbotapi.NewInlineKeyboardButtonData("Высокая", "issue:prio:high"),
		),
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("Назад", "issue:back_desc"),
			tgbotapi.NewInlineKeyboardButtonData("Отмена", cbCancel),
		),
	)
	return &kb
}

func visibilityKeyboard() *tgbotapi.InlineKeyboardMarkup {
	kb := tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("Видно сразу", "issue:vis:visible"),
		),
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("Открыть после повторений", "issue:vis:hidden_until_repeats"),
		),
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("Пока личная", "issue:vis:private"),
		),
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("Назад", "issue:back_prio"),
			tgbotapi.NewInlineKeyboardButtonData("Отмена", cbCancel),
		),
	)
	return &kb
}

func thresholdKeyboard() *tgbotapi.InlineKeyboardMarkup {
	kb := tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("1", "issue:thr:1"),
			tgbotapi.NewInlineKeyboardButtonData("2", "issue:thr:2"),
			tgbotapi.NewInlineKeyboardButtonData("3", "issue:thr:3"),
			tgbotapi.NewInlineKeyboardButtonData("5", "issue:thr:5"),
		),
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("Назад", "issue:back_vis"),
			tgbotapi.NewInlineKeyboardButtonData("Отмена", cbCancel),
		),
	)
	return &kb
}

func issueCreatedKeyboard(issueID string) *tgbotapi.InlineKeyboardMarkup {
	kb := tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("Открыть тему", cbIssueOpenPrefix+issueID),
		),
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("Добавить еще", cbIssueNew),
		),
	)
	return &kb
}

func issueCardKeyboard(issueID string, status string) *tgbotapi.InlineKeyboardMarkup {
	var rows [][]tgbotapi.InlineKeyboardButton
	rows = append(rows, tgbotapi.NewInlineKeyboardRow(
		tgbotapi.NewInlineKeyboardButtonData("🔁 Повторилось", cbIssueRepeatPrefix+issueID),
		tgbotapi.NewInlineKeyboardButtonData("🎯 Начать фокус", cbFocusStartPrefix+issueID),
	))
	rows = append(rows, tgbotapi.NewInlineKeyboardRow(
		tgbotapi.NewInlineKeyboardButtonData("📝 Повторения", cbIssueRepeatsPrefix+issueID),
		tgbotapi.NewInlineKeyboardButtonData("📚 История", cbIssueHistoryPrefix+issueID),
	))
	if status == "resolved" {
		rows = append(rows, tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("♻️ Восстановить тему", cbIssueRestorePrefix+issueID),
		))
	} else {
		rows = append(rows, tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("✅ Закрыть тему", cbIssueClosePrefix+issueID),
		))
	}
	rows = append(rows, tgbotapi.NewInlineKeyboardRow(
		tgbotapi.NewInlineKeyboardButtonData("🗑 Удалить тему", cbIssueDeletePrefix+issueID),
	))
	rows = append(rows, tgbotapi.NewInlineKeyboardRow(
		tgbotapi.NewInlineKeyboardButtonData("В меню", cbMenu),
	))
	kb := tgbotapi.NewInlineKeyboardMarkup(rows...)
	return &kb
}

func issueCloseConfirmKeyboard(issueID string) *tgbotapi.InlineKeyboardMarkup {
	kb := tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("Подтвердить закрытие", cbIssueResolveConfirmPrefix+issueID),
			tgbotapi.NewInlineKeyboardButtonData("Назад", cbIssueOpenPrefix+issueID),
		),
	)
	return &kb
}

func issueDeleteConfirmKeyboard(issueID string) *tgbotapi.InlineKeyboardMarkup {
	kb := tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("🗑 Удалить", cbIssueDeleteYesPrefix+issueID),
			tgbotapi.NewInlineKeyboardButtonData("Назад", cbIssueOpenPrefix+issueID),
		),
	)
	return &kb
}

func repeatDeleteConfirmKeyboard(repeatID string) *tgbotapi.InlineKeyboardMarkup {
	kb := tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("🗑 Удалить", cbRepeatDeleteYesPrefix+repeatID),
			tgbotapi.NewInlineKeyboardButtonData("Назад", cbRepeatOpenPrefix+repeatID),
		),
	)
	return &kb
}

func noteDeleteConfirmKeyboard(noteID string) *tgbotapi.InlineKeyboardMarkup {
	kb := tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("🗑 Удалить", cbNoteDeleteYesPrefix+noteID),
			tgbotapi.NewInlineKeyboardButtonData("Назад", cbNavNotes),
		),
	)
	return &kb
}

func conversationKeyboard(conversationID string, paused bool) *tgbotapi.InlineKeyboardMarkup {
	var pauseBtn tgbotapi.InlineKeyboardButton
	if paused {
		pauseBtn = tgbotapi.NewInlineKeyboardButtonData("▶️ Продолжить", cbConvResumePrefix+conversationID)
	} else {
		pauseBtn = tgbotapi.NewInlineKeyboardButtonData("⏸ Пауза", cbConvPausePrefix+conversationID)
	}
	kb := tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(pauseBtn),
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("➕ Побочная тема", cbConvSidePrefix+conversationID),
			tgbotapi.NewInlineKeyboardButtonData("📝 Записать мысль", cbConvNotePrefix+conversationID),
		),
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("✅ Завершить", cbConvFinishPrefix+conversationID),
		),
	)
	return &kb
}

func finishStatusKeyboard() *tgbotapi.InlineKeyboardMarkup {
	kb := tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("Решено", cbFinishStatusPrefix+"resolved"),
			tgbotapi.NewInlineKeyboardButtonData("Частично", cbFinishStatusPrefix+"partially_resolved"),
		),
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("Вернуться позже", cbFinishStatusPrefix+"postponed"),
			tgbotapi.NewInlineKeyboardButtonData("Не договорились", cbFinishStatusPrefix+"unresolved"),
		),
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("Назад", cbConvBackToActive),
		),
	)
	return &kb
}
