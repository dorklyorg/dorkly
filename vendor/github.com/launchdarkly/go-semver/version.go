package semver

// Version is a semantic version as defined by the Semantic Versions 2.0.0 standard (http://semver.org).
//
// This type provides only parsing and simple precedence comparison, since those are the only features
// required by the LaunchDarkly Go SDK.
type Version struct {
	major      int
	minor      int
	patch      int
	prerelease string
	build      string
}

// GetMajor returns the numeric major version component.
func (v Version) GetMajor() int {
	return v.major
}

// GetMinor returns the numeric minor version component.
func (v Version) GetMinor() int {
	return v.minor
}

// GetPatch returns the numeric patch version component.
func (v Version) GetPatch() int {
	return v.patch
}

// GetPrerelease returns the prerelease version component, or "" if there is none.
func (v Version) GetPrerelease() string {
	return v.prerelease
}

// GetBuild returns the build version component, or "" if there is none.
func (v Version) GetBuild() string {
	return v.build
}
