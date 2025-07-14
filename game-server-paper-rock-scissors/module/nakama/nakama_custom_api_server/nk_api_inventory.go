package nakamacustomapiserver

import (
	"encoding/json"
	"fmt"
	"net/http"
)

const dbTableUsersInventory = "transaction_users_inventory"

func apiGetInventoryItems(w http.ResponseWriter, r *http.Request) {

	token, tokenErr := checkIsTokenValid(w, r)

	if tokenErr != nil {
		writeResultJSON(fmt.Sprintf("%v", tokenErr), http.StatusUnauthorized, w)
	}

	userid, userIDErr := getUserIDFromToken(token)

	if userIDErr != nil {
		writeResultJSON(fmt.Sprintf("%v", userIDErr), http.StatusNotFound, w)
		return
	}

	if sqlDB == nil {
		writeResultJSON("❌ sqlDB is null -> database connection not initialized.", http.StatusInternalServerError, w)
		return
	}

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

	rows, queryErr := sqlDB.Query(getInventoryQuery, userid, userid)

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

	writeResultJSON(string(JSONinventory), http.StatusOK, w)
}

func apiClearInventory(w http.ResponseWriter, r *http.Request) {

	token, tokenErr := checkIsTokenValid(w, r)
	if tokenErr != nil {
		writeResultJSON(fmt.Sprintf("%v", tokenErr), http.StatusUnauthorized, w)
		return
	}

	if sqlDB == nil {
		writeResultJSON("❌ sqlDB is null -> database connection not initialized.", http.StatusInternalServerError, w)
		return
	}

	tx, txErr := sqlDB.Begin()

	if txErr != nil {
		writeResultJSON("❌ Failed to start a transaction.", http.StatusInternalServerError, w)
		return
	}

	defer tx.Rollback()

	//Get user id
	userID, userIDErr := getUserIDFromToken(token)

	if userIDErr != nil {
		writeResultJSON(fmt.Sprintf("%v", userIDErr), http.StatusNotFound, w)
		return
	}

	InsertWipeQuery := fmt.Sprintf(`INSERT INTO %s (user_id, transaction_type, reason) VALUES(?, ?, ?)
	`, dbTableUsersInventory)

	tx.Exec(InsertWipeQuery, userID, "WIPE", "Clear inventory")

	//Commit
	commitErr := tx.Commit()
	if commitErr != nil {
		writeResultJSON(fmt.Sprintf("Failed to commit. %v", commitErr), http.StatusInternalServerError, w)
		return
	}

	writeResultJSON("Successfully cleared user's inventory", http.StatusOK, w)
}
