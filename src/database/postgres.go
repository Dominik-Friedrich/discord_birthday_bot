package database

import (
	"fmt"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func postgresConnection(c Config) (*Connection, error) {
	dsn := fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%s sslmode=disable TimeZone=UTC", c.Host, c.User, c.Password, c.Database, c.Port)
	db, err := gorm.Open(postgres.Open(dsn))

	return &Connection{db}, err
}
