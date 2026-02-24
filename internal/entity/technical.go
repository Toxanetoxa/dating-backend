package entity

import (
	"time"
)

// GeneralTechFields ...
type GeneralTechFields struct {
	ID        string    `json:"ID" gorm:"primaryKey"`
	CreatedAt time.Time `json:"createdAt"`
}
