package telegram

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/superplanehq/superplane/pkg/core"
)

const telegramAPIBase = "https://api.telegram.org/bot"

type Client struct {
	BotToken string
}

func NewClient(ctx core.IntegrationContext) (*Client, error) {
	botToken, err := ctx.GetConfig("botToken")
	if err != nil {
		return nil, fmt.Errorf("failed to get bot token: %w", err)
	}

	token := string(botToken)
	if token == "" {
		return nil, fmt.Errorf("bot token is required")
	}

	return &Client{
		BotToken: token,
	}, nil
}

// BotUser represents a Telegram bot user from the getMe API
type BotUser struct {
	ID       int    `json:"id"`
	IsBot    bool   `json:"is_bot"`
	Username string `json:"username"`
	Name     string `json:"first_name"`
}

// getMeResponse is the API response from getMe
type getMeResponse struct {
	OK     bool    `json:"ok"`
	Result BotUser `json:"result"`
}

// SendMessageRequest is the request body for sendMessage
type SendMessageRequest struct {
	ChatID    string `json:"chat_id"`
	Text      string `json:"text"`
	ParseMode string `json:"parse_mode,omitempty"`
}

// Message represents a Telegram message in the sendMessage response
type Message struct {
	MessageID int    `json:"message_id"`
	Text      string `json:"text"`
	Date      int64  `json:"date"`
}

// sendMessageResponse is the API response from sendMessage
type sendMessageResponse struct {
	OK     bool    `json:"ok"`
	Result Message `json:"result"`
}

// GetMe validates the bot token by calling the getMe endpoint
func (c *Client) GetMe() (*BotUser, error) {
	responseBody, err := c.doRequest(http.MethodGet, "getMe", nil)
	if err != nil {
		return nil, err
	}

	var resp getMeResponse
	if err := json.Unmarshal(responseBody, &resp); err != nil {
		return nil, fmt.Errorf("failed to decode getMe response: %w", err)
	}

	if !resp.OK {
		return nil, fmt.Errorf("getMe returned ok=false")
	}

	return &resp.Result, nil
}

// SendMessage sends a message to a chat
func (c *Client) SendMessage(req SendMessageRequest) (*Message, error) {
	body, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	responseBody, err := c.doRequest(http.MethodPost, "sendMessage", bytes.NewReader(body))
	if err != nil {
		return nil, err
	}

	var resp sendMessageResponse
	if err := json.Unmarshal(responseBody, &resp); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	if !resp.OK {
		return nil, fmt.Errorf("sendMessage returned ok=false")
	}

	return &resp.Result, nil
}

// doRequest executes an HTTP request to the Telegram Bot API
func (c *Client) doRequest(method, endpoint string, body io.Reader) ([]byte, error) {
	url := fmt.Sprintf("%s%s/%s", telegramAPIBase, c.BotToken, endpoint)

	req, err := http.NewRequest(method, url, body)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	responseBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, fmt.Errorf("request failed: status %d, body: %s", resp.StatusCode, string(responseBody))
	}

	return responseBody, nil
}
