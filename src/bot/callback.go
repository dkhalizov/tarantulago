package bot

import (
	"bytes"
	"fmt"
	"strconv"
	"strings"
	"time"

	tele "gopkg.in/telebot.v4"
)

type CallbackData struct {
	Action      string
	TarantulaID int
}

func parseCallback(data string) CallbackData {
	parts := strings.Split(data, ":")
	if len(parts) != 2 {
		return CallbackData{}
	}
	id, _ := strconv.Atoi(parts[1])
	return CallbackData{
		Action:      parts[0],
		TarantulaID: id,
	}
}

const selectCallback = "select"
const feedCallback = "feed"
const moltCallback = "molt"
const backToListCallback = "back_to_list"
const feedSchedulerCallback = "feed_scheduler"

func (t *TarantulaBot) setupInlineKeyboards() {
	t.bot.Handle(tele.OnCallback, func(c tele.Context) error {
		callbackData := c.Callback().Data
		callbackData = strings.TrimLeft(callbackData, "\f\n\r\t ")

		if len(callbackData) >= 15 && callbackData[:16] == "colony_maintain_" {
			parts := strings.Split(callbackData, "_")
			if len(parts) < 4 {
				return c.Send("Invalid callback data")
			}

			action := parts[2]
			idStr := parts[3]
			colonyID, err := strconv.Atoi(idStr)
			if err != nil {
				return fmt.Errorf("invalid colony ID: %w", err)
			}

			switch action {
			case "select":
				return t.handleSelectColonyForMaintenance(c, colonyID)
			case "record":
				if len(parts) < 5 {
					return c.Send("Invalid maintenance type")
				}
				typeIDStr := parts[4]
				typeID, err := strconv.Atoi(typeIDStr)
				if err != nil {
					return fmt.Errorf("invalid maintenance type ID: %w", err)
				}
				return t.handleRecordColonyMaintenance(c, colonyID, typeID)
			case "history":
				return t.handleColonyMaintenanceHistory(c, colonyID)
			case "back":
				return t.handleColonyMaintenanceMenu(c)
			}
			return nil
		}

		switch callbackData {
		case "set_notification_time":
			return t.handleSetNotificationTime(c)
		case "set_feeding_reminder":
			return t.handleSetFeedingReminder(c)
		case "toggle_molt_predictions":
			return t.handleToggleMoltPredictions(c)
		case "set_molt_prediction_days":
			return t.handleSetMoltPredictionDays(c)
		case "set_post_molt_mute_days":
			return t.handleSetPostMoltMuteDays(c)
		case "toggle_notifications":
			return t.handleToggleNotifications(c)
		case "pause_1_day":
			return t.handlePauseNotifications(c, 24*time.Hour)
		case "pause_3_days":
			return t.handlePauseNotifications(c, 72*time.Hour)
		case "pause_1_week":
			return t.handlePauseNotifications(c, 168*time.Hour)
		case "pause_indefinitely":
			return t.handlePauseNotifications(c, 0)
		case "unpause_notifications":
			return t.handleUnpauseNotifications(c)
		}

		if strings.HasPrefix(callbackData, "quick_feed:") {
			tarantulaIDStr := strings.TrimPrefix(callbackData, "quick_feed:")
			tarantulaID, err := strconv.Atoi(tarantulaIDStr)
			if err != nil {
				return c.Send("Invalid tarantula ID")
			}
			return t.handleQuickFeed(c, int32(tarantulaID))
		}

		if strings.HasPrefix(callbackData, "add_photo:") || strings.HasPrefix(callbackData, "photo:") {
			var tarantulaIDStr string
			if strings.HasPrefix(callbackData, "add_photo:") {
				tarantulaIDStr = strings.TrimPrefix(callbackData, "add_photo:")
			} else {
				tarantulaIDStr = strings.TrimPrefix(callbackData, "photo:")
			}
			tarantulaID, err := strconv.Atoi(tarantulaIDStr)
			if err != nil {
				return c.Send("Invalid tarantula ID")
			}
			return t.handleAddPhoto(c, int32(tarantulaID))
		}

		if strings.HasPrefix(callbackData, "view_photos:") {
			tarantulaIDStr := strings.TrimPrefix(callbackData, "view_photos:")
			tarantulaID, err := strconv.Atoi(tarantulaIDStr)
			if err != nil {
				return c.Send("Invalid tarantula ID")
			}
			return t.handleViewPhotos(c, int32(tarantulaID))
		}

		if strings.HasPrefix(callbackData, "intel:") {
			return t.handleFeedingIntelligence(c)
		}

		if strings.HasPrefix(callbackData, "molt_pred:") {
			return t.handleIndividualMoltPrediction(c)
		}

		if callbackData == "molt_predictions" {
			return t.handleMoltPredictionsOverview(c)
		}

		if callbackData == "back_to_list" || strings.HasPrefix(callbackData, "back_to_list:") {
			return t.showTarantulaList(c)
		}

		// Colony management callbacks
		if strings.HasPrefix(callbackData, "colony_species:") {
			speciesIDStr := strings.TrimPrefix(callbackData, "colony_species:")
			speciesID, err := strconv.Atoi(speciesIDStr)
			if err != nil {
				return c.Send("Invalid species ID")
			}
			return t.handleColonySpeciesSelected(c, speciesID)
		}

		if strings.HasPrefix(callbackData, "colony_details:") {
			colonyIDStr := strings.TrimPrefix(callbackData, "colony_details:")
			colonyID, err := strconv.Atoi(colonyIDStr)
			if err != nil {
				return c.Send("Invalid colony ID")
			}
			return t.handleColonyDetails(c, int32(colonyID))
		}

		if strings.HasPrefix(callbackData, "select_colony_for_add:") {
			colonyIDStr := strings.TrimPrefix(callbackData, "select_colony_for_add:")
			colonyID, err := strconv.Atoi(colonyIDStr)
			if err != nil {
				return c.Send("Invalid colony ID")
			}
			return t.handleColonySelectedForAdd(c, int32(colonyID))
		}

		if strings.HasPrefix(callbackData, "add_tarantula_to_colony:") {
			tarantulaIDStr := strings.TrimPrefix(callbackData, "add_tarantula_to_colony:")
			tarantulaID, err := strconv.Atoi(tarantulaIDStr)
			if err != nil {
				return c.Send("Invalid tarantula ID")
			}
			return t.handleTarantulaSelectedForColony(c, int32(tarantulaID))
		}

		if strings.HasPrefix(callbackData, "feed_colony:") {
			colonyIDStr := strings.TrimPrefix(callbackData, "feed_colony:")
			colonyID, err := strconv.Atoi(colonyIDStr)
			if err != nil {
				return c.Send("Invalid colony ID")
			}
			return t.handleFeedColony(c, int32(colonyID))
		}

		// Handle species selection during tarantula creation
		if strings.HasPrefix(callbackData, "add_tarantula_species:") {
			speciesIDStr := strings.TrimPrefix(callbackData, "add_tarantula_species:")
			speciesID, err := strconv.Atoi(speciesIDStr)
			if err != nil {
				return c.Send("Invalid species ID")
			}
			return t.handleTarantulaSpeciesSelected(c, speciesID)
		}

		callback := parseCallback(callbackData)
		switch callback.Action {
		case selectCallback:
			return t.handleTarantulaDetailsEnhanced(c)
		case feedCallback:
			return t.handleTarantulaFeed(c, callback.TarantulaID)
		case feedSchedulerCallback:
			return t.handleFeedScheduler(c, callback.TarantulaID)
		case moltCallback:
			return t.handleTarantulaMolt(c, callback.TarantulaID)
		}

		return nil
	})
}

func (t *TarantulaBot) handlePauseNotifications(c tele.Context, duration time.Duration) error {
	settings, err := t.db.GetUserSettings(t.ctx, c.Sender().ID)
	if err != nil {
		return fmt.Errorf("failed to get user settings: %w", err)
	}

	now := time.Now().UTC()
	settings.NotificationsPaused = true
	settings.PauseStartDate = &now

	if duration > 0 {
		endTime := now.Add(duration)
		settings.PauseEndDate = &endTime
		settings.PauseReason = fmt.Sprintf("Paused for %s", duration.String())
	} else {
		settings.PauseEndDate = nil
		settings.PauseReason = "Paused indefinitely"
	}

	err = t.db.UpdateUserSettings(t.ctx, settings)
	if err != nil {
		return fmt.Errorf("failed to update user settings: %w", err)
	}

	var message string
	if duration > 0 {
		message = fmt.Sprintf("⏸️ Notifications paused until %s", settings.PauseEndDate.Format("2006-01-02 15:04"))
	} else {
		message = "⏸️ Notifications paused indefinitely"
	}

	return c.Send(message)
}

func (t *TarantulaBot) handleUnpauseNotifications(c tele.Context) error {
	settings, err := t.db.GetUserSettings(t.ctx, c.Sender().ID)
	if err != nil {
		return fmt.Errorf("failed to get user settings: %w", err)
	}

	settings.NotificationsPaused = false
	settings.PauseStartDate = nil
	settings.PauseEndDate = nil
	settings.PauseReason = ""

	err = t.db.UpdateUserSettings(t.ctx, settings)
	if err != nil {
		return fmt.Errorf("failed to update user settings: %w", err)
	}

	return c.Send("▶️ Notifications resumed! You'll receive feeding and colony alerts as scheduled.")
}

func (t *TarantulaBot) handleQuickFeed(c tele.Context, tarantulaID int32) error {
	err := t.db.QuickFeed(t.ctx, tarantulaID, c.Sender().ID)
	if err != nil {
		return c.Send(fmt.Sprintf("❌ Failed to record feeding: %s", err.Error()))
	}

	tarantula, err := t.db.GetTarantulaByID(t.ctx, c.Sender().ID, tarantulaID)
	if err != nil {
		return c.Send("✅ Fed successfully! (Could not retrieve details)")
	}

	return c.Send(fmt.Sprintf("✅ %s fed with 1 cricket!", tarantula.Name))
}

func (t *TarantulaBot) handleAddPhoto(c tele.Context, tarantulaID int32) error {
	session := t.sessions.GetSession(c.Sender().ID)
	session.CurrentState = StateAddingPhoto
	session.CurrentField = FieldPhoto
	session.TarantulaData.ID = int(tarantulaID)
	t.sessions.UpdateSession(c.Sender().ID, session)

	return c.Send("📸 Send a photo of your tarantula:")
}

func (t *TarantulaBot) handleWeightHistory(c tele.Context, tarantulaID int32) error {
	weights, err := t.db.GetWeightHistory(t.ctx, tarantulaID, c.Sender().ID, 10)
	if err != nil {
		return SendError(c, fmt.Sprintf("Failed to get weight history: %v", err))
	}

	tarantula, err := t.db.GetTarantulaByID(t.ctx, c.Sender().ID, tarantulaID)
	if err != nil {
		return SendError(c, fmt.Sprintf("Failed to get tarantula: %v", err))
	}

	if len(weights) == 0 {
		return SendInfo(c, fmt.Sprintf("⚖️ No weight records found for %s.\n\n💡 Use the ⚖️ button to start tracking weight!", tarantula.Name))
	}

	msg := fmt.Sprintf("⚖️ **Weight History for %s**\n\n", tarantula.Name)

	for i, weight := range weights {
		if i >= 8 {
			break
		}

		daysAgo := int(time.Since(weight.WeighDate).Hours() / 24)
		msg += fmt.Sprintf("📅 %s (%d days ago)\n", weight.WeighDate.Format("2006-01-02"), daysAgo)
		msg += fmt.Sprintf("⚖️ %.2fg\n", weight.WeightGrams)

		if weight.Notes != "" {
			msg += fmt.Sprintf("📝 %s\n", weight.Notes)
		}
		msg += "\n"
	}

	if len(weights) > 1 {
		firstWeight := weights[len(weights)-1].WeightGrams
		lastWeight := weights[0].WeightGrams
		change := lastWeight - firstWeight

		trendEmoji := "➡️"
		trendText := "stable"
		if change > 0.5 {
			trendEmoji = "📈"
			trendText = "gaining"
		} else if change < -0.5 {
			trendEmoji = "📉"
			trendText = "losing"
		}

		msg += fmt.Sprintf("%s **Trend:** %s weight (%+.1fg)\n", trendEmoji, trendText, change)
	}

	return c.Send(msg, tele.ModeMarkdown)
}

func (t *TarantulaBot) handleViewPhotos(c tele.Context, tarantulaID int32) error {
	photos, err := t.db.GetTarantulaPhotos(t.ctx, tarantulaID, c.Sender().ID, 5)
	if err != nil {
		return fmt.Errorf("failed to get photos: %w", err)
	}

	tarantula, err := t.db.GetTarantulaByID(t.ctx, c.Sender().ID, tarantulaID)
	if err != nil {
		return fmt.Errorf("failed to get tarantula: %w", err)
	}

	if len(photos) == 0 {
		return c.Send(fmt.Sprintf("🖼️ No photos found for %s. Add some photos to track their growth!", tarantula.Name))
	}

	var media tele.Album
	for _, p := range photos {
		if len(p.PhotoData) > 0 {
			photo := tele.Photo{File: tele.FromReader(bytes.NewReader(p.PhotoData)), Caption: p.Caption}
			media = append(media, &photo)
		} else if p.PhotoURL != "" {
			photo := tele.Photo{File: tele.File{FileID: p.PhotoURL}, Caption: p.Caption}
			media = append(media, &photo)
		}
	}

	if len(media) > 0 {
		return c.SendAlbum(media)
	}

	msg := fmt.Sprintf("🖼️ *Recent Photos of %s*\n\n", tarantula.Name)
	for _, photo := range photos {
		msg += fmt.Sprintf("📅 %s", photo.TakenDate.Format("2006-01-02"))
		if photo.Caption != "" {
			msg += fmt.Sprintf(" - %s", photo.Caption)
		}
		msg += "\n"
	}

	return c.Send(msg, tele.ModeMarkdown)
}
