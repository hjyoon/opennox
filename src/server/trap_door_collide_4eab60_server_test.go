package server

import (
	"math"
	"testing"
	"unsafe"

	"github.com/opennox/libs/object"
	"github.com/opennox/libs/types"
)

func TestTrapDoorCollide4EAB60NativeLayout(t *testing.T) {
	wantObjectSize := uintptr(780)
	wantObjClass := uintptr(8)
	wantObjFlags := uintptr(16)
	wantPosVec := uintptr(56)
	wantPos39 := uintptr(156)
	wantField41 := uintptr(164)
	wantField42 := uintptr(168)
	wantShape := uintptr(172)
	wantCollideData := uintptr(700)
	if unsafe.Sizeof(uintptr(0)) == 8 {
		wantObjectSize = 928
		wantObjClass = 12
		wantObjFlags = 20
		wantPosVec = 60
		wantPos39 = 160
		wantField41 = 168
		wantField42 = 172
		wantShape = 176
		wantCollideData = 776
	}
	checks := []struct {
		name string
		got  uintptr
		want uintptr
	}{
		{"TrapDoorCollideData size", unsafe.Sizeof(TrapDoorCollideData{}), 28},
		{"TrapDoorCollideData.Script", unsafe.Offsetof(TrapDoorCollideData{}.Script), 0},
		{"TrapDoorCollideData.FallVelocityX", unsafe.Offsetof(TrapDoorCollideData{}.FallVelocityX), 8},
		{"TrapDoorCollideData.FallVelocityY", unsafe.Offsetof(TrapDoorCollideData{}.FallVelocityY), 12},
		{"TrapDoorCollideData.NextFrame", unsafe.Offsetof(TrapDoorCollideData{}.NextFrame), 16},
		{"TrapDoorCollideData.Delay", unsafe.Offsetof(TrapDoorCollideData{}.Delay), 20},
		{"TrapDoorCollideData.Reserved22", unsafe.Offsetof(TrapDoorCollideData{}.Reserved22), 22},
		{"TrapDoorCollideData.Activated", unsafe.Offsetof(TrapDoorCollideData{}.Activated), 24},
		{"ScriptCallback size", unsafe.Sizeof(ScriptCallback{}), 8},
		{"Shape size", unsafe.Sizeof(Shape{}), 52},
		{"Shape.Kind", unsafe.Offsetof(Shape{}.Kind), 0},
		{"Shape.Circle", unsafe.Offsetof(Shape{}.Circle), 4},
		{"Shape.Box", unsafe.Offsetof(Shape{}.Box), 12},
		{"ShapeBox.W", unsafe.Offsetof(ShapeBox{}.W), 0},
		{"ShapeBox.H", unsafe.Offsetof(ShapeBox{}.H), 4},
		{"Object size", unsafe.Sizeof(Object{}), wantObjectSize},
		{"Object.ObjClass", unsafe.Offsetof(Object{}.ObjClass), wantObjClass},
		{"Object.ObjFlags", unsafe.Offsetof(Object{}.ObjFlags), wantObjFlags},
		{"Object.PosVec", unsafe.Offsetof(Object{}.PosVec), wantPosVec},
		{"Object.Pos39", unsafe.Offsetof(Object{}.Pos39), wantPos39},
		{"Object.Field41", unsafe.Offsetof(Object{}.Field41), wantField41},
		{"Object.Field42", unsafe.Offsetof(Object{}.Field42), wantField42},
		{"Object.Shape", unsafe.Offsetof(Object{}.Shape), wantShape},
		{"Object.CollideData", unsafe.Offsetof(Object{}.CollideData), wantCollideData},
	}
	for _, check := range checks {
		if check.got != check.want {
			t.Errorf("%s = %d, want %d", check.name, check.got, check.want)
		}
	}
}

func TestTrapDoorCollideNative4EAB60EnabledUsesExactMapAndFloatBits(t *testing.T) {
	data := &TrapDoorCollideData{FallVelocityX: 1<<24 + 1, FallVelocityY: -31}
	source := &Object{
		ObjFlags:    object.FlagEnabled,
		PosVec:      types.Ptf(30, -10),
		CollideData: unsafe.Pointer(data),
	}
	source.Shape.Kind = ShapeKindBox
	source.Shape.Box.W = 20
	source.Shape.Box.H = 20
	source.Shape.Box.Calc()
	target := &Object{
		ObjFlags: object.Flags(0x20),
		PosVec:   source.PosVec,
		Pos39:    types.Ptf(101, 102),
		Field41:  0x11223344,
		Field42:  0x55667788,
	}
	target.Shape.Kind = ShapeKindCenter
	collision := types.Ptf(7, -9)

	trapDoorCollideNative4EAB60(source, target, &collision, trapDoorCollideNativeDeps4EAB60{
		mapPointInBox: MapPointInBox57B850,
		abilityActive: func(*Object, int32) int32 {
			t.Fatal("enabled path queried ability")
			return 0
		},
		gameFrame: func() uint32 {
			t.Fatal("enabled path read frame")
			return 0
		},
		scriptCallback: func(*ScriptCallback, *Object, *Object, ScriptEventType) unsafe.Pointer {
			t.Fatal("enabled path called script")
			return nil
		},
	})

	if target.ObjFlags != object.Flags(0x60020) {
		t.Fatalf("target flags = %#x", target.ObjFlags)
	}
	if target.Field41 != math.Float32bits(float32(data.FallVelocityX)) ||
		target.Field42 != math.Float32bits(float32(data.FallVelocityY)) {
		t.Fatalf("velocity bits = %#08x/%#08x", target.Field41, target.Field42)
	}
	if target.Pos39 != source.PosVec {
		t.Fatalf("fall position = %+v, want %+v", target.Pos39, source.PosVec)
	}
	if collision != (types.Ptf(7, -9)) {
		t.Fatalf("collision changed to %+v", collision)
	}
}

func TestTrapDoorCollideNative4EAB60MapReceivesTypedFields(t *testing.T) {
	source := &Object{ObjFlags: object.FlagEnabled}
	source.Shape.Kind = ShapeKindCenter
	target := &Object{ObjFlags: object.Flags(0x40)}
	calls := 0
	trapDoorCollideNative4EAB60(source, target, nil, trapDoorCollideNativeDeps4EAB60{
		mapPointInBox: func(sourcePos *types.Pointf, shape *Shape, targetPos *types.Pointf) bool {
			calls++
			if sourcePos != &source.PosVec || shape != &source.Shape || targetPos != &target.PosVec {
				t.Fatalf("map pointers = %p/%p/%p", sourcePos, shape, targetPos)
			}
			return false
		},
	})
	if calls != 1 || target.ObjFlags != object.Flags(0x40) {
		t.Fatalf("calls/flags = %d/%#x", calls, target.ObjFlags)
	}
}

func TestTrapDoorCollideNative4EAB60InactiveCachesDataAndUsesEvent(t *testing.T) {
	oldData := &TrapDoorCollideData{
		Script: ScriptCallback{Flags: 0xa55a5aa5, Func: -17},
	}
	replacement := &TrapDoorCollideData{Activated: 77}
	source := &Object{CollideData: unsafe.Pointer(oldData)}
	target := &Object{ObjClass: object.ClassPlayer | object.Class(0x80000000)}
	collision := types.Ptf(-3, 4)
	abilityCalls, frameCalls, scriptCalls := 0, 0, 0

	trapDoorCollideNative4EAB60(source, target, &collision, trapDoorCollideNativeDeps4EAB60{
		mapPointInBox: func(*types.Pointf, *Shape, *types.Pointf) bool {
			t.Fatal("inactive path mapped point")
			return false
		},
		abilityActive: func(obj *Object, ability int32) int32 {
			abilityCalls++
			if obj != target || ability != int32(AbilityTreadLightly) {
				t.Fatalf("ability args = %p/%d", obj, ability)
			}
			oldData.Delay = 7
			source.CollideData = unsafe.Pointer(replacement)
			target.ObjClass = 0
			return 0
		},
		gameFrame: func() uint32 {
			frameCalls++
			return 0xfffffffc
		},
		scriptCallback: func(
			block *ScriptCallback,
			caller, trigger *Object,
			event ScriptEventType,
		) unsafe.Pointer {
			scriptCalls++
			if block != &oldData.Script || caller != target || trigger != source || event != NoxEventTrapdoorCollide {
				t.Fatalf("script args = %p/%p/%p/%d", block, caller, trigger, event)
			}
			if oldData.NextFrame != 3 || oldData.Activated != 0 {
				t.Fatalf("data at script = %+v", oldData)
			}
			oldData.Activated = 99
			return unsafe.Pointer(&collision)
		},
	})

	if abilityCalls != 1 || frameCalls != 1 || scriptCalls != 1 {
		t.Fatalf("calls = ability %d frame %d script %d", abilityCalls, frameCalls, scriptCalls)
	}
	if oldData.NextFrame != 3 || oldData.Activated != 1 || replacement.Activated != 77 ||
		source.CollideData != unsafe.Pointer(replacement) {
		t.Fatalf("data = old %+v replacement %+v pointer %p", oldData, replacement, source.CollideData)
	}
	if collision != (types.Ptf(-3, 4)) {
		t.Fatalf("collision changed to %+v", collision)
	}
}

func TestTrapDoorCollide4EAB60ServerBinding(t *testing.T) {
	s := &Server{}
	s.SetFrame(100)
	data := &TrapDoorCollideData{Delay: 9}
	source := &Object{CollideData: unsafe.Pointer(data)}
	target := &Object{ObjClass: object.ClassMonster}
	calls := 0
	s.TrapDoorCollide4EAB60(source, target, nil, TrapDoorCollideRuntime4EAB60{
		ScriptCallback: func(block *ScriptCallback, caller, trigger *Object, event ScriptEventType) unsafe.Pointer {
			calls++
			if block != &data.Script || caller != target || trigger != source || event != NoxEventTrapdoorCollide {
				t.Fatalf("script args = %p/%p/%p/%d", block, caller, trigger, event)
			}
			if data.NextFrame != 109 {
				t.Fatalf("next frame at script = %d", data.NextFrame)
			}
			return nil
		},
	})
	if calls != 1 || data.Activated != 1 || data.NextFrame != 109 {
		t.Fatalf("calls/data = %d/%+v", calls, data)
	}
}

func TestTrapDoorCollideNative4EAB60FaultOrder(t *testing.T) {
	t.Run("nil source before nil target", func(t *testing.T) {
		defer func() {
			if recover() == nil {
				t.Fatal("nil source returned")
			}
		}()
		trapDoorCollideNative4EAB60(nil, nil, nil, trapDoorCollideNativeDeps4EAB60{})
	})

	t.Run("nil data safe for nil target", func(t *testing.T) {
		trapDoorCollideNative4EAB60(&Object{}, nil, nil, trapDoorCollideNativeDeps4EAB60{})
	})

	t.Run("enabled nil data after flag store", func(t *testing.T) {
		source := &Object{ObjFlags: object.FlagEnabled}
		source.Shape.Kind = ShapeKindCenter
		target := &Object{ObjFlags: object.Flags(0x20)}
		defer func() {
			if recover() == nil {
				t.Fatal("nil data returned")
			}
			if target.ObjFlags != object.Flags(0x60020) {
				t.Fatalf("flags at fault = %#x", target.ObjFlags)
			}
		}()
		trapDoorCollideNative4EAB60(source, target, nil, trapDoorCollideNativeDeps4EAB60{
			mapPointInBox: func(*types.Pointf, *Shape, *types.Pointf) bool { return true },
		})
	})
}
