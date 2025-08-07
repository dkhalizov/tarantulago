package bot

import (
	"fmt"
	"strconv"
	"strings"
	"tarantulago/models"
	"time"

	tele "gopkg.in/telebot.v4"
)

func SendSuccess(c tele.Context, message string) error {
	return c.Send("✅ " + message)
}

func SendError(c tele.Context, message string) error {
	return c.Send("❌ " + message)
}

func SendInfo(c tele.Context, message string) error {
	return c.Send("ℹ️ " + message)
}

func SendWarning(c tele.Context, message string) error {
	return c.Send("⚠️ " + message)
}

type SimpleCallback struct {
	Action string
	ID     int32
	Extra  string
}

func ParseCallback(data string) SimpleCallback {
	parts := strings.Split(data, ":")
	callback := SimpleCallback{Action: parts[0]}

	if len(parts) > 1 {
		if id, err := strconv.Atoi(parts[1]); err == nil {
			callback.ID = int32(id)
		}
	}

	if len(parts) > 2 {
		callback.Extra = parts[2]
	}

	return callback
}

func FormatDate(t *time.Time) string {
	if t == nil {
		return "Never"
	}
	return t.Format("2006-01-02")
}

func FormatDateTime(t *time.Time) string {
	if t == nil {
		return "Never"
	}
	return t.Format("2006-01-02 15:04")
}

func FormatDaysAgo(t *time.Time) string {
	if t == nil {
		return "Never"
	}

	days := int(time.Since(*t).Hours() / 24)
	if days == 0 {
		return "Today"
	} else if days == 1 {
		return "Yesterday"
	} else {
		return fmt.Sprintf("%d days ago", days)
	}
}

func GetFeedingStatus(daysSince int, minDays, maxDays int) (string, string) {
	if daysSince <= minDays {
		return "🟢", "Good"
	} else if daysSince <= maxDays {
		return "🟡", "Due Soon"
	} else {
		return "🔴", "Overdue"
	}
}

func GetFeedingStatusWithMolt(daysSince int, minDays, maxDays int, currentStatus string) (string, string) {

	if currentStatus == "Molting cycle" {
		return "🔄", "Molting"
	}

	return GetFeedingStatus(daysSince, minDays, maxDays)
}

func GetHealthStatusEmoji(statusID int) string {
	switch statusID {
	case 1:
		return "✅"
	case 2:
		return "🔄"
	case 3:
		return "🔄"
	case 4:
		return "🔄"
	case 5:
		return "🚨"
	default:
		return "❓"
	}
}

func GetMoltStageEmoji(stageID int) string {
	switch stageID {
	case 1:
		return "🕷️"
	case 2:
		return "🔄"
	case 3:
		return "🌟"
	case 4:
		return "✨"
	default:
		return "❓"
	}
}

func ValidateCricketCount(input string) (int, error) {
	count, err := strconv.Atoi(input)
	if err != nil {
		return 0, fmt.Errorf("invalid cricket count format")
	}

	if count < 1 || count > 50 {
		return 0, fmt.Errorf("cricket count must be between 1 and 50")
	}

	return count, nil
}

func TarantulaListToMarkup(tarantulas []models.TarantulaListItem) *tele.ReplyMarkup {
	markup := &tele.ReplyMarkup{}

	var rows []tele.Row
	for _, t := range tarantulas {
		emoji, _ := GetFeedingStatusWithMolt(int(t.DaysSinceFeeding), int(t.MinDays), int(t.MaxDays), t.CurrentStatus)
		btn := markup.Data(
			fmt.Sprintf("%s %s", emoji, t.Name),
			"select", strconv.Itoa(int(t.ID)),
		)
		rows = append(rows, markup.Row(btn))
	}

	markup.Inline(rows...)
	return markup
}

func BuildTarantulaActionsMarkup(tarantulaID int32) *tele.ReplyMarkup {
	markup := &tele.ReplyMarkup{}

	feedBtn := markup.Data("🍽️", fmt.Sprintf("feed:%d", tarantulaID))
	weightBtn := markup.Data("⚖️", fmt.Sprintf("weight:%d", tarantulaID))
	photoBtn := markup.Data("📸", fmt.Sprintf("photo:%d", tarantulaID))
	moltBtn := markup.Data("🔄", fmt.Sprintf("molt:%d", tarantulaID))

	historyBtn := markup.Data("📊", fmt.Sprintf("weight_history:%d", tarantulaID))
	photosBtn := markup.Data("🖼️", fmt.Sprintf("view_photos:%d", tarantulaID))
	intelligenceBtn := markup.Data("🧠", fmt.Sprintf("intel:%d", tarantulaID))
	predictionBtn := markup.Data("🔮", fmt.Sprintf("molt_pred:%d", tarantulaID))

	backBtn := markup.Data("⬅️ Back", "back_to_list")

	markup.Inline(
		markup.Row(feedBtn, weightBtn, photoBtn, moltBtn),
		markup.Row(historyBtn, photosBtn, intelligenceBtn, predictionBtn),
		markup.Row(backBtn),
	)

	return markup
}

func FormatFeedingPattern(pattern models.FeedingPattern) string {
	msg := fmt.Sprintf("*%s*\n", pattern.TarantulaName)
	msg += fmt.Sprintf("• Total feedings: %d\n", pattern.TotalFeedings)
	msg += fmt.Sprintf("• Acceptance rate: %.1f%%\n", pattern.AcceptanceRate)
	if pattern.AverageInterval > 0 {
		msg += fmt.Sprintf("• Average interval: %.1f days\n", pattern.AverageInterval)
	}
	msg += fmt.Sprintf("• Regularity: %s\n", pattern.FeedingRegularity)
	if pattern.CricketsPerWeek > 0 {
		msg += fmt.Sprintf("• Consumption: %.1f crickets/week\n", pattern.CricketsPerWeek)
	}
	msg += fmt.Sprintf("• Days since feeding: %d\n", pattern.DaysSinceLastFeeding)
	return msg
}

func FormatGrowthData(data models.GrowthData) string {
	msg := fmt.Sprintf("*%s*\n", data.TarantulaName)

	msg += fmt.Sprintf("• Current size: %.1fcm\n", data.CurrentSize)

	if data.GrowthRate != nil {
		trendEmoji := "📈"
		if *data.GrowthRate < 0 {
			trendEmoji = "📉"
		} else if *data.GrowthRate == 0 {
			trendEmoji = "➡️"
		}
		msg += fmt.Sprintf("• Growth rate: %s %.2fg/month\n", trendEmoji, *data.GrowthRate)
	}

	if len(data.SizeHistory) > 1 {
		msg += fmt.Sprintf("• Size records: %d molts\n", len(data.SizeHistory))
		msg += fmt.Sprintf("• Total size growth: %+.1fcm\n", data.SizeChangeTotal)
	}
	return msg
}

func FormatMoltPrediction(prediction models.MoltPrediction) string {
	msg := fmt.Sprintf("*%s*\n", prediction.TarantulaName)

	if prediction.LastMoltDate != nil {
		msg += fmt.Sprintf("• Last molt: %s\n", FormatDaysAgo(prediction.LastMoltDate))
	}

	if prediction.PredictedMoltDate != nil {
		msg += fmt.Sprintf("• Predicted molt: %s\n", FormatDate(prediction.PredictedMoltDate))
		if prediction.DaysUntilMolt != nil {
			if *prediction.DaysUntilMolt > 0 {
				msg += fmt.Sprintf("• Days until: %d\n", *prediction.DaysUntilMolt)
			} else {
				msg += "• Molt overdue!\n"
			}
		}
	}

	msg += fmt.Sprintf("• Confidence: %s\n", prediction.ConfidenceLevel)
	msg += fmt.Sprintf("• Size: %s\n", prediction.SizeIndicator)
	msg += fmt.Sprintf("• Feeding: %s\n", prediction.FeedingBehavior)

	if len(prediction.PreMoltSigns) > 0 {
		msg += "• Signs: " + strings.Join(prediction.PreMoltSigns, ", ") + "\n"
	}

	if prediction.Recommendation != "" {
		msg += fmt.Sprintf("• Advice: %s\n", prediction.Recommendation)
	}

	return msg
}

func IsValidState(state FormState, validStates ...FormState) bool {
	for _, valid := range validStates {
		if state == valid {
			return true
		}
	}
	return false
}

// Context helpers
func GetUserID(c tele.Context) int64 {
	return c.Sender().ID
}

func GetChatID(c tele.Context) int64 {
	return c.Chat().ID
}

func GetUsername(c tele.Context) string {
	if c.Sender().Username != "" {
		return c.Sender().Username
	}
	return c.Sender().FirstName
}

// Simplified error handling
func HandleError(c tele.Context, err error, operation string) error {
	if err != nil {
		// Log error (in real implementation, use proper logging)
		fmt.Printf("Error in %s: %v\n", operation, err)
		return SendError(c, fmt.Sprintf("Failed to %s. Please try again.", operation))
	}
	return nil
}

// Common filters for lists
func FilterOverdueTarantulas(tarantulas []models.TarantulaListItem) []models.TarantulaListItem {
	var overdue []models.TarantulaListItem
	for _, t := range tarantulas {
		// Don't include tarantulas in molt cycle as overdue
		if t.CurrentStatus != "Molting cycle" && t.DaysSinceFeeding > float64(t.MaxDays) {
			overdue = append(overdue, t)
		}
	}
	return overdue
}

func FilterDueSoonTarantulas(tarantulas []models.TarantulaListItem) []models.TarantulaListItem {
	var dueSoon []models.TarantulaListItem
	for _, t := range tarantulas {
		// Don't include tarantulas in molt cycle as due soon
		if t.CurrentStatus != "Molting cycle" && t.DaysSinceFeeding >= float64(t.MinDays) && t.DaysSinceFeeding <= float64(t.MaxDays) {
			dueSoon = append(dueSoon, t)
		}
	}
	return dueSoon
}

// Species-specific feeding intelligence
type FeedingIntelligence struct {
	RecommendedDays    int    `json:"recommended_days"`
	PreySizeAdvice     string `json:"prey_size_advice"`
	FeedingNote        string `json:"feeding_note"`
	SpeciesBasedAdvice string `json:"species_based_advice"`
	AgeBasedAdvice     string `json:"age_based_advice"`
	SizeBasedAdvice    string `json:"size_based_advice"`
	MoltStageAdvice    string `json:"molt_stage_advice"`
}

func GetSpeciesFeedingIntelligence(species models.TarantulaSpecies, currentSize float64, estimatedAgeMonths int, currentMoltStage string, daysSinceFeeding float64) FeedingIntelligence {
	intel := FeedingIntelligence{}

	// Base recommendations by species temperament and size
	switch {
	case species.Temperament == "Aggressive" || species.Temperament == "Fast":
		intel.RecommendedDays = 7
		intel.SpeciesBasedAdvice = "Active species - regular feeding schedule recommended"
	case species.Temperament == "Docile" || species.Temperament == "Slow":
		intel.RecommendedDays = 10
		intel.SpeciesBasedAdvice = "Calm species - can handle longer intervals between feedings"
	default:
		intel.RecommendedDays = 7
		intel.SpeciesBasedAdvice = "Standard feeding schedule"
	}

	// Age-based adjustments
	if estimatedAgeMonths < 12 { // Juvenile
		intel.RecommendedDays = max(3, intel.RecommendedDays-3)
		intel.AgeBasedAdvice = "Juvenile - needs frequent feeding for growth"
		intel.PreySizeAdvice = "Small crickets (1/2 to 2/3 of abdomen width)"
	} else if estimatedAgeMonths < 24 { // Sub-adult
		intel.RecommendedDays = max(5, intel.RecommendedDays-1)
		intel.AgeBasedAdvice = "Sub-adult - moderate feeding frequency"
		intel.PreySizeAdvice = "Medium crickets (2/3 of abdomen width)"
	} else { // Adult
		intel.AgeBasedAdvice = "Adult - can handle longer feeding intervals"
		intel.PreySizeAdvice = "Large crickets (equal to abdomen width)"
	}

	// Size-based adjustments relative to species adult size
	sizeRatio := currentSize / species.AdultSizeCM
	switch {
	case sizeRatio < 0.3: // Very small
		intel.SizeBasedAdvice = "Very small - increase feeding frequency"
		intel.RecommendedDays = max(3, intel.RecommendedDays-2)
	case sizeRatio < 0.6: // Growing
		intel.SizeBasedAdvice = "Still growing - maintain regular feeding"
	case sizeRatio < 0.9: // Near adult
		intel.SizeBasedAdvice = "Near adult size - reduce feeding frequency"
		intel.RecommendedDays += 2
	default: // Adult size or larger
		intel.SizeBasedAdvice = "Adult size - extended feeding intervals OK"
		intel.RecommendedDays += 3
	}

	// Molt stage considerations
	switch currentMoltStage {
	case "Pre-molt":
		intel.MoltStageAdvice = "🔄 Pre-molt: No feeding until molt complete"
		intel.FeedingNote = "Tarantula in pre-molt - will refuse food"
	case "Molting":
		intel.MoltStageAdvice = "🌟 Molting: No feeding - wait for hardening"
		intel.FeedingNote = "Recently molted - wait 1-2 weeks before feeding"
	case "Post-molt":
		intel.MoltStageAdvice = "✨ Post-molt: Wait for exoskeleton to harden"
		intel.FeedingNote = "Post-molt recovery - resume feeding gradually"
	default:
		if daysSinceFeeding > float64(intel.RecommendedDays*2) {
			intel.FeedingNote = "⚠️ Extended fasting - monitor for pre-molt signs"
		} else if daysSinceFeeding > float64(intel.RecommendedDays) {
			intel.FeedingNote = "Due for feeding"
		} else {
			intel.FeedingNote = "Feeding schedule on track"
		}
	}

	return intel
}

func FormatFeedingIntelligence(intel FeedingIntelligence, speciesName, tarantulaName string) string {
	msg := fmt.Sprintf("🧠 **Feeding Intelligence: %s**\n", tarantulaName)
	msg += fmt.Sprintf("*%s*\n\n", speciesName)

	msg += fmt.Sprintf("🎯 **Recommended interval:** %d days\n", intel.RecommendedDays)
	msg += fmt.Sprintf("🦗 **Prey size:** %s\n", intel.PreySizeAdvice)
	msg += fmt.Sprintf("📝 **Status:** %s\n\n", intel.FeedingNote)

	if intel.SpeciesBasedAdvice != "" {
		msg += fmt.Sprintf("🕷️ **Species:** %s\n", intel.SpeciesBasedAdvice)
	}
	if intel.AgeBasedAdvice != "" {
		msg += fmt.Sprintf("⏰ **Age:** %s\n", intel.AgeBasedAdvice)
	}
	if intel.SizeBasedAdvice != "" {
		msg += fmt.Sprintf("📏 **Size:** %s\n", intel.SizeBasedAdvice)
	}
	if intel.MoltStageAdvice != "" {
		msg += fmt.Sprintf("🔄 **Molt:** %s\n", intel.MoltStageAdvice)
	}

	return msg
}

// Molt prediction helpers
func GetMoltPredictionSummary(predictions []models.MoltPrediction) string {
	if len(predictions) == 0 {
		return "No molt predictions available - need more historical data"
	}

	msg := "🔮 **Molt Predictions Summary**\n\n"

	upcomingCount := 0
	overdueCount := 0

	for _, pred := range predictions {
		emoji := "🕷️"
		status := ""

		if pred.DaysUntilMolt != nil {
			if *pred.DaysUntilMolt <= 0 {
				emoji = "🔴"
				status = "Overdue!"
				overdueCount++
			} else if *pred.DaysUntilMolt <= 30 {
				emoji = "🟡"
				status = fmt.Sprintf("%d days", *pred.DaysUntilMolt)
				upcomingCount++
			} else {
				emoji = "🟢"
				status = fmt.Sprintf("%d days", *pred.DaysUntilMolt)
			}
		} else {
			status = "Unknown"
		}

		msg += fmt.Sprintf("%s **%s** - %s (%s confidence)\n",
			emoji, pred.TarantulaName, status, pred.ConfidenceLevel)
	}

	msg += "\n"
	if overdueCount > 0 {
		msg += fmt.Sprintf("⚠️ %d tarantula(s) overdue for molt\n", overdueCount)
	}
	if upcomingCount > 0 {
		msg += fmt.Sprintf("📅 %d tarantula(s) expecting molt soon\n", upcomingCount)
	}

	return msg
}

// Format enhanced tarantula details with photos and weight
func FormatTarantulaDetailsEnhanced(tarantula *models.Tarantula, photos []models.TarantulaPhoto, _ *models.WeightRecord) string {
	msg := fmt.Sprintf("🕷️ **%s**\n", tarantula.Name)
	msg += fmt.Sprintf("*%s*\n\n", tarantula.Species.ScientificName)

	// Basic info
	msg += fmt.Sprintf("📅 **Acquired:** %s\n", tarantula.AcquisitionDate.Format("2006-01-02"))
	if tarantula.EstimatedAgeMonths > 0 {
		msg += fmt.Sprintf("🎂 **Estimated age:** %d months\n", tarantula.EstimatedAgeMonths)
	}

	// Current status
	msg += fmt.Sprintf("📏 **Current size:** %.1fcm\n", tarantula.CurrentSize)
	msg += fmt.Sprintf("🔄 **Molt stage:** %s\n", tarantula.CurrentMoltStage.StageName)
	msg += fmt.Sprintf("❤️ **Health status:** %s\n", tarantula.CurrentHealthStatus.StatusName)

	// No weight tracking for home use

	// Molt information
	if tarantula.LastMoltDate != nil {
		msg += fmt.Sprintf("🦋 **Last molt:** %s\n", FormatDaysAgo(tarantula.LastMoltDate))
	}

	// Photos information
	if len(photos) > 0 {
		msg += fmt.Sprintf("📸 **Recent photos:** %d available\n", len(photos))
	}

	// Notes
	if tarantula.Notes != "" {
		msg += fmt.Sprintf("\n📝 **Notes:** %s\n", tarantula.Notes)
	}

	return msg
}
