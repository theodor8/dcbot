package main

import (
	"encoding/json"
	"io"
	"net/http"

	"github.com/bwmarrin/discordgo"

	"github.com/openai/openai-go"
)

func toolCallsHandler(params *openai.ChatCompletionNewParams, toolCalls []openai.ChatCompletionMessageToolCall, s *discordgo.Session, m *discordgo.MessageCreate) {
	for _, toolCall := range toolCalls {
		var args map[string]any
		if err := json.Unmarshal([]byte(toolCall.Function.Arguments), &args); err != nil {
			log.Error("error unmarshalling tool call arguments: ", err)
			return
		}

		log.Info("toolcall: ", toolCall.Function.Name, ", args: ", args)

		toolCallResponse := ""

		// TODO: toolcall to execute shell command
		// TODO: make map with toolcall handlers

		switch toolCall.Function.Name {
		case "do_nothing":
			toolCallResponse = "No action taken."
		case "set_status":
			status := args["status"].(string)
			s.UpdateCustomStatus(status)
			toolCallResponse = "Status updated successfully."
		case "get_sender_username":
			toolCallResponse = m.Author.Username
		case "set_username":
			username := args["username"].(string)
			s.UserUpdate(username, "")
			toolCallResponse = "Username updated successfully."
		case "http_request":
			// TODO: confirm url with buttons
			url := args["url"].(string)
			resp, err := http.Get(url)
			if err != nil {
				toolCallResponse = "Error making HTTP request: " + err.Error()
				log.Error("error making HTTP request: ", err)
				break
			}
			defer resp.Body.Close()

			bodyBytes, err := io.ReadAll(resp.Body)
			if err != nil {
				toolCallResponse = "Error reading HTTP response: " + err.Error()
				log.Error("error reading HTTP response: ", err)
				break
			}
			toolCallResponse = string(bodyBytes)
		default:
			toolCallResponse = "Unknown tool call"
			log.Error("unknown tool call:", toolCall.Function.Name)
		}

		log.Info("toolcall response: ", toolCallResponse)

		params.Messages = append(params.Messages, openai.ToolMessage(toolCallResponse, toolCall.ID))
	}

	createCompletion(params, s, m)
}
