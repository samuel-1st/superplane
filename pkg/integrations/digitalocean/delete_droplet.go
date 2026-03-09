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

type DeleteDroplet struct{}

type DeleteDropletSpec struct {
	DropletID string `json:"dropletID"`
}

func (c *DeleteDroplet) Name() string {
	return "digitalocean.deleteDroplet"
}

func (c *DeleteDroplet) Label() string {
	return "Delete Droplet"
}

func (c *DeleteDroplet) Description() string {
	return "Delete a DigitalOcean Droplet"
}

func (c *DeleteDroplet) Documentation() string {
	return `The Delete Droplet component permanently deletes a DigitalOcean droplet.

## Use Cases

- **Cleanup**: Remove droplets after a job or test run completes
- **Cost management**: Automatically terminate unused instances

## Configuration

- **Droplet ID**: The numeric ID of the droplet to delete (required, supports expressions)

## Output

Returns a confirmation object with deleted status and the droplet ID.`
}

func (c *DeleteDroplet) Icon() string {
	return "trash"
}

func (c *DeleteDroplet) Color() string {
	return "red"
}

func (c *DeleteDroplet) OutputChannels(configuration any) []core.OutputChannel {
	return []core.OutputChannel{core.DefaultOutputChannel}
}

func (c *DeleteDroplet) Configuration() []configuration.Field {
	return []configuration.Field{
		{
			Name:        "dropletID",
			Label:       "Droplet ID",
			Type:        configuration.FieldTypeString,
			Required:    true,
			Description: "The numeric ID of the droplet to delete",
		},
	}
}

func (c *DeleteDroplet) Setup(ctx core.SetupContext) error {
	spec := DeleteDropletSpec{}
	if err := mapstructure.Decode(ctx.Configuration, &spec); err != nil {
		return fmt.Errorf("error decoding configuration: %v", err)
	}

	if spec.DropletID == "" {
		return errors.New("dropletID is required")
	}

	return nil
}

func (c *DeleteDroplet) Execute(ctx core.ExecutionContext) error {
	spec := DeleteDropletSpec{}
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

	if err := client.DeleteDroplet(id); err != nil {
		return fmt.Errorf("failed to delete droplet: %v", err)
	}

	return ctx.ExecutionState.Emit(
		core.DefaultOutputChannel.Name,
		"digitalocean.droplet.deleted",
		[]any{map[string]any{"deleted": true, "dropletID": id}},
	)
}

func (c *DeleteDroplet) Cancel(ctx core.ExecutionContext) error {
	return nil
}

func (c *DeleteDroplet) ProcessQueueItem(ctx core.ProcessQueueContext) (*uuid.UUID, error) {
	return ctx.DefaultProcessing()
}

func (c *DeleteDroplet) Actions() []core.Action {
	return []core.Action{}
}

func (c *DeleteDroplet) HandleAction(ctx core.ActionContext) error {
	return fmt.Errorf("unknown action: %s", ctx.Name)
}

func (c *DeleteDroplet) HandleWebhook(ctx core.WebhookRequestContext) (int, error) {
	return http.StatusOK, nil
}

func (c *DeleteDroplet) Cleanup(ctx core.SetupContext) error {
	return nil
}
