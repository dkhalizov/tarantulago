package bot

import (
	"fmt"
	tele "gopkg.in/telebot.v4"
	"strconv"
	"strings"
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
		case "toggle_notifications":
			return t.handleToggleNotifications(c)
		case "set_notification_time":
			return t.handleSetNotificationTime(c)
		case "set_feeding_reminder":
			return t.handleSetFeedingReminder(c)
		case "set_colony_threshold":
			return t.handleSetColonyThreshold(c)
		}

		callback := parseCallback(callbackData)
		switch callback.Action {
		case selectCallback:
			return t.handleTarantulaSelect(c, callback.TarantulaID)
		case feedCallback:
			return t.handleTarantulaFeed(c, callback.TarantulaID)
		case feedSchedulerCallback:
			return t.handleFeedScheduler(c, callback.TarantulaID)
		case moltCallback:
			return t.handleTarantulaMolt(c, callback.TarantulaID)
		case backToListCallback:
			return t.showTarantulaList(c)
		}

		return nil
	})
}
