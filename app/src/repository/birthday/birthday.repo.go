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
	log.Printf(log.INFO, "ADDED BIRTHDAY: %v", user)
	return r.db.Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "user_id"}, {Name: "guild_id"}},
		DoUpdates: clause.AssignmentColumns([]string{"birthday"})}).
		Create(&user).Error
}

func (r Repo) RemoveBirthday(user User) error {
	log.Printf(log.INFO, "REMOVED BIRTHDAY: %v", user)

	return r.db.Unscoped().Where(User{GuildId: user.GuildId, UserId: user.UserId}).Delete(&User{}).Error
}

func (r Repo) GetBirthdayUsers(birthday time.Time) ([]User, error) {
	log.Printf(log.INFO, "GET BIRTHDAYS FOR: %d/%d", birthday.Day(), birthday.Month())

	var birthdayUsers []User
	month := birthday.Month()
	day := birthday.Day()
	err := r.db.Where("EXTRACT(MONTH FROM birthday) = ? AND EXTRACT(DAY FROM birthday) = ?", month, day).Find(&birthdayUsers).Error

	return birthdayUsers, err
}

func (r Repo) SetBirthdayRoleId(guildId, roleId string) error {
	log.Printf(log.INFO, "SET BIRTHDAY ROLE: guildId=%s, roleId=%s", guildId, roleId)

	return r.db.Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "guild_id"}},
		DoUpdates: clause.AssignmentColumns([]string{"role_id"})}).
		Create(&Role{GuildId: guildId, RoleId: roleId}).Error
}

func (r Repo) GetBirthdayRoleId(guildId string) (string, error) {
	log.Printf(log.INFO, "GET BIRTHDAY ROLE: guildId=%s", guildId)

	var birthdayRole Role
	err := r.db.Where(&Role{GuildId: guildId}).First(&birthdayRole).Error

	// acceptable error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return "", nil
	}

	return birthdayRole.RoleId, err
}
