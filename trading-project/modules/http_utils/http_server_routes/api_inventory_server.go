package httpserverroutes

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	"github.com/google/uuid"
)

const dbTableUsersInventory = "users_inventory_transactions"

func apiAddItemIntoInventory(w http.ResponseWriter, r *http.Request, db *sql.DB) {

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

	var body map[string]interface{}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		writeResultJSON(w, fmt.Sprintf("Invalid JSON %v", err), http.StatusBadRequest)
		return
	}

	item_id, bItem_id := body["item_id"].(string)
	value, bValue := body["value"]
	status, bStatus := body["status"].(string)

	if !bItem_id {
		writeResultJSON(w, "item_id cannot be found in body", http.StatusBadRequest)
		return
	}

	if !bValue {
		writeResultJSON(w, "value cannot be found in body", http.StatusBadRequest)
		return
	}

	if !bStatus {
		writeResultJSON(w, "status cannot be found in body", http.StatusBadRequest)
		return
	}

	jwtClaim, jwtClaimErr := getClaimsFromToken(strToken, jwtServerAccessSecret)

	if jwtClaimErr != nil {
		writeResultJSON(w, fmt.Sprintf("Failed to retrieve data from token %v", jwtClaimErr), http.StatusInternalServerError)
		return
	}

	transactionsID := uuid.New().String()

	addItemQuery := fmt.Sprintf(`INSERT INTO %s (transactions_ID, user_id, item_id, value, status) VALUES (?, ?, ?, ?, ?)`, dbTableUsersInventory)

	_, InsertErr := db.Exec(addItemQuery, transactionsID, jwtClaim.UserID, item_id, value, status)

	if InsertErr != nil {
		writeResultJSON(w, fmt.Sprintf("Failed to add item %v", InsertErr), http.StatusInternalServerError)
		return
	}

	//Fetch the latest inventory data from db
	fetchItemsQuery := fmt.Sprintf(`
	SELECT
	uit.item_id,
	(SELECT 
		SUM(
		CASE 
			WHEN uit2.status = 'CREDIT' THEN uit2.value 
			ELSE -uit2.value
		END) FROM %s uit2 WHERE uit.item_id = uit2.item_id AND uit.user_id = uit2.user_id) as total_value
	FROM %s uit
	WHERE uit.user_id = ?
	GROUP BY uit.item_id
	HAVING SUM(
	CASE WHEN status = 'CREDIT' THEN value ELSE -value END 
	) > 0
	`, dbTableUsersInventory, dbTableUsersInventory)

	itemsRows, queryErr := db.Query(fetchItemsQuery, jwtClaim.UserID)

	if queryErr != nil {
		writeResultJSON(w, fmt.Sprintf("Failed to fetch items %v", queryErr), http.StatusInternalServerError)
		return
	}

	var allItems []Item

	for itemsRows.Next() {
		var I Item

		scanItemErr := itemsRows.Scan(&I.ItemID, &I.Quantity)

		if scanItemErr != nil {
			continue
		}

		allItems = append(allItems, I)
	}

	marshalItems, marshaleErr := json.Marshal(allItems)

	if marshaleErr != nil {
		writeResultJSON(w, fmt.Sprintf("Failed to marshale items %v", marshaleErr), http.StatusInternalServerError)
		return
	}

	writeResultJSON(w, string(marshalItems), http.StatusOK)
}

func apiGetInventoryItems(w http.ResponseWriter, r *http.Request, db *sql.DB) {
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

	getInventoryItemsQuery := fmt.Sprintf(`
	SELECT
	uit.item_id,
	(SELECT 
		SUM(
		CASE 
			WHEN uit2.status = 'CREDIT' THEN uit2.value 
			ELSE -uit2.value
		END) FROM %s uit2 WHERE uit.item_id = uit2.item_id AND uit.user_id = uit2.user_id) as total_value
	FROM %s uit
	WHERE uit.user_id = ?
	GROUP BY uit.item_id
	HAVING SUM(
	CASE WHEN status = 'CREDIT' THEN value ELSE -value END 
	) > 0
	`, dbTableUsersInventory, dbTableUsersInventory)

	itemsRows, queryErr := db.Query(getInventoryItemsQuery, jwtClaim.UserID)

	if queryErr != nil {
		writeResultJSON(w, fmt.Sprintf("Failed to get inventory items %v", queryErr), http.StatusInternalServerError)
		return
	}

	var allItems []Item

	for itemsRows.Next() {
		var I Item

		scanItemErr := itemsRows.Scan(&I.ItemID, &I.Quantity)

		if scanItemErr != nil {
			continue
		}

		allItems = append(allItems, I)
	}

	marshalItems, marshaleErr := json.Marshal(allItems)

	if marshaleErr != nil {
		writeResultJSON(w, fmt.Sprintf("Failed to marshale items %v", marshaleErr), http.StatusInternalServerError)
		return
	}

	writeResultJSON(w, string(marshalItems), http.StatusOK)
}

func apiGiveItemsToUser(w http.ResponseWriter, r *http.Request, db *sql.DB) {

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

	type Response struct {
		Items        []Item `json:"items"`
		TargetUserID string `json:"target_user_id"`
	}

	var body Response
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		writeResultJSON(w, fmt.Sprintf("Invalid JSON %v", err), http.StatusBadRequest)
		return
	}

	tx, txErr := db.Begin()

	if txErr != nil {
		writeResultJSON(w, fmt.Sprintf("failed to start a transaction. %s", txErr), http.StatusBadRequest)
		return
	}

	//Check if the owner has enough items
	checkItemsQuery := fmt.Sprintf(`
	SELECT
	(SELECT
		SUM(
		CASE
			WHEN uit2.status = 'CREDIT' THEN uit2.value
			ELSE -uit2.value
		END) FROM %s uit2 WHERE uit.item_id = uit2.item_id AND uit.user_id = uit2.user_id) as total_value
	FROM %s uit
	WHERE uit.user_id = ? AND uit.item_id = ?
	GROUP BY uit.item_id
	HAVING SUM(
	CASE WHEN status = 'CREDIT' THEN value ELSE -value END
	) > 0
	`, dbTableUsersInventory, dbTableUsersInventory)

	var grantableItems []Item

	for _, i := range body.Items {

		var numberOfItem int

		scanItemErr := tx.QueryRow(checkItemsQuery, jwtClaim.UserID, i.ItemID).Scan(&numberOfItem)

		if scanItemErr != nil {
			log.Printf(" item id: %v - failed to scan %v \n", i.ItemID, scanItemErr)
			continue
		}

		var newItem Item

		remainingCount := numberOfItem - i.Quantity

		if remainingCount >= 0 {
			newItem.ItemID = i.ItemID
			newItem.Quantity = i.Quantity
		} else {
			newItem.ItemID = i.ItemID
			newItem.Quantity = numberOfItem
		}

		grantableItems = append(grantableItems, newItem)

	}

	transactionsID := uuid.New().String()

	// Consume the required items from the owner's inventory
	addItemRecordQuery := fmt.Sprintf(`INSERT INTO %s (transactions_ID, user_id, item_id, value, status) VALUES (?, ?, ?, ?, ?)`, dbTableUsersInventory)

	for _, i := range grantableItems {
		_, insertErr := tx.Exec(addItemRecordQuery, transactionsID, jwtClaim.UserID, i.ItemID, i.Quantity, "DEBIT")

		if insertErr != nil {
			writeResultJSON(w, fmt.Sprintf("DEBIT: failed to insert a transaction for %s. %v", i.ItemID, txErr), http.StatusInternalServerError)
			tx.Rollback()
			return
		}
	}

	// give items to user
	for _, i := range grantableItems {
		_, insertErr := tx.Exec(addItemRecordQuery, transactionsID, body.TargetUserID, i.ItemID, i.Quantity, "CREDIT")

		if insertErr != nil {
			writeResultJSON(w, fmt.Sprintf("CREDIT: failed to insert a transaction for %s. %v", i.ItemID, insertErr), http.StatusInternalServerError)
			tx.Rollback()
			return
		}
	}

	commitErr := tx.Commit()

	if commitErr != nil {
		writeResultJSON(w, fmt.Sprintf("failed to commit %v", commitErr), http.StatusInternalServerError)
		return
	}

	//Fetch the latest inventory data
	getInventoryItemsQuery := fmt.Sprintf(`
	SELECT
	uit.item_id,
	(SELECT 
		SUM(
		CASE 
			WHEN uit2.status = 'CREDIT' THEN uit2.value 
			ELSE -uit2.value
		END) FROM %s uit2 WHERE uit.item_id = uit2.item_id  AND uit.user_id = uit2.user_id) as total_value
	FROM %s uit
	WHERE uit.user_id = ?
	GROUP BY uit.item_id
	HAVING SUM(
	CASE WHEN status = 'CREDIT' THEN value ELSE -value END 
	) > 0
	`, dbTableUsersInventory, dbTableUsersInventory)

	itemsRows, queryErr := db.Query(getInventoryItemsQuery, jwtClaim.UserID)

	if queryErr != nil {
		writeResultJSON(w, fmt.Sprintf("Failed to get inventory items %v", queryErr), http.StatusInternalServerError)
		return
	}

	var allItems []Item

	for itemsRows.Next() {
		var I Item

		scanItemErr := itemsRows.Scan(&I.ItemID, &I.Quantity)

		if scanItemErr != nil {
			continue
		}

		allItems = append(allItems, I)
	}

	marshalItems, marshaleErr := json.Marshal(allItems)

	if marshaleErr != nil {
		writeResultJSON(w, fmt.Sprintf("Failed to marshale items %v", marshaleErr), http.StatusInternalServerError)
		return
	}

	writeResultJSON(w, string(marshalItems), http.StatusOK)
}

func apiSellItemsToUser(w http.ResponseWriter, r *http.Request, db *sql.DB) {

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

	type Response struct {
		Items        []Item `json:"items"`
		TargetUserID string `json:"target_user_id"`
		Cost         []Item `json:"cost"`
	}

	var body Response
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		writeResultJSON(w, fmt.Sprintf("Invalid JSON %v", err), http.StatusBadRequest)
		return
	}

	tx, txErr := db.Begin()

	if txErr != nil {
		writeResultJSON(w, fmt.Sprintf("failed to start a transaction. %s", txErr), http.StatusBadRequest)
		return
	}

	//////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////
	//Check if the owner and buyer have enough items

	checkItemsQuery := fmt.Sprintf(`
	SELECT
	(SELECT
		SUM(
		CASE
			WHEN uit2.status = 'CREDIT' THEN uit2.value
			ELSE -uit2.value
		END) FROM %s uit2 WHERE uit.item_id = uit2.item_id AND uit.user_id = uit2.user_id) as total_value
	FROM %s uit
	WHERE uit.user_id = ? AND uit.item_id = ?
	GROUP BY uit.item_id
	HAVING SUM(
	CASE WHEN status = 'CREDIT' THEN value ELSE -value END
	) > 0
	`, dbTableUsersInventory, dbTableUsersInventory)

	var ownerProvidedItemsCount int

	//Check owner
	for _, i := range body.Items {

		var ownerItemCount int

		scanItemErr := tx.QueryRow(checkItemsQuery, jwtClaim.UserID, i.ItemID).Scan(&ownerItemCount)

		if scanItemErr != nil {
			log.Printf(" item id: %v - failed to scan %v \n", i.ItemID, scanItemErr)
			continue
		}

		if (ownerItemCount - i.Quantity) >= 0 {
			ownerProvidedItemsCount++
		}

		/*var newItem Item

		if (ownerItemCount - i.Quantity) >= 0 {

			newItem.ItemID = i.ItemID
			newItem.Quantity = i.Quantity
			grantableItems = append(grantableItems, newItem)

		}*/

	}

	if !(ownerProvidedItemsCount >= len(body.Items)) {
		writeResultJSON(w, "The owner doesn't have enough items to trade with you.", http.StatusConflict)
		return
	}

	//Check buyer
	var buyerProvidedItemsCount int
	for _, i := range body.Cost {

		var buyerItemCount int

		scanItemErr := tx.QueryRow(checkItemsQuery, body.TargetUserID, i.ItemID).Scan(&buyerItemCount)

		if scanItemErr != nil {
			log.Printf(" item id: %v - failed to scan %v \n", i.ItemID, scanItemErr)
			continue
		}

		if (buyerItemCount - i.Quantity) >= 0 {
			buyerProvidedItemsCount++
		}

	}

	if !(buyerProvidedItemsCount >= len(body.Cost)) {
		writeResultJSON(w, "The buyer doesn't have enough items to trade with the seller.", http.StatusConflict)
		return
	}

	//////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////

	// Consume and Give
	addItemRecordQuery := fmt.Sprintf(`INSERT INTO %s (transactions_ID, user_id, item_id, value, status) VALUES (?, ?, ?, ?, ?)`, dbTableUsersInventory)

	transactionsID := uuid.New().String()

	// Consume the required items from the owner's inventory, and give it to the buyer
	for _, i := range body.Items {

		//Consume the required items from the owner's inventory
		_, insertErr := tx.Exec(addItemRecordQuery, transactionsID, jwtClaim.UserID, i.ItemID, i.Quantity, "DEBIT")

		if insertErr != nil {
			writeResultJSON(w, fmt.Sprintf("DEBIT: failed to insert a transaction for %s. %v", i.ItemID, insertErr), http.StatusInternalServerError)
			tx.Rollback()
			return
		}

		//Give items from the buyer
		_, insertErr = tx.Exec(addItemRecordQuery, transactionsID, body.TargetUserID, i.ItemID, i.Quantity, "CREDIT")

		if insertErr != nil {
			writeResultJSON(w, fmt.Sprintf("CREDIT: failed to insert a transaction for %s. %v", i.ItemID, insertErr), http.StatusInternalServerError)
			tx.Rollback()
			return
		}
	}

	// Consume the required items from the buyer's inventory, and give it to the vendor
	for _, i := range body.Cost {

		//Consume the required items from the buyer's inventory
		_, insertErr := tx.Exec(addItemRecordQuery, transactionsID, body.TargetUserID, i.ItemID, i.Quantity, "DEBIT")

		if insertErr != nil {
			writeResultJSON(w, fmt.Sprintf("DEBIT: failed to insert a transaction for %s. %v", i.ItemID, insertErr), http.StatusInternalServerError)
			tx.Rollback()
			return
		}

		//Give items from the buyer
		_, insertErr = tx.Exec(addItemRecordQuery, transactionsID, jwtClaim.UserID, i.ItemID, i.Quantity, "CREDIT")

		if insertErr != nil {
			writeResultJSON(w, fmt.Sprintf("CREDIT: failed to insert a transaction for %s. %v", i.ItemID, insertErr), http.StatusInternalServerError)
			tx.Rollback()
			return
		}
	}

	commitErr := tx.Commit()

	if commitErr != nil {
		writeResultJSON(w, fmt.Sprintf("failed to commit %v", commitErr), http.StatusInternalServerError)
		return
	}

	//////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////

	//Fetch the latest inventory data
	getInventoryItemsQuery := fmt.Sprintf(`
	SELECT
	uit.item_id,
	(SELECT 
		SUM(
		CASE 
			WHEN uit2.status = 'CREDIT' THEN uit2.value 
			ELSE -uit2.value
		END) FROM %s uit2 WHERE uit.item_id = uit2.item_id  AND uit.user_id = uit2.user_id) as total_value
	FROM %s uit
	WHERE uit.user_id = ?
	GROUP BY uit.item_id
	HAVING SUM(
	CASE WHEN status = 'CREDIT' THEN value ELSE -value END 
	) > 0
	`, dbTableUsersInventory, dbTableUsersInventory)

	itemsRows, queryErr := db.Query(getInventoryItemsQuery, jwtClaim.UserID)

	if queryErr != nil {
		writeResultJSON(w, fmt.Sprintf("Failed to get inventory items %v", queryErr), http.StatusInternalServerError)
		return
	}

	var allItems []Item

	for itemsRows.Next() {
		var I Item

		scanItemErr := itemsRows.Scan(&I.ItemID, &I.Quantity)

		if scanItemErr != nil {
			continue
		}

		allItems = append(allItems, I)
	}

	marshalItems, marshaleErr := json.Marshal(allItems)

	if marshaleErr != nil {
		writeResultJSON(w, fmt.Sprintf("Failed to marshale items %v", marshaleErr), http.StatusInternalServerError)
		return
	}

	writeResultJSON(w, string(marshalItems), http.StatusOK)
}

func checkUserHasEnoughItems() {

}
