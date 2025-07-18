package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"
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
		text := c.Text()
		if isAudioURL(text) {
			origMsg := c.Message()
			go sendToWorker(text, origMsg.ID)
			return c.Reply("Аудио получено, ожидайте результат!", &telebot.SendOptions{
				ReplyTo: origMsg,
			})
		}
		return c.Send("Пожалуйста, пришлите ссылку на аудиофайл.")
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

func isAudioURL(text string) bool {
	audioExtensions := []string{".mp3", ".wav", ".ogg", ".m4a"}
	for _, ext := range audioExtensions {
		if strings.HasSuffix(strings.ToLower(text), ext) &&
			(strings.HasPrefix(text, "http://") || strings.HasPrefix(text, "https://")) {
			return true
		}
	}
	return false
}

func sendToWorker(audioURL string, messageID int) {
	workerURL := os.Getenv("WORKER_URL")
	if workerURL == "" {
		log.Println("WORKER_URL is not set")
		return
	}

	body := strings.NewReader(
		`{"audio_url":"` + audioURL + `","message_id":"` + toStr(messageID) + `"}`,
	)
	resp, err := http.Post(workerURL+"/transcribe", "application/json", body)
	if err != nil {
		log.Printf("Sent to worker error: %v", err)
		return
	}
	defer resp.Body.Close()
	log.Printf("Воркер принял ссылку: %s", resp)
}

// Перевод int в string
func toStr(i int) string {
	return fmt.Sprintf("%d", i)
}
