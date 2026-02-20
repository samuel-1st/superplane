package telegram

import (
	"errors"
	"fmt"
	"strings"

	"github.com/google/uuid"
	"github.com/mitchellh/mapstructure"
	"github.com/superplanehq/superplane/pkg/configuration"
	"github.com/superplanehq/superplane/pkg/core"
)

const (
	ParseModeNone = ""
	ParseModeMdv2 = "mdv2"
)

// markdownV2Reserved lists characters that must be escaped in MarkdownV2 mode.
// Backslash must be first to avoid double-escaping.
var markdownV2Reserved = []string{`\`, "_", "*", "[", "]", "(", ")", "~", "`", ">", "#", "+", "-", "=", "|", "{", "}", ".", "!"}

type SendTextMessage struct{}

type SendTextMessageConfiguration struct {
	ChatID    string `json:"chatId" mapstructure:"chatId"`
	Text      string `json:"text" mapstructure:"text"`
	ParseMode string `json:"parseMode" mapstructure:"parseMode"`
}

func (c *SendTextMessage) Name() string {
	return "telegram.sendTextMessage"
}

func (c *SendTextMessage) Label() string {
	return "Send Text Message"
}

func (c *SendTextMessage) Description() string {
	return "Send a text message to a Telegram chat"
}

func (c *SendTextMessage) Documentation() string {
	return `The Send Text Message component sends a message to a Telegram chat using the Telegram Bot API.

## Use Cases

- **Notifications**: Send notifications about workflow events or system status
- **Alerts**: Alert teams about important events or errors
- **Updates**: Provide status updates on long-running processes

## Configuration

- **Chat ID**: The Telegram chat ID to send the message to (e.g. ` + "`" + `-1001234567890` + "`" + ` for a group)
- **Text**: The message text to send
- **Parse Mode**: Optional formatting mode. Select ` + "`" + `MarkdownV2` + "`" + ` to enable Telegram's MarkdownV2 formatting

## Output

Returns metadata about the sent message including message ID and text.

## Notes

- The Telegram bot must be a member of the target chat and have permission to send messages
- When using MarkdownV2 formatting, reserved characters in the text are automatically escaped`
}

func (c *SendTextMessage) Icon() string {
	return "telegram"
}

func (c *SendTextMessage) Color() string {
	return "blue"
}

func (c *SendTextMessage) OutputChannels(configuration any) []core.OutputChannel {
	return []core.OutputChannel{core.DefaultOutputChannel}
}

func (c *SendTextMessage) Configuration() []configuration.Field {
	return []configuration.Field{
		{
			Name:        "chatId",
			Label:       "Chat ID",
			Type:        configuration.FieldTypeString,
			Required:    true,
			Description: "Telegram chat ID to send the message to",
		},
		{
			Name:        "text",
			Label:       "Text",
			Type:        configuration.FieldTypeText,
			Required:    true,
			Description: "Message text to send",
		},
		{
			Name:     "parseMode",
			Label:    "Parse Mode",
			Type:     configuration.FieldTypeSelect,
			Required: false,
			Default:  ParseModeNone,
			TypeOptions: &configuration.TypeOptions{
				Select: &configuration.SelectTypeOptions{
					Options: []configuration.FieldOption{
						{Label: "None", Value: ParseModeNone},
						{Label: "MarkdownV2", Value: ParseModeMdv2},
					},
				},
			},
			Description: "Optional message formatting mode",
		},
	}
}

func (c *SendTextMessage) Setup(ctx core.SetupContext) error {
	var config SendTextMessageConfiguration
	if err := mapstructure.Decode(ctx.Configuration, &config); err != nil {
		return fmt.Errorf("failed to decode configuration: %w", err)
	}

	if config.ChatID == "" {
		return errors.New("chatId is required")
	}

	if config.Text == "" {
		return errors.New("text is required")
	}

	if config.ParseMode != ParseModeNone && config.ParseMode != ParseModeMdv2 {
		return fmt.Errorf("invalid parseMode %q: must be %q or %q", config.ParseMode, ParseModeNone, ParseModeMdv2)
	}

	return nil
}

func (c *SendTextMessage) ProcessQueueItem(ctx core.ProcessQueueContext) (*uuid.UUID, error) {
	return ctx.DefaultProcessing()
}

func (c *SendTextMessage) Execute(ctx core.ExecutionContext) error {
	var config SendTextMessageConfiguration
	if err := mapstructure.Decode(ctx.Configuration, &config); err != nil {
		return fmt.Errorf("failed to decode configuration: %w", err)
	}

	if config.ChatID == "" {
		return errors.New("chatId is required")
	}

	if config.Text == "" {
		return errors.New("text is required")
	}

	client, err := NewClient(ctx.Integration)
	if err != nil {
		return fmt.Errorf("failed to create Telegram client: %w", err)
	}

	text := config.Text
	req := SendMessageRequest{
		ChatID: config.ChatID,
		Text:   text,
	}

	if config.ParseMode == ParseModeMdv2 {
		req.Text = EscapeMarkdownV2(text)
		req.ParseMode = "MarkdownV2"
	}

	message, err := client.SendMessage(req)
	if err != nil {
		return fmt.Errorf("failed to send message: %w", err)
	}

	return ctx.ExecutionState.Emit(
		core.DefaultOutputChannel.Name,
		"telegram.message.sent",
		[]any{map[string]any{
			"message_id": message.MessageID,
			"text":       message.Text,
			"date":       message.Date,
		}},
	)
}

func (c *SendTextMessage) HandleWebhook(ctx core.WebhookRequestContext) (int, error) {
	return 200, nil
}

func (c *SendTextMessage) Actions() []core.Action {
	return []core.Action{}
}

func (c *SendTextMessage) HandleAction(ctx core.ActionContext) error {
	return nil
}

func (c *SendTextMessage) Cancel(ctx core.ExecutionContext) error {
	return nil
}

func (c *SendTextMessage) Cleanup(ctx core.SetupContext) error {
	return nil
}

// EscapeMarkdownV2 escapes all reserved MarkdownV2 characters in text.
// Backslashes are escaped first to prevent double-escaping.
func EscapeMarkdownV2(text string) string {
	for _, char := range markdownV2Reserved {
		text = strings.ReplaceAll(text, char, `\`+char)
	}
	return text
}
