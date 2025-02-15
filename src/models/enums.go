package models

import "encoding/json"

type HealthStatusEnum int

const (
	HealthStatusHealthy  HealthStatusEnum = 1
	HealthStatusMonitor  HealthStatusEnum = 2
	HealthStatusCritical HealthStatusEnum = 3
)

func (h HealthStatusEnum) ToDBName() string {
	switch h {
	case HealthStatusHealthy:
		return "Healthy"
	case HealthStatusMonitor:
		return "Monitor"
	case HealthStatusCritical:
		return "Critical"
	default:
		return "Healthy"
	}
}

func (h HealthStatusEnum) Description() string {
	switch h {
	case HealthStatusHealthy:
		return "Normal health status with no concerns"
	case HealthStatusMonitor:
		return "Requires extra attention and monitoring"
	case HealthStatusCritical:
		return "Immediate attention required"
	default:
		return "Normal health status with no concerns"
	}
}

func HealthStatusFromID(id int32) HealthStatusEnum {
	switch id {
	case 1:
		return HealthStatusHealthy
	case 2:
		return HealthStatusMonitor
	case 3:
		return HealthStatusCritical
	default:
		return HealthStatusHealthy
	}
}

type FeedingStatusEnum int

const (
	FeedingStatusAccepted FeedingStatusEnum = 1
	FeedingStatusRejected FeedingStatusEnum = 2
	FeedingStatusPartial  FeedingStatusEnum = 3
	FeedingStatusPreMolt  FeedingStatusEnum = 4
	FeedingStatusDead     FeedingStatusEnum = 5
	FeedingStatusOverflow FeedingStatusEnum = 6
)

func (f FeedingStatusEnum) ToDBName() string {
	switch f {
	case FeedingStatusAccepted:
		return "Accepted"
	case FeedingStatusRejected:
		return "Rejected"
	case FeedingStatusPartial:
		return "Partial"
	case FeedingStatusPreMolt:
		return "Pre-molt"
	case FeedingStatusDead:
		return "Dead"
	case FeedingStatusOverflow:
		return "Overflow"
	default:
		return "Unknown"
	}
}

func (f FeedingStatusEnum) Description() string {
	switch f {
	case FeedingStatusAccepted:
		return "Food was accepted normally"
	case FeedingStatusRejected:
		return "Food was rejected"
	case FeedingStatusPartial:
		return "Only part of the food was consumed"
	case FeedingStatusPreMolt:
		return "Refused food due to pre-molt state"
	case FeedingStatusDead:
		return "Prey died without being eaten"
	case FeedingStatusOverflow:
		return "Too many prey items left in enclosure"
	default:
		return "Unknown feeding status"
	}
}

type MoltStageEnum int

const (
	MoltStageNormal   MoltStageEnum = 1
	MoltStagePreMolt  MoltStageEnum = 2
	MoltStageMolting  MoltStageEnum = 3
	MoltStagePostMolt MoltStageEnum = 4
	MoltStageFailed   MoltStageEnum = 5
)

func (m MoltStageEnum) ToDBName() string {
	switch m {
	case MoltStageNormal:
		return "Normal"
	case MoltStagePreMolt:
		return "Pre-molt"
	case MoltStageMolting:
		return "Molting"
	case MoltStagePostMolt:
		return "Post-molt"
	case MoltStageFailed:
		return "Failed"
	default:
		return "Unknown"
	}
}

func (m MoltStageEnum) Description() string {
	switch m {
	case MoltStageNormal:
		return "Regular feeding/activity cycle"
	case MoltStagePreMolt:
		return "Showing signs of upcoming molt"
	case MoltStageMolting:
		return "Currently in molt"
	case MoltStagePostMolt:
		return "Recently molted, needs time to harden"
	case MoltStageFailed:
		return "Experiencing molt complications"
	default:
		return "Unknown molt stage"
	}
}

type CricketSizeEnum int

const (
	CricketSizePinhead CricketSizeEnum = 1
	CricketSizeSmall   CricketSizeEnum = 2
	CricketSizeMedium  CricketSizeEnum = 3
	CricketSizeLarge   CricketSizeEnum = 4
	CricketSizeAdult   CricketSizeEnum = 5
	CricketSizeUnknown CricketSizeEnum = 6
)

func (c CricketSizeEnum) ToDBName() string {
	switch c {
	case CricketSizePinhead:
		return "Pinhead"
	case CricketSizeSmall:
		return "Small"
	case CricketSizeMedium:
		return "Medium"
	case CricketSizeLarge:
		return "Large"
	case CricketSizeAdult:
		return "Adult"
	case CricketSizeUnknown:
		return "Unknown"
	default:
		return "Unknown"
	}
}

func (c CricketSizeEnum) LengthMM() float32 {
	switch c {
	case CricketSizePinhead:
		return 2.0
	case CricketSizeSmall:
		return 5.0
	case CricketSizeMedium:
		return 10.0
	case CricketSizeLarge:
		return 15.0
	case CricketSizeAdult:
		return 20.0
	case CricketSizeUnknown:
		return 0.0
	default:
		return 0.0
	}
}

func (c CricketSizeEnum) MarshalJSON() ([]byte, error) {
	return []byte(`"` + c.ToDBName() + `"`), nil
}

func (c *CricketSizeEnum) UnmarshalJSON(data []byte) error {
	var s string
	if err := json.Unmarshal(data, &s); err != nil {
		return err
	}

	switch s {
	case "Pinhead":
		*c = CricketSizePinhead
	case "Small":
		*c = CricketSizeSmall
	case "Medium":
		*c = CricketSizeMedium
	case "Large":
		*c = CricketSizeLarge
	case "Adult":
		*c = CricketSizeAdult
	default:
		*c = CricketSizeUnknown
	}
	return nil
}
