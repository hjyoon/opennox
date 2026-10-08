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

// Ordered service boundaries, not injected HP/durability results. The native
// companion exercises the registered C dispatcher and real HP/status services.
func chargePrefixFixture4E17B0(t *testing.T, observe bool) (*Object, *Object, *PlayerUpdateData, PlayerDamageRuntime4E17B0) {
	t.Helper()
	target, source, cached, r := bitePrefixFixture4E17B0(t, observe)
	source.ObjClass = object.ClassPlayer
	source.UpdateData = unsafe.Pointer(&PlayerUpdateData{Player: &Player{PlayerInd: 9}, State: PlayerState13})
	r.GameplayFlag1 = func() bool { return true }
	r.PlayerSetState = func(*Object, PlayerState) bool { t.Fatal("unexpected charge hurt/block state"); return false }
	return target, source, cached, r
}

func chargePrefixShieldServices4E17B0(t *testing.T, target, source, shield *Object, raw int32, events *[]string, r *PlayerDamageRuntime4E17B0) {
	t.Helper()
	r.CanDamageBlockItem = func(item *Object) bool {
		if item != shield {
			t.Fatal("cached shield used instead of live inventory")
		}
		*events = append(*events, "block-admission")
		return true
	}
	r.Audio = func(id int, obj *Object) {
		if id != 878 || obj != target {
			t.Fatal("charge block sound")
		}
		*events = append(*events, "block-audio")
	}
	r.BlockDamagePercent = func() float64 { *events = append(*events, "block-percent"); return 0.25 }
	r.DamageBlockItem = func(item, owner, attacker, weapon *Object, amount float32, typ object.DamageType) bool {
		if item != shield || owner != target || attacker != source || weapon != source || amount != float32(raw)*0.25 || typ != object.DamageCrush {
			t.Fatal("charge shield arguments")
		}
		*events = append(*events, "block-wear")
		return true
	}
}

func TestPlayerDamageChargePrefixOrder4E17B0(t *testing.T) {
	bitePrefixCases4E17B0(t, func(t *testing.T, observe bool) {
		for _, shielded := range []bool{false, true} {
			for _, facing := range []string{"front", "rear", "excluded"} {
				for _, raw := range []int32{1, 5, 21} {
					t.Run(fmt.Sprintf("shield-%t/%s/raw-%d", shielded, facing, raw), func(t *testing.T) {
						target, source, cached, r := chargePrefixFixture4E17B0(t, observe)
						shield := &Object{ObjSubClass: 2, ObjFlags: object.FlagEquipped, HealthData: &HealthData{Cur: 10, Max: 10}, InitData: unsafe.Pointer(new(ModifierInitData))}
						if shielded {
							cached.Player.ArmorEquip, target.InvFirstItem = 0x1000000, shield
						}
						var events []string
						wantPos := source.PrevPos
						if observe {
							wantPos = types.Ptf(30, 7)
						}
						clear := r.ObserveClear
						r.ObserveClear = func(obj *Object) { clear(obj); events = append(events, "observe"); source.PrevPos = wantPos }
						r.BlockSourceExcluded = func(obj *Object) bool {
							if obj != source || cached.Field76 != 0 || cached.Field75 != 77 {
								t.Fatal("self-weapon charge prefix must precede exclusion")
							}
							events = append(events, "excluded")
							source.PrevPos, source.TypeInd = types.Ptf(-100, 99), 888
							cached.Field57, cached.Field21, cached.Player.ArmorEquip = math.Float32bits(0.875), math.Float32bits(0.5), 0
							return facing == "excluded"
						}
						r.BlockDirection = func(obj *Object, pos types.Pointf) bool {
							if obj != target || pos != wantPos || cached.Field76 != 0 || cached.Field75 != 77 {
								t.Fatal("charge facing lost position snapshot/self-weapon clear")
							}
							events = append(events, "direction")
							cached.State, cached.Player.WeaponEquip = PlayerState16, 0x400
							return facing == "front"
						}
						chargePrefixShieldServices4E17B0(t, target, source, shield, raw, &events, &r)
						r.GodMode = func() bool { events = append(events, "god"); return false }
						r.QuestMode = func() bool { events = append(events, "quest"); return false }
						r.BuffOff = func(obj *Object, id EnchantID) {
							if obj != target || id != playerDamageInvisibleEnchant4E17B0 {
								t.Fatal("charge BuffOff arguments")
							}
							events = append(events, "buff")
						}
						accumulated := float32(0.875*float64(raw)) + 0.5
						effective := playerDamageRound4E17B0(accumulated)
						r.DamageClear = func(obj *Object, amount int32) {
							if obj != target || amount != effective {
								t.Fatalf("charge HP argument=%d want=%d", amount, effective)
							}
							events = append(events, "hp")
						}
						h, result := PlayerDamageNative4E17B0(target, source, source, raw, object.DamageCrush, r)
						blocked := shielded && facing == "front"
						want := []string{"excluded"}
						if observe {
							want = append([]string{"observe"}, want...)
						}
						if facing != "excluded" {
							want = append(want, "direction")
						}
						carry, marker, markerType := accumulated-float32(effective), uint32(2), uint32(object.DamageCrush)
						if blocked {
							want = append(want, "block-admission", "block-audio", "block-percent", "block-wear")
							carry, marker, markerType = 0.5, 0, 77
						} else {
							want = append(want, "god", "quest", "buff", "hp")
						}
						if !h || result == blocked || !reflect.DeepEqual(events, want) || cached.Field76 != marker || cached.Field75 != markerType || cached.Field21 != math.Float32bits(carry) || *target.HealthData != (HealthData{Cur: 200, Field2: 200, Max: 200}) || *shield.HealthData != (HealthData{Cur: 10, Max: 10}) {
							t.Fatalf("charge order=%t/%t events=%v want=%v marker=%d/%d carry=%g", h, result, events, want, cached.Field76, cached.Field75, math.Float32frombits(cached.Field21))
						}
					})
				}
			}
		}
	})
}

func TestPlayerDamageChargePrefixLiveRecord4E17B0(t *testing.T) {
	bitePrefixCases4E17B0(t, func(t *testing.T, observe bool) {
		for _, stage := range []string{"exclusion", "direction"} {
			for _, change := range []string{"replacement-front", "replacement-rear", "class", "nil-update"} {
				t.Run(stage+"/"+change, func(t *testing.T) {
					target, source, cached, r := chargePrefixFixture4E17B0(t, observe)
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
					chargePrefixShieldServices4E17B0(t, target, source, shield, 5, &events, &r)
					calls, reason := 0, ""
					r.DamageClear = func(obj *Object, amount int32) {
						if change != "replacement-rear" || obj != target || amount != 5 {
							t.Fatal("invalid charge live carry/record reached HP")
						}
						calls++
					}
					r.Unsupported = func(why string, _, _, _ *Object, _ int32, _ object.DamageType) { reason = why }
					h, result := PlayerDamageNative4E17B0(target, source, source, 5, object.DamageCrush, r)
					marker, markerType := uint32(0), uint32(77)
					switch change {
					case "replacement-front":
						if !h || result || calls != 0 || reason != "" || len(events) != 4 || live.Field21 != math.Float32bits(0.5) {
							t.Fatal("charge block lost cached mask/stance/live inventory")
						}
					case "replacement-rear":
						marker, markerType = 2, uint32(object.DamageCrush)
						if !h || !result || calls != 1 || reason != "" || len(events) != 0 || live.Field21 != math.Float32bits(-0.125) {
							t.Fatal("charge tail lost entry half-armor/live carry")
						}
					default:
						if h || result || reason != "unsupported live player charge record" || calls != 0 || len(events) != 0 {
							t.Fatalf("charge live guard=%t/%t reason=%q events=%v", h, result, reason, events)
						}
					}
					if cached.Field76 != marker || cached.Field75 != markerType || cached.Field21 != math.Float32bits(0.125) || live.Field76 != 31 || live.Field75 != 33 || live.Player.Field3680 != 3 || live.Player.CameraFollowObj != source || target.HealthData.Cur != 200 {
						t.Fatal("charge cached marker/live carry or single ObserveClear lost")
					}
				})
			}
		}
	})
}

func TestPlayerDamageChargePrefixLateTail4E17B0(t *testing.T) {
	bitePrefixCases4E17B0(t, func(t *testing.T, observe bool) {
		for _, clearMarker := range []bool{false, true} {
			for _, mode := range []string{"quest", "god-player", "god-nonplayer"} {
				t.Run(fmt.Sprintf("clear-marker-%t/%s", clearMarker, mode), func(t *testing.T) {
					target, source, cached, r := chargePrefixFixture4E17B0(t, observe)
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
					r.ItemArmorValue = func(*Object) float32 { return 0.5 }
					r.CanDamageArmor = func(item *Object) bool { return item == armor }
					r.DamageArmor = func(item, attacker, weapon *Object, amount int32, typ object.DamageType) bool {
						if item != armor || attacker != source || weapon != source || amount != 2 || typ != object.DamageCrush {
							t.Fatalf("charge wear=%d, want cached half-armor remaining 2", amount)
						}
						worn, quest = true, true
						cached.Field76, cached.Field75 = 9, 123
						if clearMarker {
							cached.Field76, cached.Field75 = 0, 77
						}
						if mode == "god-nonplayer" {
							target.ObjClass = object.ClassSimple
						}
						events = append(events, "wear")
						return true
					}
					r.GodMode = func() bool {
						if !worn {
							t.Fatal("GodMode before charge wear")
						}
						events = append(events, "god")
						return mode != "quest"
					}
					r.QuestMode = func() bool {
						if !worn {
							t.Fatal("Quest before charge wear/GodMode")
						}
						events = append(events, "quest")
						return quest
					}
					r.QuestDamageScale = func() float32 { events = append(events, "scale"); return 0.5 }
					r.BuffOff = func(*Object, EnchantID) { events = append(events, "buff") }
					calls := 0
					r.DamageClear = func(obj *Object, amount int32) {
						if obj != target || amount != 10 {
							t.Fatalf("late Quest HP argument=%d want=10", amount)
						}
						calls++
						events = append(events, "hp")
					}
					h, result := PlayerDamageNative4E17B0(target, source, source, 21, object.DamageCrush, r)
					want := []string{"excluded", "direction", "wear", "god"}
					if mode != "god-player" {
						want = append(want, "quest", "scale", "buff", "hp")
					}
					marker, markerType := uint32(9), uint32(123)
					if clearMarker {
						marker, markerType = 2, uint32(object.DamageCrush)
						if mode == "god-nonplayer" {
							marker, markerType = 0, 77
						}
					}
					wantCalls := 1
					if mode == "god-player" {
						wantCalls = 0
					}
					if !h || !result || calls != wantCalls || !reflect.DeepEqual(events, want) || cached.Field76 != marker || cached.Field75 != markerType || cached.Field21 != math.Float32bits(-0.125) || target.HealthData.Cur != 200 || armor.HealthData.Cur != 10 {
						t.Fatalf("charge late tail=%t/%t events=%v marker=%d/%d carry=%g", h, result, events, cached.Field76, cached.Field75, math.Float32frombits(cached.Field21))
					}
				})
			}
		}
	})
}

func TestPlayerDamageChargePrefixMissingService4E17B0(t *testing.T) {
	bitePrefixCases4E17B0(t, func(t *testing.T, observe bool) {
		services := []string{"exclusion", "direction"}
		if observe {
			services = append(services, "observe")
		}
		for _, service := range services {
			t.Run(service, func(t *testing.T) {
				target, source, cached, r := chargePrefixFixture4E17B0(t, observe)
				switch service {
				case "exclusion":
					r.BlockSourceExcluded = nil
				case "direction":
					r.BlockDirection = nil
				case "observe":
					r.ObserveClear = nil
				}
				reason, calls := "", 0
				r.Unsupported = func(why string, _, _, _ *Object, _ int32, _ object.DamageType) { reason = why }
				r.DamageClear = func(*Object, int32) { calls++ }
				before, beforeUpdate, beforePlayer, beforeSource, beforeSourceUpdate, beforeHP := *target, *cached, *cached.Player, *source, *source.UpdateDataPlayer(), *target.HealthData
				h, result := PlayerDamageNative4E17B0(target, source, source, 5, object.DamageCrush, r)
				if h || result || calls != 0 || reason != "missing player charge prefix service" || *target != before || *cached != beforeUpdate || *cached.Player != beforePlayer || *source != beforeSource || *source.UpdateDataPlayer() != beforeSourceUpdate || *target.HealthData != beforeHP {
					t.Fatalf("missing charge prefix=%t/%t reason=%q calls=%d", h, result, reason, calls)
				}
			})
		}
	})
}

func TestPlayerDamageChargePrefixBlockWithoutTail4E17B0(t *testing.T) {
	bitePrefixCases4E17B0(t, func(t *testing.T, observe bool) {
		target, source, cached, r := chargePrefixFixture4E17B0(t, observe)
		cached.Player.ArmorEquip, cached.State = 0x1000000, PlayerState16
		shield := &Object{ObjSubClass: 2, ObjFlags: object.FlagEquipped, HealthData: &HealthData{Cur: 10, Max: 10}, InitData: unsafe.Pointer(new(ModifierInitData))}
		target.InvFirstItem = shield
		var events []string
		chargePrefixShieldServices4E17B0(t, target, source, shield, 21, &events, &r)
		r.BlockDirection = func(*Object, types.Pointf) bool { return true }
		r.GameplayFlag1, r.IsEnemy, r.DefaultDamage, r.DamageClear, r.BuffOff, r.ItemArmorValue, r.QuestDamageScale = nil, nil, nil, nil, nil, nil, nil
		r.GodMode = func() bool { t.Fatal("blocked charge read GodMode"); return false }
		r.QuestMode = func() bool { t.Fatal("blocked charge read Quest"); return false }
		h, result := PlayerDamageNative4E17B0(target, source, source, 21, object.DamageCrush, r)
		if !h || result || !reflect.DeepEqual(events, []string{"block-admission", "block-audio", "block-percent", "block-wear"}) || cached.Field76 != 0 || cached.Field75 != 77 || cached.Field21 != math.Float32bits(0.125) || target.HealthData.Cur != 200 || shield.HealthData.Cur != 10 {
			t.Fatalf("charge block depended on unused HP/friendly-fire/armor/Quest tail: %t/%t %v", h, result, events)
		}
	})
}

func TestPlayerDamageChargePrefixEarlyGate4E17B0(t *testing.T) {
	bitePrefixCases4E17B0(t, func(t *testing.T, observe bool) {
		for _, gate := range []string{"no-update", "dead", "invulnerable", "observer-bit", "coop-self"} {
			t.Run(gate, func(t *testing.T) {
				target, source, cached, r := chargePrefixFixture4E17B0(t, observe)
				switch gate {
				case "no-update":
					target.ObjFlags |= object.FlagNoUpdate
				case "dead":
					target.ObjFlags |= object.FlagDead
				case "invulnerable":
					target.Buffs |= 1 << playerDamageInvulnerableEnchant4E17B0
				case "observer-bit":
					cached.Player.Field3680 |= 1
				case "coop-self":
					source = target
					r.CoopMode = func() bool { return true }
				}
				audios := 0
				r.Audio = func(id int, obj *Object) {
					if gate != "invulnerable" || id != playerDamageInvulnerableSound4E17B0 || obj != target {
						t.Fatal("early charge sound")
					}
					audios++
				}
				r.BlockSourceExcluded = func(*Object) bool { t.Fatal("early gate reached charge prefix"); return false }
				before, beforeUpdate, beforePlayer, beforeHP := *target, *cached, *cached.Player, *target.HealthData
				h, result := PlayerDamageNative4E17B0(target, source, source, 5, object.DamageCrush, r)
				wantSound := 0
				if gate == "invulnerable" {
					wantSound = 1
				}
				if !h || result != (gate == "invulnerable") || audios != wantSound || *target != before || *cached != beforeUpdate || *cached.Player != beforePlayer || *target.HealthData != beforeHP {
					t.Fatalf("charge early gate=%t/%t sound=%d", h, result, audios)
				}
			})
		}
	})
}

func TestPlayerDamageChargePrefixShapeBoundary4E17B0(t *testing.T) {
	bitePrefixCases4E17B0(t, func(t *testing.T, observe bool) {
		// Pure MONSTER self-weapons have their own native CRUSH route;
		// PlayerDamageMonsterSelfStrike tests cover its positive admission.
		for _, shape := range []string{"zero", "negative", "nil-source", "player-monster", "player-weapon", "player-wand", "simple"} {
			t.Run(shape, func(t *testing.T) {
				target, source, cached, r := chargePrefixFixture4E17B0(t, observe)
				sourceUpdate := source.UpdateDataPlayer()
				attacker, raw := source, int32(5)
				switch shape {
				case "zero":
					raw = 0
				case "negative":
					raw = -5
				case "nil-source":
					attacker = nil
				case "player-monster":
					source.ObjClass |= object.ClassMonster
				case "player-weapon":
					source.ObjClass |= object.ClassWeapon
				case "player-wand":
					source.ObjClass |= object.ClassWand
				case "simple":
					source.ObjClass = object.ClassSimple
				}
				reason := ""
				r.Unsupported = func(why string, _, _, _ *Object, _ int32, _ object.DamageType) { reason = why }
				r.BlockSourceExcluded = func(*Object) bool { t.Fatal("unsupported shape entered charge prefix"); return false }
				before, beforeUpdate, beforePlayer, beforeSource, beforeSourceUpdate, beforeHP := *target, *cached, *cached.Player, *source, *sourceUpdate, *target.HealthData
				h, result := PlayerDamageNative4E17B0(target, attacker, attacker, raw, object.DamageCrush, r)
				if h || result || reason != "unsupported player damage shape" || *target != before || *cached != beforeUpdate || *cached.Player != beforePlayer || *source != beforeSource || *sourceUpdate != beforeSourceUpdate || *target.HealthData != beforeHP {
					t.Fatalf("charge boundary=%t/%t reason=%q", h, result, reason)
				}
			})
		}
	})
}

func TestPlayerDamageChargePrefixReflectDoesNotBlockPlayer4E17B0(t *testing.T) {
	bitePrefixCases4E17B0(t, func(t *testing.T, observe bool) {
		target, source, cached, r := chargePrefixFixture4E17B0(t, observe)
		target.Buffs |= 1 << playerDamageReflectEnchant4E17B0
		var events []string
		clear := r.ObserveClear
		r.ObserveClear = func(obj *Object) { clear(obj); events = append(events, "observe") }
		r.BlockSourceExcluded = func(obj *Object) bool {
			if obj != source || cached.Field76 != 0 {
				t.Fatal("charge prefix before Reflect non-missile path")
			}
			events = append(events, "excluded")
			return false
		}
		r.BlockDirection = func(obj *Object, pos types.Pointf) bool {
			if obj != target || pos != source.PrevPos {
				t.Fatal("Reflect used current position for a non-missile charge")
			}
			events = append(events, "direction")
			return false
		}
		r.Audio = func(int, *Object) { t.Fatal("Reflect intercepted player self-weapon CRUSH") }
		r.DamageClear = func(obj *Object, amount int32) {
			if obj != target || amount != 4 {
				t.Fatal("charge half-armor value")
			}
			events = append(events, "hp")
		}
		h, result := PlayerDamageNative4E17B0(target, source, source, 5, object.DamageCrush, r)
		want := []string{"excluded", "direction", "hp"}
		if observe {
			want = append([]string{"observe"}, want...)
		}
		if !h || !result || !reflect.DeepEqual(events, want) || cached.Field76 != 2 || cached.Field75 != uint32(object.DamageCrush) || cached.Field21 != math.Float32bits(0.5) || target.HealthData.Cur != 200 {
			t.Fatalf("charge/Reflect=%t/%t events=%v", h, result, events)
		}
	})
}

func TestPlayerDamageChargePrefixLiveHurt4E17B0(t *testing.T) {
	bitePrefixCases4E17B0(t, func(t *testing.T, observe bool) {
		for _, mode := range []string{"cached-1/live-13", "cached-13/live-1", "cached-13/live-15", "non-player", "nil-update"} {
			t.Run(mode, func(t *testing.T) {
				target, source, cached, r := chargePrefixFixture4E17B0(t, observe)
				live := &PlayerUpdateData{Player: &Player{PlayerInd: 7}, State: PlayerState13}
				var events []string
				reason := ""
				r.Unsupported = func(why string, _, _, _ *Object, _ int32, _ object.DamageType) {
					reason = why
					events = append(events, "unsupported")
				}
				r.GameBallOnDamage = func(attacker, obj *Object, amount int32) {
					if attacker != source || obj != target || amount != 54 || cached.Field76 != 2 || cached.Field75 != uint32(object.DamageCrush) || cached.Field21 != math.Float32bits(-0.5) {
						t.Fatal("charge live hurt must follow armor/carry/marker and the ordinary damage tail")
					}
					events = append(events, "game-ball")
					target.UpdateData = unsafe.Pointer(live)
					switch mode {
					case "cached-1/live-13":
						cached.State = PlayerState1
					case "cached-13/live-1":
						live.State = PlayerState1
					case "cached-13/live-15":
						live.State = PlayerState15
					case "non-player":
						target.ObjClass = object.ClassSimple
					case "nil-update":
						target.UpdateData = nil
					}
				}
				r.PlayerSetState = func(obj *Object, state PlayerState) bool {
					if mode != "cached-1/live-13" || obj != target || state != PlayerState30 || obj.UpdateData != unsafe.Pointer(live) || live.State != PlayerState13 || cached.State != PlayerState1 {
						t.Fatal("charge hurt used cached state/class instead of the post-tail live record")
					}
					events = append(events, "hurt")
					live.State = state
					return true
				}
				r.DamageClear = func(obj *Object, amount int32) {
					if obj != target || amount != 54 || mode == "nil-update" {
						t.Fatal("charge HP after live hurt")
					}
					events = append(events, "hp")
				}
				h, result := PlayerDamageNative4E17B0(target, source, source, 61, object.DamageCrush, r)
				want := []string{"game-ball", "hp"}
				wantState := PlayerState13
				switch mode {
				case "cached-1/live-13":
					want, wantState = []string{"game-ball", "hurt", "hp"}, PlayerState30
				case "cached-13/live-1":
					wantState = PlayerState1
				case "cached-13/live-15":
					wantState = PlayerState15
				case "nil-update":
					want = []string{"game-ball", "unsupported"}
				}
				wantHandled := mode != "nil-update"
				wantReason := ""
				if mode == "nil-update" {
					wantReason = "unsupported live player charge hurt record"
				}
				if h != wantHandled || result != wantHandled || reason != wantReason || !reflect.DeepEqual(events, want) || live.State != wantState || live.Field21 != 0 || live.Field76 != 0 || live.Field75 != 0 || target.HealthData.Cur != 200 || cached.Field21 != math.Float32bits(-0.5) || cached.Field76 != 2 || cached.Field75 != uint32(object.DamageCrush) {
					t.Fatalf("charge live hurt=%t/%t mode=%s events=%v reason=%q state=%d", h, result, mode, events, reason, live.State)
				}
			})
		}
	})
}

func TestPlayerDamageChargePrefixLateHurtService4E17B0(t *testing.T) {
	bitePrefixCases4E17B0(t, func(t *testing.T, observe bool) {
		target, source, cached, r := chargePrefixFixture4E17B0(t, observe)
		r.PlayerSetState = nil
		var events []string
		r.QuestMode = func() bool {
			if cached.Field76 != 2 || cached.Field21 != math.Float32bits(0.5) {
				t.Fatal("Quest reached charge before armor/carry/marker")
			}
			events = append(events, "quest")
			return true
		}
		r.QuestDamageScale = func() float32 { events = append(events, "scale"); return 10 }
		r.BuffOff = func(obj *Object, id EnchantID) {
			if obj != target || id != playerDamageInvisibleEnchant4E17B0 {
				t.Fatal("charge late hurt BuffOff")
			}
			events = append(events, "buff")
		}
		r.DamageClear = func(*Object, int32) { t.Fatal("missing live hurt service reached HP") }
		reason := ""
		r.Unsupported = func(why string, obj, attacker, weapon *Object, amount int32, typ object.DamageType) {
			if obj != target || attacker != source || weapon != source || amount != 40 || typ != object.DamageCrush {
				t.Fatal("late hurt admission used raw damage instead of post-Quest amount")
			}
			reason = why
			events = append(events, "unsupported")
		}
		h, result := PlayerDamageNative4E17B0(target, source, source, 5, object.DamageCrush, r)
		if h || result || reason != "missing live player charge hurt-state service" || !reflect.DeepEqual(events, []string{"quest", "scale", "buff", "unsupported"}) || target.HealthData.Cur != 200 || cached.Field76 != 2 || cached.Field75 != uint32(object.DamageCrush) || cached.Field21 != math.Float32bits(0.5) || cached.State != PlayerState13 {
			t.Fatalf("charge late hurt service=%t/%t reason=%q events=%v", h, result, reason, events)
		}
	})
}
