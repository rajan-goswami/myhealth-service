package entity

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type HealthRecord struct {
	ID        uint `gorm:"primarykey" json:"-"`
	CreatedAt time.Time
	UpdatedAt time.Time
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`

	UUID           uuid.UUID `gorm:"type:uuid;uniqueIndex;not null;default:uuid_generate_v4()" json:"recordUuid"`
	DeviceRecordID string    `json:"deviceRecordId"`
	UserID         string    `json:"userId"`
}
