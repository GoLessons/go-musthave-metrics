package agent

import (
	"math/rand"
	"time"
)

type Randomizer struct {
}

func NewRandomizer() *Randomizer {
	return &Randomizer{}
}

func (r *Randomizer) Randomize() GaugeValue {
	rand.Seed(time.Now().UnixNano())
	return GaugeValue(rand.Float64())
}
