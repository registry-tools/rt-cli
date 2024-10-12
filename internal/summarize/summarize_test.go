package summarize

import (
	"testing"

	"github.com/andreyvit/diff"
	svchost "github.com/hashicorp/terraform-svchost"
	"github.com/registry-tools/publish/internal/publish"
)

func TestHumanizeBytes(t *testing.T) {
	items := map[int64]string{
		1024:       "1 kB",
		884:        "884 B",
		9999:       "9 kB",
		7145859:    "7 MB",
		91957860:   "91 MB",
		188596999:  "188 MB",
		9992010501: "9 GB",
	}

	for size, result := range items {
		if actual := HumanizeBytes(size); actual != result {
			t.Errorf("expected %d to be %q, but was %q", size, result, actual)
		}
	}
}

func TestSummary(t *testing.T) {
	mod := publish.ModuleVersion{
		Name:      "computer",
		System:    "aws",
		Version:   "1.0.0",
		Namespace: "spacepioneer",
	}

	summary := NewSummary(1024, svchost.Hostname("registrytools.cloud"), &mod)
	output, err := summary.HTML()
	if err != nil {
		t.Errorf("expected no error, got %s", err)
	}

	expected := `
<h3>Module Published</h3>

Example Usage:
<pre>module &#34;computer&#34; {
  source  = &#34;registrytools.cloud/spacepioneer/computer/aws&#34;
  version = &#34;1.0.0&#34;
}
</pre>

<p>Consume this module by <a target="_blank" href="https://registrytools.cloud/provision">authenticating to the registry</a> using this ENV variable when running terraform:</p>

<code>TF_TOKEN_registrytools_cloud=&lt;token&gt;</code>

<p>The size of this module, gzipped, was 1 kB</p>


`

	if output != expected {
		t.Errorf("unexpected summary HTML:\n%v", diff.LineDiff(output, expected))
	}
}
