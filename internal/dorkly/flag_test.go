package dorkly

import (
	"github.com/launchdarkly/go-sdk-common/v3/ldvalue"
	"github.com/launchdarkly/go-server-sdk-evaluation/v3/ldmodel"
	"github.com/stretchr/testify/assert"
	"testing"
)

func Test_FlagBoolean_ToLdFlag(t *testing.T) {
	cases := []struct {
		name     string
		flag     FlagBoolean
		flagBase FlagBase
		expected ldmodel.FeatureFlag
	}{
		{
			name:     "true,client-side ok",
			flag:     FlagBoolean{Variation: true},
			flagBase: FlagBase{key: "test-key", EnableBrowser: true, EnableMobileKey: true},
			expected: ldmodel.FeatureFlag{
				Key: "test-key",
				ClientSideAvailability: ldmodel.ClientSideAvailability{
					UsingMobileKey:     true,
					UsingEnvironmentID: true,
				},
				Salt:         "dGVzdC1rZXk=",
				Variations:   []ldvalue.Value{ldvalue.Bool(true), ldvalue.Bool(false)},
				OffVariation: ldvalue.NewOptionalInt(1),
				Fallthrough:  ldmodel.VariationOrRollout{Variation: ldvalue.NewOptionalInt(0)},
				On:           true,
			},
		},
		{
			name: "false,server-side only",
			flag: FlagBoolean{Variation: false},
			flagBase: FlagBase{
				key:             "test-key-2",
				EnableBrowser:   false,
				EnableMobileKey: false,
			},
			expected: ldmodel.FeatureFlag{
				Key: "test-key-2",
				ClientSideAvailability: ldmodel.ClientSideAvailability{
					UsingMobileKey:     false,
					UsingEnvironmentID: false,
				},
				Salt:         "dGVzdC1rZXktMg==",
				Variations:   []ldvalue.Value{ldvalue.Bool(true), ldvalue.Bool(false)},
				OffVariation: ldvalue.NewOptionalInt(1),
				Fallthrough:  ldmodel.VariationOrRollout{Variation: ldvalue.NewOptionalInt(0)},
				On:           false,
			},
		},
	}

	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, tt.flag.ToLdFlag(tt.flagBase))
		})
	}
}

func Test_FlagBooleanRollout_ToLdFlag(t *testing.T) {
	cases := []struct {
		name     string
		flag     FlagBooleanRollout
		flagBase FlagBase
		expected ldmodel.FeatureFlag
	}{
		{
			name: "100% rollout",
			flag: FlagBooleanRollout{
				PercentRollout: BooleanRolloutVariation{True: 100.0, False: 0.0},
			},
			flagBase: FlagBase{
				key:             "test-key",
				EnableBrowser:   true,
				EnableMobileKey: true,
			},
			expected: ldmodel.FeatureFlag{
				Key: "test-key",
				ClientSideAvailability: ldmodel.ClientSideAvailability{
					UsingMobileKey:     true,
					UsingEnvironmentID: true,
				},
				Salt:         "dGVzdC1rZXk=",
				Variations:   []ldvalue.Value{ldvalue.Bool(true), ldvalue.Bool(false)},
				OffVariation: ldvalue.NewOptionalInt(1),
				Fallthrough: ldmodel.VariationOrRollout{
					Rollout: ldmodel.Rollout{
						Kind:        ldmodel.RolloutKindRollout,
						ContextKind: "user",
						Variations: []ldmodel.WeightedVariation{
							{
								Variation: 0,
								Weight:    100000,
							},
							{
								Variation: 1,
								Weight:    0,
							},
						},
					},
				},
				On: true,
			},
		},
		{
			name: "10% rollout",
			flag: FlagBooleanRollout{
				PercentRollout: BooleanRolloutVariation{True: 10.0, False: 90.0},
			},
			flagBase: FlagBase{
				key:             "test-key-10",
				EnableBrowser:   true,
				EnableMobileKey: true,
			},
			expected: ldmodel.FeatureFlag{
				Key: "test-key-10",
				ClientSideAvailability: ldmodel.ClientSideAvailability{
					UsingMobileKey:     true,
					UsingEnvironmentID: true,
				},
				Salt:         "dGVzdC1rZXktMTA=",
				Variations:   []ldvalue.Value{ldvalue.Bool(true), ldvalue.Bool(false)},
				OffVariation: ldvalue.NewOptionalInt(1),
				Fallthrough: ldmodel.VariationOrRollout{
					Rollout: ldmodel.Rollout{
						Kind:        ldmodel.RolloutKindRollout,
						ContextKind: "user",
						Variations: []ldmodel.WeightedVariation{
							{
								Variation: 0,
								Weight:    10000,
							},
							{
								Variation: 1,
								Weight:    90000,
							},
						},
					},
				},
				On: true,
			},
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			assert.Equal(t, c.expected, c.flag.ToLdFlag(c.flagBase))
		})
	}
}
