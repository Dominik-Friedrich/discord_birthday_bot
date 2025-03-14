package repository

import "gorm.io/gorm"

type Complaint struct {
	gorm.Model
	ComplainantID uint
	Complainant   *User `gorm:"foreignKey:ComplainantID"`

	AgainstUserID *uint
	AgainstUser   *User `gorm:"foreignKey:AgainstUserID"`

	Text string
}

func (r Repo) AddComplaint(complaint Complaint) error {
	// too lazy to use a transaction here. its fiiiiine
	if complaint.Complainant != nil {
		if err := r.GetOrCreateUser(complaint.Complainant); err != nil {
			return err
		}
		complaint.ComplainantID = complaint.Complainant.ID
	}

	if complaint.AgainstUser != nil {
		if err := r.GetOrCreateUser(complaint.AgainstUser); err != nil {
			return err
		}
		complaint.AgainstUserID = &complaint.AgainstUser.ID
	}

	return r.db.Create(&complaint).Error
}
