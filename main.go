package main

import (
	"os"
	"os/signal"
	"syscall"

	"github.com/bwmarrin/discordgo"
	"github.com/sirupsen/logrus"
)

var log = logrus.New()
var dg *discordgo.Session = nil

func main() {

	log.SetFormatter(&logrus.TextFormatter{
		ForceColors: true,
	})

	token := os.Getenv("DISCORD_TOKEN")
	if token == "" {
		log.Fatal("No token provided. Please set the DISCORD_TOKEN environment variable.")
	}

	var err error
	dg, err = discordgo.New("Bot " + token)
	if err != nil {
		log.Fatal("error creating Discord session:", err)
	}

	dg.AddHandler(func(s *discordgo.Session, r *discordgo.Ready) {
		log.Printf("Logged in as: %v#%v", s.State.User.Username, s.State.User.Discriminator)
	})

	dg.AddHandler(messageCreate)

	dg.Identify.Intents = discordgo.IntentsGuildMessages

	err = dg.Open()
	if err != nil {
		log.Fatal("error opening connection:", err)
	}
	defer dg.Close()

	sc := make(chan os.Signal, 1)
	signal.Notify(sc, syscall.SIGINT, syscall.SIGTERM, os.Interrupt)
	<-sc
	log.Info("Gracefully shutting down...")
}
