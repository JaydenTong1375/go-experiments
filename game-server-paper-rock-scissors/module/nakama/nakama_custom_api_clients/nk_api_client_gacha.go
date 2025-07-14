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

func CallUpdateUserCreditAPI(token string, status string, value int) (string, error) {
	host := os.Getenv("HOST")
	url := host + "apiUpdateUserCredit"
	bearerToken := token

	// Prepare JSON payload
	payload := map[string]interface{}{
		"status": status,
		"value":  value,
	}

	jsonData, MarshalErr := json.Marshal(payload)
	if MarshalErr != nil {
		return "", &runtime.Error{Message: fmt.Sprintf("cannot marshal payload: %v", MarshalErr)}
	}

	// Create POST request with JSON body
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		return "", &runtime.Error{Message: fmt.Sprintf("request (POST) error: %v", err)}
	}

	// Set headers
	req.Header.Set("Authorization", "Bearer "+bearerToken)
	req.Header.Set("Accept", "application/json")
	req.Header.Set("Content-Type", "application/json")

	// Send request
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return "", &runtime.Error{Message: fmt.Sprintf("error sending request: %v", err)}
	}
	defer resp.Body.Close()

	// Read response
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", &runtime.Error{Message: fmt.Sprintf("Cannot read response body: %v", err)}
	}

	var response map[string]interface{}

	unmarshalResErr := json.Unmarshal([]byte(body), &response)

	if unmarshalResErr != nil {
		return "", &runtime.Error{Message: fmt.Sprintf("Failed to unmarshal body: %v", unmarshalResErr)}
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

func CallSpinForRewardAPI(token string, items string) (string, error) {
	host := os.Getenv("HOST")
	url := host + "apiSpinForReward"
	bearerToken := token

	// Prepare JSON payload
	payload := map[string]string{
		"items": items,
	}

	jsonData, MarshalErr := json.Marshal(payload)
	if MarshalErr != nil {
		return "", &runtime.Error{Message: fmt.Sprintf("cannot marshal payload: %v", MarshalErr)}
	}

	// Create POST request with JSON body
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		return "", &runtime.Error{Message: fmt.Sprintf("request (POST) error: %v", err)}
	}

	// Set headers
	req.Header.Set("Authorization", "Bearer "+bearerToken)
	req.Header.Set("Accept", "application/json")
	req.Header.Set("Content-Type", "application/json")

	// Send request
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return "", &runtime.Error{Message: fmt.Sprintf("error sending request: %v", err)}
	}
	defer resp.Body.Close()

	// Read response
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", &runtime.Error{Message: fmt.Sprintf("Cannot read response body: %v", err)}
	}

	var response map[string]interface{}

	unmarshalResErr := json.Unmarshal([]byte(body), &response)

	if unmarshalResErr != nil {
		return "", &runtime.Error{Message: fmt.Sprintf("Failed to unmarshal body: %v", unmarshalResErr)}
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

func CallGetUserCredit(token string) (string, error) {
	host := os.Getenv("HOST")
	url := host + "apiGetUserCredit"
	bearerToken := token

	req, err := http.NewRequest("POST", url, nil)
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
