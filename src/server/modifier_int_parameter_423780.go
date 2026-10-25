package server

// modifierIntParameterValue423780 preserves the original decimal atoi prefix
// and its modulo-32-bit accumulation. Missing input is -1; a present token
// without an initial signed decimal integer is zero.
func modifierIntParameterValue423780(word string) int32 {
	if word == "" {
		return -1
	}
	i := 0
	negative := word[i] == '-'
	if negative || word[i] == '+' {
		i++
	}
	var value uint32
	for i < len(word) && word[i] >= '0' && word[i] <= '9' {
		value = value*10 + uint32(word[i]-'0')
		i++
	}
	if negative {
		value = 0 - value
	}
	return int32(value)
}
