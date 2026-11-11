package legacy

/*
#include <stdint.h>
#include <stdlib.h>
#include "native_fixture_alloc.h"
#include <string.h>
#if defined(__clang__)
#pragma clang diagnostic push
#pragma clang diagnostic ignored "-Wint-to-pointer-cast"
#endif
#include "game4_map_archive_reference.h"
#if defined(__clang__)
#pragma clang diagnostic pop
#endif

typedef struct game4_map_archive_test_bank {
	void* old_first;
	void* old_second;
	char* old_records;
	uint32_t old_count;
	FILE* old_file;
	uint64_t before;
	char first[2048];
	char second[2048];
	char records[0x26000];
	uint64_t after;
} game4_map_archive_test_bank;

static game4_map_archive_test_bank* game4_map_archive_test_begin(void) {
	game4_map_archive_test_bank* b = nox_native_fixture_calloc_high(1, sizeof(*b));
	if (!b) return NULL;
	b->old_first = dword_5d4594_1599588;
	b->old_second = dword_5d4594_1599592;
	b->old_records = dword_5d4594_1599576;
	b->old_count = dword_5d4594_1599596;
	b->old_file = nox_file_8;
	dword_5d4594_1599588 = b->first;
	dword_5d4594_1599592 = b->second;
	dword_5d4594_1599576 = b->records;
	nox_file_8 = NULL;
	return b;
}

static void game4_map_archive_test_seed(game4_map_archive_test_bank* b) {
	b->before = UINT64_C(0x7060504030201000);
	b->after = UINT64_C(0xffeeddccbbaa9988);
	memset(b->first, 0xa5, sizeof(b->first));
	memset(b->second, 0xa5, sizeof(b->second));
	memset(b->records, 0xa5, sizeof(b->records));
	b->first[0] = b->second[0] = 0;
	b->first[2047] = b->second[2047] = 0;
	strcpy(b->records, "Alpha");
	strcpy(b->records + 76, "Beta");
	strcpy(b->records + 152, "Gamma");
	uint32_t x = 0x3fc00000, y = 0xc0200000;
	memcpy(b->records + 64, &x, 4);
	memcpy(b->records + 76 + 68, &y, 4);
	dword_5d4594_1599596 = 3;
}

static void game4_map_archive_test_end(game4_map_archive_test_bank* b) {
	if (nox_file_8 && nox_file_8 != b->old_file) sub_502DF0();
	dword_5d4594_1599588 = b->old_first;
	dword_5d4594_1599592 = b->old_second;
	dword_5d4594_1599576 = b->old_records;
	dword_5d4594_1599596 = b->old_count;
	nox_file_8 = b->old_file;
	nox_native_fixture_free_high(b);
}

typedef struct game4_map_archive_test_leaf_result {
	int index[5];
	uintptr_t pointers[4];
	uint64_t coordinate_bits[2];
	int setters[4];
} game4_map_archive_test_leaf_result;

static game4_map_archive_test_leaf_result game4_map_archive_test_leaf(game4_map_archive_test_bank* b, int reference) {
	game4_map_archive_test_leaf_result r = {0};
	if (reference) {
		r.index[0] = game4_ref_sub_5029A0("alpha");
		r.index[1] = game4_ref_sub_5029A0("BETA");
		r.index[2] = game4_ref_sub_5029A0("Gamma");
		r.index[3] = game4_ref_sub_5029A0("absent");
		r.index[4] = game4_ref_sub_502A20();
		r.pointers[0] = (uintptr_t)game4_ref_sub_5029F0(2);
		r.pointers[1] = (uintptr_t)game4_ref_sub_5029F0(3);
		double x = game4_ref_sub_502E70(0), y = game4_ref_sub_502EA0(1);
		memcpy(r.coordinate_bits, &x, 8); memcpy(r.coordinate_bits + 1, &y, 8);
		r.setters[0] = game4_ref_sub_502A50("FirstMap");
		r.setters[1] = game4_ref_sub_502AB0("SecondMap");
		r.pointers[2] = (uintptr_t)game4_ref_sub_502A90();
		r.pointers[3] = (uintptr_t)game4_ref_sub_502AF0();
		r.setters[2] = game4_ref_sub_502A50(NULL);
		r.setters[3] = game4_ref_sub_502AB0(NULL);
	} else {
		r.index[0] = sub_5029A0("alpha");
		r.index[1] = sub_5029A0("BETA");
		r.index[2] = sub_5029A0("Gamma");
		r.index[3] = sub_5029A0("absent");
		r.index[4] = sub_502A20();
		r.pointers[0] = (uintptr_t)sub_5029F0(2);
		r.pointers[1] = (uintptr_t)sub_5029F0(3);
		double x = sub_502E70(0), y = sub_502EA0(1);
		memcpy(r.coordinate_bits, &x, 8); memcpy(r.coordinate_bits + 1, &y, 8);
		r.setters[0] = sub_502A50("FirstMap");
		r.setters[1] = sub_502AB0("SecondMap");
		r.pointers[2] = (uintptr_t)sub_502A90();
		r.pointers[3] = (uintptr_t)sub_502AF0();
		r.setters[2] = sub_502A50(NULL);
		r.setters[3] = sub_502AB0(NULL);
	}
	return r;
}

static int game4_map_archive_test_owner(game4_map_archive_test_bank* b) {
	return dword_5d4594_1599588 == b->first && dword_5d4594_1599592 == b->second &&
		dword_5d4594_1599576 == b->records && b->before == UINT64_C(0x7060504030201000) &&
		b->after == UINT64_C(0xffeeddccbbaa9988);
}

static void* game4_map_archive_test_payload(game4_map_archive_test_bank* b) { return b->first; }
static size_t game4_map_archive_test_payload_bytes(void) { return 4096 + 0x26000; }
static uintptr_t game4_map_archive_test_record_address(game4_map_archive_test_bank* b) { return (uintptr_t)b->records; }
static uintptr_t game4_map_archive_test_first_address(game4_map_archive_test_bank* b) { return (uintptr_t)b->first; }
static uintptr_t game4_map_archive_test_second_address(game4_map_archive_test_bank* b) { return (uintptr_t)b->second; }

static int game4_map_archive_test_index(game4_map_archive_test_bank* b, const char* path, int reference) {
	if (strlen(path) >= sizeof(b->first)) return -1;
	strcpy(b->first, path);
	return reference ? game4_ref_sub_502B10() : sub_502B10();
}
static uint32_t game4_map_archive_test_count(void) { return dword_5d4594_1599596; }
static int game4_map_archive_test_closed(void) { return nox_file_8 == NULL; }

static int game4_map_archive_test_extract(game4_map_archive_test_bank* b,
	const char* source, const char* directory, char* name, char* attachment, int reference) {
	char* live_directory = getMemAt(0x973F18, 42152);
	char old_directory[2048];
	if (strlen(source) >= sizeof(b->first) || strlen(directory) >= sizeof(old_directory)) return -1;
	memcpy(old_directory, live_directory, sizeof(old_directory));
	strcpy(b->first, source);
	strcpy(live_directory, directory);
	int r;
	if (reference) {
		r = game4_ref_sub_502B10() ? 1 : 0;
		r |= game4_ref_sub_5034B0(name) ? 2 : 0;
		r |= game4_ref_sub_5036D0(name, attachment) ? 4 : 0;
	} else {
		r = sub_502B10() ? 1 : 0;
		r |= sub_5034B0(name) ? 2 : 0;
		r |= sub_5036D0(name, attachment) ? 4 : 0;
	}
	memcpy(live_directory, old_directory, sizeof(old_directory));
	return r;
}

typedef struct game4_map_archive_test_geometry {
	int32_t corners[8];
	int32_t bounds[4];
	int32_t delta[2];
	uint32_t wall_bits;
	int canaries;
	uintptr_t record_address;
} game4_map_archive_test_geometry;

static game4_map_archive_test_geometry game4_map_archive_test_geometry_run(float x, float y, float w, float h,
	int32_t ox, int32_t oy, uint32_t wall_count, int reference) {
	struct { uint64_t before; uint8_t record[77]; uint64_t after; } b;
	memset(&b, 0xa5, sizeof(b));
	// Deliberately unaligned wire-record address; all record fields use memcpy.
	uint8_t* record = b.record + 1;
	memcpy(record + 64, &w, 4); memcpy(record + 68, &h, 4);
	uint8_t before[sizeof(b)]; memcpy(before, &b, sizeof(b));
	float2 at = {x, y}, fixed;
	nox_xxx_mapGenFixCoords_4D3D90(&at, &fixed);
	game4_map_archive_test_geometry r = {0};
	r.record_address = (uintptr_t)record;
	int4 bounds;
	float span;
	if (reference) {
		game4_ref_nox_mapgenBuildPlaceBounds_503B30(&at, &fixed, record, r.corners, &bounds);
		r.delta[0] = game4_ref_nox_mapgenPlaceOffset_503B30(fixed.field_0, ox);
		r.delta[1] = game4_ref_nox_mapgenPlaceOffset_503B30(fixed.field_4, oy);
		span = game4_ref_nox_mapgenWallSpan_503B30(wall_count);
	} else {
		nox_mapgenBuildPlaceBounds_503B30(&at, &fixed, record, r.corners, &bounds);
		r.delta[0] = nox_mapgenPlaceOffset_503B30(fixed.field_0, ox);
		r.delta[1] = nox_mapgenPlaceOffset_503B30(fixed.field_4, oy);
		span = nox_mapgenWallSpan_503B30(wall_count);
	}
	memcpy(r.bounds, &bounds, sizeof(bounds));
	memcpy(&r.wall_bits, &span, 4);
	r.canaries = memcmp(before, &b, sizeof(b)) == 0;
	return r;
}
static int32_t game4_map_archive_test_offset(float coordinate, int32_t origin, int reference) {
	return reference ? game4_ref_nox_mapgenPlaceOffset_503B30(coordinate, origin) :
		nox_mapgenPlaceOffset_503B30(coordinate, origin);
}
*/
import "C"

import "unsafe"

type game4ArchiveBank struct {
	ptr *C.game4_map_archive_test_bank
}

func game4ArchiveBankNew() *game4ArchiveBank {
	p := C.game4_map_archive_test_begin()
	if p == nil {
		return nil
	}
	return &game4ArchiveBank{ptr: p}
}
func (b *game4ArchiveBank) close() { C.game4_map_archive_test_end(b.ptr) }
func (b *game4ArchiveBank) seed()  { C.game4_map_archive_test_seed(b.ptr) }
func (b *game4ArchiveBank) address() uintptr {
	return uintptr(C.game4_map_archive_test_record_address(b.ptr))
}
func (b *game4ArchiveBank) nameAddresses() [2]uintptr {
	return [2]uintptr{uintptr(C.game4_map_archive_test_first_address(b.ptr)), uintptr(C.game4_map_archive_test_second_address(b.ptr))}
}
func (b *game4ArchiveBank) payload() []byte {
	return C.GoBytes(C.game4_map_archive_test_payload(b.ptr), C.int(C.game4_map_archive_test_payload_bytes()))
}
func (b *game4ArchiveBank) ownersIntact() bool { return C.game4_map_archive_test_owner(b.ptr) != 0 }

type game4ArchiveLeaf struct {
	indices     [5]int32
	pointers    [4]uintptr
	coordinates [2]uint64
	setters     [4]int32
}

func (b *game4ArchiveBank) leaf(reference bool) game4ArchiveLeaf {
	r := C.game4_map_archive_test_leaf(b.ptr, C.int(bool2int(reference)))
	var out game4ArchiveLeaf
	for i := range out.indices {
		out.indices[i] = int32(r.index[i])
	}
	for i := range out.pointers {
		out.pointers[i] = uintptr(r.pointers[i])
	}
	for i := range out.coordinates {
		out.coordinates[i] = uint64(r.coordinate_bits[i])
	}
	for i := range out.setters {
		out.setters[i] = int32(r.setters[i])
	}
	return out
}
func (b *game4ArchiveBank) index(path string, reference bool) (int, uint32, bool) {
	p := C.CString(path)
	defer C.free(unsafe.Pointer(p))
	r := C.game4_map_archive_test_index(b.ptr, p, C.int(bool2int(reference)))
	return int(r), uint32(C.game4_map_archive_test_count()), C.game4_map_archive_test_closed() != 0
}
func (b *game4ArchiveBank) extract(source, directory, name, attachment string, reference bool) int {
	s := C.CString(source)
	defer C.free(unsafe.Pointer(s))
	d := C.CString(directory)
	defer C.free(unsafe.Pointer(d))
	n := C.CString(name)
	defer C.free(unsafe.Pointer(n))
	a := C.CString(attachment)
	defer C.free(unsafe.Pointer(a))
	return int(C.game4_map_archive_test_extract(b.ptr, s, d, n, a, C.int(bool2int(reference))))
}

type game4ArchiveGeometry struct {
	corners  [8]int32
	bounds   [4]int32
	delta    [2]int32
	wallBits uint32
	canaries bool
}

func game4ArchiveGeometryReference(x, y, w, h float32, ox, oy int32, wall uint32, reference bool) game4ArchiveGeometry {
	r := C.game4_map_archive_test_geometry_run(C.float(x), C.float(y), C.float(w), C.float(h), C.int32_t(ox), C.int32_t(oy), C.uint32_t(wall), C.int(bool2int(reference)))
	out := game4ArchiveGeometry{wallBits: uint32(r.wall_bits), canaries: r.canaries != 0}
	for i := range out.corners {
		out.corners[i] = int32(r.corners[i])
	}
	for i := range out.bounds {
		out.bounds[i] = int32(r.bounds[i])
	}
	for i := range out.delta {
		out.delta[i] = int32(r.delta[i])
	}
	return out
}
func game4ArchiveOffsetReference(coordinate float32, origin int32, reference bool) int32 {
	return int32(C.game4_map_archive_test_offset(C.float(coordinate), C.int32_t(origin), C.int(bool2int(reference))))
}
