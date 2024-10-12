package module

import (
	"fmt"
	"strings"

	svchost "github.com/hashicorp/terraform-svchost"
)

// Module represents information about a module to be published to a particular
// registry from a particular source directory.
type Module struct {
	Namespace string
	Name      string
	System    string
	Version   string
}

// ToTerraformExample returns a string that shows the caller usage of the module.
func (m Module) ToTerraformExample(hostname svchost.Hostname) string {
	return fmt.Sprintf(`module "%s" {
  source  = "%s"
  version = "%s"
}
`, m.Name, m.Source(hostname), m.Version)
}

// Source returns the source string for the module.
func (m Module) Source(hostname svchost.Hostname) string {
	return strings.Join([]string{hostname.String(), m.Namespace, m.Name, m.System}, "/")
}
