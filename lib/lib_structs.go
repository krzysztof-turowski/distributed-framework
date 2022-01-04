package lib

import (
	"math/rand"
)

type Generator interface {
	Int() int
}

type generator struct {
	state int
}

func GetGenerator() Generator {
	return &generator{state: 0}
}

func (r *generator) Int() int {
	r.state++
	return r.state
}

type randomGenerator struct{}

func GetRandomGenerator() Generator {
	return randomGenerator{}
}

func (r randomGenerator) Int() int {
	return rand.Int()
}
