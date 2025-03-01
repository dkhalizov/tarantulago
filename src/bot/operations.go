package bot

import (
	"context"
	"tarantulago/models"
)

type TarantulaOperations interface {
	AddTarantula(ctx context.Context, t models.Tarantula) error
	GetTarantulaByID(ctx context.Context, userID int64, id int32) (*models.Tarantula, error)
	GetAllTarantulas(ctx context.Context, userID int64) ([]models.TarantulaListItem, error)
	GetTarantulasDueFeeding(ctx context.Context, userID int64) ([]models.TarantulaListItem, error)
	UpdateTarantulaEnclosure(ctx context.Context, tarantulaID, enclosureID, userID int64) error

	RecordFeeding(ctx context.Context, event models.FeedingEvent) (int64, error)
	GetRecentFeedingRecords(ctx context.Context, userID int64, limit int32) ([]models.FeedingEvent, error)
	GetFeedingSchedule(ctx context.Context, speciesID int64, bodyLengthCM float32) (*models.FeedingSchedule, error)
	GetFeedingFrequency(ctx context.Context, id int64) (*models.FeedingFrequency, error)

	RecordHealthCheck(ctx context.Context, healthCheck models.HealthCheckRecord) error
	GetRecentHealthRecords(ctx context.Context, userID int64, limit int32) ([]models.HealthCheckRecord, error)
	GetHealthAlerts(ctx context.Context, userID int64) ([]models.HealthAlert, error)

	RecordMolt(ctx context.Context, molt models.MoltRecord) error
	GetRecentMoltRecords(ctx context.Context, userID int64, limit int32) ([]models.MoltRecord, error)

	AddColony(ctx context.Context, colony models.CricketColony) error
	GetColonyStatus(ctx context.Context, userID int64) ([]models.ColonyStatus, error)
	UpdateColonyCount(ctx context.Context, colonyID int32, adjustment int32, userID int64) error

	CreateMaintenanceRecord(ctx context.Context, record models.MaintenanceRecord) (int64, error)
	GetMaintenanceHistory(ctx context.Context, enclosureID, userID int64) ([]models.MaintenanceRecord, error)
	GetMaintenanceTasks(ctx context.Context, userID int64) ([]models.MaintenanceRecord, error)

	CreateEnclosure(ctx context.Context, enclosure models.Enclosure) (int64, error)
	GetEnclosure(ctx context.Context, id, userID int64) (*models.Enclosure, error)

	EnsureUserExists(ctx context.Context, user *models.TelegramUser) error
	GetCurrentSize(ctx context.Context, tarantulaID int32) (float32, error)

	GetUserSettings(ctx context.Context, userID int64) (*models.UserSettings, error)
	UpdateUserSettings(ctx context.Context, settings *models.UserSettings) error

	RecordColonyMaintenance(ctx context.Context, record models.ColonyMaintenanceRecord) (int64, error)
	GetColonyMaintenanceAlerts(ctx context.Context, userID int64) ([]models.ColonyMaintenanceAlert, error)
	GetColonyMaintenanceHistory(ctx context.Context, colonyID int64, userID int64, limit int32) ([]models.ColonyMaintenanceRecord, error)
	GetMaintenanceTypes(ctx context.Context) ([]models.ColonyMaintenanceType, error)

	NotificationOperations
}
