package Complaint

import (
	log "github.com/chris-dot-exe/AwesomeLog"
	"main/src/bot"
	"main/src/features/Complaint/commands"
	"main/src/repository"
)

const (
	featureComplaint = "featureComplaint"
)

type ComplaintFeature struct {
	session *bot.Session
	repo    repository.Repository
	replies *commands.Cache
}

func Complaint(repo repository.Repository) bot.Feature {
	b := new(ComplaintFeature)

	b.repo = repo

	b.replies = new(commands.Cache)

	b.replies.Lock()
	defer b.replies.Unlock()
	replies, err := b.repo.GetComplaintReplies()
	if err != nil {
		log.Println(log.WARN, "unable to load complaint replies: ", err.Error())
	}
	b.replies.Refresh(replies)

	return b
}

func (b *ComplaintFeature) Init(session *bot.Session) error {
	b.session = session

	return nil
}

func (b *ComplaintFeature) Name() string {
	return featureComplaint
}

func (b *ComplaintFeature) Commands() []bot.Command {
	return []bot.Command{
		commands.Complain(b.repo, b.replies),
		commands.AddComplainReply(b.repo, b.replies),
	}
}
