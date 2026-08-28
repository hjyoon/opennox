package legacy

import (
	"fmt"
	"math"
	"reflect"
	"testing"
)

type objectReadOldTestObject4F4170 struct {
	name          string
	field34       uint32
	extent        uint32
	flags         uint32
	x             float32
	y             float32
	newX          float32
	newY          float32
	id            int
	team          uint8
	inventoryHead *objectReadOldTestObject4F4170
	inventoryNext *objectReadOldTestObject4F4170
	scriptID      int32
	field129      *objectReadOldTestObject4F4170
	field128      *objectReadOldTestObject4F4170
	typeInd       uint16
	field5        uint32
}

type objectReadOldTestWorld4F4170 struct {
	readMode       bool
	readOnlyValues []int32
	readOnlyValue  int32
	extentInput    uint32
	positionXInput float32
	positionYInput float32
	oldPosition    [2]int32
	teamInput      uint8
	scriptIDInput  int32
	ownedIDInputs  []int32
	u32Inputs      []uint32
	u16Inputs      []uint16
	u8Inputs       []uint8

	u32Args      []uint32
	u16Args      []uint16
	u8Args       []uint8
	i32Args      []int32
	oldPosArgs   [][2]int32
	allocated    []uint16
	allocateID   int
	idLengths    map[int]uintptr
	idTransfers  [][2]int
	terminations [][2]int
	gameResults  map[uint32]int32
	allowedTypes map[uint16]int32
	nextID       int32
	pending      [][2]int32
	ownedWrites  []string

	events  []string
	faultAt int
	after   map[string]func()
}

func objectReadOldTestName4F4170(object *objectReadOldTestObject4F4170) string {
	if object == nil {
		return "nil"
	}
	return object.name
}

func (w *objectReadOldTestWorld4F4170) event(name string) {
	w.events = append(w.events, name)
	if callback := w.after[name]; callback != nil {
		callback()
	}
	if w.faultAt != 0 && len(w.events) == w.faultAt {
		panic(name)
	}
}

func popFirst4F4170[T any](values *[]T, fallback T) T {
	if len(*values) == 0 {
		return fallback
	}
	value := (*values)[0]
	*values = (*values)[1:]
	return value
}

func (w *objectReadOldTestWorld4F4170) deps() objectReadOldDeps4F4170[*objectReadOldTestObject4F4170, int] {
	return objectReadOldDeps4F4170[*objectReadOldTestObject4F4170, int]{
		readOnly: func() int32 {
			w.event("read-only")
			return popFirst4F4170(&w.readOnlyValues, w.readOnlyValue)
		},
		storeField34: func(object *objectReadOldTestObject4F4170, value uint32) {
			w.event("store-field34:" + objectReadOldTestName4F4170(object))
			object.field34 = value
		},
		rwExtent: func(object *objectReadOldTestObject4F4170) {
			w.event("rw-extent:" + objectReadOldTestName4F4170(object))
			if w.readMode {
				object.extent = w.extentInput
			}
		},
		loadFlags: func(object *objectReadOldTestObject4F4170) uint32 {
			w.event("load-flags:" + objectReadOldTestName4F4170(object))
			return object.flags
		},
		rwU32: func(value uint32) uint32 {
			w.event("rw-u32")
			w.u32Args = append(w.u32Args, value)
			if w.readMode {
				return popFirst4F4170(&w.u32Inputs, value)
			}
			return value
		},
		storeFlags: func(object *objectReadOldTestObject4F4170, value uint32) {
			w.event("store-flags:" + objectReadOldTestName4F4170(object))
			object.flags = value
		},
		setOn: func(object *objectReadOldTestObject4F4170) {
			w.event("set-on:" + objectReadOldTestName4F4170(object))
		},
		setOff: func(object *objectReadOldTestObject4F4170) {
			w.event("set-off:" + objectReadOldTestName4F4170(object))
		},
		rwPositionX: func(object *objectReadOldTestObject4F4170) {
			w.event("rw-position-x:" + objectReadOldTestName4F4170(object))
			if w.readMode {
				object.x = w.positionXInput
			}
		},
		rwPositionY: func(object *objectReadOldTestObject4F4170) {
			w.event("rw-position-y:" + objectReadOldTestName4F4170(object))
			if w.readMode {
				object.y = w.positionYInput
			}
		},
		rwOldPosition: func(x, y int32) (int32, int32) {
			w.event("rw-old-position")
			w.oldPosArgs = append(w.oldPosArgs, [2]int32{x, y})
			if w.readMode {
				return w.oldPosition[0], w.oldPosition[1]
			}
			return x, y
		},
		storePositionX: func(object *objectReadOldTestObject4F4170, value float32) {
			w.event("store-position-x:" + objectReadOldTestName4F4170(object))
			object.x = value
		},
		storePositionY: func(object *objectReadOldTestObject4F4170, value float32) {
			w.event("store-position-y:" + objectReadOldTestName4F4170(object))
			object.y = value
		},
		loadPositionX: func(object *objectReadOldTestObject4F4170) float32 {
			w.event("load-position-x:" + objectReadOldTestName4F4170(object))
			return object.x
		},
		loadPositionY: func(object *objectReadOldTestObject4F4170) float32 {
			w.event("load-position-y:" + objectReadOldTestName4F4170(object))
			return object.y
		},
		storeNewPositionX: func(object *objectReadOldTestObject4F4170, value float32) {
			w.event("store-new-position-x:" + objectReadOldTestName4F4170(object))
			object.newX = value
		},
		storeNewPositionY: func(object *objectReadOldTestObject4F4170, value float32) {
			w.event("store-new-position-y:" + objectReadOldTestName4F4170(object))
			object.newY = value
		},
		loadIDPointer: func(object *objectReadOldTestObject4F4170) int {
			w.event("load-id:" + objectReadOldTestName4F4170(object))
			return object.id
		},
		stringLength: func(pointer int) uintptr {
			w.event("strlen")
			return w.idLengths[pointer]
		},
		rwU8: func(value uint8) uint8 {
			w.event("rw-u8")
			w.u8Args = append(w.u8Args, value)
			if w.readMode {
				return popFirst4F4170(&w.u8Inputs, value)
			}
			return value
		},
		allocateID: func(size uint16) int {
			w.event("allocate-id")
			w.allocated = append(w.allocated, size)
			return w.allocateID
		},
		storeIDPointer: func(object *objectReadOldTestObject4F4170, pointer int) {
			w.event("store-id:" + objectReadOldTestName4F4170(object))
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
		rwTeamID: func(object *objectReadOldTestObject4F4170) {
			w.event("rw-team:" + objectReadOldTestName4F4170(object))
			if w.readMode {
				object.team = w.teamInput
			}
		},
		loadInventoryHead: func(object *objectReadOldTestObject4F4170) *objectReadOldTestObject4F4170 {
			w.event("load-inventory-head:" + objectReadOldTestName4F4170(object))
			return object.inventoryHead
		},
		loadInventoryNext: func(object *objectReadOldTestObject4F4170) *objectReadOldTestObject4F4170 {
			w.event("load-inventory-next:" + objectReadOldTestName4F4170(object))
			return object.inventoryNext
		},
		rwScriptID: func(object *objectReadOldTestObject4F4170) {
			w.event("rw-script-id:" + objectReadOldTestName4F4170(object))
			if w.readMode {
				object.scriptID = w.scriptIDInput
			}
		},
		loadScriptID: func(object *objectReadOldTestObject4F4170) int32 {
			w.event("load-script-id:" + objectReadOldTestName4F4170(object))
			return object.scriptID
		},
		storeScriptID: func(object *objectReadOldTestObject4F4170, value int32) {
			w.event("store-script-id:" + objectReadOldTestName4F4170(object))
			object.scriptID = value
		},
		gameFlags: func(mask uint32) int32 {
			w.event(fmt.Sprintf("game-flags:%08x", mask))
			return w.gameResults[mask]
		},
		nextScriptID: func() int32 {
			w.event("next-script-id")
			return w.nextID
		},
		loadField129: func(object *objectReadOldTestObject4F4170) *objectReadOldTestObject4F4170 {
			w.event("load-field129:" + objectReadOldTestName4F4170(object))
			return object.field129
		},
		loadTypeInd: func(object *objectReadOldTestObject4F4170) uint16 {
			w.event("load-type-ind:" + objectReadOldTestName4F4170(object))
			return object.typeInd
		},
		ownedTypeAllowed: func(typeInd uint16) int32 {
			w.event(fmt.Sprintf("owned-type-allowed:%04x", typeInd))
			return w.allowedTypes[typeInd]
		},
		loadField128: func(object *objectReadOldTestObject4F4170) *objectReadOldTestObject4F4170 {
			w.event("load-field128:" + objectReadOldTestName4F4170(object))
			return object.field128
		},
		rwU16: func(value uint16) uint16 {
			w.event("rw-u16")
			w.u16Args = append(w.u16Args, value)
			if w.readMode {
				return popFirst4F4170(&w.u16Inputs, value)
			}
			return value
		},
		rwI32: func(value int32) int32 {
			w.event("rw-i32")
			w.i32Args = append(w.i32Args, value)
			if w.readMode {
				return popFirst4F4170(&w.ownedIDInputs, value)
			}
			return value
		},
		addPendingOwn: func(ownerScriptID, ownedScriptID int32) {
			w.event("add-pending-own")
			w.pending = append(w.pending, [2]int32{ownerScriptID, ownedScriptID})
		},
		rwOwnedScriptID: func(object *objectReadOldTestObject4F4170) {
			w.event("rw-owned-script-id:" + objectReadOldTestName4F4170(object))
			w.ownedWrites = append(w.ownedWrites, object.name)
		},
		loadField5: func(object *objectReadOldTestObject4F4170) uint32 {
			w.event("load-field5:" + objectReadOldTestName4F4170(object))
			return object.field5
		},
		unsetStatus: func(object *objectReadOldTestObject4F4170, status uint32) {
			w.event("unset-status:" + objectReadOldTestName4F4170(object))
			object.field5 &^= status
		},
		setStatus: func(object *objectReadOldTestObject4F4170, status uint32) {
			w.event("set-status:" + objectReadOldTestName4F4170(object))
			object.field5 |= status
		},
	}
}

func newObjectReadOldReadWorld4F4170() (*objectReadOldTestWorld4F4170, *objectReadOldTestObject4F4170) {
	object := &objectReadOldTestObject4F4170{
		name:    "object",
		field34: 99,
		flags:   0xa2400040,
		field5:  0xa5,
	}
	w := &objectReadOldTestWorld4F4170{
		readMode:       true,
		readOnlyValue:  1,
		extentInput:    0xdeadbeef,
		positionXInput: math.Float32frombits(0x7fc12345),
		positionYInput: math.Float32frombits(0x80000000),
		teamInput:      0x80,
		scriptIDInput:  0,
		ownedIDInputs:  []int32{-7, 42},
		u32Inputs:      []uint32{0x01000002, 0x0000005e},
		u16Inputs:      []uint16{2},
		u8Inputs:       []uint8{3, 2},
		allocateID:     7,
		idLengths:      make(map[int]uintptr),
		gameResults:    make(map[uint32]int32),
		allowedTypes:   make(map[uint16]int32),
		nextID:         123,
		after:          make(map[string]func()),
	}
	return w, object
}

func TestObjectReadOldVer4F4170FullReadOrderAndValues(t *testing.T) {
	w, object := newObjectReadOldReadWorld4F4170()
	if got := objectReadOldVer4F4170(object, 5, 40, w.deps()); got != 1 {
		t.Fatalf("result = %d, want 1", got)
	}
	if object.field34 != 2 || object.extent != 0xdeadbeef || object.team != 0x80 || object.scriptID != 123 {
		t.Fatalf("object scalar state = field34:%d extent:%08x team:%02x script:%d", object.field34, object.extent, object.team, object.scriptID)
	}
	if math.Float32bits(object.x) != 0x7fc12345 || math.Float32bits(object.newX) != 0x7fc12345 ||
		math.Float32bits(object.y) != 0x80000000 || math.Float32bits(object.newY) != 0x80000000 {
		t.Fatalf("position bits = x:%08x y:%08x newX:%08x newY:%08x", math.Float32bits(object.x), math.Float32bits(object.y), math.Float32bits(object.newX), math.Float32bits(object.newY))
	}
	if object.id != 7 || !reflect.DeepEqual(w.allocated, []uint16{4}) ||
		!reflect.DeepEqual(w.idTransfers, [][2]int{{7, 3}}) ||
		!reflect.DeepEqual(w.terminations, [][2]int{{7, 3}}) {
		t.Fatalf("ID state = pointer:%d allocations:%v transfers:%v terminations:%v", object.id, w.allocated, w.idTransfers, w.terminations)
	}
	if !reflect.DeepEqual(w.pending, [][2]int32{{123, -7}, {123, 42}}) {
		t.Fatalf("pending owns = %v", w.pending)
	}
	if !reflect.DeepEqual(w.i32Args, []int32{40, 40}) {
		t.Fatalf("owned ID transfer initial values = %v, want map version twice", w.i32Args)
	}
	if object.field5 != 0xff {
		t.Fatalf("field5 = %#x, want 0xff", object.field5)
	}

	wantEvents := []string{
		"read-only", "store-field34:object", "rw-extent:object",
		"load-flags:object", "rw-u32", "load-flags:object", "store-flags:object", "store-flags:object", "load-flags:object", "store-flags:object",
		"read-only", "set-on:object", "read-only",
		"rw-position-x:object", "rw-position-y:object", "load-position-x:object", "load-position-y:object", "store-new-position-x:object", "store-new-position-y:object",
		"load-id:object", "rw-u8", "read-only", "allocate-id", "store-id:object", "load-id:object", "rw-id", "load-id:object", "terminate-id",
		"rw-team:object", "load-inventory-head:object", "rw-u8", "read-only", "store-field34:object",
		"rw-script-id:object", "load-script-id:object", "read-only", "game-flags:00200000", "game-flags:00400000", "next-script-id", "store-script-id:object",
		"load-field129:object", "rw-u16", "read-only",
		"rw-i32", "game-flags:00200000", "game-flags:00400000", "load-script-id:object", "add-pending-own",
		"rw-i32", "game-flags:00200000", "game-flags:00400000", "load-script-id:object", "add-pending-own",
		"load-field5:object", "rw-u32", "unset-status:object", "set-status:object",
	}
	if !reflect.DeepEqual(w.events, wantEvents) {
		t.Fatalf("events =\n%q\nwant =\n%q", w.events, wantEvents)
	}
}

func TestObjectReadOldVer4F4170FullReadFaultPrefixes(t *testing.T) {
	baseline, baselineObject := newObjectReadOldReadWorld4F4170()
	objectReadOldVer4F4170(baselineObject, 5, 40, baseline.deps())
	want := baseline.events
	for faultAt := 1; faultAt <= len(want); faultAt++ {
		t.Run(fmt.Sprintf("fault-%d", faultAt), func(t *testing.T) {
			w, object := newObjectReadOldReadWorld4F4170()
			w.faultAt = faultAt
			defer func() {
				if got := recover(); got != want[faultAt-1] {
					t.Fatalf("panic = %v, want %q", got, want[faultAt-1])
				}
				if !reflect.DeepEqual(w.events, want[:faultAt]) {
					t.Fatalf("events = %v, want %v", w.events, want[:faultAt])
				}
			}()
			objectReadOldVer4F4170(object, 5, 40, w.deps())
		})
	}
}

func TestObjectReadOldVer4F4170OldPositionAndReadOnlyPredicates(t *testing.T) {
	w, object := newObjectReadOldReadWorld4F4170()
	w.readOnlyValue = 2
	w.oldPosition = [2]int32{16777217, -16777217}
	w.u32Inputs = []uint32{0}
	w.u8Inputs = nil
	object.field34 = 77
	object.flags = 0
	if got := objectReadOldVer4F4170(object, 3, 9, w.deps()); got != 1 {
		t.Fatalf("result = %d, want 1", got)
	}
	if object.field34 != 77 {
		t.Fatalf("field34 = %d, exact-one entry reset must not run", object.field34)
	}
	if object.x != float32(int32(16777217)) || object.y != float32(int32(-16777217)) || object.newX != object.x || object.newY != object.y {
		t.Fatalf("old position = (%v,%v), new = (%v,%v)", object.x, object.y, object.newX, object.newY)
	}
	if !reflect.DeepEqual(w.oldPosArgs, [][2]int32{{0, 0}}) {
		t.Fatalf("old position initial values = %v", w.oldPosArgs)
	}
	for _, event := range w.events {
		if event == "set-on:object" || event == "set-off:object" || event == "store-field34:object" {
			t.Fatalf("exact-one branch unexpectedly ran: %s; events=%v", event, w.events)
		}
	}
}

func TestObjectReadOldVer4F4170IDAllocationFailureStopsImmediately(t *testing.T) {
	w, object := newObjectReadOldReadWorld4F4170()
	w.allocateID = 0
	w.u32Inputs = []uint32{0}
	w.u8Inputs = []uint8{255}
	object.flags = 0
	if got := objectReadOldVer4F4170(object, 0, 10, w.deps()); got != 0 {
		t.Fatalf("result = %d, want 0", got)
	}
	if object.id != 0 || !reflect.DeepEqual(w.allocated, []uint16{256}) {
		t.Fatalf("ID state = pointer:%d allocations:%v", object.id, w.allocated)
	}
	if len(w.idTransfers) != 0 || len(w.terminations) != 0 {
		t.Fatalf("transfer continued after allocation failure: transfers=%v terminations=%v", w.idTransfers, w.terminations)
	}
	if got := w.events[len(w.events)-1]; got != "store-id:object" {
		t.Fatalf("last event = %q, want store-id:object", got)
	}
}

func TestObjectReadOldVer4F4170InventoryCountWrapsAtUint8(t *testing.T) {
	w := &objectReadOldTestWorld4F4170{
		readMode:     false,
		idLengths:    make(map[int]uintptr),
		gameResults:  make(map[uint32]int32),
		allowedTypes: make(map[uint16]int32),
		after:        make(map[string]func()),
	}
	object := &objectReadOldTestObject4F4170{name: "object"}
	items := make([]objectReadOldTestObject4F4170, 256)
	for i := range items {
		items[i].name = fmt.Sprintf("item-%d", i)
		if i+1 < len(items) {
			items[i].inventoryNext = &items[i+1]
		}
	}
	object.inventoryHead = &items[0]
	if got := objectReadOldVer4F4170(object, 0, 30, w.deps()); got != 1 {
		t.Fatalf("result = %d, want 1", got)
	}
	if !reflect.DeepEqual(w.u8Args, []uint8{0, 0}) {
		t.Fatalf("uint8 transfers = %v, want empty ID and wrapped inventory count", w.u8Args)
	}
}

func TestObjectReadOldVer4F4170OwnedWriteFiltersAndCountWidth(t *testing.T) {
	for _, test := range []struct {
		name          string
		objectVersion int32
		wantU32       []uint32
		wantU16       []uint16
	}{
		{name: "legacy uint32 count", objectVersion: 4, wantU32: []uint32{0, 1, 0x04}, wantU16: nil},
		{name: "uint16 count", objectVersion: 5, wantU32: []uint32{0, 0x04}, wantU16: []uint16{1}},
	} {
		t.Run(test.name, func(t *testing.T) {
			eligible := &objectReadOldTestObject4F4170{name: "eligible", typeInd: 0xff01, scriptID: -1}
			skipped := &objectReadOldTestObject4F4170{name: "skipped", flags: 0x20, typeInd: 0xff01}
			disallowed := &objectReadOldTestObject4F4170{name: "disallowed", typeInd: 0x8002}
			eligible.field128 = skipped
			skipped.field128 = disallowed
			object := &objectReadOldTestObject4F4170{name: "object", field129: eligible, scriptID: 9, field5: 0x84}
			w := &objectReadOldTestWorld4F4170{
				readMode:     false,
				idLengths:    make(map[int]uintptr),
				gameResults:  make(map[uint32]int32),
				allowedTypes: map[uint16]int32{0xff01: 1, 0x8002: 0},
				after:        make(map[string]func()),
			}
			if got := objectReadOldVer4F4170(object, test.objectVersion, 40, w.deps()); got != 1 {
				t.Fatalf("result = %d, want 1", got)
			}
			if !reflect.DeepEqual(w.u32Args, test.wantU32) || !reflect.DeepEqual(w.u16Args, test.wantU16) {
				t.Fatalf("count/status widths: u32=%v u16=%v, want u32=%v u16=%v", w.u32Args, w.u16Args, test.wantU32, test.wantU16)
			}
			if !reflect.DeepEqual(w.ownedWrites, []string{"eligible"}) {
				t.Fatalf("owned writes = %v, want eligible only", w.ownedWrites)
			}
		})
	}
}

func TestObjectReadOldVer4F4170LiveReloadsControlPendingOwner(t *testing.T) {
	w, object := newObjectReadOldReadWorld4F4170()
	w.u32Inputs = []uint32{0, 0}
	w.u16Inputs = []uint16{1}
	w.ownedIDInputs = []int32{456}
	w.u8Inputs = []uint8{0, 0}
	w.gameResults[objectReadOldGameFlag22_4F4170] = 1
	object.scriptID = 10
	w.scriptIDInput = 10
	w.after["rw-i32"] = func() {
		object.scriptID = 99
		w.gameResults[objectReadOldGameFlag22_4F4170] = 0
	}
	w.after["game-flags:00200000"] = func() {
		w.gameResults[objectReadOldGameFlag23_4F4170] = 0
	}
	if got := objectReadOldVer4F4170(object, 5, 40, w.deps()); got != 1 {
		t.Fatalf("result = %d, want 1", got)
	}
	if !reflect.DeepEqual(w.pending, [][2]int32{{99, 456}}) {
		t.Fatalf("pending owns = %v, want live owner ScriptID", w.pending)
	}
}

func TestObjectReadOldVer4F4170GameFlagShortCircuit(t *testing.T) {
	w, object := newObjectReadOldReadWorld4F4170()
	w.u32Inputs = []uint32{0, 1}
	w.u16Inputs = nil
	w.ownedIDInputs = []int32{1}
	w.u8Inputs = []uint8{0, 0}
	w.scriptIDInput = 7
	w.gameResults[objectReadOldGameFlag22_4F4170] = 1
	if got := objectReadOldVer4F4170(object, 2, 40, w.deps()); got != 1 {
		t.Fatalf("result = %d, want 1", got)
	}
	if len(w.pending) != 0 {
		t.Fatalf("pending owns = %v, want none", w.pending)
	}
	want22 := 0
	want23 := 0
	for _, event := range w.events {
		switch event {
		case "game-flags:00200000":
			want22++
		case "game-flags:00400000":
			want23++
		}
	}
	if want22 != 1 || want23 != 0 {
		t.Fatalf("mode gate calls = flag22:%d flag23:%d, want 1/0; events=%v", want22, want23, w.events)
	}
}
