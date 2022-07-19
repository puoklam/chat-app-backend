package db

import (
	"context"

	"github.com/puoklam/chat-app-backend/db/model"
	"github.com/puoklam/chat-app-backend/env"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

var db *gorm.DB

func init() {
	conn := env.DB_CONN
	var err error
	db, err = gorm.Open(postgres.Open(conn), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Error),
	})
	// db, err = gorm.Open(postgres.Open(conn), &gorm.Config{})
	if err != nil {
		panic(err)
	}
	db.AutoMigrate(&model.User{})
	db.AutoMigrate(&model.Group{})
	db.AutoMigrate(&model.Membership{})
	db.AutoMigrate(&model.Session{})
	db.AutoMigrate(&model.Conn{})
}

func GetDB(ctx context.Context) *gorm.DB {
	return db.WithContext(ctx)
}
