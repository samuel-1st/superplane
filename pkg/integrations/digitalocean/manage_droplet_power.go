package digitalocean

import (
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/google/uuid"
	"github.com/mitchellh/mapstructure"
	"github.com/superplanehq/superplane/pkg/configuration"
	"github.com/superplanehq/superplane/pkg/core"
)

const actionPollInterval = 10 * time.Second

type ManageDropletPower struct{}

type ManageDropletPowerSpec struct {
	DropletID string `json:"dropletID"`
	Operation string `json:"operation"`
}

func (c *ManageDropletPower) Name() string {
	return "digitalocean.manageDropletPower"
}

func (c *ManageDropletPower) Label() string {
	return "Manage Droplet Power"
}

func (c *ManageDropletPower) Description() string {
	return "Manage the power state of a DigitalOcean Droplet"
}

func (c *ManageDropletPower) Documentation() string {
	return `The Manage Droplet Power component controls the power state of a DigitalOcean droplet.

## Use Cases

- **Scheduled shutdowns**: Power off droplets during off-hours
- **Reboots**: Restart droplets after configuration changes
- **Power cycling**: Force restart unresponsive droplets

## Configuration

- **Droplet ID**: The numeric ID of the droplet (required, supports expressions)
- **Operation**: The power action to perform (required)

## Output

Returns the completed DigitalOcean action object.`
}

func (c *ManageDropletPower) Icon() string {
	return "power"
}

func (c *ManageDropletPower) Color() string {
	return "yellow"
}

func (c *ManageDropletPower) OutputChannels(configuration any) []core.OutputChannel {
	return []core.OutputChannel{core.DefaultOutputChannel}
}

func (c *ManageDropletPower) Configuration() []configuration.Field {
	return []configuration.Field{
		{
			Name:        "dropletID",
			Label:       "Droplet ID",
			Type:        configuration.FieldTypeString,
			Required:    true,
			Description: "The numeric ID of the droplet",
		},
		{
			Name:        "operation",
			Label:       "Operation",
			Type:        configuration.FieldTypeSelect,
			Required:    true,
			Description: "The power action to perform on the droplet",
			TypeOptions: &configuration.TypeOptions{
				Select: &configuration.SelectTypeOptions{
					Options: []configuration.FieldOption{
						{Label: "Power On", Value: "power_on"},
						{Label: "Shutdown", Value: "shutdown"},
						{Label: "Reboot", Value: "reboot"},
						{Label: "Power Cycle", Value: "power_cycle"},
						{Label: "Power Off", Value: "power_off"},
					},
				},
			},
		},
	}
}

func (c *ManageDropletPower) Setup(ctx core.SetupContext) error {
	spec := ManageDropletPowerSpec{}
	if err := mapstructure.Decode(ctx.Configuration, &spec); err != nil {
		return fmt.Errorf("error decoding configuration: %v", err)
	}

	if spec.DropletID == "" {
		return errors.New("dropletID is required")
	}

	if spec.Operation == "" {
		return errors.New("operation is required")
	}

	return nil
}

func (c *ManageDropletPower) Execute(ctx core.ExecutionContext) error {
	spec := ManageDropletPowerSpec{}
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

	action, err := client.PostDropletAction(id, map[string]any{"type": spec.Operation})
	if err != nil {
		return fmt.Errorf("failed to perform power action: %v", err)
	}

	if err := ctx.Metadata.Set(map[string]any{"actionID": action.ID}); err != nil {
		return fmt.Errorf("failed to store metadata: %v", err)
	}

	return ctx.Requests.ScheduleActionCall("poll", map[string]any{}, actionPollInterval)
}

func (c *ManageDropletPower) Cancel(ctx core.ExecutionContext) error {
	return nil
}

func (c *ManageDropletPower) ProcessQueueItem(ctx core.ProcessQueueContext) (*uuid.UUID, error) {
	return ctx.DefaultProcessing()
}

func (c *ManageDropletPower) Actions() []core.Action {
	return []core.Action{
		{
			Name:           "poll",
			UserAccessible: false,
		},
	}
}

func (c *ManageDropletPower) HandleAction(ctx core.ActionContext) error {
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
			"digitalocean.droplet.power.managed",
			[]any{action},
		)
	case "errored":
		return fmt.Errorf("action %d errored", action.ID)
	default:
		return ctx.Requests.ScheduleActionCall("poll", map[string]any{}, actionPollInterval)
	}
}

func (c *ManageDropletPower) HandleWebhook(ctx core.WebhookRequestContext) (int, error) {
	return http.StatusOK, nil
}

func (c *ManageDropletPower) Cleanup(ctx core.SetupContext) error {
	return nil
}
