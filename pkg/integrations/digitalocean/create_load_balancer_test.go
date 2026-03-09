package digitalocean

import (
	"io"
	"net/http"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/superplanehq/superplane/pkg/core"
	"github.com/superplanehq/superplane/test/support/contexts"
)

func Test__CreateLoadBalancer__Setup(t *testing.T) {
	component := &CreateLoadBalancer{}

	t.Run("missing name returns error", func(t *testing.T) {
		err := component.Setup(core.SetupContext{
			Configuration: map[string]any{
				"region":         "nyc3",
				"entryProtocol":  "http",
				"entryPort":      "80",
				"targetProtocol": "http",
				"targetPort":     "8080",
			},
		})

		require.ErrorContains(t, err, "name is required")
	})

	t.Run("missing region returns error", func(t *testing.T) {
		err := component.Setup(core.SetupContext{
			Configuration: map[string]any{
				"name":           "my-lb",
				"entryProtocol":  "http",
				"entryPort":      "80",
				"targetProtocol": "http",
				"targetPort":     "8080",
			},
		})

		require.ErrorContains(t, err, "region is required")
	})

	t.Run("missing entryProtocol returns error", func(t *testing.T) {
		err := component.Setup(core.SetupContext{
			Configuration: map[string]any{
				"name":           "my-lb",
				"region":         "nyc3",
				"entryPort":      "80",
				"targetProtocol": "http",
				"targetPort":     "8080",
			},
		})

		require.ErrorContains(t, err, "entryProtocol is required")
	})

	t.Run("missing entryPort returns error", func(t *testing.T) {
		err := component.Setup(core.SetupContext{
			Configuration: map[string]any{
				"name":           "my-lb",
				"region":         "nyc3",
				"entryProtocol":  "http",
				"targetProtocol": "http",
				"targetPort":     "8080",
			},
		})

		require.ErrorContains(t, err, "entryPort is required")
	})

	t.Run("missing targetProtocol returns error", func(t *testing.T) {
		err := component.Setup(core.SetupContext{
			Configuration: map[string]any{
				"name":          "my-lb",
				"region":        "nyc3",
				"entryProtocol": "http",
				"entryPort":     "80",
				"targetPort":    "8080",
			},
		})

		require.ErrorContains(t, err, "targetProtocol is required")
	})

	t.Run("missing targetPort returns error", func(t *testing.T) {
		err := component.Setup(core.SetupContext{
			Configuration: map[string]any{
				"name":           "my-lb",
				"region":         "nyc3",
				"entryProtocol":  "http",
				"entryPort":      "80",
				"targetProtocol": "http",
			},
		})

		require.ErrorContains(t, err, "targetPort is required")
	})

	t.Run("valid configuration -> no error", func(t *testing.T) {
		err := component.Setup(core.SetupContext{
			Configuration: map[string]any{
				"name":           "my-lb",
				"region":         "nyc3",
				"entryProtocol":  "http",
				"entryPort":      "80",
				"targetProtocol": "http",
				"targetPort":     "8080",
			},
		})

		require.NoError(t, err)
	})
}

func Test__CreateLoadBalancer__Execute(t *testing.T) {
	component := &CreateLoadBalancer{}

	t.Run("successful creation -> emits load balancer", func(t *testing.T) {
		httpContext := &contexts.HTTPContext{
			Responses: []*http.Response{
				{
					StatusCode: http.StatusAccepted,
					Body: io.NopCloser(strings.NewReader(`{
						"load_balancer": {
							"id": "4de7ac8b-495b-4884-9a69-1050bx3b15ed",
							"name": "my-lb",
							"ip": "",
							"status": "new",
							"size_slug": "lb-small",
							"region": {"name": "New York 3", "slug": "nyc3"},
							"forwarding_rules": [
								{"entry_protocol": "http", "entry_port": 80, "target_protocol": "http", "target_port": 8080}
							],
							"droplet_ids": [],
							"tag": []
						}
					}`)),
				},
			},
		}

		integrationCtx := &contexts.IntegrationContext{
			Configuration: map[string]any{
				"apiToken": "test-token",
			},
		}

		executionState := &contexts.ExecutionStateContext{
			KVs: map[string]string{},
		}

		err := component.Execute(core.ExecutionContext{
			Configuration: map[string]any{
				"name":           "my-lb",
				"region":         "nyc3",
				"entryProtocol":  "http",
				"entryPort":      "80",
				"targetProtocol": "http",
				"targetPort":     "8080",
			},
			HTTP:           httpContext,
			Integration:    integrationCtx,
			ExecutionState: executionState,
		})

		require.NoError(t, err)
		assert.True(t, executionState.Passed)
		assert.Equal(t, "default", executionState.Channel)
		assert.Equal(t, "digitalocean.load_balancer.created", executionState.Type)
	})

	t.Run("API error -> returns error", func(t *testing.T) {
		httpContext := &contexts.HTTPContext{
			Responses: []*http.Response{
				{
					StatusCode: http.StatusUnprocessableEntity,
					Body:       io.NopCloser(strings.NewReader(`{"id":"unprocessable_entity","message":"Name is already in use"}`)),
				},
			},
		}

		integrationCtx := &contexts.IntegrationContext{
			Configuration: map[string]any{
				"apiToken": "test-token",
			},
		}

		executionState := &contexts.ExecutionStateContext{
			KVs: map[string]string{},
		}

		err := component.Execute(core.ExecutionContext{
			Configuration: map[string]any{
				"name":           "my-lb",
				"region":         "nyc3",
				"entryProtocol":  "http",
				"entryPort":      "80",
				"targetProtocol": "http",
				"targetPort":     "8080",
			},
			HTTP:           httpContext,
			Integration:    integrationCtx,
			ExecutionState: executionState,
		})

		require.Error(t, err)
		assert.Contains(t, err.Error(), "failed to create load balancer")
	})
}
