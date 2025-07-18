package storage

import (
	"database/sql"
	"encoding/json"
	"time"

	_ "github.com/lib/pq"
)

type Meeting struct {
	ID                  int64
	AudioURL            string
	Status              string
	CreatedAt           time.Time
	UpdatedAt           time.Time
	TranscriberID       string
	TranscriptionResult json.RawMessage
	BeautifulResult     string
	MessageID           string
}

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

func (s *Storage) CreateMeeting(audioURL string, transcriber_id string, message_id string) (*Meeting, error) {
	var id int64
	created_at := time.Now()
	err := s.db.QueryRow(`
		INSERT INTO meetings (audio_url, status, created_at, transcriber_id, message_id)
		VALUES ($1, 'processing', $2, $3, $4)
		RETURNING id
	`, audioURL, created_at, transcriber_id, message_id).Scan(&id)
	if err != nil {
		return nil, err
	}

	return &Meeting{
		ID:            id,
		AudioURL:      audioURL,
		Status:        "processing",
		CreatedAt:     created_at,
		TranscriberID: transcriber_id,
	}, nil
}

func (s *Storage) GetProcessingMeetings() ([]Meeting, error) {
	rows, err := s.db.Query(`
        SELECT id, audio_url, status, created_at, updated_at, transcriber_id
        FROM meetings
        WHERE status = 'processing'
    `)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var meetings []Meeting
	for rows.Next() {
		var m Meeting
		err := rows.Scan(&m.ID, &m.AudioURL, &m.Status, &m.CreatedAt, &m.UpdatedAt, &m.TranscriberID)
		if err != nil {
			return nil, err
		}
		meetings = append(meetings, m)
	}
	return meetings, nil
}

func (s *Storage) UpdateMeetingStatusAndResult(id int64, status string, result map[string]interface{}, beautifulText string) error {
	resultJSON, err := json.Marshal(result)
	if err != nil {
		return err
	}
	_, err = s.db.Exec(`
        UPDATE meetings
        SET status = $1,
            transcription_result = $2,
            updated_at = now(),
            beautiful_result = $3
        WHERE id = $4
    `, status, resultJSON, beautifulText, id)
	return err
}
