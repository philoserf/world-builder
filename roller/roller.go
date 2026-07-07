// Package roller provides dice-rolling abstractions used throughout the world-builder module.
//
// Every random draw in the library passes through a Roller. This makes
// seeded reproducibility and scripted-test injection both straightforward.
package roller

import (
	"fmt"
	"hash/fnv"
	"math/rand"

	"github.com/philoserf/world-builder/dice"
)

// Roller is the interface every dice-driven procedure depends on.
type Roller interface {
	// Roll executes the given dice notation (e.g. "2D", "2D-7", "d10")
	// and returns the result, including any modifier in the notation.
	Roll(notation string) int

	// Fork returns a child Roller whose stream is deterministically
	// derived from this Roller's identity and key. Forking never
	// consumes a draw from the parent, so a parent and its forks are
	// independent: rolls taken from one do not perturb the other.
	//
	// SPIKE C1 (docs/rebuild-spec.md § C1). The pipeline uses Fork to
	// give every (body, procedure-family) its own substream keyed by
	// stable body identity, so reordering a stage or re-rolling one body
	// (the tidal-lock cascade) leaves every other body's values byte-
	// identical. Seeded branches; Scripted and Fixed are transparent
	// (return a view over the same sequence / value) so per-procedure
	// worked-example fixtures that feed a flat dice list are unaffected.
	Fork(key string) Roller
}

// Seeded is a production roller backed by a seeded *math/rand.Rand.
type Seeded struct {
	seed int64 // immutable construction seed; Fork derives children from it
	rng  *rand.Rand
}

// NewSeeded constructs a Seeded roller with the given seed.
func NewSeeded(seed int64) *Seeded {
	//nolint:gosec // math/rand is intentional; we are not generating crypto material.
	return &Seeded{seed: seed, rng: rand.New(rand.NewSource(seed))}
}

// Fork derives an independent child Seeded from this roller's immutable
// construction seed and key. Deterministic in (seed, key) and
// independent of how many draws the parent has taken — so forking is
// position-independent: perturbing the parent's stream never shifts a
// fork. See the Roller.Fork contract and docs/rebuild-spec.md § C1.
func (s *Seeded) Fork(key string) Roller {
	h := fnv.New64a()
	// Mix the immutable seed (not the live rng state) so the child is
	// stable regardless of parent consumption.
	var b [8]byte
	u := uint64(s.seed)
	for i := range b {
		b[i] = byte(u >> (8 * i))
	}
	_, _ = h.Write(b[:])
	_, _ = h.Write([]byte(key))
	//nolint:gosec // hash→seed conversion; not crypto material.
	return NewSeeded(int64(h.Sum64()))
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

// Fork returns the same Scripted, so forks share one flat result
// sequence consumed in call order. This keeps worked-example fixtures
// that feed a single narrated dice list unaffected by the pipeline's
// per-body forking — the topology is invisible to a scripted stream.
func (s *Scripted) Fork(string) Roller { return s }

// Fixed is a roller that always returns the same value. Useful for
// property tests where you want to pin one variable while exercising
// others, or for deterministic property-style assertions.
type Fixed int

// Roll implements the Roller interface.
func (f Fixed) Roll(string) int { return int(f) }

// Fork returns the same Fixed value; a pinned variable stays pinned
// across the fork tree.
func (f Fixed) Fork(string) Roller { return f }
