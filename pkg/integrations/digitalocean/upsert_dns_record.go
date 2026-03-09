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

type UpsertDNSRecord struct{}

type UpsertDNSRecordSpec struct {
	Domain   string `json:"domain"`
	Type     string `json:"type"`
	Name     string `json:"name"`
	Data     string `json:"data"`
	TTL      string `json:"ttl"`
	Priority string `json:"priority"`
}

func (c *UpsertDNSRecord) Name() string {
	return "digitalocean.upsertDNSRecord"
}

func (c *UpsertDNSRecord) Label() string {
	return "Upsert DNS Record"
}

func (c *UpsertDNSRecord) Description() string {
	return "Create or update a DNS record for a DigitalOcean domain"
}

func (c *UpsertDNSRecord) Documentation() string {
	return `The Upsert DNS Record component creates or updates a DNS record in a DigitalOcean domain. If a record with the same name and type already exists it will be updated; otherwise a new record is created.

## Use Cases

- **Idempotent DNS management**: Ensure a record exists with the correct value without worrying about duplicates
- **IP updates**: Point an existing hostname to a new IP address

## Configuration

- **Domain**: The domain name to manage the record in (required)
- **Type**: The DNS record type (required)
- **Name**: The host name, alias, or service being defined (required)
- **Data**: The value of the record, e.g. an IP address (required)
- **TTL**: Time to live in seconds (optional, default 1800)
- **Priority**: Priority for MX or SRV records (optional)

## Output

Returns the created or updated DNS record object.`
}

func (c *UpsertDNSRecord) Icon() string {
	return "globe"
}

func (c *UpsertDNSRecord) Color() string {
	return "blue"
}

func (c *UpsertDNSRecord) OutputChannels(configuration any) []core.OutputChannel {
	return []core.OutputChannel{core.DefaultOutputChannel}
}

func (c *UpsertDNSRecord) Configuration() []configuration.Field {
	return []configuration.Field{
		{
			Name:        "domain",
			Label:       "Domain",
			Type:        configuration.FieldTypeString,
			Required:    true,
			Description: "The domain name to manage the record in",
		},
		{
			Name:        "type",
			Label:       "Record Type",
			Type:        configuration.FieldTypeSelect,
			Required:    true,
			Description: "The DNS record type",
			TypeOptions: &configuration.TypeOptions{
				Select: &configuration.SelectTypeOptions{
					Options: []configuration.FieldOption{
						{Label: "A", Value: "A"},
						{Label: "AAAA", Value: "AAAA"},
						{Label: "CNAME", Value: "CNAME"},
						{Label: "MX", Value: "MX"},
						{Label: "TXT", Value: "TXT"},
						{Label: "NS", Value: "NS"},
						{Label: "SRV", Value: "SRV"},
						{Label: "CAA", Value: "CAA"},
					},
				},
			},
		},
		{
			Name:        "name",
			Label:       "Name",
			Type:        configuration.FieldTypeString,
			Required:    true,
			Description: "The host name, alias, or service being defined",
		},
		{
			Name:        "data",
			Label:       "Value",
			Type:        configuration.FieldTypeString,
			Required:    true,
			Description: "The value of the record, e.g. an IP address or hostname",
		},
		{
			Name:        "ttl",
			Label:       "TTL",
			Type:        configuration.FieldTypeString,
			Required:    false,
			Togglable:   true,
			Description: "Time to live in seconds (default: 1800)",
			Placeholder: "1800",
		},
		{
			Name:        "priority",
			Label:       "Priority",
			Type:        configuration.FieldTypeString,
			Required:    false,
			Togglable:   true,
			Description: "Priority for MX or SRV records",
		},
	}
}

func (c *UpsertDNSRecord) Setup(ctx core.SetupContext) error {
	spec := UpsertDNSRecordSpec{}
	if err := mapstructure.Decode(ctx.Configuration, &spec); err != nil {
		return fmt.Errorf("error decoding configuration: %v", err)
	}

	if spec.Domain == "" {
		return errors.New("domain is required")
	}

	if spec.Type == "" {
		return errors.New("type is required")
	}

	if spec.Name == "" {
		return errors.New("name is required")
	}

	if spec.Data == "" {
		return errors.New("data is required")
	}

	return nil
}

func (c *UpsertDNSRecord) Execute(ctx core.ExecutionContext) error {
	spec := UpsertDNSRecordSpec{}
	if err := mapstructure.Decode(ctx.Configuration, &spec); err != nil {
		return fmt.Errorf("error decoding configuration: %v", err)
	}

	ttl := 1800
	if spec.TTL != "" {
		parsed, err := strconv.Atoi(spec.TTL)
		if err != nil {
			return fmt.Errorf("invalid ttl %q: %v", spec.TTL, err)
		}
		ttl = parsed
	}

	req := DNSRecordRequest{
		Type: spec.Type,
		Name: spec.Name,
		Data: spec.Data,
		TTL:  ttl,
	}

	if spec.Priority != "" {
		p, err := strconv.Atoi(spec.Priority)
		if err != nil {
			return fmt.Errorf("invalid priority %q: %v", spec.Priority, err)
		}
		req.Priority = p
	}

	client, err := NewClient(ctx.HTTP, ctx.Integration)
	if err != nil {
		return fmt.Errorf("error creating client: %v", err)
	}

	existing, err := client.ListDNSRecords(spec.Domain, fmt.Sprintf("type=%s&name=%s", spec.Type, spec.Name))
	if err != nil {
		return fmt.Errorf("failed to list DNS records: %v", err)
	}

	var record *DNSRecord
	if len(existing) > 0 {
		record, err = client.UpdateDNSRecord(spec.Domain, existing[0].ID, req)
		if err != nil {
			return fmt.Errorf("failed to update DNS record: %v", err)
		}
	} else {
		record, err = client.CreateDNSRecord(spec.Domain, req)
		if err != nil {
			return fmt.Errorf("failed to create DNS record: %v", err)
		}
	}

	return ctx.ExecutionState.Emit(
		core.DefaultOutputChannel.Name,
		"digitalocean.dns.record.upserted",
		[]any{record},
	)
}

func (c *UpsertDNSRecord) Cancel(ctx core.ExecutionContext) error {
	return nil
}

func (c *UpsertDNSRecord) ProcessQueueItem(ctx core.ProcessQueueContext) (*uuid.UUID, error) {
	return ctx.DefaultProcessing()
}

func (c *UpsertDNSRecord) Actions() []core.Action {
	return []core.Action{}
}

func (c *UpsertDNSRecord) HandleAction(ctx core.ActionContext) error {
	return fmt.Errorf("unknown action: %s", ctx.Name)
}

func (c *UpsertDNSRecord) HandleWebhook(ctx core.WebhookRequestContext) (int, error) {
	return http.StatusOK, nil
}

func (c *UpsertDNSRecord) Cleanup(ctx core.SetupContext) error {
	return nil
}
