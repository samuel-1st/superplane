package github

import (
	"context"
	_ "embed"
	"fmt"
	"sync"

	gh "github.com/google/go-github/v74/github"
	"github.com/google/uuid"
	"github.com/mitchellh/mapstructure"
	"github.com/superplanehq/superplane/pkg/configuration"
	"github.com/superplanehq/superplane/pkg/core"
	"github.com/superplanehq/superplane/pkg/utils"
)

type GetWorkflowUsage struct{}

//go:embed get_workflow_usage_example_output.json
var getWorkflowUsageExampleOutputBytes []byte

var getWorkflowUsageExampleOutputOnce sync.Once
var getWorkflowUsageExampleOutput map[string]any

type GetWorkflowUsageConfiguration struct {
	Repositories []string `mapstructure:"repositories"`
}

type WorkflowUsageResult struct {
	MinutesUsed          float64                 `json:"minutes_used" mapstructure:"minutes_used"`
	MinutesUsedBreakdown gh.MinutesUsedBreakdown `json:"minutes_used_breakdown" mapstructure:"minutes_used_breakdown"`
	IncludedMinutes      float64                 `json:"included_minutes" mapstructure:"included_minutes"`
	TotalPaidMinutesUsed float64                 `json:"total_paid_minutes_used" mapstructure:"total_paid_minutes_used"`
}

func (g *GetWorkflowUsage) Name() string {
	return "github.getWorkflowUsage"
}

func (g *GetWorkflowUsage) Label() string {
	return "Get Workflow Usage"
}

func (g *GetWorkflowUsage) Description() string {
	return "Retrieve billable GitHub Actions usage (minutes) for the organization"
}

func (g *GetWorkflowUsage) Documentation() string {
	return `The Get Workflow Usage component retrieves billable GitHub Actions usage (minutes) for the installation's organization.

## Prerequisites

This action calls GitHub's **billing usage** API, which requires the GitHub App to have **Organization permission: Organization administration (read)**. 

**Important**: Existing installations will need to approve the new permission when prompted by GitHub. Until the permission is granted, this action will return a 403 error.

## Behavior

- Returns billing data for the **current billing cycle** using GitHub's enhanced billing API
- Only private repositories on GitHub-hosted runners accrue billable minutes
- Public repositories and self-hosted runners show zero billable usage
- Usage can be filtered by a specific repository or retrieved organization-wide

## Configuration

- **Repositories** (optional, multiselect): Select one or more specific repositories to check usage for. Leave empty for organization-wide usage.

## Output

Returns usage data with:
- ` + "`minutes_used`" + `: Total billable minutes used in the current billing cycle
- ` + "`minutes_used_breakdown`" + `: Map of minutes by runner SKU (e.g., "Actions Linux": 120, "Actions Windows": 60, "Actions macOS": 30)
- ` + "`included_minutes`" + `: Not provided by enhanced billing API (always 0)
- ` + "`total_paid_minutes_used`" + `: Estimated paid minutes based on cost data

**Note**: Breakdown is by runner SKU (OS and type), not by individual workflow.

## Use Cases

- **Billing Monitoring**: Track GitHub Actions usage for billing purposes
- **Quota Management**: Monitor usage to avoid exceeding billing quotas
- **Cost Control**: Alert when usage approaches limits or budget thresholds
- **Usage Reporting**: Generate monthly or periodic usage reports for compliance
- **Repository Analysis**: Analyze runner usage patterns by repository
- **Resource Planning**: Analyze runner usage patterns by OS type

## References

- [GitHub Billing Usage API](https://docs.github.com/rest/billing/usage)
- [Permissions required for GitHub Apps - Organization Administration](https://docs.github.com/en/rest/overview/permissions-required-for-github-apps#organization-permissions-for-administration)
- [Viewing your usage of metered products](https://docs.github.com/en/billing/managing-billing-for-github-actions/viewing-your-github-actions-usage)`
}

func (g *GetWorkflowUsage) Icon() string {
	return "github"
}

func (g *GetWorkflowUsage) Color() string {
	return "gray"
}

func (g *GetWorkflowUsage) OutputChannels(configuration any) []core.OutputChannel {
	return []core.OutputChannel{core.DefaultOutputChannel}
}

func (g *GetWorkflowUsage) Configuration() []configuration.Field {
	return []configuration.Field{
		{
			Name:        "repositories",
			Label:       "Repositories",
			Type:        configuration.FieldTypeIntegrationResource,
			Required:    false,
			Description: "Select specific repositories to check usage for. Leave empty for organization-wide usage.",
			TypeOptions: &configuration.TypeOptions{
				Resource: &configuration.ResourceTypeOptions{
					Type:           "repository",
					UseNameAsValue: true,
					Multi:          true,
				},
			},
		},
	}
}

func (g *GetWorkflowUsage) Setup(ctx core.SetupContext) error {
	var config GetWorkflowUsageConfiguration
	if err := mapstructure.Decode(ctx.Configuration, &config); err != nil {
		return fmt.Errorf("failed to decode configuration: %w", err)
	}

	// If repositories are specified, validate they exist in metadata
	if len(config.Repositories) > 0 {
		// Validate each repository
		var appMetadata Metadata
		if err := mapstructure.Decode(ctx.Integration.GetMetadata(), &appMetadata); err != nil {
			return fmt.Errorf("failed to decode application metadata: %w", err)
		}

		// Check each repository exists
		for _, repo := range config.Repositories {
			found := false
			for _, availableRepo := range appMetadata.Repositories {
				if availableRepo.Name == repo {
					found = true
					break
				}
			}
			if !found {
				return fmt.Errorf("repository %s is not accessible to app installation", repo)
			}
		}
	}

	return nil
}

func (g *GetWorkflowUsage) Execute(ctx core.ExecutionContext) error {
	var config GetWorkflowUsageConfiguration
	if err := mapstructure.Decode(ctx.Configuration, &config); err != nil {
		return fmt.Errorf("failed to decode configuration: %w", err)
	}

	var appMetadata Metadata
	if err := mapstructure.Decode(ctx.Integration.GetMetadata(), &appMetadata); err != nil {
		return fmt.Errorf("failed to decode application metadata: %w", err)
	}

	client, err := NewClient(ctx.Integration, appMetadata.GitHubApp.ID, appMetadata.InstallationID)
	if err != nil {
		return fmt.Errorf("failed to initialize GitHub client: %w", err)
	}

	// Use the enhanced billing API for more detailed usage data
	// This API returns itemized usage with repository-level breakdown
	usageReport, _, err := client.Billing.GetUsageReportOrg(
		context.Background(),
		appMetadata.Owner,
		nil, // No time filtering - returns current billing cycle
	)
	if err != nil {
		return fmt.Errorf("failed to get billing usage: %w", err)
	}

	// Aggregate usage data, optionally filtering by repository
	result := WorkflowUsageResult{
		MinutesUsed:          0,
		MinutesUsedBreakdown: make(gh.MinutesUsedBreakdown),
		IncludedMinutes:      0, // Enhanced billing API doesn't include this field
		TotalPaidMinutesUsed: 0,
	}

	// Process usage items
	for _, item := range usageReport.UsageItems {
		// Filter by repositories if specified
		if len(config.Repositories) > 0 {
			found := false
			repoName := item.GetRepositoryName()
			for _, repo := range config.Repositories {
				if repoName == repo {
					found = true
					break
				}
			}
			if !found {
				continue
			}
		}

		// Only process Actions usage (skip other products like Copilot, Packages, etc.)
		if item.GetProduct() != "Actions" {
			continue
		}

		// Aggregate total minutes (quantity represents minutes for Actions)
		if item.Quantity != nil {
			result.MinutesUsed += *item.Quantity
		}

		// Aggregate by SKU (runner OS type)
		sku := item.GetSKU()
		if sku != "" && item.Quantity != nil {
			result.MinutesUsedBreakdown[sku] = result.MinutesUsedBreakdown[sku] + int(*item.Quantity)
		}

		// Calculate paid minutes from net amount
		// Note: Enhanced billing API provides cost, not included minutes
		if item.NetAmount != nil {
			result.TotalPaidMinutesUsed += *item.NetAmount / 0.008 // Approximate minutes from cost (assuming $0.008/min average)
		}
	}

	return ctx.ExecutionState.Emit(
		core.DefaultOutputChannel.Name,
		"github.workflowUsage",
		[]any{result},
	)
}

func (g *GetWorkflowUsage) ProcessQueueItem(ctx core.ProcessQueueContext) (*uuid.UUID, error) {
	return ctx.DefaultProcessing()
}

func (g *GetWorkflowUsage) HandleWebhook(ctx core.WebhookRequestContext) (int, error) {
	return 200, nil
}

func (g *GetWorkflowUsage) Actions() []core.Action {
	return []core.Action{}
}

func (g *GetWorkflowUsage) HandleAction(ctx core.ActionContext) error {
	return nil
}

func (g *GetWorkflowUsage) Cancel(ctx core.ExecutionContext) error {
	return nil
}

func (g *GetWorkflowUsage) Cleanup(ctx core.SetupContext) error {
	return nil
}

func (g *GetWorkflowUsage) ExampleOutput() map[string]any {
	return utils.UnmarshalEmbeddedJSON(&getWorkflowUsageExampleOutputOnce, getWorkflowUsageExampleOutputBytes, &getWorkflowUsageExampleOutput)
}
