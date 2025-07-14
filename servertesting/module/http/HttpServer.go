package http

import (
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	wasabi "servertesting/module/Wasabi"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
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

func apiAddUser(w http.ResponseWriter, r *http.Request) {
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

	_, ValidateJWTErr := validateJWT(token)

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

	//Authenticate successfully

	// Set content-type to JSON
	w.Header().Set("Content-Type", "application/json")

	var userdata map[string]interface{}
	if err := json.NewDecoder(r.Body).Decode(&userdata); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	fmt.Printf("trying to add user: %v | %v\n\n", userdata["name"], userdata["email"])

	//Save data into the data base
	if SQLDB == nil {
		http.Error(w, "Invalid database", http.StatusBadRequest)
		fmt.Print("Invalid database!!!")
		return
	}

	// Example insert statement
	insertQuery := "INSERT INTO users (name, email) VALUES (?, ?)"

	// Execute the insert query with values
	_, err := SQLDB.Exec(insertQuery, userdata["name"], userdata["email"])

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

func apiSaveUserAvatar(w http.ResponseWriter, r *http.Request) {
	// Parse the form
	err := r.ParseMultipartForm(10 << 20) // 10 MB max memory
	if err != nil {
		http.Error(w, "Error parsing form", http.StatusBadRequest)
		return
	}

	// Get form values
	name := r.FormValue("name")
	age := r.FormValue("age")

	// Get uploaded file
	file, avatar, err := r.FormFile("avatar") // 'image' is the name of the form field
	if err != nil {
		http.Error(w, "Error retrieving the avatar", http.StatusBadRequest)
		return
	}
	defer file.Close()

	fmt.Fprintf(w, "Trying to save user avatar -> name: %s, age: %s, avatar: %s\n\n", name, age, avatar.Filename)

	//Save data into the data base
	if SQLDB == nil {
		http.Error(w, "Invalid database", http.StatusBadRequest)
		fmt.Print("Invalid database!!!")
		return
	}

	AvatarUuid := uuid.New()

	// Example insert statement
	insertQuery := `
  	INSERT INTO useravatars (name, age, avatar)
  	VALUES (?, ?, ?)
  	ON DUPLICATE KEY UPDATE
    age = VALUES(age),
    avatar = VALUES(avatar);`

	// Execute the insert query with values
	//_, ExecErr := SQLDB.Exec(insertQuery, name, age, avatar.Filename)
	_, ExecErr := SQLDB.Exec(insertQuery, name, age, AvatarUuid)

	if ExecErr != nil {
		fmt.Fprintf(w, "Insert failed: %v \n", ExecErr)
		return
	}

	//Save to wasabi
	UploadFileErr := wasabi.UploadFile(file, avatar, AvatarUuid.String())

	if UploadFileErr != nil {
		fmt.Fprintf(w, "Error: %v", UploadFileErr)
		return
	}

	fmt.Fprintf(w, "Successfully uploaded file to wasabi\n\n")

	// Save to disk (optional)
	dst, err := os.Create("./app/Images/" + avatar.Filename)
	if err != nil {
		http.Error(w, fmt.Sprintf("Unable to save the file: %v", err), http.StatusInternalServerError)
		return
	}
	defer dst.Close()
	io.Copy(dst, file)

	fmt.Fprintf(w, "Successfully saved file to disk\n\n")

	resp := Message{
		Status:  "success",
		Message: "",
	}

	json.NewEncoder(w).Encode(resp)
}

func apiRemoveUserAvatar(w http.ResponseWriter, r *http.Request) {

	var data map[string]interface{}
	if err := json.NewDecoder(r.Body).Decode(&data); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	// Now you can access values from data, e.g.
	name, ok := data["name"].(string)

	if !ok {
		fmt.Fprintf(w, "Error: Couldn't find the field called 'name'")
		return
	}

	query := "SELECT avatar FROM useravatars WHERE name = ?"

	var avatar string

	ScanErr := SQLDB.QueryRow(query, name).Scan(&avatar)

	if ScanErr != nil {
		fmt.Fprintf(w, "Failed to retrieve user avatar.\n")
		return
	}

	DeleteErr := wasabi.DeleteFile(avatar)

	if DeleteErr != nil {
		fmt.Fprintf(w, "Wasabi: Failed to delete user avatar.\n")
	}

	fmt.Fprintf(w, "Wasabi: Successfully deleted user avatar.\n")

	DeleteQuery := "DELETE FROM useravatars WHERE name = ?"
	_, ExecErr := SQLDB.Exec(DeleteQuery, name)

	if ExecErr != nil {
		fmt.Fprintf(w, "Failed to delete user avatar from useravatars: %v\n", ExecErr)
	}

	fmt.Fprintf(w, "Successfully delete user avatar from useravatars.\n")
}

func apiDownloadUserAvatar(w http.ResponseWriter, r *http.Request) {

	var data map[string]interface{}
	if err := json.NewDecoder(r.Body).Decode(&data); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	// Now you can access values from data, e.g.
	name, ok := data["name"].(string)

	if !ok {
		fmt.Fprintf(w, "Error: Couldn't find the field called 'name'")
		return
	}

	query := "SELECT avatar FROM useravatars WHERE name = ?"

	var avatar string

	ScanErr := SQLDB.QueryRow(query, name).Scan(&avatar)

	if ScanErr != nil {
		fmt.Fprintf(w, "Failed to retrieve user avatar.\n")
		return
	}

	DownloadPath := "./app/Images/wasabi/"

	DownloadErr := wasabi.DownloadImage(name+"_"+avatar, DownloadPath)

	if DownloadErr != nil {
		fmt.Fprintf(w, "Wasabi Error: %v", DownloadErr)
	}

	fmt.Fprintf(w, "Successfully downloaded image to %s\n\n", DownloadPath+name+"_"+avatar)
}

func apiGetUserAvatar(w http.ResponseWriter, r *http.Request) {
	var data map[string]interface{}
	if err := json.NewDecoder(r.Body).Decode(&data); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	// Now you can access values from data, e.g.
	name, ok := data["name"].(string)

	if !ok {
		fmt.Fprintf(w, "Error: Couldn't find the field called 'name'")
		return
	}

	query := "SELECT name, avatar FROM useravatars WHERE name = ?"

	var uname string
	var avatar string

	ScanErr := SQLDB.QueryRow(query, name).Scan(&uname, &avatar)

	if ScanErr != nil {
		fmt.Fprintf(w, "Failed to retrieve user avatar.\n")
		return
	}

	fmt.Fprintf(w, "user name: %v\n\n", uname)
	fmt.Fprintf(w, "avatar name: %v\n\n", avatar)

	resp := Message{
		Status:  "success",
		Message: "./app/Images/" + avatar,
	}

	json.NewEncoder(w).Encode(resp)
}

func apiGetImagePath(w http.ResponseWriter, r *http.Request) {

	// Retrieve the 'filename' query parameter from the URL
	filename := r.URL.Query().Get("filename")

	if filename == "" {
		http.Error(w, "Missing 'filename' query parameter", http.StatusBadRequest)
		return
	}

	// Construct the file path
	filePath := "./app/Images/" + filename

	data, err := os.ReadFile(filePath)
	if err != nil {
		fmt.Fprintf(w, "Error reading file: %v", err)
		return
	}

	fmt.Fprintf(w, "Read %d bytes from %s\n", len(data), filePath)
}

func StartHosting() {

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

	http.HandleFunc("/apiAddUser", apiAddUser)

	http.HandleFunc("/apiSaveUserAvatar", apiSaveUserAvatar)

	http.HandleFunc("/apiGetImagePath", apiGetImagePath)

	http.HandleFunc("/apiGetUserAvatar", apiGetUserAvatar)

	http.HandleFunc("/apiRemoveUserAvatar", apiRemoveUserAvatar)

	http.HandleFunc("/apiDownloadUserAvatar", apiDownloadUserAvatar)

	// Use HTTPS with your cert and key
	fmt.Println("HTTPS server running on port 8090")
	err := http.ListenAndServeTLS("0.0.0.0:8090", "./Cert/fucherng1375.ddns.net-crt.pem", "./Cert/fucherng1375.ddns.net-key-decrypted.pem", nil)
	if err != nil {
		panic(err)
	}
}

func openDB() (*sql.DB, error) {

	dsn := "myuser:mypassword@tcp(db:3306)/mydb"

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
