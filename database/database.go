package database

import (
	"os"
	"store_backend/environment"
	"store_backend/models"

	slogGorm "github.com/orandin/slog-gorm"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func Initialize(env environment.Environment) *gorm.DB {
	gormLogger := slogGorm.New(
		slogGorm.WithHandler(env.Logger.Handler()),
	)

	db, err := gorm.Open(sqlite.Open(env.DSN), &gorm.Config{Logger: gormLogger})
	if err != nil {
		panic(err)
	}

	err = db.AutoMigrate(
		&models.Product{},
		&models.Category{},
		&models.Cart{},
	)

	if err != nil {
		panic(err)
	}

	if env.ENV == environment.Development && os.Getenv("SEED") == "true" {
		if err := Seed(db); err != nil {
			panic(err)
		}
	}

	return db
}
