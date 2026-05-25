package telegram

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"
)

type Client struct {
	token      string
	baseURL    string
	httpClient *http.Client
}

func NewClient(token string) *Client {
	return &Client{
		token:   strings.TrimSpace(token),
		baseURL: "https://api.telegram.org",
		httpClient: &http.Client{
			Timeout: 20 * time.Second,
		},
	}
}

func (c *Client) SendMessage(ctx context.Context, chatID string, text string, parseMode string) error {
	if c.token == "" {
		return fmt.Errorf("telegram token is required")
	}
	payload := sendMessageRequest{
		ChatID:    strings.TrimSpace(chatID),
		Text:      text,
		ParseMode: strings.TrimSpace(parseMode),
	}
	if payload.ChatID == "" {
		return fmt.Errorf("telegram chat id is required")
	}

	body, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("marshal telegram request: %w", err)
	}

	endpoint := fmt.Sprintf("%s/bot%s/sendMessage", c.baseURL, c.token)
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint, bytes.NewReader(body))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("telegram sendMessage request: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("telegram sendMessage failed: %s", resp.Status)
	}
	return nil
}

type sendMessageRequest struct {
	ChatID    string `json:"chat_id"`
	Text      string `json:"text"`
	ParseMode string `json:"parse_mode,omitempty"`
}
