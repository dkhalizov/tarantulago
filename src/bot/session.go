package bot

import (
	"context"
	"fmt"
	"strconv"
	"sync"
	"tarantulago/models"
	"time"

	tele "gopkg.in/telebot.v4"
)

type FormState string

const (
	StateIdle                 FormState = "idle"
	StateAddingTarantula      FormState = "adding_tarantula"
	StateAddingMolt           FormState = "adding_molt"
	StateAddingCrickets       FormState = "adding_crickets"
	StateAddingColony         FormState = "adding_colony"
	StateFeeding              FormState = "adding_feeding"
	StateNotificationSettings FormState = "notification_settings"
	StateRecordingMolt        FormState = "recording_molt"
	StateRecordingFeeding     FormState = "recording_feeding"

	StateAddingPhoto FormState = "adding_photo"
)

type TarantulaFormField string

const (
	FieldName            TarantulaFormField = "name"
	FieldSpecies         TarantulaFormField = "species"
	FieldAcquisitionDate TarantulaFormField = "acquisition_date"
	FieldAge             TarantulaFormField = "age"
	FieldCurrentSize     TarantulaFormField = "current_size"
	FieldHealthStatus    TarantulaFormField = "health_status"
	FieldNotes           TarantulaFormField = "notes"

	FieldPreMoltLengthCM  TarantulaFormField = "pre_molt_length_cm"
	FieldPostMoltLengthCM TarantulaFormField = "post_molt_length_cm"
	FieldMoltNotes        TarantulaFormField = "molt_notes"
	FieldSuccess          TarantulaFormField = "success"

	FieldColonyName  TarantulaFormField = "colony_name"
	FieldColonyCount TarantulaFormField = "colony_count"

	FieldColonyID     TarantulaFormField = "colony_id"
	FieldFeedingCount TarantulaFormField = "feeding_count"

	FieldPhoto TarantulaFormField = "photo"
)

type UserSession struct {
	CurrentState     FormState
	CurrentField     TarantulaFormField
	TarantulaData    models.Tarantula
	MoltData         models.MoltRecord
	Colony           models.CricketColony
	FeedEvent        models.FeedingEvent
	LastActivityTime time.Time
}

func (s *UserSession) reset() {
	s.CurrentState = StateIdle
	s.CurrentField = ""
	s.TarantulaData = models.Tarantula{}
	s.MoltData = models.MoltRecord{}
	s.Colony = models.CricketColony{}
	s.FeedEvent = models.FeedingEvent{}
}

type SessionManager struct {
	sessions map[int64]*UserSession
	mu       sync.RWMutex
}

func NewSessionManager() *SessionManager {
	return &SessionManager{
		sessions: make(map[int64]*UserSession),
	}
}

func (sm *SessionManager) GetSession(userID int64) *UserSession {
	sm.mu.RLock()
	defer sm.mu.RUnlock()
	if session, exists := sm.sessions[userID]; exists {
		return session
	}
	return &UserSession{
		CurrentState:     StateIdle,
		LastActivityTime: time.Now(),
	}
}

func (sm *SessionManager) UpdateSession(userID int64, session *UserSession) {
	sm.mu.Lock()
	defer sm.mu.Unlock()
	session.LastActivityTime = time.Now()
	sm.sessions[userID] = session
}

func (t *TarantulaBot) handleTarantulaFormInput(c tele.Context, session *UserSession) error {
	switch session.CurrentField {
	case FieldName:
		session.TarantulaData.Name = c.Text()
		session.CurrentField = FieldSpecies
		err := c.Send("Great name! Now, what species is your tarantula? (Please enter the species ID)")
		if err != nil {
			return err
		}

	case FieldSpecies:
		speciesID, err := strconv.Atoi(c.Text())
		if err != nil {
			return c.Send("Please enter a valid species ID number")
		}
		session.TarantulaData.SpeciesID = speciesID
		session.CurrentField = FieldAcquisitionDate
		err = c.Send("When did you acquire this tarantula? (YYYY-MM-DD)")
		if err != nil {
			return err
		}

	case FieldAcquisitionDate:
		date, ok := t.parseDate(c)
		if !ok {
			return nil
		}
		session.TarantulaData.AcquisitionDate = date
		session.CurrentField = FieldAge
		err := c.Send("What's the estimated age in months?")
		if err != nil {
			return err
		}

	case FieldAge:
		age, err := strconv.Atoi(c.Text())
		if err != nil {
			return c.Send("Please enter a valid number for age in months")
		}
		session.TarantulaData.EstimatedAgeMonths = age
		session.CurrentField = FieldCurrentSize
		err = c.Send("What's the current size in cm?")
		if err != nil {
			return err
		}

	case FieldCurrentSize:
		size, err := strconv.ParseFloat(c.Text(), 64)
		if err != nil {
			return c.Send("Please enter a valid size in centimeters")
		}
		session.TarantulaData.CurrentMoltStageID = int(models.MoltStageNormal)
		session.TarantulaData.CurrentSize = size
		session.CurrentField = FieldHealthStatus
		err = c.Send("What's the current health status ID?")
		if err != nil {
			return err
		}

	case FieldHealthStatus:
		healthStatus, err := strconv.Atoi(c.Text())
		if err != nil {
			return c.Send("Please enter a valid health status ID")
		}
		session.TarantulaData.CurrentHealthStatusID = healthStatus
		session.CurrentField = FieldNotes
		err = c.Send("Any additional notes? (or type 'skip' to leave empty)")
		if err != nil {
			return err
		}

	case FieldNotes:
		if c.Text() != "skip" {
			session.TarantulaData.Notes = c.Text()
		}
		session.TarantulaData.UserID = c.Sender().ID
		err := t.db.AddTarantula(context.Background(), session.TarantulaData)
		if err != nil {
			return fmt.Errorf("failed to save tarantula: %w", err)
		}

		session.reset()

		err = sendSuccess(c, "Tarantula added!")
		if err != nil {
			return err
		}
	}

	t.sessions.UpdateSession(c.Sender().ID, session)
	return nil
}

func (t *TarantulaBot) parseDate(c tele.Context) (time.Time, bool) {
	date, err := time.Parse("2006-01-02", c.Text())
	if err != nil {
		_ = c.Send("Please enter the date in YYYY-MM-DD format")
	}
	return date, err == nil
}

func (t *TarantulaBot) handleMoltFormInput(c tele.Context, session *UserSession) error {
	var err error

	switch session.CurrentField {
	case FieldPreMoltLengthCM:
		session.MoltData.MoltDate = time.Now()
		session.MoltData.PreMoltLengthCM, err = strconv.ParseFloat(c.Text(), 64)
		if err != nil {
			return c.Send("I'm sorry, I didn't understand that number. Please enter the length in centimeters.")
		}
		session.CurrentField = FieldPostMoltLengthCM
		err = c.Send("How long is your tarantula now (in cm)?")

	case FieldPostMoltLengthCM:
		session.MoltData.PostMoltLengthCM, err = strconv.ParseFloat(c.Text(), 64)
		if err != nil {
			return c.Send("I'm sorry, I didn't understand that number. Please enter the length in centimeters.")
		}
		session.CurrentField = FieldMoltNotes
		err = c.Send("Do you have any notes or observations you'd like to add?")

	case FieldMoltNotes:
		session.MoltData.Notes = c.Text()
		session.MoltData.UserID = c.Sender().ID
		session.CurrentField = FieldSuccess
		err = c.Send("Was the molt successful? (true/false)")

	case FieldSuccess:
		success, err := strconv.ParseBool(c.Text())
		if err != nil {
			return c.Send("Please enter a valid boolean value (true/false)")
		}
		if success {
			session.MoltData.MoltStageID = int(models.MoltStagePostMolt)
		} else {
			session.MoltData.MoltStageID = int(models.MoltStageFailed)
		}
		err = t.db.RecordMolt(context.Background(), session.MoltData)
		if err != nil {
			_ = sendError(c, err.Error())
			return nil
		}
		err = sendSuccess(c, "Molt recorded!")
		session.reset()
	}
	t.sessions.UpdateSession(c.Sender().ID, session)
	return err
}

func (t *TarantulaBot) handleColonyFormInput(c tele.Context, session *UserSession) error {
	var err error

	switch session.CurrentField {
	case FieldColonyName:
		session.Colony.ColonyName = c.Text()
		session.CurrentField = FieldColonyCount
		err = c.Send("How many crickets are in the colony?")

	case FieldColonyCount:
		count, err := strconv.Atoi(c.Text())
		if err != nil {
			return c.Send("Please enter a valid number for the colony count")
		}
		session.Colony.CurrentCount = count
		session.Colony.UserID = c.Sender().ID
		session.Colony.Notes = "Initial colony setup"
		session.Colony.LastCountDate = time.Now()
		err = t.db.AddColony(context.Background(), session.Colony)
		if err != nil {
			return fmt.Errorf("failed to save colony: %w", err)
		}

		session.reset()
		err = sendSuccess(c, "Colony added!")
	}

	t.sessions.UpdateSession(c.Sender().ID, session)
	return err
}

func (t *TarantulaBot) handleFeedingFormInput(c tele.Context, session *UserSession) error {
	var err error

	switch session.CurrentField {
	case FieldColonyID:
		colonyID, err := strconv.Atoi(c.Text())
		if err != nil {
			return c.Send("Please enter a valid colony ID")
		}
		session.FeedEvent.CricketColonyID = colonyID
		session.CurrentField = FieldFeedingCount
		err = c.Send("How many crickets did you feed?")
	case FieldFeedingCount:
		count, err := strconv.Atoi(c.Text())
		if err != nil {
			return c.Send("Please enter a valid number for the feeding count")
		}
		session.FeedEvent.NumberOfCrickets = count
		session.FeedEvent.FeedingDate = time.Now()
		session.FeedEvent.UserID = c.Sender().ID
		session.FeedEvent.FeedingStatusID = int(models.FeedingStatusAccepted)
		_, err = t.db.RecordFeeding(context.Background(), session.FeedEvent)
		if err != nil {
			return fmt.Errorf("failed to save feeding event: %w", err)
		}

		session.reset()
		err = sendSuccess(c, "Feeding event recorded!")
	}
	t.sessions.UpdateSession(c.Sender().ID, session)

	return err
}

func (t *TarantulaBot) handleCricketsFormInput(c tele.Context, session *UserSession) error {
	var err error

	switch session.CurrentField {
	case FieldColonyCount:
		count, err := strconv.Atoi(c.Text())
		if err != nil {
			return c.Send("Please enter a valid number for the cricket count")
		}

		colonies, err := t.db.GetColonyStatus(t.ctx, c.Sender().ID)
		if err != nil {
			return fmt.Errorf("failed to get colony: %w", err)
		}

		if len(colonies) == 0 {

			colony := models.CricketColony{
				ColonyName:    "Cricket Colony",
				CurrentCount:  count,
				LastCountDate: time.Now(),
				UserID:        c.Sender().ID,
				Notes:         "Initial setup",
			}
			err = t.db.AddColony(context.Background(), colony)
			if err != nil {
				return fmt.Errorf("failed to create colony: %w", err)
			}
		} else {

			colonyID := colonies[0].ID
			err = t.db.UpdateColonyCount(t.ctx, colonyID, int32(count), c.Sender().ID)
			if err != nil {
				return fmt.Errorf("failed to update colony count: %w", err)
			}
		}

		session.reset()
		err = sendSuccess(c, fmt.Sprintf("Cricket count updated to %d!", count))
	}

	return err
}
