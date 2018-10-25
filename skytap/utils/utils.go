package utils

import "github.com/skytap/skytap-sdk-go/skytap"

// String returns a pointer to a string literal
func String(s string) *string {
	return &s
}

// Int returns a pointer to an int literal
func Int(v int) *int {
	return &v
}

// NetworkType returns a pointer to a NetworkType literal
func NetworkType(networkType skytap.NetworkType) *skytap.NetworkType {
	return &networkType
}
