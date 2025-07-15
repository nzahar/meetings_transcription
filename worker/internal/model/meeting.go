package model

import "time"

type Meeting struct {
	ID            int64
	AudioURL      string
	Status        string
	CreatedAt     time.Time
	TranscriberID string
}
