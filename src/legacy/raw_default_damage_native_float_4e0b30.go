package legacy

import (
	"encoding/binary"
	"math/big"
	"math/bits"
)

// This value models a numerical x87 register, not the host FPU environment.
// A finite value is (-1)^negative * significand * 2^exponent. Loads retain
// all input significand bits; only the actual arithmetic/store boundaries
// round them. NaNs and unsupported extended encodings never enter big.Int.
// Physical x87 status/register equivalence is outside this numerical helper.
type rawDefaultDamageFloat4E0B30 struct {
	kind                uint8
	negative            bool
	significand         uint64
	exponent            int32
	rawExtendedExponent uint16
}

const (
	rawDefaultDamageFloatFinite4E0B30 uint8 = iota
	rawDefaultDamageFloatInfinity4E0B30
	rawDefaultDamageFloatNaN4E0B30
	rawDefaultDamageFloatUnsupported4E0B30
)

func rawDefaultDamageFloatFromReply4E0B30(reply rawDefaultDamageNativeReply4E0B30) rawDefaultDamageFloat4E0B30 {
	switch reply.FloatFormat {
	case "f32":
		return rawDefaultDamageFloatFrom32Bits4E0B30(binary.LittleEndian.Uint32(reply.FloatPayload[:4]))
	case "f64":
		word := binary.LittleEndian.Uint64(reply.FloatPayload[:8])
		value := rawDefaultDamageFloat4E0B30{negative: word>>63 != 0}
		exponent := uint16(word >> 52 & 0x7ff)
		fraction := word & ((uint64(1) << 52) - 1)
		if exponent == 0x7ff {
			value.kind = rawDefaultDamageFloatInfinity4E0B30
			if fraction != 0 {
				value.kind = rawDefaultDamageFloatNaN4E0B30
				value.significand = uint64(1)<<63 | fraction<<11
			}
			return value
		}
		value.significand = fraction
		value.exponent = 1 - 1023 - 52
		if exponent != 0 {
			value.significand |= uint64(1) << 52
			value.exponent = int32(exponent) - 1023 - 52
		}
		return value
	case "ext80":
		significand := binary.LittleEndian.Uint64(reply.FloatPayload[:8])
		rawExponent := binary.LittleEndian.Uint16(reply.FloatPayload[8:10])
		exponent := rawExponent & 0x7fff
		value := rawDefaultDamageFloat4E0B30{
			negative: rawExponent>>15 != 0, significand: significand,
			rawExtendedExponent: rawExponent,
		}
		integerBit := significand >> 63
		fraction := significand & ((uint64(1) << 63) - 1)
		if exponent != 0 && integerBit == 0 {
			// Unnormals and pseudo-Inf/NaN remain distinguishable from
			// supported NaN/Inf; retain the original J bit and exponent.
			value.kind = rawDefaultDamageFloatUnsupported4E0B30
			return value
		}
		if exponent == 0x7fff {
			value.kind = rawDefaultDamageFloatInfinity4E0B30
			if fraction != 0 {
				value.kind = rawDefaultDamageFloatNaN4E0B30
			}
			return value
		}
		// E=0,J=1 pseudo-denormals denote the same finite value as
		// E=1,J=1. The full 64-bit significand is retained either way.
		value.exponent = 1 - 16383 - 63
		if exponent != 0 {
			value.exponent = int32(exponent) - 16383 - 63
		}
		return value
	default:
		panic("raw default damage float reply has no declared format")
	}
}

func rawDefaultDamageFloatFrom32Bits4E0B30(word uint32) rawDefaultDamageFloat4E0B30 {
	value := rawDefaultDamageFloat4E0B30{negative: word>>31 != 0}
	exponent := word >> 23 & 0xff
	fraction := word & 0x7fffff
	if exponent == 0xff {
		value.kind = rawDefaultDamageFloatInfinity4E0B30
		if fraction != 0 {
			value.kind = rawDefaultDamageFloatNaN4E0B30
			value.significand = uint64(1)<<63 | uint64(fraction)<<40
		}
		return value
	}
	value.significand = uint64(fraction)
	value.exponent = 1 - 127 - 23
	if exponent != 0 {
		value.significand |= uint64(1) << 23
		value.exponent = int32(exponent) - 127 - 23
	}
	return value
}

func rawDefaultDamageFloatNearest4E0B30(cw uint16) bool {
	switch cw {
	case 0x0e3f:
		return false
	case 0x023f:
		return true
	default:
		panic("raw default damage supports only the two sealed x87 control words")
	}
}

func rawDefaultDamageFloatIndefinite4E0B30() rawDefaultDamageFloat4E0B30 {
	return rawDefaultDamageFloat4E0B30{
		kind: rawDefaultDamageFloatNaN4E0B30, negative: true,
		significand: uint64(3) << 62, rawExtendedExponent: 0xffff,
	}
}

func rawDefaultDamageFloatQuiet4E0B30(value rawDefaultDamageFloat4E0B30) rawDefaultDamageFloat4E0B30 {
	if value.kind == rawDefaultDamageFloatUnsupported4E0B30 {
		return rawDefaultDamageFloatIndefinite4E0B30()
	}
	if value.kind == rawDefaultDamageFloatNaN4E0B30 {
		value.significand |= uint64(3) << 62
	}
	return value
}

// Return a positive magnitude on a fixed binary grid. Using a positive
// quotient/remainder avoids big.Int's arithmetic right shift for negatives.
func rawDefaultDamageFloatRoundGrid4E0B30(magnitude *big.Int, exponent, quantum int32, nearest bool) *big.Int {
	shift := quantum - exponent
	if shift <= 0 {
		return new(big.Int).Lsh(new(big.Int).Set(magnitude), uint(-shift))
	}
	quotient := new(big.Int).Rsh(new(big.Int).Set(magnitude), uint(shift))
	if !nearest {
		return quotient
	}
	remainder := new(big.Int).Sub(new(big.Int).Set(magnitude), new(big.Int).Lsh(new(big.Int).Set(quotient), uint(shift)))
	half := new(big.Int).Lsh(big.NewInt(1), uint(shift-1))
	comparison := remainder.Cmp(half)
	if comparison > 0 || comparison == 0 && quotient.Bit(0) != 0 {
		quotient.Add(quotient, big.NewInt(1))
	}
	return quotient
}

// PC selects precision; it does not narrow the extended exponent to binary64.
// Masked underflow is rounded on the selected format's subnormal grid, and
// masked overflow chooses Inf for nearest or maximum finite for toward-zero.
func rawDefaultDamageFloatRound4E0B30(magnitude *big.Int, exponent int32, negative bool, precision uint, minimumNormalExponent, maximumNormalExponent int32, cw uint16) rawDefaultDamageFloat4E0B30 {
	nearest := rawDefaultDamageFloatNearest4E0B30(cw)
	if magnitude.Sign() < 0 {
		panic("raw default damage float rounding requires a positive magnitude")
	}
	value := rawDefaultDamageFloat4E0B30{negative: negative}
	if magnitude.Sign() == 0 {
		return value
	}
	quantum := exponent + int32(magnitude.BitLen()) - int32(precision)
	minimumQuantum := minimumNormalExponent - int32(precision) + 1
	if quantum < minimumQuantum {
		quantum = minimumQuantum
	}
	rounded := rawDefaultDamageFloatRoundGrid4E0B30(magnitude, exponent, quantum, nearest)
	if rounded.Sign() == 0 {
		return value
	}
	if int32(rounded.BitLen())-1+quantum > maximumNormalExponent {
		if nearest {
			value.kind = rawDefaultDamageFloatInfinity4E0B30
			return value
		}
		value.significand = uint64(1)<<precision - 1
		value.exponent = maximumNormalExponent - int32(precision) + 1
		return value
	}
	if rounded.BitLen() > int(precision) {
		// A normal rounding carry is exactly 2^precision.
		rounded.Rsh(rounded, 1)
		quantum++
	}
	value.significand = rounded.Uint64()
	value.exponent = quantum
	return value
}

func rawDefaultDamageFloatStore32Bits4E0B30(value rawDefaultDamageFloat4E0B30, cw uint16) uint32 {
	_ = rawDefaultDamageFloatNearest4E0B30(cw)
	value = rawDefaultDamageFloatQuiet4E0B30(value)
	sign := uint32(0)
	if value.negative {
		sign = uint32(1) << 31
	}
	switch value.kind {
	case rawDefaultDamageFloatNaN4E0B30:
		// Drop low extended payload bits and quiet the memory NaN. The
		// sign and the payload bits which fit binary32 stay independent.
		return sign | 0x7f800000 | uint32(value.significand>>40)&0x7fffff | 0x400000
	case rawDefaultDamageFloatInfinity4E0B30:
		return sign | 0x7f800000
	}
	value = rawDefaultDamageFloatRound4E0B30(new(big.Int).SetUint64(value.significand), value.exponent, value.negative, 24, -126, 127, cw)
	if value.kind == rawDefaultDamageFloatInfinity4E0B30 {
		return sign | 0x7f800000
	}
	if value.significand == 0 {
		return sign
	}
	width := bits.Len64(value.significand)
	exponent := int32(width) - 1 + value.exponent
	if exponent < -126 {
		// A subnormal result has already been rounded on 2^-149.
		return sign | uint32(value.significand)
	}
	fraction := uint32(value.significand<<uint(24-width)) & 0x7fffff
	return sign | uint32(exponent+127)<<23 | fraction
}

// FCOMP's C3 bit is set for zero equality AND unordered NaNs/unsupported
// encodings. This is intentionally different from the Go expression x != 0.
func rawDefaultDamageFloatOrderedNonzero4E0B30(value rawDefaultDamageFloat4E0B30) bool {
	return value.kind == rawDefaultDamageFloatInfinity4E0B30 || value.kind == rawDefaultDamageFloatFinite4E0B30 && value.significand != 0
}

func rawDefaultDamageFloatMultiplyDamage4E0B30(value rawDefaultDamageFloat4E0B30, damage int32, cw uint16) rawDefaultDamageFloat4E0B30 {
	_ = rawDefaultDamageFloatNearest4E0B30(cw)
	value = rawDefaultDamageFloatQuiet4E0B30(value)
	if value.kind == rawDefaultDamageFloatNaN4E0B30 {
		return value
	}
	if value.kind == rawDefaultDamageFloatInfinity4E0B30 && damage == 0 {
		return rawDefaultDamageFloatIndefinite4E0B30()
	}
	negative := value.negative != (damage < 0)
	if value.kind == rawDefaultDamageFloatInfinity4E0B30 {
		value.negative = negative
		return value
	}
	integer := int64(damage)
	if integer < 0 {
		integer = -integer
	}
	magnitude := new(big.Int).Mul(new(big.Int).SetUint64(value.significand), big.NewInt(integer))
	return rawDefaultDamageFloatRound4E0B30(magnitude, value.exponent, negative, 53, -16382, 16383, cw)
}

// Execute only the original numerical boundaries: FLD1; FSUB stored p32;
// FIMUL signed damage; FSTP32. The caller owns the unspilled audio gate and
// the later converter/zero-only minimum assignment.
func rawDefaultDamageFloatProtectionProduct4E0B30(p32bits, damage uint32, cw uint16) uint32 {
	_ = rawDefaultDamageFloatNearest4E0B30(cw)
	protection := rawDefaultDamageFloatQuiet4E0B30(rawDefaultDamageFloatFrom32Bits4E0B30(p32bits))
	var difference rawDefaultDamageFloat4E0B30
	switch protection.kind {
	case rawDefaultDamageFloatNaN4E0B30:
		difference = protection
	case rawDefaultDamageFloatInfinity4E0B30:
		difference = protection
		difference.negative = !protection.negative
	default:
		exponent := protection.exponent
		if exponent > 0 {
			exponent = 0
		}
		left := new(big.Int).Lsh(big.NewInt(1), uint(-exponent))
		right := new(big.Int).Lsh(new(big.Int).SetUint64(protection.significand), uint(protection.exponent-exponent))
		if protection.negative {
			left.Add(left, right)
		} else {
			left.Sub(left, right)
		}
		negative := left.Sign() < 0
		left.Abs(left)
		difference = rawDefaultDamageFloatRound4E0B30(left, exponent, negative, 53, -16382, 16383, cw)
	}
	product := rawDefaultDamageFloatMultiplyDamage4E0B30(difference, int32(damage), cw)
	return rawDefaultDamageFloatStore32Bits4E0B30(product, cw)
}

// The converter first loads binary32, then FISTP uses inherited RC. PC53
// does not introduce a new floating-point store here. The masked invalid
// integer result and the valid integer -2^31 have the same return bits.
func rawDefaultDamageFloatConvert32Bits4E0B30(word uint32, cw uint16) int32 {
	nearest := rawDefaultDamageFloatNearest4E0B30(cw)
	value := rawDefaultDamageFloatFrom32Bits4E0B30(word)
	if value.kind != rawDefaultDamageFloatFinite4E0B30 {
		return -1 << 31
	}
	magnitude := rawDefaultDamageFloatRoundGrid4E0B30(new(big.Int).SetUint64(value.significand), value.exponent, 0, nearest)
	limit := new(big.Int).Lsh(big.NewInt(1), 31)
	if value.negative {
		if magnitude.Cmp(limit) > 0 {
			return -1 << 31
		}
		return int32(-magnitude.Int64())
	}
	if magnitude.Cmp(limit) >= 0 {
		return -1 << 31
	}
	return int32(magnitude.Int64())
}
