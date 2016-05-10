package main

import (
	"log"
	"os"

	"github.com/go-telegram-bot-api/telegram-bot-api"
)

const tokenEnv string = "TOKEN"
const ipEnv string = "OPENSHIFT_GO_IP"
const portEnv string = "OPENSHIFT_GO_PORT"

type State uint

const (
	InitState   = 0
	ConfigState = 1
	ReadyState  = 2
)

func main() {

	token := os.Getenv(tokenEnv)
	if token == "" {
		log.Panic("TOKEN ENV NOT FOUND!")
	}

	bot, err := tgbotapi.NewBotAPI(token)
	if err != nil {
		log.Panic(err)
	}

	// Start this webserver just to never puts this instance idle
	StartWebServer()

	bot.Debug = false

	log.Printf("Authorized on account %s\n", bot.Self.UserName)

	news := tgbotapi.NewUpdate(0)
	news.Timeout = 60

	updates, err := bot.GetUpdatesChan(news)

	for update := range updates {

		//log.Printf("[%s] %s", update.Message.From.UserName, update.Message.Text)

		// This bot doesn't answare for inline queries
		if update.InlineQuery != nil {
			continue
		}

		u := GetUser(update)

		// Block if user not in session
		if u == nil {
			u = NewUser(bot, update)
			u.State = InitState
		}

		// Handle the actual command
		go Handle(u, update)

	}

}
