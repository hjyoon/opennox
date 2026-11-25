package legacycrt

import (
	"math"
	"strings"
)

// ScanFloat32 reads one %f field and returns the unconsumed input.
// GAME.EXE 0056A63C converts the scanned field directly to binary32.
func ScanFloat32(text string) (value float32, rest string, ok bool) {
	text = strings.TrimLeft(text, asciiWhitespace)
	// The original scanf producer holds 349 characters and a terminator.
	word := text[:min(len(text), 349)]
	start := skipFloatSign(word, 0)
	end := scanFloatDigits(word, start)
	hasDigit := end > start
	if end < len(word) && word[end] == '.' {
		start = end + 1
		end = scanFloatDigits(word, start)
		hasDigit = hasDigit || end > start
	}
	if !hasDigit {
		return 0, text, false
	}
	if end < len(word) && (word[end] == 'e' || word[end] == 'E') {
		// Scanning consumes the exponent marker and sign even when it
		// finds no exponent digits.
		end = scanFloatDigits(word, skipFloatSign(word, end+1))
	}
	return math.Float32frombits(uint32(decimalBits(word[:end], 24, 127, 1<<31))), text[end:], true
}

func skipFloatSign(word string, offset int) int {
	if offset < len(word) && (word[offset] == '+' || word[offset] == '-') {
		return offset + 1
	}
	return offset
}

func scanFloatDigits(word string, offset int) int {
	for offset < len(word) && word[offset]-'0' < 10 {
		offset++
	}
	return offset
}
