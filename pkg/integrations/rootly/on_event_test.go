package rootly

import (
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/superplanehq/superplane/pkg/core"
	"github.com/superplanehq/superplane/test/support/contexts"
)

func Test__OnEvent__HandleWebhook(t *testing.T) {
	trigger := &OnEvent{}

	validConfig := map[string]any{
		"events": []string{"incident_event.created"},
	}

	signatureFor := func(secret string, timestamp string, body []byte) string {
		payload := append([]byte(timestamp), body...)
		sig := computeHMACSHA256([]byte(secret), payload)
		return "t=" + timestamp + ",v1=" + sig
	}

	t.Run("missing X-Rootly-Signature -> 403", func(t *testing.T) {
		code, err := trigger.HandleWebhook(core.WebhookRequestContext{
			Headers:       http.Header{},
			Configuration: validConfig,
			Webhook:       &contexts.WebhookContext{Secret: "test-secret"},
			Events:        &contexts.EventContext{},
		})

		assert.Equal(t, http.StatusForbidden, code)
		assert.ErrorContains(t, err, "invalid signature")
	})

	t.Run("invalid signature -> 403", func(t *testing.T) {
		body := []byte(`{"event":{"type":"incident_event.created","id":"evt-123","issued_at":"2025-01-01T00:00:00Z"},"data":{"id":"ev-456","kind":"note","visibility":"internal"}}`)
		headers := http.Header{}
		headers.Set("X-Rootly-Signature", "t=1234567890,v1=invalid")

		code, err := trigger.HandleWebhook(core.WebhookRequestContext{
			Body:          body,
			Headers:       headers,
			Configuration: validConfig,
			Webhook:       &contexts.WebhookContext{Secret: "test-secret"},
			Events:        &contexts.EventContext{},
		})

		assert.Equal(t, http.StatusForbidden, code)
		assert.ErrorContains(t, err, "invalid signature")
	})

	t.Run("invalid JSON body -> 400", func(t *testing.T) {
		body := []byte("invalid json")
		secret := "test-secret"
		timestamp := "1234567890"

		headers := http.Header{}
		headers.Set("X-Rootly-Signature", signatureFor(secret, timestamp, body))

		code, err := trigger.HandleWebhook(core.WebhookRequestContext{
			Body:          body,
			Headers:       headers,
			Configuration: validConfig,
			Webhook:       &contexts.WebhookContext{Secret: secret},
			Events:        &contexts.EventContext{},
		})

		assert.Equal(t, http.StatusBadRequest, code)
		assert.ErrorContains(t, err, "error parsing request body")
	})

	t.Run("event type not configured -> no emit", func(t *testing.T) {
		body := []byte(`{"event":{"type":"incident_event.updated","id":"evt-123","issued_at":"2025-01-01T00:00:00Z"},"data":{"id":"ev-456","kind":"note","visibility":"internal"}}`)
		secret := "test-secret"
		timestamp := "1234567890"

		headers := http.Header{}
		headers.Set("X-Rootly-Signature", signatureFor(secret, timestamp, body))

		eventContext := &contexts.EventContext{}
		code, err := trigger.HandleWebhook(core.WebhookRequestContext{
			Body:          body,
			Headers:       headers,
			Configuration: validConfig, // only incident_event.created configured
			Webhook:       &contexts.WebhookContext{Secret: secret},
			Events:        eventContext,
		})

		assert.Equal(t, http.StatusOK, code)
		assert.NoError(t, err)
		assert.Equal(t, 0, eventContext.Count())
	})

	t.Run("valid event -> emitted", func(t *testing.T) {
		body := []byte(`{"event":{"type":"incident_event.created","id":"evt-123","issued_at":"2025-01-01T00:00:00Z"},"data":{"id":"ev-456","kind":"note","visibility":"internal","source":"web_ui","occurred_at":"2025-01-01T00:00:00Z","created_at":"2025-01-01T00:00:00Z","user_display_name":"Jane Smith","incident":{"id":"inc-789","title":"Production down","status":"started","severity":"sev1"}}}`)
		secret := "test-secret"
		timestamp := "1234567890"

		headers := http.Header{}
		headers.Set("X-Rootly-Signature", signatureFor(secret, timestamp, body))

		eventContext := &contexts.EventContext{}
		code, err := trigger.HandleWebhook(core.WebhookRequestContext{
			Body:          body,
			Headers:       headers,
			Configuration: validConfig,
			Webhook:       &contexts.WebhookContext{Secret: secret},
			Events:        eventContext,
		})

		require.Equal(t, http.StatusOK, code)
		require.NoError(t, err)
		require.Equal(t, 1, eventContext.Count())

		payload := eventContext.Payloads[0]
		assert.Equal(t, "rootly.incident_event.created", payload.Type)
		data := payload.Data.(map[string]any)
		assert.Equal(t, "incident_event.created", data["event"])
		assert.Equal(t, "evt-123", data["event_id"])
		assert.Equal(t, "note", data["kind"])
		assert.Equal(t, "Jane Smith", data["user_display_name"])
		assert.NotNil(t, data["incident"])
	})

	t.Run("visibility filter matches -> emitted", func(t *testing.T) {
		config := map[string]any{
			"events":     []string{"incident_event.created"},
			"visibility": "internal",
		}

		body := []byte(`{"event":{"type":"incident_event.created","id":"evt-1","issued_at":"2025-01-01T00:00:00Z"},"data":{"id":"ev-1","kind":"note","visibility":"internal","incident":{"id":"inc-1","status":"started","severity":"sev1"}}}`)
		secret := "test-secret"
		timestamp := "1234567890"

		headers := http.Header{}
		headers.Set("X-Rootly-Signature", signatureFor(secret, timestamp, body))

		eventContext := &contexts.EventContext{}
		code, err := trigger.HandleWebhook(core.WebhookRequestContext{
			Body:          body,
			Headers:       headers,
			Configuration: config,
			Webhook:       &contexts.WebhookContext{Secret: secret},
			Events:        eventContext,
		})

		require.Equal(t, http.StatusOK, code)
		require.NoError(t, err)
		assert.Equal(t, 1, eventContext.Count())
	})

	t.Run("visibility filter no match -> no emit", func(t *testing.T) {
		config := map[string]any{
			"events":     []string{"incident_event.created"},
			"visibility": "external",
		}

		body := []byte(`{"event":{"type":"incident_event.created","id":"evt-1","issued_at":"2025-01-01T00:00:00Z"},"data":{"id":"ev-1","kind":"note","visibility":"internal","incident":{"id":"inc-1","status":"started","severity":"sev1"}}}`)
		secret := "test-secret"
		timestamp := "1234567890"

		headers := http.Header{}
		headers.Set("X-Rootly-Signature", signatureFor(secret, timestamp, body))

		eventContext := &contexts.EventContext{}
		code, err := trigger.HandleWebhook(core.WebhookRequestContext{
			Body:          body,
			Headers:       headers,
			Configuration: config,
			Webhook:       &contexts.WebhookContext{Secret: secret},
			Events:        eventContext,
		})

		require.Equal(t, http.StatusOK, code)
		require.NoError(t, err)
		assert.Equal(t, 0, eventContext.Count())
	})

	t.Run("event kind filter matches -> emitted", func(t *testing.T) {
		config := map[string]any{
			"events":    []string{"incident_event.created"},
			"eventKind": []string{"note"},
		}

		body := []byte(`{"event":{"type":"incident_event.created","id":"evt-1","issued_at":"2025-01-01T00:00:00Z"},"data":{"id":"ev-1","kind":"note","visibility":"internal","incident":{"id":"inc-1","status":"started"}}}`)
		secret := "test-secret"
		timestamp := "1234567890"

		headers := http.Header{}
		headers.Set("X-Rootly-Signature", signatureFor(secret, timestamp, body))

		eventContext := &contexts.EventContext{}
		code, err := trigger.HandleWebhook(core.WebhookRequestContext{
			Body:          body,
			Headers:       headers,
			Configuration: config,
			Webhook:       &contexts.WebhookContext{Secret: secret},
			Events:        eventContext,
		})

		require.Equal(t, http.StatusOK, code)
		require.NoError(t, err)
		assert.Equal(t, 1, eventContext.Count())
	})

	t.Run("event kind filter no match -> no emit", func(t *testing.T) {
		config := map[string]any{
			"events":    []string{"incident_event.created"},
			"eventKind": []string{"annotation"},
		}

		body := []byte(`{"event":{"type":"incident_event.created","id":"evt-1","issued_at":"2025-01-01T00:00:00Z"},"data":{"id":"ev-1","kind":"note","visibility":"internal","incident":{"id":"inc-1","status":"started"}}}`)
		secret := "test-secret"
		timestamp := "1234567890"

		headers := http.Header{}
		headers.Set("X-Rootly-Signature", signatureFor(secret, timestamp, body))

		eventContext := &contexts.EventContext{}
		code, err := trigger.HandleWebhook(core.WebhookRequestContext{
			Body:          body,
			Headers:       headers,
			Configuration: config,
			Webhook:       &contexts.WebhookContext{Secret: secret},
			Events:        eventContext,
		})

		require.Equal(t, http.StatusOK, code)
		require.NoError(t, err)
		assert.Equal(t, 0, eventContext.Count())
	})

	t.Run("incident status filter matches -> emitted", func(t *testing.T) {
		config := map[string]any{
			"events":         []string{"incident_event.created"},
			"incidentStatus": []string{"started", "mitigated"},
		}

		body := []byte(`{"event":{"type":"incident_event.created","id":"evt-1","issued_at":"2025-01-01T00:00:00Z"},"data":{"id":"ev-1","kind":"note","incident":{"id":"inc-1","status":"started"}}}`)
		secret := "test-secret"
		timestamp := "1234567890"

		headers := http.Header{}
		headers.Set("X-Rootly-Signature", signatureFor(secret, timestamp, body))

		eventContext := &contexts.EventContext{}
		code, err := trigger.HandleWebhook(core.WebhookRequestContext{
			Body:          body,
			Headers:       headers,
			Configuration: config,
			Webhook:       &contexts.WebhookContext{Secret: secret},
			Events:        eventContext,
		})

		require.Equal(t, http.StatusOK, code)
		require.NoError(t, err)
		assert.Equal(t, 1, eventContext.Count())
	})

	t.Run("incident status filter no match -> no emit", func(t *testing.T) {
		config := map[string]any{
			"events":         []string{"incident_event.created"},
			"incidentStatus": []string{"resolved"},
		}

		body := []byte(`{"event":{"type":"incident_event.created","id":"evt-1","issued_at":"2025-01-01T00:00:00Z"},"data":{"id":"ev-1","kind":"note","incident":{"id":"inc-1","status":"started"}}}`)
		secret := "test-secret"
		timestamp := "1234567890"

		headers := http.Header{}
		headers.Set("X-Rootly-Signature", signatureFor(secret, timestamp, body))

		eventContext := &contexts.EventContext{}
		code, err := trigger.HandleWebhook(core.WebhookRequestContext{
			Body:          body,
			Headers:       headers,
			Configuration: config,
			Webhook:       &contexts.WebhookContext{Secret: secret},
			Events:        eventContext,
		})

		require.Equal(t, http.StatusOK, code)
		require.NoError(t, err)
		assert.Equal(t, 0, eventContext.Count())
	})

	t.Run("severity filter matches -> emitted", func(t *testing.T) {
		config := map[string]any{
			"events":   []string{"incident_event.created"},
			"severity": []string{"sev1"},
		}

		body := []byte(`{"event":{"type":"incident_event.created","id":"evt-1","issued_at":"2025-01-01T00:00:00Z"},"data":{"id":"ev-1","kind":"note","incident":{"id":"inc-1","status":"started","severity":"sev1"}}}`)
		secret := "test-secret"
		timestamp := "1234567890"

		headers := http.Header{}
		headers.Set("X-Rootly-Signature", signatureFor(secret, timestamp, body))

		eventContext := &contexts.EventContext{}
		code, err := trigger.HandleWebhook(core.WebhookRequestContext{
			Body:          body,
			Headers:       headers,
			Configuration: config,
			Webhook:       &contexts.WebhookContext{Secret: secret},
			Events:        eventContext,
		})

		require.Equal(t, http.StatusOK, code)
		require.NoError(t, err)
		assert.Equal(t, 1, eventContext.Count())
	})

	t.Run("service filter matches -> emitted", func(t *testing.T) {
		config := map[string]any{
			"events":  []string{"incident_event.created"},
			"service": []string{"payments-service"},
		}

		body := []byte(`{"event":{"type":"incident_event.created","id":"evt-1","issued_at":"2025-01-01T00:00:00Z"},"data":{"id":"ev-1","kind":"note","incident":{"id":"inc-1","status":"started","services":["payments-service","auth-service"]}}}`)
		secret := "test-secret"
		timestamp := "1234567890"

		headers := http.Header{}
		headers.Set("X-Rootly-Signature", signatureFor(secret, timestamp, body))

		eventContext := &contexts.EventContext{}
		code, err := trigger.HandleWebhook(core.WebhookRequestContext{
			Body:          body,
			Headers:       headers,
			Configuration: config,
			Webhook:       &contexts.WebhookContext{Secret: secret},
			Events:        eventContext,
		})

		require.Equal(t, http.StatusOK, code)
		require.NoError(t, err)
		assert.Equal(t, 1, eventContext.Count())
	})

	t.Run("service filter no match -> no emit", func(t *testing.T) {
		config := map[string]any{
			"events":  []string{"incident_event.created"},
			"service": []string{"database-service"},
		}

		body := []byte(`{"event":{"type":"incident_event.created","id":"evt-1","issued_at":"2025-01-01T00:00:00Z"},"data":{"id":"ev-1","kind":"note","incident":{"id":"inc-1","status":"started","services":["payments-service"]}}}`)
		secret := "test-secret"
		timestamp := "1234567890"

		headers := http.Header{}
		headers.Set("X-Rootly-Signature", signatureFor(secret, timestamp, body))

		eventContext := &contexts.EventContext{}
		code, err := trigger.HandleWebhook(core.WebhookRequestContext{
			Body:          body,
			Headers:       headers,
			Configuration: config,
			Webhook:       &contexts.WebhookContext{Secret: secret},
			Events:        eventContext,
		})

		require.Equal(t, http.StatusOK, code)
		require.NoError(t, err)
		assert.Equal(t, 0, eventContext.Count())
	})

	t.Run("event source filter matches -> emitted", func(t *testing.T) {
		config := map[string]any{
			"events":      []string{"incident_event.created"},
			"eventSource": "web_ui",
		}

		body := []byte(`{"event":{"type":"incident_event.created","id":"evt-1","issued_at":"2025-01-01T00:00:00Z"},"data":{"id":"ev-1","kind":"note","source":"web_ui","incident":{"id":"inc-1","status":"started"}}}`)
		secret := "test-secret"
		timestamp := "1234567890"

		headers := http.Header{}
		headers.Set("X-Rootly-Signature", signatureFor(secret, timestamp, body))

		eventContext := &contexts.EventContext{}
		code, err := trigger.HandleWebhook(core.WebhookRequestContext{
			Body:          body,
			Headers:       headers,
			Configuration: config,
			Webhook:       &contexts.WebhookContext{Secret: secret},
			Events:        eventContext,
		})

		require.Equal(t, http.StatusOK, code)
		require.NoError(t, err)
		assert.Equal(t, 1, eventContext.Count())
	})
}

func Test__OnEvent__Setup(t *testing.T) {
	trigger := &OnEvent{}

	t.Run("invalid configuration -> decode error", func(t *testing.T) {
		integrationCtx := &contexts.IntegrationContext{}
		err := trigger.Setup(core.TriggerContext{
			Integration:   integrationCtx,
			Configuration: "invalid-config",
		})

		require.ErrorContains(t, err, "failed to decode configuration")
	})

	t.Run("at least one event required", func(t *testing.T) {
		integrationCtx := &contexts.IntegrationContext{}
		err := trigger.Setup(core.TriggerContext{
			Integration:   integrationCtx,
			Configuration: OnEventConfiguration{Events: []string{}},
		})

		require.ErrorContains(t, err, "at least one event type must be chosen")
	})

	t.Run("valid configuration -> webhook request", func(t *testing.T) {
		integrationCtx := &contexts.IntegrationContext{}
		err := trigger.Setup(core.TriggerContext{
			Integration:   integrationCtx,
			Configuration: OnEventConfiguration{Events: []string{"incident_event.created"}},
		})

		require.NoError(t, err)
		require.Len(t, integrationCtx.WebhookRequests, 1)

		webhookConfig := integrationCtx.WebhookRequests[0].(WebhookConfiguration)
		assert.Equal(t, []string{"incident_event.created"}, webhookConfig.Events)
	})

	t.Run("both event types -> webhook request with both", func(t *testing.T) {
		integrationCtx := &contexts.IntegrationContext{}
		err := trigger.Setup(core.TriggerContext{
			Integration: integrationCtx,
			Configuration: OnEventConfiguration{
				Events: []string{"incident_event.created", "incident_event.updated"},
			},
		})

		require.NoError(t, err)
		require.Len(t, integrationCtx.WebhookRequests, 1)

		webhookConfig := integrationCtx.WebhookRequests[0].(WebhookConfiguration)
		assert.Len(t, webhookConfig.Events, 2)
	})
}
