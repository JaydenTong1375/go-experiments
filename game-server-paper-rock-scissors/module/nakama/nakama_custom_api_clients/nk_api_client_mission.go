package nakamacustomapiclients

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"

	"github.com/heroiclabs/nakama-common/runtime"
)

// Quest
func CallSaveUserQuest(token string, Quests string) (string, error) {

	host := os.Getenv("HOST")
	url := host + "apiSaveUserQuest"
	bearerToken := token

	// Prepare JSON payload
	payload := map[string]string{
		"quests": Quests,
	}

	jsonData, MarshalErr := json.Marshal(payload)
	if MarshalErr != nil {
		return "", &runtime.Error{Message: fmt.Sprintf("cannot marshal payload: %v", MarshalErr)}
	}

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		return "", fmt.Errorf("request (POST) error: %v", err)
	}

	// Set headers
	req.Header.Set("Authorization", "Bearer "+bearerToken)
	req.Header.Set("Accept", "application/json")
	req.Header.Set("Content-Type", "application/json")

	// Send request
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return "", &runtime.Error{Message: fmt.Sprintf("Error sending request: %v", err)}
	}
	defer resp.Body.Close()

	// Read response
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", &runtime.Error{Message: fmt.Sprintf("cannot read response body: %v", err)}
	}

	var response map[string]interface{}

	unmarshalResErr := json.Unmarshal([]byte(body), &response)

	if unmarshalResErr != nil {
		return "", &runtime.Error{Message: "Failed to unmarshal body."}
	}

	if resp.StatusCode != 200 {
		errMessage, _ := response["result"].(string)
		return "", &runtime.Error{Message: errMessage}
	}

	fmt.Printf("Status Code: %d\n", resp.StatusCode)
	fmt.Println("Response Body:", string(body))

	result, _ := response["result"].(string)

	return result, nil
}

// Quest
func CallGetUserQuest(token string, missionType string) (string, error) {

	host := os.Getenv("HOST")
	url := host + "apiGetUserQuest"
	bearerToken := token

	// Prepare JSON payload
	payload := map[string]string{
		"type": missionType,
	}

	jsonData, MarshalErr := json.Marshal(payload)
	if MarshalErr != nil {
		return "", &runtime.Error{Message: fmt.Sprintf("cannot marshal payload: %v", MarshalErr)}
	}

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		return "", fmt.Errorf("request (POST) error: %v", err)
	}

	// Set headers
	req.Header.Set("Authorization", "Bearer "+bearerToken)
	req.Header.Set("Accept", "application/json")
	req.Header.Set("Content-Type", "application/json")

	// Send request
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return "", &runtime.Error{Message: fmt.Sprintf("Error sending request: %v", err)}
	}
	defer resp.Body.Close()

	// Read response
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", &runtime.Error{Message: fmt.Sprintf("cannot read response body: %v", err)}
	}

	var response map[string]interface{}

	unmarshalResErr := json.Unmarshal([]byte(body), &response)

	if unmarshalResErr != nil {
		return "", &runtime.Error{Message: "Failed to unmarshal body."}
	}

	if resp.StatusCode != 200 {
		errMessage, _ := response["result"].(string)
		return "", &runtime.Error{Message: errMessage}
	}

	fmt.Printf("Status Code: %d\n", resp.StatusCode)
	fmt.Println("Response Body:", string(body))

	result, _ := response["result"].(string)

	return fmt.Sprintf(`{"result": %s }`, result), nil
}

// Quest
func CallUpdateUserQuest(token string, objectiveID string, objectiveName string, objectiveDes string, objectiveProgress float64) (string, error) {

	host := os.Getenv("HOST")
	url := host + "apiUpdateUserQuest"
	bearerToken := token

	// Prepare JSON payload
	payload := map[string]interface{}{
		"objective_id":          objectiveID,
		"objective_name":        objectiveName,
		"objective_description": objectiveDes,
		"objective_progress":    objectiveProgress,
	}

	jsonData, MarshalErr := json.Marshal(payload)
	if MarshalErr != nil {
		return "", &runtime.Error{Message: fmt.Sprintf("cannot marshal payload: %v", MarshalErr)}
	}

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		return "", fmt.Errorf("request (POST) error: %v", err)
	}

	// Set headers
	req.Header.Set("Authorization", "Bearer "+bearerToken)
	req.Header.Set("Accept", "application/json")
	req.Header.Set("Content-Type", "application/json")

	// Send request
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return "", &runtime.Error{Message: fmt.Sprintf("Error sending request: %v", err)}
	}
	defer resp.Body.Close()

	// Read response
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", &runtime.Error{Message: fmt.Sprintf("cannot read response body: %v", err)}
	}

	var response map[string]interface{}

	unmarshalResErr := json.Unmarshal([]byte(body), &response)

	if unmarshalResErr != nil {
		return "", &runtime.Error{Message: "Failed to unmarshal body."}
	}

	if resp.StatusCode != 200 {
		errMessage, _ := response["result"].(string)
		return "", &runtime.Error{Message: errMessage}
	}

	fmt.Printf("Status Code: %d\n", resp.StatusCode)
	fmt.Println("Response Body:", string(body))

	result, _ := response["result"].(string)

	return fmt.Sprintf(`{"result": %s }`, result), nil
}
