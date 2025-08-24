package main

import (
	"context"
	"os"
	"os/exec"
	"syscall"

	"github.com/bwmarrin/discordgo"
)

var commands = []*discordgo.ApplicationCommand{
	{
		Name:        "exit",
		Description: "Exits the bot",
	},
	{
		Name:        "bal",
		Description: "Check your balance",
	},
	{
		Name:        "exec",
		Description: "Execute a command",
		Options: []*discordgo.ApplicationCommandOption{
			{
				Type:        discordgo.ApplicationCommandOptionString,
				Name:        "cmd",
				Description: "The command to execute",
				Required:    true,
			},
		},
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
	"bal": func(s *discordgo.Session, i *discordgo.InteractionCreate) {
		s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{
				Content: getUserBalance(),
			},
		})
	},
	"exec": func(s *discordgo.Session, i *discordgo.InteractionCreate) {
		options := i.ApplicationCommandData().Options
		optionMap := make(map[string]*discordgo.ApplicationCommandInteractionDataOption, len(options))
		for _, opt := range options {
			optionMap[opt.Name] = opt
		}

		cmd := optionMap["cmd"].StringValue()
		out, err := execCommand(cmd)
		if err != nil {
			log.Error("error executing command: ", err.Error())
		}

		s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{
				Content: "```\n" + string(out) + "\n```",
			},
		})
	},
}

func execCommand(command string) (string, error) {
	// TODO: make executing commands work with docker containers
	cmd := exec.Command("bash", "-c", command)
	out, err := cmd.CombinedOutput()
	return string(out), err
}

type getUserBalanceResponse struct {
	BalanceInfos []struct {
		Currency     string `json:"currency"`
		TotalBalance string `json:"total_balance"`
	} `json:"balance_infos"`
}

func getUserBalance() string {
	res := &getUserBalanceResponse{}
	err := client.Get(context.TODO(), os.Getenv("URL")+"user/balance", nil, &res)
	if err != nil {
		log.Error("error getting user balance: ", err)
		return "Error fetching balance"
	}
	return res.BalanceInfos[0].TotalBalance + " " + res.BalanceInfos[0].Currency
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
