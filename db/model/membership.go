package model

import (
	"database/sql/driver"
	"time"
)

type membershipStatus string

const (
	StatusActive   = "active"
	StatusDeleting = "deleting"
)

type Membership struct {
	CreatedAt time.Time        `json:"created_at"`
	UserID    uint             `gorm:"primaryKey"`
	GroupID   uint             `gorm:"primaryKey"`
	Status    membershipStatus `gorm:"type:membership_status"`
}

func (s *membershipStatus) Scan(value any) error {
	*s = membershipStatus(value.(string))
	return nil
}

func (s membershipStatus) Value() (driver.Value, error) {
	return string(s), nil
}
