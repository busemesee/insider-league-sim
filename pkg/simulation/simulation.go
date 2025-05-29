package simulation

import (
    "math/rand"
    "time"
)

// Simulator interface: allows different match simulation algorithms
type Simulator interface {
    PlayMatch(homeLevel, awayLevel int) (homeGoals, awayGoals int)
}

// SimpleSimulator: strength-based random score generator
type SimpleSimulator struct {
    rnd *rand.Rand
}

func NewSimpleSimulator() *SimpleSimulator {
    return &SimpleSimulator{rnd: rand.New(rand.NewSource(time.Now().UnixNano()))}
}

func (s *SimpleSimulator) PlayMatch(homeLevel, awayLevel int) (int, int) {
    avgHome := float64(homeLevel) / float64(homeLevel+awayLevel) * 3.0
    avgAway := float64(awayLevel) / float64(homeLevel+awayLevel) * 3.0
    home := int(s.rnd.NormFloat64()*1.0 + avgHome)
    away := int(s.rnd.NormFloat64()*1.0 + avgAway)
    if home < 0 {
        home = 0
    }
    if away < 0 {
        away = 0
    }
    return home, away
}
