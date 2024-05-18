package dorkly

import (
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"os"
	"path/filepath"
	"testing"
)

func Test_ArchiveEnvironmentRep_IncrementDataId(t *testing.T) {
	tests := []struct {
		name     string
		input    RelayArchiveEnv
		expected RelayArchiveEnv
	}{
		{
			name:     "zero",
			input:    RelayArchiveEnv{},
			expected: RelayArchiveEnv{dataId: 1, DataId: "\"1\""},
		},
		{
			name:     "one",
			input:    RelayArchiveEnv{dataId: 1, DataId: "\"1\""},
			expected: RelayArchiveEnv{dataId: 2, DataId: "\"2\""},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.input.incrementDataId()
			assert.Equal(t, tt.expected, tt.input)
		})
	}
}

// Ensures the round trip of archive -> files -> unmarshaling -> marshaling -> files -> archive works as expected.
func Test_RelayArchive_RoundTrip(t *testing.T) {
	// 1. Load a relay archive from a tar.gz file in testdata:
	testdataArchive, err := LoadRelayArchive("testdata/flags.tar.gz")
	require.Nil(t, err)

	assert.Len(t, testdataArchive.envs, 2)
	//TODO: Add more assertions here

	// 2. Create all files for a new archive in a temporary directory:
	archiveFileInputDir := filepath.Join(t.TempDir(), "archiveFileInputDir")
	err = testdataArchive.CreateArchiveFilesAndComputeChecksum(archiveFileInputDir)
	require.Nil(t, err)

	// 3. Create a new archive from the files in the temporary directory:
	archiveOutputPath := filepath.Join(t.TempDir(), "flags.tar.gz")
	err = DirectoryToTarGz(archiveFileInputDir, archiveOutputPath)
	require.Nil(t, err)

	// 4. Load the new archive and compare it to the original:
	newArchive, err := LoadRelayArchive(archiveOutputPath)
	require.Nil(t, err)
	assert.Equal(t, testdataArchive, newArchive)
}

func Test_DirectoryToTarGz(t *testing.T) {
	archiveFilePath := filepath.Join(t.TempDir(), "Test_DirectoryToTarGz.tar.gz")
	err := DirectoryToTarGz("testdata", archiveFilePath)
	require.Nil(t, err)

	fileInfo, err := os.Stat(archiveFilePath)
	require.Nil(t, err)

	// We just sanity check that the archive is non-empty
	assert.Greaterf(t, fileInfo.Size(), int64(0), "Got unexpected 0 length file at %s", archiveFilePath)
}
