package repository

import (
	"errors"
	log "github.com/chris-dot-exe/AwesomeLog"
	"time"
)

type User struct {
	UserId   string
	UserName string
	Birthday time.Time
}

type BirthdayRepo interface {
	AddBirthday(user User) error
	RemoveBirthday(user User) error
	GetBirthday(birthday time.Time) (User, error)
}

type Dummy struct {
	birthdays map[time.Time]User
}

func (d *Dummy) AddBirthday(user User) error {
	log.Println(log.INFO, "ADDED BIRTHDAY")
	d.birthdays[user.Birthday] = user
	log.PrettyPrint(log.INFO, d.birthdays)

	return nil
}

func (d *Dummy) RemoveBirthday(user User) error {
	log.Println(log.INFO, "REMOVED BIRTHDAY")
	d.birthdays[user.Birthday] = user
	log.PrettyPrint(log.INFO, d.birthdays)

	return nil
}

func (d *Dummy) GetBirthday(birthday time.Time) (User, error) {
	if bd, ok := d.birthdays[birthday]; ok {
		return bd, nil
	}
	return User{}, errors.New("no user with birthday today")
}

func NewBirthdayRepo() BirthdayRepo {
	repo := new(Dummy)
	repo.birthdays = make(map[time.Time]User)
	return repo
}
