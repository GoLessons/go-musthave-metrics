package agent

import (
	"math/rand"
)

type Randomizer struct {
}

func NewRandomizer() *Randomizer {
	return &Randomizer{}
}

func (r *Randomizer) Randomize() GaugeValue {
	return GaugeValue(rand.Float64())
}
