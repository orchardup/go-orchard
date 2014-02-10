package main

import "fmt"
import "os"
import "path"
import "strings"
import "errors"
import "text/tabwriter"
import "github.com/orchardup/orchard/cli"
import "github.com/orchardup/orchard/proxy"
import "github.com/orchardup/orchard/github.com/docopt/docopt.go"

import "net"
import "crypto/tls"
import "crypto/x509"
import "io/ioutil"
import "os/exec"
import "os/signal"
import "syscall"

func main() {
	usage := `Orchard.

Usage:
  orchard hosts
  orchard hosts create NAME
  orchard hosts rm NAME
  orchard docker [COMMAND...]
  orchard proxy

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
	} else if args["docker"] == true || args["proxy"] == true {
		socketPath := "/tmp/orchard.sock"

		p, err := MakeProxy(socketPath, "default")
		if err != nil {
			fmt.Printf("Error starting proxy: %v\n", err)
			return
		}

		go p.Start()
		defer p.Stop()

		if err := <-p.ErrorChannel; err != nil {
			fmt.Printf("Error starting proxy: %v\n", err)
			return
		}

		if args["docker"] == true {
			err := CallDocker(args["COMMAND"].([]string), []string{"DOCKER_HOST=unix://" + socketPath})
			if err != nil {
				fmt.Println(err)
			}
		} else {
			fmt.Println("Started proxy at unix://" + socketPath)

			c := make(chan os.Signal)
			signal.Notify(c, syscall.SIGINT, syscall.SIGKILL)
			<-c

			fmt.Println("\nStopping proxy")
		}
	}
}

func MakeProxy(socketPath string, hostName string) (*proxy.Proxy, error) {
	// httpClient, err := authenticator.Authenticate()
	// if err != nil {
	// 	return nil, err
	// }

	// host, err := httpClient.GetHost(hostName)
	// if err != nil {
	// 	return nil, err
	// }
	// destination := host.IPv4_Address+":443"

	destination := "107.170.41.173:4243"
	certData, err := ioutil.ReadFile("client-cert.pem")
	if err != nil {
		return nil, err
	}
	keyData, err := ioutil.ReadFile("client-key.pem")
	if err != nil {
		return nil, err
	}

	config, err := GetTLSConfig(certData, keyData)
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

func GetTLSConfig(clientCertPEMData, clientKeyPEMData []byte) (*tls.Config, error) {
	pemData, err := ioutil.ReadFile("orchard-certs.pem")
	if err != nil {
		return nil, err
	}
	certPool := x509.NewCertPool()
	certPool.AppendCertsFromPEM(pemData)

	clientCert, err := tls.X509KeyPair(clientCertPEMData, clientKeyPEMData)
	if err != nil {
		return nil, err
	}

	config := new(tls.Config)
	config.RootCAs = certPool
	config.Certificates = []tls.Certificate{clientCert}
	config.BuildNameToCertificate()

	return config, nil
}
