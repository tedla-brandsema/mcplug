package project

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoadOrCreateRegistryWritesSeedConfig(t *testing.T) {
	configPath := filepath.Join(t.TempDir(), "mcpfs", projectConfigFileName)

	registry, err := LoadOrCreateRegistry(configPath)
	if err != nil {
		t.Fatalf("LoadOrCreateRegistry returned error: %v", err)
	}

	if !registry.IsImportantFile("README.md") {
		t.Fatal("README.md is not important")
	}

	if _, err := os.Stat(configPath); err != nil {
		t.Fatalf("project config was not written: %v", err)
	}
}

func TestLoadOrCreateRegistryLoadsExistingConfig(t *testing.T) {
	configPath := filepath.Join(t.TempDir(), "mcpfs", projectConfigFileName)

	data := []byte(`{
		"project": {
			"important_files": ["Cloudberry.project"],
			"source_extensions": [".cloudberry"],
			"test_patterns": ["*.cloudberry.test"],
			"documentation_extensions": [],
			"documentation_files": [],
			"configuration_extensions": [],
			"configuration_files": ["Cloudberry.project"]
		}
	}`)

	if err := os.MkdirAll(filepath.Dir(configPath), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(configPath, data, 0o644); err != nil {
		t.Fatal(err)
	}

	registry, err := LoadOrCreateRegistry(configPath)
	if err != nil {
		t.Fatalf("LoadOrCreateRegistry returned error: %v", err)
	}

	if !registry.IsImportantFile("Cloudberry.project") {
		t.Fatal("Cloudberry.project is not important")
	}
	if registry.IsImportantFile("README.md") {
		t.Fatal("README.md should not be important for custom registry")
	}
	if !registry.IsSourceFile("main.cloudberry") {
		t.Fatal("main.cloudberry is not a source file")
	}
	if !registry.IsTestFile("main.cloudberry.test") {
		t.Fatal("main.cloudberry.test is not a test file")
	}
	if !registry.IsConfigurationFile("Cloudberry.project") {
		t.Fatal("Cloudberry.project is not a configuration file")
	}
}

func TestLoadOrCreateRegistryRejectsInvalidJSON(t *testing.T) {
	configPath := filepath.Join(t.TempDir(), "mcpfs", projectConfigFileName)

	if err := os.MkdirAll(filepath.Dir(configPath), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(configPath, []byte(`{invalid`), 0o644); err != nil {
		t.Fatal(err)
	}

	_, err := LoadOrCreateRegistry(configPath)
	if err == nil {
		t.Fatal("LoadOrCreateRegistry returned nil error")
	}
}
