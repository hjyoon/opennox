package legacy

import (
	"fmt"
	"math"
	"reflect"
	"testing"
)

type doorXferTestUpdate4F4CB0 struct {
	lockCode         uint8
	targetDirection  int32
	syncedDirection  int32
	currentDirection int32
	tileX            int32
	tileY            int32
	fractionalDir    int16
}

type doorXferTestObject4F4CB0 struct {
	field34 uint32
	update  *doorXferTestUpdate4F4CB0
	posX    float32
	posY    float32
}

type doorXferDirection4F4CB0 struct {
	x int32
	y int32
}

type doorXferMapCall4F4CB0 struct {
	object  *doorXferTestObject4F4CB0
	version int32
}

type doorXferInventoryCall4F4CB0 struct {
	version uint16
	object  *doorXferTestObject4F4CB0
	count   int32
}

type doorXferAttachCall4F4CB0 struct {
	object *doorXferTestObject4F4CB0
	tileX  int32
	tileY  int32
}

type doorXferTestWorld4F4CB0 struct {
	version         uint16
	mapResult       int32
	readOnlyDefault int32
	readOnlyValues  []int32
	directionResult int32
	lockResult      int32
	targetResult    int32
	inventoryResult int32
	directions      map[int32]doorXferDirection4F4CB0

	field34Loads       int
	updateLoads        int
	versionTransfers   []uint16
	mapCalls           []doorXferMapCall4F4CB0
	readOnlyCalls      int
	currentLoads       []*doorXferTestUpdate4F4CB0
	lockLoads          []*doorXferTestUpdate4F4CB0
	targetLoads        []*doorXferTestUpdate4F4CB0
	directionTransfers []int32
	lockTransfers      []int32
	targetTransfers    []int32
	directionXLoads    []int32
	directionYLoads    []int32
	positionXLoads     int
	positionYLoads     int
	truncValues        []float64
	attachCalls        []doorXferAttachCall4F4CB0
	inventoryCalls     []doorXferInventoryCall4F4CB0
	field34Stores      int
	currentStores      int
	fractionalStores   int
	targetStores       int
	syncedStores       int
	tileXStores        int
	tileYStores        int
	lockStores         int
	events             []string
	after              map[string]func()
	faultAt            int
}

func newDoorXferTestWorld4F4CB0() *doorXferTestWorld4F4CB0 {
	return &doorXferTestWorld4F4CB0{
		version:         doorXferCurrentVersion4F4CB0,
		mapResult:       1,
		readOnlyDefault: 1,
		directionResult: 7,
		lockResult:      0x12345678,
		targetResult:    12,
		inventoryResult: 1,
		directions: map[int32]doorXferDirection4F4CB0{
			12: {x: 32, y: 0},
		},
		after: make(map[string]func()),
	}
}

func (w *doorXferTestWorld4F4CB0) event(name string) {
	w.events = append(w.events, name)
	if callback := w.after[name]; callback != nil {
		callback()
	}
	if w.faultAt != 0 && len(w.events) == w.faultAt {
		panic(name)
	}
}

func (w *doorXferTestWorld4F4CB0) deps() doorXferDeps4F4CB0[
	*doorXferTestObject4F4CB0,
	*doorXferTestUpdate4F4CB0,
] {
	return doorXferDeps4F4CB0[
		*doorXferTestObject4F4CB0,
		*doorXferTestUpdate4F4CB0,
	]{
		loadField34: func(object *doorXferTestObject4F4CB0) uint32 {
			w.field34Loads++
			value := object.field34
			w.event(fmt.Sprintf("load-field34:%d", w.field34Loads))
			return value
		},
		loadUpdateData: func(object *doorXferTestObject4F4CB0) *doorXferTestUpdate4F4CB0 {
			w.updateLoads++
			value := object.update
			w.event("load-update")
			return value
		},
		rwVersion: func(value uint16) uint16 {
			w.versionTransfers = append(w.versionTransfers, value)
			w.event("rw-version")
			return w.version
		},
		mapReadWrite: func(object *doorXferTestObject4F4CB0, version int32) int32 {
			w.mapCalls = append(w.mapCalls, doorXferMapCall4F4CB0{object: object, version: version})
			w.event("map-read-write")
			return w.mapResult
		},
		readOnly: func() int32 {
			w.readOnlyCalls++
			value := w.readOnlyDefault
			if w.readOnlyCalls <= len(w.readOnlyValues) {
				value = w.readOnlyValues[w.readOnlyCalls-1]
			}
			w.event(fmt.Sprintf("read-only:%d", w.readOnlyCalls))
			return value
		},
		loadCurrentDirection: func(update *doorXferTestUpdate4F4CB0) int32 {
			w.currentLoads = append(w.currentLoads, update)
			value := update.currentDirection
			w.event("load-current")
			return value
		},
		loadLockCode: func(update *doorXferTestUpdate4F4CB0) uint8 {
			w.lockLoads = append(w.lockLoads, update)
			value := update.lockCode
			w.event("load-lock")
			return value
		},
		loadTargetDirection: func(update *doorXferTestUpdate4F4CB0) int32 {
			w.targetLoads = append(w.targetLoads, update)
			value := update.targetDirection
			w.event("load-target")
			return value
		},
		rwDirection: func(value int32) int32 {
			w.directionTransfers = append(w.directionTransfers, value)
			w.event("rw-direction")
			return w.directionResult
		},
		rwLockCode: func(value int32) int32 {
			w.lockTransfers = append(w.lockTransfers, value)
			w.event("rw-lock")
			return w.lockResult
		},
		rwTargetDirection: func(value int32) int32 {
			w.targetTransfers = append(w.targetTransfers, value)
			w.event("rw-target")
			return w.targetResult
		},
		storeCurrentDirection: func(update *doorXferTestUpdate4F4CB0, value int32) {
			w.currentStores++
			w.event("store-current")
			update.currentDirection = value
		},
		storeFractionalDir: func(update *doorXferTestUpdate4F4CB0, value int16) {
			w.fractionalStores++
			w.event("store-fractional")
			update.fractionalDir = value
		},
		storeTargetDirection: func(update *doorXferTestUpdate4F4CB0, value int32) {
			w.targetStores++
			w.event("store-target")
			update.targetDirection = value
		},
		storeSyncedDirection: func(update *doorXferTestUpdate4F4CB0, value int32) {
			w.syncedStores++
			w.event("store-synced")
			update.syncedDirection = value
		},
		loadDirectionX: func(direction int32) int32 {
			w.directionXLoads = append(w.directionXLoads, direction)
			value := w.directions[direction].x
			w.event("load-direction-x")
			return value
		},
		loadPositionX: func(object *doorXferTestObject4F4CB0) float32 {
			w.positionXLoads++
			value := object.posX
			w.event("load-position-x")
			return value
		},
		loadDirectionY: func(direction int32) int32 {
			w.directionYLoads = append(w.directionYLoads, direction)
			value := w.directions[direction].y
			w.event("load-direction-y")
			return value
		},
		loadPositionY: func(object *doorXferTestObject4F4CB0) float32 {
			w.positionYLoads++
			value := object.posY
			w.event("load-position-y")
			return value
		},
		truncQwordLow: func(value float64) int32 {
			w.truncValues = append(w.truncValues, value)
			w.event(fmt.Sprintf("trunc:%d", len(w.truncValues)))
			return doorXferTruncSignedQwordLow4F4CB0(value)
		},
		attachWall: func(object *doorXferTestObject4F4CB0, tileX, tileY int32) {
			w.attachCalls = append(w.attachCalls, doorXferAttachCall4F4CB0{
				object: object, tileX: tileX, tileY: tileY,
			})
			w.event("attach-wall")
		},
		storeTileX: func(update *doorXferTestUpdate4F4CB0, value int32) {
			w.tileXStores++
			w.event("store-tile-x")
			update.tileX = value
		},
		storeTileY: func(update *doorXferTestUpdate4F4CB0, value int32) {
			w.tileYStores++
			w.event("store-tile-y")
			update.tileY = value
		},
		storeLockCode: func(update *doorXferTestUpdate4F4CB0, value uint8) {
			w.lockStores++
			w.event("store-lock")
			update.lockCode = value
		},
		transferInventory: func(version uint16, object *doorXferTestObject4F4CB0, count int32) int32 {
			w.inventoryCalls = append(w.inventoryCalls, doorXferInventoryCall4F4CB0{
				version: version, object: object, count: count,
			})
			w.event("transfer-inventory")
			return w.inventoryResult
		},
		storeField34: func(object *doorXferTestObject4F4CB0, value uint32) {
			w.field34Stores++
			w.event("store-field34")
			object.field34 = value
		},
	}
}

func TestDoorXfer4F4CB0PreservesEntryCachesAndReadOrder(t *testing.T) {
	entryUpdate := &doorXferTestUpdate4F4CB0{lockCode: 0x11, targetDirection: 3, currentDirection: 4}
	liveUpdate := &doorXferTestUpdate4F4CB0{lockCode: 0x22, targetDirection: 5, currentDirection: 6}
	object := &doorXferTestObject4F4CB0{
		field34: 0x11223344,
		update:  entryUpdate,
		posX:    30.5,
		posY:    -14.25,
	}
	w := newDoorXferTestWorld4F4CB0()
	w.directionResult = -3
	w.lockResult = 0x1234abcd
	w.targetResult = 11
	w.directions[11] = doorXferDirection4F4CB0{x: 31, y: -6}
	w.after["load-field34:1"] = func() { object.field34 = 0x55667788 }
	w.after["load-update"] = func() { object.update = liveUpdate }
	w.after["store-lock"] = func() { object.field34 = 0x80000003 }
	w.after["load-field34:2"] = func() { object.field34 = 9 }

	if got := doorXfer4F4CB0(object, w.deps()); got != 1 {
		t.Fatalf("result = %d, want 1", got)
	}
	if object.field34 != 0x11223344 {
		t.Fatalf("Field34 = %#08x, want entry cache %#08x", object.field34, uint32(0x11223344))
	}
	if object.update != liveUpdate {
		t.Fatalf("live UpdateData pointer changed: got %p, want %p", object.update, liveUpdate)
	}
	if entryUpdate.currentDirection != -3 || entryUpdate.fractionalDir != -24 ||
		entryUpdate.targetDirection != 11 || entryUpdate.syncedDirection != -3 ||
		entryUpdate.tileX != 1 || entryUpdate.tileY != 0 || entryUpdate.lockCode != 0xcd {
		t.Fatalf("cached update = %+v, want direction -3/-24 target 11 tiles 1/0 lock cd", *entryUpdate)
	}
	if *liveUpdate != (doorXferTestUpdate4F4CB0{lockCode: 0x22, targetDirection: 5, currentDirection: 6}) {
		t.Fatalf("live update mutated: %+v", *liveUpdate)
	}
	if !reflect.DeepEqual(w.directionTransfers, []int32{0}) ||
		!reflect.DeepEqual(w.lockTransfers, []int32{0}) ||
		!reflect.DeepEqual(w.targetTransfers, []int32{0}) {
		t.Fatalf("read-mode transfer inputs = %v/%v/%v, want zero locals", w.directionTransfers, w.lockTransfers, w.targetTransfers)
	}
	if !reflect.DeepEqual(w.directionXLoads, []int32{11}) ||
		!reflect.DeepEqual(w.directionYLoads, []int32{11}) {
		t.Fatalf("direction indices = %v/%v, want cached target 11", w.directionXLoads, w.directionYLoads)
	}
	if !reflect.DeepEqual(w.attachCalls, []doorXferAttachCall4F4CB0{{object: object, tileX: 1, tileY: 0}}) {
		t.Fatalf("attach calls = %#v, want object tiles 1/0", w.attachCalls)
	}
	if !reflect.DeepEqual(w.inventoryCalls, []doorXferInventoryCall4F4CB0{{
		version: 60, object: object, count: -2147483645,
	}}) {
		t.Fatalf("inventory calls = %#v, want raw version and signed live count", w.inventoryCalls)
	}
	wantEvents := []string{
		"load-field34:1", "load-update", "rw-version", "map-read-write",
		"read-only:1", "rw-direction", "rw-lock", "rw-target", "read-only:2",
		"store-current", "store-fractional", "store-target", "store-synced",
		"load-direction-x", "load-position-x", "trunc:1",
		"load-direction-y", "load-position-y", "trunc:2", "attach-wall",
		"store-tile-x", "store-tile-y", "store-lock",
		"load-field34:2", "read-only:3", "transfer-inventory", "store-field34",
	}
	if !reflect.DeepEqual(w.events, wantEvents) {
		t.Fatalf("events = %v, want %v", w.events, wantEvents)
	}
}

func TestDoorXfer4F4CB0VersionAndModeBranches(t *testing.T) {
	t.Run("negative version uses write caches and old target rule", func(t *testing.T) {
		update := &doorXferTestUpdate4F4CB0{lockCode: 0xfe, targetDirection: 29, currentDirection: -17}
		object := &doorXferTestObject4F4CB0{update: update}
		w := newDoorXferTestWorld4F4CB0()
		w.version = 0xffff
		w.readOnlyDefault = 0
		w.directionResult = -17
		w.lockResult = 0xfe

		if got := doorXfer4F4CB0(object, w.deps()); got != 1 {
			t.Fatalf("result = %d, want 1", got)
		}
		if !reflect.DeepEqual(w.mapCalls, []doorXferMapCall4F4CB0{{object: object, version: -1}}) {
			t.Fatalf("map calls = %#v, want signed version -1", w.mapCalls)
		}
		if !reflect.DeepEqual(w.currentLoads, []*doorXferTestUpdate4F4CB0{update}) ||
			!reflect.DeepEqual(w.lockLoads, []*doorXferTestUpdate4F4CB0{update}) ||
			!reflect.DeepEqual(w.targetLoads, []*doorXferTestUpdate4F4CB0{update}) {
			t.Fatalf("write cache loads = %v/%v/%v, want cached update", w.currentLoads, w.lockLoads, w.targetLoads)
		}
		if !reflect.DeepEqual(w.directionTransfers, []int32{-17}) ||
			!reflect.DeepEqual(w.lockTransfers, []int32{0xfe}) || len(w.targetTransfers) != 0 {
			t.Fatalf("wire inputs = %v/%v/%v, want -17/254/no target", w.directionTransfers, w.lockTransfers, w.targetTransfers)
		}
		if w.readOnlyCalls != 2 || w.currentStores != 0 || len(w.inventoryCalls) != 0 || w.field34Stores != 1 {
			t.Fatalf("mode/stores/inventory/restore = %d/%d/%d/%d, want 2/0/0/1",
				w.readOnlyCalls, w.currentStores, len(w.inventoryCalls), w.field34Stores)
		}
	})

	t.Run("nonzero non-one mode neither caches nor applies", func(t *testing.T) {
		update := &doorXferTestUpdate4F4CB0{lockCode: 7, targetDirection: 8, currentDirection: 9}
		object := &doorXferTestObject4F4CB0{field34: 5, update: update}
		w := newDoorXferTestWorld4F4CB0()
		w.readOnlyDefault = 2

		if got := doorXfer4F4CB0(object, w.deps()); got != 1 {
			t.Fatalf("result = %d, want 1", got)
		}
		if len(w.currentLoads)+len(w.lockLoads)+len(w.targetLoads) != 0 ||
			w.currentStores+w.fractionalStores+w.targetStores+w.syncedStores != 0 {
			t.Fatalf("cache/store counts = %d/%d/%d/%d and %d/%d/%d/%d, want zero",
				len(w.currentLoads), len(w.lockLoads), len(w.targetLoads), 0,
				w.currentStores, w.fractionalStores, w.targetStores, w.syncedStores)
		}
		if !reflect.DeepEqual(w.directionTransfers, []int32{0}) ||
			!reflect.DeepEqual(w.lockTransfers, []int32{0}) ||
			!reflect.DeepEqual(w.targetTransfers, []int32{0}) {
			t.Fatalf("transfer inputs = %v/%v/%v, want zero locals", w.directionTransfers, w.lockTransfers, w.targetTransfers)
		}
		if w.readOnlyCalls != 3 || len(w.inventoryCalls) != 0 || object.field34 != 5 {
			t.Fatalf("read-only/inventory/Field34 = %d/%d/%d, want 3/0/5",
				w.readOnlyCalls, len(w.inventoryCalls), object.field34)
		}
	})

	t.Run("version forty omits target while forty-one includes it", func(t *testing.T) {
		for _, tc := range []struct {
			version     uint16
			targetCalls int
		}{
			{version: 40, targetCalls: 0},
			{version: 41, targetCalls: 1},
		} {
			t.Run(fmt.Sprintf("version-%d", tc.version), func(t *testing.T) {
				object := &doorXferTestObject4F4CB0{update: &doorXferTestUpdate4F4CB0{}}
				w := newDoorXferTestWorld4F4CB0()
				w.version = tc.version
				w.readOnlyDefault = 2
				if got := doorXfer4F4CB0(object, w.deps()); got != 1 {
					t.Fatalf("result = %d, want 1", got)
				}
				if len(w.targetTransfers) != tc.targetCalls {
					t.Fatalf("target transfers = %d, want %d", len(w.targetTransfers), tc.targetCalls)
				}
			})
		}
	})
}

func TestDoorXfer4F4CB0FailurePrefixesAndInventoryGates(t *testing.T) {
	t.Run("version greater than sixty", func(t *testing.T) {
		object := &doorXferTestObject4F4CB0{field34: 7, update: &doorXferTestUpdate4F4CB0{}}
		w := newDoorXferTestWorld4F4CB0()
		w.version = 61
		w.after["rw-version"] = func() { object.field34 = 19 }

		if got := doorXfer4F4CB0(object, w.deps()); got != 0 {
			t.Fatalf("result = %d, want 0", got)
		}
		want := []string{"load-field34:1", "load-update", "rw-version"}
		if !reflect.DeepEqual(w.events, want) || object.field34 != 19 || w.field34Stores != 0 {
			t.Fatalf("events/Field34/stores = %v/%d/%d, want %v/19/0", w.events, object.field34, w.field34Stores, want)
		}
	})

	t.Run("map serializer failure", func(t *testing.T) {
		object := &doorXferTestObject4F4CB0{field34: 11, update: &doorXferTestUpdate4F4CB0{}}
		w := newDoorXferTestWorld4F4CB0()
		w.mapResult = 0
		w.after["map-read-write"] = func() { object.field34 = 29 }

		if got := doorXfer4F4CB0(object, w.deps()); got != 0 {
			t.Fatalf("result = %d, want 0", got)
		}
		want := []string{"load-field34:1", "load-update", "rw-version", "map-read-write"}
		if !reflect.DeepEqual(w.events, want) || object.field34 != 29 || w.readOnlyCalls != 0 || w.field34Stores != 0 {
			t.Fatalf("events/Field34/mode/stores = %v/%d/%d/%d, want %v/29/0/0",
				w.events, object.field34, w.readOnlyCalls, w.field34Stores, want)
		}
	})

	t.Run("zero live count skips final mode read", func(t *testing.T) {
		object := &doorXferTestObject4F4CB0{field34: 13, update: &doorXferTestUpdate4F4CB0{}}
		w := newDoorXferTestWorld4F4CB0()
		w.after["store-lock"] = func() { object.field34 = 0 }

		if got := doorXfer4F4CB0(object, w.deps()); got != 1 {
			t.Fatalf("result = %d, want 1", got)
		}
		if w.readOnlyCalls != 2 || len(w.inventoryCalls) != 0 || object.field34 != 13 {
			t.Fatalf("read-only/inventory/Field34 = %d/%d/%d, want 2/0/13",
				w.readOnlyCalls, len(w.inventoryCalls), object.field34)
		}
	})

	t.Run("non-one final mode skips inventory", func(t *testing.T) {
		object := &doorXferTestObject4F4CB0{field34: 17, update: &doorXferTestUpdate4F4CB0{}}
		w := newDoorXferTestWorld4F4CB0()
		w.readOnlyValues = []int32{1, 1, 2}

		if got := doorXfer4F4CB0(object, w.deps()); got != 1 {
			t.Fatalf("result = %d, want 1", got)
		}
		if w.readOnlyCalls != 3 || len(w.inventoryCalls) != 0 || object.field34 != 17 {
			t.Fatalf("read-only/inventory/Field34 = %d/%d/%d, want 3/0/17",
				w.readOnlyCalls, len(w.inventoryCalls), object.field34)
		}
	})

	t.Run("inventory failure keeps live field and applied update", func(t *testing.T) {
		update := &doorXferTestUpdate4F4CB0{}
		object := &doorXferTestObject4F4CB0{field34: 23, update: update}
		w := newDoorXferTestWorld4F4CB0()
		w.inventoryResult = 0
		w.after["store-lock"] = func() { object.field34 = 5 }
		w.after["transfer-inventory"] = func() { object.field34 = 31 }

		if got := doorXfer4F4CB0(object, w.deps()); got != 0 {
			t.Fatalf("result = %d, want 0", got)
		}
		if object.field34 != 31 || w.field34Stores != 0 || update.currentDirection != w.directionResult {
			t.Fatalf("Field34/stores/current = %d/%d/%d, want 31/0/%d",
				object.field34, w.field34Stores, update.currentDirection, w.directionResult)
		}
	})
}

func TestDoorXfer4F4CB0TreatsAnyNonzeroCallbackResultAsSuccess(t *testing.T) {
	object := &doorXferTestObject4F4CB0{field34: 37, update: &doorXferTestUpdate4F4CB0{}}
	w := newDoorXferTestWorld4F4CB0()
	w.mapResult = -7
	w.inventoryResult = -9

	if got := doorXfer4F4CB0(object, w.deps()); got != 1 {
		t.Fatalf("result = %d, want canonical 1", got)
	}
	if len(w.inventoryCalls) != 1 || object.field34 != 37 {
		t.Fatalf("inventory/Field34 = %d/%d, want 1/37", len(w.inventoryCalls), object.field34)
	}
}

func TestDoorXfer4F4CB0Arithmetic(t *testing.T) {
	if got := math.Float32bits(math.Float32frombits(doorXferGridInverseBits4F4CB0)); got != 0x3d321643 {
		t.Fatalf("grid inverse bits = %#x, want 0x3d321643", got)
	}
	for _, tc := range []struct {
		direction int32
		want      int16
	}{
		{direction: 0, want: 0},
		{direction: 1, want: 8},
		{direction: -1, want: -8},
		{direction: 0x007fffff, want: -8},
		{direction: 0x00800000, want: 0},
		{direction: math.MaxInt32, want: -8},
		{direction: math.MinInt32, want: 0},
	} {
		if got := doorXferFractionalDirection4F4CB0(tc.direction); got != tc.want {
			t.Errorf("fractional(%#08x) = %d, want %d", uint32(tc.direction), got, tc.want)
		}
	}

	for _, tc := range []struct {
		name     string
		offset   int32
		position float32
		want     int32
	}{
		{name: "one tile", offset: 0, position: 23, want: 1},
		{name: "negative fraction", offset: -11, position: 0, want: 0},
		{name: "offset plus position", offset: 16, position: 7, want: 1},
		{name: "negative two tiles", offset: 0, position: -46, want: -2},
		{name: "NaN invalid", offset: 0, position: float32(math.NaN()), want: 0},
		{name: "range invalid", offset: 0, position: math.MaxFloat32, want: 0},
	} {
		t.Run(tc.name, func(t *testing.T) {
			if got := doorXferTileCoordinate4F4CB0(
				tc.offset, tc.position, doorXferTruncSignedQwordLow4F4CB0,
			); got != tc.want {
				t.Fatalf("tile(%d,%v) = %d, want %d", tc.offset, tc.position, got, tc.want)
			}
		})
	}

	for _, value := range []float64{math.NaN(), math.Inf(1), math.Inf(-1), 0x1p63, -0x1p63 - 2048} {
		if got := doorXferTruncSignedQwordLow4F4CB0(value); got != 0 {
			t.Errorf("invalid trunc(%v) = %d, want 0", value, got)
		}
	}
}

func TestDoorXfer4F4CB0FaultPrefixes(t *testing.T) {
	wantEvents := []string{
		"load-field34:1", "load-update", "rw-version", "map-read-write",
		"read-only:1", "rw-direction", "rw-lock", "rw-target", "read-only:2",
		"store-current", "store-fractional", "store-target", "store-synced",
		"load-direction-x", "load-position-x", "trunc:1",
		"load-direction-y", "load-position-y", "trunc:2", "attach-wall",
		"store-tile-x", "store-tile-y", "store-lock",
		"load-field34:2", "read-only:3", "transfer-inventory", "store-field34",
	}
	for faultAt := 1; faultAt <= len(wantEvents); faultAt++ {
		t.Run(fmt.Sprintf("fault-%d", faultAt), func(t *testing.T) {
			object := &doorXferTestObject4F4CB0{
				field34: 3,
				update:  &doorXferTestUpdate4F4CB0{},
			}
			w := newDoorXferTestWorld4F4CB0()
			w.faultAt = faultAt

			var recovered any
			func() {
				defer func() { recovered = recover() }()
				_ = doorXfer4F4CB0(object, w.deps())
			}()
			if recovered == nil {
				t.Fatal("expected injected fault")
			}
			if !reflect.DeepEqual(w.events, wantEvents[:faultAt]) {
				t.Fatalf("events = %v, want prefix %v", w.events, wantEvents[:faultAt])
			}
		})
	}
}
