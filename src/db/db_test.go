package db

import (
	"context"
	"fmt"
	"log"
	"tarantulago/models"
	"testing"
	"time"
)

func TestDatabaseOperations(t *testing.T) {
	connectionString := "postgres://postgres:postgrespassword@localhost:5432/postgres?sslmode=disable"

	database, err := NewTarantulaDB(connectionString)
	if err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
	}

	ctx := context.Background()

	// Test user ID
	testUser := &models.TelegramUser{
		TelegramID: int64(1),
		Username:   "testuser",
		FirstName:  "Test",
		LastName:   "User",
	}
	err = database.EnsureUserExists(ctx, testUser)
	if err != nil {
		t.Fatalf("Failed to ensure user exists: %v", err)
	}

	userID := testUser.TelegramID
	enclosure := models.Enclosure{
		Name:             "Test Enclosure",
		HeightCM:         20,
		WidthCM:          20,
		LengthCM:         30,
		SubstrateDepthCM: 5,
		UserID:           userID,
	}
	enclosureId, err := database.CreateEnclosure(ctx, enclosure)
	if err != nil {
		t.Fatalf("Failed to add enclosure: %v", err)
	}
	t.Logf("Enclosure ID: %d", enclosureId)
	// 1. Create a new tarantula
	tarantula := models.Tarantula{
		Name:                  "Test Spider",
		SpeciesID:             1,
		AcquisitionDate:       time.Now(),
		UserID:                userID,
		LastHealthCheckDate:   time.Now(),
		LastMoltDate:          nil,
		CurrentMoltStageID:    4,
		CurrentHealthStatusID: 1,
		//EnclosureID:           int(enclosureId),
	}
	err = database.AddTarantula(ctx, tarantula)
	if err != nil {
		t.Fatalf("Failed to add tarantula: %v", err)
	}

	// 2. Get all tarantulas
	tarantulas, err := database.GetAllTarantulas(ctx, userID)
	if err != nil {
		t.Fatalf("Failed to get tarantulas: %v", err)
	}
	fmt.Printf("Found %d tarantulas\n", len(tarantulas))

	// 3. Create a new cricket colony
	colony := models.CricketColony{
		ColonyName:   "Test Colony",
		CurrentCount: 100,
		UserID:       userID,
	}
	err = database.AddColony(ctx, colony)
	if err != nil {
		t.Fatalf("Failed to add colony: %v", err)
	}

	// 4. Get colony status
	colonies, err := database.GetColonyStatus(ctx, userID)
	if err != nil {
		t.Fatalf("Failed to get colony status: %v", err)
	}
	fmt.Printf("Found %d colonies\n", len(colonies))

	// 5. Record a feeding event
	if len(tarantulas) > 0 && len(colonies) > 0 {
		feedingEvent := models.FeedingEvent{
			TarantulaID:      int(tarantulas[0].ID),
			CricketColonyID:  int(colonies[0].ID),
			NumberOfCrickets: 2,
			Notes:            "Test feeding",
			UserID:           userID,
		}
		feedingID, err := database.RecordFeeding(ctx, feedingEvent)
		if err != nil {
			t.Fatalf("Failed to record feeding: %v", err)
		}
		fmt.Printf("Recorded feeding with ID: %d\n", feedingID)
	}

	//// 6. Create an enclosure
	//enclosure := models.Enclosure{
	//	EnclosureType:    "Terrestrial",
	//	Dimensions:       "30x30x30",
	//	SubstrateType:    "Coco fiber",
	//	HumidityRange:    "60-70%",
	//	TemperatureRange: "22-26°C",
	//	UserID:           userID,
	//}
	//enclosureID, err := database.CreateEnclosure(ctx, enclosure)
	//if err != nil {
	//	t.Fatalf("Failed to create enclosure: %v", err)
	//}

	// 7. Record a health check
	if len(tarantulas) > 0 {
		healthCheck := models.HealthCheckRecord{
			TarantulaID:        int(tarantulas[0].ID),
			CheckDate:          time.Now(),
			HealthStatusID:     1,
			WeightGrams:        5.5,
			HumidityPercent:    65,
			TemperatureCelsius: 24,
			Notes:              "Test health check",
			UserID:             userID,
		}
		err = database.RecordHealthCheck(ctx, healthCheck)
		if err != nil {
			t.Fatalf("Failed to record health check: %v", err)
		}
	}

	// 8. Record a molt
	if len(tarantulas) > 0 {
		lengthCM := float64(8.5)
		molt := models.MoltRecord{
			TarantulaID:      int(tarantulas[0].ID),
			MoltDate:         time.Now(),
			MoltStageID:      int(models.MoltStagePostMolt),
			PostMoltLengthCM: lengthCM,
			Complications:    "false",
			Notes:            "Test molt record",
			UserID:           userID,
		}
		err = database.RecordMolt(ctx, molt)
		if err != nil {
			t.Fatalf("Failed to record molt: %v", err)
		}
	}

	// 9. Get maintenance tasks
	tasks, err := database.GetMaintenanceTasks(ctx, userID)
	if err != nil {
		t.Fatalf("Failed to get maintenance tasks: %v", err)
	}
	fmt.Printf("Found %d maintenance tasks\n", len(tasks))

	// 11. Get recent records
	recentFeedings, err := database.GetRecentFeedingRecords(ctx, userID, 5)
	if err != nil {
		t.Fatalf("Failed to get recent feeding records: %v", err)
	}
	fmt.Printf("Found %d recent feeding records\n", len(recentFeedings))

	recentMolts, err := database.GetRecentMoltRecords(ctx, userID, 5)
	if err != nil {
		t.Fatalf("Failed to get recent molt records: %v", err)
	}
	fmt.Printf("Found %d recent molt records\n", len(recentMolts))

	recentHealth, err := database.GetRecentHealthRecords(ctx, userID, 5)
	if err != nil {
		t.Fatalf("Failed to get recent health records: %v", err)
	}
	fmt.Printf("Found %d recent health records\n", len(recentHealth))

	// 12. Get tarantulas due for feeding
	dueFeedingList, err := database.GetTarantulasDueFeeding(ctx, userID)
	if err != nil {
		t.Fatalf("Failed to get tarantulas due feeding: %v", err)
	}
	fmt.Printf("Found %d tarantulas due for feeding\n", len(dueFeedingList))

	// 13. Ensure test user exists

	fmt.Println("Database operations test completed!")
}
