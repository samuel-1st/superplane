package telegram

import (
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/superplanehq/superplane/pkg/core"
	"github.com/superplanehq/superplane/test/support/contexts"
)

func Test__Telegram__Sync(t *testing.T) {
	tg := &Telegram{}

	t.Run("missing bot token -> error", func(t *testing.T) {
		integrationCtx := &contexts.IntegrationContext{
			Configuration: map[string]any{},
		}

		err := tg.Sync(core.SyncContext{
			Configuration: map[string]any{},
			Integration:   integrationCtx,
		})

		require.Error(t, err)
		assert.Contains(t, err.Error(), "botToken is required")
	})

	t.Run("empty bot token -> error", func(t *testing.T) {
		integrationCtx := &contexts.IntegrationContext{
			Configuration: map[string]any{
				"botToken": "",
			},
		}

		err := tg.Sync(core.SyncContext{
			Configuration: map[string]any{"botToken": ""},
			Integration:   integrationCtx,
		})

		require.Error(t, err)
		assert.Contains(t, err.Error(), "botToken is required")
	})

	t.Run("valid bot token -> verifies and sets ready", func(t *testing.T) {
		withDefaultTransport(t, func(req *http.Request) (*http.Response, error) {
			assert.Contains(t, req.URL.String(), "/getMe")
			assert.Equal(t, http.MethodGet, req.Method)
			return jsonResponse(http.StatusOK, `{
				"ok": true,
				"result": {
					"id": 123456789,
					"is_bot": true,
					"first_name": "TestBot",
					"username": "testbot"
				}
			}`), nil
		})

		botToken := "test-bot-token"
		integrationCtx := &contexts.IntegrationContext{
			Configuration: map[string]any{
				"botToken": botToken,
			},
		}

		err := tg.Sync(core.SyncContext{
			Configuration: map[string]any{"botToken": botToken},
			Integration:   integrationCtx,
		})

		require.NoError(t, err)
		assert.Equal(t, "ready", integrationCtx.State)

		metadata, ok := integrationCtx.Metadata.(Metadata)
		require.True(t, ok)
		assert.Equal(t, 123456789, metadata.BotID)
		assert.Equal(t, "testbot", metadata.Username)
	})

	t.Run("bot token verification fails -> error", func(t *testing.T) {
		withDefaultTransport(t, func(req *http.Request) (*http.Response, error) {
			return jsonResponse(http.StatusUnauthorized, `{"ok": false, "error_code": 401, "description": "Unauthorized"}`), nil
		})

		botToken := "invalid-token"
		integrationCtx := &contexts.IntegrationContext{
			Configuration: map[string]any{
				"botToken": botToken,
			},
		}

		err := tg.Sync(core.SyncContext{
			Configuration: map[string]any{"botToken": botToken},
			Integration:   integrationCtx,
		})

		require.Error(t, err)
		assert.Contains(t, err.Error(), "failed to verify bot token")
	})
}
