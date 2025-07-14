package nkamacustomrpc

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"

	ncac "game-server-paper-rock-scissors/module/nakama/nakama_custom_api_clients"

	"github.com/heroiclabs/nakama-common/runtime"
)

//Quest

func RPCSaveUserQuest(ctx context.Context, logger runtime.Logger, db *sql.DB, nk runtime.NakamaModule, payload string) (string, error) {

	logger.Debug("payload: %v", payload)

	var jsonPayload map[string]interface{}

	if UnmarshalErr := json.Unmarshal([]byte(payload), &jsonPayload); UnmarshalErr != nil {
		return "", &runtime.Error{Message: "Invalid JSON payload"}
	}

	token, bIsTokenValid := jsonPayload["token"].(string)

	quests, bIsQuestsValid := jsonPayload["quests"].([]interface{})

	if bIsTokenValid == false {
		return "", &runtime.Error{Message: "token not found in payload"}
	}

	if bIsQuestsValid == false {
		return "", &runtime.Error{Message: "quests not found in payload"}
	}

	jsonQuest, questMarshalErr := json.Marshal(quests)

	if questMarshalErr != nil {
		return "", &runtime.Error{Message: fmt.Sprintf("failed to marshal quest data. %v", questMarshalErr)}
	}

	return ncac.CallSaveUserQuest(token, string(jsonQuest))
}

func RPCGetUserQuest(ctx context.Context, logger runtime.Logger, db *sql.DB, nk runtime.NakamaModule, payload string) (string, error) {

	logger.Debug("RPCGetUserQuest payload: %v", payload)

	var jsonPayload map[string]interface{}

	if UnmarshalErr := json.Unmarshal([]byte(payload), &jsonPayload); UnmarshalErr != nil {
		return "", &runtime.Error{Message: "Invalid JSON payload"}
	}

	token, bIsTokenValid := jsonPayload["token"].(string)

	missionType, bIsTypeValid := jsonPayload["type"].(string)

	if bIsTokenValid == false {
		return "", &runtime.Error{Message: "token not found in payload"}
	}

	if bIsTypeValid == false {
		return "", &runtime.Error{Message: "type not found in payload"}
	}

	return ncac.CallGetUserQuest(token, missionType)
}

func RPCUpdateUserQuest(ctx context.Context, logger runtime.Logger, db *sql.DB, nk runtime.NakamaModule, payload string) (string, error) {

	logger.Debug("RPCUpdateUserQuest payload: %v", payload)

	var jsonPayload map[string]interface{}

	if UnmarshalErr := json.Unmarshal([]byte(payload), &jsonPayload); UnmarshalErr != nil {
		return "", &runtime.Error{Message: "Invalid JSON payload"}
	}

	token, bIsTokenValid := jsonPayload["token"].(string)

	objectiveID, bIsObjectiveIDValid := jsonPayload["objective_id"].(string)

	objectiveName, bIsObjectiveNameValid := jsonPayload["objective_name"].(string)

	objectiveDes, bIsObjectiveDesValid := jsonPayload["objective_description"].(string)

	objectiveProgress, bIsObjectiveProgressValid := jsonPayload["objective_progress"].(float64)

	if bIsTokenValid == false {
		return "", &runtime.Error{Message: "token not found in payload"}
	}

	if bIsObjectiveIDValid == false {
		return "", &runtime.Error{Message: "objective_id not found in payload"}
	}

	if bIsObjectiveNameValid == false {
		return "", &runtime.Error{Message: "objective_name not found in payload"}
	}

	if bIsObjectiveDesValid == false {
		return "", &runtime.Error{Message: "objective_description not found in payload"}
	}

	if bIsObjectiveProgressValid == false {
		return "", &runtime.Error{Message: "objective_progress not found in payload"}
	}

	return ncac.CallUpdateUserQuest(token, objectiveID, objectiveName, objectiveDes, objectiveProgress)
}
