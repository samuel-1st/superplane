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

func Test__UpsertDNSRecord__Setup(t *testing.T) {
	component := &UpsertDNSRecord{}

	t.Run("missing domain returns error", func(t *testing.T) {
		err := component.Setup(core.SetupContext{
			Configuration: map[string]any{
				"type": "A",
				"name": "www",
				"data": "1.2.3.4",
			},
		})

		require.ErrorContains(t, err, "domain is required")
	})

	t.Run("missing type returns error", func(t *testing.T) {
		err := component.Setup(core.SetupContext{
			Configuration: map[string]any{
				"domain": "example.com",
				"name":   "www",
				"data":   "1.2.3.4",
			},
		})

		require.ErrorContains(t, err, "type is required")
	})

	t.Run("missing name returns error", func(t *testing.T) {
		err := component.Setup(core.SetupContext{
			Configuration: map[string]any{
				"domain": "example.com",
				"type":   "A",
				"data":   "1.2.3.4",
			},
		})

		require.ErrorContains(t, err, "name is required")
	})

	t.Run("missing data returns error", func(t *testing.T) {
		err := component.Setup(core.SetupContext{
			Configuration: map[string]any{
				"domain": "example.com",
				"type":   "A",
				"name":   "www",
			},
		})

		require.ErrorContains(t, err, "data is required")
	})

	t.Run("valid configuration -> no error", func(t *testing.T) {
		err := component.Setup(core.SetupContext{
			Configuration: map[string]any{
				"domain": "example.com",
				"type":   "A",
				"name":   "www",
				"data":   "1.2.3.4",
			},
		})

		require.NoError(t, err)
	})
}

func Test__UpsertDNSRecord__Execute(t *testing.T) {
	component := &UpsertDNSRecord{}

	t.Run("no existing record -> creates new record", func(t *testing.T) {
		httpContext := &contexts.HTTPContext{
			Responses: []*http.Response{
				// ListDNSRecords returns empty
				{
					StatusCode: http.StatusOK,
					Body:       io.NopCloser(strings.NewReader(`{"domain_records": []}`)),
				},
				// CreateDNSRecord
				{
					StatusCode: http.StatusCreated,
					Body: io.NopCloser(strings.NewReader(`{
						"domain_record": {
							"id": 12345678,
							"type": "A",
							"name": "www",
							"data": "1.2.3.4",
							"ttl": 1800
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
				"domain": "example.com",
				"type":   "A",
				"name":   "www",
				"data":   "1.2.3.4",
			},
			HTTP:           httpContext,
			Integration:    integrationCtx,
			ExecutionState: executionState,
		})

		require.NoError(t, err)
		assert.True(t, executionState.Passed)
		assert.Equal(t, "default", executionState.Channel)
		assert.Equal(t, "digitalocean.dns.record.upserted", executionState.Type)
	})

	t.Run("existing record -> updates it", func(t *testing.T) {
		httpContext := &contexts.HTTPContext{
			Responses: []*http.Response{
				// ListDNSRecords returns existing
				{
					StatusCode: http.StatusOK,
					Body: io.NopCloser(strings.NewReader(`{
						"domain_records": [
							{"id": 12345678, "type": "A", "name": "www", "data": "1.1.1.1", "ttl": 1800}
						]
					}`)),
				},
				// UpdateDNSRecord
				{
					StatusCode: http.StatusOK,
					Body: io.NopCloser(strings.NewReader(`{
						"domain_record": {
							"id": 12345678,
							"type": "A",
							"name": "www",
							"data": "1.2.3.4",
							"ttl": 1800
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
				"domain": "example.com",
				"type":   "A",
				"name":   "www",
				"data":   "1.2.3.4",
			},
			HTTP:           httpContext,
			Integration:    integrationCtx,
			ExecutionState: executionState,
		})

		require.NoError(t, err)
		assert.True(t, executionState.Passed)
		assert.Equal(t, "digitalocean.dns.record.upserted", executionState.Type)
	})

	t.Run("API error on list -> returns error", func(t *testing.T) {
		httpContext := &contexts.HTTPContext{
			Responses: []*http.Response{
				{
					StatusCode: http.StatusUnauthorized,
					Body:       io.NopCloser(strings.NewReader(`{"id":"unauthorized","message":"Unable to authenticate you."}`)),
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
				"domain": "example.com",
				"type":   "A",
				"name":   "www",
				"data":   "1.2.3.4",
			},
			HTTP:           httpContext,
			Integration:    integrationCtx,
			ExecutionState: executionState,
		})

		require.Error(t, err)
		assert.Contains(t, err.Error(), "failed to list DNS records")
	})
}
