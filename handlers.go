package main

import (
	"slices"

	"github.com/bwmarrin/discordgo"

	"github.com/openai/openai-go"
	"github.com/openai/openai-go/packages/param"
)

func addChannelMessages(params *openai.ChatCompletionNewParams, s *discordgo.Session, m *discordgo.MessageCreate, n int) error {
	messages, err := s.ChannelMessages(m.ChannelID, n, "", "", "")
	if err != nil {
		return err
	}

	slices.Reverse(messages)

	for _, message := range messages {
		if message.Author.ID == s.State.User.ID {
			params.Messages = append(params.Messages, openai.AssistantMessage(message.Content))
		} else {
			userMessage := openai.UserMessage(message.Content)
			userMessage.OfUser.Name = param.NewOpt(message.Author.Username)
			params.Messages = append(params.Messages, userMessage)
		}
	}

	// TODO: add slash commands to params

	return nil
}

func messageCreate(s *discordgo.Session, m *discordgo.MessageCreate) {

	if m.Author.ID == s.State.User.ID {
		return
	}

	log.Info("message received by ", m.Author.Username, ": ", m.Content)

	params := newParams()
	if err := addChannelMessages(params, s, m, 20); err != nil {
		log.Error("error adding channel messages: ", err)
		return
	}

	s.ChannelTyping(m.ChannelID)

	createCompletion(params, s, m)
}

func interactionCreate(s *discordgo.Session, i *discordgo.InteractionCreate) {
	log.Infof("interaction received by %v", i.Member.User.Username)
	switch i.Type {
	case discordgo.InteractionApplicationCommand:
		if h, ok := commandHandlers[i.ApplicationCommandData().Name]; ok {
			h(s, i)
		}
	case discordgo.InteractionMessageComponent:
		if h, ok := componentHandlers[i.MessageComponentData().CustomID]; ok {
			h(s, i)
		}
	}
}
