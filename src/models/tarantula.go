package models

import (
	"gorm.io/gorm"
	"time"
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
	ID                  int       `json:"id" gorm:"primaryKey"`
	UserID              int64     `json:"user_id" gorm:"uniqueIndex;not null"`
	NotificationEnabled bool      `json:"notification_enabled" gorm:"default:true"`
	NotificationTimeUTC string    `json:"notification_time_utc" gorm:"type:varchar(5);default:'12:00'"`
	FeedingReminderDays int       `json:"feeding_reminder_days" gorm:"default:7"`
	LowColonyThreshold  int       `json:"low_colony_threshold" gorm:"default:50"`
	CreatedAt           time.Time `json:"created_at" gorm:"default:CURRENT_TIMESTAMP"`
	UpdatedAt           time.Time `json:"updated_at" gorm:"default:CURRENT_TIMESTAMP"`

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
