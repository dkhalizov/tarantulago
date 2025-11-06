package bot

import (
	"context"
	"tarantulago/models"
)

type TarantulaOperations interface {
	EnsureUserExists(ctx context.Context, user *models.TelegramUser) error
	GetUserSettings(ctx context.Context, userID int64) (*models.UserSettings, error)
	UpdateUserSettings(ctx context.Context, settings *models.UserSettings) error

	TarantulaService

	FeedingService

	ColonyService

	AnalyticsService

	NotificationOperations
}

type TarantulaService interface {
	AddTarantula(ctx context.Context, t models.Tarantula) error
	GetTarantulaByID(ctx context.Context, userID int64, id int32) (*models.Tarantula, error)
	GetTarantulaWithSpeciesData(ctx context.Context, tarantulaID int32, userID int64) (*models.Tarantula, error)
	GetAllTarantulas(ctx context.Context, userID int64) ([]models.TarantulaListItem, error)
	GetTarantulasDueFeeding(ctx context.Context, userID int64) ([]models.TarantulaListItem, error)
	UpdateTarantulaEnclosure(ctx context.Context, tarantulaID, enclosureID, userID int64) error

	RecordWeight(ctx context.Context, weight models.WeightRecord) (int64, error)
	GetWeightHistory(ctx context.Context, tarantulaID int32, userID int64, limit int32) ([]models.WeightRecord, error)
	GetLatestWeight(ctx context.Context, tarantulaID int32, userID int64) (*models.WeightRecord, error)
	AddPhoto(ctx context.Context, photo models.TarantulaPhoto) (int64, error)
	GetPhotos(ctx context.Context, tarantulaID int32, userID int64) ([]models.TarantulaPhoto, error)
	GetTarantulaPhotos(ctx context.Context, tarantulaID int32, userID int64, limit int32) ([]models.TarantulaPhoto, error)
	UpdateTarantulaProfilePhoto(ctx context.Context, tarantulaID int32, photoURL string, userID int64) error

	RecordHealthCheck(ctx context.Context, healthCheck models.HealthCheckRecord) error
	RecordMolt(ctx context.Context, molt models.MoltRecord) error
	GetHealthAlerts(ctx context.Context, userID int64) ([]models.HealthAlert, error)
	GetRecentMoltRecords(ctx context.Context, userID int64, limit int32) ([]models.MoltRecord, error)
}

type FeedingService interface {
	RecordFeeding(ctx context.Context, event models.FeedingEvent) (int64, error)
	QuickFeed(ctx context.Context, tarantulaID int32, userID int64) error
	GetFeedingHistory(ctx context.Context, userID int64, limit int32) ([]models.FeedingEvent, error)
	GetRecentFeedingRecords(ctx context.Context, userID int64, limit int32) ([]models.FeedingEvent, error)
	GetFeedingSchedule(ctx context.Context, speciesID int64, bodyLengthCM float32) (*models.FeedingSchedule, error)
}

type ColonyService interface {
	AddColony(ctx context.Context, colony models.CricketColony) error
	GetColonyStatus(ctx context.Context, userID int64) ([]models.ColonyStatus, error)
	UpdateColonyCount(ctx context.Context, colonyID int32, adjustment int32, userID int64) error
	RecordColonyMaintenance(ctx context.Context, record models.ColonyMaintenanceRecord) (int64, error)
	GetColonyMaintenanceHistory(ctx context.Context, colonyID int64, userID int64, limit int32) ([]models.ColonyMaintenanceRecord, error)
	GetMaintenanceTypes(ctx context.Context) ([]models.ColonyMaintenanceType, error)
}

type AnalyticsService interface {
	GetFeedingPatterns(ctx context.Context, userID int64) ([]models.FeedingPattern, error)
	GetAllFeedingPatterns(ctx context.Context, userID int64) ([]models.FeedingPattern, error)
	GetGrowthData(ctx context.Context, userID int64) ([]models.GrowthData, error)
	GetAllGrowthData(ctx context.Context, userID int64) ([]models.GrowthData, error)
	GenerateAnnualReport(ctx context.Context, userID int64, year int) ([]models.AnnualReport, error)
	GetAllAnnualReports(ctx context.Context, year int, userID int64) ([]models.AnnualReport, error)
	GetMoltPredictions(ctx context.Context, userID int64) ([]models.MoltPrediction, error)
	GetAllMoltPredictions(ctx context.Context, userID int64) ([]models.MoltPrediction, error)
}

type NotificationOperations interface {
	GetColonyStatus(ctx context.Context, userID int64) ([]models.ColonyStatus, error)
	GetTarantulasDueFeeding(ctx context.Context, userID int64) ([]models.TarantulaListItem, error)
	GetUserSettings(ctx context.Context, userID int64) (*models.UserSettings, error)
	GetActiveUsers(ctx context.Context) ([]models.TelegramUser, error)
	GetColonyMaintenanceAlerts(ctx context.Context, userID int64) ([]models.ColonyMaintenanceAlert, error)
	GetUpcomingMoltPredictions(ctx context.Context, userID int64, withinDays int) ([]models.MoltPrediction, error)
}
