package legacy

/*
#include "GAME1.h"
#include <stddef.h>
#include <stdint.h>
#include <stdlib.h>
#include <string.h>

_Static_assert(sizeof(int) == 4, "two settings pointer C return width");
_Static_assert(sizeof(uintptr_t) == sizeof(void*), "native pointer witness width");
typedef int (*two_settings_pointer_fixture_fn_40a740)(void);
_Static_assert(_Generic(&sub_40A740, two_settings_pointer_fixture_fn_40a740: 1, default: 0), "40A740 int(void) ABI");
_Static_assert(_Generic(&sub_40A770, two_settings_pointer_fixture_fn_40a740: 1, default: 0), "40A770 int(void) ABI");

static int two_settings_pointer_fixture_call_40a740(void) {
    two_settings_pointer_fixture_fn_40a740 fn = &sub_40A740;
    return fn();
}

static int two_settings_pointer_fixture_call_40a770(void) {
    two_settings_pointer_fixture_fn_40a740 fn = &sub_40A770;
    return fn();
}

static size_t two_settings_pointer_fixture_offset_40a740(uint32_t field) {
    switch (field) {
    case 0: return offsetof(nox_object_t, data_update);
    case 1: return offsetof(nox_player_update_data_t, player);
    case 2: return offsetof(nox_playerInfo, field_3680);
    case 3: return offsetof(nox_playerInfo, playerInd);
    default: abort();
    }
}

static size_t two_settings_pointer_fixture_width_40a740(uint32_t field) {
    switch (field) {
    case 0: return sizeof(((nox_object_t*)0)->data_update);
    case 1: return sizeof(((nox_player_update_data_t*)0)->player);
    case 2: return sizeof(((nox_playerInfo*)0)->field_3680);
    case 3: return sizeof(((nox_playerInfo*)0)->playerInd);
    default: abort();
    }
}

static size_t two_settings_pointer_fixture_pointer_width_40a740(void) {
    return sizeof(void*);
}

// These memcpy operations do not require an aligned pointer slot or a complete
// object/update/player allocation. The caller must provide sizeof(void*) bytes.
static void two_settings_pointer_fixture_store_pointer_40a740(void* dst, void* value) {
    memcpy(dst, &value, sizeof(value));
}

static void* two_settings_pointer_fixture_load_pointer_40a740(const void* src) {
    void* value;
    memcpy(&value, src, sizeof(value));
    return value;
}

typedef struct two_settings_pointer_fixture_witness_40a740 {
    uintptr_t native;
    uint32_t low32;
    int above_4GiB;
    int low32_loses_bits;
} two_settings_pointer_fixture_witness_40a740;

static two_settings_pointer_fixture_witness_40a740 two_settings_pointer_fixture_pointer_witness_40a740(void* ptr) {
    uintptr_t native = (uintptr_t)ptr;
    uint32_t low32 = (uint32_t)native;
    two_settings_pointer_fixture_witness_40a740 out = {
        native, low32, native > UINT64_C(0x100000000), (uintptr_t)low32 != native
    };
    return out;
}
*/
import "C"

import (
	"runtime"
	"unsafe"
)

const (
	twoSettingsPointerFixtureObjectUpdate40A740 = iota
	twoSettingsPointerFixtureUpdatePlayer40A740
	twoSettingsPointerFixturePlayerStatus40A740
	twoSettingsPointerFixturePlayerIndex40A740
)

type twoSettingsPointerFixtureField40A740 struct {
	Offset uintptr
	Width  uintptr
}

// The values come from the C declarations compiled with the consuming test
// main. This helper does not access an instance or admit a complete C struct.
func twoSettingsPointerFixtureCLayout40A740() (out [4]twoSettingsPointerFixtureField40A740) {
	for i := range out {
		out[i] = twoSettingsPointerFixtureField40A740{
			Offset: uintptr(C.two_settings_pointer_fixture_offset_40a740(C.uint32_t(i))),
			Width:  uintptr(C.two_settings_pointer_fixture_width_40a740(C.uint32_t(i))),
		}
	}
	return out
}

func twoSettingsPointerFixtureCPointerWidth40A740() uintptr {
	return uintptr(C.two_settings_pointer_fixture_pointer_width_40a740())
}

func twoSettingsPointerFixtureCall40A740(entry uint32) int32 {
	switch entry {
	case 0x40a740:
		return int32(C.two_settings_pointer_fixture_call_40a740())
	case 0x40a770:
		return int32(C.two_settings_pointer_fixture_call_40a770())
	default:
		panic("unsupported two settings pointer C entry")
	}
}

// The consumer owns destination extent, pointer validity and lifetime. For a
// Go allocation it must retain a runtime.Pinner owner and pin every reachable
// Go pointee for the entire interval in which C memory retains that pointer.
// KeepAlive alone does not pin memory. No pointer is hidden in an integer.
func twoSettingsPointerFixtureStorePointer40A740(dst, value unsafe.Pointer) {
	C.two_settings_pointer_fixture_store_pointer_40a740(dst, value)
	runtime.KeepAlive(dst)
	runtime.KeepAlive(value)
}

// src must name a full native pointer slot. Partial PE pointer bytes remain
// opaque input bytes; this helper does not turn them into a native identity.
func twoSettingsPointerFixtureLoadPointer40A740(src unsafe.Pointer) unsafe.Pointer {
	value := C.two_settings_pointer_fixture_load_pointer_40a740(src)
	runtime.KeepAlive(src)
	return value
}

type twoSettingsPointerFixtureWitness40A740 struct {
	Native         uintptr
	Low32          uint32
	Above4GiB      bool
	Low32LosesBits bool
}

// This negative witness is numeric only. The low word is never converted back
// to a pointer or dereferenced. The current old C callers already use native
// pointer types; this helper does not claim they truncate those pointers.
func twoSettingsPointerFixturePointerWitness40A740(ptr unsafe.Pointer) twoSettingsPointerFixtureWitness40A740 {
	w := C.two_settings_pointer_fixture_pointer_witness_40a740(ptr)
	runtime.KeepAlive(ptr)
	return twoSettingsPointerFixtureWitness40A740{
		Native:         uintptr(w.native),
		Low32:          uint32(w.low32),
		Above4GiB:      w.above_4GiB != 0,
		Low32LosesBits: w.low32_loses_bits != 0,
	}
}
