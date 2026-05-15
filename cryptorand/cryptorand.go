// Package cryptorand is adapted from https://yourbasic.org/golang/crypto-rand-int/.
package cryptorand

import (
	crand "crypto/rand"
	"encoding/binary"
	mrand "math/rand/v2"
)

type cryptoSource struct{}

func (s cryptoSource) Uint64() (v uint64) {
	err := binary.Read(crand.Reader, binary.BigEndian, &v)
	if err != nil {
		panic(err)
	}
	return v
}

func Source() mrand.Source {
	return cryptoSource{}
}

func New() *mrand.Rand {
	return mrand.New(Source())
}
