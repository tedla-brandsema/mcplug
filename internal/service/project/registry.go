package project

import (
	_ "embed"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path"
	"path/filepath"
	"strings"

	"github.com/bmatcuk/doublestar/v4"
)

//go:embed project.cfg.json
var embeddedProjectConfig []byte

const projectConfigFileName = "project.cfg.json"

type Registry struct {
	Project ProjectRules `json:"project"`
}

type ProjectRules struct {
	ImportantFiles          []string `json:"important_files"`
	SourceExtensions        []string `json:"source_extensions"`
	TestPatterns            []string `json:"test_patterns"`
	DocumentationExtensions []string `json:"documentation_extensions"`
	DocumentationFiles      []string `json:"documentation_files"`
	ConfigurationExtensions []string `json:"configuration_extensions"`
	ConfigurationFiles      []string `json:"configuration_files"`
}

func LoadOrCreateDefaultRegistry() (Registry, string, error) {
	configPath, err := DefaultRegistryPath()
	if err != nil {
		return Registry{}, "", err
	}

	registry, err := LoadOrCreateRegistry(configPath)
	if err != nil {
		return Registry{}, "", err
	}

	return registry, configPath, nil
}

func DefaultRegistryPath() (string, error) {
	dir, err := os.UserConfigDir()
	if err != nil {
		return "", fmt.Errorf("resolve user config dir: %w", err)
	}

	return filepath.Join(dir, "mcpfs", projectConfigFileName), nil
}

func LoadOrCreateRegistry(configPath string) (Registry, error) {
	if configPath == "" {
		return Registry{}, fmt.Errorf("config path is required")
	}

	if _, err := os.Stat(configPath); err != nil {
		if !errors.Is(err, os.ErrNotExist) {
			return Registry{}, fmt.Errorf("stat project config: %w", err)
		}

		if err := os.MkdirAll(filepath.Dir(configPath), 0o755); err != nil {
			return Registry{}, fmt.Errorf("create project config dir: %w", err)
		}

		if err := os.WriteFile(configPath, embeddedProjectConfig, 0o644); err != nil {
			return Registry{}, fmt.Errorf("write project config: %w", err)
		}
	}

	data, err := os.ReadFile(configPath)
	if err != nil {
		return Registry{}, fmt.Errorf("read project config: %w", err)
	}

	var registry Registry
	if err := json.Unmarshal(data, &registry); err != nil {
		return Registry{}, fmt.Errorf("parse project config %q: %w", configPath, err)
	}

	return registry, nil
}

func MustDefaultRegistryForTests() Registry {
	var registry Registry
	if err := json.Unmarshal(embeddedProjectConfig, &registry); err != nil {
		panic(err)
	}
	return registry
}

func (r Registry) IsImportantFile(rel string) bool {
	return matchFileRules(rel, r.Project.ImportantFiles)
}

func (r Registry) IsSourceFile(rel string) bool {
	return matchExtension(rel, r.Project.SourceExtensions)
}

func (r Registry) IsTestFile(rel string) bool {
	return matchFileRules(rel, r.Project.TestPatterns)
}

func (r Registry) IsDocumentationFile(rel string) bool {
	return matchExtension(rel, r.Project.DocumentationExtensions) ||
		matchFileRules(rel, r.Project.DocumentationFiles)
}

func (r Registry) IsConfigurationFile(rel string) bool {
	return matchExtension(rel, r.Project.ConfigurationExtensions) ||
		matchFileRules(rel, r.Project.ConfigurationFiles)
}

func matchExtension(rel string, extensions []string) bool {
	ext := strings.ToLower(path.Ext(rel))
	if ext == "" {
		return false
	}

	for _, candidate := range extensions {
		candidate = strings.ToLower(strings.TrimSpace(candidate))
		if candidate == "" {
			continue
		}
		if !strings.HasPrefix(candidate, ".") {
			candidate = "." + candidate
		}
		if ext == candidate {
			return true
		}
	}

	return false
}

func matchFileRules(rel string, rules []string) bool {
	rel = path.Clean(filepath.ToSlash(rel))
	relLower := strings.ToLower(rel)
	baseLower := strings.ToLower(path.Base(rel))

	for _, rule := range rules {
		rule = strings.ToLower(strings.TrimSpace(filepath.ToSlash(rule)))
		if rule == "" {
			continue
		}

		if strings.ContainsAny(rule, "*?[") || strings.Contains(rule, "/") {
			if ok, _ := doublestar.PathMatch(rule, relLower); ok {
				return true
			}
			if ok, _ := doublestar.PathMatch(rule, baseLower); ok {
				return true
			}
			continue
		}

		if baseLower == rule || relLower == rule {
			return true
		}
	}

	return false
}
