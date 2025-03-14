package repository

import (
	log "github.com/chris-dot-exe/AwesomeLog"
	"gorm.io/gorm"
)

type Reply struct {
	gorm.Model
	// GuildId where the reply will be used
	GuildId string `gorm:"index:idx_guildUserText,unique"`
	// UserId of the person that added the reply
	UserId uint `gorm:"index:idx_guildUserText,unique"`
	User   *User

	Text string `gorm:"index:idx_guildUserText,unique"`
}

func (r Repo) AddComplaintReply(reply *Reply) error {
	if reply.User != nil {
		if err := r.GetOrCreateUser(reply.User); err != nil {
			return err
		}
		reply.UserId = reply.User.ID
	}

	if err := r.db.Create(reply).Error; err != nil {
		return err
	}

	log.Printf(log.INFO, "Added complaint reply: %+v\n", reply)
	return nil
}

func (r Repo) GetComplaintReplies() ([]Reply, error) {
	var complaintReplies []Reply

	err := r.db.Find(&complaintReplies).Error

	return complaintReplies, err
}
