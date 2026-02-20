package telegram

import (
	"fmt"

	"github.com/mitchellh/mapstructure"
	"github.com/superplanehq/superplane/pkg/configuration"
	"github.com/superplanehq/superplane/pkg/core"
)

type OnMessageReceived struct{}

type OnMessageReceivedMetadata struct {
	SubscriptionID *string `json:"subscriptionId,omitempty" mapstructure:"subscriptionId,omitempty"`
}

func (t *OnMessageReceived) Name() string {
	return "telegram.onMessageReceived"
}

func (t *OnMessageReceived) Label() string {
	return "On Message Received"
}

func (t *OnMessageReceived) Description() string {
	return "Fires when the bot receives a message or is mentioned"
}

func (t *OnMessageReceived) Documentation() string {
	return `The On Message Received trigger fires when the Telegram bot receives a message.

## Use Cases

- **Bot commands**: React to commands sent to the bot
- **Mentions**: Start a workflow when the bot is mentioned in a group
- **Direct messages**: Process messages sent directly to the bot

## Event Data

Each event includes the full Telegram message object:
- **message_id**: Unique message identifier
- **from**: Sender information (id, first_name, username)
- **chat**: Chat information (id, type, title)
- **text**: Message text
- **date**: Unix timestamp of the message
- **entities**: Special entities in the message (mentions, commands, etc.)

## Notes

- Requires the Telegram integration to be configured with a valid bot token
- SuperPlane automatically registers a webhook with Telegram on integration setup`
}

func (t *OnMessageReceived) Icon() string {
	return "telegram"
}

func (t *OnMessageReceived) Color() string {
	return "blue"
}

func (t *OnMessageReceived) Configuration() []configuration.Field {
	return []configuration.Field{}
}

func (t *OnMessageReceived) Setup(ctx core.TriggerContext) error {
	var metadata OnMessageReceivedMetadata
	if err := mapstructure.Decode(ctx.Metadata.Get(), &metadata); err != nil {
		return fmt.Errorf("failed to decode metadata: %w", err)
	}

	if metadata.SubscriptionID != nil {
		return nil
	}

	subscriptionID, err := ctx.Integration.Subscribe(SubscriptionConfiguration{
		EventTypes: []string{"message.received"},
	})
	if err != nil {
		return fmt.Errorf("failed to subscribe to messages: %w", err)
	}

	s := subscriptionID.String()
	return ctx.Metadata.Set(OnMessageReceivedMetadata{
		SubscriptionID: &s,
	})
}

func (t *OnMessageReceived) OnIntegrationMessage(ctx core.IntegrationMessageContext) error {
	return ctx.Events.Emit("telegram.message.received", ctx.Message)
}

func (t *OnMessageReceived) HandleWebhook(ctx core.WebhookRequestContext) (int, error) {
	return 200, nil
}

func (t *OnMessageReceived) Actions() []core.Action {
	return []core.Action{}
}

func (t *OnMessageReceived) HandleAction(ctx core.TriggerActionContext) (map[string]any, error) {
	return nil, nil
}

func (t *OnMessageReceived) Cleanup(ctx core.TriggerContext) error {
	return nil
}
