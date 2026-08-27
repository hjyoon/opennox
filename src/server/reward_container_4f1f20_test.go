package server

import (
	"fmt"
	"math"
	"reflect"
	"testing"
	"unsafe"

	"github.com/opennox/libs/types"
)

type rewardContainerTestData4F1F20 struct {
	field216 uint8
}

type rewardContainerTestObject4F1F20 struct {
	name      string
	typeInd   uint16
	init      unsafe.Pointer
	data      *rewardContainerTestData4F1F20
	next      *rewardContainerTestObject4F1F20
	firstItem *rewardContainerTestObject4F1F20
	nextItem  *rewardContainerTestObject4F1F20
	class     uint32
	subclass  uint8
	position  types.Pointf
}

type rewardContainerTestState4F1F20 struct {
	stage       uint32
	markerCache uint32
	plusCache   uint32
	lookups     map[string]uint32
	first       *rewardContainerTestObject4F1F20
	chestInit   unsafe.Pointer
	activation  map[*rewardContainerTestObject4F1F20]*rewardContainerTestObject4F1F20
	newObjects  []*rewardContainerTestObject4F1F20
	events      []string
	deleted     []*rewardContainerTestObject4F1F20
	onActivate  func(*rewardContainerTestObject4F1F20, uint32)
	onDetach    func(*rewardContainerTestObject4F1F20, *rewardContainerTestObject4F1F20)
	onDelete    func(*rewardContainerTestObject4F1F20)
	onRandom    func(float32, *rewardContainerTestObject4F1F20, *types.Pointf)
}

func rewardContainerTestHooks4F1F20(state *rewardContainerTestState4F1F20) rewardContainerHooks4F1F20[
	*rewardContainerTestObject4F1F20,
	*rewardContainerTestData4F1F20,
] {
	return rewardContainerHooks4F1F20[*rewardContainerTestObject4F1F20, *rewardContainerTestData4F1F20]{
		loadQuestStage: func() uint32 {
			state.events = append(state.events, "stage")
			return state.stage
		},
		loadMarkerCache: func() uint32 {
			state.events = append(state.events, "cache:marker")
			return state.markerCache
		},
		storeMarkerCache: func(value uint32) {
			state.events = append(state.events, fmt.Sprintf("store:marker:%d", value))
			state.markerCache = value
		},
		loadPlusCache: func() uint32 {
			state.events = append(state.events, "cache:plus")
			return state.plusCache
		},
		storePlusCache: func(value uint32) {
			state.events = append(state.events, fmt.Sprintf("store:plus:%d", value))
			state.plusCache = value
		},
		lookupType: func(name string) uint32 {
			state.events = append(state.events, "lookup:"+name)
			return state.lookups[name]
		},
		preprocessMarkers: func() {
			state.events = append(state.events, "preprocess:markers")
		},
		preprocessRewards: func() {
			state.events = append(state.events, "preprocess:rewards")
		},
		firstObject: func() *rewardContainerTestObject4F1F20 {
			state.events = append(state.events, "first")
			return state.first
		},
		nextObject: func(object *rewardContainerTestObject4F1F20) *rewardContainerTestObject4F1F20 {
			state.events = append(state.events, "next:"+object.name)
			return object.next
		},
		loadTypeInd: func(object *rewardContainerTestObject4F1F20) uint16 {
			state.events = append(state.events, "type:"+object.name)
			return object.typeInd
		},
		loadInit: func(object *rewardContainerTestObject4F1F20) unsafe.Pointer {
			state.events = append(state.events, "init:"+object.name)
			return object.init
		},
		isChestInit: func(init unsafe.Pointer) bool {
			state.events = append(state.events, "is-chest")
			return init == state.chestInit
		},
		loadInitData: func(object *rewardContainerTestObject4F1F20) *rewardContainerTestData4F1F20 {
			state.events = append(state.events, "data:"+object.name)
			return object.data
		},
		loadField216Low: func(data *rewardContainerTestData4F1F20) uint8 {
			state.events = append(state.events, "field216")
			return data.field216
		},
		firstInventory: func(object *rewardContainerTestObject4F1F20) *rewardContainerTestObject4F1F20 {
			state.events = append(state.events, "first-item:"+object.name)
			return object.firstItem
		},
		nextInventoryItem: func(object *rewardContainerTestObject4F1F20) *rewardContainerTestObject4F1F20 {
			state.events = append(state.events, "next-item:"+object.name)
			return object.nextItem
		},
		activateMarker: func(object *rewardContainerTestObject4F1F20, stage uint32) *rewardContainerTestObject4F1F20 {
			state.events = append(state.events, fmt.Sprintf("activate:%s:%d", object.name, stage))
			if state.onActivate != nil {
				state.onActivate(object, stage)
			}
			return state.activation[object]
		},
		detachInventory: func(owner, item *rewardContainerTestObject4F1F20) {
			state.events = append(state.events, "detach:"+owner.name+":"+item.name)
			if state.onDetach != nil {
				state.onDetach(owner, item)
			}
		},
		delayedDelete: func(object *rewardContainerTestObject4F1F20) {
			state.events = append(state.events, "delete:"+object.name)
			state.deleted = append(state.deleted, object)
			if state.onDelete != nil {
				state.onDelete(object)
			}
		},
		inventoryPut: func(owner, item *rewardContainerTestObject4F1F20, mode uint32) {
			state.events = append(state.events, fmt.Sprintf("put:%s:%s:%d", owner.name, item.name, mode))
		},
		loadClass: func(object *rewardContainerTestObject4F1F20) uint32 {
			state.events = append(state.events, "class:"+object.name)
			return object.class
		},
		loadSubclassLow: func(object *rewardContainerTestObject4F1F20) uint8 {
			state.events = append(state.events, "subclass:"+object.name)
			return object.subclass
		},
		newObject: func(name string) *rewardContainerTestObject4F1F20 {
			state.events = append(state.events, "new:"+name)
			if len(state.newObjects) == 0 {
				return nil
			}
			object := state.newObjects[0]
			state.newObjects = state.newObjects[1:]
			return object
		},
		loadPosX: func(object *rewardContainerTestObject4F1F20) float32 {
			state.events = append(state.events, "x:"+object.name)
			return object.position.X
		},
		loadPosY: func(object *rewardContainerTestObject4F1F20) float32 {
			state.events = append(state.events, "y:"+object.name)
			return object.position.Y
		},
		createAt: func(object, owner *rewardContainerTestObject4F1F20, point types.Pointf) {
			ownerName := "nil"
			if owner != nil {
				ownerName = owner.name
			}
			state.events = append(state.events, fmt.Sprintf(
				"create:%s:%s:%g:%g", object.name, ownerName, point.X, point.Y,
			))
			object.position = point
		},
		randomReachable: func(radius float32, center *rewardContainerTestObject4F1F20, output *types.Pointf) {
			state.events = append(state.events, fmt.Sprintf(
				"random:%g:%s:%g:%g", radius, center.name, output.X, output.Y,
			))
			if state.onRandom != nil {
				state.onRandom(radius, center, output)
			}
		},
	}
}

func TestRewardContainerCacheInitializationAndPreprocessOrder4F1F20(t *testing.T) {
	state := &rewardContainerTestState4F1F20{
		lookups: map[string]uint32{
			rewardContainerMarkerTypeName4F1F20:     17,
			rewardContainerMarkerPlusTypeName4F1F20: 19,
		},
		activation: make(map[*rewardContainerTestObject4F1F20]*rewardContainerTestObject4F1F20),
	}
	rewardContainerProcess4F1F20(rewardContainerTestHooks4F1F20(state))
	want := []string{
		"stage", "cache:marker", "lookup:RewardMarker", "store:marker:17",
		"lookup:RewardMarkerPlus", "store:plus:19", "preprocess:markers",
		"preprocess:rewards", "first",
	}
	if state.markerCache != 17 || state.plusCache != 19 || !reflect.DeepEqual(state.events, want) {
		t.Fatalf("marker/plus/events = %d/%d/%v, want 17/19/%v", state.markerCache, state.plusCache, state.events, want)
	}

	state.events = nil
	state.lookups = map[string]uint32{}
	state.markerCache = 17
	state.plusCache = 0
	rewardContainerProcess4F1F20(rewardContainerTestHooks4F1F20(state))
	want = []string{"stage", "cache:marker", "preprocess:markers", "preprocess:rewards", "first"}
	if state.plusCache != 0 || !reflect.DeepEqual(state.events, want) {
		t.Fatalf("cached entry plus/events = %d/%v, want 0/%v", state.plusCache, state.events, want)
	}
}

func TestRewardContainerWorldMarkerQuiverOrder4F1F20(t *testing.T) {
	marker := &rewardContainerTestObject4F1F20{
		name: "marker", typeInd: 10, data: &rewardContainerTestData4F1F20{field216: 0x80},
		position: types.Pointf{X: 7, Y: 9},
	}
	reward := &rewardContainerTestObject4F1F20{
		name: "reward", class: rewardContainerWeaponClass4F1F20,
		subclass: rewardContainerBowSubclassMask4F1F20,
		position: types.Pointf{X: -1, Y: -2},
	}
	quiver := &rewardContainerTestObject4F1F20{name: "quiver"}
	state := &rewardContainerTestState4F1F20{
		stage: 23, markerCache: 10, plusCache: 11, first: marker,
		activation: map[*rewardContainerTestObject4F1F20]*rewardContainerTestObject4F1F20{marker: reward},
		newObjects: []*rewardContainerTestObject4F1F20{quiver},
	}
	state.onRandom = func(radius float32, center *rewardContainerTestObject4F1F20, output *types.Pointf) {
		if radius != 30 || center != reward || center.position != (types.Pointf{X: 7, Y: 9}) ||
			*output != (types.Pointf{X: 7, Y: 9}) {
			t.Fatalf("random input = %g/%p/%+v/%+v", radius, center, center.position, *output)
		}
		*output = types.Pointf{X: 13, Y: 14}
	}
	rewardContainerProcess4F1F20(rewardContainerTestHooks4F1F20(state))
	want := []string{
		"stage", "cache:marker", "preprocess:markers", "preprocess:rewards", "first",
		"next:marker", "cache:marker", "type:marker", "data:marker", "field216",
		"activate:marker:23", "y:marker", "x:marker", "create:reward:nil:7:9",
		"class:reward", "subclass:reward", "new:Quiver", "x:reward", "y:reward",
		"random:30:reward:7:9", "create:quiver:nil:13:14", "delete:marker",
	}
	if !reflect.DeepEqual(state.events, want) {
		t.Fatalf("events =\n%v\nwant\n%v", state.events, want)
	}
	if reward.position != (types.Pointf{X: 7, Y: 9}) || quiver.position != (types.Pointf{X: 13, Y: 14}) {
		t.Fatalf("reward/quiver positions = %+v/%+v", reward.position, quiver.position)
	}
}

func TestRewardContainerCachesNextWorldObjectAndReloadsTypes4F1F20(t *testing.T) {
	first := &rewardContainerTestObject4F1F20{
		name: "first", typeInd: 10, data: &rewardContainerTestData4F1F20{},
	}
	second := &rewardContainerTestObject4F1F20{
		name: "second", typeInd: 20, data: &rewardContainerTestData4F1F20{},
	}
	first.next = second
	state := &rewardContainerTestState4F1F20{
		markerCache: 10, plusCache: 20, first: first,
		activation: make(map[*rewardContainerTestObject4F1F20]*rewardContainerTestObject4F1F20),
	}
	state.onDelete = func(object *rewardContainerTestObject4F1F20) {
		if object == first {
			first.next = nil
			state.markerCache = 99
		}
	}
	rewardContainerProcess4F1F20(rewardContainerTestHooks4F1F20(state))
	if !reflect.DeepEqual(state.deleted, []*rewardContainerTestObject4F1F20{first, second}) {
		t.Fatalf("deleted = %v, want first and cached-next second", state.deleted)
	}
	wantSuffix := []string{
		"next:first", "cache:marker", "type:first", "data:first", "field216", "delete:first",
		"next:second", "cache:marker", "type:second", "cache:plus", "data:second", "field216", "delete:second",
	}
	if got := state.events[len(state.events)-len(wantSuffix):]; !reflect.DeepEqual(got, wantSuffix) {
		t.Fatalf("traversal suffix = %v, want %v", got, wantSuffix)
	}
}

func TestRewardContainerChestMutationAndWrappedStage4F1F20(t *testing.T) {
	chestIdentity := new(byte)
	chest := &rewardContainerTestObject4F1F20{name: "chest", typeInd: 1, init: unsafe.Pointer(chestIdentity)}
	firstMarker := &rewardContainerTestObject4F1F20{name: "marker-a", typeInd: 10}
	secondMarker := &rewardContainerTestObject4F1F20{name: "marker-b", typeInd: 20}
	firstMarker.nextItem = secondMarker
	chest.firstItem = firstMarker
	reward := &rewardContainerTestObject4F1F20{
		name: "reward", class: rewardContainerWeaponClass4F1F20,
		subclass: rewardContainerBowSubclassMask4F1F20,
	}
	quiver := &rewardContainerTestObject4F1F20{name: "quiver"}
	state := &rewardContainerTestState4F1F20{
		stage: math.MaxUint32, markerCache: 10, plusCache: 20,
		first: chest, chestInit: unsafe.Pointer(chestIdentity),
		activation: map[*rewardContainerTestObject4F1F20]*rewardContainerTestObject4F1F20{
			firstMarker:  reward,
			secondMarker: nil,
		},
		newObjects: []*rewardContainerTestObject4F1F20{quiver},
	}
	state.onActivate = func(object *rewardContainerTestObject4F1F20, stage uint32) {
		if stage != 0 {
			t.Fatalf("activation stage = %d, want wrapped zero", stage)
		}
		if object == firstMarker {
			firstMarker.nextItem = nil
			state.markerCache = 99
		}
	}
	rewardContainerProcess4F1F20(rewardContainerTestHooks4F1F20(state))
	wantSuffix := []string{
		"next:chest", "cache:marker", "type:chest", "cache:plus", "init:chest", "is-chest", "first-item:chest",
		"next-item:marker-a", "cache:marker", "type:marker-a", "activate:marker-a:0",
		"detach:chest:marker-a", "delete:marker-a", "put:chest:reward:0", "class:reward",
		"subclass:reward", "new:Quiver", "put:chest:quiver:0",
		"next-item:marker-b", "cache:marker", "type:marker-b", "cache:plus", "activate:marker-b:0",
		"detach:chest:marker-b", "delete:marker-b",
	}
	if got := state.events[len(state.events)-len(wantSuffix):]; !reflect.DeepEqual(got, wantSuffix) {
		t.Fatalf("chest suffix =\n%v\nwant\n%v", got, wantSuffix)
	}
	if !reflect.DeepEqual(state.deleted, []*rewardContainerTestObject4F1F20{firstMarker, secondMarker}) {
		t.Fatalf("deleted = %v, want marker items only", state.deleted)
	}
}

func TestRewardContainerComparesZeroExtendedTypeAgainstFullCaches4F1F20(t *testing.T) {
	notChest := new(byte)
	object := &rewardContainerTestObject4F1F20{
		name: "ordinary", typeInd: 10, init: unsafe.Pointer(notChest),
	}
	state := &rewardContainerTestState4F1F20{
		markerCache: 0x0001000a,
		plusCache:   0x0002000a,
		first:       object,
		chestInit:   unsafe.Pointer(new(byte)),
		activation:  make(map[*rewardContainerTestObject4F1F20]*rewardContainerTestObject4F1F20),
	}
	rewardContainerProcess4F1F20(rewardContainerTestHooks4F1F20(state))
	wantSuffix := []string{"next:ordinary", "cache:marker", "type:ordinary", "cache:plus", "init:ordinary", "is-chest"}
	if got := state.events[len(state.events)-len(wantSuffix):]; !reflect.DeepEqual(got, wantSuffix) {
		t.Fatalf("events = %v, want suffix %v", state.events, wantSuffix)
	}
	if len(state.deleted) != 0 {
		t.Fatalf("zero-extended TypeInd falsely matched full cache: deleted=%v", state.deleted)
	}
}

func TestRewardContainerNilInitDataFaultPrefix4F1F20(t *testing.T) {
	marker := &rewardContainerTestObject4F1F20{name: "marker", typeInd: 10}
	state := &rewardContainerTestState4F1F20{
		markerCache: 10,
		plusCache:   20,
		first:       marker,
		activation:  make(map[*rewardContainerTestObject4F1F20]*rewardContainerTestObject4F1F20),
	}
	defer func() {
		if recover() == nil {
			t.Fatal("nil RewardMarker InitData did not fault at Field216 load")
		}
		wantSuffix := []string{"next:marker", "cache:marker", "type:marker", "data:marker", "field216"}
		if got := state.events[len(state.events)-len(wantSuffix):]; !reflect.DeepEqual(got, wantSuffix) {
			t.Fatalf("fault prefix = %v, want %v", got, wantSuffix)
		}
		if len(state.deleted) != 0 {
			t.Fatalf("marker was deleted after the original fault point: %v", state.deleted)
		}
	}()
	rewardContainerProcess4F1F20(rewardContainerTestHooks4F1F20(state))
}
