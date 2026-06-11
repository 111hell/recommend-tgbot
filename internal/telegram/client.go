package telegram

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"
)

type Client struct {
	token      string
	baseURL    string
	httpClient *http.Client
}

type SendOptions struct {
	ParseMode string
	Buttons   []Button
}

type Button struct {
	Text string
	URL  string
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
	return c.SendMessageWithOptions(ctx, chatID, text, SendOptions{ParseMode: parseMode})
}

func (c *Client) SendMessageWithOptions(ctx context.Context, chatID string, text string, options SendOptions) error {
	if c.token == "" {
		return fmt.Errorf("telegram token is required")
	}
	payload := sendMessageRequest{
		ChatID:    strings.TrimSpace(chatID),
		Text:      text,
		ParseMode: strings.TrimSpace(options.ParseMode),
	}
	if payload.ChatID == "" {
		return fmt.Errorf("telegram chat id is required")
	}
	if len(options.Buttons) > 0 {
		row := make([]inlineKeyboardButton, 0, len(options.Buttons))
		for _, button := range options.Buttons {
			text := strings.TrimSpace(button.Text)
			buttonURL := strings.TrimSpace(button.URL)
			if text == "" || !isAllowedButtonURL(buttonURL) {
				continue
			}
			row = append(row, inlineKeyboardButton{Text: text, URL: buttonURL})
		}
		if len(row) > 0 {
			payload.ReplyMarkup = &inlineKeyboardMarkup{InlineKeyboard: [][]inlineKeyboardButton{row}}
		}
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
		body, _ := io.ReadAll(io.LimitReader(resp.Body, 4096))
		message := strings.TrimSpace(string(body))
		if message != "" {
			return fmt.Errorf("telegram sendMessage failed: %s: %s", resp.Status, message)
		}
		return fmt.Errorf("telegram sendMessage failed: %s", resp.Status)
	}
	return nil
}

func isAllowedButtonURL(rawURL string) bool {
	parsed, err := url.Parse(strings.TrimSpace(rawURL))
	if err != nil {
		return false
	}
	switch parsed.Scheme {
	case "http", "https", "tg":
		return true
	default:
		return false
	}
}

type sendMessageRequest struct {
	ChatID      string                `json:"chat_id"`
	Text        string                `json:"text"`
	ParseMode   string                `json:"parse_mode,omitempty"`
	ReplyMarkup *inlineKeyboardMarkup `json:"reply_markup,omitempty"`
}

type inlineKeyboardMarkup struct {
	InlineKeyboard [][]inlineKeyboardButton `json:"inline_keyboard"`
}

type inlineKeyboardButton struct {
	Text string `json:"text"`
	URL  string `json:"url"`
}
