package model

import "github.com/google/uuid"

type Group struct {
	Base
	Name        string        `json:"name"`
	Host        string        `json:"-"`
	Topic       uuid.UUID     `json:"-" gorm:"type:uuid;default:uuid_generate_v4()"`
	Memberships []*Membership `json:"members"`
}
