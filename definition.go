package main

import (
	"fmt"
	"log"
	"os"
	"regexp"
	"strings"
	"sync"

	"gopkg.in/yaml.v2"
)

var (
	registryMu sync.RWMutex
	Registry   = map[string]ErrorDefinition{}
	validKey   = regexp.MustCompile(`^[a-z0-9_.-]+$`)
)

type ErrorDefinition struct {
	Key          string            `yaml:"key"`
	Code         int               `yaml:"code"`
	Default      string            `yaml:"default"`
	I18n         map[string]string `yaml:"i18n"`
	Expose       bool              `yaml:"expose"`
	InternalCode string            `yaml:"internal_code"`
	Group        string            `yaml:"group"`
}

func Register(key string, def ErrorDefinition) {
	registryMu.Lock()
	defer registryMu.Unlock()

	if _, exists := Registry[key]; exists && !AllowDuplicateKeys {
		if strings.Contains(strings.ToLower(os.Getenv("APP_ENV")), "dev") {
			panic(fmt.Sprintf("duplicate error key: %s", key))
		}
		log.Printf("[WARN] duplicate error key ignored: %s", key)
		return
	}

	if !validKey.MatchString(key) && !RelaxedKeyFormat {
		if strings.Contains(strings.ToLower(os.Getenv("APP_ENV")), "dev") {
			panic(fmt.Sprintf("Invalid error key format: %s", key))
		}
		log.Printf("[WARN] Invalid error key format: %s", key)
	}
	Registry[key] = def
}

func LoadDefinitionsFromYAML(filePath string) error {
	data, err := os.ReadFile(filePath)
	if err != nil {
		return fmt.Errorf("failed to read error definitions file: %w", err)
	}

	var definitions []ErrorDefinition
	if err := yaml.Unmarshal(data, &definitions); err != nil {
		return fmt.Errorf("failed to parse error definitions: %w", err)
	}

	for _, def := range definitions {
		if def.Key == "" || def.Code == 0 || def.Default == "" {
			return fmt.Errorf("invalid error definition: %+v", def)
		}
		Register(def.Key, def)
	}
	return nil
}
