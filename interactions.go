package main

import (
	"context"
	"os"
	"os/exec"
	"sync"
	"syscall"
	"time"

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
		s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseDeferredChannelMessageWithSource,
		})

		cmd := optionMap["cmd"].StringValue()
		if !confirmAction(s, i.ChannelID, "Execute command `"+cmd+"`?") {
			out := "Command execution canceled."
			s.InteractionResponseEdit(i.Interaction, &discordgo.WebhookEdit{
				Content: &out,
			})
			return
		}
		out, err := execCommand(cmd)
		if err != nil {
			log.Error("error executing command: ", err.Error())
		}
		out = "```\n" + out + "\n```"
		s.InteractionResponseEdit(i.Interaction, &discordgo.WebhookEdit{
			Content: &out,
		})

	},
}

var componentHandlers = map[string]func(s *discordgo.Session, i *discordgo.InteractionCreate){
	"confirm_action": func(s *discordgo.Session, i *discordgo.InteractionCreate) {
		s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseUpdateMessage,
			Data: &discordgo.InteractionResponseData{
				Content:    "Action confirmed ✅",
			},
		})
		go func() {
			<-time.After(1 * time.Second)
			err := s.ChannelMessageDelete(i.ChannelID, i.Message.ID)
			if err != nil {
				log.Error("error deleting confirmation message: ", err)
			}
		}()

		confirmedActions.Lock()
		if ch, ok := confirmedActions.actions[i.Message.ID]; ok {
			ch <- true
			close(ch)
			delete(confirmedActions.actions, i.Message.ID)
		} else {
			log.Warn("no confirmation channel found for message ID: ", i.Message.ID)
		}
		confirmedActions.Unlock()
	},
	"cancel_action": func(s *discordgo.Session, i *discordgo.InteractionCreate) {
		s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseUpdateMessage,
			Data: &discordgo.InteractionResponseData{
				Content:    "Action canceled ❌",
			},
		})
		go func() {
			<-time.After(1 * time.Second)
			err := s.ChannelMessageDelete(i.ChannelID, i.Message.ID)
			if err != nil {
				log.Error("error deleting confirmation message: ", err)
			}
		}()

		confirmedActions.Lock()
		if ch, ok := confirmedActions.actions[i.Message.ID]; ok {
			ch <- false
			close(ch)
			delete(confirmedActions.actions, i.Message.ID)
		} else {
			log.Warn("no confirmation channel found for message ID: ", i.Message.ID)
		}
		confirmedActions.Unlock()
	},
}

var confirmedActions = struct {
	sync.RWMutex
	actions map[string]chan bool
}{actions: make(map[string]chan bool)}

func confirmAction(s *discordgo.Session, channelID string, action string) bool {
	st, err := s.ChannelMessageSendComplex(channelID, &discordgo.MessageSend{
		Content: action,
		Components: []discordgo.MessageComponent{
			discordgo.ActionsRow{
				Components: []discordgo.MessageComponent{
					discordgo.Button{
						Label:    "Confirm",
						Style:    discordgo.SuccessButton,
						CustomID: "confirm_action",
					},
					discordgo.Button{
						Label:    "Cancel",
						Style:    discordgo.DangerButton,
						CustomID: "cancel_action",
					},
				},
			},
		},
	})
	if err != nil {
		log.Error("error sending confirmation message: ", err)
		return false
	}
	confirmedActions.Lock()
	confirmedActions.actions[st.ID] = make(chan bool)
	confirmedActions.Unlock()

	return <-confirmedActions.actions[st.ID]
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
