package nakamacustomapiserver

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/golang-jwt/jwt"
)

//////////////////////////////////////// Generic function ///////////////////////////////////////////

func openDB() (*sql.DB, error) {

	dsn := "myuser:mypassword@tcp(mariadb:3306)/paper-rock-scissors"

	db, OpenErr := sql.Open("mysql", dsn)

	if OpenErr != nil {
		return nil, fmt.Errorf("failed to open db: %v", OpenErr)
	}

	ticker := time.NewTicker(1 * time.Second)
	defer ticker.Stop()

	bConnectSuccessfully := false

	for {

		if bConnectSuccessfully {
			break
		}

		<-ticker.C
		fmt.Println("Trying to connect to DB...")
		PingErr := db.Ping()

		if PingErr != nil {
			fmt.Printf("failed to connect to the database: %v \n", PingErr)
		}

		bConnectSuccessfully = PingErr == nil

	}

	fmt.Println("Successfully connected to MariaDB using go-sql-driver/mysql!")

	sqlDB = db

	return db, nil
}

func checkIsTokenValid(w http.ResponseWriter, r *http.Request) (*jwt.Token, error) {

	authHeader := r.Header.Get("Authorization")

	if !strings.HasPrefix(authHeader, "Bearer ") {

		http.Error(w, "Missing token", http.StatusUnauthorized)

		return nil, fmt.Errorf("Missing token %v", http.StatusUnauthorized)
	}

	tokenStr := strings.TrimPrefix(authHeader, "Bearer ")

	token, ParseErr := jwt.Parse(tokenStr, func(token *jwt.Token) (interface{}, error) {
		// Validate signing method
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method")
		}
		return nakamaSecret, nil
	})

	if ParseErr != nil || !token.Valid {
		http.Error(w, "nvalid token", http.StatusUnauthorized)
		return nil, fmt.Errorf("Invalid token %v", http.StatusUnauthorized)
	}

	return token, nil
}

func getUserIDFromToken(token *jwt.Token) (string, error) {

	if token == nil {
		return "", fmt.Errorf("token is nil")
	}
	// Extract claims
	claims := token.Claims.(jwt.MapClaims)
	return claims["uid"].(string), nil
}

func writeResultJSON(result string, status int, w http.ResponseWriter) {

	log.Printf("Status: %d | Result: %s \n", status, result)

	w.Header().Set("Content-Type", "application/json")

	res := map[string]interface{}{
		"status": status,
		"result": result,
	}

	jsonResponse, MarshalErr := json.Marshal(res)

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
	w.Write([]byte(jsonResponse))
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}
