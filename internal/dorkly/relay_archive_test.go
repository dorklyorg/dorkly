package dorkly

import (
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"os"
	"path/filepath"
	"testing"
)

// Ensures the round trip of archive -> files -> unmarshaling -> marshaling -> files -> archive works as expected.
func Test_RelayArchive_RoundTrip(t *testing.T) {
	var err error
	var testdataArchive *RelayArchive
	archiveOutputPath := filepath.Join(t.TempDir(), "flags.tar.gz")

	// 1. Load a relay archive from a tar.gz file in testdata:
	t.Run("loadRelayArchiveFromTarGzFile", func(t *testing.T) {
		testdataArchive, err = loadRelayArchiveFromTarGzFile("testdata/flags.tar.gz")
		require.Nil(t, err)

		assert.Len(t, testdataArchive.envs, 2)
		//TODO: Add more assertions here
	})

	// 2. Write the relay archive to a new tar.gz file:
	t.Run("toTarGzFile", func(t *testing.T) {
		err = testdataArchive.toTarGzFile(archiveOutputPath)
		require.Nil(t, err)
	})

	// 3. Load the new archive and compare it to the original:
	t.Run("loadRelayArchiveAgain", func(t *testing.T) {
		newArchive, err := loadRelayArchiveFromTarGzFile(archiveOutputPath)
		require.Nil(t, err)
		assert.Equal(t, testdataArchive, newArchive)
	})
}

func Test_DirectoryToTarGz(t *testing.T) {
	archiveFilePath := filepath.Join(t.TempDir(), "Test_DirectoryToTarGz.tar.gz")
	err := directoryToTarGz("testdata", archiveFilePath)
	require.Nil(t, err)

	fileInfo, err := os.Stat(archiveFilePath)
	require.Nil(t, err)

	// We just sanity check that the archive is non-empty
	assert.Greaterf(t, fileInfo.Size(), int64(0), "Got unexpected 0 length file at %s", archiveFilePath)
}
