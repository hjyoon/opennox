package legacy

/*
#include <stdint.h>

*/
import "C"

const minimumPlayerProtectionToken = 0x2734945f

func updatePlayerProtectedStat(token uint32, update func(uint32) uint32) int32 {
	if int32(token) < minimumPlayerProtectionToken {
		return int32(token)
	}
	record := findPlayerProtectionRecord(token)
	if record == nil {
		return 0
	}
	oldValue := C.uint32_t(record.value)
	legacyWords.dword_5d4594_2516328 ^= oldValue
	key := legacyWords.dword_5d4594_2516348
	newValue := C.uint32_t(update(uint32(oldValue^key))) ^ key
	record.value = uint32(newValue)
	legacyWords.dword_5d4594_2516328 ^= newValue
	return int32(nox_xxx_protectData_56F5C0())
}

func setPlayerProtectedStat(token, value uint32) int32 {
	return updatePlayerProtectedStat(token, func(uint32) uint32 { return value })
}

func addPlayerProtectedStat(token, delta uint32) int32 {
	return updatePlayerProtectedStat(token, func(value uint32) uint32 { return value + delta })
}

func addPlayerProtectedExperience(token uint32, delta float32) int32 {
	return updatePlayerProtectedStat(token, func(value uint32) uint32 {
		return uint32(x87TruncQwordLow(addFloat64TowardZero(float64(value), float64(delta))))
	})
}

//export sub_56F780
func sub_56F780(token C.int, value C.int) C.int32_t {
	return C.int32_t(setPlayerProtectedStat(uint32(token), uint32(value)))
}

//export nox_xxx_playerResetProtectionCRC_56F7D0
func nox_xxx_playerResetProtectionCRC_56F7D0(token C.int, value C.int) C.int32_t {
	return C.int32_t(setPlayerProtectedStat(uint32(token), uint32(value)))
}

//export sub_56F820
func sub_56F820(token C.int, value C.uchar) C.int32_t {
	return C.int32_t(setPlayerProtectedStat(uint32(token), uint32(value)))
}

//export nox_xxx_protectPlayerHPMana_56F870
func nox_xxx_protectPlayerHPMana_56F870(token C.int, value C.ushort) C.int32_t {
	return C.int32_t(setPlayerProtectedStat(uint32(token), uint32(value)))
}

//export sub_56F8C0
func sub_56F8C0(token C.int, value C.float) C.int32_t {
	return C.int32_t(setPlayerProtectedStat(uint32(token), uint32(x87TruncQwordLow(float64(value)))))
}

//export sub_56F920
func sub_56F920(token C.int, delta C.int) C.int32_t {
	return C.int32_t(addPlayerProtectedStat(uint32(token), uint32(delta)))
}

//export sub_56F980
func sub_56F980(token C.int, delta C.uchar) C.int32_t {
	return C.int32_t(addPlayerProtectedStat(uint32(token), uint32(delta)))
}

//export nox_xxx_protectMana_56F9E0
func nox_xxx_protectMana_56F9E0(token C.int, delta C.short) C.int32_t {
	return C.int32_t(addPlayerProtectedStat(uint32(token), uint32(delta)))
}

//export sub_56FA40
func sub_56FA40(token C.int, delta C.float) C.int32_t {
	return C.int32_t(addPlayerProtectedExperience(uint32(token), float32(delta)))
}
