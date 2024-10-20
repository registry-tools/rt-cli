package main

import (
	"log"
	"os"

	"github.com/hashicorp/cli"
	"github.com/registry-tools/rt-cli/internal/commands"
	"github.com/registry-tools/rt-cli/version"
)

func main() {
	c := cli.NewCLI("rt", version.Version)
	c.Args = os.Args[1:]
	c.Commands = map[string]cli.CommandFactory{
		"publish": commands.PublishCommandFactory,
		"gha":     commands.GHACommandFactory,
		"login":   commands.LoginCommandFactory,
	}

	c.HiddenCommands = []string{"gha"}

	exitStatus, err := c.Run()
	if err != nil {
		log.Printf("Error: %v", err)
	}

	os.Exit(exitStatus)
}
