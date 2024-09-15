package check_test

import (
	"testing"

	"github.com/jimmysawczuk/kit/check"
	"github.com/stretchr/testify/require"
)

func TestCheck(t *testing.T) {
	type V struct {
		Foo string `check:"required"`
	}

	tests := []struct {
		In    V
		Valid bool
	}{
		{
			In:    V{Foo: "Bar"},
			Valid: true,
		},
		{
			In:    V{Foo: ""},
			Valid: false,
		},
	}

	for _, test := range tests {
		err := check.Check(test.In)

		require.Equal(t, test.Valid, err == nil)
	}
}
