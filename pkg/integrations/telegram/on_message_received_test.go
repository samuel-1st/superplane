package telegram

import (
	"testing"

	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/superplanehq/superplane/pkg/core"
	"github.com/superplanehq/superplane/test/support/contexts"
)

func Test__OnMessageReceived__Setup(t *testing.T) {
	trigger := &OnMessageReceived{}

	t.Run("no existing subscription -> creates subscription", func(t *testing.T) {
		metadata := &contexts.MetadataContext{}
		integrationCtx := &contexts.IntegrationContext{}

		err := trigger.Setup(core.TriggerContext{
			Integration:   integrationCtx,
			Metadata:      metadata,
			Configuration: map[string]any{},
		})

		require.NoError(t, err)
		require.Len(t, integrationCtx.Subscriptions, 1)

		subConfig, ok := integrationCtx.Subscriptions[0].Configuration.(SubscriptionConfiguration)
		require.True(t, ok)
		assert.Equal(t, []string{"message.received"}, subConfig.EventTypes)

		stored, ok := metadata.Metadata.(OnMessageReceivedMetadata)
		require.True(t, ok)
		require.NotNil(t, stored.SubscriptionID)
	})

	t.Run("existing subscription -> no-op", func(t *testing.T) {
		existingID := uuid.NewString()
		metadata := &contexts.MetadataContext{
			Metadata: OnMessageReceivedMetadata{
				SubscriptionID: &existingID,
			},
		}
		integrationCtx := &contexts.IntegrationContext{}

		err := trigger.Setup(core.TriggerContext{
			Integration:   integrationCtx,
			Metadata:      metadata,
			Configuration: map[string]any{},
		})

		require.NoError(t, err)
		assert.Empty(t, integrationCtx.Subscriptions)
	})
}

func Test__OnMessageReceived__OnIntegrationMessage(t *testing.T) {
	trigger := &OnMessageReceived{}

	t.Run("emits telegram.message.received with message payload", func(t *testing.T) {
		message := &TelegramMessage{
			MessageID: 42,
			Chat:      TelegramChat{ID: -1001234567890, Type: "group", Title: "My Team"},
			Text:      "@mybot deploy!",
			Date:      1737028800,
		}

		events := &contexts.EventContext{}
		err := trigger.OnIntegrationMessage(core.IntegrationMessageContext{
			Configuration: map[string]any{},
			Message:       message,
			Logger:        logrus.NewEntry(logrus.New()),
			Events:        events,
		})

		require.NoError(t, err)
		require.Equal(t, 1, events.Count())
		assert.Equal(t, "telegram.message.received", events.Payloads[0].Type)
		assert.Equal(t, message, events.Payloads[0].Data)
	})
}
