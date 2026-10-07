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
	"github.com/opennox/opennox/v1/legacy/common/alloc"
)

func TestPlayerDamageNPCCasterImpact4E17B0CachedLiveAndQuest(t *testing.T) {
	for _, owner := range []string{"player-self", "player-owned-NPC", "imaginary-owned-NPC", "world-self"} {
		for _, excluded := range []bool{false, true} {
			for _, replace := range []bool{false, true} {
				for _, entering := range []bool{false, true} {
					t.Run(fmt.Sprintf("%s/excluded-%t/replace-%t/Quest-enter-%t", owner, excluded, replace, entering), func(t *testing.T) {
						v, a, w := defaultDamageCasterImpactFixture4E0B30(t, owner, "NPC")
						armor := worldFlameArmorFixture4E17B0(t, v)
						cached, live := v.UpdateDataMonster(), v.UpdateDataMonster()
						cached.Field518, cached.Field1 = math.Float32bits(0.25), math.Float32bits(0.125)
						cached.Field547, cached.Field546 = 99, 77
						pos := w.PrevPos
						quest := !entering
						var events []string
						r := damageMeleeRuntimeFixture4E17B0(t)
						damageMeleeArmorRuntime4E17B0(&r, armor, 0.875)
						r.Frame = func() uint32 { t.Fatal("early frame query"); return 0 }
						r.BlockSourceExcluded = func(attack *Object) bool {
							if attack != w || cached.Field547 != 0 || cached.Field546 != 77 {
								t.Fatal("cached marker before exclusion")
							}
							events = append(events, "exclude")
							if replace {
								var free func()
								live, free = alloc.New(MonsterUpdateData{})
								t.Cleanup(free)
								*live = MonsterUpdateData{Field547: 55, Field546: 66}
								v.UpdateData = unsafe.Pointer(live)
							}
							live.Field518, live.Field1 = math.Float32bits(0.875), math.Float32bits(0.5)
							w.TypeInd, w.PrevPos = 432, types.Ptf(-100, 99)
							return excluded
						}
						r.BlockDirection = func(target *Object, p types.Pointf) bool {
							marker, kind := uint32(0), uint32(77)
							if a != w {
								marker, kind = 1, 432
							}
							if excluded || target != v || p != pos || cached.Field547 != marker || cached.Field546 != kind {
								t.Fatal("snapshot position/live type after exclusion")
							}
							events = append(events, "facing")
							return false
						}
						r.ItemArmorValue = func(item *Object) float32 {
							if item != armor || live.Field1 != math.Float32bits(0.25) {
								t.Fatal("live carry precedes real armor lookup")
							}
							events = append(events, "lookup")
							return 0.875
						}
						r.DamageArmor = func(item, source, weapon *Object, amount int32, typ object.DamageType) bool {
							marker := uint32(0)
							if a != w {
								marker = 1
							}
							if item != armor || source != a || weapon != w || amount != 1 || typ != object.DamageImpact || cached.Field547 != marker {
								t.Fatal("cached absorption/live armor denominator before Quest")
							}
							events = append(events, "wear")
							item.HealthData.Cur -= uint16(amount)
							cached.Field547 = 0 // An actual armor callback may clear attribution.
							quest = entering
							return true
						}
						r.GodMode = func() bool {
							if cached.Field547 != 2 || cached.Field546 != 11 || armor.HealthData.Cur != 49 {
								t.Fatal("wear and cached fallback precede God")
							}
							events = append(events, "god")
							return true // A live NPC still reaches Quest and DefaultDamage.
						}
						r.QuestMode = func() bool { events = append(events, "quest"); return quest }
						r.QuestDamageScale = func() float32 { events = append(events, "scale"); return 0.5 }
						r.DefaultDamage = func(target, source, weapon *Object, amount int32, typ object.DamageType) bool {
							want := int32(4)
							if entering {
								want = 2
							}
							if target != v || source != a || weapon != w || amount != want || typ != object.DamageImpact {
								t.Fatal("live Quest/default identities")
							}
							events = append(events, "default")
							return true
						}
						if h, result := PlayerDamageNative4E17B0(v, a, w, 5, object.DamageImpact, r); !h || !result {
							t.Fatal("cached/live NPC caster hit rejected")
						}
						want := []string{"exclude"}
						if !excluded {
							want = append(want, "facing")
						}
						want = append(want, "lookup", "wear", "god", "quest")
						if entering {
							want = append(want, "scale")
						}
						want = append(want, "default")
						if !slices.Equal(events, want) || v.HealthData.Cur != 20 || live.Field1 != math.Float32bits(0.25) ||
							(replace && (cached.Field1 != math.Float32bits(0.125) || live.Field547 != 55 || live.Field546 != 66)) {
							t.Fatalf("cached/live events=%v want=%v", events, want)
						}
					})
				}
			}
		}
	}
}

// Mutation cases deliberately enter as a nonmissile caster. They cover the
// live class/subclass reads after exclusion/audio/reflection without broadening
// the public admission to preexisting missile or weapon paths.
func TestPlayerDamageNPCCasterImpact4E17B0LiveDefense(t *testing.T) {
	for _, equipment := range []string{"shield", "sword", "staff"} {
		for _, raw := range []int32{-3, 0, 8} {
			for _, mutation := range []string{"none", "exclude-missile", "audio-missile", "reflect-removes", "reflect-sub2"} {
				t.Run(fmt.Sprintf("%s/raw-%d/%s", equipment, raw, mutation), func(t *testing.T) {
					v, a, w := defaultDamageCasterImpactFixture4E0B30(t, "player-owned-NPC", "NPC")
					w.ObjSubClass = 0 // A mutated missile has an independently selected subclass.
					cached := v.UpdateDataMonster()
					cached.Field547, cached.Field546, cached.Field1 = 99, 77, 0
					mask, sound := uint32(2), 878
					switch equipment {
					case "shield":
						cached.ArmorEquipFlags, cached.AIStack[0].Action = 0x1000000, uint32(ai.ACTION_BLOCK_ATTACK)
					case "sword":
						cached.WeaponEquipFlags, mask, sound = 0x400, 0x400, 890
					case "staff":
						cached.WeaponEquipFlags, mask, sound = 0x8000, 0x8000, 894
					}
					old, freeOld := alloc.New(Object{})
					t.Cleanup(freeOld)
					*old = Object{ObjSubClass: object.SubClass(mask), ObjFlags: object.FlagEquipped}
					// 004E22A0/004E2330 test equipment flag/subclass, not item class.
					item, freeItem := alloc.New(Object{})
					t.Cleanup(freeItem)
					*item = Object{ObjClass: object.ClassSimple, ObjSubClass: object.SubClass(mask), ObjFlags: object.FlagEquipped}
					v.InvFirstItem = old
					live, free := alloc.New(MonsterUpdateData{})
					t.Cleanup(free)
					*live = MonsterUpdateData{Field547: 55, Field546: 66, Field1: math.Float32bits(0.5)}
					initialMissile := mutation == "exclude-missile" || mutation == "reflect-removes" || mutation == "reflect-sub2"
					blocked := equipment != "staff" || !initialMissile
					var events []string
					r := damageMeleeRuntimeFixture4E17B0(t)
					r.BlockSourceExcluded = func(attack *Object) bool {
						if attack != w || cached.Field547 != 0 {
							t.Fatal("defense entry marker/identity")
						}
						events = append(events, "exclude")
						if initialMissile {
							w.ObjClass = object.ClassMissile
						}
						return false
					}
					r.BlockDirection = func(target *Object, p types.Pointf) bool {
						if target != v || p != w.PrevPos || cached.Field547 != 1 || cached.Field546 != uint32(w.TypeInd) {
							t.Fatal("defense distinct type/facing")
						}
						events = append(events, "front")
						return true
					}
					r.Audio = func(id int, target *Object) {
						if !blocked || id != sound || target != v || cached.Field547 != 1 {
							t.Fatal("defense audio")
						}
						events = append(events, "audio")
						v.UpdateData = unsafe.Pointer(live)
						if mutation == "audio-missile" {
							w.ObjClass = object.ClassMissile
						}
					}
					r.ProjectileReflect = func(attack, target *Object) {
						if attack != w || target != v {
							t.Fatal("defense reflection identity")
						}
						events = append(events, "reflect")
						if mutation == "reflect-removes" {
							w.ObjClass = 0
						} else if mutation == "reflect-sub2" {
							w.ObjSubClass = 2
						}
					}
					r.ClearOwner = func(attack *Object) {
						if attack != w {
							t.Fatal("clear owner identity")
						}
						events = append(events, "clear")
					}
					r.SetOwner = func(target, attack *Object) {
						if target != v || attack != w {
							t.Fatal("set owner identity")
						}
						events = append(events, "owner")
					}
					r.Melee.MonsterBlockAction = func(target *Object) {
						if target != v || target.UpdateData != unsafe.Pointer(live) {
							t.Fatal("monster action did not reload live record")
						}
						events = append(events, "action")
					}
					r.BlockDamagePercent = func() float64 { events = append(events, "percent"); v.InvFirstItem = item; return 0.375 }
					admit := func(selected *Object) bool {
						if selected != item {
							t.Fatal("stale block item")
						}
						events = append(events, "admit")
						return true
					}
					r.CanDamageBlockItem, r.Melee.CanDamageBlockWeapon = admit, admit
					wear := func(selected, target, source, attack *Object, amount float32, typ object.DamageType) bool {
						if selected != item || target != v || source != a || attack != w || amount != float32(float64(raw)*0.375) || typ != object.DamageImpact ||
							cached.Field547 != 1 || cached.Field1 != 0 || live.Field547 != 55 || live.Field1 != math.Float32bits(0.5) {
							t.Fatal("live durability/signed amount/cached attribution")
						}
						events = append(events, "wear")
						item.ObjFlags |= object.FlagDestroyed
						return true
					}
					r.DamageBlockItem, r.Melee.DamageBlockWeapon = wear, wear
					r.Melee.MonsterPopBlockAction = func(target *Object) {
						if target != v {
							t.Fatal("pop target")
						}
						events = append(events, "pop")
					}
					r.DefaultDamage = func(target, source, attack *Object, amount int32, typ object.DamageType) bool {
						if blocked || target != v || source != a || attack != w || amount != raw || typ != object.DamageImpact {
							t.Fatal("missile-mutated caster reached wrong tail")
						}
						events = append(events, "default")
						return true
					}
					if h, result := PlayerDamageNative4E17B0(v, a, w, raw, object.DamageImpact, r); !h || result == blocked {
						t.Fatal("live defense result")
					}
					want := []string{"exclude", "front"}
					appendReflect := func() {
						want = append(want, "reflect")
						if mutation != "reflect-removes" && mutation != "reflect-sub2" {
							want = append(want, "clear", "owner")
						}
					}
					if blocked {
						if equipment == "sword" && initialMissile {
							appendReflect()
						}
						want = append(want, "audio")
						if equipment == "shield" && (initialMissile || mutation == "audio-missile") {
							appendReflect()
						}
						if equipment != "shield" {
							want = append(want, "action")
						}
						want = append(want, "percent", "admit", "wear", "pop")
					} else {
						want = append(want, "default")
					}
					if !slices.Equal(events, want) || v.HealthData.Cur != 20 || old.Flags().Has(object.FlagDestroyed) {
						t.Fatalf("defense events=%v want=%v", events, want)
					}
				})
			}
		}
	}
}

func TestPlayerDamageNPCCasterImpact4E17B0NilBlockItemFaultOrder(t *testing.T) {
	for _, equipment := range []string{"shield", "sword", "staff"} {
		t.Run(equipment, func(t *testing.T) {
			v, a, w := defaultDamageCasterImpactFixture4E0B30(t, "player-self", "NPC")
			ud := v.UpdateDataMonster()
			mask, sound := uint32(2), 878
			switch equipment {
			case "shield":
				ud.ArmorEquipFlags, ud.AIStack[0].Action = 0x1000000, uint32(ai.ACTION_BLOCK_ATTACK)
			case "sword":
				ud.WeaponEquipFlags, mask, sound = 0x400, 0x400, 890
			case "staff":
				ud.WeaponEquipFlags, mask, sound = 0x8000, 0x8000, 894
			}
			var events []string
			r := damageMeleeRuntimeFixture4E17B0(t)
			r.BlockDirection = func(*Object, types.Pointf) bool { return true }
			r.Audio = func(id int, target *Object) {
				if id != sound || target != v {
					t.Fatal("nil item audio")
				}
				events = append(events, "audio")
			}
			r.Melee.MonsterBlockAction = func(*Object) { events = append(events, "action") }
			r.BlockDamagePercent = func() float64 { events = append(events, "percent"); return 0.25 }
			wear := func(item, target, source, attack *Object, amount float32, typ object.DamageType) bool {
				if item != nil || target != v || source != a || attack != w || amount != 2 || typ != object.DamageImpact {
					t.Fatal("original nil item durability arguments")
				}
				events = append(events, "wear-nil")
				return true
			}
			r.DamageBlockItem, r.Melee.DamageBlockWeapon, r.DefaultDamage = wear, wear, nil
			want := []string{"audio"}
			if mask != 2 {
				want = append(want, "action")
			}
			want = append(want, "percent", "wear-nil")
			var fault any
			func() {
				defer func() { fault = recover() }()
				if h, result := PlayerDamageNative4E17B0(v, a, w, 8, object.DamageImpact, r); !h || result {
					t.Fatal("nil item block result")
				}
			}()
			if (fault != nil) != (mask != 2) || !slices.Equal(events, want) || v.HealthData.Cur != 20 {
				t.Fatalf("nil item fault=%v events=%v want=%v", fault, events, want)
			}
		})
	}
}

func TestPlayerDamageNPCCasterImpact4E17B0EarlyAndNilCarry(t *testing.T) {
	for _, gate := range []string{"no-update", "dead", "invulnerable-audio", "invulnerable-quiet", "entry-nil", "live-nil"} {
		t.Run(gate, func(t *testing.T) {
			v, a, w := defaultDamageCasterImpactFixture4E0B30(t, "player-self", "NPC")
			cached := v.UpdateDataMonster()
			cached.Field547, cached.Field546, cached.Field1 = 99, 77, math.Float32bits(0.125)
			var events []string
			r := damageMeleeRuntimeFixture4E17B0(t)
			r.Frame = func() uint32 {
				events = append(events, "frame")
				if gate == "invulnerable-quiet" {
					return 1401
				}
				return 1400
			}
			r.Audio = func(id int, target *Object) {
				if id != 71 || target != v {
					t.Fatal("invulnerable audio")
				}
				events = append(events, "audio")
			}
			r.CoopMode = func() bool { events = append(events, "coop"); return false }
			r.BlockSourceExcluded = func(*Object) bool {
				events = append(events, "exclude")
				if gate == "live-nil" {
					v.UpdateData = nil
				}
				return false
			}
			r.BlockDirection = func(*Object, types.Pointf) bool { events = append(events, "facing"); return false }
			r.DefaultDamage = func(*Object, *Object, *Object, int32, object.DamageType) bool {
				t.Fatal("fault/early gate reached HP")
				return false
			}
			if gate != "live-nil" {
				v.UpdateData = nil
			}
			switch gate {
			case "no-update":
				v.ObjFlags |= object.FlagNoUpdate
			case "dead":
				v.ObjFlags |= object.FlagDead
			case "invulnerable-audio", "invulnerable-quiet":
				v.Buffs |= 1 << playerDamageInvulnerableEnchant4E17B0
			}
			var fault any
			func() {
				defer func() { fault = recover() }()
				h, result := PlayerDamageNative4E17B0(v, a, w, 8, object.DamageImpact, r)
				if !h || result != (gate == "invulnerable-audio" || gate == "invulnerable-quiet") {
					t.Fatal("early result")
				}
			}()
			var want []string
			marker := uint32(99)
			switch gate {
			case "invulnerable-audio":
				want = []string{"frame", "audio"}
			case "invulnerable-quiet":
				want = []string{"frame"}
			case "entry-nil":
				want = []string{"coop"}
			case "live-nil":
				want, marker = []string{"coop", "exclude", "facing"}, 0
			}
			if (fault != nil) != (gate == "entry-nil" || gate == "live-nil") || !slices.Equal(events, want) ||
				cached.Field547 != marker || cached.Field546 != 77 || cached.Field1 != math.Float32bits(0.125) || v.HealthData.Cur != 20 {
				t.Fatalf("early fault=%v events=%v want=%v marker=%d", fault, events, want, cached.Field547)
			}
		})
	}
}
