package model

import (
	"time"

	"gorm.io/gorm"
)

const (
	StatusOnline  = "online"
	StatusOffline = "offline"
)

type Session struct {
	UserID        uint           `json:"user_id" gorm:"primaryKey"`
	IP            string         `json:"ip" gorm:"primaryKey"`
	CreatedAt     time.Time      `json:"created_at"`
	UpdatedAt     time.Time      `json:"updated_at"`
	DeletedAt     gorm.DeletedAt `gorm:"index" json:"deleted_at"`
	Ch            string         `json:"-"`
	ExpoPushToken string         `json:"-"`
	Status        string         `json:"status"`
}
