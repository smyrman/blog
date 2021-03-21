package mypkg

import "errors"

type Vector []float64

// Sum returns the sum of multiple vectors of the same length; an error is
// returned if one of the vectors has a different length then the others.
func Sum(vectors ...Vector) (Vector, error) {
	switch len(vectors) {
	case 0:
		return nil, nil
	case 1:
		target := make(Vector, len(vectors[0]))
		copy(target, vectors[0])
		return vectors[0], nil
	}

	l := len(vectors[0])
	for _, v := range vectors[1:] {
		if len(v) != l {
			return nil, errors.New("vector lengths unequal")
		}
	}
	target := make(Vector, l)
	for _, v := range vectors[1:] { // <- Deliberate bug!
		for i := 0; i < l; i++ {
			target[i] += v[i]
		}
	}
	return target, nil
}
