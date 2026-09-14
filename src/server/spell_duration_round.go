package server

import "math"

// spellDurationRoundNearestEven models the original x87 FISTP conversion.
// Invalid or out-of-range operands produce the integer-indefinite value.
func spellDurationRoundNearestEven(value float32) int32 {
	if math.IsNaN(float64(value)) || value >= 2147483648 || value < -2147483648 {
		return math.MinInt32
	}
	return int32(math.RoundToEven(float64(value)))
}
