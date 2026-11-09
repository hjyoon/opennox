package legacy

/*
#include <stddef.h>
#include <stdlib.h>
#include "GAME1.h"

extern void* nox_access_native_head_416720(int banned);
extern nox_list_item_t nox_access_allowed_list_4168E0;
extern nox_list_item_t nox_access_banned_list_416900;

static void* access_support_head_416720(int banned) {
    return nox_access_native_head_416720(banned);
}

static void* access_support_owner_416720(int banned) {
    return banned ? (void*)&nox_access_banned_list_416900 : (void*)&nox_access_allowed_list_4168E0;
}

static void* access_support_field_416720(int field, void* entry) {
    switch (field) {
    case 0: return nox_access_allowed_name_4168E0((nox_access_allowed_entry_t*)entry);
    case 1: return nox_access_banned_name_416900((nox_access_banned_entry_t*)entry);
    case 2: return nox_access_banned_serial_416900((nox_access_banned_entry_t*)entry);
    default: abort();
    }
}

static size_t access_support_offset_416720(int field) {
    switch (field) {
    case 0: return offsetof(nox_access_allowed_entry_t, name);
    case 1: return offsetof(nox_access_banned_entry_t, name);
    case 2: return offsetof(nox_access_banned_entry_t, serial);
    default: abort();
    }
}

static size_t access_support_size_416720(int banned) {
    return banned ? sizeof(nox_access_banned_entry_t) : sizeof(nox_access_allowed_entry_t);
}

// Read the actual C layout independently of the Go list helper.
static void* access_support_link_416720(void* entry, int field) {
    nox_list_item_t* item = (nox_list_item_t*)entry;
    switch (field) {
    case 0: return item->field_0;
    case 1: return item->field_1;
    case 2: return item->field_2;
    default: abort();
    }
}

// Named-field writes witness both directions of the returned field alias.
static void access_support_store_416720(int field, void* entry, unsigned index, unsigned value) {
    switch (field) {
    case 0:
        if (index >= 26) abort();
        ((nox_access_allowed_entry_t*)entry)->name[index] = (wchar2_t)value;
        return;
    case 1:
        if (index >= 26) abort();
        ((nox_access_banned_entry_t*)entry)->name[index] = (wchar2_t)value;
        return;
    case 2:
        if (index >= 24) abort();
        ((nox_access_banned_entry_t*)entry)->serial[index] = (char)value;
        return;
    default: abort();
    }
}

static unsigned access_support_read_416720(int field, void* entry, unsigned index) {
    switch (field) {
    case 0:
        if (index >= 26) abort();
        return ((nox_access_allowed_entry_t*)entry)->name[index];
    case 1:
        if (index >= 26) abort();
        return ((nox_access_banned_entry_t*)entry)->name[index];
    case 2:
        if (index >= 24) abort();
        return (unsigned char)((nox_access_banned_entry_t*)entry)->serial[index];
    default: abort();
    }
}
*/
import "C"

import (
	"runtime"
	"unsafe"
)

func accessSupportFixtureHead416720(banned int32) unsafe.Pointer {
	return C.access_support_head_416720(C.int(banned))
}

func accessSupportFixtureOwner416720(banned bool) unsafe.Pointer {
	v := C.int(0)
	if banned {
		v = 1
	}
	return C.access_support_owner_416720(v)
}

func accessSupportFixtureField416720(field int, entry unsafe.Pointer) unsafe.Pointer {
	p := C.access_support_field_416720(C.int(field), entry)
	runtime.KeepAlive(entry)
	return p
}

func accessSupportFixtureOffset416720(field int) uintptr {
	return uintptr(C.access_support_offset_416720(C.int(field)))
}

func accessSupportFixtureSize416720(banned bool) uintptr {
	v := C.int(0)
	if banned {
		v = 1
	}
	return uintptr(C.access_support_size_416720(v))
}

func accessSupportFixtureLinks416720(entry unsafe.Pointer) [3]unsafe.Pointer {
	var links [3]unsafe.Pointer
	for i := range links {
		links[i] = C.access_support_link_416720(entry, C.int(i))
	}
	runtime.KeepAlive(entry)
	return links
}

func accessSupportFixtureStore416720(field int, entry unsafe.Pointer, index int, value uint32) {
	C.access_support_store_416720(C.int(field), entry, C.uint(index), C.uint(value))
	runtime.KeepAlive(entry)
}

func accessSupportFixtureRead416720(field int, entry unsafe.Pointer, index int) uint32 {
	v := uint32(C.access_support_read_416720(C.int(field), entry, C.uint(index)))
	runtime.KeepAlive(entry)
	return v
}
