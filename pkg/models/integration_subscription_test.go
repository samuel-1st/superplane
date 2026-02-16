package models

import (
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/superplanehq/superplane/pkg/database"
	"gorm.io/datatypes"
)

func TestFindIntegrationSubscriptionByConfigFields(t *testing.T) {
	require.NoError(t, database.TruncateTables())

	// Create test data
	installationID := uuid.New()
	workflowID := uuid.New()
	nodeID := "test-node-1"

	t.Run("should find subscription with valid fields", func(t *testing.T) {
		// Create subscription with test configuration
		config := map[string]any{
			"message_ts": "1234567890.123456",
			"channel_id": "C123ABC456",
			"type":       "button_click",
		}
		subscription := &IntegrationSubscription{
			InstallationID: installationID,
			WorkflowID:     workflowID,
			NodeID:         nodeID,
			Configuration:  datatypes.NewJSONType[any](config),
		}
		err := database.Conn().Create(subscription).Error
		require.NoError(t, err)

		// Test finding by valid fields
		found, err := FindIntegrationSubscriptionByConfigFields(database.Conn(), installationID, map[string]string{
			"message_ts": "1234567890.123456",
			"channel_id": "C123ABC456",
			"type":       "button_click",
		})
		require.NoError(t, err)
		assert.Equal(t, subscription.ID, found.ID)
		assert.Equal(t, nodeID, found.NodeID)
	})

	t.Run("should return error for invalid field name", func(t *testing.T) {
		// Try to use an invalid/unsafe field name
		_, err := FindIntegrationSubscriptionByConfigFields(database.Conn(), installationID, map[string]string{
			"invalid_field": "some_value",
		})
		require.Error(t, err)
		assert.Contains(t, err.Error(), "unsupported configuration field")
	})

	t.Run("should reject SQL injection attempts", func(t *testing.T) {
		// Try to inject SQL via field name
		_, err := FindIntegrationSubscriptionByConfigFields(database.Conn(), installationID, map[string]string{
			"'; DROP TABLE app_installation_subscriptions; --": "malicious",
		})
		require.Error(t, err)
		assert.Contains(t, err.Error(), "unsupported configuration field")

		// Verify table still exists by querying it
		var count int64
		err = database.Conn().Model(&IntegrationSubscription{}).Count(&count).Error
		require.NoError(t, err, "Table should still exist after injection attempt")
	})

	t.Run("should return error when subscription not found", func(t *testing.T) {
		nonExistentID := uuid.New()
		_, err := FindIntegrationSubscriptionByConfigFields(database.Conn(), nonExistentID, map[string]string{
			"message_ts": "nonexistent",
		})
		require.Error(t, err)
	})

	t.Run("should support querying with subset of fields", func(t *testing.T) {
		// Create another subscription
		newInstallationID := uuid.New()
		config := map[string]any{
			"message_ts": "9999999999.999999",
			"channel_id": "C999XYZ999",
			"type":       "button_click",
		}
		subscription := &IntegrationSubscription{
			InstallationID: newInstallationID,
			WorkflowID:     workflowID,
			NodeID:         "test-node-2",
			Configuration:  datatypes.NewJSONType[any](config),
		}
		err := database.Conn().Create(subscription).Error
		require.NoError(t, err)

		// Find using only one field
		found, err := FindIntegrationSubscriptionByConfigFields(database.Conn(), newInstallationID, map[string]string{
			"message_ts": "9999999999.999999",
		})
		require.NoError(t, err)
		assert.Equal(t, subscription.ID, found.ID)
	})
}
