// Original GAME.EXE 0x428170..0x4281eb: signed JGE at 0x428180/0x42819d,
// lower clamps at 0x4281b4/0x4281c1, signed upper JL at 0x4281d9/0x4281e4.
// Expected values are literal original-derived outputs, not current-source
// calculations. Inverted outputs are intentional: the original clamps only
// minima below zero and maxima above 5887, not all four bounds independently.
#include <stdint.h>
#include <stdio.h>
#include <stdlib.h>
#include <string.h>
#include "GAME1_2.h"

// PRODUCTION_REGION_BOUNDS_BODY

_Static_assert(sizeof(int) == sizeof(int32_t), "original coordinate DWORD width");
_Static_assert(sizeof(int4) == 16, "original four signed coordinate fields");
struct bounds_case {
    const char *name;
    int32_t x0, y0, x1, y1;
    int32_t expected[4];
};
static const struct bounds_case cases[] = {
    {"positive-ordered", 46, 69, 92, 115, {46, 69, 92, 115}},
    {"positive-reversed", 92, 115, 46, 69, {46, 69, 92, 115}},
    {"negative-x-mixed", -23, 69, 92, 115, {0, 69, 92, 115}},
    {"negative-y-mixed", 46, -23, 92, 69, {46, 0, 92, 69}},
    {"negative-both-mixed", -23, -46, 92, 115, {0, 0, 92, 115}},
    {"negative-mixed-reversed", 92, 115, -23, -46, {0, 0, 92, 115}},
    {"entirely-negative", -92, -115, -46, -69, {0, 0, -46, -69}},
    {"negative-reversed", -46, -69, -92, -115, {0, 0, -46, -69}},
    {"equal-zero", 0, 0, 0, 0, {0, 0, 0, 0}},
    {"equal-negative", -23, -46, -23, -46, {0, 0, -23, -46}},
    {"upper-boundary", 0, 0, 5888, 5888, {0, 0, 5887, 5887}},
    {"upper-last-valid", 0, 0, 5887, 5887, {0, 0, 5887, 5887}},
    {"entirely-above-upper", 6000, 7000, 8000, 9000, {6000, 7000, 5887, 5887}},
    {"int-min-max", INT32_MIN, INT32_MIN, INT32_MAX, INT32_MAX, {0, 0, 5887, 5887}},
    {"int-max-min", INT32_MAX, INT32_MAX, INT32_MIN, INT32_MIN, {0, 0, 5887, 5887}},
    {"equal-int-min", INT32_MIN, INT32_MIN, INT32_MIN, INT32_MIN, {0, 0, INT32_MIN, INT32_MIN}},
    {"equal-int-max", INT32_MAX, INT32_MAX, INT32_MAX, INT32_MAX, {INT32_MAX, INT32_MAX, 5887, 5887}},
    {"equal-above-upper", 5888, 5888, 5888, 5888, {5888, 5888, 5887, 5887}},
    {"int-min-and-upper", INT32_MIN, 69, 92, INT32_MAX, {0, 69, 92, 5887}},
    {"opposite-extreme-corners", INT32_MAX, INT32_MIN, INT32_MIN, INT32_MAX, {0, 0, 5887, 5887}},
};
struct input_box { uint32_t prefix[4]; int32_t region[8]; uint32_t suffix[4]; };
struct output_box { uint32_t prefix[4]; int4 bounds; uint32_t suffix[4]; };
static int test_case(const struct bounds_case *c) {
    struct input_box *input = malloc(sizeof(*input));
    struct output_box *output = malloc(sizeof(*output));
    if (!input || !output) { fprintf(stderr, "R303 FAIL allocation\n"); exit(72); }
    memset(input, 0xa7, sizeof(*input)); memset(output, 0xa7, sizeof(*output));
    input->region[0] = INT32_MIN; input->region[1] = c->y0;
    input->region[2] = c->x0; input->region[3] = INT32_MAX;
    input->region[4] = c->x1; input->region[5] = -7654321;
    input->region[6] = 0x13579bdf; input->region[7] = c->y1;
    struct input_box before = *input;
    struct output_box before_output = *output;
    if (sizeof(void*) == 8 && ((uintptr_t)input->region <= UINT32_MAX || (uintptr_t)&output->bounds <= UINT32_MAX)) {
        fprintf(stderr, "R303 FAIL native addresses case=%s\n", c->name); exit(72);
    }
    // Bind through the public native-pointer declaration; signedness belongs
    // to the descriptor's DWORD fields, not to the pointer-sized ABI.
    int (*entry)(void*, int4*) = sub_428170;
    int result = entry(input->region, &output->bounds);
    int32_t actual[4] = {output->bounds.field_0, output->bounds.field_4, output->bounds.field_8, output->bounds.field_C};
    int preserved = memcmp(input, &before, sizeof(*input)) == 0 &&
        memcmp(output->prefix, before_output.prefix, sizeof(output->prefix)) == 0 &&
        memcmp(output->suffix, before_output.suffix, sizeof(output->suffix)) == 0;
    int passed = result == 0 && preserved && memcmp(actual, c->expected, sizeof(actual)) == 0;
    printf("R303 %s case=%s actual=[%d,%d,%d,%d] expected=[%d,%d,%d,%d] input_canaries_unchanged=%s\n",
        passed ? "PASS" : "FAIL", c->name, actual[0], actual[1], actual[2], actual[3],
        c->expected[0], c->expected[1], c->expected[2], c->expected[3], preserved ? "true" : "false");
    free(input); free(output); return passed;
}
int main(int argc, char **argv) {
    if (argc != 2) return 64;
    const size_t count = sizeof(cases) / sizeof(cases[0]);
    if (!strcmp(argv[1], "--list")) {
        for (size_t i = 0; i < count; ++i) puts(cases[i].name);
        return 0;
    }
    int selected = 0, failed = 0;
    for (size_t i = 0; i < count; ++i) {
        if (strcmp(argv[1], "all") && strcmp(argv[1], cases[i].name)) continue;
        ++selected; failed += !test_case(&cases[i]);
    }
    if (!selected) return 64;
    printf("R303 SUMMARY selected=%d passed=%d failed=%d\n", selected, selected - failed, failed);
    return failed != 0;
}
