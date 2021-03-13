// Package cryptorand is adapted from https://yourbasic.org/golang/crypto-rand-int/.
package cryptorand

import (
	crand "crypto/rand"
	"encoding/binary"
	mrand "math/rand"
)

type cryptoSource struct{}

func (s cryptoSource) Seed(seed int64) {}

func (s cryptoSource) Int63() int64 {
	return int64(s.Uint64() & ^uint64(1<<63))
}

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
