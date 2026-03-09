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

type DeleteLoadBalancer struct{}

type DeleteLoadBalancerSpec struct {
	LoadBalancerID string `json:"loadBalancerID"`
}

func (c *DeleteLoadBalancer) Name() string {
	return "digitalocean.deleteLoadBalancer"
}

func (c *DeleteLoadBalancer) Label() string {
	return "Delete Load Balancer"
}

func (c *DeleteLoadBalancer) Description() string {
	return "Delete a DigitalOcean Load Balancer"
}

func (c *DeleteLoadBalancer) Documentation() string {
	return `The Delete Load Balancer component permanently removes a DigitalOcean load balancer.

## Use Cases

- **Cleanup**: Remove load balancers when decommissioning infrastructure
- **Cost management**: Delete unused load balancers to avoid billing

## Configuration

- **Load Balancer ID**: The ID of the load balancer to delete (required, supports expressions)

## Output

Returns a confirmation object with deleted status and the load balancer ID.`
}

func (c *DeleteLoadBalancer) Icon() string {
	return "trash"
}

func (c *DeleteLoadBalancer) Color() string {
	return "red"
}

func (c *DeleteLoadBalancer) OutputChannels(configuration any) []core.OutputChannel {
	return []core.OutputChannel{core.DefaultOutputChannel}
}

func (c *DeleteLoadBalancer) Configuration() []configuration.Field {
	return []configuration.Field{
		{
			Name:        "loadBalancerID",
			Label:       "Load Balancer ID",
			Type:        configuration.FieldTypeString,
			Required:    true,
			Description: "The ID of the load balancer to delete",
		},
	}
}

func (c *DeleteLoadBalancer) Setup(ctx core.SetupContext) error {
	spec := DeleteLoadBalancerSpec{}
	if err := mapstructure.Decode(ctx.Configuration, &spec); err != nil {
		return fmt.Errorf("error decoding configuration: %v", err)
	}

	if spec.LoadBalancerID == "" {
		return errors.New("loadBalancerID is required")
	}

	return nil
}

func (c *DeleteLoadBalancer) Execute(ctx core.ExecutionContext) error {
	spec := DeleteLoadBalancerSpec{}
	if err := mapstructure.Decode(ctx.Configuration, &spec); err != nil {
		return fmt.Errorf("error decoding configuration: %v", err)
	}

	client, err := NewClient(ctx.HTTP, ctx.Integration)
	if err != nil {
		return fmt.Errorf("error creating client: %v", err)
	}

	if err := client.DeleteLoadBalancer(spec.LoadBalancerID); err != nil {
		return fmt.Errorf("failed to delete load balancer: %v", err)
	}

	return ctx.ExecutionState.Emit(
		core.DefaultOutputChannel.Name,
		"digitalocean.load_balancer.deleted",
		[]any{map[string]any{"deleted": true, "loadBalancerID": spec.LoadBalancerID}},
	)
}

func (c *DeleteLoadBalancer) Cancel(ctx core.ExecutionContext) error {
	return nil
}

func (c *DeleteLoadBalancer) ProcessQueueItem(ctx core.ProcessQueueContext) (*uuid.UUID, error) {
	return ctx.DefaultProcessing()
}

func (c *DeleteLoadBalancer) Actions() []core.Action {
	return []core.Action{}
}

func (c *DeleteLoadBalancer) HandleAction(ctx core.ActionContext) error {
	return fmt.Errorf("unknown action: %s", ctx.Name)
}

func (c *DeleteLoadBalancer) HandleWebhook(ctx core.WebhookRequestContext) (int, error) {
	return http.StatusOK, nil
}

func (c *DeleteLoadBalancer) Cleanup(ctx core.SetupContext) error {
	return nil
}
