package commands

import (
	"disqo/player"
	"fmt"
	"github.com/disgoorg/disgo/discord"
	"github.com/disgoorg/disgo/events"
)

var skip = Command{
	Name:        "skip",
	Description: "Skip to the next song",
	Callback: func(interaction *events.InteractionCreate) {
		_ = deferReply(interaction)
		if interaction.Member() == nil {
			_ = editReply(interaction, discord.MessageUpdate{Content: point("This command must be ran in a server.")})
			return
		}
		channelId := getMemberVoiceChannelId(interaction)
		if channelId == nil {
			_ = editReply(interaction, discord.MessageUpdate{Content: point("You must be in a voice channel.")})
			return
		}
		player := player.GetPlayer(*interaction.GuildID(), interaction.Client().VoiceManager())
		if player.Channel() != *channelId {
			_ = editReply(interaction, discord.MessageUpdate{Content: point("You are not in the same voice channel as me")})
			return
		}

		if player.QueueLen() < 2 {
			_ = editReply(interaction, discord.MessageUpdate{Content: point("There is no song to skip to")})
			return
		}
		go func() {
			err := player.Skip()
			fmt.Println(err)
			_ = editReply(interaction, discord.MessageUpdate{Content: point("⏩ Successfully skipped to the next song.")})
		}()
	},
}
