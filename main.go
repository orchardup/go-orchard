package main

import "fmt"
import "github.com/orchardup/orchard/cli"
import "github.com/orchardup/orchard/github.com/docopt/docopt.go"

func main() {
	usage := `Orchard.

Usage:
  orchard hosts
  orchard hosts create NAME
  orchard hosts rm NAME
  orchard docker COMMAND...

Options:
  -h --help   Show this screen.
  --version   Show version.`

	args, err := docopt.Parse(usage, nil, true, "Orchard 2.0.0", true)
	if err != nil {
		fmt.Println("Error parsing arguments: %s\n", err)
		return
	}

	if args["hosts"] == true {
		if err := Hosts(args); err != nil {
			fmt.Println(err)
		}
	} else {
		fmt.Println(args)
	}
}

func Hosts(args map[string]interface{}) error {
	httpClient, err := authenticator.Authenticate()
	if err != nil {
		fmt.Printf("Error authenticating:\n%s\n", err)
	}

	if args["create"] == true {
		host, err := httpClient.CreateHost(args["NAME"].(string))
		if err != nil {
			return err
		}
		fmt.Printf("Created %s\n", host.Name)
	} else if args["rm"] == true {
		err := httpClient.DeleteHost(args["NAME"].(string))
		if err != nil {
			return err
		}
		fmt.Printf("Removed %s\n", args["NAME"].(string))
	} else {
		hosts, err := httpClient.GetHosts()
		if err != nil {
			return err
		}

		fmt.Printf("hosts: %v\n", hosts)
	}

	return nil
}
