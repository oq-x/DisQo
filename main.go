package main

import (
	"context"
	"disqo/commands"
	"disqo/log"
	"disqo/youtube"
	"github.com/disgoorg/disgo"
	"github.com/disgoorg/disgo/bot"
	"github.com/disgoorg/disgo/cache"
	"github.com/disgoorg/disgo/events"
	"github.com/disgoorg/disgo/gateway"
	"github.com/fatih/color"
	"github.com/joho/godotenv"
	"os"
	"os/signal"
	"syscall"
)

func main() {
	color.Blue("    .___.__                     \n  __| _/|__| ____________ ____  \n / __ | |  |/  ___/ ____//  _ \\ \n/ /_/ | |  |\\___ < <_|  (  <_> )\n\\____ | |__/____  >__   |\\____/ \n     \\/         \\/   |__|       ")
	log.INFO("made by oq - https://github.com/oq-x")
	_ = godotenv.Load()
	token := os.Getenv("TOKEN")
	if token == "" {
		log.ERROR("no token found in env!")
		return
	}
	youtubeKey := os.Getenv("YT_API_KEY")
	if youtubeKey == "" {
		log.ERROR("a youtube api key was not found in env!")
		return
	}
	if err := youtube.InitService(youtubeKey); err != nil {
		log.ERRORf("Error creating youtube service: %v", err)
		return
	}

	client, err := disgo.New(token,
		bot.WithGatewayConfigOpts(gateway.WithIntents(gateway.IntentGuilds|gateway.IntentGuildVoiceStates)),
		bot.WithCacheConfigOpts(
			cache.WithCaches(cache.FlagVoiceStates),
		),
		bot.WithEventListenerFunc(func(e *events.Ready) {
			log.INFO("VinylCord is running!")
			_ = commands.RegisterCommands(e.Client())
		}),
		bot.WithEventListenerFunc(commands.Handle),
	)

	if err != nil {
		log.ERROR("Error running bot: ", err)
		return
	}
	// connect to the gateway
	if err = client.OpenGateway(context.TODO()); err != nil {
		log.ERROR("Error running bot: ", err)
	}

	s := make(chan os.Signal, 1)
	signal.Notify(s, syscall.SIGINT, syscall.SIGTERM)
	<-s
}
