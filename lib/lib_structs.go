package lib

import (
	"math/rand"
	"time"
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

type randomGenerator struct {
	state *rand.Rand
}

func GetRandomGenerator() Generator {
	return &randomGenerator{state: rand.New(rand.NewSource(time.Now().UnixNano()))}
}

func (r *randomGenerator) Int() int {
	return r.state.Int()
}
