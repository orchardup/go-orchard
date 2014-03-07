package commands

import (
	"errors"
	"flag"
	"fmt"
	"github.com/orchardup/go-orchard/api"
	"github.com/orchardup/go-orchard/authenticator"
	"github.com/orchardup/go-orchard/proxy"
	"github.com/orchardup/go-orchard/tlsconfig"
	"github.com/orchardup/go-orchard/utils"
	"github.com/orchardup/go-orchard/vendor/crypto/tls"
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

type Command struct {
	Run       func(cmd *Command, args []string) error
	UsageLine string
	Short     string
	Long      string
	Flag      flag.FlagSet
}

func (c *Command) Name() string {
	name := c.UsageLine
	i := strings.Index(name, " ")
	if i >= 0 {
		name = name[:i]
	}
	return name
}

func (c *Command) Usage() {
	fmt.Fprintf(os.Stderr, "Usage: %s\n\n", c.UsageLine)
	fmt.Fprintf(os.Stderr, "%s\n", strings.TrimSpace(c.Long))
	os.Exit(2)
}

func (c *Command) UsageError(format string, args ...interface{}) error {
	fmt.Fprintf(os.Stderr, format, args...)
	fmt.Fprintf(os.Stderr, "\nUsage: %s\n", c.UsageLine)
	os.Exit(2)
	return fmt.Errorf(format, args...)
}

var All = []*Command{
	Hosts,
	Docker,
	Proxy,
	IP,
}

var HostSubcommands = []*Command{
	CreateHost,
	RemoveHost,
}

func init() {
	Hosts.Run = RunHosts
	CreateHost.Run = RunCreateHost
	RemoveHost.Run = RunRemoveHost
	Docker.Run = RunDocker
	Proxy.Run = RunProxy
	IP.Run = RunIP
}

var Hosts = &Command{
	UsageLine: "hosts",
	Short:     "Manage hosts",
	Long: `Manage hosts.

Usage: orchard hosts [COMMAND] [ARGS...]

Commands:
  ls          List hosts (default)
  create      Create a host
  rm          Remove a host

Run 'orchard hosts COMMAND -h' for more information on a command.
`,
}

var CreateHost = &Command{
	UsageLine: "create [-m MEMORY] [NAME]",
	Short:     "Create a host",
	Long: fmt.Sprintf(`Create a host.

You can optionally specify a name for the host - if not, it will be
named 'default', and 'orchard docker' commands will use it automatically.

You can also specify how much RAM the host should have with -m.
Valid amounts are %s.`, validSizes),
}

var flCreateSize = CreateHost.Flag.String("m", "512M", "")
var validSizes = "512M, 1G, 2G, 4G and 8G"

var RemoveHost = &Command{
	UsageLine: "rm [-f] [NAME]",
	Short:     "Remove a host",
	Long: `Remove a host.

You can optionally specify which host to remove - if you don't, the default
host (named 'default') will be removed.

Set -f to bypass the confirmation step, at your peril.
`,
}

var flRemoveHostForce = RemoveHost.Flag.Bool("f", false, "")

var Docker = &Command{
	UsageLine: "docker [-H HOST] [COMMAND...]",
	Short:     "Run a Docker command against a host",
	Long: `Run a Docker command against a host.

Wraps the 'docker' command-line tool - see the Docker website for reference:

    http://docs.docker.io/en/latest/reference/commandline/

You can optionally specify a host by name - if you don't, the default host
will be used.`,
}

var flDockerHost = Docker.Flag.String("H", "", "")

var Proxy = &Command{
	UsageLine: "proxy [-H HOST] [LISTEN_URL]",
	Short:     "Start a local proxy to a host's Docker daemon",
	Long: `Start a local proxy to a host's Docker daemon.

By default, listens on a Unix socket at a random path, e.g.

    $ orchard proxy
    Started proxy at unix:///tmp/orchard-12345/orchard.sock

    $ docker -H unix:///tmp/orchard-12345/orchard.sock run ubuntu echo hello world
    hello world

Instead, you can specify a URL to listen on, which can be a socket or TCP address:

    $ orchard proxy unix:///path/to/socket
    $ orchard proxy tcp://localhost:1234
`,
}

var flProxyHost = Proxy.Flag.String("H", "", "")

var IP = &Command{
	UsageLine: "ip [NAME]",
	Short:     "Print a hosts's IP address to stdout",
	Long: `Print a hosts's IP address to stdout.

You can optionally specify which host - if you don't, the default
host (named 'default') will be assumed.
`,
}

func RunHosts(cmd *Command, args []string) error {
	list := len(args) == 0 || (len(args) == 1 && args[0] == "ls")

	if !list {
		for _, subcommand := range HostSubcommands {
			if subcommand.Name() == args[0] {
				subcommand.Flag.Usage = func() { subcommand.Usage() }
				subcommand.Flag.Parse(args[1:])
				args = subcommand.Flag.Args()
				err := subcommand.Run(subcommand, args)
				return err
			}
		}

		return fmt.Errorf("Unknown `hosts` subcommand: %s", args[0])
	}

	httpClient, err := authenticator.Authenticate()
	if err != nil {
		return err
	}

	hosts, err := httpClient.GetHosts()
	if err != nil {
		return err
	}

	writer := tabwriter.NewWriter(os.Stdout, 20, 1, 3, ' ', 0)
	fmt.Fprintln(writer, "NAME\tSIZE\tIP")
	for _, host := range hosts {
		fmt.Fprintf(writer, "%s\t%s\t%s\n", host.Name, utils.HumanSize(host.Size*1024*1024), host.IPAddress)
	}
	writer.Flush()

	return nil
}

func RunCreateHost(cmd *Command, args []string) error {
	if len(args) > 1 {
		return cmd.UsageError("`orchard hosts create` expects at most 1 argument, but got more: %s", strings.Join(args[1:], " "))
	}

	httpClient, err := authenticator.Authenticate()
	if err != nil {
		return err
	}

	hostName, humanName := GetHostName(args)
	humanName = utils.Capitalize(humanName)

	size, sizeString := GetHostSize()
	if size == -1 {
		fmt.Fprintf(os.Stderr, "Sorry, %q isn't a size we support.\nValid sizes are %s.\n", sizeString, validSizes)
		return nil
	}

	host, err := httpClient.CreateHost(hostName, size)
	if err != nil {
		// HACK. api.go should decode JSON and return a specific type of error for this case.
		if strings.Contains(err.Error(), "already exists") {
			fmt.Fprintf(os.Stderr, "%s is already running.\nYou can create additional hosts with `orchard hosts create [NAME]`.\n", humanName)
			return nil
		}
		if strings.Contains(err.Error(), "Invalid value") {
			fmt.Fprintf(os.Stderr, "Sorry, '%s' isn't a valid host name.\nHost names can only contain lowercase letters, numbers and underscores.\n", hostName)
			return nil
		}
		if strings.Contains(err.Error(), "Unsupported size") {
			fmt.Fprintf(os.Stderr, "Sorry, %q isn't a size we support.\nValid sizes are %s.\n", sizeString, validSizes)
			return nil
		}

		return err
	}
	fmt.Fprintf(os.Stderr, "%s running at %s\n", humanName, host.IPAddress)

	return nil
}

func RunRemoveHost(cmd *Command, args []string) error {
	if len(args) > 1 {
		return cmd.UsageError("`orchard hosts rm` expects at most 1 argument, but got more: %s", strings.Join(args[1:], " "))
	}

	hostName, humanName := GetHostName(args)

	if !*flRemoveHostForce {
		var confirm string
		fmt.Printf("Going to remove %s. All data on it will be lost.\n", humanName)
		fmt.Print("Are you sure you're ready? [yN] ")
		fmt.Scanln(&confirm)

		if strings.ToLower(confirm) != "y" {
			return nil
		}
	}

	httpClient, err := authenticator.Authenticate()
	if err != nil {
		return err
	}

	err = httpClient.DeleteHost(hostName)
	if err != nil {
		// HACK. api.go should decode JSON and return a specific type of error for this case.
		if strings.Contains(err.Error(), "Not found") {
			fmt.Fprintf(os.Stderr, "%s doesn't seem to be running.\nYou can view your running hosts with `orchard hosts`.\n", utils.Capitalize(humanName))
			return nil
		}

		return err
	}
	fmt.Fprintf(os.Stderr, "Removed %s\n", humanName)

	return nil
}

func RunDocker(cmd *Command, args []string) error {
	return WithDockerProxy("", *flDockerHost, func(listenURL string) error {
		err := CallDocker(args, []string{"DOCKER_HOST=" + listenURL})
		if err != nil {
			return fmt.Errorf("Docker exited with error")
		}
		return nil
	})
}

func RunProxy(cmd *Command, args []string) error {
	specifiedURL := ""

	if len(args) == 1 {
		specifiedURL = args[0]
	} else if len(args) > 1 {
		return cmd.UsageError("`orchard proxy` expects at most 1 argument, but got: %s", strings.Join(args, " "))
	}

	return WithDockerProxy(specifiedURL, *flProxyHost, func(listenURL string) error {
		fmt.Fprintf(os.Stderr, `Started proxy. Use it by setting your Docker host:
export DOCKER_HOST=%s
`, listenURL)

		c := make(chan os.Signal)
		signal.Notify(c, syscall.SIGINT, syscall.SIGKILL)
		<-c

		fmt.Fprintln(os.Stderr, "\nStopping proxy")
		return nil
	})
}

func RunIP(cmd *Command, args []string) error {
	if len(args) > 1 {
		return cmd.UsageError("`orchard ip` expects at most 1 argument, but got more: %s", strings.Join(args[1:], " "))
	}

	hostName, _ := GetHostName(args)

	host, err := GetHost(hostName)
	if err != nil {
		return err
	}

	fmt.Fprintln(os.Stdout, host.IPAddress)
	return nil
}

func WithDockerProxy(listenURL, hostName string, callback func(string) error) error {
	if hostName == "" {
		hostName = "default"
	}

	if listenURL == "" {
		dirname, err := ioutil.TempDir("/tmp", "orchard-")
		if err != nil {
			return fmt.Errorf("Error creating temporary directory: %s\n", err)
		}
		defer os.RemoveAll(dirname)
		listenURL = fmt.Sprintf("unix://%s", path.Join(dirname, "orchard.sock"))
	}

	listenType, listenAddr, err := ListenArgs(listenURL)
	if err != nil {
		return err
	}

	p, err := MakeProxy(listenType, listenAddr, hostName)
	if err != nil {
		return fmt.Errorf("Error starting proxy: %v\n", err)
	}

	go p.Start()
	defer p.Stop()

	if err := <-p.ErrorChannel; err != nil {
		return fmt.Errorf("Error starting proxy: %v\n", err)
	}

	if err := callback(listenURL); err != nil {
		return err
	}

	return nil
}

var validListenTypes = []string{"tcp", "tcp4", "tcp6", "unix", "unixpacket"}

func ListenArgs(url string) (string, string, error) {
	parts := strings.SplitN(url, "://", 2)
	if len(parts) != 2 {
		return "", "", fmt.Errorf("Invalid URL: %q", url)
	}
	for _, validType := range validListenTypes {
		if parts[0] == validType {
			return parts[0], parts[1], nil
		}
	}
	return "", "", fmt.Errorf("Invalid URL type: %q", parts[0])
}

func MakeProxy(listenType, listenAddr string, hostName string) (*proxy.Proxy, error) {
	host, err := GetHost(hostName)
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
		func() (net.Listener, error) { return net.Listen(listenType, listenAddr) },
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

func GetHostName(args []string) (string, string) {
	hostName := "default"

	if len(args) > 0 {
		hostName = args[0]
	}

	return hostName, GetHumanHostName(hostName)
}

func GetHumanHostName(hostName string) string {
	if hostName == "default" {
		return "default host"
	} else {
		return fmt.Sprintf("host '%s'", hostName)
	}
}

func GetHostSize() (int, string) {
	sizeString := *flCreateSize

	bytes, err := utils.RAMInBytes(sizeString)
	if err != nil {
		return -1, sizeString
	}

	megs := bytes / (1024 * 1024)
	if megs < 1 {
		return -1, sizeString
	}

	return int(megs), sizeString
}

func GetHost(hostName string) (*api.Host, error) {
	httpClient, err := authenticator.Authenticate()
	if err != nil {
		return nil, err
	}

	host, err := httpClient.GetHost(hostName)
	if err != nil {
		// HACK. api.go should decode JSON and return a specific type of error for this case.
		if strings.Contains(err.Error(), "Not found") {
			humanName := GetHumanHostName(hostName)
			return nil, fmt.Errorf("%s doesn't seem to be running.\nYou can create it with `orchard hosts create %s`.", utils.Capitalize(humanName), hostName)
		}

		return nil, err
	}

	return host, nil
}
