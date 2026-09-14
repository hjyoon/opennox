package legacy

/*
#include <stdint.h>
#include <string.h>
#include "GAME2.h"

typedef struct nox_test_spell_slot_45DA50_result {
	int offset;
	uint32_t spell;
	uint8_t flags;
} nox_test_spell_slot_45DA50_result;

static nox_test_spell_slot_45DA50_result nox_test_spell_slot_45DA50(int row, int slot, int nil_base) {
	unsigned char quickbar[240] = {0};
	quickbar[200] = (unsigned char)row;
	// The old 32-bit selected-row pointer at +204 remains zero on native builds.
	if (row >= 0 && row < 5 && slot >= 0 && slot < 5) {
		unsigned char* chosen = quickbar + 40 * row + 8 * slot;
		uint32_t spell = 0x12345678;
		memcpy(chosen, &spell, sizeof(spell));
		chosen[4] = 3;
	}
	unsigned char* chosen = nox_quickbar_selected_slot(nil_base ? 0 : quickbar, slot);
	nox_test_spell_slot_45DA50_result out = {.offset = -1};
	if (chosen) {
		out.offset = (int)(chosen - quickbar);
		memcpy(&out.spell, chosen, sizeof(out.spell));
		out.flags = chosen[4];
	}
	return out;
}
*/
import "C"

type spellSlotFixture45DA50 struct {
	Offset int
	Spell  uint32
	Flags  uint8
}

func quickbarSelectedSlotFixture45DA50(row, slot int, nilBase bool) spellSlotFixture45DA50 {
	var cNilBase C.int
	if nilBase {
		cNilBase = 1
	}
	out := C.nox_test_spell_slot_45DA50(C.int(row), C.int(slot), cNilBase)
	return spellSlotFixture45DA50{Offset: int(out.offset), Spell: uint32(out.spell), Flags: uint8(out.flags)}
}
