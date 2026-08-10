package avro

import (
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
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
