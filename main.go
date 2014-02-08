package main

import "fmt"
import "os"
import "text/tabwriter"
import "github.com/orchardup/orchard/cli"
import "github.com/orchardup/orchard/proxy"
import "github.com/orchardup/orchard/github.com/docopt/docopt.go"

import "os/exec"

func main() {
	usage := `Orchard.

Usage:
  orchard hosts
  orchard hosts create NAME
  orchard hosts rm NAME
  orchard docker [COMMAND...]

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
	} else if args["docker"] == true {
		p := proxy.New("unix", "/tmp/orchard.sock", "tcp", "localdocker:4243")
		go p.Start()

		err := <-p.ErrorChannel
		if err != nil {
			fmt.Printf("proxy failed to start: '%v'\n", err)
		} else {
			err := CallDocker(
				args["COMMAND"].([]string),
				[]string{
					"DOCKER_HOST=unix:///tmp/orchard.sock",
					"DEBUG=1",
				},
			)
			if err != nil {
				fmt.Printf("docker failed: %v\n", err)
			}
		}

		fmt.Println("stopping proxy")
		p.Stop()
	} else {
		fmt.Println(args)
	}
}

func CallDocker(args []string, env []string) error {
	// TODO: handle case where docker isn't installed
	cmd := exec.Command("/usr/local/bin/docker", args...)
	cmd.Env = env
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
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

		writer := tabwriter.NewWriter(os.Stdout, 20, 1, 3, ' ', 0)
		fmt.Fprintln(writer, "ID\tNAME")
		for _, host := range hosts {
			fmt.Fprintf(writer, "%s\t%s\n", host.ID, host.Name)
		}
		writer.Flush()
	}

	return nil
}
