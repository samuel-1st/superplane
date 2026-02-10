package slack

import (
	"encoding/json"
	"io"
	"net/http"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/superplanehq/superplane/pkg/core"
	"github.com/superplanehq/superplane/test/support/contexts"
)

func Test__SendAndWaitForResponse__Setup(t *testing.T) {
	component := &SendAndWaitForResponse{}

	t.Run("invalid configuration -> error", func(t *testing.T) {
		err := component.Setup(core.SetupContext{
			Integration:   &contexts.IntegrationContext{},
			Metadata:      &contexts.MetadataContext{},
			Configuration: "invalid",
		})

		require.ErrorContains(t, err, "failed to decode configuration")
	})

	t.Run("missing channel -> error", func(t *testing.T) {
		err := component.Setup(core.SetupContext{
			Integration:   &contexts.IntegrationContext{},
			Metadata:      &contexts.MetadataContext{},
			Configuration: map[string]any{
				"channel": "",
				"message": "test",
				"buttons": []map[string]any{{"name": "Yes", "value": "yes"}},
			},
		})

		require.ErrorContains(t, err, "channel is required")
	})

	t.Run("missing message -> error", func(t *testing.T) {
		err := component.Setup(core.SetupContext{
			Integration:   &contexts.IntegrationContext{},
			Metadata:      &contexts.MetadataContext{},
			Configuration: map[string]any{
				"channel": "C123",
				"message": "",
				"buttons": []map[string]any{{"name": "Yes", "value": "yes"}},
			},
		})

		require.ErrorContains(t, err, "message is required")
	})

	t.Run("no buttons -> error", func(t *testing.T) {
		err := component.Setup(core.SetupContext{
			Integration:   &contexts.IntegrationContext{},
			Metadata:      &contexts.MetadataContext{},
			Configuration: map[string]any{
				"channel": "C123",
				"message": "test",
				"buttons": []map[string]any{},
			},
		})

		require.ErrorContains(t, err, "at least one button is required")
	})

	t.Run("too many buttons -> error", func(t *testing.T) {
		err := component.Setup(core.SetupContext{
			Integration:   &contexts.IntegrationContext{},
			Metadata:      &contexts.MetadataContext{},
			Configuration: map[string]any{
				"channel": "C123",
				"message": "test",
				"buttons": []map[string]any{
					{"name": "1", "value": "1"},
					{"name": "2", "value": "2"},
					{"name": "3", "value": "3"},
					{"name": "4", "value": "4"},
					{"name": "5", "value": "5"},
				},
			},
		})

		require.ErrorContains(t, err, "maximum of 4 buttons allowed")
	})

	t.Run("button missing name -> error", func(t *testing.T) {
		err := component.Setup(core.SetupContext{
			Integration:   &contexts.IntegrationContext{},
			Metadata:      &contexts.MetadataContext{},
			Configuration: map[string]any{
				"channel": "C123",
				"message": "test",
				"buttons": []map[string]any{{"value": "yes"}},
			},
		})

		require.ErrorContains(t, err, "name is required")
	})

	t.Run("button missing value -> error", func(t *testing.T) {
		err := component.Setup(core.SetupContext{
			Integration:   &contexts.IntegrationContext{},
			Metadata:      &contexts.MetadataContext{},
			Configuration: map[string]any{
				"channel": "C123",
				"message": "test",
				"buttons": []map[string]any{{"name": "Yes"}},
			},
		})

		require.ErrorContains(t, err, "value is required")
	})

	t.Run("valid configuration -> stores metadata", func(t *testing.T) {
		withDefaultTransport(t, func(req *http.Request) (*http.Response, error) {
			assert.Equal(t, "https://slack.com/api/conversations.info", req.URL.Scheme+"://"+req.URL.Host+req.URL.Path)
			assert.Equal(t, "C123", req.URL.Query().Get("channel"))
			return jsonResponse(http.StatusOK, `{"ok": true, "channel": {"id": "C123", "name": "general"}}`), nil
		})

		metadata := &contexts.MetadataContext{}
		integrationCtx := &contexts.IntegrationContext{
			Configuration: map[string]any{
				"botToken": "token-123",
			},
		}

		err := component.Setup(core.SetupContext{
			Integration: integrationCtx,
			Metadata:    metadata,
			Configuration: map[string]any{
				"channel": "C123",
				"message": "Choose one:",
				"buttons": []map[string]any{
					{"name": "Approve", "value": "approve"},
					{"name": "Reject", "value": "reject"},
				},
			},
		})

		require.NoError(t, err)
		stored, ok := metadata.Metadata.(SendAndWaitForResponseMetadata)
		require.True(t, ok)
		require.NotNil(t, stored.Channel)
		assert.Equal(t, "C123", stored.Channel.ID)
		assert.Equal(t, "general", stored.Channel.Name)
		assert.Equal(t, "pending", stored.State)
		assert.NotNil(t, stored.SubscriptionID)
		assert.Len(t, integrationCtx.Subscriptions, 1)
	})
}

func Test__SendAndWaitForResponse__Execute(t *testing.T) {
	component := &SendAndWaitForResponse{}

	t.Run("missing channel -> error", func(t *testing.T) {
		err := component.Execute(core.ExecutionContext{
			Integration:    &contexts.IntegrationContext{},
			ExecutionState: &contexts.ExecutionStateContext{KVs: map[string]string{}},
			Metadata:       &contexts.MetadataContext{},
			Requests:       &contexts.RequestContext{},
			Configuration: map[string]any{
				"channel": "",
				"message": "test",
				"buttons": []map[string]any{{"name": "Yes", "value": "yes"}},
			},
		})

		require.ErrorContains(t, err, "channel is required")
	})

	t.Run("valid configuration -> sends message with buttons", func(t *testing.T) {
		executionID := uuid.New()

		withDefaultTransport(t, func(req *http.Request) (*http.Response, error) {
			if req.URL.String() == "https://slack.com/api/chat.postMessage" {
				body, err := io.ReadAll(req.Body)
				require.NoError(t, err)
				var payload ChatPostMessageWithBlocksRequest
				require.NoError(t, json.Unmarshal(body, &payload))
				assert.Equal(t, "C123", payload.Channel)
				assert.Equal(t, "Choose one:", payload.Text)
				assert.Len(t, payload.Blocks, 2) // text section + buttons
				return jsonResponse(http.StatusOK, `{"ok": true, "ts": "1234567890.123456"}`), nil
			}
			return jsonResponse(http.StatusNotFound, `{"ok": false}`), nil
		})

		metadata := &contexts.MetadataContext{
			Metadata: SendAndWaitForResponseMetadata{
				State: "pending",
			},
		}
		integrationCtx := &contexts.IntegrationContext{
			Configuration: map[string]any{
				"botToken": "token-123",
			},
		}

		err := component.Execute(core.ExecutionContext{
			ID:             executionID,
			Integration:    integrationCtx,
			ExecutionState: &contexts.ExecutionStateContext{KVs: map[string]string{}},
			Metadata:       metadata,
			Requests:       &contexts.RequestContext{},
			Configuration: map[string]any{
				"channel": "C123",
				"message": "Choose one:",
				"buttons": []map[string]any{
					{"name": "Approve", "value": "approve"},
					{"name": "Reject", "value": "reject"},
				},
			},
		})

		require.NoError(t, err)
		stored, ok := metadata.Metadata.(SendAndWaitForResponseMetadata)
		require.True(t, ok)
		assert.Equal(t, "1234567890.123456", stored.MessageTS)
		assert.Equal(t, "waiting", stored.State)
	})

	t.Run("with timeout -> schedules timeout action", func(t *testing.T) {
		executionID := uuid.New()
		timeout := 60

		withDefaultTransport(t, func(req *http.Request) (*http.Response, error) {
			return jsonResponse(http.StatusOK, `{"ok": true, "ts": "1234567890.123456"}`), nil
		})

		metadata := &contexts.MetadataContext{
			Metadata: SendAndWaitForResponseMetadata{
				State: "pending",
			},
		}
		requestCtx := &contexts.RequestContext{}

		err := component.Execute(core.ExecutionContext{
			ID:             executionID,
			Integration:    &contexts.IntegrationContext{Configuration: map[string]any{"botToken": "token-123"}},
			ExecutionState: &contexts.ExecutionStateContext{KVs: map[string]string{}},
			Metadata:       metadata,
			Requests:       requestCtx,
			Configuration: map[string]any{
				"channel": "C123",
				"message": "Choose one:",
				"timeout": timeout,
				"buttons": []map[string]any{
					{"name": "Yes", "value": "yes"},
				},
			},
		})

		require.NoError(t, err)
		assert.Equal(t, "timeout", requestCtx.Action)
		assert.Equal(t, time.Duration(timeout)*time.Second, requestCtx.Duration)
	})
}

func Test__SendAndWaitForResponse__HandleAction(t *testing.T) {
	component := &SendAndWaitForResponse{}

	t.Run("timeout action when waiting -> emits timeout", func(t *testing.T) {
		metadata := &contexts.MetadataContext{
			Metadata: SendAndWaitForResponseMetadata{
				MessageTS: "1234567890.123456",
				State:     "waiting",
			},
		}
		execState := &contexts.ExecutionStateContext{}

		err := component.HandleAction(core.ActionContext{
			Name:           "timeout",
			Metadata:       metadata,
			ExecutionState: execState,
		})

		require.NoError(t, err)
		assert.Equal(t, ChannelTimeout, execState.Channel)
		assert.Equal(t, "slack.response.timeout", execState.Type)

		stored, ok := metadata.Metadata.(SendAndWaitForResponseMetadata)
		require.True(t, ok)
		assert.Equal(t, "timeout", stored.State)
	})

	t.Run("timeout action when already responded -> does not emit", func(t *testing.T) {
		metadata := &contexts.MetadataContext{
			Metadata: SendAndWaitForResponseMetadata{
				MessageTS: "1234567890.123456",
				State:     "responded",
			},
		}
		execState := &contexts.ExecutionStateContext{}

		err := component.HandleAction(core.ActionContext{
			Name:           "timeout",
			Metadata:       metadata,
			ExecutionState: execState,
		})

		require.NoError(t, err)
		assert.Empty(t, execState.Channel)
	})

	t.Run("buttonClicked action -> emits received", func(t *testing.T) {
		metadata := &contexts.MetadataContext{
			Metadata: SendAndWaitForResponseMetadata{
				MessageTS: "1234567890.123456",
				State:     "waiting",
			},
		}
		execState := &contexts.ExecutionStateContext{}

		err := component.HandleAction(core.ActionContext{
			Name:     "buttonClicked",
			Metadata: metadata,
			Parameters: map[string]any{
				"value":      "approve",
				"responseTs": "1234567891.123456",
				"user": map[string]any{
					"id":   "U123",
					"name": "john",
				},
			},
			ExecutionState: execState,
		})

		require.NoError(t, err)
		assert.Equal(t, ChannelReceived, execState.Channel)
		assert.Equal(t, "slack.response.received", execState.Type)

		stored, ok := metadata.Metadata.(SendAndWaitForResponseMetadata)
		require.True(t, ok)
		assert.Equal(t, "responded", stored.State)
		assert.Equal(t, "approve", stored.SelectedValue)
		assert.Equal(t, "1234567891.123456", stored.ResponseTS)
	})

	t.Run("buttonClicked when already responded -> does not emit", func(t *testing.T) {
		metadata := &contexts.MetadataContext{
			Metadata: SendAndWaitForResponseMetadata{
				MessageTS:     "1234567890.123456",
				State:         "responded",
				SelectedValue: "approve",
			},
		}
		execState := &contexts.ExecutionStateContext{}

		err := component.HandleAction(core.ActionContext{
			Name:     "buttonClicked",
			Metadata: metadata,
			Parameters: map[string]any{
				"value": "reject",
			},
			ExecutionState: execState,
		})

		require.NoError(t, err)
		assert.Empty(t, execState.Channel)

		// State should not change
		stored, ok := metadata.Metadata.(SendAndWaitForResponseMetadata)
		require.True(t, ok)
		assert.Equal(t, "responded", stored.State)
		assert.Equal(t, "approve", stored.SelectedValue) // original value preserved
	})
}
