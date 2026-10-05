package server

import (
	"fmt"
	"math"
	"reflect"
	"testing"
	"unsafe"

	"github.com/opennox/libs/object"
	"github.com/opennox/libs/player"
	"github.com/opennox/libs/types"
	noxflags "github.com/opennox/opennox/v1/common/flags"
	"github.com/opennox/opennox/v1/common/unit/ai"
)

func TestMonsterMainProgress547210AllMovingActions(t *testing.T) {
	for _, action := range []ai.ActionType{ai.ACTION_MOVE_TO, ai.ACTION_FAR_MOVE_TO, ai.ACTION_MOVE_TO_HOME, ai.ACTION_ROAM, ai.ACTION_FLEE} {
		for _, mode := range []string{"passive", "active", "caster", "equipped-NPC", "disabled-stunned"} {
			t.Run(fmt.Sprintf("%s/%s", action, mode), func(t *testing.T) {
				s, unit, update, runtime := monsterMainDodgeServiceFixture547C50(t)
				update.AIStack[0].Action = uint32(action)
				update.Field124, update.Field125, update.Field126 = 10, math.Float32bits(100), math.Float32bits(100)
				runtime.TestShield = func(*Object) int { return 0 }
				switch mode {
				case "passive":
					update.Aggression = 0
				case "caster":
					update.StatusFlags = object.MonStatusCanCastSpells
				case "equipped-NPC":
					unit.ObjSubClass = object.SubClass(object.MonsterNPC)
					update.WeaponEquipFlags = 0x400
					update.ArmorEquipFlags = 0x1000000
				case "disabled-stunned":
					unit.ObjFlags = object.FlagDestroyed
					unit.Buffs = 1<<ENCHANT_HELD | 1<<ENCHANT_FREEZE
				}
				stack := update.AIStack
				if !s.MonsterMainNativeRuntime547210(unit, runtime) || update.Field124 != 100 || update.Field125 != math.Float32bits(100) || update.Field126 != math.Float32bits(200) || update.AIStack != stack {
					t.Fatalf("original common movement tracking missing: action=%v frame=%d position=%08x/%08x", action, update.Field124, update.Field125, update.Field126)
				}
			})
		}
	}
}

func TestMonsterMainProgress547210DistanceAndUnsignedTimeout(t *testing.T) {
	for _, tc := range []struct {
		name                 string
		old                  types.Pointf
		frame, previous, fps uint32
		moved, stalled       bool
	}{
		{"exact-225", types.Ptf(115, 200), 116, 100, 31, false, true},
		{"next-below-225", types.Ptf(math.Nextafter32(115, 100), 200), 116, 100, 31, false, true},
		{"next-above-225", types.Ptf(math.Nextafter32(115, 120), 200), 116, 100, 31, true, false},
		{"half-fps-equality", types.Ptf(100, 200), 115, 100, 31, false, false},
		{"frame-wrap-equality", types.Ptf(100, 200), 5, ^uint32(0) - 9, 30, false, false},
		{"frame-wrap-expired", types.Ptf(100, 200), 6, ^uint32(0) - 9, 30, false, true},
		{"unsigned-half-fps", types.Ptf(100, 200), 0x80000001, 0, 0xffffffff, false, true},
		{"NaN-uses-timeout", types.Ptf(float32(math.NaN()), 200), 116, 100, 31, false, true},
		{"infinity-moved", types.Ptf(float32(math.Inf(1)), 200), 116, 100, 31, true, false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			s, unit, update, runtime := monsterMainDodgeServiceFixture547C50(t)
			s.SetFrame(tc.frame)
			s.SetTickRate(tc.fps)
			update.Field363 = tc.frame + 1
			update.AIStack[0].Action = uint32(ai.ACTION_MOVE_TO)
			update.Field124, update.Field125, update.Field126 = tc.previous, math.Float32bits(tc.old.X), math.Float32bits(tc.old.Y)
			calls := 0
			runtime.RandomInt = func(min, max int) int {
				calls++
				if calls == 1 {
					if min != 0 || max != 100 {
						t.Fatal("admission bounds")
					}
					return 33
				}
				if min != int(tc.fps>>1) || max != int(int32(2*tc.fps)) {
					t.Fatal("duration bounds")
				}
				return 20
			}
			if !s.MonsterMainNativeRuntime547210(unit, runtime) {
				t.Fatal("common progress tail rejected")
			}
			if tc.stalled {
				if calls != 2 || !update.StatusFlags.Has(object.MonStatusFrustrated) || update.AIStackHead().Type() != ai.ACTION_WAIT || update.AIStackHead().ArgU32(0) != tc.frame+20 {
					t.Fatal("unsigned stalled transition missing")
				}
			} else if calls != 0 || update.AIStackInd != 0 || update.StatusFlags.Has(object.MonStatusFrustrated) {
				t.Fatal("non-stalled branch scheduled an action")
			}
			if tc.moved || tc.stalled {
				if update.Field124 != tc.frame || update.Field125 != math.Float32bits(unit.PosVec.X) || update.Field126 != math.Float32bits(unit.PosVec.Y) {
					t.Fatal("progress writeback missing")
				}
			} else if update.Field124 != tc.previous || update.Field125 != math.Float32bits(tc.old.X) || update.Field126 != math.Float32bits(tc.old.Y) {
				t.Fatal("pre-timeout progress changed")
			}
		})
	}
}

func TestMonsterMainProgress547210EntryCachedHeadAcrossCallback(t *testing.T) {
	for _, mode := range []string{"cached-move/live-wait", "cached-wait/live-move", "cached-move-becomes-wait"} {
		t.Run(mode, func(t *testing.T) {
			s, unit, cached, runtime := monsterMainDodgeServiceFixture547C50(t)
			noxflags.SetGame(noxflags.GameModeCoop)
			cached.AIStack[0].Action = uint32(ai.ACTION_MOVE_TO)
			live := &MonsterUpdateData{Field127: 160}
			live.AIStack[0].Action = uint32(ai.ACTION_WAIT)
			if mode == "cached-wait/live-move" {
				cached.AIStack[0].Action = uint32(ai.ACTION_WAIT)
				live.AIStack[0].Action = uint32(ai.ACTION_MOVE_TO)
			}
			runtime.GUICursorActive = func() bool {
				unit.UpdateData = unsafe.Pointer(live)
				unit.PosVec = types.Ptf(300, 400)
				s.SetFrame(160)
				if mode == "cached-move-becomes-wait" {
					cached.AIStack[0].Action = uint32(ai.ACTION_WAIT)
				}
				return true
			}
			if !s.MonsterMainNativeRuntime547210(unit, runtime) {
				t.Fatal("cached-head tail rejected")
			}
			if mode == "cached-move/live-wait" {
				if cached.Field124 != 160 || cached.Field125 != math.Float32bits(300) || cached.Field126 != math.Float32bits(400) {
					t.Fatal("cached head and record lost")
				}
			} else if cached.Field124 != 0 || cached.Field125 != 0 || cached.Field126 != 0 {
				t.Fatal("used live head or cached action value instead of entry head pointer")
			}
			if live.Field124 != 0 || live.Field125 != 0 || live.Field126 != 0 {
				t.Fatal("movement bookkeeping wrote replacement record")
			}
		})
	}
}

func TestMonsterMainProgress547210FoodAndWeaponTailOrder(t *testing.T) {
	for _, stalled := range []bool{false, true} {
		t.Run(fmt.Sprintf("stalled-%t", stalled), func(t *testing.T) {
			s, unit, update, runtime := monsterMainDodgeServiceFixture547C50(t)
			s.SetFrame(160)
			unit.HealthData.Cur = 50
			update.Field127 = 160 // no earlier health retreat this tick
			update.AIStack[0].Action = uint32(ai.ACTION_FAR_MOVE_TO)
			if stalled {
				update.Field125, update.Field126 = math.Float32bits(unit.PosVec.X), math.Float32bits(unit.PosVec.Y)
			}
			food := &Object{ObjSubClass: object.SubClass(object.FoodApple)}
			owner := &Player{}
			owner.Info().SetPlayerClass(player.Warrior)
			pud := &PlayerUpdateData{Player: owner, Field73: update}
			weapon := &Object{ObjClass: object.ClassWeapon}
			events := []string{}
			runtime.RandomInt = func(min, max int) int {
				if min == 0 && max == 100 {
					return 33
				}
				return 20
			}
			runtime.SearchEdible = func(*Object, float32) *Object {
				if update.Field124 != 160 {
					t.Fatal("food must follow movement bookkeeping")
				}
				events = append(events, "food")
				return food
			}
			runtime.SearchWeapon = func(*Object, float32) *Object { events = append(events, "weapon"); return weapon }
			runtime.PlaceInventory = func(owner, item *Object, a, b int) bool {
				if item == food {
					events = append(events, "place-food")
					update.StatusFlags |= object.MonStatusBot
					update.Field545 = pud
				} else if item == weapon {
					if unit.UpdateData != unsafe.Pointer(pud) {
						t.Fatal("bot not morphed")
					}
					events = append(events, "place-weapon")
				} else {
					t.Fatal("unknown inventory item")
				}
				return false
			}
			if !s.MonsterMainNativeRuntime547210(unit, runtime) {
				t.Fatal("ordered tail rejected")
			}
			want := []string{}
			if !stalled {
				want = []string{"food", "place-food", "weapon", "place-weapon"}
			}
			if !reflect.DeepEqual(events, want) || unit.UpdateData != unsafe.Pointer(update) {
				t.Fatalf("tail events=%v, want %v", events, want)
			}
		})
	}
}
