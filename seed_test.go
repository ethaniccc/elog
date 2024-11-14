package elog

import (
	"testing"
)

func VerifySeedFunction(f SeedFunc, initalSeed, times uint64) bool {
	seed, oldSeed := initalSeed, initalSeed
	for i := uint64(0); i < times; i++ {
		seed = f(seed)
		if oldSeed == seed {
			return false
		}

		oldSeed = seed
	}

	return true
}

func TestRandomSeedFunction(t *testing.T) {
	var seed_f SeedFunc = func(seed uint64) uint64 {
		seed *= 10
		if seed > 1_000_000 {
			seed -= 400_000
		}
		seed /= 2

		return seed
	}

	if !VerifySeedFunction(seed_f, 69420, 1_000_000_000) {
		panic("seed function is faulty")
	}
}
