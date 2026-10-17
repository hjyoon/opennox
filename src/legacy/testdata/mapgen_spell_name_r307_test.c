#include <ctype.h>
#include <limits.h>
#include <stdint.h>
#include <stdio.h>
#include <stdlib.h>
#include <string.h>
#include "noxstring.h"

#pragma GCC diagnostic push
#pragma GCC diagnostic ignored "-Wsign-compare"
#ifdef __clang__
#pragma GCC diagnostic ignored "-Wincompatible-pointer-types-discards-qualifiers"
#else
#pragma GCC diagnostic ignored "-Wdiscarded-qualifiers"
#endif
#include "noxstring.c"
#pragma GCC diagnostic pop

/* The exercised formatter has only %s; integer formatting is not substituted. */
char* nox_itoa(int value, char* dest, int radix) {
    (void)value; (void)dest; (void)radix; abort();
}

static char expected[60];
static int expected_result, trace, lookup_calls, capacity_diagnostic;
static unsigned cases;
static char* formatted;
#define REQUIRE(x) do { if (!(x)) { \
    fprintf(stderr, "spell-name case %u line %d: %s\n", cases, __LINE__, #x); exit(42); \
} } while (0)

static int observed_format(char* dest, size_t capacity, const char* format, const char* input) {
    REQUIRE(trace == 0);
    REQUIRE(strcmp(format, "SPELL_%s") == 0);
    REQUIRE(strcmp(input, expected + 6) == 0);
    trace = 1;
    /* Diagnostic mode reports the actual compiler-known callee object before
       any out-of-bounds write. Sanitized mode executes the real formatter. */
    if (capacity_diagnostic && capacity != 60) {
        fprintf(stderr, "spell-name actual output capacity %zu, expected 60\n", capacity);
        exit(41);
    }
    if (capacity == 60) memset(dest, 0xCD, 60);
    formatted = dest;
    int result = nox_sprintf(dest, format, input);
    REQUIRE(result == (int)strlen(expected));
    REQUIRE(strcmp(dest, expected) == 0);
    if (capacity == 60) {
        for (size_t i = strlen(expected) + 1; i < 60; i++) {
            REQUIRE((unsigned char)dest[i] == 0xCD);
        }
    }
    trace = 2;
    return result;
}

int nox_xxx_spellNameToN_4243F0(const char* name) {
    REQUIRE(trace == 2);
    REQUIRE(name == formatted);
    REQUIRE(strcmp(name, expected) == 0);
    trace = 3;
    lookup_calls++;
    return expected_result;
}

#define nox_sprintf(dest, format, input) \
    observed_format((dest), __builtin_object_size((dest), 0), (format), (input))
// PRODUCTION_SPELL_NAME_51E1D0
#undef nox_sprintf

static void one(const char* input, const char* upper, int result) {
    struct { unsigned char before[16]; char input[60]; unsigned char after[16]; } guard;
    unsigned char snapshot[sizeof(guard)];
    REQUIRE(strlen(input) <= 53 && strlen(input) == strlen(upper));
    memset(&guard, 0xA5, sizeof(guard));
    memcpy(guard.input, input, strlen(input) + 1);
    memcpy(snapshot, &guard, sizeof(guard));
    memcpy(expected, "SPELL_", 6);
    memcpy(expected + 6, upper, strlen(upper) + 1);
    expected_result = result; trace = 0; lookup_calls = 0; formatted = NULL;
    REQUIRE(nox_xxx_mapGenSpellIdByName_51E1D0(guard.input) == result);
    REQUIRE(trace == 3 && lookup_calls == 1);
    REQUIRE(memcmp(&guard, snapshot, sizeof(guard)) == 0);
    cases++;
}

int main(int argc, char** argv) {
    REQUIRE(argc == 2);
    capacity_diagnostic = strcmp(argv[1], "capacity") == 0;
    REQUIRE(capacity_diagnostic || strcmp(argv[1], "native") == 0);
    const int results[] = {0, 137, -137, INT_MAX, INT_MIN};
    const char* names[][2] = {{"fireball", "FIREBALL"}, {"Magic_Missile", "MAGIC_MISSILE"},
        {"SPELL_1", "SPELL_1"}, {"a b-c.d!", "A B-C.D!"}, {"z", "Z"},
        {"%s", "%S"}, {"123_", "123_"}, {"MiXeD", "MIXED"}};
    for (size_t r = 0; r < sizeof(results)/sizeof(results[0]); r++) {
        for (unsigned length = 0; length <= 53; length++) {
            char input[54], upper[54];
            const char lower_pattern[] = "abZ_09-";
            const char upper_pattern[] = "ABZ_09-";
            for (unsigned i = 0; i < length; i++) {
                input[i] = lower_pattern[i % 7]; upper[i] = upper_pattern[i % 7];
            }
            input[length] = 0; upper[length] = 0;
            one(input, upper, results[r]);
        }
        for (size_t n = 0; n < sizeof(names)/sizeof(names[0]); n++) {
            one(names[n][0], names[n][1], results[r]);
        }
    }
    REQUIRE(cases == 310);
    printf("Spell-name 51E1D0: %u bounded cases; real formatter, lookup identity/order, input canaries, output tail and signed returns PASS\n", cases);
    return 0;
}
