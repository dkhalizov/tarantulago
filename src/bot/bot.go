package bot

import (
	"context"
	"fmt"
	"log/slog"
	"sync"
	"tarantulago/models"
	"time"

	tele "gopkg.in/telebot.v4"
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
				slog.Error("Failed to send error message", "error", err)
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

func (t *TarantulaBot) handleTarantulaDetailsEnhanced(c tele.Context) error {

	callback := parseCallback(c.Callback().Data)

	tarantula, err := t.db.GetTarantulaWithSpeciesData(t.ctx, int32(callback.TarantulaID), c.Sender().ID)
	if err != nil {
		return SendError(c, fmt.Sprintf("Failed to get tarantula details: %v", err))
	}

	photos, err := t.db.GetTarantulaPhotos(t.ctx, int32(callback.TarantulaID), c.Sender().ID, 3)
	if err != nil {
		photos = nil
	}

	msg := FormatTarantulaDetailsEnhanced(tarantula, photos, nil)

	markup := BuildTarantulaActionsMarkup(int32(callback.TarantulaID))

	return c.Send(msg, markup, tele.ModeMarkdown)
}

func (t *TarantulaBot) handleFeedingIntelligence(c tele.Context) error {

	callback := parseCallback(c.Callback().Data)

	tarantula, err := t.db.GetTarantulaWithSpeciesData(t.ctx, int32(callback.TarantulaID), c.Sender().ID)
	if err != nil {
		return SendError(c, fmt.Sprintf("Failed to get tarantula data: %v", err))
	}

	recentFeedings, err := t.db.GetRecentFeedingRecords(t.ctx, c.Sender().ID, 5)
	if err != nil {
		return SendError(c, fmt.Sprintf("Failed to get feeding history: %v", err))
	}

	daysSinceFeeding := 999.0
	for _, feeding := range recentFeedings {
		if feeding.TarantulaID != nil && *feeding.TarantulaID == callback.TarantulaID {
			daysSinceFeeding = time.Since(feeding.FeedingDate).Hours() / 24
			break
		}
	}

	intelligence := GetSpeciesFeedingIntelligence(
		tarantula.Species,
		tarantula.CurrentSize,
		tarantula.EstimatedAgeMonths,
		tarantula.CurrentMoltStage.StageName,
		daysSinceFeeding,
	)

	msg := FormatFeedingIntelligence(intelligence, tarantula.Species.ScientificName, tarantula.Name)

	return c.Send(msg, tele.ModeMarkdown)
}

func (t *TarantulaBot) handleIndividualMoltPrediction(c tele.Context) error {

	callback := parseCallback(c.Callback().Data)

	predictions, err := t.db.GetAllMoltPredictions(t.ctx, c.Sender().ID)
	if err != nil {
		return SendError(c, fmt.Sprintf("Failed to get molt predictions: %v", err))
	}

	var targetPrediction *models.MoltPrediction
	for _, pred := range predictions {
		if pred.TarantulaID == int32(callback.TarantulaID) {
			targetPrediction = &pred
			break
		}
	}

	if targetPrediction == nil {
		return SendInfo(c, "No molt prediction available for this tarantula. Need more historical molt data.")
	}

	msg := "🔮 **Individual Molt Prediction**\n\n"
	msg += FormatMoltPrediction(*targetPrediction)

	return c.Send(msg, tele.ModeMarkdown)
}

func (t *TarantulaBot) handleMoltPredictionsOverview(c tele.Context) error {

	predictions, err := t.db.GetAllMoltPredictions(t.ctx, c.Sender().ID)
	if err != nil {
		return SendError(c, fmt.Sprintf("Failed to get molt predictions: %v", err))
	}

	if len(predictions) == 0 {
		return SendInfo(c, "No molt predictions available yet. Record some molts to generate predictions!")
	}

	msg := GetMoltPredictionSummary(predictions)

	return c.Send(msg, tele.ModeMarkdown)
}
