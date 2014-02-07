package main

import "fmt"
import "github.com/orchardup/orchard/api"
import "code.google.com/p/gopass"

func main() {
	httpClient, err := Authenticate()
	if err != nil {
		fmt.Printf("Error authenticating:\n%s\n", err)
	}

	hosts, err := httpClient.GetHosts()
	if err != nil {
		fmt.Printf("Error getting hosts:\n%s\n", err)
		return
	}

	fmt.Printf("hosts: %v\n", hosts)
}

func Authenticate() (*api.HTTPClient, error) {
	httpClient := api.HTTPClient{"http://localdocker:8000/api/v1", ""}

	username, password := Prompt()

	token, err := httpClient.GetAuthToken(username, password)
	if err != nil {
		return nil, err
	}
	httpClient.Token = token

	return &httpClient, nil
}

func Prompt() (string, string) {
	var (
		username string
		password string
	)
	fmt.Print("Username you signed up for Orchard with: ")
	fmt.Scanln(&username)
	password, _ = gopass.GetPass("Password: ")
	return username, password
}
