package model

type User struct {
	Base
	Username    string    `gorm:"unique" json:"username"`
	Displayname string    `json:"displayname"`
	Pass        string    `json:"-"`
	Groups      []*Group  `gorm:"many2many:memberships" json:"groups"`
	Sessions    []Session `json:"sessions"`
}
