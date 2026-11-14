package legacy

/*
#include <stddef.h>
#include "GAME1_1.h"
#include "GAME4_2.h"

_Static_assert(_Generic(&nox_xxx_setNPCVoiceSet_424320, int (*)(nox_object_t*, void*): 1, default: 0), "native NPC voice setter ABI");
_Static_assert(_Generic(&nox_xxx_getDefaultSoundSet_424350, void* (*)(const char*): 1, default: 0), "opaque sound data and const name ABI");
_Static_assert(_Generic(&nox_xxx_readNPCVoiceSet_52AD10, size_t (*)(nox_object_t*): 1, default: 0), "native NPC voice transfer ABI");
_Static_assert(_Generic(&nox_xxx_monsterGetSoundSet_424300, void* (*)(nox_object_t*): 1, default: 0), "existing native voice getter ABI");

static size_t npc_voice_object_class(void) { return offsetof(nox_object_t, obj_class); }
static size_t npc_voice_object_update(void) { return offsetof(nox_object_t, data_update); }
static int npc_voice_set(nox_object_t* obj, void* data) { return nox_xxx_setNPCVoiceSet_424320(obj, data); }
static void* npc_voice_lookup(const char* name) { return nox_xxx_getDefaultSoundSet_424350(name); }
static void* npc_voice_get(nox_object_t* obj) { return nox_xxx_monsterGetSoundSet_424300(obj); }
static size_t npc_voice_transfer(nox_object_t* obj) { return nox_xxx_readNPCVoiceSet_52AD10(obj); }
*/
import "C"

import "unsafe"

func npcVoiceLayoutC() (uintptr, uintptr, uintptr) {
	return C.sizeof_nox_object_t, uintptr(C.npc_voice_object_class()), uintptr(C.npc_voice_object_update())
}

func npcVoiceSetC(obj, data unsafe.Pointer) int32 {
	return int32(C.npc_voice_set((*C.nox_object_t)(obj), data))
}

func npcVoiceLookupC(name unsafe.Pointer) unsafe.Pointer {
	return C.npc_voice_lookup((*C.char)(name))
}

func npcVoiceGetC(obj unsafe.Pointer) unsafe.Pointer {
	return C.npc_voice_get((*C.nox_object_t)(obj))
}

func npcVoiceTransferC(obj unsafe.Pointer) uintptr {
	return uintptr(C.npc_voice_transfer((*C.nox_object_t)(obj)))
}
