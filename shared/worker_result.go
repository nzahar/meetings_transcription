package shared

type WorkerResult struct {
	ChatID     int64  `json:"chat_id"`
	MessageID  int    `json:"message_id"`
	Transcript string `json:"transcript"`
	Protocol   string `json:"protocol"`
}
