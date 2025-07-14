package apiserver

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

// Define a response structure
type CustomClaims struct {
	Username string `json:"username"`
	jwt.RegisteredClaims
}

var (
	jwtServerAccessSecret = []byte("your_secret_key")
	//jwtRefreshSecret = []byte("your-refresh-secret")
)

func generateJWT(username string) (string, error) {
	// Create token with claims
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"username": username,
		"exp":      time.Now().Add(time.Hour * 1).Unix(), // expiration: time.Now().Add(10 * time.Second).Unix()
	})

	// Sign and return the complete token
	tokenString, err := token.SignedString(jwtServerAccessSecret)
	if err != nil {
		return "", err
	}
	return tokenString, nil
}

// ValidateJWT parses and validates a JWT token string
func validateJWT(tokenString string) (*CustomClaims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &CustomClaims{}, func(token *jwt.Token) (interface{}, error) {
		// Check the signing method
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return jwtServerAccessSecret, nil
	})

	if err != nil {
		return nil, err
	}

	claims, ok := token.Claims.(*CustomClaims)
	if !ok || !token.Valid {
		return nil, errors.New("invalid token claims")
	}

	return claims, nil
}

func writeResultJSON(w http.ResponseWriter, result string, status int) {

	log.Printf("Status: %d | Result: %s \n", status, json.RawMessage(result))

	w.Header().Set("Content-Type", "application/json")

	res := map[string]interface{}{
		"status": status,
		"result": json.RawMessage(result),
	}

	jsonResponse, MarshalErr := json.MarshalIndent(res, "", "  ")

	if MarshalErr != nil {

		errRes := map[string]interface{}{
			"status": http.StatusBadRequest,
			"result": fmt.Sprintf("❌ failed to marhsal result. Result: %v", result),
		}

		jsonResponse, _ := json.Marshal(errRes)

		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(jsonResponse))
		return
	}

	w.WriteHeader(status)
	w.Write(jsonResponse)
}
