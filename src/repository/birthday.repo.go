package repository

import (
	"errors"
	"fmt"
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
	GetBirthdayUsers(birthday time.Time) ([]User, error)

	SetBirthdayRoleId(guildId, roleId string) error
	GetBirthdayRoleId(guildId string) (string, error)
}

type Dummy struct {
	birthdays      map[string][]User
	birthdayRoleId map[string]string
}

func (d *Dummy) SetBirthdayRoleId(guildId, userId string) error {
	d.birthdayRoleId[guildId] = userId

	return nil
}

func (d *Dummy) GetBirthdayRoleId(guildId string) (string, error) {
	return d.birthdayRoleId[guildId], nil
}

func NewBirthdayRepo() BirthdayRepo {
	repo := new(Dummy)
	repo.birthdays = make(map[string][]User)
	repo.birthdays[asKey(time.Now())] = append(repo.birthdays[asKey(time.Now())], User{UserId: "806972973696155720"})
	return repo
}

func (d *Dummy) AddBirthday(user User) error {
	if _, ok := d.birthdays[asKey(user.Birthday)]; !ok {
		d.birthdays[asKey(user.Birthday)] = make([]User, 0)
	}

	log.Println(log.INFO, "ADDED BIRTHDAY")
	d.birthdays[asKey(user.Birthday)] = append(d.birthdays[asKey(user.Birthday)], user)
	log.PrettyPrint(log.INFO, d.birthdays)

	return nil
}

func (d *Dummy) RemoveBirthday(user User) error {
	log.Println(log.INFO, "REMOVED BIRTHDAY")
	d.birthdays[asKey(user.Birthday)] = append(d.birthdays[asKey(user.Birthday)], user)
	log.PrettyPrint(log.INFO, d.birthdays)

	return nil
}

func (d *Dummy) GetBirthdayUsers(birthday time.Time) ([]User, error) {
	if bd, ok := d.birthdays[asKey(birthday)]; ok {
		return bd, nil
	}
	return []User{}, errors.New("no user with birthday today")
}

func asKey(birthday time.Time) string {
	return fmt.Sprintf("%d/%s", birthday.Day(), birthday.Month())
}
