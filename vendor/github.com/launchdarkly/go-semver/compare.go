package semver

// ComparePrecedence compares this Version to another Version according to the canonical precedence rules. It
// returns -1 if v has lower precedence than other, 1 if v has higher precedence, or 0 if the same.
func (v Version) ComparePrecedence(other Version) int {
	if v.major < other.major {
		return -1
	}
	if v.major > other.major {
		return 1
	}
	if v.minor < other.minor {
		return -1
	}
	if v.minor > other.minor {
		return 1
	}
	if v.patch < other.patch {
		return -1
	}
	if v.patch > other.patch {
		return 1
	}
	if v.prerelease == "" && other.prerelease == "" {
		return 0
	}
	// *no* prerelease component always has higher precedence than *any* prerelease component
	if v.prerelease == "" {
		return 1
	}
	if other.prerelease == "" {
		return -1
	}
	// compare prerelease components - the build component, if any, has no effect on precedence
	return comparePrereleaseIdentifiers(v.prerelease, other.prerelease)
}

func comparePrereleaseIdentifiers(prerel1, prerel2 string) int {
	// The parser has already validated the syntax of both of these strings, so we know that they both
	// contain one or more identifiers separated by a period, and that they contain only alphanumerics.
	// If an identifier contains only digits, and does not have a leading zero, then we treat it as a
	// number.

	scanner1 := newSimpleASCIIScanner(prerel1)
	scanner2 := newSimpleASCIIScanner(prerel2)

	for {
		// all components up to this point have been determined to be equal
		if scanner1.eof() {
			if scanner2.eof() {
				return 0
			}
			return -1 // x.y is always less than x.y.z
		} else {
			if scanner2.eof() {
				return 1
			}
		}

		identifier1, _ := scanner1.readUntil(dotTerminator)
		identifier2, _ := scanner2.readUntil(dotTerminator)

		// each sub-identifier is compared numerically if both are numeric; if both are non-numeric,
		// they're compared as strings; otherwise, the numeric one is the lesser one
		var n1, n2, d int
		var isNum1, isNum2 bool
		n1, isNum1 = parsePositiveNumericString(identifier1)
		n2, isNum2 = parsePositiveNumericString(identifier2)
		if isNum1 && isNum2 {
			if n1 < n2 {
				d = -1
			} else if n1 > n2 {
				d = 1
			}
		} else {
			if isNum1 {
				d = -1
			} else if isNum2 {
				d = 1
			} else { // string comparison
				if identifier1 < identifier2 {
					d = -1
				} else if identifier1 > identifier2 {
					d = 1
				}
			}
		}

		if d != 0 {
			return d
		}
	}
}
