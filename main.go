package main

import (
	"flag"
	"fmt"
	"github.com/orchardup/orchard/commands"
	"io"
	"os"
	"strings"
	"text/template"
)

func main() {
	flag.Usage = usage
	flag.Parse()

	args := flag.Args()
	if len(args) < 1 {
		usage()
	}

	if args[0] == "help" {
		help(args[1:])
		return
	}

	for _, cmd := range commands.All {
		if cmd.Name() == args[0] {
			cmd.Flag.Usage = func() { cmd.Usage() }
			cmd.Flag.Parse(args[1:])
			args = cmd.Flag.Args()
			err := cmd.Run(cmd, args)
			if err != nil {
				fmt.Fprintln(os.Stderr, err)
				os.Exit(1)
			}
			return
		}
	}

	fmt.Fprintf(os.Stderr, "Unknown subcommand: %q\n\n", args[0])
	usage()
}

var usageTemplate = `Orchard command-line client.

Usage: orchard COMMAND [ARG...]

Commands:
{{range .}}
  {{.Name | printf "%-11s"}} {{.Short}}{{end}}

Run 'orchard help command' for more information on a command.
`

var helpTemplate = `Usage: orchard {{.UsageLine}}

{{.Long | trim}}
`

func tmpl(w io.Writer, text string, data interface{}) {
	t := template.New("top")
	t.Funcs(template.FuncMap{"trim": strings.TrimSpace})
	template.Must(t.Parse(text))
	if err := t.Execute(w, data); err != nil {
		panic(err)
	}
}

func printUsage(w io.Writer) {
	tmpl(w, usageTemplate, commands.All)
}

func usage() {
	printUsage(os.Stderr)
	os.Exit(2)
}

func help(args []string) {
	if len(args) == 0 {
		printUsage(os.Stdout)
		return
	}

	arg := args[0]

	for _, cmd := range commands.All {
		if cmd.Name() == arg {
			tmpl(os.Stdout, helpTemplate, cmd)
			return
		}
	}

	fmt.Fprintf(os.Stderr, "Unknown help topic %#q.\n\n", arg)
	printUsage(os.Stderr)
	os.Exit(2)
}
