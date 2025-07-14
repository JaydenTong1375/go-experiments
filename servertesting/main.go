package main

import (
	"servertesting/module/http"

	_ "github.com/go-sql-driver/mysql"
)

func main() {

	http.StartHosting()

}
