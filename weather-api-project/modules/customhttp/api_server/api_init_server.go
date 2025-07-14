package apiserver

import (
	"database/sql"
	"fmt"
	"net/http"
)

func RegisterServerAPI(db *sql.DB) error {

	if db == nil {
		return fmt.Errorf("failed to register API because the database is invalid")
	}
	http.HandleFunc("/apiGetWeatherData", func(w http.ResponseWriter, r *http.Request) {
		apiGetWeatherData(w, r, db)
	})
	http.HandleFunc("/apiUpdateWeatherData", func(w http.ResponseWriter, r *http.Request) {
		apiUpdateWeatherData(w, r, db)
	})

	return nil
}
