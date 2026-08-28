package legacy

import (
	"fmt"
	"math"
	"reflect"
	"runtime"
	"testing"
	"unsafe"

	"github.com/opennox/opennox/v1/common/ntype"
	"github.com/opennox/opennox/v1/server"
)

type mapLoadPlaceTestObject4F3F50 struct {
	name     string
	typeInd  uint16
	x        float32
	y        float32
	first    *mapLoadPlaceTestObject4F3F50
	nextItem *mapLoadPlaceTestObject4F3F50
}

type mapLoadPlaceTestWall4F3F50 struct {
	width  uint32
	height uint32
}

func mapLoadPlaceTestName4F3F50(object *mapLoadPlaceTestObject4F3F50) string {
	if object == nil {
		return "nil"
	}
	return object.name
}

func TestMapLoadPlaceObject4F3F50EarlyRejectCleanupDoesNotClearHead(t *testing.T) {
	item2 := &mapLoadPlaceTestObject4F3F50{name: "item2"}
	item1 := &mapLoadPlaceTestObject4F3F50{name: "item1", nextItem: item2}
	rogue := &mapLoadPlaceTestObject4F3F50{name: "rogue"}
	object := &mapLoadPlaceTestObject4F3F50{name: "object", typeInd: 7, first: item1}
	var events []string

	result := mapLoadPlaceObject4F3F50(object, (*mapLoadPlaceTestObject4F3F50)(nil), false,
		mapLoadPlaceDeps4F3F50[*mapLoadPlaceTestObject4F3F50, struct{}]{
			gameFlag23: func() int32 {
				events = append(events, "flag23")
				return 0
			},
			loadTypeInd: func(got *mapLoadPlaceTestObject4F3F50) uint16 {
				events = append(events, "type:"+got.name)
				return got.typeInd
			},
			typeAllowed: func(typeInd uint16) int32 {
				events = append(events, fmt.Sprintf("allowed:%d", typeInd))
				return 0
			},
			loadFirstItem: func(got *mapLoadPlaceTestObject4F3F50) *mapLoadPlaceTestObject4F3F50 {
				events = append(events, "first:"+got.name)
				return got.first
			},
			loadNextItem: func(item *mapLoadPlaceTestObject4F3F50) *mapLoadPlaceTestObject4F3F50 {
				events = append(events, "next:"+item.name)
				return item.nextItem
			},
			freeObject: func(got *mapLoadPlaceTestObject4F3F50) {
				events = append(events, "free:"+got.name)
				if got == item1 {
					got.nextItem = rogue
				}
			},
			storeFirstItem: func(*mapLoadPlaceTestObject4F3F50, *mapLoadPlaceTestObject4F3F50) {
				t.Fatal("early reject cleared the inventory head")
			},
		})
	if result != 0 {
		t.Fatalf("result = %d, want 0", result)
	}
	want := []string{
		"flag23", "type:object", "allowed:7", "first:object",
		"next:item1", "free:item1", "next:item2", "free:item2", "free:object",
	}
	if !reflect.DeepEqual(events, want) {
		t.Fatalf("events = %v, want %v", events, want)
	}
	if object.first != item1 {
		t.Fatalf("early cleanup head = %p, want original %p", object.first, item1)
	}
}

func TestMapLoadPlaceObject4F3F50TranslationWrapAndSecondFlagStages(t *testing.T) {
	object := &mapLoadPlaceTestObject4F3F50{name: "object", x: 100.5, y: -5.25}
	wall := &mapLoadPlaceTestWall4F3F50{width: 3, height: 4}
	var events []string
	flagCalls := 0

	result := mapLoadPlaceObject4F3F50(object, (*mapLoadPlaceTestObject4F3F50)(nil), true,
		mapLoadPlaceDeps4F3F50[*mapLoadPlaceTestObject4F3F50, *mapLoadPlaceTestWall4F3F50]{
			gameFlag23: func() int32 {
				flagCalls++
				events = append(events, fmt.Sprintf("flag23:%d", flagCalls))
				if flagCalls == 1 {
					return math.MinInt32
				}
				return -1
			},
			wallSize: func() *mapLoadPlaceTestWall4F3F50 {
				events = append(events, "wall")
				return wall
			},
			loadWallWidth: func(got *mapLoadPlaceTestWall4F3F50) uint32 {
				events = append(events, "width")
				return got.width
			},
			loadWallHeight: func(got *mapLoadPlaceTestWall4F3F50) uint32 {
				events = append(events, "height")
				return got.height
			},
			loadX: func(got *mapLoadPlaceTestObject4F3F50) float32 {
				events = append(events, "x")
				return got.x
			},
			loadY: func(got *mapLoadPlaceTestObject4F3F50) float32 {
				events = append(events, "y")
				return got.y
			},
			loadTranslationX: func() int32 {
				events = append(events, "translation-x")
				return -7
			},
			loadTranslationY: func() int32 {
				events = append(events, "translation-y")
				return 30
			},
			storeX: func(got *mapLoadPlaceTestObject4F3F50, value float32) {
				events = append(events, "store-x")
				got.x = value
				wall.height = math.MaxUint32
			},
			storeY: func(got *mapLoadPlaceTestObject4F3F50, value float32) {
				events = append(events, "store-y")
				got.y = value
			},
			stageObject: func(got *mapLoadPlaceTestObject4F3F50) {
				events = append(events, "stage:"+got.name)
			},
		})
	if result != 1 {
		t.Fatalf("result = %d, want 1", result)
	}
	want := []string{
		"flag23:1", "wall", "width", "x", "translation-x", "store-x",
		"height", "y", "translation-y", "store-y", "flag23:2", "stage:object",
	}
	if !reflect.DeepEqual(events, want) {
		t.Fatalf("events = %v, want %v", events, want)
	}
	if got := math.Float32bits(object.x); got != math.Float32bits(13.5) {
		t.Fatalf("translated X bits = %#08x, want %#08x", got, math.Float32bits(13.5))
	}
	// 23*0xffffffff wraps to 0xffffffe9, which FILD reads as signed -23.
	if got := math.Float32bits(object.y); got != math.Float32bits(36.75) {
		t.Fatalf("translated Y bits = %#08x, want %#08x", got, math.Float32bits(36.75))
	}
}

func TestMapLoadPlaceObject4F3F50GameFlag22ExactOneAndLiveCreateReads(t *testing.T) {
	tests := []struct {
		name       string
		flag22     int32
		wantEvents []string
	}{
		{
			name:   "exact one bypasses placement predicate",
			flag22: 1,
			wantEvents: []string{
				"flag23:1", "type:7", "allowed:7", "flag23:2", "flag22:1",
				"y", "x", "create:object:owner:00000000",
			},
		},
		{
			name:   "other nonzero value calls placement predicate",
			flag22: 2,
			wantEvents: []string{
				"flag23:1", "type:7", "allowed:7", "flag23:2", "flag22:2",
				"type:9", "place:9", "y", "x", "create:object:owner:00000000",
			},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			object := &mapLoadPlaceTestObject4F3F50{name: "object", typeInd: 7, x: 12.25, y: -31.5}
			owner := &mapLoadPlaceTestObject4F3F50{name: "owner"}
			var events []string
			flag23Calls := 0
			result := mapLoadPlaceObject4F3F50(object, owner, false,
				mapLoadPlaceDeps4F3F50[*mapLoadPlaceTestObject4F3F50, struct{}]{
					gameFlag23: func() int32 {
						flag23Calls++
						events = append(events, fmt.Sprintf("flag23:%d", flag23Calls))
						return 0
					},
					loadTypeInd: func(got *mapLoadPlaceTestObject4F3F50) uint16 {
						events = append(events, fmt.Sprintf("type:%d", got.typeInd))
						return got.typeInd
					},
					typeAllowed: func(typeInd uint16) int32 {
						events = append(events, fmt.Sprintf("allowed:%d", typeInd))
						object.typeInd = 9
						return 1
					},
					gameFlag22: func() int32 {
						events = append(events, fmt.Sprintf("flag22:%d", test.flag22))
						return test.flag22
					},
					placeAllowed: func(typeInd uint16) int32 {
						events = append(events, fmt.Sprintf("place:%d", typeInd))
						return 1
					},
					loadY: func(got *mapLoadPlaceTestObject4F3F50) float32 {
						events = append(events, "y")
						got.x = 99.75
						return got.y
					},
					loadX: func(got *mapLoadPlaceTestObject4F3F50) float32 {
						events = append(events, "x")
						return got.x
					},
					createAt: func(got, gotOwner *mapLoadPlaceTestObject4F3F50, x, y float32, reserved int32) {
						events = append(events, fmt.Sprintf("create:%s:%s:%08x", got.name, gotOwner.name, uint32(reserved)))
						if x != 99.75 || y != -31.5 {
							t.Fatalf("create coordinates = %v,%v, want 99.75,-31.5", x, y)
						}
					},
				})
			if result != 1 {
				t.Fatalf("result = %d, want 1", result)
			}
			if !reflect.DeepEqual(events, test.wantEvents) {
				t.Fatalf("events = %v, want %v", events, test.wantEvents)
			}
		})
	}
}

func TestMapLoadPlaceObject4F3F50LateRejectClearsLiveHeadBeforeObjectFree(t *testing.T) {
	item2 := &mapLoadPlaceTestObject4F3F50{name: "item2"}
	item1 := &mapLoadPlaceTestObject4F3F50{name: "item1", nextItem: item2}
	stale := &mapLoadPlaceTestObject4F3F50{name: "stale"}
	rogue := &mapLoadPlaceTestObject4F3F50{name: "rogue"}
	object := &mapLoadPlaceTestObject4F3F50{name: "object", typeInd: 4, first: stale}
	var events []string
	flag23Calls := 0

	result := mapLoadPlaceObject4F3F50(object, (*mapLoadPlaceTestObject4F3F50)(nil), false,
		mapLoadPlaceDeps4F3F50[*mapLoadPlaceTestObject4F3F50, struct{}]{
			gameFlag23: func() int32 {
				flag23Calls++
				events = append(events, fmt.Sprintf("flag23:%d", flag23Calls))
				return 0
			},
			loadTypeInd: func(got *mapLoadPlaceTestObject4F3F50) uint16 {
				events = append(events, fmt.Sprintf("type:%d", got.typeInd))
				return got.typeInd
			},
			typeAllowed: func(typeInd uint16) int32 {
				events = append(events, fmt.Sprintf("allowed:%d", typeInd))
				object.typeInd = 6
				return 1
			},
			gameFlag22: func() int32 {
				events = append(events, "flag22:2")
				return 2
			},
			placeAllowed: func(typeInd uint16) int32 {
				events = append(events, fmt.Sprintf("place:%d", typeInd))
				object.first = item1
				return 0
			},
			loadFirstItem: func(got *mapLoadPlaceTestObject4F3F50) *mapLoadPlaceTestObject4F3F50 {
				events = append(events, "first:"+got.name)
				return got.first
			},
			loadNextItem: func(item *mapLoadPlaceTestObject4F3F50) *mapLoadPlaceTestObject4F3F50 {
				events = append(events, "next:"+item.name)
				return item.nextItem
			},
			freeObject: func(got *mapLoadPlaceTestObject4F3F50) {
				events = append(events, "free:"+got.name)
				if got == item1 {
					got.nextItem = rogue
				}
				if got == object && object.first != nil {
					t.Fatalf("object freed with non-nil inventory head %p", object.first)
				}
			},
			storeFirstItem: func(got, first *mapLoadPlaceTestObject4F3F50) {
				events = append(events, "store-first:"+mapLoadPlaceTestName4F3F50(first))
				got.first = first
			},
		})
	if result != 0 {
		t.Fatalf("result = %d, want 0", result)
	}
	want := []string{
		"flag23:1", "type:4", "allowed:4", "flag23:2", "flag22:2",
		"type:6", "place:6", "first:object", "next:item1", "free:item1",
		"next:item2", "free:item2", "store-first:nil", "free:object",
	}
	if !reflect.DeepEqual(events, want) {
		t.Fatalf("events = %v, want %v", events, want)
	}
	if object.first != nil {
		t.Fatalf("late cleanup head = %p, want nil", object.first)
	}
}

func TestMapLoadPlaceTranslatedCoordinate4F3F50WrapsProductAsSignedDword(t *testing.T) {
	tests := []struct {
		name        string
		position    float32
		wallCount   uint32
		translation int32
		wantBits    uint32
	}{
		{name: "ordinary", position: 100.5, wallCount: 3, translation: -7, wantBits: math.Float32bits(13.5)},
		{name: "negative wrapped product", position: -5.25, wallCount: math.MaxUint32, translation: 30, wantBits: math.Float32bits(36.75)},
		{name: "zero", position: 11, wallCount: 0, translation: 0, wantBits: 0},
		{name: "signed minimum product", position: 0, wallCount: 0x80000000, translation: math.MinInt32, wantBits: math.Float32bits(-11)},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			got := mapLoadPlaceTranslatedCoordinate4F3F50(test.position, test.wallCount, test.translation)
			if bits := math.Float32bits(got); bits != test.wantBits {
				t.Fatalf("result bits = %#08x (%v), want %#08x", bits, got, test.wantBits)
			}
		})
	}
}

func TestMapLoadPlaceObject4F3F50NativeLayout(t *testing.T) {
	type check struct {
		name string
		got  uintptr
		want uintptr
	}
	checks32 := []check{
		{"Object.size", unsafe.Sizeof(server.Object{}), 780},
		{"Object.TypeInd", unsafe.Offsetof(server.Object{}.TypeInd), 4},
		{"Object.PosVec", unsafe.Offsetof(server.Object{}.PosVec), 56},
		{"Object.InvNextItem", unsafe.Offsetof(server.Object{}.InvNextItem), 496},
		{"Object.InvFirstItem", unsafe.Offsetof(server.Object{}.InvFirstItem), 504},
	}
	checks64 := []check{
		{"Object.size", unsafe.Sizeof(server.Object{}), 928},
		{"Object.TypeInd", unsafe.Offsetof(server.Object{}.TypeInd), 8},
		{"Object.PosVec", unsafe.Offsetof(server.Object{}.PosVec), 60},
		{"Object.InvNextItem", unsafe.Offsetof(server.Object{}.InvNextItem), 528},
		{"Object.InvFirstItem", unsafe.Offsetof(server.Object{}.InvFirstItem), 544},
	}
	checks := checks64
	if unsafe.Sizeof(uintptr(0)) == 4 {
		checks = checks32
	}
	checks = append(checks,
		check{"Point32.size", unsafe.Sizeof(ntype.Point32{}), 8},
		check{"Point32.X", unsafe.Offsetof(ntype.Point32{}.X), 0},
		check{"Point32.Y", unsafe.Offsetof(ntype.Point32{}.Y), 4},
	)
	for _, test := range checks {
		if test.got != test.want {
			t.Errorf("%s on %s/%s = %d, want %d", test.name, runtime.GOOS, runtime.GOARCH, test.got, test.want)
		}
	}
}
