package server

import (
	"fmt"
	"math"
	"reflect"
	"testing"
)

func castSpellByUserTestHooks4FDD20(t *testing.T) castSpellByUserHooks4FDD20[int, int] {
	t.Helper()
	return castSpellByUserHooks4FDD20[int, int]{
		loadCasterArg: func() int { return 2 },
		loadSpellArg:  func() int32 { return 7 },
		spellGetPower: func(int32, int) int32 { return 3 },
		spellHasFlags: func(int32, uint32) int32 { return 0 },
		disableEnchant: func(int, int32) {
			t.Fatal("unexpected enchant cancellation")
		},
		cancelDuration: func(int32, int) {
			t.Fatal("unexpected duration cancellation")
		},
		loadAcceptArg: func() int { return 5 },
		loadTarget: func(int) int {
			t.Fatal("non-targeted spell loaded target")
			return 0
		},
		createProjectile: func(int, int, int32) {
			t.Fatal("unexpected projectile")
		},
		spellAccept: func(int32, int, int, int, int, int32) int32 { return 1 },
	}
}

func TestCastSpellByUser4FDD20ExactOrderCachesEntryAndReturnsDword(t *testing.T) {
	const (
		entryCaster = 0x1002
		entrySpell  = int32(0x70000007)
		lateArg     = 0x2005
		power       = int32(math.MinInt32)
		result      = int32(math.MaxInt32)
	)
	caster := entryCaster
	spellID := entrySpell
	arg := 5
	var events []string
	hooks := castSpellByUserTestHooks4FDD20(t)
	hooks.loadCasterArg = func() int {
		events = append(events, "caster")
		return caster
	}
	hooks.loadSpellArg = func() int32 {
		events = append(events, "spell")
		return spellID
	}
	hooks.spellGetPower = func(gotID int32, gotCaster int) int32 {
		events = append(events, "power")
		if gotID != entrySpell || gotCaster != entryCaster {
			t.Fatalf("power args = %#x/%#x", gotID, gotCaster)
		}
		caster = 0x3002
		spellID = -1
		return power
	}
	hooks.spellHasFlags = func(gotID int32, mask uint32) int32 {
		events = append(events, fmt.Sprintf("flags:%#x", mask))
		if gotID != entrySpell {
			t.Fatalf("flags spell = %#x, want cached %#x", gotID, entrySpell)
		}
		if mask == castSpellByUserTargetedFlag4FDD20 {
			arg = lateArg
			return 1
		}
		return 1
	}
	hooks.disableEnchant = func(gotCaster int, enchant int32) {
		events = append(events, fmt.Sprintf("disable:%d", enchant))
		if gotCaster != entryCaster {
			t.Fatalf("disable caster = %#x, want cached %#x", gotCaster, entryCaster)
		}
	}
	hooks.cancelDuration = func(gotSpell int32, gotCaster int) {
		events = append(events, fmt.Sprintf("cancel:%d", gotSpell))
		if gotSpell != castSpellByUserOvalShieldSpell4FDD20 || gotCaster != entryCaster {
			t.Fatalf("cancel args = %d/%#x", gotSpell, gotCaster)
		}
	}
	hooks.loadAcceptArg = func() int {
		events = append(events, "arg")
		return arg
	}
	hooks.loadTarget = func(gotArg int) int {
		events = append(events, "target")
		if gotArg != lateArg {
			t.Fatalf("target arg = %#x, want late %#x", gotArg, lateArg)
		}
		return entryCaster
	}
	hooks.spellAccept = func(gotID int32, second, third, fourth, gotArg int, gotPower int32) int32 {
		events = append(events, "accept")
		if gotID != entrySpell || second != entryCaster || third != entryCaster || fourth != entryCaster || gotArg != lateArg || gotPower != power {
			t.Fatalf("accept args = %#x/%#x/%#x/%#x/%#x/%d", gotID, second, third, fourth, gotArg, gotPower)
		}
		return result
	}
	if got := castSpellByUser4FDD20(hooks); got != result {
		t.Fatalf("result = %d, want verbatim %d", got, result)
	}
	want := []string{
		"caster", "spell", "power", "flags:0x20",
		"disable:0", "disable:23", "cancel:67",
		"flags:0x4", "arg", "target", "accept",
	}
	if !reflect.DeepEqual(events, want) {
		t.Fatalf("events = %v, want %v", events, want)
	}
}

func TestCastSpellByUser4FDD20EveryNonzeroFlagValueIsTrue(t *testing.T) {
	for _, value := range []int32{-1, 1, 2, math.MinInt32, math.MaxInt32} {
		t.Run(fmt.Sprintf("offensive_%d", value), func(t *testing.T) {
			var events []string
			hooks := castSpellByUserTestHooks4FDD20(t)
			hooks.spellHasFlags = func(_ int32, mask uint32) int32 {
				if mask == castSpellByUserOffensiveFlag4FDD20 {
					return value
				}
				return 0
			}
			hooks.disableEnchant = func(_ int, enchant int32) {
				events = append(events, fmt.Sprintf("disable:%d", enchant))
			}
			hooks.cancelDuration = func(id int32, _ int) {
				events = append(events, fmt.Sprintf("cancel:%d", id))
			}
			if got := castSpellByUser4FDD20(hooks); got != 1 {
				t.Fatalf("result = %d, want 1", got)
			}
			want := []string{"disable:0", "disable:23", "cancel:67"}
			if !reflect.DeepEqual(events, want) {
				t.Fatalf("events = %v, want %v", events, want)
			}
		})

		t.Run(fmt.Sprintf("targeted_%d", value), func(t *testing.T) {
			projectiles := 0
			hooks := castSpellByUserTestHooks4FDD20(t)
			hooks.spellHasFlags = func(_ int32, mask uint32) int32 {
				if mask == castSpellByUserTargetedFlag4FDD20 {
					return value
				}
				return 0
			}
			hooks.loadTarget = func(int) int { return 9 }
			hooks.createProjectile = func(caster, target int, id int32) {
				projectiles++
				if caster != 2 || target != 9 || id != 7 {
					t.Fatalf("projectile args = %d/%d/%d", caster, target, id)
				}
			}
			hooks.spellAccept = func(int32, int, int, int, int, int32) int32 {
				t.Fatal("projectile path reached acceptance")
				return 0
			}
			if got := castSpellByUser4FDD20(hooks); got != 1 {
				t.Fatalf("result = %d, want canonical 1", got)
			}
			if projectiles != 1 {
				t.Fatalf("projectiles = %d, want 1", projectiles)
			}
		})
	}
}

func TestCastSpellByUser4FDD20NonTargetedNilArgumentIsForwarded(t *testing.T) {
	type acceptArg struct{ target int }
	hooks := castSpellByUserHooks4FDD20[int, *acceptArg]{
		loadCasterArg:  func() int { return 2 },
		loadSpellArg:   func() int32 { return 7 },
		spellGetPower:  func(int32, int) int32 { return math.MaxInt32 },
		spellHasFlags:  func(int32, uint32) int32 { return 0 },
		disableEnchant: func(int, int32) { t.Fatal("unexpected disable") },
		cancelDuration: func(int32, int) { t.Fatal("unexpected cancel") },
		loadAcceptArg:  func() *acceptArg { return nil },
		loadTarget: func(*acceptArg) int {
			t.Fatal("non-targeted nil argument was dereferenced")
			return 0
		},
		createProjectile: func(int, int, int32) { t.Fatal("unexpected projectile") },
		spellAccept: func(id int32, second, third, fourth int, arg *acceptArg, power int32) int32 {
			if id != 7 || second != 2 || third != 2 || fourth != 2 || arg != nil || power != math.MaxInt32 {
				t.Fatalf("accept args = %d/%d/%d/%d/%p/%d", id, second, third, fourth, arg, power)
			}
			return math.MinInt32
		},
	}
	if got := castSpellByUser4FDD20(hooks); got != math.MinInt32 {
		t.Fatalf("result = %d, want verbatim %d", got, int32(math.MinInt32))
	}
}

func TestCastSpellByUser4FDD20TargetLoadedOnceAndCachedForProjectile(t *testing.T) {
	argTarget := 9
	loads := 0
	hooks := castSpellByUserTestHooks4FDD20(t)
	hooks.spellHasFlags = func(_ int32, mask uint32) int32 {
		if mask == castSpellByUserTargetedFlag4FDD20 {
			return 1
		}
		return 0
	}
	hooks.loadTarget = func(int) int {
		loads++
		got := argTarget
		argTarget = 11
		return got
	}
	hooks.createProjectile = func(caster, target int, id int32) {
		if caster != 2 || target != 9 || id != 7 {
			t.Fatalf("projectile args = %d/%d/%d, want cached 2/9/7", caster, target, id)
		}
	}
	hooks.spellAccept = func(int32, int, int, int, int, int32) int32 {
		t.Fatal("distinct target reached acceptance")
		return 0
	}
	if got := castSpellByUser4FDD20(hooks); got != 1 {
		t.Fatalf("result = %d, want canonical 1", got)
	}
	if loads != 1 {
		t.Fatalf("target loads = %d, want 1", loads)
	}
}

func TestCastSpellByUser4FDD20TargetedNilArgumentFaultPrefix(t *testing.T) {
	type acceptArg struct{ target int }
	var events []string
	hooks := castSpellByUserHooks4FDD20[int, *acceptArg]{
		loadCasterArg: func() int { events = append(events, "caster"); return 2 },
		loadSpellArg:  func() int32 { events = append(events, "spell"); return 7 },
		spellGetPower: func(int32, int) int32 { events = append(events, "power"); return 3 },
		spellHasFlags: func(_ int32, mask uint32) int32 {
			events = append(events, fmt.Sprintf("flags:%#x", mask))
			if mask == castSpellByUserTargetedFlag4FDD20 {
				return 1
			}
			return 0
		},
		disableEnchant: func(int, int32) { t.Fatal("unexpected disable") },
		cancelDuration: func(int32, int) { t.Fatal("unexpected cancel") },
		loadAcceptArg:  func() *acceptArg { events = append(events, "arg"); return nil },
		loadTarget: func(arg *acceptArg) int {
			events = append(events, "target")
			return arg.target
		},
		createProjectile: func(int, int, int32) { t.Fatal("unexpected projectile") },
		spellAccept: func(int32, int, int, int, *acceptArg, int32) int32 {
			t.Fatal("faulted target reached acceptance")
			return 0
		},
	}
	func() {
		defer func() {
			if recover() == nil {
				t.Fatal("targeted nil argument did not fault")
			}
		}()
		castSpellByUser4FDD20(hooks)
	}()
	want := []string{"caster", "spell", "power", "flags:0x20", "flags:0x4", "arg", "target"}
	if !reflect.DeepEqual(events, want) {
		t.Fatalf("events = %v, want fault prefix %v", events, want)
	}
}

func TestCastSpellByUser4FDD20EveryObservableFaultPrefix(t *testing.T) {
	tests := []struct {
		name   string
		target int
		tail   string
	}{
		{name: "accept", target: 2, tail: "accept"},
		{name: "projectile", target: 9, tail: "projectile"},
	}
	for _, test := range tests {
		want := []string{
			"caster", "spell", "power", "flags:0x20",
			"disable:0", "disable:23", "cancel:67",
			"flags:0x4", "arg", "target", test.tail,
		}
		for faultAt := range want {
			t.Run(fmt.Sprintf("%s/%02d_%s", test.name, faultAt, want[faultAt]), func(t *testing.T) {
				var events []string
				record := func(event string) {
					events = append(events, event)
					if len(events)-1 == faultAt {
						panic("injected fault")
					}
				}
				hooks := castSpellByUserHooks4FDD20[int, int]{
					loadCasterArg: func() int { record("caster"); return 2 },
					loadSpellArg:  func() int32 { record("spell"); return 7 },
					spellGetPower: func(int32, int) int32 { record("power"); return 3 },
					spellHasFlags: func(_ int32, mask uint32) int32 {
						record(fmt.Sprintf("flags:%#x", mask))
						return 1
					},
					disableEnchant:   func(_ int, enchant int32) { record(fmt.Sprintf("disable:%d", enchant)) },
					cancelDuration:   func(id int32, _ int) { record(fmt.Sprintf("cancel:%d", id)) },
					loadAcceptArg:    func() int { record("arg"); return 5 },
					loadTarget:       func(int) int { record("target"); return test.target },
					createProjectile: func(int, int, int32) { record("projectile") },
					spellAccept:      func(int32, int, int, int, int, int32) int32 { record("accept"); return 1 },
				}
				func() {
					defer func() {
						if recover() == nil {
							t.Fatal("injected fault did not propagate")
						}
					}()
					castSpellByUser4FDD20(hooks)
				}()
				prefix := want[:faultAt+1]
				if !reflect.DeepEqual(events, prefix) {
					t.Fatalf("events = %v, want prefix %v", events, prefix)
				}
			})
		}
	}
}
