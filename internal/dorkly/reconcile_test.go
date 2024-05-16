package dorkly

import (
	"fmt"
	"github.com/stretchr/testify/assert"
	"testing"
)

func Test_Reconcile(t *testing.T) {
	type args struct {
		old RelayArchive
		new RelayArchive
	}
	tests := []struct {
		name    string
		args    args
		want    RelayArchive
		wantErr assert.ErrorAssertionFunc
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := Reconcile(tt.args.old, tt.args.new)
			if !tt.wantErr(t, err, fmt.Sprintf("Reconcile(%v, %v)", tt.args.old, tt.args.new)) {
				return
			}
			assert.Equalf(t, tt.want, got, "Reconcile(%v, %v)", tt.args.old, tt.args.new)
		})
	}
}

func Test_compareMaps(t *testing.T) {
	type testCase struct {
		name     string
		old      map[string]string
		new      map[string]string
		expected compareResult
	}
	tests := []testCase{
		{
			name: "nil input",
			old:  nil,
			new:  nil,
			expected: compareResult{
				new:      []string{},
				existing: []string{},
				deleted:  []string{},
			},
		},
		{
			name: "nil new",
			old:  map[string]string{"aKey": "aValue", "bKey": "bValue"},
			new:  nil,
			expected: compareResult{
				new:      []string{},
				existing: []string{},
				deleted:  []string{"aKey", "bKey"},
			},
		},
		{
			name: "nil old",
			old:  nil,
			new:  map[string]string{"aKey": "aValue", "bKey": "bValue"},
			expected: compareResult{
				new:      []string{"aKey", "bKey"},
				existing: []string{},
				deleted:  []string{},
			},
		},
		{
			name: "same",
			old:  map[string]string{"aKey": "aValue", "bKey": "bValue"},
			new:  map[string]string{"aKey": "aValue", "bKey": "bValue"},
			expected: compareResult{
				new:      []string{},
				existing: []string{"aKey", "bKey"},
				deleted:  []string{},
			},
		},
		{
			name: "mixed",
			old:  map[string]string{"deletedKey": "deletedValue", "bKey": "bValue"},
			new:  map[string]string{"newKey": "newValue", "bKey": "bValue"},
			expected: compareResult{
				new:      []string{"newKey"},
				existing: []string{"bKey"},
				deleted:  []string{"deletedKey"},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			actual := compareMapKeys(tt.old, tt.new)
			assert.ElementsMatch(t, tt.expected.new, actual.new, "compareMapKeys(%v, %v)", tt.old, tt.new)
			assert.ElementsMatch(t, tt.expected.existing, actual.existing, "compareMapKeys(%v, %v)", tt.old, tt.new)
			assert.ElementsMatch(t, tt.expected.deleted, actual.deleted, "compareMapKeys(%v, %v)", tt.old, tt.new)
		})
	}
}
