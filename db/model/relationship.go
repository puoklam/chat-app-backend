package model

import "github.com/google/uuid"

// status "default", "removed", , "accepted", "blocked"
const (
	StatusDefault  = "default"
	StatusRemoved  = "removed"
	StatusAccepted = "accepted"
	StatusBlocked  = "blocked"
)

type Relationship struct {
	User1ID               uint      `json:"user1_id" gorm:"primaryKey"`
	User2ID               uint      `json:"user2_id" gorm:"primaryKey"`
	User1                 *User     `json:"user1" gorm:"foreignKey:User1ID"`
	User2                 *User     `json:"user2" gorm:"foreignKey:User2ID"`
	ForwardStatus         string    `json:"forward_status"`
	BackwardStatus        string    `json:"backward_status"`
	ForwardNotifications  bool      `json:"forward_notifications"`
	BackwardNotifications bool      `json:"backward_notifications"`
	Host                  string    `json:"-"`
	Topic                 uuid.UUID `json:"-" gorm:"type:uuid;default:uuid_generate_v4()"`
}
