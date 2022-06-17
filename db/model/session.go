package model

import (
	"time"

	"gorm.io/gorm"
)

type Session struct {
	UserID    uint           `gorm:"primaryKey"`
	IP        string         `gorm:"primaryKey"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"deleted_at"`
	Ch        string
}
