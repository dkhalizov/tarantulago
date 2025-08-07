package bot

import (
	"context"
	"fmt"
	"tarantulago/models"
	"time"
)

type TarantulaServiceImpl struct {
	db TarantulaOperations
}

type FeedingServiceImpl struct {
	db TarantulaOperations
}

type ColonyServiceImpl struct {
	db TarantulaOperations
}

type AnalyticsServiceImpl struct {
	db TarantulaOperations
}

func NewTarantulaService(db TarantulaOperations) *TarantulaServiceImpl {
	return &TarantulaServiceImpl{db: db}
}

func (s *TarantulaServiceImpl) QuickAdd(ctx context.Context, userID int64, name string, speciesID int32) error {
	tarantula := models.Tarantula{
		Name:                  name,
		SpeciesID:             int(speciesID),
		AcquisitionDate:       time.Now(),
		EstimatedAgeMonths:    6,
		CurrentMoltStageID:    1,
		CurrentHealthStatusID: 1,
		LastHealthCheckDate:   time.Now(),
		UserID:                userID,
	}
	return s.db.AddTarantula(ctx, tarantula)
}

func (s *TarantulaServiceImpl) GetWithDetails(ctx context.Context, userID int64, id int32) (*TarantulaDetails, error) {
	tarantula, err := s.db.GetTarantulaByID(ctx, userID, id)
	if err != nil {
		return nil, err
	}

	photos, _ := s.db.GetTarantulaPhotos(ctx, id, userID, 3)

	weights, _ := s.db.GetWeightHistory(ctx, id, userID, 1)
	var latestWeight *models.WeightRecord
	if len(weights) > 0 {
		latestWeight = &weights[0]
	}

	return &TarantulaDetails{
		Tarantula:    tarantula,
		RecentPhotos: photos,
		LatestWeight: latestWeight,
	}, nil
}

type TarantulaDetails struct {
	Tarantula    *models.Tarantula
	RecentPhotos []models.TarantulaPhoto
	LatestWeight *models.WeightRecord
}

func NewFeedingService(db TarantulaOperations) *FeedingServiceImpl {
	return &FeedingServiceImpl{db: db}
}

func (s *FeedingServiceImpl) QuickFeedWithCrickets(ctx context.Context, tarantulaID int32, userID int64, cricketCount int) error {
	event := models.FeedingEvent{
		TarantulaID:      int(tarantulaID),
		FeedingDate:      time.Now(),
		NumberOfCrickets: cricketCount,
		Notes:            fmt.Sprintf("Quick feed - %d crickets", cricketCount),
		UserID:           userID,
	}

	_, err := s.db.RecordFeeding(ctx, event)
	return err
}

func (s *FeedingServiceImpl) GetRecentFeedingsForTarantula(ctx context.Context, tarantulaID int32, userID int64) ([]models.FeedingEvent, error) {

	return s.db.GetRecentFeedingRecords(ctx, userID, 10)
}

func NewColonyService(db TarantulaOperations) *ColonyServiceImpl {
	return &ColonyServiceImpl{db: db}
}

func (s *ColonyServiceImpl) GetStatusSummary(ctx context.Context, userID int64) (*ColonySummary, error) {
	statuses, err := s.db.GetColonyStatus(ctx, userID)
	if err != nil {
		return nil, err
	}

	summary := &ColonySummary{
		Colonies:       statuses,
		TotalCrickets:  0,
		NeedsAttention: false,
	}

	for _, status := range statuses {
		summary.TotalCrickets += int(status.CurrentCount)
		if status.CurrentCount < 20 {
			summary.NeedsAttention = true
		}
	}

	return summary, nil
}

type ColonySummary struct {
	Colonies       []models.ColonyStatus
	TotalCrickets  int
	NeedsAttention bool
}

func NewAnalyticsService(db TarantulaOperations) *AnalyticsServiceImpl {
	return &AnalyticsServiceImpl{db: db}
}

func (s *AnalyticsServiceImpl) GetDashboard(ctx context.Context, userID int64) (*AnalyticsDashboard, error) {
	patterns, err := s.db.GetAllFeedingPatterns(ctx, userID)
	if err != nil {
		return nil, err
	}

	growthData, err := s.db.GetAllGrowthData(ctx, userID)
	if err != nil {
		return nil, err
	}

	predictions, err := s.db.GetAllMoltPredictions(ctx, userID)
	if err != nil {
		return nil, err
	}

	return &AnalyticsDashboard{
		FeedingPatterns: patterns,
		GrowthData:      growthData,
		MoltPredictions: predictions,
		GeneratedAt:     time.Now(),
	}, nil
}

type AnalyticsDashboard struct {
	FeedingPatterns []models.FeedingPattern
	GrowthData      []models.GrowthData
	MoltPredictions []models.MoltPrediction
	GeneratedAt     time.Time
}

type TarantulaBuilder struct {
	tarantula models.Tarantula
}

func NewTarantulaBuilder() *TarantulaBuilder {
	return &TarantulaBuilder{
		tarantula: models.Tarantula{
			AcquisitionDate:       time.Now(),
			CurrentMoltStageID:    1,
			CurrentHealthStatusID: 1,
			LastHealthCheckDate:   time.Now(),
		},
	}
}

func (b *TarantulaBuilder) Name(name string) *TarantulaBuilder {
	b.tarantula.Name = name
	return b
}

func (b *TarantulaBuilder) Species(speciesID int) *TarantulaBuilder {
	b.tarantula.SpeciesID = speciesID
	return b
}

func (b *TarantulaBuilder) Age(months int) *TarantulaBuilder {
	b.tarantula.EstimatedAgeMonths = months
	return b
}

func (b *TarantulaBuilder) User(userID int64) *TarantulaBuilder {
	b.tarantula.UserID = userID
	return b
}

func (b *TarantulaBuilder) Build() models.Tarantula {
	return b.tarantula
}

type QueryOptions struct {
	Limit     int32
	Offset    int32
	SortBy    string
	SortOrder string
	DateFrom  *time.Time
	DateTo    *time.Time
}

func DefaultQueryOptions() QueryOptions {
	return QueryOptions{
		Limit:     10,
		Offset:    0,
		SortBy:    "created_at",
		SortOrder: "desc",
	}
}

type ServiceError struct {
	Operation string
	Err       error
	UserID    int64
}

func (e ServiceError) Error() string {
	return fmt.Sprintf("service error in %s for user %d: %v", e.Operation, e.UserID, e.Err)
}

func NewServiceError(operation string, userID int64, err error) ServiceError {
	return ServiceError{
		Operation: operation,
		UserID:    userID,
		Err:       err,
	}
}
