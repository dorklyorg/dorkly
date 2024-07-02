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
	// Consider it the canonical Project.
	testProject1 = Project{
		Name:         "testProject1",
		Description:  "Human-readable description of the project.",
		environments: []string{"production", "staging"},
		Flags: map[string]Flag{
			"boolean1": {
				FlagBase: FlagBase{
					key:             "boolean1",
					Description:     "Human-readable description of the flag.",
					Type:            "boolean",
					EnableBrowser:   true,
					EnableMobileKey: true,
				},
				envConfigs: map[string]FlagConfigForEnv{
					"production": &FlagBoolean{
						FlagBase: FlagBase{
							key:             "boolean1",
							Description:     "Human-readable description of the flag.",
							Type:            "boolean",
							EnableBrowser:   true,
							EnableMobileKey: true,
						},
						Variation: false,
					},
					"staging": &FlagBoolean{
						FlagBase: FlagBase{
							key:             "boolean1",
							Description:     "Human-readable description of the flag.",
							Type:            "boolean",
							EnableBrowser:   true,
							EnableMobileKey: true,
						},
						Variation: true,
					},
				},
			},
			"rollout1": {
				FlagBase: FlagBase{
					key:             "rollout1",
					Description:     "Human-readable description of the flag.",
					Type:            "booleanRollout",
					EnableBrowser:   true,
					EnableMobileKey: true,
				},
				envConfigs: map[string]FlagConfigForEnv{
					"production": &FlagBooleanRollout{
						FlagBase: FlagBase{
							key:             "rollout1",
							Description:     "Human-readable description of the flag.",
							Type:            "booleanRollout",
							EnableBrowser:   true,
							EnableMobileKey: true,
						},
						PercentRollout: BooleanRolloutVariation{True: 31.0, False: 69.0},
					},
					"staging": &FlagBooleanRollout{
						FlagBase: FlagBase{
							key:             "rollout1",
							Description:     "Human-readable description of the flag.",
							Type:            "booleanRollout",
							EnableBrowser:   true,
							EnableMobileKey: true,
						},
						PercentRollout: BooleanRolloutVariation{True: 100.0, False: 0.0}},
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
			got, err := loadProjectYamlFiles(tt.path)
			if !tt.wantErr(t, err, fmt.Sprintf("loadProjectYamlFiles(%v)", tt.path)) {
				return
			}
			assert.Equalf(t, tt.want, got, "loadProjectYamlFiles(%v)", tt.path)
		})
	}
}

// To update golden files run this from the root of the repo:
// go test ./internal/dorkly -update
// then check in the resulting changes.
func Test_ToRelayArchive(t *testing.T) {
	expectedArchiveFiles := []string{"staging.json", "staging-data.json", "production.json", "production-data.json"}
	actualRelayArchive := testProject1.toRelayArchive()
	assert.NotNil(t, actualRelayArchive)

	// The actualRelayArchive struct is big and gnarly, so instead of comparing it directly,
	// we'll marshal it to JSON and compare it to the golden files.
	actualArchiveFiles, err := actualRelayArchive.marshalArchiveFilesJson()
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
