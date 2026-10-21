package avro

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"sync/atomic"
	"testing"
)

const testSchema = `{"type":"record","name":"test","fields":[{"name":"id","type":"string"}]}`

func TestSchemaSourceResolveFromFile(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	schemaFile := filepath.Join(dir, "test.avsc")
	if err := os.WriteFile(schemaFile, []byte(testSchema), 0o600); err != nil {
		t.Fatalf("failed to write schema file: %v", err)
	}

	source := NewSchemaSource(Standard)

	codec, err := source.Resolve(schemaFile)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if codec == nil {
		t.Fatal("expected codec, got nil")
	}
}

func TestSchemaSourceResolveFromFileURI(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	schemaFile := filepath.Join(dir, "test.avsc")
	if err := os.WriteFile(schemaFile, []byte(testSchema), 0o600); err != nil {
		t.Fatalf("failed to write schema file: %v", err)
	}

	source := NewSchemaSource(Standard)

	codec, err := source.Resolve((&url.URL{Scheme: "file", Path: filepath.ToSlash(schemaFile)}).String())
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if codec == nil {
		t.Fatal("expected codec, got nil")
	}
}

func TestSchemaSourceResolveFromFileURIWithRemoteHost(t *testing.T) {
	t.Parallel()

	source := NewSchemaSource(Standard)

	if _, err := source.Resolve("file://remote-host/schemas/test.avsc"); err == nil {
		t.Fatal("expected error for file uri with remote host, got nil")
	}
}

func TestIsSchemaURI(t *testing.T) {
	t.Parallel()

	for location, expected := range map[string]bool{
		"http://example.com/schema.avsc":  true,
		"https://example.com/schema.avsc": true,
		"file:///tmp/schema.avsc":         true,
		"/tmp/schema.avsc":                false,
		"schema.avsc":                     false,
		"urn:example:schema":              false,
	} {
		if got := IsSchemaURI(location); got != expected {
			t.Errorf("IsSchemaURI(%q) = %v, expected %v", location, got, expected)
		}
	}
}

func TestSchemaSourceResolveFromURL(t *testing.T) {
	t.Parallel()

	var requestCount int32

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		atomic.AddInt32(&requestCount, 1)
		_, _ = w.Write([]byte(testSchema))
	}))
	defer server.Close()

	source := NewSchemaSource(Standard)

	codec, err := source.Resolve(server.URL)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if codec == nil {
		t.Fatal("expected codec, got nil")
	}

	if _, err := source.Resolve(server.URL); err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if got := atomic.LoadInt32(&requestCount); got != 1 {
		t.Fatalf("expected schema to be fetched once and then served from cache, got %d requests", got)
	}
}

func TestSchemaSourceResolveFromURLNotFound(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusNotFound)
	}))
	defer server.Close()

	source := NewSchemaSource(Standard)

	if _, err := source.Resolve(server.URL); err == nil {
		t.Fatal("expected error for 404 response, got nil")
	}
}

func TestSchemaSourceResolveInvalidSchema(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	schemaFile := filepath.Join(dir, "invalid.avsc")
	if err := os.WriteFile(schemaFile, []byte("not a schema"), 0o600); err != nil {
		t.Fatalf("failed to write schema file: %v", err)
	}

	source := NewSchemaSource(Standard)

	if _, err := source.Resolve(schemaFile); err == nil {
		t.Fatal("expected error for invalid schema, got nil")
	}
}

func TestSchemaSourceCachesFailures(t *testing.T) {
	t.Parallel()

	var requestCount int32

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		atomic.AddInt32(&requestCount, 1)
		w.WriteHeader(http.StatusNotFound)
	}))
	defer server.Close()

	source := NewSchemaSource(Standard)

	for i := range 3 {
		if _, err := source.Resolve(server.URL); err == nil {
			t.Fatalf("attempt %d: expected error for 404 response, got nil", i)
		}
	}

	// an unresolvable location is retried once per message otherwise, each time waiting for the timeout
	if got := atomic.LoadInt32(&requestCount); got != 1 {
		t.Fatalf("expected the failure to be cached after one request, got %d requests", got)
	}
}

func TestSchemaSourceRejectsOversizedSchema(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		for written := 0; written <= maxSchemaBytes; written += 1024 {
			if _, err := w.Write(make([]byte, 1024)); err != nil {
				return
			}
		}
	}))
	defer server.Close()

	source := NewSchemaSource(Standard)

	_, err := source.Resolve(server.URL)
	if err == nil {
		t.Fatal("expected error for an oversized schema, got nil")
	}
	if !strings.Contains(err.Error(), "larger than the maximum") {
		t.Fatalf("expected a size limit error, got: %v", err)
	}
}
