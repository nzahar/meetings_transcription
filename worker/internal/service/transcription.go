package service

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"github.com/nzahar/meetings_transcription/worker/config"
	"github.com/nzahar/meetings_transcription/worker/internal/storage"
	"io"
	"log"
	"net/http"
	"time"
)

type AgentRequest struct {
	ModelName string `json:"model"`
	AudioURL  string `json:"url"`
}
type AgentResponse struct {
	ID string `json:"id"`
}

func SendToAgent(agentURL, audioURL string) (string, error) {
	cfg := config.Load()

	body, _ := json.Marshal(AgentRequest{
		AudioURL:  audioURL,
		ModelName: "assemblyai-universal"})
	req, err := http.NewRequest("POST", agentURL, bytes.NewReader(body))
	if err != nil {
		return "", fmt.Errorf("failed to create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+cfg.AgentToken)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("request to agent failed: %w", err)
	}
	defer resp.Body.Close()

	respBody, _ := io.ReadAll(resp.Body)
	log.Printf("Agent response: %s", string(respBody))

	if resp.StatusCode >= 300 {
		return "", fmt.Errorf("agent returned status %d", resp.StatusCode)
	}

	var agentResp AgentResponse
	if err := json.Unmarshal(respBody, &agentResp); err != nil {
		return "", fmt.Errorf("failed to parse agent response: %w", err)
	}

	return agentResp.ID, nil
}

type Poller struct {
	DB *storage.Storage
}

func (p *Poller) Start(ctx context.Context) {
	ticker := time.NewTicker(10 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			p.pollPendingTranscriptions()
		}
	}
}

func (p *Poller) pollPendingTranscriptions() {
	meetings, err := p.DB.GetProcessingMeetings()
	if err != nil {
		log.Printf("error getting processing meetings: %v", err)
		return
	}

	for _, m := range meetings {
		status, result, err := p.checkStatus(m.TranscriberID)
		if err != nil {
			log.Printf("error checking status for meeting %d: %v", m.ID, err)
			continue
		}

		if status == "completed" || status == "error" {
			err := p.DB.UpdateMeetingStatusAndResult(m.ID, status, result)
			if err != nil {
				log.Printf("error updating meeting %d: %v", m.ID, err)
			}
		}
	}
}

func (p *Poller) checkStatus(transcriberID string) (string, map[string]interface{}, error) {
	cfg := config.Load()

	bodyreq, _ := json.Marshal(AgentRequest{
		ModelName: "assemblyai-universal"})
	req, err := http.NewRequest("GET", cfg.AgentURL+"/"+transcriberID, bytes.NewReader(bodyreq))
	if err != nil {
		return "", nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+cfg.AgentToken)

	log.Printf("Status request. %s, body: %s", transcriberID, string(bodyreq))

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return "", nil, err
	}
	defer resp.Body.Close()

	var body map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&body); err != nil {
		return "", nil, err
	}

	status, ok := body["status"].(string)
	log.Printf("Status result. %s, status: %s", transcriberID, status)

	if !ok {
		return "", nil, err
	}

	return status, body, nil
}
