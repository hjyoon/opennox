package server

import (
	"math"
	"testing"
	"unsafe"

	"github.com/opennox/libs/object"

	"github.com/opennox/opennox/v1/common/unit/ai"
)

func blockMonsterTestObject532070(t *testing.T) *Object {
	t.Helper()
	unit := monsterActionTestObject50A910(t)
	update := unit.UpdateDataMonster()
	update.AIStackInd = 1
	update.AIStack[1].Action = uint32(ai.ACTION_BLOCK_ATTACK)
	return unit
}

func assertBlockEvents532070(t *testing.T, got, want []string) {
	t.Helper()
	if len(got) != len(want) {
		t.Fatalf("events = %v, want %v", got, want)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("events = %v, want %v", got, want)
		}
	}
}

func TestMonsterActionBlockAttack532070RefreshesCachedHeadInOracleOrder(t *testing.T) {
	unit := blockMonsterTestObject532070(t)
	update := unit.UpdateDataMonster()
	update.AIStack[0].Args[0] = 7
	var events []string
	frames := []uint32{100, 115}
	frameIndex := 0
	handled := monsterActionBlockAttack532070(unit, monsterActionBlockHooks532070{
		testShield: func(got *Object) int {
			events = append(events, "shield")
			if got != unit || uintptr(unsafe.Pointer(got)) != uintptr(unsafe.Pointer(unit)) {
				t.Fatalf("shield unit = %p, want exact native pointer %p", got, unit)
			}
			update.AIStackInd = 0
			return 1
		},
		tickRate: func() uint32 {
			events = append(events, "fps")
			return 31
		},
		frame: func() uint32 {
			events = append(events, "frame")
			if frameIndex >= len(frames) {
				t.Fatal("too many frame reads")
			}
			frame := frames[frameIndex]
			frameIndex++
			return frame
		},
		pop: func() int {
			t.Fatal("deadline equality popped action")
			return 0
		},
		push: func(ai.ActionType, ...any) *AIStackItem {
			t.Fatal("deadline equality pushed finish action")
			return nil
		},
	})
	if !handled || frameIndex != 2 {
		t.Fatalf("handled/frame reads = %v/%d, want true/2", handled, frameIndex)
	}
	if got := update.AIStack[1].ArgU32(0); got != 115 {
		t.Fatalf("cached head deadline = %d, want 115", got)
	}
	if got := update.AIStack[0].ArgU32(0); got != 7 {
		t.Fatalf("live head was overwritten: %d", got)
	}
	assertBlockEvents532070(t, events, []string{"shield", "fps", "frame", "frame"})
}

func TestMonsterActionBlockAttack532070ExpiryReadsSubclassAfterPop(t *testing.T) {
	for _, tc := range []struct {
		name           string
		setNPCAfterPop bool
		wantPush       bool
	}{
		{name: "monster pushes finish", wantPush: true},
		{name: "NPC after pop suppresses finish", setNPCAfterPop: true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			unit := blockMonsterTestObject532070(t)
			unit.UpdateDataMonster().AIStack[1].Args[0] = 41
			var events []string
			handled := monsterActionBlockAttack532070(unit, monsterActionBlockHooks532070{
				testShield: func(got *Object) int {
					if got != unit {
						t.Fatalf("shield unit = %p, want %p", got, unit)
					}
					events = append(events, "shield")
					return 0
				},
				tickRate: func() uint32 {
					t.Fatal("failed shield test read FPS")
					return 0
				},
				frame: func() uint32 {
					events = append(events, "frame")
					return 42
				},
				pop: func() int {
					events = append(events, "pop")
					if tc.setNPCAfterPop {
						unit.ObjSubClass |= object.SubClass(object.MonsterNPC)
					}
					return 0
				},
				push: func(action ai.ActionType, args ...any) *AIStackItem {
					if action != ai.ACTION_BLOCK_FINISH || len(args) != 0 {
						t.Fatalf("push = %v/%v, want block finish without args", action, args)
					}
					events = append(events, "push")
					return &AIStackItem{Action: uint32(action)}
				},
			})
			if !handled {
				t.Fatal("block action was not handled")
			}
			want := []string{"shield", "frame", "pop"}
			if tc.wantPush {
				want = append(want, "push")
			}
			assertBlockEvents532070(t, events, want)
		})
	}
}

func TestMonsterActionBlockAttack532070UsesUnsignedFrameComparison(t *testing.T) {
	unit := blockMonsterTestObject532070(t)
	unit.UpdateDataMonster().AIStack[1].Args[0] = uintptr(uint32(math.MaxUint32))
	handled := monsterActionBlockAttack532070(unit, monsterActionBlockHooks532070{
		frame:      func() uint32 { return 0 },
		tickRate:   func() uint32 { t.Fatal("shield miss read FPS"); return 0 },
		testShield: func(*Object) int { return 0 },
		pop:        func() int { t.Fatal("wrapped frame popped action"); return 0 },
		push:       func(ai.ActionType, ...any) *AIStackItem { t.Fatal("wrapped frame pushed action"); return nil },
	})
	if !handled {
		t.Fatal("block action was not handled")
	}
}

func TestMonsterActionBlockTerminal5320E0(t *testing.T) {
	for _, done := range []uint8{0, 1} {
		t.Run(string(rune('0'+done)), func(t *testing.T) {
			unit := blockMonsterTestObject532070(t)
			unit.UpdateDataMonster().Field120_3 = done
			pops := 0
			if !monsterActionBlockTerminal5320E0(unit, func() int { pops++; return 0 }) {
				t.Fatal("terminal action was not handled")
			}
			if pops != int(done) {
				t.Fatalf("pops = %d, want %d", pops, done)
			}
		})
	}
}

func TestMonsterActionBlock532070RejectsInvalidMetadataBeforeHooks(t *testing.T) {
	badHook := func() uint32 { t.Fatal("invalid metadata called hook"); return 0 }
	for _, unit := range []*Object{
		nil,
		{},
		{ObjClass: object.ClassPlayer, UpdateData: unsafe.Pointer(new(MonsterUpdateData))},
		{ObjClass: object.ClassMonster, UpdateData: unsafe.Pointer(&MonsterUpdateData{AIStackInd: -1})},
		{ObjClass: object.ClassMonster, UpdateData: unsafe.Pointer(&MonsterUpdateData{AIStackInd: int8(len(MonsterUpdateData{}.AIStack))})},
	} {
		if monsterActionBlockAttack532070(unit, monsterActionBlockHooks532070{
			frame:      badHook,
			tickRate:   badHook,
			testShield: func(*Object) int { t.Fatal("invalid metadata tested shield"); return 0 },
			pop:        func() int { t.Fatal("invalid metadata popped"); return 0 },
			push:       func(ai.ActionType, ...any) *AIStackItem { t.Fatal("invalid metadata pushed"); return nil },
		}) {
			t.Fatalf("invalid unit %#v was handled", unit)
		}
	}
}
