package models

import (
	"time"

	"gorm.io/gorm"
)

type CricketSizeType struct {
	ID                  int     `json:"id" gorm:"primaryKey"`
	SizeName            string  `json:"size_name" gorm:"unique;not null"`
	ApproximateLengthMM float64 `json:"approximate_length_mm"`
}

type TelegramUser struct {
	ID         int       `json:"id" gorm:"primaryKey"`
	TelegramID int64     `json:"telegram_id" gorm:"unique;not null"`
	Username   string    `json:"username"`
	FirstName  string    `json:"first_name"`
	LastName   string    `json:"last_name"`
	ChatID     int64     `json:"chat_id"`
	IsActive   bool      `json:"is_active" gorm:"default:true"`
	CreatedAt  time.Time `json:"created_at" gorm:"default:CURRENT_TIMESTAMP"`
	LastActive time.Time `json:"last_active" gorm:"default:CURRENT_TIMESTAMP"`
}

type TarantulaSpecies struct {
	ID                            int     `json:"id" gorm:"primaryKey"`
	ScientificName                string  `json:"scientific_name" gorm:"unique;not null"`
	CommonName                    string  `json:"common_name" gorm:"index"`
	AdultSizeCM                   float64 `json:"adult_size_cm"`
	Temperament                   string  `json:"temperament"`
	HumidityRequirementPercent    int     `json:"humidity_requirement_percent"`
	TemperatureRequirementCelsius float64 `json:"temperature_requirement_celsius"`
}

type MoltStage struct {
	ID          int    `json:"id" gorm:"primaryKey"`
	StageName   string `json:"stage_name" gorm:"unique;not null"`
	Description string `json:"description"`
}

type CricketColony struct {
	ID            int          `json:"id" gorm:"primaryKey"`
	ColonyName    string       `json:"colony_name"`
	CurrentCount  int          `json:"current_count"`
	LastCountDate time.Time    `json:"last_count_date" gorm:"index"`
	Notes         string       `json:"notes"`
	CreatedAt     time.Time    `json:"created_at" gorm:"default:CURRENT_TIMESTAMP"`
	UpdatedAt     time.Time    `json:"updated_at" gorm:"default:CURRENT_TIMESTAMP"`
	UserID        int64        `json:"user_id" gorm:"index"`
	User          TelegramUser `json:"user" gorm:"foreignKey:UserID;references:TelegramID"`
}

type Enclosure struct {
	ID               int          `json:"id" gorm:"primaryKey"`
	Name             string       `json:"name"`
	HeightCM         int          `json:"height_cm"`
	WidthCM          int          `json:"width_cm"`
	LengthCM         int          `json:"length_cm"`
	SubstrateDepthCM int          `json:"substrate_depth_cm"`
	Notes            string       `json:"notes"`
	UserID           int64        `json:"user_id" gorm:"index"`
	CreatedAt        time.Time    `json:"created_at" gorm:"default:CURRENT_TIMESTAMP"`
	UpdatedAt        time.Time    `json:"updated_at" gorm:"default:CURRENT_TIMESTAMP"`
	User             TelegramUser `json:"user" gorm:"foreignKey:UserID;references:TelegramID"`
}

type FeedingFrequency struct {
	ID            int    `json:"id" gorm:"primaryKey"`
	FrequencyName string `json:"frequency_name" gorm:"unique;not null"`
	MinDays       int    `json:"min_days" gorm:"not null"`
	MaxDays       int    `json:"max_days" gorm:"not null"`
	Description   string `json:"description"`
}

type FeedingStatus struct {
	ID          int    `json:"id" gorm:"primaryKey"`
	StatusName  string `json:"status_name" gorm:"unique;not null"`
	Description string `json:"description"`
}

type HealthStatus struct {
	ID          int    `json:"id" gorm:"primaryKey"`
	StatusName  string `json:"status_name" gorm:"unique;not null"`
	Description string `json:"description"`
}

type Tarantula struct {
	ID                    int        `json:"id" gorm:"primaryKey"`
	Name                  string     `json:"name"`
	SpeciesID             int        `json:"species_id" gorm:"index"`
	AcquisitionDate       time.Time  `json:"acquisition_date" gorm:"index;not null"`
	LastMoltDate          *time.Time `json:"last_molt_date" gorm:"index"`
	EstimatedAgeMonths    int        `json:"estimated_age_months"`
	CurrentMoltStageID    int        `json:"current_molt_stage_id" gorm:"index"`
	CurrentHealthStatusID int        `json:"current_health_status_id" gorm:"index"`
	LastHealthCheckDate   time.Time  `json:"last_health_check_date"`
	Notes                 string     `json:"notes"`
	CreatedAt             time.Time  `json:"created_at" gorm:"default:CURRENT_TIMESTAMP"`
	UpdatedAt             time.Time  `json:"updated_at" gorm:"default:CURRENT_TIMESTAMP"`
	UserID                int64      `json:"user_id" gorm:"index"`
	EnclosureID           *int       `json:"enclosure_id"`
	CurrentSize           float64    `json:"current_size"`

	// New fields for enhanced tracking
	ProfilePhotoURL    string     `json:"profile_photo_url"`
	CurrentWeightGrams *float64   `json:"current_weight_grams"`
	LastWeighDate      *time.Time `json:"last_weigh_date"`

	Species             TarantulaSpecies `json:"species" gorm:"foreignKey:SpeciesID"`
	CurrentMoltStage    MoltStage        `json:"current_molt_stage" gorm:"foreignKey:CurrentMoltStageID"`
	CurrentHealthStatus HealthStatus     `json:"current_health_status" gorm:"foreignKey:CurrentHealthStatusID"`
	User                TelegramUser     `json:"user" gorm:"foreignKey:UserID;references:TelegramID"`
	Enclosure           Enclosure        `json:"enclosure" gorm:"foreignKey:EnclosureID"`
}

type FeedingSchedule struct {
	ID               int     `json:"id" gorm:"primaryKey"`
	SpeciesID        int     `json:"species_id"`
	SizeCategory     string  `json:"size_category"`
	BodyLengthCM     float64 `json:"body_length_cm"`
	PreySize         string  `json:"prey_size"`
	FeedingFrequency string  `json:"feeding_frequency"`
	PreyType         string  `json:"prey_type"`
	Notes            string  `json:"notes"`
	FrequencyID      int     `json:"frequency_id"`

	Species   TarantulaSpecies `json:"species" gorm:"foreignKey:SpeciesID"`
	Frequency FeedingFrequency `json:"frequency" gorm:"foreignKey:FrequencyID"`
}

type FeedingEvent struct {
	ID               int       `json:"id" gorm:"primaryKey"`
	TarantulaID      int       `json:"tarantula_id" gorm:"index"`
	FeedingDate      time.Time `json:"feeding_date" gorm:"index;not null"`
	CricketColonyID  int       `json:"cricket_colony_id" gorm:"index"`
	NumberOfCrickets int       `json:"number_of_crickets" gorm:"not null"`
	FeedingStatusID  int       `json:"feeding_status_id" gorm:"index"`
	Notes            string    `json:"notes"`
	CreatedAt        time.Time `json:"created_at" gorm:"default:CURRENT_TIMESTAMP"`
	UserID           int64     `json:"user_id" gorm:"index"`

	Tarantula     Tarantula     `json:"tarantula" gorm:"foreignKey:TarantulaID"`
	CricketColony CricketColony `json:"cricket_colony" gorm:"foreignKey:CricketColonyID"`
	FeedingStatus FeedingStatus `json:"feeding_status" gorm:"foreignKey:FeedingStatusID"`
	User          TelegramUser  `json:"user" gorm:"foreignKey:UserID;references:TelegramID"`
}

type HealthCheckRecord struct {
	ID                 int       `json:"id" gorm:"primaryKey"`
	TarantulaID        int       `json:"tarantula_id" gorm:"index"`
	CheckDate          time.Time `json:"check_date" gorm:"index;not null"`
	HealthStatusID     int       `json:"health_status_id" gorm:"index"`
	WeightGrams        float64   `json:"weight_grams"`
	HumidityPercent    int       `json:"humidity_percent"`
	TemperatureCelsius float64   `json:"temperature_celsius"`
	Abnormalities      string    `json:"abnormalities"`
	Notes              string    `json:"notes"`
	CreatedAt          time.Time `json:"created_at" gorm:"default:CURRENT_TIMESTAMP"`
	UserID             int64     `json:"user_id" gorm:"index"`

	Tarantula    Tarantula    `json:"tarantula" gorm:"foreignKey:TarantulaID"`
	HealthStatus HealthStatus `json:"health_status" gorm:"foreignKey:HealthStatusID"`
	User         TelegramUser `json:"user" gorm:"foreignKey:UserID;references:TelegramID"`
}

type MaintenanceRecord struct {
	ID                 int       `json:"id" gorm:"primaryKey"`
	EnclosureID        int       `json:"enclosure_id" gorm:"index"`
	MaintenanceDate    time.Time `json:"maintenance_date" gorm:"index;not null"`
	TemperatureCelsius float64   `json:"temperature_celsius"`
	HumidityPercent    int       `json:"humidity_percent"`
	Notes              string    `json:"notes"`
	UserID             int64     `json:"user_id" gorm:"index"`
	CreatedAt          time.Time `json:"created_at" gorm:"default:CURRENT_TIMESTAMP"`

	Enclosure Enclosure    `json:"enclosure" gorm:"foreignKey:EnclosureID"`
	User      TelegramUser `json:"user" gorm:"foreignKey:UserID;references:TelegramID"`
}

type MoltRecord struct {
	ID               int       `json:"id" gorm:"primaryKey"`
	TarantulaID      int       `json:"tarantula_id" gorm:"index"`
	MoltDate         time.Time `json:"molt_date" gorm:"index;not null"`
	MoltStageID      int       `json:"molt_stage_id" gorm:"index"`
	PreMoltLengthCM  float64   `json:"pre_molt_length_cm"`
	PostMoltLengthCM float64   `json:"post_molt_length_cm"`
	Complications    string    `json:"complications"`
	Notes            string    `json:"notes"`
	CreatedAt        time.Time `json:"created_at" gorm:"default:CURRENT_TIMESTAMP"`
	UserID           int64     `json:"user_id" gorm:"index"`

	Tarantula Tarantula    `json:"tarantula" gorm:"foreignKey:TarantulaID"`
	MoltStage MoltStage    `json:"molt_stage" gorm:"foreignKey:MoltStageID"`
	User      TelegramUser `json:"user" gorm:"foreignKey:UserID;references:TelegramID"`
}

type HealthAlert struct {
	ID             int32  `json:"id" gorm:"column:id"`
	Name           string `json:"name" gorm:"column:name"`
	ScientificName string `json:"scientific_name" gorm:"column:scientific_name"`
	AlertType      string `json:"alert_type" gorm:"column:alert_type"`
	DaysInState    int32  `json:"days_in_state" gorm:"column:days_in_state"`
}

func (HealthAlert) TableName() string {
	return "spider_bot.health_alerts"
}

type ColonyStatus struct {
	ID                int32    `json:"id" gorm:"column:id"`
	ColonyName        string   `json:"colony_name" gorm:"column:colony_name"`
	CurrentCount      int32    `json:"current_count" gorm:"column:current_count"`
	SizeName          string   `json:"size_name" gorm:"column:size_name"` // Changed from SizeType
	CricketsUsed7Days int32    `json:"crickets_used_7_days" gorm:"column:crickets_used_7_days"`
	WeeksRemaining    *float32 `json:"weeks_remaining,omitempty" gorm:"column:weeks_remaining"`
}

func (ColonyStatus) TableName() string {
	return "spider_bot.cricket_colonies"
}

type AddTarantulaParams struct {
	Name               string
	SpeciesID          int32
	AcquisitionDate    string
	EstimatedAgeMonths int32
	EnclosureNumber    *string
	Notes              *string
}

type AddColonyParams struct {
	ColonyName      string
	SizeTypeID      int64
	CurrentCount    int32
	ContainerNumber string
	Notes           *string
}

type TarantulaListItem struct {
	ID               int32   `json:"id"`
	Name             string  `json:"name"`
	SpeciesID        int32   `json:"species_id"`
	SpeciesName      string  `json:"species_name"`
	DaysSinceFeeding float64 `json:"days_since_feeding"`
	CurrentStatus    string  `json:"current_status"`
	FrequencyID      int32   `json:"frequency_id"`
	MinDays          int32   `json:"min_days"`
	MaxDays          int32   `json:"max_days"`
}

type UserSettings struct {
	ID                         int       `json:"id" gorm:"primaryKey"`
	UserID                     int64     `json:"user_id" gorm:"uniqueIndex;not null"`
	NotificationEnabled        bool      `json:"notification_enabled" gorm:"default:true"`
	NotificationTimeUTC        string    `json:"notification_time_utc" gorm:"type:varchar(5);default:'12:00'"`
	FeedingReminderDays        int       `json:"feeding_reminder_days" gorm:"default:7"`
	LowColonyThreshold         int       `json:"low_colony_threshold" gorm:"default:50"`
	CreatedAt                  time.Time `json:"created_at" gorm:"default:CURRENT_TIMESTAMP"`
	UpdatedAt                  time.Time `json:"updated_at" gorm:"default:CURRENT_TIMESTAMP"`
	MaintenanceReminderEnabled bool      `json:"maintenance_reminder_enabled" gorm:"default:true"`
	FoodWaterFrequencyDays     int       `json:"food_water_frequency_days" gorm:"default:3"`
	CleaningFrequencyDays      int       `json:"cleaning_frequency_days" gorm:"default:14"`
	AdultRemovalFrequencyDays  int       `json:"adult_removal_frequency_days" gorm:"default:7"`

	// New pause notification feature
	NotificationsPaused bool       `json:"notifications_paused" gorm:"default:false"`
	PauseStartDate      *time.Time `json:"pause_start_date"`
	PauseEndDate        *time.Time `json:"pause_end_date"`
	PauseReason         string     `json:"pause_reason" gorm:"type:varchar(255)"`

	User TelegramUser `json:"user" gorm:"foreignKey:UserID;references:TelegramID"`
}

func (UserSettings) Defaults() UserSettings {
	return UserSettings{
		NotificationEnabled: true,
		NotificationTimeUTC: "12:00",
		FeedingReminderDays: 7,
		LowColonyThreshold:  50,
	}
}

func (UserSettings) TableName() string {
	return "spider_bot.user_settings"
}

func (u *TelegramUser) BeforeUpdate(tx *gorm.DB) error {
	u.LastActive = time.Now()
	return nil
}

func (s *UserSettings) BeforeUpdate(tx *gorm.DB) error {
	s.UpdatedAt = time.Now()
	return nil
}

type ColonyMaintenanceType struct {
	ID            int    `json:"id" gorm:"primaryKey"`
	TypeName      string `json:"type_name" gorm:"unique;not null"`
	Description   string `json:"description"`
	FrequencyDays int    `json:"frequency_days" gorm:"default:7"`
}

type ColonyMaintenanceRecord struct {
	ID                int       `json:"id" gorm:"primaryKey"`
	ColonyID          int       `json:"colony_id" gorm:"index"`
	MaintenanceTypeID int       `json:"maintenance_type_id" gorm:"index"`
	MaintenanceDate   time.Time `json:"maintenance_date" gorm:"index;not null"`
	Notes             string    `json:"notes"`
	UserID            int64     `json:"user_id" gorm:"index"`
	CreatedAt         time.Time `json:"created_at" gorm:"default:CURRENT_TIMESTAMP"`

	Colony          CricketColony         `json:"colony" gorm:"foreignKey:ColonyID"`
	MaintenanceType ColonyMaintenanceType `json:"maintenance_type" gorm:"foreignKey:MaintenanceTypeID"`
	User            TelegramUser          `json:"user" gorm:"foreignKey:UserID;references:TelegramID"`
}

type ColonyMaintenanceSchedule struct {
	ID                int        `json:"id" gorm:"primaryKey"`
	ColonyID          int        `json:"colony_id" gorm:"index"`
	MaintenanceTypeID int        `json:"maintenance_type_id" gorm:"index"`
	FrequencyDays     int        `json:"frequency_days" gorm:"not null"`
	Enabled           bool       `json:"enabled" gorm:"default:true"`
	LastPerformedDate *time.Time `json:"last_performed_date"`
	UserID            int64      `json:"user_id" gorm:"index"`
	CreatedAt         time.Time  `json:"created_at" gorm:"default:CURRENT_TIMESTAMP"`
	UpdatedAt         time.Time  `json:"updated_at" gorm:"default:CURRENT_TIMESTAMP"`

	Colony          CricketColony         `json:"colony" gorm:"foreignKey:ColonyID"`
	MaintenanceType ColonyMaintenanceType `json:"maintenance_type" gorm:"foreignKey:MaintenanceTypeID"`
	User            TelegramUser          `json:"user" gorm:"foreignKey:UserID;references:TelegramID"`
}

type ColonyMaintenanceAlert struct {
	ID                int32  `json:"id" gorm:"column:id;primaryKey"`
	ColonyName        string `json:"colony_name" gorm:"column:colony_name"`
	MaintenanceType   string `json:"maintenance_type" gorm:"column:maintenance_type"`
	DaysSinceLastDone int32  `json:"days_since_last_done" gorm:"column:days_since_last_done"`
	DaysOverdue       int32  `json:"days_overdue" gorm:"column:days_overdue"`
}

func (ColonyMaintenanceAlert) TableName() string {
	return "spider_bot.colony_maintenance_alerts"
}

// New model for weight tracking
type WeightRecord struct {
	ID          int       `json:"id" gorm:"primaryKey"`
	TarantulaID int       `json:"tarantula_id" gorm:"index"`
	WeightGrams float64   `json:"weight_grams" gorm:"not null"`
	WeighDate   time.Time `json:"weigh_date" gorm:"index;not null"`
	Notes       string    `json:"notes"`
	UserID      int64     `json:"user_id" gorm:"index"`
	CreatedAt   time.Time `json:"created_at" gorm:"default:CURRENT_TIMESTAMP"`

	Tarantula Tarantula    `json:"tarantula" gorm:"foreignKey:TarantulaID"`
	User      TelegramUser `json:"user" gorm:"foreignKey:UserID;references:TelegramID"`
}

// New model for photos
type TarantulaPhoto struct {
	ID          int       `json:"id" gorm:"primaryKey"`
	TarantulaID int       `json:"tarantula_id" gorm:"index"`
	PhotoURL    string    `json:"photo_url" gorm:"not null"`
	PhotoData   []byte    `json:"photo_data" gorm:"type:bytea"`
	PhotoType   string    `json:"photo_type" gorm:"default:'general'"` // general, pre-molt, post-molt
	Caption     string    `json:"caption"`
	TakenDate   time.Time `json:"taken_date" gorm:"index;not null"`
	UserID      int64     `json:"user_id" gorm:"index"`
	CreatedAt   time.Time `json:"created_at" gorm:"default:CURRENT_TIMESTAMP"`

	Tarantula Tarantula    `json:"tarantula" gorm:"foreignKey:TarantulaID"`
	User      TelegramUser `json:"user" gorm:"foreignKey:UserID;references:TelegramID"`
}

// Analytics and reporting models
type FeedingPattern struct {
	TarantulaID          int32      `json:"tarantula_id"`
	TarantulaName        string     `json:"tarantula_name"`
	TotalFeedings        int32      `json:"total_feedings"`
	AverageInterval      float64    `json:"average_interval_days"`
	AcceptanceRate       float64    `json:"acceptance_rate_percent"`
	LastFeedingDate      *time.Time `json:"last_feeding_date"`
	DaysSinceLastFeeding int32      `json:"days_since_last_feeding"`
	FeedingRegularity    string     `json:"feeding_regularity"` // "Regular", "Irregular", "Inconsistent"
	CricketsPerWeek      float64    `json:"crickets_per_week"`
}

type GrowthData struct {
	TarantulaID       int32         `json:"tarantula_id"`
	TarantulaName     string        `json:"tarantula_name"`
	CurrentWeight     *float64      `json:"current_weight_grams"`
	CurrentSize       float64       `json:"current_size_cm"`
	WeightHistory     []WeightPoint `json:"weight_history"`
	SizeHistory       []SizePoint   `json:"size_history"`
	GrowthRate        *float64      `json:"growth_rate_grams_per_month"`
	WeightChangeTotal float64       `json:"total_weight_change"`
	SizeChangeTotal   float64       `json:"total_size_change"`
}

type WeightPoint struct {
	Date   time.Time `json:"date"`
	Weight float64   `json:"weight_grams"`
}

type SizePoint struct {
	Date time.Time `json:"date"`
	Size float64   `json:"size_cm"`
}

type AnnualReport struct {
	Year          int    `json:"year"`
	TarantulaID   int32  `json:"tarantula_id"`
	TarantulaName string `json:"tarantula_name"`

	// Feeding statistics
	TotalFeedings   int32   `json:"total_feedings"`
	TotalCrickets   int32   `json:"total_crickets_consumed"`
	AcceptanceRate  float64 `json:"acceptance_rate"`
	AverageInterval float64 `json:"average_feeding_interval"`

	// Growth statistics
	StartWeight *float64 `json:"start_weight_grams"`
	EndWeight   *float64 `json:"end_weight_grams"`
	WeightGain  *float64 `json:"weight_gain_grams"`
	StartSize   float64  `json:"start_size_cm"`
	EndSize     float64  `json:"end_size_cm"`
	SizeGrowth  float64  `json:"size_growth_cm"`

	// Health and events
	MoltCount    int32 `json:"molt_count"`
	PhotosAdded  int32 `json:"photos_added"`
	HealthIssues int32 `json:"health_issues"`

	// Milestones
	Milestones []string `json:"milestones"`

	// Cost estimates
	EstimatedCost float64 `json:"estimated_cricket_cost"`
}

type MoltPrediction struct {
	TarantulaID   int32  `json:"tarantula_id"`
	TarantulaName string `json:"tarantula_name"`

	// Historical data
	LastMoltDate      *time.Time `json:"last_molt_date"`
	DaysSinceLastMolt int32      `json:"days_since_last_molt"`
	AverageMoltCycle  *float64   `json:"average_molt_cycle_days"`
	MoltCount         int32      `json:"total_molts"`

	// Prediction
	PredictedMoltDate *time.Time `json:"predicted_molt_date"`
	ConfidenceLevel   string     `json:"confidence_level"` // "High", "Medium", "Low"
	DaysUntilMolt     *int32     `json:"days_until_predicted_molt"`

	// Indicators
	PreMoltSigns    []string `json:"pre_molt_signs"`
	SizeIndicator   string   `json:"size_indicator"`   // "Small", "Medium", "Large", "Adult"
	FeedingBehavior string   `json:"feeding_behavior"` // "Normal", "Reduced", "Stopped"

	// Reasoning
	PredictionBasis string `json:"prediction_basis"`
	Recommendation  string `json:"recommendation"`
}
