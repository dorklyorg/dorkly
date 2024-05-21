package ldsampling

import (
	"math/rand"
	"time"
)

// NewSampler creates a *RatioSampler instance that can be used to
// determine sampling selections.
//
// The random number generator used is seeded with the current system time.
func NewSampler() *RatioSampler {
	return NewSamplerFromSource(rand.NewSource(time.Now().UnixNano()))
}

// NewSamplerFromSource creates a *RatioSampler instance similar to
// NewSampler with the additional benefit of providing the random number
// source.
func NewSamplerFromSource(source rand.Source) *RatioSampler {
	return &RatioSampler{rng: rand.New(source)} //nolint:gosec // doesn't need cryptographic security
}

// RatioSampler provides a simple interface for determining sample selections.
//
// The distribution calculation relies on random number generation, it does not
// perform any tracking to ensure a strict 1 in x outcome.
//
// A non-positive ratio effectively disables sampling, resulting in Sample always
// returning false. A ratio of 1 will result in every call being true. Any other
// ratio results in a 1 in x chance of being sampled.
//
// As a RatioSampler relies on rand.Source, the sampler is not safe for
// concurrent use.
type RatioSampler struct {
	rng *rand.Rand
}

// Sample returns a boolean to determine whether or not something should be
// sampled. It should not be called concurrently.
func (r *RatioSampler) Sample(ratio int) bool {
	if ratio <= 0 {
		return false
	}

	if ratio == 1 {
		return true
	}

	return r.rng.Float64() < 1/float64(ratio)
}
