package httpserverroutes

import (
	"database/sql"
	"log"
	"net/http"
)

func RegisterAPI(db *sql.DB) error {

	if db == nil {
		log.Println("Failed to register API because data base is invalid")
	}

	//Authentication & Registration
	http.HandleFunc("/apiRefreshToken", func(w http.ResponseWriter, r *http.Request) {
		apiRefreshToken(w, r)
	})
	http.HandleFunc("/apiRegisterUser", func(w http.ResponseWriter, r *http.Request) {
		apiRegisterUser(w, r, db)
	})
	http.HandleFunc("/apiUserAuthentication", func(w http.ResponseWriter, r *http.Request) {
		apiUserAuthentication(w, r, db)
	})
	http.HandleFunc("/apiTestAuthentication", func(w http.ResponseWriter, r *http.Request) {
		apiTestAuthentication(w, r)
	})

	//Inventory & Trading
	http.HandleFunc("/apiAddItemIntoInventory", func(w http.ResponseWriter, r *http.Request) {
		apiAddItemIntoInventory(w, r, db)
	})
	http.HandleFunc("/apiGetInventoryItems", func(w http.ResponseWriter, r *http.Request) {
		apiGetInventoryItems(w, r, db)
	})
	http.HandleFunc("/apiGiveItemsToUser", func(w http.ResponseWriter, r *http.Request) {
		apiGiveItemsToUser(w, r, db)
	})
	http.HandleFunc("/apiSellItemsToUser", func(w http.ResponseWriter, r *http.Request) {
		apiSellItemsToUser(w, r, db)
	})
	return nil
}
