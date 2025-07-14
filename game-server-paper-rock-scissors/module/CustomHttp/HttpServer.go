package CustomHttp

import (
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	wasabi "game-server-paper-rock-scissors/module/Wasabi"
	"net/http"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

// Secret
var (
	jwtServerAccessSecret = []byte("your_secret_key")
	//jwtRefreshSecret = []byte("your-refresh-secret")
	SQLDB *sql.DB
)

// Define a response structure
type Message struct {
	Status  string `json:"status"`
	Message string `json:"message"`
}

type CustomClaims struct {
	Username string `json:"username"`
	jwt.RegisteredClaims
}

// A simple token (usually a long random string) issued to clients to identify and authorize them
// Simple API access, service auth
// Keep key secret, use HTTPS
const API_KEY = "my-secret-api-key"

// A token (usually a JWT or opaque token) that proves the user/client has been authenticated.
// OAuth2, token-based auth
// Tokens expire, revoke possible, use HTTPS
const BearerToken = "your-secret-token"

func apiHandlerAPIKeyAuth(w http.ResponseWriter, r *http.Request) {

	apiKey := r.Header.Get("Authorization")

	//Implement basic authorization using an API key
	if apiKey != API_KEY {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// Set content-type to JSON
	w.Header().Set("Content-Type", "application/json")

	SendData := make(map[string]interface{})
	SendData["name"] = "Jayden"

	JsonMarshalData, _ := json.Marshal(SendData)

	// Example response
	resp := Message{
		Status:  "success",
		Message: string(JsonMarshalData),
	}

	// Convert to JSON and write to response
	json.NewEncoder(w).Encode(resp)
}

func apiHandlerBearerAuth(w http.ResponseWriter, r *http.Request) {

	ctx := r.Context()
	select {
	case <-ctx.Done():
		fmt.Println("Request cancelled by client:", ctx.Err())
		return
	default:
	}

	// Extract token from Authorization header
	authHeader := r.Header.Get("Authorization")
	const prefix = "Bearer "

	if !strings.HasPrefix(authHeader, prefix) {
		http.Error(w, "Unauthorized - missing bearer prefix", http.StatusUnauthorized)
		return
	}

	token := strings.TrimPrefix(authHeader, prefix)

	// Compare with your predefined token

	claims, ValidateJWTErr := validateJWT(token)

	if ValidateJWTErr != nil {
		fmt.Println("Unauthorized - invalid token: ", ValidateJWTErr)

		// Example response
		resp := Message{
			Status:  "failed",
			Message: "Unauthorized - invalid token",
		}

		// Convert to JSON and write to response
		json.NewEncoder(w).Encode(resp)
		return
	}

	// Set content-type to JSON
	w.Header().Set("Content-Type", "application/json")

	SendData := make(map[string]interface{})
	SendData["name"] = claims.Username

	JsonMarshalData, _ := json.Marshal(SendData)

	// Example response
	resp := Message{
		Status:  "success",
		Message: string(JsonMarshalData),
	}

	// Convert to JSON and write to response
	json.NewEncoder(w).Encode(resp)
}

func apiAuthenticateUser(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	select {
	case <-ctx.Done():
		fmt.Println("Request cancelled by client:", ctx.Err())
		return
	default:
	}

	// Decode JSON body into a dynamic struct
	var auth map[string]interface{}
	if err := json.NewDecoder(r.Body).Decode(&auth); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	username, bHasUsername := auth["username"]
	password, bHasPassword := auth["password"]

	if !bHasUsername {
		fmt.Fprint(w, "username key doesn't exist in the map.")
	}

	if !bHasPassword {
		fmt.Fprint(w, "password key doesn't exist in the map.")
	}

	// Example: Check credentials (replace with real validation)
	if username != "Jayden" || password != "123456" {
		http.Error(w, "Unauthorized - invalid credentials", http.StatusUnauthorized)
		return
	}

	// Set content-type to JSON
	w.Header().Set("Content-Type", "application/json")

	Token, JWTErr := generateJWT(username.(string))

	if JWTErr != nil {
		fmt.Println("Failed to generate JWT token: ", JWTErr)
		return
	}

	resp := Message{
		Status:  "success",
		Message: Token,
	}

	json.NewEncoder(w).Encode(resp)
}

func apiUserRegistration(w http.ResponseWriter, r *http.Request) {

	ctx := r.Context()
	select {
	case <-ctx.Done():
		fmt.Println("Request cancelled by client:", ctx.Err())
		return
	default:
	}

	// Set content-type to JSON
	w.Header().Set("Content-Type", "application/json")

	var userdata map[string]interface{}
	if err := json.NewDecoder(r.Body).Decode(&userdata); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	fmt.Printf("trying to register user: %v | %v | %v\n\n", userdata["user_id"], userdata["name"], userdata["email"])

	//Save data into the data base
	if SQLDB == nil {
		http.Error(w, "Invalid database", http.StatusBadRequest)
		fmt.Print("Invalid database!!!")
		return
	}

	// Example insert statement
	insertQuery := "INSERT INTO users (user_id, name, email) VALUES (?, ?, ?)"

	// Execute the insert query with values
	_, err := SQLDB.Exec(insertQuery, userdata["user_id"], userdata["name"], userdata["email"])

	if err != nil {
		fmt.Printf("Insert failed: %v \n", err)
		return
	}

	resp := Message{
		Status:  "success",
		Message: "",
	}

	json.NewEncoder(w).Encode(resp)

}

func StartHosting() {

	fmt.Println("Start hosting......")

	_, OpenDBErr := openDB()

	defer SQLDB.Close()

	if OpenDBErr != nil {
		fmt.Println(OpenDBErr.Error())
	}

	wasabi.LoadConfig()

	//Initialize API
	http.HandleFunc("/apiHandlerAPIKeyAuth", apiHandlerAPIKeyAuth) //http://localhost:8090/api https://fucherng1375.ddns.net:8090/ApiHandlerAPIKeyAuth

	http.HandleFunc("/apiHandlerBearerAuth", apiHandlerBearerAuth)

	http.HandleFunc("/apiAuthenticateUser", apiAuthenticateUser)

	http.HandleFunc("/apiUserRegistration", apiUserRegistration)

	// Use HTTPS with your cert and key
	fmt.Println("HTTPS server running on port 8090")
	err := http.ListenAndServeTLS("0.0.0.0:8090", "/Cert/fucherng1375.ddns.net-crt.pem", "/Cert/fucherng1375.ddns.net-key-decrypted.pem", nil)
	if err != nil {
		panic(err)
	}
}

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

	SQLDB = db

	return db, nil
}

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
