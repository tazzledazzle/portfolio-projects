package main

import (
	"sync"
	"testing"
	"time"
)

func TestFlakeScore_InitialState(t *testing.T) {
	fs := NewFlakeScore()
	if got := fs.Score(); got != 0 {
		t.Errorf("NewFlakeScore().Score() = %v, want 0", got)
	}
}

func TestFlakeScore_SinglePass(t *testing.T) {
	fs := NewFlakeScore()
	fs.Update(true)
	score := fs.Score()
	if score >= 10 {
		t.Errorf("Score after single pass = %v, want < 10 (low flake)", score)
	}
}

func TestFlakeScore_SingleFail(t *testing.T) {
	fs := NewFlakeScore()
	fs.Update(false)
	score := fs.Score()
	if score <= 40 {
		t.Errorf("Score after single fail = %v, want > 40 (high flake)", score)
	}
}

func TestFlakeScore_FlipSequence(t *testing.T) {
	fs := NewFlakeScore()
	fs.Update(true)
	fs.Update(false)
	fs.Update(true)
	fs.Update(false)
	score := fs.Score()
	if score < 40 || score > 60 {
		t.Errorf("Score after pass,fail,pass,fail = %v, want 40-60", score)
	}
}

func TestFlakeScore_DecayBounds(t *testing.T) {
	fs := NewFlakeScore()
	fs.Update(true)
	fs.Update(true)

	fs.SetLastUpdate(time.Now().Add(-30 * 24 * time.Hour))

	fs.Update(true)
	score := fs.Score()

	if score != score {
		t.Error("Score is NaN after decay")
	}
	if score < 0 || score > 100 {
		t.Errorf("Score = %v, want 0-100 (no Inf)", score)
	}
}

func TestFlakeScore_QuarantineThreshold(t *testing.T) {
	fs := NewFlakeScore()
	fs.Update(false)
	fs.Update(false)

	if fs.Score() <= 50 {
		t.Skip("Score not > 50, cannot test quarantine threshold")
	}

	if !fs.IsQuarantined() {
		t.Error("IsQuarantined() = false, want true when Score() > 50")
	}
}

func TestFlakeScore_ConcurrentUpdate(t *testing.T) {
	fs := NewFlakeScore()
	var wg sync.WaitGroup
	for i := 0; i < 50; i++ {
		wg.Add(1)
		go func(pass bool) {
			defer wg.Done()
			fs.Update(pass)
		}(i%2 == 0)
	}
	wg.Wait()

	score := fs.Score()
	if score != score {
		t.Error("Score is NaN after concurrent updates")
	}
}
