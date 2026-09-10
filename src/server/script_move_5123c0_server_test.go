package server

import (
	"math"
	"runtime"
	"testing"
	"unsafe"

	"github.com/opennox/libs/object"
	"github.com/opennox/libs/types"

	"github.com/opennox/opennox/v1/common/unit/ai"
)

func TestScriptMoveNative5123C0BuildsMonsterStackWithNativeWaypoint(t *testing.T) {
	s := newMoverStateServer(t)
	update := new(MonsterUpdateData)
	update.AIStackInd = 0
	update.AIStack[0] = AIStackItem{Action: uint32(ai.ACTION_WAIT)}
	update.Field333 = 0xdeadbea5
	unit := &Object{
		ObjClass:     object.ClassMonster,
		UpdateData:   unsafe.Pointer(update),
		serverHandle: s.handle,
	}
	waypoint := &Waypoint{
		Index:     0x10203040,
		PosVec:    types.Pointf{X: 1.5, Y: -2.25},
		PointsCnt: 1,
	}
	if unsafe.Sizeof(uintptr(0)) == 8 {
		if uintptr(unsafe.Pointer(unit)) <= math.MaxUint32 {
			t.Fatalf("unit pointer = %p, want native address above 4 GiB", unit)
		}
		if uintptr(unsafe.Pointer(update)) <= math.MaxUint32 {
			t.Fatalf("UpdateData pointer = %p, want native address above 4 GiB", update)
		}
		if uintptr(unsafe.Pointer(waypoint)) <= math.MaxUint32 {
			t.Fatalf("waypoint pointer = %p, want native address above 4 GiB", waypoint)
		}
	}

	s.ScriptMoveTo5123C0(unit, waypoint, ScriptMoveRuntime5123C0{
		MoverTypeID: func() uint32 {
			t.Fatal("monster branch must not read the mover type")
			return 0
		},
		SetOn: func(*Object) { t.Fatal("monster branch must not set a mover on") },
	})

	if update.AIStackInd != 2 {
		t.Fatalf("AIStackInd = %d, want 2", update.AIStackInd)
	}
	report := &update.AIStack[0]
	if report.Type() != ai.ACTION_REPORT || report.Args != [4]uintptr{8, 0, 0, 0} {
		t.Fatalf("REPORT = %s/%#v, want REPORT/[8 0 0 0]", report.Type(), report.Args)
	}
	roam := &update.AIStack[1]
	if roam.Type() != ai.ACTION_ROAM {
		t.Fatalf("ROAM action = %s", roam.Type())
	}
	if roam.Args[0] != uintptr(unsafe.Pointer(waypoint)) {
		t.Fatalf("ROAM waypoint = %#x, want native pointer %#x", roam.Args[0], uintptr(unsafe.Pointer(waypoint)))
	}
	if roam.Args[1] != 0 || roam.Args[2] != 0xa5 || roam.Args[3] != 0 {
		t.Fatalf("ROAM args = %#v, want [waypoint 0 0xa5 0]", roam.Args)
	}
	move := &update.AIStack[2]
	if move.Type() != ai.ACTION_FAR_MOVE_TO {
		t.Fatalf("move action = %s, want FAR_MOVE_TO", move.Type())
	}
	if move.Args != [4]uintptr{uintptr(math.Float32bits(1.5)), uintptr(math.Float32bits(-2.25)), 0, 0} {
		t.Fatalf("FAR_MOVE_TO args = %#v", move.Args)
	}
	if update.Field333 != 0xdeadbea5 {
		t.Fatalf("Field333 = %#08x, want unchanged 0xdeadbea5", update.Field333)
	}
	if !s.AI.StackChanged {
		t.Fatal("action-stack callbacks did not mark the stack changed")
	}
	runtime.KeepAlive(unit)
	runtime.KeepAlive(waypoint)
}

func TestScriptMoveNative5123C0SkipsRoamForSimpleWaypoint(t *testing.T) {
	s := newMoverStateServer(t)
	update := new(MonsterUpdateData)
	unit := &Object{
		ObjClass:     object.ClassMonster,
		UpdateData:   unsafe.Pointer(update),
		serverHandle: s.handle,
	}
	waypoint := &Waypoint{PosVec: types.Pointf{X: 3, Y: 4}}
	s.ScriptMoveTo5123C0(unit, waypoint, ScriptMoveRuntime5123C0{})
	if update.AIStackInd != 1 || update.AIStack[0].Type() != ai.ACTION_REPORT || update.AIStack[1].Type() != ai.ACTION_FAR_MOVE_TO {
		t.Fatalf("stack = %#v, want REPORT then FAR_MOVE_TO", update.GetAIStack())
	}
}

func TestScriptMoveNative5123C0PartialStoresPreserveUpperBits(t *testing.T) {
	item := new(AIStackItem)
	if unsafe.Sizeof(uintptr(0)) == 8 {
		highArg0 := uint64(0xfedcba9800000000)
		highArg2 := uint64(0x0123456789abcd00)
		item.Args[0] = uintptr(highArg0)
		item.Args[2] = uintptr(highArg2)
	}
	scriptMoveStoreArgU325123C0(item, 0, 0x76543210)
	scriptMoveStoreArgLow5123C0(item, 2, 0xa5)
	if unsafe.Sizeof(uintptr(0)) == 8 {
		wantArg0 := uint64(0xfedcba9876543210)
		wantArg2 := uint64(0x0123456789abcda5)
		if item.Args[0] != uintptr(wantArg0) {
			t.Fatalf("dword store = %#x, want upper native bits preserved", item.Args[0])
		}
		if item.Args[2] != uintptr(wantArg2) {
			t.Fatalf("byte store = %#x, want upper native bits preserved", item.Args[2])
		}
	} else if item.Args[0] != uintptr(0x76543210) || item.Args[2] != uintptr(0xa5) {
		t.Fatalf("32-bit stores = %#x/%#x", item.Args[0], item.Args[2])
	}
}

func TestScriptMoveNative5123C0DirectMoverUsesCachedUpdate(t *testing.T) {
	s := newMoverStateServer(t)
	entry := &MoverUpdateData{Field_0: 9, Field_2: -1}
	replacement := &MoverUpdateData{Field_0: 7, Field_2: -2}
	unit := &Object{
		TypeInd:      0x1234,
		VelVec:       types.Pointf{X: 12.5, Y: -7.25},
		UpdateData:   unsafe.Pointer(entry),
		serverHandle: s.handle,
	}
	waypoint := &Waypoint{Index: 0x10203040}
	var setOnCalls int
	s.ScriptMoveTo5123C0(unit, waypoint, ScriptMoveRuntime5123C0{
		MoverTypeID: func() uint32 { return 0x1234 },
		SetOn: func(got *Object) {
			setOnCalls++
			if got != unit {
				t.Fatalf("SetOn object = %p, want %p", got, unit)
			}
			unit.UpdateData = unsafe.Pointer(replacement)
			waypoint.Index = 0xa1b2c3d4
		},
	})

	if setOnCalls != 1 {
		t.Fatalf("SetOn calls = %d, want 1", setOnCalls)
	}
	if unit.VelVec != (types.Pointf{}) {
		t.Fatalf("velocity = %+v, want zero", unit.VelVec)
	}
	if entry.Field_0 != 0 || uint32(entry.Field_2) != 0xa1b2c3d4 {
		t.Fatalf("entry update = state %d/index %#x", entry.Field_0, uint32(entry.Field_2))
	}
	if replacement.Field_0 != 7 || replacement.Field_2 != -2 {
		t.Fatalf("replacement update was modified: %+v", *replacement)
	}
	if unit.IsUpdatable != 1 || s.Objs.UpdatableList != unit {
		t.Fatalf("updatable state = %d/%p, want 1/%p", unit.IsUpdatable, s.Objs.UpdatableList, unit)
	}
}

func TestScriptMoveNative5123C0FindsMoversBySourceExtent(t *testing.T) {
	s := newMoverStateServer(t)
	const moverType = uint16(0x1234)
	source := &Object{TypeInd: 1, Extent: 0xaabbccdd}
	plain := &Object{TypeInd: 2}
	matchingData := &MoverUpdateData{Field_0: 9, Field_8: source.Extent}
	matching := &Object{TypeInd: moverType, UpdateData: unsafe.Pointer(matchingData)}
	otherData := &MoverUpdateData{Field_0: 8, Field_8: 0x11223344}
	other := &Object{TypeInd: moverType, UpdateData: unsafe.Pointer(otherData)}
	plain.ObjNext = matching
	matching.ObjNext = other
	s.Objs.SetObjects(plain)
	waypoint := &Waypoint{Index: 73}
	var moverTypeLoads int
	var setOn []*Object
	s.ScriptMoveTo5123C0(source, waypoint, ScriptMoveRuntime5123C0{
		MoverTypeID: func() uint32 {
			moverTypeLoads++
			return uint32(moverType)
		},
		SetOn: func(obj *Object) { setOn = append(setOn, obj) },
	})

	if moverTypeLoads != 4 {
		t.Fatalf("mover type loads = %d, want direct check plus three list iterations", moverTypeLoads)
	}
	if len(setOn) != 1 || setOn[0] != matching {
		t.Fatalf("SetOn objects = %p, want only %p", setOn, matching)
	}
	if matchingData.Field_0 != 0 || matchingData.Field_2 != 73 || matching.IsUpdatable != 1 {
		t.Fatalf("matching mover = %+v/updatable %d", *matchingData, matching.IsUpdatable)
	}
	if otherData.Field_0 != 8 || otherData.Field_2 != 0 || other.IsUpdatable != 0 {
		t.Fatalf("nonmatching mover changed = %+v/updatable %d", *otherData, other.IsUpdatable)
	}
}

func TestScriptMoveNative5123C0BlockedObjectDoesNothing(t *testing.T) {
	s := newMoverStateServer(t)
	unit := &Object{ObjFlags: object.Flags(scriptMoveBlockedFlag5123C0)}
	s.ScriptMoveTo5123C0(unit, &Waypoint{}, ScriptMoveRuntime5123C0{
		MoverTypeID: func() uint32 {
			t.Fatal("blocked object must return before mover lookup")
			return 0
		},
		SetOn: func(*Object) { t.Fatal("blocked object must not move") },
	})
}

func TestScriptMoveNativeLayout5123C0(t *testing.T) {
	wantObjectType := uintptr(4)
	wantObjectClass := uintptr(8)
	wantObjectFlags := uintptr(16)
	wantObjectExtent := uintptr(40)
	wantObjectVelocity := uintptr(80)
	wantObjectUpdate := uintptr(748)
	wantActionArgs := uintptr(4)
	wantWaypointPoints := uintptr(476)
	if unsafe.Sizeof(uintptr(0)) == 8 {
		wantObjectType = 8
		wantObjectClass = 12
		wantObjectFlags = 20
		wantObjectExtent = 44
		wantObjectVelocity = 84
		wantObjectUpdate = 872
		wantActionArgs = 8
		wantWaypointPoints = 736
	}
	checks := []struct {
		name string
		got  uintptr
		want uintptr
	}{
		{"Object.TypeInd", unsafe.Offsetof(Object{}.TypeInd), wantObjectType},
		{"Object.ObjClass", unsafe.Offsetof(Object{}.ObjClass), wantObjectClass},
		{"Object.ObjFlags", unsafe.Offsetof(Object{}.ObjFlags), wantObjectFlags},
		{"Object.Extent", unsafe.Offsetof(Object{}.Extent), wantObjectExtent},
		{"Object.VelVec", unsafe.Offsetof(Object{}.VelVec), wantObjectVelocity},
		{"Object.UpdateData", unsafe.Offsetof(Object{}.UpdateData), wantObjectUpdate},
		{"Waypoint.Index", unsafe.Offsetof(Waypoint{}.Index), 0},
		{"Waypoint.PosVec", unsafe.Offsetof(Waypoint{}.PosVec), 8},
		{"Waypoint.PointsCnt", unsafe.Offsetof(Waypoint{}.PointsCnt), wantWaypointPoints},
		{"MonsterUpdateData.Field333", unsafe.Offsetof(MonsterUpdateData{}.Field333), map[bool]uintptr{false: 1332, true: 2072}[unsafe.Sizeof(uintptr(0)) == 8]},
		{"MoverUpdateData.Field_0", unsafe.Offsetof(MoverUpdateData{}.Field_0), 0},
		{"MoverUpdateData.Field_2", unsafe.Offsetof(MoverUpdateData{}.Field_2), 8},
		{"MoverUpdateData.Field_8", unsafe.Offsetof(MoverUpdateData{}.Field_8), 32},
		{"AIStackItem.Args", unsafe.Offsetof(AIStackItem{}.Args), wantActionArgs},
	}
	for _, check := range checks {
		if check.got != check.want {
			t.Errorf("%s on %s/%s = %d, want %d", check.name, runtime.GOOS, runtime.GOARCH, check.got, check.want)
		}
	}
}
