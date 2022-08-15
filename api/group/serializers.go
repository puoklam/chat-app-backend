package group

import (
	"github.com/puoklam/chat-app-backend/db/model"
)

type OutGetGroup struct {
	model.Base
	Name         string              `json:"name"`
	Description  string              `json:"description"`
	ImageURL     string              `json:"image_url"`
	Joined       bool                `json:"joined"`
	Notification bool                `json:"notification"`
	OwnerID      uint                `json:"owner_id"`
	Owner        *model.User         `json:"owner"`
	Memberships  []*model.Membership `json:"memberships"`
}

type InCreateGroup struct {
	Name        *string `json:"name"`
	Description *string `json:"description"`
}

type OutCreateGroup struct {
	model.Base
	Name        string
	Description string `json:"description"`
	ImageURL    string `json:"image_url"`
}

type OutListGroups struct {
	model.Base
	Name        string      `json:"name"`
	Description string      `json:"description"`
	ImageURL    string      `json:"image_url"`
	OwnerID     uint        `json:"owner_id"`
	Owner       *model.User `json:"owner"`
}

type InUpdateGroup struct {
	Name        *string `json:"name"`
	Description *string `json:"description"`
}

type OutUpdateGroup struct {
	model.Base
	Name        string              `json:"name"`
	Description string              `json:"description"`
	ImageURL    string              `json:"image_url"`
	OwnerID     uint                `json:"owner_id"`
	Owner       *model.User         `json:"owner"`
	Memberships []*model.Membership `json:"memberships"`
}

type InCreateMsg struct {
	Message *string `json:"message"`
}
