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

type DeleteDNSRecord struct{}

type DeleteDNSRecordSpec struct {
	Domain   string `json:"domain"`
	RecordID string `json:"recordID"`
}

func (c *DeleteDNSRecord) Name() string {
	return "digitalocean.deleteDNSRecord"
}

func (c *DeleteDNSRecord) Label() string {
	return "Delete DNS Record"
}

func (c *DeleteDNSRecord) Description() string {
	return "Delete a DNS record from a DigitalOcean domain"
}

func (c *DeleteDNSRecord) Documentation() string {
	return `The Delete DNS Record component removes a DNS record from a DigitalOcean domain.

## Use Cases

- **Cleanup**: Remove DNS records when infrastructure is decommissioned
- **DNS rotation**: Delete old records as part of a DNS update workflow

## Configuration

- **Domain**: The domain name containing the record (required)
- **Record ID**: The numeric ID of the DNS record to delete (required, supports expressions)

## Output

Returns a confirmation object with deleted status, domain, and record ID.`
}

func (c *DeleteDNSRecord) Icon() string {
	return "trash"
}

func (c *DeleteDNSRecord) Color() string {
	return "red"
}

func (c *DeleteDNSRecord) OutputChannels(configuration any) []core.OutputChannel {
	return []core.OutputChannel{core.DefaultOutputChannel}
}

func (c *DeleteDNSRecord) Configuration() []configuration.Field {
	return []configuration.Field{
		{
			Name:        "domain",
			Label:       "Domain",
			Type:        configuration.FieldTypeString,
			Required:    true,
			Description: "The domain name containing the record",
		},
		{
			Name:        "recordID",
			Label:       "Record ID",
			Type:        configuration.FieldTypeString,
			Required:    true,
			Description: "The numeric ID of the DNS record to delete",
		},
	}
}

func (c *DeleteDNSRecord) Setup(ctx core.SetupContext) error {
	spec := DeleteDNSRecordSpec{}
	if err := mapstructure.Decode(ctx.Configuration, &spec); err != nil {
		return fmt.Errorf("error decoding configuration: %v", err)
	}

	if spec.Domain == "" {
		return errors.New("domain is required")
	}

	if spec.RecordID == "" {
		return errors.New("recordID is required")
	}

	return nil
}

func (c *DeleteDNSRecord) Execute(ctx core.ExecutionContext) error {
	spec := DeleteDNSRecordSpec{}
	if err := mapstructure.Decode(ctx.Configuration, &spec); err != nil {
		return fmt.Errorf("error decoding configuration: %v", err)
	}

	id, err := strconv.Atoi(spec.RecordID)
	if err != nil {
		return fmt.Errorf("invalid recordID %q: %v", spec.RecordID, err)
	}

	client, err := NewClient(ctx.HTTP, ctx.Integration)
	if err != nil {
		return fmt.Errorf("error creating client: %v", err)
	}

	if err := client.DeleteDNSRecord(spec.Domain, id); err != nil {
		return fmt.Errorf("failed to delete DNS record: %v", err)
	}

	return ctx.ExecutionState.Emit(
		core.DefaultOutputChannel.Name,
		"digitalocean.dns.record.deleted",
		[]any{map[string]any{"deleted": true, "domain": spec.Domain, "recordID": id}},
	)
}

func (c *DeleteDNSRecord) Cancel(ctx core.ExecutionContext) error {
	return nil
}

func (c *DeleteDNSRecord) ProcessQueueItem(ctx core.ProcessQueueContext) (*uuid.UUID, error) {
	return ctx.DefaultProcessing()
}

func (c *DeleteDNSRecord) Actions() []core.Action {
	return []core.Action{}
}

func (c *DeleteDNSRecord) HandleAction(ctx core.ActionContext) error {
	return fmt.Errorf("unknown action: %s", ctx.Name)
}

func (c *DeleteDNSRecord) HandleWebhook(ctx core.WebhookRequestContext) (int, error) {
	return http.StatusOK, nil
}

func (c *DeleteDNSRecord) Cleanup(ctx core.SetupContext) error {
	return nil
}
