#include <inttypes.h>
#include <stdint.h>
#include <stdio.h>
#include <stdlib.h>
#include <string.h>

float* sub_5245A0(int, float*, int, int);
float* sub_524610(int, float*, int);
float* sub_51D5E0(float*);
_Static_assert(sizeof(float) == 4, "original binary32 fields");

static uint32_t observed[8][2];
static size_t calls;
static float acknowledgement[2];

static void need(int ok, const char* message) {
    if (!ok) {
        fprintf(stderr, "rectangle point contract: %s\n", message);
        exit(86);
    }
}

// PRODUCTION_RECTANGLE_POINT_BODIES

float* sub_51D5E0(float* point) {
    need(calls < 8, "bounded callback count");
    memcpy(observed[calls++], point, 2 * sizeof(float));
    return acknowledgement;
}

int main(int argc, char** argv) {
    need(argc == 2, "named case");
    struct {
        uint32_t prefix;
        float point[2];
        uint32_t suffix;
    } input = {0x12345678, {100.25f, -50.75f}, 0x9abcdef0}, before;
    static const uint32_t rectangle[][2] = {
        {0x42e906e4, 0xc209f238},
        {0x43150a56, 0xc209f238},
        {0x42e906e4, 0xbffad520},
        {0x43150a56, 0xbffad520},
    };
    static const uint32_t column[][2] = {
        {0x41d21b8f, 0x42110dc8},
        {0x41d21b8f, 0x428994ac},
        {0x41d21b8f, 0x42caa274},
    };
    const uint32_t (*expected)[2] = NULL;
    size_t count = 0;
    float* result;
    float* want_result;
    const char* name = argv[1];
    if (!strcmp(name, "column-three") || !strcmp(name, "column-zero") || !strcmp(name, "column-negative")) {
        input.point[0] = 10.0f;
        input.point[1] = 20.0f;
        before = input;
        int height = !strcmp(name, "column-three") ? 3 : !strcmp(name, "column-zero") ? 0 : -3;
        result = sub_524610(0, input.point, height);
        if (height > 0) {
            count = 3;
            expected = column;
            want_result = acknowledgement;
        } else want_result = input.point;
    } else {
        int width = 2, height = 2;
        if (!strcmp(name, "rectangle-two-by-two")) {
            count = 4;
            expected = rectangle;
        } else if (!strcmp(name, "rectangle-zero-width")) width = 0;
        else if (!strcmp(name, "rectangle-zero-height")) height = 0;
        else if (!strcmp(name, "rectangle-negative-width")) width = -2;
        else if (!strcmp(name, "rectangle-negative-height")) height = -2;
        else need(0, "known case");
        before = input;
        result = sub_5245A0(0, input.point, width, height);
        want_result = count ? acknowledgement : (float*)(intptr_t)height;
    }
    need(calls == count, "exact callback count");
    for (size_t i = 0; i < count; ++i) {
        if (memcmp(observed[i], expected[i], sizeof(observed[i]))) {
            fprintf(stderr, "point[%zu] got %08" PRIx32 "/%08" PRIx32 " want %08" PRIx32 "/%08" PRIx32 "\n",
                i, observed[i][0], observed[i][1], expected[i][0], expected[i][1]);
            need(0, "both binary32 coordinates and row/column order");
        }
    }
    need(result == want_result, "retained C return propagation");
    need(!memcmp(&input, &before, sizeof(input)), "input and adjacent bytes preserved");
    printf("PASS %s: %zu exact point callbacks\n", name, calls);
    return 0;
}
