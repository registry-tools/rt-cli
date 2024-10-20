package commands

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"github.com/cli/oauth"
	"github.com/fatih/color"
	"github.com/hashicorp/cli"

	userconfig "github.com/registry-tools/rt-cli/internal/userconfig"
	"github.com/registry-tools/rt-cli/version"
)

func LoginCommandFactory() (cli.Command, error) {
	return &loginCommand{}, nil
}

type loginCommand struct{}

func (c *loginCommand) Help() string {
	return `
Usage: rt login [hostname]

  Login to a Registry Tools private registry. Optionally, provide a hostname
	to login to. If none is given, defaults to "registrytools.cloud".
`
}

func (c *loginCommand) Run(args []string) int {
	hostname := DefaultHostname
	if len(args) == 1 {
		hostname = args[0]
	}

	hostname = strings.TrimPrefix(hostname, "https://")
	hostname = strings.TrimPrefix(hostname, "http://")

	host := oauth.Host{
		DeviceCodeURL: "https://" + hostname + "/auth/device/code",
		TokenURL:      "https://" + hostname + "/auth/token",
		AuthorizeURL:  "https://" + hostname + "/login",
	}

	colorWarn := color.New(color.FgHiYellow, color.Bold)
	colorErr := color.New(color.FgRed, color.Bold)
	colorSuccess := color.New(color.FgCyan, color.Faint)

	colorSuccess.Printf("Logging in to %s...\n", hostname)

	flow := oauth.Flow{
		Host:     &host,
		Scopes:   []string{"owner"},
		ClientID: "rt-cli",
		DisplayCode: func(code string, url string) error {
			fmt.Print("First, copy your one-time code: ")
			colorWarn.Printf("%s\n\n", code)
			fmt.Printf("Press [Enter] to continue in the web browser... ")

			scanner := bufio.NewScanner(os.Stdin)
			scanner.Scan()

			return nil
		},
	}

	accessToken, err := flow.DeviceFlow()
	if err != nil {
		colorErr.Printf("Login failed: %s\n", err)
		return 1
	}

	config, err := userconfig.LoadFromUserConfigDirectory()
	if err != nil {
		colorErr.Printf("Login failed: %s\n", err)
		return 1
	}

	config.SetHostToken(hostname, accessToken.Token)

	if err = config.SaveToUserConfigDirectory(version.Version); err != nil {
		colorErr.Printf("Login failed: %s\n", err)
		return 1
	}

	return 0
}

func (c *loginCommand) Synopsis() string {
	return "Publish a module to the registry"
}
