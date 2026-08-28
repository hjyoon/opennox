package server

import (
	"fmt"
	"math"
	"reflect"
	"testing"
)

type audEventPickupTestWorld4F3D50 struct {
	ownerArg     string
	itemArg      string
	arg3         int32
	arg4         int32
	defaultRet   int32
	typeInd      map[string]uint16
	rows         []audEventPickupSoundRow4F3D50
	events       []string
	faultAt      int
	afterDefault func(*audEventPickupTestWorld4F3D50)
	afterRowType func(*audEventPickupTestWorld4F3D50, int)
	afterTypeInd func(*audEventPickupTestWorld4F3D50, string)
}

func newAudEventPickupTestWorld4F3D50() *audEventPickupTestWorld4F3D50 {
	return &audEventPickupTestWorld4F3D50{
		ownerArg:   "owner-a",
		itemArg:    "item-a",
		arg3:       31,
		arg4:       47,
		defaultRet: 1,
		typeInd:    map[string]uint16{"item-a": 7},
		rows: []audEventPickupSoundRow4F3D50{
			{typeInd: 7, sound: 321},
			{typeInd: audEventPickupSentinel4F3D50},
		},
	}
}

func (w *audEventPickupTestWorld4F3D50) event(value string) {
	w.events = append(w.events, value)
	if w.faultAt != 0 && len(w.events) == w.faultAt {
		panic(value)
	}
}

func (w *audEventPickupTestWorld4F3D50) hooks() audEventPickupHooks4F3D50[string] {
	return audEventPickupHooks4F3D50[string]{
		loadOwnerArg: func() string {
			w.event("owner-arg:" + w.ownerArg)
			return w.ownerArg
		},
		loadItemArg: func() string {
			w.event("item-arg:" + w.itemArg)
			return w.itemArg
		},
		loadArg4: func() int32 {
			w.event(fmt.Sprintf("arg4:%d", w.arg4))
			return w.arg4
		},
		loadArg3: func() int32 {
			w.event(fmt.Sprintf("arg3:%d", w.arg3))
			return w.arg3
		},
		defaultPickup: func(owner, item string, arg3, arg4 int32) int32 {
			w.event(fmt.Sprintf("default:%s:%s:%d:%d", owner, item, arg3, arg4))
			result := w.defaultRet
			if w.afterDefault != nil {
				w.afterDefault(w)
			}
			return result
		},
		loadRowType: func(row int) uint16 {
			value := w.rows[row].typeInd
			w.event(fmt.Sprintf("row-type:%d=%04x", row, value))
			if w.afterRowType != nil {
				w.afterRowType(w, row)
			}
			return value
		},
		loadTypeInd: func(item string) uint16 {
			value := w.typeInd[item]
			w.event(fmt.Sprintf("type:%s=%04x", item, value))
			if w.afterTypeInd != nil {
				w.afterTypeInd(w, item)
			}
			return value
		},
		loadRowSound: func(row int) uint16 {
			value := w.rows[row].sound
			w.event(fmt.Sprintf("row-sound:%d=%04x", row, value))
			return value
		},
		audio: func(sound uint32, owner string, kind int32, code uint32) {
			w.event(fmt.Sprintf("audio:%d:%s:%d:%08x", sound, owner, kind, code))
		},
	}
}

func verifyAudEventPickupFaultPrefixes4F3D50(t *testing.T, want []string, build func() *audEventPickupTestWorld4F3D50) {
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
			audEventPickup4F3D50(w.hooks())
		})
	}
}

func TestAudEventPickup4F3D50NilGatesAndScalarLoadOrder(t *testing.T) {
	tests := []struct {
		name  string
		build func(*audEventPickupTestWorld4F3D50)
		want  []string
	}{
		{name: "owner", build: func(w *audEventPickupTestWorld4F3D50) { w.ownerArg = "" }, want: []string{"owner-arg:"}},
		{name: "item", build: func(w *audEventPickupTestWorld4F3D50) { w.itemArg = "" }, want: []string{"owner-arg:owner-a", "item-arg:"}},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			w := newAudEventPickupTestWorld4F3D50()
			tc.build(w)
			if got := audEventPickup4F3D50(w.hooks()); got != 0 {
				t.Fatalf("result = %d, want 0", got)
			}
			if !reflect.DeepEqual(w.events, tc.want) {
				t.Fatalf("events = %v, want %v", w.events, tc.want)
			}
		})
	}

	w := newAudEventPickupTestWorld4F3D50()
	w.defaultRet = 0
	if got := audEventPickup4F3D50(w.hooks()); got != 0 {
		t.Fatalf("result = %d, want 0", got)
	}
	want := []string{
		"owner-arg:owner-a",
		"item-arg:item-a",
		"arg4:47",
		"arg3:31",
		"default:owner-a:item-a:31:47",
	}
	if !reflect.DeepEqual(w.events, want) {
		t.Fatalf("events = %v, want %v", w.events, want)
	}
}

func TestAudEventPickup4F3D50EmptyTableSkipsTypeIndAndPreservesResult(t *testing.T) {
	w := newAudEventPickupTestWorld4F3D50()
	w.defaultRet = math.MinInt32
	w.rows = []audEventPickupSoundRow4F3D50{{typeInd: audEventPickupSentinel4F3D50}}
	if got := audEventPickup4F3D50(w.hooks()); got != math.MinInt32 {
		t.Fatalf("result = %d, want %d", got, int32(math.MinInt32))
	}
	if got := w.events[len(w.events)-1]; got != "row-type:0=ffff" {
		t.Fatalf("last event = %q; all = %v", got, w.events)
	}
	for _, event := range w.events {
		if event == "type:item-a=0007" {
			t.Fatalf("empty table read TypeInd: %v", w.events)
		}
	}
}

func TestAudEventPickup4F3D50FirstMatchCallsZeroSoundAndReturnsExact(t *testing.T) {
	build := func() *audEventPickupTestWorld4F3D50 {
		w := newAudEventPickupTestWorld4F3D50()
		w.defaultRet = -17
		w.rows = []audEventPickupSoundRow4F3D50{
			{typeInd: 7, sound: 0},
			{typeInd: 7, sound: 999},
			{typeInd: audEventPickupSentinel4F3D50},
		}
		return w
	}
	want := []string{
		"owner-arg:owner-a",
		"item-arg:item-a",
		"arg4:47",
		"arg3:31",
		"default:owner-a:item-a:31:47",
		"row-type:0=0007",
		"type:item-a=0007",
		"row-sound:0=0000",
		"audio:0:owner-a:0:00000000",
	}
	w := build()
	if got := audEventPickup4F3D50(w.hooks()); got != -17 {
		t.Fatalf("result = %d, want -17", got)
	}
	if !reflect.DeepEqual(w.events, want) {
		t.Fatalf("events = %v, want %v", w.events, want)
	}
	verifyAudEventPickupFaultPrefixes4F3D50(t, want, build)
}

func TestAudEventPickup4F3D50CachesArgumentsAndTypeInd(t *testing.T) {
	w := newAudEventPickupTestWorld4F3D50()
	w.rows = []audEventPickupSoundRow4F3D50{
		{typeInd: 3, sound: 100},
		{typeInd: 7, sound: 200},
		{typeInd: audEventPickupSentinel4F3D50},
	}
	w.afterDefault = func(w *audEventPickupTestWorld4F3D50) {
		w.ownerArg = "owner-mutated"
		w.itemArg = "item-mutated"
		w.arg3 = 300
		w.arg4 = 400
		w.typeInd["item-a"] = 9
	}
	w.afterRowType = func(w *audEventPickupTestWorld4F3D50, row int) {
		if row == 0 {
			w.typeInd["item-a"] = 7
		}
	}
	w.afterTypeInd = func(w *audEventPickupTestWorld4F3D50, item string) {
		w.typeInd[item] = 3
	}
	if got := audEventPickup4F3D50(w.hooks()); got != 1 {
		t.Fatalf("result = %d, want 1", got)
	}
	want := []string{
		"owner-arg:owner-a",
		"item-arg:item-a",
		"arg4:47",
		"arg3:31",
		"default:owner-a:item-a:31:47",
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

func TestAudEventPickup4F3D50SentinelMissFaultPrefixes(t *testing.T) {
	build := func() *audEventPickupTestWorld4F3D50 {
		w := newAudEventPickupTestWorld4F3D50()
		w.defaultRet = math.MinInt32
		w.typeInd["item-a"] = 9
		return w
	}
	want := []string{
		"owner-arg:owner-a",
		"item-arg:item-a",
		"arg4:47",
		"arg3:31",
		"default:owner-a:item-a:31:47",
		"row-type:0=0007",
		"type:item-a=0009",
		"row-type:1=ffff",
	}
	w := build()
	if got := audEventPickup4F3D50(w.hooks()); got != math.MinInt32 {
		t.Fatalf("result = %d, want %d", got, int32(math.MinInt32))
	}
	if !reflect.DeepEqual(w.events, want) {
		t.Fatalf("events = %v, want %v", w.events, want)
	}
	verifyAudEventPickupFaultPrefixes4F3D50(t, want, build)
}

type audEventPickupParseTestWorld5367B0 struct {
	init       uint32
	rows       [audEventPickupRowStorage4F3D50]audEventPickupSoundRow4F3D50
	token      string
	tokenBytes map[string]byte
	sounds     map[string]uint16
	typeInd    uint16
	events     []string
	faultAt    int
}

func newAudEventPickupParseTestWorld5367B0() *audEventPickupParseTestWorld5367B0 {
	return &audEventPickupParseTestWorld5367B0{
		token:      "token-a",
		tokenBytes: map[string]byte{"token-a": 'S', "empty-token": 0},
		sounds:     map[string]uint16{"token-a": 700},
		typeInd:    12,
	}
}

func (w *audEventPickupParseTestWorld5367B0) event(value string) {
	w.events = append(w.events, value)
	if w.faultAt != 0 && len(w.events) == w.faultAt {
		panic(value)
	}
}

func (w *audEventPickupParseTestWorld5367B0) hooks() audEventPickupParseHooks5367B0[string] {
	return audEventPickupParseHooks5367B0[string]{
		loadInit: func() uint32 {
			value := w.init
			w.event(fmt.Sprintf("init=%08x", value))
			return value
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
			value := w.rows[row].typeInd
			w.event(fmt.Sprintf("row-type:%d=%04x", row, value))
			return value
		},
		nextToken: func() string {
			w.event("token:" + w.token)
			return w.token
		},
		loadTokenByte: func(token string) byte {
			value := w.tokenBytes[token]
			w.event(fmt.Sprintf("token-byte:%s=%02x", token, value))
			return value
		},
		resolveSound: func(token string) uint16 {
			value := w.sounds[token]
			w.event(fmt.Sprintf("resolve:%s=%04x", token, value))
			return value
		},
		loadTypeInd: func() uint16 {
			w.event(fmt.Sprintf("type=%04x", w.typeInd))
			return w.typeInd
		},
	}
}

func audEventPickupParseInitEvents5367B0() []string {
	events := []string{"init=00000000"}
	for row := 0; row < audEventPickupRowStorage4F3D50; row++ {
		events = append(events,
			fmt.Sprintf("store-type:%d=ffff", row),
			fmt.Sprintf("store-sound:%d=0000", row),
		)
	}
	return append(events, "store-init=00000001")
}

func verifyAudEventPickupParseFaultPrefixes5367B0(t *testing.T, want []string, build func() *audEventPickupParseTestWorld5367B0) {
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
			audEventPickupParse5367B0(w.hooks())
		})
	}
}

func TestAudEventPickupParse5367B0InitializesAndPublishesSoundBeforeType(t *testing.T) {
	build := newAudEventPickupParseTestWorld5367B0
	want := append(audEventPickupParseInitEvents5367B0(),
		"row-type:0=ffff",
		"token:token-a",
		"token-byte:token-a=53",
		"resolve:token-a=02bc",
		"type=000c",
		"store-sound:0=02bc",
		"store-type:0=000c",
	)
	w := build()
	if !audEventPickupParse5367B0(w.hooks()) {
		t.Fatal("row was not published")
	}
	if !reflect.DeepEqual(w.events, want) {
		t.Fatalf("events = %v, want %v", w.events, want)
	}
	if w.init != 1 || w.rows[0] != (audEventPickupSoundRow4F3D50{typeInd: 12, sound: 700}) {
		t.Fatalf("state = init %d row %+v", w.init, w.rows[0])
	}
	if w.rows[1] != (audEventPickupSoundRow4F3D50{typeInd: audEventPickupSentinel4F3D50}) {
		t.Fatalf("sentinel row = %+v", w.rows[1])
	}
	verifyAudEventPickupParseFaultPrefixes5367B0(t, want, build)
}

func TestAudEventPickupParse5367B0AppendsDuplicatesAndCapsAtFifty(t *testing.T) {
	w := newAudEventPickupParseTestWorld5367B0()
	w.init = math.MaxUint32
	for row := range w.rows {
		w.rows[row].typeInd = audEventPickupSentinel4F3D50
	}
	w.rows[0] = audEventPickupSoundRow4F3D50{typeInd: 12, sound: 111}
	if !audEventPickupParse5367B0(w.hooks()) {
		t.Fatal("duplicate row was not appended")
	}
	if w.rows[0].sound != 111 || w.rows[1] != (audEventPickupSoundRow4F3D50{typeInd: 12, sound: 700}) {
		t.Fatalf("rows = %+v %+v", w.rows[0], w.rows[1])
	}

	full := newAudEventPickupParseTestWorld5367B0()
	full.init = 1
	for row := 0; row < audEventPickupRowCapacity4F3D50; row++ {
		full.rows[row] = audEventPickupSoundRow4F3D50{typeInd: uint16(row), sound: uint16(row + 1)}
	}
	full.rows[audEventPickupRowCapacity4F3D50].typeInd = audEventPickupSentinel4F3D50
	if audEventPickupParse5367B0(full.hooks()) {
		t.Fatal("full table unexpectedly published a row")
	}
	if len(full.events) != 1+audEventPickupRowCapacity4F3D50 {
		t.Fatalf("event count = %d, want %d", len(full.events), 1+audEventPickupRowCapacity4F3D50)
	}
	if got := full.events[len(full.events)-1]; got != "row-type:49=0031" {
		t.Fatalf("last event = %q", got)
	}
}

func TestAudEventPickupParse5367B0TokenAndSoundGatesDelayTypeRead(t *testing.T) {
	tests := []struct {
		name  string
		build func(*audEventPickupParseTestWorld5367B0)
		last  string
	}{
		{name: "nil-token", build: func(w *audEventPickupParseTestWorld5367B0) { w.token = "" }, last: "token:"},
		{name: "empty-token", build: func(w *audEventPickupParseTestWorld5367B0) { w.token = "empty-token" }, last: "token-byte:empty-token=00"},
		{name: "unknown-sound", build: func(w *audEventPickupParseTestWorld5367B0) { w.sounds["token-a"] = 0 }, last: "resolve:token-a=0000"},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			w := newAudEventPickupParseTestWorld5367B0()
			w.init = 1
			for row := range w.rows {
				w.rows[row].typeInd = audEventPickupSentinel4F3D50
			}
			tc.build(w)
			if audEventPickupParse5367B0(w.hooks()) {
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

func TestAudEventPickupParse5367B0FaultBeforeTypePublishLeavesSentinel(t *testing.T) {
	w := newAudEventPickupParseTestWorld5367B0()
	w.init = 1
	for row := range w.rows {
		w.rows[row].typeInd = audEventPickupSentinel4F3D50
	}
	w.faultAt = 8
	defer func() {
		if got := recover(); got != "store-type:0=000c" {
			t.Fatalf("panic = %v", got)
		}
		if w.rows[0].typeInd != audEventPickupSentinel4F3D50 || w.rows[0].sound != 700 {
			t.Fatalf("partially published row = %+v", w.rows[0])
		}
	}()
	audEventPickupParse5367B0(w.hooks())
}
