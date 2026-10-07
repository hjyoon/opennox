package opennox

import (
	"testing"

	"github.com/opennox/libs/spell"
)

func TestE2ERestoreNPCClientHPOriginalPacking(t *testing.T) {
	for _, tc := range []struct {
		name    string
		id      spell.ID
		injured uint16
		maximum uint16
		want    uint16
	}{
		{"health stock NPC", spell.SPELL_RESTORE_HEALTH, 143, 150, 150},
		{"wink stock NPC", spell.SPELL_WINK, 143, 150, 150},
		{"mana NPC remains injured", spell.SPELL_RESTORE_MANA, 143, 150, 142},
		{"health odd maximum", spell.SPELL_RESTORE_HEALTH, 143, 151, 150},
		{"wink odd maximum", spell.SPELL_WINK, 143, 151, 150},
		{"maximum byte boundary", spell.SPELL_RESTORE_HEALTH, 143, 511, 510},
		{"maximum byte wrap", spell.SPELL_RESTORE_HEALTH, 143, 512, 0},
		{"unsigned maximum", spell.SPELL_WINK, 143, 65535, 510},
		{"mana even injury", spell.SPELL_RESTORE_MANA, 142, 65535, 142},
		{"mana byte boundary", spell.SPELL_RESTORE_MANA, 511, 65535, 510},
		{"mana byte wrap", spell.SPELL_RESTORE_MANA, 512, 65535, 0},
		{"mana unsigned injury", spell.SPELL_RESTORE_MANA, 65535, 150, 510},
		{"zero maximum", spell.SPELL_RESTORE_HEALTH, 143, 0, 0},
		{"zero injury", spell.SPELL_RESTORE_MANA, 0, 150, 0},
	} {
		t.Run(tc.name, func(t *testing.T) {
			if got := e2eRestoreNPCClientHP(tc.id, tc.injured, tc.maximum); got != tc.want {
				t.Fatalf("client HP = %d, want original unsigned-byte report %d", got, tc.want)
			}
		})
	}
}
