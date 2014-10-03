package balancer

import (
	"math"
	"math/rand"
)

type Backends []*Backend

// Up returns all backends that are up
func (be Backends) Up() Backends {
	return be.all(func(b *Backend) bool { return b.Up() })
}

// FirstUp returns the first backend that is up
func (be Backends) FirstUp() *Backend {
	return be.first(func(b *Backend) bool { return b.Up() })
}

// MinUp returns the backend with the minumum result that is up
func (be Backends) MinUp(minimum func(*Backend) int64) *Backend {
	min := int64(math.MaxInt64)
	pos := -1
	for n, b := range be {
		if b.Up() {
			if num := minimum(b); num < min {
				pos, min = n, num
			}
		}
	}

	if pos < 0 {
		return nil
	}
	return be[pos]
}

// Random returns a random backend
func (be Backends) Random() *Backend {
	if size := len(be); size > 0 {
		return be[rand.Intn(size)]
	}
	return nil
}

// WeightedRandom returns a weighted-random backend
func (be Backends) WeightedRandom(weight func(*Backend) int64) *Backend {
	if len(be) < 1 {
		return nil
	}

	var min, max int64 = math.MaxInt64, 0
	weights := make([]int64, len(be))
	for n, b := range be {
		w := weight(b)
		if w > max {
			max = w
		}
		if w < min {
			min = w
		}
		weights[n] = w
	}

	var sum int64
	for n, w := range weights {
		w = min + max - w
		sum = sum + w
		weights[n] = w
	}

	mark := rand.Int63n(sum)
	for n, w := range weights {
		if mark -= w; mark <= 0 {
			return be[n]
		}
	}

	// We should never reach this point if the slice wasn't empty
	return nil
}

// selects all backends given a criteria
func (be Backends) all(criteria func(*Backend) bool) Backends {
	res := make(Backends, 0, len(be))
	for _, b := range be {
		if criteria(b) {
			res = append(res, b)
		}
	}
	return res
}

// returns the first matching backend given a criteria, or nil when nothing matches
func (be Backends) first(criteria func(*Backend) bool) *Backend {
	for _, b := range be {
		if criteria(b) {
			return b
		}
	}
	return nil
}
