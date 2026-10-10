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

func TestPlayerDamageArmedExplosion4E17B0ServiceUseBoundaries(t *testing.T) {
	for _, player := range []bool{false, true} {
		for _, boundary := range []string{"exclusion", "direction", "armor", "Quest", "default", "nil-shield", "live-class", "live-update", "invulnerable-frame", "invulnerable-audio", "invulnerable-quiet", "dead-before-update", "no-update-before-update", "excluded-before-direction"} {
			t.Run(fmt.Sprintf("player-%t/%s", player, boundary), func(t *testing.T) {
				v, a := damageMeleeUnitFixture4E17B0(t, player), damageMeleeUnitFixture4E17B0(t, !player)
				w := &Object{TypeInd: 1174, ObjClass: object.Class(85458944), ObjSubClass: 256, InitData: unsafe.Pointer(&ModifierInitData{}), PrevPos: types.Ptf(17, 9)}
				entry := v.UpdateData
				if player {
					u := v.UpdateDataPlayer()
					u.Field57, u.Field21, u.Field76, u.Field75 = math.Float32bits(.5), math.Float32bits(.25), 99, 77
				} else {
					u := v.UpdateDataMonster()
					u.Field518, u.Field1, u.Field547, u.Field546 = math.Float32bits(.5), math.Float32bits(.25), 99, 77
				}
				reasons, events := []string{}, []string{}
				defaults := 0
				r := PlayerDamageRuntime4E17B0{
					Frame:               func() uint32 { return 1400 },
					BlockSourceExcluded: func(*Object) bool { events = append(events, "exclusion"); return false },
					BlockDirection:      func(*Object, types.Pointf) bool { events = append(events, "direction"); return false },
					DefaultDamage: func(target, source, weapon *Object, d int32, typ object.DamageType) bool {
						if target != v || source != a || weapon != w || d != 5 || typ != object.DamageExplosion {
							t.Fatal("default args")
						}
						defaults++
						return true
					},
					Unsupported: func(reason string, _, _, _ *Object, _ int32, _ object.DamageType) {
						reasons = append(reasons, reason)
						events = append(events, "unsupported")
					},
				}
				wantMarker, wantKind, wantCarry := uint32(1), uint32(1174), float32(.25)
				wantHandled, wantResult, wantDefaults := false, false, 0
				var wantReason string
				switch boundary {
				case "exclusion":
					r.BlockSourceExcluded = nil
					wantMarker, wantKind, wantReason = 0, 77, "missing armed EXPLOSION weapon exclusion"
				case "direction":
					r.BlockDirection = nil
					wantReason = "missing armed EXPLOSION block direction"
				case "armor":
					armor := damageMeleeArmorFixture4E17B0(v, .5, .25)
					r.CanDamageArmor = func(*Object) bool { t.Fatal("missing lookup should stop before durability"); return false }
					_ = armor
					wantCarry, wantReason = -.25, "unsupported live armed EXPLOSION armor durability"
				case "Quest":
					r.QuestMode = func() bool { return true }
					wantCarry, wantReason = -.25, "missing live armed EXPLOSION Quest scale"
				case "default":
					r.DefaultDamage = nil
					wantCarry, wantReason = -.25, "missing armed EXPLOSION default damage"
				case "nil-shield":
					if player {
						v.UpdateDataPlayer().Player.ArmorEquip, v.UpdateDataPlayer().State = 0x1000000, PlayerState16
					} else {
						v.UpdateDataMonster().ArmorEquipFlags, v.UpdateDataMonster().AIStack[0].Action = 0x1000000, uint32(ai.ACTION_BLOCK_ATTACK)
					}
					r.BlockDirection = func(*Object, types.Pointf) bool { events = append(events, "direction"); return true }
					r.Audio = func(id int, _ *Object) {
						if id != 878 {
							t.Fatal("shield sound")
						}
						events = append(events, "audio")
					}
					r.BlockDamagePercent = func() float64 { events = append(events, "balance"); return .25 }
					r.CanDamageBlockItem = func(*Object) bool { t.Fatal("nil shield must not preflight item"); return false }
					r.DamageBlockItem = func(item, target, source, weapon *Object, d float32, typ object.DamageType) bool {
						if item != nil || target != v || source != a || weapon != w || d != 2.25 || typ != object.DamageExplosion {
							t.Fatal("original nil no-op wear args")
						}
						events = append(events, "wear")
						return true
					}
					wantReason = "original armed EXPLOSION nil shield fault"
				case "live-class":
					r.BlockSourceExcluded = func(*Object) bool { v.ObjClass = object.ClassArmor; return false }
					wantMarker, wantKind, wantReason = 0, 77, "unsupported live armed EXPLOSION cached marker layout"
				case "live-update":
					r.BlockDirection = func(*Object, types.Pointf) bool { v.UpdateData = nil; return false }
					wantReason = "unsupported live armed EXPLOSION carry update"
				case "invulnerable-frame", "invulnerable-audio", "invulnerable-quiet":
					v.Buffs |= 1 << playerDamageInvulnerableEnchant4E17B0
					wantMarker, wantKind = 99, 77
					if boundary == "invulnerable-frame" {
						r.Frame = nil
						wantReason = "missing armed EXPLOSION invulnerability frame"
					} else if boundary == "invulnerable-audio" {
						wantReason = "missing armed EXPLOSION invulnerability audio"
					} else {
						r.Frame = func() uint32 { return 1401 }
						wantHandled, wantResult = true, true
					}
				case "dead-before-update", "no-update-before-update":
					v.UpdateData = nil
					if boundary == "dead-before-update" {
						v.ObjFlags |= object.FlagDead
					} else {
						v.ObjFlags |= object.FlagNoUpdate
					}
					wantMarker, wantKind, wantHandled = 99, 77, true
				case "excluded-before-direction":
					r.BlockSourceExcluded = func(*Object) bool { return true }
					r.BlockDirection = nil
					wantCarry, wantHandled, wantResult, wantDefaults = -.25, true, true, 1
				}
				h, result := PlayerDamageNative4E17B0(v, a, w, 9, object.DamageExplosion, r)
				var marker, kind, carry uint32
				if player {
					u := (*PlayerUpdateData)(entry)
					marker, kind, carry = u.Field76, u.Field75, u.Field21
				} else {
					u := (*MonsterUpdateData)(entry)
					marker, kind, carry = u.Field547, u.Field546, u.Field1
				}
				if h != wantHandled || result != wantResult || defaults != wantDefaults || marker != wantMarker || kind != wantKind || carry != math.Float32bits(wantCarry) {
					t.Fatalf("%s handled/result=%t/%t default=%d marker=%d/%d carry=%g", boundary, h, result, defaults, marker, kind, math.Float32frombits(carry))
				}
				if wantReason == "" {
					if len(reasons) != 0 {
						t.Fatalf("unused service rejected %v", reasons)
					}
				} else if !slices.Equal(reasons, []string{wantReason}) {
					t.Fatalf("reports=%v want=%s", reasons, wantReason)
				}
				if boundary == "nil-shield" && !slices.Equal(events, []string{"exclusion", "direction", "audio", "balance", "wear", "unsupported"}) {
					t.Fatalf("original fault effects must be retained in order: %v", events)
				}
			})
		}
	}
}

func TestPlayerDamageArmedExplosion4E17B0AdmissionDisjoint(t *testing.T) {
	for _, shape := range []string{"Sword", "renamed-retyped-Sword", "qualifying-Wand", "Hammer", "nil-source", "nil-weapon", "source-is-weapon", "missile", "unit-weapon", "unqualified-Wand", "nil-Wand-use", "weaponless", "non-unit-source", "ranged-source", "BLADE", "CRUSH", "POISON"} {
		t.Run(shape, func(t *testing.T) {
			a := damageMeleeUnitFixture4E17B0(t, false)
			w := &Object{TypeInd: 1174, ObjClass: object.Class(85458944), ObjSubClass: 256, InitData: unsafe.Pointer(&ModifierInitData{})}
			typ, want := object.DamageExplosion, false
			switch shape {
			case "Sword":
				want = true
			case "renamed-retyped-Sword":
				w.TypeInd = 60000
				want = true
			case "qualifying-Wand":
				w.ObjClass, w.ObjSubClass, w.UseData.Ptr = object.ClassWand, 0x10000, unsafe.Pointer(&WandUseData{Flags: 2})
				want = true
			case "Hammer":
				w.ObjClass, w.ObjSubClass = object.ClassWeapon, 0x4000
				want = true
			case "nil-source":
				a = nil
			case "nil-weapon", "weaponless":
				w = nil
			case "source-is-weapon":
				w = a
			case "missile":
				w.ObjClass |= object.ClassMissile
			case "unit-weapon":
				w.ObjClass |= object.ClassMonster
			case "unqualified-Wand":
				w.ObjClass, w.ObjSubClass, w.UseData.Ptr = object.ClassWand, 0x10000, unsafe.Pointer(&WandUseData{})
			case "nil-Wand-use":
				w.ObjClass, w.ObjSubClass = object.ClassWand, 0x10000
			case "non-unit-source":
				a.ObjClass = object.ClassSimple
			case "ranged-source":
				a.ObjClass |= object.ClassMissile
			case "BLADE":
				typ = object.DamageBlade
			case "CRUSH":
				typ = object.DamageCrush
			case "POISON":
				typ = object.DamagePoison
			}
			if got := playerDamageArmedExplosionShape4E17B0(a, w, typ); got != want {
				t.Fatalf("shape=%t want=%t", got, want)
			}
		})
	}
}

// OpenNox's gameex extension is deliberately separate from stock GAME.EXE
// parity. A pure type-7 hit never uses a staff/GreatSword BLADE block.
func TestPlayerDamageArmedExplosion4E17B0OptionalGameexBlock(t *testing.T) {
	for _, enabled := range []bool{false, true} {
		t.Run(fmt.Sprintf("enabled-%t", enabled), func(t *testing.T) {
			v, a := damageMeleeUnitFixture4E17B0(t, true), damageMeleeUnitFixture4E17B0(t, false)
			v.UpdateDataPlayer().State, v.UpdateDataPlayer().Player.ArmorEquip = PlayerState1, 0x1000000
			w := &Object{TypeInd: 1174, ObjClass: object.ClassWeapon, ObjSubClass: 256}
			shield := &Object{ObjFlags: object.FlagEquipped, ObjSubClass: 2}
			v.InvFirstItem = shield
			queried, defaults, wears := 0, 0, 0
			r := PlayerDamageRuntime4E17B0{BlockSourceExcluded: func(*Object) bool { return false }, BlockDirection: func(*Object, types.Pointf) bool { return true }, BerserkShieldBlock: func(*Object) bool { queried++; return enabled }, Audio: func(id int, _ *Object) {
				if id != 878 {
					t.Fatal("gameex sound")
				}
			}, BlockDamagePercent: func() float64 { return .25 }, DamageBlockItem: func(got, owner, source, weapon *Object, amount float32, typ object.DamageType) bool {
				if got != shield || owner != v || source != a || weapon != w || amount != .75 || typ != object.DamageExplosion {
					t.Fatal("gameex wear")
				}
				wears++
				return true
			}, DefaultDamage: func(_, _, _ *Object, d int32, typ object.DamageType) bool {
				if d != 3 || typ != object.DamageExplosion {
					t.Fatal("unchanged type7")
				}
				defaults++
				return true
			}, Unsupported: func(reason string, _, _, _ *Object, _ int32, _ object.DamageType) { t.Fatal(reason) }}
			h, result := PlayerDamageNative4E17B0(v, a, w, 3, object.DamageExplosion, r)
			if !h || result == enabled || queried != 1 || defaults != map[bool]int{false: 1, true: 0}[enabled] || wears != map[bool]int{false: 0, true: 1}[enabled] {
				t.Fatalf("extension handled/result=%t/%t query/default/wear=%d/%d/%d", h, result, queried, defaults, wears)
			}
		})
	}
}
