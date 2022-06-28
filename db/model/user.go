package model

type User struct {
	Base
	Username    string        `gorm:"unique" json:"username"`
	Displayname string        `json:"displayname"`
	Pass        string        `json:"-"`
	Memberships []*Membership `json:"memberships"`
	Sessions    []Session     `json:"sessions"`
}
