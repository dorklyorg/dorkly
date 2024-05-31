package main

import (
	"github.com/dorklyorg/dorkly/internal/dorkly"
	"os"
)

const (
	dorklyYamlEnvVar           = "DORKLY_YAML"
	defaultDorklyYamlInputPath = "project"
)

var logger = dorkly.GetLogger().Named("Validator")

func main() {
	dorklyYamlInputPath := os.Getenv(dorklyYamlEnvVar)
	if dorklyYamlInputPath == "" {
		logger.Debugf("Env var [%s] not set. Using default: %s", dorklyYamlEnvVar, defaultDorklyYamlInputPath)
		dorklyYamlInputPath = defaultDorklyYamlInputPath
	}
	l := logger.With("dorklyYamlInputPath", dorklyYamlInputPath)

	p, err := dorkly.ValidateYamlProject(dorklyYamlInputPath)
	if err != nil {
		l.Fatalf("Failed to validate project yaml files: %v", err)
	}
	l.With("project", p).Info("Project yaml files validated successfully")
}
