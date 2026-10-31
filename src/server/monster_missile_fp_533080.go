package server

import "math"

// Memory operands retain SNaN identity until the original arithmetic
// instruction, while FLD quiets its input. Avoid host conversion here.
func monsterMissileMemoryF32_533080(value float32) float64 {
	word := math.Float32bits(value)
	if word&0x7f800000 == 0x7f800000 && word&0x007fffff != 0 {
		return math.Float64frombits(uint64(word&0x80000000)<<32 | 0x7ff0000000000000 | uint64(word&0x007fffff)<<29)
	}
	return float64(value)
}

func monsterMissileLoadF32_533080(value float32) float64 {
	x := monsterMissileMemoryF32_533080(value)
	word := math.Float64bits(x)
	if word&0x7fffffffffffffff > 0x7ff0000000000000 {
		return math.Float64frombits(word | 0x0008000000000000)
	}
	return x
}

// The masked x87 operation prefers QNaN to SNaN; otherwise it chooses the
// larger significand and then the positive sign for equal significands.
func monsterMissileNaN533080(a, b float64) (float64, bool) {
	ax, bx := math.Float64bits(a), math.Float64bits(b)
	an, bn := ax&0x7fffffffffffffff > 0x7ff0000000000000, bx&0x7fffffffffffffff > 0x7ff0000000000000
	if !an && !bn {
		return 0, false
	}
	chosen := ax
	if !an {
		chosen = bx
	} else if bn {
		aq, bq := ax&0x0008000000000000 != 0, bx&0x0008000000000000 != 0
		if !aq && bq || aq == bq && (bx&0x000fffffffffffff > ax&0x000fffffffffffff || bx&0x000fffffffffffff == ax&0x000fffffffffffff && bx < ax) {
			chosen = bx
		}
	}
	return math.Float64frombits(chosen | 0x0008000000000000), true
}

func monsterMissileInvalid533080(value float64) float64 {
	if math.IsNaN(value) {
		return math.Float64frombits(0xfff8000000000000)
	}
	return value
}

func monsterMissileAdd533080(a, b float64) float64 {
	if x, ok := monsterMissileNaN533080(a, b); ok {
		return x
	}
	return monsterMissileInvalid533080(pointOnLineAddChop53_57C8A0(a, b))
}

func monsterMissileSub533080(a, b float64) float64 {
	if x, ok := monsterMissileNaN533080(a, b); ok {
		return x
	}
	return monsterMissileInvalid533080(pointOnLineAddChop53_57C8A0(a, -b))
}

func monsterMissileMul533080(a, b float64) float64 {
	if x, ok := monsterMissileNaN533080(a, b); ok {
		return x
	}
	return monsterMissileInvalid533080(pointOnLineMulChop53_57C8A0(a, b))
}

func monsterMissileDiv533080(a, b float64) float64 {
	if x, ok := monsterMissileNaN533080(a, b); ok {
		return x
	}
	return monsterMissileInvalid533080(pointOnLineDivChop53_57C8A0(a, b))
}

func monsterMissileSqrt533080(a float64) float64 {
	if x, ok := monsterMissileNaN533080(a, a); ok {
		return x
	}
	return monsterMissileInvalid533080(monsterMoveForceSqrtChop53_50D50C(a))
}

func monsterMissileSpill533080(a float64) float32 {
	word := math.Float64bits(a)
	if word&0x7fffffffffffffff > 0x7ff0000000000000 {
		return math.Float32frombits(uint32(word>>32)&0x80000000 | 0x7fc00000 | uint32(word>>29)&0x007fffff)
	}
	return pointOnLineSpill57C8A0(a)
}
