package ws

import "github.com/puoklam/chat-app-backend/model"

type User struct {
	model.User
	sessions map[string]*Client
}
