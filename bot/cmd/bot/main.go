package main

import (
	"encoding/json"
	"fmt"
	"github.com/nzahar/meetings_transcription/shared"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/phpdave11/gofpdf"
	"gopkg.in/telebot.v3"
)

var botInstance *telebot.Bot
var allowedUsers map[string]bool

func main() {
	initAllowedUsers()

	pref := telebot.Settings{
		Token:  mustToken(),
		Poller: &telebot.LongPoller{Timeout: 10 * time.Second},
	}

	bot, err := telebot.NewBot(pref)
	if err != nil {
		log.Fatal(err)
	}
	botInstance = bot

	bot.Handle(telebot.OnText, func(c telebot.Context) error {
		if !isUserAuthorized(c.Sender()) {
			log.Printf("Неавторизованный пользователь попытался использовать бота: @%s (ID: %d)",
				c.Sender().Username, c.Sender().ID)
			return nil
		}

		text := c.Text()
		if isAudioURL(text) {
			origMsg := c.Message()
			go sendToWorker(text, origMsg.Chat.ID, origMsg.ID)
			return c.Reply("Аудио получено, ожидайте результат!", &telebot.SendOptions{
				ReplyTo: origMsg,
			})
		}
		return c.Send("Пожалуйста, пришлите ссылку на аудиофайл.")
	})

	go func() {
		http.HandleFunc("/worker_result", workerResultHandler)
		addr := os.Getenv("RESULT_LISTEN_ADDR")
		if addr == "" {
			addr = ":8082"
		}
		log.Println("HTTP handler for worker results listening", addr)
		http.ListenAndServe(addr, nil)
	}()

	log.Println("Bot started")
	if len(allowedUsers) > 0 {
		log.Printf("Авторизация включена для %d пользователей", len(allowedUsers))
	} else {
		log.Println("Авторизация отключена - бот доступен всем пользователям")
	}
	bot.Start()
}

func initAllowedUsers() {
	allowedUsers = make(map[string]bool)

	usersEnv := os.Getenv("ALLOWED_USERS")
	if usersEnv == "" {
		log.Println("ALLOWED_USERS не установлена - бот будет доступен всем пользователям")
		return
	}

	users := strings.Split(usersEnv, ",")
	for _, user := range users {
		username := strings.TrimSpace(user)
		if username != "" {
			username = strings.TrimPrefix(username, "@")
			username = strings.ToLower(username)
			allowedUsers[username] = true
		}
	}

	log.Printf("Загружено %d разрешённых пользователей", len(allowedUsers))
}

func isUserAuthorized(user *telebot.User) bool {
	if len(allowedUsers) == 0 {
		return false
	}

	if user.Username != "" {
		username := strings.ToLower(user.Username)
		if allowedUsers[username] {
			return true
		}
	}

	return false
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

func sendToWorker(audioURL string, chatID int64, messageID int) {
	workerURL := os.Getenv("WORKER_URL")
	if workerURL == "" {
		log.Println("WORKER_URL is not set")
		return
	}

	body := strings.NewReader(
		fmt.Sprintf(`{"audio_url":"%s","chat_id":%d,"message_id":%d}`, audioURL, chatID, messageID),
	)
	resp, err := http.Post(workerURL+"/transcribe", "application/json", body)
	if err != nil {
		log.Printf("Sent to worker error: %v", err)
		return
	}
	defer resp.Body.Close()
	log.Printf("Worker received the task: %s", resp.Status)
}

func workerResultHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
	var res shared.WorkerResult
	if err := json.NewDecoder(r.Body).Decode(&res); err != nil {
		log.Println("JSON decode error:", err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	log.Printf("Result for chat %d message %d received", res.ChatID, res.MessageID)

	outputPath := fmt.Sprintf("result_%d_%d.pdf", res.ChatID, res.MessageID)
	err := generateResultPDF(res.Protocol, res.Transcript, outputPath)
	if err != nil {
		log.Printf("PDF generation error: %v", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	defer os.Remove(outputPath)

	if botInstance != nil {
		recipient := &telebot.Chat{ID: res.ChatID}
		doc := &telebot.Document{File: telebot.FromDisk(outputPath), FileName: "результат.pdf"}
		_, err := botInstance.Send(recipient, doc, &telebot.SendOptions{
			ReplyTo: &telebot.Message{ID: res.MessageID, Chat: recipient},
		})
		if err != nil {
			log.Printf("Sending PDF error: %v", err)
		}
	}

	w.WriteHeader(http.StatusOK)
}

func generateResultPDF(protocol, transcript, tempfilepath string) error {
	pdf := gofpdf.New("P", "mm", "A4", "")

	fontsDir := "fonts"
	pdf.AddUTF8Font("Roboto", "", filepath.Join(fontsDir, "Roboto-Regular.ttf"))
	pdf.AddUTF8Font("Roboto", "B", filepath.Join(fontsDir, "Roboto-Bold.ttf"))

	// Протокол
	pdf.AddPage()
	pdf.SetFont("Roboto", "B", 16)
	pdf.Cell(0, 12, "Протокол")
	pdf.Ln(14)
	pdf.SetFont("Roboto", "", 12)
	pdf.MultiCell(0, 8, protocol, "", "", false)

	// Новая страница — транскрипт
	pdf.AddPage()
	pdf.SetFont("Roboto", "B", 16)
	pdf.Cell(0, 12, "Сырой транскрипт")
	pdf.Ln(14)
	pdf.SetFont("Roboto", "", 12)
	pdf.MultiCell(0, 8, transcript, "", "", false)

	return pdf.OutputFileAndClose(tempfilepath)
}
