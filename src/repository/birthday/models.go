package birthday

import (
	"gorm.io/gorm"
	"time"
)

// Very simple and not database design conform models

type User struct {
	gorm.Model
	GuildId  string
	UserId   string
	UserName string
	Birthday time.Time
}

type Role struct {
	gorm.Model
	GuildId string
	RoleId  string
}
