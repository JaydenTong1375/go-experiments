package nakamacustomapiserver

import (
	"database/sql"
	"encoding/json"
	"fmt"
	nkAPIClient "game-server-paper-rock-scissors/module/nakama/nakama_custom_api_clients"
	"log"
	"net/http"
	"strings"

	"github.com/golang-jwt/jwt"
	"github.com/heroiclabs/nakama-common/runtime"
)

// The Nakama server secret is defined in local.yml
var nakamaSecret = []byte("your_super_secret_server_key")

var (
	jwtServerAccessSecret = []byte("your_secret_key")
	sqlDB                 *sql.DB
)

const dbTableUsers = "users"

func RegisterCustomAPI(initializer runtime.Initializer) error {

	openDB()

	if err := initializer.RegisterHttp("/apiTestAuthenticate", apiTestAuthenticate); err != nil {
		return err
	}

	if err := initializer.RegisterHttp("/apiRegisterUser", apiRegisterUser); err != nil {
		return err
	}

	if err := initializer.RegisterHttp("/apiGetInventoryItems", apiGetInventoryItems); err != nil {
		return err
	}

	if err := initializer.RegisterHttp("/apiUpdateUserCredit", apiUpdateUserCredit); err != nil {
		return err
	}

	if err := initializer.RegisterHttp("/apiGetUserCredit", apiGetUserCredit); err != nil {
		return err
	}

	if err := initializer.RegisterHttp("/apiSpinForReward", apiSpinForReward); err != nil {
		return err
	}

	if err := initializer.RegisterHttp("/apiClearInventory", apiClearInventory); err != nil {
		return err
	}

	if err := initializer.RegisterHttp("/apiSaveUserQuest", apiSaveUserQuest); err != nil {
		return err
	}

	if err := initializer.RegisterHttp("/apiGetUserQuest", apiGetUserQuest); err != nil {
		return err
	}

	if err := initializer.RegisterHttp("/apiUpdateUserQuest", apiUpdateUserQuest); err != nil {
		return err
	}

	return nil
}

//////////////////////////////////////// Rest API ///////////////////////////////////////////

func apiRegisterUser(w http.ResponseWriter, r *http.Request) {

	authHeader := r.Header.Get("Authorization")
	if !strings.HasPrefix(authHeader, "Bearer ") {
		http.Error(w, "Missing token", http.StatusUnauthorized)
		return
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
		http.Error(w, "Invalid token", http.StatusUnauthorized)
		return
	}

	//Get payload from body
	var payload map[string]interface{}

	DecodeErr := json.NewDecoder(r.Body).Decode(&payload)
	if DecodeErr != nil {
		http.Error(w, "Invalid JSON body", http.StatusBadRequest)
		return
	}

	// Access data using type assertions
	username, _ := payload["username"].(string)
	email, _ := payload["email"].(string)

	// Extract claims
	claims := token.Claims.(jwt.MapClaims)
	userID := claims["uid"].(string)

	if sqlDB == nil {
		log.Printf("❌ sqlDB is null -> database connection not initialized")
		return
	}

	const registerUserQuery = "INSERT INTO " + dbTableUsers + " (user_id, username, email) VALUES (?, ?, ?)"

	_, queryErr := sqlDB.Query(registerUserQuery, userID, username, email)

	if queryErr != nil {
		log.Printf("❌ Failed to query: %v", queryErr)
		return
	}

	log.Printf("✅ Successfully registered userid: %s | username: %s | email: %s", userID, username, email)

	//Initialize credit
	_, updateCreditErr := nkAPIClient.CallUpdateUserCreditAPI(tokenStr, "CREDIT", 12)

	if updateCreditErr != nil {
		writeResultJSON(fmt.Sprintf("Failed to initialize user's credit. %v", updateCreditErr), http.StatusInternalServerError, w)
		return
	}
}

func apiTestAuthenticate(w http.ResponseWriter, r *http.Request) {

	authHeader := r.Header.Get("Authorization")
	if !strings.HasPrefix(authHeader, "Bearer ") {
		http.Error(w, "Missing token", http.StatusUnauthorized)
		return
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
		http.Error(w, "Invalid token", http.StatusUnauthorized)
		return
	}

}
