package opennox

import (
	"strings"
	"testing"
	"time"
	"unsafe"

	"github.com/opennox/libs/spell"
	"github.com/opennox/libs/types"

	"github.com/opennox/opennox/v1/legacy/common/alloc"
	"github.com/opennox/opennox/v1/server"
)

func TestE2EPlayerStatusSpellInitializesNativeArgument(t *testing.T) {
	unit, freeUnit := alloc.New(server.Object{})
	t.Cleanup(freeUnit)
	*unit = server.Object{PosVec: types.Pointf{X: -33.5, Y: 12.25}}
	arg, freeArg := e2ePlayerStatusSpellArg(unit)
	t.Cleanup(freeArg)
	if arg.Obj != unit || arg.Pos != unit.Pos() {
		t.Fatalf("spell argument = %+v, want actual target/position", *arg)
	}
	if unsafe.Sizeof(uintptr(0)) > 4 {
		for _, ptr := range []unsafe.Pointer{unsafe.Pointer(unit), unsafe.Pointer(arg)} {
			if uintptr(ptr) <= 0xffffffff {
				t.Fatalf("native pointer = %p, want above 4 GiB", ptr)
			}
		}
	}
	unit.PosVec = types.Pointf{X: 1, Y: 2}
	if arg.Obj != unit || arg.Pos != (types.Pointf{X: -33.5, Y: 12.25}) {
		t.Fatal("argument did not retain the target identity and cast-time position")
	}
}

func TestE2EPlayerStatusSpellSchedule(t *testing.T) {
	for _, tc := range []struct {
		kind           string
		id             spell.ID
		animation, key string
	}{
		{"confused", spell.SPELL_CONFUSE, "confused", "ConfuseEnchantDuration"},
		{"stun", spell.SPELL_STUN, "stun", "StunEnchantDuration"},
		{"stun-warrior", spell.SPELL_STUN, "slow", "StunEnchantDuration"},
	} {
		t.Run(tc.kind, func(t *testing.T) {
			id, animation, key, ok := e2ePlayerStatusSpell(tc.kind)
			if !ok || id != tc.id || animation != tc.animation || key != tc.key {
				t.Fatalf("spell case = %s/%s/%s/%t", id, animation, key, ok)
			}
			var sc e2eScenario
			sc.CheckPlayerStatusSpell(tc.kind, "spell")
			if len(sc.steps) != 7 {
				t.Fatalf("steps = %d, want prepare/two pixel checks/expiry/three screens", len(sc.steps))
			}
			for suffix, timeout := range map[string]time.Duration{" prepare": 1200, " first visible": 120, " advanced visible": 120, " natural expiry": 3000} {
				found := 0
				for _, step := range sc.steps {
					if strings.HasSuffix(step.name, suffix) {
						found++
						if step.ready == nil || step.fnc == nil || step.waitTimeout != timeout {
							t.Fatalf("%q lacks a bounded live predicate/result check", suffix)
						}
					}
				}
				if found != 1 {
					t.Fatalf("%q found %d times, want one", suffix, found)
				}
			}
			for _, label := range []string{"first", "advanced", "expired"} {
				found := 0
				for _, step := range sc.steps {
					if step.name == "player status spell "+tc.kind+" "+label {
						found++
					}
				}
				if found != 1 {
					t.Fatalf("%s screenshot count = %d", label, found)
				}
			}
		})
	}
}

func TestE2EPlayerStatusSpellRejectsUnknown(t *testing.T) {
	for _, kind := range []string{"", "STUN", "slow", "poison", "shield"} {
		var sc e2eScenario
		func() {
			defer func() {
				if recover() == nil || len(sc.steps) != 0 {
					t.Errorf("unknown status spell %q scheduled a mutation", kind)
				}
			}()
			sc.CheckPlayerStatusSpell(kind, "invalid")
		}()
	}
}
