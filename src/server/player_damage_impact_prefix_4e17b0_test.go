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

// Service-boundary fixtures: HP and durability callbacks record arguments,
// never manufacture outcomes. The native fixture exercises real HP bindings.
func impactPrefixFixture4E17B0(t *testing.T, observe bool) (*Object, *Object, *Object, *PlayerUpdateData, PlayerDamageRuntime4E17B0) {
	t.Helper()
	target, source, cached, r := bitePrefixFixture4E17B0(t, observe)
	missile := &Object{ObjClass: object.ClassMissile, TypeInd: 801, PrevPos: types.Ptf(20, 0), PosVec: types.Ptf(-20, 0)}
	return target, source, missile, cached, r
}

func impactPrefixShieldServices4E17B0(t *testing.T, target, source, missile, shield *Object, raw int32, events *[]string, r *PlayerDamageRuntime4E17B0) {
	t.Helper()
	r.CanDamageBlockItem = func(item *Object) bool {
		if item != shield {
			t.Fatal("block used a stale shield")
		}
		*events = append(*events, "block-admission")
		return true
	}
	r.Audio = func(id int, obj *Object) {
		if id != 878 || obj != target {
			t.Fatal("wrong shield sound")
		}
		*events = append(*events, "block-audio")
	}
	r.ProjectileReflect = func(attack, owner *Object) {
		if attack != missile || owner != target {
			t.Fatal("wrong reflected projectile")
		}
		*events = append(*events, "reflect")
	}
	r.ClearOwner = func(attack *Object) {
		if attack != missile {
			t.Fatal("wrong cleared projectile")
		}
		*events = append(*events, "clear-owner")
	}
	r.SetOwner = func(owner, attack *Object) {
		if owner != target || attack != missile {
			t.Fatal("wrong new projectile owner")
		}
		*events = append(*events, "set-owner")
	}
	r.BlockDamagePercent = func() float64 { *events = append(*events, "block-percent"); return 0.25 }
	r.DamageBlockItem = func(item, owner, attacker, attack *Object, amount float32, typ object.DamageType) bool {
		if item != shield || owner != target || attacker != source || attack != missile || amount != float32(raw)*0.25 || typ != object.DamageImpact {
			t.Fatal("IMPACT shield arguments changed")
		}
		*events = append(*events, "block-wear")
		return true
	}
	r.PlayerSetState = func(*Object, PlayerState) bool { t.Fatal("intact shield changed stance"); return false }
}

func TestPlayerDamageImpactPrefixOrder4E17B0(t *testing.T) {
	bitePrefixCases4E17B0(t, func(t *testing.T, observe bool) {
		for _, shielded := range []bool{false, true} {
			for _, facing := range []string{"front", "rear", "excluded"} {
				for _, raw := range []int32{1, 5, 21} {
					t.Run(fmt.Sprintf("shield-%t/%s/raw-%d", shielded, facing, raw), func(t *testing.T) {
						target, source, missile, cached, r := impactPrefixFixture4E17B0(t, observe)
						shield := &Object{ObjSubClass: 2, ObjFlags: object.FlagEquipped, HealthData: &HealthData{Cur: 10, Max: 10}, InitData: unsafe.Pointer(new(ModifierInitData))}
						if shielded {
							cached.Player.ArmorEquip, target.InvFirstItem = 0x1000000, shield
						}
						var events []string
						wantPos := missile.PrevPos
						if observe {
							wantPos = types.Ptf(30, 7)
						}
						clear := r.ObserveClear
						r.ObserveClear = func(obj *Object) { clear(obj); events = append(events, "observe"); missile.PrevPos = wantPos }
						r.BlockSourceExcluded = func(obj *Object) bool {
							if obj != missile || cached.Field76 != 0 || cached.Field75 != 77 {
								t.Fatal("prefix must precede distinct-missile exclusion")
							}
							events = append(events, "excluded")
							missile.PrevPos, missile.TypeInd = types.Ptf(-100, 99), 888
							cached.Field57, cached.Field21, cached.Player.ArmorEquip = math.Float32bits(0.875), math.Float32bits(0.5), 0
							return facing == "excluded"
						}
						r.BlockDirection = func(obj *Object, pos types.Pointf) bool {
							if obj != target || pos != wantPos || cached.Field76 != 1 || cached.Field75 != 888 {
								t.Fatal("facing lost position snapshot/live missile attribution")
							}
							events = append(events, "direction")
							cached.State, cached.Player.WeaponEquip = PlayerState16, 0x400
							return facing == "front"
						}
						impactPrefixShieldServices4E17B0(t, target, source, missile, shield, raw, &events, &r)
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
						r.DamageClear = func(obj *Object, amount int32) {
							if obj != target || amount != effective {
								t.Fatalf("HP argument=%d, want %d", amount, effective)
							}
							events = append(events, "hp")
						}
						h, result := PlayerDamageNative4E17B0(target, source, missile, raw, object.DamageImpact, r)
						blocked := shielded && facing == "front"
						want := []string{"excluded"}
						if observe {
							want = append([]string{"observe"}, want...)
						}
						if facing != "excluded" {
							want = append(want, "direction")
						}
						carry := accumulated - float32(effective)
						if blocked {
							want = append(want, "block-admission", "block-audio", "reflect", "clear-owner", "set-owner", "block-percent", "block-wear")
							carry = 0.5
						} else {
							want = append(want, "god", "quest", "buff", "hp")
						}
						if !h || result == blocked || !reflect.DeepEqual(events, want) || cached.Field76 != 1 || cached.Field75 != 888 || cached.Field21 != math.Float32bits(carry) || *target.HealthData != (HealthData{Cur: 200, Field2: 200, Max: 200}) {
							t.Fatalf("IMPACT order=%t/%t events=%v want=%v marker=%d/%d carry=%g", h, result, events, want, cached.Field76, cached.Field75, math.Float32frombits(cached.Field21))
						}
					})
				}
			}
		}
	})
}

func TestPlayerDamageImpactPrefixLiveRecord4E17B0(t *testing.T) {
	bitePrefixCases4E17B0(t, func(t *testing.T, observe bool) {
		for _, stage := range []string{"exclusion", "direction"} {
			for _, change := range []string{"replacement-front", "replacement-rear", "class", "nil-update"} {
				t.Run(stage+"/"+change, func(t *testing.T) {
					target, source, missile, cached, r := impactPrefixFixture4E17B0(t, observe)
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
						return change == "replacement-front"
					}
					var events []string
					impactPrefixShieldServices4E17B0(t, target, source, missile, shield, 5, &events, &r)
					calls, reason := 0, ""
					r.DamageClear = func(obj *Object, amount int32) {
						if change != "replacement-rear" || obj != target || amount != 4 {
							t.Fatal("invalid live carry or record reached HP")
						}
						calls++
					}
					r.Unsupported = func(why string, _, _, _ *Object, _ int32, _ object.DamageType) { reason = why }
					h, result := PlayerDamageNative4E17B0(target, source, missile, 5, object.DamageImpact, r)
					marker, markerType := uint32(1), uint32(801)
					if stage == "exclusion" && change == "class" {
						marker, markerType = 0, 77
					}
					switch change {
					case "replacement-front":
						if !h || result || calls != 0 || reason != "" || len(events) != 7 || live.Field21 != math.Float32bits(0.5) {
							t.Fatal("block lost cached mask/stance/live shield")
						}
					case "replacement-rear":
						if !h || !result || calls != 1 || reason != "" || len(events) != 0 || live.Field21 != math.Float32bits(0.25) {
							t.Fatal("tail lost live carry")
						}
					default:
						if h || result || reason != "unsupported live player impact record" || calls != 0 || len(events) != 0 {
							t.Fatalf("live guard=%t/%t reason=%q events=%v", h, result, reason, events)
						}
					}
					if cached.Field76 != marker || cached.Field75 != markerType || cached.Field21 != math.Float32bits(0.125) || live.Field76 != 31 || live.Field75 != 33 || live.Player.Field3680 != 3 || live.Player.CameraFollowObj != source || target.HealthData.Cur != 200 {
						t.Fatal("cached marker/live carry or single ObserveClear lost")
					}
				})
			}
		}
	})
}

func TestPlayerDamageImpactPrefixLateTail4E17B0(t *testing.T) {
	bitePrefixCases4E17B0(t, func(t *testing.T, observe bool) {
		for _, clearMarker := range []bool{false, true} {
			for _, mode := range []string{"quest", "god-player", "god-nonplayer"} {
				t.Run(fmt.Sprintf("clear-marker-%t/%s", clearMarker, mode), func(t *testing.T) {
					target, source, missile, cached, r := impactPrefixFixture4E17B0(t, observe)
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
					r.DamageArmor = func(item, attacker, attack *Object, amount int32, typ object.DamageType) bool {
						if item != armor || attacker != source || attack != missile || amount != 1 || typ != object.DamageImpact || cached.Field21 != math.Float32bits(0.25) || cached.Field76 != 1 || cached.Field75 != 801 {
							t.Fatal("wear lost entry armor/live carry/attribution order")
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
							t.Fatal("GodMode preceded armor")
						}
						events = append(events, "god")
						return mode != "quest"
					}
					r.QuestMode = func() bool {
						if !worn {
							t.Fatal("QuestMode cached before wear")
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
					h, result := PlayerDamageNative4E17B0(target, source, missile, 5, object.DamageImpact, r)
					want := []string{"excluded", "direction", "armor", "god"}
					if mode != "god-player" {
						want = append(want, "quest", "scale", "buff", "hp")
					}
					marker, markerType := uint32(9), uint32(123)
					if clearMarker {
						marker = 0
						if mode != "god-nonplayer" {
							marker, markerType = 2, uint32(object.DamageImpact)
						}
					}
					if !h || !result || !reflect.DeepEqual(events, want) || cached.Field76 != marker || cached.Field75 != markerType || cached.Field21 != math.Float32bits(0.25) || target.HealthData.Cur != 200 || armor.HealthData.Cur != 10 {
						t.Fatalf("late IMPACT=%t/%t events=%v want=%v marker=%d/%d", h, result, events, want, cached.Field76, cached.Field75)
					}
				})
			}
		}
	})
}

func TestPlayerDamageImpactPrefixAdmission4E17B0(t *testing.T) {
	bitePrefixCases4E17B0(t, func(t *testing.T, observe bool) {
		for _, service := range []string{"exclusion", "direction", "observe"} {
			if service == "observe" && !observe {
				continue
			}
			t.Run(service, func(t *testing.T) {
				target, source, missile, cached, r := impactPrefixFixture4E17B0(t, observe)
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
				h, result := PlayerDamageNative4E17B0(target, source, missile, 5, object.DamageImpact, r)
				if h || result || reason != "missing player impact prefix service" || *target != beforeObj || *cached != beforeUD || *cached.Player != beforePlayer || *target.HealthData != beforeHP {
					t.Fatalf("admission=%t/%t reason=%q", h, result, reason)
				}
			})
		}
	})
}

func TestPlayerDamageImpactPrefixBlockUnusedTail4E17B0(t *testing.T) {
	bitePrefixCases4E17B0(t, func(t *testing.T, observe bool) {
		for _, defense := range []string{"shield", "greatsword"} {
			t.Run(defense, func(t *testing.T) {
				target, source, missile, cached, r := impactPrefixFixture4E17B0(t, observe)
				var events []string
				if defense == "shield" {
					cached.Player.ArmorEquip, cached.State = 0x1000000, PlayerState16
					shield := &Object{ObjClass: object.ClassArmor, ObjSubClass: 2, ObjFlags: object.FlagEquipped}
					target.InvFirstItem = shield
					impactPrefixShieldServices4E17B0(t, target, source, missile, shield, 5, &events, &r)
				} else {
					cached.Player.WeaponEquip = 0x400
					sword := &Object{ObjClass: object.ClassWeapon, ObjSubClass: 0x400, ObjFlags: object.FlagEquipped}
					target.InvFirstItem = sword
					r.Melee.CanDamageBlockWeapon = func(item *Object) bool {
						if item != sword {
							t.Fatal("wrong sword")
						}
						events = append(events, "admission")
						return true
					}
					r.ProjectileReflect = func(attack, owner *Object) {
						if attack != missile || owner != target || cached.Field76 != 1 || cached.Field75 != 801 {
							t.Fatal("sword reflect before attributed prefix")
						}
						events = append(events, "reflect")
					}
					r.ClearOwner = func(*Object) { events = append(events, "clear-owner") }
					r.SetOwner = func(*Object, *Object) { events = append(events, "set-owner") }
					r.Audio = func(id int, obj *Object) {
						if id != 890 || obj != target {
							t.Fatal("wrong sword audio")
						}
						events = append(events, "audio")
					}
					r.Melee.RandomInt = func(lo, hi int) int {
						if lo != 18 || hi != 20 {
							t.Fatal("sword random range")
						}
						events = append(events, "random")
						return 19
					}
					r.PlayerSetState = func(obj *Object, state PlayerState) bool {
						if obj != target || state != PlayerState19 {
							t.Fatal("sword stance")
						}
						events = append(events, "state")
						return true
					}
					r.BlockDamagePercent = func() float64 { events = append(events, "percent"); return 0.25 }
					r.Melee.DamageBlockWeapon = func(item, owner, attacker, attack *Object, amount float32, typ object.DamageType) bool {
						if item != sword || owner != target || attacker != source || attack != missile || amount != 1.25 || typ != object.DamageImpact {
							t.Fatal("sword wear arguments")
						}
						events = append(events, "wear")
						return true
					}
				}
				r.BlockDirection = func(*Object, types.Pointf) bool { return true }
				r.DefaultDamage, r.DamageClear, r.BuffOff, r.QuestDamageScale, r.ItemArmorValue = nil, nil, nil, nil, nil
				r.QuestMode = func() bool { t.Fatal("block evaluated unused Quest tail"); return true }
				r.GodMode = func() bool { t.Fatal("block evaluated unused GodMode tail"); return true }
				h, result := PlayerDamageNative4E17B0(target, source, missile, 5, object.DamageImpact, r)
				want := []string{"block-admission", "block-audio", "reflect", "clear-owner", "set-owner", "block-percent", "block-wear"}
				if defense == "greatsword" {
					want = []string{"admission", "reflect", "clear-owner", "set-owner", "audio", "random", "state", "percent", "admission", "wear"}
				}
				if !h || result || !reflect.DeepEqual(events, want) || cached.Field76 != 1 || cached.Field75 != 801 || cached.Field21 != math.Float32bits(0.125) || target.HealthData.Cur != 200 {
					t.Fatalf("block=%t/%t events=%v want=%v marker=%d/%d", h, result, events, want, cached.Field76, cached.Field75)
				}
			})
		}
	})
}

func TestPlayerDamageImpactPrefixReflect4E17B0(t *testing.T) {
	bitePrefixCases4E17B0(t, func(t *testing.T, observe bool) {
		modes := []string{"front", "rear", "missing-effect", "live-class"}
		if observe {
			modes = append(modes, "enable", "disable")
		}
		for _, mode := range modes {
			t.Run(mode, func(t *testing.T) {
				target, source, missile, cached, r := impactPrefixFixture4E17B0(t, observe)
				target.Buffs |= 1 << playerDamageReflectEnchant4E17B0
				if mode == "enable" {
					target.Buffs = 0
				}
				var events []string
				clear := r.ObserveClear
				r.ObserveClear = func(obj *Object) {
					clear(obj)
					events = append(events, "observe")
					if mode == "enable" {
						target.Buffs |= 1 << playerDamageReflectEnchant4E17B0
					}
					if mode == "disable" {
						target.Buffs = 0
					}
					if mode == "live-class" {
						target.ObjClass = object.ClassSimple
					}
				}
				if mode == "live-class" && !observe {
					target.Buffs = 0
				}
				r.BlockDirection = func(obj *Object, pos types.Pointf) bool {
					if obj != target {
						t.Fatal("wrong facing target")
					}
					if pos == missile.PosVec {
						if cached.Field76 != 0 || cached.Field75 != 77 {
							t.Fatal("Reflect facing ran before entry clear")
						}
						events = append(events, "reflect-facing")
						return mode != "rear"
					}
					if pos != missile.PrevPos || cached.Field76 != 1 || cached.Field75 != 801 {
						t.Fatal("ordinary facing lost distinct projectile attribution")
					}
					events = append(events, "ordinary-facing")
					return false
				}
				r.BlockSourceExcluded = func(*Object) bool {
					if cached.Field76 != 0 {
						t.Fatal("exclusion after attribution")
					}
					events = append(events, "excluded")
					return false
				}
				r.ProjectileReflect = func(attack, owner *Object) {
					if attack != missile || owner != target {
						t.Fatal("reflection arguments")
					}
					events = append(events, "reflect")
				}
				r.ClearOwner = func(attack *Object) {
					if attack != missile {
						t.Fatal("owner clear arguments")
					}
					events = append(events, "clear-owner")
				}
				r.SetOwner = func(owner, attack *Object) {
					if owner != target || attack != missile {
						t.Fatal("owner set arguments")
					}
					events = append(events, "set-owner")
				}
				r.Audio = func(id int, obj *Object) {
					if id != 122 || obj != target {
						t.Fatal("Reflect audio")
					}
					events = append(events, "audio")
				}
				if mode == "missing-effect" {
					r.ProjectileReflect = nil
				}
				r.DamageClear = func(obj *Object, amount int32) {
					if obj != target || amount != 4 {
						t.Fatal("wrong unblocked HP argument")
					}
					events = append(events, "hp")
				}
				reason := ""
				r.Unsupported = func(why string, _, _, _ *Object, _ int32, _ object.DamageType) { reason = why }
				h, result := PlayerDamageNative4E17B0(target, source, missile, 5, object.DamageImpact, r)
				var want []string
				if observe {
					want = append(want, "observe")
				}
				marker, markerType, carry := uint32(0), uint32(77), float32(0.125)
				if mode == "live-class" && observe {
					if h || result || reason != "unsupported live possessed player record" {
						t.Fatalf("ObserveClear guard=%t/%t reason=%q", h, result, reason)
					}
				} else if mode == "rear" || mode == "disable" || mode == "live-class" {
					if mode == "rear" {
						want = append(want, "reflect-facing")
					}
					want = append(want, "excluded", "ordinary-facing", "hp")
					marker, markerType, carry = 1, 801, -0.125
					if !h || !result || reason != "" {
						t.Fatalf("unblocked=%t/%t reason=%q", h, result, reason)
					}
				} else {
					want = append(want, "reflect-facing")
					if mode == "missing-effect" {
						if h || result || reason != "missing Reflect Shield effect service" {
							t.Fatalf("Reflect admission=%t/%t reason=%q", h, result, reason)
						}
					} else {
						want = append(want, "reflect", "clear-owner", "set-owner", "audio")
						if !h || result || reason != "" {
							t.Fatalf("reflected=%t/%t reason=%q", h, result, reason)
						}
					}
				}
				if !reflect.DeepEqual(events, want) || cached.Field76 != marker || cached.Field75 != markerType || cached.Field21 != math.Float32bits(carry) || target.HealthData.Cur != 200 {
					t.Fatalf("Reflect events=%v want=%v marker=%d/%d carry=%g", events, want, cached.Field76, cached.Field75, math.Float32frombits(cached.Field21))
				}
			})
		}
	})
}

func TestPlayerDamageImpactPrefixEarlyGate4E17B0(t *testing.T) {
	bitePrefixCases4E17B0(t, func(t *testing.T, observe bool) {
		for _, gate := range []string{"no-update", "dead", "invulnerable", "observer-bit-one", "coop-self"} {
			t.Run(gate, func(t *testing.T) {
				target, source, missile, cached, r := impactPrefixFixture4E17B0(t, observe)
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
				h, result := PlayerDamageNative4E17B0(target, source, missile, 5, object.DamageImpact, r)
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

func TestPlayerDamageImpactPrefixShapeBoundary4E17B0(t *testing.T) {
	bitePrefixCases4E17B0(t, func(t *testing.T, observe bool) {
		for _, shape := range []string{"zero", "negative", "nil-source", "nil-weapon", "self-weapon", "player-source", "nil-source-update", "non-missile"} {
			t.Run(shape, func(t *testing.T) {
				target, source, missile, cached, r := impactPrefixFixture4E17B0(t, observe)
				weapon, raw := missile, int32(5)
				switch shape {
				case "zero":
					raw = 0
				case "negative":
					raw = -1
				case "nil-source":
					source = nil
				case "nil-weapon":
					weapon = nil
				case "self-weapon":
					weapon = source
				case "player-source":
					source.ObjClass = object.ClassPlayer
				case "nil-source-update":
					source.UpdateData = nil
				case "non-missile":
					missile.ObjClass = object.ClassSimple
				}
				beforeObj, beforeUD, beforePlayer, beforeHP := *target, *cached, *cached.Player, *target.HealthData
				reason := ""
				r.Unsupported = func(why string, _, _, _ *Object, _ int32, _ object.DamageType) { reason = why }
				r.ObserveClear = func(*Object) { t.Fatal("unsupported shape entered prefix") }
				r.BlockSourceExcluded = func(*Object) bool { t.Fatal("unsupported shape entered prefix"); return false }
				r.BlockDirection = func(*Object, types.Pointf) bool { t.Fatal("unsupported shape entered prefix"); return false }
				h, result := PlayerDamageNative4E17B0(target, source, weapon, raw, object.DamageImpact, r)
				if h || result || reason != "unsupported player damage shape" || *target != beforeObj || *cached != beforeUD || *cached.Player != beforePlayer || *target.HealthData != beforeHP {
					t.Fatalf("shape boundary=%t/%t reason=%q", h, result, reason)
				}
			})
		}
	})
}
