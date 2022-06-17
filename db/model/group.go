package model

import "github.com/google/uuid"

type Group struct {
	Base
	Name    string    `json:"name"`
	Topic   uuid.UUID `json:"-" gorm:"type:uuid;default:uuid_generate_v4()"`
	Members []*User   `gorm:"many2many:memberships" json:"members"`
}
