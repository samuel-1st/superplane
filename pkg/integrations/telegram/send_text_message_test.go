package telegram

import (
	"encoding/json"
	"io"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/superplanehq/superplane/pkg/core"
	"github.com/superplanehq/superplane/test/support/contexts"
)

func Test__EscapeMarkdownV2(t *testing.T) {
	t.Run("escapes all reserved characters", func(t *testing.T) {
		input := `_ * [ ] ( ) ~ ` + "`" + ` > # + - = | { } . !`
		result := EscapeMarkdownV2(input)
		assert.Equal(t, `\_ \* \[ \] \( \) \~ \`+"`"+` \> \# \+ \- \= \| \{ \} \. \!`, result)
	})

	t.Run("escapes backslash first to avoid double-escaping", func(t *testing.T) {
		result := EscapeMarkdownV2(`hello\world`)
		assert.Equal(t, `hello\\world`, result)
	})

	t.Run("plain text is unchanged", func(t *testing.T) {
		result := EscapeMarkdownV2("hello world")
		assert.Equal(t, "hello world", result)
	})

	t.Run("mixed text escapes only reserved chars", func(t *testing.T) {
		result := EscapeMarkdownV2("Deploy to prod_env failed!")
		assert.Equal(t, `Deploy to prod\_env failed\!`, result)
	})
}

func Test__SendTextMessage__Setup(t *testing.T) {
	component := &SendTextMessage{}

	t.Run("invalid configuration -> error", func(t *testing.T) {
		err := component.Setup(core.SetupContext{
			Integration:   &contexts.IntegrationContext{},
			Metadata:      &contexts.MetadataContext{},
			Configuration: "invalid",
		})

		require.ErrorContains(t, err, "failed to decode configuration")
	})

	t.Run("missing chatId with no integration default -> error", func(t *testing.T) {
		err := component.Setup(core.SetupContext{
			Integration:   &contexts.IntegrationContext{},
			Metadata:      &contexts.MetadataContext{},
			Configuration: map[string]any{"text": "Hello"},
		})

		require.ErrorContains(t, err, "chatId is required")
	})

	t.Run("missing text -> error", func(t *testing.T) {
		err := component.Setup(core.SetupContext{
			Integration:   &contexts.IntegrationContext{},
			Metadata:      &contexts.MetadataContext{},
			Configuration: map[string]any{"chatId": "123456"},
		})

		require.ErrorContains(t, err, "text is required")
	})

	t.Run("invalid parseMode -> error", func(t *testing.T) {
		err := component.Setup(core.SetupContext{
			Integration:   &contexts.IntegrationContext{},
			Metadata:      &contexts.MetadataContext{},
			Configuration: map[string]any{"chatId": "123456", "text": "Hello", "parseMode": "html"},
		})

		require.ErrorContains(t, err, "invalid parseMode")
	})

	t.Run("chatId from integration default -> ok", func(t *testing.T) {
		err := component.Setup(core.SetupContext{
			Integration: &contexts.IntegrationContext{
				Configuration: map[string]any{"defaultChatId": "999888"},
			},
			Metadata:      &contexts.MetadataContext{},
			Configuration: map[string]any{"text": "Hello"},
		})

		require.NoError(t, err)
	})

	t.Run("valid config with no parseMode -> ok", func(t *testing.T) {
		err := component.Setup(core.SetupContext{
			Integration:   &contexts.IntegrationContext{},
			Metadata:      &contexts.MetadataContext{},
			Configuration: map[string]any{"chatId": "123456", "text": "Hello"},
		})

		require.NoError(t, err)
	})

	t.Run("valid config with mdv2 parseMode -> ok", func(t *testing.T) {
		err := component.Setup(core.SetupContext{
			Integration:   &contexts.IntegrationContext{},
			Metadata:      &contexts.MetadataContext{},
			Configuration: map[string]any{"chatId": "123456", "text": "Hello", "parseMode": "mdv2"},
		})

		require.NoError(t, err)
	})
}

func Test__SendTextMessage__Execute(t *testing.T) {
	component := &SendTextMessage{}

	t.Run("sends message without parse_mode when not set", func(t *testing.T) {
		withDefaultTransport(t, func(req *http.Request) (*http.Response, error) {
			assert.Contains(t, req.URL.String(), "/sendMessage")
			assert.Equal(t, http.MethodPost, req.Method)

			body, err := io.ReadAll(req.Body)
			require.NoError(t, err)

			var payload SendMessageRequest
			require.NoError(t, json.Unmarshal(body, &payload))
			assert.Equal(t, "123456", payload.ChatID)
			assert.Equal(t, "Hello, Telegram!", payload.Text)
			assert.Empty(t, payload.ParseMode)

			return jsonResponse(http.StatusOK, `{
				"ok": true,
				"result": {"message_id": 42, "text": "Hello, Telegram!", "date": 1737028800}
			}`), nil
		})

		execState := &contexts.ExecutionStateContext{KVs: map[string]string{}}
		integrationCtx := &contexts.IntegrationContext{
			Configuration: map[string]any{"botToken": "test-bot-token"},
		}

		err := component.Execute(core.ExecutionContext{
			Integration:    integrationCtx,
			ExecutionState: execState,
			Configuration: map[string]any{
				"chatId": "123456",
				"text":   "Hello, Telegram!",
			},
		})

		require.NoError(t, err)
		assert.Equal(t, core.DefaultOutputChannel.Name, execState.Channel)
		assert.Equal(t, "telegram.message.sent", execState.Type)
		require.Len(t, execState.Payloads, 1)

		payload := execState.Payloads[0].(map[string]any)
		data := payload["data"].(map[string]any)
		assert.Equal(t, 42, data["message_id"])
		assert.Equal(t, "Hello, Telegram!", data["text"])
	})

	t.Run("sends message with MarkdownV2 when parseMode is mdv2", func(t *testing.T) {
		withDefaultTransport(t, func(req *http.Request) (*http.Response, error) {
			body, err := io.ReadAll(req.Body)
			require.NoError(t, err)

			var payload SendMessageRequest
			require.NoError(t, json.Unmarshal(body, &payload))
			assert.Equal(t, "MarkdownV2", payload.ParseMode)
			// Reserved chars in the text should be escaped
			assert.Equal(t, `Deploy to prod\_env failed\!`, payload.Text)

			return jsonResponse(http.StatusOK, `{
				"ok": true,
				"result": {"message_id": 43, "text": "Deploy to prod_env failed!", "date": 1737028800}
			}`), nil
		})

		execState := &contexts.ExecutionStateContext{KVs: map[string]string{}}
		integrationCtx := &contexts.IntegrationContext{
			Configuration: map[string]any{"botToken": "test-bot-token"},
		}

		err := component.Execute(core.ExecutionContext{
			Integration:    integrationCtx,
			ExecutionState: execState,
			Configuration: map[string]any{
				"chatId":    "123456",
				"text":      "Deploy to prod_env failed!",
				"parseMode": "mdv2",
			},
		})

		require.NoError(t, err)
	})

	t.Run("API failure -> error", func(t *testing.T) {
		withDefaultTransport(t, func(req *http.Request) (*http.Response, error) {
			return jsonResponse(http.StatusForbidden, `{"ok": false, "error_code": 403, "description": "Forbidden"}`), nil
		})

		execState := &contexts.ExecutionStateContext{KVs: map[string]string{}}
		integrationCtx := &contexts.IntegrationContext{
			Configuration: map[string]any{"botToken": "test-bot-token"},
		}

		err := component.Execute(core.ExecutionContext{
			Integration:    integrationCtx,
			ExecutionState: execState,
			Configuration: map[string]any{
				"chatId": "123456",
				"text":   "Hello",
			},
		})

		require.Error(t, err)
		assert.Contains(t, err.Error(), "failed to send message")
	})

	t.Run("missing chatId with no default -> error", func(t *testing.T) {
		execState := &contexts.ExecutionStateContext{KVs: map[string]string{}}
		integrationCtx := &contexts.IntegrationContext{
			Configuration: map[string]any{"botToken": "test-bot-token"},
		}

		err := component.Execute(core.ExecutionContext{
			Integration:    integrationCtx,
			ExecutionState: execState,
			Configuration:  map[string]any{"text": "Hello"},
		})

		require.Error(t, err)
		assert.Contains(t, err.Error(), "chatId is required")
	})
}
