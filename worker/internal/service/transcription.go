package service

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"github.com/nzahar/meetings_transcription/shared"
	"github.com/nzahar/meetings_transcription/worker/config"
	"github.com/nzahar/meetings_transcription/worker/internal/storage"
	"io"
	"log"
	"net/http"
	"os"
	"strings"
	"time"
)

type TranscribeAgentRequest struct {
	ModelName string `json:"model"`
	AudioURL  string `json:"url"`
}
type TranscribeAgentResponse struct {
	ID string `json:"id"`
}

func SendToTranscribeAgent(agentURL, audioURL string) (string, error) {
	cfg := config.Load()

	body, _ := json.Marshal(TranscribeAgentRequest{
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

	var agentResp TranscribeAgentResponse
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

		beautifulText := ""
		if status == "completed" {
			utterancesRaw := result["utterances"]
			cleanedUtterances, err := cleanUtterances(utterancesRaw)
			if err != nil {
				fmt.Println("Ошибка:", err)
				return
			}

			utterancesJSON, err := json.Marshal(cleanedUtterances)
			if err != nil {
				log.Printf("failed to marshal utterances: %w", err)
			}

			beautifulText, err = SummarizeMeeting(utterancesJSON)
			if err != nil {
				log.Printf("failed to make text beautiful: %w", err)
				return
			}

			result := shared.WorkerResult{
				ChatID:     m.ChatID,
				MessageID:  m.MessageID,
				Transcript: string(utterancesJSON),
				Protocol:   beautifulText,
			}
			if err := sendResultToBot(result); err != nil {
				log.Printf("Error sendint result to the bot: %v", err)
			}

		}

		if status == "error" {
			resultJSON, err := json.Marshal(result)
			if err != nil {
				log.Printf("failed to marshal result: %w", err)
			}
			failResult := shared.WorkerResult{
				ChatID:     m.ChatID,
				MessageID:  m.MessageID,
				Transcript: string(resultJSON),
				Protocol:   "",
			}
			if err := sendResultToBot(failResult); err != nil {
				log.Printf("Error sending error to the bot: %v", err)
			}

		}

		if status == "completed" || status == "error" {
			err := p.DB.UpdateMeetingStatusAndResult(m.ID, status, result, beautifulText)
			if err != nil {
				log.Printf("error updating meeting %d: %v", m.ID, err)
			}
		}
	}
}

func (p *Poller) checkStatus(transcriberID string) (string, map[string]interface{}, error) {
	cfg := config.Load()

	bodyreq, _ := json.Marshal(TranscribeAgentRequest{
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

type Message struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type ChatRequest struct {
	Model    string    `json:"model"`
	Messages []Message `json:"messages"`
	Stream   bool      `json:"stream"`
}

type ChatResponse struct {
	Choices []struct {
		Message Message `json:"message"`
	} `json:"choices"`
}

func SummarizeMeeting(input json.RawMessage) (string, error) {
	cfg := config.Load()

	systemPrompt := cfg.SummarizationPrompt
	agentURL := cfg.AgentLLMURL

	messages := []Message{
		{
			Role:    "system",
			Content: systemPrompt,
		},
		{
			Role:    "user",
			Content: string(input),
		},
	}

	reqBody := ChatRequest{
		Model:    cfg.LLMModel,
		Messages: messages,
		Stream:   false,
	}

	body, err := json.Marshal(reqBody)
	if err != nil {
		return "", fmt.Errorf("failed to marshal request body: %w", err)
	}
	log.Printf("sending text to make it beautiful")

	req, err := http.NewRequest("POST", agentURL, bytes.NewReader(body))
	if err != nil {
		return "", fmt.Errorf("failed to send request to agent: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+cfg.AgentToken)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		data, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("non-200 response: %s - %s", resp.Status, string(data))
	}
	log.Printf("answer received")

	var response ChatResponse
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return "", fmt.Errorf("failed to decode response: %w", err)
	}

	if len(response.Choices) == 0 {
		return "", fmt.Errorf("no choices returned")
	}
	//log.Printf("answer is %s", response.Choices[0].Message.Content)
	log.Printf("The answer is ready")

	return response.Choices[0].Message.Content, nil
}

func sendResultToBot(result shared.WorkerResult) error {
	botURL := os.Getenv("BOT_URL")
	if botURL == "" {
		return fmt.Errorf("BOT_URL не задан")
	}

	payload, err := json.Marshal(result)
	if err != nil {
		return err
	}
	resp, err := http.Post(botURL+"/worker_result", "application/json", strings.NewReader(string(payload)))
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("бот ответил статусом %d", resp.StatusCode)
	}
	return nil
}

type Utterance struct {
	Speaker string `json:"speaker"`
	Text    string `json:"text"`
}

func cleanUtterances(utterancesRaw interface{}) ([]Utterance, error) {
	var cleaned []Utterance

	rawList, ok := utterancesRaw.([]interface{})
	if !ok {
		return nil, fmt.Errorf("utterancesRaw не является []interface{}, а %T", utterancesRaw)
	}

	for _, item := range rawList {
		m, ok := item.(map[string]interface{})
		if !ok {
			continue
		}
		speaker, _ := m["speaker"].(string)
		text, _ := m["text"].(string)
		cleaned = append(cleaned, Utterance{
			Speaker: speaker,
			Text:    text,
		})
	}
	return cleaned, nil
}
