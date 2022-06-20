package group

import (
	"github.com/puoklam/chat-app-backend/db/model"
)

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
