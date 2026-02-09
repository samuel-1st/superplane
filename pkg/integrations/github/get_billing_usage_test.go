package github

import (
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/superplanehq/superplane/pkg/core"
	contexts "github.com/superplanehq/superplane/test/support/contexts"
)

func Test__GetBillingUsage__Setup(t *testing.T) {
	helloRepo := Repository{ID: 123456, Name: "hello", URL: "https://github.com/testhq/hello"}
	worldRepo := Repository{ID: 789012, Name: "world", URL: "https://github.com/testhq/world"}
	component := GetBillingUsage{}

	t.Run("validates year format", func(t *testing.T) {
		integrationCtx := &contexts.IntegrationContext{
			Metadata: Metadata{
				Repositories: []Repository{helloRepo},
			},
		}
		err := component.Setup(core.SetupContext{
			Integration:   integrationCtx,
			Metadata:      &contexts.MetadataContext{},
			Configuration: map[string]any{"year": "invalid"},
		})

		require.ErrorContains(t, err, "invalid year")
	})

	t.Run("validates month range", func(t *testing.T) {
		integrationCtx := &contexts.IntegrationContext{
			Metadata: Metadata{
				Repositories: []Repository{helloRepo},
			},
		}
		err := component.Setup(core.SetupContext{
			Integration:   integrationCtx,
			Metadata:      &contexts.MetadataContext{},
			Configuration: map[string]any{"month": "13"},
		})

		require.ErrorContains(t, err, "invalid month")
	})

	t.Run("validates day range", func(t *testing.T) {
		integrationCtx := &contexts.IntegrationContext{
			Metadata: Metadata{
				Repositories: []Repository{helloRepo},
			},
		}
		err := component.Setup(core.SetupContext{
			Integration:   integrationCtx,
			Metadata:      &contexts.MetadataContext{},
			Configuration: map[string]any{"day": "32"},
		})

		require.ErrorContains(t, err, "invalid day")
	})

	t.Run("validates repository accessibility when specified", func(t *testing.T) {
		integrationCtx := &contexts.IntegrationContext{
			Metadata: Metadata{
				Repositories: []Repository{helloRepo},
			},
		}
		err := component.Setup(core.SetupContext{
			Integration:   integrationCtx,
			Metadata:      &contexts.MetadataContext{},
			Configuration: map[string]any{"repositories": []string{"nonexistent"}},
		})

		require.ErrorContains(t, err, "repository nonexistent is not accessible")
	})

	t.Run("setup succeeds with valid configuration", func(t *testing.T) {
		integrationCtx := &contexts.IntegrationContext{
			Metadata: Metadata{
				Repositories: []Repository{helloRepo, worldRepo},
			},
		}

		nodeMetadataCtx := contexts.MetadataContext{}
		require.NoError(t, component.Setup(core.SetupContext{
			Integration: integrationCtx,
			Metadata:    &nodeMetadataCtx,
			Configuration: map[string]any{
				"repositories": []string{"hello", "world"},
				"year":         "2026",
				"month":        "2",
			},
		}))
	})

	t.Run("setup succeeds with no repositories specified", func(t *testing.T) {
		integrationCtx := &contexts.IntegrationContext{
			Metadata: Metadata{
				Repositories: []Repository{helloRepo},
			},
		}

		nodeMetadataCtx := contexts.MetadataContext{}
		require.NoError(t, component.Setup(core.SetupContext{
			Integration:   integrationCtx,
			Metadata:      &nodeMetadataCtx,
			Configuration: map[string]any{},
		}))
	})
}

func Test__GetBillingUsage__ParseAPIResponse(t *testing.T) {
	component := GetBillingUsage{}

	t.Run("parses usage items format", func(t *testing.T) {
		apiResponse := map[string]any{
			"usageItems": []any{
				map[string]any{
					"sku":       "UBUNTU",
					"quantity":  float64(1000),
					"netAmount": float64(10.50),
				},
				map[string]any{
					"sku":       "WINDOWS",
					"quantity":  float64(500),
					"netAmount": float64(5.25),
				},
			},
		}

		output := component.parseAPIResponse(apiResponse)

		require.Equal(t, int64(1500), output.MinutesUsed)
		require.Equal(t, int64(1000), output.MinutesUsedBreakdown["UBUNTU"])
		require.Equal(t, int64(500), output.MinutesUsedBreakdown["WINDOWS"])
		require.Equal(t, 15.75, output.TotalCost)
	})

	t.Run("parses direct fields format", func(t *testing.T) {
		apiResponse := map[string]any{
			"total_minutes_used": float64(2000),
			"minutes_used_breakdown": map[string]any{
				"UBUNTU":  float64(1500),
				"WINDOWS": float64(500),
			},
		}

		output := component.parseAPIResponse(apiResponse)

		require.Equal(t, int64(2000), output.MinutesUsed)
		require.Equal(t, int64(1500), output.MinutesUsedBreakdown["UBUNTU"])
		require.Equal(t, int64(500), output.MinutesUsedBreakdown["WINDOWS"])
	})

	t.Run("handles empty response", func(t *testing.T) {
		apiResponse := map[string]any{}

		output := component.parseAPIResponse(apiResponse)

		require.Equal(t, int64(0), output.MinutesUsed)
		require.NotNil(t, output.MinutesUsedBreakdown)
		require.Equal(t, 0, len(output.MinutesUsedBreakdown))
	})
}
