package player

import (
	"context"
	"errors"
	"log/slog"

	"github.com/bwmarrin/discordgo"

	"github.com/Dominik-Friedrich/discord_birthday_bot/internal/bot"
)

const (
	playCommandName = "play"
	paramQuery      = "query"
)

type playCommand struct {
	player IPlayer
}

func Play(player IPlayer) bot.Command {
	return &playCommand{player: player}
}

func (p *playCommand) Name() string {
	return playCommandName
}

func (p *playCommand) Command() *discordgo.ApplicationCommand {
	neededPermissions := int64(
		discordgo.PermissionViewChannel |
			discordgo.PermissionVoiceConnect |
			discordgo.PermissionVoiceSpeak |
			discordgo.PermissionSendMessages,
	)

	return &discordgo.ApplicationCommand{
		Name:                     playCommandName,
		Description:              "Play a song either from a URL or search.",
		DefaultMemberPermissions: &neededPermissions,
		Options: []*discordgo.ApplicationCommandOption{
			{
				Type:        discordgo.ApplicationCommandOptionString,
				Name:        paramQuery,
				Description: "The query to search for.",
				Required:    true,
			},
		},
	}
}

func (p *playCommand) Handle(s *discordgo.Session, i *discordgo.InteractionCreate) {
	// Resolving can take a few seconds -- defer to avoid missing Discord's
	// 3-second response window.
	if err := s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseDeferredChannelMessageWithSource,
	}); err != nil {
		slog.Warn("error deferring interaction response", "error", err)
		return
	}

	track, startedImmediately, err := p.playAudio(i)
	if err != nil {
		slog.Warn("error playing audio", "error", err)
		errMsg := err.Error()
		if _, err := s.InteractionResponseEdit(i.Interaction, &discordgo.WebhookEdit{Content: &errMsg}); err != nil {
			slog.Warn("error editing interaction response", "error", err)
		}
		scheduleCleanup(s, i.Interaction)
		return
	}

	if startedImmediately {
		// The player announces "now playing" itself for this one -- deleting
		// the deferred response instead of editing it avoids showing both
		// that a "queued" message and a "now playing" message for one song.
		if err := s.InteractionResponseDelete(i.Interaction); err != nil {
			slog.Warn("error deleting interaction response", "error", err)
		}
		return
	}

	embed := track.Embed("Added to the queue", embedColorQueued)
	if _, err := s.InteractionResponseEdit(i.Interaction, &discordgo.WebhookEdit{
		Embeds: &[]*discordgo.MessageEmbed{embed},
	}); err != nil {
		slog.Warn("error editing interaction response", "error", err)
	}
	scheduleCleanup(s, i.Interaction)
}

func (p *playCommand) playAudio(i *discordgo.InteractionCreate) (TrackInfo, bool, error) {
	if i.Member == nil {
		return TrackInfo{}, false, errors.New("unable to determine who to play for")
	}

	optionMap := bot.OptionMap(i.ApplicationCommandData().Options)
	opt, ok := optionMap[paramQuery]
	if !ok {
		return TrackInfo{}, false, errors.New("query is a required field")
	}

	return p.player.Play(context.Background(), i.Interaction, opt.StringValue())
}
