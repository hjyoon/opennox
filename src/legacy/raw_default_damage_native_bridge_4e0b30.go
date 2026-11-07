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

// C canonicalizes its own bool result. Handled observer reply.Word values do
// not pass through these product fallbacks and must remain full-width words.
func rawDefaultDamageGameFlagsNative4E0B30(flags uint32) uint32 {
	return uint32(C.rdd_native_game_flags_word_4e0b30(C.uint32_t(flags)))
}

func rawDefaultDamageGameplayFlagsNative4E0B30(flags int32) uint32 {
	return uint32(C.rdd_native_gameplay_flags_word_4e0b30(C.int32_t(flags)))
}
