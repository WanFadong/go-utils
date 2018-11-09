package utils

// IsSliceEqual returns whether slice a is equal to slice b element by element
func IsSliceEqual(a []byte, b []byte) bool {
	if len(a) != len(b) {
		return false
	}

	for i := 0; i < len(a); i++ {
		if a[i] != b[i] {
			return false
		}
	}

	return true
}

// ContainString returns whether s is in m
// Specially:
// if len(m) == 0, always return false
func ContainString(m []string, s string) bool {
	for _, a := range m {
		if s == a {
			return true
		}
	}
	return false
}

// EqualToLast returns whether s is equal to the last element in m
// if len(m) == 0, return false;
func EqualToLast(m []string, s string) bool {
	if len(m) == 0 {
		return false
	}
	return s == m[len(m)-1]
}
