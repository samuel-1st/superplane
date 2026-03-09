package digitalocean

import (
	"errors"
	"fmt"
	"net/http"
	"strconv"

	"github.com/google/uuid"
	"github.com/mitchellh/mapstructure"
	"github.com/superplanehq/superplane/pkg/configuration"
	"github.com/superplanehq/superplane/pkg/core"
)

type CreateSnapshot struct{}

type CreateSnapshotSpec struct {
	DropletID    string `json:"dropletID"`
	SnapshotName string `json:"snapshotName"`
}

func (c *CreateSnapshot) Name() string {
	return "digitalocean.createSnapshot"
}

func (c *CreateSnapshot) Label() string {
	return "Create Snapshot"
}

func (c *CreateSnapshot) Description() string {
	return "Create a snapshot of a DigitalOcean Droplet"
}

func (c *CreateSnapshot) Documentation() string {
	return `The Create Snapshot component takes a snapshot of a DigitalOcean droplet.

## Use Cases

- **Backups**: Capture the state of a droplet before changes
- **Cloning**: Create snapshots to use as base images for new droplets

## Configuration

- **Droplet ID**: The numeric ID of the droplet to snapshot (required, supports expressions)
- **Snapshot Name**: The name for the new snapshot (required, supports expressions)

## Output

Returns the completed DigitalOcean action object once the snapshot is created.`
}

func (c *CreateSnapshot) Icon() string {
	return "camera"
}

func (c *CreateSnapshot) Color() string {
	return "blue"
}

func (c *CreateSnapshot) OutputChannels(configuration any) []core.OutputChannel {
	return []core.OutputChannel{core.DefaultOutputChannel}
}

func (c *CreateSnapshot) Configuration() []configuration.Field {
	return []configuration.Field{
		{
			Name:        "dropletID",
			Label:       "Droplet ID",
			Type:        configuration.FieldTypeString,
			Required:    true,
			Description: "The numeric ID of the droplet to snapshot",
		},
		{
			Name:        "snapshotName",
			Label:       "Snapshot Name",
			Type:        configuration.FieldTypeString,
			Required:    true,
			Description: "The name for the new snapshot",
		},
	}
}

func (c *CreateSnapshot) Setup(ctx core.SetupContext) error {
	spec := CreateSnapshotSpec{}
	if err := mapstructure.Decode(ctx.Configuration, &spec); err != nil {
		return fmt.Errorf("error decoding configuration: %v", err)
	}

	if spec.DropletID == "" {
		return errors.New("dropletID is required")
	}

	if spec.SnapshotName == "" {
		return errors.New("snapshotName is required")
	}

	return nil
}

func (c *CreateSnapshot) Execute(ctx core.ExecutionContext) error {
	spec := CreateSnapshotSpec{}
	if err := mapstructure.Decode(ctx.Configuration, &spec); err != nil {
		return fmt.Errorf("error decoding configuration: %v", err)
	}

	id, err := strconv.Atoi(spec.DropletID)
	if err != nil {
		return fmt.Errorf("invalid dropletID %q: %v", spec.DropletID, err)
	}

	client, err := NewClient(ctx.HTTP, ctx.Integration)
	if err != nil {
		return fmt.Errorf("error creating client: %v", err)
	}

	action, err := client.PostDropletAction(id, map[string]any{
		"type": "snapshot",
		"name": spec.SnapshotName,
	})
	if err != nil {
		return fmt.Errorf("failed to create snapshot: %v", err)
	}

	if err := ctx.Metadata.Set(map[string]any{"actionID": action.ID}); err != nil {
		return fmt.Errorf("failed to store metadata: %v", err)
	}

	return ctx.Requests.ScheduleActionCall("poll", map[string]any{}, actionPollInterval)
}

func (c *CreateSnapshot) Cancel(ctx core.ExecutionContext) error {
	return nil
}

func (c *CreateSnapshot) ProcessQueueItem(ctx core.ProcessQueueContext) (*uuid.UUID, error) {
	return ctx.DefaultProcessing()
}

func (c *CreateSnapshot) Actions() []core.Action {
	return []core.Action{
		{
			Name:           "poll",
			UserAccessible: false,
		},
	}
}

func (c *CreateSnapshot) HandleAction(ctx core.ActionContext) error {
	if ctx.Name != "poll" {
		return fmt.Errorf("unknown action: %s", ctx.Name)
	}

	if ctx.ExecutionState.IsFinished() {
		return nil
	}

	var metadata struct {
		ActionID int `mapstructure:"actionID"`
	}

	if err := mapstructure.Decode(ctx.Metadata.Get(), &metadata); err != nil {
		return fmt.Errorf("failed to decode metadata: %v", err)
	}

	client, err := NewClient(ctx.HTTP, ctx.Integration)
	if err != nil {
		return fmt.Errorf("error creating client: %v", err)
	}

	action, err := client.GetAction(metadata.ActionID)
	if err != nil {
		return fmt.Errorf("failed to get action: %v", err)
	}

	switch action.Status {
	case "completed":
		return ctx.ExecutionState.Emit(
			core.DefaultOutputChannel.Name,
			"digitalocean.droplet.snapshot.created",
			[]any{action},
		)
	case "errored":
		return fmt.Errorf("action %d errored", action.ID)
	default:
		return ctx.Requests.ScheduleActionCall("poll", map[string]any{}, actionPollInterval)
	}
}

func (c *CreateSnapshot) HandleWebhook(ctx core.WebhookRequestContext) (int, error) {
	return http.StatusOK, nil
}

func (c *CreateSnapshot) Cleanup(ctx core.SetupContext) error {
	return nil
}
