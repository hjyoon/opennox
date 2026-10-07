package opennox

import (
	"fmt"
	"testing"

	"github.com/opennox/libs/spell"
)

func TestE2ERestoreSpellModes(t *testing.T) {
	for _, tc := range []struct {
		mode string
		id   spell.ID
		npc  bool
	}{
		{"health-player", spell.SPELL_RESTORE_HEALTH, false},
		{"health-npc", spell.SPELL_RESTORE_HEALTH, true},
		{"wink-player", spell.SPELL_WINK, false},
		{"wink-npc", spell.SPELL_WINK, true},
		{"mana-player", spell.SPELL_RESTORE_MANA, false},
		{"mana-npc", spell.SPELL_RESTORE_MANA, true},
	} {
		for level := 1; level <= 5; level++ {
			t.Run(fmt.Sprintf("%s/%d", tc.mode, level), func(t *testing.T) {
				id, npc, ok := e2eRestoreSpellMode(level, tc.mode)
				if !ok || id != tc.id || npc != tc.npc {
					t.Fatalf("id/npc/ok = %s/%t/%t", id, npc, ok)
				}
			})
		}
		for _, level := range []int{-1, 0, 6} {
			if id, npc, ok := e2eRestoreSpellMode(level, tc.mode); ok || id != 0 || npc {
				t.Fatalf("accepted invalid level %d for %s", level, tc.mode)
			}
		}
	}
	for _, mode := range []string{"", "health", "mana", "Health-player", "health-npc-extra"} {
		if id, npc, ok := e2eRestoreSpellMode(1, mode); ok || id != 0 || npc {
			t.Fatalf("accepted invalid mode %q", mode)
		}
	}
}
