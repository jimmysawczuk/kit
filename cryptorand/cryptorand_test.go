package cryptorand_test

import (
	"testing"

	"github.com/jimmysawczuk/kit/cryptorand"
	"github.com/stretchr/testify/assert"
)

func TestUint64NoPanic(t *testing.T) {
	src := cryptorand.Source()

	assert.NotPanics(t, func() {
		for range 100 {
			src.Uint64()
		}
	})
}
