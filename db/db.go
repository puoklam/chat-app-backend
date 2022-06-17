package db

import (
	"context"
	"os"

	"github.com/puoklam/chat-app-backend/db/model"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

var db *gorm.DB

func init() {
	conn := os.Getenv("DB_CONN")
	var err error
	db, err = gorm.Open(postgres.Open(conn), &gorm.Config{})
	if err != nil {
		panic(err)
	}
	db.AutoMigrate(&model.User{})
	db.AutoMigrate(&model.Group{})
	db.AutoMigrate(&model.Session{})
}

func GetDB(ctx context.Context) *gorm.DB {
	return db.WithContext(ctx)
}
