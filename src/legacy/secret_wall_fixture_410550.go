package legacy

/*
#include "GAME1.h"
#include "secret_wall.h"
#include "map_temp_lists_503f40.h"
#include <stddef.h>
#include <stdint.h>
#include <stdlib.h>
#include <string.h>
extern void* dword_5d4594_251560;
typedef void* (*secret_wall_lookup_fn_410550)(short);
typedef nox_secret_wall_t* (*secret_wall_clear_fn_410550)(void);
typedef nox_secret_wall_t* (*secret_wall_remove_fn_410550)(void*);
_Static_assert(_Generic(&sub_410550,secret_wall_lookup_fn_410550:1,default:0),"410550 public short to pointer ABI");
_Static_assert(_Generic(&sub_410730,secret_wall_clear_fn_410550:1,default:0),"410730 public void to typed pointer ABI");
_Static_assert(_Generic(&sub_4107A0,secret_wall_remove_fn_410550:1,default:0),"4107A0 public void pointer to typed pointer ABI");
_Static_assert(sizeof(short)==2 && sizeof(uint32_t)==4,"SecretWall original WORD and DWORD widths");
_Static_assert(sizeof(((nox_map_wall_504290*)0)->field_10)==2 && sizeof(((nox_secret_wall_t*)0)->next)==sizeof(void*) && sizeof(((nox_secret_wall_t*)0)->wall)==sizeof(void*),"SecretWall actual field widths");
static void* secret_wall_fixture_call_410550(uint32_t entry,uint32_t word,void* pointer) {
    switch(entry) {
    case 0x410550: { secret_wall_lookup_fn_410550 fn=&sub_410550; return fn((short)(uint16_t)word); }
    case 0x410730: { secret_wall_clear_fn_410550 fn=&sub_410730; return fn(); }
    case 0x4107a0: { secret_wall_remove_fn_410550 fn=&sub_4107A0; return fn(pointer); }
    default:return NULL;
    }
}
static size_t secret_wall_fixture_layout_410550(int index) {
    switch(index) {
    case 0:return offsetof(nox_secret_wall_t,next);
    case 1:return offsetof(nox_secret_wall_t,wall);
    case 2:return offsetof(nox_map_wall_504290,field_10);
    case 3:return sizeof(nox_secret_wall_t);
    default:return (size_t)-1;
    }
}
static size_t secret_wall_fixture_scalar_offset_410550(int index) {
    switch(index) {
    case 0:return offsetof(nox_secret_wall_t,x);
    case 1:return offsetof(nox_secret_wall_t,y);
    case 2:return offsetof(nox_secret_wall_t,open_wait);
    case 3:return offsetof(nox_secret_wall_t,flags);
    case 4:return offsetof(nox_secret_wall_t,last_open);
    case 5:return offsetof(nox_secret_wall_t,player_bits);
    default:return (size_t)-1;
    }
}
_Static_assert(offsetof(nox_secret_wall_t,last_open)-offsetof(nox_secret_wall_t,flags)==4,"four adjacent SecretWall BYTEs");
static size_t secret_wall_fixture_pointer_width_410550(void) { return sizeof(void*); }
static size_t secret_wall_fixture_wall_size_410550(void) { return sizeof(nox_map_wall_504290); }
static void* secret_wall_fixture_head_cell_410550(void) { return &dword_5d4594_251560; }
static void secret_wall_fixture_store_pointer_410550(void* slot,void* pointer) { memcpy(slot,&pointer,sizeof(pointer)); }
static void* secret_wall_fixture_load_pointer_410550(const void* slot) { void* pointer;memcpy(&pointer,slot,sizeof(pointer));return pointer; }
static void* secret_wall_fixture_alloc_410550(size_t n) { return calloc(1,n); }
static void secret_wall_fixture_free_410550(void* p) { free(p); }
*/
import "C"
import "unsafe"

// These thunks call the actual public C entries. They never restore PE DWORDs
// as pointers and never dereference an address obtained from a low32 witness.
func secretWallFixtureCall410550(entry uint32, argWord uint32, argPointer unsafe.Pointer) unsafe.Pointer {
	switch entry {
	case 0x410550, 0x410730, 0x4107a0:
	default:
		panic("SecretWall fixture entry")
	}
	return C.secret_wall_fixture_call_410550(C.uint32_t(entry), C.uint32_t(argWord), argPointer)
}
func secretWallFixtureLayout410550() [4]uintptr {
	var out [4]uintptr
	for i := range out {
		out[i] = uintptr(C.secret_wall_fixture_layout_410550(C.int(i)))
	}
	return out
}
func secretWallFixturePointerWidth410550() uintptr {
	return uintptr(C.secret_wall_fixture_pointer_width_410550())
}
func secretWallFixtureWallSize410550() uintptr {
	return uintptr(C.secret_wall_fixture_wall_size_410550())
}
func secretWallFixtureHeadCell410550() unsafe.Pointer {
	return C.secret_wall_fixture_head_cell_410550()
}
func secretWallFixtureStorePointer410550(slot, p unsafe.Pointer) {
	C.secret_wall_fixture_store_pointer_410550(slot, p)
}
func secretWallFixtureLoadPointer410550(slot unsafe.Pointer) unsafe.Pointer {
	return C.secret_wall_fixture_load_pointer_410550(slot)
}

// Finite default-linked controls own these C allocations. Goheap controls pin
// every transitive owner explicitly and clear C-retained pointers before Unpin.
func secretWallFixtureAlloc410550(n uintptr) unsafe.Pointer {
	return C.secret_wall_fixture_alloc_410550(C.size_t(n))
}
func secretWallFixtureFree410550(p unsafe.Pointer) { C.secret_wall_fixture_free_410550(p) }

func secretWallFixtureScalarOffsets410550() [6]uintptr {
	var out [6]uintptr
	for i := range out {
		out[i] = uintptr(C.secret_wall_fixture_scalar_offset_410550(C.int(i)))
	}
	return out
}
