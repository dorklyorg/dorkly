package dorkly

import (
	"errors"
	"github.com/launchdarkly/go-sdk-common/v3/ldvalue"
	"github.com/launchdarkly/go-server-sdk-evaluation/v3/ldmodel"
)

// This ensures that FlagBoolean implements FlagConfigForEnv
var _ FlagConfigForEnv = &FlagBoolean{}

// FlagBoolean is a boolean flag that is either on (true) or off (false)
type FlagBoolean struct {
	FlagBase
	Variation bool `yaml:"variation"`
}

func (f *FlagBoolean) Validate() error {
	return nil
}

func (f *FlagBoolean) ToLdFlag() ldmodel.FeatureFlag {
	return f.ldFeatureFlagBoolean(f.Variation)
}

// This ensures that FlagBooleanRollout implements FlagConfigForEnv
var _ FlagConfigForEnv = &FlagBooleanRollout{}

// FlagBooleanRollout is a boolean flag that is on (true) for a percentage of users based on the id field
type FlagBooleanRollout struct {
	FlagBase
	PercentRollout BooleanRolloutVariation `yaml:"percentRollout"`
}

func (f *FlagBooleanRollout) Validate() error {
	if f.PercentRollout.True < 0.0 {
		return errors.New("percentRollout.true must be >= 0")
	}
	if f.PercentRollout.False < 0.0 {
		return errors.New("percentRollout.false must be >= 0")
	}

	if f.PercentRollout.True+f.PercentRollout.False > 100 {
		return errors.New("sum of percentRollout values must be <= 100")
	}

	if f.PercentRollout.True == 0.0 && f.PercentRollout.False == 0 {
		return errors.New("at least one of percentRollout.true or percentRollout.false must be > 0")
	}

	if f.PercentRollout.True == 0.0 {
		f.PercentRollout.True = 100.0 - f.PercentRollout.False
	}

	if f.PercentRollout.False == 0.0 {
		f.PercentRollout.False = 100.0 - f.PercentRollout.True
	}

	return nil
}

type BooleanRolloutVariation struct {
	True  float64 `yaml:"true"`
	False float64 `yaml:"false"`
}

func (f *FlagBooleanRollout) ToLdFlag() ldmodel.FeatureFlag {
	ldFlag := f.ldFeatureFlagBoolean(true)
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

func (f *FlagBase) ldFeatureFlagBoolean(on bool) ldmodel.FeatureFlag {
	ldFlag := f.ldFeatureFlagBase()
	ldFlag.Variations = []ldvalue.Value{ldvalue.Bool(true), ldvalue.Bool(false)}
	ldFlag.OffVariation = ldvalue.NewOptionalInt(1)
	ldFlag.Fallthrough = ldmodel.VariationOrRollout{Variation: ldvalue.NewOptionalInt(0)}
	ldFlag.On = on
	return ldFlag
}
