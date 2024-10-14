package summarize

import (
	"bytes"
	"fmt"
	"html/template"
	"strings"

	"github.com/fatih/color"
	svchost "github.com/hashicorp/terraform-svchost"
	"github.com/registry-tools/publish/internal/module"
	"github.com/registry-tools/publish/internal/publish"
)

func HumanizeBytes(i int64) string {
	if i > 1_000_000_000 {
		return fmt.Sprintf("%d GB", i/1_000_000_000)
	}
	if i > 1_000_000 {
		return fmt.Sprintf("%d MB", i/1_000_000)
	}
	if i > 1_000 {
		return fmt.Sprintf("%d kB", i/1_000)
	}
	return fmt.Sprintf("%d B", i)
}

var tmplHTML = `
<h3>Module Published</h3>

Example Usage:
<pre>{{.TerraformExample}}</pre>

<p>Consume this module by <a target="_blank" href="{{.ProvisionURL}}">authenticating to the registry</a> using this ENV variable when running terraform:</p>

<code>{{.TFTokenExample}}</code>

<p>The size of this module, gzipped, was {{.SizeHuman}}</p>

{{if gt .Size 1000000}}
<p>This seems to be extraordinarily large for an IaC module. Use a .terraformignore file to exclude files that aren't needed by the module.</p>
{{end}}
`

type templateData struct {
	Size             int64
	SizeHuman        string
	TerraformExample string
	TFTokenExample   string
	ProvisionURL     string
}

type Summary struct {
	Size      int64
	Namespace string
	Module    *publish.ModuleVersion
	Host      svchost.Hostname
}

func NewSummary(size int64, host svchost.Hostname, mod *publish.ModuleVersion) Summary {
	return Summary{
		Size:      size,
		Namespace: mod.Namespace,
		Module:    mod,
		Host:      host,
	}
}

func (s Summary) module() module.Module {
	return s.Module.Module(s.Namespace)
}

func (s Summary) getTemplateData() templateData {
	hostEscaped := strings.ReplaceAll(s.Host.String(), "-", "__")
	hostEscaped = strings.ReplaceAll(hostEscaped, ".", "_")

	mod := s.module()

	return templateData{
		Size:             s.Size,
		SizeHuman:        HumanizeBytes(s.Size),
		TerraformExample: mod.ToTerraformExample(s.Host),
		TFTokenExample:   fmt.Sprintf("TF_TOKEN_%s=<token>", hostEscaped),
		ProvisionURL:     fmt.Sprintf("https://%s/provision", s.Host.String()),
	}
}

func (s Summary) HTML() (string, error) {
	data := s.getTemplateData()
	t, err := template.New("").Parse(tmplHTML)
	if err != nil {
		return "", fmt.Errorf("failed to parse template: %w", err)
	}

	var b bytes.Buffer
	if err := t.Execute(&b, data); err != nil {
		return "", fmt.Errorf("failed to execute template: %w", err)
	}

	return b.String(), nil
}

func indent(spaces int, s string) string {
	pad := strings.Repeat(" ", spaces)
	return pad + strings.Replace(s, "\n", "\n"+pad, -1)
}

func (s Summary) CLI() string {
	result := strings.Builder{}

	codeSample := color.New(color.FgWhite, color.Italic)
	success := color.New(color.FgGreen)
	warning := color.New(color.FgYellow, color.Bold)

	result.WriteString(success.Sprint("Module published successfully."))
	result.WriteString("\n\nExample Usage:\n\n")

	data := s.getTemplateData()
	result.WriteString(indent(2, codeSample.Sprint(data.TerraformExample)))

	if data.Size > 1000000 {
		result.WriteString(warning.Sprintf("The size of this module, gzipped, was %s. This seems to be extraordinarily large for an IaC module. Use a .terraformignore file to exclude files that aren't needed by the module.", data.SizeHuman))
	}

	result.WriteString("\n")
	return result.String()
}
