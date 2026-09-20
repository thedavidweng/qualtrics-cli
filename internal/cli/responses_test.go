package cli

import (
	"archive/zip"
	"bytes"
	"os"
	"path/filepath"
	"testing"
)

func buildZip(t *testing.T, entries map[string]string) []byte {
	t.Helper()
	var buf bytes.Buffer
	w := zip.NewWriter(&buf)
	for name, content := range entries {
		f, err := w.Create(name)
		if err != nil {
			t.Fatal(err)
		}
		if _, err := f.Write([]byte(content)); err != nil {
			t.Fatal(err)
		}
	}
	if err := w.Close(); err != nil {
		t.Fatal(err)
	}
	return buf.Bytes()
}

func TestExtractZip(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "export.zip")
	if err := os.WriteFile(path, buildZip(t, map[string]string{
		"responses.csv":    "ResponseId,Q1\nR_1,2\n",
		"nested/inner.csv": "a,b\n",
	}), 0o600); err != nil {
		t.Fatal(err)
	}

	if err := extractZip(path); err != nil {
		t.Fatal(err)
	}

	out := filepath.Join(dir, "export_extracted", "responses.csv")
	data, err := os.ReadFile(out)
	if err != nil {
		t.Fatal(err)
	}
	if string(data) != "ResponseId,Q1\nR_1,2\n" {
		t.Fatalf("content = %q", data)
	}
	nested := filepath.Join(dir, "export_extracted", "nested", "inner.csv")
	if _, err := os.Stat(nested); err != nil {
		t.Fatalf("nested entry missing: %v", err)
	}
}

func TestExtractZipRejectsEscape(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "evil.zip")
	var buf bytes.Buffer
	w := zip.NewWriter(&buf)
	f, err := w.CreateHeader(&zip.FileHeader{Name: "../escaped.txt"})
	if err != nil {
		t.Fatal(err)
	}
	if _, err := f.Write([]byte("nope")); err != nil {
		t.Fatal(err)
	}
	if err := w.Close(); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, buf.Bytes(), 0o600); err != nil {
		t.Fatal(err)
	}

	if err := extractZip(path); err == nil {
		t.Fatal("expected zip-slip rejection")
	}
}
