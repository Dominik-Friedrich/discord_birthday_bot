package birthday

import (
	"main/src/database"
	"time"
)

type Repository interface {
	AddBirthday(user User) error
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

	return br
}

func (r Repo) AddBirthday(user User) error {
	//TODO implement me
	panic("implement me")
}

func (r Repo) RemoveBirthday(user User) error {
	//TODO implement me
	panic("implement me")
}

func (r Repo) GetBirthdayUsers(birthday time.Time) ([]User, error) {
	//TODO implement me
	panic("implement me")
}

func (r Repo) SetBirthdayRoleId(guildId, roleId string) error {
	//TODO implement me
	panic("implement me")
}

func (r Repo) GetBirthdayRoleId(guildId string) (string, error) {
	//TODO implement me
	panic("implement me")
}
