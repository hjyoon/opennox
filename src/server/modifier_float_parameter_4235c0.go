package server

import (
	"math"
	"strconv"
)

// modifierFloatParameterValue4235C0 preserves the original decimal atof prefix
// and the binary32 store at 4235D3 under startup rounding toward zero.
// A missing token is -1; a present token without a decimal prefix is zero.
func modifierFloatParameterValue4235C0(word string) float32 {
	if word == "" {
		return -1
	}
	i := 0
	if word[i] == '+' || word[i] == '-' {
		i++
	}
	digits := 0
	for i < len(word) && word[i] >= '0' && word[i] <= '9' {
		i++
		digits++
	}
	if i < len(word) && word[i] == '.' {
		i++
		for i < len(word) && word[i] >= '0' && word[i] <= '9' {
			i++
			digits++
		}
	}
	if digits == 0 {
		return 0
	}
	end := i
	if i < len(word) && (word[i] == 'e' || word[i] == 'E') {
		i++
		if i < len(word) && (word[i] == '+' || word[i] == '-') {
			i++
		}
		start := i
		for i < len(word) && word[i] >= '0' && word[i] <= '9' {
			i++
		}
		if i != start {
			end = i
		}
	}
	// The scanned prefix is syntactically valid. Range results retain their
	// signed infinity or zero, matching the original static CRT observations.
	value, _ := strconv.ParseFloat(word[:end], 64)
	out := float32(value)
	if (value > 0 && float64(out) > value) || (value < 0 && float64(out) < value) {
		out = math.Nextafter32(out, 0)
	}
	return out
}
