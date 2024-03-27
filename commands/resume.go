package commands

import (
	"disqo/player"
	"github.com/disgoorg/disgo/discord"
	"github.com/disgoorg/disgo/events"
)

var resume = Command{
	Name:        "resume",
	Description: "Resume the current song",
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
		if !player.Paused() {
			_ = editReply(interaction, discord.MessageUpdate{Content: point("The song is not paused")})
			return
		}
		player.Resume()
		_ = editReply(interaction, discord.MessageUpdate{Content: point("The song has been resumed")})
	},
}
