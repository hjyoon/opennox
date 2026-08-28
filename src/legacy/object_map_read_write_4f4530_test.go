package legacy

import (
	"fmt"
	"math"
	"reflect"
	"testing"
)

type objectMapTestObject4F4530 struct {
	name          string
	field32       uint32
	field34       uint32
	extent        uint32
	scriptID      int32
	flags         uint32
	x             float32
	y             float32
	newX          float32
	newY          float32
	id            int
	team          uint8
	inventoryHead *objectMapTestObject4F4530
	inventoryNext *objectMapTestObject4F4530
	field129      *objectMapTestObject4F4530
	field128      *objectMapTestObject4F4530
	typeInd       uint16
	field5        uint32
	field189      int
}

type objectMapOldCall4F4530 struct {
	object        *objectMapTestObject4F4530
	objectVersion int32
	mapVersion    int32
}

type objectMapTestWorld4F4530 struct {
	readMode       bool
	readOnlyValue  int32
	readOnlyValues []int32
	u16Inputs      []uint16
	u8Inputs       []uint8
	u32Inputs      []uint32
	i32Inputs      []int32
	ownedInputs    []int32
	extentInput    uint32
	scriptIDInput  int32
	positionXInput float32
	positionYInput float32
	teamInput      uint8
	admission      int8
	oldResult      int32
	allocateID     int
	idLengths      map[int]uintptr
	allowedTypes   map[uint16]int32
	gameResults    map[uint32]int32
	nextScriptID   int32
	scriptResult   int32
	frame          uint32

	readOnlyCalls  int
	u16Calls       int
	u8Calls        int
	u32Calls       int
	i32Calls       int
	u16Args        []uint16
	u8Args         []uint8
	u32Args        []uint32
	i32Args        []int32
	oldCalls       []objectMapOldCall4F4530
	allocations    []uint16
	idTransfers    [][2]int
	terminations   [][2]int
	pending        [][2]int32
	ownedWrites    []string
	scriptContexts []int

	events  []string
	faultAt int
	after   map[string]func()
}

func objectMapTestName4F4530(object *objectMapTestObject4F4530) string {
	if object == nil {
		return "nil"
	}
	return object.name
}

func (w *objectMapTestWorld4F4530) event(name string) {
	w.events = append(w.events, name)
	if callback := w.after[name]; callback != nil {
		callback()
	}
	if w.faultAt != 0 && len(w.events) == w.faultAt {
		panic(name)
	}
}

func popObjectMap4F4530[T any](values *[]T, fallback T) T {
	if len(*values) == 0 {
		return fallback
	}
	value := (*values)[0]
	*values = (*values)[1:]
	return value
}

func (w *objectMapTestWorld4F4530) deps() objectMapReadWriteDeps4F4530[*objectMapTestObject4F4530, int] {
	return objectMapReadWriteDeps4F4530[*objectMapTestObject4F4530, int]{
		loadField34: func(object *objectMapTestObject4F4530) uint32 {
			w.event("load-field34:" + objectMapTestName4F4530(object))
			return object.field34
		},
		readOnly: func() int32 {
			w.readOnlyCalls++
			w.event(fmt.Sprintf("read-only:%d", w.readOnlyCalls))
			return popObjectMap4F4530(&w.readOnlyValues, w.readOnlyValue)
		},
		rwU16: func(value uint16) uint16 {
			w.u16Calls++
			w.event(fmt.Sprintf("rw-u16:%d", w.u16Calls))
			w.u16Args = append(w.u16Args, value)
			if w.readMode {
				return popObjectMap4F4530(&w.u16Inputs, value)
			}
			return value
		},
		readOld: func(object *objectMapTestObject4F4530, objectVersion, mapVersion int32) int32 {
			w.event("read-old:" + objectMapTestName4F4530(object))
			w.oldCalls = append(w.oldCalls, objectMapOldCall4F4530{object, objectVersion, mapVersion})
			return w.oldResult
		},
		storeField34: func(object *objectMapTestObject4F4530, value uint32) {
			w.event("store-field34:" + objectMapTestName4F4530(object))
			object.field34 = value
		},
		rwExtent: func(object *objectMapTestObject4F4530) {
			w.event("rw-extent:" + objectMapTestName4F4530(object))
			if w.readMode {
				object.extent = w.extentInput
			}
		},
		rwScriptID: func(object *objectMapTestObject4F4530) {
			w.event("rw-script-id:" + objectMapTestName4F4530(object))
			if w.readMode {
				object.scriptID = w.scriptIDInput
			}
		},
		loadScriptID: func(object *objectMapTestObject4F4530) int32 {
			w.event("load-script-id:" + objectMapTestName4F4530(object))
			return object.scriptID
		},
		gameFlags: func(mask uint32) int32 {
			w.event(fmt.Sprintf("game-flags:%08x", mask))
			return w.gameResults[mask]
		},
		nextScriptID: func() int32 {
			w.event("next-script-id")
			return w.nextScriptID
		},
		storeScriptID: func(object *objectMapTestObject4F4530, value int32) {
			w.event("store-script-id:" + objectMapTestName4F4530(object))
			object.scriptID = value
		},
		rwPositionX: func(object *objectMapTestObject4F4530) {
			w.event("rw-position-x:" + objectMapTestName4F4530(object))
			if w.readMode {
				object.x = w.positionXInput
			}
		},
		rwPositionY: func(object *objectMapTestObject4F4530) {
			w.event("rw-position-y:" + objectMapTestName4F4530(object))
			if w.readMode {
				object.y = w.positionYInput
			}
		},
		loadPositionX: func(object *objectMapTestObject4F4530) float32 {
			w.event("load-position-x:" + objectMapTestName4F4530(object))
			return object.x
		},
		loadPositionY: func(object *objectMapTestObject4F4530) float32 {
			w.event("load-position-y:" + objectMapTestName4F4530(object))
			return object.y
		},
		storeNewPositionX: func(object *objectMapTestObject4F4530, value float32) {
			w.event("store-new-position-x:" + objectMapTestName4F4530(object))
			object.newX = value
		},
		storeNewPositionY: func(object *objectMapTestObject4F4530, value float32) {
			w.event("store-new-position-y:" + objectMapTestName4F4530(object))
			object.newY = value
		},
		extendedAdmission: func(object *objectMapTestObject4F4530) int8 {
			w.event("extended-admission:" + objectMapTestName4F4530(object))
			return w.admission
		},
		rwU8: func(value uint8) uint8 {
			w.u8Calls++
			w.event(fmt.Sprintf("rw-u8:%d", w.u8Calls))
			w.u8Args = append(w.u8Args, value)
			if w.readMode {
				return popObjectMap4F4530(&w.u8Inputs, value)
			}
			return value
		},
		loadFlags: func(object *objectMapTestObject4F4530) uint32 {
			w.event("load-flags:" + objectMapTestName4F4530(object))
			return object.flags
		},
		rwU32: func(value uint32) uint32 {
			w.u32Calls++
			w.event(fmt.Sprintf("rw-u32:%d", w.u32Calls))
			w.u32Args = append(w.u32Args, value)
			if w.readMode {
				return popObjectMap4F4530(&w.u32Inputs, value)
			}
			return value
		},
		storeFlags: func(object *objectMapTestObject4F4530, value uint32) {
			w.event("store-flags:" + objectMapTestName4F4530(object))
			object.flags = value
		},
		setOn: func(object *objectMapTestObject4F4530) {
			w.event("set-on:" + objectMapTestName4F4530(object))
		},
		setOff: func(object *objectMapTestObject4F4530) {
			w.event("set-off:" + objectMapTestName4F4530(object))
		},
		loadIDPointer: func(object *objectMapTestObject4F4530) int {
			w.event("load-id:" + objectMapTestName4F4530(object))
			return object.id
		},
		stringLength: func(pointer int) uintptr {
			w.event("strlen")
			return w.idLengths[pointer]
		},
		allocateID: func(size uint16) int {
			w.event("allocate-id")
			w.allocations = append(w.allocations, size)
			return w.allocateID
		},
		storeIDPointer: func(object *objectMapTestObject4F4530, pointer int) {
			w.event("store-id:" + objectMapTestName4F4530(object))
			object.id = pointer
		},
		rwIDBytes: func(pointer int, length uint8) {
			w.event("rw-id")
			w.idTransfers = append(w.idTransfers, [2]int{pointer, int(length)})
		},
		terminateID: func(pointer int, length uint8) {
			w.event("terminate-id")
			w.terminations = append(w.terminations, [2]int{pointer, int(length)})
		},
		rwTeamID: func(object *objectMapTestObject4F4530) {
			w.event("rw-team:" + objectMapTestName4F4530(object))
			if w.readMode {
				object.team = w.teamInput
			}
		},
		loadInventoryHead: func(object *objectMapTestObject4F4530) *objectMapTestObject4F4530 {
			w.event("load-inventory-head:" + objectMapTestName4F4530(object))
			return object.inventoryHead
		},
		loadInventoryNext: func(object *objectMapTestObject4F4530) *objectMapTestObject4F4530 {
			w.event("load-inventory-next:" + objectMapTestName4F4530(object))
			return object.inventoryNext
		},
		loadField129: func(object *objectMapTestObject4F4530) *objectMapTestObject4F4530 {
			w.event("load-field129:" + objectMapTestName4F4530(object))
			return object.field129
		},
		loadTypeInd: func(object *objectMapTestObject4F4530) uint16 {
			w.event("load-type-ind:" + objectMapTestName4F4530(object))
			return object.typeInd
		},
		ownedTypeAllowed: func(typeInd uint16) int32 {
			w.event(fmt.Sprintf("owned-type-allowed:%04x", typeInd))
			return w.allowedTypes[typeInd]
		},
		loadField128: func(object *objectMapTestObject4F4530) *objectMapTestObject4F4530 {
			w.event("load-field128:" + objectMapTestName4F4530(object))
			return object.field128
		},
		rwI32: func(value int32) int32 {
			w.i32Calls++
			w.event(fmt.Sprintf("rw-i32:%d", w.i32Calls))
			w.i32Args = append(w.i32Args, value)
			if w.readMode {
				return popObjectMap4F4530(&w.i32Inputs, value)
			}
			return value
		},
		readOwnedScriptID: func() int32 {
			w.event("read-owned-script-id")
			return popObjectMap4F4530(&w.ownedInputs, int32(0))
		},
		addPendingOwn: func(ownerScriptID, ownedScriptID int32) {
			w.event("add-pending-own")
			w.pending = append(w.pending, [2]int32{ownerScriptID, ownedScriptID})
		},
		rwOwnedScriptID: func(object *objectMapTestObject4F4530) {
			w.event("rw-owned-script-id:" + objectMapTestName4F4530(object))
			w.ownedWrites = append(w.ownedWrites, object.name)
		},
		loadField5: func(object *objectMapTestObject4F4530) uint32 {
			w.event("load-field5:" + objectMapTestName4F4530(object))
			return object.field5
		},
		unsetStatus: func(object *objectMapTestObject4F4530, status uint32) {
			w.event("unset-status:" + objectMapTestName4F4530(object))
			object.field5 &^= status
		},
		setStatus: func(object *objectMapTestObject4F4530, status uint32) {
			w.event("set-status:" + objectMapTestName4F4530(object))
			object.field5 |= status
		},
		loadField189: func(object *objectMapTestObject4F4530) int {
			w.event("load-field189:" + objectMapTestName4F4530(object))
			return object.field189
		},
		scriptHandler: func(object *objectMapTestObject4F4530, context int) int32 {
			w.event("script-handler:" + objectMapTestName4F4530(object))
			w.scriptContexts = append(w.scriptContexts, context)
			return w.scriptResult
		},
		gameFrame: func() uint32 {
			w.event("game-frame")
			return w.frame
		},
		storeField32: func(object *objectMapTestObject4F4530, value uint32) {
			w.event("store-field32:" + objectMapTestName4F4530(object))
			object.field32 = value
		},
	}
}

func newObjectMapReadWorld4F4530() (*objectMapTestWorld4F4530, *objectMapTestObject4F4530) {
	inv2 := &objectMapTestObject4F4530{name: "inv2"}
	inv1 := &objectMapTestObject4F4530{name: "inv1", inventoryNext: inv2}
	owned2 := &objectMapTestObject4F4530{name: "owned2", typeInd: 2}
	owned1 := &objectMapTestObject4F4530{name: "owned1", typeInd: 1, field128: owned2}
	object := &objectMapTestObject4F4530{
		name:          "object",
		field34:       1000,
		flags:         0x80000040,
		field5:        0xa5,
		inventoryHead: inv1,
		field129:      owned1,
		field189:      77,
	}
	w := &objectMapTestWorld4F4530{
		readMode:       true,
		readOnlyValue:  1,
		u16Inputs:      []uint16{64, 2},
		u8Inputs:       []uint8{1, 3, 2},
		u32Inputs:      []uint32{0x01400002, 0x0000005e},
		i32Inputs:      []int32{25},
		ownedInputs:    []int32{-7, 42},
		extentInput:    0xdeadbeef,
		scriptIDInput:  0,
		positionXInput: math.Float32frombits(0x7fc12345),
		positionYInput: math.Float32frombits(0x80000000),
		teamInput:      9,
		admission:      -1,
		oldResult:      1,
		allocateID:     7,
		idLengths:      make(map[int]uintptr),
		allowedTypes:   map[uint16]int32{1: 1, 2: 1},
		gameResults:    make(map[uint32]int32),
		nextScriptID:   123,
		scriptResult:   1,
		frame:          900,
		after:          make(map[string]func()),
	}
	return w, object
}

func objectMapFullReadEvents4F4530() []string {
	return []string{
		"load-field34:object", "rw-u16:1", "read-only:1", "store-field34:object",
		"rw-extent:object", "rw-script-id:object", "load-script-id:object", "read-only:2",
		"game-flags:00200000", "game-flags:00400000", "next-script-id", "store-script-id:object",
		"read-only:3", "rw-position-x:object", "rw-position-y:object", "load-position-x:object", "load-position-y:object",
		"store-new-position-x:object", "store-new-position-y:object", "extended-admission:object", "rw-u8:1",
		"load-flags:object", "rw-u32:1", "load-flags:object", "store-flags:object", "store-flags:object",
		"load-flags:object", "store-flags:object", "read-only:4", "set-on:object",
		"load-id:object", "rw-u8:2", "read-only:5", "allocate-id", "store-id:object", "load-id:object", "rw-id",
		"load-id:object", "terminate-id", "rw-team:object", "load-inventory-head:object",
		"load-inventory-next:inv1", "load-inventory-next:inv2", "rw-u8:3", "read-only:6", "store-field34:object",
		"load-field129:object", "load-flags:owned1", "load-type-ind:owned1", "owned-type-allowed:0001", "load-field128:owned1",
		"load-flags:owned2", "load-type-ind:owned2", "owned-type-allowed:0002", "load-field128:owned2", "rw-u16:2",
		"read-only:7", "read-owned-script-id", "game-flags:00200000", "game-flags:00400000", "load-script-id:object", "add-pending-own",
		"read-owned-script-id", "game-flags:00200000", "game-flags:00400000", "load-script-id:object", "add-pending-own",
		"load-field5:object", "rw-u32:2", "unset-status:object", "set-status:object",
		"load-field189:object", "script-handler:object", "game-frame", "rw-i32:1", "read-only:8", "load-flags:object", "store-field32:object",
	}
}

func TestObjectMapReadWrite4F4530FullReadOrderAndValues(t *testing.T) {
	w, object := newObjectMapReadWorld4F4530()
	if got := objectMapReadWrite4F4530(object, 40, w.deps()); got != 1 {
		t.Fatalf("result = %d, want 1", got)
	}
	if !reflect.DeepEqual(w.events, objectMapFullReadEvents4F4530()) {
		t.Fatalf("events =\n%q\nwant\n%q", w.events, objectMapFullReadEvents4F4530())
	}
	if object.field34 != 2 || object.extent != 0xdeadbeef || object.scriptID != 123 || object.team != 9 || object.field32 != 25 {
		t.Fatalf("object scalars = field34:%d extent:%08x script:%d team:%d field32:%d", object.field34, object.extent, object.scriptID, object.team, object.field32)
	}
	if math.Float32bits(object.x) != 0x7fc12345 || math.Float32bits(object.newX) != 0x7fc12345 ||
		math.Float32bits(object.y) != 0x80000000 || math.Float32bits(object.newY) != 0x80000000 {
		t.Fatalf("position bits = x:%08x y:%08x newX:%08x newY:%08x", math.Float32bits(object.x), math.Float32bits(object.y), math.Float32bits(object.newX), math.Float32bits(object.newY))
	}
	if !reflect.DeepEqual(w.u8Args, []uint8{0xff, 0, 2}) || !reflect.DeepEqual(w.u16Args, []uint16{64, 2}) {
		t.Fatalf("wire args = u8:%v u16:%v", w.u8Args, w.u16Args)
	}
	if object.id != 7 || !reflect.DeepEqual(w.allocations, []uint16{4}) ||
		!reflect.DeepEqual(w.idTransfers, [][2]int{{7, 3}}) || !reflect.DeepEqual(w.terminations, [][2]int{{7, 3}}) {
		t.Fatalf("ID state = %d allocations:%v transfers:%v terminations:%v", object.id, w.allocations, w.idTransfers, w.terminations)
	}
	if !reflect.DeepEqual(w.pending, [][2]int32{{123, -7}, {123, 42}}) {
		t.Fatalf("pending owns = %v", w.pending)
	}
	if object.field5 != 0xff || !reflect.DeepEqual(w.scriptContexts, []int{77}) {
		t.Fatalf("status/context = %#x/%v", object.field5, w.scriptContexts)
	}
	if !reflect.DeepEqual(w.i32Args, []int32{100}) {
		t.Fatalf("frame delta arg = %v, want [100]", w.i32Args)
	}
}

func TestObjectMapReadWrite4F4530FullReadFaultPrefixes(t *testing.T) {
	want := objectMapFullReadEvents4F4530()
	for faultAt := 1; faultAt <= len(want); faultAt++ {
		w, object := newObjectMapReadWorld4F4530()
		w.faultAt = faultAt
		func() {
			defer func() {
				if recover() == nil {
					t.Fatalf("fault %d did not panic", faultAt)
				}
			}()
			objectMapReadWrite4F4530(object, 40, w.deps())
		}()
		if !reflect.DeepEqual(w.events, want[:faultAt]) {
			t.Fatalf("fault %d events = %q, want %q", faultAt, w.events, want[:faultAt])
		}
	}
}

func TestObjectMapReadWrite4F4530VersionRouting(t *testing.T) {
	tests := []struct {
		name           string
		mapVersion     int32
		readOnly       int32
		version        uint16
		wantResult     int32
		wantOld        *objectMapOldCall4F4530
		wantEvents     []string
		wantVersionArg []uint16
	}{
		{name: "pre40 read omits version", mapVersion: 39, readOnly: 1, wantResult: 7,
			wantOld:    &objectMapOldCall4F4530{objectVersion: 0, mapVersion: 39},
			wantEvents: []string{"load-field34:object", "read-only:1", "read-old:object"}},
		{name: "pre40 write emits current version", mapVersion: 39, readOnly: 0, wantResult: 7,
			wantOld:    &objectMapOldCall4F4530{objectVersion: 64, mapVersion: 39},
			wantEvents: []string{"load-field34:object", "read-only:1", "rw-u16:1", "read-old:object"}, wantVersionArg: []uint16{64}},
		{name: "version60 delegates", mapVersion: 40, readOnly: 1, version: 60, wantResult: 7,
			wantOld:    &objectMapOldCall4F4530{objectVersion: 60, mapVersion: 40},
			wantEvents: []string{"load-field34:object", "rw-u16:1", "read-old:object"}, wantVersionArg: []uint16{64}},
		{name: "negative version delegates", mapVersion: 40, readOnly: 1, version: 0xffff, wantResult: 7,
			wantOld:    &objectMapOldCall4F4530{objectVersion: -1, mapVersion: 40},
			wantEvents: []string{"load-field34:object", "rw-u16:1", "read-old:object"}, wantVersionArg: []uint16{64}},
		{name: "version65 rejects", mapVersion: 40, readOnly: 1, version: 65, wantResult: 0,
			wantEvents: []string{"load-field34:object", "rw-u16:1"}, wantVersionArg: []uint16{64}},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			w, object := newObjectMapReadWorld4F4530()
			w.readOnlyValue = tc.readOnly
			w.oldResult = 7
			w.u16Inputs = nil
			if tc.mapVersion >= 40 {
				w.u16Inputs = []uint16{tc.version}
			}
			if got := objectMapReadWrite4F4530(object, tc.mapVersion, w.deps()); got != tc.wantResult {
				t.Fatalf("result = %d, want %d", got, tc.wantResult)
			}
			if !reflect.DeepEqual(w.events, tc.wantEvents) || !reflect.DeepEqual(w.u16Args, tc.wantVersionArg) {
				t.Fatalf("events/version args = %q/%v, want %q/%v", w.events, w.u16Args, tc.wantEvents, tc.wantVersionArg)
			}
			if tc.wantOld == nil {
				if len(w.oldCalls) != 0 {
					t.Fatalf("old calls = %+v, want none", w.oldCalls)
				}
				return
			}
			if len(w.oldCalls) != 1 {
				t.Fatalf("old calls = %+v", w.oldCalls)
			}
			got := w.oldCalls[0]
			if got.object != object || got.objectVersion != tc.wantOld.objectVersion || got.mapVersion != tc.wantOld.mapVersion {
				t.Fatalf("old call = %+v, want version/map %d/%d", got, tc.wantOld.objectVersion, tc.wantOld.mapVersion)
			}
		})
	}
}

func TestObjectMapReadWrite4F4530ZeroAdmissionStopsAfterByte(t *testing.T) {
	w, object := newObjectMapReadWorld4F4530()
	w.u16Inputs = []uint16{61}
	w.u8Inputs = []uint8{0}
	if got := objectMapReadWrite4F4530(object, 40, w.deps()); got != 1 {
		t.Fatalf("result = %d, want 1", got)
	}
	if got := w.events[len(w.events)-1]; got != "rw-u8:1" {
		t.Fatalf("last event = %q, want admission byte", got)
	}
	for _, event := range w.events {
		if event == "load-flags:object" || event == "rw-team:object" {
			t.Fatalf("zero admission reached extended event %q", event)
		}
	}
}

func TestObjectMapReadWrite4F4530AllocationFailureStopsBeforeIDTransfer(t *testing.T) {
	w, object := newObjectMapReadWorld4F4530()
	w.allocateID = 0
	if got := objectMapReadWrite4F4530(object, 40, w.deps()); got != 0 {
		t.Fatalf("result = %d, want 0", got)
	}
	if got := w.events[len(w.events)-1]; got != "store-id:object" {
		t.Fatalf("last event = %q, want failed pointer store", got)
	}
	if len(w.idTransfers) != 0 {
		t.Fatalf("ID transfers after failure = %v", w.idTransfers)
	}
}

func TestObjectMapReadWrite4F4530WritePreservesBytesAndLiveChains(t *testing.T) {
	w, object := newObjectMapReadWorld4F4530()
	w.readMode = false
	w.readOnlyValue = 0
	w.u16Inputs = nil
	w.u8Inputs = nil
	w.u32Inputs = nil
	w.i32Inputs = nil
	w.ownedInputs = nil
	object.scriptID = 55
	object.id = 9
	w.idLengths[9] = 260
	object.team = 11
	originalFlags := object.flags
	originalStatus := object.field5
	w.frame = 1005
	if got := objectMapReadWrite4F4530(object, 40, w.deps()); got != 1 {
		t.Fatalf("result = %d, want 1", got)
	}
	if !reflect.DeepEqual(w.u16Args, []uint16{64, 2}) || !reflect.DeepEqual(w.u8Args, []uint8{0xff, 4, 2}) {
		t.Fatalf("write wire args = u16:%v u8:%v", w.u16Args, w.u8Args)
	}
	if !reflect.DeepEqual(w.idTransfers, [][2]int{{9, 4}}) || !reflect.DeepEqual(w.terminations, [][2]int{{9, 4}}) {
		t.Fatalf("write ID = transfers:%v terminations:%v", w.idTransfers, w.terminations)
	}
	if !reflect.DeepEqual(w.ownedWrites, []string{"owned1", "owned2"}) {
		t.Fatalf("owned writes = %v", w.ownedWrites)
	}
	if object.flags != originalFlags || object.field5 != originalStatus || object.field34 != 1000 || object.field32 != 0 {
		t.Fatalf("post-write object = flags:%08x status:%08x field34:%d field32:%d", object.flags, object.field5, object.field34, object.field32)
	}
	if !reflect.DeepEqual(w.i32Args, []int32{-5}) {
		t.Fatalf("wrapping frame delta = %v, want [-5]", w.i32Args)
	}
}

func TestObjectMapReadWrite4F4530ScriptFailureSkipsFrame(t *testing.T) {
	w, object := newObjectMapReadWorld4F4530()
	w.scriptResult = 0
	if got := objectMapReadWrite4F4530(object, 40, w.deps()); got != 0 {
		t.Fatalf("result = %d, want 0", got)
	}
	if got := w.events[len(w.events)-1]; got != "script-handler:object" {
		t.Fatalf("last event = %q, want script handler", got)
	}
	if len(w.i32Args) != 0 {
		t.Fatalf("frame transfer after script failure = %v", w.i32Args)
	}
}

func TestObjectMapReadWrite4F4530NonExactReadSkipsExactOneEffects(t *testing.T) {
	w, object := newObjectMapReadWorld4F4530()
	w.readOnlyValue = 2
	w.u8Inputs = []uint8{1, 0, 0}
	w.u16Inputs = []uint16{64, 0}
	w.u32Inputs = []uint32{objectMapDecayFlag4F4530, 0}
	w.ownedInputs = nil
	w.i32Inputs = []int32{9}
	object.inventoryHead = nil
	object.field129 = nil
	object.scriptID = 8
	if got := objectMapReadWrite4F4530(object, 40, w.deps()); got != 1 {
		t.Fatalf("result = %d, want 1", got)
	}
	if object.field34 != 1000 || object.field32 != 0 {
		t.Fatalf("exact-one fields changed = field34:%d field32:%d", object.field34, object.field32)
	}
	for _, event := range w.events {
		if event == "set-on:object" || event == "set-off:object" || event == "store-field32:object" {
			t.Fatalf("non-exact read executed %q", event)
		}
	}
}

func TestObjectMapReadWrite4F4530UsesLiveFlagsAndOwnerScriptID(t *testing.T) {
	w, object := newObjectMapReadWorld4F4530()
	w.after["rw-u32:1"] = func() { object.flags = 0x60000040 }
	w.after["read-owned-script-id"] = func() { object.scriptID++ }
	if got := objectMapReadWrite4F4530(object, 40, w.deps()); got != 1 {
		t.Fatalf("result = %d, want 1", got)
	}
	wantFlags := uint32(0x60000040)&objectMapFlagsKeepMask4F4530 | objectMapPreserveFlag4F4530 | 0x01400002
	if object.flags != wantFlags {
		t.Fatalf("live flags = %08x, want %08x", object.flags, wantFlags)
	}
	if !reflect.DeepEqual(w.pending, [][2]int32{{124, -7}, {125, 42}}) {
		t.Fatalf("live pending owners = %v", w.pending)
	}
}
