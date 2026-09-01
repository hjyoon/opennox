package server

import (
	"fmt"
	"reflect"
	"strings"
	"testing"
)

type mapTransitionPlayerInitTestWorld4FC6D0 struct {
	events  []string
	faultAt int

	mapInitState  int32
	mapEntryState int32
	firstResults  []string
	firstCalls    int
	next          map[string]string
	flags         map[int32]int32

	questStage       int32
	restorePredicate int32
	restoreReady     int32
	queuedRestore    int32
	dataRoot         string
	tempPath         string

	dataByUnit      map[string]string
	playerByData    map[string]string
	playerField4792 map[string]int32
	updateField138  map[string]int32
	playerIndex     map[string]uint8
	playerField3680 map[string]uint8
	saveResult      int32
	prepareResult   int32
	restoreResult   int32

	onFirst    func(int)
	onSave     func(string, uint8)
	onPrepare  func(uint8)
	onGauntlet func(uint8, int32)
	onRestore  func(string, uint8)
	onDelete   func(string)
	onFinish   func(uint8)
	onEnchant  func(string)
}

func (w *mapTransitionPlayerInitTestWorld4FC6D0) record(event string) {
	w.events = append(w.events, event)
	if w.faultAt != 0 && len(w.events) == w.faultAt {
		panic(event)
	}
}

func (w *mapTransitionPlayerInitTestWorld4FC6D0) hooks() mapTransitionPlayerInitHooks4FC6D0[string, string, string] {
	return mapTransitionPlayerInitHooks4FC6D0[string, string, string]{
		loadMapInitState: func() int32 {
			value := w.mapInitState
			w.record(fmt.Sprintf("map-init:%d", value))
			return value
		},
		loadMapEntryState: func() int32 {
			value := w.mapEntryState
			w.record(fmt.Sprintf("map-entry:%d", value))
			return value
		},
		firstPlayerUnit: func() string {
			w.firstCalls++
			var unit string
			if w.firstCalls <= len(w.firstResults) {
				unit = w.firstResults[w.firstCalls-1]
			}
			w.record(fmt.Sprintf("first%d:%s", w.firstCalls, unit))
			if w.onFirst != nil {
				w.onFirst(w.firstCalls)
			}
			return unit
		},
		nextPlayerUnit: func(unit string) string {
			next := w.next[unit]
			w.record(fmt.Sprintf("next:%s:%s", unit, next))
			return next
		},
		hasGame: func(flag int32) int32 {
			value := w.flags[flag]
			w.record(fmt.Sprintf("flag:%#x:%d", flag, value))
			return value
		},
		loadQuestStage: func() int32 {
			value := w.questStage
			w.record(fmt.Sprintf("quest-stage:%d", value))
			return value
		},
		loadRestorePredicate: func() int32 {
			value := w.restorePredicate
			w.record(fmt.Sprintf("restore-predicate:%d", value))
			return value
		},
		loadRestoreReady: func() int32 {
			value := w.restoreReady
			w.record(fmt.Sprintf("restore-ready:%d", value))
			return value
		},
		loadQueuedRestore: func() int32 {
			value := w.queuedRestore
			w.record(fmt.Sprintf("queued-restore:%d", value))
			return value
		},
		sendQuestStage: func(index int32) {
			w.record(fmt.Sprintf("quest-stage-send:%d", index))
		},
		sendQuestRestore: func(index, state int32) {
			w.record(fmt.Sprintf("quest-restore-send:%d:%d", index, state))
		},
		storeQueuedRestore: func(value int32) {
			w.record(fmt.Sprintf("queued-restore-store:%d", value))
			w.queuedRestore = value
		},
		markQuestReady: func(value int32) {
			w.record(fmt.Sprintf("quest-ready:%d", value))
		},
		finishQuestTransition: func() {
			w.record("quest-finish")
		},
		fadeBegin: func(a1, a2 int32) {
			w.record(fmt.Sprintf("fade:%d:%d", a1, a2))
		},
		loadDataRoot: func() string {
			root := w.dataRoot
			w.record("data-root:" + root)
			return root
		},
		formatTempSavePath: func(root string) string {
			path := w.tempPath
			w.record(fmt.Sprintf("temp-path:%s:%s", root, path))
			return path
		},
		loadDeleteTempFile: func() func(string) {
			w.record("delete-load")
			return func(path string) {
				w.record("delete:" + path)
				if w.onDelete != nil {
					w.onDelete(path)
				}
			}
		},
		loadUpdateData: func(unit string) string {
			data := w.dataByUnit[unit]
			w.record(fmt.Sprintf("update:%s:%s", unit, data))
			return data
		},
		loadPlayer: func(data string) string {
			player := w.playerByData[data]
			w.record(fmt.Sprintf("player:%s:%s", data, player))
			return player
		},
		loadPlayerField4792: func(player string) int32 {
			value := w.playerField4792[player]
			w.record(fmt.Sprintf("field4792:%s:%d", player, value))
			return value
		},
		loadUpdateField138: func(data string) int32 {
			value := w.updateField138[data]
			w.record(fmt.Sprintf("field138:%s:%d", data, value))
			return value
		},
		loadPlayerIndex: func(player string) uint8 {
			value := w.playerIndex[player]
			w.record(fmt.Sprintf("index:%s:%d", player, value))
			return value
		},
		loadPlayerField3680: func(player string) uint8 {
			value := w.playerField3680[player]
			w.record(fmt.Sprintf("field3680:%s:%d", player, value))
			return value
		},
		savePlayerData: func(path string, index uint8) int32 {
			value := w.saveResult
			w.record(fmt.Sprintf("save:%s:%d:%d", path, index, value))
			if w.onSave != nil {
				w.onSave(path, index)
			}
			return value
		},
		preparePlayerData: func(index uint8) int32 {
			value := w.prepareResult
			w.record(fmt.Sprintf("prepare:%d:%d", index, value))
			if w.onPrepare != nil {
				w.onPrepare(index)
			}
			return value
		},
		sendGauntlet: func(index uint8, state int32) {
			w.record(fmt.Sprintf("gauntlet:%d:%d", index, state))
			if w.onGauntlet != nil {
				w.onGauntlet(index, state)
			}
		},
		restorePlayerData: func(path string, index uint8) int32 {
			value := w.restoreResult
			w.record(fmt.Sprintf("restore:%s:%d:%d", path, index, value))
			if w.onRestore != nil {
				w.onRestore(path, index)
			}
			return value
		},
		finishPlayerData: func(index uint8) {
			w.record(fmt.Sprintf("finish-player:%d", index))
			if w.onFinish != nil {
				w.onFinish(index)
			}
		},
		applyEnchant: func(unit string, enchant, strength, duration int32) {
			w.record(fmt.Sprintf("enchant:%s:%d:%d:%d", unit, enchant, strength, duration))
			if w.onEnchant != nil {
				w.onEnchant(unit)
			}
		},
	}
}

func TestMapTransitionPlayerInit4FC6D0ExactStateGates(t *testing.T) {
	for _, tc := range []struct {
		name     string
		mapInit  int32
		mapEntry int32
		want     []string
	}{
		{name: "both zero", want: []string{"map-init:0", "map-entry:0"}},
		{name: "noncanonical states", mapInit: -1, mapEntry: 2, want: []string{"map-init:-1", "map-entry:2"}},
		{name: "map initialize exact one short circuits entry", mapInit: 1, mapEntry: -1, want: []string{"map-init:1", "first1:"}},
		{name: "map entry exact one", mapInit: 2, mapEntry: 1, want: []string{"map-init:2", "map-entry:1", "first1:"}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			w := &mapTransitionPlayerInitTestWorld4FC6D0{
				mapInitState:  tc.mapInit,
				mapEntryState: tc.mapEntry,
			}
			mapTransitionPlayerInit4FC6D0(w.hooks())
			if !reflect.DeepEqual(w.events, tc.want) {
				t.Fatalf("events = %#v, want %#v", w.events, tc.want)
			}
		})
	}
}

func TestMapTransitionPlayerInit4FC6D0QuestBranchOrder(t *testing.T) {
	for _, tc := range []struct {
		name             string
		flags            map[int32]int32
		questStage       int32
		restorePredicate int32
		restoreReady     int32
		queuedRestore    int32
		firstResults     []string
		want             []string
	}{
		{
			name:         "non quest fades then checks online",
			flags:        map[int32]int32{},
			firstResults: []string{"gate"},
			want: []string{
				"map-init:1", "first1:gate", "flag:0x1000:0", "fade:1:1", "flag:0x2000:0",
			},
		},
		{
			name:         "quest stage exact one",
			flags:        map[int32]int32{mapTransitionQuestFlag4FC6D0: -1},
			questStage:   1,
			firstResults: []string{"gate"},
			want: []string{
				"map-init:1", "first1:gate", "flag:0x1000:-1", "quest-stage:1",
				"quest-stage-send:255", "quest-ready:1", "quest-finish", "flag:0x2000:0",
			},
		},
		{
			name:             "second predicate exact zero selects stage path",
			flags:            map[int32]int32{mapTransitionQuestFlag4FC6D0: 1},
			questStage:       -1,
			restorePredicate: 2,
			restoreReady:     0,
			firstResults:     []string{"gate"},
			want: []string{
				"map-init:1", "first1:gate", "flag:0x1000:1", "quest-stage:-1",
				"restore-predicate:2", "restore-ready:0", "quest-stage-send:255",
				"quest-ready:1", "quest-finish", "flag:0x2000:0",
			},
		},
		{
			name:             "first predicate zero skips second and consumes queued restore",
			flags:            map[int32]int32{mapTransitionQuestFlag4FC6D0: 1},
			questStage:       2,
			restorePredicate: 0,
			restoreReady:     -1,
			queuedRestore:    1,
			firstResults:     []string{"gate"},
			want: []string{
				"map-init:1", "first1:gate", "flag:0x1000:1", "quest-stage:2",
				"restore-predicate:0", "queued-restore:1", "quest-restore-send:255:1",
				"queued-restore-store:0", "quest-ready:1", "quest-finish", "flag:0x2000:0",
			},
		},
		{
			name:             "noncanonical queued state takes temporary path even with no second unit",
			flags:            map[int32]int32{mapTransitionQuestFlag4FC6D0: 1},
			questStage:       2,
			restorePredicate: 2,
			restoreReady:     -1,
			queuedRestore:    -1,
			firstResults:     []string{"gate", ""},
			want: []string{
				"map-init:1", "first1:gate", "flag:0x1000:1", "quest-stage:2",
				"restore-predicate:2", "restore-ready:-1", "queued-restore:-1",
				"data-root:/nox", "temp-path:/nox:/nox/save/_temp_.dat", "first2:",
				"quest-restore-send:255:0", "quest-ready:1", "quest-finish", "flag:0x2000:0",
			},
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			w := &mapTransitionPlayerInitTestWorld4FC6D0{
				mapInitState:     1,
				flags:            tc.flags,
				questStage:       tc.questStage,
				restorePredicate: tc.restorePredicate,
				restoreReady:     tc.restoreReady,
				queuedRestore:    tc.queuedRestore,
				firstResults:     tc.firstResults,
				dataRoot:         "/nox",
				tempPath:         "/nox/save/_temp_.dat",
			}
			mapTransitionPlayerInit4FC6D0(w.hooks())
			if !reflect.DeepEqual(w.events, tc.want) {
				t.Fatalf("events = %#v, want %#v", w.events, tc.want)
			}
			if tc.queuedRestore == 1 && w.queuedRestore != 0 {
				t.Fatalf("queued restore = %d, want zero", w.queuedRestore)
			}
		})
	}
}

func newMapTransitionDenseTempWorld4FC6D0() (*mapTransitionPlayerInitTestWorld4FC6D0, []string) {
	w := &mapTransitionPlayerInitTestWorld4FC6D0{
		mapInitState:     1,
		firstResults:     []string{"u1", "u1"},
		next:             map[string]string{"u1": "u2", "u3": ""},
		flags:            map[int32]int32{mapTransitionQuestFlag4FC6D0: 1},
		questStage:       2,
		restorePredicate: 0,
		queuedRestore:    2,
		dataRoot:         "/nox",
		tempPath:         "/nox/save/_temp_.dat",
		dataByUnit:       map[string]string{"u1": "d1", "u3": "d3"},
		playerByData:     map[string]string{"d1": "p1", "d3": "p3"},
		playerField4792:  map[string]int32{"p1": 1, "p2": 1, "p3": 0},
		updateField138:   map[string]int32{"d1": 0, "d3": 99},
		playerIndex:      map[string]uint8{"p1": 1, "p2": 2, "p3": 9},
		saveResult:       1,
	}
	w.onSave = func(_ string, _ uint8) {
		w.playerByData["d1"] = "p2"
	}
	w.onPrepare = func(_ uint8) {
		w.playerIndex["p2"] = 3
	}
	w.onGauntlet = func(_ uint8, state int32) {
		if state == 1 {
			w.playerIndex["p2"] = 4
		} else {
			w.playerIndex["p2"] = 6
		}
	}
	w.onRestore = func(_ string, _ uint8) {
		w.playerIndex["p2"] = 5
	}
	w.onDelete = func(_ string) {
		w.playerIndex["p2"] = 7
	}
	w.onFinish = func(index uint8) {
		if index == 7 {
			w.next["u1"] = "u3"
		}
	}
	want := []string{
		"map-init:1", "first1:u1", "flag:0x1000:1", "quest-stage:2",
		"restore-predicate:0", "queued-restore:2", "data-root:/nox",
		"temp-path:/nox:/nox/save/_temp_.dat", "first2:u1", "delete-load",
		"update:u1:d1", "player:d1:p1", "field4792:p1:1", "field138:d1:0",
		"player:d1:p1", "index:p1:1", "save:/nox/save/_temp_.dat:1:1",
		"player:d1:p2", "index:p2:2", "prepare:2:0",
		"player:d1:p2", "index:p2:3", "gauntlet:3:1",
		"player:d1:p2", "index:p2:4", "restore:/nox/save/_temp_.dat:4:0",
		"player:d1:p2", "index:p2:5", "gauntlet:5:0", "delete:/nox/save/_temp_.dat",
		"player:d1:p2", "index:p2:7", "finish-player:7", "next:u1:u3",
		"update:u3:d3", "player:d3:p3", "field4792:p3:0",
		"player:d3:p3", "index:p3:9", "finish-player:9", "next:u3:",
		"quest-restore-send:255:0", "quest-ready:1", "quest-finish", "flag:0x2000:0",
	}
	return w, want
}

func TestMapTransitionPlayerInit4FC6D0ReloadsPlayerAndIndexAndUsesLiveNext(t *testing.T) {
	w, want := newMapTransitionDenseTempWorld4FC6D0()
	mapTransitionPlayerInit4FC6D0(w.hooks())
	if !reflect.DeepEqual(w.events, want) {
		t.Fatalf("events =\n%#v\nwant =\n%#v", w.events, want)
	}
}

func TestMapTransitionPlayerInit4FC6D0TemporaryPathFaultPrefixes(t *testing.T) {
	_, want := newMapTransitionDenseTempWorld4FC6D0()
	for faultAt := 1; faultAt <= len(want); faultAt++ {
		t.Run(fmt.Sprintf("fault-%d", faultAt), func(t *testing.T) {
			w, _ := newMapTransitionDenseTempWorld4FC6D0()
			w.faultAt = faultAt
			defer func() {
				if got := recover(); got != want[faultAt-1] {
					t.Fatalf("panic = %v, want %q", got, want[faultAt-1])
				}
				if prefix := want[:faultAt]; !reflect.DeepEqual(w.events, prefix) {
					t.Fatalf("events = %#v, want prefix %#v", w.events, prefix)
				}
			}()
			mapTransitionPlayerInit4FC6D0(w.hooks())
		})
	}
}

func countMapTransitionEventsWithPrefix4FC6D0(events []string, prefix string) int {
	count := 0
	for _, event := range events {
		if strings.HasPrefix(event, prefix) {
			count++
		}
	}
	return count
}

func TestMapTransitionPlayerInit4FC6D0TemporaryRestoreConditions(t *testing.T) {
	for _, tc := range []struct {
		name             string
		field4792        int32
		field138         int32
		saveResult       int32
		prepareResult    int32
		restoreResult    int32
		wantField138Read bool
		wantSave         bool
		wantRestore      bool
		wantRollback     bool
		wantDelete       bool
	}{
		{name: "field 4792 must be exact one", field4792: 2},
		{name: "field 138 must be exact zero", field4792: 1, field138: -1, wantField138Read: true},
		{name: "zero save result skips restore and delete", field4792: 1, wantField138Read: true, wantSave: true},
		{name: "nonzero prepare suppresses rollback", field4792: 1, saveResult: -1, prepareResult: 2, wantField138Read: true, wantSave: true, wantRestore: true, wantDelete: true},
		{name: "nonzero restore suppresses rollback", field4792: 1, saveResult: 2, restoreResult: -1, wantField138Read: true, wantSave: true, wantRestore: true, wantDelete: true},
		{name: "both zero results request rollback", field4792: 1, saveResult: 1, wantField138Read: true, wantSave: true, wantRestore: true, wantRollback: true, wantDelete: true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			w := &mapTransitionPlayerInitTestWorld4FC6D0{
				mapInitState:     1,
				firstResults:     []string{"gate", "u1"},
				next:             map[string]string{"u1": ""},
				flags:            map[int32]int32{mapTransitionQuestFlag4FC6D0: 1},
				questStage:       2,
				restorePredicate: 0,
				queuedRestore:    2,
				dataRoot:         "/nox",
				tempPath:         "/nox/save/_temp_.dat",
				dataByUnit:       map[string]string{"u1": "d1"},
				playerByData:     map[string]string{"d1": "p1"},
				playerField4792:  map[string]int32{"p1": tc.field4792},
				updateField138:   map[string]int32{"d1": tc.field138},
				playerIndex:      map[string]uint8{"p1": 7},
				saveResult:       tc.saveResult,
				prepareResult:    tc.prepareResult,
				restoreResult:    tc.restoreResult,
			}
			mapTransitionPlayerInit4FC6D0(w.hooks())

			checks := []struct {
				prefix string
				want   bool
			}{
				{prefix: "field138:", want: tc.wantField138Read},
				{prefix: "save:", want: tc.wantSave},
				{prefix: "prepare:", want: tc.wantRestore},
				{prefix: "restore:", want: tc.wantRestore},
				{prefix: "gauntlet:7:0", want: tc.wantRollback},
				{prefix: "delete:", want: tc.wantDelete},
			}
			for _, check := range checks {
				if got := countMapTransitionEventsWithPrefix4FC6D0(w.events, check.prefix) != 0; got != check.want {
					t.Fatalf("event prefix %q present = %t, want %t; events = %#v", check.prefix, got, check.want, w.events)
				}
			}
			if got := countMapTransitionEventsWithPrefix4FC6D0(w.events, "finish-player:7"); got != 1 {
				t.Fatalf("finish count = %d, want one; events = %#v", got, w.events)
			}
		})
	}
}

func TestMapTransitionPlayerInit4FC6D0OnlineTraversalAndHostShortCircuit(t *testing.T) {
	w := &mapTransitionPlayerInitTestWorld4FC6D0{
		mapInitState: 1,
		firstResults: []string{"gate", "host"},
		next:         map[string]string{"host": "guest", "guest": "", "late": ""},
		flags: map[int32]int32{
			mapTransitionOnlineFlag4FC6D0: -1,
		},
		dataByUnit:      map[string]string{"host": "dh", "guest": "dg", "late": "dl"},
		playerByData:    map[string]string{"dh": "ph", "dg": "pg", "dl": "pl"},
		playerIndex:     map[string]uint8{"ph": 31, "pg": 2, "pl": 3},
		playerField3680: map[string]uint8{"ph": 0, "pg": 2, "pl": 1},
	}
	w.onEnchant = func(unit string) {
		if unit == "guest" {
			w.next["guest"] = "late"
		}
	}
	want := []string{
		"map-init:1", "first1:gate", "flag:0x1000:0", "fade:1:1",
		"flag:0x2000:-1", "flag:0x80:0", "first2:host",
		"update:host:dh", "player:dh:ph", "index:ph:31", "next:host:guest",
		"update:guest:dg", "player:dg:pg", "index:pg:2", "field3680:pg:2",
		"enchant:guest:23:0:5", "next:guest:late",
		"update:late:dl", "player:dl:pl", "index:pl:3", "field3680:pl:1", "next:late:",
	}
	mapTransitionPlayerInit4FC6D0(w.hooks())
	if !reflect.DeepEqual(w.events, want) {
		t.Fatalf("events = %#v, want %#v", w.events, want)
	}
}

func TestMapTransitionPlayerInit4FC6D0OnlineFlagShortCircuit(t *testing.T) {
	for _, tc := range []struct {
		name     string
		online   int32
		chat     int32
		wantTail []string
	}{
		{name: "offline skips chat", wantTail: []string{"flag:0x2000:0"}},
		{name: "chat nonzero skips traversal", online: 2, chat: -1, wantTail: []string{"flag:0x2000:2", "flag:0x80:-1"}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			w := &mapTransitionPlayerInitTestWorld4FC6D0{
				mapInitState: 1,
				firstResults: []string{"gate", "must-not-be-read"},
				flags: map[int32]int32{
					mapTransitionOnlineFlag4FC6D0: tc.online,
					mapTransitionChatFlag4FC6D0:   tc.chat,
				},
			}
			mapTransitionPlayerInit4FC6D0(w.hooks())
			want := append([]string{"map-init:1", "first1:gate", "flag:0x1000:0", "fade:1:1"}, tc.wantTail...)
			if !reflect.DeepEqual(w.events, want) {
				t.Fatalf("events = %#v, want %#v", w.events, want)
			}
		})
	}
}

func TestMapTransitionPlayerInit4FC6D0QueuedAndOnlineFaultPrefixes(t *testing.T) {
	want := []string{
		"map-init:1", "first1:u1", "flag:0x1000:1", "quest-stage:2",
		"restore-predicate:0", "queued-restore:1", "quest-restore-send:255:1",
		"queued-restore-store:0", "quest-ready:1", "quest-finish",
		"flag:0x2000:1", "flag:0x80:0", "first2:u1",
		"update:u1:d1", "player:d1:p1", "index:p1:2", "field3680:p1:0",
		"enchant:u1:23:0:5", "next:u1:",
	}
	for faultAt := 1; faultAt <= len(want); faultAt++ {
		t.Run(fmt.Sprintf("fault-%d", faultAt), func(t *testing.T) {
			w := &mapTransitionPlayerInitTestWorld4FC6D0{
				faultAt:         faultAt,
				mapInitState:    1,
				firstResults:    []string{"u1", "u1"},
				next:            map[string]string{"u1": ""},
				flags:           map[int32]int32{mapTransitionQuestFlag4FC6D0: 1, mapTransitionOnlineFlag4FC6D0: 1},
				questStage:      2,
				queuedRestore:   1,
				dataByUnit:      map[string]string{"u1": "d1"},
				playerByData:    map[string]string{"d1": "p1"},
				playerIndex:     map[string]uint8{"p1": 2},
				playerField3680: map[string]uint8{"p1": 0},
			}
			defer func() {
				if got := recover(); got != want[faultAt-1] {
					t.Fatalf("panic = %v, want %q", got, want[faultAt-1])
				}
				if prefix := want[:faultAt]; !reflect.DeepEqual(w.events, prefix) {
					t.Fatalf("events = %#v, want prefix %#v", w.events, prefix)
				}
				if faultAt <= 8 {
					if w.queuedRestore != 1 {
						t.Fatalf("queued restore = %d, want one before store returns", w.queuedRestore)
					}
				} else if w.queuedRestore != 0 {
					t.Fatalf("queued restore = %d, want zero after store returns", w.queuedRestore)
				}
			}()
			mapTransitionPlayerInit4FC6D0(w.hooks())
		})
	}
}
