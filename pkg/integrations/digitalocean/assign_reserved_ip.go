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

type AssignReservedIP struct{}

type AssignReservedIPSpec struct {
	ReservedIP string `json:"reservedIP"`
	Action     string `json:"action"`
	DropletID  string `json:"dropletID"`
}

func (c *AssignReservedIP) Name() string {
	return "digitalocean.assignReservedIP"
}

func (c *AssignReservedIP) Label() string {
	return "Assign Reserved IP"
}

func (c *AssignReservedIP) Description() string {
	return "Assign or unassign a DigitalOcean Reserved IP to a Droplet"
}

func (c *AssignReservedIP) Documentation() string {
	return `The Assign Reserved IP component assigns or unassigns a DigitalOcean reserved IP address to a droplet.

## Use Cases

- **Failover**: Reassign a reserved IP from a failed droplet to a healthy one
- **Blue/green deployments**: Move a reserved IP between deployments
- **Release IPs**: Unassign a reserved IP when a droplet is decommissioned

## Configuration

- **Reserved IP**: The reserved IP address to act on (required)
- **Action**: Whether to assign or unassign (required)
- **Droplet ID**: The numeric ID of the target droplet (required when assigning)

## Output

Returns the DigitalOcean reserved IP action result.`
}

func (c *AssignReservedIP) Icon() string {
	return "network"
}

func (c *AssignReservedIP) Color() string {
	return "blue"
}

func (c *AssignReservedIP) OutputChannels(configuration any) []core.OutputChannel {
	return []core.OutputChannel{core.DefaultOutputChannel}
}

func (c *AssignReservedIP) Configuration() []configuration.Field {
	return []configuration.Field{
		{
			Name:        "reservedIP",
			Label:       "Reserved IP",
			Type:        configuration.FieldTypeString,
			Required:    true,
			Description: "The reserved IP address to act on",
		},
		{
			Name:        "action",
			Label:       "Action",
			Type:        configuration.FieldTypeSelect,
			Required:    true,
			Description: "Whether to assign or unassign the reserved IP",
			TypeOptions: &configuration.TypeOptions{
				Select: &configuration.SelectTypeOptions{
					Options: []configuration.FieldOption{
						{Label: "Assign", Value: "assign"},
						{Label: "Unassign", Value: "unassign"},
					},
				},
			},
		},
		{
			Name:        "dropletID",
			Label:       "Droplet ID",
			Type:        configuration.FieldTypeString,
			Required:    false,
			Togglable:   true,
			Description: "The numeric ID of the droplet to assign the IP to (required when assigning)",
		},
	}
}

func (c *AssignReservedIP) Setup(ctx core.SetupContext) error {
	spec := AssignReservedIPSpec{}
	if err := mapstructure.Decode(ctx.Configuration, &spec); err != nil {
		return fmt.Errorf("error decoding configuration: %v", err)
	}

	if spec.ReservedIP == "" {
		return errors.New("reservedIP is required")
	}

	if spec.Action == "" {
		return errors.New("action is required")
	}

	return nil
}

func (c *AssignReservedIP) Execute(ctx core.ExecutionContext) error {
	spec := AssignReservedIPSpec{}
	if err := mapstructure.Decode(ctx.Configuration, &spec); err != nil {
		return fmt.Errorf("error decoding configuration: %v", err)
	}

	body := map[string]any{"type": spec.Action}

	if spec.Action == "assign" {
		if spec.DropletID == "" {
			return errors.New("dropletID is required when action is assign")
		}
		id, err := strconv.Atoi(spec.DropletID)
		if err != nil {
			return fmt.Errorf("invalid dropletID %q: %v", spec.DropletID, err)
		}
		body["droplet_id"] = id
	}

	client, err := NewClient(ctx.HTTP, ctx.Integration)
	if err != nil {
		return fmt.Errorf("error creating client: %v", err)
	}

	action, err := client.PostReservedIPAction(spec.ReservedIP, body)
	if err != nil {
		return fmt.Errorf("failed to perform reserved IP action: %v", err)
	}

	return ctx.ExecutionState.Emit(
		core.DefaultOutputChannel.Name,
		"digitalocean.reserved_ip.action",
		[]any{action},
	)
}

func (c *AssignReservedIP) Cancel(ctx core.ExecutionContext) error {
	return nil
}

func (c *AssignReservedIP) ProcessQueueItem(ctx core.ProcessQueueContext) (*uuid.UUID, error) {
	return ctx.DefaultProcessing()
}

func (c *AssignReservedIP) Actions() []core.Action {
	return []core.Action{}
}

func (c *AssignReservedIP) HandleAction(ctx core.ActionContext) error {
	return fmt.Errorf("unknown action: %s", ctx.Name)
}

func (c *AssignReservedIP) HandleWebhook(ctx core.WebhookRequestContext) (int, error) {
	return http.StatusOK, nil
}

func (c *AssignReservedIP) Cleanup(ctx core.SetupContext) error {
	return nil
}
