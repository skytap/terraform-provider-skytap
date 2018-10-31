package skytap

// ptrToStr returns a string value for the passed string pointer.
// It returns the empty string if the pointer is nil.
func ptrToStr(s *string) string {
	if s != nil {
		return *s
	}
	return ""
}

// strToPtr returns a pointer to the passed string.
func strToPtr(s string) *string {
	return &s
}

// ptrToInt returns an int value for the passed int pointer.
// It returns 0 if the pointer is nil.
func ptrToInt(i *int) int {
	if i != nil {
		return *i
	}
	return 0
}

// intToPtr returns a pointer to the passed int.
func intToPtr(i int) *int {
	return &i
}

// boolToPtr returns a pointer to the passed bool.
func boolToPtr(i bool) *bool {
	return &i
}

// networkTypeToPtr returns a pointer to the passed NetworkType.
func networkTypeToPtr(networkType NetworkType) *NetworkType {
	return &networkType
}

// vmRunStateToPtr returns a pointer to the passed VMRunstate.
func vmRunStateToPtr(vmRunState VMRunstate) *VMRunstate {
	return &vmRunState
}

// nicTypeToPtr returns a pointer to the passed NICType.
func nicTypeToPtr(nicType NICType) *NICType {
	return &nicType
}
