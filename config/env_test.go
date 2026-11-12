package config

import (
	"io"
	"os"
	"path/filepath"
	"testing"

	logging "github.com/sirupsen/logrus"
)

func TestResolveEnvironmentFileUsesCanonicalDefault(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)

	path, err := ResolveEnvironmentFile(testLogger(), "", "", "")
	if err != nil {
		t.Fatalf("ResolveEnvironmentFile() error = %v", err)
	}

	want := filepath.Join(home, ".tp", "default.env.yml")
	if path != want {
		t.Fatalf("ResolveEnvironmentFile() = %q, want %q", path, want)
	}
}

func TestResolveEnvironmentFileMigratesLegacyDefault(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)

	legacyPath := filepath.Join(home, ".tp", "default.yml")
	canonicalPath := filepath.Join(home, ".tp", "default.env.yml")
	if err := os.MkdirAll(filepath.Dir(legacyPath), 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(legacyPath, []byte("user: alice\n"), 0600); err != nil {
		t.Fatal(err)
	}

	path, err := ResolveEnvironmentFile(testLogger(), "", "", "")
	if err != nil {
		t.Fatalf("ResolveEnvironmentFile() error = %v", err)
	}
	if path != canonicalPath {
		t.Fatalf("ResolveEnvironmentFile() = %q, want %q", path, canonicalPath)
	}
	if _, err := os.Stat(legacyPath); !os.IsNotExist(err) {
		t.Fatalf("legacy default file still exists; stat error = %v", err)
	}
	if _, err := os.Stat(canonicalPath); err != nil {
		t.Fatalf("canonical default file was not created by migration: %v", err)
	}
}

func TestResolveEnvironmentFileMigratesLegacyNamedFile(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)

	legacyPath := filepath.Join(home, ".tp", "dev.yml")
	canonicalPath := filepath.Join(home, ".tp", "dev.env.yml")
	content := []byte("user: alice\n")
	if err := os.MkdirAll(filepath.Dir(legacyPath), 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(legacyPath, content, 0600); err != nil {
		t.Fatal(err)
	}

	path, err := ResolveEnvironmentFile(testLogger(), "", "dev", "")
	if err != nil {
		t.Fatalf("ResolveEnvironmentFile() error = %v", err)
	}
	if path != canonicalPath {
		t.Fatalf("ResolveEnvironmentFile() = %q, want %q", path, canonicalPath)
	}

	got, err := os.ReadFile(canonicalPath)
	if err != nil {
		t.Fatalf("reading migrated file: %v", err)
	}
	if string(got) != string(content) {
		t.Fatalf("migrated file content = %q, want %q", got, content)
	}
	if _, err := os.Stat(legacyPath); !os.IsNotExist(err) {
		t.Fatalf("legacy file still exists; stat error = %v", err)
	}
	info, err := os.Stat(canonicalPath)
	if err != nil {
		t.Fatal(err)
	}
	if gotMode := info.Mode().Perm(); gotMode != 0600 {
		t.Fatalf("migrated file mode = %04o, want %04o", gotMode, 0600)
	}
}

func TestResolveEnvironmentFilePrefersCanonicalWhenBothExist(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)

	legacyPath := filepath.Join(home, ".tp", "dev.yml")
	canonicalPath := filepath.Join(home, ".tp", "dev.env.yml")
	if err := os.MkdirAll(filepath.Dir(legacyPath), 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(legacyPath, []byte("old: value\n"), 0600); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(canonicalPath, []byte("new: value\n"), 0600); err != nil {
		t.Fatal(err)
	}

	path, err := ResolveEnvironmentFile(testLogger(), "", "dev", "")
	if err != nil {
		t.Fatalf("ResolveEnvironmentFile() error = %v", err)
	}
	if path != canonicalPath {
		t.Fatalf("ResolveEnvironmentFile() = %q, want %q", path, canonicalPath)
	}
	if _, err := os.Stat(legacyPath); err != nil {
		t.Fatalf("legacy file should be left untouched: %v", err)
	}
	got, err := os.ReadFile(canonicalPath)
	if err != nil {
		t.Fatal(err)
	}
	if string(got) != "new: value\n" {
		t.Fatalf("canonical file was changed: %q", got)
	}
}

func TestResolveEnvironmentFileFailsIfLegacyMigrationFails(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)

	legacyPath := filepath.Join(home, ".tp", "dev.yml")
	if err := os.MkdirAll(legacyPath, 0755); err != nil {
		t.Fatal(err)
	}

	path, err := ResolveEnvironmentFile(testLogger(), "", "dev", "")
	if err == nil {
		t.Fatal("ResolveEnvironmentFile() succeeded, want migration error")
	}
	if path != "" {
		t.Fatalf("ResolveEnvironmentFile() path = %q, want empty on migration error", path)
	}
	if _, err := os.Stat(legacyPath); err != nil {
		t.Fatalf("legacy path should remain after failed migration: %v", err)
	}
}

func TestResolveEnvironmentFileDoesNotMigrateConfiguredOrExplicitPaths(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)

	legacyPath := filepath.Join(home, ".tp", "vars.yml")
	if err := os.MkdirAll(filepath.Dir(legacyPath), 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(legacyPath, []byte("id: 1\n"), 0600); err != nil {
		t.Fatal(err)
	}

	path, err := ResolveEnvironmentFile(testLogger(), legacyPath, "", "")
	if err != nil {
		t.Fatalf("resolving configured path: %v", err)
	}
	if path != legacyPath {
		t.Fatalf("configured path = %q, want %q", path, legacyPath)
	}
	if _, err := os.Stat(legacyPath); err != nil {
		t.Fatalf("configured path was unexpectedly moved: %v", err)
	}

	explicitPath := filepath.Join(home, "custom.yml")
	path, err = ResolveEnvironmentFile(testLogger(), legacyPath, "dev", explicitPath)
	if err != nil {
		t.Fatalf("resolving explicit path: %v", err)
	}
	if path != explicitPath {
		t.Fatalf("explicit path = %q, want %q", path, explicitPath)
	}
}

func TestLoadEnvironmentReadsOnlySpecifiedFile(t *testing.T) {
	dir := t.TempDir()
	canonicalPath := filepath.Join(dir, "default.env.yml")
	otherPath := filepath.Join(dir, "default.yml")
	if err := os.WriteFile(canonicalPath, []byte("selected: canonical\n"), 0600); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(otherPath, []byte("selected: legacy\n"), 0600); err != nil {
		t.Fatal(err)
	}

	path, variables, err := LoadEnvironment(testLogger(), canonicalPath)
	if err != nil {
		t.Fatalf("LoadEnvironment() error = %v", err)
	}
	if path != canonicalPath {
		t.Fatalf("LoadEnvironment() path = %q, want %q", path, canonicalPath)
	}
	if got := variables["selected"]; got != "canonical" {
		t.Fatalf("loaded selected value = %#v, want %q", got, "canonical")
	}
}

func testLogger() *logging.Logger {
	logger := logging.New()
	logger.SetOutput(io.Discard)
	return logger
}
