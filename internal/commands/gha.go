package commands

import (
	"context"
	"errors"
	"log"
	"os"
	"strings"

	"github.com/hashicorp/cli"
	svchost "github.com/hashicorp/terraform-svchost"
	"github.com/registry-tools/publish/internal/publish"
	sdk "github.com/registry-tools/rt-sdk"
	"github.com/sethvargo/go-githubactions"
)

func GHACommandFactory() (cli.Command, error) {
	return &ghaCommand{}, nil
}

type ghaCommand struct{}

func (c *ghaCommand) Help() string {
	return "This text should not be displayed."
}

func (c *ghaCommand) sdkFromAction() (sdk.SDK, error) {
	hostName := os.Getenv("REGISTRY_TOOLS_HOSTNAME")
	if hostName == "" {
		hostName = DefaultHostname
	}

	envToken := os.Getenv("REGISTRY_TOOLS_TOKEN")
	if envToken == "" {
		return nil, errors.New("REGISTRY_TOOLS_TOKEN must be set")
	}

	return sdk.NewSDKWithAccessToken(hostName, envToken)
}

// ModuleArgsFromAction returns a ModuleArgs from GitHub Actions inputs. If
// any input is missing or invalid, this function will perform a fatal program exit.
func ModuleArgsFromAction() (*ModuleArgs, error) {
	moduleName := githubactions.GetInput("module")
	if moduleName == "" {
		var ok bool
		moduleName, ok = os.LookupEnv("GITHUB_REPOSITORY")
		if ok {
			_, moduleName, ok = strings.Cut(moduleName, "/")
			if !ok {
				return nil, errors.New("expected organization/repository format for GITHUB_REPOSITORY")
			}
		} else {
			return nil, errors.New("module input is required")
		}
	}

	system := githubactions.GetInput("system")
	if system == "" {
		system = "null"
	}

	version := githubactions.GetInput("version")
	if version == "" {
		return nil, errors.New("version input is required")
	}

	directory := githubactions.GetInput("directory")
	if directory == "" {
		directory = "."
	}

	namespace := githubactions.GetInput("namespace")
	if namespace == "" {
		return nil, errors.New("namespace input is required")
	}

	return &ModuleArgs{
		Namespace: namespace,
		Name:      moduleName,
		System:    system,
		Version:   version,
		Directory: directory,
	}, nil
}

func (c *ghaCommand) Run(args []string) int {
	// Application can be run as a GitHub Action or as a normal executable
	var isGitHubAction = os.Getenv("GITHUB_ACTIONS") == "true"
	if !isGitHubAction {
		log.Printf("[ERROR] This command can only be run as a GitHub Action")
		return 1
	}

	var ma, err = ModuleArgsFromAction()
	if err != nil {
		log.Printf("[ERROR] Failed to fetch required input arguments: %s", err)
		return 1
	}

	sdkclient, err := c.sdkFromAction()
	if err != nil {
		log.Printf("[ERROR] Failed to create SDK client: %s", err)
		return 127
	}

	hostname, err := svchost.ForComparison(sdkclient.Endpoint().Host)
	if err != nil {
		log.Printf("[ERROR] Failed to parse hostname: %s", err)
		return 127
	}

	// Pack the source directory into a temporary file
	path, size, err := publish.PackAsFile(ma.Directory)
	if err != nil {
		log.Printf("[ERROR] Failed to pack directory %q: %s", ma.Directory, err)
		return 2
	}
	defer os.Remove(path)

	file, err := os.Open(path)
	if err != nil {
		log.Printf("[ERROR] Failed to open archive file: %s", err)
		return 2
	}
	defer file.Close()

	summary, err := publishModuleArchive(context.TODO(), file, size, sdkclient, *ma)
	if err != nil {
		log.Printf("[ERROR] Failed to publish module: %s", err)
		return 1
	}

	githubactions.SetOutput("source", ma.Module().Source(hostname))

	html, err := summary.HTML()
	if err != nil {
		log.Printf("[ERROR] Module was published successfully, but this program failed to generate a summary: %s", err)
	} else {
		githubactions.AddStepSummary(html)
	}

	return 0
}

func (c *ghaCommand) Synopsis() string {
	return "This text should not be displayed."
}
