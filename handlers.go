package main

import (
	"github.com/bwmarrin/discordgo"

	"github.com/openai/openai-go"
	"github.com/openai/openai-go/packages/param"
)

func messageCreate(s *discordgo.Session, m *discordgo.MessageCreate) {

	if m.Author.ID == s.State.User.ID {
		return
	}

	s.ChannelTyping(m.ChannelID)

	userMessage := openai.UserMessage(m.Content)
	userMessage.OfUser.Name = param.NewOpt(m.Author.Username)
	params.Messages = append(params.Messages, userMessage)

	createCompletion(s, m)
}
