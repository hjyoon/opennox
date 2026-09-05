package server

import (
	"math"
	"reflect"
	"runtime"
	"testing"
	"unsafe"

	"github.com/opennox/libs/spell"
	"github.com/opennox/libs/things"
)

func TestCastSpellByUserNative4FDD20PreservesPointersAndSignedDwords(t *testing.T) {
	caster := new(Object)
	target := new(Object)
	arg := &SpellAcceptArg{Obj: target}
	var events []string
	got := castSpellByUserNative4FDD20(math.MinInt32, caster, arg, castSpellByUserNativeDeps4FDD20{
		spellHasFlags: func(id int32, mask uint32) int32 {
			events = append(events, "flags")
			if id != math.MinInt32 {
				t.Fatalf("flags spell = %d, want %d", id, int32(math.MinInt32))
			}
			if mask == castSpellByUserTargetedFlag4FDD20 {
				return math.MaxInt32
			}
			return 0
		},
		runtime: CastSpellByUserRuntime4FDD20{
			SpellGetPower: func(id spell.ID, gotCaster *Object) int32 {
				events = append(events, "power")
				if int32(id) != math.MinInt32 || gotCaster != caster {
					t.Fatalf("power args = %d/%p, want %d/%p", id, gotCaster, int32(math.MinInt32), caster)
				}
				return math.MaxInt32
			},
			DisableEnchant: func(*Object, EnchantID) { t.Fatal("unexpected enchant cancellation") },
			CancelDuration: func(spell.ID, *Object) { t.Fatal("unexpected duration cancellation") },
			CreateProjectile: func(gotCaster, gotTarget *Object, id spell.ID) {
				events = append(events, "projectile")
				if gotCaster != caster || gotTarget != target || int32(id) != math.MinInt32 {
					t.Fatalf("projectile args = %p/%p/%d", gotCaster, gotTarget, id)
				}
			},
			SpellAccept: func(spell.ID, *Object, *Object, *Object, *SpellAcceptArg, int32) int32 {
				t.Fatal("distinct target reached acceptance")
				return 0
			},
		},
	})
	if got != 1 {
		t.Fatalf("result = %d, want canonical 1", got)
	}
	if !reflect.DeepEqual(events, []string{"power", "flags", "flags", "projectile"}) {
		t.Fatalf("events = %v", events)
	}
	if unsafe.Sizeof(uintptr(0)) == 8 {
		for name, ptr := range map[string]uintptr{
			"caster": uintptr(unsafe.Pointer(caster)),
			"target": uintptr(unsafe.Pointer(target)),
			"arg":    uintptr(unsafe.Pointer(arg)),
		} {
			if ptr <= math.MaxUint32 {
				t.Fatalf("%s pointer = %#x, want native address above 4 GiB", name, ptr)
			}
		}
	}
	runtime.KeepAlive(caster)
	runtime.KeepAlive(target)
	runtime.KeepAlive(arg)
}

func TestCastSpellByUserNative4FDD20ReturnsAcceptanceDwordUnchanged(t *testing.T) {
	caster := new(Object)
	arg := &SpellAcceptArg{Obj: caster}
	got := castSpellByUserNative4FDD20(math.MaxInt32, caster, arg, castSpellByUserNativeDeps4FDD20{
		spellHasFlags: func(int32, uint32) int32 { return 0 },
		runtime: CastSpellByUserRuntime4FDD20{
			SpellGetPower: func(id spell.ID, gotCaster *Object) int32 {
				if int32(id) != math.MaxInt32 || gotCaster != caster {
					t.Fatalf("power args = %d/%p", id, gotCaster)
				}
				return math.MinInt32
			},
			DisableEnchant:   func(*Object, EnchantID) { t.Fatal("unexpected disable") },
			CancelDuration:   func(spell.ID, *Object) { t.Fatal("unexpected cancel") },
			CreateProjectile: func(*Object, *Object, spell.ID) { t.Fatal("unexpected projectile") },
			SpellAccept: func(id spell.ID, second, third, fourth *Object, gotArg *SpellAcceptArg, power int32) int32 {
				if int32(id) != math.MaxInt32 || second != caster || third != caster || fourth != caster || gotArg != arg || power != math.MinInt32 {
					t.Fatalf("accept args = %d/%p/%p/%p/%p/%d", id, second, third, fourth, gotArg, power)
				}
				return math.MinInt32
			},
		},
	})
	if got != math.MinInt32 {
		t.Fatalf("result = %d, want verbatim %d", got, int32(math.MinInt32))
	}
	runtime.KeepAlive(caster)
	runtime.KeepAlive(arg)
}

func TestCastSpellByUserServer4FDD20UsesLiveFlagsAndExactEffects(t *testing.T) {
	const id = spell.ID(91)
	s := new(Server)
	s.Spells.byID = map[spell.ID]*SpellDef{
		id: {
			ID: id,
			Def: things.Spell{
				Flags: things.SpellOffensive | things.SpellTargeted,
			},
		},
	}
	caster := new(Object)
	arg := &SpellAcceptArg{Obj: caster}
	var events []string
	got := s.CastSpellByUser4FDD20(id, caster, arg, CastSpellByUserRuntime4FDD20{
		SpellGetPower: func(gotID spell.ID, gotCaster *Object) int32 {
			events = append(events, "power")
			if gotID != id || gotCaster != caster {
				t.Fatalf("power args = %d/%p", gotID, gotCaster)
			}
			return 6
		},
		DisableEnchant: func(gotCaster *Object, enchant EnchantID) {
			events = append(events, "disable:"+enchant.String())
			if gotCaster != caster {
				t.Fatalf("disable caster = %p, want %p", gotCaster, caster)
			}
		},
		CancelDuration: func(gotID spell.ID, gotCaster *Object) {
			events = append(events, "cancel")
			if gotID != spell.SPELL_OVAL_SHIELD || gotCaster != caster {
				t.Fatalf("cancel args = %d/%p", gotID, gotCaster)
			}
		},
		CreateProjectile: func(*Object, *Object, spell.ID) { t.Fatal("self target created projectile") },
		SpellAccept: func(gotID spell.ID, second, third, fourth *Object, gotArg *SpellAcceptArg, power int32) int32 {
			events = append(events, "accept")
			if gotID != id || second != caster || third != caster || fourth != caster || gotArg != arg || power != 6 {
				t.Fatalf("accept args = %d/%p/%p/%p/%p/%d", gotID, second, third, fourth, gotArg, power)
			}
			return math.MaxInt32
		},
	})
	if got != math.MaxInt32 {
		t.Fatalf("result = %d, want verbatim %d", got, int32(math.MaxInt32))
	}
	want := []string{
		"power",
		"disable:ENCHANT_INVISIBLE",
		"disable:ENCHANT_INVULNERABLE",
		"cancel",
		"accept",
	}
	if !reflect.DeepEqual(events, want) {
		t.Fatalf("events = %v, want %v", events, want)
	}
	runtime.KeepAlive(caster)
	runtime.KeepAlive(arg)
}

func TestCastSpellByUserServer4FDD20NonTargetedNilArgIsForwarded(t *testing.T) {
	const id = spell.ID(7)
	s := new(Server)
	s.Spells.byID = map[spell.ID]*SpellDef{id: {ID: id}}
	caster := new(Object)
	got := s.CastSpellByUser4FDD20(id, caster, nil, CastSpellByUserRuntime4FDD20{
		SpellGetPower:    func(spell.ID, *Object) int32 { return 3 },
		DisableEnchant:   func(*Object, EnchantID) { t.Fatal("unexpected disable") },
		CancelDuration:   func(spell.ID, *Object) { t.Fatal("unexpected cancel") },
		CreateProjectile: func(*Object, *Object, spell.ID) { t.Fatal("unexpected projectile") },
		SpellAccept: func(_ spell.ID, second, third, fourth *Object, arg *SpellAcceptArg, power int32) int32 {
			if second != caster || third != caster || fourth != caster || arg != nil || power != 3 {
				t.Fatalf("accept args = %p/%p/%p/%p/%d", second, third, fourth, arg, power)
			}
			return -17
		},
	})
	if got != -17 {
		t.Fatalf("result = %d, want -17", got)
	}
	runtime.KeepAlive(caster)
}
