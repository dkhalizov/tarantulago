package bot

import (
	"context"
	"fmt"
	"log/slog"
	"tarantulago/models"
	"time"

	tele "gopkg.in/telebot.v4"
)

type NotificationSystem struct {
	bot    *tele.Bot
	db     NotificationOperations
	ctx    context.Context
	cancel context.CancelFunc
}

func NewNotificationSystem(bot *tele.Bot, db NotificationOperations) *NotificationSystem {
	ctx, cancel := context.WithCancel(context.Background())
	return &NotificationSystem{
		bot:    bot,
		db:     db,
		ctx:    ctx,
		cancel: cancel,
	}
}

func (n *NotificationSystem) Start() {
	go n.runNotificationScheduler()
}

func (n *NotificationSystem) Stop() {
	n.cancel()
}

func (n *NotificationSystem) shouldSendNotification(settings *models.UserSettings) bool {
	if !settings.NotificationEnabled {
		return false
	}

	if settings.NotificationsPaused {
		now := time.Now().UTC()

		if settings.PauseEndDate != nil && now.After(*settings.PauseEndDate) {

		} else {
			return false
		}
	}

	notificationTime, err := time.Parse("15:04", settings.NotificationTimeUTC)
	if err != nil {
		slog.Error("Invalid notification time format", "time", settings.NotificationTimeUTC)
		return false
	}

	now := time.Now().UTC()
	currentTime := time.Date(0, 1, 1, now.Hour(), now.Minute(), 0, 0, time.UTC)

	diff := currentTime.Sub(notificationTime)
	return diff >= 0 && diff < time.Minute
}

func (n *NotificationSystem) runNotificationScheduler() {
	ticker := time.NewTicker(1 * time.Minute)
	defer ticker.Stop()

	for {
		select {
		case <-n.ctx.Done():
			return
		case <-ticker.C:
			n.processScheduledNotifications()
		}
	}
}

func (n *NotificationSystem) processScheduledNotifications() {
	users, err := n.db.GetActiveUsers(n.ctx)
	if err != nil {
		slog.Error("Failed to get active users", "error", err)
		return
	}

	for _, user := range users {
		settings, err := n.db.GetUserSettings(n.ctx, user.TelegramID)
		if err != nil {
			slog.Error("Failed to get user settings", "user_id", user.TelegramID, "error", err)
			continue
		}

		if n.shouldSendNotification(settings) {
			n.triggerChecks(user, settings)
		}
	}
}

func (n *NotificationSystem) triggerChecks(user models.TelegramUser, settings *models.UserSettings) {
	n.checkFeedings(user.TelegramID, user.ChatID, settings)

}

func (n *NotificationSystem) checkFeedings(userID int64, chatID int64, settings *models.UserSettings) {
	feedings, err := n.db.GetTarantulasDueFeeding(n.ctx, userID)
	if err != nil {
		slog.Error("Error checking feedings", "user_id", userID, "error", err)
		return
	}

	overdueFeedings := make([]models.TarantulaListItem, 0)
	dueFeedings := make([]models.TarantulaListItem, 0)

	for _, t := range feedings {
		if t.CurrentStatus == "Pre-molt" || t.CurrentStatus == "Molting" || t.CurrentStatus == "Post-molt" {
			continue
		}

		if t.DaysSinceFeeding > float64(t.MaxDays) {
			overdueFeedings = append(overdueFeedings, t)
		} else if t.DaysSinceFeeding >= float64(t.MinDays) {
			dueFeedings = append(dueFeedings, t)
		}
	}

	if len(overdueFeedings) > 0 || len(dueFeedings) > 0 {
		message := "🕷 *Feeding Schedule Update*\n\n"

		if len(overdueFeedings) > 0 {
			message += "⚠️ *Overdue Feedings:*\n"
			for _, t := range overdueFeedings {
				message += fmt.Sprintf("• %s (%s) - %.0f days since last feeding (recommended: %d-%d days)\n",
					t.Name, t.SpeciesName, t.DaysSinceFeeding, t.MinDays, t.MaxDays)
			}
			message += "\n"
		}

		if len(dueFeedings) > 0 {
			message += "📅 *Due for Feeding:*\n"
			for _, t := range dueFeedings {
				message += fmt.Sprintf("• %s (%s) - %.0f days since last feeding (recommended: %d-%d days)\n",
					t.Name, t.SpeciesName, t.DaysSinceFeeding, t.MinDays, t.MaxDays)
			}
		}

		if _, err = n.bot.Send(&tele.Chat{ID: chatID}, message, tele.ModeMarkdown); err != nil {
			slog.Error("Error sending feeding notification", "user_id", userID, "error", err)
		}
	}
}

func (n *NotificationSystem) checkColonies(userID int64, chatID int64, settings *models.UserSettings) {
	colonies, err := n.db.GetColonyStatus(n.ctx, userID)
	if err != nil {
		slog.Error("Error checking colonies", "user_id", userID, "error", err)
		return
	}

	if len(colonies) > 0 {
		colony := colonies[0]
		if colony.CurrentCount <= int32(settings.LowColonyThreshold) {
			message := fmt.Sprintf("🦗 *Low Cricket Alert*\n\nYour cricket colony has %d crickets remaining\n\n💡 Consider breeding more crickets soon!", colony.CurrentCount)

			if _, err = n.bot.Send(&tele.Chat{ID: chatID}, message, tele.ModeMarkdown); err != nil {
				slog.Error("Error sending colony notification", "user_id", userID, "error", err)
			}
		}
	}
}

func (n *NotificationSystem) checkColonyMaintenance(userID int64, chatID int64, settings *models.UserSettings) {
	if !settings.MaintenanceReminderEnabled {
		return
	}

	alerts, err := n.db.GetColonyMaintenanceAlerts(n.ctx, userID)
	if err != nil {
		slog.Error("Error checking colony maintenance alerts", "user_id", userID, "error", err)
		return
	}

	if len(alerts) == 0 {
		return
	}

	coloniesWithAlerts := make(map[string][]models.ColonyMaintenanceAlert)
	for _, alert := range alerts {
		coloniesWithAlerts[alert.ColonyName] = append(coloniesWithAlerts[alert.ColonyName], alert)
	}

	message := "🧹 *Cricket Colony Maintenance Reminder*\n\n"

	for colonyName, alerts := range coloniesWithAlerts {
		message += fmt.Sprintf("*%s*:\n", colonyName)
		for _, alert := range alerts {
			message += fmt.Sprintf("• %s - %d days overdue\n", alert.MaintenanceType, alert.DaysOverdue)
		}
		message += "\n"
	}

	if _, err = n.bot.Send(&tele.Chat{ID: chatID}, message, tele.ModeMarkdown); err != nil {
		slog.Error("Error sending maintenance notification", "user_id", userID, "error", err)
	}
}
