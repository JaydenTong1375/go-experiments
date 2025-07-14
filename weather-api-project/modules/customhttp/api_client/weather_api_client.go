package apiclient

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
)

func GetWeatherDataFromOfficialWeb(country string, date string) (string, error) {
	urlBase := os.Getenv("urlBase_weather")
	apiKey := os.Getenv("APIKey_VisualCrossing")

	fullUrl := fmt.Sprintf(`%s%s/%s`, urlBase, country, date)

	// Parse the base URL
	u, err := url.Parse(fullUrl)
	if err != nil {
		return "", err
	}

	// Add query parameters
	query := u.Query()
	query.Set("key", apiKey)
	u.RawQuery = query.Encode()

	log.Printf("api url = %s", u.String())

	// Create the request
	req, err := http.NewRequest("GET", u.String(), nil)
	if err != nil {
		return "", err
	}

	// Send the request
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	// Read the response body using io.ReadAll (recommended since Go 1.16)
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	return string(body), nil
}
