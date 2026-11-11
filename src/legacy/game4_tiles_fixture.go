package legacy

/*
#include <stdint.h>
#include <stdlib.h>
#include <string.h>
#include "GAME4.h"
extern void* game4_fixture_calloc_go(size_t count, size_t size);
extern void game4_fixture_release_go(void* p);

extern nox_tile_coord_entry_t* dword_5d4594_1599556;
extern uint32_t nox_tile_def_cnt;
extern nox_tileDef_t nox_tile_defs_arr[176];

typedef struct game4_tiles_test_event { uint32_t code, value, x_bits, y_bits; } game4_tiles_test_event;
typedef struct game4_tiles_test_data {
	uint64_t before;
	nox_tile_coord_entry_t entries[3];
	nox_tile_layer_t layers[4];
	nox_tile_list_node_t nodes[3];
	uint8_t state[72];
	game4_tiles_test_event events[64];
	uint32_t event_count, overflow, mutate, first_image, first_mode, first_subtile;
	uint64_t after;
} game4_tiles_test_data;
typedef struct game4_tiles_test_bank {
	nox_tile_coord_entry_t* old_head;
	uint32_t old_count, old_def_count;
	int32_t old_transparent;
	nox_tileDef_t old_defs[3];
	struct game4_tiles_test_bank* old_active;
	nox_tile_coord_entry_t* allocated[8];
	uint32_t allocated_count;
	game4_tiles_test_data data;
} game4_tiles_test_bank;
static game4_tiles_test_bank* game4_tiles_test_active;

static void game4_tiles_test_event_add(uint32_t code, int value, const float2* at) {
	game4_tiles_test_data* d = &game4_tiles_test_active->data;
	if (d->event_count >= 64) { d->overflow = 1; return; }
	game4_tiles_test_event* e = &d->events[d->event_count++];
	e->code = code; e->value = (uint32_t)value;
	if (at) { memcpy(&e->x_bits, &at->field_0, 4); memcpy(&e->y_bits, &at->field_4, 4); }
	d->state[code] ^= (uint8_t)(code * 7);
}
static unsigned char* game4_tiles_test_snapshot(void) {
	game4_tiles_test_event_add(1, 0, NULL);
	if (game4_tiles_test_active->data.mutate) dword_5d4594_1599556 = &game4_tiles_test_active->data.entries[0];
	return game4_tiles_test_active->data.state;
}
static void game4_tiles_test_snapshot_copy(void* dst) { memcpy(dst, game4_tiles_test_snapshot(), 72); }
static void game4_tiles_test_restore(const void* saved) {
	game4_tiles_test_event_add(9, 0, NULL);
	memcpy(game4_tiles_test_active->data.state, saved, 72);
}
static int game4_tiles_test_image(int value) {
	game4_tiles_test_event_add(2, value, NULL);
	game4_tiles_test_data* d = &game4_tiles_test_active->data;
	if (d->mutate && !d->first_image++) {
		d->entries[0].layer = &d->layers[2];
		d->entries[0].x = 999; d->entries[0].y = -999;
		dword_5d4594_1599556 = &d->entries[1];
	}
	return -17;
}
static int game4_tiles_test_variation(int value) { game4_tiles_test_event_add(3, value, NULL); return 0; }
static int game4_tiles_test_mode(int value) {
	game4_tiles_test_event_add(4, value, NULL);
	game4_tiles_test_data* d = &game4_tiles_test_active->data;
	if (d->mutate && !d->first_mode++) *getMemIntPtr(0x587000, 229704) = -1;
	return -23;
}
static int game4_tiles_test_base(float2* at) { game4_tiles_test_event_add(5, 0, at); return 0; }
static int game4_tiles_test_byte3(int value) { game4_tiles_test_event_add(6, value, NULL); return -1; }
static int game4_tiles_test_byte4(int value) { game4_tiles_test_event_add(7, value, NULL); return 0; }
static int game4_tiles_test_subtile(float2* at) {
	game4_tiles_test_event_add(8, 0, at);
	game4_tiles_test_data* d = &game4_tiles_test_active->data;
	if (d->mutate && !d->first_subtile++) {
		d->nodes[0].next = &d->nodes[2];
		d->entries[0].prev = &d->entries[2];
	}
	return -31;
}

#define calloc game4_fixture_calloc_go
#define free game4_fixture_release_go
#include "game4_tiles_reference.h"
#undef free
#undef calloc

static game4_tiles_test_bank* game4_tiles_test_begin(void) {
	game4_tiles_test_bank* b = calloc(1, sizeof(*b));
	if (!b) return NULL;
	b->old_head = dword_5d4594_1599556;
	b->old_count = *getMemU32Ptr(0x5D4594, 1599560);
	b->old_def_count = nox_tile_def_cnt;
	b->old_transparent = *getMemIntPtr(0x587000, 229704);
	memcpy(b->old_defs, nox_tile_defs_arr, sizeof(b->old_defs));
	b->old_active = game4_tiles_test_active;
	game4_tiles_test_active = b;
	dword_5d4594_1599556 = NULL;
	return b;
}
static void game4_tiles_test_end(game4_tiles_test_bank* b) {
	// Remove every public owner before freeing any fixture-owned C memory.
	dword_5d4594_1599556 = b->old_head;
	*getMemU32Ptr(0x5D4594, 1599560) = b->old_count;
	nox_tile_def_cnt = b->old_def_count;
	*getMemIntPtr(0x587000, 229704) = b->old_transparent;
	memcpy(nox_tile_defs_arr, b->old_defs, sizeof(b->old_defs));
	game4_tiles_test_active = b->old_active;
	for (uint32_t i = 0; i < b->allocated_count; ++i) {
		game4_fixture_release_go(b->allocated[i]->layer); game4_fixture_release_go(b->allocated[i]);
	}
	free(b);
}
static void game4_tiles_test_seed(game4_tiles_test_bank* b, int shape, int definitions, int32_t cached) {
	memset(&b->data, 0, sizeof(b->data));
	game4_tiles_test_data* d = &b->data;
	d->before = UINT64_C(0x123456789abcdef0); d->after = UINT64_C(0xfedcba9876543210);
	for (int i = 0; i < 72; ++i) d->state[i] = (uint8_t)(7 + 3*i);
	for (int i = 0; i < 3; ++i) {
		d->entries[i].layer = &d->layers[i];
		d->entries[i].x = (float)(10 + i); d->entries[i].y = (float)(-20 - i);
		d->layers[i].field_0 = i + 1; d->layers[i].field_1 = 100 + i;
		d->nodes[i].field_0 = (int)(0x80000000u + i);
		d->nodes[i].field_1 = -200 - i;
		d->nodes[i].field_2 = (int)(0xfedcba98u + i);
		d->nodes[i].field_3 = 300 + i;
	}
	d->nodes[0].next = &d->nodes[1]; d->nodes[1].next = &d->nodes[2];
	d->entries[0].prev = &d->entries[1];
	d->layers[0].subtiles = &d->nodes[0];
	d->layers[2].field_0 = UINT32_MAX; d->layers[2].field_1 = 0x800000ffu;
	d->layers[2].subtiles = &d->nodes[0];
	d->layers[3].field_0 = 33; d->layers[3].field_1 = 44;
	d->entries[2].layer = &d->layers[3];
	d->mutate = shape == 1;
	if (d->mutate) d->layers[0].subtiles = &d->nodes[1];
	dword_5d4594_1599556 = shape == 2 ? NULL : &d->entries[d->mutate ? 1 : 0];
	memset(nox_tile_defs_arr, 0, sizeof(b->old_defs));
	strcpy(nox_tile_defs_arr[0].name, "Stone");
	strcpy(nox_tile_defs_arr[1].name, definitions == 1 ? "transparentfloor" : "TransparentFloor");
	strcpy(nox_tile_defs_arr[2].name, "Other");
	nox_tile_def_cnt = definitions == 2 ? 0 : 3;
	*getMemIntPtr(0x587000, 229704) = cached;
	*getMemU32Ptr(0x5D4594, 1599560) = UINT32_MAX;
}
static int game4_tiles_test_init(int x, int y, int reference) {
	return reference ? game4_tiles_ref_init(x, y) : nox_xxx_tileInit_504150(x, y);
}
static int32_t game4_tiles_test_cached(void) { return *getMemIntPtr(0x587000, 229704); }
static void* game4_tiles_test_data_ptr(game4_tiles_test_bank* b) { return &b->data; }
static size_t game4_tiles_test_data_bytes(void) { return sizeof(game4_tiles_test_data); }
static int game4_tiles_test_canaries(game4_tiles_test_bank* b) {
	return b->data.before == UINT64_C(0x123456789abcdef0) && b->data.after == UINT64_C(0xfedcba9876543210) && !b->data.overflow;
}
static void game4_tiles_test_alloc_reset(void) { dword_5d4594_1599556 = NULL; *getMemU32Ptr(0x5D4594, 1599560) = UINT32_MAX; }
typedef struct game4_tiles_test_alloc_result {
	uintptr_t entry, layer;
	uint32_t x_bits, y_bits, count, kind, links, zero_layer;
} game4_tiles_test_alloc_result;
static game4_tiles_test_alloc_result game4_tiles_test_alloc(game4_tiles_test_bank* b, int x, int y, uint32_t bits, int reference) {
	game4_tiles_test_alloc_result r = {0};
	if (b->allocated_count >= 8) return r;
	nox_tile_coord_entry_t* old = dword_5d4594_1599556;
	float kind; memcpy(&kind, &bits, 4);
	nox_tile_coord_entry_t* p = reference ? game4_tiles_ref_alloc(x,y,kind) : nox_xxx_tileAllocTileInCoordList_5040A0(x,y,kind);
	if (!p) return r;
	b->allocated[b->allocated_count++] = p;
	r.entry = (uintptr_t)p; r.layer = (uintptr_t)p->layer;
	memcpy(&r.x_bits, &p->x, 4); memcpy(&r.y_bits, &p->y, 4);
	r.count = *getMemU32Ptr(0x5D4594, 1599560); r.kind = p->kind;
	r.links = dword_5d4594_1599556 == p && p->prev == old && p->next == NULL && (!old || old->next == p);
	nox_tile_layer_t zero = {0}; r.zero_layer = memcmp(&zero, p->layer, sizeof(zero)) == 0;
	return r;
}
static game4_tiles_test_alloc_result game4_tiles_test_alloc_existing(game4_tiles_test_bank* b, uint32_t index) {
	game4_tiles_test_alloc_result r = {0};
	if (b->allocated_count < 4 || index >= 4) return r;
	uint32_t start = b->allocated_count - 4;
	nox_tile_coord_entry_t* p = b->allocated[start + index];
	r.entry = (uintptr_t)p; r.layer = (uintptr_t)p->layer;
	memcpy(&r.x_bits, &p->x, 4); memcpy(&r.y_bits, &p->y, 4);
	r.count = *getMemU32Ptr(0x5D4594, 1599560); r.kind = p->kind;
	r.links = dword_5d4594_1599556 == b->allocated[start+3] &&
		p->prev == (index ? b->allocated[start+index-1] : NULL) &&
		p->next == (index < 3 ? b->allocated[start+index+1] : NULL);
	nox_tile_layer_t zero = {0}; r.zero_layer = memcmp(&zero, p->layer, sizeof(zero)) == 0;
	return r;
}
*/
import "C"

import "unsafe"

type game4TilesTestBank struct {
	ptr           *C.game4_tiles_test_bank
	releaseMemory func()
}

func game4TilesTestBegin() *game4TilesTestBank {
	releaseMemory := game4SupportHighFixtureMemory()
	p := C.game4_tiles_test_begin()
	if p == nil {
		releaseMemory()
		return nil
	}
	return &game4TilesTestBank{ptr: p, releaseMemory: releaseMemory}
}
func (b *game4TilesTestBank) close() {
	C.game4_tiles_test_end(b.ptr)
	b.releaseMemory()
}
func (b *game4TilesTestBank) seed(shape, definitions int, cached int32) {
	C.game4_tiles_test_seed(b.ptr, C.int(shape), C.int(definitions), C.int32_t(cached))
}
func (b *game4TilesTestBank) data() []byte {
	return C.GoBytes(C.game4_tiles_test_data_ptr(b.ptr), C.int(C.game4_tiles_test_data_bytes()))
}
func (b *game4TilesTestBank) canaries() bool { return C.game4_tiles_test_canaries(b.ptr) != 0 }
func game4TilesTestInit(x, y int32, reference bool) (int32, int32) {
	return int32(C.game4_tiles_test_init(C.int(x), C.int(y), C.int(bool2int(reference)))), int32(C.game4_tiles_test_cached())
}
func game4TilesTestServices() game4TilesServices {
	return game4TilesServices{
		snapshot:  func() [72]byte { var b [72]byte; C.game4_tiles_test_snapshot_copy(unsafe.Pointer(&b[0])); return b },
		restore:   func(b *[72]byte) { C.game4_tiles_test_restore(unsafe.Pointer(&b[0])) },
		image:     func(v int32) { C.game4_tiles_test_image(C.int(v)) },
		variation: func(v int32) { C.game4_tiles_test_variation(C.int(v)) },
		mode:      func(v int32) { C.game4_tiles_test_mode(C.int(v)) },
		base:      func(p *C.float2) { C.game4_tiles_test_base(p) },
		byte3:     func(v int32) { C.game4_tiles_test_byte3(C.int(v)) },
		byte4:     func(v int32) { C.game4_tiles_test_byte4(C.int(v)) },
		subtile:   func(p *C.float2) { C.game4_tiles_test_subtile(p) },
	}
}

type game4TilesAllocObservation struct {
	entry, layer              uintptr
	xBits, yBits, count, kind uint32
	links, zeroLayer          bool
}

func game4TilesTestAllocReset() { C.game4_tiles_test_alloc_reset() }
func (b *game4TilesTestBank) alloc(x, y int32, bits uint32, reference bool) game4TilesAllocObservation {
	r := C.game4_tiles_test_alloc(b.ptr, C.int(x), C.int(y), C.uint32_t(bits), C.int(bool2int(reference)))
	return game4TilesAllocObservation{uintptr(r.entry), uintptr(r.layer), uint32(r.x_bits), uint32(r.y_bits), uint32(r.count), uint32(r.kind), r.links != 0, r.zero_layer != 0}
}
func (b *game4TilesTestBank) existing(index uint32) game4TilesAllocObservation {
	r := C.game4_tiles_test_alloc_existing(b.ptr, C.uint32_t(index))
	return game4TilesAllocObservation{uintptr(r.entry), uintptr(r.layer), uint32(r.x_bits), uint32(r.y_bits), uint32(r.count), uint32(r.kind), r.links != 0, r.zero_layer != 0}
}
