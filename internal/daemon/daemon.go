package daemon

import (
	"fmt"
	"github.com/strickdd/refressh/internal/api"
)

type Daemon struct {
	// State management will go here
}

func New() *Daemon {
	return &Daemon{}
}

func (d *Daemon) Start() error {
	fmt.Println("Daemon starting...")
	// TODO: Initialize PTY manager
	// TODO: Start API server
	return api.Start()
}
