package server

import (
	"fmt"
	"math"
	"runtime"
	"testing"
	"unsafe"

	"github.com/opennox/libs/object"
	"github.com/opennox/libs/types"

	"github.com/opennox/opennox/v1/common/unit/ai"
)

func monsterActionRefreshCallbackFixture50A910(t *testing.T, action ai.ActionType) (*Object, *MonsterUpdateData, *Object, *Object) {
	t.Helper()
	unit := monsterActionTestObject50A910(t)
	oldTarget := monsterActionTestObject50A910(t)
	newTarget := monsterActionTestObject50A910(t)
	oldTarget.PosVec = types.Ptf(12.5, -7.25)
	newTarget.PosVec = types.Ptf(-19.75, 31.5)
	update := unit.UpdateDataMonster()
	update.AIStackInd = 1
	update.AIStack[0].Action = uint32(ai.ACTION_IDLE)
	update.AIStack[1] = AIStackItem{
		Action: uint32(action),
		Args:   [4]uintptr{0x7fa01234, 0x80000000, 0, 0x12345678},
		Field5: 0x76543210,
	}
	monsterActionSetTarget50A910(&update.AIStack[1], oldTarget)
	return unit, update, oldTarget, newTarget
}

func monsterActionRefreshCallbackPosition50A910(t *testing.T, item *AIStackItem, want types.Pointf) {
	t.Helper()
	wantX, wantY := uintptr(math.Float32bits(want.X)), uintptr(math.Float32bits(want.Y))
	if item.Args[0] != wantX || item.Args[1] != wantY {
		t.Errorf("raw position = %#x/%#x, want %#x/%#x", item.Args[0], item.Args[1], wantX, wantY)
	}
	if item.Args[3] != 0x12345678 || item.Field5 != 0x76543210 {
		t.Errorf("unrelated slot words changed: %#x/%#x", item.Args[3], item.Field5)
	}
}

// GAME.EXE 0050A9E9 and 0050AA2D reload the object argument from the
// entry-cached slot AFTER 005370E0. The pointer passed to visibility is not
// the pointer used for coordinates. Missile's false branch never probes ESCORT.
func TestMonsterActionRefresh50A910CallbackTargetReload(t *testing.T) {
	for _, action := range []ai.ActionType{ai.ACTION_MOVE_TO, ai.ACTION_FAR_MOVE_TO, ai.ACTION_MISSILE_ATTACK} {
		for _, visible := range []bool{false, true} {
			for _, escort := range []bool{false, true} {
				t.Run(fmt.Sprintf("%s/visible=%t/escort=%t", action, visible, escort), func(t *testing.T) {
					unit, update, oldTarget, newTarget := monsterActionRefreshCallbackFixture50A910(t, action)
					defer runtime.KeepAlive([]*Object{unit, oldTarget, newTarget})
					if escort {
						update.AIStack[0].Action = uint32(ai.ACTION_ESCORT)
					}
					item := &update.AIStack[1]
					before := item.ArgPos(0)
					calls := 0
					got := monsterActionRefresh50A910(unit, func(gotUnit, gotTarget *Object, flags int) bool {
						calls++
						if gotUnit != unit || gotTarget != oldTarget || flags != 0 {
							t.Fatalf("visibility arguments = %p/%p/%d, want entry unit/target/0", gotUnit, gotTarget, flags)
						}
						monsterActionSetTarget50A910(item, newTarget)
						return visible
					})
					if got != 0 || calls != 1 {
						t.Fatalf("result/calls = %d/%d, want 0/1", got, calls)
					}
					refresh := visible || action != ai.ACTION_MISSILE_ATTACK && escort
					wantPosition := before
					wantTarget := newTarget
					if refresh {
						wantPosition = newTarget.PosVec
					} else if action != ai.ACTION_MISSILE_ATTACK {
						wantTarget = nil
					}
					monsterActionRefreshCallbackPosition50A910(t, item, wantPosition)
					if item.ArgObj(2) != wantTarget {
						t.Errorf("target = %p, want %p", item.ArgObj(2), wantTarget)
					}
				})
			}
		}
	}
}

// 0050AA1C calls 0050A0D0 with the UNIT, so ESCORT membership comes from
// its live update record. Iteration and coordinate writes retain the old slot.
func TestMonsterActionRefresh50A910CallbackLiveEscort(t *testing.T) {
	for _, action := range []ai.ActionType{ai.ACTION_MOVE_TO, ai.ACTION_FAR_MOVE_TO} {
		for _, entryEscort := range []bool{false, true} {
			for _, liveEscort := range []bool{false, true} {
				t.Run(fmt.Sprintf("%s/entry=%t/live=%t", action, entryEscort, liveEscort), func(t *testing.T) {
					unit, update, oldTarget, newTarget := monsterActionRefreshCallbackFixture50A910(t, action)
					defer runtime.KeepAlive([]*Object{unit, oldTarget, newTarget})
					if entryEscort {
						update.AIStack[0].Action = uint32(ai.ACTION_ESCORT)
					}
					live := new(MonsterUpdateData)
					live.AIStackInd = 0
					live.AIStack[0] = AIStackItem{Action: uint32(ai.ACTION_IDLE), Args: [4]uintptr{11, 22, 0, 44}, Field5: 55}
					if liveEscort {
						live.AIStack[0].Action = uint32(ai.ACTION_ESCORT)
					}
					liveSlot := live.AIStack[0]
					item := &update.AIStack[1]
					before := item.ArgPos(0)
					calls := 0
					monsterActionRefresh50A910(unit, func(*Object, *Object, int) bool {
						calls++
						unit.UpdateData = unsafe.Pointer(live)
						monsterActionSetTarget50A910(item, newTarget)
						return false
					})
					wantPosition, wantTarget := before, (*Object)(nil)
					if liveEscort {
						wantPosition, wantTarget = newTarget.PosVec, newTarget
					}
					monsterActionRefreshCallbackPosition50A910(t, item, wantPosition)
					if calls != 1 || item.ArgObj(2) != wantTarget || unit.UpdateDataMonster() != live {
						t.Errorf("calls/target/live = %d/%p/%p, want 1/%p/%p", calls, item.ArgObj(2), unit.UpdateDataMonster(), wantTarget, live)
					}
					if live.AIStackInd != 0 || live.AIStack[0] != liveSlot {
						t.Fatal("refresh wrote into the callback-replaced stack")
					}
					if update.AIStackInd != 1 {
						t.Fatal("refresh changed the entry stack index")
					}
					runtime.KeepAlive(update)
				})
			}
		}
	}
}

// 0050AA4B's loop count and slot base are cached before visibility. Each
// subsequent slot's action/enemy is live within the original update record.
func TestMonsterActionRefresh50A910CallbackCachedIteration(t *testing.T) {
	for _, action := range []ai.ActionType{ai.ACTION_MOVE_TO, ai.ACTION_FAR_MOVE_TO, ai.ACTION_MISSILE_ATTACK} {
		for _, visible := range []bool{false, true} {
			t.Run(fmt.Sprintf("%s/visible=%t", action, visible), func(t *testing.T) {
				unit, update, oldTarget, newTarget := monsterActionRefreshCallbackFixture50A910(t, action)
				defer runtime.KeepAlive([]*Object{unit, oldTarget, newTarget})
				live := new(MonsterUpdateData)
				live.AIStackInd = 0
				live.AIStack[0].Action = uint32(ai.ACTION_ESCORT)
				live.CurrentEnemy = oldTarget
				update.AIStack[0] = AIStackItem{Action: uint32(ai.ACTION_IDLE), Args: [4]uintptr{1, 2, 0, 4}, Field5: 5}
				item := &update.AIStack[1]
				before := item.ArgPos(0)
				calls := 0
				monsterActionRefresh50A910(unit, func(*Object, *Object, int) bool {
					calls++
					unit.UpdateData = unsafe.Pointer(live)
					update.AIStackInd = -1
					update.CurrentEnemy = newTarget
					update.AIStack[0].Action = uint32(ai.ACTION_FIGHT)
					item.Action = uint32(ai.ACTION_IDLE)
					monsterActionSetTarget50A910(item, newTarget)
					return visible
				})
				wantPosition := newTarget.PosVec
				if action == ai.ACTION_MISSILE_ATTACK && !visible {
					wantPosition = before
				}
				monsterActionRefreshCallbackPosition50A910(t, item, wantPosition)
				if calls != 1 || item.ArgObj(2) != newTarget || item.Type() != ai.ACTION_IDLE {
					t.Errorf("selected branch/target/calls not retained: %s/%p/%d", item.Type(), item.ArgObj(2), calls)
				}
				bottom := &update.AIStack[0]
				if bottom.ArgPos(0) != newTarget.PosVec || bottom.Args[3] != 4 || bottom.Field5 != 5 {
					t.Fatal("entry-cached descending iteration did not refresh the live FIGHT slot")
				}
				if update.AIStackInd != -1 || live.AIStackInd != 0 || live.AIStack[0].Args != [4]uintptr{} {
					t.Fatal("refresh changed a stack index or wrote to the replacement record")
				}
				runtime.KeepAlive(update)
			})
		}
	}
}

// The post-visibility coordinate reloads have no new nil gate. Record the
// original fault prefix; missile visibility failure must still avoid the read.
func TestMonsterActionRefresh50A910CallbackRequiredTargetFaultPrefix(t *testing.T) {
	for _, action := range []ai.ActionType{ai.ACTION_MOVE_TO, ai.ACTION_FAR_MOVE_TO, ai.ACTION_MISSILE_ATTACK} {
		for _, visible := range []bool{false, true} {
			t.Run(fmt.Sprintf("%s/visible=%t", action, visible), func(t *testing.T) {
				unit, update, oldTarget, newTarget := monsterActionRefreshCallbackFixture50A910(t, action)
				defer runtime.KeepAlive([]*Object{unit, oldTarget, newTarget})
				update.AIStack[0].Action = uint32(ai.ACTION_ESCORT)
				item := &update.AIStack[1]
				before := item.ArgPos(0)
				calls := 0
				var fault any
				func() {
					defer func() { fault = recover() }()
					monsterActionRefresh50A910(unit, func(*Object, *Object, int) bool {
						calls++
						item.Args[2] = 0
						return visible
					})
				}()
				wantFault := visible || action != ai.ACTION_MISSILE_ATTACK
				if (fault != nil) != wantFault {
					t.Errorf("post-visibility fault = %v, want fault=%t", fault, wantFault)
				}
				if calls != 1 || item.Args[2] != 0 || update.AIStackInd != 1 {
					t.Errorf("fault prefix calls/target/index = %d/%#x/%d, want 1/0/1", calls, item.Args[2], update.AIStackInd)
				}
				monsterActionRefreshCallbackPosition50A910(t, item, before)
			})
		}
	}
}

func TestMonsterActionRefresh50A910CallbackAbsentEntryTarget(t *testing.T) {
	for _, action := range []ai.ActionType{ai.ACTION_MOVE_TO, ai.ACTION_FAR_MOVE_TO, ai.ACTION_MISSILE_ATTACK} {
		for _, destroyed := range []bool{false, true} {
			t.Run(fmt.Sprintf("%s/destroyed=%t", action, destroyed), func(t *testing.T) {
				unit, update, oldTarget, newTarget := monsterActionRefreshCallbackFixture50A910(t, action)
				defer runtime.KeepAlive([]*Object{unit, oldTarget, newTarget})
				item := &update.AIStack[1]
				before := item.ArgPos(0)
				if destroyed {
					oldTarget.ObjFlags = object.FlagDestroyed
				} else {
					item.Args[2] = 0
				}
				monsterActionRefresh50A910(unit, func(*Object, *Object, int) bool {
					t.Fatal("visibility called after absent/destroyed entry target")
					return true
				})
				monsterActionRefreshCallbackPosition50A910(t, item, before)
				if item.Args[2] != 0 || update.AIStackInd != 1 {
					t.Fatal("absent target or stack index changed")
				}
			})
		}
	}
}
