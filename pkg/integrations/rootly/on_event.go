package rootly

import (
	"encoding/json"
	"fmt"
	"net/http"
	"slices"

	"github.com/mitchellh/mapstructure"
	"github.com/superplanehq/superplane/pkg/configuration"
	"github.com/superplanehq/superplane/pkg/core"
)

type OnEvent struct{}

type OnEventConfiguration struct {
	Events         []string `json:"events"`
	IncidentStatus []string `json:"incidentStatus,omitempty"`
	Severity       []string `json:"severity,omitempty"`
	Service        []string `json:"service,omitempty"`
	Team           []string `json:"team,omitempty"`
	EventSource    string   `json:"eventSource,omitempty"`
	Visibility     string   `json:"visibility,omitempty"`
	EventKind      []string `json:"eventKind,omitempty"`
}

func (t *OnEvent) Name() string {
	return "rootly.onEvent"
}

func (t *OnEvent) Label() string {
	return "On Event"
}

func (t *OnEvent) Description() string {
	return "Listen to incident timeline events"
}

func (t *OnEvent) Documentation() string {
	return `The On Event trigger starts a workflow execution when Rootly incident timeline events are created or updated.

## Use Cases

- **Note automation**: Run a workflow when someone adds a note to an incident
- **Timeline sync**: Sync timeline events to Slack or external systems
- **Investigation tracking**: Run automation when investigation notes are added

## Configuration

- **Events**: Select which incident event types to listen for (created, updated)
- **Incident status** (optional): Filter by the status of the associated incident
- **Severity** (optional): Filter by the severity of the associated incident
- **Service** (optional): Filter by service associated with the incident
- **Team** (optional): Filter by team associated with the incident
- **Event source** (optional): Filter by event source
- **Visibility** (optional): Filter by event visibility (external or internal)
- **Event kind** (optional): Filter by event kind (e.g. note, annotation)

## Event Data

Each incident event includes:
- **id**: Event ID
- **event**: Event content
- **kind**: Event kind (note, annotation, etc.)
- **occurred_at**: When the event occurred
- **created_at**: When the event was created
- **user_display_name**: Display name of the user who created the event
- **incident**: The associated incident data

## Webhook Setup

This trigger automatically sets up a Rootly webhook endpoint when configured. The endpoint is managed by SuperPlane and will be cleaned up when the trigger is removed.`
}

func (t *OnEvent) Icon() string {
	return "file-text"
}

func (t *OnEvent) Color() string {
	return "gray"
}

func (t *OnEvent) Configuration() []configuration.Field {
	return []configuration.Field{
		{
			Name:     "events",
			Label:    "Events",
			Type:     configuration.FieldTypeMultiSelect,
			Required: true,
			Default:  []string{"incident_event.created"},
			TypeOptions: &configuration.TypeOptions{
				MultiSelect: &configuration.MultiSelectTypeOptions{
					Options: []configuration.FieldOption{
						{Label: "Created", Value: "incident_event.created"},
						{Label: "Updated", Value: "incident_event.updated"},
					},
				},
			},
		},
		{
			Name:     "incidentStatus",
			Label:    "Incident Status",
			Type:     configuration.FieldTypeMultiSelect,
			Required: false,
			TypeOptions: &configuration.TypeOptions{
				MultiSelect: &configuration.MultiSelectTypeOptions{
					Options: []configuration.FieldOption{
						{Label: "Started", Value: "started"},
						{Label: "Mitigated", Value: "mitigated"},
						{Label: "Resolved", Value: "resolved"},
						{Label: "Cancelled", Value: "cancelled"},
					},
				},
			},
			Description: "Filter by incident status. Leave empty to match all statuses.",
		},
		{
			Name:     "severity",
			Label:    "Severity",
			Type:     configuration.FieldTypeIntegrationResource,
			Required: false,
			TypeOptions: &configuration.TypeOptions{
				Resource: &configuration.ResourceTypeOptions{
					Type:           "severity",
					UseNameAsValue: true,
					Multi:          true,
				},
			},
			Description: "Filter by incident severity. Leave empty to match all severities.",
		},
		{
			Name:     "service",
			Label:    "Service",
			Type:     configuration.FieldTypeIntegrationResource,
			Required: false,
			TypeOptions: &configuration.TypeOptions{
				Resource: &configuration.ResourceTypeOptions{
					Type:           "service",
					UseNameAsValue: true,
					Multi:          true,
				},
			},
			Description: "Filter by incident service. Leave empty to match all services.",
		},
		{
			Name:     "team",
			Label:    "Team",
			Type:     configuration.FieldTypeIntegrationResource,
			Required: false,
			TypeOptions: &configuration.TypeOptions{
				Resource: &configuration.ResourceTypeOptions{
					Type:           "team",
					UseNameAsValue: true,
					Multi:          true,
				},
			},
			Description: "Filter by incident team. Leave empty to match all teams.",
		},
		{
			Name:        "eventSource",
			Label:       "Event Source",
			Type:        configuration.FieldTypeString,
			Required:    false,
			Description: "Filter by event source. Leave empty to match all sources.",
		},
		{
			Name:     "visibility",
			Label:    "Visibility",
			Type:     configuration.FieldTypeSelect,
			Required: false,
			TypeOptions: &configuration.TypeOptions{
				Select: &configuration.SelectTypeOptions{
					Options: []configuration.FieldOption{
						{Label: "Internal", Value: "internal"},
						{Label: "External", Value: "external"},
					},
				},
			},
			Description: "Filter by event visibility. Leave empty to match all.",
		},
		{
			Name:     "eventKind",
			Label:    "Event Kind",
			Type:     configuration.FieldTypeMultiSelect,
			Required: false,
			TypeOptions: &configuration.TypeOptions{
				MultiSelect: &configuration.MultiSelectTypeOptions{
					Options: []configuration.FieldOption{
						{Label: "Note", Value: "note"},
						{Label: "Annotation", Value: "annotation"},
					},
				},
			},
			Description: "Filter by event kind. Leave empty to match all kinds.",
		},
	}
}

func (t *OnEvent) Setup(ctx core.TriggerContext) error {
	config := OnEventConfiguration{}
	err := mapstructure.Decode(ctx.Configuration, &config)
	if err != nil {
		return fmt.Errorf("failed to decode configuration: %w", err)
	}

	if len(config.Events) == 0 {
		return fmt.Errorf("at least one event type must be chosen")
	}

	return ctx.Integration.RequestWebhook(WebhookConfiguration{
		Events: config.Events,
	})
}

func (t *OnEvent) Actions() []core.Action {
	return []core.Action{}
}

func (t *OnEvent) HandleAction(ctx core.TriggerActionContext) (map[string]any, error) {
	return nil, nil
}

func (t *OnEvent) HandleWebhook(ctx core.WebhookRequestContext) (int, error) {
	config := OnEventConfiguration{}
	err := mapstructure.Decode(ctx.Configuration, &config)
	if err != nil {
		return http.StatusInternalServerError, fmt.Errorf("failed to decode configuration: %w", err)
	}

	// Verify signature
	signature := ctx.Headers.Get("X-Rootly-Signature")
	secret, err := ctx.Webhook.GetSecret()
	if err != nil {
		return http.StatusInternalServerError, fmt.Errorf("error getting secret: %v", err)
	}

	if err := verifyWebhookSignature(signature, ctx.Body, secret); err != nil {
		return http.StatusForbidden, fmt.Errorf("invalid signature: %v", err)
	}

	// Parse webhook payload
	var webhook WebhookPayload
	err = json.Unmarshal(ctx.Body, &webhook)
	if err != nil {
		return http.StatusBadRequest, fmt.Errorf("error parsing request body: %v", err)
	}

	eventType := webhook.Event.Type

	// Filter by event type
	if !slices.Contains(config.Events, eventType) {
		return http.StatusOK, nil
	}

	// Filter by optional fields
	if !matchesEventFilters(webhook.Data, config) {
		return http.StatusOK, nil
	}

	err = ctx.Events.Emit(
		fmt.Sprintf("rootly.%s", eventType),
		buildEventPayload(webhook),
	)

	if err != nil {
		return http.StatusInternalServerError, fmt.Errorf("error emitting event: %v", err)
	}

	return http.StatusOK, nil
}

// matchesEventFilters checks whether the webhook data passes all configured optional filters.
func matchesEventFilters(data map[string]any, config OnEventConfiguration) bool {
	if config.Visibility != "" {
		visibility, _ := data["visibility"].(string)
		if visibility != config.Visibility {
			return false
		}
	}

	if len(config.EventKind) > 0 {
		kind, _ := data["kind"].(string)
		if !slices.Contains(config.EventKind, kind) {
			return false
		}
	}

	if config.EventSource != "" {
		source, _ := data["source"].(string)
		if source != config.EventSource {
			return false
		}
	}

	// Filters based on the nested incident object
	incident, _ := data["incident"].(map[string]any)
	if incident == nil {
		// If there's no incident data and we have incident-level filters, skip the event
		if len(config.IncidentStatus) > 0 || len(config.Severity) > 0 ||
			len(config.Service) > 0 || len(config.Team) > 0 {
			return false
		}
		return true
	}

	if len(config.IncidentStatus) > 0 {
		status, _ := incident["status"].(string)
		if !slices.Contains(config.IncidentStatus, status) {
			return false
		}
	}

	if len(config.Severity) > 0 {
		severity := severityString(incident["severity"])
		if !slices.Contains(config.Severity, severity) {
			return false
		}
	}

	if len(config.Service) > 0 {
		if !incidentMatchesNames(incident["services"], config.Service) {
			return false
		}
	}

	if len(config.Team) > 0 {
		if !incidentMatchesNames(incident["groups"], config.Team) {
			return false
		}
	}

	return true
}

// incidentMatchesNames checks if any of the resource entries (as list of strings or objects with "name")
// match at least one of the required names.
func incidentMatchesNames(value any, names []string) bool {
	switch v := value.(type) {
	case []any:
		for _, item := range v {
			switch s := item.(type) {
			case string:
				if slices.Contains(names, s) {
					return true
				}
			case map[string]any:
				if name, ok := s["name"].(string); ok && slices.Contains(names, name) {
					return true
				}
			}
		}
	case []string:
		for _, s := range v {
			if slices.Contains(names, s) {
				return true
			}
		}
	}

	return false
}

func buildEventPayload(webhook WebhookPayload) map[string]any {
	payload := map[string]any{
		"event":     webhook.Event.Type,
		"event_id":  webhook.Event.ID,
		"issued_at": webhook.Event.IssuedAt,
	}

	if webhook.Data != nil {
		for k, v := range webhook.Data {
			payload[k] = v
		}
	}

	return payload
}

func (t *OnEvent) Cleanup(ctx core.TriggerContext) error {
	return nil
}
