package server

import (
	"fmt"
	"reflect"
	"testing"
)

type playerRespawnItemInit4EF750 struct {
	name string
}

type playerRespawnItemAttrs4EF750 struct {
	name string
}

type playerRespawnItemUpdate4EF750 struct {
	name string
	mark uint32
}

type playerRespawnItemObject4EF750 struct {
	name   string
	init   *playerRespawnItemInit4EF750
	flags  uint32
	class  uint32
	update *playerRespawnItemUpdate4EF750
}

type playerRespawnItemWorld4EF750 struct {
	events       []string
	faultAt      int
	onEvent      func(string)
	typeID       string
	created      *playerRespawnItemObject4EF750
	attrs        *playerRespawnItemAttrs4EF750
	player       *playerRespawnItemObject4EF750
	a4           int32
	a5           int32
	calledInit   *playerRespawnItemInit4EF750
	placedPlayer *playerRespawnItemObject4EF750
	placedItem   *playerRespawnItemObject4EF750
	placedA4     int32
	placedA5     int32
}

func (w *playerRespawnItemWorld4EF750) event(event string) {
	w.events = append(w.events, event)
	if w.onEvent != nil {
		w.onEvent(event)
	}
	if w.faultAt != 0 && len(w.events) == w.faultAt {
		panic("injected fault")
	}
}

func playerRespawnItemObjectName4EF750(obj *playerRespawnItemObject4EF750) string {
	if obj == nil {
		return "nil"
	}
	return obj.name
}

func playerRespawnItemInitName4EF750(init *playerRespawnItemInit4EF750) string {
	if init == nil {
		return "nil"
	}
	return init.name
}

func playerRespawnItemAttrsName4EF750(attrs *playerRespawnItemAttrs4EF750) string {
	if attrs == nil {
		return "nil"
	}
	return attrs.name
}

func (w *playerRespawnItemWorld4EF750) hooks() playerRespawnItemHooks4EF750[
	*playerRespawnItemObject4EF750,
	*playerRespawnItemInit4EF750,
	*playerRespawnItemAttrs4EF750,
	*playerRespawnItemUpdate4EF750,
] {
	return playerRespawnItemHooks4EF750[
		*playerRespawnItemObject4EF750,
		*playerRespawnItemInit4EF750,
		*playerRespawnItemAttrs4EF750,
		*playerRespawnItemUpdate4EF750,
	]{
		loadTypeIDArg: func() string {
			value := w.typeID
			w.event("type:" + value)
			return value
		},
		newObject: func(typeID string) *playerRespawnItemObject4EF750 {
			value := w.created
			w.event("new:" + typeID)
			return value
		},
		loadInit: func(item *playerRespawnItemObject4EF750) *playerRespawnItemInit4EF750 {
			value := item.init
			w.event("init:" + playerRespawnItemInitName4EF750(value))
			return value
		},
		callInit: func(init *playerRespawnItemInit4EF750, item *playerRespawnItemObject4EF750, value uint32) {
			w.event(fmt.Sprintf("call-init:%s:%s:%d", init.name, item.name, value))
			w.calledInit = init
		},
		loadAttrsArg: func() *playerRespawnItemAttrs4EF750 {
			value := w.attrs
			w.event("attrs:" + playerRespawnItemAttrsName4EF750(value))
			return value
		},
		applyAttrs: func(item *playerRespawnItemObject4EF750, attrs *playerRespawnItemAttrs4EF750) {
			w.event("apply:" + item.name + ":" + attrs.name)
		},
		loadPlaceA5Arg: func() int32 {
			value := w.a5
			w.event(fmt.Sprintf("a5:%d", value))
			return value
		},
		loadPlaceA4Arg: func() int32 {
			value := w.a4
			w.event(fmt.Sprintf("a4:%d", value))
			return value
		},
		loadPlayerArg: func() *playerRespawnItemObject4EF750 {
			value := w.player
			w.event("player:" + playerRespawnItemObjectName4EF750(value))
			return value
		},
		placeInventory: func(player, item *playerRespawnItemObject4EF750, a4, a5 int32) {
			w.event(fmt.Sprintf("place:%s:%s:%d:%d", playerRespawnItemObjectName4EF750(player), item.name, a4, a5))
			w.placedPlayer = player
			w.placedItem = item
			w.placedA4 = a4
			w.placedA5 = a5
		},
		loadFlags: func(item *playerRespawnItemObject4EF750) uint32 {
			value := item.flags
			w.event(fmt.Sprintf("flags:%#x", value))
			return value
		},
		loadClass: func(item *playerRespawnItemObject4EF750) uint32 {
			value := item.class
			w.event(fmt.Sprintf("class:%#x", value))
			return value
		},
		storeFlags: func(item *playerRespawnItemObject4EF750, flags uint32) {
			w.event(fmt.Sprintf("store-flags:%#x", flags))
			item.flags = flags
		},
		loadUpdateData: func(item *playerRespawnItemObject4EF750) *playerRespawnItemUpdate4EF750 {
			value := item.update
			name := "nil"
			if value != nil {
				name = value.name
			}
			w.event("update:" + name)
			return value
		},
		loadUpdateMark: func(update *playerRespawnItemUpdate4EF750) uint32 {
			value := update.mark
			w.event(fmt.Sprintf("mark:%s:%#x", update.name, value))
			return value
		},
		storeUpdateMark: func(update *playerRespawnItemUpdate4EF750, mark uint32) {
			w.event(fmt.Sprintf("store-mark:%s:%#x", update.name, mark))
			update.mark = mark
		},
	}
}

func newPlayerRespawnItemWorld4EF750() *playerRespawnItemWorld4EF750 {
	return &playerRespawnItemWorld4EF750{
		typeID: "Longsword",
		created: &playerRespawnItemObject4EF750{
			name:   "item",
			init:   &playerRespawnItemInit4EF750{name: "init-a"},
			flags:  0xa5a80040,
			class:  playerRespawnItemUpdateMask4EF750,
			update: &playerRespawnItemUpdate4EF750{name: "update-a", mark: 0x10},
		},
		attrs:  &playerRespawnItemAttrs4EF750{name: "attrs-a"},
		player: &playerRespawnItemObject4EF750{name: "player-a"},
		a4:     4,
		a5:     5,
	}
}

func TestPlayerRespawnItem4EF750ExactOrderAndState(t *testing.T) {
	world := newPlayerRespawnItemWorld4EF750()
	got := playerRespawnItem4EF750(world.hooks())
	wantEvents := []string{
		"type:Longsword", "new:Longsword", "init:init-a",
		"call-init:init-a:item:0", "attrs:attrs-a", "apply:item:attrs-a",
		"a5:5", "a4:4", "player:player-a", "place:player-a:item:4:5",
		"flags:0xa5a80040", "class:0x3001000", "store-flags:0xa5a00040",
		"update:update-a", "mark:update-a:0x10", "store-mark:update-a:0x11",
	}
	if !reflect.DeepEqual(world.events, wantEvents) {
		t.Fatalf("events = %v, want %v", world.events, wantEvents)
	}
	if got != world.created || world.calledInit != world.created.init {
		t.Fatalf("result/init = %p/%p, want %p/%p", got, world.calledInit, world.created, world.created.init)
	}
	if world.placedPlayer != world.player || world.placedItem != world.created || world.placedA4 != 4 || world.placedA5 != 5 {
		t.Fatalf("placement = %p/%p/%d/%d", world.placedPlayer, world.placedItem, world.placedA4, world.placedA5)
	}
	if world.created.flags != 0xa5a00040 || world.created.update.mark != 0x11 {
		t.Fatalf("item state = flags %#x mark %#x", world.created.flags, world.created.update.mark)
	}
}

func TestPlayerRespawnItem4EF750NilCreationReturnsBeforeLaterLoads(t *testing.T) {
	world := newPlayerRespawnItemWorld4EF750()
	world.created = nil
	if got := playerRespawnItem4EF750(world.hooks()); got != nil {
		t.Fatalf("result = %p, want nil", got)
	}
	want := []string{"type:Longsword", "new:Longsword"}
	if !reflect.DeepEqual(world.events, want) {
		t.Fatalf("events = %v, want %v", world.events, want)
	}
}

func TestPlayerRespawnItem4EF750SkipsNilInitAttrsAndUnmarkedUpdate(t *testing.T) {
	world := newPlayerRespawnItemWorld4EF750()
	world.created.init = nil
	world.attrs = nil
	world.created.class = 0xfcffe000
	got := playerRespawnItem4EF750(world.hooks())
	if got != world.created {
		t.Fatalf("result = %p, want %p", got, world.created)
	}
	want := []string{
		"type:Longsword", "new:Longsword", "init:nil", "attrs:nil",
		"a5:5", "a4:4", "player:player-a", "place:player-a:item:4:5",
		"flags:0xa5a80040", "class:0xfcffe000", "store-flags:0xa5a00040",
	}
	if !reflect.DeepEqual(world.events, want) {
		t.Fatalf("events = %v, want %v", world.events, want)
	}
	if world.created.update.mark != 0x10 {
		t.Fatalf("unmarked update = %#x, want unchanged", world.created.update.mark)
	}
}

func TestPlayerRespawnItem4EF750UsesCachedAndLiveValues(t *testing.T) {
	world := newPlayerRespawnItemWorld4EF750()
	initA := world.created.init
	initB := &playerRespawnItemInit4EF750{name: "init-b"}
	attrsB := &playerRespawnItemAttrs4EF750{name: "attrs-b"}
	playerB := &playerRespawnItemObject4EF750{name: "player-b"}
	updateA := world.created.update
	updateB := &playerRespawnItemUpdate4EF750{name: "update-b", mark: 0x20}

	world.onEvent = func(event string) {
		switch event {
		case "init:init-a":
			world.created.init = initB
		case "call-init:init-a:item:0":
			world.attrs = attrsB
		case "apply:item:attrs-b":
			world.a5 = 55
			world.a4 = 44
			world.player = playerB
		case "a5:55":
			world.a5 = 99
			world.a4 = 66
		case "a4:66":
			world.a4 = 88
		case "player:player-b":
			world.player = nil
		case "place:player-b:item:66:55":
			world.created.flags = 0x123c5678
			world.created.class = 0
		case "flags:0x123c5678":
			world.created.flags = 0xffffffff
			world.created.class = 0x01000000
		case "class:0x1000000":
			world.created.class = 0
		case "store-flags:0x12345678":
			world.created.update = updateB
		case "update:update-b":
			world.created.update = updateA
		case "mark:update-b:0x20":
			updateB.mark = 0x80
		}
	}

	got := playerRespawnItem4EF750(world.hooks())
	if got != world.created || world.calledInit != initA {
		t.Fatalf("result/init = %p/%p, want %p/%p", got, world.calledInit, world.created, initA)
	}
	if world.placedPlayer != playerB || world.placedA4 != 66 || world.placedA5 != 55 {
		t.Fatalf("live placement = %p/%d/%d, want %p/66/55", world.placedPlayer, world.placedA4, world.placedA5, playerB)
	}
	if world.created.flags != 0x12345678 {
		t.Fatalf("cached flags store = %#x, want 0x12345678", world.created.flags)
	}
	if updateA.mark != 0x10 || updateB.mark != 0x21 {
		t.Fatalf("live/cached update marks = %#x/%#x, want 0x10/0x21", updateA.mark, updateB.mark)
	}
}

func TestPlayerRespawnItem4EF750EachEquipmentClassMarksUpdate(t *testing.T) {
	for _, class := range []uint32{0x00001000, 0x01000000, 0x02000000, 0x03001000} {
		t.Run(fmt.Sprintf("class_%08x", class), func(t *testing.T) {
			world := newPlayerRespawnItemWorld4EF750()
			world.created.class = class
			world.created.update.mark = 0xfffffffe
			playerRespawnItem4EF750(world.hooks())
			if world.created.update.mark != 0xffffffff {
				t.Fatalf("mark = %#x, want %#x", world.created.update.mark, uint32(0xffffffff))
			}
		})
	}
}

func TestPlayerRespawnItem4EF750EveryObservableFaultPrefix(t *testing.T) {
	base := newPlayerRespawnItemWorld4EF750()
	playerRespawnItem4EF750(base.hooks())
	want := append([]string(nil), base.events...)

	for faultAt := 1; faultAt <= len(want); faultAt++ {
		t.Run(fmt.Sprintf("event_%02d", faultAt), func(t *testing.T) {
			world := newPlayerRespawnItemWorld4EF750()
			world.faultAt = faultAt
			func() {
				defer func() {
					if recover() == nil {
						t.Fatal("expected injected fault")
					}
				}()
				playerRespawnItem4EF750(world.hooks())
			}()
			if !reflect.DeepEqual(world.events, want[:faultAt]) {
				t.Fatalf("events = %v, want prefix %v", world.events, want[:faultAt])
			}
		})
	}
}

func TestPlayerRespawnItem4EF750EveryNilCreationFaultPrefix(t *testing.T) {
	base := newPlayerRespawnItemWorld4EF750()
	base.created = nil
	playerRespawnItem4EF750(base.hooks())
	want := append([]string(nil), base.events...)

	for faultAt := 1; faultAt <= len(want); faultAt++ {
		t.Run(fmt.Sprintf("event_%02d", faultAt), func(t *testing.T) {
			world := newPlayerRespawnItemWorld4EF750()
			world.created = nil
			world.faultAt = faultAt
			func() {
				defer func() {
					if recover() == nil {
						t.Fatal("expected injected fault")
					}
				}()
				playerRespawnItem4EF750(world.hooks())
			}()
			if !reflect.DeepEqual(world.events, want[:faultAt]) {
				t.Fatalf("events = %v, want prefix %v", world.events, want[:faultAt])
			}
		})
	}
}

func TestPlayerRespawnItem4EF750EverySparsePathFaultPrefix(t *testing.T) {
	configure := func(world *playerRespawnItemWorld4EF750) {
		world.created.init = nil
		world.attrs = nil
		world.created.class = 0xfcffe000
	}
	base := newPlayerRespawnItemWorld4EF750()
	configure(base)
	playerRespawnItem4EF750(base.hooks())
	want := append([]string(nil), base.events...)

	for faultAt := 1; faultAt <= len(want); faultAt++ {
		t.Run(fmt.Sprintf("event_%02d", faultAt), func(t *testing.T) {
			world := newPlayerRespawnItemWorld4EF750()
			configure(world)
			world.faultAt = faultAt
			func() {
				defer func() {
					if recover() == nil {
						t.Fatal("expected injected fault")
					}
				}()
				playerRespawnItem4EF750(world.hooks())
			}()
			if !reflect.DeepEqual(world.events, want[:faultAt]) {
				t.Fatalf("events = %v, want prefix %v", world.events, want[:faultAt])
			}
		})
	}
}
