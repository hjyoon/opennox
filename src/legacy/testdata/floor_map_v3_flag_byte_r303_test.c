// Preserve the original BYTE flag contract beside poisoned reserved bytes.
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

// PRODUCTION_V3_FLAG_BYTE_BODIES

_Static_assert(sizeof(((obj_5D4594_2650668_t*)0)->field_0) == 1, "tile flag is one BYTE");
_Static_assert(offsetof(obj_5D4594_2650668_t, field_0_1) == 1, "first reserved flag neighbor");
_Static_assert(offsetof(obj_5D4594_2650668_t, field_0_2) == 2, "remaining reserved neighbors");
enum { CAP = 128, X = 64, Y = 32 };
struct row_box { unsigned char prefix[16]; obj_5D4594_2650668_t row[CAP]; unsigned char suffix[16]; };
static struct row_box* rows[CAP];
static unsigned char wire[256], expected[256];
static size_t cursor, wire_size, expected_size, widths[32], expected_widths[32], calls, expected_calls;
static int reading;
static const char* case_name;
static void need(int ok, const char* why) {
    if (!ok) { fprintf(stderr, "R303 BYTE FAIL case=%s reason=%s cursor=%zu calls=%zu\n", case_name, why, cursor, calls); exit(71); }
}
int nox_crypt_IsReadOnly(void) { return reading; }
bool nox_common_gameFlags_check_40A5C0(unsigned int flags) { (void)flags; need(0, "unexpected region flags callback"); return false; }
nox_tile_coord_entry_t* nox_xxx_tileAllocTileInCoordList_5040A0(int x, int y, float kind) { (void)x; (void)y; (void)kind; need(0, "unexpected region allocation"); return NULL; }
nox_tile_list_node_t* nox_xxx_tileListAddNewSubtile_422160(int a, int b, int c, int d) { (void)a; (void)b; (void)c; (void)d; need(0, "unexpected node allocation"); return NULL; }
int nox_xxx_wallMath_427F30(int2* point, int* region) { (void)point; (void)region; need(0, "unexpected geometry callback"); return 0; }
size_t nox_xxx_fileReadWrite_426AC0_file3_fread_impl(unsigned char* p, size_t n, char* file, int line) {
    (void)file; (void)line;
    need(calls < 32 && cursor <= sizeof(wire) && n <= sizeof(wire) - cursor, "I/O bounds"); widths[calls++] = n;
    if (reading) { need(cursor <= wire_size && n <= wire_size - cursor, "read bounds"); memcpy(p, wire + cursor, n); }
    else { memcpy(wire + cursor, p, n); wire_size = cursor + n; }
    cursor += n; return n;
}
static void pack(unsigned value, size_t n) {
    need(expected_calls < 32 && expected_size + n <= sizeof(expected), "expected I/O bounds");
    expected_widths[expected_calls++] = n;
    for (size_t i = 0; i < n; ++i) expected[expected_size++] = (unsigned char)(value >> (8 * i));
}
static void fields(int a, int b, int c, int d) {
    size_t start = expected_calls; pack((unsigned)a, 4); pack((unsigned)b, 4); pack((unsigned)c, 4); pack((unsigned)d, 4);
    expected_calls = start; expected_widths[expected_calls++] = 16;
}
static void setup(const char* name) {
    case_name = name; cursor = wire_size = expected_size = calls = expected_calls = 0; reading = 0;
    memset(wire, 0, sizeof(wire)); memset(expected, 0, sizeof(expected));
    ptr_5D4594_2650668_cap = CAP; ptr_5D4594_2650668 = calloc(CAP, sizeof(*ptr_5D4594_2650668)); need(ptr_5D4594_2650668 != NULL, "grid allocation");
    for (int x = 0; x < CAP; ++x) {
        rows[x] = malloc(sizeof(*rows[x])); need(rows[x] != NULL, "row allocation"); memset(rows[x], 0xa7, sizeof(*rows[x])); memset(rows[x]->row, 0, sizeof(rows[x]->row));
        ptr_5D4594_2650668[x] = rows[x]->row;
        need(sizeof(void*) != 8 || (uintptr_t)rows[x]->row > UINT32_MAX, "native row address");
        for (int y = 0; y < CAP; ++y) { rows[x]->row[y].field_0_1 = 0x91; rows[x]->row[y].field_0_2 = 0xb2c3; }
    }
}
static void finish(int expected_flag) {
    need(cursor == expected_size && wire_size == expected_size && memcmp(wire, expected, expected_size) == 0, "exact original BYTE/scalar wire");
    need(calls == expected_calls && memcmp(widths, expected_widths, calls * sizeof(widths[0])) == 0, "exact callback widths/order");
    for (int x = 0; x < CAP; ++x) {
        for (int i = 0; i < 16; ++i) need(rows[x]->prefix[i] == 0xa7 && rows[x]->suffix[i] == 0xa7, "row allocation canary");
        for (int y = 0; y < CAP; ++y) {
            obj_5D4594_2650668_t* tile = &rows[x]->row[y];
            need(tile->field_0_1 == 0x91 && tile->field_0_2 == 0xb2c3, "upper three flag neighbors preserved");
            need(tile->field_0 == ((x == X && y == Y) ? expected_flag : 0), "only one BYTE cleared/read");
        }
        free(rows[x]);
    }
    free(ptr_5D4594_2650668);
    printf("R303 BYTE PASS case=%s bytes=%zu callbacks=%zu reserved_bytes_unchanged=true\n", case_name, cursor, calls);
}
static void test_empty(void) {
    setup("empty-upper-bytes"); need(nox_xxx_tile_422C10(3, NULL) == 1, "empty writer result"); finish(0);
}
static void test_bounds(void) {
    setup("bounds-upper-bytes"); obj_5D4594_2650668_t* tile = &rows[X]->row[Y];
    tile->field_0 = 3; tile->field_1 = 11; tile->field_2 = -22; tile->field_3 = 33; tile->field_4 = -44;
    tile->field_6 = 55; tile->field_7 = -66; tile->field_8 = 77; tile->field_9 = -88;
    pack(X, 4); pack(Y, 4); pack(1, 4); pack(1, 4); pack(3, 1); fields(11, -22, 33, -44); pack(0, 1); fields(55, -66, 77, -88); pack(0, 1);
    need(nox_xxx_tile_422C10(3, NULL) == 1, "bounded writer result"); finish(3);
}
static void test_read(void) {
    setup("read-upper-bytes"); reading = 1;
    for (int x = 0; x < CAP; ++x) for (int y = 0; y < CAP; ++y) rows[x]->row[y].field_0 = 0xab;
    pack(X, 4); pack(Y, 4); pack(1, 4); pack(1, 4); pack(1, 1); fields(11, -22, 33, -44); pack(0, 1);
    memcpy(wire, expected, expected_size); wire_size = expected_size;
    need(nox_xxx_tile_422C10(3, NULL) == 1, "reader result");
    obj_5D4594_2650668_t* tile = &rows[X]->row[Y];
    need(tile->field_1 == 11 && tile->field_2 == -22 && tile->field_3 == 33 && tile->field_4 == -44 && !tile->field_5, "reader first-layer scalar/head fields"); finish(1);
}
int main(int argc, char** argv) {
    if (argc != 2) return 64;
    if (!strcmp(argv[1], "empty-upper-bytes")) test_empty();
    else if (!strcmp(argv[1], "bounds-upper-bytes")) test_bounds();
    else if (!strcmp(argv[1], "read-upper-bytes")) test_read();
    else return 64;
    return 0;
}
