package repository

import (
	"errors"
	log "github.com/chris-dot-exe/AwesomeLog"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type Role struct {
	gorm.Model
	GuildId string `gorm:"unique"`
	RoleId  string
}

func (r Repo) SetBirthdayRoleId(guildId, roleId string) error {
	log.Printf(log.INFO, "SET BIRTHDAY ROLE: guildId=%s, roleId=%s\n", guildId, roleId)

	return r.db.Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "guild_id"}},
		DoUpdates: clause.AssignmentColumns([]string{"role_id"})}).
		Create(&Role{GuildId: guildId, RoleId: roleId}).Error
}

func (r Repo) GetBirthdayRoleId(guildId string) (string, error) {
	log.Printf(log.INFO, "GET BIRTHDAY ROLE: guildId=%s\n", guildId)

	var birthdayRole Role
	err := r.db.Where(&Role{GuildId: guildId}).First(&birthdayRole).Error

	// acceptable error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return "", nil
	}

	return birthdayRole.RoleId, err
}
