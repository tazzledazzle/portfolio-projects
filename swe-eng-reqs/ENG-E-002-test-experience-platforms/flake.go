package main

import (
	"math"
	"sync"
	"time"
)

const (
	decayHalfLife       = 14 * 24 * time.Hour
	quarantineThreshold = 50.0
	minCount            = 0.0
)

type FlakeScore struct {
	passes     float64
	fails      float64
	lastUpdate time.Time
	mu         sync.Mutex
}

func NewFlakeScore() *FlakeScore {
	return &FlakeScore{
		passes:     0,
		fails:      0,
		lastUpdate: time.Now(),
	}
}

func (fs *FlakeScore) Update(passed bool) {
	fs.mu.Lock()
	defer fs.mu.Unlock()

	fs.applyDecay()

	if passed {
		fs.passes++
	} else {
		fs.fails++
	}

	fs.lastUpdate = time.Now()
}

func (fs *FlakeScore) applyDecay() {
	elapsed := time.Since(fs.lastUpdate)
	if elapsed <= 0 {
		return
	}

	decayFactor := math.Pow(0.5, float64(elapsed)/float64(decayHalfLife))

	fs.passes = math.Max(minCount, fs.passes*decayFactor)
	fs.fails = math.Max(minCount, fs.fails*decayFactor)
}

func (fs *FlakeScore) Score() float64 {
	fs.mu.Lock()
	defer fs.mu.Unlock()

	total := fs.passes + fs.fails
	if total == 0 {
		return 0
	}

	score := (fs.fails / total) * 100

	if math.IsNaN(score) || math.IsInf(score, 0) {
		return 0
	}

	return score
}

func (fs *FlakeScore) IsQuarantined() bool {
	return fs.Score() > quarantineThreshold
}

func (fs *FlakeScore) SetLastUpdate(t time.Time) {
	fs.mu.Lock()
	defer fs.mu.Unlock()
	fs.lastUpdate = t
}
