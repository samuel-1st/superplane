package digitalocean

import (
	"io"
	"net/http"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/superplanehq/superplane/pkg/core"
	"github.com/superplanehq/superplane/test/support/contexts"
)

func Test__ManageDropletPower__Setup(t *testing.T) {
	component := &ManageDropletPower{}

	t.Run("missing dropletID returns error", func(t *testing.T) {
		err := component.Setup(core.SetupContext{
			Configuration: map[string]any{
				"operation": "power_on",
			},
		})

		require.ErrorContains(t, err, "dropletID is required")
	})

	t.Run("missing operation returns error", func(t *testing.T) {
		err := component.Setup(core.SetupContext{
			Configuration: map[string]any{
				"dropletID": "98765432",
			},
		})

		require.ErrorContains(t, err, "operation is required")
	})

	t.Run("valid configuration -> no error", func(t *testing.T) {
		err := component.Setup(core.SetupContext{
			Configuration: map[string]any{
				"dropletID": "98765432",
				"operation": "reboot",
			},
		})

		require.NoError(t, err)
	})
}

func Test__ManageDropletPower__Execute(t *testing.T) {
	component := &ManageDropletPower{}

	t.Run("successful action dispatch -> stores metadata and schedules poll", func(t *testing.T) {
		httpContext := &contexts.HTTPContext{
			Responses: []*http.Response{
				{
					StatusCode: http.StatusCreated,
					Body: io.NopCloser(strings.NewReader(`{
						"action": {
							"id": 36804751,
							"status": "in-progress",
							"type": "power_on",
							"started_at": "2014-11-14T16:29:21Z",
							"completed_at": null,
							"resource_id": 98765432,
							"resource_type": "droplet",
							"region_slug": "nyc3"
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

		metadataCtx := &contexts.MetadataContext{}
		requestCtx := &contexts.RequestContext{}
		executionState := &contexts.ExecutionStateContext{
			KVs: map[string]string{},
		}

		err := component.Execute(core.ExecutionContext{
			Configuration: map[string]any{
				"dropletID": "98765432",
				"operation": "power_on",
			},
			HTTP:           httpContext,
			Integration:    integrationCtx,
			ExecutionState: executionState,
			Metadata:       metadataCtx,
			Requests:       requestCtx,
		})

		require.NoError(t, err)

		metadata, ok := metadataCtx.Metadata.(map[string]any)
		require.True(t, ok)
		assert.Equal(t, 36804751, metadata["actionID"])

		assert.Equal(t, "poll", requestCtx.Action)
		assert.Equal(t, 10*time.Second, requestCtx.Duration)
		assert.False(t, executionState.Passed)
	})

	t.Run("API error -> returns error", func(t *testing.T) {
		httpContext := &contexts.HTTPContext{
			Responses: []*http.Response{
				{
					StatusCode: http.StatusUnprocessableEntity,
					Body:       io.NopCloser(strings.NewReader(`{"id":"unprocessable_entity","message":"Droplet is already powered on"}`)),
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
				"dropletID": "98765432",
				"operation": "power_on",
			},
			HTTP:           httpContext,
			Integration:    integrationCtx,
			ExecutionState: executionState,
		})

		require.Error(t, err)
		assert.Contains(t, err.Error(), "failed to perform power action")
	})
}

func Test__ManageDropletPower__HandleAction(t *testing.T) {
	component := &ManageDropletPower{}

	t.Run("action completed -> emits result", func(t *testing.T) {
		httpContext := &contexts.HTTPContext{
			Responses: []*http.Response{
				{
					StatusCode: http.StatusOK,
					Body: io.NopCloser(strings.NewReader(`{
						"action": {
							"id": 36804751,
							"status": "completed",
							"type": "power_on",
							"started_at": "2014-11-14T16:29:21Z",
							"completed_at": "2014-11-14T16:30:06Z",
							"resource_id": 98765432,
							"resource_type": "droplet",
							"region_slug": "nyc3"
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

		metadataCtx := &contexts.MetadataContext{
			Metadata: map[string]any{"actionID": 36804751},
		}

		executionState := &contexts.ExecutionStateContext{
			KVs: map[string]string{},
		}

		err := component.HandleAction(core.ActionContext{
			Name:           "poll",
			HTTP:           httpContext,
			Integration:    integrationCtx,
			Metadata:       metadataCtx,
			ExecutionState: executionState,
			Requests:       &contexts.RequestContext{},
		})

		require.NoError(t, err)
		assert.True(t, executionState.Passed)
		assert.Equal(t, "default", executionState.Channel)
		assert.Equal(t, "digitalocean.droplet.power.managed", executionState.Type)
	})

	t.Run("action in progress -> schedules another poll", func(t *testing.T) {
		httpContext := &contexts.HTTPContext{
			Responses: []*http.Response{
				{
					StatusCode: http.StatusOK,
					Body: io.NopCloser(strings.NewReader(`{
						"action": {
							"id": 36804751,
							"status": "in-progress",
							"type": "power_on",
							"resource_id": 98765432,
							"resource_type": "droplet",
							"region_slug": "nyc3"
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

		metadataCtx := &contexts.MetadataContext{
			Metadata: map[string]any{"actionID": 36804751},
		}

		requestCtx := &contexts.RequestContext{}
		executionState := &contexts.ExecutionStateContext{
			KVs: map[string]string{},
		}

		err := component.HandleAction(core.ActionContext{
			Name:           "poll",
			HTTP:           httpContext,
			Integration:    integrationCtx,
			Metadata:       metadataCtx,
			ExecutionState: executionState,
			Requests:       requestCtx,
		})

		require.NoError(t, err)
		assert.False(t, executionState.Passed)
		assert.Equal(t, "poll", requestCtx.Action)
		assert.Equal(t, 10*time.Second, requestCtx.Duration)
	})

	t.Run("action errored -> returns error", func(t *testing.T) {
		httpContext := &contexts.HTTPContext{
			Responses: []*http.Response{
				{
					StatusCode: http.StatusOK,
					Body: io.NopCloser(strings.NewReader(`{
						"action": {
							"id": 36804751,
							"status": "errored",
							"type": "power_on",
							"resource_id": 98765432,
							"resource_type": "droplet",
							"region_slug": "nyc3"
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

		metadataCtx := &contexts.MetadataContext{
			Metadata: map[string]any{"actionID": 36804751},
		}

		executionState := &contexts.ExecutionStateContext{
			KVs: map[string]string{},
		}

		err := component.HandleAction(core.ActionContext{
			Name:           "poll",
			HTTP:           httpContext,
			Integration:    integrationCtx,
			Metadata:       metadataCtx,
			ExecutionState: executionState,
			Requests:       &contexts.RequestContext{},
		})

		require.Error(t, err)
		assert.Contains(t, err.Error(), "errored")
		assert.False(t, executionState.Passed)
	})
}
