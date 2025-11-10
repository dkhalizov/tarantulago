package bot

import (
	"context"
	"fmt"
	"io"
	"strconv"
	"strings"
	"time"

	"tarantulago/models"

	tele "gopkg.in/telebot.v4"
)

type Menu struct {
	main             *tele.ReplyMarkup
	tarantula        *tele.ReplyMarkup
	settings         *tele.ReplyMarkup
	colony           *tele.ReplyMarkup
	tarantulaColony  *tele.ReplyMarkup
	analytics        *tele.ReplyMarkup
	back             tele.Btn
}

var (
	mainMarkup    = tele.ReplyMarkup{ResizeKeyboard: true}
	btnBackToMain = mainMarkup.Text("⬅️ Back to Main Menu")
	menu          = &Menu{
		main:            &mainMarkup,
		tarantula:       &tele.ReplyMarkup{ResizeKeyboard: true},
		colony:          &tele.ReplyMarkup{ResizeKeyboard: true},
		tarantulaColony: &tele.ReplyMarkup{ResizeKeyboard: true},
		settings:        &tele.ReplyMarkup{ResizeKeyboard: true},
		analytics:       &tele.ReplyMarkup{ResizeKeyboard: true},
		back:            btnBackToMain,
	}
	btnAddTarantula     = menu.tarantula.Text("➕ Add New Tarantula")
	btnListTarantulas   = menu.tarantula.Text("📋 List Tarantulas")
	btnViewMolts        = menu.tarantula.Text("📊 View Molt History")
	btnQuickActions     = menu.tarantula.Text("⚡ Quick Actions")
	btnManageColonies   = menu.tarantula.Text("👥 Manage Colonies")

	btnColonyStatus   = menu.colony.Text("📊 Cricket Status")
	btnUpdateCount    = menu.colony.Text("🔢 Update Cricket Count")
	btnFeedingHistory = menu.colony.Text("📈 Feeding History")

	btnCreateColony    = menu.tarantulaColony.Text("➕ Create Colony")
	btnListColonies    = menu.tarantulaColony.Text("📋 List Colonies")
	btnAddToColony     = menu.tarantulaColony.Text("👤 Add Member")
	btnBackToTarantula = menu.tarantulaColony.Text("⬅️ Back to Tarantulas")

	btnFeedingPatterns = menu.analytics.Text("🍽️ Feeding Patterns")
	btnGrowthCharts    = menu.analytics.Text("📈 Growth Charts")
	btnAnnualReports   = menu.analytics.Text("📋 Annual Reports")
	btnMoltPredictions = menu.analytics.Text("🔮 Molt Predictions")

	btnTarantulas = menu.main.Text("🕷 Tarantulas")
	btnFeeding    = menu.main.Text("🪱 Feeding")
	btnColony     = menu.main.Text("🦗 Quick Feed")
	btnAnalytics  = menu.main.Text("📊 Analytics")
	btnSettings   = menu.main.Text("⚙️ Settings")

	btnNotifications      = menu.settings.Text("🔔 Notification Settings")
	btnPauseNotifications = menu.settings.Text("⏸️ Pause Notifications")
)

func (m *Menu) init() {
	m.main.Reply(
		m.main.Row(btnTarantulas, btnFeeding),
		m.main.Row(btnColony, btnAnalytics),
		m.main.Row(btnSettings),
	)

	m.tarantula.Reply(
		m.tarantula.Row(btnAddTarantula, btnListTarantulas),
		m.tarantula.Row(btnViewMolts, btnQuickActions),
		m.tarantula.Row(btnManageColonies),
		m.tarantula.Row(m.back),
	)

	m.tarantulaColony.Reply(
		m.tarantulaColony.Row(btnCreateColony, btnListColonies),
		m.tarantulaColony.Row(btnAddToColony),
		m.tarantulaColony.Row(btnBackToTarantula),
	)

	m.colony.Reply(
		m.colony.Row(btnColonyStatus, btnUpdateCount),
		m.colony.Row(btnFeedingHistory),
		m.colony.Row(m.back),
	)

	m.analytics.Reply(
		m.analytics.Row(btnFeedingPatterns, btnGrowthCharts),
		m.analytics.Row(btnAnnualReports, btnMoltPredictions),
		m.analytics.Row(m.back),
	)

	menu.settings.Reply(
		menu.settings.Row(btnNotifications),
		menu.settings.Row(btnPauseNotifications),
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
			ChatID:     c.Chat().ID,
			LastName:   c.Sender().LastName,
			Username:   c.Sender().Username,
		})
		if err != nil {
			return fmt.Errorf("failed to ensure user exists: %w", err)
		}
		return c.Send("🕷 Welcome to TarantulaGo! Choose an option:", menu.main)
	})

	b.Handle("/check", func(c tele.Context) error {
		settings, err := t.db.GetUserSettings(t.ctx, c.Sender().ID)
		if err != nil {
			return fmt.Errorf("failed to get user settings: %w", err)
		}
		t.notifications.triggerChecks(models.TelegramUser{
			TelegramID: c.Sender().ID,
			FirstName:  c.Sender().FirstName,
			ChatID:     c.Chat().ID,
			LastName:   c.Sender().LastName,
			Username:   c.Sender().Username,
		}, settings)
		return c.Send("🔔 Notifications checked!")
	})

	b.Handle(&btnTarantulas, func(c tele.Context) error {
		return c.Send("Tarantula Management:", menu.tarantula)
	})

	b.Handle(&btnColony, func(c tele.Context) error {
		return t.handleQuickActions(c)
	})

	b.Handle(&btnQuickActions, func(c tele.Context) error {
		return t.handleQuickActions(c)
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
	b.Handle(&btnPauseNotifications, t.handlePauseNotificationSettings)

	b.Handle(&btnSettings, func(c tele.Context) error {
		return c.Send("⚙️ Settings:", menu.settings)
	})

	b.Handle(&btnAnalytics, func(c tele.Context) error {
		return c.Send("📊 Analytics & Reports:", menu.analytics)
	})

	b.Handle(&btnFeedingPatterns, func(c tele.Context) error {
		patterns, err := t.db.GetAllFeedingPatterns(context.Background(), c.Sender().ID)
		if err != nil {
			return SendError(c, fmt.Sprintf("Error: %v", err))
		}

		if len(patterns) == 0 {
			return SendInfo(c, "No feeding data available yet. Feed your tarantulas to generate patterns!")
		}

		msg := "🍽️ *Feeding Pattern Analysis*\n\n"
		for _, pattern := range patterns {
			msg += FormatFeedingPattern(pattern) + "\n"
		}

		return c.Send(msg, tele.ModeMarkdown)
	})

	b.Handle(&btnGrowthCharts, func(c tele.Context) error {
		growthData, err := t.db.GetAllGrowthData(context.Background(), c.Sender().ID)
		if err != nil {
			return SendError(c, fmt.Sprintf("Error: %v", err))
		}

		if len(growthData) == 0 {
			return SendInfo(c, "No growth data available yet. Record molts to generate size progression charts!")
		}

		msg := "📈 *Growth Tracking Charts*\n\n"
		for _, data := range growthData {
			msg += FormatGrowthData(data) + "\n"
		}

		return c.Send(msg, tele.ModeMarkdown)
	})

	b.Handle(&btnAnnualReports, func(c tele.Context) error {
		currentYear := time.Now().Year()
		reports, err := t.db.GetAllAnnualReports(context.Background(), currentYear, c.Sender().ID)
		if err != nil {
			return fmt.Errorf("failed to get annual reports: %w", err)
		}

		if len(reports) == 0 {
			return c.Send("📋 No data for " + strconv.Itoa(currentYear) + " yet. Keep tracking your tarantulas!")
		}

		msg := fmt.Sprintf("📋 *Annual Report %d*\n\n", currentYear)

		var totalCrickets int32
		var totalCost float64

		for _, report := range reports {
			msg += fmt.Sprintf("*%s*\n", report.TarantulaName)
			msg += fmt.Sprintf("🍽️ Fed %d times (%d crickets)\n", report.TotalFeedings, report.TotalCrickets)
			msg += fmt.Sprintf("✅ %.1f%% acceptance rate\n", report.AcceptanceRate)

			if report.MoltCount > 0 {
				msg += fmt.Sprintf("🔄 Molted %d time(s)\n", report.MoltCount)
			}

			if report.PhotosAdded > 0 {
				msg += fmt.Sprintf("📸 Added %d photos\n", report.PhotosAdded)
			}

			if len(report.Milestones) > 0 {
				msg += "🏆 Milestones:\n"
				for _, milestone := range report.Milestones {
					msg += fmt.Sprintf("  • %s\n", milestone)
				}
			}

			msg += fmt.Sprintf("💰 Est. cost: $%.2f\n\n", report.EstimatedCost)

			totalCrickets += report.TotalCrickets
			totalCost += report.EstimatedCost
		}

		msg += "📊 *Total Summary*\n"
		msg += fmt.Sprintf("🦗 %d crickets consumed\n", totalCrickets)
		msg += fmt.Sprintf("💰 $%.2f estimated cost\n", totalCost)

		return c.Send(msg, tele.ModeMarkdown)
	})

	b.Handle(&btnMoltPredictions, func(c tele.Context) error {
		return t.handleMoltPredictionsOverview(c)
	})

	b.Handle(&btnColonyStatus, func(c tele.Context) error {
		colonyStatuses, err := t.db.GetColonyStatus(context.Background(), c.Sender().ID)
		if err != nil {
			return fmt.Errorf("failed to get colony status: %w", err)
		}
		var msg string
		if len(colonyStatuses) == 0 {
			msg = "🦗 No cricket colony found.\n\n💡 Use 'Update Cricket Count' to set your current cricket amount!"
		} else {
			colony := colonyStatuses[0]
			msg = fmt.Sprintf("🦗 *Cricket Status*\n\n"+
				"Current Count: *%d crickets*\n"+
				"Used in last 7 days: *%d crickets*\n",
				colony.CurrentCount, colony.CricketsUsed7Days)

			if colony.WeeksRemaining != nil {
				if *colony.WeeksRemaining < 2 {
					msg += fmt.Sprintf("⚠️ Low stock: ~%.1f weeks remaining\n", *colony.WeeksRemaining)
				} else {
					msg += fmt.Sprintf("✅ Stock: ~%.1f weeks remaining\n", *colony.WeeksRemaining)
				}
			}
		}
		return c.Send(msg, tele.ModeMarkdown)
	})

	b.Handle(&btnAddTarantula, func(c tele.Context) error {
		session := t.sessions.GetSession(c.Sender().ID)
		session.CurrentState = StateAddingTarantula
		session.CurrentField = FieldName
		t.sessions.UpdateSession(c.Sender().ID, session)

		return c.Send("Let's add a new tarantula! What's their name?")
	})

	b.Handle(&btnListTarantulas, func(c tele.Context) error {
		return t.showTarantulaList(c)
	})

	// Tarantula Colony Management Handlers
	b.Handle(&btnManageColonies, func(c tele.Context) error {
		return c.Send("👥 Colony Management:", menu.tarantulaColony)
	})

	b.Handle(&btnBackToTarantula, func(c tele.Context) error {
		return c.Send("Tarantula Management:", menu.tarantula)
	})

	b.Handle(&btnCreateColony, func(c tele.Context) error {
		return t.handleCreateColony(c)
	})

	b.Handle(&btnListColonies, func(c tele.Context) error {
		return t.handleListColonies(c)
	})

	b.Handle(&btnAddToColony, func(c tele.Context) error {
		return t.handleAddToColony(c)
	})

	b.Handle(&btnUpdateCount, func(c tele.Context) error {
		session := t.sessions.GetSession(c.Sender().ID)
		session.CurrentState = StateAddingCrickets
		session.CurrentField = FieldColonyCount
		t.sessions.UpdateSession(c.Sender().ID, session)

		return c.Send("🦗 Enter your current cricket count:")
	})

	b.Handle(&btnFeeding, func(c tele.Context) error {

		recentFeedings, err := t.db.GetRecentFeedingRecords(t.ctx, c.Sender().ID, 10)
		if err != nil {
			return SendError(c, fmt.Sprintf("Failed to get feeding records: %v", err))
		}

		tarantulas, err := t.db.GetAllTarantulas(t.ctx, c.Sender().ID)
		if err != nil {
			return SendError(c, fmt.Sprintf("Failed to get tarantulas: %v", err))
		}

		var msg strings.Builder
		msg.WriteString("🍽️ *Feeding Dashboard*\n\n")

		if len(tarantulas) > 0 {
			msg.WriteString("📊 *Feeding Status Overview:*\n")
			for _, spider := range tarantulas {
				daysSince := int(spider.DaysSinceFeeding)
				statusEmoji, _ := GetFeedingStatusWithMolt(daysSince, int(spider.MinDays), int(spider.MaxDays), spider.CurrentStatus)

				lastFed := "Never"
				if daysSince < 999 {
					lastFed = fmt.Sprintf("%d days ago", daysSince)
				}

				msg.WriteString(fmt.Sprintf("%s *%s* - %s\n", statusEmoji, spider.Name, lastFed))
			}
			msg.WriteString("\n")
		}

		if len(recentFeedings) == 0 {
			msg.WriteString("📝 *Recent Activity:*\nNo feeding records found.\n")
		} else {
			msg.WriteString("📝 *Recent Feeding History:*\n")
			for i, record := range recentFeedings {
				if i >= 8 {
					break
				}

				status := "✅"
				if record.FeedingStatus.StatusName == "Rejected" {
					status = "❌"
				}

				msg.WriteString(fmt.Sprintf("%s *%s* • %s • %s\n",
					status,
					record.Tarantula.Name,
					FormatDate(&record.FeedingDate),
					FormatDaysAgo(&record.FeedingDate)))
			}
		}

		markup := &tele.ReplyMarkup{}
		var buttons [][]tele.InlineButton

		if len(tarantulas) > 0 {

			needsFeeding := 0
			overdue := 0
			for _, spider := range tarantulas {
				daysSince := int(spider.DaysSinceFeeding)
				if daysSince >= int(spider.MaxDays) {
					overdue++
				} else if daysSince >= int(spider.MinDays) {
					needsFeeding++
				}
			}

			if overdue > 0 || needsFeeding > 0 {
				msg.WriteString("\n⚠️ *Attention Needed:*\n")
				if overdue > 0 {
					msg.WriteString(fmt.Sprintf("🔴 %d tarantula(s) overdue for feeding\n", overdue))
				}
				if needsFeeding > 0 {
					msg.WriteString(fmt.Sprintf("🟡 %d tarantula(s) ready for feeding\n", needsFeeding))
				}
			}

			btnQuickFeed := tele.InlineButton{
				Text: "🚀 Quick Feed",
				Data: "quick_actions",
			}

			btnFeedingHistory := tele.InlineButton{
				Text: "📊 Full History",
				Data: "feeding_history",
			}

			buttons = append(buttons, []tele.InlineButton{btnQuickFeed, btnFeedingHistory})
		}

		btnBack := tele.InlineButton{
			Text: "⬅️ Back to Main",
			Data: "back_to_main",
		}
		buttons = append(buttons, []tele.InlineButton{btnBack})

		markup.InlineKeyboard = buttons
		return c.Send(msg.String(), markup, tele.ModeMarkdown)
	})

	b.Handle(&btnFeedingHistory, func(c tele.Context) error {
		feedings, err := t.db.GetRecentFeedingRecords(t.ctx, c.Sender().ID, 20)
		if err != nil {
			return SendError(c, fmt.Sprintf("Failed to get feeding records: %v", err))
		}

		var msg strings.Builder
		msg.WriteString("📊 *Complete Feeding History*\n\n")

		if len(feedings) == 0 {
			msg.WriteString("No feeding records found.\n")
		} else {

			spiderFeedings := make(map[string][]models.FeedingEvent)
			for _, feeding := range feedings {
				spiderFeedings[feeding.Tarantula.Name] = append(spiderFeedings[feeding.Tarantula.Name], feeding)
			}

			for spiderName, records := range spiderFeedings {
				msg.WriteString(fmt.Sprintf("🕷 *%s:*\n", spiderName))

				for i, record := range records {
					if i >= 5 {
						msg.WriteString("   _...and more_\n")
						break
					}

					status := "✅"
					if record.FeedingStatus.StatusName == "Rejected" {
						status = "❌"
					}

					msg.WriteString(fmt.Sprintf("  %s %s • %d 🦗 • %s\n",
						status,
						FormatDate(&record.FeedingDate),
						record.NumberOfCrickets,
						FormatDaysAgo(&record.FeedingDate)))
				}
				msg.WriteString("\n")
			}
		}

		markup := &tele.ReplyMarkup{
			InlineKeyboard: [][]tele.InlineButton{
				{{Text: "⬅️ Back to Feeding", Data: "feeding_dashboard"}},
			},
		}

		return c.Send(msg.String(), markup, tele.ModeMarkdown)
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
		case StateCreatingColony:
			return t.handleTarantulaColonyFormInput(c, session)

		default:
			return nil
		}
	})

	b.Handle(tele.OnPhoto, func(c tele.Context) error {
		session := t.sessions.GetSession(c.Sender().ID)
		if session.CurrentState == StateAddingPhoto {
			return t.handlePhotoInput(c, session)
		}
		return nil
	})

	b.Handle(&btnViewMolts, t.handleViewMolts)

	t.setupColonyMaintenanceHandlers()
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
	msg := fmt.Sprintf("🕷 *%s*\n\n", tarantula.Name)
	msg += fmt.Sprintf("Species: %s\n", tarantula.Species.CommonName)
	msg += fmt.Sprintf("Acquired: %s\n", tarantula.AcquisitionDate.Format("2006-01-02"))

	msg += fmt.Sprintf("Size: %.1fcm\n", tarantula.CurrentSize)
	msg += fmt.Sprintf("Age: %d months\n", tarantula.EstimatedAgeMonths)
	msg += fmt.Sprintf("Health: %s\n", getHealthStatus(tarantula.CurrentHealthStatusID))

	if tarantula.LastMoltDate != nil {
		msg += fmt.Sprintf("Last molt: %s\n", formatDate(tarantula.LastMoltDate))
	}

	if tarantula.Notes != "" {
		msg += fmt.Sprintf("Notes: %s\n", tarantula.Notes)
	}

	// Add photo action buttons only
	markup := &tele.ReplyMarkup{}

	photoBtn := tele.InlineButton{
		Text: "📸 Add Photo",
		Data: fmt.Sprintf("add_photo:%d", tarantula.ID),
	}

	viewPhotosBtn := tele.InlineButton{
		Text: "🖼️ View Photos",
		Data: fmt.Sprintf("view_photos:%d", tarantula.ID),
	}

	markup.InlineKeyboard = [][]tele.InlineButton{
		{photoBtn, viewPhotosBtn},
	}

	return c.Send(msg, markup, tele.ModeMarkdown)
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

	// Molt prediction settings
	moltPredictionToggleText := "🔕 Disable Molt Predictions"
	if !settings.MoltPredictionEnabled {
		moltPredictionToggleText = "🔔 Enable Molt Predictions"
	}

	moltPredictionToggleBtn := tele.InlineButton{
		Text: moltPredictionToggleText,
		Data: "toggle_molt_predictions",
	}

	moltPredictionDaysBtn := tele.InlineButton{
		Text: fmt.Sprintf("🦗 Molt Alert Days: %d days before", settings.MoltPredictionDays),
		Data: "set_molt_prediction_days",
	}

	postMoltMuteBtn := tele.InlineButton{
		Text: fmt.Sprintf("🤫 Post-Molt Mute: %d days", settings.PostMoltMuteDays),
		Data: "set_post_molt_mute_days",
	}

	markup.InlineKeyboard = [][]tele.InlineButton{
		{toggleBtn},
		{timeBtn},
		{reminderBtn},
		{moltPredictionToggleBtn},
		{moltPredictionDaysBtn},
		{postMoltMuteBtn},
	}

	return c.Send("🔔 Notification Settings:", markup)
}

func (t *TarantulaBot) handlePauseNotificationSettings(c tele.Context) error {
	settings, err := t.db.GetUserSettings(t.ctx, c.Sender().ID)
	if err != nil {
		return fmt.Errorf("failed to get user settings: %w", err)
	}

	markup := &tele.ReplyMarkup{}

	var pauseBtn, unpauseBtn, pause1Day, pause3Day, pause1Week tele.InlineButton

	if settings.NotificationsPaused {
		unpauseBtn = tele.InlineButton{
			Text: "▶️ Resume Notifications",
			Data: "unpause_notifications",
		}

		var statusText string
		if settings.PauseEndDate != nil {
			statusText = fmt.Sprintf("⏸️ Paused until %s", settings.PauseEndDate.Format("2006-01-02 15:04"))
		} else {
			statusText = "⏸️ Paused indefinitely"
		}

		markup.InlineKeyboard = [][]tele.InlineButton{
			{unpauseBtn},
		}

		return c.Send(fmt.Sprintf("🔕 Notifications Status: %s", statusText), markup)
	} else {
		pause1Day = tele.InlineButton{
			Text: "⏸️ Pause 1 Day",
			Data: "pause_1_day",
		}
		pause3Day = tele.InlineButton{
			Text: "⏸️ Pause 3 Days",
			Data: "pause_3_days",
		}
		pause1Week = tele.InlineButton{
			Text: "⏸️ Pause 1 Week",
			Data: "pause_1_week",
		}
		pauseBtn = tele.InlineButton{
			Text: "⏸️ Pause Indefinitely",
			Data: "pause_indefinitely",
		}

		markup.InlineKeyboard = [][]tele.InlineButton{
			{pause1Day, pause3Day},
			{pause1Week},
			{pauseBtn},
		}

		return c.Send("🔔 Notifications are currently active. Choose pause duration:", markup)
	}
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
	return c.Send(fmt.Sprintf("✅ Notifications %s!", status))
}

func (t *TarantulaBot) handleToggleMoltPredictions(c tele.Context) error {
	settings, err := t.db.GetUserSettings(t.ctx, c.Sender().ID)
	if err != nil {
		return fmt.Errorf("failed to get user settings: %w", err)
	}

	settings.MoltPredictionEnabled = !settings.MoltPredictionEnabled
	err = t.db.UpdateUserSettings(t.ctx, settings)
	if err != nil {
		return fmt.Errorf("failed to update user settings: %w", err)
	}

	status := "enabled"
	if !settings.MoltPredictionEnabled {
		status = "disabled"
	}
	return c.Send(fmt.Sprintf("✅ Molt prediction notifications %s!", status))
}

func (t *TarantulaBot) handleSetMoltPredictionDays(c tele.Context) error {
	session := t.sessions.GetSession(c.Sender().ID)
	session.CurrentField = "molt_prediction_days"
	t.sessions.UpdateSession(c.Sender().ID, session)

	return c.Send("How many days before a predicted molt should you be notified?")
}

func (t *TarantulaBot) handleSetPostMoltMuteDays(c tele.Context) error {
	session := t.sessions.GetSession(c.Sender().ID)
	session.CurrentField = "post_molt_mute_days"
	t.sessions.UpdateSession(c.Sender().ID, session)

	return c.Send("How many days after a molt should feeding notifications be muted?")
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

	case "molt_prediction_days":
		days, err := strconv.Atoi(c.Text())
		if err != nil || days <= 0 {
			return c.Send("Please enter a valid number of days (greater than 0)")
		}
		settings.MoltPredictionDays = days

	case "post_molt_mute_days":
		days, err := strconv.Atoi(c.Text())
		if err != nil || days < 0 {
			return c.Send("Please enter a valid number of days (0 or greater)")
		}
		settings.PostMoltMuteDays = days
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

func (t *TarantulaBot) setupColonyMaintenanceHandlers() {
	b := t.bot

	btnColonyMaintenance := menu.colony.Text("🧹 Colony Maintenance")
	menu.colony.Reply(
		menu.colony.Row(btnColonyStatus, btnUpdateCount),
		menu.colony.Row(btnFeedingHistory, btnColonyMaintenance),
		menu.colony.Row(menu.back),
	)

	b.Handle(&btnColonyMaintenance, t.handleColonyMaintenanceMenu)
}

func (t *TarantulaBot) handleColonyMaintenanceMenu(c tele.Context) error {
	colonies, err := t.db.GetColonyStatus(t.ctx, c.Sender().ID)
	if err != nil {
		return fmt.Errorf("failed to get colonies: %w", err)
	}

	if len(colonies) == 0 {
		return c.Send("No cricket colonies found. Add a colony first.")
	}

	var rows [][]tele.InlineButton
	for _, colony := range colonies {
		btn := tele.InlineButton{
			Text: fmt.Sprintf("🦗 %s (%d crickets)", colony.ColonyName, colony.CurrentCount),
			Data: fmt.Sprintf("colony_maintain_select_%d", colony.ID),
		}
		rows = append(rows, []tele.InlineButton{btn})
	}

	return c.Send("Select a colony to maintain:", &tele.ReplyMarkup{
		InlineKeyboard: rows,
	})
}

func (t *TarantulaBot) handleSelectColonyForMaintenance(c tele.Context, colonyID int) error {
	alerts, err := t.db.GetColonyMaintenanceAlerts(t.ctx, c.Sender().ID)
	if err != nil {
		return fmt.Errorf("failed to get maintenance alerts: %w", err)
	}

	colonies, err := t.db.GetColonyStatus(t.ctx, c.Sender().ID)
	if err != nil {
		return fmt.Errorf("failed to get colony: %w", err)
	}

	var colony models.ColonyStatus
	for _, col := range colonies {
		if int(col.ID) == colonyID {
			colony = col
			break
		}
	}

	if colony.ID == 0 {
		return c.Send("Colony not found.")
	}

	maintenanceTypes, err := t.db.GetMaintenanceTypes(t.ctx)
	if err != nil {
		return fmt.Errorf("failed to get maintenance types: %w", err)
	}

	msg := fmt.Sprintf("🦗 *%s* (%d crickets)\n\n", colony.ColonyName, colony.CurrentCount)

	colonyAlerts := make(map[string]models.ColonyMaintenanceAlert)
	for _, alert := range alerts {
		if int(alert.ID) == colonyID {
			colonyAlerts[alert.MaintenanceType] = alert
		}
	}

	if len(colonyAlerts) > 0 {
		msg += "⚠️ *Maintenance Due:*\n"
		for _, alert := range colonyAlerts {
			msg += fmt.Sprintf("• %s - %d days overdue\n", alert.MaintenanceType, alert.DaysOverdue)
		}
		msg += "\n"
	}

	msg += "Select a maintenance action to record:"

	var rows [][]tele.InlineButton
	for _, mType := range maintenanceTypes {
		var alertIndicator string
		if _, exists := colonyAlerts[mType.TypeName]; exists {
			alertIndicator = "⚠️ "
		}

		btn := tele.InlineButton{
			Text: fmt.Sprintf("%s%s", alertIndicator, mType.TypeName),
			Data: fmt.Sprintf("colony_maintain_record_%d_%d", colonyID, mType.ID),
		}
		rows = append(rows, []tele.InlineButton{btn})
	}

	historyBtn := tele.InlineButton{
		Text: "📜 View Maintenance History",
		Data: fmt.Sprintf("colony_maintain_history_%d", colonyID),
	}
	rows = append(rows, []tele.InlineButton{historyBtn})

	keyboard := &tele.ReplyMarkup{
		InlineKeyboard: rows,
	}

	if c.Callback() != nil {
		return c.Edit(msg, keyboard)
	}

	return c.Send(msg, keyboard)
}

func (t *TarantulaBot) handleRecordColonyMaintenance(c tele.Context, colonyID, typeID int) error {
	record := models.ColonyMaintenanceRecord{
		ColonyID:          colonyID,
		MaintenanceTypeID: typeID,
		MaintenanceDate:   time.Now(),
		UserID:            c.Sender().ID,
	}

	_, err := t.db.RecordColonyMaintenance(t.ctx, record)
	if err != nil {
		return fmt.Errorf("failed to record maintenance: %w", err)
	}

	maintenanceTypes, err := t.db.GetMaintenanceTypes(t.ctx)
	if err != nil {
		return fmt.Errorf("failed to get maintenance types: %w", err)
	}

	var typeName string
	for _, mt := range maintenanceTypes {
		if mt.ID == typeID {
			typeName = mt.TypeName
			break
		}
	}

	msg := fmt.Sprintf("✅ %s maintenance recorded!", typeName)

	return c.Respond(&tele.CallbackResponse{
		Text:      msg,
		ShowAlert: true,
	})
}

func (t *TarantulaBot) handleColonyMaintenanceHistory(c tele.Context, colonyID int) error {
	history, err := t.db.GetColonyMaintenanceHistory(t.ctx, int64(colonyID), c.Sender().ID, 10)
	if err != nil {
		return fmt.Errorf("failed to get maintenance history: %w", err)
	}

	var msg string
	if len(history) == 0 {
		msg = "No maintenance records found for this colony."
	} else {
		msg = "Recent maintenance records:\n\n"
		for _, record := range history {
			formattedDate := record.MaintenanceDate.Format("2006-01-02")
			msg += fmt.Sprintf("• %s on %s\n", record.MaintenanceType.TypeName, formattedDate)
		}
	}
	var rows [][]tele.InlineButton

	backBtn := tele.InlineButton{
		Text: "⬅️ Back to Colony List",
		Data: "colony_maintain_back",
	}
	rows = append(rows, []tele.InlineButton{backBtn})

	if c.Callback() != nil {
		return c.Edit(msg, &tele.ReplyMarkup{
			InlineKeyboard: rows,
		})
	}

	return c.Send(msg, &tele.ReplyMarkup{
		InlineKeyboard: rows,
	})
}

func (t *TarantulaBot) handlePhotoInput(c tele.Context, session *UserSession) error {
	if session.CurrentField != FieldPhoto {
		return nil
	}

	if c.Message().Photo == nil {
		return c.Send("Please send a photo")
	}

	photo := c.Message().Photo
	photoFile := photo.File

	// Keep FileID as reference
	photoURL := photoFile.FileID

	// Try downloading the binary photo data from Telegram
	buf, err := t.bot.File(&photoFile)
	if err != nil {
		return fmt.Errorf("failed to download photo: %w", err)
	}

	photoBytes, err := io.ReadAll(buf)
	if err != nil {
		return fmt.Errorf("failed to read photo data: %w", err)
	}

	photoRecord := models.TarantulaPhoto{
		TarantulaID: session.TarantulaData.ID,
		PhotoURL:    photoURL,
		PhotoData:   photoBytes,
		PhotoType:   "general",
		Caption:     c.Message().Caption,
		UserID:      c.Sender().ID,
	}

	_, err = t.db.AddPhoto(t.ctx, photoRecord)
	if err != nil {
		return fmt.Errorf("failed to save photo: %w", err)
	}

	// Update profile photo if it's the first one
	tarantula, err := t.db.GetTarantulaByID(t.ctx, c.Sender().ID, int32(session.TarantulaData.ID))
	if err == nil && tarantula.ProfilePhotoURL == "" {
		_ = t.db.UpdateTarantulaProfilePhoto(t.ctx, int32(session.TarantulaData.ID), photoURL, c.Sender().ID)
	}

	session.reset()
	t.sessions.UpdateSession(c.Sender().ID, session)

	return c.Send("📸 Photo added successfully!")
}

func (t *TarantulaBot) handleViewMolts(c tele.Context) error {
	moltRecords, err := t.db.GetRecentMoltRecords(t.ctx, c.Sender().ID, 20)
	if err != nil {
		return SendError(c, fmt.Sprintf("Failed to get molt records: %v", err))
	}

	if len(moltRecords) == 0 {
		return SendInfo(c, "🔄 No molt records found yet.\n\n💡 Record Molts when your tarantulas molt to track their growth and development!")
	}

	msg := "🔄 **Recent Molt History**\n\n"
	for i, record := range moltRecords {
		if i >= 15 { // Limit display
			break
		}

		daysAgo := int(time.Since(record.MoltDate).Hours() / 24)
		msg += fmt.Sprintf("🕷️ **%s**\n", record.Tarantula.Name)
		msg += fmt.Sprintf("📅 %s (%d days ago)\n", record.MoltDate.Format("2006-01-02"), daysAgo)

		if record.PreMoltLengthCM > 0 && record.PostMoltLengthCM > 0 {
			growth := record.PostMoltLengthCM - record.PreMoltLengthCM
			msg += fmt.Sprintf("📏 %.1fcm → %.1fcm (+%.1fcm)\n",
				record.PreMoltLengthCM, record.PostMoltLengthCM, growth)
		} else if record.PostMoltLengthCM > 0 {
			msg += fmt.Sprintf("📏 Size after molt: %.1fcm\n", record.PostMoltLengthCM)
		}

		if record.Notes != "" {
			msg += fmt.Sprintf("📝 %s\n", record.Notes)
		}

		msg += "\n"
	}

	return c.Send(msg, tele.ModeMarkdown)
}

func (t *TarantulaBot) handleQuickActions(c tele.Context) error {
	tarantulas, err := t.db.GetAllTarantulas(context.Background(), c.Sender().ID)
	if err != nil {
		return fmt.Errorf("failed to get tarantulas: %w", err)
	}

	if len(tarantulas) == 0 {
		return c.Send("No tarantulas found. Add one first!")
	}

	markup := &tele.ReplyMarkup{}
	var buttons [][]tele.InlineButton

	for _, spider := range tarantulas {
		daysSince := int(spider.DaysSinceFeeding)
		statusEmoji := "🟢"
		if daysSince > int(spider.MaxDays) {
			statusEmoji = "🔴"
		} else if daysSince >= int(spider.MinDays) {
			statusEmoji = "🟡"
		}

		button := tele.InlineButton{
			Text: fmt.Sprintf("%s %s (%dd)", statusEmoji, spider.Name, daysSince),
			Data: fmt.Sprintf("quick_feed:%d", spider.ID),
		}
		buttons = append(buttons, []tele.InlineButton{button})
	}

	markup.InlineKeyboard = buttons
	return c.Send("🚀 Quick Feed - Tap to feed with 1 cricket:", markup)
}

// Temporary debug function to troubleshoot feeding status
func (t *TarantulaBot) handleDebugStatus(c tele.Context) error {
	userID := GetUserID(c)

	tarantulas, err := t.db.GetAllTarantulas(t.ctx, userID)
	if err != nil {
		return SendError(c, "Failed to load tarantula data")
	}

	msg := "🐛 **Debug: Feeding Status Data**\n\n"
	for _, spider := range tarantulas {
		emoji, status := GetFeedingStatusWithMolt(int(spider.DaysSinceFeeding), int(spider.MinDays), int(spider.MaxDays), spider.CurrentStatus)

		msg += fmt.Sprintf("**%s %s**\n", emoji, spider.Name)
		msg += fmt.Sprintf("• Days since feeding: %.1f\n", spider.DaysSinceFeeding)
		msg += fmt.Sprintf("• Min days: %d\n", spider.MinDays)
		msg += fmt.Sprintf("• Max days: %d\n", spider.MaxDays)
		msg += fmt.Sprintf("• Current status: %s\n", spider.CurrentStatus)
		msg += fmt.Sprintf("• Feeding status: %s\n", status)
		msg += fmt.Sprintf("• Species ID: %d\n", spider.SpeciesID)
		msg += fmt.Sprintf("• Frequency ID: %d\n\n", spider.FrequencyID)
	}

	return SendInfo(c, msg)
}

// Temporary debug function to troubleshoot molt predictions
func (t *TarantulaBot) handleDebugMolts(c tele.Context) error {
	userID := GetUserID(c)

	// Get recent molt records
	molts, err := t.db.GetRecentMoltRecords(t.ctx, userID, 20)
	if err != nil {
		return SendError(c, "Failed to get molt records")
	}

	msg := "🐛 **Debug: Molt Records**\n\n"
	if len(molts) == 0 {
		msg += "No molt records found.\n"
	} else {
		for _, molt := range molts {
			msg += fmt.Sprintf("**%s**\n", molt.Tarantula.Name)
			msg += fmt.Sprintf("• Molt date: %s\n", molt.MoltDate.Format("2006-01-02"))
			msg += fmt.Sprintf("• Days ago: %d\n", int(time.Since(molt.MoltDate).Hours()/24))
			if molt.PostMoltLengthCM > 0 {
				msg += fmt.Sprintf("• Size: %.1f cm\n", molt.PostMoltLengthCM)
			}
			msg += "\n"
		}
	}

	return SendInfo(c, msg)
}

// Helper function to safely send or edit message based on context
func (t *TarantulaBot) sendOrEdit(c tele.Context, text string, options ...interface{}) error {
	// Try to edit first (if this is a callback from inline button)
	if c.Callback() != nil {
		return c.Edit(text, options...)
	}
	// Otherwise send new message (if this is from menu button)
	return c.Send(text, options...)
}

// ========== Tarantula Colony Management Handlers ==========

func (t *TarantulaBot) handleCreateColony(c tele.Context) error {
	session := t.sessions.GetSession(c.Sender().ID)
	session.CurrentState = StateCreatingColony
	session.CurrentField = FieldColonyName
	t.sessions.UpdateSession(c.Sender().ID, session)

	return c.Send("👥 Let's create a tarantula colony!\n\nWhat would you like to name this colony?\n(e.g., 'Balfouri Group', 'Main Colony')")
}

func (t *TarantulaBot) handleListColonies(c tele.Context) error {
	colonies, err := t.db.GetUserColonies(t.ctx, c.Sender().ID)
	if err != nil {
		return SendError(c, fmt.Sprintf("Failed to get colonies: %v", err))
	}

	if len(colonies) == 0 {
		return SendInfo(c, "👥 You don't have any colonies yet.\n\nUse 'Create Colony' to start a communal setup!")
	}

	msg := "👥 *Your Tarantula Colonies*\n\n"

	markup := &tele.ReplyMarkup{}
	var buttons [][]tele.InlineButton

	for _, colony := range colonies {
		activeMembers := 0
		for _, member := range colony.Members {
			if member.IsActive {
				activeMembers++
			}
		}

		msg += fmt.Sprintf("*%s*\n", colony.ColonyName)
		msg += fmt.Sprintf("Species: %s\n", colony.Species.CommonName)
		msg += fmt.Sprintf("Members: %d tarantulas\n", activeMembers)
		msg += fmt.Sprintf("Formed: %s\n", colony.FormationDate.Format("Jan 2, 2006"))
		if colony.Notes != "" {
			msg += fmt.Sprintf("Notes: %s\n", colony.Notes)
		}
		msg += "\n"

		btn := tele.InlineButton{
			Text: fmt.Sprintf("📋 %s (%d)", colony.ColonyName, activeMembers),
			Data: fmt.Sprintf("colony_details:%d", colony.ID),
		}
		buttons = append(buttons, []tele.InlineButton{btn})
	}

	markup.InlineKeyboard = buttons
	return c.Send(msg, markup, tele.ModeMarkdown)
}

func (t *TarantulaBot) handleAddToColony(c tele.Context) error {
	// First check if user has any colonies
	colonies, err := t.db.GetUserColonies(t.ctx, c.Sender().ID)
	if err != nil {
		return SendError(c, fmt.Sprintf("Failed to get colonies: %v", err))
	}

	if len(colonies) == 0 {
		return SendInfo(c, "❌ You need to create a colony first!\n\nUse 'Create Colony' to start.")
	}

	session := t.sessions.GetSession(c.Sender().ID)
	session.CurrentState = StateAddingToColony
	session.CurrentField = FieldColonySelection
	t.sessions.UpdateSession(c.Sender().ID, session)

	// Show colony selection
	msg := "👤 Add a tarantula to a colony\n\nSelect the colony:"

	markup := &tele.ReplyMarkup{}
	var buttons [][]tele.InlineButton

	for _, colony := range colonies {
		btn := tele.InlineButton{
			Text: fmt.Sprintf("%s (%s)", colony.ColonyName, colony.Species.CommonName),
			Data: fmt.Sprintf("select_colony_for_add:%d", colony.ID),
		}
		buttons = append(buttons, []tele.InlineButton{btn})
	}

	markup.InlineKeyboard = buttons
	return c.Send(msg, markup)
}

func (t *TarantulaBot) handleColonySpeciesSelected(c tele.Context, speciesID int) error {
	session := t.sessions.GetSession(c.Sender().ID)

	// Verify we're in colony creation mode
	if session.CurrentState != StateCreatingColony {
		return SendError(c, "Invalid session state. Please start over by selecting 'Create Colony'.")
	}

	session.TarantulaColony.SpeciesID = speciesID
	session.CurrentField = FieldFormationDate
	t.sessions.UpdateSession(c.Sender().ID, session)

	return c.Send("When was this colony formed? (YYYY-MM-DD)\n(Or enter today's date if forming now)")
}

func (t *TarantulaBot) handleColonyDetails(c tele.Context, colonyID int32) error {
	colony, err := t.db.GetColony(t.ctx, colonyID, c.Sender().ID)
	if err != nil {
		return SendError(c, fmt.Sprintf("Failed to get colony details: %v", err))
	}

	activeMembers := 0
	var membersList strings.Builder
	for _, member := range colony.Members {
		if member.IsActive {
			activeMembers++
			membersList.WriteString(fmt.Sprintf("  • %s (joined %s)\n",
				member.Tarantula.Name,
				member.JoinedDate.Format("Jan 2, 2006")))
		}
	}

	msg := fmt.Sprintf("👥 *Colony: %s*\n\n", colony.ColonyName)
	msg += fmt.Sprintf("Species: %s\n", colony.Species.CommonName)
	msg += fmt.Sprintf("Scientific: %s\n", colony.Species.ScientificName)
	msg += fmt.Sprintf("Formed: %s\n", colony.FormationDate.Format("Jan 2, 2006"))
	msg += fmt.Sprintf("\n*Members (%d):*\n", activeMembers)
	if activeMembers > 0 {
		msg += membersList.String()
	} else {
		msg += "  No members yet\n"
	}

	if colony.Notes != "" {
		msg += fmt.Sprintf("\nNotes: %s\n", colony.Notes)
	}

	markup := &tele.ReplyMarkup{}
	btnAddMember := tele.InlineButton{
		Text: "➕ Add Member",
		Data: fmt.Sprintf("select_colony_for_add:%d", colony.ID),
	}
	btnFeedColony := tele.InlineButton{
		Text: "🍽️ Feed Colony",
		Data: fmt.Sprintf("feed_colony:%d", colony.ID),
	}
	markup.InlineKeyboard = [][]tele.InlineButton{
		{btnFeedColony},
		{btnAddMember},
	}

	return c.Send(msg, markup, tele.ModeMarkdown)
}

func (t *TarantulaBot) handleColonySelectedForAdd(c tele.Context, colonyID int32) error {
	session := t.sessions.GetSession(c.Sender().ID)
	session.SelectedColonyID = int(colonyID)
	t.sessions.UpdateSession(c.Sender().ID, session)

	// Get colony to show species
	colony, err := t.db.GetColony(t.ctx, colonyID, c.Sender().ID)
	if err != nil {
		return SendError(c, fmt.Sprintf("Failed to get colony: %v", err))
	}

	// Get user's tarantulas of the same species
	allTarantulas, err := t.db.GetAllTarantulas(t.ctx, c.Sender().ID)
	if err != nil {
		return SendError(c, fmt.Sprintf("Failed to get tarantulas: %v", err))
	}

	var availableTarantulas []models.TarantulaListItem
	for _, t := range allTarantulas {
		if t.SpeciesID == int32(colony.SpeciesID) {
			availableTarantulas = append(availableTarantulas, t)
		}
	}

	if len(availableTarantulas) == 0 {
		return SendInfo(c, fmt.Sprintf("You don't have any %s to add to this colony.\n\nAdd some %s first!",
			colony.Species.CommonName, colony.Species.CommonName))
	}

	msg := fmt.Sprintf("Select a %s to add to the colony:", colony.Species.CommonName)

	markup := &tele.ReplyMarkup{}
	var buttons [][]tele.InlineButton

	for _, tarantula := range availableTarantulas {
		btn := tele.InlineButton{
			Text: tarantula.Name,
			Data: fmt.Sprintf("add_tarantula_to_colony:%d", tarantula.ID),
		}
		buttons = append(buttons, []tele.InlineButton{btn})
	}

	markup.InlineKeyboard = buttons
	return c.Send(msg, markup)
}

func (t *TarantulaBot) handleTarantulaSelectedForColony(c tele.Context, tarantulaID int32) error {
	session := t.sessions.GetSession(c.Sender().ID)
	colonyID := session.SelectedColonyID

	if colonyID == 0 {
		return SendError(c, "Session expired. Please try again.")
	}

	// Add the tarantula to the colony
	member := models.TarantulaColonyMember{
		ColonyID:    colonyID,
		TarantulaID: int(tarantulaID),
		JoinedDate:  time.Now(),
		IsActive:    true,
		UserID:      c.Sender().ID,
	}

	err := t.db.AddMemberToColony(t.ctx, member)
	if err != nil {
		return SendError(c, fmt.Sprintf("Failed to add member: %v", err))
	}

	session.reset()
	t.sessions.UpdateSession(c.Sender().ID, session)

	return sendSuccess(c, "Tarantula added to colony successfully!")
}

func (t *TarantulaBot) handleTarantulaSpeciesSelected(c tele.Context, speciesID int) error {
	session := t.sessions.GetSession(c.Sender().ID)

	if session.CurrentState != StateAddingTarantula || session.CurrentField != FieldSpecies {
		return SendError(c, "Invalid session state. Please start over.")
	}

	session.TarantulaData.SpeciesID = speciesID
	session.CurrentField = FieldAcquisitionDate
	t.sessions.UpdateSession(c.Sender().ID, session)

	return c.Send("When did you acquire this tarantula? (YYYY-MM-DD)")
}

func (t *TarantulaBot) handleFeedColony(c tele.Context, colonyID int32) error {
	// Get colony details
	colony, err := t.db.GetColony(t.ctx, colonyID, c.Sender().ID)
	if err != nil {
		return SendError(c, fmt.Sprintf("Failed to get colony: %v", err))
	}

	// Count active members
	activeMembers := 0
	for _, member := range colony.Members {
		if member.IsActive {
			activeMembers++
		}
	}

	if activeMembers == 0 {
		return SendInfo(c, "This colony has no members yet. Add some tarantulas first!")
	}

	// Prompt for number of crickets
	session := t.sessions.GetSession(c.Sender().ID)
	session.CurrentState = StateFeeding
	session.SelectedColonyID = int(colonyID)
	session.CurrentField = FieldFeedingCount
	t.sessions.UpdateSession(c.Sender().ID, session)

	return c.Send(fmt.Sprintf("🍽️ Feeding colony: %s (%d members)\n\nHow many crickets are you offering?",
		colony.ColonyName, activeMembers))
}
