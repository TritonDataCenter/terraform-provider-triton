package triton

// wildcardMatch performs a simple wildcard pattern match against a string
// value. This implementation uses Russ Cox's linear match algorithm, and
// supports only two types of common wildcards:
//
// * - matches any number of any characters including none; and
// ? - matches one occurrence of any character.
//
// There is no support for either ranges or character classes.
func wildcardMatch(pattern, s string) bool {
	p := 0
	n := 0
	nextP := 0
	nextN := 0

	// Would always match.
	if pattern == "*" {
		return true
	}

	for n < len(s) || p < len(pattern) {
		if p < len(pattern) {
			c := pattern[p]

			switch c {
			case '?':
				if n < len(s) {
					p++
					n++
					continue
				}
			case '*':
				nextP = p
				nextN = n + 1
				p++
				continue
			default:
				if n < len(s) && s[n] == c {
					p++
					n++
					continue
				}
			}
		}

		// Restart.
		if 0 < nextN && nextN <= len(s) {
			p = nextP
			n = nextN
			continue
		}
		return false
	}
	return true
}
