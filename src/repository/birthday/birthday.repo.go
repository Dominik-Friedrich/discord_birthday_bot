package birthday

import (
	"errors"
	log "github.com/chris-dot-exe/AwesomeLog"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
	"main/src/database"
	"time"
)

type Repository interface {
	UpsertBirthday(user User) error
	RemoveBirthday(user User) error
	GetBirthdayUsers(birthday time.Time) ([]User, error)

	SetBirthdayRoleId(guildId, roleId string) error
	GetBirthdayRoleId(guildId string) (string, error)
}

type Repo struct {
	db *database.Connection
}

func NewRepository(connection *database.Connection) Repository {
	br := new(Repo)
	br.db = connection

	err := br.initDatabase()
	if err != nil {
		log.Panicf("error initialising birthday repo: %v", err.Error())
	}

	return br
}

func (r Repo) initDatabase() error {
	return r.db.AutoMigrate(
		User{},
		Role{},
	)
}

func (r Repo) UpsertBirthday(user User) error {
	return r.db.Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "birthday"}},
		DoUpdates: clause.AssignmentColumns([]string{"birthday"})}).
		Create(&user).Error
}

func (r Repo) RemoveBirthday(user User) error {
	return r.db.Delete(&user).Error
}

func (r Repo) GetBirthdayUsers(birthday time.Time) ([]User, error) {
	var birthdayUsers []User
	err := r.db.Where(&User{Birthday: birthday}).Find(&birthdayUsers).Error

	return birthdayUsers, err
}

func (r Repo) SetBirthdayRoleId(guildId, roleId string) error {
	return r.db.Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "guild_id"}},
		DoUpdates: clause.AssignmentColumns([]string{"role_id"})}).
		Create(&Role{GuildId: guildId, RoleId: roleId}).Error
}

func (r Repo) GetBirthdayRoleId(guildId string) (string, error) {
	var birthdayRole Role
	err := r.db.Where(&Role{GuildId: guildId}).First(&birthdayRole).Error

	// acceptable error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return "", nil
	}

	return birthdayRole.RoleId, err
}
