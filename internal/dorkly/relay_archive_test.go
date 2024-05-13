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
			tt.input.IncrementDataId()
			assert.Equal(t, tt.expected, tt.input)
		})
	}
}

func Test_DirectoryToTarGz(t *testing.T) {
	archiveFilePath := filepath.Join(os.TempDir(), "Test_DirectoryToTarGz.tar.gz")
	err := DirectoryToTarGz("testdata", archiveFilePath)
	require.Nil(t, err)

	fileInfo, err := os.Stat(archiveFilePath)
	require.Nil(t, err)

	// We just sanity check that the archive is non-empty
	assert.Greaterf(t, fileInfo.Size(), int64(0), "Got unexpected 0 length file at %s", archiveFilePath)
}
