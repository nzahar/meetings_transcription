package storage

import (
	"database/sql"
	"github.com/nzahar/meetings_transcription/worker/internal/model"
	"time"

	_ "github.com/lib/pq"
)

type Storage struct {
	db *sql.DB
}

func NewPostgres(dsn string) (*Storage, error) {
	db, err := sql.Open("postgres", dsn)
	if err != nil {
		return nil, err
	}
	return &Storage{db: db}, nil
}

func (s *Storage) Close() error {
	return s.db.Close()
}

func (s *Storage) CreateMeeting(audioURL string, transcriber_id string) (*model.Meeting, error) {
	var id int64
	created_at := time.Now()
	err := s.db.QueryRow(`
		INSERT INTO meetings (audio_url, status, created_at, transcriber_id)
		VALUES ($1, 'processing', $2, $3)
		RETURNING id
	`, audioURL, created_at, transcriber_id).Scan(&id)
	if err != nil {
		return nil, err
	}

	return &model.Meeting{
		ID:            id,
		AudioURL:      audioURL,
		Status:        "processing",
		CreatedAt:     created_at,
		TranscriberID: transcriber_id,
	}, nil
}
