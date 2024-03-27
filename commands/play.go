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
		song := player.NewSong(stream, vid)
		player := player.GetPlayer(*interaction.GuildID(), interaction.Client().VoiceManager())
		if !player.Connected() {
			go func() {
				if err := player.Connect(*channelId, interaction.Channel().ID()); err != nil {
					_ = editReply(interaction, discord.MessageUpdate{Content: point(fmt.Sprintf("Error connecting: %v", err))})
				}
				if err := player.Play(song); err != nil {
					player.Kill()
					_ = editReply(interaction, discord.MessageUpdate{Content: point(fmt.Sprintf("Error playing: %v", err))})
					return
				} else {
					embed := discord.NewEmbedBuilder().
						SetTitlef("Now playing: %s", vid.Title).
						SetURLf("https://youtu.be/%s", vid.ID).
						SetThumbnail(vid.Thumbnails[0].URL).
						SetDescriptionf("Duration: %s", vid.Duration.String()).
						SetAuthorName(vid.Author).Build()
					_ = editReply(interaction, discord.MessageUpdate{Embeds: &[]discord.Embed{embed}})

					<-player.NowPlaying

					for s := range player.NowPlaying {
						embed := discord.NewEmbedBuilder().
							SetTitlef("Now playing: %s", s.Snippet.Title).
							SetURLf("https://youtu.be/%s", s.Snippet.ID).
							SetThumbnail(s.Snippet.Thumbnails[0].URL).
							SetDescriptionf("Duration: %s", s.Snippet.Duration.String()).
							SetAuthorName(s.Snippet.Author).Build()
						_, _ = interaction.Client().Rest().CreateMessage(interaction.Channel().ID(), discord.MessageCreate{Embeds: []discord.Embed{embed}})
					}
				}
			}()
		} else {
			player.AddToQueue(song)
			embed := discord.NewEmbedBuilder().
				SetTitlef("Added to queue: %s", vid.Title).
				SetURLf("https://youtu.be/%s", vid.ID).
				SetThumbnail(vid.Thumbnails[0].URL).
				SetDescriptionf("Duration: %s", vid.Duration.String()).
				SetAuthorName(vid.Author).Build()
			_ = editReply(interaction, discord.MessageUpdate{Embeds: &[]discord.Embed{embed}})
		}
	},
}

func getMemberVoiceChannelId(interaction *events.InteractionCreate) *snowflake.ID {
	state, _ := interaction.Client().Caches().VoiceState(*interaction.GuildID(), interaction.User().ID)
	return state.ChannelID
}
