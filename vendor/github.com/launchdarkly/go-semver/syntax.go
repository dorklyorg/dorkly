package semver

func dotTerminator(ch rune) bool {
	return ch == '.'
}

func hyphenOrPlusTerminator(ch rune) bool {
	return ch == '-' || ch == '+'
}

func dotOrHyphenOrPlusTerminator(ch rune) bool {
	return ch == '.' || ch == '-' || ch == '+'
}

func plusTerminator(ch rune) bool {
	return ch == '+'
}

func isDigit(ch rune) bool {
	return ch >= '0' && ch <= '9'
}

func isAlphanumericOrHyphen(ch rune) bool {
	return (ch >= '0' && ch <= '9') || (ch >= 'a' && ch <= 'z') || (ch >= 'A' && ch <= 'Z') || ch == '-'
}

// Attempts to parse a string as an integer greater than or equal to zero. A zero value must be
// only "0"; otherwise leading zeroes are not allowed. Non-ASCII strings are not supported.
func parsePositiveNumericString(s string) (int, bool) {
	max := len(s)
	if max == 0 {
		return 0, false
	}
	n := 0
	for i := 0; i < max; i++ {
		ch := rune(s[i])
		if ch < '0' || ch > '9' {
			return 0, false
		}
		if ch == '0' && i == 0 && max > 1 {
			return 0, false // leading zeroes aren't allowed
		}
		n = n*10 + (int(ch) - int('0'))
	}
	return n, true
}

func everyChar(s string, validatorFn func(rune) bool) bool {
	n := len(s)
	for i := 0; i < n; i++ {
		if !validatorFn(rune(s[i])) { // we can assume it's an ASCII string due to prior validation
			return false
		}
	}
	return true
}
