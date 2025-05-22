package main

import (
	"slices"

	"github.com/bwmarrin/discordgo"

	"github.com/openai/openai-go"
	"github.com/openai/openai-go/packages/param"
)

func messageCreate(s *discordgo.Session, m *discordgo.MessageCreate) {

	if m.Author.ID == s.State.User.ID {
		return
	}

	messages, err := s.ChannelMessages(m.ChannelID, 20, "", "", "")
	if err != nil {
		log.Error("Error fetching messages: ", err)
		return
	}
	slices.Reverse(messages)

	params := newParams()

	for _, message := range messages {
		if message.Author.ID == s.State.User.ID {
			params.Messages = append(params.Messages, openai.AssistantMessage(message.Content))
		} else {
			userMessage := openai.UserMessage(message.Content)
			userMessage.OfUser.Name = param.NewOpt(message.Author.Username)
			params.Messages = append(params.Messages, userMessage)
		}
	}

	s.ChannelTyping(m.ChannelID)

	createCompletion(params, s, m)
}
