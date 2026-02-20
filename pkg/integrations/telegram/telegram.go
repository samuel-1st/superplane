package telegram

import (
	"fmt"

	"github.com/superplanehq/superplane/pkg/configuration"
	"github.com/superplanehq/superplane/pkg/core"
	"github.com/superplanehq/superplane/pkg/registry"
)

func init() {
	registry.RegisterIntegration("telegram", &Telegram{})
}

type Telegram struct{}

type Configuration struct {
	BotToken string `json:"botToken" mapstructure:"botToken"`
}

type Metadata struct {
	BotID    int    `json:"botId" mapstructure:"botId"`
	Username string `json:"username" mapstructure:"username"`
}

func (t *Telegram) Name() string {
	return "telegram"
}

func (t *Telegram) Label() string {
	return "Telegram"
}

func (t *Telegram) Icon() string {
	return "telegram"
}

func (t *Telegram) Description() string {
	return "Send messages via Telegram Bot API"
}

func (t *Telegram) Instructions() string {
	return `To set up Telegram integration:

1. Open Telegram and search for **@BotFather**
2. Send **/newbot** and follow the prompts to create a new bot
3. BotFather will provide a **Bot Token** (e.g. ` + "`" + `123456:ABC-DEF...` + "`" + `)
4. To find your **Chat ID**, send a message to your bot and call:
   ` + "`" + `https://api.telegram.org/bot<token>/getUpdates` + "`" + `
   The ` + "`" + `chat.id` + "`" + ` field in the response is your Chat ID
5. Paste the **Bot Token** in the **Bot Token** field below`
}

func (t *Telegram) Configuration() []configuration.Field {
	return []configuration.Field{
		{
			Name:        "botToken",
			Label:       "Bot Token",
			Type:        configuration.FieldTypeString,
			Required:    true,
			Sensitive:   true,
			Description: "Telegram bot token from BotFather",
		},
	}
}

func (t *Telegram) Components() []core.Component {
	return []core.Component{
		&SendTextMessage{},
	}
}

func (t *Telegram) Triggers() []core.Trigger {
	return []core.Trigger{}
}

func (t *Telegram) Sync(ctx core.SyncContext) error {
	botTokenBytes, err := ctx.Integration.GetConfig("botToken")
	if err != nil {
		return fmt.Errorf("botToken is required")
	}

	botToken := string(botTokenBytes)
	if botToken == "" {
		return fmt.Errorf("botToken is required")
	}

	client, err := NewClient(ctx.Integration)
	if err != nil {
		return err
	}

	me, err := client.GetMe()
	if err != nil {
		return fmt.Errorf("failed to verify bot token: %v", err)
	}

	ctx.Integration.SetMetadata(Metadata{
		BotID:    me.ID,
		Username: me.Username,
	})

	ctx.Integration.Ready()
	return nil
}

func (t *Telegram) HandleRequest(ctx core.HTTPRequestContext) {
}

func (t *Telegram) Cleanup(ctx core.IntegrationCleanupContext) error {
	return nil
}

func (t *Telegram) ListResources(resourceType string, ctx core.ListResourcesContext) ([]core.IntegrationResource, error) {
	return []core.IntegrationResource{}, nil
}

func (t *Telegram) Actions() []core.Action {
	return []core.Action{}
}

func (t *Telegram) HandleAction(ctx core.IntegrationActionContext) error {
	return nil
}
