package auth

import "github.com/puoklam/chat-app-backend/db/model"

type OutUser struct {
	model.Base
	Username    string           `json:"username"`
	Displayname string           `json:"displayname"`
	Pass        string           `json:"-"`
	Memberships []*OutMemberShip `json:"memberships"`
	Sessions    []*model.Session `json:"sessions"`
}

type OutMemberShip struct {
	model.Membership
	Group *model.Group `json:"group"`
}
