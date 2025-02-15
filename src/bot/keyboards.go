package bot

import (
	"context"
	"fmt"
	tele "gopkg.in/telebot.v4"
	"strconv"
	"strings"
	"tarantulago/models"
	"time"
)

type Menu struct {
	main      *tele.ReplyMarkup
	tarantula *tele.ReplyMarkup
	settings  *tele.ReplyMarkup
	colony    *tele.ReplyMarkup
	back      tele.Btn
}

var (
	mainMarkup    = tele.ReplyMarkup{ResizeKeyboard: true}
	btnBackToMain = mainMarkup.Text("⬅️ Back to Main Menu")
	menu          = &Menu{
		main:      &mainMarkup,
		tarantula: &tele.ReplyMarkup{ResizeKeyboard: true},
		colony:    &tele.ReplyMarkup{ResizeKeyboard: true},
		settings:  &tele.ReplyMarkup{ResizeKeyboard: true},
		back:      btnBackToMain,
	}
	btnAddTarantula   = menu.tarantula.Text("➕ Add New Tarantula")
	btnListTarantulas = menu.tarantula.Text("📋 List Tarantulas")
	btnViewMolts      = menu.tarantula.Text("📊 View Molt History")

	btnColonyStatus = menu.colony.Text("📊 Colony Status")
	btnUpdateCount  = menu.colony.Text("🔢 Update Count")
	btnAddColony    = menu.colony.Text("➕ Add New Colony")

	btnTarantulas = menu.main.Text("🕷 Tarantulas")
	btnFeeding    = menu.main.Text("🪱 Feeding")
	btnColony     = menu.main.Text("🦗 Colony Management")
	btnSettings   = menu.main.Text("⚙️ Settings")

	btnNotifications = menu.settings.Text("🔔 Notification Settings")
	btnColonyAlerts  = menu.settings.Text("🦗 Colony Alerts")
)

func (m *Menu) init() {
	m.main.Reply(
		m.main.Row(btnTarantulas, btnFeeding),
		m.main.Row(btnColony, btnSettings),
	)

	m.tarantula.Reply(
		m.tarantula.Row(btnAddTarantula, btnListTarantulas),
		m.tarantula.Row(btnViewMolts),
		m.tarantula.Row(m.back),
	)

	m.colony.Reply(
		m.colony.Row(btnColonyStatus, btnUpdateCount),
		m.colony.Row(btnAddColony),
		m.colony.Row(m.back),
	)

	menu.settings.Reply(
		menu.settings.Row(btnNotifications),
		menu.settings.Row(btnColonyAlerts),
		menu.settings.Row(menu.back),
	)
}

func (t *TarantulaBot) setupHandlers() {
	menu.init()
	b := t.bot
	b.Handle("/start", func(c tele.Context) error {
		err := t.db.EnsureUserExists(context.Background(), &models.TelegramUser{
			TelegramID: c.Sender().ID,
			FirstName:  c.Sender().FirstName,
			LastName:   c.Sender().LastName,
			Username:   c.Sender().Username,
		})
		if err != nil {
			return fmt.Errorf("failed to ensure user exists: %w", err)
		}
		return c.Send("🕷 Welcome to TarantulaGo! Choose an option:", menu.main)
	})

	b.Handle(&btnTarantulas, func(c tele.Context) error {
		return c.Send("Tarantula Management:", menu.tarantula)
	})

	b.Handle(&btnColony, func(c tele.Context) error {
		return c.Send("Cricket Colony Management:", menu.colony)
	})

	b.Handle(&btnBackToMain, func(c tele.Context) error {
		return c.Send("Main Menu:", menu.main)
	})

	b.Handle(&btnBackToMain, func(c tele.Context) error {
		session := t.sessions.GetSession(c.Sender().ID)
		if session != nil {
			session.reset()
			t.sessions.UpdateSession(c.Sender().ID, session)
		}
		return c.Send("Main Menu:", menu.main)
	})

	b.Handle(&btnNotifications, t.handleNotificationSettings)
	b.Handle(&btnColonyAlerts, t.handleColonyAlertSettings)

	b.Handle(&btnSettings, func(c tele.Context) error {
		return c.Send("⚙️ Settings:", menu.settings)
	})

	b.Handle(&btnColonyStatus, func(c tele.Context) error {
		colonyStatuses, err := t.db.GetColonyStatus(context.Background(), c.Sender().ID)
		if err != nil {
			return fmt.Errorf("failed to get colony status: %w", err)
		}
		var msg string
		if colonyStatuses == nil {
			msg = "No cricket colonies found."
		} else {
			msg = "Colony Statuses:\n"
			for _, status := range colonyStatuses {
				msg += fmt.Sprintf("🦗 %s: %d\n", status.ColonyName, status.CurrentCount)
			}
		}
		return c.Send(msg)
	})

	b.Handle(&btnAddTarantula, func(c tele.Context) error {
		session := t.sessions.GetSession(c.Sender().ID)
		session.CurrentState = StateAddingTarantula
		session.CurrentField = FieldName
		t.sessions.UpdateSession(c.Sender().ID, session)

		return c.Send("Let's add a new tarantula! What's their name?")
	})

	t.bot.Handle(&btnListTarantulas, func(c tele.Context) error {
		return t.showTarantulaList(c)
	})

	b.Handle(&btnUpdateCount, func(c tele.Context) error {
		session := t.sessions.GetSession(c.Sender().ID)
		session.CurrentState = StateAddingCrickets
		session.CurrentField = FieldColonyID
		t.sessions.UpdateSession(c.Sender().ID, session)

		return c.Send("What's the colony ID?")
	})

	b.Handle(&btnAddColony, func(c tele.Context) error {
		session := t.sessions.GetSession(c.Sender().ID)
		session.CurrentState = StateAddingColony
		session.CurrentField = FieldColonyName
		t.sessions.UpdateSession(c.Sender().ID, session)

		return c.Send("Let's add a new cricket colony! What's the name?")
	})

	b.Handle(&btnViewMolts, func(c tele.Context) error {
		moltRecords, err := t.db.GetRecentMoltRecords(context.Background(), c.Sender().ID, 10)
		if err != nil {
			return fmt.Errorf("failed to get recent molt records: %w", err)
		}
		var msg string
		if len(moltRecords) == 0 {
			msg = "No molt records found."
		} else {
			msg = "Recent molt records:\n"
			for _, record := range moltRecords {
				msg += fmt.Sprintf("🕷 %s molted on %s\n", record.Tarantula.Name, record.MoltDate.Format("2006-01-02"))
			}
		}
		return c.Send(msg)
	})

	b.Handle(tele.OnText, func(c tele.Context) error {
		session := t.sessions.GetSession(c.Sender().ID)

		switch session.CurrentState {
		case StateAddingTarantula:
			return t.handleTarantulaFormInput(c, session)
		case StateFeeding:
			return t.handleFeedingFormInput(c, session)
		case StateAddingMolt:
			return t.handleMoltFormInput(c, session)
		case StateAddingColony:
			return t.handleColonyFormInput(c, session)
		case StateAddingCrickets:
			return t.handleCricketsFormInput(c, session)
		case StateNotificationSettings:
			return t.handleSettingsInput(c, session)
		default:
			return nil
		}
	})

	t.setupInlineKeyboards()
}

func sendError(c tele.Context, msg string) error {
	return c.Send(fmt.Sprintf("❌ Error: %s", msg))
}

func sendSuccess(c tele.Context, msg string) error {
	return c.Send(fmt.Sprintf("✅ Success: %s", msg))
}

func (t *TarantulaBot) showTarantulaList(c tele.Context) error {
	tarantulas, err := t.db.GetAllTarantulas(t.ctx, c.Sender().ID)
	if err != nil {
		return fmt.Errorf("failed to get tarantulas: %w", err)
	}

	if len(tarantulas) == 0 {
		return c.Send("No tarantulas found.")
	}

	var rows [][]tele.InlineButton
	for _, tarantula := range tarantulas {
		btn := tele.InlineButton{
			Text: fmt.Sprintf("🕷 %s", tarantula.Name),
			Data: fmt.Sprintf("select:%d", tarantula.ID),
		}
		rows = append(rows, []tele.InlineButton{btn})
	}

	return c.Send("Select a tarantula:", &tele.ReplyMarkup{
		InlineKeyboard: rows,
	})
}

func (t *TarantulaBot) handleTarantulaSelect(c tele.Context, tarantulaID int) error {
	tarantula, err := t.db.GetTarantulaByID(t.ctx, c.Sender().ID, int32(tarantulaID))
	if err != nil {
		return fmt.Errorf("failed to get tarantula: %w", err)
	}

	markup := makeTarantulaMarkup(tarantulaID)

	msg := fmt.Sprintf(
		"🕷 *%s*\n"+
			"Species: %s\n"+
			"Last Molt: %s\n"+
			"Health Status: %s",
		tarantula.Name,
		tarantula.Species.CommonName,
		formatDate(tarantula.LastMoltDate),
		getHealthStatus(tarantula.CurrentHealthStatusID),
	)

	if c.Callback() != nil {
		return c.Edit(msg, markup)
	}

	return c.Send(msg, markup)
}

func makeTarantulaMarkup(tarantulaID int) *tele.ReplyMarkup {
	markup := &tele.ReplyMarkup{}

	feedBtn := tele.InlineButton{
		Data: fmt.Sprintf("%s:%d", feedCallback, tarantulaID),
		Text: "🍽 Feed",
	}

	moltBtn := tele.InlineButton{
		Text: "🔄 Record Molt", Data: fmt.Sprintf("%s:%d", moltCallback, tarantulaID)}

	feedSchedulerBtn := tele.InlineButton{
		Text: "📅 Schedule Feedings", Data: fmt.Sprintf("%s:%d", feedSchedulerCallback, tarantulaID),
	}

	backBtn := tele.InlineButton{Text: "⬅️ Back to List",
		Data: backToListCallback + ":0"}

	markup.InlineKeyboard = [][]tele.InlineButton{
		{feedBtn, moltBtn, feedSchedulerBtn},
		{backBtn},
	}
	return markup
}

func (t *TarantulaBot) handleTarantulaFeed(c tele.Context, tarantulaID int) error {
	session := t.sessions.GetSession(c.Sender().ID)
	session.CurrentState = StateFeeding
	session.CurrentField = FieldColonyID
	session.FeedEvent.TarantulaID = tarantulaID
	t.sessions.UpdateSession(c.Sender().ID, session)

	return c.Send("What's colony ID?")
}

func (t *TarantulaBot) handleTarantulaMolt(c tele.Context, tarantulaID int) error {
	session := t.sessions.GetSession(c.Sender().ID)
	session.CurrentState = StateAddingMolt
	session.CurrentField = FieldPreMoltLengthCM
	session.MoltData.TarantulaID = tarantulaID
	t.sessions.UpdateSession(c.Sender().ID, session)

	return c.Send("How long was your tarantula before it molted (in cm)?")
}

func (t *TarantulaBot) handleTarantulaInfo(c tele.Context, id int) error {
	tarantula, err := t.db.GetTarantulaByID(t.ctx, c.Sender().ID, int32(id))
	if err != nil {
		return fmt.Errorf("failed to get tarantula: %w", err)
	}

	return sendTarantulaInfo(c, tarantula)
}

func sendTarantulaInfo(c tele.Context, tarantula *models.Tarantula) error {
	msg := fmt.Sprintf(
		"🕷 *%s*\n"+
			"Species: %s\n"+
			"Last Molt: %s\n"+
			"Health Status: %s",
		tarantula.Name,
		tarantula.Species.CommonName,
		formatDate(tarantula.LastMoltDate),
		getHealthStatus(tarantula.CurrentHealthStatusID),
	)

	return c.Send(msg)
}

func formatDate(t *time.Time) string {
	if t == nil {
		return "Never"
	}
	return t.Format("2006-01-02")
}

func getHealthStatus(statusID int) string {
	return models.HealthStatusFromID(int32(statusID)).Description()
}

func (t *TarantulaBot) handleNotificationSettings(c tele.Context) error {
	settings, err := t.db.GetUserSettings(t.ctx, c.Sender().ID)
	if err != nil {
		return fmt.Errorf("failed to get user settings: %w", err)
	}
	session := t.sessions.GetSession(c.Sender().ID)
	session.CurrentState = StateNotificationSettings
	t.sessions.UpdateSession(c.Sender().ID, session)
	markup := &tele.ReplyMarkup{}

	toggleText := "🔕 Disable Notifications"
	if !settings.NotificationEnabled {
		toggleText = "🔔 Enable Notifications"
	}

	toggleBtn := tele.InlineButton{
		Text: toggleText,
		Data: "toggle_notifications",
	}

	timeBtn := tele.InlineButton{
		Text: fmt.Sprintf("⏰ Notification Time: %s UTC", settings.NotificationTimeUTC),
		Data: "set_notification_time",
	}

	reminderBtn := tele.InlineButton{
		Text: fmt.Sprintf("📅 Feeding Reminder: %d days", settings.FeedingReminderDays),
		Data: "set_feeding_reminder",
	}

	markup.InlineKeyboard = [][]tele.InlineButton{
		{toggleBtn},
		{timeBtn},
		{reminderBtn},
	}

	return c.Send("🔔 Notification Settings:", markup)
}

func (t *TarantulaBot) handleColonyAlertSettings(c tele.Context) error {
	settings, err := t.db.GetUserSettings(t.ctx, c.Sender().ID)
	if err != nil {
		return fmt.Errorf("failed to get user settings: %w", err)
	}

	markup := &tele.ReplyMarkup{}

	thresholdBtn := tele.InlineButton{
		Text: fmt.Sprintf("🦗 Low Colony Alert: %d crickets", settings.LowColonyThreshold),
		Data: "set_colony_threshold",
	}

	markup.InlineKeyboard = [][]tele.InlineButton{
		{thresholdBtn},
	}

	return c.Send("🦗 Colony Alert Settings:", markup)
}

func (t *TarantulaBot) handleToggleNotifications(c tele.Context) error {
	settings, err := t.db.GetUserSettings(t.ctx, c.Sender().ID)
	if err != nil {
		return fmt.Errorf("failed to get user settings: %w", err)
	}

	settings.NotificationEnabled = !settings.NotificationEnabled
	err = t.db.UpdateUserSettings(t.ctx, settings)
	if err != nil {
		return fmt.Errorf("failed to update user settings: %w", err)
	}

	status := "enabled"
	if !settings.NotificationEnabled {
		status = "disabled"
	}

	return c.Send(fmt.Sprintf("✅ Notifications %s", status))
}

func (t *TarantulaBot) handleSetNotificationTime(c tele.Context) error {
	session := t.sessions.GetSession(c.Sender().ID)
	session.CurrentField = "notification_time"
	t.sessions.UpdateSession(c.Sender().ID, session)

	return c.Send("Please enter the time you want to receive notifications (HH:MM in UTC)")
}

func (t *TarantulaBot) handleSetFeedingReminder(c tele.Context) error {
	session := t.sessions.GetSession(c.Sender().ID)
	session.CurrentField = "feeding_reminder"
	t.sessions.UpdateSession(c.Sender().ID, session)

	return c.Send("How many days before feeding would you like to be reminded?")
}

func (t *TarantulaBot) handleSetColonyThreshold(c tele.Context) error {
	session := t.sessions.GetSession(c.Sender().ID)
	session.CurrentField = "colony_threshold"
	t.sessions.UpdateSession(c.Sender().ID, session)

	return c.Send("At what number of crickets should I warn you about low colony count?")
}

func (t *TarantulaBot) handleSettingsInput(c tele.Context, session *UserSession) error {
	settings, err := t.db.GetUserSettings(t.ctx, c.Sender().ID)
	if err != nil {
		return fmt.Errorf("failed to get user settings: %w", err)
	}

	switch session.CurrentField {
	case "notification_time":
		if len(c.Text()) != 5 || c.Text()[2] != ':' {
			return c.Send("Please use HH:MM format (e.g., 14:30)")
		}
		settings.NotificationTimeUTC = c.Text()

	case "feeding_reminder":
		days, err := strconv.Atoi(c.Text())
		if err != nil || days <= 0 {
			return c.Send("Please enter a valid number of days (greater than 0)")
		}
		settings.FeedingReminderDays = days

	case "colony_threshold":
		threshold, err := strconv.Atoi(c.Text())
		if err != nil || threshold <= 0 {
			return c.Send("Please enter a valid number of crickets (greater than 0)")
		}
		settings.LowColonyThreshold = threshold
	}

	err = t.db.UpdateUserSettings(t.ctx, settings)
	if err != nil {
		return fmt.Errorf("failed to update user settings: %w", err)
	}

	session.reset()
	t.sessions.UpdateSession(c.Sender().ID, session)

	return c.Send("✅ Settings updated successfully!")
}

func (t *TarantulaBot) handleFeedScheduler(c tele.Context, tarantulaId int) error {
	tarantula, err := t.db.GetTarantulaByID(t.ctx, c.Sender().ID, int32(tarantulaId))
	if err != nil {
		return fmt.Errorf("failed to get tarantula: %w", err)
	}
	schedule, err := t.db.GetFeedingSchedule(t.ctx, int64(tarantula.ID), float32(tarantula.CurrentSize))
	if err != nil {
		return fmt.Errorf("failed to get feeding schedule: %w", err)
	}
	var msg strings.Builder
	if schedule == nil {
		msg.WriteString("No feeding schedule found.")
	} else {
		msg.WriteString(fmt.Sprintf("🕷 Feeding Schedule for %s %.2f:\n", tarantula.Species.CommonName, schedule.BodyLengthCM))
		msg.WriteString(fmt.Sprintf("📏 Prey size: %s\n", schedule.PreySize))
		msg.WriteString(fmt.Sprintf("🦗 Prey type: %s\n", schedule.PreyType))

		msg.WriteString(fmt.Sprintf("⏰ Frequency: %s\n", schedule.Frequency.FrequencyName))
		msg.WriteString(fmt.Sprintf("ℹ️ Additional: %s\n", schedule.Frequency.Description))
		msg.WriteString(fmt.Sprintf("📝 Notes: %s\n", schedule.Notes))
	}
	return c.Send(msg.String())
}
