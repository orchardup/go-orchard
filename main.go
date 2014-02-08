package main

import "fmt"
import "github.com/orchardup/orchard/cli"

func main() {
	httpClient, err := authenticator.Authenticate()
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
