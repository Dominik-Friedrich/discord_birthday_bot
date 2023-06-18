package database

import (
	"fmt"
	"gorm.io/gorm"
)

type Config struct {
	Type     string `json:"type"`
	Host     string `json:"host"`
	Port     string `json:"port"`
	User     string `json:"user"`
	Password string `json:"password"`
	Database string `json:"database"`
}

type Connection struct {
	*gorm.DB
}

func NewConnection(c Config) (*Connection, error) {
	switch c.Type {
	case "postgres":
		return postgresConnection(c)
	default:
		return nil, fmt.Errorf("unsupported database type: %s", c.Type)
	}
}
