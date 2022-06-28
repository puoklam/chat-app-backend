package group

import (
	"github.com/puoklam/chat-app-backend/db/model"
)

type OutGetGroup struct {
	model.Base
	Name   string `json:"name"`
	Joined bool   `json:"joined"`
}

type InCreateGroup struct {
	Name *string `json:"name"`
}

type OutCreateGroup struct {
	model.Base
	Name string
}

type OutListGroups struct {
	model.Base
	Name string `json:"name"`
}

type InCreateMsg struct {
	Message *string `json:"message"`
}
