package commands

import (
	"disqo/player"
	"disqo/youtube"
	"fmt"
	"github.com/disgoorg/disgo/discord"
	"github.com/disgoorg/disgo/events"
	"github.com/disgoorg/snowflake/v2"
)

func point[T any](t T) *T {
	return &t
}

var play = Command{
	Name: "play",
	Options: []discord.ApplicationCommandOption{
		discord.ApplicationCommandOptionString{Name: "query", Description: "The song to search", Required: true},
	},
	Description: "Play a song from YouTube.",
	Callback: func(interaction *events.InteractionCreate) {
		_ = deferReply(interaction)
		d := data(interaction)
		query := d.String("query")
		var id string
		if i, err := youtube.GetVideoIDFromURL(query); err == nil {
			id = i
		} else {
			videos, err := youtube.SearchVideos(query)
			if err != nil {
				_ = editReply(interaction, discord.MessageUpdate{Content: point(fmt.Sprintf("Failed to search videos: %v", err))})
				return
			}
			id = videos.Items[0].Id.VideoId
		}

		vid, stream, _ := youtube.GetVideo(id)
		duration, _ := youtube.ParseVideoDuration(vid.ContentDetails.Duration)
		if vid == nil {
			_ = editReply(interaction, discord.MessageUpdate{Content: point("No video found.")})
			return
		}

		if interaction.Member() == nil {
			_ = editReply(interaction, discord.MessageUpdate{Content: point("This command must be ran in a server.")})
			return
		}
		channelId := getMemberVoiceChannelId(interaction)
		if channelId == nil {
			_ = editReply(interaction, discord.MessageUpdate{Content: point("You must be in a voice channel.")})
			return
		}
		song := player.NewSong(stream, vid.Snippet)
		player := player.GetPlayer(*interaction.GuildID(), interaction.Client().VoiceManager())
		if !player.Connected() {
			go func() {
				if err := player.Connect(*channelId); err != nil {
					_ = editReply(interaction, discord.MessageUpdate{Content: point(fmt.Sprintf("Error connecting: %v", err))})
				}
				if err := player.Play(song); err != nil {
					player.Kill()
					_ = editReply(interaction, discord.MessageUpdate{Content: point(fmt.Sprintf("Error playing: %v", err))})
					return
				} else {
					embed := discord.NewEmbedBuilder().
						SetTitlef("Now playing: %s", vid.Snippet.Title).
						SetURLf("https://youtu.be/%s", vid.Id).
						SetThumbnail(vid.Snippet.Thumbnails.Default.Url).
						SetDescriptionf("Duration: %s", duration.String()).
						SetAuthorName(vid.Snippet.ChannelTitle).Build()
					_ = editReply(interaction, discord.MessageUpdate{Embeds: &[]discord.Embed{embed}})
				}
			}()
		} else {
			if player.Playing() {
				player.AddToQueue(song)
				embed := discord.NewEmbedBuilder().
					SetTitlef("Added to queue: %s", vid.Snippet.Title).
					SetURLf("https://youtu.be/%s", vid.Id).
					SetThumbnail(vid.Snippet.Thumbnails.Default.Url).
					SetDescriptionf("Duration: %s", duration.String()).
					SetAuthorName(vid.Snippet.ChannelTitle).Build()
				_ = editReply(interaction, discord.MessageUpdate{Embeds: &[]discord.Embed{embed}})
			} else {
				if err := player.Play(song); err != nil {
					player.Kill()
					_ = editReply(interaction, discord.MessageUpdate{Content: point(fmt.Sprintf("Error playing: %v", err))})
					return
				} else {
					embed := discord.NewEmbedBuilder().
						SetTitlef("Now playing: %s", vid.Snippet.Title).
						SetURLf("https://youtu.be/%s", vid.Id).
						SetThumbnail(vid.Snippet.Thumbnails.Default.Url).
						SetDescriptionf("Duration: %s", duration.String()).
						SetAuthorName(vid.Snippet.ChannelTitle).Build()
					_ = editReply(interaction, discord.MessageUpdate{Embeds: &[]discord.Embed{embed}})
				}
			}
		}
	},
}

func getMemberVoiceChannelId(interaction *events.InteractionCreate) *snowflake.ID {
	state, _ := interaction.Client().Caches().VoiceState(*interaction.GuildID(), interaction.User().ID)
	return state.ChannelID
}
