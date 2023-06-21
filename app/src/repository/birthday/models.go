package birthday

import (
	"gorm.io/gorm"
	"time"
)

// Very simple and not database design conform models

type User struct {
	gorm.Model
	GuildId  string `gorm:"index:idx_guildUser,unique"`
	UserId   string `gorm:"index:idx_guildUser,unique"`
	UserName string
	Birthday time.Time
}

type Role struct {
	gorm.Model
	GuildId string `gorm:"unique"`
	RoleId  string
}
