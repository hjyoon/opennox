package legacy

/*
#include <stdint.h>

*/
import "C"

const minimumPlayerProtectionToken = 0x2734945f

func updatePlayerProtectedStat(token C.int, update func(uint32) uint32) C.int32_t {
	if token < minimumPlayerProtectionToken {
		return C.int32_t(token)
	}
	record := findPlayerProtectionRecord(uint32(token))
	if record == nil {
		return 0
	}
	oldValue := C.uint32_t(record.value)
	legacyWords.dword_5d4594_2516328 ^= oldValue
	key := legacyWords.dword_5d4594_2516348
	newValue := C.uint32_t(update(uint32(oldValue^key))) ^ key
	record.value = uint32(newValue)
	legacyWords.dword_5d4594_2516328 ^= newValue
	return C.int32_t(nox_xxx_protectData_56F5C0())
}

func setPlayerProtectedStat(token C.int, value uint32) C.int32_t {
	return updatePlayerProtectedStat(token, func(uint32) uint32 { return value })
}

func addPlayerProtectedStat(token C.int, delta uint32) C.int32_t {
	return updatePlayerProtectedStat(token, func(value uint32) uint32 { return value + delta })
}

//export sub_56F780
func sub_56F780(token C.int, value C.int) C.int32_t {
	return setPlayerProtectedStat(token, uint32(value))
}

//export nox_xxx_playerResetProtectionCRC_56F7D0
func nox_xxx_playerResetProtectionCRC_56F7D0(token C.int, value C.int) C.int32_t {
	return setPlayerProtectedStat(token, uint32(value))
}

//export sub_56F820
func sub_56F820(token C.int, value C.uchar) C.int32_t {
	return setPlayerProtectedStat(token, uint32(value))
}

//export nox_xxx_protectPlayerHPMana_56F870
func nox_xxx_protectPlayerHPMana_56F870(token C.int, value C.ushort) C.int32_t {
	return setPlayerProtectedStat(token, uint32(value))
}

//export sub_56F8C0
func sub_56F8C0(token C.int, value C.float) C.int32_t {
	return setPlayerProtectedStat(token, uint32(x87TruncQwordLow(float64(value))))
}

//export sub_56F920
func sub_56F920(token C.int, delta C.int) C.int32_t {
	return addPlayerProtectedStat(token, uint32(delta))
}

//export sub_56F980
func sub_56F980(token C.int, delta C.uchar) C.int32_t {
	return addPlayerProtectedStat(token, uint32(delta))
}

//export nox_xxx_protectMana_56F9E0
func nox_xxx_protectMana_56F9E0(token C.int, delta C.short) C.int32_t {
	return addPlayerProtectedStat(token, uint32(delta))
}

//export sub_56FA40
func sub_56FA40(token C.int, delta C.float) C.int32_t {
	return updatePlayerProtectedStat(token, func(value uint32) uint32 {
		return uint32(x87TruncQwordLow(addFloat64TowardZero(float64(value), float64(delta))))
	})
}
