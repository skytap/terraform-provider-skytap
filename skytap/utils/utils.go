package utils

import (
	"github.com/skytap/skytap-sdk-go/skytap"
)

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

// VMRunstate returns a pointer to a VMRunstate literal
func VMRunstate(runstate skytap.VMRunstate) *skytap.VMRunstate {
	return &runstate
}

// NICType returns a pointer to a NICType literal
func NICType(nicType skytap.NICType) *skytap.NICType {
	return &nicType
}
