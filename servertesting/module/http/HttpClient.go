package http

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

func TestHttpsClientAPIKeyAuth() {

	fmt.Println("API Key Authentication")

	// Create a new HTTP request
	req, err := http.NewRequest("POST", "https://fucherng1375.ddns.net:8090/apiHandlerAPIKeyAuth", nil)
	if err != nil {
		panic(err)
	}

	// Set the Authorization header BEFORE sending
	req.Header.Set("Authorization", "my-secret-api-key")

	// Send the request
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		panic(err)
	}
	defer resp.Body.Close()

	fmt.Println("Response status:", resp.Status)

	// Read the whole body
	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		panic("Failed to read body: " + err.Error())
	}

	fmt.Println("Response body:", string(bodyBytes))

	// Unmarshal the outer message
	var message Message
	if err := json.Unmarshal(bodyBytes, &message); err != nil {
		fmt.Println("Failed to Unmarshal message:", err)
		return
	}

	// Unmarshal the nested Person struct
	var person map[string]interface{}
	if err := json.Unmarshal([]byte(message.Message), &person); err != nil {
		fmt.Println("Failed to Unmarshal person:", err)
		return
	}

	fmt.Println("Name:", person["name"])
	println()
}

func TestHttpsClientBearerTokenAuth() {

	fmt.Println("API Bearer Authentication")

	// Create a new HTTP request
	req, err := http.NewRequest("POST", "https://fucherng1375.ddns.net:8090/apiHandlerBearerAuth", nil)
	if err != nil {
		panic(err)
	}

	// Set the Authorization header BEFORE sending
	req.Header.Set("Authorization", "Bearer your-secret-token")

	// Send the request
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		panic(err)
	}
	defer resp.Body.Close()

	fmt.Println("Response status:", resp.Status)

	// Read the whole body
	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		panic("Failed to read body: " + err.Error())
	}

	fmt.Println("Response body:", string(bodyBytes))

	// Unmarshal the outer message
	var message Message
	if err := json.Unmarshal(bodyBytes, &message); err != nil {
		fmt.Println("Failed to Unmarshal message:", err)
		return
	}

	// Unmarshal the nested Person struct
	var person map[string]interface{}
	if err := json.Unmarshal([]byte(message.Message), &person); err != nil {
		fmt.Println("Failed to Unmarshal person:", err)
		return
	}

	fmt.Println("Hi, ", person["name"])
	println()
}

func AuthenticateUser(username, password string) (string, error) {
	fmt.Println("Authenticating User...")

	// Prepare the request body
	authData := make(map[string]interface{})
	authData["username"] = username
	authData["password"] = password

	jsonData, err := json.Marshal(authData)
	if err != nil {
		return "", fmt.Errorf("failed to marshal json: %v", err)
	}

	// Create a new POST request with JSON body
	req, err := http.NewRequest("POST", "https://fucherng1375.ddns.net:8090/apiAuthenticateUser", bytes.NewBuffer(jsonData))
	if err != nil {
		panic(err)
	}

	// Set headers
	req.Header.Set("Content-Type", "application/json")
	// Optional: Set an API key or token if needed
	// req.Header.Set("Authorization", "my-secret-api-key")

	// Send the request
	resp, HttpSendErr := http.DefaultClient.Do(req)
	if HttpSendErr != nil {
		return "", fmt.Errorf("failed to send: %v", HttpSendErr)
	}
	defer resp.Body.Close()

	fmt.Println("Response status:", resp.Status)

	// Read the response body
	bodyBytes, bodyBytesErr := io.ReadAll(resp.Body)
	if bodyBytesErr != nil {
		return "", fmt.Errorf("failed to read body: %v", bodyBytesErr)
	}

	fmt.Println("Response body:", string(bodyBytes))

	// Unmarshal the outer message
	var message Message
	if UnmarshalMessageErr := json.Unmarshal(bodyBytes, &message); UnmarshalMessageErr != nil {
		fmt.Println("Failed to Unmarshal message:", UnmarshalMessageErr)
		return "", fmt.Errorf("failed to marshal json: %v", UnmarshalMessageErr)
	}

	fmt.Println("Received Token: ", message.Message)

	return message.Message, nil
}
