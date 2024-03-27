package commands

import (
	"github.com/disgoorg/disgo/bot"
	"github.com/disgoorg/disgo/discord"
	"github.com/disgoorg/disgo/events"
)

type Command struct {
	Name, Description string
	Options           []discord.ApplicationCommandOption
	Callback          func(interaction *events.InteractionCreate)
}

var commands = []Command{
	play, stop, skip, pause, resume, queue,
}

func data(i *events.InteractionCreate) discord.SlashCommandInteractionData {
	return i.Interaction.(discord.ApplicationCommandInteraction).Data.(discord.SlashCommandInteractionData)
}

func deferReply(interaction *events.InteractionCreate) error {
	return interaction.Respond(discord.InteractionResponseTypeDeferredCreateMessage, nil)
}

func editReply(interaction *events.InteractionCreate, msg discord.MessageUpdate) error {
	_, err := interaction.Client().Rest().
		UpdateInteractionResponse(interaction.Client().ID(), interaction.Token(), msg)
	return err
}

func Handle(interaction *events.InteractionCreate) {
	if i, ok := interaction.Interaction.(discord.ApplicationCommandInteraction); !ok {
		return
	} else {
		var command *Command
		for _, cmd := range commands {
			if cmd.Name == i.Data.CommandName() {
				command = &cmd
				break
			}
		}
		if command == nil {
			return
		}

		command.Callback(interaction)
	}
}

func RegisterCommands(c bot.Client) error {
	cmds := make([]discord.ApplicationCommandCreate, len(commands))
	for i, cmd := range commands {
		cmds[i] = discord.SlashCommandCreate{Name: cmd.Name, Description: cmd.Description, Options: cmd.Options}
	}
	_, err := c.Rest().SetGlobalCommands(c.ID(), cmds)
	return err
}
