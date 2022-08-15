package model

type User struct {
	Base
	Email                 string          `gorm:"unique" json:"email"`
	Username              string          `gorm:"unique" json:"username"`
	Displayname           string          `json:"displayname"`
	ImageURL              string          `json:"image_url"`
	Pass                  string          `json:"-"`
	Bio                   string          `json:"bio"`
	Memberships           []*Membership   `json:"memberships"`
	Sessions              []*Session      `json:"sessions"`
	ForwardRelationships  []*Relationship `json:"forward_relationships" gorm:"foreignKey:User1ID"`
	BackwardRelationships []*Relationship `json:"backward_relationships" gorm:"foreignKey:User2ID"`
}
