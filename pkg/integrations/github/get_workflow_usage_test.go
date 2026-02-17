package github

import (
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/superplanehq/superplane/pkg/core"
	contexts "github.com/superplanehq/superplane/test/support/contexts"
)

func Test__GetWorkflowUsage__Setup(t *testing.T) {
	component := GetWorkflowUsage{}

	t.Run("setup succeeds with no configuration", func(t *testing.T) {
		integrationCtx := &contexts.IntegrationContext{}
		err := component.Setup(core.SetupContext{
			Integration:   integrationCtx,
			Metadata:      &contexts.MetadataContext{},
			Configuration: map[string]any{},
		})

		require.NoError(t, err)
	})
}

func Test__GetWorkflowUsage__Execute(t *testing.T) {
	component := GetWorkflowUsage{}

	t.Run("fails when configuration decode fails", func(t *testing.T) {
		err := component.Execute(core.ExecutionContext{
			Integration:    &contexts.IntegrationContext{},
			ExecutionState: &contexts.ExecutionStateContext{},
			Configuration:  "not a map",
		})

		require.ErrorContains(t, err, "failed to decode configuration")
	})

	t.Run("fails when metadata decode fails", func(t *testing.T) {
		integrationCtx := &contexts.IntegrationContext{
			Metadata: "not a valid metadata",
		}
		err := component.Execute(core.ExecutionContext{
			Integration:    integrationCtx,
			ExecutionState: &contexts.ExecutionStateContext{},
			Configuration:  map[string]any{},
		})

		require.ErrorContains(t, err, "failed to decode application metadata")
	})
}

func Test__GetWorkflowUsage__Name(t *testing.T) {
	component := GetWorkflowUsage{}
	require.Equal(t, "github.getWorkflowUsage", component.Name())
}

func Test__GetWorkflowUsage__Label(t *testing.T) {
	component := GetWorkflowUsage{}
	require.Equal(t, "Get Workflow Usage", component.Label())
}

func Test__GetWorkflowUsage__ExampleOutput(t *testing.T) {
	component := GetWorkflowUsage{}
	output := component.ExampleOutput()

	require.NotNil(t, output)
	require.Contains(t, output, "data")
	require.Contains(t, output, "type")
	require.Equal(t, "github.workflowUsage", output["type"])
}
