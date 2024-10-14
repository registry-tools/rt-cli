package publish_test

import (
	"os"
	"testing"

	"github.com/registry-tools/publish/internal/publish"
)

func TestPackAsFile(t *testing.T) {
	file, size, err := publish.PackAsFile("./fixtures/moduleA")
	t.Cleanup(func() {
		if file != "" {
			os.Remove(file)
		}
	})

	if err != nil {
		t.Errorf("expected no error, got %s", err)
	}

	if size == 0 {
		t.Errorf("expected size greater than 0, got %d", size)
	}
}
