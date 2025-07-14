package main

import (
	"fmt"
	"servertesting/module/http"
	"testing"
)

func TestRun(t *testing.T) {
	token, AuthenErr := http.AuthenticateUser("Jayden", "123456")

	if AuthenErr != nil {
		fmt.Printf("Authentication failed: %v", AuthenErr)
	}

	fmt.Printf("Successfully authenticated: %s", token)

}
