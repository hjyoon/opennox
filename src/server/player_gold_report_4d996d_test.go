package server

import (
	"fmt"
	"reflect"
	"testing"
)

type playerGoldReportTestPlayer4D996D struct {
	name     string
	gold     uint32
	reported uint32
	index    int32
}

type playerGoldReportTestUpdate4D996D struct {
	player *playerGoldReportTestPlayer4D996D
}

type playerGoldReportTestWorld4D996D struct {
	update   *playerGoldReportTestUpdate4D996D
	events   []string
	faultAt  int
	onReport func()
}

func (w *playerGoldReportTestWorld4D996D) record(event string) {
	w.events = append(w.events, event)
	if w.faultAt != 0 && len(w.events) == w.faultAt {
		panic(event)
	}
}

func (w *playerGoldReportTestWorld4D996D) hooks() playerGoldReportHooks4D996D[
	string,
	*playerGoldReportTestUpdate4D996D,
	*playerGoldReportTestPlayer4D996D,
] {
	return playerGoldReportHooks4D996D[string, *playerGoldReportTestUpdate4D996D, *playerGoldReportTestPlayer4D996D]{
		loadPlayer: func(update *playerGoldReportTestUpdate4D996D) *playerGoldReportTestPlayer4D996D {
			player := update.player
			w.record("player:" + player.name)
			return player
		},
		loadReportedGold: func(player *playerGoldReportTestPlayer4D996D) uint32 {
			w.record(fmt.Sprintf("reported:%s:%d", player.name, player.reported))
			return player.reported
		},
		loadGold: func(player *playerGoldReportTestPlayer4D996D) uint32 {
			w.record(fmt.Sprintf("gold:%s:%d", player.name, player.gold))
			return player.gold
		},
		loadPlayerIndex: func(player *playerGoldReportTestPlayer4D996D) int32 {
			w.record(fmt.Sprintf("index:%s:%d", player.name, player.index))
			return player.index
		},
		reportGold: func(index int32, unit string) {
			w.record(fmt.Sprintf("report:%d:%s", index, unit))
			if w.onReport != nil {
				w.onReport()
			}
		},
		storeReported: func(player *playerGoldReportTestPlayer4D996D, value uint32) {
			w.record(fmt.Sprintf("store:%s:%d", player.name, value))
			player.reported = value
		},
	}
}

func TestPlayerGoldReport4D996DExactChangedTraceAndLiveReload(t *testing.T) {
	first := &playerGoldReportTestPlayer4D996D{name: "first", gold: 37, reported: 12, index: 7}
	second := &playerGoldReportTestPlayer4D996D{name: "second", gold: 91, reported: 3, index: 9}
	w := &playerGoldReportTestWorld4D996D{update: &playerGoldReportTestUpdate4D996D{player: first}}
	w.onReport = func() {
		w.update.player = second
		first.gold = 1000
	}
	playerGoldReport4D996D("unit", w.update, w.hooks())
	want := []string{
		"player:first", "reported:first:12", "gold:first:37", "index:first:7",
		"report:7:unit", "player:second", "gold:second:91", "store:second:91",
	}
	if !reflect.DeepEqual(w.events, want) {
		t.Fatalf("events = %v, want %v", w.events, want)
	}
	if first.reported != 12 || second.reported != 91 {
		t.Fatalf("reported caches = first:%d second:%d, want 12/91", first.reported, second.reported)
	}
}

func TestPlayerGoldReport4D996DEqualStopsBeforeIndex(t *testing.T) {
	player := &playerGoldReportTestPlayer4D996D{name: "same", gold: 37, reported: 37, index: 7}
	w := &playerGoldReportTestWorld4D996D{update: &playerGoldReportTestUpdate4D996D{player: player}}
	playerGoldReport4D996D("unit", w.update, w.hooks())
	want := []string{"player:same", "reported:same:37", "gold:same:37"}
	if !reflect.DeepEqual(w.events, want) {
		t.Fatalf("events = %v, want %v", w.events, want)
	}
}

func TestPlayerGoldReport4D996DAllChangedFaultPrefixes(t *testing.T) {
	want := []string{
		"player:first", "reported:first:12", "gold:first:37", "index:first:7",
		"report:7:unit", "player:first", "gold:first:37", "store:first:37",
	}
	for faultAt := 1; faultAt <= len(want); faultAt++ {
		t.Run(fmt.Sprintf("event-%d", faultAt), func(t *testing.T) {
			player := &playerGoldReportTestPlayer4D996D{name: "first", gold: 37, reported: 12, index: 7}
			w := &playerGoldReportTestWorld4D996D{
				update:  &playerGoldReportTestUpdate4D996D{player: player},
				faultAt: faultAt,
			}
			defer func() {
				if got := recover(); got != want[faultAt-1] {
					t.Fatalf("panic = %v, want %q", got, want[faultAt-1])
				}
				if prefix := want[:faultAt]; !reflect.DeepEqual(w.events, prefix) {
					t.Fatalf("events = %v, want %v", w.events, prefix)
				}
			}()
			playerGoldReport4D996D("unit", w.update, w.hooks())
		})
	}
}
