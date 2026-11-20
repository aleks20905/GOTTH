package auth

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSignVerifyRoundtrip(t *testing.T) {
	signed := Sign("secret", "value")

	value, ok := Verify("secret", signed)

	assert.True(t, ok)
	assert.Equal(t, "value", value)
}

func TestVerifyRejections(t *testing.T) {
	signed := Sign("secret", "value")

	_, ok := Verify("other-key", signed)
	assert.False(t, ok)

	_, ok = Verify("secret", "value")
	assert.False(t, ok)

	_, ok = Verify("secret", "value.deadbeef")
	assert.False(t, ok)
}
