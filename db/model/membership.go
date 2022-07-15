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
	UserID    uint             `json:"user_id" gorm:"primaryKey"`
	GroupID   uint             `json:"group_id" gorm:"primaryKey"`
	Status    membershipStatus `json:"status" gorm:"type:membership_status"`
	User      *User            `json:"user"`
	Group     *Group           `json:"group"`
}

func (s *membershipStatus) Scan(value any) error {
	*s = membershipStatus(value.(string))
	return nil
}

func (s membershipStatus) Value() (driver.Value, error) {
	return string(s), nil
}
