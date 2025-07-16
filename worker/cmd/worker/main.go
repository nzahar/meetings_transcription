package main

import (
	"context"
	"github.com/nzahar/meetings_transcription/worker/config"
	"github.com/nzahar/meetings_transcription/worker/internal/handler"
	"github.com/nzahar/meetings_transcription/worker/internal/service"
	"github.com/nzahar/meetings_transcription/worker/internal/storage"
	"log"
	"net/http"
)

func main() {
	cfg := config.Load()

	db, err := storage.NewPostgres(cfg.DatabaseURL)
	if err != nil {
		log.Fatalf("DB connection failed: %v", err)
	}
	defer db.Close()

	go func() {
		p := &service.Poller{
			DB: db,
		}
		ctx := context.Background()
		p.Start(ctx)
	}()

	h := handler.New(db, cfg.AgentURL)
	http.HandleFunc("/transcribe", h.HandleTranscriptionRequest)

	log.Printf("Listening on %s...", cfg.ListenAddr)
	if err := http.ListenAndServe(cfg.ListenAddr, nil); err != nil {
		log.Fatal(err)
	}
}
