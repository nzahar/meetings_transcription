package handler

import (
	"encoding/json"
	"github.com/nzahar/meetings_transcription/worker/internal/service"
	"github.com/nzahar/meetings_transcription/worker/internal/storage"
	"net/http"
)

type Handler struct {
	storage  *storage.Storage
	agentURL string
}

func New(storage *storage.Storage, agentURL string) *Handler {
	return &Handler{storage: storage, agentURL: agentURL}
}

type transcriptionRequest struct {
	AudioURL  string `json:"audio_url"`
	MessageID int    `json:"message_id"`
	ChatID    int64  `json:"chat_id"`
}

func (h *Handler) HandleTranscriptionRequest(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "only POST allowed", http.StatusMethodNotAllowed)
		return
	}

	var req transcriptionRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request", http.StatusBadRequest)
		return
	}

	res_id, err := service.SendToTranscribeAgent(h.agentURL, req.AudioURL)
	if err != nil {
		http.Error(w, "failed to send to agent", http.StatusBadGateway)
		return
	}

	meeting, err := h.storage.CreateMeeting(req.AudioURL, res_id, req.MessageID, req.ChatID)
	if err != nil {
		http.Error(w, "failed to create meeting", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusAccepted)
	json.NewEncoder(w).Encode(meeting)
}
