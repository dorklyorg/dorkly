package dorkly

import (
	"encoding/base64"
	"github.com/launchdarkly/go-server-sdk-evaluation/v3/ldmodel"
)

type FlagType string

const (
	FlagTypeBoolean        FlagType = "boolean"
	FlagTypeBooleanRollout FlagType = "booleanRollout"
	FlagTypeString         FlagType = "string"

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

// FlagConfigForEnv is an interface for flag configs that are specific to an environment
// when combined with a FlagBase, it can be converted to a LaunchDarkly FeatureFlag
type FlagConfigForEnv interface {
	ToLdFlag(flagBase FlagBase) ldmodel.FeatureFlag
	Validate(flagBase FlagBase) error
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
