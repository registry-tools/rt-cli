package publish_test

import (
	"context"
	"fmt"
	"io"
	"mime"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"testing"

	"github.com/registry-tools/publish/internal/module"
	"github.com/registry-tools/publish/internal/publish"
	sdk "github.com/registry-tools/rt-sdk"
)

var uploadBlobCalled = false

func newTestServer(t *testing.T) *httptest.Server {
	t.Helper()

	mux := http.NewServeMux()

	mux.HandleFunc("/auth/token", func(res http.ResponseWriter, req *http.Request) {
		res.Header().Set("Content-Type", "application/json")
		_, err := res.Write([]byte(`{"access_token": "test-token","expires_in": 3600}`))
		if err != nil {
			t.Fatalf("Failed to write response: %s", err)
		}
	})

	mux.HandleFunc("/.well-known/terraform.json", func(res http.ResponseWriter, req *http.Request) {
		t.Fatal("here")
		res.Header().Set("Content-Type", "application/json")
		_, err := res.Write([]byte(`{"rt.v1": "/"}`))
		if err != nil {
			t.Fatalf("Failed to write response: %s", err)
		}
	})

	mux.HandleFunc("/upload-blob", func(res http.ResponseWriter, req *http.Request) {
		file, err := os.CreateTemp("", "test-uploaded")
		t.Cleanup(func() {
			file.Close()
			if os.Getenv("KEEP_ARCHIVE") == "true" {
				t.Logf("KEEP_ARCHIVE enabled. Archive stored at %q", file.Name())
			} else {
				os.Remove(file.Name())
			}
		})
		if err != nil {
			t.Fatalf("Could not create upload file: %s", err)
		}

		copied, err := io.Copy(file, req.Body)
		if err != nil {
			t.Fatalf("Failed to download request body to file: %s", err)
		}

		t.Logf("Uploaded %d bytes to %q", copied, file.Name())
		t.Logf("Request Headers: %v", req.Header)

		if req.Header.Get("Content-Type") != "application/gzip" {
			t.Errorf("Expected Content-Type to be application/gzip, got %q", req.Header.Get("Content-Type"))
		}

		uploadBlobCalled = true
		res.WriteHeader(http.StatusNoContent)
	})

	mux.HandleFunc("/api/archives", func(res http.ResponseWriter, req *http.Request) {
		res.Header().Set("Content-Type", "application/json")
		res.WriteHeader(http.StatusCreated)
		res.Write([]byte(fmt.Sprintf(`{"data": {"signed-id": "123","filename":"slug.tar.gz"}, "meta":{"upload-url": "https://%s/upload-blob", "headers": {"Content-Type":"application/gzip"}}}`, req.Host)))
	})

	mux.HandleFunc("/api/terraform-module-versions", func(res http.ResponseWriter, req *http.Request) {
		if req.Header.Get("Authorization") != "Bearer insecure-testing-only" {
			t.Fatalf("Expected Authorization header to be set with Bearer insecure-testing-only, but was %q", req.Header.Get("Authorization"))
		}

		mimeType, _, err := mime.ParseMediaType(req.Header.Get("Content-Type"))
		if err != nil {
			t.Fatalf("Expected no error, got %q", err)
		}
		if mimeType != "application/json" {
			t.Fatalf("Expected Content-Type to be application/json, got %q", mimeType)
		}

		if req.Method != "POST" {
			t.Fatalf("Expected method to be POST, got %q", req.Method)
		}

		res.Header().Set("Content-Type", "application/json")
		res.WriteHeader(http.StatusCreated)
		_, err = res.Write([]byte(`{"data":{"id":"mv-1234","name":"moduleB","system":"null","version":"1.2.3","namespace":"my-example-org"}}`))
		if err != nil {
			t.Fatalf("Failed to write response: %s", err)
		}
	})

	server := httptest.NewTLSServer(mux)
	t.Cleanup(func() {
		server.Close()
	})
	return server
}

func TestPackAndPublish(t *testing.T) {
	server := newTestServer(t)
	serverURL, err := url.Parse(server.URL)

	t.Logf("Server URL: %s", serverURL)

	if err != nil {
		t.Fatalf("Failed to parse server URL: %s", err)
	}

	sdk, err := sdk.NewInsecureSDKForTesting(fmt.Sprintf("%s:%s", serverURL.Hostname(), serverURL.Port()))
	if err != nil {
		t.Fatalf("Failed to create SDK client: %s", err)
	}

	publisher := publish.Publisher{
		SDK: sdk,
	}

	path, _, err := publish.PackAsFile("./fixtures/moduleA")
	t.Cleanup(func() {
		if path != "" {
			os.Remove(path)
		}
	})
	if err != nil {
		t.Fatalf("Failed to pack directory: %s", err)
	}

	file, err := os.Open(path)
	if err != nil {
		t.Fatalf("Failed to open packed file: %s", err)
	}

	publishedModule, err := publisher.Publish(context.TODO(), module.Module{
		Namespace: "my-example-org",
		Name:      "moduleB",
		System:    "null",
		Version:   "1.2.3",
	}, file)

	if err != nil {
		t.Fatalf("Failed to publish module: %s", err)
	}

	if !uploadBlobCalled {
		t.Error("Expected upload-blob to be called, but it was not")
	}

	if publishedModule.Name != "moduleB" {
		t.Errorf("Expected module name to be \"moduleB\", got %q", publishedModule.Name)
	}

	if publishedModule.System != "null" {
		t.Errorf("Expected module system to be \"null\", got %q", publishedModule.System)
	}

	if publishedModule.Version != "1.2.3" {
		t.Errorf("Expected module version to be \"1.2.3\", got %q", publishedModule.Version)
	}

	if publishedModule.Namespace != "my-example-org" {
		t.Errorf("Expected module namespace to be \"my-example-org\", got %q", publishedModule.Namespace)
	}

	if err != nil {
		t.Fatalf("Failed to publish module: %s", err)
	}
}
