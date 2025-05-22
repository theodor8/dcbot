package main

import (
	"encoding/json"

	"github.com/bwmarrin/discordgo"

	"github.com/openai/openai-go"
)

func toolCallsHandler(params *openai.ChatCompletionNewParams, toolCalls []openai.ChatCompletionMessageToolCall, s *discordgo.Session, m *discordgo.MessageCreate) {
	for _, toolCall := range toolCalls {
		var args map[string]interface{}
		if err := json.Unmarshal([]byte(toolCall.Function.Arguments), &args); err != nil {
			log.Error("Error unmarshalling tool call arguments: ", err)
			return
		}

		toolCallResponse := ""

		switch toolCall.Function.Name {
		case "do_nothing":
			log.Info("Doing nothing")
		case "set_status":
			status := args["status"].(string)
			log.Info("Setting status to: ", status)
			s.UpdateCustomStatus(status)
			toolCallResponse = "Status updated successfully."
		case "get_sender_username":
			log.Info("Getting sender username: ", m.Author.Username)
			toolCallResponse = m.Author.Username
		case "set_username":
			username := args["username"].(string)
			log.Info("Setting username to: ", username)
			s.UserUpdate(username, "")
			toolCallResponse = "Username updated successfully."
		default:
			toolCallResponse = "Unknown tool call"
			log.Error("Error: unknown tool call:", toolCall.Function.Name)
		}

		params.Messages = append(params.Messages, openai.ToolMessage(toolCallResponse, toolCall.ID))
	}

	createCompletion(params, s, m)
}
