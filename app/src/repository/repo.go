package repository

import (
	log "github.com/chris-dot-exe/AwesomeLog"
	"main/src/lib/database"
	"time"
)

type Repository interface {
	UpsertUser(user *User) error
	RemoveBirthday(user User) error
	GetUsers(birthday time.Time) ([]User, error)

	SetBirthdayRoleId(guildId, roleId string) error
	GetBirthdayRoleId(guildId string) (string, error)

	AddComplaintReply(reply *Reply) error
	GetComplaintReplies() ([]Reply, error)

	AddComplaint(complaint Complaint) error
}

type Repo struct {
	db *database.Connection
}

func NewRepository(connection *database.Connection) Repository {
	br := new(Repo)
	br.db = connection

	err := br.initDatabase()
	if err != nil {
		log.Panicf("error initialising complaint repo: %v", err.Error())
	}

	return br
}

func (r Repo) initDatabase() error {
	return r.db.AutoMigrate(
		Reply{},
		User{},
		Role{},
		Complaint{},
	)
}
