package digitalocean

import (
	"errors"
	"fmt"
	"net/http"

	"github.com/google/uuid"
	"github.com/mitchellh/mapstructure"
	"github.com/superplanehq/superplane/pkg/configuration"
	"github.com/superplanehq/superplane/pkg/core"
)

type DeleteSnapshot struct{}

type DeleteSnapshotSpec struct {
	SnapshotID string `json:"snapshotID"`
}

func (c *DeleteSnapshot) Name() string {
	return "digitalocean.deleteSnapshot"
}

func (c *DeleteSnapshot) Label() string {
	return "Delete Snapshot"
}

func (c *DeleteSnapshot) Description() string {
	return "Delete a DigitalOcean snapshot"
}

func (c *DeleteSnapshot) Documentation() string {
	return `The Delete Snapshot component permanently deletes a DigitalOcean snapshot.

## Use Cases

- **Storage cleanup**: Remove outdated snapshots to reduce storage costs
- **Post-migration cleanup**: Delete temporary snapshots after they are no longer needed

## Configuration

- **Snapshot ID**: The ID of the snapshot to delete (required, supports expressions)

## Output

Returns a confirmation object with deleted status and the snapshot ID.`
}

func (c *DeleteSnapshot) Icon() string {
	return "trash"
}

func (c *DeleteSnapshot) Color() string {
	return "red"
}

func (c *DeleteSnapshot) OutputChannels(configuration any) []core.OutputChannel {
	return []core.OutputChannel{core.DefaultOutputChannel}
}

func (c *DeleteSnapshot) Configuration() []configuration.Field {
	return []configuration.Field{
		{
			Name:        "snapshotID",
			Label:       "Snapshot ID",
			Type:        configuration.FieldTypeString,
			Required:    true,
			Description: "The ID of the snapshot to delete",
		},
	}
}

func (c *DeleteSnapshot) Setup(ctx core.SetupContext) error {
	spec := DeleteSnapshotSpec{}
	if err := mapstructure.Decode(ctx.Configuration, &spec); err != nil {
		return fmt.Errorf("error decoding configuration: %v", err)
	}

	if spec.SnapshotID == "" {
		return errors.New("snapshotID is required")
	}

	return nil
}

func (c *DeleteSnapshot) Execute(ctx core.ExecutionContext) error {
	spec := DeleteSnapshotSpec{}
	if err := mapstructure.Decode(ctx.Configuration, &spec); err != nil {
		return fmt.Errorf("error decoding configuration: %v", err)
	}

	client, err := NewClient(ctx.HTTP, ctx.Integration)
	if err != nil {
		return fmt.Errorf("error creating client: %v", err)
	}

	if err := client.DeleteSnapshot(spec.SnapshotID); err != nil {
		return fmt.Errorf("failed to delete snapshot: %v", err)
	}

	return ctx.ExecutionState.Emit(
		core.DefaultOutputChannel.Name,
		"digitalocean.snapshot.deleted",
		[]any{map[string]any{"deleted": true, "snapshotID": spec.SnapshotID}},
	)
}

func (c *DeleteSnapshot) Cancel(ctx core.ExecutionContext) error {
	return nil
}

func (c *DeleteSnapshot) ProcessQueueItem(ctx core.ProcessQueueContext) (*uuid.UUID, error) {
	return ctx.DefaultProcessing()
}

func (c *DeleteSnapshot) Actions() []core.Action {
	return []core.Action{}
}

func (c *DeleteSnapshot) HandleAction(ctx core.ActionContext) error {
	return fmt.Errorf("unknown action: %s", ctx.Name)
}

func (c *DeleteSnapshot) HandleWebhook(ctx core.WebhookRequestContext) (int, error) {
	return http.StatusOK, nil
}

func (c *DeleteSnapshot) Cleanup(ctx core.SetupContext) error {
	return nil
}
