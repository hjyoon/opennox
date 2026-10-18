#include <assert.h>
#include <stdio.h>
#include "GAME1.h"
#include "GAME4_2.h"
#include "client__gui__guigen.h"
#include "input.h"
#include "operators.h"

static uint32_t current_tick, game_flags;
static int rendering_disabled;
static unsigned engine_queries, game_queries, tick_reads, input_polls, displays;
static uint32_t threshold = 250;
static uint32_t expected_sequence_at_display, expected_previous_at_display;
static int mutate_poll, mutate_display;
static struct {
    uint32_t before, value, after;
} sequence = {0x13579BDF, 0, 0x2468ACE0};
static unsigned char packet[3];
uint32_t dword_5d4594_2487564, dword_5d4594_2487568;

bool nox_common_getEngineFlag(const nox_engine_flag flags) {
    assert(flags == NOX_ENGINE_FLAG_DISABLE_GRAPHICS_RENDERING);
    ++engine_queries;
    return rendering_disabled != 0;
}
bool nox_common_gameFlags_check_40A5C0(unsigned int flags) {
    assert(!rendering_disabled && flags == 0x200000);
    ++game_queries;
    return (game_flags & flags) != 0;
}
unsigned int nox_platform_get_ticks(void) {
    assert(!rendering_disabled && !game_flags && !tick_reads && !input_polls && !displays);
    ++tick_reads;
    return current_tick;
}
int nox_input_pollEvents_4453A0(void) {
    assert(!rendering_disabled && tick_reads == 1 && !input_polls && !displays);
    ++input_polls;
    if (mutate_poll) {
        dword_5d4594_2487564 = 400;
        dword_5d4594_2487568 = 40;
    }
    return 0;
}
short nox_xxx_mapGenClientText_4A9D00(unsigned char* value) {
    assert(!rendering_disabled && tick_reads == 1 && input_polls == 1 && !displays);
    memcpy(packet, value, sizeof(packet));
    ++displays;
    // Last-progress time is stored after the display callback.
    assert(dword_5d4594_2487568 == expected_previous_at_display);
    assert(sequence.value == expected_sequence_at_display);
    if (mutate_display) {
        dword_5d4594_2487564 = 777;
        dword_5d4594_2487568 = 888;
        sequence.value = 0xABCDEF12;
    }
    return 0;
}
static uint32_t* progress_sequence(uintptr_t base, uintptr_t offset) {
    assert(base == 0x5D4594 && offset == 2487572 && !rendering_disabled);
    return &sequence.value;
}
static int32_t* progress_threshold(uintptr_t base, uintptr_t offset) {
    assert(base == 0x587000 && offset == 254948 && !rendering_disabled);
    return (int32_t*)&threshold;
}
#undef getMemU16Ptr
#undef getMemU32Ptr
#undef getMemIntPtr
#define getMemU16Ptr(B,O) ((uint16_t*)progress_sequence(B,O))
#define getMemU32Ptr(B,O) progress_sequence(B,O)
#define getMemIntPtr(B,O) progress_threshold(B,O)

#pragma GCC diagnostic push
#pragma GCC diagnostic ignored "-Wsign-compare"
// COMPLETE_MAPGEN_PROGRESS_R314
#pragma GCC diagnostic pop

static void reset(uint32_t now, uint32_t previous, uint32_t number) {
    current_tick = now;
    dword_5d4594_2487564 = 0xAA55AA55;
    dword_5d4594_2487568 = previous;
    sequence.value = number;
    expected_sequence_at_display = number + 1;
    expected_previous_at_display = (int32_t)previous > (int32_t)now ? 0 : previous;
    mutate_poll = mutate_display = 0;
    engine_queries = game_queries = tick_reads = input_polls = displays = 0;
    memset(packet, 0xCC, sizeof(packet));
}
static void canaries(void) {
    assert(sequence.before == 0x13579BDF && sequence.after == 0x2468ACE0);
}
int main(int argc, char** argv) {
    assert(argc == 2);
    if (!strcmp(argv[1], "headless")) {
        rendering_disabled = 1;
        for (unsigned editor = 0; editor < 2; ++editor) {
            game_flags = editor ? 0x200000 : 0;
            reset(5000, 23, UINT32_MAX);
            nox_xxx_mapGenSetFlags_5235F0((char)157);
            assert(engine_queries == 1 && !game_queries && !tick_reads && !input_polls && !displays);
            assert(dword_5d4594_2487564 == 0xAA55AA55 && dword_5d4594_2487568 == 23);
            assert(sequence.value == UINT32_MAX);
            canaries();
        }
    } else if (!strcmp(argv[1], "editor")) {
        game_flags = 0x200000;
        reset(5000, 23, 9);
        nox_xxx_mapGenSetFlags_5235F0((char)155);
        assert(game_queries == 1 && !tick_reads && !input_polls && !displays);
        assert(dword_5d4594_2487564 == 0xAA55AA55 && dword_5d4594_2487568 == 23 && sequence.value == 9);
        canaries();
    } else {
        assert(!strcmp(argv[1], "rendered"));
        const struct { uint32_t now, previous, number; int draw; uint32_t stored; } cases[] = {
            {250, 0, 7, 0, 0},
            {251, 0, 0xABCDFFFF, 1, 251},
            {251, 1, 7, 0, 1},
            {252, 1, UINT32_MAX, 1, 252},
            {10, 5000, 7, 0, 0},
            {251, 5000, 7, 1, 251},
            {0xFFFFFFFF, 0, 7, 0, 0},
            {0x80000010, 0x80000000, 7, 0, 0x80000000},
            {0x80000400, 0x7FFFFFF0, 7, 0, 0},
            {10, 0xFFFFFFF0, 7, 0, 0xFFFFFFF0},
            {1000, 0xFFFFFE0B, 7, 1, 1000},
        };
        for (unsigned i = 0; i < sizeof(cases)/sizeof(cases[0]); ++i) {
            for (unsigned code = 155; code <= 157; ++code) {
                reset(cases[i].now, cases[i].previous, cases[i].number);
                nox_xxx_mapGenSetFlags_5235F0((char)code);
                assert(game_queries == 1 && tick_reads == 1 && input_polls == 1 && displays == (unsigned)cases[i].draw);
                assert(dword_5d4594_2487564 == cases[i].now && dword_5d4594_2487568 == cases[i].stored);
                assert(sequence.value == cases[i].number + (uint32_t)cases[i].draw);
                if (cases[i].draw) {
                    assert(packet[0] == code);
                    assert(((unsigned)packet[1] | ((unsigned)packet[2] << 8)) == (cases[i].number & 0xFFFF));
                }
                canaries();
            }
        }
        reset(300, 0, 0xFFFF);
        mutate_poll = mutate_display = 1;
        expected_previous_at_display = 40;
        nox_xxx_mapGenSetFlags_5235F0((char)157);
        assert(displays == 1 && packet[0] == 157 && packet[1] == 255 && packet[2] == 255);
        assert(dword_5d4594_2487564 == 777 && dword_5d4594_2487568 == 777);
        assert(sequence.value == 0xABCDEF12);
        canaries();
    }
    printf("Mapgen progress %s: PASS\n", argv[1]);
    return 0;
}
