package server

import (
	"fmt"
	"math"
	"reflect"
	"testing"
)

type treasureDropTestPoint4ED710 struct {
	name string
}

type treasureDropTestPlayer4ED710 struct {
	name          string
	treasureCount uint32
	treasureMax   uint32
}

type treasureDropTestUpdate4ED710 struct {
	name   string
	player *treasureDropTestPlayer4ED710
}

type treasureDropTestObject4ED710 struct {
	name     string
	classLow uint8
	update   *treasureDropTestUpdate4ED710
}

type treasureDropTestWorld4ED710 struct {
	owner, treasure *treasureDropTestObject4ED710
	point           *treasureDropTestPoint4ED710
	ownerArg        *treasureDropTestObject4ED710
	treasureArg     *treasureDropTestObject4ED710
	pointArg        *treasureDropTestPoint4ED710

	defaultResult int32
	gameResult    int32
	maximum       uint32
	events        []string
	faultAt       int

	afterDefault func(*treasureDropTestWorld4ED710)
	afterGame    func(*treasureDropTestWorld4ED710)
	afterMaximum func(*treasureDropTestWorld4ED710)
}

func newTreasureDropTestWorld4ED710() *treasureDropTestWorld4ED710 {
	player := &treasureDropTestPlayer4ED710{name: "player-a", treasureCount: 0}
	update := &treasureDropTestUpdate4ED710{name: "update-a", player: player}
	owner := &treasureDropTestObject4ED710{name: "owner-a", classLow: 4, update: update}
	treasure := &treasureDropTestObject4ED710{name: "treasure-a"}
	point := &treasureDropTestPoint4ED710{name: "point-a"}
	return &treasureDropTestWorld4ED710{
		owner: owner, treasure: treasure, point: point,
		ownerArg: owner, treasureArg: treasure, pointArg: point,
		defaultResult: 1, gameResult: 1, maximum: 7,
	}
}

func (w *treasureDropTestWorld4ED710) event(name string) {
	w.events = append(w.events, name)
	if w.faultAt != 0 && len(w.events) == w.faultAt {
		panic(name)
	}
}

func treasureDropObjectName4ED710(obj *treasureDropTestObject4ED710) string {
	if obj == nil {
		return "nil"
	}
	return obj.name
}

func treasureDropPlayerName4ED710(player *treasureDropTestPlayer4ED710) string {
	if player == nil {
		return "nil"
	}
	return player.name
}

func (w *treasureDropTestWorld4ED710) hooks() treasureDropHooks4ED710[
	*treasureDropTestObject4ED710,
	*treasureDropTestUpdate4ED710,
	*treasureDropTestPlayer4ED710,
	*treasureDropTestPoint4ED710,
] {
	return treasureDropHooks4ED710[
		*treasureDropTestObject4ED710,
		*treasureDropTestUpdate4ED710,
		*treasureDropTestPlayer4ED710,
		*treasureDropTestPoint4ED710,
	]{
		loadPointArg: func() *treasureDropTestPoint4ED710 {
			w.event("point-arg:" + w.pointArg.name)
			return w.pointArg
		},
		loadTreasureArg: func() *treasureDropTestObject4ED710 {
			w.event("treasure-arg:" + treasureDropObjectName4ED710(w.treasureArg))
			return w.treasureArg
		},
		loadOwnerArg: func() *treasureDropTestObject4ED710 {
			w.event("owner-arg:" + treasureDropObjectName4ED710(w.ownerArg))
			return w.ownerArg
		},
		defaultDrop: func(owner, treasure *treasureDropTestObject4ED710, point *treasureDropTestPoint4ED710) int32 {
			w.event(fmt.Sprintf("default:%s:%s:%s", owner.name, treasure.name, point.name))
			result := w.defaultResult
			if w.afterDefault != nil {
				w.afterDefault(w)
			}
			return result
		},
		loadClassLow: func(owner *treasureDropTestObject4ED710) uint8 {
			w.event(fmt.Sprintf("class:%s:%#x", owner.name, owner.classLow))
			return owner.classLow
		},
		gameFlag: func(flag uint32) int32 {
			w.event(fmt.Sprintf("game:%d", flag))
			result := w.gameResult
			if w.afterGame != nil {
				w.afterGame(w)
			}
			return result
		},
		loadUpdate: func(owner *treasureDropTestObject4ED710) *treasureDropTestUpdate4ED710 {
			w.event("update:" + owner.name + ":" + owner.update.name)
			return owner.update
		},
		loadPlayer: func(update *treasureDropTestUpdate4ED710) *treasureDropTestPlayer4ED710 {
			w.event("player:" + update.name + ":" + treasureDropPlayerName4ED710(update.player))
			return update.player
		},
		loadCount: func(player *treasureDropTestPlayer4ED710) uint32 {
			w.event(fmt.Sprintf("count:%s:%#x", player.name, player.treasureCount))
			return player.treasureCount
		},
		storeCount: func(player *treasureDropTestPlayer4ED710, value uint32) {
			w.event(fmt.Sprintf("store-count:%s:%#x", player.name, value))
			player.treasureCount = value
		},
		treasureMax: func() uint32 {
			w.event(fmt.Sprintf("maximum:%#x", w.maximum))
			value := w.maximum
			if w.afterMaximum != nil {
				w.afterMaximum(w)
			}
			return value
		},
		storeMax: func(player *treasureDropTestPlayer4ED710, value uint32) {
			w.event(fmt.Sprintf("store-max:%s:%#x", player.name, value))
			player.treasureMax = value
		},
		report: func(owner *treasureDropTestObject4ED710) {
			w.event("report:" + owner.name)
		},
		audio: func(id uint32, owner *treasureDropTestObject4ED710, kind int32, code uint32) {
			w.event(fmt.Sprintf("audio:%d:%s:%d:%d", id, owner.name, kind, code))
		},
	}
}

func treasureDropSuccessEvents4ED710() []string {
	return []string{
		"point-arg:point-a", "treasure-arg:treasure-a", "owner-arg:owner-a",
		"default:owner-a:treasure-a:point-a", "class:owner-a:0x4", "game:64",
		"update:owner-a:update-a", "player:update-a:player-a", "count:player-a:0x0",
		"store-count:player-a:0xffffffff", "maximum:0x7", "player:update-a:player-a",
		"store-max:player-a:0x7", "report:owner-a", "audio:308:owner-a:0:0",
	}
}

func verifyTreasureDropFaultPrefixes4ED710(t *testing.T, want []string) {
	t.Helper()
	for faultAt := 1; faultAt <= len(want); faultAt++ {
		t.Run(fmt.Sprintf("fault-%d", faultAt), func(t *testing.T) {
			w := newTreasureDropTestWorld4ED710()
			w.faultAt = faultAt
			defer func() {
				if got := recover(); got != want[faultAt-1] {
					t.Fatalf("panic = %v, want %q", got, want[faultAt-1])
				}
				if !reflect.DeepEqual(w.events, want[:faultAt]) {
					t.Fatalf("events = %#v, want %#v", w.events, want[:faultAt])
				}
			}()
			treasureDrop4ED710(w.hooks())
		})
	}
}

func TestTreasureDrop4ED710ExactSuccessTraceAndWrappingDecrement(t *testing.T) {
	w := newTreasureDropTestWorld4ED710()
	if got := treasureDrop4ED710(w.hooks()); got != 1 {
		t.Fatalf("result = %d, want 1", got)
	}
	want := treasureDropSuccessEvents4ED710()
	if !reflect.DeepEqual(w.events, want) {
		t.Fatalf("events = %#v, want %#v", w.events, want)
	}
	player := w.owner.update.player
	if player.treasureCount != math.MaxUint32 || player.treasureMax != 7 {
		t.Fatalf("count/max = %#x/%#x, want %#x/7", player.treasureCount, player.treasureMax, uint32(math.MaxUint32))
	}
	verifyTreasureDropFaultPrefixes4ED710(t, want)
}

func TestTreasureDrop4ED710DefaultAndGameUseWholeEAX(t *testing.T) {
	for _, result := range []int32{0, 1, -1, math.MinInt32} {
		t.Run(fmt.Sprintf("default-%08x", uint32(result)), func(t *testing.T) {
			w := newTreasureDropTestWorld4ED710()
			w.defaultResult = result
			got := treasureDrop4ED710(w.hooks())
			if result == 0 {
				if got != 0 || w.events[len(w.events)-1] != "default:owner-a:treasure-a:point-a" {
					t.Fatalf("result/events = %d/%#v", got, w.events)
				}
				return
			}
			if got != 1 || w.events[len(w.events)-1] != "audio:308:owner-a:0:0" {
				t.Fatalf("result/events = %d/%#v", got, w.events)
			}
		})
	}

	for _, result := range []int32{0, 1, -1, math.MinInt32} {
		t.Run(fmt.Sprintf("game-%08x", uint32(result)), func(t *testing.T) {
			w := newTreasureDropTestWorld4ED710()
			w.gameResult = result
			got := treasureDrop4ED710(w.hooks())
			if got != 1 {
				t.Fatalf("result = %d, want 1", got)
			}
			wantLast := "game:64"
			if result != 0 {
				wantLast = "audio:308:owner-a:0:0"
			}
			if w.events[len(w.events)-1] != wantLast {
				t.Fatalf("events = %#v, want last %q", w.events, wantLast)
			}
		})
	}
}

func TestTreasureDrop4ED710NonPlayerSkipsGameAndPlayerData(t *testing.T) {
	w := newTreasureDropTestWorld4ED710()
	w.owner.classLow = 0xf8
	if got := treasureDrop4ED710(w.hooks()); got != 1 {
		t.Fatalf("result = %d, want 1", got)
	}
	want := []string{
		"point-arg:point-a", "treasure-arg:treasure-a", "owner-arg:owner-a",
		"default:owner-a:treasure-a:point-a", "class:owner-a:0xf8",
	}
	if !reflect.DeepEqual(w.events, want) {
		t.Fatalf("events = %#v, want %#v", w.events, want)
	}
}

func TestTreasureDrop4ED710CachesArgumentsAndUpdateButReloadsPlayer(t *testing.T) {
	w := newTreasureDropTestWorld4ED710()
	updateB := &treasureDropTestUpdate4ED710{
		name:   "update-b",
		player: &treasureDropTestPlayer4ED710{name: "player-b", treasureCount: 5},
	}
	playerC := &treasureDropTestPlayer4ED710{name: "player-c", treasureCount: 99}
	w.afterDefault = func(w *treasureDropTestWorld4ED710) {
		w.ownerArg = &treasureDropTestObject4ED710{name: "owner-b"}
		w.treasureArg = &treasureDropTestObject4ED710{name: "treasure-b"}
		w.pointArg = &treasureDropTestPoint4ED710{name: "point-b"}
	}
	w.afterGame = func(w *treasureDropTestWorld4ED710) {
		w.owner.update = updateB
	}
	w.afterMaximum = func(w *treasureDropTestWorld4ED710) {
		w.owner.update = &treasureDropTestUpdate4ED710{name: "update-c", player: playerC}
		updateB.player = playerC
	}

	if got := treasureDrop4ED710(w.hooks()); got != 1 {
		t.Fatalf("result = %d, want 1", got)
	}
	if updateB.player != playerC || playerC.treasureMax != 7 {
		t.Fatalf("reloaded player/max = %p/%d, want %p/7", updateB.player, playerC.treasureMax, playerC)
	}
	if updateB.player.treasureCount != 99 {
		t.Fatalf("new player count = %d, want 99", updateB.player.treasureCount)
	}
	if !containsTreasureDropEvent4ED710(w.events, "store-count:player-b:0x4") ||
		!containsTreasureDropEvent4ED710(w.events, "player:update-b:player-c") ||
		!containsTreasureDropEvent4ED710(w.events, "report:owner-a") {
		t.Fatalf("events = %#v", w.events)
	}
	for _, event := range w.events {
		if event == "update:owner-b:update-c" || event == "report:owner-b" {
			t.Fatalf("uncached argument/update event %q in %#v", event, w.events)
		}
	}
}

func containsTreasureDropEvent4ED710(events []string, want string) bool {
	for _, event := range events {
		if event == want {
			return true
		}
	}
	return false
}
