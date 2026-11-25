package legacycrt

import (
	"math"
	"math/big"
	"math/bits"
	"strings"
	"sync"
)

// Atof converts the decimal prefix accepted at GAME.EXE 005668EE.
func Atof(text string) float64 {
	digits, scale, negative := atofMantissa(text)
	var sign uint64
	if negative {
		sign = 1 << 63
	}
	// Even the largest 24-digit mantissa underflows binary64 at this scale.
	if len(digits) == 0 || scale < -350 {
		return math.Float64frombits(sign)
	}
	if scale > 308 {
		return math.Float64frombits(sign | 0x7ff0000000000000)
	}

	value := atofNormalize(digits)
	if scale != 0 {
		value.tail = 0
		powers := atofDecimalPowers()
		direction := 0
		if scale < 0 {
			direction, scale = 1, -scale
		}
		for place := 0; scale != 0; place++ {
			if digit := scale % 8; digit != 0 {
				value = value.multiply(powers[direction][place][digit-1])
			}
			scale /= 8
		}
	}
	return math.Float64frombits(sign | value.binary64())
}

// AtofFloat32 applies the game's startup rounding toward zero to the
// binary32 store following an atof call.
func AtofFloat32(text string) float32 {
	value := Atof(text)
	out := float32(value)
	if value > 0 && float64(out) > value || value < 0 && float64(out) < value {
		out = math.Nextafter32(out, 0)
	}
	return out
}

func atofMantissa(text string) (digits []byte, scale int, negative bool) {
	text = strings.TrimLeft(text, " \t\n\v\f\r")
	if text == "" {
		return nil, 0, false
	}
	negative = text[0] == '-'
	if text[0] == '+' || negative {
		text = text[1:]
	}

	var significant [25]byte
	var count, offset int
	var fractional, hasDigit bool
	for offset < len(text) {
		ch := text[offset]
		if ch == '.' && !fractional {
			fractional = true
			offset++
			continue
		}
		if digit := ch - '0'; digit < 10 {
			hasDigit = true
			switch {
			case count < len(significant):
				if count != 0 || digit != 0 {
					significant[count] = digit
					count++
				}
				if fractional {
					scale--
				}
			case !fractional:
				scale++
			}
			offset++
			continue
		}
		break
	}
	if !hasDigit {
		return nil, 0, false
	}
	if count == len(significant) {
		// The original adjusts the last retained digit by testing that digit.
		count--
		if significant[count-1] >= 5 {
			significant[count-1]++
		}
		scale++
	}
	for count > 0 && significant[count-1] == 0 {
		count--
		scale++
	}
	return significant[:count], scale + atofExponent(text[offset:]), negative
}

func atofExponent(text string) int {
	if len(text) == 0 || text[0] != 'E' && text[0] != 'e' && text[0] != 'D' && text[0] != 'd' {
		return 0
	}
	text = text[1:]
	if text == "" {
		return 0
	}
	negative := text[0] == '-'
	if text[0] == '+' || negative {
		text = text[1:]
	}
	var exponent int
	for i := 0; i < len(text); i++ {
		digit := text[i] - '0'
		if digit >= 10 {
			break
		}
		exponent = 10*exponent + int(digit)
		if exponent > 5200 {
			exponent = 5201
			break
		}
	}
	if negative {
		return -exponent
	}
	return exponent
}

type atofExtended struct {
	significand uint64
	tail        uint16
	exponent    int
}

func atofNormalize(digits []byte) atofExtended {
	var high, low uint64
	for _, digit := range digits {
		productHigh, productLow := bits.Mul64(low, 10)
		var carry uint64
		low, carry = bits.Add64(productLow, uint64(digit), 0)
		high = high*10 + productHigh + carry
	}
	length := bits.Len64(low)
	if high != 0 {
		length = 64 + bits.Len64(high)
	}
	shift := 80 - length
	if shift >= 64 {
		high = low << (shift - 64)
	} else {
		high = high<<shift | low>>(64-shift)
	}
	low <<= shift
	return atofExtended{significand: high<<48 | low>>16, tail: uint16(low), exponent: length - 1}
}

// The original scales the significand by successive base-eight digits of
// the decimal exponent. These powers round to the same 80-bit values.
var atofDecimalPowers = sync.OnceValue(func() *[2][3][7]atofExtended {
	var powers [2][3][7]atofExtended
	for place, step := range []int64{1, 8, 64} {
		for digit := range powers[0][place] {
			magnitude := new(big.Int).Exp(big.NewInt(10), big.NewInt(int64(digit+1)*step), nil)
			var value big.Float
			value.SetPrec(80).SetInt(magnitude)
			powers[0][place][digit] = atofPower(&value)
			value.Quo(big.NewFloat(1), new(big.Float).SetInt(magnitude))
			powers[1][place][digit] = atofPower(&value)
		}
	}
	return &powers
})

func atofPower(value *big.Float) atofExtended {
	var mantissa big.Float
	exponent := value.MantExp(&mantissa) - 1
	mantissa.SetMantExp(&mantissa, 80)
	integer, _ := mantissa.Int(nil)
	tail := uint16(integer.Uint64())
	return atofExtended{significand: integer.Rsh(integer, 16).Uint64(), tail: tail, exponent: exponent}
}

func (v atofExtended) words() [5]uint16 {
	return [5]uint16{
		v.tail, uint16(v.significand), uint16(v.significand >> 16),
		uint16(v.significand >> 32), uint16(v.significand >> 48),
	}
}

func (v atofExtended) multiply(power atofExtended) atofExtended {
	// 0056DE5C accumulates the upper diagonals of the 16-bit word product.
	// The retained terms occupy six words before normalization and rounding.
	var product [6]uint64
	a, b := v.words(), power.words()
	for i, left := range a {
		for j := 4 - i; j < len(b); j++ {
			product[i+j-4] += uint64(left) * uint64(b[j])
		}
	}
	for i := 0; i < len(product)-1; i++ {
		product[i+1] += product[i] >> 16
		product[i] &= 0xffff
	}
	high := product[5]<<48 | product[4]<<32 | product[3]<<16 | product[2]
	low := uint32(product[1]<<16 | product[0])
	exponent := v.exponent + power.exponent + 1
	if high < 1<<63 {
		high = high<<1 | uint64(low>>31)
		low <<= 1
		exponent--
	}
	tail, discarded := uint16(low>>16), uint16(low)
	if discarded > 0x8000 || discarded == 0x8000 && tail&1 != 0 {
		tail++
		if tail == 0 {
			high++
			if high == 0 {
				high = 1 << 63
				exponent++
			}
		}
	}
	return atofExtended{significand: high, tail: tail, exponent: exponent}
}

func (v atofExtended) binary64() uint64 {
	tail := uint32(v.tail) << 16
	value, carry := atofRound53(v.significand, tail)
	exponent := v.exponent + int(carry)
	switch {
	case exponent > 1023:
		return 0x7ff0000000000000
	case exponent < -1076:
		return 0
	case exponent < -1022:
		shift := -1023 - exponent
		high := v.significand >> shift
		if shift <= 32 {
			tail = tail>>shift | uint32(v.significand<<(32-shift))
		} else {
			tail = uint32(v.significand >> (shift - 32))
		}
		value, _ := atofRound53(high, tail)
		return value >> 1
	default:
		return uint64(exponent+1023)<<52 | value&((1<<52)-1)
	}
}

func atofRound53(high uint64, tail uint32) (value, carry uint64) {
	const guard = 1 << 10
	// 0056BE23 checks the guard and the bits below the next lower bit.
	if high&guard != 0 && (high&(guard/2-1) != 0 || tail != 0) {
		high, carry = bits.Add64(high, guard*2, 0)
	}
	return high >> 11, carry
}
