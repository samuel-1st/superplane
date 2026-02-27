package cloudsmith

import (
	"encoding/json"
	"io"
	"net/http"
	"strings"
	"testing"

	log "github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/superplanehq/superplane/pkg/core"
	"github.com/superplanehq/superplane/test/support/contexts"
)

func Test__OnPackageEvent__Setup(t *testing.T) {
	trigger := &OnPackageEvent{}

	t.Run("repository is required", func(t *testing.T) {
		err := trigger.Setup(core.TriggerContext{
			Integration:   &contexts.IntegrationContext{},
			Metadata:      &contexts.MetadataContext{},
			Configuration: map[string]any{"repository": ""},
		})

		require.ErrorContains(t, err, "repository is required")
	})

	t.Run("invalid repository format -> error", func(t *testing.T) {
		err := trigger.Setup(core.TriggerContext{
			Integration:   &contexts.IntegrationContext{},
			Metadata:      &contexts.MetadataContext{},
			Configuration: map[string]any{"repository": "invalid"},
		})

		require.ErrorContains(t, err, "repository must be in the format of namespace/repo")
	})

	t.Run("valid configuration -> creates webhook and stores metadata", func(t *testing.T) {
		httpCtx := &contexts.HTTPContext{
			Responses: []*http.Response{
				{
					StatusCode: http.StatusCreated,
					Body:       io.NopCloser(strings.NewReader(`{"slug_perm":"abc123"}`)),
				},
			},
		}

		metadata := &contexts.MetadataContext{}
		integrationCtx := &contexts.IntegrationContext{
			Configuration: map[string]any{
				"apiKey":    "test-api-key",
				"workspace": "my-org",
			},
		}

		err := trigger.Setup(core.TriggerContext{
			HTTP:        httpCtx,
			Integration: integrationCtx,
			Metadata:    metadata,
			Webhook:     &contexts.WebhookContext{},
			Configuration: map[string]any{
				"repository": "my-org/my-repo",
				"events":     []string{"package.synced"},
			},
		})

		require.NoError(t, err)
		stored, ok := metadata.Metadata.(OnPackageEventMetadata)
		require.True(t, ok)
		assert.Equal(t, "my-org/my-repo", stored.Repository)
		assert.Equal(t, "abc123", stored.WebhookSlug)
		assert.NotEmpty(t, stored.WebhookURL)

		// Verify the HTTP request used the correct "target_url" field (not "webhook_url")
		require.Len(t, httpCtx.Requests, 1)
		bodyBytes, err := io.ReadAll(httpCtx.Requests[0].Body)
		require.NoError(t, err)
		var reqBody map[string]any
		require.NoError(t, json.Unmarshal(bodyBytes, &reqBody))
		assert.Contains(t, reqBody, "target_url", "request body must use 'target_url' per the Cloudsmith API spec")
		assert.NotContains(t, reqBody, "webhook_url", "request body must not use 'webhook_url'")
	})

	t.Run("already configured -> skips setup", func(t *testing.T) {
		httpCtx := &contexts.HTTPContext{}
		metadata := &contexts.MetadataContext{
			Metadata: OnPackageEventMetadata{
				Repository:  "my-org/my-repo",
				WebhookSlug: "existing-slug",
				WebhookURL:  "https://example.com/hook",
			},
		}
		integrationCtx := &contexts.IntegrationContext{
			Configuration: map[string]any{
				"apiKey":    "test-api-key",
				"workspace": "my-org",
			},
		}

		err := trigger.Setup(core.TriggerContext{
			HTTP:        httpCtx,
			Integration: integrationCtx,
			Metadata:    metadata,
			Webhook:     &contexts.WebhookContext{},
			Configuration: map[string]any{
				"repository": "my-org/my-repo",
				"events":     []string{"package.synced"},
			},
		})

		require.NoError(t, err)
		// No HTTP requests should have been made
		assert.Len(t, httpCtx.Requests, 0)
	})
}

func Test__OnPackageEvent__HandleWebhook(t *testing.T) {
	trigger := &OnPackageEvent{}

	t.Run("invalid JSON -> 400", func(t *testing.T) {
		code, err := trigger.HandleWebhook(core.WebhookRequestContext{
			Body:          []byte(`invalid`),
			Events:        &contexts.EventContext{},
			Configuration: map[string]any{"repository": "my-org/my-repo", "events": []string{"package.synced"}},
			Metadata:      &contexts.MetadataContext{},
			Logger:        log.NewEntry(log.New()),
		})

		assert.Equal(t, http.StatusBadRequest, code)
		assert.ErrorContains(t, err, "error parsing request body")
	})

	t.Run("event filter mismatch -> ignored", func(t *testing.T) {
		body := []byte(`{"event":"package.deleted"}`)
		events := &contexts.EventContext{}
		code, err := trigger.HandleWebhook(core.WebhookRequestContext{
			Body:   body,
			Events: events,
			Logger: log.NewEntry(log.New()),
			Metadata: &contexts.MetadataContext{
				Metadata: OnPackageEventMetadata{
					Repository: "my-org/my-repo",
				},
			},
			Configuration: map[string]any{
				"repository": "my-org/my-repo",
				"events":     []string{"package.synced"},
			},
		})

		assert.Equal(t, http.StatusOK, code)
		require.NoError(t, err)
		assert.Equal(t, 0, events.Count())
	})

	t.Run("matching event -> event emitted", func(t *testing.T) {
		body := []byte(`{"event":"package.synced","data":{"package":{"name":"my-lib","version":"1.0.0"}}}`)
		events := &contexts.EventContext{}
		code, err := trigger.HandleWebhook(core.WebhookRequestContext{
			Body:   body,
			Events: events,
			Logger: log.NewEntry(log.New()),
			Metadata: &contexts.MetadataContext{
				Metadata: OnPackageEventMetadata{
					Repository: "my-org/my-repo",
				},
			},
			Configuration: map[string]any{
				"repository": "my-org/my-repo",
				"events":     []string{"package.synced"},
			},
		})

		assert.Equal(t, http.StatusOK, code)
		require.NoError(t, err)
		require.Equal(t, 1, events.Count())
		assert.Equal(t, "cloudsmith.package.event", events.Payloads[0].Type)
	})
}
