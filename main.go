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

var commands = []*discordgo.ApplicationCommand{
	{
		Name:        "exit",
		Description: "Exits the bot",
	},
}
var commandHandlers = map[string]func(s *discordgo.Session, i *discordgo.InteractionCreate){
	"exit": func(s *discordgo.Session, i *discordgo.InteractionCreate) {
		log.Info("exiting bot...")
		s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{
				Content: "Exiting bot...",
			},
		})
		stop <- syscall.SIGINT
	},
}

func deleteCommands(s *discordgo.Session) {
	log.Info("deleting all application commands...")
	commands, err := s.ApplicationCommands(s.State.User.ID, "")
	if err != nil {
		log.Error("error fetching commands:", err)
		return
	}

	for _, cmd := range commands {
		err = s.ApplicationCommandDelete(s.State.User.ID, "", cmd.ID)
		if err != nil {
			log.Error("error deleting command:", err)
		} else {
			log.Infof("deleted command: %s", cmd.Name)
		}
	}
}

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

	s.AddHandler(messageCreate)

	s.AddHandler(func(s *discordgo.Session, i *discordgo.InteractionCreate) {
		if handler, ok := commandHandlers[i.ApplicationCommandData().Name]; ok {
			handler(s, i)
		}
	})
}

func main() {
	s.AddHandler(func(s *discordgo.Session, r *discordgo.Ready) {
		log.Printf("logged in as: %v#%v", s.State.User.Username, s.State.User.Discriminator)
	})

	err := s.Open()
	if err != nil {
		log.Fatal("error opening Discord connection:", err)
	}
	defer s.Close()

	log.Info("adding commands...")
	for _, v := range commands {
		_, err := s.ApplicationCommandCreate(s.State.User.ID, "", v)
		if err != nil {
			log.Panicf("cannot create '%v' command: %v", v.Name, err)
		}
	}

	stop = make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM, os.Interrupt)
	<-stop
	log.Info("gracefully shutting down...")
	// deleteCommands(s)
}
