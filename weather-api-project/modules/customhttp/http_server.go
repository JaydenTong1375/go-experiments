package customhttp

import (
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"time"

	apiServer "weather-api-project/modules/customhttp/api_server"

	_ "github.com/go-sql-driver/mysql"
)

func StartHosting() {

	db, initDBErr := initDB()

	if initDBErr != nil {
		fmt.Println(initDBErr.Error())
	}

	defer db.Close()

	resgiterErr := apiServer.RegisterServerAPI(db)

	if resgiterErr != nil {
		panic(resgiterErr)
	}

	// Use HTTPS with your cert and key
	fmt.Println("HTTPS server running on port 8090")
	err := http.ListenAndServeTLS("0.0.0.0:8090", "./Cert/fucherng1375.ddns.net-crt.pem", "./Cert/fucherng1375.ddns.net-key-decrypted.pem", nil)
	if err != nil {
		panic(err)
	}
}

func initDB() (*sql.DB, error) {

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
		log.Println("Trying to connect to DB...")
		PingErr := db.Ping()

		if PingErr != nil {
			log.Printf("failed to connect to the database: %v \n", PingErr)
		}

		bConnectSuccessfully = PingErr == nil

	}

	log.Println("Successfully connected to MariaDB using go-sql-driver/mysql!")

	return db, nil
}
