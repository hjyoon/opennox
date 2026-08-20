package server

import (
	"fmt"
	"math"
	"reflect"
	"testing"

	"github.com/opennox/libs/types"
)

type dropPlayerInventoryClassTestItem4EDD70 struct {
	name  string
	next  *dropPlayerInventoryClassTestItem4EDD70
	class uint32
}

type dropPlayerInventoryClassTestPlayer4EDD70 struct {
	name string
	next *dropPlayerInventoryClassTestPlayer4EDD70
	head *dropPlayerInventoryClassTestItem4EDD70
	pos  types.Pointf
}

type dropPlayerInventoryClassTestDrop4EDD70 struct {
	player *dropPlayerInventoryClassTestPlayer4EDD70
	item   *dropPlayerInventoryClassTestItem4EDD70
	point  types.Pointf
}

type dropPlayerInventoryClassTestWorld4EDD70 struct {
	first *dropPlayerInventoryClassTestPlayer4EDD70

	events        []string
	faultAt       int
	randomPoints  []types.Pointf
	randomCalls   int
	randomOutputs []*types.Pointf
	randomReturn  *types.Pointf
	dropResults   []int32
	drops         []dropPlayerInventoryClassTestDrop4EDD70
	afterNext     func(*dropPlayerInventoryClassTestWorld4EDD70, *dropPlayerInventoryClassTestItem4EDD70)
	afterDrop     func(*dropPlayerInventoryClassTestWorld4EDD70, *dropPlayerInventoryClassTestPlayer4EDD70, *dropPlayerInventoryClassTestItem4EDD70)
}

func dropPlayerInventoryClassTestItemName4EDD70(item *dropPlayerInventoryClassTestItem4EDD70) string {
	if item == nil {
		return "nil"
	}
	return item.name
}

func dropPlayerInventoryClassTestPlayerName4EDD70(player *dropPlayerInventoryClassTestPlayer4EDD70) string {
	if player == nil {
		return "nil"
	}
	return player.name
}

func dropPlayerInventoryClassTestPoint4EDD70(point types.Pointf) string {
	return fmt.Sprintf("%08x,%08x", math.Float32bits(point.X), math.Float32bits(point.Y))
}

func (w *dropPlayerInventoryClassTestWorld4EDD70) event(event string) {
	w.events = append(w.events, event)
	if w.faultAt != 0 && len(w.events) == w.faultAt {
		panic(event)
	}
}

func (w *dropPlayerInventoryClassTestWorld4EDD70) hooks() dropPlayerInventoryClassHooks4EDD70[
	*dropPlayerInventoryClassTestPlayer4EDD70,
	*dropPlayerInventoryClassTestItem4EDD70,
	types.Pointf,
] {
	return dropPlayerInventoryClassHooks4EDD70[
		*dropPlayerInventoryClassTestPlayer4EDD70,
		*dropPlayerInventoryClassTestItem4EDD70,
		types.Pointf,
	]{
		firstPlayer: func() *dropPlayerInventoryClassTestPlayer4EDD70 {
			first := w.first
			w.event("first-player:" + dropPlayerInventoryClassTestPlayerName4EDD70(first))
			return first
		},
		loadInventoryHead: func(player *dropPlayerInventoryClassTestPlayer4EDD70) *dropPlayerInventoryClassTestItem4EDD70 {
			head := player.head
			w.event("head:" + dropPlayerInventoryClassTestPlayerName4EDD70(player) + ":" + dropPlayerInventoryClassTestItemName4EDD70(head))
			return head
		},
		loadItemClass: func(item *dropPlayerInventoryClassTestItem4EDD70) uint32 {
			class := item.class
			w.event(fmt.Sprintf("class:%s:%08x", dropPlayerInventoryClassTestItemName4EDD70(item), class))
			return class
		},
		loadInventoryNext: func(item *dropPlayerInventoryClassTestItem4EDD70) *dropPlayerInventoryClassTestItem4EDD70 {
			next := item.next
			w.event("next:" + dropPlayerInventoryClassTestItemName4EDD70(item) + ":" + dropPlayerInventoryClassTestItemName4EDD70(next))
			if w.afterNext != nil {
				w.afterNext(w, item)
			}
			return next
		},
		playerPosition: func(player *dropPlayerInventoryClassTestPlayer4EDD70) *types.Pointf {
			w.event("position:" + dropPlayerInventoryClassTestPlayerName4EDD70(player) + ":" + dropPlayerInventoryClassTestPoint4EDD70(player.pos))
			return &player.pos
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
				dropPlayerInventoryClassTestPoint4EDD70(*center),
				dropPlayerInventoryClassTestPoint4EDD70(point),
			))
			*output = point
			return w.randomReturn
		},
		drop: func(player *dropPlayerInventoryClassTestPlayer4EDD70, item *dropPlayerInventoryClassTestItem4EDD70, point *types.Pointf) int32 {
			call := len(w.drops)
			result := int32(0)
			if call < len(w.dropResults) {
				result = w.dropResults[call]
			}
			w.event(fmt.Sprintf(
				"drop:%d:%s:%s:%s:%d",
				call+1,
				dropPlayerInventoryClassTestPlayerName4EDD70(player),
				dropPlayerInventoryClassTestItemName4EDD70(item),
				dropPlayerInventoryClassTestPoint4EDD70(*point),
				result,
			))
			w.drops = append(w.drops, dropPlayerInventoryClassTestDrop4EDD70{player: player, item: item, point: *point})
			if w.afterDrop != nil {
				w.afterDrop(w, player, item)
			}
			return result
		},
		nextPlayer: func(player *dropPlayerInventoryClassTestPlayer4EDD70) *dropPlayerInventoryClassTestPlayer4EDD70 {
			next := player.next
			w.event("next-player:" + dropPlayerInventoryClassTestPlayerName4EDD70(player) + ":" + dropPlayerInventoryClassTestPlayerName4EDD70(next))
			return next
		},
	}
}

func newDropPlayerInventoryClassTrace4EDD70() (*dropPlayerInventoryClassTestWorld4EDD70, []string) {
	itemD := &dropPlayerInventoryClassTestItem4EDD70{name: "item-d", class: 0x10000001}
	itemC := &dropPlayerInventoryClassTestItem4EDD70{name: "item-c", class: 0x10000000}
	itemB := &dropPlayerInventoryClassTestItem4EDD70{name: "item-b", class: 0x00000010}
	itemA := &dropPlayerInventoryClassTestItem4EDD70{name: "item-a", next: itemB, class: 0x10000000}
	playerC := &dropPlayerInventoryClassTestPlayer4EDD70{
		name: "player-c",
		head: itemD,
		pos:  types.Pointf{X: 7, Y: 8},
	}
	playerB := &dropPlayerInventoryClassTestPlayer4EDD70{
		name: "player-b",
		head: itemC,
		pos:  types.Pointf{X: 5, Y: 6},
	}
	playerA := &dropPlayerInventoryClassTestPlayer4EDD70{
		name: "player-a",
		next: playerB,
		head: itemA,
		pos:  types.Pointf{X: 1, Y: 2},
	}
	w := &dropPlayerInventoryClassTestWorld4EDD70{
		first: playerA,
		randomPoints: []types.Pointf{
			{X: 10, Y: 20},
			{X: 30, Y: 40},
		},
		randomReturn: &types.Pointf{X: 999, Y: 999},
		dropResults:  []int32{math.MinInt32, math.MaxInt32},
	}
	w.afterNext = func(w *dropPlayerInventoryClassTestWorld4EDD70, item *dropPlayerInventoryClassTestItem4EDD70) {
		if item == itemA {
			item.class = 0
			item.next = nil
		}
	}
	w.afterDrop = func(w *dropPlayerInventoryClassTestWorld4EDD70, player *dropPlayerInventoryClassTestPlayer4EDD70, item *dropPlayerInventoryClassTestItem4EDD70) {
		if player == playerA && item == itemA {
			player.next = playerC
		}
	}
	want := []string{
		"first-player:player-a",
		"head:player-a:item-a",
		"class:item-a:10000000",
		"next:item-a:item-b",
		"position:player-a:3f800000,40000000",
		"random:1:42480000:3f800000,40000000:41200000,41a00000",
		"drop:1:player-a:item-a:41200000,41a00000:-2147483648",
		"class:item-b:00000010",
		"next:item-b:nil",
		"next-player:player-a:player-c",
		"head:player-c:item-d",
		"class:item-d:10000001",
		"next:item-d:nil",
		"position:player-c:40e00000,41000000",
		"random:2:42480000:40e00000,41000000:41f00000,42200000",
		"drop:2:player-c:item-d:41f00000,42200000:2147483647",
		"next-player:player-c:nil",
	}
	return w, want
}

func verifyDropPlayerInventoryClassFaultPrefixes4EDD70(
	t *testing.T,
	want []string,
	build func() *dropPlayerInventoryClassTestWorld4EDD70,
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
			dropPlayerInventoryClass4EDD70(w.hooks())
		})
	}
}

func TestDropPlayerInventoryClass4EDD70ExactOrderMutationAndFaultPrefixes(t *testing.T) {
	build := func() *dropPlayerInventoryClassTestWorld4EDD70 {
		w, _ := newDropPlayerInventoryClassTrace4EDD70()
		return w
	}
	w, want := newDropPlayerInventoryClassTrace4EDD70()
	dropPlayerInventoryClass4EDD70(w.hooks())
	if !reflect.DeepEqual(w.events, want) {
		t.Fatalf("events = %v, want %v", w.events, want)
	}
	if len(w.drops) != 2 || w.drops[0].player.name != "player-a" || w.drops[1].player.name != "player-c" {
		t.Fatalf("drops = %+v", w.drops)
	}
	if len(w.randomOutputs) != 2 || w.randomOutputs[0] != w.randomOutputs[1] || w.randomOutputs[0] == w.randomReturn {
		t.Fatalf("random output identities = %p/%p, returned = %p", w.randomOutputs[0], w.randomOutputs[1], w.randomReturn)
	}
	verifyDropPlayerInventoryClassFaultPrefixes4EDD70(t, want, build)
}

func TestDropPlayerInventoryClass4EDD70EmptyPlayerListStopsAfterFirst(t *testing.T) {
	w := &dropPlayerInventoryClassTestWorld4EDD70{}
	dropPlayerInventoryClass4EDD70(w.hooks())
	want := []string{"first-player:nil"}
	if !reflect.DeepEqual(w.events, want) {
		t.Fatalf("events = %v, want %v", w.events, want)
	}
	verifyDropPlayerInventoryClassFaultPrefixes4EDD70(t, want, func() *dropPlayerInventoryClassTestWorld4EDD70 {
		return &dropPlayerInventoryClassTestWorld4EDD70{}
	})
}

func TestDropPlayerInventoryClass4EDD70Constants(t *testing.T) {
	if dropPlayerInventoryClassMask4EDD70 != 0x10000000 {
		t.Fatalf("class mask = %#08x, want 0x10000000", dropPlayerInventoryClassMask4EDD70)
	}
	if got := math.Float32bits(dropPlayerInventoryRadius4EDD70); got != 0x42480000 {
		t.Fatalf("radius bits = %#08x, want 0x42480000", got)
	}
}
