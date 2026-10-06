package server

import (
	"bytes"
	"fmt"
	"math"
	"reflect"
	"sync/atomic"
	"testing"
	"unsafe"

	"github.com/opennox/libs/object"
	"github.com/opennox/libs/prand"
	"github.com/opennox/libs/types"

	"github.com/opennox/opennox/v1/common/unit/ai"
	"github.com/opennox/opennox/v1/legacy/common/alloc"
)

func moveToNativeRetryFixture5443F0(t *testing.T, action ai.ActionType) (*Server, *Object, *MonsterUpdateData) {
	t.Helper()
	s := new(Server)
	s.handle = atomic.AddUintptr(&serverLast, 1)
	servers.Store(s.handle, s)
	t.Cleanup(func() { servers.Delete(s.handle) })
	s.SetFrame(1000)
	s.SetTickRate(30)
	s.Rand.Logic, s.Rand.Other = prand.New(1496), prand.New(3905)
	unit, freeUnit := alloc.New(Object{})
	t.Cleanup(freeUnit)
	update, freeUpdate := alloc.New(MonsterUpdateData{})
	t.Cleanup(freeUpdate)
	*unit = Object{ObjClass: object.ClassMonster, SpeedBase: 2, PosVec: types.Ptf(100, 200), serverHandle: s.handle, UpdateData: unsafe.Pointer(update)}
	*update = MonsterUpdateData{AIStackInd: 0, StatusFlags: object.MonStatusCanSeeFriends}
	update.AIStack[0].Action = uint32(action)
	update.AIStack[0].SetArgs(types.Ptf(300, 400), uint32(0))
	for _, p := range []unsafe.Pointer{unsafe.Pointer(unit), unsafe.Pointer(update)} {
		if unsafe.Sizeof(uintptr(0)) == 8 && uintptr(p) <= math.MaxUint32 {
			t.Fatalf("native MOVE_TO retry fixture below 4 GiB: %p", p)
		}
	}
	return s, unit, update
}

func moveToRawSnapshot5443F0(p unsafe.Pointer, size uintptr) []byte {
	return append([]byte(nil), unsafe.Slice((*byte)(p), int(size))...)
}

var moveToSharedActions5443F0 = []ai.ActionType{ai.ACTION_MOVE_TO, ai.ACTION_MOVE_TO_HOME, ai.ACTION_FAR_MOVE_TO, ai.ACTION_FIGHT}

// GAME.EXE 00534324..0053433F compares SpeedBase >= the binary32 word
// 3C23D70A at 00583A9C. The unordered x87 result is false, including sNaNs.
func TestMonsterActionMoveTo5443F0NativeOrderedSpeed(t *testing.T) {
	for _, tc := range []struct {
		word uint32
		move bool
	}{
		{0, false}, {0x80000000, false}, {1, false}, {0x80000001, false},
		{0x3c23d709, false}, {0x3c23d70a, true}, {0x3c23d70b, true},
		{0x3f800000, true}, {0xbf800000, false}, {0x7f7fffff, true},
		{0x7f800000, true}, {0xff800000, false},
		{0x7fc12345, false}, {0xffc12345, false}, {0x7f800001, false}, {0xff800001, false},
	} {
		for _, action := range moveToSharedActions5443F0 {
			t.Run(fmt.Sprintf("%08x/%s", tc.word, action), func(t *testing.T) {
				s, unit, update := moveToNativeRetryFixture5443F0(t, action)
				unit.SpeedBase = math.Float32frombits(tc.word)
				beforeUnit := moveToRawSnapshot5443F0(unsafe.Pointer(unit), unsafe.Sizeof(*unit))
				beforeUpdate := *update
				var events []string
				hooks := s.monsterActionMoveToHooks5443F0(unit, nil, nil)
				hooks.setMovePath = func(*Object, types.Pointf) bool { events = append(events, "path"); return false }
				hooks.moveAudio = func(*Object) { events = append(events, "audio") }
				hooks.pop = func() int { events = append(events, "pop"); return 0 }
				if !monsterActionMoveToForAction5443F0(unit, action, hooks) {
					t.Fatal("valid native action was rejected")
				}
				want := []string{"pop"}
				if tc.move {
					want = []string{"path", "audio"}
				}
				if !reflect.DeepEqual(events, want) || *update != beforeUpdate ||
					!bytes.Equal(beforeUnit, moveToRawSnapshot5443F0(unsafe.Pointer(unit), unsafe.Sizeof(*unit))) ||
					s.Rand.Logic.Index() != 1496 || s.Rand.Other.Index() != 3905 {
					t.Fatalf("ordered speed: events=%v want=%v raw speed=%08x", events, want, math.Float32bits(unit.SpeedBase))
				}
			})
		}
	}
}

// 00544527 and 00544588 gate each RNG draw on a successful 0050A260 push.
// Exercise actual native stack capacity and real server Logic RNG, including
// partial success; no manually injected pop or compensating RNG is allowed.
func TestMonsterActionMoveTo5443F0NativeStackCapacityRNG(t *testing.T) {
	for available := 0; available <= 3; available++ {
		for _, action := range moveToSharedActions5443F0 {
			t.Run(fmt.Sprintf("slots-%d/%s", available, action), func(t *testing.T) {
				s, unit, update := moveToNativeRetryFixture5443F0(t, action)
				update.Field71 = 0xabcdef02
				update.AIStackInd = int8(len(update.AIStack) - 1 - available)
				for i := 0; i < int(update.AIStackInd); i++ {
					update.AIStack[i] = AIStackItem{Action: uint32(ai.ACTION_GUARD), Args: [4]uintptr{uintptr(i), 2, 3, 4}}
				}
				update.AIStackHead().Action = uint32(action)
				update.AIStackHead().SetArgs(types.Ptf(300, 400), uint32(0))
				beforeUnit, beforeUpdate := *unit, *update
				wantRNG := prand.New(1496)
				var deadlineTime, deadlineWait uint32
				if available > 0 {
					deadlineTime = s.Frame() + uint32(wantRNG.IntClamp(60, 120))
				}
				if available == 3 {
					deadlineWait = s.Frame() + uint32(wantRNG.IntClamp(15, 30))
				}
				hooks := s.monsterActionMoveToHooks5443F0(unit, nil, nil)
				hooks.setMovePath = func(*Object, types.Pointf) bool { return true }
				audio := 0
				hooks.moveAudio = func(got *Object) {
					audio++
					if got != unit {
						t.Fatal("audio lost native unit identity")
					}
				}
				if !monsterActionMoveToForAction5443F0(unit, action, hooks) || audio != 1 {
					t.Fatal("native capacity branch failed")
				}
				if s.Rand.Logic.Index() != wantRNG.Index() || s.Rand.Other.Index() != 3905 || *unit != beforeUnit ||
					update.AIStackInd != beforeUpdate.AIStackInd+int8(available) ||
					update.StatusFlags != beforeUpdate.StatusFlags|object.MonStatusFrustrated || s.AI.StackChanged != (available > 0) {
					t.Fatalf("capacity=%d Logic=%d want=%d stack=%+v", available, s.Rand.Logic.Index(), wantRNG.Index(), update.GetAIStack())
				}
				for i := 0; i <= int(beforeUpdate.AIStackInd); i++ {
					if update.AIStack[i] != beforeUpdate.AIStack[i] {
						t.Fatalf("retry changed underlying action %d", i)
					}
				}
				for i, kind := range []ai.ActionType{ai.DEPENDENCY_TIME, ai.ACTION_RANDOM_WALK, ai.ACTION_WAIT}[:available] {
					item := update.AIStack[int(beforeUpdate.AIStackInd)+1+i]
					var deadline uint32
					if i == 0 {
						deadline = deadlineTime
					} else if i == 2 {
						deadline = deadlineWait
					}
					if item.Type() != kind || item.Args != [4]uintptr{uintptr(deadline)} || item.Field5 != 0 {
						t.Fatalf("new stack item %d=%+v want=%s/%d", i, item, kind, deadline)
					}
				}
			})
		}
	}
}

// 0054452B/0054458C load FPS once AFTER each accepted push; 00544547 and
// 005445A4 load the frame AFTER random. The PE32 bounds and deadlines wrap
// as DWORDs and Random receives signed 32-bit bounds, even on ARM64.
func TestMonsterActionMoveTo5443F0NativeDeadlineCallbackOrder(t *testing.T) {
	for _, fps := range []uint32{0, 1, 30, 0x20000000, 0x40000000, 0x80000001, 0xffffffff} {
		for _, action := range moveToSharedActions5443F0 {
			t.Run(fmt.Sprintf("fps-%08x/%s", fps, action), func(t *testing.T) {
				s, unit, update := moveToNativeRetryFixture5443F0(t, action)
				update.Field71 = 2
				timed, waited := &update.AIStack[1], &update.AIStack[3]
				timed.Args, waited.Args = [4]uintptr{9, uintptr(unsafe.Pointer(unit)), 0xfedcba98, 0x12345678}, [4]uintptr{8, uintptr(unsafe.Pointer(update)), 0x87654321, 0xabcdef01}
				beforeTimed, beforeWaited := *timed, *waited
				var events []string
				randomCalls := 0
				hooks := s.monsterActionMoveToHooks5443F0(unit, nil, nil)
				hooks.setMovePath = func(*Object, types.Pointf) bool { events = append(events, "path"); return true }
				hooks.frame = func() uint32 { events = append(events, "frame"); return s.Frame() }
				hooks.tickRate = func() uint32 { events = append(events, "fps"); return s.TickRate() }
				hooks.random = func(minimum, maximum int) int {
					randomCalls++
					events = append(events, fmt.Sprintf("random-%d", randomCalls))
					wantMin, wantMax := int(int32(2*fps)), int(int32(4*fps))
					if randomCalls == 2 {
						wantMin, wantMax = int(fps>>1), int(int32(fps))
					}
					if minimum != wantMin || maximum != wantMax {
						t.Errorf("random %d bounds=%d/%d want=%d/%d", randomCalls, minimum, maximum, wantMin, wantMax)
					}
					if randomCalls == 1 {
						s.SetFrame(0xfffffffe)
					} else {
						s.SetFrame(5)
					}
					return minimum
				}
				hooks.push = func(kind ai.ActionType, args ...any) *AIStackItem {
					events = append(events, fmt.Sprintf("push-%d", kind))
					if len(args) != 0 {
						t.Errorf("push %s evaluated deadline before acceptance: %v", kind, args)
					}
					s.SetTickRate(fps)
					switch kind {
					case ai.DEPENDENCY_TIME:
						return timed
					case ai.ACTION_WAIT:
						return waited
					default:
						return nil
					}
				}
				hooks.moveAudio = func(*Object) { events = append(events, "audio") }
				monsterActionMoveToForAction5443F0(unit, action, hooks)
				wantEvents := []string{"path", "frame", "push-41", "fps", "random-1", "frame", "push-29", "push-1", "fps", "random-2", "frame", "audio"}
				beforeTimed.Args[0] = uintptr(uint32(0xfffffffe) + 2*fps)
				beforeWaited.Args[0] = uintptr(uint32(5) + (fps >> 1))
				if !reflect.DeepEqual(events, wantEvents) || *timed != beforeTimed || *waited != beforeWaited || randomCalls != 2 {
					t.Fatalf("deadline order=%v want=%v timed=%+v waited=%+v", events, wantEvents, timed, waited)
				}
			})
		}
	}
}

// EDX's entry frame is reused at 005444E8, not queried a second time.
func TestMonsterActionMoveTo5443F0NativeStatusOneEntryFrame(t *testing.T) {
	for _, elapsed := range []uint32{149, 150, 151} {
		for _, action := range moveToSharedActions5443F0 {
			t.Run(fmt.Sprintf("elapsed-%d/%s", elapsed, action), func(t *testing.T) {
				s, unit, update := moveToNativeRetryFixture5443F0(t, action)
				update.Field71, update.Field135 = 0x123401, s.Frame()-elapsed
				var pushed []ai.ActionType
				hooks := s.monsterActionMoveToHooks5443F0(unit, nil, nil)
				hooks.setMovePath = func(*Object, types.Pointf) bool { return true }
				frames := 0
				hooks.frame = func() uint32 { frames++; return uint32(1000 + 7*(frames-1)) }
				hooks.push = func(kind ai.ActionType, args ...any) *AIStackItem {
					pushed = append(pushed, kind)
					return unit.MonsterPushAction(kind, args...)
				}
				hooks.moveAudio = nil
				monsterActionMoveToForAction5443F0(unit, action, hooks)
				want := []ai.ActionType{ai.ACTION_WAIT}
				wantFrames := 2
				if elapsed < 150 {
					want = []ai.ActionType{ai.DEPENDENCY_TIME, ai.ACTION_RANDOM_WALK, ai.ACTION_WAIT}
					wantFrames = 3
				}
				if update.Field135 != 1000 || frames != wantFrames || !reflect.DeepEqual(pushed, want) {
					t.Fatalf("entry frame=%d frame reads=%d actions=%v", update.Field135, frames, pushed)
				}
			})
		}
	}
}

// 00544571 reloads the LOW byte of the entry-cached update after callbacks.
// Keep the real native record even when a callback replaces unit.UpdateData.
func TestMonsterActionMoveTo5443F0NativeLiveTailStatus(t *testing.T) {
	for _, tc := range []struct {
		name        string
		entry, tail uint32
		replace     bool
	}{
		{"retry-clears", 2, 0, false}, {"retry-high-word-only", 2, 0xffffff00, false},
		{"reset-sets", 0, 0xffffff01, false}, {"cached-clear", 2, 0, true}, {"cached-set", 0, 2, true},
	} {
		for _, action := range moveToSharedActions5443F0 {
			t.Run(tc.name+"/"+action.String(), func(t *testing.T) {
				s, unit, update := moveToNativeRetryFixture5443F0(t, action)
				alternate, freeAlternate := alloc.New(MonsterUpdateData{})
				t.Cleanup(freeAlternate)
				*alternate = MonsterUpdateData{Field71: 2}
				if byte(tc.tail) != 0 {
					alternate.Field71 = 0
				}
				if unsafe.Sizeof(uintptr(0)) == 8 && uintptr(unsafe.Pointer(alternate)) <= math.MaxUint32 {
					t.Fatalf("alternate native update below 4 GiB: %p", alternate)
				}
				beforeAlternate := *alternate
				update.Field71 = tc.entry
				var pushed []ai.ActionType
				mutate := func() {
					update.Field71 = tc.tail
					if tc.replace {
						unit.UpdateData = unsafe.Pointer(alternate)
					}
				}
				hooks := s.monsterActionMoveToHooks5443F0(unit, nil, nil)
				hooks.setMovePath = func(*Object, types.Pointf) bool { return true }
				hooks.pathReset = func() bool { mutate(); return true }
				hooks.push = func(kind ai.ActionType, args ...any) *AIStackItem {
					pushed = append(pushed, kind)
					if kind == ai.ACTION_RANDOM_WALK {
						mutate()
					}
					return nil
				}
				hooks.random = func(int, int) int { t.Error("rejected push consumed RNG"); return 0 }
				hooks.moveAudio = nil
				monsterActionMoveToForAction5443F0(unit, action, hooks)
				var want []ai.ActionType
				if tc.entry == 2 {
					want = append(want, ai.DEPENDENCY_TIME, ai.ACTION_RANDOM_WALK)
				}
				if byte(tc.tail) != 0 {
					want = append(want, ai.ACTION_WAIT)
				}
				if !reflect.DeepEqual(pushed, want) || *alternate != beforeAlternate ||
					update.StatusFlags.Has(object.MonStatusFrustrated) != (tc.entry == 2) {
					t.Fatalf("tail status: actions=%v want=%v entry=%08x tail=%08x", pushed, want, tc.entry, update.Field71)
				}
			})
		}
	}
}

// 00544502 passes the cached head's LIVE coordinate slots to DirCalc after
// 00547F10. A nonnil object argument still suppresses the arrival pop.
func TestMonsterActionMoveTo5443F0NativeArrivalLiveCoordinates(t *testing.T) {
	for _, tracked := range []bool{false, true} {
		for _, action := range moveToSharedActions5443F0 {
			t.Run(fmt.Sprintf("tracked-%t/%s", tracked, action), func(t *testing.T) {
				s, unit, update := moveToNativeRetryFixture5443F0(t, action)
				head := update.AIStackHead()
				if tracked {
					head.Args[2] = uintptr(unsafe.Pointer(unit))
				}
				hooks := s.monsterActionMoveToHooks5443F0(unit, nil, nil)
				hooks.setMovePath = func(*Object, types.Pointf) bool { return true }
				hooks.pathReset = func() bool {
					head.SetArgs(types.Ptf(100, 100))
					return false
				}
				pops := 0
				hooks.pop = func() int { pops++; return 0 }
				hooks.moveAudio = nil
				monsterActionMoveToForAction5443F0(unit, action, hooks)
				wantPops, wantDir := 1, DirFromVec(types.Ptf(0, -100))
				if tracked {
					wantPops, wantDir = 0, 0
				}
				if pops != wantPops || unit.Direction2 != wantDir {
					t.Fatalf("arrival pop/direction=%d/%d want=%d/%d", pops, unit.Direction2, wantPops, wantDir)
				}
			})
		}
	}
}
