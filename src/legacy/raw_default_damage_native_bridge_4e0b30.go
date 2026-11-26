package legacy

/*
#include "raw_default_damage_native_bridge_4e0b30.h"
*/
import "C"

import (
	"runtime"
	"unsafe"
)

// These synchronous bridges preserve actual pointer arguments. KeepAlive is
// a lifetime barrier, not a pinning/transitive-Go-graph guarantee. Go-owned
// graphs still need the caller's complete cgo/pinning contract; the current
// raw consumer graph is foreign C-owned storage.
func rawDefaultDamageCallDamageNativeCallback4E0B30(cb, target, source, weapon unsafe.Pointer, damage, typ int32) {
	C.rdd_native_damage_callback_4e0b30((*[0]byte)(cb), target, source, weapon, C.int32_t(damage), C.int32_t(typ))
	runtime.KeepAlive(cb)
	runtime.KeepAlive(target)
	runtime.KeepAlive(source)
	runtime.KeepAlive(weapon)
}

func rawDefaultDamageCallSoundNativeCallback4E0B30(cb, target, source unsafe.Pointer) {
	C.rdd_native_sound_callback_4e0b30((*[0]byte)(cb), target, source)
	runtime.KeepAlive(cb)
	runtime.KeepAlive(target)
	runtime.KeepAlive(source)
}

// The product flag checks return a canonical word. An observer's reply is
// handled by the caller and can contain a full-width value.
func rawDefaultDamageGameFlagsNative4E0B30(flags uint32) uint32 {
	return uint32(bool2int(bool(nox_common_gameFlags_check_40A5C0(C.uint(flags)))))
}

func rawDefaultDamageGameplayFlagsNative4E0B30(flags int32) uint32 {
	return uint32(bool2int(bool(nox_xxx_CheckGameplayFlags_417DA0(flags))))
}

//export rdd_native_game_flags_word_4e0b30
func rdd_native_game_flags_word_4e0b30(flags C.uint32_t) C.uint32_t {
	return C.uint32_t(rawDefaultDamageGameFlagsNative4E0B30(uint32(flags)))
}

//export rdd_native_gameplay_flags_word_4e0b30
func rdd_native_gameplay_flags_word_4e0b30(flags C.int32_t) C.uint32_t {
	return C.uint32_t(rawDefaultDamageGameplayFlagsNative4E0B30(int32(flags)))
}
