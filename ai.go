package main

import (
	"os"

	"github.com/openai/openai-go"
	"github.com/openai/openai-go/option"

	"github.com/bwmarrin/discordgo"
)

var client = openai.NewClient(
	option.WithAPIKey(os.Getenv("API_KEY")),
	option.WithBaseURL(os.Getenv("URL")),
)

var params = openai.ChatCompletionNewParams{
	Messages: []openai.ChatCompletionMessageParamUnion{
		openai.SystemMessage("Du är en discord bot. Du ska svara på frågor och ge information om olika ämnen. Du ska vara hjälpsam och vänlig. Du ska inte ge några personliga åsikter eller känslor. Du ska alltid svara på svenska. Om du inte vet svaret på en fråga ska du säga att du inte vet. Du ska svara kort och koncist. Du ska bara svara om meddelandet är avsett för dig. Om du inte förväntas svara ska du göra inget."),
	},
	Tools: []openai.ChatCompletionToolParam{
		{
			Function: openai.FunctionDefinitionParam{
				Name:        "do_nothing",
				Description: openai.String("Do nothing"),
			},
		},
		{
			Function: openai.FunctionDefinitionParam{
				Name:        "set_status",
				Description: openai.String("Set your status"),
				Parameters: openai.FunctionParameters{
					"type": "object",
					"properties": map[string]interface{}{
						"status": map[string]string{
							"type": "string",
						},
					},
					"required": []string{"status"},
				},
			},
		},
		{
			Function: openai.FunctionDefinitionParam{
				Name:        "set_username",
				Description: openai.String("Set your username"),
				Parameters: openai.FunctionParameters{
					"type": "object",
					"properties": map[string]interface{}{
						"username": map[string]string{
							"type": "string",
						},
					},
					"required": []string{"username"},
				},
			},
		},
		{
			Function: openai.FunctionDefinitionParam{
				Name:        "get_sender_username",
				Description: openai.String("Get the username of the sender"),
			},
		},
	},
	Model: os.Getenv("MODEL"),
}

func completionHandler(completion *openai.ChatCompletion, s *discordgo.Session, m *discordgo.MessageCreate) {
	log.Infof("Prompt tokens: %v, completion tokens: %v", completion.Usage.PromptTokens, completion.Usage.CompletionTokens)

	if len(completion.Choices) != 1 {
		log.Error("Error: expected 1 choice, got ", len(completion.Choices))
	}

	params.Messages = append(params.Messages, completion.Choices[0].Message.ToParam())

	toolCalls := completion.Choices[0].Message.ToolCalls
	if len(toolCalls) > 0 {
		toolCallsHandler(toolCalls, s, m)
		return
	}

	s.ChannelTyping(m.ChannelID)
	s.ChannelMessageSend(m.ChannelID, completion.Choices[0].Message.Content)
}
