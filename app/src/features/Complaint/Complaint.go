package Complaint

import (
	"main/src/bot"
	"main/src/database"
	"main/src/features/Complaint/commands"
	"main/src/repository/complaint"
)

const (
	featureComplaint = "featureComplaint"
)

type ComplaintFeature struct {
	session *bot.Session
	repo    complaint.Repository
	replies *commands.Cache
}

func Complaint(connection *database.Connection) bot.Feature {
	b := new(ComplaintFeature)

	b.repo = complaint.NewRepository(connection)

	return b
}

func (b *ComplaintFeature) Init(session *bot.Session) error {
	b.session = session
	b.replies = new(commands.Cache)

	return nil
}

func (b *ComplaintFeature) Name() string {
	return featureComplaint
}

func (b *ComplaintFeature) Commands() []bot.Command {
	return []bot.Command{
		commands.Complain(b.repo, b.replies),
		//commands.AddComplaintResponse(b.repo), TODO implement command
	}
}
