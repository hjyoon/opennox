package legacy

/*
#include <stddef.h>
#include <stdint.h>
#include <string.h>
typedef struct { uintptr_t object, older, newer; } spawn_owner_fixture_link;
static uintptr_t spawn_owner_fixture_read(void* owner, size_t offset) {
	uintptr_t value;
	memcpy(&value, (unsigned char*)owner + offset, sizeof(value));
	return value;
}
static size_t spawn_owner_fixture_size(void) { return sizeof(spawn_owner_fixture_link); }
static size_t spawn_owner_fixture_align(void) { return _Alignof(spawn_owner_fixture_link); }
static size_t spawn_owner_fixture_older(void) { return offsetof(spawn_owner_fixture_link, older); }
static size_t spawn_owner_fixture_newer(void) { return offsetof(spawn_owner_fixture_link, newer); }
*/
import "C"

import "unsafe"

func monsterSpawnOwnerNativeRead(pointer unsafe.Pointer, offset uintptr) uintptr {
	return uintptr(C.spawn_owner_fixture_read(pointer, C.size_t(offset)))
}

func monsterSpawnOwnerNativeCLayout() [4]uintptr {
	return [4]uintptr{uintptr(C.spawn_owner_fixture_size()), uintptr(C.spawn_owner_fixture_align()), uintptr(C.spawn_owner_fixture_older()), uintptr(C.spawn_owner_fixture_newer())}
}
