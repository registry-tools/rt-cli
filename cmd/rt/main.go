package main

import (
	"log"
	"os"

	"github.com/hashicorp/cli"
	"github.com/registry-tools/publish/internal/commands"
)

var version = "0.1.0-dev"

func main() {
	c := cli.NewCLI("rt", version)
	c.Args = os.Args[1:]
	c.Commands = map[string]cli.CommandFactory{
		"publish": commands.PublishCommandFactory,
		"gha":     commands.GHACommandFactory,
	}

	c.HiddenCommands = []string{"gha"}

	exitStatus, err := c.Run()
	if err != nil {
		log.Printf("Error: %v", err)
	}

	os.Exit(exitStatus)
}
