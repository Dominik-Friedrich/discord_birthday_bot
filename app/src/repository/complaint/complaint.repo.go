package complaint

import (
	log "github.com/chris-dot-exe/AwesomeLog"
	"main/src/lib/database"
	"main/src/repository/birthday"
)

type Repository interface {
	AddComplaintReply(reply Reply) error
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
		birthday.User{},
		Complaint{},
		Reply{},
	)
}

func (r Repo) AddComplaintReply(reply Reply) error {
	//TODO implement me
	panic("implement me")
}

func (r Repo) GetComplaintReplies() ([]Reply, error) {
	var complaintReplies []Reply

	err := r.db.Find(&complaintReplies).Error

	return complaintReplies, err
}

func (r Repo) AddComplaint(complaint Complaint) error {
	// TODO: bug, user not being saved
	err := r.db.FirstOrCreate(&complaint.User).Error
	if err != nil {
		return err
	}
	return r.db.Create(&complaint).Error
}
