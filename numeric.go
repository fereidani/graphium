package graphium

import "math"

// Number is the type constraint for numeric edge weights usable by the generic
// shortest-path algorithms. It admits every integer and floating-point kind so
// that callers may use the most natural weight type.
type Number interface {
	~int | ~int8 | ~int16 | ~int32 | ~int64 |
		~uint | ~uint8 | ~uint16 | ~uint32 | ~uint64 | ~uintptr |
		~float32 | ~float64
}

// Float is the constraint for floating-point edge weights. Bellman-Ford, which
// detects negative cycles, is restricted to floats to mirror petgraph's
// FloatMeasure and to avoid integer-overflow hazards at the unreachable
// sentinel.
type Float interface {
	~float32 | ~float64
}

// floatInf returns positive infinity for the floating-point type T. It is the
// sentinel meaning "no path known yet" used by Bellman-Ford.
func floatInf[T Float]() T {
	var z T
	switch any(z).(type) {
	case float32:
		return T(math.Float32frombits(0x7f800000)) // +Inf
	default:
		return T(math.Inf(1))
	}
}

// zero returns the zero value of T, used as the source distance in Dijkstra and
// as the identity for path lengths.
func zero[T any]() T {
	var z T
	return z
}
