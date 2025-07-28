package httpserverroutes

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

const dbTableUsers = "users"

func apiRegisterUser(w http.ResponseWriter, r *http.Request, db *sql.DB) {

	var body map[string]interface{}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		writeResultJSON(w, fmt.Sprintf("Invalid JSON %v", err), http.StatusBadRequest)
		return
	}

	username, bUsername := body["username"].(string)
	password, bPassword := body["password"].(string)

	if !bUsername {
		writeResultJSON(w, "username not found in request body during API call.", http.StatusBadRequest)
		return
	}

	if !bPassword {
		writeResultJSON(w, "password not found in request body during API call.", http.StatusBadRequest)
		return
	}

	tx, dbErr := db.Begin()

	if dbErr != nil {
		writeResultJSON(w, fmt.Sprintf("failed to start a transaction. %s", dbErr), http.StatusBadRequest)
		return
	}

	checkUserQuery := fmt.Sprintf(`SELECT COUNT(*) FROM %s WHERE username = ?`, dbTableUsers)

	var rowCount int
	scanErr := tx.QueryRow(checkUserQuery, username).Scan(&rowCount)

	if scanErr != nil {
		tx.Rollback()
		writeResultJSON(w, fmt.Sprintf("failed to scan. %s", scanErr), http.StatusInternalServerError)
		return
	}

	log.Printf("row count = %d\n", rowCount)

	if rowCount > 0 {
		tx.Rollback()
		writeResultJSON(w, "Failed to register user. An account with this username already exists.", http.StatusBadRequest)
		return
	}

	userID := uuid.New().String()

	// Generate hashed password (bcrypt automatically salts)
	hashedPassword, hashedErr := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if hashedErr != nil {
		writeResultJSON(w, fmt.Sprintf("Failed to hash password. %s", hashedErr), http.StatusInternalServerError)
		return
	}

	addUserQuery := fmt.Sprintf(`INSERT INTO %s (user_id, username, hashed_password) VALUES (?, ?, ?)`, dbTableUsers)

	_, insertErr := tx.Exec(addUserQuery, userID, username, hashedPassword)

	if insertErr != nil {
		tx.Rollback()
		writeResultJSON(w, fmt.Sprintf("Failed to add user. %s", insertErr), http.StatusInternalServerError)
		return
	}

	commitErr := tx.Commit()

	if commitErr != nil {
		writeResultJSON(w, fmt.Sprintf("Failed to commit. %s", commitErr), http.StatusInternalServerError)
		return
	}

	writeResultJSON(w, "user registered susscessfully", http.StatusOK)
}

func apiUserAuthentication(w http.ResponseWriter, r *http.Request, db *sql.DB) {

	var body map[string]interface{}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		writeResultJSON(w, fmt.Sprintf("Invalid JSON %v", err), http.StatusBadRequest)
		return
	}

	username, bUsername := body["username"].(string)
	password, bPassword := body["password"].(string)

	if !bUsername {
		writeResultJSON(w, "username not found in request body during API call.", http.StatusBadRequest)
		return
	}

	if !bPassword {
		writeResultJSON(w, "password not found in request body during API call.", http.StatusBadRequest)
		return
	}

	tx, dbErr := db.Begin()
	if dbErr != nil {
		writeResultJSON(w, fmt.Sprintf("failed to start a transaction. %v", dbErr), http.StatusBadRequest)
		return
	}

	getUserPassQuery := fmt.Sprintf(`SELECT user_id, hashed_password FROM %s WHERE username = ?`, dbTableUsers)

	var dbUserID string
	var dbhashedPassword string
	scanPassErr := tx.QueryRow(getUserPassQuery, username).Scan(&dbUserID, &dbhashedPassword)

	if scanPassErr != nil {
		writeResultJSON(w, fmt.Sprintf("failed to get user's ID and password. %v", scanPassErr), http.StatusBadRequest)
		return
	}

	if dbhashedPassword == "" {
		writeResultJSON(w, "hashed Password is empty.", http.StatusInternalServerError)
		return
	}

	if !checkPassword(dbhashedPassword, password) {
		writeResultJSON(w, "❌ incorrect password", http.StatusBadRequest)
		return
	}

	strAccessToken, jwtAccessErr := generateJWT(username, dbUserID, "player", time.Hour*1, "access", jwtServerAccessSecret)
	strRefreshToken, jwtRefreshErr := generateJWT(username, dbUserID, "player", time.Hour*168, "refresh", jwtServerRefreshSecret) //7 days

	if jwtAccessErr != nil {
		writeResultJSON(w, fmt.Sprintf("❌ failed to generate JWT access token. %v", jwtAccessErr), http.StatusInternalServerError)
		return
	}

	if jwtRefreshErr != nil {
		writeResultJSON(w, fmt.Sprintf("❌ failed to generate JWT refresh token. %v", jwtRefreshErr), http.StatusInternalServerError)
		return
	}

	var response Response_UserAuth
	response.AccessToken = strAccessToken
	response.RefreshToken = strRefreshToken
	response.UserID = dbUserID
	response.Username = username

	marshalRes, marshalResErr := json.Marshal(response)

	if marshalResErr != nil {
		writeResultJSON(w, fmt.Sprintf("❌ failed to marshal response%v", marshalResErr), http.StatusInternalServerError)
		return
	}

	writeResultJSON(w, string(marshalRes), http.StatusOK)
}

func apiTestAuthentication(w http.ResponseWriter, r *http.Request) {

	strToken, getTokenErr := getTokenFromHeader(w, r)

	if getTokenErr != nil {
		writeResultJSON(w, fmt.Sprintf("Failed to fetch token from header %v", getTokenErr), http.StatusUnauthorized)
		return
	}

	_, validateJWTErr := validateJWT(strToken, jwtServerAccessSecret)

	if validateJWTErr != nil {
		writeResultJSON(w, fmt.Sprintf("Failed to validate token %v", validateJWTErr), http.StatusUnauthorized)
		return
	}

	jwtClaim, jwtClaimErr := getClaimsFromToken(strToken, jwtServerAccessSecret)

	if jwtClaimErr != nil {
		writeResultJSON(w, fmt.Sprintf("Failed to retrieve data from token %v", jwtClaimErr), http.StatusInternalServerError)
		return
	}

	jsonRes, marshalResErr := json.Marshal(jwtClaim)

	if marshalResErr != nil {
		writeResultJSON(w, fmt.Sprintf("Failed to marshal token claims %v", marshalResErr), http.StatusInternalServerError)
		return
	}

	writeResultJSON(w, string(jsonRes), http.StatusOK)
}

func apiRefreshToken(w http.ResponseWriter, r *http.Request) {

	strRefreshToken, tokenErr := getTokenFromHeader(w, r)

	if tokenErr != nil {
		writeResultJSON(w, fmt.Sprintf("1. Couldn't get the refresh token from header %v", tokenErr), http.StatusUnauthorized)
		return
	}

	_, validateErr := validateJWT(strRefreshToken, jwtServerRefreshSecret)

	if validateErr != nil {
		writeResultJSON(w, fmt.Sprintf("2. Refresh token: %v", validateErr), http.StatusUnauthorized)
		return
	}

	jwtClaim, getClaimErr := getClaimsFromToken(strRefreshToken, jwtServerRefreshSecret)

	if getClaimErr != nil {
		writeResultJSON(w, fmt.Sprintf("3. Refresh token: %v", getClaimErr), http.StatusUnauthorized)
		return
	}

	if jwtClaim.AccessType != "refresh" {
		writeResultJSON(w, "4. Failed to refresh token. The provided token is not a valid refresh token.", http.StatusUnauthorized)
		return
	}

	strAccessToken, accessTokenErr := RefreshAccessToken(strRefreshToken)

	if accessTokenErr != nil {
		writeResultJSON(w, fmt.Sprintf("5. Refresh token: %v", accessTokenErr), http.StatusUnauthorized)
		return
	}

	var response Response_UserAuth
	response.AccessToken = strAccessToken
	response.RefreshToken = strRefreshToken
	response.UserID = jwtClaim.UserID
	response.Username = jwtClaim.Username

	marshalRes, marshalResErr := json.Marshal(response)

	if marshalResErr != nil {
		writeResultJSON(w, fmt.Sprintf("6. ❌ failed to marshal response%v", marshalResErr), http.StatusInternalServerError)
		return
	}

	writeResultJSON(w, string(marshalRes), http.StatusOK)

}

func checkPassword(hashedPassword string, plainPassword string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(plainPassword))
	return err == nil
}
