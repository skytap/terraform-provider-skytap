package skytap

// stString returns a string value for the passed string pointer.
// It returns the empty string if the pointer is nil.
func stString(s *string) string {
	if s != nil {
		return *s
	}
	return ""
}

// stStringPtr returns a pointer to the passed string.
func stStringPtr(s string) *string {
	return &s
}

// stInt returns an int value for the passed int pointer.
// It returns 0 if the pointer is nil.
func stInt(i *int) int {
	if i != nil {
		return *i
	}
	return 0
}

// stIntPtr returns a pointer to the passed int.
func stIntPtr(i int) *int {
	return &i
}
