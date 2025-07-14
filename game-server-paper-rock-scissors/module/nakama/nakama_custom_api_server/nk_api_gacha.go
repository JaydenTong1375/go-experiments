package nakamacustomapiserver

import (
	"database/sql"
	"encoding/json"
	"fmt"
	rd "game-server-paper-rock-scissors/module/redis"
	"log"
	"net/http"
	"strings"

	"github.com/google/uuid"
)

const dbTableUsersCredit = "transaction_users_credit"
const dbTableGachaResult = "transaction_gacha_result"

func apiSpinForReward(w http.ResponseWriter, r *http.Request) {

	token, tokenErr := checkIsTokenValid(w, r)

	if tokenErr != nil {
		writeResultJSON("❌Invalid token", http.StatusUnauthorized, w)
		return
	}

	//Get payload from body
	var payload map[string]interface{}

	DecodeErr := json.NewDecoder(r.Body).Decode(&payload)
	if DecodeErr != nil {
		writeResultJSON("Invalid JSON body", http.StatusBadRequest, w)
		return
	}

	userID, userIDErr := getUserIDFromToken(token)

	if userIDErr != nil {
		writeResultJSON("failed to get user id from the token", http.StatusNotFound, w)
		return
	}

	//Use a Redis lock to prevent users from spamming this API
	lockErr := rd.RedisLock("lock:spin", userID)
	if lockErr != nil {
		writeResultJSON(fmt.Sprintf("Request already in progress. %v", lockErr), http.StatusTooManyRequests, w)
		return
	}

	if sqlDB == nil {
		writeResultJSON("❌ sqlDB is null -> database connection not initialized.", http.StatusInternalServerError, w)
		return
	}

	transactionID := uuid.New().String()

	tx, txErr := sqlDB.Begin()

	if txErr != nil {
		writeResultJSON("❌ Failed to start a transaction.", http.StatusInternalServerError, w)
		return
	}

	defer tx.Rollback()

	Response := map[string]interface{}{
		"inventory": "",
		"credit":    0,
	}

	//Check balance first
	var currentBalance int

	checkBalanceQuery := fmt.Sprintf(`
	SELECT SUM(CASE WHEN transaction_type = 'CREDIT' THEN amount ELSE - amount END)
	FROM %s
	WHERE user_id = ?
	`, dbTableUsersCredit)

	scanBalanceErr := tx.QueryRow(checkBalanceQuery, userID).Scan(&currentBalance)

	if scanBalanceErr != nil {
		writeResultJSON(fmt.Sprintf("Failed to check user credit. %v", scanBalanceErr), http.StatusInternalServerError, w)
		return
	}

	if currentBalance <= 0 {
		writeResultJSON("Insufficient user credit.", http.StatusPaymentRequired, w)
		return
	}

	//Update balance
	updateUserCreditQuery := fmt.Sprintf(`INSERT INTO %s (user_id, transaction_id, amount, transaction_type) VALUES (?, ?, ?, ?)`, dbTableUsersCredit)

	_, updateCreditErr := tx.Exec(updateUserCreditQuery, userID, transactionID, 1, "DEBIT")

	if updateCreditErr != nil {
		writeResultJSON(fmt.Sprintf("Failed to update user credit. %v", updateCreditErr), http.StatusInternalServerError, w)
		return
	}

	Response["credit"] = currentBalance - 1

	//Add item into the inventory
	//Get items from the payload
	payloadItems, bIsItemsValid := payload["items"].(string)

	payloadItems = strings.ReplaceAll(payloadItems, "'", "\"")

	var items []map[string]interface{}
	itemsUnmarshalErr := json.Unmarshal([]byte(payloadItems), &items)

	if itemsUnmarshalErr != nil {
		writeResultJSON(fmt.Sprintf("Failed to unmarshal payload items: %v\n", itemsUnmarshalErr), http.StatusBadRequest, w)
		return
	}

	if bIsItemsValid == false {
		writeResultJSON(fmt.Sprintf("Invalid format: expected 'items' field to be an array of strings: %v\n", items), http.StatusBadRequest, w)
		return
	}

	for _, item := range items {

		itemID, bIsItemIDValid := item["item_id"].(string)
		rarity, bIsRarityValid := item["rarity"].(string)
		qtyFloat, bIsQtyValid := item["quantity"].(float64)

		qty := int(qtyFloat)

		if bIsItemIDValid == false {
			log.Println("Missing 'item_id' feild")
			continue
		}

		if bIsRarityValid == false {
			log.Println("Missing 'rarity' feild")
			continue
		}

		if bIsQtyValid == false {
			log.Println("Missing 'quantity' feild")
			continue
		}

		//Sava gacha result
		insertNewGachaQuery := fmt.Sprintf(`INSERT INTO %s (user_id, item_id, transaction_id, quantity, rarity) VALUES (?, ?, ?, ?, ?)`, dbTableGachaResult)

		_, insertGachaErr := tx.Exec(insertNewGachaQuery, userID, itemID, transactionID, 1, rarity)

		if insertGachaErr != nil {
			writeResultJSON(fmt.Sprintf("Failed to insert gacha result into %s table", dbTableUsersCredit), http.StatusInternalServerError, w)
			continue
		}

		//Save to user's inventory
		InsertItemQuery := fmt.Sprintf(`INSERT INTO %s (user_id, item_id, quantity, transaction_type, reason) VALUES (?, ?, ?, ?, ?)`, dbTableUsersInventory)

		_, insertItemErr := tx.Exec(InsertItemQuery, userID, itemID, qty, "CREDIT", "Obtained from Gacha")

		if insertItemErr != nil {
			log.Printf("Failed to add %s item into inventory. %v\n", itemID, insertItemErr)
			continue
		}

	}

	//Get inventory items
	getInventoryQuery := fmt.Sprintf(`
	SELECT item_id, SUM(CASE 
	WHEN transaction_type = 'CREDIT' THEN quantity 
	WHEN transaction_type = 'DEBIT' THEN - quantity 
	ELSE 0 END) as item_quantity
	FROM %s
	WHERE user_id = ?
	AND created_at > (
		SELECT COALESCE(MAX(created_at), '1970-01-01 00:00:00')
		FROM %s
		WHERE user_id = ? 
		AND transaction_type = 'WIPE'
	)
	GROUP BY item_id
	`, dbTableUsersInventory, dbTableUsersInventory)

	rows, queryErr := tx.Query(getInventoryQuery, userID, userID)

	if queryErr != nil {
		writeResultJSON(fmt.Sprintf("❌ Scan error: %v\n", queryErr), http.StatusBadRequest, w)
		return
	}

	defer rows.Close()

	type InventoryItem struct {
		ItemID   string
		Quantity int
	}

	var inventory []InventoryItem

	for rows.Next() {
		var item InventoryItem
		scanItemErr := rows.Scan(&item.ItemID, &item.Quantity)

		if scanItemErr != nil {
			continue
		}

		inventory = append(inventory, item)
	}

	JSONinventory, marshalErr := json.Marshal(inventory)

	if marshalErr != nil {
		writeResultJSON(fmt.Sprintf("Failed to marshal inventory. %v", marshalErr), http.StatusInternalServerError, w)
		return
	}

	Response["inventory"] = string(JSONinventory)

	//Commit
	commitErr := tx.Commit()
	if commitErr != nil {
		writeResultJSON(fmt.Sprintf("Failed to commit. %v", commitErr), http.StatusInternalServerError, w)
		return
	}

	jsonRes, marshalErr := json.Marshal(Response)

	if marshalErr != nil {
		writeResultJSON(fmt.Sprintf("Failed to marshal response. %v", marshalErr), http.StatusBadRequest, w)
		return
	}

	writeResultJSON(string(jsonRes), http.StatusOK, w)

}

func apiUpdateUserCredit(w http.ResponseWriter, r *http.Request) {

	token, tokenErr := checkIsTokenValid(w, r)
	if tokenErr != nil {
		writeResultJSON(fmt.Sprintf("%v", tokenErr), http.StatusUnauthorized, w)
		return
	}

	//Get user id
	userID, userIDErr := getUserIDFromToken(token)

	if userIDErr != nil {
		writeResultJSON(fmt.Sprintf("%v", userIDErr), http.StatusNotFound, w)
		return
	}

	if sqlDB == nil {
		writeResultJSON("❌ sqlDB is null -> database connection not initialized.", http.StatusInternalServerError, w)
		return
	}

	//Get payload from body
	var payload map[string]interface{}

	DecodeErr := json.NewDecoder(r.Body).Decode(&payload)
	if DecodeErr != nil {
		writeResultJSON("Invalid JSON body", http.StatusBadRequest, w)
		return
	}

	status, bIsStatusValid := payload["status"]

	if bIsStatusValid == false {
		writeResultJSON("Missing 'status' field in payload", http.StatusBadRequest, w)
		return
	}

	value, bIsValueValid := payload["value"]

	if bIsValueValid == false {
		writeResultJSON("Missing 'value' field in payload", http.StatusBadRequest, w)
		return
	}

	transactionID := uuid.New().String()

	//Insert a new row into the table
	insertNewCreditQuery := fmt.Sprintf(`INSERT INTO %s (user_id, transaction_id, amount, transaction_type) VALUES (?, ?, ?, ?)`, dbTableUsersCredit)

	_, insertUserCreditErr := sqlDB.Exec(insertNewCreditQuery, userID, transactionID, value, status)

	if insertUserCreditErr != nil {
		writeResultJSON(fmt.Sprintf("Failed to insert user credit into %s table", dbTableUsersCredit), http.StatusInternalServerError, w)
		return
	}

	writeResultJSON("Successfully insert user credit", http.StatusOK, w)
}

func apiGetUserCredit(w http.ResponseWriter, r *http.Request) {

	token, tokenErr := checkIsTokenValid(w, r)
	if tokenErr != nil {
		writeResultJSON(fmt.Sprintf("%v", tokenErr), http.StatusUnauthorized, w)
		return
	}

	if sqlDB == nil {
		writeResultJSON("❌ sqlDB is null -> database connection not initialized.", http.StatusInternalServerError, w)
		return
	}

	//Get user id
	userID, userIDErr := getUserIDFromToken(token)

	if userIDErr != nil {
		writeResultJSON(fmt.Sprintf("%v", userIDErr), http.StatusNotFound, w)
		return
	}

	getUserCreditQuery := fmt.Sprintf(`
	SELECT 
	SUM(CASE 
	WHEN transaction_type = 'CREDIT' AND deleted_at IS NULL THEN amount 
	WHEN transaction_type = 'DEBIT' AND deleted_at IS NULL THEN -amount
	ELSE 0 END) as current_balance
	FROM %s 
	WHERE user_id = ?
	`, dbTableUsersCredit)

	var jsonUserCredit string
	scanErr := sqlDB.QueryRow(getUserCreditQuery, userID).Scan(&jsonUserCredit)

	if scanErr != nil && scanErr != sql.ErrNoRows {
		writeResultJSON(fmt.Sprintf("Scan error: %v", scanErr), http.StatusBadRequest, w)
		return
	}

	writeResultJSON(jsonUserCredit, http.StatusOK, w)
}
