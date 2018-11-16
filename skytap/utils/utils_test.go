package utils

import (
	"testing"

	"github.com/skytap/skytap-sdk-go/skytap"

	"github.com/stretchr/testify/assert"
)

func TestString(t *testing.T) {
	v := ""
	assert.Equal(t, v, *String(v))
}

func TestInt(t *testing.T) {
	v := 1
	assert.Equal(t, v, *Int(v))
}

func TestNetworkType(t *testing.T) {
	v := skytap.NetworkTypeAutomatic
	assert.Equal(t, v, *NetworkType(v))
}

func TestRunstate(t *testing.T) {
	v := skytap.VMRunstateRunning
	assert.Equal(t, skytap.VMRunstateRunning, *VMRunstate(v))
}
