package commands

import (
	"disqo/player"
	"github.com/disgoorg/disgo/discord"
	"github.com/disgoorg/disgo/events"
)

var stop = Command{
	Name:        "stop",
	Description: "Stop the music",
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
		go player.Kill()
		_ = editReply(interaction, discord.MessageUpdate{Content: point("⏹ Stopped the music successfully.")})
	},
}
