// The complete production codec is inserted below; I/O, allocation and region
// selection are controlled to expose exact wire bytes and callback ordering.
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

// PRODUCTION_FLOOR_V3_BODIES

enum { CAP = 128, MAX_NODES = 32 };
static unsigned char wire[4096], expected[4096];
static size_t wire_size, cursor, calls, expected_size, expected_calls;
static size_t widths[256], expected_widths[256];
static int readonly_mode, geometry_calls;
static const char *case_name;
struct region_box { uint32_t prefix[4], region[8], suffix[4]; };
struct row_box { unsigned char prefix[16]; obj_5D4594_2650668_t row[CAP]; unsigned char suffix[16]; };
struct node_box { unsigned char prefix[16]; nox_tile_list_node_t node; unsigned char suffix[16]; };
static struct region_box *box;
static struct row_box *rows[CAP];
static struct node_box *nodes[MAX_NODES];
static size_t node_count;
static void need(int ok, const char *why) {
    if (!ok) { fprintf(stderr, "R302 FAIL case=%s reason=%s cursor=%zu calls=%zu\n", case_name, why, cursor, calls); exit(71); }
}
int nox_crypt_IsReadOnly(void) { return readonly_mode; }
bool nox_common_gameFlags_check_40A5C0(unsigned int flags) {
    need(flags == 0x400000, "unexpected flags query"); return false;
}
size_t nox_xxx_fileReadWrite_426AC0_file3_fread_impl(unsigned char *p, size_t n, char *file, int line) {
    (void)file; (void)line;
    need(calls < sizeof(widths) / sizeof(widths[0]), "I/O call bound"); widths[calls++] = n;
    need(cursor <= sizeof(wire) && n <= sizeof(wire) - cursor, "buffer bounds");
    if (readonly_mode) { need(cursor <= wire_size && n <= wire_size - cursor, "read past input"); memcpy(p, wire + cursor, n); }
    else { memcpy(wire + cursor, p, n); wire_size = cursor + n; }
    cursor += n; return n;
}
nox_tile_coord_entry_t* nox_xxx_tileAllocTileInCoordList_5040A0(int x, int y, float kind) {
    (void)x; (void)y; (void)kind; need(0, "unexpected mapgen allocation"); return NULL;
}
nox_tile_list_node_t* nox_xxx_tileListAddNewSubtile_422160(int a, int b, int c, int d) {
    need(node_count < MAX_NODES, "node allocation bound");
    struct node_box *n = malloc(sizeof(*n)); need(n != NULL, "node allocation");
    memset(n, 0xa7, sizeof(*n)); n->node = (nox_tile_list_node_t){a, b, c, d, NULL};
    need((uintptr_t)&n->node > UINT32_MAX, "native node above 4GiB"); nodes[node_count++] = n; return &n->node;
}
int nox_xxx_wallMath_427F30(int2 *point, int *region) {
    static const int2 points[] = {{80, 103}, {57, 126}, {80, 149}};
    static const size_t positions[] = {17, 68, 105}, call_counts[] = {5, 11, 18};
    need(!readonly_mode && geometry_calls < 3, "geometry call phase/count");
    need(region == (int*)box->region, "original native region pointer");
    need(point->field_0 == points[geometry_calls].field_0 && point->field_4 == points[geometry_calls].field_4, "geometry point/order");
    need(cursor == positions[geometry_calls] && calls == call_counts[geometry_calls], "geometry callback order relative to I/O");
    return ++geometry_calls < 3; // The third tile is outside the selected shape.
}
static void pack(unsigned value, size_t n) {
    need(expected_calls < sizeof(expected_widths) / sizeof(expected_widths[0]), "expected I/O bound");
    expected_widths[expected_calls++] = n;
    for (size_t i = 0; i < n; ++i) expected[expected_size++] = (unsigned char)(value >> (8 * i));
}
static void fields(int a, int b, int c, int d) {
    size_t start = expected_calls; pack((unsigned)a, 4); pack((unsigned)b, 4); pack((unsigned)c, 4); pack((unsigned)d, 4);
    expected_calls = start; expected_widths[expected_calls++] = 16;
}
static void header(int x, int y, int w, int h) { pack((unsigned)x, 4); pack((unsigned)y, 4); pack((unsigned)w, 4); pack((unsigned)h, 4); }
static void layer1(void) { fields(0x11223344, -1234567, 0x55667788, -7654321); pack(2, 1); fields(11, -22, 33, -44); fields(55, -66, 77, -88); }
static void layer2(void) { fields(0x12345678, -2222222, 0x34567812, -3333333); pack(1, 1); fields(99, -111, 222, -333); }
static void grid_wire(void) {
    header(1, 2, 3, 3);
    for (int y = 2; y <= 4; ++y) for (int x = 1; x <= 3; ++x) {
        unsigned mask = x == 1 && y == 2 ? 3 : x == 3 && y == 4 ? 2 : 0;
        pack(mask, 1);
        if (mask & 1) layer1();
        if (mask & 2) { if (mask == 3) layer2(); else { fields(-11, 22, -33, 44); pack(0, 1); } }
    }
}
static void set_layer1(obj_5D4594_2650668_t *tile) {
    tile->field_0 |= 1; tile->field_1 = 0x11223344; tile->field_2 = -1234567; tile->field_3 = 0x55667788; tile->field_4 = -7654321;
    tile->field_5 = nox_xxx_tileListAddNewSubtile_422160(11, -22, 33, -44);
    tile->field_5->next = nox_xxx_tileListAddNewSubtile_422160(55, -66, 77, -88);
}
static void set_layer2(obj_5D4594_2650668_t *tile) {
    tile->field_0 |= 2; tile->field_6 = 0x12345678; tile->field_7 = -2222222; tile->field_8 = 0x34567812; tile->field_9 = -3333333;
    tile->field_10 = nox_xxx_tileListAddNewSubtile_422160(99, -111, 222, -333);
}
static void check_layers(obj_5D4594_2650668_t *tile) {
    need(tile->field_0 == 3 && tile->field_1 == 0x11223344 && tile->field_2 == -1234567 && tile->field_3 == 0x55667788 && tile->field_4 == -7654321, "first native layer fields");
    need(tile->field_6 == 0x12345678 && tile->field_7 == -2222222 && tile->field_8 == 0x34567812 && tile->field_9 == -3333333, "second native layer fields");
    nox_tile_list_node_t *a = tile->field_5, *b = a ? a->next : NULL, *c = tile->field_10;
    need(a == &nodes[0]->node && b == &nodes[1]->node && c == &nodes[2]->node && !b->next && !c->next, "native head/next pointers and order");
    need(a->field_0 == 11 && a->field_1 == -22 && a->field_2 == 33 && a->field_3 == -44, "first node fields");
    need(b->field_0 == 55 && b->field_1 == -66 && b->field_2 == 77 && b->field_3 == -88, "second node fields");
    need(c->field_0 == 99 && c->field_1 == -111 && c->field_2 == 222 && c->field_3 == -333, "second layer node fields");
}
static void setup(const char *name) {
    case_name = name; wire_size = cursor = calls = expected_size = expected_calls = node_count = 0; geometry_calls = 0; readonly_mode = 0;
    memset(wire, 0, sizeof(wire)); memset(expected, 0, sizeof(expected));
    box = malloc(sizeof(*box)); need(box != NULL, "region allocation"); memset(box, 0xa7, sizeof(*box));
    const uint32_t region[8] = {0x13243546, 69, 46, 0x24354657, 92, 0x35465768, 0x46576879, 115};
    memcpy(box->region, region, sizeof(region)); need((uintptr_t)box->region > UINT32_MAX, "native region above 4GiB");
    ptr_5D4594_2650668_cap = CAP; ptr_5D4594_2650668 = calloc(CAP, sizeof(*ptr_5D4594_2650668)); need(ptr_5D4594_2650668 != NULL, "grid allocation");
    for (int x = 0; x < CAP; ++x) {
        rows[x] = malloc(sizeof(*rows[x])); need(rows[x] != NULL, "row allocation"); memset(rows[x], 0xa7, sizeof(*rows[x])); memset(rows[x]->row, 0, sizeof(rows[x]->row));
        ptr_5D4594_2650668[x] = rows[x]->row; need((uintptr_t)rows[x]->row > UINT32_MAX, "native row above 4GiB");
    }
    printf("R302 START case=%s grid=%p region=%p controlled_services=true\n", name, (void*)rows[0]->row, (void*)box->region); fflush(stdout);
}
static void finish(struct region_box before) {
    need(memcmp(box, &before, sizeof(before)) == 0, "region and canaries unchanged");
    need(cursor == expected_size && wire_size == expected_size && memcmp(wire, expected, expected_size) == 0, "exact v3 wire bytes");
    need(calls == expected_calls && memcmp(widths, expected_widths, calls * sizeof(widths[0])) == 0, "exact callback widths and order");
    for (int x = 0; x < CAP; ++x) for (int i = 0; i < 16; ++i) need(rows[x]->prefix[i] == 0xa7 && rows[x]->suffix[i] == 0xa7, "row canary");
    for (size_t n = 0; n < node_count; ++n) for (int i = 0; i < 16; ++i) need(nodes[n]->prefix[i] == 0xa7 && nodes[n]->suffix[i] == 0xa7, "node canary");
    printf("R302 RESULT case=%s passed=true bytes=%zu callbacks=%zu nodes=%zu geometry=%d\n", case_name, cursor, calls, node_count, geometry_calls);
    for (size_t n = 0; n < node_count; ++n) free(nodes[n]);
    for (int x = 0; x < CAP; ++x) free(rows[x]);
    free(ptr_5D4594_2650668); free(box);
}
static void test_grid(int read) {
    setup(read ? "nil-region-read" : "nil-region-write"); struct region_box before = *box; grid_wire();
    obj_5D4594_2650668_t *tile = &ptr_5D4594_2650668[1][2], *last = &ptr_5D4594_2650668[3][4];
    if (read) {
        readonly_mode = 1; memcpy(wire, expected, expected_size); wire_size = expected_size;
        for (int x = 0; x < CAP; ++x) for (int y = 0; y < CAP; ++y) {
            ptr_5D4594_2650668[x][y].field_0 = 3;
            ptr_5D4594_2650668[x][y].field_5 = (nox_tile_list_node_t*)(uintptr_t)0x123456789abcdef0;
            ptr_5D4594_2650668[x][y].field_10 = (nox_tile_list_node_t*)(uintptr_t)0x23456789abcdef01;
        }
    } else {
        set_layer1(tile); set_layer2(tile); last->field_0 = 2; last->field_6 = -11; last->field_7 = 22; last->field_8 = -33; last->field_9 = 44;
    }
    obj_5D4594_2650668_t snapshot[CAP][CAP]; for (int x = 0; x < CAP; ++x) memcpy(snapshot[x], rows[x]->row, sizeof(snapshot[x]));
    need(nox_xxx_tile_422C10(3, NULL) == 1, "actual nil-region codec result"); check_layers(tile);
    need(last->field_0 == 2 && last->field_6 == -11 && last->field_7 == 22 && last->field_8 == -33 && last->field_9 == 44 && !last->field_10, "last tile count-zero layer");
    for (int x = 0; x < CAP; ++x) for (int y = 0; y < CAP; ++y) {
        obj_5D4594_2650668_t *t = &ptr_5D4594_2650668[x][y];
        if (read && t != tile && t != last) {
            obj_5D4594_2650668_t unchanged = snapshot[x][y]; unchanged.field_0 = 0;
            need(memcmp(t, &unchanged, sizeof(*t)) == 0, "nil reader clears flags and preserves unrelated fields/heads");
        }
        if (!read) need(memcmp(t, &snapshot[x][y], sizeof(*t)) == 0, "writer preserves native tile fields and links");
    }
    need(geometry_calls == 0 && node_count == 3, "nil codec node and geometry counts"); finish(before);
}
static void test_region_write(void) {
    setup("geometry-region-write"); struct region_box before = *box;
    obj_5D4594_2650668_t *tile = &ptr_5D4594_2650668[1][2]; set_layer1(tile); set_layer2(tile);
    ptr_5D4594_2650668[1][3].field_0 = 1;
    header(2, 3, 3, 3);
    for (int y = 3; y <= 5; ++y) for (int x = 2; x <= 4; ++x) {
        unsigned selected = (x == 3 && y == 3) || (x == 2 && y == 4); pack(selected, 1);
        if (x == 3 && y == 3) layer1();
        if (x == 2 && y == 4) layer2();
    }
    obj_5D4594_2650668_t snapshot[CAP][CAP]; for (int x = 0; x < CAP; ++x) memcpy(snapshot[x], rows[x]->row, sizeof(snapshot[x]));
    need(nox_xxx_tile_422C10(3, box->region) == 1, "actual selected-region writer result"); check_layers(tile);
    for (int x = 0; x < CAP; ++x) need(memcmp(snapshot[x], rows[x]->row, sizeof(snapshot[x])) == 0, "selected writer preserves grid");
    need(geometry_calls == 3 && node_count == 3, "selected/rejected geometry callbacks"); finish(before);
}
static void test_region_read(void) {
    setup("region-read"); struct region_box before = *box;
    header(99, -99, 2, 2); pack(0, 1); pack(1, 1); layer1(); pack(1, 1); layer2(); pack(0, 1);
    memcpy(wire, expected, expected_size); wire_size = expected_size; readonly_mode = 1;
    need(nox_xxx_tile_422C10(3, box->region) == 1, "actual selected-region reader result");
    // v3's legacy row mapping is retained, including (row + 23) / 46.
    check_layers(&ptr_5D4594_2650668[1][0]);
    for (int x = 0; x < CAP; ++x) for (int y = 0; y < CAP; ++y) if (x != 1 || y != 0) need(ptr_5D4594_2650668[x][y].field_0 == 0, "region reader target mapping");
    need(geometry_calls == 0 && node_count == 3, "region reader node/geometry counts"); finish(before);
}
static void test_empty(void) {
    setup("empty-write"); struct region_box before = *box;
    need(nox_xxx_tile_422C10(3, NULL) == 1, "empty writer result"); need(node_count == 0 && geometry_calls == 0, "empty writer callbacks"); finish(before);
}
static void test_edge(int read) {
    setup(read ? "edge-read" : "edge-write"); struct region_box before = *box;
    header(127, 127, 1, 1); pack(1, 1); fields(-11, 22, -33, 44); pack(0, 1);
    obj_5D4594_2650668_t *tile = &ptr_5D4594_2650668[127][127];
    if (read) { readonly_mode = 1; memcpy(wire, expected, expected_size); wire_size = expected_size; }
    else { tile->field_0 = 1; tile->field_1 = -11; tile->field_2 = 22; tile->field_3 = -33; tile->field_4 = 44; }
    need(nox_xxx_tile_422C10(3, NULL) == 1, "actual edge codec result");
    need(tile->field_0 == 1 && tile->field_1 == -11 && tile->field_2 == 22 && tile->field_3 == -33 && tile->field_4 == 44 && !tile->field_5, "last native tile fields/head");
    need(node_count == 0 && geometry_calls == 0, "edge codec callbacks"); finish(before);
}
int main(int argc, char **argv) {
    if (argc != 2) return 64;
    if (!strcmp(argv[1], "nil-region-write")) test_grid(0);
    else if (!strcmp(argv[1], "geometry-region-write")) test_region_write();
    else if (!strcmp(argv[1], "nil-region-read")) test_grid(1);
    else if (!strcmp(argv[1], "region-read")) test_region_read();
    else if (!strcmp(argv[1], "edge-write")) test_edge(0);
    else if (!strcmp(argv[1], "edge-read")) test_edge(1);
    else if (!strcmp(argv[1], "empty-write")) test_empty();
    else if (!strcmp(argv[1], "all")) { test_grid(0); test_region_write(); test_grid(1); test_region_read(); test_edge(0); test_edge(1); test_empty(); }
    else return 64;
    return 0;
}
