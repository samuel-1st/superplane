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

type GetDroplet struct{}

type GetDropletSpec struct {
	DropletID string `json:"dropletID"`
}

func (c *GetDroplet) Name() string {
	return "digitalocean.getDroplet"
}

func (c *GetDroplet) Label() string {
	return "Get Droplet"
}

func (c *GetDroplet) Description() string {
	return "Retrieve a DigitalOcean Droplet by ID"
}

func (c *GetDroplet) Documentation() string {
	return `The Get Droplet component retrieves the details of an existing DigitalOcean droplet.

## Use Cases

- **Status checking**: Retrieve the current state of a droplet
- **Inventory lookups**: Fetch droplet metadata for downstream workflow steps

## Configuration

- **Droplet ID**: The numeric ID of the droplet to retrieve (required, supports expressions)

## Output

Returns the droplet object including id, name, status, region, networks, and image information.`
}

func (c *GetDroplet) Icon() string {
	return "server"
}

func (c *GetDroplet) Color() string {
	return "gray"
}

func (c *GetDroplet) OutputChannels(configuration any) []core.OutputChannel {
	return []core.OutputChannel{core.DefaultOutputChannel}
}

func (c *GetDroplet) Configuration() []configuration.Field {
	return []configuration.Field{
		{
			Name:        "dropletID",
			Label:       "Droplet ID",
			Type:        configuration.FieldTypeString,
			Required:    true,
			Description: "The numeric ID of the droplet to retrieve",
		},
	}
}

func (c *GetDroplet) Setup(ctx core.SetupContext) error {
	spec := GetDropletSpec{}
	if err := mapstructure.Decode(ctx.Configuration, &spec); err != nil {
		return fmt.Errorf("error decoding configuration: %v", err)
	}

	if spec.DropletID == "" {
		return errors.New("dropletID is required")
	}

	return nil
}

func (c *GetDroplet) Execute(ctx core.ExecutionContext) error {
	spec := GetDropletSpec{}
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

	droplet, err := client.GetDroplet(id)
	if err != nil {
		return fmt.Errorf("failed to get droplet: %v", err)
	}

	return ctx.ExecutionState.Emit(
		core.DefaultOutputChannel.Name,
		"digitalocean.droplet.retrieved",
		[]any{droplet},
	)
}

func (c *GetDroplet) Cancel(ctx core.ExecutionContext) error {
	return nil
}

func (c *GetDroplet) ProcessQueueItem(ctx core.ProcessQueueContext) (*uuid.UUID, error) {
	return ctx.DefaultProcessing()
}

func (c *GetDroplet) Actions() []core.Action {
	return []core.Action{}
}

func (c *GetDroplet) HandleAction(ctx core.ActionContext) error {
	return fmt.Errorf("unknown action: %s", ctx.Name)
}

func (c *GetDroplet) HandleWebhook(ctx core.WebhookRequestContext) (int, error) {
	return http.StatusOK, nil
}

func (c *GetDroplet) Cleanup(ctx core.SetupContext) error {
	return nil
}
