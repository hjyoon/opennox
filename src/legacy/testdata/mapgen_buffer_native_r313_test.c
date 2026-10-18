#include <assert.h>
#include <stdint.h>
#include <stdio.h>
#include <stdlib.h>
#include <string.h>

static int fail_allocation, fail_registration;
static unsigned allocations, releases;
static void* live_buffer;
static struct {
    uint32_t before, token, after;
} legacy_slot = {UINT32_C(0x13579BDF), 0, UINT32_C(0x2468ACE0)};
#define dword_5d4594_2487556 legacy_slot.token

static void* registry_malloc(size_t size) {
    return fail_registration ? NULL : malloc(size);
}
#define malloc registry_malloc
#include "mapgen_legacy_ptr.c"
#undef malloc

static void* buffer_calloc(size_t count, size_t size) {
    assert(count == 1 && size == 0x2000 && !live_buffer);
    if (fail_allocation) return NULL;
    live_buffer = calloc(count, size);
    assert(live_buffer);
#if UINTPTR_MAX > UINT32_MAX
    assert((uintptr_t)live_buffer > UINT32_MAX);
#endif
    ++allocations;
    return live_buffer;
}

static void buffer_free(void* pointer) {
    if (!pointer) return;
    assert(pointer == live_buffer);
    ++releases;
    free(pointer);
    live_buffer = NULL;
}

#define calloc buffer_calloc
#define free buffer_free
// COMPLETE_MAPGEN_BUFFER_R313
#undef free
#undef calloc

static void assert_clean(void) {
    assert(!live_buffer && !legacy_slot.token && !nox_mapgen_legacy_ptrs);
    assert(legacy_slot.before == UINT32_C(0x13579BDF));
    assert(legacy_slot.after == UINT32_C(0x2468ACE0));
    assert(allocations == releases);
}

int main(void) {
    nox_xxx_mapgenFreeBuffer_521400();
    assert_clean();
    fail_allocation = 1;
    assert(!nox_xxx_mapgenAllocBuffer_5213E0());
    assert_clean();
    fail_allocation = 0;
    fail_registration = 1;
    assert(!nox_xxx_mapgenAllocBuffer_5213E0());
    assert_clean();
    fail_registration = 0;
    for (unsigned iteration = 0; iteration < 12; ++iteration) {
        assert(nox_xxx_mapgenAllocBuffer_5213E0());
        assert(legacy_slot.token);
        assert(nox_mapgenLegacyPtrResolve(legacy_slot.token) == live_buffer);
        for (unsigned byte = 0; byte < 0x2000; ++byte) {
            assert(((unsigned char*)live_buffer)[byte] == 0);
        }
        memset(live_buffer, 0xA5, 0x2000);
        nox_xxx_mapgenFreeBuffer_521400();
        assert_clean();
        nox_xxx_mapgenFreeBuffer_521400();
        assert_clean();
    }
    puts("Mapgen buffer native R313: 12 repeated lifetimes and allocation failures PASS");
    return 0;
}
