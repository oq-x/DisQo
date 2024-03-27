package commands

import (
	"disqo/player"
	"fmt"
	"github.com/disgoorg/disgo/discord"
	"github.com/disgoorg/disgo/events"
	"math"
)

var queue = Command{
	Name:        "queue",
	Description: "Show the queue for this server",
	Options: []discord.ApplicationCommandOption{
		discord.ApplicationCommandOptionInt{
			Name:        "page",
			Description: "The page of the queue",
		},
	},
	Callback: func(interaction *events.InteractionCreate) {
		_ = deferReply(interaction)
		if interaction.GuildID() == nil {
			_ = editReply(interaction, discord.MessageUpdate{Content: point("This command must be ran in a server.")})
			return
		}
		if !player.HasPlayer(*interaction.GuildID()) {
			_ = editReply(interaction, discord.MessageUpdate{Content: point("Nothing is playing.")})
			return
		}
		player := player.GetPlayer(*interaction.GuildID(), interaction.Client().VoiceManager())
		queue := player.Queue()
		if len(queue) == 0 {
			_ = editReply(interaction, discord.MessageUpdate{Content: point("Nothing is playing.")})
			return
		}
		pages := int(math.Ceil(float64(len(queue)) / 10))
		page := data(interaction).Int("page")
		if page == 0 {
			page = 1
		}

		if pages < page {
			page = pages
		}

		startIndex := (page - 1) * 10
		endIndex := page * 10

		if endIndex > len(queue) {
			endIndex = len(queue)
		}

		var description string

		for i, song := range queue {
			if i < startIndex || i > endIndex-1 {
				continue
			}
			description += fmt.Sprintf("**%d.** [%s - %s](https://youtu.be/%s)", i+1, song.Snippet.Title, song.Snippet.Author, song.Snippet.ID)
			if i == 0 {
				description += " [Now Playing]"
			}
			description += "\n"
		}

		embed := discord.NewEmbedBuilder().
			SetTitle("Queue").
			SetDescription(description).
			SetFooterTextf("Page %d/%d", page, pages).
			Build()

		_ = editReply(interaction, discord.MessageUpdate{Embeds: &[]discord.Embed{embed}})
	},
}
