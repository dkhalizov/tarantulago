package bot

import (
	"context"
	"fmt"
	tele "gopkg.in/telebot.v4"
	"log/slog"
	"sync"
	"time"
)

type TarantulaBot struct {
	bot           *tele.Bot
	db            TarantulaOperations
	userChats     sync.Map
	notifications *NotificationSystem
	ctx           context.Context
	cancelFunc    context.CancelFunc
	sessions      *SessionManager
}

func NewTarantulaBot(token string, database TarantulaOperations) (*TarantulaBot, error) {
	bot, err := tele.NewBot(tele.Settings{
		Token:  token,
		Poller: &tele.LongPoller{Timeout: 10 * time.Second},
		OnError: func(e error, ctx tele.Context) {
			if err := sendError(ctx, e.Error()); err != nil {
				slog.Error("Failed to send error message:", err)
			}
		},
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create bot: %w", err)
	}

	ctx, cancel := context.WithCancel(context.Background())
	tarantulaBot := TarantulaBot{
		bot:           bot,
		notifications: NewNotificationSystem(bot, database),
		db:            database,
		ctx:           ctx,
		cancelFunc:    cancel,
		sessions:      NewSessionManager(),
	}

	tarantulaBot.setupHandlers()
	return &tarantulaBot, nil
}

func (t *TarantulaBot) Start() {
	t.notifications.Start()
	t.bot.Start()
}

func (t *TarantulaBot) Stop() {
	t.cancelFunc()
}
