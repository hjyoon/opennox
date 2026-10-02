package server

import (
	"fmt"
	"math"
	"slices"
	"testing"
	"unsafe"

	"github.com/opennox/libs/object"
	"github.com/opennox/libs/types"

	"github.com/opennox/opennox/v1/legacy/common/alloc"
)

func defaultDamagePoisonFixture4E0B30(t *testing.T, subclass object.SubClass) *Object {
	t.Helper()
	target, freeTarget := alloc.New(Object{})
	t.Cleanup(freeTarget)
	update, freeUpdate := alloc.New(MonsterUpdateData{})
	t.Cleanup(freeUpdate)
	health, freeHealth := alloc.New(HealthData{})
	t.Cleanup(freeHealth)
	*update = MonsterUpdateData{Field547: 99, Field546: 91, Field1: math.Float32bits(0.25)}
	*health = HealthData{Cur: 20, Max: 20, Field2: 20}
	*target = Object{ObjClass: object.ClassMonster, ObjSubClass: subclass,
		UpdateData: unsafe.Pointer(update), HealthData: health,
		Pos132: types.Ptf(31, 47), Frame134: 77,
		Buffs: 1<<defaultDamageInvisibleEnchant4E0B30 | 1<<defaultDamageShieldEnchant4E0B30 | 1<<defaultDamageShockEnchant4E0B30,
	}
	if unsafe.Sizeof(uintptr(0)) == 8 {
		for _, ptr := range []unsafe.Pointer{unsafe.Pointer(target), unsafe.Pointer(update), unsafe.Pointer(health)} {
			if uintptr(ptr) <= math.MaxUint32 {
				t.Fatalf("native poison pointer=%p, want >4 GiB", ptr)
			}
		}
	}
	return target
}

func TestDefaultDamageWorld4E0B30SourceLessMonsterPoison(t *testing.T) {
	for _, subclass := range []object.SubClass{1, 0x10, 0x400, 0x800, 0x11012} {
		for _, amount := range []int32{-3, 0, 1, 25} {
			t.Run(fmt.Sprintf("subclass-%x/damage-%d", uint32(subclass), amount), func(t *testing.T) {
				target := defaultDamagePoisonFixture4E0B30(t, subclass)
				update := target.UpdateDataMonster()
				before := *target
				var events []string
				r := DefaultDamageWorldRuntime4E0B30{
					Frame:              func() uint32 { return 1400 },
					IsEnemy:            func(*Object, *Object) bool { t.Fatal("source-less poison tested enemy"); return false },
					BuffOff:            func(*Object, EnchantID) { t.Fatal("source-less poison removed invisibility") },
					FireProtection:     func(*Object) float64 { t.Fatal("poison used fire protection"); return 0 },
					ElectricProtection: func(*Object) float64 { t.Fatal("poison used electric protection"); return 0 },
					ShieldReduce:       func(*Object, *int32, object.DamageType, *Object) { t.Fatal("Shield reduced poison") },
					DefaultDamageSound: func(got, source *Object) {
						if got != target || source != nil || update.Field547 != 2 || update.Field546 != 5 || target.Field131 != 5 {
							t.Fatal("damage-sound target or raw poison marker changed")
						}
						events = append(events, "sound")
					},
					AdjustFieldGuide: func(source, got *Object, damage int32) int32 {
						if source != nil || got != target || damage != amount {
							t.Fatal("field-guide args")
						}
						events = append(events, "field-guide")
						return damage
					},
					DamageClear: func(got *Object, damage int32) {
						if got != target || damage != amount {
							t.Fatal("damage-clear args")
						}
						remaining := max(int32(target.HealthData.Cur)-damage, 0)
						target.HealthData.Cur = uint16(remaining)
						events = append(events, "damage")
					},
					Unsupported: func(reason string, _, _, _ *Object, _ int32, _ object.DamageType) {
						t.Fatalf("source-less poison rejected: %s", reason)
					},
				}
				if !DefaultDamageWorld4E0B30(target, nil, nil, amount, object.DamagePoison, r) {
					t.Fatal("source-less poison returned false")
				}
				if !slices.Equal(events, []string{"sound", "field-guide", "damage"}) ||
					target.HealthData.Cur != uint16(max(20-amount, 0)) || target.Pos132 != (types.Pointf{}) ||
					target.Obj130 != nil || target.Frame134 != 1400 || target.Buffs != before.Buffs ||
					update.Field1 != math.Float32bits(0.25) {
					t.Fatalf("events=%v HP=%d metadata=%d/%d status=%x", events, target.HealthData.Cur, target.Field131, target.Frame134, update.StatusFlags)
				}
				if !update.StatusFlags.Has(object.MonStatusInjured) || update.StatusFlags.Has(object.MonStatusOnFire) {
					t.Fatalf("poison status=%x, want injured without fire", update.StatusFlags)
				}
			})
		}
	}
}

func TestDefaultDamageWorld4E0B30MonsterPoisonEarlyGates(t *testing.T) {
	for _, gate := range []string{"invulnerable", "dead", "no-update", "poison-immune"} {
		t.Run(gate, func(t *testing.T) {
			target := defaultDamagePoisonFixture4E0B30(t, 0x10)
			update := target.UpdateDataMonster()
			switch gate {
			case "invulnerable":
				target.Buffs |= 1 << defaultDamageInvulnerableEnchant4E0B30
			case "dead":
				target.ObjFlags |= object.FlagDead
			case "no-update":
				target.ObjFlags |= object.FlagNoUpdate
			case "poison-immune":
				target.ObjSubClass |= 0x200
				target.HealthData = nil
			}
			audio := 0
			r := DefaultDamageWorldRuntime4E0B30{
				Frame: func() uint32 { return 1400 },
				Audio: func(id int, got *Object) {
					if gate != "invulnerable" || id != 71 || got != target {
						t.Fatal("early gate audio")
					}
					audio++
				},
				DamageClear: func(*Object, int32) { t.Fatal("early gate damaged target") },
				Unsupported: func(reason string, _, _, _ *Object, _ int32, _ object.DamageType) {
					t.Fatalf("early gate rejected: %s", reason)
				},
			}
			if !DefaultDamageWorld4E0B30(target, nil, nil, 1, object.DamagePoison, r) {
				t.Fatal("early gate result")
			}
			wantLatch, wantAudio := uint32(0), 0
			if gate == "invulnerable" {
				wantLatch, wantAudio = 99, 1
			}
			if audio != wantAudio || update.Field547 != wantLatch || update.Field546 != 91 ||
				update.StatusFlags != 0 || target.Pos132 != types.Ptf(31, 47) || target.Frame134 != 77 {
				t.Fatalf("gate=%s audio=%d latch=%d marker=%d", gate, audio, update.Field547, update.Field546)
			}
		})
	}
}

func TestDefaultDamageWorld4E0B30MonsterPoisonNPCDefend(t *testing.T) {
	target := defaultDamagePoisonFixture4E0B30(t, 0x10)
	modifier, freeModifier := alloc.New(ModifierEff{})
	defer freeModifier()
	fn, freeFn := alloc.New(byte(0))
	defer freeFn()
	modifier.Defend76.Fnc = unsafe.Pointer(fn)
	initData, freeInit := alloc.New(ModifierInitData{})
	defer freeInit()
	initData.Modifiers[2] = modifier
	armor, freeArmor := alloc.New(Object{})
	defer freeArmor()
	*armor = Object{ObjClass: object.ClassArmor, ObjFlags: object.FlagEquipped,
		InitData: unsafe.Pointer(initData)}
	target.InvFirstItem = armor
	var events []string
	r := DefaultDamageWorldRuntime4E0B30{
		CanApplyLateDefend: func(got *ModifierEff) bool { return got == modifier },
		ApplyLateDefend: func(got *ModifierEff, item, unit, weapon, source *Object, damage int32, typ object.DamageType) int32 {
			if got != modifier || item != armor || unit != target || weapon != nil || source != nil || damage != 4 || typ != object.DamagePoison || target.UpdateDataMonster().Field547 != 0 {
				t.Fatal("NPC late defend args/order")
			}
			events = append(events, "defend")
			return damage - 1
		},
		DefaultDamageSound: func(*Object, *Object) { events = append(events, "sound") },
		DamageClear: func(got *Object, damage int32) {
			if got != target || damage != 3 {
				t.Fatal("NPC defended damage")
			}
			events = append(events, "damage")
		},
		Unsupported: func(reason string, _, _, _ *Object, _ int32, _ object.DamageType) {
			t.Fatalf("NPC poison defense rejected: %s", reason)
		},
	}
	if !DefaultDamageWorld4E0B30(target, nil, nil, 4, object.DamagePoison, r) || !slices.Equal(events, []string{"defend", "sound", "damage"}) {
		t.Fatalf("events=%v", events)
	}
}

func TestDefaultDamageWorld4E0B30MonsterPoisonAdmissionBoundary(t *testing.T) {
	for _, weaponOnly := range []bool{false, true} {
		t.Run(fmt.Sprintf("weapon-only-%t", weaponOnly), func(t *testing.T) {
			target := defaultDamagePoisonFixture4E0B30(t, 0x10)
			other := &Object{ObjClass: object.ClassMonster, UpdateData: unsafe.Pointer(&MonsterUpdateData{})}
			source, weapon := other, (*Object)(nil)
			if weaponOnly {
				source, weapon = nil, other
			}
			reason := ""
			DefaultDamageWorld4E0B30(target, source, weapon, 1, object.DamagePoison, DefaultDamageWorldRuntime4E0B30{
				GameplayFlag1: func() bool { return true },
				DamageClear:   func(*Object, int32) { t.Fatal("admitted sourced poison") },
				Unsupported:   func(s string, _, _, _ *Object, _ int32, _ object.DamageType) { reason = s },
			})
			if reason != "unsupported monster damage shape" || target.HealthData.Cur != 20 || target.Frame134 != 77 || target.UpdateDataMonster().Field547 != 0 {
				t.Fatalf("reason=%q HP=%d", reason, target.HealthData.Cur)
			}
		})
	}
}
