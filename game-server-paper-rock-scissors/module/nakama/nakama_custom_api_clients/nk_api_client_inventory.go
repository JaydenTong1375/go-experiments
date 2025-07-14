package nakamacustomapiclients

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"

	"github.com/heroiclabs/nakama-common/runtime"
)

func CallClearInventory(token string) (string, error) {

	host := os.Getenv("HOST")
	url := host + "apiClearInventory"
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

func CallGetInventoryItems(token string) (string, error) {

	host := os.Getenv("HOST")
	url := host + "apiGetInventoryItems"
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
