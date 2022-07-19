package model

type Conn struct {
	UserID uint   `json:"user_id" gorm:"primaryKey"`
	Topic  string `json:"topic" gorm:"primaryKey"`
	Count  int64  `json:"count"`
	User   *User  `json:"user"`
}
