// Package main is the entry point for the refreSSH CLI application.
package main

import (
	"fmt"
	"os"

	"github.com/strickdd/refressh/internal/cli"
)

func main() {
	if err := cli.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err) //nolint:errcheck
		os.Exit(1)
	}
}
