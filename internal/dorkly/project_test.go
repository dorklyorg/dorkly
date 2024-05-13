package dorkly

import (
	"fmt"
	"github.com/stretchr/testify/assert"
	"gotest.tools/v3/golden"
	"testing"
)

var (

	// This struct is re-used across tests. It is defined here to avoid duplication.
	// It must be kept in sync with the files in testdata/testProject1 in order for tests to pass.
	testProject1 = Project{
		Key:          "testProject1",
		Description:  "Human-readable description of the project.",
		Environments: []string{"staging", "production"},
		Flags: map[string]Flag{
			"boolean1": {
				FlagBase: FlagBase{
					Key:            "boolean1",
					Description:    "Human-readable description of the flag.",
					Type:           "boolean",
					ServerSideOnly: false,
				},
				envConfigs: map[string]FlagConfigForEnv{
					"production": &FlagBoolean{Variation: false},
					"staging":    &FlagBoolean{Variation: true},
				},
			},
			"rollout1": {
				FlagBase: FlagBase{
					Key:            "rollout1",
					Description:    "Human-readable description of the flag.",
					Type:           "booleanRollout",
					ServerSideOnly: false,
				},
				envConfigs: map[string]FlagConfigForEnv{
					"production": &FlagBooleanRollout{PercentRollout: 31.0},
					"staging":    &FlagBooleanRollout{PercentRollout: 100.0},
				},
			},
		},
	}
)

func Test_getFileNameNoSuffix(t *testing.T) {
	cases := []struct {
		name     string
		path     string
		expected string
	}{
		{name: "noSuffix", path: "blah", expected: "blah"},
		{name: "suffix", path: "blah.yml", expected: "blah"},
		{name: "suffixWithPath", path: "/var/log/blah.yml", expected: "blah"},
	}

	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equalf(t, tt.expected, getFileNameNoSuffix(tt.path), "getFileNameNoSuffix(%v)", tt.path)
		})
	}
}

func Test_LoadProjectYaml(t *testing.T) {
	cases := []struct {
		name    string
		path    string
		want    *Project
		wantErr assert.ErrorAssertionFunc
	}{
		{
			name:    "testProject1",
			path:    "testdata/testProject1/inputYaml",
			want:    &testProject1,
			wantErr: assert.NoError,
		},
	}

	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {
			tt.want.path = tt.path
			got, err := LoadProjectYamlFiles(tt.path)
			if !tt.wantErr(t, err, fmt.Sprintf("LoadProjectYamlFiles(%v)", tt.path)) {
				return
			}
			assert.Equalf(t, tt.want, got, "LoadProjectYamlFiles(%v)", tt.path)
		})
	}
}

// To update golden files run this from the root of the repo:
// go test ./internal/dorkly -update
// then check in the resulting changes.
func Test_ToRelayArchive(t *testing.T) {
	expectedArchiveFiles := []string{"staging.json", "staging-data.json", "production.json", "production-data.json"}
	actualRelayArchive, err := testProject1.ToRelayArchive()
	assert.NoError(t, err)
	assert.NotNil(t, actualRelayArchive)
	assert.NoError(t, err)

	// The actualRelayArchive struct is big and gnarly, so instead of comparing it directly,
	// we'll marshal it to JSON and compare it to the golden files.
	actualArchiveFiles, err := actualRelayArchive.MarshalArchiveFilesJson()
	assert.NoError(t, err)
	assert.Len(t, actualArchiveFiles, len(expectedArchiveFiles))

	for _, expectedFile := range expectedArchiveFiles {
		t.Run(expectedFile, func(t *testing.T) {
			fileBytes, ok := actualArchiveFiles[expectedFile]
			assert.True(t, ok, "expected file %v not found", expectedFile)
			golden.Assert(t, string(fileBytes), fmt.Sprintf("testProject1/outputJson/%s.golden", expectedFile))
		})
	}
}
