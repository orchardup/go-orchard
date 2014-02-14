package main

import (
	"crypto/tls"
	"errors"
	"fmt"
	"github.com/orchardup/orchard/cli"
	"github.com/orchardup/orchard/proxy"
	"github.com/orchardup/orchard/tlsconfig"
	"github.com/orchardup/orchard/vendor/github.com/docopt/docopt.go"
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
  orchard start [NAME]
  orchard stop [NAME]
  orchard [options] docker [COMMAND...]
  orchard [options] proxy

Options:
  -h --help               Show this screen.
  --version               Show version.
  -H NAME, --host=NAME    Name of host to connect to (instead of 'default')`

	args, err := docopt.Parse(usage, nil, true, "Orchard 2.0.0", true)
	if err != nil {
		fmt.Fprintln(os.Stderr, "Error parsing arguments: %s\n", err)
		return
	}

	var cmdErr error = nil

	if args["hosts"] == true {
		cmdErr = Hosts(args)
	} else if args["start"] == true {
		cmdErr = Start(args)
	} else if args["stop"] == true {
		cmdErr = Stop(args)
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
		fmt.Fprintln(os.Stderr, "Started proxy at unix://"+socketPath)

		c := make(chan os.Signal)
		signal.Notify(c, syscall.SIGINT, syscall.SIGKILL)
		<-c

		fmt.Fprintln(os.Stderr, "\nStopping proxy")
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
	destination := host.IPAddress + ":4243"

	certData := []byte(host.ClientCert)
	keyData := []byte(host.ClientKey)
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

func Start(args map[string]interface{}) error {
	httpClient, err := authenticator.Authenticate()
	if err != nil {
		return err
	}

	hostName, humanName := GetHostName(args)
	humanName = strings.ToUpper(humanName[0:1]) + humanName[1:]

	host, err := httpClient.CreateHost(hostName)
	if err != nil {
		// HACK. api.go should decode JSON and return a specific type of error for this case.
		if strings.Contains(err.Error(), "already exists") {
			fmt.Fprintf(os.Stderr, "%s is already running.\nYou can create additional hosts with `orchard start NAME`.\n", humanName)
			return nil
		}
		if strings.Contains(err.Error(), "Invalid value") {
			fmt.Fprintf(os.Stderr, "Sorry, '%s' isn't a valid host name.\nHost names can only contain lowercase letters, numbers and underscores.\n", hostName)
			return nil
		}

		return err
	}
	fmt.Fprintf(os.Stderr, "%s running at %s\n", humanName, host.IPAddress)

	return nil
}

func Stop(args map[string]interface{}) error {
	hostName, humanName := GetHostName(args)

	var confirm string
	fmt.Printf("Going to stop and delete %s. All data on it will be lost.\n", humanName)
	fmt.Print("Are you sure you're ready? [yN] ")
	fmt.Scanln(&confirm)

	if strings.ToLower(confirm) != "y" {
		return nil
	}

	httpClient, err := authenticator.Authenticate()
	if err != nil {
		return err
	}

	err = httpClient.DeleteHost(hostName)
	if err != nil {
		return err
	}
	fmt.Fprintf(os.Stderr, "Stopped %s\n", humanName)

	return nil
}

func GetHostName(args map[string]interface{}) (string, string) {
	hostName := "default"
	humanName := "default host"

	if args["NAME"] != nil {
		hostName = args["NAME"].(string)
		humanName = fmt.Sprintf("host '%s'", hostName)
	}

	return hostName, humanName
}

func Hosts(args map[string]interface{}) error {
	httpClient, err := authenticator.Authenticate()
	if err != nil {
		return err
	}

	hosts, err := httpClient.GetHosts()
	if err != nil {
		return err
	}

	writer := tabwriter.NewWriter(os.Stdout, 20, 1, 3, ' ', 0)
	fmt.Fprintln(writer, "NAME\tIP")
	for _, host := range hosts {
		fmt.Fprintf(writer, "%s\t%s\n", host.Name, host.IPAddress)
	}
	writer.Flush()

	return nil
}
