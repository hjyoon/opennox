package opennox

import (
	"fmt"
	"math"
	"testing"
	"unsafe"

	"github.com/opennox/libs/object"
	"github.com/opennox/libs/strman"

	noxflags "github.com/opennox/opennox/v1/common/flags"
	"github.com/opennox/opennox/v1/common/unit/ai"
	"github.com/opennox/opennox/v1/legacy/common/alloc"
	"github.com/opennox/opennox/v1/server"
)

func aiAliveDependencyNative546A70(t *testing.T) (*Server, *server.Object, *server.Object, *server.HealthData) {
	t.Helper()
	s := NewServer(nil, nil, strman.New())
	t.Cleanup(s.Close)
	if !s.Objs.Init(2) {
		t.Fatal("cannot initialize native AI object allocation")
	}
	t.Cleanup(s.Objs.FreeObjects)
	unit, target := s.Objs.NewObject(&server.ObjectType{}), s.Objs.NewObject(&server.ObjectType{})
	update, freeUpdate := alloc.New(server.MonsterUpdateData{})
	t.Cleanup(freeUpdate)
	health, freeHealth := alloc.New(server.HealthData{})
	t.Cleanup(freeHealth)
	unit.ObjClass, unit.UpdateData = object.ClassMonster, unsafe.Pointer(update)
	target.ObjClass, target.HealthData = object.ClassMonster, health
	for _, pointer := range []unsafe.Pointer{unsafe.Pointer(unit), unit.UpdateData, unsafe.Pointer(target), unsafe.Pointer(health)} {
		if unsafe.Sizeof(uintptr(0)) == 8 && uintptr(pointer) <= math.MaxUint32 {
			t.Fatalf("native AI object/update/health below 4 GiB: %p", pointer)
		}
	}
	oldEngine := noxflags.GetEngine() & noxflags.EngineShowAI
	noxflags.UnsetEngine(noxflags.EngineShowAI)
	t.Cleanup(func() { noxflags.SetEngine(oldEngine) })
	s.SetFrame(702)
	s.SetTickRate(30)
	return s, unit, target, health
}

var aiAliveDependencyShapes546A70 = []string{
	"and", "or-false", "or-true-before-alive", "or-true-after-alive",
}

func aiAliveDependencyStack546A70(unit, target *server.Object, shape string) *server.MonsterUpdateData {
	update := unit.UpdateDataMonster()
	*update = server.MonsterUpdateData{
		Field2: 71, Field67: 72, Field74: 73, Field91: 74,
		Field97: 75, Field98: 76, Field101: 77, Field102: 78,
		Field120_0: 81, Field120_1: 82, Field120_2: 83, Field120_3: 84,
		Field124: 11, Field137: 12, Field136_1: 85, Field136_2: 86,
	}
	high := uintptr(0x12345678)
	high <<= 32
	update.AIStack[0] = server.AIStackItem{Action: uint32(ai.ACTION_WAIT), Args: [4]uintptr{7, high | 8, 9, 10}, Field5: 0}
	condition := server.AIStackItem{
		Action: uint32(ai.DEPENDENCY_ALIVE), Field5: 0xfedcba98,
		Args: [4]uintptr{uintptr(unsafe.Pointer(target)), high | 0x80000000, high | 0x7fc12345, high | 0x10203040},
	}
	index := 1
	switch shape {
	case "and":
		update.AIStack[index] = condition
	case "or-false", "or-true-before-alive":
		// Literal TIME 0 is false even at frame 0; NOT_MOVED is true for
		// this stationary unit, including uint32 frame-wrap fixtures.
		companion := ai.DEPENDENCY_TIME
		if shape == "or-true-before-alive" {
			companion = ai.DEPENDENCY_NOT_MOVED
		}
		update.AIStack[index] = server.AIStackItem{Action: uint32(companion)}
		index++
		update.AIStack[index] = condition
	case "or-true-after-alive":
		update.AIStack[index] = condition
		index++
		update.AIStack[index] = server.AIStackItem{Action: uint32(ai.DEPENDENCY_NOT_MOVED)}
	default:
		panic("unknown original ALIVE stack shape")
	}
	if shape != "and" {
		index++
		update.AIStack[index] = server.AIStackItem{Action: uint32(ai.DEPENDENCY_OR)}
	}
	index++
	update.AIStack[index] = server.AIStackItem{Action: uint32(ai.ACTION_WAIT), Args: [4]uintptr{900, high | 88, 99, 100}}
	update.AIStackInd = int8(index)
	update.AIStack[20] = server.AIStackItem{Action: 17, Args: [4]uintptr{high | 1, high | 2, high | 3, high | 4}, Field5: 18}
	return update
}

func aiAliveDependencyCheck546A70(t *testing.T, s *Server, unit, target *server.Object, shape string, validTarget, wantFault bool, deadline uint32) {
	t.Helper()
	update := aiAliveDependencyStack546A70(unit, target, shape)
	before := *update
	s.AI.StackChanged = false
	var fault any
	func() {
		defer func() { fault = recover() }()
		(&aiData{s: s}).nox_xxx_mobActionDependency(unit)
	}()
	if (fault != nil) != wantFault {
		t.Fatalf("original ALIVE gate fault=%v wantFault=%t target=%p stack=%s", fault, wantFault, target, shape)
	}
	want := before
	wantPopped := false
	if !wantFault && !validTarget {
		// 00546BE9 clears the cached update's heard sound before recording
		// frame + FPS, even if another member makes the OR group pass.
		want.Field97, want.Field101 = 0, deadline
		wantPopped = shape == "and" || shape == "or-false"
		if wantPopped {
			want.AIStackInd = 0
			want.Field2, want.Field67, want.Field74, want.Field91 = 0, 0, 0, 0
			want.Field120_1, want.Field120_2, want.Field120_3 = 0, 0, 0
			want.Field124, want.Field137 = s.Frame(), s.Frame()
		}
	}
	if *update != want || s.AI.StackChanged != wantPopped {
		t.Fatalf("original ALIVE condition/pop prefix differs: index=%d/%d heard=%d/%d deadline=%d/%d reset=%d,%d/%d,%d changed=%t/%t", update.AIStackInd, want.AIStackInd, update.Field97, want.Field97, update.Field101, want.Field101, update.Field124, update.Field137, want.Field124, want.Field137, s.AI.StackChanged, wantPopped)
	}
}

func TestAIAliveDependency546A70NativeTargetAndGroup(t *testing.T) {
	s, unit, target, health := aiAliveDependencyNative546A70(t)
	for _, tc := range []struct {
		name      string
		hasTarget bool
		class     object.Class
		flags     object.Flags
		hasHealth bool
		cur, max  uint16
		valid     bool
	}{
		{"nil-target", false, 0, 0, false, 0, 0, false},
		{"food-nil-health", true, object.ClassFood, 0, false, 0, 0, false},
		{"food-positive-health", true, object.ClassFood, 0, true, 1, 5, false},
		{"nonunit-high-byte-nil-health", true, object.Class(0x60000600), 0, false, 0, 0, false},
		{"monster-positive", true, object.ClassMonster, 0, true, 1, 5, true},
		{"monster-zero-current", true, object.ClassMonster, 0, true, 0, 5, false},
		{"monster-zero-current-and-max", true, object.ClassMonster, 0, true, 0, 0, true},
		{"monster-maximum-current-zero-max", true, object.ClassMonster, 0, true, 0xffff, 0, true},
		{"player-unsigned-current", true, object.ClassPlayer, 0, true, 0x8000, 0xffff, true},
		{"player-zero-current", true, object.ClassPlayer, 0, true, 0, 0xffff, false},
		{"mixed-unit-zero-max", true, object.MaskUnits | object.ClassFood, 0, true, 0, 0, true},
		{"monster-positive-ignored-flags", true, object.ClassMonster, object.FlagDead | object.FlagNoUpdate | object.FlagDestroyed, true, 1, 1, true},
	} {
		for _, shape := range aiAliveDependencyShapes546A70 {
			t.Run(tc.name+"/"+shape, func(t *testing.T) {
				target.ObjClass, target.ObjFlags, target.HealthData = tc.class, tc.flags, nil
				*health = server.HealthData{Cur: tc.cur, Max: tc.max, Field2: 0xabcd, Field16: 0x12345678}
				if tc.hasHealth {
					target.HealthData = health
				}
				selected := target
				if !tc.hasTarget {
					selected = nil
				}
				before := *health
				healthPointer := target.HealthData
				aiAliveDependencyCheck546A70(t, s, unit, selected, shape, tc.valid, false, 732)
				if *health != before || target.HealthData != healthPointer {
					t.Fatal("ALIVE changed target health/native identity")
				}
			})
		}
	}
}

func TestAIAliveDependency546A70NativeNilHealthFaultPrefix(t *testing.T) {
	s, unit, target, _ := aiAliveDependencyNative546A70(t)
	for _, class := range []object.Class{object.ClassMonster, object.ClassPlayer} {
		for _, shape := range aiAliveDependencyShapes546A70 {
			t.Run(fmt.Sprintf("class-%d/%s", class, shape), func(t *testing.T) {
				target.ObjClass, target.HealthData = class, nil
				// A unit's missing health faults before heard/deadline/reset;
				// an already-true OR companion must not suppress that read.
				aiAliveDependencyCheck546A70(t, s, unit, target, shape, false, true, 732)
			})
		}
	}
}

func TestAIAliveDependency546A70NativeDeadlineWrap(t *testing.T) {
	s, unit, target, health := aiAliveDependencyNative546A70(t)
	target.ObjClass, target.HealthData = object.ClassMonster, health
	*health = server.HealthData{Cur: 0, Max: 1}
	for _, tc := range []struct {
		frame, fps, want uint32
	}{
		{702, 30, 732}, {0xfffffffe, 30, 28}, {0xffffffff, 1, 0},
		{2, 0xffffffff, 1}, {51, 0, 51}, {0, 0, 0},
	} {
		for _, missing := range []bool{false, true} {
			for _, shape := range aiAliveDependencyShapes546A70 {
				t.Run(fmt.Sprintf("frame-%08x/fps-%08x/missing-%t/%s", tc.frame, tc.fps, missing, shape), func(t *testing.T) {
					s.SetFrame(tc.frame)
					s.SetTickRate(tc.fps)
					selected := target
					if missing {
						selected = nil
					}
					aiAliveDependencyCheck546A70(t, s, unit, selected, shape, false, false, tc.want)
				})
			}
		}
	}
}

func TestAIAliveDependency546A70NativeOriginalClassByte(t *testing.T) {
	s, unit, target, health := aiAliveDependencyNative546A70(t)
	*health = server.HealthData{Cur: 1, Max: 1}
	for bits := uint32(0); bits < 256; bits++ {
		target.ObjClass, target.HealthData = object.Class(0x60010000|bits), health
		// Literal original low BYTE test at 00546BC8, not the Go class
		// predicate, defines which native target may read its health.
		aiAliveDependencyCheck546A70(t, s, unit, target, "and", bits&0x06 != 0, false, 732)
	}
}
