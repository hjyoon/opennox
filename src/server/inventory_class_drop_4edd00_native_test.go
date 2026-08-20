package server

import (
	"math"
	"reflect"
	"testing"
	"unsafe"

	"github.com/opennox/libs/object"
	"github.com/opennox/libs/types"
)

func TestInventoryClassDropNative4EDD00ObjectLayout(t *testing.T) {
	wantObjectSize := uintptr(780)
	wantClass := uintptr(8)
	wantPosition := uintptr(56)
	wantNext := uintptr(496)
	wantHead := uintptr(504)
	if unsafe.Sizeof(uintptr(0)) == 8 {
		wantObjectSize = 928
		wantClass = 12
		wantPosition = 60
		wantNext = 528
		wantHead = 544
	}
	checks := []struct {
		name string
		got  uintptr
		want uintptr
	}{
		{"Object size", unsafe.Sizeof(Object{}), wantObjectSize},
		{"Object.ObjClass", unsafe.Offsetof(Object{}.ObjClass), wantClass},
		{"Object.PosVec", unsafe.Offsetof(Object{}.PosVec), wantPosition},
		{"Object.InvNextItem", unsafe.Offsetof(Object{}.InvNextItem), wantNext},
		{"Object.InvFirstItem", unsafe.Offsetof(Object{}.InvFirstItem), wantHead},
		{"Pointf size", unsafe.Sizeof(types.Pointf{}), 8},
	}
	for _, check := range checks {
		if check.got != check.want {
			t.Errorf("%s = %d, want %d", check.name, check.got, check.want)
		}
	}
}

func TestRespawnInventoryClassNative4EDD00UsesNativePointersAndCachedSuccessor(t *testing.T) {
	itemB := &Object{ObjClass: object.ClassKey}
	itemA := &Object{ObjClass: object.ClassKey, InvNextItem: itemB}
	owner := &Object{
		PosVec:       types.Pointf{X: 1, Y: 2},
		InvFirstItem: itemA,
	}
	foreignPoint := &types.Pointf{X: 999, Y: 999}
	randomPoints := []types.Pointf{{X: 10, Y: 20}, {X: 30, Y: 40}}
	var events []string
	var outputs []*types.Pointf
	var creates []struct {
		item  *Object
		owner *Object
		point types.Pointf
	}
	deps := respawnInventoryClassNativeDeps4EDD00{
		detachInventory: func(gotOwner, item *Object) {
			if gotOwner != owner {
				t.Fatalf("detach owner = %p, want %p", gotOwner, owner)
			}
			switch item {
			case itemA:
				events = append(events, "detach-a")
				itemA.InvNextItem = nil
				owner.PosVec = types.Pointf{X: 3, Y: 4}
			case itemB:
				events = append(events, "detach-b")
				owner.PosVec = types.Pointf{X: 5, Y: 6}
			default:
				t.Fatalf("detach item = %p", item)
			}
		},
		randomReachable: func(radius float32, center, output *types.Pointf) *types.Pointf {
			call := len(outputs)
			if math.Float32bits(radius) != 0x42700000 || center != &owner.PosVec {
				t.Fatalf("random radius/center = %#08x/%p, want 0x42700000/%p", math.Float32bits(radius), center, &owner.PosVec)
			}
			wantCenter := types.Pointf{X: 3, Y: 4}
			if call == 1 {
				wantCenter = types.Pointf{X: 5, Y: 6}
			}
			if *center != wantCenter {
				t.Fatalf("random center = %+v, want %+v", *center, wantCenter)
			}
			outputs = append(outputs, output)
			*output = randomPoints[call]
			events = append(events, "random")
			return foreignPoint
		},
		createAt: func(item, gotOwner *Object, point types.Pointf) {
			if gotOwner != nil {
				t.Fatalf("create owner = %p, want nil", gotOwner)
			}
			creates = append(creates, struct {
				item  *Object
				owner *Object
				point types.Pointf
			}{item: item, owner: gotOwner, point: point})
			events = append(events, "create")
		},
	}

	respawnInventoryClassNative4EDD00(owner, uint32(object.ClassKey), deps)
	if want := []string{"detach-a", "random", "create", "detach-b", "random", "create"}; !reflect.DeepEqual(events, want) {
		t.Fatalf("events = %v, want %v", events, want)
	}
	if len(creates) != 2 || creates[0].item != itemA || creates[1].item != itemB ||
		creates[0].point != randomPoints[0] || creates[1].point != randomPoints[1] {
		t.Fatalf("creates = %+v", creates)
	}
	if len(outputs) != 2 || outputs[0] != outputs[1] || outputs[0] == foreignPoint {
		t.Fatalf("output identities = %p/%p, foreign = %p", outputs[0], outputs[1], foreignPoint)
	}
}

func TestDropPlayerInventoryClassNative4EDD70UsesLivePlayerNextAndLocalPoint(t *testing.T) {
	itemC := &Object{ObjClass: object.Class(0x10000001)}
	itemB := &Object{ObjClass: object.ClassKey}
	itemA := &Object{ObjClass: object.Class(0x10000000), InvNextItem: itemB}
	playerB := &Object{PosVec: types.Pointf{X: 5, Y: 6}, InvFirstItem: itemC}
	playerA := &Object{PosVec: types.Pointf{X: 1, Y: 2}, InvFirstItem: itemA}
	foreignPoint := &types.Pointf{X: 999, Y: 999}
	nextPlayers := map[*Object]*Object{playerA: playerB, playerB: nil}
	var events []string
	var outputs []*types.Pointf
	var dropPoints []*types.Pointf
	deps := dropPlayerInventoryClassNativeDeps4EDD70{
		firstPlayer: func() *Object {
			events = append(events, "first")
			return playerA
		},
		nextPlayer: func(player *Object) *Object {
			events = append(events, "next-player")
			return nextPlayers[player]
		},
		randomReachable: func(radius float32, center, output *types.Pointf) *types.Pointf {
			call := len(outputs)
			wantCenter := &playerA.PosVec
			if call == 1 {
				wantCenter = &playerB.PosVec
			}
			if math.Float32bits(radius) != 0x42480000 || center != wantCenter {
				t.Fatalf("random radius/center = %#08x/%p, want 0x42480000/%p", math.Float32bits(radius), center, wantCenter)
			}
			outputs = append(outputs, output)
			*output = types.Pointf{X: float32(10 + 20*call), Y: float32(20 + 20*call)}
			events = append(events, "random")
			return foreignPoint
		},
		drop: func(player, item *Object, point *types.Pointf) int32 {
			dropPoints = append(dropPoints, point)
			switch len(dropPoints) {
			case 1:
				if player != playerA || item != itemA || *point != (types.Pointf{X: 10, Y: 20}) {
					t.Fatalf("first drop = %p/%p/%+v", player, item, *point)
				}
				itemA.InvNextItem = nil
				nextPlayers[playerA] = playerB
			case 2:
				if player != playerB || item != itemC || *point != (types.Pointf{X: 30, Y: 40}) {
					t.Fatalf("second drop = %p/%p/%+v", player, item, *point)
				}
			default:
				t.Fatalf("unexpected drop %d", len(dropPoints))
			}
			events = append(events, "drop")
			return math.MinInt32
		},
	}

	dropPlayerInventoryClassNative4EDD70(deps)
	if want := []string{"first", "random", "drop", "next-player", "random", "drop", "next-player"}; !reflect.DeepEqual(events, want) {
		t.Fatalf("events = %v, want %v", events, want)
	}
	if len(outputs) != 2 || outputs[0] != outputs[1] || outputs[0] == foreignPoint ||
		len(dropPoints) != 2 || dropPoints[0] != outputs[0] || dropPoints[1] != outputs[1] {
		t.Fatalf("output/drop identities = %p/%p and %p/%p, foreign = %p", outputs[0], outputs[1], dropPoints[0], dropPoints[1], foreignPoint)
	}
}
