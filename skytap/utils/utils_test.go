package utils

import (
	"testing"

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
