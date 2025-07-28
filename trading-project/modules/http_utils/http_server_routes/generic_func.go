package httpserverroutes

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

// Define a response structure
type CustomClaims struct {
	Username string `json:"username"`
	jwt.RegisteredClaims
}

var (
	jwtServerAccessSecret  = []byte(os.Getenv("JWT_SECRET"))
	jwtServerRefreshSecret = []byte(os.Getenv("JWT_REFRESH_SECRET"))
)

func generateJWT(username string, userID string, role string, d time.Duration, accessType string, secret []byte) (string, error) {
	// Create token with claims
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"user_id":     userID,
		"username":    username,
		"role":        role,
		"access_type": accessType,
		"exp":         time.Now().Add(d).Unix(), // expiration: time.Now().Add(10 * time.Second).Unix()
	})

	// Sign and return the complete token
	tokenString, err := token.SignedString(secret)
	if err != nil {
		return "", err
	}
	return tokenString, nil
}

// RefreshAccessToken - validate refresh token & issue new access token
func RefreshAccessToken(refreshToken string) (string, error) {
	token, err := jwt.Parse(refreshToken, func(t *jwt.Token) (interface{}, error) {
		return jwtServerRefreshSecret, nil
	})

	if err != nil || !token.Valid {
		return "", fmt.Errorf("invalid refresh token")
	}

	claims := token.Claims.(jwt.MapClaims)

	userID := claims["user_id"].(string)
	username := claims["username"].(string)
	role := claims["role"].(string)

	// Issue a new access token
	return generateJWT(username, userID, role, time.Hour*1, "access", jwtServerAccessSecret)
}

// ValidateJWT parses and validates a JWT token string
func validateJWT(tokenString string, secretKey []byte) (*CustomClaims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &CustomClaims{}, func(token *jwt.Token) (interface{}, error) {
		// Check the signing method
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return secretKey, nil
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

	var tempResult interface{}

	tempResult = result

	if IsValidJSON(result) {
		tempResult = json.RawMessage(result)
	}

	res := map[string]interface{}{
		"status": status,
		"result": tempResult,
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

func getTokenFromHeader(w http.ResponseWriter, r *http.Request) (string, error) {
	// Extract token from Authorization header
	authHeader := r.Header.Get("Authorization")
	const prefix = "Bearer "

	if !strings.HasPrefix(authHeader, prefix) {
		return "", fmt.Errorf("unauthorized - missing bearer prefix")
	}

	strToken := strings.TrimPrefix(authHeader, prefix)

	return strToken, nil
}

func IsValidJSON(s string) bool {
	var js json.RawMessage
	return json.Unmarshal([]byte(s), &js) == nil
}

func getClaimsFromToken(tokenString string, secret []byte) (*Jwt_Claims, error) {
	// Parse token
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		// Ensure the signing method is correct
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return secret, nil
	})

	if err != nil || !token.Valid {
		return nil, fmt.Errorf("invalid token: %v", err)
	}

	// Extract claims
	if claims, ok := token.Claims.(jwt.MapClaims); ok {

		var c Jwt_Claims
		c.UserID = claims["user_id"].(string)
		c.Username = claims["username"].(string)
		c.Role = claims["role"].(string)
		c.AccessType = claims["access_type"].(string)

		return &c, nil
	}

	return nil, fmt.Errorf("cannot parse claims")
}
