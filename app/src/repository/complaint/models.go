package complaint

import (
	"gorm.io/gorm"
	"main/src/repository/birthday"
)

type Complaint struct {
	gorm.Model
	UserId        uint
	birthday.User // reusing this because im lazy
	text          string
}

type Reply struct {
	gorm.Model
	// GuildId where the reply will be used
	GuildId string `gorm:"index:idx_guildUser,unique"`
	// UserId of the person that added the reply
	UserId string `gorm:"index:idx_guildUser,unique"`
	text   string `gorm:"uniqueIndex"`
}
