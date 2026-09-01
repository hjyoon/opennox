package server

import (
	"fmt"
	"math"
	"reflect"
	"testing"
)

type beastGuideAwardTestObject4FAE80 struct {
	name   string
	class  uint8
	update *beastGuideAwardTestUpdate4FAE80
}

type beastGuideAwardTestUpdate4FAE80 struct {
	player *beastGuideAwardTestPlayer4FAE80
}

type beastGuideAwardTestPlayer4FAE80 struct {
	name       string
	unit       *beastGuideAwardTestObject4FAE80
	levels     map[int32]uint32
	protection uint32
}

type beastGuideAwardTestWorld4FAE80 struct {
	unit       *beastGuideAwardTestObject4FAE80
	first      *beastGuideAwardTestPlayer4FAE80
	next       map[*beastGuideAwardTestPlayer4FAE80]*beastGuideAwardTestPlayer4FAE80
	related    []int32
	events     []string
	faultAt    int
	afterStore func(*beastGuideAwardTestPlayer4FAE80, int32, uint32)
}

func beastGuideAwardTestObjectName4FAE80(obj *beastGuideAwardTestObject4FAE80) string {
	if obj == nil {
		return "nil"
	}
	return obj.name
}

func beastGuideAwardTestPlayerName4FAE80(player *beastGuideAwardTestPlayer4FAE80) string {
	if player == nil {
		return "nil"
	}
	return player.name
}

func (w *beastGuideAwardTestWorld4FAE80) event(value string) {
	w.events = append(w.events, value)
	if w.faultAt != 0 && len(w.events) == w.faultAt {
		panic(value)
	}
}

func (w *beastGuideAwardTestWorld4FAE80) hooks() beastGuideAwardHooks4FAE80[
	*beastGuideAwardTestObject4FAE80,
	*beastGuideAwardTestUpdate4FAE80,
	*beastGuideAwardTestPlayer4FAE80,
	*beastGuideAwardTestPlayer4FAE80,
	string,
] {
	return beastGuideAwardHooks4FAE80[
		*beastGuideAwardTestObject4FAE80,
		*beastGuideAwardTestUpdate4FAE80,
		*beastGuideAwardTestPlayer4FAE80,
		*beastGuideAwardTestPlayer4FAE80,
		string,
	]{
		loadUnitArg: func() *beastGuideAwardTestObject4FAE80 {
			w.event("unit:" + beastGuideAwardTestObjectName4FAE80(w.unit))
			return w.unit
		},
		loadClassLow: func(unit *beastGuideAwardTestObject4FAE80) uint8 {
			w.event(fmt.Sprintf("class:%s=%02x", beastGuideAwardTestObjectName4FAE80(unit), unit.class))
			return unit.class
		},
		loadUpdateData: func(unit *beastGuideAwardTestObject4FAE80) *beastGuideAwardTestUpdate4FAE80 {
			w.event("update:" + beastGuideAwardTestObjectName4FAE80(unit))
			return unit.update
		},
		loadPlayer: func(update *beastGuideAwardTestUpdate4FAE80) *beastGuideAwardTestPlayer4FAE80 {
			w.event("player:" + beastGuideAwardTestPlayerName4FAE80(update.player))
			return update.player
		},
		loadGuideLevel: func(player *beastGuideAwardTestPlayer4FAE80, guide int32) uint32 {
			level := player.levels[guide]
			w.event(fmt.Sprintf("level:%s:%08x=%08x", player.name, uint32(guide), level))
			return level
		},
		storeGuideLevel: func(player *beastGuideAwardTestPlayer4FAE80, guide int32, level uint32) {
			w.event(fmt.Sprintf("store:%s:%08x=%08x", player.name, uint32(guide), level))
			player.levels[guide] = level
			if w.afterStore != nil {
				w.afterStore(player, guide, level)
			}
		},
		loadProtection: func(player *beastGuideAwardTestPlayer4FAE80) uint32 {
			w.event(fmt.Sprintf("protection:%s=%08x", player.name, player.protection))
			return player.protection
		},
		loadString: func(key, path string, line int) string {
			w.event(fmt.Sprintf("string:%s:%s:%d", key, path, line))
			return "localized:" + key
		},
		sendLineMessage: func(unit *beastGuideAwardTestObject4FAE80, message string) {
			w.event("line:" + beastGuideAwardTestObjectName4FAE80(unit) + ":" + message)
		},
		awardProtection: func(token uint32, guide, level int32) {
			w.event(fmt.Sprintf("award:%08x:%08x:%08x", token, uint32(guide), uint32(level)))
		},
		audio: func(id uint32, unit *beastGuideAwardTestObject4FAE80, kind int32, code uint32) {
			w.event(fmt.Sprintf("audio:%d:%s:%d:%08x", id, beastGuideAwardTestObjectName4FAE80(unit), kind, code))
		},
		rewardNotify: func(recipient *beastGuideAwardTestObject4FAE80, kind int32, source *beastGuideAwardTestObject4FAE80, guide int32) {
			w.event(fmt.Sprintf("notify:%s:%d:%s:%08x", beastGuideAwardTestObjectName4FAE80(recipient), kind, beastGuideAwardTestObjectName4FAE80(source), uint32(guide)))
		},
		relatedGuides: func(guide int32) []int32 {
			w.event(fmt.Sprintf("relations:%08x=%v", uint32(guide), w.related))
			return w.related
		},
		firstPlayer: func() *beastGuideAwardTestPlayer4FAE80 {
			w.event("first:" + beastGuideAwardTestPlayerName4FAE80(w.first))
			return w.first
		},
		nextPlayer: func(player *beastGuideAwardTestPlayer4FAE80) *beastGuideAwardTestPlayer4FAE80 {
			next := w.next[player]
			w.event("next:" + beastGuideAwardTestPlayerName4FAE80(player) + "=" + beastGuideAwardTestPlayerName4FAE80(next))
			return next
		},
		loadPlayerUnit: func(player *beastGuideAwardTestPlayer4FAE80) *beastGuideAwardTestObject4FAE80 {
			w.event("player-unit:" + player.name + "=" + beastGuideAwardTestObjectName4FAE80(player.unit))
			return player.unit
		},
		reportGuideAward: func(unit *beastGuideAwardTestObject4FAE80, guide, notify, shop int32) {
			w.event(fmt.Sprintf("report:%s:%08x:%08x:%08x", beastGuideAwardTestObjectName4FAE80(unit), uint32(guide), uint32(notify), uint32(shop)))
		},
	}
}

func newBeastGuideAwardTestWorld4FAE80() *beastGuideAwardTestWorld4FAE80 {
	player := &beastGuideAwardTestPlayer4FAE80{
		name:       "source-player",
		levels:     make(map[int32]uint32),
		protection: 0x89abcdef,
	}
	unit := &beastGuideAwardTestObject4FAE80{name: "source", class: 0xf4}
	unit.update = &beastGuideAwardTestUpdate4FAE80{player: player}
	player.unit = unit
	nilUnitPlayer := &beastGuideAwardTestPlayer4FAE80{name: "nil-unit", levels: make(map[int32]uint32)}
	otherUnit := &beastGuideAwardTestObject4FAE80{name: "other", class: 4}
	otherPlayer := &beastGuideAwardTestPlayer4FAE80{name: "other-player", unit: otherUnit, levels: make(map[int32]uint32)}
	return &beastGuideAwardTestWorld4FAE80{
		unit:    unit,
		first:   player,
		related: []int32{8, 9},
		next: map[*beastGuideAwardTestPlayer4FAE80]*beastGuideAwardTestPlayer4FAE80{
			player:        nilUnitPlayer,
			nilUnitPlayer: otherPlayer,
			otherPlayer:   nil,
		},
	}
}

func beastGuideAwardSuccessTrace4FAE80() []string {
	return []string{
		"unit:source",
		"class:source=f4",
		"update:source",
		"player:source-player",
		"level:source-player:00000018=00000000",
		"store:source-player:00000018=00000001",
		"player:source-player",
		"level:source-player:00000018=00000001",
		"protection:source-player=89abcdef",
		"award:89abcdef:00000018:00000001",
		"audio:227:source:0:00000000",
		"notify:source:1:source:00000018",
		"relations:00000018=[8 9]",
		"player:source-player",
		"store:source-player:00000008=00000001",
		"player:source-player",
		"level:source-player:00000008=00000001",
		"protection:source-player=89abcdef",
		"award:89abcdef:00000008:00000001",
		"player:source-player",
		"store:source-player:00000009=00000001",
		"player:source-player",
		"level:source-player:00000009=00000001",
		"protection:source-player=89abcdef",
		"award:89abcdef:00000009:00000001",
		"first:source-player",
		"player-unit:source-player=source",
		"next:source-player=nil-unit",
		"player-unit:nil-unit=nil",
		"next:nil-unit=other-player",
		"player-unit:other-player=other",
		"notify:other:1:source:00000018",
		"next:other-player=nil",
		"report:source:00000018:80000000:00000000",
	}
}

func TestBeastGuideAward4FAE80ExactSuccessTraceAndFaultPrefixes(t *testing.T) {
	want := beastGuideAwardSuccessTrace4FAE80()
	build := newBeastGuideAwardTestWorld4FAE80

	w := build()
	if got := beastGuideAward4FAE80(24, math.MinInt32, w.hooks()); got != 1 {
		t.Fatalf("result = %d, want 1", got)
	}
	if !reflect.DeepEqual(w.events, want) {
		t.Fatalf("events = %v, want %v", w.events, want)
	}

	for faultAt := 1; faultAt <= len(want); faultAt++ {
		t.Run(fmt.Sprintf("fault-%02d", faultAt), func(t *testing.T) {
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
			beastGuideAward4FAE80(24, math.MinInt32, w.hooks())
		})
	}
}

func TestBeastGuideAward4FAE80ClassGuideOwnedAndNotifyGates(t *testing.T) {
	t.Run("non-player", func(t *testing.T) {
		w := newBeastGuideAwardTestWorld4FAE80()
		w.unit.class = 0xf0
		if got := beastGuideAward4FAE80(math.MinInt32, 1, w.hooks()); got != 0 {
			t.Fatalf("result = %d, want 0", got)
		}
		want := []string{"unit:source", "class:source=f0"}
		if !reflect.DeepEqual(w.events, want) {
			t.Fatalf("events = %v, want %v", w.events, want)
		}
	})

	for _, guide := range []int32{math.MinInt32, -1, 0, 41, math.MaxInt32} {
		t.Run(fmt.Sprintf("invalid-%08x", uint32(guide)), func(t *testing.T) {
			w := newBeastGuideAwardTestWorld4FAE80()
			if got := beastGuideAward4FAE80(guide, 1, w.hooks()); got != 0 {
				t.Fatalf("result = %d, want 0", got)
			}
			want := []string{
				"unit:source",
				"class:source=f4",
				fmt.Sprintf("string:AwardGuideError:%s:39", beastGuideAwardMessagePath4FAE80),
				"line:source:localized:AwardGuideError",
			}
			if !reflect.DeepEqual(w.events, want) {
				t.Fatalf("events = %v, want %v", w.events, want)
			}
		})
	}

	t.Run("already-owned", func(t *testing.T) {
		w := newBeastGuideAwardTestWorld4FAE80()
		w.unit.update.player.levels[40] = math.MaxUint32
		if got := beastGuideAward4FAE80(40, 1, w.hooks()); got != 0 {
			t.Fatalf("result = %d, want 0", got)
		}
		wantTail := []string{
			"update:source",
			"player:source-player",
			"level:source-player:00000028=ffffffff",
		}
		if got := w.events[len(w.events)-len(wantTail):]; !reflect.DeepEqual(got, wantTail) {
			t.Fatalf("tail = %v, want %v", got, wantTail)
		}
	})

	t.Run("zero-notify-skips-live-events-but-reports", func(t *testing.T) {
		w := newBeastGuideAwardTestWorld4FAE80()
		w.related = nil
		if got := beastGuideAward4FAE80(1, 0, w.hooks()); got != 1 {
			t.Fatalf("result = %d, want 1", got)
		}
		wantTail := []string{
			"award:89abcdef:00000001:00000001",
			"relations:00000001=[]",
			"report:source:00000001:00000000:00000000",
		}
		if got := w.events[len(w.events)-len(wantTail):]; !reflect.DeepEqual(got, wantTail) {
			t.Fatalf("tail = %v, want %v", got, wantTail)
		}
	})
}

func TestBeastGuideAward4FAE80ReloadsPlayerAfterStores(t *testing.T) {
	w := newBeastGuideAwardTestWorld4FAE80()
	w.related = nil
	replacement := &beastGuideAwardTestPlayer4FAE80{
		name:       "replacement",
		levels:     map[int32]uint32{24: math.MaxUint32},
		protection: 0x10203040,
	}
	w.afterStore = func(player *beastGuideAwardTestPlayer4FAE80, guide int32, level uint32) {
		if player.name == "source-player" && guide == 24 && level == 1 {
			w.unit.update.player = replacement
		}
	}
	if got := beastGuideAward4FAE80(24, 0, w.hooks()); got != 1 {
		t.Fatalf("result = %d, want 1", got)
	}
	want := []string{
		"player:replacement",
		"level:replacement:00000018=ffffffff",
		"protection:replacement=10203040",
		"award:10203040:00000018:ffffffff",
	}
	for index := range w.events {
		if index+len(want) <= len(w.events) && reflect.DeepEqual(w.events[index:index+len(want)], want) {
			return
		}
	}
	t.Fatalf("events = %v, missing %v", w.events, want)
}
