package main

import (
	"log"
	"os"
	"time"

	"gopkg.in/telebot.v3"
)

func main() {
	pref := telebot.Settings{
		Token:  mustToken(),
		Poller: &telebot.LongPoller{Timeout: 10 * time.Second},
	}

	bot, err := telebot.NewBot(pref)
	if err != nil {
		log.Fatal(err)
	}

	bot.Handle(telebot.OnText, func(c telebot.Context) error {
		return c.Send(c.Text())
	})

	log.Println("Bot started")
	bot.Start()
}

func mustToken() string {
	token := os.Getenv("TELEGRAM_BOT_TOKEN")
	if token == "" {
		log.Fatal("TELEGRAM_BOT_TOKEN is not set")
	}
	return token
}
