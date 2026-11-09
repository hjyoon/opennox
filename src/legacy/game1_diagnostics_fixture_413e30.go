package legacy

/*
#include "GAME1.h"
#include "game1_mem_dump_native_413e30.h"
#include <stddef.h>
#include <stdlib.h>
static void nox_mem_dump_fixture_entry_413e30(void) { nox_xxx_gameLoopMemDump_413E30(); }
static unsigned char nox_trig_fixture_entry_414c90(void) { return (unsigned char)nox_xxx_initSinCosTables_414C90(); }
static void* nox_mem_dump_fixture_new_413e30(void) { return calloc(1, sizeof(nox_mem_dump_node_413e30)); }
static size_t nox_mem_dump_fixture_layout_413e30(int field) {
    switch(field) {
    case 0: return sizeof(nox_mem_dump_node_413e30);
    case 1: return offsetof(nox_mem_dump_node_413e30, next);
    case 2: return offsetof(nox_mem_dump_node_413e30, count);
    default: return offsetof(nox_mem_dump_node_413e30, name);
    }
}
*/
import "C"
import "unsafe"

func memDumpFixturePublic413E30()               { C.nox_mem_dump_fixture_entry_413e30() }
func trigFixturePublic414C90() byte             { return byte(C.nox_trig_fixture_entry_414c90()) }
func memDumpFixtureNew413E30() unsafe.Pointer   { return C.nox_mem_dump_fixture_new_413e30() }
func memDumpFixtureFree413E30(p unsafe.Pointer) { C.free(p) }
func memDumpFixtureLayout413E30() (v [4]uintptr) {
	for i := range v {
		v[i] = uintptr(C.nox_mem_dump_fixture_layout_413e30(C.int(i)))
	}
	return
}
