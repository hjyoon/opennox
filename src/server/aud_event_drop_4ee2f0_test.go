package server

import (
	"fmt"
	"math"
	"reflect"
	"testing"
)

type audEventDropTestWorld4EE2F0 struct {
	ownerArg     string
	itemArg      string
	pointArg     string
	defaultRet   int32
	typeInd      map[string]uint16
	rows         []audEventDropSoundRow4EE2F0
	events       []string
	faultAt      int
	afterDefault func(*audEventDropTestWorld4EE2F0)
	afterRowType func(*audEventDropTestWorld4EE2F0, int)
	afterTypeInd func(*audEventDropTestWorld4EE2F0, string)
}

func newAudEventDropTestWorld4EE2F0() *audEventDropTestWorld4EE2F0 {
	return &audEventDropTestWorld4EE2F0{
		ownerArg:   "owner-a",
		itemArg:    "item-a",
		pointArg:   "point-a",
		defaultRet: 1,
		typeInd:    map[string]uint16{"item-a": 7},
		rows: []audEventDropSoundRow4EE2F0{
			{typeInd: 7, sound: 321},
			{typeInd: audEventDropSentinel4EE2F0},
		},
	}
}

func (w *audEventDropTestWorld4EE2F0) event(value string) {
	w.events = append(w.events, value)
	if w.faultAt != 0 && len(w.events) == w.faultAt {
		panic(value)
	}
}

func (w *audEventDropTestWorld4EE2F0) hooks() audEventDropHooks4EE2F0[string, string] {
	return audEventDropHooks4EE2F0[string, string]{
		loadOwnerArg: func() string {
			w.event("owner-arg:" + w.ownerArg)
			return w.ownerArg
		},
		loadItemArg: func() string {
			w.event("item-arg:" + w.itemArg)
			return w.itemArg
		},
		loadPointArg: func() string {
			w.event("point-arg:" + w.pointArg)
			return w.pointArg
		},
		defaultDrop: func(owner, item, point string) int32 {
			w.event("default:" + owner + ":" + item + ":" + point)
			ret := w.defaultRet
			if w.afterDefault != nil {
				w.afterDefault(w)
			}
			return ret
		},
		loadRowType: func(row int) uint16 {
			v := w.rows[row].typeInd
			w.event(fmt.Sprintf("row-type:%d=%04x", row, v))
			if w.afterRowType != nil {
				w.afterRowType(w, row)
			}
			return v
		},
		loadTypeInd: func(item string) uint16 {
			v := w.typeInd[item]
			w.event(fmt.Sprintf("type:%s=%04x", item, v))
			if w.afterTypeInd != nil {
				w.afterTypeInd(w, item)
			}
			return v
		},
		loadRowSound: func(row int) uint16 {
			v := w.rows[row].sound
			w.event(fmt.Sprintf("row-sound:%d=%04x", row, v))
			return v
		},
		audio: func(sound uint32, owner string, kind int32, code uint32) {
			w.event(fmt.Sprintf("audio:%d:%s:%d:%08x", sound, owner, kind, code))
		},
	}
}

func verifyAudEventDropFaultPrefixes4EE2F0(t *testing.T, want []string, build func() *audEventDropTestWorld4EE2F0) {
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
			audEventDrop4EE2F0(w.hooks())
		})
	}
}

func TestAudEventDrop4EE2F0NilGates(t *testing.T) {
	tests := []struct {
		name  string
		build func(*audEventDropTestWorld4EE2F0)
		want  []string
	}{
		{name: "owner", build: func(w *audEventDropTestWorld4EE2F0) { w.ownerArg = "" }, want: []string{"owner-arg:"}},
		{name: "item", build: func(w *audEventDropTestWorld4EE2F0) { w.itemArg = "" }, want: []string{"owner-arg:owner-a", "item-arg:"}},
		{name: "point", build: func(w *audEventDropTestWorld4EE2F0) { w.pointArg = "" }, want: []string{"owner-arg:owner-a", "item-arg:item-a", "point-arg:"}},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			w := newAudEventDropTestWorld4EE2F0()
			tc.build(w)
			if got := audEventDrop4EE2F0(w.hooks()); got != 0 {
				t.Fatalf("result = %d, want 0", got)
			}
			if !reflect.DeepEqual(w.events, tc.want) {
				t.Fatalf("events = %v, want %v", w.events, tc.want)
			}
		})
	}
}

func TestAudEventDrop4EE2F0DefaultZeroSkipsTable(t *testing.T) {
	w := newAudEventDropTestWorld4EE2F0()
	w.defaultRet = 0
	if got := audEventDrop4EE2F0(w.hooks()); got != 0 {
		t.Fatalf("result = %d, want 0", got)
	}
	want := []string{
		"owner-arg:owner-a",
		"item-arg:item-a",
		"point-arg:point-a",
		"default:owner-a:item-a:point-a",
	}
	if !reflect.DeepEqual(w.events, want) {
		t.Fatalf("events = %v, want %v", w.events, want)
	}
}

func TestAudEventDrop4EE2F0EmptyTableSkipsTypeIndAndPreservesResult(t *testing.T) {
	w := newAudEventDropTestWorld4EE2F0()
	w.defaultRet = math.MinInt32
	w.rows = []audEventDropSoundRow4EE2F0{{typeInd: audEventDropSentinel4EE2F0}}
	if got := audEventDrop4EE2F0(w.hooks()); got != math.MinInt32 {
		t.Fatalf("result = %d, want %d", got, int32(math.MinInt32))
	}
	wantLast := "row-type:0=ffff"
	if got := w.events[len(w.events)-1]; got != wantLast {
		t.Fatalf("last event = %q, want %q; all = %v", got, wantLast, w.events)
	}
	for _, event := range w.events {
		if event == "type:item-a=0007" {
			t.Fatalf("empty table read TypeInd: %v", w.events)
		}
	}
}

func TestAudEventDrop4EE2F0FirstMatchCallsZeroSoundAndReturnsExact(t *testing.T) {
	build := func() *audEventDropTestWorld4EE2F0 {
		w := newAudEventDropTestWorld4EE2F0()
		w.defaultRet = -17
		w.rows = []audEventDropSoundRow4EE2F0{
			{typeInd: 7, sound: 0},
			{typeInd: 7, sound: 999},
			{typeInd: audEventDropSentinel4EE2F0},
		}
		return w
	}
	want := []string{
		"owner-arg:owner-a",
		"item-arg:item-a",
		"point-arg:point-a",
		"default:owner-a:item-a:point-a",
		"row-type:0=0007",
		"type:item-a=0007",
		"row-sound:0=0000",
		"audio:0:owner-a:0:00000000",
	}
	w := build()
	if got := audEventDrop4EE2F0(w.hooks()); got != -17 {
		t.Fatalf("result = %d, want -17", got)
	}
	if !reflect.DeepEqual(w.events, want) {
		t.Fatalf("events = %v, want %v", w.events, want)
	}
	verifyAudEventDropFaultPrefixes4EE2F0(t, want, build)
}

func TestAudEventDrop4EE2F0TypeIndLoadedOnceAfterFirstRow(t *testing.T) {
	w := newAudEventDropTestWorld4EE2F0()
	w.rows = []audEventDropSoundRow4EE2F0{
		{typeInd: 3, sound: 100},
		{typeInd: 7, sound: 200},
		{typeInd: audEventDropSentinel4EE2F0},
	}
	w.afterDefault = func(w *audEventDropTestWorld4EE2F0) {
		w.ownerArg = "owner-mutated"
		w.itemArg = "item-mutated"
		w.pointArg = "point-mutated"
		w.typeInd["item-a"] = 9
	}
	w.afterRowType = func(w *audEventDropTestWorld4EE2F0, row int) {
		if row == 0 {
			w.typeInd["item-a"] = 7
		}
	}
	w.afterTypeInd = func(w *audEventDropTestWorld4EE2F0, item string) {
		w.typeInd[item] = 3
	}
	if got := audEventDrop4EE2F0(w.hooks()); got != 1 {
		t.Fatalf("result = %d, want 1", got)
	}
	want := []string{
		"owner-arg:owner-a",
		"item-arg:item-a",
		"point-arg:point-a",
		"default:owner-a:item-a:point-a",
		"row-type:0=0003",
		"type:item-a=0007",
		"row-type:1=0007",
		"row-sound:1=00c8",
		"audio:200:owner-a:0:00000000",
	}
	if !reflect.DeepEqual(w.events, want) {
		t.Fatalf("events = %v, want %v", w.events, want)
	}
}

func TestAudEventDrop4EE2F0SentinelMissFaultPrefixes(t *testing.T) {
	build := func() *audEventDropTestWorld4EE2F0 {
		w := newAudEventDropTestWorld4EE2F0()
		w.defaultRet = math.MinInt32
		w.typeInd["item-a"] = 9
		return w
	}
	want := []string{
		"owner-arg:owner-a",
		"item-arg:item-a",
		"point-arg:point-a",
		"default:owner-a:item-a:point-a",
		"row-type:0=0007",
		"type:item-a=0009",
		"row-type:1=ffff",
	}
	w := build()
	if got := audEventDrop4EE2F0(w.hooks()); got != math.MinInt32 {
		t.Fatalf("result = %d, want %d", got, int32(math.MinInt32))
	}
	if !reflect.DeepEqual(w.events, want) {
		t.Fatalf("events = %v, want %v", w.events, want)
	}
	verifyAudEventDropFaultPrefixes4EE2F0(t, want, build)
}

type audEventDropParseTestWorld536AC0 struct {
	init       uint32
	rows       [audEventDropRowStorage4EE2F0]audEventDropSoundRow4EE2F0
	token      string
	tokenBytes map[string]byte
	sounds     map[string]uint16
	typeInd    uint16
	events     []string
	faultAt    int
}

func newAudEventDropParseTestWorld536AC0() *audEventDropParseTestWorld536AC0 {
	return &audEventDropParseTestWorld536AC0{
		token:      "token-a",
		tokenBytes: map[string]byte{"token-a": 'S', "empty-token": 0},
		sounds:     map[string]uint16{"token-a": 700},
		typeInd:    12,
	}
}

func (w *audEventDropParseTestWorld536AC0) event(value string) {
	w.events = append(w.events, value)
	if w.faultAt != 0 && len(w.events) == w.faultAt {
		panic(value)
	}
}

func (w *audEventDropParseTestWorld536AC0) hooks() audEventDropParseHooks536AC0[string] {
	return audEventDropParseHooks536AC0[string]{
		loadInit: func() uint32 {
			v := w.init
			w.event(fmt.Sprintf("init=%08x", v))
			return v
		},
		storeRowType: func(row int, value uint16) {
			w.event(fmt.Sprintf("store-type:%d=%04x", row, value))
			w.rows[row].typeInd = value
		},
		storeRowSound: func(row int, value uint16) {
			w.event(fmt.Sprintf("store-sound:%d=%04x", row, value))
			w.rows[row].sound = value
		},
		storeInit: func(value uint32) {
			w.event(fmt.Sprintf("store-init=%08x", value))
			w.init = value
		},
		loadRowType: func(row int) uint16 {
			v := w.rows[row].typeInd
			w.event(fmt.Sprintf("row-type:%d=%04x", row, v))
			return v
		},
		nextToken: func() string {
			v := w.token
			w.event("token:" + v)
			return v
		},
		loadTokenByte: func(token string) byte {
			v := w.tokenBytes[token]
			w.event(fmt.Sprintf("token-byte:%s=%02x", token, v))
			return v
		},
		resolveSound: func(token string) uint16 {
			v := w.sounds[token]
			w.event(fmt.Sprintf("resolve:%s=%04x", token, v))
			return v
		},
		loadTypeInd: func() uint16 {
			v := w.typeInd
			w.event(fmt.Sprintf("type=%04x", v))
			return v
		},
	}
}

func audEventDropParseInitEvents536AC0() []string {
	events := []string{"init=00000000"}
	for row := 0; row < audEventDropRowStorage4EE2F0; row++ {
		events = append(events,
			fmt.Sprintf("store-type:%d=ffff", row),
			fmt.Sprintf("store-sound:%d=0000", row),
		)
	}
	return append(events, "store-init=00000001")
}

func verifyAudEventDropParseFaultPrefixes536AC0(t *testing.T, want []string, build func() *audEventDropParseTestWorld536AC0) {
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
			audEventDropParse536AC0(w.hooks())
		})
	}
}

func TestAudEventDropParse536AC0InitializesAndPublishesSoundBeforeType(t *testing.T) {
	build := newAudEventDropParseTestWorld536AC0
	want := append(audEventDropParseInitEvents536AC0(),
		"row-type:0=ffff",
		"token:token-a",
		"token-byte:token-a=53",
		"resolve:token-a=02bc",
		"type=000c",
		"store-sound:0=02bc",
		"store-type:0=000c",
	)
	w := build()
	if !audEventDropParse536AC0(w.hooks()) {
		t.Fatal("row was not published")
	}
	if !reflect.DeepEqual(w.events, want) {
		t.Fatalf("events = %v, want %v", w.events, want)
	}
	if w.init != 1 || w.rows[0] != (audEventDropSoundRow4EE2F0{typeInd: 12, sound: 700}) {
		t.Fatalf("state = init %d row %+v", w.init, w.rows[0])
	}
	if w.rows[1] != (audEventDropSoundRow4EE2F0{typeInd: audEventDropSentinel4EE2F0}) {
		t.Fatalf("sentinel row = %+v", w.rows[1])
	}
	verifyAudEventDropParseFaultPrefixes536AC0(t, want, build)
}

func TestAudEventDropParse536AC0SkipsInitForAnyNonzeroFlagAndAppendsDuplicates(t *testing.T) {
	w := newAudEventDropParseTestWorld536AC0()
	w.init = math.MaxUint32
	for i := range w.rows {
		w.rows[i].typeInd = audEventDropSentinel4EE2F0
	}
	w.rows[0] = audEventDropSoundRow4EE2F0{typeInd: 12, sound: 111}
	if !audEventDropParse536AC0(w.hooks()) {
		t.Fatal("duplicate row was not appended")
	}
	if w.rows[0].sound != 111 || w.rows[1] != (audEventDropSoundRow4EE2F0{typeInd: 12, sound: 700}) {
		t.Fatalf("rows = %+v %+v", w.rows[0], w.rows[1])
	}
	wantPrefix := []string{"init=ffffffff", "row-type:0=000c", "row-type:1=ffff"}
	if !reflect.DeepEqual(w.events[:len(wantPrefix)], wantPrefix) {
		t.Fatalf("prefix = %v, want %v", w.events[:len(wantPrefix)], wantPrefix)
	}
}

func TestAudEventDropParse536AC0FullTableStopsBeforeSentinelAndToken(t *testing.T) {
	build := func() *audEventDropParseTestWorld536AC0 {
		w := newAudEventDropParseTestWorld536AC0()
		w.init = 1
		for row := 0; row < audEventDropRowCapacity4EE2F0; row++ {
			w.rows[row] = audEventDropSoundRow4EE2F0{typeInd: uint16(row), sound: uint16(row + 1)}
		}
		w.rows[audEventDropRowCapacity4EE2F0].typeInd = audEventDropSentinel4EE2F0
		return w
	}
	w := build()
	if audEventDropParse536AC0(w.hooks()) {
		t.Fatal("full table unexpectedly published a row")
	}
	if len(w.events) != 1+audEventDropRowCapacity4EE2F0 {
		t.Fatalf("event count = %d, want %d", len(w.events), 1+audEventDropRowCapacity4EE2F0)
	}
	if got := w.events[len(w.events)-1]; got != "row-type:49=0031" {
		t.Fatalf("last event = %q", got)
	}
	for _, event := range w.events {
		if event == "token:token-a" || event == "row-type:50=ffff" {
			t.Fatalf("full table read forbidden state: %v", w.events)
		}
	}
}

func TestAudEventDropParse536AC0TokenAndSoundGatesDelayTypeRead(t *testing.T) {
	tests := []struct {
		name  string
		build func(*audEventDropParseTestWorld536AC0)
		last  string
	}{
		{name: "nil-token", build: func(w *audEventDropParseTestWorld536AC0) { w.token = "" }, last: "token:"},
		{name: "empty-token", build: func(w *audEventDropParseTestWorld536AC0) { w.token = "empty-token" }, last: "token-byte:empty-token=00"},
		{name: "unknown-sound", build: func(w *audEventDropParseTestWorld536AC0) { w.sounds["token-a"] = 0 }, last: "resolve:token-a=0000"},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			w := newAudEventDropParseTestWorld536AC0()
			w.init = 1
			for row := range w.rows {
				w.rows[row].typeInd = audEventDropSentinel4EE2F0
			}
			tc.build(w)
			if audEventDropParse536AC0(w.hooks()) {
				t.Fatal("invalid input unexpectedly published a row")
			}
			if got := w.events[len(w.events)-1]; got != tc.last {
				t.Fatalf("last event = %q, want %q; all = %v", got, tc.last, w.events)
			}
			for _, event := range w.events {
				if event == "type=000c" {
					t.Fatalf("gate read type: %v", w.events)
				}
			}
		})
	}
}

func TestAudEventDropParse536AC0FaultBeforeTypePublishLeavesSentinel(t *testing.T) {
	w := newAudEventDropParseTestWorld536AC0()
	w.init = 1
	for row := range w.rows {
		w.rows[row].typeInd = audEventDropSentinel4EE2F0
	}
	w.faultAt = 8
	defer func() {
		if got := recover(); got != "store-type:0=000c" {
			t.Fatalf("panic = %v", got)
		}
		if w.rows[0].typeInd != audEventDropSentinel4EE2F0 || w.rows[0].sound != 700 {
			t.Fatalf("partially published row = %+v", w.rows[0])
		}
	}()
	audEventDropParse536AC0(w.hooks())
}
