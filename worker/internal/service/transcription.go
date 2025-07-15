package service

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/nzahar/meetings_transcription/worker/config"
	"io"
	"log"
	"net/http"
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
