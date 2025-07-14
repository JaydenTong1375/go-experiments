package nakamacustomapiclients

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
)

func CallRegisterUserAPI(token string, username string, email string) error {
	host := os.Getenv("HOST")
	url := host + "apiRegisterUser"
	bearerToken := token

	// Prepare JSON payload
	payload := map[string]string{
		"username": username,
		"email":    email,
	}
	jsonData, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("cannot marshal payload: %v", err)
	}

	// Create POST request with JSON body
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		return fmt.Errorf("request (POST) error: %v", err)
	}

	// Set headers
	req.Header.Set("Authorization", "Bearer "+bearerToken)
	req.Header.Set("Accept", "application/json")
	req.Header.Set("Content-Type", "application/json")

	// Send request
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("error sending request: %v", err)
	}
	defer resp.Body.Close()

	// Read response
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("cannot read response body: %v", err)
	}

	fmt.Printf("Status Code: %d\n", resp.StatusCode)
	fmt.Println("Response Body:", string(body))

	return nil
}
