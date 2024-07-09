package database

import (
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func migrate(db *gorm.DB) {
	db.AutoMigrate(&Node{})
}

func ConnectSQLite(name string) (*gorm.DB, error) {
	if name == "" {
		name = "tartarus.db"
	}
	db, err := gorm.Open(sqlite.Open(name), &gorm.Config{})
	if err != nil {
		return nil, err
	}

	migrate(db)

	return db, nil
}

// add more db connection functions here
