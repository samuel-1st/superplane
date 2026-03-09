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

type CreateLoadBalancer struct{}

type CreateLoadBalancerSpec struct {
	Name           string   `json:"name"`
	Region         string   `json:"region"`
	Size           string   `json:"size"`
	EntryProtocol  string   `json:"entryProtocol"`
	EntryPort      string   `json:"entryPort"`
	TargetProtocol string   `json:"targetProtocol"`
	TargetPort     string   `json:"targetPort"`
	DropletIDs     []string `json:"dropletIDs"`
	Tags           []string `json:"tags"`
}

func (c *CreateLoadBalancer) Name() string {
	return "digitalocean.createLoadBalancer"
}

func (c *CreateLoadBalancer) Label() string {
	return "Create Load Balancer"
}

func (c *CreateLoadBalancer) Description() string {
	return "Create a DigitalOcean Load Balancer"
}

func (c *CreateLoadBalancer) Documentation() string {
	return `The Create Load Balancer component provisions a new DigitalOcean load balancer.

## Use Cases

- **Traffic distribution**: Automatically set up load balancing for new application deployments
- **High availability**: Create load balancers as part of a multi-droplet infrastructure workflow

## Configuration

- **Name**: The name of the load balancer (required)
- **Region**: The region where the load balancer will be created (required)
- **Size**: The size of the load balancer (optional)
- **Entry Protocol**: The protocol for incoming traffic (required)
- **Entry Port**: The port for incoming traffic (required)
- **Target Protocol**: The protocol for forwarding to backend droplets (required)
- **Target Port**: The port on backend droplets (required)
- **Droplet IDs**: Numeric IDs of droplets to add as backends (optional)
- **Tags**: Tags to apply to the load balancer (optional)

## Output

Returns the created load balancer object including id, name, ip, and status.`
}

func (c *CreateLoadBalancer) Icon() string {
	return "sliders"
}

func (c *CreateLoadBalancer) Color() string {
	return "blue"
}

func (c *CreateLoadBalancer) OutputChannels(configuration any) []core.OutputChannel {
	return []core.OutputChannel{core.DefaultOutputChannel}
}

func (c *CreateLoadBalancer) Configuration() []configuration.Field {
	return []configuration.Field{
		{
			Name:        "name",
			Label:       "Name",
			Type:        configuration.FieldTypeString,
			Required:    true,
			Description: "The name of the load balancer",
		},
		{
			Name:        "region",
			Label:       "Region",
			Type:        configuration.FieldTypeIntegrationResource,
			Required:    true,
			Description: "The region where the load balancer will be created",
			Placeholder: "Select a region",
			TypeOptions: &configuration.TypeOptions{
				Resource: &configuration.ResourceTypeOptions{
					Type: "region",
				},
			},
		},
		{
			Name:        "size",
			Label:       "Size",
			Type:        configuration.FieldTypeSelect,
			Required:    false,
			Togglable:   true,
			Description: "The size of the load balancer",
			TypeOptions: &configuration.TypeOptions{
				Select: &configuration.SelectTypeOptions{
					Options: []configuration.FieldOption{
						{Label: "Small", Value: "lb-small"},
						{Label: "Medium", Value: "lb-medium"},
						{Label: "Large", Value: "lb-large"},
					},
				},
			},
		},
		{
			Name:        "entryProtocol",
			Label:       "Entry Protocol",
			Type:        configuration.FieldTypeSelect,
			Required:    true,
			Description: "The protocol for incoming traffic",
			TypeOptions: &configuration.TypeOptions{
				Select: &configuration.SelectTypeOptions{
					Options: []configuration.FieldOption{
						{Label: "HTTP", Value: "http"},
						{Label: "HTTPS", Value: "https"},
						{Label: "TCP", Value: "tcp"},
					},
				},
			},
		},
		{
			Name:        "entryPort",
			Label:       "Entry Port",
			Type:        configuration.FieldTypeString,
			Required:    true,
			Description: "The port for incoming traffic",
		},
		{
			Name:        "targetProtocol",
			Label:       "Target Protocol",
			Type:        configuration.FieldTypeSelect,
			Required:    true,
			Description: "The protocol for forwarding to backend droplets",
			TypeOptions: &configuration.TypeOptions{
				Select: &configuration.SelectTypeOptions{
					Options: []configuration.FieldOption{
						{Label: "HTTP", Value: "http"},
						{Label: "HTTPS", Value: "https"},
						{Label: "TCP", Value: "tcp"},
					},
				},
			},
		},
		{
			Name:        "targetPort",
			Label:       "Target Port",
			Type:        configuration.FieldTypeString,
			Required:    true,
			Description: "The port on backend droplets to forward traffic to",
		},
		{
			Name:        "dropletIDs",
			Label:       "Droplet IDs",
			Type:        configuration.FieldTypeList,
			Required:    false,
			Togglable:   true,
			Description: "Numeric IDs of droplets to add as backends",
			TypeOptions: &configuration.TypeOptions{
				List: &configuration.ListTypeOptions{
					ItemLabel: "Droplet ID",
					ItemDefinition: &configuration.ListItemDefinition{
						Type: configuration.FieldTypeString,
					},
				},
			},
		},
		{
			Name:        "tags",
			Label:       "Tags",
			Type:        configuration.FieldTypeList,
			Required:    false,
			Togglable:   true,
			Description: "Tags to apply to the load balancer",
			TypeOptions: &configuration.TypeOptions{
				List: &configuration.ListTypeOptions{
					ItemLabel: "Tag",
					ItemDefinition: &configuration.ListItemDefinition{
						Type: configuration.FieldTypeString,
					},
				},
			},
		},
	}
}

func (c *CreateLoadBalancer) Setup(ctx core.SetupContext) error {
	spec := CreateLoadBalancerSpec{}
	if err := mapstructure.Decode(ctx.Configuration, &spec); err != nil {
		return fmt.Errorf("error decoding configuration: %v", err)
	}

	if spec.Name == "" {
		return errors.New("name is required")
	}

	if spec.Region == "" {
		return errors.New("region is required")
	}

	if spec.EntryProtocol == "" {
		return errors.New("entryProtocol is required")
	}

	if spec.EntryPort == "" {
		return errors.New("entryPort is required")
	}

	if spec.TargetProtocol == "" {
		return errors.New("targetProtocol is required")
	}

	if spec.TargetPort == "" {
		return errors.New("targetPort is required")
	}

	return nil
}

func (c *CreateLoadBalancer) Execute(ctx core.ExecutionContext) error {
	spec := CreateLoadBalancerSpec{}
	if err := mapstructure.Decode(ctx.Configuration, &spec); err != nil {
		return fmt.Errorf("error decoding configuration: %v", err)
	}

	entryPort, err := strconv.Atoi(spec.EntryPort)
	if err != nil {
		return fmt.Errorf("invalid entryPort %q: %v", spec.EntryPort, err)
	}

	targetPort, err := strconv.Atoi(spec.TargetPort)
	if err != nil {
		return fmt.Errorf("invalid targetPort %q: %v", spec.TargetPort, err)
	}

	dropletIDs := make([]int, 0, len(spec.DropletIDs))
	for _, s := range spec.DropletIDs {
		id, err := strconv.Atoi(s)
		if err != nil {
			return fmt.Errorf("invalid droplet ID %q: %v", s, err)
		}
		dropletIDs = append(dropletIDs, id)
	}

	req := CreateLoadBalancerRequest{
		Name:     spec.Name,
		Region:   spec.Region,
		SizeSlug: spec.Size,
		ForwardingRules: []ForwardingRule{
			{
				EntryProtocol:  spec.EntryProtocol,
				EntryPort:      entryPort,
				TargetProtocol: spec.TargetProtocol,
				TargetPort:     targetPort,
			},
		},
		DropletIDs: dropletIDs,
		Tags:       spec.Tags,
	}

	client, err := NewClient(ctx.HTTP, ctx.Integration)
	if err != nil {
		return fmt.Errorf("error creating client: %v", err)
	}

	lb, err := client.CreateLoadBalancer(req)
	if err != nil {
		return fmt.Errorf("failed to create load balancer: %v", err)
	}

	return ctx.ExecutionState.Emit(
		core.DefaultOutputChannel.Name,
		"digitalocean.load_balancer.created",
		[]any{lb},
	)
}

func (c *CreateLoadBalancer) Cancel(ctx core.ExecutionContext) error {
	return nil
}

func (c *CreateLoadBalancer) ProcessQueueItem(ctx core.ProcessQueueContext) (*uuid.UUID, error) {
	return ctx.DefaultProcessing()
}

func (c *CreateLoadBalancer) Actions() []core.Action {
	return []core.Action{}
}

func (c *CreateLoadBalancer) HandleAction(ctx core.ActionContext) error {
	return fmt.Errorf("unknown action: %s", ctx.Name)
}

func (c *CreateLoadBalancer) HandleWebhook(ctx core.WebhookRequestContext) (int, error) {
	return http.StatusOK, nil
}

func (c *CreateLoadBalancer) Cleanup(ctx core.SetupContext) error {
	return nil
}
