// Package roller provides dice-rolling abstractions used throughout wbh.
//
// Every random draw in the library passes through a Roller. This makes
// seeded reproducibility and scripted-test injection both straightforward.
package roller

import (
	"fmt"
	"math/rand"
	"wbh/dice"
)

// Roller is the interface every dice-driven procedure depends on.
type Roller interface {
	// Roll executes the given dice notation (e.g. "2D", "2D-7", "d10")
	// and returns the result, including any modifier in the notation.
	Roll(notation string) int
}

// Seeded is a production roller backed by a seeded *math/rand.Rand.
type Seeded struct {
	rng *rand.Rand
}

// NewSeeded constructs a Seeded roller with the given seed.
func NewSeeded(seed int64) *Seeded {
	//nolint:gosec // math/rand is intentional; we are not generating crypto material.
	return &Seeded{rng: rand.New(rand.NewSource(seed))}
}

// Roll implements the Roller interface.
func (s *Seeded) Roll(notation string) int {
	spec, err := dice.Parse(notation)
	if err != nil {
		panic(fmt.Errorf("roller.Seeded: %w", err))
	}
	total := spec.Modifier
	for range spec.Count {
		total += s.rng.Intn(spec.Sides) + 1
	}
	return total
}

// Scripted is a test roller that yields preset results in order.
//
// The scripted values are *final results* — the natural roll plus any
// modifier already applied at the call site if appropriate. This keeps
// book worked-examples readable: the book reports e.g. "a 2D roll of 9"
// and the test feeds 9 directly.
type Scripted struct {
	results []int
	idx     int
}

// NewScripted constructs a Scripted roller that returns the supplied
// values in order.
func NewScripted(results ...int) *Scripted {
	return &Scripted{results: results}
}

// Roll implements the Roller interface. Panics if the scripted sequence
// is exhausted; an exhausted Scripted in a test always indicates a bug
// in the test or the procedure being tested.
func (s *Scripted) Roll(notation string) int {
	if s.idx >= len(s.results) {
		panic(fmt.Sprintf("roller.Scripted: exhausted on Roll(%q)", notation))
	}
	v := s.results[s.idx]
	s.idx++
	return v
}

// Fixed is a roller that always returns the same value. Useful for
// property tests where you want to pin one variable while exercising
// others, or for deterministic property-style assertions.
type Fixed int

// Roll implements the Roller interface.
func (f Fixed) Roll(string) int { return int(f) }
