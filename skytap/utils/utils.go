package utils

// String returns a pointer to a string literal
func String(s string) *string {
	return &s
}

// Int returns a pointer to an int literal
func Int(v int) *int {
	return &v
}
