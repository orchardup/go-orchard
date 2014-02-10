package main

import (
	"crypto/tls"
	"errors"
	"fmt"
	"github.com/orchardup/orchard/cli"
	"github.com/orchardup/orchard/github.com/docopt/docopt.go"
	"github.com/orchardup/orchard/proxy"
	"github.com/orchardup/orchard/tlsconfig"
	"io/ioutil"
	"net"
	"os"
	"os/exec"
	"os/signal"
	"path"
	"strings"
	"syscall"
	"text/tabwriter"
)

func main() {
	usage := `Orchard.

Usage:
  orchard hosts
  orchard hosts create NAME
  orchard hosts rm NAME
  orchard [options] docker [COMMAND...]
  orchard [options] proxy

Options:
  -h --help               Show this screen.
  --version               Show version.
  -H NAME, --host=NAME    Name of host to connect to (instead of 'default')`

	args, err := docopt.Parse(usage, nil, true, "Orchard 2.0.0", true)
	if err != nil {
		fmt.Println("Error parsing arguments: %s\n", err)
		return
	}

	var cmdErr error = nil

	if args["hosts"] == true {
		cmdErr = Hosts(args)
	} else if args["docker"] == true || args["proxy"] == true {
		cmdErr = Docker(args)
	}

	if cmdErr != nil {
		fmt.Fprintln(os.Stderr, cmdErr)
		os.Exit(1)
	}
}

func Docker(args map[string]interface{}) error {
	hostName := "default"
	if args["--host"] != nil {
		hostName = args["--host"].(string)
	}

	dirname, err := ioutil.TempDir("/tmp", "orchard-")
	if err != nil {
		return fmt.Errorf("Error creating temporary directory: %s\n", err)
	}
	defer os.RemoveAll(dirname)
	socketPath := path.Join(dirname, "orchard.sock")

	p, err := MakeProxy(socketPath, hostName)
	if err != nil {
		return fmt.Errorf("Error starting proxy: %v\n", err)
	}

	go p.Start()
	defer p.Stop()

	if err := <-p.ErrorChannel; err != nil {
		return fmt.Errorf("Error starting proxy: %v\n", err)
	}

	if args["docker"] == true {
		err := CallDocker(args["COMMAND"].([]string), []string{"DOCKER_HOST=unix://" + socketPath})
		if err != nil {
			return fmt.Errorf("Docker exited with error")
		}
	} else {
		fmt.Println("Started proxy at unix://" + socketPath)

		c := make(chan os.Signal)
		signal.Notify(c, syscall.SIGINT, syscall.SIGKILL)
		<-c

		fmt.Println("\nStopping proxy")
	}

	return nil
}

func MakeProxy(socketPath string, hostName string) (*proxy.Proxy, error) {
	httpClient, err := authenticator.Authenticate()
	if err != nil {
		return nil, err
	}

	host, err := httpClient.GetHost(hostName)
	if err != nil {
		return nil, err
	}
	destination := host.IPv4_Address + ":4243"

	fmt.Printf("Connecting to %s...\n", destination)

	certData := []byte(host.Client_Cert)
	keyData := []byte(host.Client_Key)
	config, err := tlsconfig.GetTLSConfig(certData, keyData)
	if err != nil {
		return nil, err
	}

	return proxy.New(
		func() (net.Listener, error) { return net.Listen("unix", socketPath) },
		func() (net.Conn, error) { return tls.Dial("tcp", destination, config) },
	), nil
}

func CallDocker(args []string, env []string) error {
	dockerPath := GetDockerPath()
	if dockerPath == "" {
		return errors.New("Can't find `docker` executable in $PATH.\nYou might need to install it: http://docs.docker.io/en/latest/installation/#installation-list")
	}

	cmd := exec.Command(dockerPath, args...)
	cmd.Env = env
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

func GetDockerPath() string {
	for _, dir := range strings.Split(os.Getenv("PATH"), ":") {
		dockerPath := path.Join(dir, "docker")
		_, err := os.Stat(dockerPath)
		if err == nil {
			return dockerPath
		}
	}
	return ""
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
		fmt.Printf("Created %s with IP address %s\n", host.Name, host.IPv4_Address)
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
		fmt.Fprintln(writer, "NAME\tIP")
		for _, host := range hosts {
			fmt.Fprintf(writer, "%s\t%s\n", host.Name, host.IPv4_Address)
		}
		writer.Flush()
	}

	return nil
}
