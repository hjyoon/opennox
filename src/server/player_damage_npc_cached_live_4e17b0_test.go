package server

import (
	"fmt"
	"math"
	"slices"
	"testing"
	"unsafe"

	"github.com/opennox/libs/object"
	"github.com/opennox/libs/types"

	"github.com/opennox/opennox/v1/common/unit/ai"
)

// 004E1898 caches absorption before direction callbacks, while 004E20F0
// reloads UpdateData for carry. 004E2180 independently reloads the armor
// denominator; its callback must still see the cached hit marker.
func TestPlayerDamageNPCCachedLive4E17B0Crush(t *testing.T) {
	for _, scripted := range []bool{false, true} {
		for _, replace := range []bool{false, true} {
			t.Run(fmt.Sprintf("scripted-%t/replace-%t", scripted, replace), func(t *testing.T) {
				target, source, r := npcChargeFixture4E17B0(t)
				weapon := source
				if scripted {
					source = damageMeleeUnitFixture4E17B0(t, false)
					source.ObjSubClass = 0
					// This excluded 004E1400 weapon bit takes the scripted NPC
					// tail, not the distinct ordinary-melee path.
					weapon = &Object{TypeInd: 65000, ObjClass: object.ClassWeapon, ObjSubClass: 2, InitData: unsafe.Pointer(&ModifierInitData{})}
				}
				old := target.UpdateDataMonster()
				old.Field518, old.Field1, old.Field547, old.Field546 = math.Float32bits(0.5), math.Float32bits(0.4), 99, 77
				old.ArmorEquipFlags, old.AIStack[0].Action = 0x1000000, uint32(ai.ACTION_BLOCK_ATTACK)
				live := old
				if replace {
					live = &MonsterUpdateData{Field547: 7, Field546: 88}
				}
				var events []string
				r.BlockSourceExcluded = func(*Object) bool { return false }
				r.BlockDirection = func(v *Object, p types.Pointf) bool {
					if v != target || p != weapon.PrevPos {
						t.Fatal("CRUSH defense did not use prior attack position")
					}
					events = append(events, "direction")
					old.Field518 = math.Float32bits(0.75)
					live.Field518, live.Field1 = math.Float32bits(1), math.Float32bits(0.25)
					target.UpdateData = unsafe.Pointer(live)
					return false
				}
				carry := damageArmorCarryFixture4E17B0(0)
				armor := &Object{ObjClass: object.ClassArmor, ObjFlags: object.FlagEquipped, Damage: unsafe.Pointer(new(byte)),
					HealthData: &HealthData{Cur: 50, Max: 50}, UpdateData: unsafe.Pointer(carry), InitData: unsafe.Pointer(&ModifierInitData{})}
				target.InvFirstItem = armor
				r.ItemArmorValue = func(v *Object) float32 {
					if v != armor || live.Field1 != 0 {
						t.Fatal("armor lookup preceded live carry store")
					}
					events = append(events, "armor-value")
					return 0.5
				}
				r.CanDamageArmor = func(v *Object) bool { return v == armor }
				r.DamageArmor = func(v, a, w *Object, d int32, typ object.DamageType) bool {
					marker, markerType := uint32(0), uint32(77)
					if scripted {
						marker, markerType = 1, uint32(weapon.TypeInd)
					}
					if v != armor || a != source || w != weapon || d != 1 || typ != object.DamageCrush || old.Field547 != marker || old.Field546 != markerType {
						t.Fatalf("cached absorption/marker and live armor: wear=%d marker=%d/%d", d, old.Field547, old.Field546)
					}
					events = append(events, "armor")
					v.HealthData.Cur--
					return true
				}
				r.QuestMode = func() bool { events = append(events, "quest-mode"); return false }
				r.QuestDamageScale = func() float32 { t.Fatal("non-Quest scale"); return 0 }
				world := damageMeleeWorldRuntime4E0B30(t)
				r.DefaultDamage = func(v, a, w *Object, d int32, typ object.DamageType) bool {
					if d != 7 {
						t.Fatalf("cached armor + live carry damage=%d, want 7", d)
					}
					events = append(events, "default")
					return DefaultDamageWorld4E0B30(v, a, w, d, typ, world)
				}
				if h, result := PlayerDamageNative4E17B0(target, source, weapon, 9, object.DamageCrush, r); !h || !result {
					t.Fatalf("CRUSH=%t/%t", h, result)
				}
				if !slices.Equal(events, []string{"direction", "armor-value", "armor", "quest-mode", "default"}) || target.HealthData.Cur != 193 ||
					armor.HealthData.Cur != 49 || *carry != 0 || live.Field1 != 0 || source.HealthData.Cur != 200 {
					t.Fatalf("events=%v HP=%d armor=%d carry=%g", events, target.HealthData.Cur, armor.HealthData.Cur, math.Float32frombits(live.Field1))
				}
				if replace && (old.Field1 != math.Float32bits(0.4) || live.Field547 != map[bool]uint32{false: 2, true: 1}[scripted]) {
					t.Fatal("cached carry was written or the shared tail did not use live metadata")
				}
			})
		}
	}
}

// 004E1DF2 calls electric armor first; 004E20F0 then reads AND writes the
// newly live carry. Electric armor wear uses raw damage, not raw minus HP.
func TestPlayerDamageNPCCachedLive4E17B0Electric(t *testing.T) {
	for _, typ := range []object.DamageType{object.DamageElectric, object.DamageAirborneElectric} {
		for _, replace := range []bool{false, true} {
			t.Run(fmt.Sprintf("type-%d/replace-%t", typ, replace), func(t *testing.T) {
				target, source, r := npcChargeFixture4E17B0(t)
				old := target.UpdateDataMonster()
				old.Field518, old.Field1, old.Field547, old.Field546 = math.Float32bits(0.5), math.Float32bits(-0.25), 99, 77
				live := old
				if replace {
					live = &MonsterUpdateData{Field547: 7, Field546: 88}
				}
				var events []string
				r.ElectricArmorScale = func(v *Object) float32 {
					if v != target || old.Field547 != 0 {
						t.Fatal("electric armor preceded cached marker reset")
					}
					events = append(events, "electric")
					live.Field518, live.Field1 = math.Float32bits(1), math.Float32bits(0.25)
					target.UpdateData = unsafe.Pointer(live)
					return 0.5
				}
				carry := damageArmorCarryFixture4E17B0(0)
				armor := &Object{ObjClass: object.ClassArmor, ObjFlags: object.FlagEquipped, Damage: unsafe.Pointer(new(byte)),
					HealthData: &HealthData{Cur: 50, Max: 50}, UpdateData: unsafe.Pointer(carry), InitData: unsafe.Pointer(&ModifierInitData{})}
				target.InvFirstItem = armor
				r.ItemArmorValue = func(*Object) float32 { events = append(events, "armor-value"); return 1 }
				r.CanDamageArmor = func(v *Object) bool { return v == armor }
				r.DamageArmor = func(v, a, w *Object, d int32, gotType object.DamageType) bool {
					if v != armor || a != source || w != source || d != 3 || gotType != typ || live.Field1 != math.Float32bits(-0.25) || old.Field547 != 0 || old.Field546 != 77 {
						t.Fatalf("electric raw wear/carry/marker=%d/%g/%d/%d", d, math.Float32frombits(live.Field1), old.Field547, old.Field546)
					}
					events = append(events, "armor")
					v.HealthData.Cur -= 3
					return true
				}
				r.QuestMode = func() bool { events = append(events, "quest-mode"); return false }
				r.QuestDamageScale = func() float32 { t.Fatal("non-Quest scale"); return 0 }
				world := damageMeleeWorldRuntime4E0B30(t)
				world.ElectricProtection = func(*Object) float64 { return 0 }
				r.DefaultDamage = func(v, a, w *Object, d int32, gotType object.DamageType) bool {
					if d != 2 || old.Field547 != 2 || old.Field546 != uint32(typ) {
						t.Fatalf("electric damage/marker=%d/%d/%d", d, old.Field547, old.Field546)
					}
					events = append(events, "default")
					return DefaultDamageWorld4E0B30(v, a, w, d, gotType, world)
				}
				if h, result := PlayerDamageNative4E17B0(target, source, source, 3, typ, r); !h || !result ||
					target.HealthData.Cur != 198 || armor.HealthData.Cur != 47 || *carry != 0 ||
					!slices.Equal(events, []string{"electric", "armor-value", "armor", "quest-mode", "default"}) {
					t.Fatalf("electric=%t/%t HP=%d armor=%d events=%v", h, result, target.HealthData.Cur, armor.HealthData.Cur, events)
				}
				if replace && (old.Field1 != math.Float32bits(-0.25) || live.Field547 != 2 || live.Field546 != uint32(typ)) {
					t.Fatal("electric wrote stale carry/live tail metadata")
				}
			})
		}
	}
}

func TestPlayerDamageNPCCachedLive4E17B0QuestAfterWear(t *testing.T) {
	for _, typ := range []object.DamageType{object.DamageCrush, object.DamageElectric, object.DamageAirborneElectric} {
		for _, entering := range []bool{false, true} {
			t.Run(fmt.Sprintf("type-%d/entering-%t", typ, entering), func(t *testing.T) {
				target, source, r := npcChargeFixture4E17B0(t)
				armor := damageMeleeArmorFixture4E17B0(target, 0.5, 0)
				damageMeleeArmorRuntime4E17B0(&r, armor, 0.5)
				quest := !entering
				var events []string
				r.ElectricArmorScale = func(*Object) float32 { return 0.5 }
				r.DamageArmor = func(v, a, w *Object, _ int32, gotType object.DamageType) bool {
					if v != armor || a != source || w != source || gotType != typ {
						t.Fatal("Quest armor identity")
					}
					events = append(events, "armor")
					quest = entering
					return true
				}
				r.QuestMode = func() bool { events = append(events, "quest-mode"); return quest }
				r.QuestDamageScale = func() float32 { events = append(events, "quest-scale"); return 0.5 }
				r.DefaultDamage = func(v, a, w *Object, d int32, gotType object.DamageType) bool {
					want := int32(2)
					if typ == object.DamageCrush {
						want = 3
					}
					if entering {
						want = playerDamageRound4E17B0(float32(want) * 0.5)
					}
					if v != target || a != source || w != source || d != want || gotType != typ {
						t.Fatalf("live Quest damage=%d, want %d", d, want)
					}
					events = append(events, "default")
					return false
				}
				if h, result := PlayerDamageNative4E17B0(target, source, source, 4, typ, r); !h || result {
					t.Fatal("Quest default return propagation")
				}
				want := []string{"armor", "quest-mode", "default"}
				if entering {
					want = []string{"armor", "quest-mode", "quest-scale", "default"}
				}
				if !slices.Equal(events, want) {
					t.Fatalf("Quest ordering=%v, want %v", events, want)
				}
			})
		}
	}
}

func TestPlayerDamageNPCCachedLive4E17B0ElectricNilCarryPrefix(t *testing.T) {
	target, source, r := npcChargeFixture4E17B0(t)
	old := target.UpdateDataMonster()
	old.Field1, old.Field547, old.Field546 = math.Float32bits(0.25), 99, 77
	r.ElectricArmorScale = func(*Object) float32 { target.UpdateData = nil; return 0.5 }
	r.QuestMode = func() bool { t.Fatal("Quest read before live carry fault"); return false }
	r.QuestDamageScale = func() float32 { t.Fatal("Quest scale before live carry fault"); return 0 }
	r.DefaultDamage = func(*Object, *Object, *Object, int32, object.DamageType) bool {
		t.Fatal("DefaultDamage before live carry fault")
		return false
	}
	defer func() {
		if recover() == nil {
			t.Fatal("stale carry hid nil live UpdateData")
		}
		if old.Field1 != math.Float32bits(0.25) || old.Field547 != 0 || old.Field546 != 77 || target.HealthData.Cur != 200 {
			t.Fatal("nil carry fault lost the prior cached marker prefix or changed carry/HP")
		}
	}()
	PlayerDamageNative4E17B0(target, source, source, 3, object.DamageElectric, r)
}
