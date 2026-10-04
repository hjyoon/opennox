package server

import (
	"fmt"
	"math"
	"reflect"
	"testing"
	"unsafe"

	"github.com/opennox/libs/object"
	"github.com/opennox/libs/types"
)

// Ordered service-boundary tests, not GUI damage or actual durability/HP
// callbacks. The separate native fixture covers the real C dispatcher and HP.
func bitePrefixCases4E17B0(t *testing.T, run func(*testing.T, bool)) {
	t.Helper()
	for _, observe := range []bool{false, true} {
		t.Run(fmt.Sprintf("observe-%t", observe), func(t *testing.T) { run(t, observe) })
	}
}

func bitePrefixFixture4E17B0(t *testing.T, observe bool) (*Object, *Object, *PlayerUpdateData, PlayerDamageRuntime4E17B0) {
	t.Helper()
	target, source, sound := playerDamageFixture4E17B0(t)
	update := target.UpdateDataPlayer()
	update.Field57, update.Field21 = math.Float32bits(0.25), math.Float32bits(0.125)
	update.Field76, update.Field75, update.State = 88, 77, PlayerState13
	update.Player.ArmorEquip, update.Player.WeaponEquip = 0, 0
	if observe {
		update.Player.Field3680, update.Player.CameraFollowObj = 0x22, source
	}
	target.HealthData.Cur, target.HealthData.Field2, target.HealthData.Max = 200, 200, 200
	source.TypeInd, source.PrevPos, source.PosVec = 777, types.Ptf(20, 0), types.Ptf(-20, 0)
	r := playerDamageRuntime4E17B0(t, sound, new([]int32))
	r.DamageClear = func(*Object, int32) { t.Fatal("unexpected HP callback") }
	r.ObserveClear = func(*Object) {
		if !observe || update.Field76 != 0 || update.Field75 != 77 {
			t.Fatal("ObserveClear ran outside the cleared entry prefix")
		}
		update.Player.Field3680 &^= 2
		update.Player.CameraFollowObj = nil
	}
	return target, source, update, r
}

func bitePrefixBlockServices4E17B0(t *testing.T, target, source, shield *Object, raw int32, events *[]string, r *PlayerDamageRuntime4E17B0) {
	t.Helper()
	r.CanDamageBlockItem = func(item *Object) bool {
		if item != shield {
			t.Fatal("block used a stale inventory item")
		}
		*events = append(*events, "block-admission")
		return true
	}
	r.Audio = func(id int, obj *Object) {
		if id != 878 || obj != target {
			t.Fatal("wrong block sound")
		}
		*events = append(*events, "block-audio")
	}
	r.BlockDamagePercent = func() float64 {
		*events = append(*events, "block-percent")
		return 0.25
	}
	r.DamageBlockItem = func(item, owner, attacker, weapon *Object, amount float32, typ object.DamageType) bool {
		if item != shield || owner != target || attacker != source || weapon != source || amount != float32(raw)*0.25 || typ != object.DamageBite {
			t.Fatal("BITE block arguments changed")
		}
		*events = append(*events, "block-wear")
		return true
	}
	r.PlayerSetState = func(*Object, PlayerState) bool { t.Fatal("intact shield changed stance"); return false }
}

func TestPlayerDamageBitePrefixOrder4E17B0(t *testing.T) {
	bitePrefixCases4E17B0(t, func(t *testing.T, observe bool) {
		for _, shielded := range []bool{false, true} {
			for _, facing := range []string{"front", "rear", "excluded"} {
				for _, raw := range []int32{1, 5, 21} {
					t.Run(fmt.Sprintf("shield-%t/%s/raw-%d", shielded, facing, raw), func(t *testing.T) {
						target, source, cached, r := bitePrefixFixture4E17B0(t, observe)
						shield := &Object{ObjSubClass: 2, ObjFlags: object.FlagEquipped, HealthData: &HealthData{Cur: 10, Max: 10}, InitData: unsafe.Pointer(new(ModifierInitData))}
						if shielded {
							cached.Player.ArmorEquip = 0x1000000
							target.InvFirstItem = shield
						}
						var events []string
						wantPos := source.PrevPos
						if observe {
							wantPos = types.Ptf(30, 7)
						}
						r.ObserveClear = func(*Object) {
							if !observe || cached.Field76 != 0 || cached.Field75 != 77 {
								t.Fatal("ObserveClear ordering")
							}
							events = append(events, "observe")
							cached.Player.Field3680 &^= 2
							cached.Player.CameraFollowObj = nil
							source.PrevPos = wantPos
						}
						r.BlockSourceExcluded = func(obj *Object) bool {
							if obj != source || cached.Field76 != 0 || cached.Field75 != 77 {
								t.Fatal("self-weapon prefix/exclusion ordering")
							}
							events = append(events, "excluded")
							source.PrevPos, source.TypeInd = types.Ptf(-100, 99), 888
							cached.Field57, cached.Field21, cached.Player.ArmorEquip = math.Float32bits(0.875), math.Float32bits(0.5), 0
							return facing == "excluded"
						}
						r.BlockDirection = func(obj *Object, pos types.Pointf) bool {
							if obj != target || pos != wantPos || cached.Field76 != 0 || cached.Field75 != 77 {
								t.Fatal("facing lost the pre-exclusion PrevPos snapshot")
							}
							events = append(events, "direction")
							cached.State, cached.Player.WeaponEquip = PlayerState16, 0x400
							return facing == "front"
						}
						bitePrefixBlockServices4E17B0(t, target, source, shield, raw, &events, &r)
						r.GodMode = func() bool { events = append(events, "god"); return false }
						r.QuestMode = func() bool { events = append(events, "quest"); return false }
						r.BuffOff = func(obj *Object, id EnchantID) {
							if obj != target || id != playerDamageInvisibleEnchant4E17B0 {
								t.Fatal("BuffOff arguments")
							}
							events = append(events, "buff")
						}
						accumulated := float32(0.75*float64(raw)) + 0.5
						effective := playerDamageRound4E17B0(accumulated)
						r.DamageClear = func(obj *Object, damage int32) {
							if obj != target || damage != effective {
								t.Fatalf("HP argument=%d, want %d", damage, effective)
							}
							events = append(events, "hp")
						}
						h, result := PlayerDamageNative4E17B0(target, source, source, raw, object.DamageBite, r)
						blocked := shielded && facing == "front"
						want := []string{"excluded"}
						if observe {
							want = append([]string{"observe"}, want...)
						}
						if facing != "excluded" {
							want = append(want, "direction")
						}
						marker, markerType, carry := uint32(2), uint32(object.DamageBite), accumulated-float32(effective)
						if blocked {
							want = append(want, "block-admission", "block-audio", "block-percent", "block-wear")
							marker, markerType, carry = 0, 77, 0.5
						} else {
							want = append(want, "god", "quest", "buff", "hp")
						}
						if !h || result == blocked || !reflect.DeepEqual(events, want) || cached.Field76 != marker || cached.Field75 != markerType || cached.Field21 != math.Float32bits(carry) || *target.HealthData != (HealthData{Cur: 200, Field2: 200, Max: 200}) {
							t.Fatalf("BITE order=%t/%t events=%v want=%v marker=%d/%d carry=%g", h, result, events, want, cached.Field76, cached.Field75, math.Float32frombits(cached.Field21))
						}
					})
				}
			}
		}
	})
}

func TestPlayerDamageBitePrefixLiveRecord4E17B0(t *testing.T) {
	bitePrefixCases4E17B0(t, func(t *testing.T, observe bool) {
		for _, stage := range []string{"exclusion", "direction"} {
			for _, change := range []string{"replacement-front", "replacement-rear", "class", "nil-update"} {
				t.Run(stage+"/"+change, func(t *testing.T) {
					target, source, cached, r := bitePrefixFixture4E17B0(t, observe)
					cached.Player.ArmorEquip = 0x1000000
					live := &PlayerUpdateData{Player: &Player{Field3680: 3, CameraFollowObj: source}, State: PlayerState13, Field57: math.Float32bits(0.9), Field21: math.Float32bits(0.5), Field76: 31, Field75: 33}
					oldShield, shield := &Object{ObjSubClass: 2, ObjFlags: object.FlagEquipped, InitData: unsafe.Pointer(new(ModifierInitData))}, &Object{ObjSubClass: 2, ObjFlags: object.FlagEquipped, InitData: unsafe.Pointer(new(ModifierInitData))}
					target.InvFirstItem = oldShield
					mutate := func() {
						cached.Player.ArmorEquip, cached.Field57, cached.State = 0, math.Float32bits(0.875), PlayerState16
						target.InvFirstItem = shield
						switch change {
						case "class":
							target.ObjClass = object.ClassSimple
						case "nil-update":
							target.UpdateData = nil
						default:
							target.UpdateData = unsafe.Pointer(live)
						}
					}
					front := change == "replacement-front"
					r.BlockSourceExcluded = func(*Object) bool {
						if stage == "exclusion" {
							mutate()
						}
						return false
					}
					r.BlockDirection = func(*Object, types.Pointf) bool {
						if stage == "direction" {
							mutate()
						}
						return front
					}
					var events []string
					bitePrefixBlockServices4E17B0(t, target, source, shield, 5, &events, &r)
					calls, reason := 0, ""
					r.DamageClear = func(obj *Object, amount int32) {
						if change != "replacement-rear" || obj != target || amount != 4 {
							t.Fatal("invalid live record reached HP or cached armor/live carry lost")
						}
						calls++
					}
					r.Unsupported = func(why string, _, _, _ *Object, _ int32, _ object.DamageType) { reason = why }
					h, result := PlayerDamageNative4E17B0(target, source, source, 5, object.DamageBite, r)
					marker, markerType := uint32(0), uint32(77)
					switch change {
					case "replacement-front":
						if !h || result || calls != 0 || reason != "" || len(events) != 4 || live.Field21 != math.Float32bits(0.5) {
							t.Fatal("replacement block did not use entry mask/post-facing cached stance/live shield")
						}
					case "replacement-rear":
						marker, markerType = 2, uint32(object.DamageBite)
						if !h || !result || calls != 1 || reason != "" || len(events) != 0 || live.Field21 != math.Float32bits(0.25) {
							t.Fatal("replacement tail did not use live carry")
						}
					default:
						if h || result || reason == "" || calls != 0 || len(events) != 0 {
							t.Fatalf("live guard=%t/%t reason=%q events=%v", h, result, reason, events)
						}
					}
					if cached.Field76 != marker || cached.Field75 != markerType || cached.Field21 != math.Float32bits(0.125) || live.Field76 != 31 || live.Field75 != 33 || live.Player.Field3680 != 3 || live.Player.CameraFollowObj != source || target.HealthData.Cur != 200 {
						t.Fatal("cached marker/live carry separation or no repeated ObserveClear was lost")
					}
				})
			}
		}
	})
}

func TestPlayerDamageBitePrefixLateTail4E17B0(t *testing.T) {
	bitePrefixCases4E17B0(t, func(t *testing.T, observe bool) {
		for _, clearMarker := range []bool{false, true} {
			for _, mode := range []string{"quest", "god-player", "god-nonplayer"} {
				t.Run(fmt.Sprintf("clear-marker-%t/%s", clearMarker, mode), func(t *testing.T) {
					target, source, cached, r := bitePrefixFixture4E17B0(t, observe)
					armor := &Object{ObjClass: object.ClassArmor, ObjFlags: object.FlagEquipped, HealthData: &HealthData{Cur: 10, Max: 10}, UpdateData: unsafe.Pointer(new(WeaponArmorUpdateData)), InitData: unsafe.Pointer(new(ModifierInitData)), Damage: unsafe.Pointer(new(byte))}
					target.InvFirstItem = armor
					var events []string
					worn, quest := false, false
					r.BlockSourceExcluded = func(*Object) bool { events = append(events, "excluded"); return false }
					r.BlockDirection = func(*Object, types.Pointf) bool {
						cached.Field57, cached.Field21 = math.Float32bits(0.5), math.Float32bits(0.5)
						events = append(events, "direction")
						return false
					}
					r.CanDamageArmor = func(item *Object) bool { return item == armor }
					r.ItemArmorValue = func(item *Object) float32 {
						if item != armor {
							t.Fatal("wrong armor lookup")
						}
						return 0.5
					}
					r.DamageArmor = func(item, attacker, weapon *Object, amount int32, typ object.DamageType) bool {
						if item != armor || attacker != source || weapon != source || amount != 1 || typ != object.DamageBite || cached.Field21 != math.Float32bits(0.25) || cached.Field76 != 0 {
							t.Fatal("wear lost entry armor/live carry/prefix ordering")
						}
						events = append(events, "armor")
						worn, quest = true, true
						cached.Field76, cached.Field75 = 9, 123
						if clearMarker {
							cached.Field76 = 0
						}
						if mode == "god-nonplayer" {
							target.ObjClass = object.ClassSimple
						}
						return true
					}
					r.GodMode = func() bool {
						if !worn {
							t.Fatal("GodMode ran before armor")
						}
						events = append(events, "god")
						return mode != "quest"
					}
					r.QuestMode = func() bool {
						if !worn {
							t.Fatal("QuestMode was cached before armor")
						}
						events = append(events, "quest")
						return quest
					}
					r.QuestDamageScale = func() float32 { events = append(events, "scale"); return 0.5 }
					r.BuffOff = func(*Object, EnchantID) { events = append(events, "buff") }
					r.DamageClear = func(obj *Object, amount int32) {
						if obj != target || amount != 2 {
							t.Fatal("late Quest scaling lost")
						}
						events = append(events, "hp")
					}
					h, result := PlayerDamageNative4E17B0(target, source, source, 5, object.DamageBite, r)
					want := []string{"excluded", "direction", "armor", "god"}
					if mode != "god-player" {
						want = append(want, "quest", "scale", "buff", "hp")
					}
					marker, markerType := uint32(9), uint32(123)
					if clearMarker {
						marker = 0
						if mode != "god-nonplayer" {
							marker, markerType = 2, uint32(object.DamageBite)
						}
					}
					if !h || !result || !reflect.DeepEqual(events, want) || cached.Field76 != marker || cached.Field75 != markerType || cached.Field21 != math.Float32bits(0.25) || target.HealthData.Cur != 200 || armor.HealthData.Cur != 10 {
						t.Fatalf("late BITE=%t/%t events=%v want=%v marker=%d/%d", h, result, events, want, cached.Field76, cached.Field75)
					}
				})
			}
		}
	})
}

func TestPlayerDamageBitePrefixAdmission4E17B0(t *testing.T) {
	bitePrefixCases4E17B0(t, func(t *testing.T, observe bool) {
		for _, service := range []string{"exclusion", "direction", "observe"} {
			if service == "observe" && !observe {
				continue
			}
			t.Run(service, func(t *testing.T) {
				target, source, cached, r := bitePrefixFixture4E17B0(t, observe)
				beforeObj, beforeUD, beforePlayer, beforeHP := *target, *cached, *cached.Player, *target.HealthData
				switch service {
				case "exclusion":
					r.BlockSourceExcluded = nil
				case "direction":
					r.BlockDirection = nil
				case "observe":
					r.ObserveClear = nil
				}
				reason := ""
				r.Unsupported = func(why string, _, _, _ *Object, _ int32, _ object.DamageType) { reason = why }
				h, result := PlayerDamageNative4E17B0(target, source, source, 5, object.DamageBite, r)
				if h || result || reason != "missing player bite prefix service" || *target != beforeObj || *cached != beforeUD || *cached.Player != beforePlayer || *target.HealthData != beforeHP {
					t.Fatalf("prefix admission=%t/%t reason=%q mutated=%t", h, result, reason, *cached != beforeUD)
				}
			})
		}
	})
}

func TestPlayerDamageBitePrefixBlockUnusedTail4E17B0(t *testing.T) {
	bitePrefixCases4E17B0(t, func(t *testing.T, observe bool) {
		target, source, cached, r := bitePrefixFixture4E17B0(t, observe)
		cached.Player.ArmorEquip, cached.State = 0x1000000, PlayerState16
		shield := &Object{ObjSubClass: 2, ObjFlags: object.FlagEquipped}
		target.InvFirstItem = shield
		var events []string
		bitePrefixBlockServices4E17B0(t, target, source, shield, 5, &events, &r)
		r.BlockDirection = func(*Object, types.Pointf) bool { return true }
		r.DefaultDamage, r.DamageClear, r.BuffOff, r.QuestDamageScale, r.ItemArmorValue = nil, nil, nil, nil, nil
		r.QuestMode = func() bool { t.Fatal("block evaluated unused Quest tail"); return true }
		r.GodMode = func() bool { t.Fatal("block evaluated unused GodMode tail"); return true }
		h, result := PlayerDamageNative4E17B0(target, source, source, 5, object.DamageBite, r)
		if !h || result || len(events) != 4 || cached.Field76 != 0 || cached.Field75 != 77 || cached.Field21 != math.Float32bits(0.125) || target.HealthData.Cur != 200 {
			t.Fatal("unused HP/armor/Quest callbacks became block prerequisites")
		}
	})
}

func TestPlayerDamageBitePrefixEarlyGate4E17B0(t *testing.T) {
	bitePrefixCases4E17B0(t, func(t *testing.T, observe bool) {
		for _, gate := range []string{"no-update", "dead", "invulnerable", "observer-bit-one", "coop-self"} {
			t.Run(gate, func(t *testing.T) {
				target, source, cached, r := bitePrefixFixture4E17B0(t, observe)
				want := false
				switch gate {
				case "no-update":
					target.ObjFlags |= object.FlagNoUpdate
				case "dead":
					target.ObjFlags |= object.FlagDead
				case "invulnerable":
					target.Buffs |= 1 << playerDamageInvulnerableEnchant4E17B0
					want = true
				case "observer-bit-one":
					cached.Player.Field3680 |= 1
				case "coop-self":
					source.ObjOwner = target
					r.CoopMode = func() bool { return true }
				}
				beforeObj, beforeUD, beforePlayer, beforeHP := *target, *cached, *cached.Player, *target.HealthData
				r.ObserveClear = func(*Object) { t.Fatal("early gate ran ObserveClear") }
				r.BlockSourceExcluded = func(*Object) bool { t.Fatal("early gate ran exclusion"); return false }
				r.BlockDirection = func(*Object, types.Pointf) bool { t.Fatal("early gate ran facing"); return false }
				sounds := 0
				r.Audio = func(id int, obj *Object) {
					if gate != "invulnerable" || id != playerDamageInvulnerableSound4E17B0 || obj != target {
						t.Fatal("unexpected early sound")
					}
					sounds++
				}
				h, result := PlayerDamageNative4E17B0(target, source, source, 5, object.DamageBite, r)
				wantSounds := 0
				if gate == "invulnerable" {
					wantSounds = 1
				}
				if !h || result != want || sounds != wantSounds || *target != beforeObj || *cached != beforeUD || *cached.Player != beforePlayer || *target.HealthData != beforeHP {
					t.Fatalf("early gate=%t/%t sounds=%d", h, result, sounds)
				}
			})
		}
	})
}

func TestPlayerDamageBitePrefixShapeBoundary4E17B0(t *testing.T) {
	bitePrefixCases4E17B0(t, func(t *testing.T, observe bool) {
		for _, shape := range []string{"zero", "negative", "nil-source", "nil-weapon", "distinct-weapon", "player-source", "nil-source-update"} {
			t.Run(shape, func(t *testing.T) {
				target, source, cached, r := bitePrefixFixture4E17B0(t, observe)
				weapon, raw := source, int32(5)
				switch shape {
				case "zero":
					raw = 0
				case "negative":
					raw = -1
				case "nil-source":
					source = nil
				case "nil-weapon":
					weapon = nil
				case "distinct-weapon":
					weapon = &Object{ObjClass: object.ClassMissile}
				case "player-source":
					source.ObjClass = object.ClassPlayer
				case "nil-source-update":
					source.UpdateData = nil
				}
				beforeObj, beforeUD, beforePlayer, beforeHP := *target, *cached, *cached.Player, *target.HealthData
				reason := ""
				r.Unsupported = func(why string, _, _, _ *Object, _ int32, _ object.DamageType) { reason = why }
				r.ObserveClear = func(*Object) { t.Fatal("unsupported shape entered prefix") }
				r.BlockSourceExcluded = func(*Object) bool { t.Fatal("unsupported shape entered prefix"); return false }
				r.BlockDirection = func(*Object, types.Pointf) bool { t.Fatal("unsupported shape entered prefix"); return false }
				h, result := PlayerDamageNative4E17B0(target, source, weapon, raw, object.DamageBite, r)
				if h || result || reason != "unsupported player damage shape" || *target != beforeObj || *cached != beforeUD || *cached.Player != beforePlayer || *target.HealthData != beforeHP {
					t.Fatalf("shape boundary=%t/%t reason=%q", h, result, reason)
				}
			})
		}
	})
}
