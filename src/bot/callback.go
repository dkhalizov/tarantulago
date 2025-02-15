package bot

import (
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
		callback := parseCallback(c.Callback().Data)
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
		switch c.Callback().Data {
		case "toggle_notifications":
			return t.handleToggleNotifications(c)
		case "set_notification_time":
			return t.handleSetNotificationTime(c)
		case "set_feeding_reminder":
			return t.handleSetFeedingReminder(c)
		case "set_colony_threshold":
			return t.handleSetColonyThreshold(c)
		}

		return nil
	})
}
