package main

import (
	"os"
	"os/signal"
	"syscall"

	"github.com/bwmarrin/discordgo"
	"github.com/sirupsen/logrus"
)

var log = logrus.New()
var s *discordgo.Session = nil
var stop chan os.Signal

func init() {
	log.SetFormatter(&logrus.TextFormatter{
		ForceColors: true,
	})

	token := os.Getenv("DISCORD_TOKEN")
	if token == "" {
		log.Fatal("no DISCORD_TOKEN environment variable found")
	}

	var err error
	s, err = discordgo.New("Bot " + token)
	if err != nil {
		log.Fatal("error creating Discord session:", err)
	}
	s.Identify.Intents = discordgo.IntentsGuildMessages
	log.Info("created Discord session")

	s.AddHandler(messageCreate)
	s.AddHandler(interactionCreate)
	s.AddHandler(func(s *discordgo.Session, r *discordgo.Ready) {
		log.Printf("logged in as: %v#%v", s.State.User.Username, s.State.User.Discriminator)
	})
	log.Info("added handlers")
}

func main() {

	err := s.Open()
	if err != nil {
		log.Fatal("error opening Discord connection:", err)
	}
	defer s.Close()
	log.Info("opened Discord connection")

	for _, v := range commands {
		_, err := s.ApplicationCommandCreate(s.State.User.ID, "", v)
		if err != nil {
			log.Panicf("cannot create '%v' command: %v", v.Name, err)
		}
	}
	log.Info("added commands")

	stop = make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM, os.Interrupt)
	<-stop
	log.Info("gracefully shutting down...")
}
