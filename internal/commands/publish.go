package commands

import (
	"context"
	"flag"
	"fmt"
	"io"
	"log"
	"os"
	"path"
	"strings"

	"github.com/fatih/color"
	"github.com/hashicorp/cli"
	svchost "github.com/hashicorp/terraform-svchost"
	sdk "github.com/registry-tools/rt-sdk"

	"github.com/registry-tools/rt-cli/internal/module"
	"github.com/registry-tools/rt-cli/internal/publish"
	"github.com/registry-tools/rt-cli/internal/summarize"
)

// DefaultHostname is the default hostname for Registry Tools Cloud
const DefaultHostname = "registrytools.cloud"

func PublishCommandFactory() (cli.Command, error) {
	return &publishCommand{}, nil
}

type publishCommand struct {
}

type ModuleArgs struct {
	Namespace string
	Version   string
	Name      string
	System    string
	Directory string
}

func (m ModuleArgs) Module() module.Module {
	return module.Module{
		Namespace: m.Namespace,
		Name:      m.Name,
		System:    m.System,
		Version:   m.Version,
	}
}

func (c *publishCommand) Help() string {
	return `
Usage: rt publish [options] <source_directory>

  Publish a module to the registry.

Options:

  --namespace=<namespace>  (Required) The namespace of the module. This is the
                           first part of the path to a moodule, Ex: "platform".

  --version=<version>      (Required) The version of the module, Ex: "2.1.0".
  
  --name=<name>            The name of the module. Defaults to being derived from
                           the source directory. This is the second part of the
                           path to a module. Ex: "networking".

  --system=<provider>      The provider system associated with the module. Defaults
                           to being derived from the source directory. Ex: "aws".

  --directory=<dir>        The directory containing the module source code. Defaults
                           to the current directory.
`
}

func publishModuleArchive(ctx context.Context, reader io.ReadSeeker, size int64, sdkclient sdk.SDK, margs ModuleArgs) (*summarize.Summary, error) {
	// Publish the module and summarize the result
	publisher := publish.Publisher{
		SDK: sdkclient,
	}

	host, err := svchost.ForComparison(sdkclient.Endpoint().Host)
	if err != nil {
		return nil, fmt.Errorf("invalid host: %w", err)
	}

	log.Printf("[INFO] Publishing module \"%s/%s/%s\" version %q", host, margs.Name, margs.System, margs.Version)

	ver, err := publisher.Publish(ctx, margs.Module(), reader)
	if err != nil {
		return nil, err
	}

	result := summarize.NewSummary(size, host, ver)
	return &result, nil
}

func moduleArgsFromCWD() (*ModuleArgs, error) {
	pwd, err := os.Getwd()
	if err != nil {
		return nil, fmt.Errorf("failed to get current working directory: %w", err)
	}

	base := path.Base(pwd)
	name := base
	system := "null"

	if strings.HasPrefix(base, "terraform-") {
		parts := strings.SplitN(base, "-", 3)
		if len(parts) > 2 {
			system = parts[1]
			name = parts[2]
		} else {
			name = parts[1]
		}
	}

	return &ModuleArgs{
		Directory: pwd,
		Name:      name,
		System:    system,
	}, nil
}

func (c *publishCommand) sdkFromEnvironment() (sdk.SDK, error) {
	host := os.Getenv("REGISTRY_TOOLS_HOSTNAME")
	if host == "" {
		host = DefaultHostname
	}

	var result sdk.SDK
	var err error
	if envToken := os.Getenv("REGISTRY_TOOLS_TOKEN"); envToken != "" {
		result, err = sdk.NewSDKWithAccessToken(host, envToken)
	} else if envClientID := os.Getenv("REGISTRY_TOOLS_CLIENT_ID"); envClientID != "" {
		result, err = sdk.NewSDK(host, envClientID, os.Getenv("REGISTRY_TOOLS_CLIENT_SECRET"))
	} else {
		err = ErrLoginRequired
	}

	return result, err
}

func (c *publishCommand) requireArgumentOrExit(name, value string) {
	if value == "" {
		log.Printf("[ERROR] Required argument %q is missing", name)
		os.Exit(1)
	}
}

func (c *publishCommand) Run(args []string) int {
	f := flag.NewFlagSet("", flag.ExitOnError)
	f.SetOutput(io.Discard)
	f.Usage = func() {}

	var ma ModuleArgs
	defaults, err := moduleArgsFromCWD()
	if err != nil {
		log.Printf("[ERROR] Failed to read current working directory: %s", err)
		return 1
	}

	f.StringVar(&ma.Namespace, "namespace", "", "")
	f.StringVar(&ma.Version, "version", "", "")
	f.StringVar(&ma.Name, "name", defaults.Name, "")
	f.StringVar(&ma.System, "system", defaults.System, "")
	f.StringVar(&ma.Directory, "directory", defaults.Directory, "")

	if err := f.Parse(args); err != nil {
		return 1
	}

	c.requireArgumentOrExit("namespace", ma.Namespace)
	c.requireArgumentOrExit("version", ma.Version)
	c.requireArgumentOrExit("name", ma.Name)
	c.requireArgumentOrExit("system", ma.System)

	// Env config
	sdkclient, err := c.sdkFromEnvironment()
	if err != nil {
		log.Printf("[ERROR] Failed to create SDK client: %s", err)
		return 127
	}

	// Pack the source directory into a temporary file
	path, size, err := publish.PackAsFile(ma.Directory)
	if err != nil {
		log.Printf("[ERROR] Failed to pack directory %q: %s", ma.Directory, err)
		return 2
	}
	defer os.Remove(path)

	info, err := os.Stat(path)
	if err != nil {
		log.Printf("[ERROR] Failed to stat archive file: %s", err)
		return 2
	}

	if !c.confirm(size, info.Size(), ma, svchost.ForDisplay(sdkclient.Endpoint().Host)) {
		log.Printf("[ERROR] User did not confirm")
		return 1
	}

	file, err := os.Open(path)
	if err != nil {
		log.Printf("[ERROR] Failed to open archive file: %s", err)
		return 2
	}
	defer file.Close()

	ctx := context.Background()
	summary, err := publishModuleArchive(ctx, file, size, sdkclient, ma)
	if err != nil {
		log.Printf("[ERROR] Failed to publish module: %s", err)
		return 1
	}

	fmt.Print(summary.CLI())

	return 0
}

func (c *publishCommand) confirm(size int64, sizeCompressed int64, ma ModuleArgs, hostnameForDisplay string) bool {
	baseUI := &cli.BasicUi{
		Reader:      os.Stdin,
		Writer:      os.Stdout,
		ErrorWriter: os.Stderr,
	}

	label := color.New(color.FgCyan, color.Faint)
	value := color.New(color.FgCyan, color.Bold)
	host := color.New(color.FgYellow, color.Bold)

	cwd, err := os.Getwd()
	if err != nil {
		log.Printf("[ERROR] Failed to get current working directory: %s", err)
		return false
	}

	label.Print("Version:   ")
	value.Println(ma.Version)
	label.Print("Name:      ")
	value.Println(ma.Name)
	label.Print("System:    ")
	value.Println(ma.System)
	label.Print("Directory: ")
	if ma.Directory != cwd {
		value.Println(ma.Directory)
	} else {
		value.Println(".")
	}
	label.Print("Size:      ")
	value.Print(summarize.HumanizeBytes(size))
	value.Println(fmt.Sprintf(" (%s compressed)", summarize.HumanizeBytes(sizeCompressed)))

	answer, err := baseUI.Ask(color.YellowString(fmt.Sprintf("Publish to %s? You must type 'yes' to confirm:", host.Sprint(hostnameForDisplay))))
	if err != nil {
		log.Printf("[ERROR] cannot to read user input: %s", err)
		return false
	}

	return answer == "yes"
}

func (c *publishCommand) Synopsis() string {
	return "Publish a module to the registry"
}
