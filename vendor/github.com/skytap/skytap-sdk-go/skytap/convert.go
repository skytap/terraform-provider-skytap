package skytap

// String returns a string value for the passed string pointer.
// It returns the empty string if the pointer is nil.
func String(s *string) string {
	if s != nil {
		return *s
	}
	return ""
}

// StringPtr returns a pointer to the passed string.
func StringPtr(s string) *string {
	return &s
}

// Int returns an int value for the passed int pointer.
// It returns 0 if the pointer is nil.
func Int(i *int) int {
	if i != nil {
		return *i
	}
	return 0
}

// IntPtr returns a pointer to the passed int.
func IntPtr(i int) *int {
	return &i
}
