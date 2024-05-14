package main

import (
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

	proj, err := dorkly.LoadProjectYamlFiles(dorklyYamlInputPath)
	if err != nil {
		log.Fatal(err)
	}

	relayArchive, err := proj.ToRelayArchive()
	if err != nil {
		log.Fatal(err)
	}

	existingFlags, err := dorkly.LoadRelayArchive(existingRelayArchivePath)
	if err != nil {
		log.Fatal(err)
	}

	reconciledRelayArchive, err := dorkly.Reconcile(*existingFlags, *relayArchive)
	if err != nil {
		log.Fatal(err)
	}

	err = reconciledRelayArchive.CreateArchiveFilesAndComputeChecksum(newRelayArchiveDirPath)
	if err != nil {
		log.Fatal(err)
	}

	err = dorkly.DirectoryToTarGz(newRelayArchiveDirPath, newRelayArchivePath)
	if err != nil {
		log.Fatal(err)
	}

}
