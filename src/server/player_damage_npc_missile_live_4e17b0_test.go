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

// Cases 3/11 share 004E1F84..004E2098. The Quest query at 004E2046 is
// after durability, the cached marker fallback and the positive minimum.
func TestPlayerDamageNPCMissileLive4E17B0QuestAfterArmor(t *testing.T) {
	for _, typ := range []object.DamageType{object.DamageImpale, object.DamageImpact} {
		for _, playerSource := range []bool{false, true} {
			if typ == object.DamageImpact && !playerSource {
				continue // HarpoonCollide supplies a player owner.
			}
			for _, pure := range []bool{false, true} {
				for _, entering := range []bool{false, true} {
					for _, inDefend := range []bool{false, true} {
						t.Run(fmt.Sprintf("type-%d/player-%t/pure-%t/entering-%t/defend-%t", typ, playerSource, pure, entering, inDefend), func(t *testing.T) {
							target, source, missile, r, damages := playerDamageNPCPierceSourceFixture4E17B0(t, playerSource)
							if pure {
								missile.ObjClass = object.ClassMissile
							}
							beforeSource, beforeMissile := *source, *missile
							var beforePlayer PlayerUpdateData
							if playerSource {
								beforePlayer = *source.UpdateDataPlayer()
							}
							armor := damageMeleeArmorFixture4E17B0(target, 0.25, 0.25)
							damageMeleeArmorRuntime4E17B0(&r, armor, 0.25)
							update := target.UpdateDataMonster()
							quest := !entering
							var events []string
							r.ItemArmorValue = func(v *Object) float32 {
								if v != armor {
									t.Fatal("armor definition identity")
								}
								events = append(events, "armor-value")
								return 0.25
							}
							if inDefend {
								modifier := &ModifierEff{Defend76: ModifierEffFnc{Fnc: unsafe.Pointer(new(byte))}}
								(*ModifierInitData)(armor.InitData).Modifiers[1] = modifier
								r.CanApplyArmorDefend = func(m *ModifierEff) bool { return m == modifier }
								r.ApplyArmorDefend = func(m *ModifierEff, item, owner, weapon, attacker *Object, amount *float32) bool {
									if m != modifier || item != armor || owner != target || weapon != missile || attacker != source || *amount != 2 {
										t.Fatal("armor Defend arguments")
									}
									events = append(events, "defend")
									quest = entering
									return true
								}
							}
							r.DamageArmor = func(item, attacker, weapon *Object, amount int32, gotType object.DamageType) bool {
								if item != armor || attacker != source || weapon != missile || amount != 2 || gotType != typ ||
									update.Field547 != 1 || update.Field546 != uint32(missile.TypeInd) || update.Field1 != math.Float32bits(0.25) {
									t.Fatal("raw-minus-effective wear or pre-Quest marker/carry")
								}
								events = append(events, "armor")
								if !inDefend {
									quest = entering
								}
								item.HealthData.Cur -= uint16(amount)
								return true
							}
							r.ReportArmorHealth = func(*Object, *Object, uint16, uint16) { t.Fatal("NPC reported player armor health") }
							r.QuestMode = func() bool {
								events = append(events, "quest-mode")
								return quest
							}
							r.QuestDamageScale = func() float32 {
								events = append(events, "quest-scale")
								return 0.5
							}
							tail := r.DefaultDamage
							r.DefaultDamage = func(v, a, w *Object, d int32, gotType object.DamageType) bool {
								want := int32(6)
								if entering {
									want = 3
								}
								if v != target || a != source || w != missile || d != want || gotType != typ {
									t.Fatalf("live Quest damage=%d, want %d", d, want)
								}
								events = append(events, "default")
								return tail(v, a, w, d, gotType)
							}
							wantDamage := int32(6)
							wantEvents := []string{"armor-value"}
							if inDefend {
								wantEvents = append(wantEvents, "defend")
							}
							wantEvents = append(wantEvents, "armor", "quest-mode")
							if entering {
								wantDamage = 3
								wantEvents = append(wantEvents, "quest-scale")
							}
							wantEvents = append(wantEvents, "default")
							if h, result := PlayerDamageNative4E17B0(target, source, missile, 8, typ, r); !h || !result ||
								!slices.Equal(events, wantEvents) || !slices.Equal(*damages, []int32{wantDamage}) ||
								target.HealthData.Cur != uint16(20-wantDamage) || armor.HealthData.Cur != 23 ||
								update.Field1 != math.Float32bits(0.25) || *source != beforeSource || *missile != beforeMissile {
								t.Fatalf("live Quest=%t/%t HP=%d armor=%d events=%v, want %v", h, result, target.HealthData.Cur, armor.HealthData.Cur, events, wantEvents)
							}
							if playerSource && *source.UpdateDataPlayer() != beforePlayer {
								t.Fatal("missile path treated the player source as a monster")
							}
						})
					}
				}
			}
		}
	}
}

func TestPlayerDamageNPCMissileLive4E17B0QuestAfterMarkerAndMinimum(t *testing.T) {
	for _, typ := range []object.DamageType{object.DamageImpale, object.DamageImpact} {
		t.Run(fmt.Sprint(typ), func(t *testing.T) {
			target, source, missile, r, _ := playerDamageNPCPierceSourceFixture4E17B0(t, true)
			armor := damageMeleeArmorFixture4E17B0(target, 1, 0)
			damageMeleeArmorRuntime4E17B0(&r, armor, 1)
			cached := target.UpdateDataMonster()
			live := &MonsterUpdateData{Field1: math.Float32bits(0.75), Field518: math.Float32bits(0.125)}
			var events []string
			r.DamageArmor = func(item, a, w *Object, d int32, gotType object.DamageType) bool {
				if item != armor || a != source || w != missile || d != 1 || gotType != typ || cached.Field547 != 1 || cached.Field1 != 0 {
					t.Fatal("minimum applied before full-armor wear")
				}
				events = append(events, "armor")
				cached.Field547 = 0
				target.UpdateData = unsafe.Pointer(live)
				return true
			}
			r.QuestMode = func() bool {
				if cached.Field547 != 2 || cached.Field546 != uint32(typ) || cached.Field1 != 0 || live.Field1 != math.Float32bits(0.75) {
					t.Fatal("Quest queried before the cached marker fallback")
				}
				events = append(events, "quest-mode")
				return true
			}
			r.QuestDamageScale = func() float32 { events = append(events, "quest-scale"); return 0.5 }
			r.DefaultDamage = func(v, a, w *Object, d int32, gotType object.DamageType) bool {
				if v != target || a != source || w != missile || d != 1 || gotType != typ || live.Field547 != 0 {
					t.Fatal("positive minimum before and after Quest or cached/live marker")
				}
				events = append(events, "default")
				return false
			}
			if h, result := PlayerDamageNative4E17B0(target, source, missile, 1, typ, r); !h || result ||
				!slices.Equal(events, []string{"armor", "quest-mode", "quest-scale", "default"}) || target.HealthData.Cur != 20 {
				t.Fatalf("minimum/false tail=%t/%t events=%v HP=%d", h, result, events, target.HealthData.Cur)
			}
		})
	}
}

func TestPlayerDamageNPCMissileLive4E17B0MissingQuestService(t *testing.T) {
	for _, typ := range []object.DamageType{object.DamageImpale, object.DamageImpact} {
		for _, initialQuest := range []bool{false, true} {
			t.Run(fmt.Sprintf("type-%d/initial-Quest-%t", typ, initialQuest), func(t *testing.T) {
				target, source, missile, r, _ := playerDamageNPCPierceSourceFixture4E17B0(t, true)
				armor := damageMeleeArmorFixture4E17B0(target, 0.25, 0.25)
				damageMeleeArmorRuntime4E17B0(&r, armor, 0.25)
				update := target.UpdateDataMonster()
				beforeTarget, beforeUpdate, beforeArmor, beforeHealth := *target, *update, *armor, *armor.HealthData
				beforeCarry := *(*WeaponArmorUpdateData)(armor.UpdateData)
				quest := initialQuest
				var events []string
				r.QuestMode = func() bool { events = append(events, "quest-mode"); return quest }
				r.QuestDamageScale = nil
				r.DamageArmor = func(item, a, w *Object, d int32, gotType object.DamageType) bool {
					if initialQuest || item != armor || a != source || w != missile || d != 2 || gotType != typ {
						t.Fatal("missing Quest service wear prefix")
					}
					events = append(events, "armor")
					quest = true
					item.HealthData.Cur -= uint16(d)
					return true
				}
				r.DefaultDamage = func(*Object, *Object, *Object, int32, object.DamageType) bool {
					t.Fatal("missing live Quest service reached HP")
					return false
				}
				r.Unsupported = func(reason string, v, a, w *Object, d int32, gotType object.DamageType) {
					want := "missing live quest damage service"
					if initialQuest {
						want = "missing quest damage service"
					}
					if reason != want || v != target || a != source || w != missile || d != 8 || gotType != typ {
						t.Fatalf("unsupported=%q, want %q", reason, want)
					}
					events = append(events, "unsupported")
				}
				if h, result := PlayerDamageNative4E17B0(target, source, missile, 8, typ, r); h || result || target.HealthData.Cur != 20 {
					t.Fatalf("missing-service result=%t/%t HP=%d", h, result, target.HealthData.Cur)
				}
				if initialQuest {
					if !slices.Equal(events, []string{"quest-mode", "unsupported"}) ||
						*target != beforeTarget || *update != beforeUpdate || *armor != beforeArmor || *armor.HealthData != beforeHealth ||
						*(*WeaponArmorUpdateData)(armor.UpdateData) != beforeCarry {
						t.Fatal("initial missing service mutated state")
					}
				} else if !slices.Equal(events, []string{"quest-mode", "armor", "quest-mode", "unsupported"}) ||
					update.Field547 != 1 || update.Field546 != uint32(missile.TypeInd) || update.Field1 != math.Float32bits(0.25) || armor.HealthData.Cur != 23 {
					t.Fatalf("live missing service lost the completed prefix: events=%v marker=%d/%d carry=%g armor=%d", events, update.Field547, update.Field546, math.Float32frombits(update.Field1), armor.HealthData.Cur)
				}
			})
		}
	}
}

func TestPlayerDamageNPCMissileLive4E17B0NilCarryBeforeQuest(t *testing.T) {
	for _, typ := range []object.DamageType{object.DamageImpale, object.DamageImpact} {
		t.Run(fmt.Sprint(typ), func(t *testing.T) {
			target, source, missile, r, _ := playerDamageNPCPierceSourceFixture4E17B0(t, true)
			cached := target.UpdateDataMonster()
			cached.ArmorEquipFlags, cached.AIStackInd = 0x1000000, 0
			cached.AIStack[0].Action = uint32(ai.ACTION_BLOCK_ATTACK)
			r.BlockDirection = func(v *Object, _ types.Pointf) bool {
				if v != target {
					t.Fatal("shield target")
				}
				target.UpdateData = nil
				return false
			}
			r.QuestMode = func() bool { t.Fatal("Quest query before live carry fault"); return false }
			r.QuestDamageScale = func() float32 { t.Fatal("Quest scale before live carry fault"); return 0 }
			r.DefaultDamage = func(*Object, *Object, *Object, int32, object.DamageType) bool {
				t.Fatal("HP before live carry fault")
				return false
			}
			defer func() {
				if recover() == nil {
					t.Fatal("nil live carry was hidden")
				}
				if target.HealthData.Cur != 20 || cached.Field547 != 99 || cached.Field546 != 77 || cached.Field1 != 0 {
					t.Fatal("Quest ordering change altered the existing carry-fault prefix")
				}
			}()
			PlayerDamageNative4E17B0(target, source, missile, 8, typ, r)
		})
	}
}
