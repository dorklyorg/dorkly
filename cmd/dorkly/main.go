package main

import (
	"context"
	"github.com/dorklyorg/dorkly/internal/dorkly"
	"log"
	"os"
)

const (
	dorklyYamlEnvVar           = "DORKLY_YAML"
	newRelayArchiveDirEnvVar   = "NEW_RELAY_ARCHIVE_DIR"
	existingRelayArchiveEnvVar = "EXISTING_RELAY_ARCHIVE"
	newRelayArchiveEnvVar      = "NEW_RELAY_ARCHIVE"

	defaultDorklyYamlInputPath    = "project"
	defaultNewRelayArchiveDirPath = "newRelayArchive"
	defaultExistingRelayArchive   = "flags.tar.gz"
	defaultNewRelayArchive        = "flags-new.tar.gz"
)

func main() {
	//TODO: use a config library to manage these env vars
	dorklyYamlInputPath := os.Getenv(dorklyYamlEnvVar)
	if dorklyYamlInputPath == "" {
		log.Printf(dorklyYamlEnvVar+" env var not set. Using default: %s", defaultDorklyYamlInputPath)
		dorklyYamlInputPath = defaultDorklyYamlInputPath
	}

	newRelayArchiveDirPath := os.Getenv(newRelayArchiveDirEnvVar)
	if newRelayArchiveDirPath == "" {
		log.Printf(newRelayArchiveDirEnvVar + " env var not set. Using default: " + defaultNewRelayArchiveDirPath)
		newRelayArchiveDirPath = defaultNewRelayArchiveDirPath
	}

	existingRelayArchivePath := os.Getenv(existingRelayArchiveEnvVar)
	if existingRelayArchivePath == "" {
		log.Printf(existingRelayArchiveEnvVar + " env var not set. Using default: " + defaultExistingRelayArchive)
		existingRelayArchivePath = defaultExistingRelayArchive
	}

	newRelayArchivePath := os.Getenv(newRelayArchiveEnvVar)
	if newRelayArchivePath == "" {
		log.Printf(newRelayArchiveEnvVar + " env var not set. Using default: " + newRelayArchivePath)
		newRelayArchivePath = defaultNewRelayArchive
	}

	localFileArchiveService := dorkly.NewLocalFileRelayArchiveService("temp/flags.tar.gz")
	reconciler := dorkly.NewReconciler(localFileArchiveService, dorklyYamlInputPath)

	err := reconciler.Reconcile(context.Background())
	if err != nil {
		log.Fatal(err)
	}
}
