package server

import (
	"image"
	"strings"
)

// Decode the metadata of a complete name emitted by sub_5435C0. GAME.EXE's
// 00505360 keeps the full name and tokenizes a separate copy with CRT strtok.
// The ordinal stays a string; the two offsets are separate signed dwords.
func scriptFunctionQualifier505360(name string) (string, image.Point, bool) {
	if end := strings.IndexByte(name, 0); end >= 0 {
		name = name[:end]
	}
	if name != "" && !strings.ContainsRune(name, '%') {
		return "", image.Point{}, true
	}
	words := strings.FieldsFunc(name, func(c rune) bool { return c == '%' })
	if len(words) < 4 {
		// Missing tokens can reach atoi(NULL) in the original. Do not invent
		// a normal original zero-fill result for those malformed records.
		return "", image.Point{}, false
	}
	return "%" + words[1], image.Pt(
		int(scriptQualifierDecimal505360(words[2])),
		int(scriptQualifierDecimal505360(words[3])),
	), true
}

// This is the file-initialized CRT's ASCII decimal-prefix contract. Dword
// accumulation and negation wrap; a process changing its locale is not modeled.
func scriptQualifierDecimal505360(word string) int32 {
	index := 0
	for index < len(word) {
		switch word[index] {
		case ' ', '\t', '\n', '\v', '\f', '\r':
			index++
			continue
		}
		break
	}
	negative := false
	if index < len(word) {
		switch word[index] {
		case '-':
			negative = true
			index++
		case '+':
			index++
		}
	}
	var value uint32
	for index < len(word) && word[index] >= '0' && word[index] <= '9' {
		value = value*10 + uint32(word[index]-'0')
		index++
	}
	if negative {
		value = -value
	}
	return int32(value)
}
