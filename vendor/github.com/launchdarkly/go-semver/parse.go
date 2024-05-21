package semver

import (
	"errors"
)

// ParseMode is an enum-like type used with ParseAs.
type ParseMode int

var invalidSemverError = errors.New("invalid semantic version")

const (
	// ParseModeStrict is the default parsing mode, requiring a strictly correct version string with
	// all three required numeric components.
	ParseModeStrict = iota

	// ParseModeAllowMissingMinorAndPatch is a parsing mode in which the version string may omit the patch
	// version component ("2.1"), or both the minor and patch version components ("2"), in which case
	// they are assumed to be zero.
	ParseModeAllowMissingMinorAndPatch = iota
)

// Parse attempts to parse a string into a Version. It only accepts strings that strictly match the
// specification, so extensions like a "v" prefix are not allowed.
//
// If parsing fails, it returns a non-nil error as the second return value, and Version{} as the first.
func Parse(s string) (Version, error) {
	return ParseAs(s, ParseModeStrict)
}

// ParseAs attempts to parse a string into a Version, using the specified ParseMode.
//
// If parsing fails, it returns a non-nil error as the second return value, and Version{} as the first.
func ParseAs(s string, mode ParseMode) (Version, error) {
	if mode != ParseModeStrict && mode != ParseModeAllowMissingMinorAndPatch {
		return Version{}, errors.New("invalid ParseMode")
	}

	scanner := newSimpleASCIIScanner(s)

	var result Version
	var term int8
	var ok bool

	if mode == ParseModeAllowMissingMinorAndPatch {
		result.major, term, ok = requirePositiveIntegerComponent(&scanner, dotOrHyphenOrPlusTerminator)
		if !ok {
			return Version{}, invalidSemverError
		}
		if term == '.' {
			result.minor, term, ok = requirePositiveIntegerComponent(&scanner, dotOrHyphenOrPlusTerminator)
			if !ok {
				return Version{}, invalidSemverError
			}
			if term == '.' {
				result.patch, term, ok = requirePositiveIntegerComponent(&scanner, hyphenOrPlusTerminator)
				if !ok {
					return Version{}, invalidSemverError
				}
			}
		}
	} else {
		result.major, term, ok = requirePositiveIntegerComponent(&scanner, dotTerminator)
		if !ok || term != '.' {
			return Version{}, invalidSemverError
		}
		result.minor, term, ok = requirePositiveIntegerComponent(&scanner, dotTerminator)
		if !ok || term != '.' {
			return Version{}, invalidSemverError
		}
		result.patch, term, ok = requirePositiveIntegerComponent(&scanner, hyphenOrPlusTerminator)
		if !ok {
			return Version{}, invalidSemverError
		}
	}

	if term == '-' {
		result.prerelease, term = scanner.readUntil(plusTerminator)
		if result.prerelease == "" || term == scannerNonASCII || !validatePrerelease(result.prerelease) {
			return Version{}, invalidSemverError
		}
	}

	if term == '+' {
		result.build, term = scanner.readUntil(noTerminator)
		if result.build == "" || term == scannerNonASCII || !validateBuild(result.build) {
			return Version{}, invalidSemverError
		}
	}

	return result, nil
}

func requirePositiveIntegerComponent(
	scanner *simpleASCIIScanner,
	terminatorFn func(rune) bool,
) (n int, terminatedBy int8, ok bool) {
	// From spec:
	// A normal version number MUST take the form X.Y.Z where X, Y, and Z are non-negative integers, and
	// MUST NOT contain leading zeroes.
	substr, terminatedBy := scanner.readUntil(terminatorFn)
	if terminatedBy == scannerNonASCII {
		return 0, terminatedBy, false
	}
	if n, okNumber := parsePositiveNumericString(substr); okNumber {
		return n, terminatedBy, true
	}
	return 0, terminatedBy, false
}

func validatePrerelease(s string) bool {
	// BNF definition from spec:
	// <pre-release> ::= <dot-separated pre-release identifiers>
	// <dot-separated pre-release identifiers> ::=
	//     <pre-release identifier>
	//     | <pre-release identifier> "." <dot-separated pre-release identifiers>
	// <pre-release identifier> ::= <alphanumeric identifier> | <numeric identifier>
	// <alphanumeric identifier> ::=
	//     <non-digit>
	//     | <non-digit> <identifier characters>
	//     | <identifier characters> <non-digit>
	//     | <identifier characters> <non-digit> <identifier characters>
	// <numeric identifier> ::=
	//     "0"
	//     | <positive digit>
	//     | <positive digit> <digits>
	// <identifier characters> ::= <identifier character> | <identifier character> <identifier characters>
	// <identifier character> ::= <digit> | <non-digit>
	// <non-digit> ::= <letter> | "-"
	// <digits> ::= <digit> | <digit> <digits>
	// <digit> ::= "0" | <positive digit>
	// <positive digit> ::= "1" | "2" | "3" | "4" | "5" | "6" | "7" | "8" | "9"
	// <letter> ::= [A-Za-z]
	//
	// Textual definition from spec:
	// A pre-release version MAY be denoted by appending a hyphen and a series of dot separated identifiers
	// immediately following the patch version. Identifiers MUST comprise only ASCII alphanumerics and hyphen
	// [0-9A-Za-z-]. Identifiers MUST NOT be empty. Numeric identifiers MUST NOT include leading zeroes.
	// Pre-release versions have a lower precedence than the associated normal version. A pre-release version
	// indicates that the version is unstable and might not satisfy the intended compatibility requirements as
	// denoted by its associated normal version. Examples: 1.0.0-alpha, 1.0.0-alpha.1, 1.0.0-0.3.7,
	// 1.0.0-x.7.z.92.
	scanner := newSimpleASCIIScanner(s)
	for {
		substr, terminatedBy := scanner.readUntil(dotTerminator)
		if terminatedBy == scannerNonASCII || substr == "" {
			return false
		}
		if !everyChar(substr, isAlphanumericOrHyphen) {
			return false
		}
		if len(substr) > 1 && everyChar(substr, isDigit) && substr[0] == '0' {
			// leading zero is not allowed in an all-numeric string, for prerelease (OK in build)
			return false
		}
		if terminatedBy == scannerEOF {
			break
		}
	}
	return true
}

func validateBuild(s string) bool {
	// BNF definition from spec (see validatePrerelease for basic definitions)
	//
	// <build> ::= <dot-separated build identifiers>
	// <dot-separated build identifiers> ::=
	//     <build identifier>
	//     | <build identifier> "." <dot-separated build identifiers>
	// <build identifier> ::= <alphanumeric identifier> | <digits>
	//
	// Textual definition from spec:
	// Build metadata MAY be denoted by appending a plus sign and a series of dot separated identifiers
	// immediately following the patch or pre-release version. Identifiers MUST comprise only ASCII
	// alphanumerics and hyphen [0-9A-Za-z-]. Identifiers MUST NOT be empty.
	scanner := newSimpleASCIIScanner(s)
	for {
		substr, terminatedBy := scanner.readUntil(dotTerminator)
		if terminatedBy == scannerNonASCII || substr == "" || !everyChar(substr, isAlphanumericOrHyphen) {
			return false
		}
		if terminatedBy == scannerEOF {
			break
		}
	}
	return true
}
