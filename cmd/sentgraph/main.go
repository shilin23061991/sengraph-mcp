// Command sentgraph is a memory MCP server backed by Zep Cloud.
//
// It runs in three modes:
//
//	sentgraph serve [--http ADDR]   run the MCP server (stdio by default)
//	sentgraph hook <event>          handle a Claude Code lifecycle hook (reads JSON from stdin)
//	sentgraph doctor                check configuration and Zep connectivity
package main

import (
	"context"
	"fmt"
	"os"
)

const usage = `sentgraph - memory MCP server backed by Zep Cloud

Usage:
  sentgraph serve [--http ADDR]   Run the MCP server (stdio by default)
  sentgraph hook <event>          Handle a Claude Code lifecycle hook (reads JSON from stdin)
  sentgraph doctor                Check configuration and Zep connectivity
`

func main() {
	if len(os.Args) < 2 {
		fmt.Fprint(os.Stderr, usage)
		os.Exit(2)
	}

	ctx := context.Background()
	var err error
	switch os.Args[1] {
	case "serve":
		err = runServe(ctx, os.Args[2:])
	case "hook":
		err = runHook(ctx, os.Args[2:])
	case "doctor":
		err = runDoctor(ctx, os.Args[2:])
	case "-h", "--help", "help":
		fmt.Print(usage)
		return
	default:
		fmt.Fprintf(os.Stderr, "unknown command %q\n\n%s", os.Args[1], usage)
		os.Exit(2)
	}

	if err != nil {
		fmt.Fprintln(os.Stderr, "sentgraph:", err)
		os.Exit(1)
	}
}

func errNotImplemented(what string) error {
	return fmt.Errorf("%s: not implemented yet", what)
}
