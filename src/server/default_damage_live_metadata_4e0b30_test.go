package server

import (
	"math"
	"slices"
	"testing"
	"unsafe"

	"github.com/opennox/libs/object"
	"github.com/opennox/libs/types"
)

// The entry reset at 004E0B86 must not turn its update address into an
// execution plan for the three independently reloaded metadata stores.
func TestDefaultDamageWorld4E0B30LiveMetadataElectric(t *testing.T) {
	for _, tc := range []struct {
		name        string
		entry, live object.Class
		npc         bool
	}{
		{"NPC replaces update", object.ClassMonster, object.ClassMonster, true},
		{"NPC becomes player", object.ClassMonster, object.ClassPlayer, false},
		{"player becomes NPC", object.ClassPlayer, object.ClassMonster, true},
		{"world becomes NPC", object.ClassObstacle, object.ClassMonster, true},
		{"world becomes player", object.ClassObstacle, object.ClassPlayer, false},
		{"PLAYER wins combined class", object.ClassObstacle, object.ClassPlayer | object.ClassMonster, true},
		{"NPC loses subclass", object.ClassMonster, object.ClassMonster, false},
		{"NPC leaves units", object.ClassMonster, object.ClassObstacle, false},
		{"player leaves units", object.ClassPlayer, object.ClassObstacle, false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			defer func() {
				if v := recover(); v != nil {
					t.Fatalf("electric state used the wrong live update layout: %v", v)
				}
			}()
			target, source := damageMeleeUnitFixture4E17B0(t, tc.entry == object.ClassPlayer), damageMeleeUnitFixture4E17B0(t, true)
			oldM, oldP := &MonsterUpdateData{Field523_2: 0x33}, &PlayerUpdateData{Field40_0: 0x3344}
			target.ObjClass = tc.entry
			if tc.entry == object.ClassMonster {
				target.UpdateData = unsafe.Pointer(oldM)
			} else {
				target.UpdateData = unsafe.Pointer(oldP)
			}
			liveM := &MonsterUpdateData{Field523_2: 0x55, Field523_3: 0xa5}
			liveP := &PlayerUpdateData{Player: &Player{}, Field40_0: 0x5566, Field40_1: 0xabcd}
			r := damageMeleeWorldRuntime4E0B30(t)
			var events []string
			r.ElectricProtection = func(*Object) float64 {
				events = append(events, "protection")
				target.ObjClass, target.ObjSubClass, target.UpdateData = tc.live, 0, nil
				if tc.npc {
					target.ObjSubClass = 0x10
				}
				if tc.live.Has(object.ClassPlayer) {
					target.UpdateData = unsafe.Pointer(liveP)
				} else if tc.live.Has(object.ClassMonster) && tc.npc {
					target.UpdateData = unsafe.Pointer(liveM)
				}
				return 0.25
			}
			r.BuffOff = func(*Object, EnchantID) {
				wantM, wantP := uint8(0x55), uint16(0x5566)
				if tc.live.Has(object.ClassPlayer) {
					wantP = 2
				} else if tc.live.Has(object.ClassMonster) && tc.npc {
					wantM = 2
				}
				if oldM.Field523_2 != 0x33 || oldP.Field40_0 != 0x3344 || liveM.Field523_2 != wantM || liveM.Field523_3 != 0xa5 || liveP.Field40_0 != wantP || liveP.Field40_1 != 0xabcd || target.Pos132 != source.PrevPos {
					t.Fatal("electric state used an entry class/address or overwrote adjacent bytes")
				}
				events = append(events, "buff")
				// End this test's electric observation before the independent
				// injured store. A combined class need not share update layouts.
				target.ObjClass, target.UpdateData = object.ClassObstacle, nil
			}
			r.DamageClear = func(*Object, int32) { events = append(events, "HP") }
			if !DefaultDamageWorld4E0B30(target, source, nil, 8, object.DamageElectric, r) || !slices.Equal(events, []string{"protection", "buff", "HP"}) {
				t.Fatalf("electric order=%v", events)
			}
		})
	}
}

func TestDefaultDamageWorld4E0B30LiveMetadataAttackTypeBeforeBuff(t *testing.T) {
	for _, armed := range []bool{false, true} {
		for _, tc := range []struct {
			name        string
			entry, live object.Class
		}{
			{"NPC new update", object.ClassMonster, object.ClassMonster},
			{"player becomes monster", object.ClassPlayer, object.ClassMonster},
			{"world becomes monster", object.ClassObstacle, object.ClassMonster},
			{"monster leaves class", object.ClassMonster, object.ClassObstacle},
		} {
			t.Run(tc.name+map[bool]string{false: " unarmed", true: " armed"}[armed], func(t *testing.T) {
				target, source := damageMeleeUnitFixture4E17B0(t, tc.entry == object.ClassPlayer), damageMeleeUnitFixture4E17B0(t, true)
				old := &MonsterUpdateData{Field547: 99, Field546: 0xaabbccdd}
				target.ObjClass = tc.entry
				if tc.entry == object.ClassMonster {
					target.UpdateData = unsafe.Pointer(old)
				}
				live := &MonsterUpdateData{Field546: 0x11223344}
				var weapon *Object
				typ := object.DamageClaw
				wantType := uint32(0xfedd)
				if armed {
					weapon = &Object{TypeInd: 0xfedc, ObjClass: object.ClassWeapon, InitData: unsafe.Pointer(&ModifierInitData{})}
					typ, wantType = object.DamageBlade, 0xfedc
				}
				r := damageMeleeWorldRuntime4E0B30(t)
				r.GameplayFlag1 = func() bool { return false }
				changed, seen := false, false
				r.IsEnemy = func(*Object, *Object) bool {
					if !changed {
						changed = true
						target.ObjClass, target.UpdateData = tc.live, nil
						if tc.live.Has(object.ClassMonster) {
							target.UpdateData = unsafe.Pointer(live)
						}
						source.TypeInd = 0xfedd
					}
					return true
				}
				r.BuffOff = func(*Object, EnchantID) {
					wantLatch, wantKind := uint32(0), uint32(0x11223344)
					if tc.live.Has(object.ClassMonster) {
						wantLatch, wantKind = 1, wantType
					}
					wantOld := uint32(99)
					if tc.entry.Has(object.ClassMonster) {
						wantOld = 0
					}
					if !changed || live.Field547 != wantLatch || live.Field546 != wantKind || old.Field547 != wantOld || old.Field546 != 0xaabbccdd || target.Pos132 != source.PrevPos || target.Frame134 != 0 {
						t.Fatal("attack type did not use the live Monster update before BuffOff/attribution")
					}
					seen = true
					target.ObjClass, target.UpdateData = object.ClassObstacle, nil
				}
				r.DamageClear = func(*Object, int32) {}
				if !DefaultDamageWorld4E0B30(target, source, weapon, 3, typ, r) || !seen {
					t.Fatal("attack-type prefix was not observed")
				}
			})
		}
	}
}

func TestDefaultDamageWorld4E0B30LiveMetadataInjuredUpdate(t *testing.T) {
	for _, stage := range []string{"protection", "buff", "late defend"} {
		for _, latch := range []uint32{0, 7, math.MaxUint32} {
			t.Run(stage+"/"+map[uint32]string{0: "empty", 7: "nonzero", math.MaxUint32: "all bits"}[latch], func(t *testing.T) {
				target, source := damageMeleeUnitFixture4E17B0(t, false), damageMeleeUnitFixture4E17B0(t, true)
				old := target.UpdateDataMonster()
				old.StatusFlags = 0x40000000
				live := &MonsterUpdateData{StatusFlags: 0x20000000, Field547: latch, Field546: 0x11223344}
				weapon := &Object{TypeInd: 777, ObjClass: object.ClassWand, InitData: unsafe.Pointer(&ModifierInitData{})}
				typ := object.DamageBlade
				if stage == "protection" {
					weapon.ObjClass |= object.ClassMissile
					typ = object.DamageExplosion
				}
				target.Field131, target.Frame134 = 99, 77
				r := damageMeleeWorldRuntime4E0B30(t)
				r.FireProtection = func(*Object) float64 { target.UpdateData = unsafe.Pointer(live); return 0 }
				r.BuffOff = func(*Object, EnchantID) {
					if stage == "buff" {
						target.UpdateData = unsafe.Pointer(live)
					}
				}
				if stage == "late defend" {
					target.InvFirstItem = &Object{ObjFlags: object.FlagEquipped, InitData: unsafe.Pointer(&ModifierInitData{Modifiers: [4]*ModifierEff{2: defaultDamageLateEffectFixture4E0B30()}})}
					r.CanApplyLateDefend = func(*ModifierEff) bool { return true }
					r.ApplyLateDefend = func(_ *ModifierEff, _, _, _, _ *Object, d int32, _ object.DamageType) int32 {
						if target.Field131 != 99 || target.Frame134 != 77 || old.Field547 != 1 {
							t.Fatal("late Defend preceded attack marker or followed attribution")
						}
						target.UpdateData = unsafe.Pointer(live)
						return d
					}
				}
				weapon.InitData = unsafe.Pointer(&ModifierInitData{Modifiers: [4]*ModifierEff{defaultDamagePreEffectFixture4E0B30()}})
				r.CanApplyPreDamage = func(*ModifierEff) bool { return true }
				r.ApplyPreDamage = func(_ *ModifierEff, _, _, owner *Object, d *int32) {
					wantLatch, wantKind := latch, uint32(0x11223344)
					if stage == "protection" {
						wantLatch, wantKind = 1, 777
					} else if latch == 0 {
						wantLatch, wantKind = 2, uint32(typ)
					}
					if owner != target || *d != 3 || !live.StatusFlags.Has(object.MonStatusInjured) || live.StatusFlags != 0x20000000|object.MonStatusInjured || old.StatusFlags.Has(object.MonStatusInjured) || live.Field547 != wantLatch || live.Field546 != wantKind || target.Obj130 != weapon || target.Field131 != uint32(typ) || target.Frame134 != 1400 {
						t.Fatalf("injured record or prefix: old=%#x live=%#x latch=%#x kind=%#x", uint32(old.StatusFlags), uint32(live.StatusFlags), live.Field547, live.Field546)
					}
				}
				r.DamageClear = func(*Object, int32) {}
				if !DefaultDamageWorld4E0B30(target, source, weapon, 3, typ, r) {
					t.Fatal("injured tail")
				}
			})
		}
	}
}

func TestDefaultDamageWorld4E0B30LiveMetadataInjuredClass(t *testing.T) {
	for _, tc := range []struct {
		name        string
		entry, live object.Class
	}{
		{"world becomes monster", object.ClassObstacle, object.ClassMonster},
		{"player becomes monster", object.ClassPlayer, object.ClassMonster},
		{"monster leaves units", object.ClassMonster, object.ClassObstacle},
		{"monster becomes player", object.ClassMonster, object.ClassPlayer},
	} {
		t.Run(tc.name, func(t *testing.T) {
			target, source := damageMeleeUnitFixture4E17B0(t, tc.entry == object.ClassPlayer), damageMeleeUnitFixture4E17B0(t, true)
			old := &MonsterUpdateData{StatusFlags: 0x40000000}
			target.ObjClass = tc.entry
			if tc.entry == object.ClassMonster {
				target.UpdateData = unsafe.Pointer(old)
			}
			live := &MonsterUpdateData{StatusFlags: 0x20000000, Field546: 0x11223344}
			r := damageMeleeWorldRuntime4E0B30(t)
			r.BuffOff = func(*Object, EnchantID) {
				target.ObjClass, target.ObjSubClass, target.UpdateData = tc.live, 0, nil
				if tc.live == object.ClassMonster {
					target.UpdateData = unsafe.Pointer(live)
				}
			}
			r.DamageClear = func(*Object, int32) {
				if old.StatusFlags.Has(object.MonStatusInjured) || live.StatusFlags.Has(object.MonStatusInjured) != tc.live.Has(object.ClassMonster) {
					t.Fatal("injured store used entry-time class")
				}
				if tc.live.Has(object.ClassMonster) && (live.Field547 != 2 || live.Field546 != uint32(object.DamageClaw)) {
					t.Fatal("new Monster injury latch")
				}
			}
			if !DefaultDamageWorld4E0B30(target, source, nil, 3, object.DamageClaw, r) {
				t.Fatal("class tail")
			}
		})
	}
}

func TestDefaultDamageWorld4E0B30LiveMetadataNilUpdateFaultPrefixes(t *testing.T) {
	for _, stage := range []string{"electric", "attack type", "injured"} {
		t.Run(stage, func(t *testing.T) {
			target, source := damageMeleeUnitFixture4E17B0(t, false), damageMeleeUnitFixture4E17B0(t, true)
			old := target.UpdateDataMonster()
			old.Field523_2, old.Field546, old.StatusFlags = 0x55, 0x11223344, 0x40000000
			target.Pos132, target.Field131, target.Frame134 = types.Ptf(9, 10), 99, 77
			weapon := &Object{TypeInd: 777, ObjClass: object.ClassWeapon, InitData: unsafe.Pointer(&ModifierInitData{})}
			typ := object.DamageBlade
			r := damageMeleeWorldRuntime4E0B30(t)
			var events []string
			if stage == "electric" {
				weapon, typ = nil, object.DamageElectric
				r.ElectricProtection = func(*Object) float64 { target.UpdateData = nil; events = append(events, "protection"); return 0 }
			} else if stage == "attack type" {
				r.IsEnemy = func(*Object, *Object) bool { target.UpdateData = nil; events = append(events, "enemy"); return true }
			}
			r.BuffOff = func(*Object, EnchantID) {
				events = append(events, "buff")
				if stage == "injured" {
					target.UpdateData = nil
				}
			}
			r.DefaultDamageSound = func(*Object, *Object) { events = append(events, "sound") }
			r.DamageClear = func(*Object, int32) { events = append(events, "HP") }
			var caught any
			func() {
				defer func() { caught = recover() }()
				DefaultDamageWorld4E0B30(target, source, weapon, 3, typ, r)
			}()
			want := []string{"protection"}
			if stage == "attack type" {
				want = []string{"enemy"}
			} else if stage == "injured" {
				want = []string{"buff"}
			}
			if caught == nil || !slices.Equal(events, want) || old.Field523_2 != 0x55 || old.StatusFlags.Has(object.MonStatusInjured) || target.HealthData.Cur != 200 {
				t.Fatalf("nil at-use fault: stage=%s caught=%v events=%v", stage, caught, events)
			}
			if stage == "electric" && target.Pos132 != (types.Ptf(9, 10)) {
				t.Fatal("electric nil fault followed position store")
			}
			if stage == "attack type" && (target.Pos132 != source.PrevPos || target.Frame134 != 77 || old.Field547 != 0) {
				t.Fatal("attack marker fault prefix")
			}
			if stage == "injured" && (target.Obj130 != weapon || target.Field131 != uint32(typ) || target.Frame134 != 1400 || old.Field547 != 1 || old.Field546 != 777) {
				t.Fatal("injured fault preceded attribution or reused cached update")
			}
		})
	}
}

func TestDefaultDamageWorld4E0B30LiveMetadataStopsBeforePreDamage(t *testing.T) {
	target, source := damageMeleeUnitFixture4E17B0(t, false), damageMeleeUnitFixture4E17B0(t, true)
	old, live := target.UpdateDataMonster(), &MonsterUpdateData{Field547: 77, Field546: 88}
	weapon := &Object{ObjClass: object.ClassWeapon, InitData: unsafe.Pointer(&ModifierInitData{Modifiers: [4]*ModifierEff{defaultDamagePreEffectFixture4E0B30()}})}
	r := damageMeleeWorldRuntime4E0B30(t)
	r.CanApplyPreDamage = func(*ModifierEff) bool { return true }
	r.ApplyPreDamage = func(_ *ModifierEff, _, _, _ *Object, _ *int32) {
		if !old.StatusFlags.Has(object.MonStatusInjured) || old.Field547 != 1 {
			t.Fatal("injury preceded wrong callback")
		}
		target.UpdateData = unsafe.Pointer(live)
	}
	r.DamageClear = func(*Object, int32) {
		if live.StatusFlags.Has(object.MonStatusInjured) || live.Field547 != 77 || live.Field546 != 88 {
			t.Fatal("injured metadata was incorrectly replayed after pre-Damage")
		}
	}
	if !DefaultDamageWorld4E0B30(target, source, weapon, 3, object.DamageBlade, r) {
		t.Fatal("pre-Damage tail")
	}
}
