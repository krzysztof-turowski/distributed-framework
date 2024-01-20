package lib

import (
	"fmt"
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

type uniquenessGenerator struct {
	provider   func(int) int
	invocation int
	used       map[int]int
}

func GetUniquenessGenerator(provider func(int) int) Generator {
	return &uniquenessGenerator{
		provider:   provider,
		invocation: 0,
		used:       make(map[int]int),
	}
}

func (r *uniquenessGenerator) Int() int {
	r.invocation++
	value := r.provider(r.invocation)
	if invocation, present := r.used[value]; present {
		panic(
			fmt.Sprint(
				"The provided values were not unique: ",
				value,
				" occured both in invocations ",
				invocation,
				" and ",
				r.invocation,
			),
		)
	}
	r.used[value] = r.invocation
	return value
}
