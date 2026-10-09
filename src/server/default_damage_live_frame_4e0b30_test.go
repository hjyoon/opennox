package server

import (
	"fmt"
	"math"
	"slices"
	"testing"
	"unsafe"

	"github.com/opennox/libs/object"
)

// GAME.EXE reads 84EA04 at use, not on entry to DefaultDamage. The Frame
// service models that raw DWORD; changing it in a callback must not move a
// metadata store, replay attribution, or introduce a read on a skipped path.
func TestDefaultDamageWorld4E0B30LiveFrameAttribution(t *testing.T) {
	for _, stage := range []string{"fire", "electric", "audio", "buff", "late defend"} {
		for _, live := range []uint32{0, 0x80000004, math.MaxUint32} {
			t.Run(fmt.Sprintf("%s/%#x", stage, live), func(t *testing.T) {
				target, source := damageMeleeUnitFixture4E17B0(t, false), damageMeleeUnitFixture4E17B0(t, true)
				target.Field131, target.Frame134 = 99, 77
				weapon := &Object{TypeInd: 777, ObjClass: object.ClassWand, InitData: unsafe.Pointer(&ModifierInitData{})}
				typ, frame := object.DamageBlade, uint32(1400)
				r := damageMeleeWorldRuntime4E0B30(t)
				r.Frame = func() uint32 { return frame }
				changed := false
				change := func() { frame, changed = live, true }
				r.FireProtection = func(*Object) float64 { change(); return 0 }
				r.ElectricProtection = func(*Object) float64 {
					if stage == "electric" {
						change()
					}
					return 0.25
				}
				r.Audio = func(int, *Object) { change() }
				r.BuffOff = func(*Object, EnchantID) {
					if stage == "buff" {
						change()
					}
				}
				switch stage {
				case "fire":
					weapon.ObjClass |= object.ClassMissile
					typ = object.DamageExplosion
				case "electric", "audio":
					weapon, typ = nil, object.DamageElectric
				case "late defend":
					target.InvFirstItem = &Object{ObjFlags: object.FlagEquipped, InitData: unsafe.Pointer(&ModifierInitData{Modifiers: [4]*ModifierEff{2: defaultDamageLateEffectFixture4E0B30()}})}
					r.CanApplyLateDefend = func(*ModifierEff) bool { return true }
					r.ApplyLateDefend = func(_ *ModifierEff, _, _, _, _ *Object, d int32, _ object.DamageType) int32 {
						if target.Frame134 != 77 || target.Field131 != 99 {
							t.Fatal("attribution preceded late Defend")
						}
						change()
						return d
					}
				}
				r.DamageClear = func(*Object, int32) {
					want := source
					if weapon != nil {
						want = weapon
					}
					if !changed || target.Obj130 != want || target.Field131 != uint32(typ) || target.Frame134 != live || !target.UpdateDataMonster().StatusFlags.Has(object.MonStatusInjured) {
						t.Fatalf("live attribution: changed=%t frame=%#x want=%#x", changed, target.Frame134, live)
					}
				}
				if !DefaultDamageWorld4E0B30(target, source, weapon, 8, typ, r) {
					t.Fatal("attribution tail")
				}
			})
		}
	}
}

func TestDefaultDamageWorld4E0B30LiveFrameProtectionCadence(t *testing.T) {
	for _, electric := range []bool{false, true} {
		for _, tc := range []struct {
			name        string
			entry, live uint32
			protection  float64
			audio       bool
		}{
			{"leaves sound frame", 8, 9, 0.25, false},
			{"enters sound frame", 9, 12, 0.25, true},
			{"wrap to zero", math.MaxUint32, 0, 0.25, true},
			{"high bits preserved", 9, 0x80000004, 0.25, true},
			{"zero protection skips cadence read", 8, 12, 0, false},
		} {
			t.Run(fmt.Sprintf("electric-%t/%s", electric, tc.name), func(t *testing.T) {
				target, source := &Object{ObjClass: object.ClassObstacle}, &Object{}
				r, frame := damageMeleeWorldRuntime4E0B30(t), tc.entry
				var events []string
				r.Frame = func() uint32 { events = append(events, "frame"); return frame }
				protect := func(*Object) float64 { events = append(events, "protection"); frame = tc.live; return tc.protection }
				typ, sound := object.DamageFlame, 104
				r.FireProtection = protect
				if electric {
					typ, sound, r.ElectricProtection = object.DamageElectric, 108, protect
				}
				r.Audio = func(id int, owner *Object) {
					if id != sound || owner != target {
						t.Fatal("protection sound identity")
					}
					events = append(events, "audio")
					frame = 99 // Attribution reads after Audio, not before it.
				}
				r.DamageClear = func(*Object, int32) { events = append(events, "HP") }
				DefaultDamageWorld4E0B30(target, source, nil, 8, typ, r)
				want := []string{"protection"}
				if tc.protection != 0 {
					want = append(want, "frame")
				}
				wantFrame := tc.live
				if tc.audio {
					want, wantFrame = append(want, "audio"), 99
				}
				want = append(want, "frame", "HP")
				if !slices.Equal(events, want) || target.Frame134 != wantFrame {
					t.Fatalf("cadence events=%v want=%v frame=%#x want=%#x", events, want, target.Frame134, wantFrame)
				}
			})
		}
	}
}

func TestDefaultDamageWorld4E0B30LiveFrameSkippedPaths(t *testing.T) {
	for _, path := range []string{"nil", "dead non-zombie", "no update", "poison immune", "friendly", "unsupported"} {
		t.Run(path, func(t *testing.T) {
			target, source := damageMeleeUnitFixture4E17B0(t, false), damageMeleeUnitFixture4E17B0(t, true)
			r, typ := damageMeleeWorldRuntime4E0B30(t), object.DamageClaw
			r.Frame = func() uint32 { t.Error("skipped path read Frame"); return 1400 }
			r.IsZombie = func(*Object) bool { return false }
			r.DamageClear = func(*Object, int32) { t.Error("skipped path applied HP") }
			switch path {
			case "nil":
				target = nil
			case "dead non-zombie":
				target.ObjFlags |= object.FlagDead
			case "no update":
				target.ObjFlags |= object.FlagNoUpdate
			case "poison immune":
				target.ObjSubClass |= 0x200
				source, typ = nil, object.DamagePoison
			case "friendly":
				r.GameplayFlag1 = func() bool { return false }
				r.IsEnemy = func(*Object, *Object) bool { return false }
			case "unsupported":
				typ = object.DamagePlasma
				// Unit-owned weaponless Plasma is restored. Keep this skipped
				// path on the still-unported missile-source Plasma shape.
				source.ObjClass = object.ClassMissile
				r.Unsupported = func(string, *Object, *Object, *Object, int32, object.DamageType) {}
			}
			DefaultDamageWorld4E0B30(target, source, nil, 8, typ, r)
		})
	}
}

func TestDefaultDamageWorld4E0B30LiveFrameDeadZombie(t *testing.T) {
	for _, armed := range []bool{false, true} {
		for _, live := range []uint32{0, 0x80000004, math.MaxUint32} {
			t.Run(fmt.Sprintf("armed-%t/%#x", armed, live), func(t *testing.T) {
				target, source := &Object{ObjClass: object.ClassObstacle, ObjFlags: object.FlagDead, Frame134: 77}, &Object{}
				var weapon *Object
				if armed {
					weapon = &Object{}
				}
				r, frame := damageMeleeWorldRuntime4E0B30(t), uint32(1400)
				var events []string
				r.IsZombie = func(*Object) bool { events = append(events, "zombie"); frame = live; return true }
				r.Frame = func() uint32 { events = append(events, "frame"); return frame }
				r.DamageClear = func(*Object, int32) { t.Fatal("dead zombie HP") }
				DefaultDamageWorld4E0B30(target, source, weapon, 8, object.DamageClaw, r)
				want := source
				if weapon != nil {
					want = weapon
				}
				if !slices.Equal(events, []string{"zombie", "frame"}) || target.Frame134 != live || target.Obj130 != want || target.Field131 != uint32(object.DamageClaw) {
					t.Fatalf("zombie events=%v frame=%#x want=%#x", events, target.Frame134, live)
				}
			})
		}
	}
}

func TestDefaultDamageWorld4E0B30LiveFrameMonsterHit(t *testing.T) {
	for _, owned := range []bool{false, true} {
		for _, stage := range []string{"pre damage", "sound", "guide", "hit enemy"} {
			for _, latch := range []uint32{0, 77, math.MaxUint32} {
				t.Run(fmt.Sprintf("owned-%t/%s/latch-%#x", owned, stage, latch), func(t *testing.T) {
					target, monster := damageMeleeUnitFixture4E17B0(t, false), damageMeleeUnitFixture4E17B0(t, false)
					source := monster
					if owned {
						source = damageMeleeUnitFixture4E17B0(t, true)
						source.ObjOwner = monster
					}
					monster.UpdateDataMonster().Field130 = latch
					weapon := &Object{TypeInd: 777, ObjClass: object.ClassWand, InitData: unsafe.Pointer(&ModifierInitData{Modifiers: [4]*ModifierEff{defaultDamagePreEffectFixture4E0B30()}})}
					r, frame, reads := damageMeleeWorldRuntime4E0B30(t), uint32(1400), 0
					r.Frame = func() uint32 { reads++; return frame }
					r.BuffOff = func(*Object, EnchantID) { frame = 1401 }
					changed := false
					change := func() { frame, changed = math.MaxUint32, true }
					r.CanApplyPreDamage = func(*ModifierEff) bool { return true }
					r.ApplyPreDamage = func(*ModifierEff, *Object, *Object, *Object, *int32) {
						if stage == "pre damage" {
							change()
						}
					}
					r.DefaultDamageSound = func(*Object, *Object) {
						if stage == "sound" {
							change()
						}
					}
					r.AdjustFieldGuide = func(_, _ *Object, d int32) int32 {
						if stage == "guide" {
							change()
						}
						return d
					}
					r.IsEnemy = func(_, m *Object) bool {
						if stage == "hit enemy" && m == monster && target.Obj130 == weapon {
							change()
						}
						return true
					}
					DefaultDamageWorld4E0B30(target, source, weapon, 3, object.DamageBlade, r)
					wantLatch, wantReads := latch, 1
					if latch == 0 {
						wantLatch, wantReads = math.MaxUint32, 2
					}
					if !changed || target.Frame134 != 1401 || monster.UpdateDataMonster().Field130 != wantLatch || reads != wantReads || target.HealthData.Cur != 197 {
						t.Fatalf("hit frame: changed=%t target=%#x hit=%#x want=%#x reads=%d want=%d HP=%d", changed, target.Frame134, monster.UpdateDataMonster().Field130, wantLatch, reads, wantReads, target.HealthData.Cur)
					}
				})
			}
		}
	}
}

func TestDefaultDamageWorld4E0B30LiveFrameFaultPrefix(t *testing.T) {
	for _, zombie := range []bool{false, true} {
		t.Run(fmt.Sprintf("zombie-%t", zombie), func(t *testing.T) {
			target, source := &Object{ObjClass: object.ClassObstacle, Field131: 99, Frame134: 77}, damageMeleeUnitFixture4E17B0(t, true)
			r, buff, dead := damageMeleeWorldRuntime4E0B30(t), false, false
			if zombie {
				target.ObjFlags = object.FlagDead
			}
			r.IsZombie = func(*Object) bool { dead = true; return true }
			r.BuffOff = func(*Object, EnchantID) { buff = true }
			r.Frame = func() uint32 {
				if target.Obj130 != source || target.Field131 != uint32(object.DamageClaw) || target.Frame134 != 77 || (zombie && !dead) || (!zombie && (!buff || target.Pos132 != source.PrevPos)) {
					t.Error("Frame fault preceded its original attribution prefix")
				}
				panic("frame fault")
			}
			r.DamageClear = func(*Object, int32) { t.Error("HP after Frame fault") }
			defer func() {
				if v := recover(); v != "frame fault" {
					t.Fatalf("fault=%v", v)
				}
				if target.Obj130 != source || target.Field131 != uint32(object.DamageClaw) || target.Frame134 != 77 {
					t.Error("Frame fault lost the committed attribution/type prefix")
				}
			}()
			DefaultDamageWorld4E0B30(target, source, nil, 3, object.DamageClaw, r)
		})
	}
}

func TestDefaultDamageWorld4E0B30LiveFrameMonsterHitFaultPrefix(t *testing.T) {
	target, source := damageMeleeUnitFixture4E17B0(t, false), damageMeleeUnitFixture4E17B0(t, false)
	r, reads, sound, guide := damageMeleeWorldRuntime4E0B30(t), 0, false, false
	r.Frame = func() uint32 {
		reads++
		if reads == 1 {
			return 0x80000004
		}
		if !sound || !guide || target.Obj130 != source || target.Frame134 != 0x80000004 || source.UpdateDataMonster().Field130 != 0 || target.HealthData.Cur != 200 {
			t.Error("successful-hit Frame read preceded its original callback/metadata prefix")
		}
		panic("hit frame fault")
	}
	r.DefaultDamageSound = func(*Object, *Object) { sound = true }
	r.AdjustFieldGuide = func(_, _ *Object, d int32) int32 { guide = true; return d }
	r.DamageClear = func(*Object, int32) { t.Error("HP after successful-hit Frame fault") }
	defer func() {
		if v := recover(); v != "hit frame fault" || reads != 2 || source.UpdateDataMonster().Field130 != 0 || target.Frame134 != 0x80000004 || target.HealthData.Cur != 200 {
			t.Fatalf("hit Frame fault=%v reads=%d target=%#x hit=%#x HP=%d", v, reads, target.Frame134, source.UpdateDataMonster().Field130, target.HealthData.Cur)
		}
	}()
	DefaultDamageWorld4E0B30(target, source, nil, 3, object.DamageClaw, r)
}

func TestDefaultDamageWorld4E0B30LiveFrameNilService(t *testing.T) {
	target, source := damageMeleeUnitFixture4E17B0(t, false), damageMeleeUnitFixture4E17B0(t, false)
	r := damageMeleeWorldRuntime4E0B30(t)
	r.Frame = nil
	DefaultDamageWorld4E0B30(target, source, nil, 3, object.DamageClaw, r)
	if target.Frame134 != 0 || source.UpdateDataMonster().Field130 != 0 || target.HealthData.Cur != 197 {
		t.Fatal("nil Frame service must retain the native-runtime default of zero")
	}
}
