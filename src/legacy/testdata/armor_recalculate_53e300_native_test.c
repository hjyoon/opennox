#include <assert.h>
#include <math.h>
#include <stdint.h>
#include <stdio.h>
#include <stdlib.h>
#include <string.h>
#include "defs.h"
#undef _Static_assert
_Static_assert(sizeof(void*) == 8, "native pointer width");
_Static_assert(sizeof(int) == 4, "original integer width");
#include "GAME1.h"
#include "GAME1_1.h"
#include "GAME4_3.h"

static nox_object_t owner, items[16];
static float values[16];
static int present[16], count, events, stores;
static uint32_t player_bits = 0x3e800000, monster_bits = 0x3ec00000;
uint64_t qword_581450_9512 = UINT64_C(0x3ff0000000000000);

double nox_xxx_itemApplyDefendEffect_415C00(nox_object_t* item) {
    assert(item >= items && item < items + count);
    int i = (int)(item - items);
    printf("%s{\"service\":\"itemArmor\",\"item_index\":%d}", events++ ? "," : "", i);
    return present[i] ? (double)values[i] : 0.0;
}
void nox_xxx_objectSetArmorValue(nox_object_t* object, float value) {
    assert(object == &owner);
    uint32_t bits;
    memcpy(&bits, &value, sizeof(bits));
    if (object->obj_class & 4) player_bits = bits;
    else if (object->obj_class & 2) monster_bits = bits;
    stores++;
}

// PRODUCTION_DOUBLE2FLOAT_419AF0
// PRODUCTION_BODY_53E300

int main(int argc, char** argv) {
    assert(argc >= 3);
    owner.obj_class = (uint32_t)strtoul(argv[1], NULL, 0);
    count = atoi(argv[2]);
    assert(count >= 0 && count <= 16 && argc == 3 + 4 * count);
    owner.inv_first_item = count ? items : NULL;
    for (int i = 0; i < count; i++) {
        items[i].obj_class = (uint32_t)strtoul(argv[3 + 4*i], NULL, 0);
        items[i].obj_flags = (uint32_t)strtoul(argv[4 + 4*i], NULL, 0);
        uint32_t bits = (uint32_t)strtoul(argv[5 + 4*i], NULL, 0);
        memcpy(&values[i], &bits, sizeof(bits));
        present[i] = atoi(argv[6 + 4*i]);
        items[i].inv_next_item = i + 1 < count ? items + i + 1 : NULL;
    }
    uintptr_t ptrs[] = {(uintptr_t)&owner, (uintptr_t)items, (uintptr_t)values,
                        (uintptr_t)&player_bits, (uintptr_t)&monster_bits};
    for (unsigned i = 0; i < sizeof(ptrs)/sizeof(ptrs[0]); i++) assert(ptrs[i] > UINT32_MAX);
    nox_object_t before_owner = owner, before_items[16];
    memcpy(before_items, items, sizeof(items));
    float before_values[16]; memcpy(before_values, values, sizeof(values));
    printf("{\"pointer_bytes\":8,\"high_native_pointers\":true,\"events\":[");
    nox_xxx_recalculateArmorVal_53E300(&owner);
    assert(!memcmp(&owner, &before_owner, sizeof(owner)));
    assert(!memcmp(items, before_items, sizeof(items)));
    assert(!memcmp(values, before_values, sizeof(values)));
    assert(stores == 1);
    printf("],\"player_armor_bits\":%u,\"monster_armor_bits\":%u,"
           "\"nonoutput_canaries_unchanged\":true}\n", player_bits, monster_bits);
    return 0;
}
