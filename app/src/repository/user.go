package repository

import (
	"errors"
	log "github.com/chris-dot-exe/AwesomeLog"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
	"time"
)

type User struct {
	gorm.Model
	GuildId  string `gorm:"index:idx_guildUser,unique"`
	UserId   string `gorm:"index:idx_guildUser,unique"`
	Username string
	Nickname *string
	Birthday time.Time
}

func (r Repo) GetUsers(birthday time.Time) ([]User, error) {
	log.Printf(log.INFO, "GET BIRTHDAYS FOR: %d/%d\n", birthday.Day(), birthday.Month())

	var birthdayUsers []User
	month := birthday.Month()
	day := birthday.Day()
	err := r.db.Where("EXTRACT(MONTH FROM birthday) = ? AND EXTRACT(DAY FROM birthday) = ?", month, day).Find(&birthdayUsers).Error

	return birthdayUsers, err
}

func (r Repo) UpsertUser(user *User) error {
	log.Printf(log.INFO, "ADDED USER: %v\n", user)
	return r.db.Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "user_id"}, {Name: "guild_id"}},
		DoUpdates: clause.AssignmentColumns([]string{"birthday", "username", "nickname"})}).
		Create(user).Error
}

func (r Repo) GetOrCreateUser(user *User) error {
	existingUser := User{}
	err := r.db.Where("user_id = ? AND guild_id = ?", user.UserId, user.GuildId).First(&existingUser).Error

	if err == nil {
		*user = existingUser
		return nil
	}

	if errors.Is(err, gorm.ErrRecordNotFound) {
		if err := r.db.Create(user).Error; err != nil {
			return err
		}
		log.Printf("CREATED USER: %v\n", user)
		return nil
	}

	return err
}

func (r Repo) RemoveBirthday(user User) error {
	log.Printf(log.INFO, "REMOVED USER: %v\n", user)

	return r.db.Unscoped().Where(User{GuildId: user.GuildId, UserId: user.UserId}).Delete(&User{}).Error
}
