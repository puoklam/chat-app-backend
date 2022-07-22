package model

import "github.com/google/uuid"

const (
	DefaultGroupImage = "https://firebasestorage.googleapis.com/v0/b/instant-messenger-ab5e3.appspot.com/o/groups%2Fdefault.png?alt=media&token=13312afc-5f6a-4c97-ace9-8950f588e051"
)

type Group struct {
	Base
	Name        string        `json:"name"`
	Description string        `json:"description"`
	ImageURL    string        `json:"image_url"`
	Host        string        `json:"-"`
	Topic       uuid.UUID     `json:"-" gorm:"type:uuid;default:uuid_generate_v4()"`
	OwnerID     uint          `json:"owner_id"`
	Owner       *User         `json:"owner" gorm:"foreignKey:OwnerID"`
	Memberships []*Membership `json:"memberships"`
}
