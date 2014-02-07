package main

import "fmt"
import "github.com/orchardup/orchard/api"
import "code.google.com/p/gopass"

func main() {
	var httpClient api.Client
	httpClient = api.HTTPClient{"http://localdocker:8000/api/v1"}

	username, password := Prompt()
	token, err := httpClient.GetAuthToken(username, password)
	if err != nil {
		fmt.Printf("Error getting token:\n%s\n", err)
		return
	}
	hosts, err := httpClient.GetHosts(token)
	if err != nil {
		fmt.Printf("Error getting hosts:\n%s\n", err)
		return
	}
	fmt.Printf("hosts: %v\n", hosts)
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
