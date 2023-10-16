package main

import (
	"log"
	"context"
	"fmt"
	"os"	
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	openai "github.com/sashabaranov/go-openai"
)

var client *openai.Client
var req = make(map[string]openai.ChatCompletionRequest)

func Init() {
	client = openai.NewClient(os.Getenv("OPENAI_API_KEY"))	
}

func gptRequest(userName string, message string) string {	
	var result string = ""

	var userReq, ok = req[userName] 

	if (!ok) {
		userReq = openai.ChatCompletionRequest{
			Model: openai.GPT3Dot5Turbo,
			Messages: []openai.ChatCompletionMessage{
				{
					Role:    openai.ChatMessageRoleSystem,
					Content: "you are a helpful chatbot",
				},
			},
		}
	}

	fmt.Printf("request: %v\n", message)
	userReq.Messages = append(userReq.Messages, openai.ChatCompletionMessage{
		Role:    openai.ChatMessageRoleUser,
		Content: message,
	})

	resp, err := client.CreateChatCompletion(context.Background(), userReq)

	if err != nil {
		fmt.Printf("ChatCompletion error: %v\n", err)
		return err.Error()
	}

	result = resp.Choices[0].Message.Content
	fmt.Println(resp.Choices[0].Message.Content)
	userReq.Messages = append(userReq.Messages, resp.Choices[0].Message)
	req[userName] = userReq
	return result
}

func main() {
	Init()

	bot, err := tgbotapi.NewBotAPI(os.Getenv("TELEGRAM_API_KEY"))
	if err != nil {
		log.Panic(err)
	}
	
	bot.Debug = true

	log.Printf("Authorized on account %s", bot.Self.UserName)

	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60

	updates := bot.GetUpdatesChan(u)

	for update := range updates {
		if update.Message != nil { // If we got a message
			log.Printf("[%s] %s", update.Message.From.UserName, update.Message.Text)
			gptResponse := gptRequest(update.Message.From.UserName, update.Message.Text)
			msg := tgbotapi.NewMessage(update.Message.Chat.ID, gptResponse)
			msg.ReplyToMessageID = update.Message.MessageID

			bot.Send(msg)
		}
	}
}