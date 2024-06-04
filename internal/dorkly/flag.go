package dorkly

import (
	"encoding/base64"
	"github.com/launchdarkly/go-sdk-common/v3/ldvalue"
	"github.com/launchdarkly/go-server-sdk-evaluation/v3/ldmodel"
)

const (
	contextKindUser = "user"
)

// FlagBase contains the common flag fields shared between environments
type FlagBase struct {
	key             string
	Description     string   `yaml:"description"`
	Type            FlagType `yaml:"type"`
	EnableMobileKey bool     `yaml:"enableMobileKey"`
	EnableBrowser   bool     `yaml:"enableBrowser"`
}

type FlagType string

const (
	FlagTypeBoolean        FlagType = "boolean"
	FlagTypeBooleanRollout FlagType = "booleanRollout"
)

// FlagConfigForEnv is an interface for flag configs that are specific to an environment
// when combined with a FlagBase, it can be converted to a LaunchDarkly FeatureFlag
type FlagConfigForEnv interface {
	ToLdFlag(flagBase FlagBase) ldmodel.FeatureFlag
}

// FlagBoolean is a boolean flag that is either on (true) or off (false)
type FlagBoolean struct {
	Variation bool `yaml:"variation"`
}

func (f *FlagBoolean) ToLdFlag(flagBase FlagBase) ldmodel.FeatureFlag {
	return flagBase.ldFeatureFlagBoolean(f.Variation)
}

// FlagBooleanRollout is a boolean flag that is on (true) for a percentage of users based on the id field
type FlagBooleanRollout struct {
	PercentRollout BooleanRolloutVariation `yaml:"percentRollout"`
}

type BooleanRolloutVariation struct {
	True  float64 `yaml:"true"`
	False float64 `yaml:"false"`
}

func (f *FlagBooleanRollout) ToLdFlag(flagBase FlagBase) ldmodel.FeatureFlag {
	ldFlag := flagBase.ldFeatureFlagBoolean(true)
	ldFlag.Fallthrough = ldmodel.VariationOrRollout{
		Rollout: ldmodel.Rollout{
			Kind:        ldmodel.RolloutKindRollout,
			ContextKind: contextKindUser,
			Variations: []ldmodel.WeightedVariation{
				{
					Variation: 0, // true
					Weight:    percentToLdWeight(f.PercentRollout.True),
				},
				{
					Variation: 1, // false
					Weight:    percentToLdWeight(f.PercentRollout.False),
				},
			},
		},
	}
	return ldFlag
}

func percentToLdWeight(percent float64) int {
	return int(percent * 1000.0)
}

func (f *FlagBase) ldFeatureFlagBase() ldmodel.FeatureFlag {
	return ldmodel.FeatureFlag{
		Key: f.key,
		ClientSideAvailability: ldmodel.ClientSideAvailability{
			UsingMobileKey:     f.EnableMobileKey,
			UsingEnvironmentID: f.EnableBrowser,
		},
		// TODO: ld-relay archive json files also contain a "clientSide": boolean field.. do we need it?

		// TODO: is this an ok salt? users shouldn't have to manage it.
		Salt: base64.StdEncoding.EncodeToString([]byte(f.key)),
	}
}

func (f *FlagBase) ldFeatureFlagBoolean(on bool) ldmodel.FeatureFlag {
	ldFlag := f.ldFeatureFlagBase()
	ldFlag.Variations = []ldvalue.Value{ldvalue.Bool(true), ldvalue.Bool(false)}
	ldFlag.OffVariation = ldvalue.NewOptionalInt(1)
	ldFlag.Fallthrough = ldmodel.VariationOrRollout{Variation: ldvalue.NewOptionalInt(0)}
	ldFlag.On = on
	return ldFlag
}
