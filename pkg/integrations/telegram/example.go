package telegram

import (
	_ "embed"
	"sync"

	"github.com/superplanehq/superplane/pkg/utils"
)

//go:embed example_output_send_text_message.json
var exampleOutputSendTextMessageBytes []byte

//go:embed example_data_on_message_received.json
var exampleDataOnMessageReceivedBytes []byte

var exampleOutputSendTextMessageOnce sync.Once
var exampleOutputSendTextMessage map[string]any

var exampleDataOnMessageReceivedOnce sync.Once
var exampleDataOnMessageReceived map[string]any

func (c *SendTextMessage) ExampleOutput() map[string]any {
	return utils.UnmarshalEmbeddedJSON(&exampleOutputSendTextMessageOnce, exampleOutputSendTextMessageBytes, &exampleOutputSendTextMessage)
}

func (t *OnMessageReceived) ExampleData() map[string]any {
	return utils.UnmarshalEmbeddedJSON(&exampleDataOnMessageReceivedOnce, exampleDataOnMessageReceivedBytes, &exampleDataOnMessageReceived)
}

