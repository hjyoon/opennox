package opennox

import (
	"fmt"
	"math"
	"reflect"
	"testing"
)

func TestSpellResult4FB0B0ExactTableAndCallOrder(t *testing.T) {
	wantKeys := [...]string{
		"SpellOK",
		"SpellInUse",
		"execspel.c:UnseenTarget",
		"TooManySpells",
		"summon.c:CreatureControlFailed",
		"glyph.c:TooManyGlyphs",
		"Spell.c:DuplicateGlyphSpell",
		"glyph.c:CantCastGlyph",
		"BadTarget",
		"BadSkill",
		"Illegal",
		"NotEnoughManaCast",
		"spell.c:NotEnoughManaGlyph",
		"SpellRestrictedByFlag",
		"SpellNotStartedWarCry",
		"spell.c:SpellCancelledByWarCry",
		"SpellRestrictedByBall",
		"SpellRestrictedByCrown",
	}
	if spellResultKeys4FB0B0 != wantKeys {
		t.Fatalf("keys = %#v, want %#v", spellResultKeys4FB0B0, wantKeys)
	}

	for status, wantKey := range wantKeys {
		status := uint32(status)
		wantText := "translated:" + wantKey
		var events []string
		ok := spellResult4FB0B0(status, spellResultHooks4FB0B0{
			loadString: func(key, source string) string {
				events = append(events, fmt.Sprintf("load:%s:%s", key, source))
				return wantText
			},
			printCentered: func(text string) {
				events = append(events, "print:"+text)
			},
		})
		if !ok {
			t.Fatalf("status %d was rejected", status)
		}
		wantEvents := []string{
			fmt.Sprintf("load:%s:%s", wantKey, spellResultSourcePath4FB0B0),
			"print:" + wantText,
		}
		if !reflect.DeepEqual(events, wantEvents) {
			t.Fatalf("status %d events = %v, want %v", status, events, wantEvents)
		}
	}
}

func TestSpellResult4FB0B0RejectsUndefinedIndicesBeforeCallbacks(t *testing.T) {
	for _, status := range []uint32{uint32(len(spellResultKeys4FB0B0)), math.MaxUint32} {
		calls := 0
		ok := spellResult4FB0B0(status, spellResultHooks4FB0B0{
			loadString: func(string, string) string {
				calls++
				return "unexpected"
			},
			printCentered: func(string) {
				calls++
			},
		})
		if ok || calls != 0 {
			t.Fatalf("status %#x = ok %t, callbacks %d; want false, 0", status, ok, calls)
		}
	}
}
