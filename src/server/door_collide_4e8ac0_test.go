package server

import (
	"fmt"
	"math"
	"reflect"
	"strings"
	"testing"
)

type doorCollideTestUpdate4E8AC0 struct {
	lockCode         uint8
	targetDirection  int32
	currentDirection int32
	tileX            int32
	tileY            int32
}

type doorCollideTestObject4E8AC0 struct {
	name       string
	update     *doorCollideTestUpdate4E8AC0
	owner      *doorCollideTestObject4E8AC0
	ownerFrame uint32
	subclass   uint8
	class      uint8
	holder     *doorCollideTestObject4E8AC0
}

type doorCollideTestState4E8AC0 struct {
	log              []string
	frameValue       uint32
	tickValues       []uint64
	feedbackTicks    uint64
	key              *doorCollideTestObject4E8AC0
	quest            bool
	questState       int32
	storedQuestFrame uint32
	rect             doorCollideRect4E8AC0
	target           DoorTilePoint
	priorityTarget   *doorCollideTestObject4E8AC0
	deleted          *doorCollideTestObject4E8AC0
	onAudio          func(uint32, *doorCollideTestObject4E8AC0)
	onQuestSync      func(*doorCollideTestObject4E8AC0)
	onQuestState     func()
}

func (s *doorCollideTestState4E8AC0) add(format string, args ...any) {
	s.log = append(s.log, fmt.Sprintf(format, args...))
}

func doorCollideTestObjectName4E8AC0(obj *doorCollideTestObject4E8AC0) string {
	if obj == nil {
		return "nil"
	}
	return obj.name
}

func (s *doorCollideTestState4E8AC0) hooks() doorCollideHooks4E8AC0[
	*doorCollideTestObject4E8AC0,
	*doorCollideTestUpdate4E8AC0,
] {
	return doorCollideHooks4E8AC0[*doorCollideTestObject4E8AC0, *doorCollideTestUpdate4E8AC0]{
		loadUpdateData: func(obj *doorCollideTestObject4E8AC0) *doorCollideTestUpdate4E8AC0 {
			s.add("update:%s", doorCollideTestObjectName4E8AC0(obj))
			return obj.update
		},
		loadCurrentDirection: func(update *doorCollideTestUpdate4E8AC0) int32 {
			s.add("current:%d", update.currentDirection)
			return update.currentDirection
		},
		loadTargetDirection: func(update *doorCollideTestUpdate4E8AC0) int32 {
			s.add("target:%d", update.targetDirection)
			return update.targetDirection
		},
		loadOwner: func(obj *doorCollideTestObject4E8AC0) *doorCollideTestObject4E8AC0 {
			s.add("owner:%s", doorCollideTestObjectName4E8AC0(obj.owner))
			return obj.owner
		},
		loadOwnerExpiryFrame: func(obj *doorCollideTestObject4E8AC0) uint32 {
			s.add("owner-frame:%d", obj.ownerFrame)
			return obj.ownerFrame
		},
		frame: func() uint32 {
			s.add("frame:%d", s.frameValue)
			return s.frameValue
		},
		storeOwner: func(obj, owner *doorCollideTestObject4E8AC0) {
			s.add("store-owner:%s", doorCollideTestObjectName4E8AC0(owner))
			obj.owner = owner
		},
		ticks: func() uint64 {
			if len(s.tickValues) == 0 {
				panic("unexpected ticks call")
			}
			value := s.tickValues[0]
			s.tickValues = s.tickValues[1:]
			s.add("ticks:%d", value)
			return value
		},
		loadFeedbackTicks: func() uint64 {
			s.add("feedback:%d", s.feedbackTicks)
			return s.feedbackTicks
		},
		storeFeedbackTicks: func(value uint64) {
			s.add("store-feedback:%d", value)
			s.feedbackTicks = value
		},
		loadSubclassByte: func(obj *doorCollideTestObject4E8AC0) uint8 {
			s.add("subclass:%d", obj.subclass)
			return obj.subclass
		},
		audio: func(id uint32, obj *doorCollideTestObject4E8AC0) {
			s.add("audio:%d:%s", id, doorCollideTestObjectName4E8AC0(obj))
			if s.onAudio != nil {
				s.onAudio(id, obj)
			}
		},
		priorityMessage: func(obj *doorCollideTestObject4E8AC0, message string) {
			s.add("priority:%s:%s", doorCollideTestObjectName4E8AC0(obj), message)
			s.priorityTarget = obj
		},
		loadLockCode: func(update *doorCollideTestUpdate4E8AC0) uint8 {
			s.add("lock:%d", update.lockCode)
			return update.lockCode
		},
		findKey: func(unit, door *doorCollideTestObject4E8AC0) *doorCollideTestObject4E8AC0 {
			s.add("find-key:%s:%s", unit.name, door.name)
			return s.key
		},
		keyMessage: func(obj *doorCollideTestObject4E8AC0, message string, lockCode uint8) {
			s.add("key-message:%s:%s:%d", doorCollideTestObjectName4E8AC0(obj), message, lockCode)
		},
		loadTileX: func(update *doorCollideTestUpdate4E8AC0) int32 {
			s.add("tile-x:%d", update.tileX)
			return update.tileX
		},
		loadTileY: func(update *doorCollideTestUpdate4E8AC0) int32 {
			s.add("tile-y:%d", update.tileY)
			return update.tileY
		},
		storeLockCode: func(update *doorCollideTestUpdate4E8AC0, value uint8) {
			s.add("store-lock:%d", value)
			update.lockCode = value
		},
		questMode: func() bool {
			s.add("quest:%t", s.quest)
			return s.quest
		},
		questSync: func(obj *doorCollideTestObject4E8AC0) int32 {
			s.add("quest-sync:%s", obj.name)
			if s.onQuestSync != nil {
				s.onQuestSync(obj)
			}
			return math.MinInt32
		},
		storeQuestFrame: func(frame uint32) {
			s.add("store-quest-frame:%d", frame)
			s.storedQuestFrame = frame
		},
		eachObjectInRect: func(rect doorCollideRect4E8AC0, target DoorTilePoint) {
			s.add("each-rect")
			s.rect = rect
			s.target = target
		},
		loadInventoryHolder: func(obj *doorCollideTestObject4E8AC0) *doorCollideTestObject4E8AC0 {
			s.add("holder:%s=%s", obj.name, doorCollideTestObjectName4E8AC0(obj.holder))
			return obj.holder
		},
		loadClassByte: func(obj *doorCollideTestObject4E8AC0) uint8 {
			s.add("class:%s:%d", obj.name, obj.class)
			return obj.class
		},
		questKeyState: func() int32 {
			if s.onQuestState != nil {
				s.onQuestState()
			}
			s.add("quest-state:%d", s.questState)
			return s.questState
		},
		delayedDelete: func(obj *doorCollideTestObject4E8AC0) {
			s.add("delete:%s", obj.name)
			s.deleted = obj
		},
	}
}

func TestDoorCollide4E8AC0EntryAndMotionGuards(t *testing.T) {
	update := &doorCollideTestUpdate4E8AC0{targetDirection: 8, currentDirection: 8}
	door := &doorCollideTestObject4E8AC0{name: "door", update: update}

	state := new(doorCollideTestState4E8AC0)
	doorCollide4E8AC0(door, nil, struct{}{}, state.hooks())
	if want := []string{"update:door"}; !reflect.DeepEqual(state.log, want) {
		t.Fatalf("nil-unit order = %v, want %v", state.log, want)
	}

	state = new(doorCollideTestState4E8AC0)
	unit := &doorCollideTestObject4E8AC0{name: "unit"}
	update.currentDirection = 7
	doorCollide4E8AC0(door, unit, struct{}{}, state.hooks())
	want := []string{"update:door", "current:7", "target:8"}
	if !reflect.DeepEqual(state.log, want) {
		t.Fatalf("motion guard order = %v, want %v", state.log, want)
	}
}

func TestDoorCollide4E8AC0OwnerExpiryAndMagicFeedback(t *testing.T) {
	unit := &doorCollideTestObject4E8AC0{name: "unit"}
	owner := &doorCollideTestObject4E8AC0{name: "owner"}
	update := &doorCollideTestUpdate4E8AC0{targetDirection: 8, currentDirection: 8}
	door := &doorCollideTestObject4E8AC0{
		name: "door", update: update, owner: owner, ownerFrame: 10,
	}

	expired := &doorCollideTestState4E8AC0{frameValue: 10}
	doorCollide4E8AC0(door, unit, struct{}{}, expired.hooks())
	wantExpired := []string{
		"update:door", "current:8", "target:8", "owner:owner",
		"owner-frame:10", "frame:10", "store-owner:nil", "lock:0",
	}
	if !reflect.DeepEqual(expired.log, wantExpired) {
		t.Fatalf("expired owner order = %v, want %v", expired.log, wantExpired)
	}
	if door.owner != nil {
		t.Fatalf("expired owner was not cleared: %v", door.owner)
	}

	door.owner = unit
	door.ownerFrame = 11
	sameOwner := &doorCollideTestState4E8AC0{frameValue: 10}
	doorCollide4E8AC0(door, unit, struct{}{}, sameOwner.hooks())
	wantSameOwner := []string{
		"update:door", "current:8", "target:8", "owner:unit",
		"owner-frame:11", "frame:10", "lock:0",
	}
	if !reflect.DeepEqual(sameOwner.log, wantSameOwner) {
		t.Fatalf("same owner order = %v, want %v", sameOwner.log, wantSameOwner)
	}

	door.owner = owner
	door.ownerFrame = 11
	door.subclass = doorCollideGateSubclassByte4E8AC0
	magic := &doorCollideTestState4E8AC0{
		frameValue: 10, tickValues: []uint64{2000, 2500}, feedbackTicks: 400,
	}
	doorCollide4E8AC0(door, unit, struct{}{}, magic.hooks())
	wantMagic := []string{
		"update:door", "current:8", "target:8", "owner:owner",
		"owner-frame:11", "frame:10", "ticks:2000", "feedback:400",
		"subclass:4", "audio:244:door",
		"priority:unit:objcoll.c:GateLockedMagic", "ticks:2500", "store-feedback:2500",
	}
	if !reflect.DeepEqual(magic.log, wantMagic) {
		t.Fatalf("magic feedback order = %v, want %v", magic.log, wantMagic)
	}

	cooldown := &doorCollideTestState4E8AC0{
		frameValue: 10, tickValues: []uint64{1900}, feedbackTicks: 400,
	}
	doorCollide4E8AC0(door, unit, struct{}{}, cooldown.hooks())
	if got := strings.Join(cooldown.log, "|"); strings.Contains(got, "audio:") || strings.Contains(got, "priority:") {
		t.Fatalf("feedback fired at the inclusive 1500 ms boundary: %v", cooldown.log)
	}
	if cooldown.feedbackTicks != 400 {
		t.Fatalf("cooldown timestamp changed to %d", cooldown.feedbackTicks)
	}
}

func TestDoorCollide4E8AC0FeedbackUsesUint64WrapAndLiveKeyCode(t *testing.T) {
	unit := &doorCollideTestObject4E8AC0{name: "unit"}
	update := &doorCollideTestUpdate4E8AC0{
		lockCode: 3, targetDirection: 16, currentDirection: 16,
	}
	door := &doorCollideTestObject4E8AC0{name: "door", update: update}
	state := &doorCollideTestState4E8AC0{
		tickValues:    []uint64{5, 9},
		feedbackTicks: math.MaxUint64 - 2000,
	}
	state.onAudio = func(uint32, *doorCollideTestObject4E8AC0) {
		update.lockCode = 4
	}
	doorCollide4E8AC0(door, unit, struct{}{}, state.hooks())
	wantSuffix := []string{
		"ticks:5", fmt.Sprintf("feedback:%d", uint64(math.MaxUint64-2000)),
		"subclass:0", "audio:240:door", "lock:4",
		"key-message:unit:objcoll.c:DoorLockedKey:4", "ticks:9", "store-feedback:9",
	}
	if len(state.log) < len(wantSuffix) || !reflect.DeepEqual(state.log[len(state.log)-len(wantSuffix):], wantSuffix) {
		t.Fatalf("wrapped key feedback suffix = %v, want %v", state.log, wantSuffix)
	}
}

func TestDoorCollide4E8AC0KeyFeedbackDoorAndGateMessages(t *testing.T) {
	for _, tc := range []struct {
		name     string
		subclass uint8
		want     string
	}{
		{"door", 0, "audio:240:door|lock:3|key-message:unit:objcoll.c:DoorLockedKey:3"},
		{"gate", doorCollideGateSubclassByte4E8AC0, "audio:244:door|lock:3|key-message:unit:objcoll.c:GateLockedKey:3"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			unit := &doorCollideTestObject4E8AC0{name: "unit"}
			update := &doorCollideTestUpdate4E8AC0{
				lockCode: 3, targetDirection: 8, currentDirection: 8,
			}
			door := &doorCollideTestObject4E8AC0{name: "door", update: update, subclass: tc.subclass}
			state := &doorCollideTestState4E8AC0{tickValues: []uint64{1501, 1700}}
			doorCollide4E8AC0(door, unit, struct{}{}, state.hooks())
			if got := strings.Join(state.log, "|"); !strings.Contains(got, tc.want) {
				t.Fatalf("key feedback = %v", state.log)
			}
		})
	}
}

func TestDoorCollide4E8AC0MechanismFeedback(t *testing.T) {
	for _, tc := range []struct {
		name     string
		subclass uint8
		want     string
	}{
		{"door", 0, "audio:240:door|priority:unit:objcoll.c:DoorLockedMechanism"},
		{"gate", doorCollideGateSubclassByte4E8AC0, "audio:244:door|priority:unit:objcoll.c:GateLockedMechanism"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			unit := &doorCollideTestObject4E8AC0{name: "unit"}
			update := &doorCollideTestUpdate4E8AC0{
				lockCode: 5, targetDirection: 0, currentDirection: 0,
			}
			door := &doorCollideTestObject4E8AC0{name: "door", update: update, subclass: tc.subclass}
			state := &doorCollideTestState4E8AC0{tickValues: []uint64{1501, 1700}}
			doorCollide4E8AC0(door, unit, struct{}{}, state.hooks())
			got := strings.Join(state.log, "|")
			if !strings.Contains(got, tc.want) {
				t.Fatalf("mechanism feedback = %v", state.log)
			}
			if strings.Contains(got, "find-key:") {
				t.Fatalf("mechanism lock searched for a key: %v", state.log)
			}
		})
	}
}

func TestDoorCollide4E8AC0UnlockOrderGeometryAndHolderReload(t *testing.T) {
	unit := &doorCollideTestObject4E8AC0{name: "unit"}
	holder1 := &doorCollideTestObject4E8AC0{name: "holder1", class: doorCollidePlayerClassByte4E8AC0}
	holder2 := &doorCollideTestObject4E8AC0{name: "holder2", class: doorCollidePlayerClassByte4E8AC0}
	key := &doorCollideTestObject4E8AC0{name: "key", holder: holder1}
	update := &doorCollideTestUpdate4E8AC0{
		lockCode: 2, targetDirection: 16, currentDirection: 16, tileX: 10, tileY: -3,
	}
	door := &doorCollideTestObject4E8AC0{name: "door", update: update}
	state := &doorCollideTestState4E8AC0{
		frameValue: 77, key: key, quest: true, questState: 1,
	}
	sawClearedLock := false
	state.onQuestSync = func(*doorCollideTestObject4E8AC0) {
		sawClearedLock = update.lockCode == 0
		state.frameValue = 88
	}
	state.onQuestState = func() {
		key.holder = holder2
	}

	doorCollide4E8AC0(door, unit, [2]float32{1, 2}, state.hooks())
	want := []string{
		"update:door", "current:16", "target:16", "owner:nil", "lock:2",
		"find-key:unit:door", "tile-x:10", "tile-y:-3", "store-lock:0", "target:16",
		"quest:true", "quest-sync:door", "frame:88", "store-quest-frame:88",
		"each-rect", "audio:234:door", "holder:key=holder1", "class:holder1:4",
		"quest:true", "quest-state:1", "holder:key=holder2",
		"priority:holder2:GeneralPrint:KeyShared1", "delete:key",
	}
	if !reflect.DeepEqual(state.log, want) {
		t.Fatalf("unlock order =\n%v\nwant\n%v", state.log, want)
	}
	if !sawClearedLock || update.lockCode != 0 {
		t.Fatalf("Quest sync did not observe the cleared lock: saw=%t lock=%d", sawClearedLock, update.lockCode)
	}
	if state.storedQuestFrame != 88 {
		t.Fatalf("stored Quest frame = %d, want callback-reloaded 88", state.storedQuestFrame)
	}
	wantRect := doorCollideRect4E8AC0{MinX: 196, MinY: -103, MaxX: 264, MaxY: -35}
	if state.rect != wantRect {
		t.Fatalf("unlock rect = %#v, want %#v", state.rect, wantRect)
	}
	if state.target != (DoorTilePoint{X: 11, Y: -2}) {
		t.Fatalf("unlock target = %#v, want {11 -2}", state.target)
	}
	if state.priorityTarget != holder2 || state.deleted != key {
		t.Fatalf("holder/delete result = %p/%p, want %p/%p", state.priorityTarget, state.deleted, holder2, key)
	}
}

func TestDoorCollideGeometry4E8AC0DirectionsAndX87YSpill(t *testing.T) {
	for _, tc := range []struct {
		direction int32
		want      DoorTilePoint
	}{
		{0, DoorTilePoint{X: 6, Y: 10}},
		{8, DoorTilePoint{X: 8, Y: 10}},
		{16, DoorTilePoint{X: 8, Y: 12}},
		{24, DoorTilePoint{X: 6, Y: 12}},
		{-1, DoorTilePoint{}},
		{25, DoorTilePoint{}},
	} {
		_, got := doorCollideGeometry4E8AC0(7, 11, tc.direction)
		if got != tc.want {
			t.Fatalf("direction %d target = %#v, want %#v", tc.direction, got, tc.want)
		}
	}

	const tileY = int32(729445)
	rect, _ := doorCollideGeometry4E8AC0(math.MaxInt32, tileY, 0)
	tileX := int32(math.MaxInt32)
	centerX := tileX * int32(23)
	centerY := tileY * int32(23)
	if centerX != 2147483625 {
		t.Fatalf("32-bit tile multiplication did not wrap: %d", centerX)
	}
	if rect.MinX != float32(float64(centerX)-34) || rect.MaxX != float32(float64(centerX)+34) {
		t.Fatalf("X rectangle did not use the exact int32 center: %#v", rect)
	}
	wantSpilled := float32(float64(float32(centerY)) + 34)
	wantUnspilled := float32(float64(centerY) + 34)
	if wantSpilled == wantUnspilled {
		t.Fatalf("test fixture does not expose the Y spill: center=%d result=%v", centerY, wantSpilled)
	}
	if rect.MaxY != wantSpilled {
		t.Fatalf("MaxY = %v, want spilled %v (unspilled %v)", rect.MaxY, wantSpilled, wantUnspilled)
	}
}
