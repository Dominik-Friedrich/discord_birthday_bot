package complaint

import (
	"context"
	"log/slog"

	"github.com/Dominik-Friedrich/discord_birthday_bot/internal/user"

	"gorm.io/gorm"
)

type Reply struct {
	gorm.Model
	// GuildId where the reply will be used
	GuildId string `gorm:"index:idx_guildUserText,unique"`
	// UserId of the person that added the reply
	UserId uint `gorm:"index:idx_guildUserText,unique"`
	User   *user.User

	Text string `gorm:"index:idx_guildUserText,unique"`
}

func (r *Repository) AddComplaintReply(ctx context.Context, reply *Reply) error {
	if reply.User != nil {
		if err := r.users.GetOrCreateUser(ctx, reply.User); err != nil {
			return err
		}
		reply.UserId = reply.User.ID
	}

	if err := gorm.G[Reply](r.db.DB).Create(ctx, reply); err != nil {
		return err
	}

	slog.Info("added complaint reply", "guild_id", reply.GuildId, "reply_id", reply.ID)
	return nil
}

func (r *Repository) GetComplaintReplies(ctx context.Context) ([]Reply, error) {
	return gorm.G[Reply](r.db.DB).Find(ctx)
}
