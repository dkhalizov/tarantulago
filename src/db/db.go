package db

import (
	"context"
	"fmt"
	"strings"
	"tarantulago/models"
	"time"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/schema"
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
	db = db.Set("gorm:table_options", "")
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
		&models.WeightRecord{},
		&models.TarantulaPhoto{},
		&models.TarantulaColony{},
		&models.TarantulaColonyMember{},
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

func (db *TarantulaDB) GetAllSpecies(ctx context.Context) ([]models.TarantulaSpecies, error) {
	var species []models.TarantulaSpecies
	result := db.db.WithContext(ctx).Order("common_name ASC").Find(&species)
	if result.Error != nil {
		return nil, fmt.Errorf("failed to get species: %w", result.Error)
	}
	return species, nil
}

func (db *TarantulaDB) RecordFeeding(ctx context.Context, event models.FeedingEvent) (int64, error) {
	var id int64
	err := db.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {

		// Validate either individual tarantula or colony (but not both)
		if event.TarantulaColonyID != nil && *event.TarantulaColonyID > 0 {
			// Colony feeding - validate colony exists
			var tarantulaColony models.TarantulaColony
			if err := tx.Where("id = ? AND user_id = ?", *event.TarantulaColonyID, event.UserID).First(&tarantulaColony).Error; err != nil {
				return fmt.Errorf("tarantula colony not found or access denied: %w", err)
			}
		} else if event.TarantulaID != nil && *event.TarantulaID > 0 {
			// Individual feeding - validate tarantula exists
			var tarantula models.Tarantula
			if err := tx.Where("id = ? AND user_id = ?", *event.TarantulaID, event.UserID).First(&tarantula).Error; err != nil {
				return fmt.Errorf("tarantula not found or access denied: %w", err)
			}
		} else {
			return fmt.Errorf("feeding event must specify either a tarantula or a colony")
		}

		var colony models.CricketColony
		if err := tx.Where("user_id = ?", event.UserID).First(&colony).Error; err != nil {
			return fmt.Errorf("no cricket colony found for user: %w", err)
		}

		if colony.CurrentCount < event.NumberOfCrickets {
			colony.CurrentCount = 100
			if err := tx.Model(&colony).Update("current_count", colony.CurrentCount).Error; err != nil {
				return fmt.Errorf("failed to auto-refill colony: %w", err)
			}
		}

		result := tx.Model(&colony).
			UpdateColumn("current_count", gorm.Expr("current_count - ?", event.NumberOfCrickets))

		if result.Error != nil {
			return fmt.Errorf("failed to update colony count: %w", result.Error)
		}

		feedingEvent := models.FeedingEvent{
			TarantulaID:       event.TarantulaID,
			TarantulaColonyID: event.TarantulaColonyID,
			FeedingDate:       time.Now(),
			CricketColonyID:   colony.ID,
			NumberOfCrickets:  event.NumberOfCrickets,
			FeedingStatusID:   int(models.FeedingStatusAccepted),
			Notes:             event.Notes,
			UserID:            event.UserID,
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
    -- Get individual feedings
    SELECT
        tarantula_id,
        MAX(feeding_date) as last_feeding_date,
        EXTRACT(DAY FROM (CURRENT_DATE - MAX(feeding_date))) as days_since_feeding
    FROM spider_bot.feeding_events
    WHERE tarantula_id IS NOT NULL
    GROUP BY tarantula_id

    UNION ALL

    -- Get colony feedings for tarantulas that are members
    SELECT
        tcm.tarantula_id,
        MAX(fe.feeding_date) as last_feeding_date,
        EXTRACT(DAY FROM (CURRENT_DATE - MAX(fe.feeding_date))) as days_since_feeding
    FROM spider_bot.feeding_events fe
    INNER JOIN spider_bot.tarantula_colony_members tcm
        ON fe.tarantula_colony_id = tcm.colony_id
    WHERE fe.tarantula_colony_id IS NOT NULL
      AND tcm.is_active = true
      AND (tcm.left_date IS NULL OR fe.feeding_date <= tcm.left_date)
      AND fe.feeding_date >= tcm.joined_date
    GROUP BY tcm.tarantula_id
),
     CombinedFeeding AS (
    SELECT
        tarantula_id,
        MAX(last_feeding_date) as last_feeding_date,
        MIN(days_since_feeding) as days_since_feeding
    FROM LastFeeding
    GROUP BY tarantula_id
),
     SizeBoundaries AS (
         SELECT
             size_category,
             MIN(body_length_cm) as min_size,
             MAX(body_length_cm) as max_size,
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
             COALESCE(
                     t.current_size,
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
SELECT
    t.id as id,
    t.name as name,
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
         LEFT JOIN CombinedFeeding lf ON t.id = lf.tarantula_id
         LEFT JOIN MatchingSchedule ms ON t.id = ms.tarantula_id
WHERE t.user_id = ?
  AND (molt.stage_name IS NULL OR molt.stage_name != 'Pre-molt')
  AND (t.post_molt_mute_until IS NULL OR t.post_molt_mute_until < CURRENT_TIMESTAMP)
  AND (
    lf.days_since_feeding IS NULL
        OR lf.days_since_feeding > ms.min_days
    )
ORDER BY days_since_feeding DESC;`, userID).
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
		Preload("TarantulaColony").
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
        SELECT DISTINCT
            t.id,
            t.name,
            t.species_id,
            ts.common_name as species_name,
            COALESCE(last_feeding.days_since_feeding, 999) as days_since_feeding,
            CASE
                WHEN ms.stage_name IN ('Pre-molt', 'Molting', 'Post-molt') THEN ms.stage_name
                WHEN hs.status_name = 'Critical' THEN 'Critical'
                WHEN COALESCE(last_feeding.days_since_feeding, 999) > COALESCE(best_schedule.max_days, 14) THEN 'Needs feeding'
                ELSE 'Normal'
            END as current_status,
            COALESCE(best_schedule.frequency_id, 1) as frequency_id,
            COALESCE(best_schedule.min_days, 7) as min_days,
            COALESCE(best_schedule.max_days, 14) as max_days
        FROM spider_bot.tarantulas t
        JOIN spider_bot.tarantula_species ts ON t.species_id = ts.id
        LEFT JOIN spider_bot.molt_stages ms ON t.current_molt_stage_id = ms.id
        LEFT JOIN spider_bot.health_statuses hs ON t.current_health_status_id = hs.id
        LEFT JOIN (
            SELECT
                t_sub.id as tarantula_id,
                EXTRACT(EPOCH FROM CURRENT_DATE::timestamp - MAX(fe.feeding_date)::timestamp)/86400 as days_since_feeding
            FROM spider_bot.tarantulas t_sub
            LEFT JOIN spider_bot.feeding_events fe ON
                fe.tarantula_id = t_sub.id OR
                (fe.tarantula_colony_id = t_sub.colony_id AND t_sub.colony_id IS NOT NULL)
            WHERE fe.id IS NOT NULL
            GROUP BY t_sub.id
        ) last_feeding ON t.id = last_feeding.tarantula_id
        LEFT JOIN (
            SELECT DISTINCT ON (fs.species_id)
                fs.species_id,
                fs.frequency_id,
                ff.min_days,
                ff.max_days
            FROM spider_bot.feeding_schedules fs
            JOIN spider_bot.feeding_frequencies ff ON fs.frequency_id = ff.id
            ORDER BY fs.species_id, fs.body_length_cm DESC
        ) best_schedule ON t.species_id = best_schedule.species_id
        WHERE t.user_id = $1
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
		// Get user settings to determine post-molt mute duration
		var settings models.UserSettings
		if err := tx.Where("user_id = ?", molt.UserID).First(&settings).Error; err != nil {
			// If settings not found, use default
			settings = models.UserSettings{PostMoltMuteDays: 7}
		}

		// Calculate post-molt mute period
		muteUntil := time.Now().AddDate(0, 0, settings.PostMoltMuteDays)

		result := tx.Model(&models.Tarantula{}).
			Where("id = ? AND user_id = ?", molt.TarantulaID, molt.UserID).
			Updates(map[string]interface{}{
				"last_molt_date":        time.Now(),
				"current_molt_stage_id": models.MoltStagePostMolt,
				"post_molt_mute_until":  muteUntil,
			})

		if result.Error != nil {
			return fmt.Errorf("failed to update tarantula molt status: %w", result.Error)
		}

		if result.RowsAffected == 0 {
			return fmt.Errorf("tarantula not found or access denied")
		}

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

	result := db.db.WithContext(ctx).
		Where("is_active = ? AND chat_id IS NOT NULL", true).
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

func (db *TarantulaDB) RecordColonyMaintenance(ctx context.Context, record models.ColonyMaintenanceRecord) (int64, error) {
	result := db.db.WithContext(ctx).Create(&record)
	if result.Error != nil {
		return 0, fmt.Errorf("failed to create colony maintenance record: %w", result.Error)
	}

	err := db.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var schedule models.ColonyMaintenanceSchedule
		err := tx.Where("colony_id = ? AND maintenance_type_id = ? AND user_id = ?",
			record.ColonyID, record.MaintenanceTypeID, record.UserID).First(&schedule).Error

		if err != nil {
			if err == gorm.ErrRecordNotFound {
				var maintenanceType models.ColonyMaintenanceType
				if err := tx.First(&maintenanceType, record.MaintenanceTypeID).Error; err != nil {
					return fmt.Errorf("failed to get maintenance type: %w", err)
				}

				schedule = models.ColonyMaintenanceSchedule{
					ColonyID:          record.ColonyID,
					MaintenanceTypeID: record.MaintenanceTypeID,
					FrequencyDays:     maintenanceType.FrequencyDays,
					Enabled:           true,
					LastPerformedDate: &record.MaintenanceDate,
					UserID:            record.UserID,
				}

				return tx.Create(&schedule).Error
			}
			return err
		}

		schedule.LastPerformedDate = &record.MaintenanceDate
		return tx.Save(&schedule).Error
	})

	if err != nil {
		return int64(record.ID), fmt.Errorf("failed to update maintenance schedule: %w", err)
	}

	return int64(record.ID), nil
}

func (db *TarantulaDB) GetColonyMaintenanceAlerts(ctx context.Context, userID int64) ([]models.ColonyMaintenanceAlert, error) {
	var alerts []models.ColonyMaintenanceAlert

	result := db.db.WithContext(ctx).Raw(`
        WITH LastMaintenance AS (
            SELECT
                cms.colony_id,
                cms.maintenance_type_id,
                cms.frequency_days,
                cms.user_id,
                cms.last_performed_date,
                cmt.type_name as maintenance_type,
                cc.colony_name,
                CASE
                    WHEN cms.last_performed_date IS NOT NULL THEN
                        (CURRENT_DATE - cms.last_performed_date)
                    ELSE
                        (CURRENT_DATE - cms.created_at::date)
                END as days_since_last_done
            FROM spider_bot.colony_maintenance_schedules cms
                JOIN spider_bot.colony_maintenance_types cmt ON cms.maintenance_type_id = cmt.id
                JOIN spider_bot.cricket_colonies cc ON cms.colony_id = cc.id
            WHERE cms.enabled = TRUE AND cms.user_id = ?
        )
        SELECT
            lm.colony_id as id,
            lm.colony_name,
            lm.maintenance_type,
            lm.days_since_last_done::INTEGER as days_since_last_done,
            (lm.days_since_last_done - lm.frequency_days)::INTEGER as days_overdue
        FROM LastMaintenance lm
        WHERE lm.days_since_last_done >= lm.frequency_days
        ORDER BY days_overdue DESC`, userID).
		Scan(&alerts)

	if result.Error != nil {
		return nil, fmt.Errorf("failed to get colony maintenance alerts: %w", result.Error)
	}

	return alerts, nil
}

func (db *TarantulaDB) GetColonyMaintenanceHistory(ctx context.Context, colonyID int64, userID int64, limit int32) ([]models.ColonyMaintenanceRecord, error) {
	var records []models.ColonyMaintenanceRecord

	result := db.db.WithContext(ctx).
		Preload("MaintenanceType").
		Preload("Colony").
		Where("colony_id = ? AND user_id = ?", colonyID, userID).
		Order("maintenance_date DESC").
		Limit(int(limit)).
		Find(&records)

	if result.Error != nil {
		return nil, fmt.Errorf("failed to get colony maintenance history: %w", result.Error)
	}

	return records, nil
}

func (db *TarantulaDB) GetMaintenanceTypes(ctx context.Context) ([]models.ColonyMaintenanceType, error) {
	var types []models.ColonyMaintenanceType

	result := db.db.WithContext(ctx).Find(&types)
	if result.Error != nil {
		return nil, fmt.Errorf("failed to get maintenance types: %w", result.Error)
	}

	return types, nil
}

func (db *TarantulaDB) RecordWeight(ctx context.Context, weight models.WeightRecord) (int64, error) {
	weight.WeighDate = time.Now()
	weight.CreatedAt = time.Now()

	err := db.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {

		if err := tx.Create(&weight).Error; err != nil {
			return fmt.Errorf("failed to create weight record: %w", err)
		}

		updateTime := time.Now()
		if err := tx.Model(&models.Tarantula{}).
			Where("id = ? AND user_id = ?", weight.TarantulaID, weight.UserID).
			Updates(map[string]interface{}{
				"current_weight_grams": weight.WeightGrams,
				"last_weigh_date":      &updateTime,
			}).Error; err != nil {
			return fmt.Errorf("failed to update tarantula weight: %w", err)
		}

		return nil
	})

	if err != nil {
		return 0, err
	}

	return int64(weight.ID), nil
}

func (db *TarantulaDB) GetWeightHistory(ctx context.Context, tarantulaID int32, userID int64, limit int32) ([]models.WeightRecord, error) {
	var weights []models.WeightRecord

	result := db.db.WithContext(ctx).
		Where("tarantula_id = ? AND user_id = ?", tarantulaID, userID).
		Order("weigh_date DESC").
		Limit(int(limit)).
		Find(&weights)

	if result.Error != nil {
		return nil, fmt.Errorf("failed to get weight history: %w", result.Error)
	}

	return weights, nil
}

func (db *TarantulaDB) GetLatestWeight(ctx context.Context, tarantulaID int32, userID int64) (*models.WeightRecord, error) {
	var weight models.WeightRecord

	result := db.db.WithContext(ctx).
		Where("tarantula_id = ? AND user_id = ?", tarantulaID, userID).
		Order("weigh_date DESC").
		First(&weight)

	if result.Error != nil {
		if result.Error == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get latest weight: %w", result.Error)
	}

	return &weight, nil
}

func (db *TarantulaDB) AddPhoto(ctx context.Context, photo models.TarantulaPhoto) (int64, error) {
	photo.TakenDate = time.Now()
	photo.CreatedAt = time.Now()

	result := db.db.WithContext(ctx).Create(&photo)
	if result.Error != nil {
		return 0, fmt.Errorf("failed to add photo: %w", result.Error)
	}

	return int64(photo.ID), nil
}

func (db *TarantulaDB) GetTarantulaPhotos(ctx context.Context, tarantulaID int32, userID int64, limit int32) ([]models.TarantulaPhoto, error) {
	var photos []models.TarantulaPhoto

	result := db.db.WithContext(ctx).
		Where("tarantula_id = ? AND user_id = ?", tarantulaID, userID).
		Order("taken_date DESC").
		Limit(int(limit)).
		Find(&photos)

	if result.Error != nil {
		return nil, fmt.Errorf("failed to get photos: %w", result.Error)
	}

	return photos, nil
}

func (db *TarantulaDB) UpdateTarantulaProfilePhoto(ctx context.Context, tarantulaID int32, photoURL string, userID int64) error {
	result := db.db.WithContext(ctx).
		Model(&models.Tarantula{}).
		Where("id = ? AND user_id = ?", tarantulaID, userID).
		Update("profile_photo_url", photoURL)

	if result.Error != nil {
		return fmt.Errorf("failed to update profile photo: %w", result.Error)
	}

	if result.RowsAffected == 0 {
		return fmt.Errorf("tarantula not found or access denied")
	}

	return nil
}

func (db *TarantulaDB) QuickFeed(ctx context.Context, tarantulaID int32, userID int64) error {
	return db.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {

		var colony models.CricketColony
		if err := tx.Where("user_id = ?", userID).First(&colony).Error; err != nil {
			return fmt.Errorf("no cricket colony found: %w", err)
		}

		if colony.CurrentCount < 1 {
			return fmt.Errorf("no crickets available in colony")
		}

		tid := int(tarantulaID)
		feedingEvent := models.FeedingEvent{
			TarantulaID:      &tid,
			FeedingDate:      time.Now(),
			CricketColonyID:  colony.ID,
			NumberOfCrickets: 1,
			FeedingStatusID:  int(models.FeedingStatusAccepted),
			Notes:            "Quick feed",
			UserID:           userID,
		}

		if err := tx.Create(&feedingEvent).Error; err != nil {
			return fmt.Errorf("failed to create feeding event: %w", err)
		}

		if err := tx.Model(&colony).
			UpdateColumn("current_count", gorm.Expr("current_count - 1")).Error; err != nil {
			return fmt.Errorf("failed to update colony count: %w", err)
		}

		return nil
	})
}

func (db *TarantulaDB) QuickFeedColony(ctx context.Context, tarantulaColonyID int32, userID int64) error {
	return db.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {

		// Verify colony exists and belongs to user
		var tarantulaColony models.TarantulaColony
		if err := tx.Where("id = ? AND user_id = ?", tarantulaColonyID, userID).First(&tarantulaColony).Error; err != nil {
			return fmt.Errorf("tarantula colony not found: %w", err)
		}

		var cricketColony models.CricketColony
		if err := tx.Where("user_id = ?", userID).First(&cricketColony).Error; err != nil {
			return fmt.Errorf("no cricket colony found: %w", err)
		}

		if cricketColony.CurrentCount < 1 {
			return fmt.Errorf("no crickets available in colony")
		}

		cid := int(tarantulaColonyID)
		feedingEvent := models.FeedingEvent{
			TarantulaColonyID: &cid,
			FeedingDate:       time.Now(),
			CricketColonyID:   cricketColony.ID,
			NumberOfCrickets:  1,
			FeedingStatusID:   int(models.FeedingStatusAccepted),
			Notes:             "Quick feed - colony",
			UserID:            userID,
		}

		if err := tx.Create(&feedingEvent).Error; err != nil {
			return fmt.Errorf("failed to create feeding event: %w", err)
		}

		if err := tx.Model(&cricketColony).
			UpdateColumn("current_count", gorm.Expr("current_count - 1")).Error; err != nil {
			return fmt.Errorf("failed to update cricket colony count: %w", err)
		}

		return nil
	})
}

func (db *TarantulaDB) GetFeedingPatterns(ctx context.Context, userID int64) ([]models.FeedingPattern, error) {
	var patterns []models.FeedingPattern

	query := `
    SELECT 
        t.id as tarantula_id,
        t.name as tarantula_name,
        COALESCE(COUNT(fe.id), 0) as total_feedings,
        COALESCE(
            CASE WHEN COUNT(fe.id) > 1 THEN
                EXTRACT(DAY FROM (MAX(fe.feeding_date) - MIN(fe.feeding_date))) / NULLIF(COUNT(fe.id) - 1, 0)
            ELSE 0 END, 0
        ) as average_interval,
        COALESCE(
            COUNT(CASE WHEN fs.status_name = 'Accepted' THEN 1 END) * 100.0 / NULLIF(COUNT(fe.id), 0), 0
        ) as acceptance_rate,
        MAX(fe.feeding_date) as last_feeding_date,
        COALESCE(EXTRACT(DAY FROM (NOW() - MAX(fe.feeding_date))), 999) as days_since_last_feeding,
        COALESCE(
            COUNT(fe.id) * 7.0 / NULLIF(EXTRACT(DAY FROM (MAX(fe.feeding_date) - MIN(fe.feeding_date))), 0), 0
        ) as crickets_per_week,
        CASE 
            WHEN COUNT(fe.id) < 3 THEN 'Insufficient Data'
            WHEN EXTRACT(DAY FROM (MAX(fe.feeding_date) - MIN(fe.feeding_date))) / NULLIF(COUNT(fe.id) - 1, 0) BETWEEN 7 AND 14 THEN 'Regular'
            WHEN EXTRACT(DAY FROM (MAX(fe.feeding_date) - MIN(fe.feeding_date))) / NULLIF(COUNT(fe.id) - 1, 0) BETWEEN 5 AND 21 THEN 'Irregular' 
            ELSE 'Inconsistent'
        END as feeding_regularity
    FROM spider_bot.tarantulas t
    LEFT JOIN spider_bot.feeding_events fe ON t.id = fe.tarantula_id
    LEFT JOIN spider_bot.feeding_statuses fs ON fe.feeding_status_id = fs.id
    WHERE t.user_id = ?
    GROUP BY t.id, t.name
    ORDER BY t.name`

	if err := db.db.WithContext(ctx).Raw(query, userID).Scan(&patterns).Error; err != nil {
		return nil, fmt.Errorf("failed to get feeding patterns: %w", err)
	}

	return patterns, nil
}

func (db *TarantulaDB) GetGrowthData(ctx context.Context, userID int64) ([]models.GrowthData, error) {

	var basicData []struct {
		TarantulaID   int32    `json:"tarantula_id"`
		TarantulaName string   `json:"tarantula_name"`
		CurrentWeight *float64 `json:"current_weight_grams"`
		CurrentSize   float64  `json:"current_size"`
	}

	query := `
    SELECT 
        t.id as tarantula_id,
        t.name as tarantula_name,
        t.current_weight_grams,
        t.current_size
    FROM spider_bot.tarantulas t
    WHERE t.user_id = ?`

	if err := db.db.WithContext(ctx).Raw(query, userID).Scan(&basicData).Error; err != nil {
		return nil, fmt.Errorf("failed to get growth data: %w", err)
	}

	growthData := make([]models.GrowthData, len(basicData))
	for i, data := range basicData {
		growthData[i] = models.GrowthData{
			TarantulaID:   data.TarantulaID,
			TarantulaName: data.TarantulaName,
			CurrentWeight: data.CurrentWeight,
			CurrentSize:   data.CurrentSize,
		}

		var weightRecords []struct {
			Date   time.Time `json:"date"`
			Weight float64   `json:"weight"`
		}

		weightQuery := `
        SELECT weigh_date as date, weight_grams as weight 
        FROM spider_bot.weight_records 
        WHERE tarantula_id = ? AND user_id = ? 
        ORDER BY weigh_date`

		if err := db.db.WithContext(ctx).Raw(weightQuery, data.TarantulaID, userID).Scan(&weightRecords).Error; err != nil {
			return nil, fmt.Errorf("failed to get weight history: %w", err)
		}

		weightPoints := make([]models.WeightPoint, len(weightRecords))
		for j, record := range weightRecords {
			weightPoints[j] = models.WeightPoint{
				Date:   record.Date,
				Weight: record.Weight,
			}
		}
		growthData[i].WeightHistory = weightPoints

		var sizeRecords []struct {
			Date time.Time `json:"date"`
			Size float64   `json:"size"`
		}

		sizeQuery := `
        SELECT molt_date as date, post_molt_length_cm as size 
        FROM spider_bot.molt_records 
        WHERE tarantula_id = ? AND user_id = ? AND post_molt_length_cm > 0
        ORDER BY molt_date`

		if err := db.db.WithContext(ctx).Raw(sizeQuery, data.TarantulaID, userID).Scan(&sizeRecords).Error; err != nil {
			return nil, fmt.Errorf("failed to get size history: %w", err)
		}

		sizePoints := make([]models.SizePoint, len(sizeRecords))
		for j, record := range sizeRecords {
			sizePoints[j] = models.SizePoint{
				Date: record.Date,
				Size: record.Size,
			}
		}
		growthData[i].SizeHistory = sizePoints

		if len(weightPoints) > 1 {
			firstWeight := weightPoints[0].Weight
			lastWeight := weightPoints[len(weightPoints)-1].Weight
			days := weightPoints[len(weightPoints)-1].Date.Sub(weightPoints[0].Date).Hours() / 24
			if days > 0 {
				monthlyRate := (lastWeight - firstWeight) / (days / 30)
				growthData[i].GrowthRate = &monthlyRate
			}
			growthData[i].WeightChangeTotal = lastWeight - firstWeight
		}

		if len(sizePoints) > 1 {
			growthData[i].SizeChangeTotal = sizePoints[len(sizePoints)-1].Size - sizePoints[0].Size
		}
	}

	return growthData, nil
}

func (db *TarantulaDB) GenerateAnnualReport(ctx context.Context, userID int64, year int) ([]models.AnnualReport, error) {

	type AnnualReportTemp struct {
		Year            int      `json:"year"`
		TarantulaID     int32    `json:"tarantula_id"`
		TarantulaName   string   `json:"tarantula_name"`
		TotalFeedings   int32    `json:"total_feedings"`
		TotalCrickets   int32    `json:"total_crickets"`
		AcceptanceRate  float64  `json:"acceptance_rate"`
		AverageInterval float64  `json:"average_interval"`
		EndWeight       *float64 `json:"end_weight"`
		EndSize         float64  `json:"end_size"`
		MoltCount       int32    `json:"molt_count"`
		PhotosAdded     int32    `json:"photos_added"`
		HealthIssues    int32    `json:"health_issues"`
		EstimatedCost   float64  `json:"estimated_cost"`
	}

	var tempReports []AnnualReportTemp

	query := `
    SELECT 
        $1 as year,
        t.id as tarantula_id,
        t.name as tarantula_name,
        -- Feeding statistics
        COALESCE(COUNT(DISTINCT fe.id), 0) as total_feedings,
        COALESCE(SUM(fe.number_of_crickets), 0) as total_crickets,
        COALESCE(
            COUNT(CASE WHEN fs.status_name = 'Accepted' THEN 1 END) * 100.0 / NULLIF(COUNT(fe.id), 0), 0
        ) as acceptance_rate,
        COALESCE(
            CASE WHEN COUNT(fe.id) > 1 THEN
                EXTRACT(DAY FROM (MAX(fe.feeding_date) - MIN(fe.feeding_date))) / NULLIF(COUNT(fe.id) - 1, 0)
            ELSE 0 END, 0
        ) as average_interval,
        
        -- Growth statistics (will be calculated separately for start/end weights)
        t.current_weight_grams as end_weight,
        t.current_size as end_size,
        
        -- Events count
        COALESCE(COUNT(DISTINCT mr.id), 0) as molt_count,
        COALESCE(COUNT(DISTINCT tp.id), 0) as photos_added,
        COALESCE(COUNT(CASE WHEN hs.status_name != 'Healthy' THEN 1 END), 0) as health_issues,
        COALESCE(SUM(fe.number_of_crickets), 0) * 0.10 as estimated_cost
        
    FROM spider_bot.tarantulas t
    LEFT JOIN spider_bot.feeding_events fe ON t.id = fe.tarantula_id 
        AND EXTRACT(YEAR FROM fe.feeding_date) = $1
    LEFT JOIN spider_bot.feeding_statuses fs ON fe.feeding_status_id = fs.id
    LEFT JOIN spider_bot.molt_records mr ON t.id = mr.tarantula_id 
        AND EXTRACT(YEAR FROM mr.molt_date) = $1
    LEFT JOIN spider_bot.tarantula_photos tp ON t.id = tp.tarantula_id 
        AND EXTRACT(YEAR FROM tp.taken_date) = $1
    LEFT JOIN spider_bot.health_check_records hcr ON t.id = hcr.tarantula_id 
        AND EXTRACT(YEAR FROM hcr.check_date) = $1
    LEFT JOIN spider_bot.health_statuses hs ON hcr.health_status_id = hs.id
    WHERE t.user_id = $2
    GROUP BY t.id, t.name, t.current_weight_grams, t.current_size
    ORDER BY t.name`

	if err := db.db.WithContext(ctx).Raw(query, year, userID).Scan(&tempReports).Error; err != nil {
		return nil, fmt.Errorf("failed to generate annual report: %w", err)
	}

	reports := make([]models.AnnualReport, len(tempReports))
	for i, temp := range tempReports {
		reports[i] = models.AnnualReport{
			Year:            temp.Year,
			TarantulaID:     temp.TarantulaID,
			TarantulaName:   temp.TarantulaName,
			TotalFeedings:   temp.TotalFeedings,
			TotalCrickets:   temp.TotalCrickets,
			AcceptanceRate:  temp.AcceptanceRate,
			AverageInterval: temp.AverageInterval,
			EndWeight:       temp.EndWeight,
			EndSize:         temp.EndSize,
			MoltCount:       temp.MoltCount,
			PhotosAdded:     temp.PhotosAdded,
			HealthIssues:    temp.HealthIssues,
			EstimatedCost:   temp.EstimatedCost,
			Milestones:      []string{},
		}
	}

	for i := range reports {

		var startWeight *float64
		startWeightQuery := `
        SELECT weight_grams FROM spider_bot.weight_records 
        WHERE tarantula_id = $1 AND EXTRACT(YEAR FROM weigh_date) = $2 
        ORDER BY weigh_date LIMIT 1`

		var weight float64
		if err := db.db.WithContext(ctx).Raw(startWeightQuery, reports[i].TarantulaID, year).Scan(&weight).Error; err == nil {
			startWeight = &weight
		}
		reports[i].StartWeight = startWeight

		if startWeight != nil && reports[i].EndWeight != nil {
			gain := *reports[i].EndWeight - *startWeight
			reports[i].WeightGain = &gain
		}

		var startSize *float64
		startSizeQuery := `
        SELECT pre_molt_length_cm FROM spider_bot.molt_records 
        WHERE tarantula_id = $1 AND EXTRACT(YEAR FROM molt_date) = $2
        ORDER BY molt_date LIMIT 1`

		if err := db.db.WithContext(ctx).Raw(startSizeQuery, reports[i].TarantulaID, year).Scan(&startSize).Error; err == nil && startSize != nil {
			reports[i].StartSize = *startSize
		}

		reports[i].SizeGrowth = reports[i].EndSize - reports[i].StartSize

		var milestones []string
		if reports[i].MoltCount > 0 {
			milestones = append(milestones, fmt.Sprintf("Molted %d times", reports[i].MoltCount))
		}
		if reports[i].WeightGain != nil && *reports[i].WeightGain > 0 {
			milestones = append(milestones, fmt.Sprintf("Gained %.1fg", *reports[i].WeightGain))
		}
		if reports[i].SizeGrowth > 0 {
			milestones = append(milestones, fmt.Sprintf("Grew %.1fcm", reports[i].SizeGrowth))
		}
		if reports[i].TotalFeedings > 52 {
			milestones = append(milestones, "Fed weekly all year")
		}
		if reports[i].PhotosAdded > 10 {
			milestones = append(milestones, "Well documented with photos")
		}
		reports[i].Milestones = milestones
	}

	return reports, nil
}

func (db *TarantulaDB) GetMoltPredictions(ctx context.Context, userID int64) ([]models.MoltPrediction, error) {

	var queryResults []struct {
		TarantulaID        int32      `json:"tarantula_id"`
		TarantulaName      string     `json:"tarantula_name"`
		LastMoltDate       *time.Time `json:"last_molt_date"`
		DaysSinceLastMolt  int32      `json:"days_since_last_molt"`
		MoltCount          int32      `json:"molt_count"`
		CurrentSize        float64    `json:"current_size"`
		AdultSizeCM        *float64   `json:"adult_size_cm"`
		EstimatedAgeMonths int32      `json:"estimated_age_months"`
		SpeciesTemperament string     `json:"species_temperament"`
		SpeciesName        string     `json:"species_name"`
		RecentFeedings     int32      `json:"recent_feedings"`
		RecentRejections   int32      `json:"recent_rejections"`
		AverageCycle       *float64   `json:"average_cycle"`
		LastFeedingDate    *time.Time `json:"last_feeding_date"`
		FeedingFreqMinDays *int32     `json:"feeding_freq_min_days"`
		FeedingFreqMaxDays *int32     `json:"feeding_freq_max_days"`
	}

	query := `
    WITH molt_stats AS (
        SELECT 
            t.id as tarantula_id,
            t.name as tarantula_name,
            t.current_size,
            t.estimated_age_months,
            ts.adult_size_cm,
            ts.temperament as species_temperament,
            ts.scientific_name as species_name,
            
            -- Use the actual most recent molt date from molt_records
            MAX(mr.molt_date) as last_molt_date,
            COALESCE(EXTRACT(DAY FROM (NOW() - MAX(mr.molt_date))), 999) as days_since_last_molt,
            COUNT(mr.id) as molt_count,
            
            -- Calculate proper average cycle from consecutive molt records
            CASE WHEN COUNT(mr.id) > 1 THEN
                EXTRACT(DAY FROM (MAX(mr.molt_date) - MIN(mr.molt_date))) / NULLIF(COUNT(mr.id) - 1, 0)
            ELSE NULL END as average_cycle
            
        FROM spider_bot.tarantulas t
        LEFT JOIN spider_bot.molt_records mr ON t.id = mr.tarantula_id
        LEFT JOIN spider_bot.tarantula_species ts ON t.species_id = ts.id
        WHERE t.user_id = $1
        GROUP BY t.id, t.name, t.current_size, t.estimated_age_months, ts.adult_size_cm, ts.temperament, ts.scientific_name
    ),
    feeding_behavior AS (
        SELECT 
            t.id as tarantula_id,
            COUNT(CASE WHEN fe.feeding_date > NOW() - INTERVAL '14 days' THEN 1 END) as recent_feedings,
            COUNT(CASE WHEN fe.feeding_date > NOW() - INTERVAL '30 days' AND fs.status_name = 'Rejected' THEN 1 END) as recent_rejections,
            MAX(fe.feeding_date) as last_feeding_date
        FROM spider_bot.tarantulas t
        LEFT JOIN spider_bot.feeding_events fe ON t.id = fe.tarantula_id
        LEFT JOIN spider_bot.feeding_statuses fs ON fe.feeding_status_id = fs.id
        WHERE t.user_id = $1
        GROUP BY t.id
    ),
    species_feeding_schedule AS (
        SELECT DISTINCT ON (fs.species_id)
            fs.species_id,
            ff.min_days,
            ff.max_days
        FROM spider_bot.feeding_schedules fs
        JOIN spider_bot.feeding_frequencies ff ON fs.frequency_id = ff.id
        JOIN spider_bot.tarantulas t ON fs.species_id = t.species_id AND t.user_id = $1
        WHERE fs.body_length_cm <= t.current_size
        ORDER BY fs.species_id, fs.body_length_cm DESC
    )
    SELECT 
        ms.*,
        COALESCE(fb.recent_feedings, 0) as recent_feedings,
        COALESCE(fb.recent_rejections, 0) as recent_rejections,
        fb.last_feeding_date,
        sfs.min_days as feeding_freq_min_days,
        sfs.max_days as feeding_freq_max_days
    FROM molt_stats ms
    LEFT JOIN feeding_behavior fb ON ms.tarantula_id = fb.tarantula_id
    LEFT JOIN species_feeding_schedule sfs ON ms.tarantula_id IN (
        SELECT t.id FROM spider_bot.tarantulas t WHERE t.species_id = sfs.species_id AND t.user_id = $1
    )
    ORDER BY ms.tarantula_name`

	if err := db.db.WithContext(ctx).Raw(query, userID).Scan(&queryResults).Error; err != nil {
		return nil, fmt.Errorf("failed to get molt predictions: %w", err)
	}

	predictions := make([]models.MoltPrediction, len(queryResults))

	for i, result := range queryResults {
		prediction := models.MoltPrediction{
			TarantulaID:       result.TarantulaID,
			TarantulaName:     result.TarantulaName,
			LastMoltDate:      result.LastMoltDate,
			DaysSinceLastMolt: result.DaysSinceLastMolt,
			AverageMoltCycle:  result.AverageCycle,
			MoltCount:         result.MoltCount,
		}

		var estimatedCycle int
		var confidenceAdjustment int

		if result.AverageCycle != nil && result.LastMoltDate != nil && result.MoltCount > 1 {
			estimatedCycle = int(*result.AverageCycle)
			confidenceAdjustment = 2 // High confidence boost
		} else if result.LastMoltDate != nil {
			// Step 2: Use species-specific intelligent estimates
			estimatedCycle = db.calculateSpeciesBasedMoltCycle(result)
			confidenceAdjustment = 0
		} else {
			// Step 3: Pure species-based estimate for no molt history
			estimatedCycle = db.calculateSpeciesBasedMoltCycle(result)
			confidenceAdjustment = -1 // Lower confidence
		}

		// Calculate prediction if we have a last molt date
		if result.LastMoltDate != nil && estimatedCycle > 0 {
			predictedDate := result.LastMoltDate.AddDate(0, 0, estimatedCycle)
			prediction.PredictedMoltDate = &predictedDate
			daysUntil := int32(time.Until(predictedDate).Hours() / 24)
			prediction.DaysUntilMolt = &daysUntil
		}

		// Enhanced confidence calculation
		prediction.ConfidenceLevel = db.calculateConfidenceLevel(result, confidenceAdjustment)

		// Set size indicator
		if result.AdultSizeCM != nil {
			ratio := result.CurrentSize / *result.AdultSizeCM
			if ratio < 0.3 {
				prediction.SizeIndicator = "Small"
			} else if ratio < 0.6 {
				prediction.SizeIndicator = "Medium"
			} else if ratio < 0.9 {
				prediction.SizeIndicator = "Large"
			} else {
				prediction.SizeIndicator = "Adult"
			}
		} else {
			prediction.SizeIndicator = "Unknown"
		}

		// Set feeding behavior
		if result.RecentRejections > 2 {
			prediction.FeedingBehavior = "Stopped"
		} else if result.RecentFeedings < 2 {
			prediction.FeedingBehavior = "Reduced"
		} else {
			prediction.FeedingBehavior = "Normal"
		}

		// Add pre-molt signs
		var preMoltSigns []string
		if prediction.FeedingBehavior == "Stopped" {
			preMoltSigns = append(preMoltSigns, "Refusing food")
		}
		if prediction.FeedingBehavior == "Reduced" {
			preMoltSigns = append(preMoltSigns, "Eating less")
		}
		prediction.PreMoltSigns = preMoltSigns

		// Build reasoning
		var reasoning strings.Builder
		if result.AverageCycle != nil && result.MoltCount > 1 {
			reasoning.WriteString(fmt.Sprintf("Based on %d previous molts with %.0f day average cycle. ",
				result.MoltCount, *result.AverageCycle))
		} else if result.LastMoltDate != nil {
			reasoning.WriteString("Based on species size estimates and last molt date. ")
		} else {
			reasoning.WriteString("No molt history available. ")
		}

		if prediction.SizeIndicator == "Adult" {
			reasoning.WriteString("Adult size reached, molts will be less frequent.")
		}
		prediction.PredictionBasis = reasoning.String()

		// Enhanced recommendations with feeding behavior integration
		prediction.Recommendation = db.generateMoltRecommendation(result, &prediction)

		predictions[i] = prediction
	}

	return predictions, nil
}

// GetUpcomingMoltPredictions returns molt predictions that are predicted within the specified number of days
// Only returns predictions with High or Medium confidence
func (db *TarantulaDB) GetUpcomingMoltPredictions(ctx context.Context, userID int64, withinDays int) ([]models.MoltPrediction, error) {
	allPredictions, err := db.GetMoltPredictions(ctx, userID)
	if err != nil {
		return nil, err
	}

	var upcomingPredictions []models.MoltPrediction
	for _, pred := range allPredictions {
		// Only include predictions with sufficient confidence
		if pred.ConfidenceLevel != "High" && pred.ConfidenceLevel != "Medium" {
			continue
		}

		// Only include if we have a prediction date and days until molt
		if pred.DaysUntilMolt == nil || pred.PredictedMoltDate == nil {
			continue
		}

		// Check if molt is predicted within the specified timeframe
		if *pred.DaysUntilMolt > 0 && *pred.DaysUntilMolt <= int32(withinDays) {
			upcomingPredictions = append(upcomingPredictions, pred)
		}
	}

	return upcomingPredictions, nil
}

// Simplified method aliases for backward compatibility
func (db *TarantulaDB) GetAllFeedingPatterns(ctx context.Context, userID int64) ([]models.FeedingPattern, error) {
	return db.GetFeedingPatterns(ctx, userID)
}

func (db *TarantulaDB) GetAllGrowthData(ctx context.Context, userID int64) ([]models.GrowthData, error) {
	return db.GetGrowthData(ctx, userID)
}

func (db *TarantulaDB) GetAllAnnualReports(ctx context.Context, year int, userID int64) ([]models.AnnualReport, error) {
	return db.GenerateAnnualReport(ctx, userID, year)
}

func (db *TarantulaDB) GetAllMoltPredictions(ctx context.Context, userID int64) ([]models.MoltPrediction, error) {
	return db.GetMoltPredictions(ctx, userID)
}

// Convenience method aliases for cleaner interface
func (db *TarantulaDB) GetFeedingHistory(ctx context.Context, userID int64, limit int32) ([]models.FeedingEvent, error) {
	return db.GetRecentFeedingRecords(ctx, userID, limit)
}

func (db *TarantulaDB) GetPhotos(ctx context.Context, tarantulaID int32, userID int64) ([]models.TarantulaPhoto, error) {
	return db.GetTarantulaPhotos(ctx, tarantulaID, userID, 10) // Default limit of 10
}

func (db *TarantulaDB) GetTarantulaWithSpeciesData(ctx context.Context, tarantulaID int32, userID int64) (*models.Tarantula, error) {
	var tarantula models.Tarantula

	result := db.db.WithContext(ctx).
		Preload("Species").
		Preload("CurrentMoltStage").
		Preload("CurrentHealthStatus").
		Where("id = ? AND user_id = ?", tarantulaID, userID).
		First(&tarantula)

	if result.Error != nil {
		if result.Error == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("tarantula not found")
		}
		return nil, fmt.Errorf("failed to get tarantula with species data: %w", result.Error)
	}

	return &tarantula, nil
}

// Enhanced species-based molt cycle calculation
func (db *TarantulaDB) calculateSpeciesBasedMoltCycle(result struct {
	TarantulaID        int32      `json:"tarantula_id"`
	TarantulaName      string     `json:"tarantula_name"`
	LastMoltDate       *time.Time `json:"last_molt_date"`
	DaysSinceLastMolt  int32      `json:"days_since_last_molt"`
	MoltCount          int32      `json:"molt_count"`
	CurrentSize        float64    `json:"current_size"`
	AdultSizeCM        *float64   `json:"adult_size_cm"`
	EstimatedAgeMonths int32      `json:"estimated_age_months"`
	SpeciesTemperament string     `json:"species_temperament"`
	SpeciesName        string     `json:"species_name"`
	RecentFeedings     int32      `json:"recent_feedings"`
	RecentRejections   int32      `json:"recent_rejections"`
	AverageCycle       *float64   `json:"average_cycle"`
	LastFeedingDate    *time.Time `json:"last_feeding_date"`
	FeedingFreqMinDays *int32     `json:"feeding_freq_min_days"`
	FeedingFreqMaxDays *int32     `json:"feeding_freq_max_days"`
}) int {
	// REALISTIC molt cycle estimates by size ratio to adult
	var baseCycle int
	if result.AdultSizeCM != nil && *result.AdultSizeCM > 0 {
		sizeRatio := result.CurrentSize / *result.AdultSizeCM

		// SPECIES-SPECIFIC molt intervals for HOME BREEDING
		switch {
		case sizeRatio < 0.2: // Spiderling (0-2cm)
			baseCycle = 60 // ~2 months (fast growth phase)
		case sizeRatio < 0.4: // Early juvenile (2-4cm for LP, 2-3cm for hamorii)
			baseCycle = 90 // ~3 months
		case sizeRatio < 0.6: // Juvenile (4-8cm for LP, 3-5cm for hamorii)
			baseCycle = 120 // ~4 months
		case sizeRatio < 0.8: // Sub-adult (8-12cm for LP, 5-8cm for hamorii)
			baseCycle = 180 // ~6 months
		case sizeRatio < 0.95: // Near adult (12-16cm for LP, 8-12cm for hamorii)
			baseCycle = 270 // ~9 months
		default: // Adult (16+ for LP, 12+ for hamorii)
			baseCycle = 450 // ~15 months (adults molt much less frequently)
		}
	} else {
		// Fallback based on estimated age (more realistic for home breeding)
		if result.EstimatedAgeMonths <= 6 {
			baseCycle = 60 // ~2 months for babies
		} else if result.EstimatedAgeMonths <= 12 {
			baseCycle = 90 // ~3 months for juveniles
		} else if result.EstimatedAgeMonths <= 24 {
			baseCycle = 150 // ~5 months for sub-adults
		} else if result.EstimatedAgeMonths <= 36 {
			baseCycle = 240 // ~8 months for near adults
		} else {
			baseCycle = 420 // ~14 months for adults
		}
	}

	// Adjust by species temperament (metabolism indicator)
	switch result.SpeciesTemperament {
	case "Fast", "Aggressive":
		baseCycle = int(float64(baseCycle) * 0.85) // 15% faster growth
	case "Slow", "Docile":
		baseCycle = int(float64(baseCycle) * 1.15) // 15% slower growth
	case "Moderate":
		// No adjustment for moderate temperament
	}

	// Adjust by feeding frequency (growth rate indicator)
	if result.FeedingFreqMinDays != nil {
		avgFeedDays := float64(*result.FeedingFreqMinDays)
		if result.FeedingFreqMaxDays != nil {
			avgFeedDays = (float64(*result.FeedingFreqMinDays) + float64(*result.FeedingFreqMaxDays)) / 2
		}

		// Species with faster feeding schedules molt more frequently
		if avgFeedDays <= 5 { // Fast feeders
			baseCycle = int(float64(baseCycle) * 0.9)
		} else if avgFeedDays >= 14 { // Slow feeders
			baseCycle = int(float64(baseCycle) * 1.1)
		}
	}

	// Minimum reasonable cycle time
	if baseCycle < 30 {
		baseCycle = 30
	}

	return baseCycle
}

// Enhanced confidence level calculation
func (db *TarantulaDB) calculateConfidenceLevel(result struct {
	TarantulaID        int32      `json:"tarantula_id"`
	TarantulaName      string     `json:"tarantula_name"`
	LastMoltDate       *time.Time `json:"last_molt_date"`
	DaysSinceLastMolt  int32      `json:"days_since_last_molt"`
	MoltCount          int32      `json:"molt_count"`
	CurrentSize        float64    `json:"current_size"`
	AdultSizeCM        *float64   `json:"adult_size_cm"`
	EstimatedAgeMonths int32      `json:"estimated_age_months"`
	SpeciesTemperament string     `json:"species_temperament"`
	SpeciesName        string     `json:"species_name"`
	RecentFeedings     int32      `json:"recent_feedings"`
	RecentRejections   int32      `json:"recent_rejections"`
	AverageCycle       *float64   `json:"average_cycle"`
	LastFeedingDate    *time.Time `json:"last_feeding_date"`
	FeedingFreqMinDays *int32     `json:"feeding_freq_min_days"`
	FeedingFreqMaxDays *int32     `json:"feeding_freq_max_days"`
}, confidenceAdjustment int) string {

	baseScore := 0

	// Historical data quality
	if result.MoltCount >= 3 && result.AverageCycle != nil {
		baseScore += 3 // Excellent historical data
	} else if result.MoltCount >= 2 {
		baseScore += 2 // Good historical data
	} else if result.MoltCount >= 1 {
		baseScore += 1 // Some historical data
	}

	// Species data availability
	if result.AdultSizeCM != nil && *result.AdultSizeCM > 0 {
		baseScore += 1 // Have adult size reference
	}
	if result.FeedingFreqMinDays != nil {
		baseScore += 1 // Have feeding schedule data
	}

	// Recent behavioral indicators
	if result.RecentRejections > 2 {
		baseScore += 1 // Pre-molt signs
	}

	// Apply adjustment
	finalScore := baseScore + confidenceAdjustment

	switch {
	case finalScore >= 5:
		return "High"
	case finalScore >= 3:
		return "Medium"
	case finalScore >= 1:
		return "Low"
	default:
		return "None"
	}
}

// Enhanced molt recommendations
func (db *TarantulaDB) generateMoltRecommendation(result struct {
	TarantulaID        int32      `json:"tarantula_id"`
	TarantulaName      string     `json:"tarantula_name"`
	LastMoltDate       *time.Time `json:"last_molt_date"`
	DaysSinceLastMolt  int32      `json:"days_since_last_molt"`
	MoltCount          int32      `json:"molt_count"`
	CurrentSize        float64    `json:"current_size"`
	AdultSizeCM        *float64   `json:"adult_size_cm"`
	EstimatedAgeMonths int32      `json:"estimated_age_months"`
	SpeciesTemperament string     `json:"species_temperament"`
	SpeciesName        string     `json:"species_name"`
	RecentFeedings     int32      `json:"recent_feedings"`
	RecentRejections   int32      `json:"recent_rejections"`
	AverageCycle       *float64   `json:"average_cycle"`
	LastFeedingDate    *time.Time `json:"last_feeding_date"`
	FeedingFreqMinDays *int32     `json:"feeding_freq_min_days"`
	FeedingFreqMaxDays *int32     `json:"feeding_freq_max_days"`
}, prediction *models.MoltPrediction) string {

	if result.LastMoltDate == nil {
		return "Record first molt to enable predictions."
	}

	// Recently molted
	if result.DaysSinceLastMolt < 14 {
		return "Recently molted. Continue normal care routine."
	}

	// Pre-molt signs detected
	if result.RecentRejections > 2 {
		return "🔄 Pre-molt signs detected! Stop feeding, ensure humidity 65-75%, avoid handling."
	}

	// Feeding behavior changes
	if result.RecentFeedings == 0 && result.LastFeedingDate != nil {
		daysSinceFeeding := int32(time.Since(*result.LastFeedingDate).Hours() / 24)
		if daysSinceFeeding > 14 {
			return "⚠️ Extended food refusal - likely in pre-molt. Monitor for molt signs."
		}
	}

	// REALISTIC prediction-based recommendations for HOME KEEPERS
	if prediction.DaysUntilMolt != nil {
		switch {
		case *prediction.DaysUntilMolt <= -60: // 2+ months overdue
			// This is unrealistic - molt predictions are estimates, not strict schedules
			return "🟢 Molt timing varies greatly. Continue normal care unless showing pre-molt signs."
		case *prediction.DaysUntilMolt <= -30: // 1 month overdue
			return "🟡 Later than estimated but normal. Some tarantulas molt unpredictably."
		case *prediction.DaysUntilMolt <= 0: // Recently "due"
			return "🟡 Around expected molt time. Watch for food refusal and increased humidity needs."
		case *prediction.DaysUntilMolt <= 14: // Within 2 weeks
			return "🟡 Molt possible soon. Monitor feeding behavior, avoid unnecessary handling."
		case *prediction.DaysUntilMolt <= 30: // Within a month
			return "🟢 Molt expected in coming weeks. Watch for behavioral changes."
		case *prediction.DaysUntilMolt <= 90: // Within 3 months
			return "🟢 Continue normal care routine. No molt expected soon."
		default:
			return "🟢 Continue normal care routine. Long time until next estimated molt."
		}
	}

	return "Continue normal care routine."
}

// ========== Tarantula Colony Service Implementation ==========

func (db *TarantulaDB) CreateColony(ctx context.Context, colony models.TarantulaColony) (int64, error) {
	result := db.db.WithContext(ctx).Create(&colony)
	if result.Error != nil {
		return 0, fmt.Errorf("failed to create tarantula colony: %w", result.Error)
	}
	return int64(colony.ID), nil
}

func (db *TarantulaDB) GetColony(ctx context.Context, colonyID int32, userID int64) (*models.TarantulaColony, error) {
	var colony models.TarantulaColony
	result := db.db.WithContext(ctx).
		Preload("Species").
		Preload("Enclosure").
		Preload("Members").
		Preload("Members.Tarantula").
		Where("id = ? AND user_id = ?", colonyID, userID).
		First(&colony)

	if result.Error != nil {
		return nil, fmt.Errorf("failed to get colony: %w", result.Error)
	}
	return &colony, nil
}

func (db *TarantulaDB) GetUserColonies(ctx context.Context, userID int64) ([]models.TarantulaColony, error) {
	var colonies []models.TarantulaColony
	result := db.db.WithContext(ctx).
		Preload("Species").
		Preload("Enclosure").
		Preload("Members", "is_active = ?", true).
		Preload("Members.Tarantula").
		Where("user_id = ?", userID).
		Order("formation_date DESC").
		Find(&colonies)

	if result.Error != nil {
		return nil, fmt.Errorf("failed to get colonies: %w", result.Error)
	}
	return colonies, nil
}

func (db *TarantulaDB) AddMemberToColony(ctx context.Context, member models.TarantulaColonyMember) error {
	return db.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		// Verify colony exists and belongs to user
		var colony models.TarantulaColony
		if err := tx.Where("id = ? AND user_id = ?", member.ColonyID, member.UserID).First(&colony).Error; err != nil {
			return fmt.Errorf("colony not found or access denied: %w", err)
		}

		// Verify tarantula exists and belongs to user
		var tarantula models.Tarantula
		if err := tx.Where("id = ? AND user_id = ?", member.TarantulaID, member.UserID).First(&tarantula).Error; err != nil {
			return fmt.Errorf("tarantula not found or access denied: %w", err)
		}

		// Verify species match
		if tarantula.SpeciesID != colony.SpeciesID {
			return fmt.Errorf("tarantula species does not match colony species")
		}

		// Create the membership record
		if err := tx.Create(&member).Error; err != nil {
			return fmt.Errorf("failed to add member to colony: %w", err)
		}

		// Update tarantula's colony_id
		if err := tx.Model(&tarantula).Update("colony_id", member.ColonyID).Error; err != nil {
			return fmt.Errorf("failed to update tarantula colony reference: %w", err)
		}

		return nil
	})
}

func (db *TarantulaDB) RemoveMemberFromColony(ctx context.Context, colonyID, tarantulaID int32, userID int64) error {
	return db.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		// Find active membership
		var member models.TarantulaColonyMember
		if err := tx.Where("colony_id = ? AND tarantula_id = ? AND is_active = ? AND user_id = ?",
			colonyID, tarantulaID, true, userID).First(&member).Error; err != nil {
			return fmt.Errorf("active membership not found: %w", err)
		}

		// Mark as inactive and set left date
		now := time.Now()
		member.IsActive = false
		member.LeftDate = &now
		if err := tx.Save(&member).Error; err != nil {
			return fmt.Errorf("failed to update membership: %w", err)
		}

		// Clear tarantula's colony_id
		var tarantula models.Tarantula
		if err := tx.Where("id = ? AND user_id = ?", tarantulaID, userID).First(&tarantula).Error; err != nil {
			return fmt.Errorf("tarantula not found: %w", err)
		}
		if err := tx.Model(&tarantula).Update("colony_id", nil).Error; err != nil {
			return fmt.Errorf("failed to clear tarantula colony reference: %w", err)
		}

		return nil
	})
}

func (db *TarantulaDB) GetColonyMembers(ctx context.Context, colonyID int32, userID int64, activeOnly bool) ([]models.TarantulaColonyMember, error) {
	var members []models.TarantulaColonyMember

	query := db.db.WithContext(ctx).
		Preload("Tarantula").
		Preload("Tarantula.Species").
		Where("colony_id = ? AND user_id = ?", colonyID, userID)

	if activeOnly {
		query = query.Where("is_active = ?", true)
	}

	result := query.Order("joined_date ASC").Find(&members)

	if result.Error != nil {
		return nil, fmt.Errorf("failed to get colony members: %w", result.Error)
	}
	return members, nil
}

func (db *TarantulaDB) UpdateColony(ctx context.Context, colony models.TarantulaColony) error {
	result := db.db.WithContext(ctx).
		Where("id = ? AND user_id = ?", colony.ID, colony.UserID).
		Updates(&colony)

	if result.Error != nil {
		return fmt.Errorf("failed to update colony: %w", result.Error)
	}
	if result.RowsAffected == 0 {
		return fmt.Errorf("colony not found or access denied")
	}
	return nil
}

// GetColoniesDueFeeding returns colonies that need feeding based on their species schedule
func (db *TarantulaDB) GetColoniesDueFeeding(ctx context.Context, userID int64) ([]models.TarantulaColony, error) {
	var colonies []models.TarantulaColony

	// Get all user's colonies with species and member info
	result := db.db.WithContext(ctx).
		Preload("Species").
		Preload("Members", "is_active = ?", true).
		Preload("Members.Tarantula").
		Where("user_id = ?", userID).
		Find(&colonies)

	if result.Error != nil {
		return nil, fmt.Errorf("failed to get colonies: %w", result.Error)
	}

	var dueColonies []models.TarantulaColony

	for _, colony := range colonies {
		// Skip colonies with no active members
		if len(colony.Members) == 0 {
			continue
		}

		// Get last colony feeding
		var lastFeeding models.FeedingEvent
		err := db.db.WithContext(ctx).
			Where("tarantula_colony_id = ?", colony.ID).
			Order("feeding_date DESC").
			First(&lastFeeding).Error

		daysSinceFeeding := 999.0
		if err == nil {
			daysSinceFeeding = time.Since(lastFeeding.FeedingDate).Hours() / 24
		}

		// Get feeding schedule for this species
		// Use average size for colony members
		var avgSize float64
		for _, member := range colony.Members {
			if member.Tarantula.CurrentSize > 0 {
				avgSize += member.Tarantula.CurrentSize
			}
		}
		if len(colony.Members) > 0 {
			avgSize /= float64(len(colony.Members))
		}

		var schedule models.FeedingSchedule
		err = db.db.WithContext(ctx).
			Where("species_id = ? AND body_length_cm >= ?", colony.SpeciesID, avgSize).
			Order("body_length_cm ASC").
			First(&schedule).Error

		if err == nil && schedule.FrequencyID > 0 {
			var frequency models.FeedingFrequency
			if err = db.db.First(&frequency, schedule.FrequencyID).Error; err == nil {
				// Check if colony is due
				if daysSinceFeeding >= float64(frequency.MinDays) {
					dueColonies = append(dueColonies, colony)
				}
			}
		}
	}

	return dueColonies, nil
}
