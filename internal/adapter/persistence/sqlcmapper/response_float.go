package sqlcmapper

import "math"

func roundResponseFloat(value float64) float64 {
	const (
		scale   = 100
		epsilon = 1e-9
	)

	// Stabilize half-up rounding for binary float edge cases such as 1.005.
	rounded := math.Round((value+math.Copysign(epsilon, value))*scale) / scale
	if rounded == 0 {
		return 0
	}

	return rounded
}

func roundOptionalResponseFloat(value *float64) *float64 {
	if value == nil {
		return nil
	}

	rounded := roundResponseFloat(*value)
	return &rounded
}
