package server

import (
	"fmt"
	"math"
	"reflect"
	"testing"

	"github.com/opennox/libs/types"
)

type respawnInventoryClassTestItem4EDD00 struct {
	name  string
	next  *respawnInventoryClassTestItem4EDD00
	class uint32
}

type respawnInventoryClassTestOwner4EDD00 struct {
	name string
	head *respawnInventoryClassTestItem4EDD00
	pos  types.Pointf
}

type respawnInventoryClassTestCreate4EDD00 struct {
	item  *respawnInventoryClassTestItem4EDD00
	owner *respawnInventoryClassTestOwner4EDD00
	point types.Pointf
}

type respawnInventoryClassTestWorld4EDD00 struct {
	owner *respawnInventoryClassTestOwner4EDD00
	mask  uint32

	events        []string
	faultAt       int
	randomPoints  []types.Pointf
	randomCalls   int
	randomOutputs []*types.Pointf
	randomReturn  *types.Pointf
	creates       []respawnInventoryClassTestCreate4EDD00
	afterNext     func(*respawnInventoryClassTestWorld4EDD00, *respawnInventoryClassTestItem4EDD00)
	afterDetach   func(*respawnInventoryClassTestWorld4EDD00, *respawnInventoryClassTestItem4EDD00)
	afterRandom   func(*respawnInventoryClassTestWorld4EDD00, int)
}

func respawnInventoryClassTestItemName4EDD00(item *respawnInventoryClassTestItem4EDD00) string {
	if item == nil {
		return "nil"
	}
	return item.name
}

func respawnInventoryClassTestOwnerName4EDD00(owner *respawnInventoryClassTestOwner4EDD00) string {
	if owner == nil {
		return "nil"
	}
	return owner.name
}

func respawnInventoryClassTestPoint4EDD00(point types.Pointf) string {
	return fmt.Sprintf("%08x,%08x", math.Float32bits(point.X), math.Float32bits(point.Y))
}

func (w *respawnInventoryClassTestWorld4EDD00) event(event string) {
	w.events = append(w.events, event)
	if w.faultAt != 0 && len(w.events) == w.faultAt {
		panic(event)
	}
}

func (w *respawnInventoryClassTestWorld4EDD00) hooks() respawnInventoryClassHooks4EDD00[
	*respawnInventoryClassTestOwner4EDD00,
	*respawnInventoryClassTestItem4EDD00,
	types.Pointf,
] {
	return respawnInventoryClassHooks4EDD00[
		*respawnInventoryClassTestOwner4EDD00,
		*respawnInventoryClassTestItem4EDD00,
		types.Pointf,
	]{
		loadOwnerArg: func() *respawnInventoryClassTestOwner4EDD00 {
			owner := w.owner
			w.event("owner-arg:" + respawnInventoryClassTestOwnerName4EDD00(owner))
			return owner
		},
		loadInventoryHead: func(owner *respawnInventoryClassTestOwner4EDD00) *respawnInventoryClassTestItem4EDD00 {
			head := owner.head
			w.event("head:" + respawnInventoryClassTestOwnerName4EDD00(owner) + ":" + respawnInventoryClassTestItemName4EDD00(head))
			return head
		},
		loadClassMaskArg: func() uint32 {
			mask := w.mask
			w.event(fmt.Sprintf("mask:%08x", mask))
			return mask
		},
		loadInventoryNext: func(item *respawnInventoryClassTestItem4EDD00) *respawnInventoryClassTestItem4EDD00 {
			next := item.next
			w.event("next:" + respawnInventoryClassTestItemName4EDD00(item) + ":" + respawnInventoryClassTestItemName4EDD00(next))
			if w.afterNext != nil {
				w.afterNext(w, item)
			}
			return next
		},
		loadItemClass: func(item *respawnInventoryClassTestItem4EDD00) uint32 {
			class := item.class
			w.event(fmt.Sprintf("class:%s:%08x", respawnInventoryClassTestItemName4EDD00(item), class))
			return class
		},
		detachInventory: func(owner *respawnInventoryClassTestOwner4EDD00, item *respawnInventoryClassTestItem4EDD00) {
			w.event("detach:" + respawnInventoryClassTestOwnerName4EDD00(owner) + ":" + respawnInventoryClassTestItemName4EDD00(item))
			if w.afterDetach != nil {
				w.afterDetach(w, item)
			}
		},
		ownerPosition: func(owner *respawnInventoryClassTestOwner4EDD00) *types.Pointf {
			w.event("position:" + respawnInventoryClassTestOwnerName4EDD00(owner) + ":" + respawnInventoryClassTestPoint4EDD00(owner.pos))
			return &owner.pos
		},
		randomReachable: func(radius float32, center, output *types.Pointf) *types.Pointf {
			call := w.randomCalls
			point := types.Pointf{}
			if call < len(w.randomPoints) {
				point = w.randomPoints[call]
			}
			w.randomCalls++
			w.randomOutputs = append(w.randomOutputs, output)
			w.event(fmt.Sprintf(
				"random:%d:%08x:%s:%s",
				w.randomCalls,
				math.Float32bits(radius),
				respawnInventoryClassTestPoint4EDD00(*center),
				respawnInventoryClassTestPoint4EDD00(point),
			))
			*output = point
			if w.afterRandom != nil {
				w.afterRandom(w, w.randomCalls)
			}
			return w.randomReturn
		},
		loadPointY: func(point *types.Pointf) float32 {
			y := point.Y
			w.event(fmt.Sprintf("point-y:%08x", math.Float32bits(y)))
			return y
		},
		loadPointX: func(point *types.Pointf) float32 {
			x := point.X
			w.event(fmt.Sprintf("point-x:%08x", math.Float32bits(x)))
			return x
		},
		createAt: func(item *respawnInventoryClassTestItem4EDD00, owner *respawnInventoryClassTestOwner4EDD00, x, y float32) {
			point := types.Pointf{X: x, Y: y}
			w.event("create:" + respawnInventoryClassTestItemName4EDD00(item) + ":" + respawnInventoryClassTestOwnerName4EDD00(owner) + ":" + respawnInventoryClassTestPoint4EDD00(point))
			w.creates = append(w.creates, respawnInventoryClassTestCreate4EDD00{item: item, owner: owner, point: point})
		},
	}
}

func newRespawnInventoryClassTrace4EDD00() (*respawnInventoryClassTestWorld4EDD00, []string) {
	itemC := &respawnInventoryClassTestItem4EDD00{name: "item-c", class: 0x80000040}
	itemB := &respawnInventoryClassTestItem4EDD00{name: "item-b", next: itemC, class: 0x00000008}
	itemA := &respawnInventoryClassTestItem4EDD00{name: "item-a", next: itemB, class: 0x00000040}
	owner := &respawnInventoryClassTestOwner4EDD00{
		name: "owner-a",
		head: itemA,
		pos:  types.Pointf{X: 1, Y: 2},
	}
	w := &respawnInventoryClassTestWorld4EDD00{
		owner: owner,
		mask:  0x00000040,
		randomPoints: []types.Pointf{
			{X: 10, Y: 20},
			{X: -30, Y: 40},
		},
		randomReturn: &types.Pointf{X: 999, Y: 999},
	}
	w.afterNext = func(w *respawnInventoryClassTestWorld4EDD00, item *respawnInventoryClassTestItem4EDD00) {
		if item == itemA {
			item.next = nil
		}
	}
	w.afterDetach = func(w *respawnInventoryClassTestWorld4EDD00, item *respawnInventoryClassTestItem4EDD00) {
		switch item {
		case itemA:
			w.owner.pos = types.Pointf{X: 3, Y: 4}
		case itemC:
			w.owner.pos = types.Pointf{X: 5, Y: 6}
		}
	}
	w.afterRandom = func(w *respawnInventoryClassTestWorld4EDD00, call int) {
		if call == 1 {
			w.mask = 0
		}
	}
	want := []string{
		"owner-arg:owner-a",
		"head:owner-a:item-a",
		"mask:00000040",
		"next:item-a:item-b",
		"class:item-a:00000040",
		"detach:owner-a:item-a",
		"position:owner-a:40400000,40800000",
		"random:1:42700000:40400000,40800000:41200000,41a00000",
		"point-y:41a00000",
		"point-x:41200000",
		"create:item-a:nil:41200000,41a00000",
		"next:item-b:item-c",
		"class:item-b:00000008",
		"next:item-c:nil",
		"class:item-c:80000040",
		"detach:owner-a:item-c",
		"position:owner-a:40a00000,40c00000",
		"random:2:42700000:40a00000,40c00000:c1f00000,42200000",
		"point-y:42200000",
		"point-x:c1f00000",
		"create:item-c:nil:c1f00000,42200000",
	}
	return w, want
}

func verifyRespawnInventoryClassFaultPrefixes4EDD00(
	t *testing.T,
	want []string,
	build func() *respawnInventoryClassTestWorld4EDD00,
) {
	t.Helper()
	for faultAt := 1; faultAt <= len(want); faultAt++ {
		t.Run(fmt.Sprintf("fault-%d", faultAt), func(t *testing.T) {
			w := build()
			w.faultAt = faultAt
			defer func() {
				if got := recover(); got != want[faultAt-1] {
					t.Fatalf("panic = %v, want %q", got, want[faultAt-1])
				}
				if !reflect.DeepEqual(w.events, want[:faultAt]) {
					t.Fatalf("events = %v, want %v", w.events, want[:faultAt])
				}
			}()
			respawnInventoryClass4EDD00(w.hooks())
		})
	}
}

func TestRespawnInventoryClass4EDD00ExactOrderMutationAndFaultPrefixes(t *testing.T) {
	build := func() *respawnInventoryClassTestWorld4EDD00 {
		w, _ := newRespawnInventoryClassTrace4EDD00()
		return w
	}
	w, want := newRespawnInventoryClassTrace4EDD00()
	respawnInventoryClass4EDD00(w.hooks())
	if !reflect.DeepEqual(w.events, want) {
		t.Fatalf("events = %v, want %v", w.events, want)
	}
	if len(w.creates) != 2 || w.creates[0].owner != nil || w.creates[1].owner != nil {
		t.Fatalf("creates = %+v, want two nil-owner creates", w.creates)
	}
	if len(w.randomOutputs) != 2 || w.randomOutputs[0] != w.randomOutputs[1] || w.randomOutputs[0] == w.randomReturn {
		t.Fatalf("random output identities = %p/%p, returned = %p", w.randomOutputs[0], w.randomOutputs[1], w.randomReturn)
	}
	verifyRespawnInventoryClassFaultPrefixes4EDD00(t, want, build)
}

func TestRespawnInventoryClass4EDD00EmptyInventoryDelaysMask(t *testing.T) {
	w := &respawnInventoryClassTestWorld4EDD00{
		owner: &respawnInventoryClassTestOwner4EDD00{name: "owner-empty"},
		mask:  0xffffffff,
	}
	respawnInventoryClass4EDD00(w.hooks())
	want := []string{"owner-arg:owner-empty", "head:owner-empty:nil"}
	if !reflect.DeepEqual(w.events, want) {
		t.Fatalf("events = %v, want %v", w.events, want)
	}
	verifyRespawnInventoryClassFaultPrefixes4EDD00(t, want, func() *respawnInventoryClassTestWorld4EDD00 {
		return &respawnInventoryClassTestWorld4EDD00{
			owner: &respawnInventoryClassTestOwner4EDD00{name: "owner-empty"},
			mask:  0xffffffff,
		}
	})
}

func TestRespawnInventoryClass4EDD00Constants(t *testing.T) {
	if got := math.Float32bits(respawnInventoryClassRadius4EDD00); got != 0x42700000 {
		t.Fatalf("radius bits = %#08x, want 0x42700000", got)
	}
}
