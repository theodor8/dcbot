package main

import (
	"context"

	"github.com/bwmarrin/discordgo"

	"github.com/openai/openai-go"
	"github.com/openai/openai-go/packages/param"
)

func messageCreate(s *discordgo.Session, m *discordgo.MessageCreate) {

	if m.Author.ID == s.State.User.ID {
		return
	}

	userMessage := openai.UserMessage(m.Content)
	userMessage.OfUser.Name = param.NewOpt(m.Author.Username)
	params.Messages = append(params.Messages, userMessage)

	completion, err := client.Chat.Completions.New(context.TODO(), params)
	if err != nil {
		log.Error("Error calling AI API: ", err)
		return
	}

	completionHandler(completion, s, m)
}
