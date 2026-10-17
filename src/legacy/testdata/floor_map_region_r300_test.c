// Controlled I/O, allocation and geometry services for the complete production
// FloorMap bodies inserted by TestFloorMapNativeRegionDataR300. The separate
// TestFloorMapReadNativeRegionR300 exercises the real Go/CGo and cryptfile path.
#include <stdio.h>
#include <stdlib.h>
#include <string.h>
#include "GAME1.h"
#include "GAME1_1.h"
#include "GAME1_2.h"
#include "GAME4.h"
#include "common__crypt.h"
#include "operators.h"

obj_5D4594_2650668_t** ptr_5D4594_2650668;
int ptr_5D4594_2650668_cap;

// PRODUCTION_FLOOR_REGION_BODIES

static unsigned char wire[4096];
static size_t wire_size, cursor, widths[128], calls;
static int readonly_mode, geometry_calls;
struct region_box { uint32_t prefix[4], region[8], suffix[4]; };
struct node_box { unsigned char prefix[16]; nox_tile_list_node_t node; unsigned char suffix[16]; };
static struct region_box *box;
static struct node_box *nodes[16];
static size_t node_count;
static const char *case_name;
static void need(int ok, const char *why) {
    if (!ok) { fprintf(stderr, "R300 FAIL case=%s reason=%s cursor=%zu calls=%zu\n", case_name, why, cursor, calls); exit(71); }
}
int nox_crypt_IsReadOnly(void) { return readonly_mode; }
bool nox_common_gameFlags_check_40A5C0(unsigned int flags) {
    need(flags == 0x400000, "unexpected flags query");
    return false;
}
size_t nox_xxx_fileReadWrite_426AC0_file3_fread_impl(unsigned char *p, size_t n, char *file, int line) {
    (void)file; (void)line;
    need(calls < sizeof(widths) / sizeof(widths[0]), "I/O call bound");
    widths[calls++] = n;
    need(cursor <= sizeof(wire) && n <= sizeof(wire) - cursor, "buffer bounds");
    if (readonly_mode) {
        need(cursor <= wire_size && n <= wire_size - cursor, "read past declared input");
        memcpy(p, wire + cursor, n);
    } else {
        memcpy(wire + cursor, p, n);
        wire_size = cursor + n;
    }
    cursor += n;
    return n;
}
nox_tile_coord_entry_t* nox_xxx_tileAllocTileInCoordList_5040A0(int x, int y, float kind) {
    (void)x; (void)y; (void)kind;
    need(0, "unexpected mapgen coordinate allocation"); return NULL;
}
nox_tile_list_node_t* nox_xxx_tileListAddNewSubtile_422160(int a, int b, int c, int d) {
    need(node_count < 16, "subtile allocation bound");
    struct node_box *n = malloc(sizeof(*n));
    need(n != NULL, "subtile allocation");
    memset(n, 0xa7, sizeof(*n));
    n->node = (nox_tile_list_node_t){a, b, c, d, NULL};
    need((uintptr_t)&n->node > UINT32_MAX, "native subtile address");
    nodes[node_count++] = n;
    return &n->node;
}
int nox_xxx_wallMath_427F30(int2 *point, int *region) {
    static const int2 expected[] = {{80, 103}, {57, 126}};
    need(!readonly_mode && geometry_calls < 2, "geometry call phase/count");
    need(region == (int*)box->region, "full original region passed to geometry");
    need(point->field_0 == expected[geometry_calls].field_0 && point->field_4 == expected[geometry_calls].field_4, "geometry call order/point");
    ++geometry_calls;
    return 1; // Selected geometry result only; not the actual polygon algorithm.
}
static int go_call_int_ptr_func(int (*fnc)(void*), void* a1) { return fnc(a1); }
static void put8(unsigned x) { need(wire_size < sizeof(wire), "pack8"); wire[wire_size++] = (unsigned char)x; }
static void put16(unsigned x) { put8(x); put8(x >> 8); }
static void put32(unsigned x) { put16(x); put16(x >> 16); }
static void header(unsigned version, unsigned x, unsigned y, unsigned w, unsigned h) { put16(version); put32(x); put32(y); put32(w); put32(h); }
static void scalar4(unsigned a, int b, unsigned c, unsigned d, unsigned count) { put8(a); put16((unsigned)b); put8(c); put8(d); put8(count); }
static void subtile4(unsigned a, int b, unsigned c, unsigned d) { put8(a); put16((unsigned)b); put8(c); put8(d); }
static void setup(const char *name) {
    case_name = name;
    wire_size = cursor = calls = node_count = 0;
    geometry_calls = 0; readonly_mode = 1;
    memset(wire, 0, sizeof(wire)); memset(widths, 0, sizeof(widths));
    box = malloc(sizeof(*box)); need(box != NULL, "region allocation");
    memset(box, 0xa7, sizeof(*box));
    uint32_t region[8] = {0x13243546, 69, 46, 0x24354657, 92, 0x35465768, 0x46576879, 115};
    memcpy(box->region, region, sizeof(region));
    need(sizeof(void*) == 8 && (uintptr_t)box->region > UINT32_MAX, "LP64 high region");
    ptr_5D4594_2650668_cap = 8;
    ptr_5D4594_2650668 = calloc(8, sizeof(*ptr_5D4594_2650668));
    need(ptr_5D4594_2650668 != NULL, "row pointer allocation");
    for (int i = 0; i < 8; ++i) {
        ptr_5D4594_2650668[i] = calloc(8, sizeof(**ptr_5D4594_2650668));
        need(ptr_5D4594_2650668[i] != NULL && (uintptr_t)ptr_5D4594_2650668[i] > UINT32_MAX, "LP64 native grid row");
    }
    printf("R300 START {\"case\":\"%s\",\"region_ptr\":\"%p\",\"grid_ptr\":\"%p\",\"controlled_services\":true}\n", name, (void*)box->region, (void*)ptr_5D4594_2650668[0]);
    fflush(stdout);
}
static void finish(struct region_box before) {
    need(memcmp(box, &before, sizeof(before)) == 0, "region/canary changed");
    for (size_t i = 0; i < node_count; ++i) {
        for (int j = 0; j < 16; ++j) need(nodes[i]->prefix[j] == 0xa7 && nodes[i]->suffix[j] == 0xa7, "subtile canary");
    }
    printf("R300 RESULT {\"case\":\"%s\",\"passed\":true,\"bytes\":%zu,\"I_O_calls\":%zu,\"native_subtiles\":%zu,\"geometry_calls\":%d,\"canaries_unchanged\":true}\n", case_name, cursor, calls, node_count, geometry_calls);
    for (size_t i = 0; i < node_count; ++i) free(nodes[i]);
    for (int i = 0; i < 8; ++i) free(ptr_5D4594_2650668[i]);
    free(ptr_5D4594_2650668); ptr_5D4594_2650668 = NULL; ptr_5D4594_2650668_cap = 0;
    free(box); box = NULL;
}
static void invoke(void) { need(go_call_int_ptr_func((int (*)(void*))nox_server_mapRWFloorMap_422230, box->region) == 1, "actual source entry result"); }
static void test_v4_read(void) {
    setup("v4-negative-row-delta-read"); struct region_box before = *box;
    header(4, 5, 7, 3, 3); // Translated origin: signed row delta is -4.
    put16(0x0708); scalar4(0xa1, -1234, 0xb2, 0xc3, 1); subtile4(0xd1, -42, 0xe1, 0xf1);
    put16(0x0809); scalar4(0xa2, -2222, 0xb3, 0xc4, 0); put16(0xffff);
    invoke();
    need(cursor == wire_size && cursor == 41, "v4 exact consumed compact wire");
    obj_5D4594_2650668_t *first = &ptr_5D4594_2650668[2][2], *second = &ptr_5D4594_2650668[2][3];
    need(first->field_0 == 2 && first->field_6 == 0xa1 && first->field_7 == -1234 && first->field_8 == 0xb2 && first->field_9 == 0xc3, "v4 first translated layer");
    need(second->field_0 == 1 && second->field_1 == 0xa2 && second->field_2 == -2222 && second->field_3 == 0xb3 && second->field_4 == 0xc4 && !second->field_5, "v4 reset signed delta for second layer");
    nox_tile_list_node_t *n = first->field_10;
    need(node_count == 1 && n == &nodes[0]->node && n->field_0 == 0xd1 && n->field_1 == -42 && n->field_2 == 0xe1 && n->field_3 == 0xf1 && !n->next, "v4 native subtile contents/link");
    finish(before);
}
static void test_v3_read(void) {
    setup("v3-region-two-column-read"); struct region_box before = *box;
    header(3, 0, 0, 2, 1);
    const int data[2][4] = {{11, -1234, 33, 44}, {55, -2222, 77, 88}};
    for (int i = 0; i < 2; ++i) { put8(1); for (int j = 0; j < 4; ++j) put32((unsigned)data[i][j]); put8(0); }
    invoke();
    need(cursor == wire_size && cursor == 54 && calls == 11, "v3 exact two-column I/O, not native layer size");
    for (int i = 0; i < 2; ++i) {
        obj_5D4594_2650668_t *tile = &ptr_5D4594_2650668[i][2];
        need(tile->field_0 == 1 && tile->field_1 == data[i][0] && tile->field_2 == data[i][1] && tile->field_3 == data[i][2] && tile->field_4 == data[i][3] && !tile->field_5, "source-derived v3 scalar fields");
    }
    need(node_count == 0, "v3 count-zero allocation"); finish(before);
}
static void test_v4_write(void) {
    setup("v4-region-write-native-links"); struct region_box before = *box;
    obj_5D4594_2650668_t *tile = &ptr_5D4594_2650668[1][2];
    tile->field_0 = 3;
    tile->field_1 = 0xa1; tile->field_2 = -1234; tile->field_3 = 0xb2; tile->field_4 = 0xc3;
    tile->field_6 = 0xa2; tile->field_7 = -2222; tile->field_8 = 0xb3; tile->field_9 = 0xc4;
    tile->field_5 = nox_xxx_tileListAddNewSubtile_422160(0xd1, -42, 0xe1, 0xf1);
    tile->field_5->next = nox_xxx_tileListAddNewSubtile_422160(0xd2, -84, 0xe2, 0xf2);
    header(4, 2, 3, 3, 3);
    put16(0x0303); scalar4(0xa1, -1234, 0xb2, 0xc3, 2); subtile4(0xd1, -42, 0xe1, 0xf1); subtile4(0xd2, -84, 0xe2, 0xf2);
    put16(0x0204); scalar4(0xa2, -2222, 0xb3, 0xc4, 0); put16(0xffff);
    unsigned char expected[4096]; size_t expected_size = wire_size;
    memcpy(expected, wire, wire_size); memset(wire, 0, sizeof(wire));
    readonly_mode = 0; wire_size = cursor = calls = 0;
    invoke();
    need(expected_size == 46 && wire_size == expected_size && cursor == expected_size && memcmp(wire, expected, expected_size) == 0, "v4 exact writer wire and 16-bit field widths");
    need(geometry_calls == 2 && tile->field_5 == &nodes[0]->node && tile->field_5->next == &nodes[1]->node && !nodes[1]->node.next, "native links and exact geometry calls");
    finish(before);
}
int main(int argc, char **argv) {
    if (argc != 2) return 64;
    if (!strcmp(argv[1], "v4-negative-row-delta-read")) test_v4_read();
    else if (!strcmp(argv[1], "v3-region-two-column-read")) test_v3_read();
    else if (!strcmp(argv[1], "v4-region-write-native-links")) test_v4_write();
    else if (!strcmp(argv[1], "all")) { test_v4_read(); test_v3_read(); test_v4_write(); }
    else return 64;
    return 0;
}
