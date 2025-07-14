package nkamacustomrpc

import (
	"context"
	"database/sql"
	"encoding/json"

	"github.com/heroiclabs/nakama-common/runtime"

	nkAPIClient "game-server-paper-rock-scissors/module/nakama/nakama_custom_api_clients"
)

func RPCgetInventoryData(ctx context.Context, logger runtime.Logger, db *sql.DB, nk runtime.NakamaModule, payload string) (string, error) {

	logger.Debug("payload: %v", payload)

	var jsonPayload map[string]interface{}

	if UnmarshalErr := json.Unmarshal([]byte(payload), &jsonPayload); UnmarshalErr != nil {
		return "", &runtime.Error{Message: "Invalid JSON payload"}
	}

	token, bIsTokenValid := jsonPayload["token"].(string)

	if bIsTokenValid == false {
		return "", &runtime.Error{Message: "Token not found in payload"}
	}

	return nkAPIClient.CallGetInventoryItems(token)
}

func RPCclearInventory(ctx context.Context, logger runtime.Logger, db *sql.DB, nk runtime.NakamaModule, payload string) (string, error) {

	logger.Debug("payload: %v", payload)

	var jsonPayload map[string]interface{}

	if UnmarshalErr := json.Unmarshal([]byte(payload), &jsonPayload); UnmarshalErr != nil {
		return "", &runtime.Error{Message: "Invalid JSON payload"}
	}

	token, bIsTokenValid := jsonPayload["token"].(string)

	if bIsTokenValid == false {
		return "", &runtime.Error{Message: "Token not found in payload"}
	}

	return nkAPIClient.CallClearInventory(token)
}
