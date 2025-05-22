package main

import (
	"context"
	"os"
	"time"

	"github.com/openai/openai-go"
	"github.com/openai/openai-go/option"

	"github.com/bwmarrin/discordgo"
)

var systemMessage = `Du är en discord bot. Du ska svara på frågor och ge information om olika ämnen. Du ska svara kort och koncist. Du formaterar alltid svar för att se tydliga och välstrukturerade ut i Discord. Använd markdown för att göra texten lättläst, men håll det enkelt och undvik komplex formatering som kan se konstigt ut i Discord. Prioritetera läsbarhet. Använd kortfattade svar med tydliga rubriker, punktlistor och kodblock endast när det behövs. Anpassa alltid för Discord:

- ` + "`**fetstil**`" + ` för betoning
- ` + "`*kursiv*`" + ` för subtil betoning eller termer
- ` + "`~~överstruken~~`" + ` för att visa ändringar eller ogiltiga påståenden
- ` + "`> citat`" + ` för att framhäva särskilda delar
- ` + "`` `kod` ``" + ` för enstaka kodsnuttar eller kommandon
- ` + "```\nblockkod\n```" + ` för längre kodexempel eller utdrag

Strukturera svaret tydligt med rubriker (#, ##, ###) och listor (-, *, 1.) där det passar. Håll språket tydligt, koncist och anpassat för Discord.`

var client = openai.NewClient(
	option.WithAPIKey(os.Getenv("API_KEY")),
	option.WithBaseURL(os.Getenv("URL")),
)

var params = openai.ChatCompletionNewParams{
	Model: os.Getenv("MODEL"),
	Messages: []openai.ChatCompletionMessageParamUnion{
		openai.SystemMessage(systemMessage),
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
}

func startStreaming(s *discordgo.Session, m *discordgo.MessageCreate) openai.ChatCompletionAccumulator {
	stream := client.Chat.Completions.NewStreaming(context.TODO(), params)
	acc := openai.ChatCompletionAccumulator{}
	var message *discordgo.Message = nil
	var ticker *time.Ticker = nil
	done := make(chan bool)
	for stream.Next() {
		chunk := stream.Current()
		if !acc.AddChunk(chunk) {
			log.Error("Error adding chunk: ", stream.Err())
		}

		if len(chunk.Choices) == 0 || chunk.Choices[0].Delta.Content == "" || message != nil {
			continue
		}

		var err error
		message, err = s.ChannelMessageSend(m.ChannelID, acc.Choices[0].Message.Content)
		if err != nil {
			log.Error("Error sending message: ", err)
		}
		ticker = time.NewTicker(400 * time.Millisecond)
		go func() {
			log.Info("Ticker started")
			editMessage := func() {
				message, err = s.ChannelMessageEdit(m.ChannelID, message.ID, acc.Choices[0].Message.Content)
				if err != nil {
					log.Error("Error editing message: ", err)
				}
			}
			for {
				select {
				case <-done:
					editMessage()
					ticker.Stop()
					log.Info("Ticker stopped")
					return
				case <-ticker.C:
					editMessage()
				}
			}
		}()

	}
	if ticker != nil {
		done <- true
	}
	if stream.Err() != nil {
		log.Error("Error streaming completion: ", stream.Err())
	}
	return acc
}

func createCompletion(s *discordgo.Session, m *discordgo.MessageCreate) {
	completion := startStreaming(s, m)

	log.Infof("Prompt tokens: %v, completion tokens: %v", completion.Usage.PromptTokens, completion.Usage.CompletionTokens)

	params.Messages = append(params.Messages, completion.Choices[0].Message.ToParam())

	toolCalls := completion.Choices[0].Message.ToolCalls
	if len(toolCalls) > 0 {
		toolCallsHandler(toolCalls, s, m)
		return
	}
}
