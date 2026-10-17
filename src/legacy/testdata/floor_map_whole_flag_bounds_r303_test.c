// Complete production v3/v4 codecs are source-bound below. Controlled file I/O
// exposes exact bytes, scalar widths and callback order; rows and lists are
// native allocations above 4GiB. The explicit low-layer removal is a fixture
// state transition with the editor's BYTE &0xFE mask, not editor execution.
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

// PRODUCTION_FLOOR_WHOLE_FLAG_BODIES

enum { CAP = 128, MAX_BYTES = 4096, MAX_CALLS = 256, MAX_NODES = 8 };
struct stream { unsigned char data[MAX_BYTES]; size_t size, calls, widths[MAX_CALLS]; };
struct row_box { unsigned char prefix[16]; obj_5D4594_2650668_t row[CAP]; unsigned char suffix[16]; };
struct node_box { unsigned char prefix[16]; nox_tile_list_node_t node; unsigned char suffix[16]; };
static struct row_box *rows[CAP];
static struct node_box *nodes[MAX_NODES];
static struct stream wire, expected;
static size_t cursor, calls, widths[MAX_CALLS], node_count;
static int readonly_mode;
static const char *case_name;

static void need(int ok, const char *why) {
    if (!ok) {
        fprintf(stderr, "R303 FLAGS FAIL case=%s reason=%s cursor=%zu expected_bytes=%zu calls=%zu expected_calls=%zu\n", case_name, why, cursor, expected.size, calls, expected.calls);
        exit(73);
    }
}
int nox_crypt_IsReadOnly(void) { return readonly_mode; }
bool nox_common_gameFlags_check_40A5C0(unsigned int flags) { (void)flags; need(0, "nil-region flags query"); return false; }
int nox_xxx_wallMath_427F30(int2 *point, int *region) { (void)point; (void)region; need(0, "nil-region geometry call"); return 0; }
nox_tile_coord_entry_t* nox_xxx_tileAllocTileInCoordList_5040A0(int x, int y, float kind) {
    (void)x; (void)y; (void)kind; need(0, "nil-region coordinate entry"); return NULL;
}
size_t nox_xxx_fileReadWrite_426AC0_file3_fread_impl(unsigned char *p, size_t n, char *file, int line) {
    (void)file; (void)line;
    need(calls < MAX_CALLS && cursor <= MAX_BYTES && n <= MAX_BYTES - cursor, "I/O bounds");
    widths[calls++] = n;
    if (readonly_mode) { need(cursor <= wire.size && n <= wire.size - cursor, "read past declared v3 input"); memcpy(p, wire.data + cursor, n); }
    else { memcpy(wire.data + cursor, p, n); wire.size = cursor + n; }
    cursor += n; return n;
}
nox_tile_list_node_t* nox_xxx_tileListAddNewSubtile_422160(int a, int b, int c, int d) {
    need(node_count < MAX_NODES, "node bound");
    struct node_box *node = malloc(sizeof(*node)); need(node != NULL, "node allocation"); memset(node, 0xa7, sizeof(*node));
    node->node = (nox_tile_list_node_t){a, b, c, d, NULL};
    need((uintptr_t)&node->node > UINT32_MAX, "native node above 4GiB"); nodes[node_count++] = node; return &node->node;
}
static void pack(unsigned value, size_t width) {
    need(expected.calls < MAX_CALLS && expected.size + width <= MAX_BYTES, "expected bound");
    expected.widths[expected.calls++] = width;
    for (size_t i = 0; i < width; ++i) expected.data[expected.size++] = (unsigned char)(value >> (8 * i));
}
static void fields(unsigned a, int b, unsigned c, unsigned d) {
    size_t start = expected.calls; pack(a, 4); pack((unsigned)b, 4); pack(c, 4); pack(d, 4);
    expected.calls = start; expected.widths[expected.calls++] = 16;
}
static void header(unsigned version, unsigned x, unsigned y, unsigned w, unsigned h) {
    pack(version, 2); pack(x, 4); pack(y, 4); pack(w, 4); pack(h, 4);
}
static void layer_v3(int layer) {
    if (layer == 1) {
        fields(0xa1, -1234, 0xb2, 0xc3); pack(2, 1);
        fields(0xd1, -42, 0xe1, 0xf1); fields(0xd2, -84, 0xe2, 0xf2);
    } else {
        fields(0xa2, -2222, 0xb3, 0xc4); pack(1, 1); fields(0xd3, -126, 0xe3, 0xf3);
    }
}
static void layer_v4(int layer) {
    if (layer == 1) {
        pack(0xa1, 1); pack((unsigned)-1234, 2); pack(0xb2, 1); pack(0xc3, 1); pack(2, 1);
        pack(0xd1, 1); pack((unsigned)-42, 2); pack(0xe1, 1); pack(0xf1, 1);
        pack(0xd2, 1); pack((unsigned)-84, 2); pack(0xe2, 1); pack(0xf2, 1);
    } else {
        pack(0xa2, 1); pack((unsigned)-2222, 2); pack(0xb3, 1); pack(0xc4, 1); pack(1, 1);
        pack(0xd3, 1); pack((unsigned)-126, 2); pack(0xe3, 1); pack(0xf3, 1);
    }
}
static void assert_stream(const char *why) {
    need(cursor == expected.size && wire.size == expected.size && memcmp(wire.data, expected.data, expected.size) == 0, why);
    need(calls == expected.calls && memcmp(widths, expected.widths, calls * sizeof(*widths)) == 0, "exact scalar widths and I/O order");
}
static void run_case(const char *name) {
    case_name = name; cursor = calls = node_count = 0; readonly_mode = 1; wire = (struct stream){0}; expected = (struct stream){0};
    ptr_5D4594_2650668_cap = CAP; ptr_5D4594_2650668 = calloc(CAP, sizeof(*ptr_5D4594_2650668)); need(ptr_5D4594_2650668 != NULL, "row pointer allocation");
    for (int x = 0; x < CAP; ++x) {
        rows[x] = malloc(sizeof(*rows[x])); need(rows[x] != NULL, "row allocation"); memset(rows[x], 0xa7, sizeof(*rows[x])); memset(rows[x]->row, 0, sizeof(rows[x]->row));
        ptr_5D4594_2650668[x] = rows[x]->row; need((uintptr_t)rows[x]->row > UINT32_MAX, "native row above 4GiB");
    }
    unsigned flag = 4, x0 = 5, y0 = 7, w = 1, h = 1, layer_x = x0, layer_y = y0;
    int bounds_case = !strcmp(name, "high-and-layer-05") || !strcmp(name, "high-and-two-layers-07");
    int clear_layer = !strcmp(name, "layer-clear-high-only");
    uint32_t adjacent = !strcmp(name, "adjacent-byte-0100") ? 0x100u : !strcmp(name, "adjacent-byte-80000000") ? 0x80000000u : 0;
    int empty = !strcmp(name, "empty-control") || adjacent != 0;
    if (!strcmp(name, "high-only-80")) flag = 0x80;
    else if (!strcmp(name, "high-only-grid-edge")) { x0 = y0 = layer_x = layer_y = 127; flag = 0xfc; }
    else if (!strcmp(name, "low-layer-control")) flag = 3;
    else if (clear_layer || bounds_case) flag = !strcmp(name, "high-and-two-layers-07") ? 7 : 5;
    else if (empty) flag = 0;
    if (bounds_case) { x0 = 1; y0 = 2; w = h = 7; layer_x = 4; layer_y = 5; }
    if (adjacent) memcpy(&ptr_5D4594_2650668[layer_x][layer_y].field_0, &adjacent, sizeof(adjacent));
    header(3, x0, y0, w, h);
    for (unsigned y = y0; y < y0 + h; ++y) for (unsigned x = x0; x < x0 + w; ++x) {
        unsigned f = !bounds_case || (x == layer_x && y == layer_y) ? flag : ((x == x0 && y == y0) || (x == x0 + w - 1 && y == y0 + h - 1) ? 4 : 0);
        pack(f, 1); if (f & 1) layer_v3(1); if (f & 2) layer_v3(2);
    }
    wire = expected;
    need(nox_server_mapRWFloorMap_422230(NULL) == 1, "actual v3 entry/delegate result"); assert_stream("exact v3 input consumed");
    obj_5D4594_2650668_t *tile = &ptr_5D4594_2650668[layer_x][layer_y];
    need(tile->field_0 == flag, "raw v3 flags BYTE retained");
    if (adjacent) { uint32_t actual; memcpy(&actual, &tile->field_0, sizeof(actual)); need(actual == adjacent, "v3 flag clear preserves adjacent three bytes"); }
    if (bounds_case) need(ptr_5D4594_2650668[x0][y0].field_0 == 4 && ptr_5D4594_2650668[x0+w-1][y0+h-1].field_0 == 4, "raw v3 high-only extrema");
    if (flag & 1) need(tile->field_1 == 0xa1 && tile->field_2 == -1234 && tile->field_3 == 0xb2 && tile->field_4 == 0xc3 && tile->field_5 == &nodes[0]->node && tile->field_5->next == &nodes[1]->node && !nodes[1]->node.next, "v3 native first-layer fields and links");
    if (flag & 2) need(tile->field_6 == 0xa2 && tile->field_7 == -2222 && tile->field_8 == 0xb3 && tile->field_9 == 0xc4 && tile->field_10 == &nodes[2]->node && !nodes[2]->node.next, "v3 native second-layer fields and links");
    if (clear_layer) { tile->field_0 = (uint8_t)(tile->field_0 & 0xfe); need(tile->field_0 == 4, "layer removal preserves high BYTE bits"); }
    obj_5D4594_2650668_t *snapshot = malloc(CAP * CAP * sizeof(*snapshot)); need(snapshot != NULL, "snapshot allocation");
    for (int x = 0; x < CAP; ++x) memcpy(snapshot + x * CAP, rows[x]->row, sizeof(rows[x]->row));
    struct node_box node_snapshot[MAX_NODES]; for (size_t n = 0; n < node_count; ++n) node_snapshot[n] = *nodes[n];
    cursor = calls = 0; readonly_mode = 0; wire = (struct stream){0}; expected = (struct stream){0};
    if (empty) pack(4, 2);
    else {
        header(4, x0, y0, w, h);
        unsigned active = clear_layer ? 0 : flag & 3;
        if (active) {
            pack((layer_x << 8) | layer_y | (active & 1 ? 0x8000 : 0) | (active & 2 ? 0x80 : 0), 2);
            if (active & 1) layer_v4(1); if (active & 2) layer_v4(2);
        }
        pack(0xffff, 2);
    }
    need(nox_server_mapRWFloorMap_422230(NULL) == 1, "actual v4 writer result"); assert_stream("exact v4 bounds/header/payload/sentinel bytes");
    for (int x = 0; x < CAP; ++x) {
        need(memcmp(snapshot + x * CAP, rows[x]->row, sizeof(rows[x]->row)) == 0, "writer preserves all native cells");
        for (int b = 0; b < 16; ++b) need(rows[x]->prefix[b] == 0xa7 && rows[x]->suffix[b] == 0xa7, "row canaries");
    }
    for (size_t n = 0; n < node_count; ++n) need(memcmp(&node_snapshot[n], nodes[n], sizeof(*nodes[n])) == 0, "node fields/links/canaries unchanged");
    printf("R303 FLAGS RESULT case=%s passed=true v3_flags=0x%02x post_clear_flags=0x%02x bytes=%zu callbacks=%zu nodes=%zu pointer_bytes=%zu\n", name, flag, tile->field_0, cursor, calls, node_count, sizeof(void*));
    free(snapshot); for (size_t n = 0; n < node_count; ++n) free(nodes[n]);
    for (int x = 0; x < CAP; ++x) free(rows[x]); free(ptr_5D4594_2650668);
}
int main(int argc, char **argv) {
    static const char *cases[] = {"high-only-04", "high-only-80", "high-only-grid-edge", "high-and-layer-05", "high-and-two-layers-07", "layer-clear-high-only", "low-layer-control", "empty-control", "adjacent-byte-0100", "adjacent-byte-80000000"};
    if (argc != 2) return 64;
    int selected = 0;
    for (size_t i = 0; i < sizeof(cases)/sizeof(*cases); ++i) if (!strcmp(argv[1], cases[i]) || !strcmp(argv[1], "all")) { run_case(cases[i]); selected++; }
    return selected ? 0 : 64;
}
