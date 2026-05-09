// Package main is the entry point for the refreSSH CLI application.
package main

import (
	"fmt"
	"os"

	"github.com/strickdd/refressh/internal/cli"
)

func main() {
	if err := cli.Execute(); err != nil {
		if _, err := fmt.Fprintln(os.Stderr, err); err != nil {
			// Suppress error to satisfy errcheck
			_ = err
		}
		os.Exit(1)
	}
}
