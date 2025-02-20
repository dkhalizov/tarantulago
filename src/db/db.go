package db

import (
	"context"
	"fmt"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/schema"
	"tarantulago/models"
	"time"
)

type TarantulaDB struct {
	db *gorm.DB
}

func NewTarantulaDB(connectionString string) (*TarantulaDB, error) {
	db, err := gorm.Open(postgres.Open(connectionString), &gorm.Config{
		NamingStrategy: schema.NamingStrategy{
			TablePrefix: "spider_bot.",
		},
	})
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}
	db = db.Set("gorm:table_options", "schema spider_bot")
	err = db.AutoMigrate(
		&models.Tarantula{},
		&models.TarantulaSpecies{},
		&models.MoltStage{},
		&models.HealthStatus{},
		&models.FeedingEvent{},
		&models.CricketColony{},
		&models.Enclosure{},
		&models.FeedingFrequency{},
		&models.FeedingSchedule{},
		&models.HealthCheckRecord{},
		&models.MaintenanceRecord{},
		&models.MoltRecord{},
		&models.TelegramUser{},
	)
	if err != nil {
		return nil, fmt.Errorf("failed to migrate database: %w", err)
	}

	return &TarantulaDB{db: db}, nil
}

func (db *TarantulaDB) GetRecentHealthRecords(ctx context.Context, userID int64, limit int32) ([]models.HealthCheckRecord, error) {
	var records []models.HealthCheckRecord

	result := db.db.WithContext(ctx).
		Table("spider_bot.health_check_records").
		Select(`
            t.name as tarantula_name,
            health_check_records.check_date,
            hs.status_name as status,
            health_check_records.weight_grams,
            health_check_records.humidity_percent,
            health_check_records.temperature_celsius,
            health_check_records.notes
        `).
		Joins("JOIN spider_bot.tarantulas t ON health_check_records.tarantula_id = t.id").
		Joins("JOIN spider_bot.health_statuses hs ON health_check_records.health_status_id = hs.id").
		Where("t.user_id = ?", userID).
		Order("health_check_records.check_date DESC").
		Limit(int(limit)).
		Find(&records)

	if result.Error != nil {
		return nil, fmt.Errorf("failed to get health records: %w", result.Error)
	}

	return records, nil
}

func (db *TarantulaDB) AddTarantula(ctx context.Context, tarantula models.Tarantula) error {
	result := db.db.WithContext(ctx).Create(&tarantula)
	if result.Error != nil {
		return fmt.Errorf("failed to create tarantula: %w", result.Error)
	}

	return nil
}

func (db *TarantulaDB) RecordFeeding(ctx context.Context, event models.FeedingEvent) (int64, error) {
	var id int64
	err := db.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {

		var tarantula models.Tarantula
		if err := tx.Where("id = ? AND user_id = ?", event.TarantulaID, event.UserID).First(&tarantula).Error; err != nil {
			return fmt.Errorf("tarantula not found or access denied: %w", err)
		}

		result := tx.Model(&models.CricketColony{}).
			Where("id = ? AND user_id = ? AND current_count >= ?",
				event.CricketColonyID, event.UserID, event.NumberOfCrickets).
			UpdateColumn("current_count", gorm.Expr("current_count - ?", event.NumberOfCrickets))

		if result.Error != nil {
			return fmt.Errorf("failed to update colony count: %w", result.Error)
		}
		if result.RowsAffected == 0 {
			return fmt.Errorf("colony not found, access denied, or insufficient crickets")
		}

		feedingEvent := models.FeedingEvent{
			TarantulaID:      event.TarantulaID,
			FeedingDate:      time.Now(),
			CricketColonyID:  event.CricketColonyID,
			NumberOfCrickets: event.NumberOfCrickets,
			FeedingStatusID:  int(models.FeedingStatusAccepted),
			Notes:            event.Notes,
			UserID:           event.UserID,
		}

		if err := tx.Create(&feedingEvent).Error; err != nil {
			return fmt.Errorf("failed to create feeding event: %w", err)
		}

		id = int64(feedingEvent.ID)
		return nil
	})

	if err != nil {
		return 0, err
	}

	return id, nil
}

func (db *TarantulaDB) GetTarantulasDueFeeding(ctx context.Context, userID int64) ([]models.TarantulaListItem, error) {
	var items []models.TarantulaListItem

	result := db.db.WithContext(ctx).Raw(`
        WITH LastFeeding AS (
    SELECT 
        tarantula_id,
        MAX(feeding_date) as last_feeding_date,
        EXTRACT(EPOCH FROM (CURRENT_DATE - MAX(feeding_date)))/86400 as days_since_feeding
    FROM spider_bot.feeding_events
    GROUP BY tarantula_id
),
SizeBoundaries AS (
    SELECT 
        size_category,
        MIN(body_length_cm) as min_size,
        MAX(body_length_cm) as max_size,
        -- Create an ordering for size categories
        CASE size_category
            WHEN 'Spiderling' THEN 1
            WHEN 'Juvenile' THEN 2
            WHEN 'Sub-Adult' THEN 3
            WHEN 'Adult' THEN 4
        END as category_order
    FROM spider_bot.feeding_schedules
    GROUP BY size_category
),
TarantulaSize AS (
    SELECT 
        t.id as tarantula_id,
        COALESCE(t.current_size,
            CASE 
                WHEN ts.adult_size_cm <= 8 THEN ts.adult_size_cm * 0.3
                ELSE ts.adult_size_cm * 0.4
            END
        ) as current_size_cm
    FROM spider_bot.tarantulas t
    JOIN spider_bot.tarantula_species ts ON t.species_id = ts.id
),
MatchingSchedule AS (
    SELECT DISTINCT ON (t.id)
        t.id as tarantula_id,
        fs.frequency_id,
        ff.min_days,
        ff.max_days
    FROM spider_bot.tarantulas t
    JOIN TarantulaSize ts ON t.id = ts.tarantula_id
    JOIN spider_bot.feeding_schedules fs ON t.species_id = fs.species_id
    JOIN spider_bot.feeding_frequencies ff ON fs.frequency_id = ff.id
    JOIN SizeBoundaries sb ON fs.size_category = sb.size_category
    WHERE ts.current_size_cm <= sb.max_size
      AND ts.current_size_cm > COALESCE(
        (SELECT MAX(max_size) 
         FROM SizeBoundaries sb2 
         WHERE sb2.category_order < sb.category_order),
        0
      )
    ORDER BY t.id, sb.category_order DESC
)
SELECT DISTINCT
    t.id,
    t.name,
    ts.id as species_id,
    ts.common_name as species_name,
    COALESCE(lf.days_since_feeding, 999) as days_since_feeding,
    ms.frequency_id,
    ms.min_days,
    ms.max_days,
    CASE
        WHEN molt.stage_name = 'Pre-molt' THEN 'In pre-molt'
        WHEN lf.days_since_feeding IS NULL THEN 'Never fed'
        WHEN lf.days_since_feeding > ms.max_days THEN 'Overdue feeding'
        WHEN lf.days_since_feeding > ms.min_days THEN 'Due for feeding'
        ELSE 'Recently fed'
    END as current_status
FROM spider_bot.tarantulas t
JOIN spider_bot.tarantula_species ts ON t.species_id = ts.id
LEFT JOIN spider_bot.molt_stages molt ON t.current_molt_stage_id = molt.id
LEFT JOIN LastFeeding lf ON t.id = lf.tarantula_id
LEFT JOIN MatchingSchedule ms ON t.id = ms.tarantula_id
WHERE t.user_id = ?
    AND molt.stage_name != 'Pre-molt'
    AND (
        lf.days_since_feeding IS NULL 
        OR lf.days_since_feeding > ms.min_days
    )
ORDER BY days_since_feeding DESC`, userID).
		Scan(&items)

	if result.Error != nil {
		return nil, fmt.Errorf("failed to get tarantulas due feeding: %w", result.Error)
	}

	return items, nil
}

func (db *TarantulaDB) GetTarantulaByID(ctx context.Context, userID int64, id int32) (*models.Tarantula, error) {
	var tarantula models.Tarantula

	result := db.db.WithContext(ctx).
		Preload("Species").
		Preload("CurrentMoltStage").
		Preload("CurrentHealthStatus").
		Preload("Enclosure").
		Where("id = ? AND user_id = ?", id, userID).
		First(&tarantula)

	if result.Error != nil {
		if result.Error == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("tarantula not found")
		}
		return nil, fmt.Errorf("failed to get tarantula: %w", result.Error)
	}

	return &tarantula, nil
}

func (db *TarantulaDB) GetRecentFeedingRecords(ctx context.Context, userID int64, limit int32) ([]models.FeedingEvent, error) {
	var records []models.FeedingEvent

	result := db.db.WithContext(ctx).
		Preload("Tarantula").
		Preload("CricketColony").
		Preload("FeedingStatus").
		Preload("User").
		Where("user_id = ?", userID).
		Order("feeding_date DESC").
		Limit(int(limit)).
		Find(&records)

	if result.Error != nil {
		return nil, fmt.Errorf("failed to get feeding records: %w", result.Error)
	}

	return records, nil
}

func (db *TarantulaDB) GetAllTarantulas(ctx context.Context, userID int64) ([]models.TarantulaListItem, error) {
	var items []models.TarantulaListItem

	result := db.db.WithContext(ctx).Raw(`
        SELECT
            t.id,
            t.name,
            ts.common_name as species_name,
            COALESCE(EXTRACT(EPOCH FROM CURRENT_DATE::timestamp - MAX(f.feeding_date)::timestamp)/86400, 365) as days_since_feeding,
            CASE
                WHEN ms.stage_name = 'Pre-molt' THEN 'In pre-molt'
                WHEN hs.status_name = 'Critical' THEN 'Critical'
                WHEN EXTRACT(EPOCH FROM CURRENT_DATE::timestamp - MAX(f.feeding_date)::timestamp)/86400 > 14 THEN 'Needs feeding'
                ELSE 'Normal'
            END as current_status
        FROM spider_bot.tarantulas t
        JOIN spider_bot.tarantula_species ts ON t.species_id = ts.id
        LEFT JOIN spider_bot.molt_stages ms ON t.current_molt_stage_id = ms.id
        LEFT JOIN spider_bot.health_statuses hs ON t.current_health_status_id = hs.id
        LEFT JOIN spider_bot.feeding_events f ON t.id = f.tarantula_id
        WHERE t.user_id = ?
        GROUP BY t.id, t.name, ts.common_name, ms.stage_name, hs.status_name
        ORDER BY t.name`, userID).
		Scan(&items)

	if result.Error != nil {
		return nil, fmt.Errorf("failed to get all tarantulas: %w", result.Error)
	}

	return items, nil
}

func (db *TarantulaDB) UpdateTarantulaEnclosure(ctx context.Context, tarantulaID, enclosureID, userID int64) error {
	result := db.db.WithContext(ctx).
		Model(&models.Tarantula{}).
		Where("id = ? AND user_id = ?", tarantulaID, userID).
		Update("enclosure_id", enclosureID)

	if result.Error != nil {
		return fmt.Errorf("failed to update tarantula enclosure: %w", result.Error)
	}

	if result.RowsAffected == 0 {
		return fmt.Errorf("tarantula not found or access denied")
	}

	return nil
}

func (db *TarantulaDB) GetFeedingSchedule(ctx context.Context, speciesID int64, bodyLengthCM float32) (*models.FeedingSchedule, error) {
	var schedule models.FeedingSchedule

	result := db.db.WithContext(ctx).
		Preload("Species").
		Preload("Frequency").
		Where("species_id = ? AND body_length_cm >= ?", speciesID, bodyLengthCM).
		Order("body_length_cm ASC").
		First(&schedule)

	if result.Error != nil {
		if result.Error == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get feeding schedule: %w", result.Error)
	}

	return &schedule, nil
}

func (db *TarantulaDB) GetFeedingFrequency(ctx context.Context, id int64) (*models.FeedingFrequency, error) {
	var frequency models.FeedingFrequency

	result := db.db.WithContext(ctx).First(&frequency, id)
	if result.Error != nil {
		if result.Error == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get feeding frequency: %w", result.Error)
	}

	return &frequency, nil
}

func (db *TarantulaDB) RecordHealthCheck(ctx context.Context, healthCheck models.HealthCheckRecord) error {
	return db.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		result := tx.Model(&models.Tarantula{}).
			Where("id = ? AND user_id = ?", healthCheck.TarantulaID, healthCheck.UserID).
			Updates(map[string]interface{}{
				"last_health_check_date":   time.Now(),
				"current_health_status_id": healthCheck.HealthStatusID,
			})

		if result.Error != nil {
			return fmt.Errorf("failed to update tarantula health status: %w", result.Error)
		}

		if result.RowsAffected == 0 {
			return fmt.Errorf("tarantula not found or access denied")
		}

		//healthCheck := models.HealthCheckRecord{
		//	TarantulaID:        tarantulaID,
		//	CheckDate:          time.Now(),
		//	HealthStatusID:     int32(status),
		//	WeightGrams:        0, // TODO: Add these to parameters
		//	HumidityPercent:    55,
		//	TemperatureCelsius: 20,
		//	Notes:              notes,
		//	UserID:             userID,
		//}

		if err := tx.Create(&healthCheck).Error; err != nil {
			return fmt.Errorf("failed to create health check record: %w", err)
		}

		return nil
	})
}

func (db *TarantulaDB) GetHealthAlerts(ctx context.Context, userID int64) ([]models.HealthAlert, error) {
	var alerts []models.HealthAlert

	result := db.db.WithContext(ctx).Raw(`
        WITH alerts AS (
            SELECT 
                t.id,
                t.name,
                ts.scientific_name,
                CASE
                    WHEN EXTRACT(EPOCH FROM CURRENT_DATE::timestamp - t.last_health_check_date::timestamp)/86400 >= 30 
                        THEN 'Overdue Health Check'
                    WHEN EXTRACT(EPOCH FROM CURRENT_DATE::timestamp - MAX(f.feeding_date)::timestamp)/86400 >= 14 
                        AND ms.stage_name != 'Pre-molt' 
                        THEN 'Extended Feeding Strike'
                    WHEN ms.stage_name = 'Pre-molt' 
                        AND EXTRACT(EPOCH FROM CURRENT_DATE::timestamp - t.last_molt_date::timestamp)/86400 >= 180
                        THEN 'Extended Pre-molt'
                    ELSE 'None'
                END as alert_type,
                CASE
                    WHEN EXTRACT(EPOCH FROM CURRENT_DATE::timestamp - t.last_health_check_date::timestamp)/86400 >= 30
                        THEN EXTRACT(EPOCH FROM CURRENT_DATE::timestamp - t.last_health_check_date::timestamp)/86400
                    WHEN EXTRACT(EPOCH FROM CURRENT_DATE::timestamp - MAX(f.feeding_date)::timestamp)/86400 >= 14
                        THEN EXTRACT(EPOCH FROM CURRENT_DATE::timestamp - MAX(f.feeding_date)::timestamp)/86400
                    WHEN ms.stage_name = 'Pre-molt'
                        THEN EXTRACT(EPOCH FROM CURRENT_DATE::timestamp - t.last_molt_date::timestamp)/86400
                    ELSE 0
                END::INTEGER as days_in_state
            FROM spider_bot.tarantulas t
            JOIN spider_bot.tarantula_species ts ON t.species_id = ts.id
            LEFT JOIN spider_bot.feeding_events f ON t.id = f.tarantula_id
            LEFT JOIN spider_bot.molt_stages ms ON t.current_molt_stage_id = ms.id
            WHERE t.user_id = ?
            GROUP BY t.id, t.name, ts.scientific_name, t.last_health_check_date, t.last_molt_date, ms.stage_name
        )
        SELECT * FROM alerts
        WHERE alert_type != 'None'
        ORDER BY days_in_state DESC`, userID).
		Scan(&alerts)

	if result.Error != nil {
		return nil, fmt.Errorf("failed to get health alerts: %w", result.Error)
	}

	return alerts, nil
}

func (db *TarantulaDB) RecordMolt(ctx context.Context, molt models.MoltRecord) error {
	return db.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {

		result := tx.Model(&models.Tarantula{}).
			Where("id = ? AND user_id = ?", molt.TarantulaID, molt.UserID).
			Updates(map[string]interface{}{
				"last_molt_date":        time.Now(),
				"current_molt_stage_id": models.MoltStagePostMolt,
			})

		if result.Error != nil {
			return fmt.Errorf("failed to update tarantula molt status: %w", result.Error)
		}

		if result.RowsAffected == 0 {
			return fmt.Errorf("tarantula not found or access denied")
		}

		//// Create molt record
		//molt := models.MoltRecord{
		//	TarantulaID:      tarantulaID,
		//	MoltDate:         time.Now(),
		//	MoltStageID:      int32(models.MoltStagePostMolt),
		//	PostMoltLengthCM: &lengthCM,
		//	Complications:    complications,
		//	Notes:            notes,
		//	UserID:           userID,
		//}

		if err := tx.Create(&molt).Error; err != nil {
			return fmt.Errorf("failed to create molt record: %w", err)
		}

		return nil
	})
}

func (db *TarantulaDB) GetRecentMoltRecords(ctx context.Context, userID int64, limit int32) ([]models.MoltRecord, error) {
	var records []models.MoltRecord

	result := db.db.WithContext(ctx).
		Preload("Tarantula").
		Preload("MoltStage").
		Model(&models.MoltRecord{}).
		Joins("JOIN spider_bot.tarantulas ON spider_bot.molt_records.tarantula_id = spider_bot.tarantulas.id").
		Where("spider_bot.tarantulas.user_id = ?", userID).
		Order("spider_bot.molt_records.molt_date DESC").
		Limit(int(limit)).
		Find(&records)

	if result.Error != nil {
		return nil, fmt.Errorf("failed to get molt records: %w", result.Error)
	}

	return records, nil
}

func (db *TarantulaDB) AddColony(ctx context.Context, colony models.CricketColony) error {
	result := db.db.WithContext(ctx).Create(&colony)
	if result.Error != nil {
		return fmt.Errorf("failed to create colony: %w", result.Error)
	}

	return nil
}

func (db *TarantulaDB) GetColonyStatus(ctx context.Context, userID int64) ([]models.ColonyStatus, error) {
	var colonies []models.ColonyStatus

	result := db.db.WithContext(ctx).Raw(`
        SELECT
            cc.id,
            cc.colony_name,
            cc.current_count,
            COALESCE(SUM(fe.number_of_crickets), 0) as crickets_used_7_days,
            CASE
                WHEN SUM(fe.number_of_crickets) > 0
                THEN cc.current_count::FLOAT / (SUM(fe.number_of_crickets)::FLOAT / 7.0)
                ELSE NULL
            END as weeks_remaining
        FROM spider_bot.cricket_colonies cc
        LEFT JOIN spider_bot.feeding_events fe ON cc.id = fe.cricket_colony_id
            AND fe.feeding_date >= CURRENT_DATE - INTERVAL '7 days'
        WHERE cc.user_id = ?
        GROUP BY cc.id, cc.colony_name, cc.current_count
        ORDER BY weeks_remaining ASC NULLS LAST`, userID).
		Scan(&colonies)

	if result.Error != nil {
		return nil, fmt.Errorf("failed to get colony status: %w", result.Error)
	}

	return colonies, nil
}

func (db *TarantulaDB) UpdateColonyCount(ctx context.Context, colonyID int32, adjustment int32, userID int64) error {
	result := db.db.WithContext(ctx).
		Model(&models.CricketColony{}).
		Where("id = ? AND user_id = ?", colonyID, userID).
		Update("current_count", gorm.Expr("current_count + ?", adjustment))

	if result.Error != nil {
		return fmt.Errorf("failed to update colony count: %w", result.Error)
	}

	if result.RowsAffected == 0 {
		return fmt.Errorf("colony not found or access denied")
	}

	return nil
}

func (db *TarantulaDB) CreateMaintenanceRecord(ctx context.Context, record models.MaintenanceRecord) (int64, error) {
	result := db.db.WithContext(ctx).Create(&record)
	if result.Error != nil {
		return 0, fmt.Errorf("failed to create maintenance record: %w", result.Error)
	}

	return int64(record.ID), nil
}

func (db *TarantulaDB) GetMaintenanceHistory(ctx context.Context, enclosureID, userID int64) ([]models.MaintenanceRecord, error) {
	var records []models.MaintenanceRecord

	result := db.db.WithContext(ctx).
		Where("enclosure_id = ? AND user_id = ?", enclosureID, userID).
		Order("maintenance_date DESC").
		Find(&records)

	if result.Error != nil {
		return nil, fmt.Errorf("failed to get maintenance history: %w", result.Error)
	}

	return records, nil
}

func (db *TarantulaDB) GetMaintenanceTasks(ctx context.Context, userID int64) ([]models.MaintenanceRecord, error) {
	var tasks []models.MaintenanceRecord

	result := db.db.WithContext(ctx).Raw(`
        SELECT
            t.id,
            t.name,
            ts.scientific_name,
            CASE
                WHEN EXTRACT(EPOCH FROM CURRENT_DATE::timestamp - t.last_health_check_date::timestamp)/86400 >= 30 
                    THEN 'Health Check Required'
                WHEN EXTRACT(EPOCH FROM CURRENT_DATE::timestamp - MAX(f.feeding_date)::timestamp)/86400 >= 7
                    AND ms.stage_name != 'Pre-molt' THEN 'Feeding Due'
                WHEN ms.stage_name = 'Pre-molt' THEN 'Monitor for Molt'
                ELSE 'Regular Check'
            END as required_action,
            CASE
                WHEN EXTRACT(EPOCH FROM CURRENT_DATE::timestamp - t.last_health_check_date::timestamp)/86400 >= 30 THEN 1
                WHEN EXTRACT(EPOCH FROM CURRENT_DATE::timestamp - MAX(f.feeding_date)::timestamp)/86400 >= 7 THEN 2
                WHEN ms.stage_name = 'Pre-molt' THEN 3
                ELSE 4
            END as priority
        FROM spider_bot.tarantulas t
        JOIN spider_bot.tarantula_species ts ON t.species_id = ts.id
        LEFT JOIN spider_bot.feeding_events f ON t.id = f.tarantula_id
        LEFT JOIN spider_bot.molt_stages ms ON t.current_molt_stage_id = ms.id
        WHERE t.user_id = ?
        GROUP BY t.id, t.name, ts.scientific_name, 
                 t.last_health_check_date, ms.stage_name
        ORDER BY priority, name`, userID).
		Scan(&tasks)

	if result.Error != nil {
		return nil, fmt.Errorf("failed to get maintenance tasks: %w", result.Error)
	}

	return tasks, nil
}

func (db *TarantulaDB) CreateEnclosure(ctx context.Context, enclosure models.Enclosure) (int64, error) {
	result := db.db.WithContext(ctx).Create(&enclosure)
	if result.Error != nil {
		return 0, fmt.Errorf("failed to create enclosure: %w", result.Error)
	}

	return int64(enclosure.ID), nil
}

func (db *TarantulaDB) GetEnclosure(ctx context.Context, id, userID int64) (*models.Enclosure, error) {
	var enclosure models.Enclosure

	result := db.db.WithContext(ctx).
		Where("id = ? AND user_id = ?", id, userID).
		First(&enclosure)

	if result.Error != nil {
		if result.Error == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("enclosure not found")
		}
		return nil, fmt.Errorf("failed to get enclosure: %w", result.Error)
	}

	return &enclosure, nil
}

func (db *TarantulaDB) GetCurrentSize(ctx context.Context, tarantulaID int32) (float32, error) {
	var size float32

	result := db.db.WithContext(ctx).Raw(`
        WITH LastMoltSize AS (
            SELECT 
                tarantula_id,
                post_molt_length_cm,
                molt_date
            FROM spider_bot.molt_records 
            WHERE molt_date = (
                SELECT MAX(molt_date)
                FROM spider_bot.molt_records mr2
                WHERE mr2.tarantula_id = molt_records.tarantula_id
                AND mr2.post_molt_length_cm IS NOT NULL
            )
        ),
        EstimatedSize AS (
            SELECT
                t.id as tarantula_id,
                COALESCE(
                    lms.post_molt_length_cm,
                    CASE 
                        WHEN t.current_molt_stage_id IS NOT NULL THEN
                            (SELECT fs.body_length_cm
                             FROM spider_bot.feeding_schedules fs
                             WHERE fs.species_id = t.species_id
                             AND CASE 
                                 WHEN ms.stage_name LIKE '%spiderling%' THEN 'Spiderling'
                                 WHEN ms.stage_name LIKE '%juvenile%' THEN 'Juvenile'
                                 WHEN ms.stage_name LIKE '%sub%adult%' THEN 'Sub-Adult'
                                 WHEN ms.stage_name LIKE '%adult%' THEN 'Adult'
                             END = fs.size_category
                             LIMIT 1)
                        WHEN t.estimated_age_months IS NOT NULL THEN
                            CASE 
                                WHEN t.estimated_age_months < 6 THEN ts.adult_size_cm * 0.2
                                WHEN t.estimated_age_months < 12 THEN ts.adult_size_cm * 0.4
                                WHEN t.estimated_age_months < 24 THEN ts.adult_size_cm * 0.6
                                ELSE ts.adult_size_cm * 0.8
                            END
                        ELSE 
                            CASE 
                                WHEN EXTRACT(EPOCH FROM CURRENT_DATE::timestamp - t.acquisition_date::timestamp)/86400 > 730 
                                    THEN ts.adult_size_cm
                                ELSE ts.adult_size_cm * 0.5
                            END
                    END
                ) as current_size_cm
            FROM spider_bot.tarantulas t
            JOIN spider_bot.tarantula_species ts ON t.species_id = ts.id
            LEFT JOIN spider_bot.molt_stages ms ON t.current_molt_stage_id = ms.id
            LEFT JOIN LastMoltSize lms ON t.id = lms.tarantula_id
            WHERE t.id = ?
        )
        SELECT current_size_cm FROM EstimatedSize`, tarantulaID).
		Scan(&size)

	if result.Error != nil {
		return 0, fmt.Errorf("failed to get current size: %w", result.Error)
	}

	return size, nil
}

func (db *TarantulaDB) GetUserSettings(ctx context.Context, userID int64) (*models.UserSettings, error) {
	var settings models.UserSettings
	result := db.db.WithContext(ctx).Where("user_id = ?", userID).First(&settings)
	if result.Error != nil {
		if result.Error == gorm.ErrRecordNotFound {
			// Create default settings if none exist
			return db.createDefaultSettings(ctx, userID)
		}
		return nil, fmt.Errorf("failed to get user settings: %w", result.Error)
	}
	return &settings, nil
}

func (db *TarantulaDB) UpdateUserSettings(ctx context.Context, settings *models.UserSettings) error {
	result := db.db.WithContext(ctx).Save(settings)
	if result.Error != nil {
		return fmt.Errorf("failed to update user settings: %w", result.Error)
	}
	return nil
}

func (db *TarantulaDB) createDefaultSettings(ctx context.Context, userID int64) (*models.UserSettings, error) {
	settings := models.UserSettings{}.Defaults()
	settings.UserID = userID
	result := db.db.WithContext(ctx).Create(&settings)
	if result.Error != nil {
		return nil, fmt.Errorf("failed to create default settings: %w", result.Error)
	}
	return &settings, nil
}
func (db *TarantulaDB) GetActiveUsers(ctx context.Context) ([]models.TelegramUser, error) {
	var users []models.TelegramUser

	thirtyDaysAgo := time.Now().AddDate(0, 0, -30)

	result := db.db.WithContext(ctx).
		Where("is_active = ? AND chat_id IS NOT NULL AND last_active > ?", true, thirtyDaysAgo).
		Find(&users)

	if result.Error != nil {
		return nil, fmt.Errorf("failed to get active users: %w", result.Error)
	}

	return users, nil
}

func (db *TarantulaDB) EnsureUserExists(ctx context.Context, user *models.TelegramUser) error {
	result := db.db.WithContext(ctx).
		Where(models.TelegramUser{TelegramID: user.TelegramID}).
		Assign(models.TelegramUser{
			Username:   user.Username,
			FirstName:  user.FirstName,
			LastName:   user.LastName,
			ChatID:     user.ChatID,
			IsActive:   true,
			LastActive: time.Now(),
		}).
		FirstOrCreate(user)

	if result.Error != nil {
		return fmt.Errorf("failed to ensure user exists: %w", result.Error)
	}

	return nil
}
