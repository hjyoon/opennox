package server

import (
	"fmt"
	"math"
	"reflect"
	"testing"
)

type activatePoisonTestHealth4EE7E0 struct {
	name  string
	frame uint32
}

type activatePoisonTestPlayer4EE7E0 struct {
	name  string
	flags uint32
}

type activatePoisonTestUpdate4EE7E0 struct {
	name   string
	player *activatePoisonTestPlayer4EE7E0
}

type activatePoisonTestObject4EE7E0 struct {
	name     string
	current  uint8
	flags    uint8
	class    uint32
	subclass uint32
	update   *activatePoisonTestUpdate4EE7E0
	health   *activatePoisonTestHealth4EE7E0
}

type activatePoisonTestWorld4EE7E0 struct {
	unit       *activatePoisonTestObject4EE7E0
	increment  int32
	maximum    int32
	buffResult int32
	protection float64
	roll       int32
	frame      uint32
	events     []string
	faultAt    int

	afterArg       func(*activatePoisonTestWorld4EE7E0, *activatePoisonTestObject4EE7E0)
	afterCurrent   func(*activatePoisonTestWorld4EE7E0, *activatePoisonTestObject4EE7E0, uint8)
	afterClass     func(*activatePoisonTestWorld4EE7E0, *activatePoisonTestObject4EE7E0, uint32)
	afterIncrement func(*activatePoisonTestWorld4EE7E0, int32)
	afterMaximum   func(*activatePoisonTestWorld4EE7E0, int32)
	afterSet       func(*activatePoisonTestWorld4EE7E0, *activatePoisonTestObject4EE7E0, int32)
	afterAudio     func(*activatePoisonTestWorld4EE7E0, *activatePoisonTestObject4EE7E0)
	afterHealth    func(*activatePoisonTestWorld4EE7E0, *activatePoisonTestHealth4EE7E0)
	afterFrame     func(*activatePoisonTestWorld4EE7E0, uint32)
}

func activatePoisonObjectName4EE7E0(obj *activatePoisonTestObject4EE7E0) string {
	if obj == nil {
		return "nil"
	}
	return obj.name
}

func activatePoisonUpdateName4EE7E0(update *activatePoisonTestUpdate4EE7E0) string {
	if update == nil {
		return "nil"
	}
	return update.name
}

func activatePoisonPlayerName4EE7E0(player *activatePoisonTestPlayer4EE7E0) string {
	if player == nil {
		return "nil"
	}
	return player.name
}

func activatePoisonHealthName4EE7E0(health *activatePoisonTestHealth4EE7E0) string {
	if health == nil {
		return "nil"
	}
	return health.name
}

func (w *activatePoisonTestWorld4EE7E0) record(format string, args ...any) {
	event := fmt.Sprintf(format, args...)
	w.events = append(w.events, event)
	if w.faultAt != 0 && len(w.events) == w.faultAt {
		panic(event)
	}
}

func (w *activatePoisonTestWorld4EE7E0) hooks() activatePoisonHooks4EE7E0[
	*activatePoisonTestObject4EE7E0,
	*activatePoisonTestUpdate4EE7E0,
	*activatePoisonTestPlayer4EE7E0,
	*activatePoisonTestHealth4EE7E0,
] {
	return activatePoisonHooks4EE7E0[
		*activatePoisonTestObject4EE7E0,
		*activatePoisonTestUpdate4EE7E0,
		*activatePoisonTestPlayer4EE7E0,
		*activatePoisonTestHealth4EE7E0,
	]{
		loadUnitArg: func() *activatePoisonTestObject4EE7E0 {
			unit := w.unit
			w.record("arg:%s", activatePoisonObjectName4EE7E0(unit))
			if w.afterArg != nil {
				w.afterArg(w, unit)
			}
			return unit
		},
		loadCurrent: func(unit *activatePoisonTestObject4EE7E0) uint8 {
			w.record("current:%s", activatePoisonObjectName4EE7E0(unit))
			if unit == nil {
				panic("current:nil")
			}
			current := unit.current
			if w.afterCurrent != nil {
				w.afterCurrent(w, unit, current)
			}
			return current
		},
		loadFlagsLow: func(unit *activatePoisonTestObject4EE7E0) uint8 {
			w.record("flags:%s", activatePoisonObjectName4EE7E0(unit))
			return unit.flags
		},
		testBuff: func(unit *activatePoisonTestObject4EE7E0, enchant uint32) int32 {
			w.record("buff:%s:%d", activatePoisonObjectName4EE7E0(unit), enchant)
			return w.buffResult
		},
		loadClass: func(unit *activatePoisonTestObject4EE7E0) uint32 {
			class := unit.class
			w.record("class:%s", activatePoisonObjectName4EE7E0(unit))
			if w.afterClass != nil {
				w.afterClass(w, unit, class)
			}
			return class
		},
		loadUpdateData: func(unit *activatePoisonTestObject4EE7E0) *activatePoisonTestUpdate4EE7E0 {
			w.record("update:%s", activatePoisonObjectName4EE7E0(unit))
			return unit.update
		},
		loadPlayer: func(update *activatePoisonTestUpdate4EE7E0) *activatePoisonTestPlayer4EE7E0 {
			w.record("player:%s", activatePoisonUpdateName4EE7E0(update))
			if update == nil {
				panic("player:nil-update")
			}
			return update.player
		},
		loadPlayerFlags: func(player *activatePoisonTestPlayer4EE7E0) uint32 {
			w.record("player-flags:%s", activatePoisonPlayerName4EE7E0(player))
			if player == nil {
				panic("flags:nil-player")
			}
			return player.flags
		},
		loadSubClass: func(unit *activatePoisonTestObject4EE7E0) uint32 {
			w.record("subclass:%s", activatePoisonObjectName4EE7E0(unit))
			return unit.subclass
		},
		poisonProtection: func(unit *activatePoisonTestObject4EE7E0) float64 {
			w.record("protection:%s", activatePoisonObjectName4EE7E0(unit))
			return w.protection
		},
		floatToInt: func(value float32) int32 {
			w.record("round:%08x", math.Float32bits(value))
			return activatePoisonRound4EE7E0(value)
		},
		randomInt: func(minimum, maximum int32, path string, line int32) int32 {
			w.record("random:%d:%d:%d:%s", minimum, maximum, line, path)
			return w.roll
		},
		priorityMessage: func(unit *activatePoisonTestObject4EE7E0, message string, value uint8) {
			w.record("message:%s:%s:%d", activatePoisonObjectName4EE7E0(unit), message, value)
		},
		loadIncrementArg: func() int32 {
			increment := w.increment
			w.record("increment:%d", increment)
			if w.afterIncrement != nil {
				w.afterIncrement(w, increment)
			}
			return increment
		},
		loadMaximumArg: func() int32 {
			maximum := w.maximum
			w.record("maximum:%d", maximum)
			if w.afterMaximum != nil {
				w.afterMaximum(w, maximum)
			}
			return maximum
		},
		setPoison: func(unit *activatePoisonTestObject4EE7E0, value int32) {
			w.record("set:%s:%d", activatePoisonObjectName4EE7E0(unit), value)
			unit.current = uint8(value)
			if w.afterSet != nil {
				w.afterSet(w, unit, value)
			}
		},
		audio: func(id uint32, unit *activatePoisonTestObject4EE7E0, kind int32, code uint32) {
			w.record("audio:%d:%s:%d:%d", id, activatePoisonObjectName4EE7E0(unit), kind, code)
			if w.afterAudio != nil {
				w.afterAudio(w, unit)
			}
		},
		loadHealth: func(unit *activatePoisonTestObject4EE7E0) *activatePoisonTestHealth4EE7E0 {
			health := unit.health
			w.record("health:%s=%s", activatePoisonObjectName4EE7E0(unit), activatePoisonHealthName4EE7E0(health))
			if w.afterHealth != nil {
				w.afterHealth(w, health)
			}
			return health
		},
		frame: func() uint32 {
			frame := w.frame
			w.record("frame:%d", frame)
			if w.afterFrame != nil {
				w.afterFrame(w, frame)
			}
			return frame
		},
		storePoisonFrame: func(health *activatePoisonTestHealth4EE7E0, frame uint32) {
			w.record("store-frame:%s:%d", activatePoisonHealthName4EE7E0(health), frame)
			health.frame = frame
		},
	}
}

func newActivatePoisonTestWorld4EE7E0() *activatePoisonTestWorld4EE7E0 {
	player := &activatePoisonTestPlayer4EE7E0{name: "player"}
	update := &activatePoisonTestUpdate4EE7E0{name: "update", player: player}
	health := &activatePoisonTestHealth4EE7E0{name: "health", frame: 11}
	return &activatePoisonTestWorld4EE7E0{
		unit: &activatePoisonTestObject4EE7E0{
			name:   "unit",
			class:  uint32(activatePoisonPlayerClassLow4EE7E0 | activatePoisonMonsterClassLow4EE7E0),
			update: update,
			health: health,
		},
		increment:  2,
		maximum:    3,
		protection: 0.125,
		roll:       100,
		frame:      77,
	}
}

func activatePoisonFullEvents4EE7E0() []string {
	return []string{
		"arg:unit",
		"current:unit",
		"flags:unit",
		"buff:unit:23",
		"class:unit",
		"update:unit",
		"player:update",
		"player-flags:player",
		"subclass:unit",
		"protection:unit",
		"round:41480000",
		"random:0:100:361:C:\\NoxPost\\src\\Server\\Object\\health.c",
		"increment:2",
		"maximum:3",
		"set:unit:2",
		"audio:100:unit:0:0",
		"health:unit=health",
		"frame:77",
		"store-frame:health:77",
	}
}

func TestActivatePoison4EE7E0AllFaultPrefixes(t *testing.T) {
	want := activatePoisonFullEvents4EE7E0()
	for faultAt := 1; faultAt <= len(want); faultAt++ {
		t.Run(fmt.Sprintf("event-%02d", faultAt), func(t *testing.T) {
			w := newActivatePoisonTestWorld4EE7E0()
			w.faultAt = faultAt
			defer func() {
				if got := recover(); got != want[faultAt-1] {
					t.Fatalf("panic = %v, want %q", got, want[faultAt-1])
				}
				if prefix := want[:faultAt]; !reflect.DeepEqual(w.events, prefix) {
					t.Fatalf("events =\n%q\nwant\n%q", w.events, prefix)
				}
			}()
			activatePoison4EE7E0(w.hooks())
		})
	}
}

func TestActivatePoison4EE7E0SuccessfulOrder(t *testing.T) {
	w := newActivatePoisonTestWorld4EE7E0()
	if got := activatePoison4EE7E0(w.hooks()); got != 1 {
		t.Fatalf("result = %d, want 1", got)
	}
	if want := activatePoisonFullEvents4EE7E0(); !reflect.DeepEqual(w.events, want) {
		t.Fatalf("events =\n%q\nwant\n%q", w.events, want)
	}
	if w.unit.current != 2 || w.unit.health.frame != 77 {
		t.Fatalf("state = poison %d, frame %d; want 2, 77", w.unit.current, w.unit.health.frame)
	}
}

func TestActivatePoison4EE7E0NilUnitFaultsBeforeNominalGate(t *testing.T) {
	w := newActivatePoisonTestWorld4EE7E0()
	w.unit = nil
	defer func() {
		if got := recover(); got != "current:nil" {
			t.Fatalf("panic = %v, want current:nil", got)
		}
		want := []string{"arg:nil", "current:nil"}
		if !reflect.DeepEqual(w.events, want) {
			t.Fatalf("events = %q, want %q", w.events, want)
		}
	}()
	activatePoison4EE7E0(w.hooks())
}

func TestActivatePoison4EE7E0EntryGates(t *testing.T) {
	tests := []struct {
		name string
		edit func(*activatePoisonTestWorld4EE7E0)
		want []string
	}{
		{
			name: "destroyed",
			edit: func(w *activatePoisonTestWorld4EE7E0) {
				w.unit.flags = activatePoisonDestroyedFlagLow4EE7E0
			},
			want: []string{"arg:unit", "current:unit", "flags:unit"},
		},
		{
			name: "exact blocking buff",
			edit: func(w *activatePoisonTestWorld4EE7E0) {
				w.buffResult = 1
			},
			want: []string{"arg:unit", "current:unit", "flags:unit", "buff:unit:23"},
		},
		{
			name: "observer player",
			edit: func(w *activatePoisonTestWorld4EE7E0) {
				w.unit.class = uint32(activatePoisonPlayerClassLow4EE7E0)
				w.unit.update.player.flags = activatePoisonObserverFlag4EE7E0
			},
			want: []string{
				"arg:unit", "current:unit", "flags:unit", "buff:unit:23", "class:unit",
				"update:unit", "player:update", "player-flags:player",
			},
		},
		{
			name: "immune monster",
			edit: func(w *activatePoisonTestWorld4EE7E0) {
				w.unit.class = uint32(activatePoisonMonsterClassLow4EE7E0)
				w.unit.subclass = activatePoisonImmuneSubClass4EE7E0
			},
			want: []string{
				"arg:unit", "current:unit", "flags:unit", "buff:unit:23",
				"class:unit", "subclass:unit",
			},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			w := newActivatePoisonTestWorld4EE7E0()
			test.edit(w)
			if got := activatePoison4EE7E0(w.hooks()); got != 0 {
				t.Fatalf("result = %d, want 0", got)
			}
			if !reflect.DeepEqual(w.events, test.want) {
				t.Fatalf("events = %q, want %q", w.events, test.want)
			}
		})
	}
}

func TestActivatePoison4EE7E0BuffRejectsOnlyExactOne(t *testing.T) {
	for _, buff := range []int32{-1, 0, 1, 2} {
		t.Run(fmt.Sprintf("%d", buff), func(t *testing.T) {
			w := newActivatePoisonTestWorld4EE7E0()
			w.buffResult = buff
			got := activatePoison4EE7E0(w.hooks())
			want := int32(1)
			if buff == 1 {
				want = 0
			}
			if got != want {
				t.Fatalf("result = %d, want %d", got, want)
			}
			reachedClass := false
			for _, event := range w.events {
				reachedClass = reachedClass || event == "class:unit"
			}
			if reachedClass != (buff != 1) {
				t.Fatalf("class reached = %t for buff result %d", reachedClass, buff)
			}
		})
	}
}

func TestActivatePoison4EE7E0PlayerPointerFaultsHaveNoNilGates(t *testing.T) {
	tests := []struct {
		name      string
		edit      func(*activatePoisonTestWorld4EE7E0)
		wantPanic any
		wantTail  string
	}{
		{
			name: "nil update data",
			edit: func(w *activatePoisonTestWorld4EE7E0) {
				w.unit.class = uint32(activatePoisonPlayerClassLow4EE7E0)
				w.unit.update = nil
			},
			wantPanic: "player:nil-update",
			wantTail:  "player:nil",
		},
		{
			name: "nil player",
			edit: func(w *activatePoisonTestWorld4EE7E0) {
				w.unit.class = uint32(activatePoisonPlayerClassLow4EE7E0)
				w.unit.update.player = nil
			},
			wantPanic: "flags:nil-player",
			wantTail:  "player-flags:nil",
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			w := newActivatePoisonTestWorld4EE7E0()
			test.edit(w)
			defer func() {
				if got := recover(); got != test.wantPanic {
					t.Fatalf("panic = %v, want %v", got, test.wantPanic)
				}
				if got := w.events[len(w.events)-1]; got != test.wantTail {
					t.Fatalf("last event = %q, want %q", got, test.wantTail)
				}
			}()
			activatePoison4EE7E0(w.hooks())
		})
	}
}

func TestActivatePoison4EE7E0ResistanceDefersArguments(t *testing.T) {
	w := newActivatePoisonTestWorld4EE7E0()
	w.unit.class = 0
	w.roll = 11
	if got := activatePoison4EE7E0(w.hooks()); got != 0 {
		t.Fatalf("result = %d, want 0", got)
	}
	want := []string{
		"arg:unit",
		"current:unit",
		"flags:unit",
		"buff:unit:23",
		"class:unit",
		"protection:unit",
		"round:41480000",
		"random:0:100:361:C:\\NoxPost\\src\\Server\\Object\\health.c",
		"message:unit:Health.c:ResistPoison:0",
	}
	if !reflect.DeepEqual(w.events, want) {
		t.Fatalf("events =\n%q\nwant\n%q", w.events, want)
	}
}

func TestActivatePoison4EE7E0NoChangeSkipsEffectsAndFrame(t *testing.T) {
	w := newActivatePoisonTestWorld4EE7E0()
	w.unit.class = 0
	w.unit.current = 9
	w.increment = 100
	w.maximum = 3
	w.buffResult = 2
	if got := activatePoison4EE7E0(w.hooks()); got != 1 {
		t.Fatalf("result = %d, want 1", got)
	}
	wantTail := []string{"increment:100", "maximum:3"}
	if got := w.events[len(w.events)-2:]; !reflect.DeepEqual(got, wantTail) {
		t.Fatalf("event tail = %q, want %q", got, wantTail)
	}
	if w.unit.current != 9 || w.unit.health.frame != 11 {
		t.Fatalf("unexpected state mutation: poison=%d frame=%d", w.unit.current, w.unit.health.frame)
	}
}

func TestActivatePoison4EE7E0NilHealthSkipsFrameLoad(t *testing.T) {
	w := newActivatePoisonTestWorld4EE7E0()
	w.unit.class = 0
	w.unit.health = nil
	if got := activatePoison4EE7E0(w.hooks()); got != 1 {
		t.Fatalf("result = %d, want 1", got)
	}
	if got := w.events[len(w.events)-1]; got != "health:unit=nil" {
		t.Fatalf("last event = %q, want health:unit=nil", got)
	}
}

func TestActivatePoison4EE7E0CachesEntryValuesAndReloadsHealthAfterEffects(t *testing.T) {
	w := newActivatePoisonTestWorld4EE7E0()
	original := w.unit
	replacement := &activatePoisonTestObject4EE7E0{name: "replacement"}
	afterSetHealth := &activatePoisonTestHealth4EE7E0{name: "after-set", frame: 21}
	afterAudioHealth := &activatePoisonTestHealth4EE7E0{name: "after-audio", frame: 22}
	afterLoadHealth := &activatePoisonTestHealth4EE7E0{name: "after-load", frame: 23}

	w.afterArg = func(w *activatePoisonTestWorld4EE7E0, _ *activatePoisonTestObject4EE7E0) {
		w.unit = replacement
	}
	w.afterCurrent = func(_ *activatePoisonTestWorld4EE7E0, unit *activatePoisonTestObject4EE7E0, _ uint8) {
		unit.current = 99
	}
	w.afterClass = func(_ *activatePoisonTestWorld4EE7E0, unit *activatePoisonTestObject4EE7E0, _ uint32) {
		unit.class = 0
	}
	w.afterIncrement = func(w *activatePoisonTestWorld4EE7E0, _ int32) {
		w.increment = 99
	}
	w.afterMaximum = func(w *activatePoisonTestWorld4EE7E0, _ int32) {
		w.maximum = 0
	}
	w.afterSet = func(_ *activatePoisonTestWorld4EE7E0, unit *activatePoisonTestObject4EE7E0, _ int32) {
		unit.health = afterSetHealth
	}
	w.afterAudio = func(_ *activatePoisonTestWorld4EE7E0, unit *activatePoisonTestObject4EE7E0) {
		unit.health = afterAudioHealth
	}
	w.afterHealth = func(_ *activatePoisonTestWorld4EE7E0, _ *activatePoisonTestHealth4EE7E0) {
		original.health = afterLoadHealth
	}
	w.afterFrame = func(w *activatePoisonTestWorld4EE7E0, _ uint32) {
		w.frame = 88
	}

	if got := activatePoison4EE7E0(w.hooks()); got != 1 {
		t.Fatalf("result = %d, want 1", got)
	}
	if original.current != 2 {
		t.Fatalf("original poison = %d, want cached-target 2", original.current)
	}
	if replacement.current != 0 {
		t.Fatalf("replacement poison = %d, want untouched", replacement.current)
	}
	if afterAudioHealth.frame != 77 {
		t.Fatalf("cached live health frame = %d, want cached frame 77", afterAudioHealth.frame)
	}
	if afterSetHealth.frame != 21 || afterLoadHealth.frame != 23 {
		t.Fatalf("wrong HealthData changed: after-set=%d after-load=%d", afterSetHealth.frame, afterLoadHealth.frame)
	}
	if w.increment != 99 || w.maximum != 0 || w.frame != 88 {
		t.Fatalf("mutation hooks did not run: increment=%d maximum=%d frame=%d", w.increment, w.maximum, w.frame)
	}
	if !reflect.DeepEqual(w.events, []string{
		"arg:unit",
		"current:unit",
		"flags:unit",
		"buff:unit:23",
		"class:unit",
		"update:unit",
		"player:update",
		"player-flags:player",
		"subclass:unit",
		"protection:unit",
		"round:41480000",
		"random:0:100:361:C:\\NoxPost\\src\\Server\\Object\\health.c",
		"increment:2",
		"maximum:3",
		"set:unit:2",
		"audio:100:unit:0:0",
		"health:unit=after-audio",
		"frame:77",
		"store-frame:after-audio:77",
	}) {
		t.Fatalf("unexpected events:\n%q", w.events)
	}
}

func TestActivatePoisonTarget4EE7E0SignedClampAndWrap(t *testing.T) {
	tests := []struct {
		name               string
		current            uint8
		increment, maximum int32
		want               int32
	}{
		{"sum below cap", 5, 2, 10, 7},
		{"sum equals cap", 5, 3, 8, 8},
		{"clamp to cap", 5, 10, 8, 8},
		{"current already above cap", 9, 10, 8, 9},
		{"negative sum", 0, -1, 10, -1},
		{"negative cap retains current", 0, 1, -1, 0},
		{"positive overflow", 255, math.MaxInt32, math.MaxInt32, -2147483394},
		{"minimum exact", 0, math.MinInt32, math.MinInt32, math.MinInt32},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if got := activatePoisonTarget4EE7E0(test.current, test.increment, test.maximum); got != test.want {
				t.Fatalf("target = %d (%#x), want %d (%#x)", got, uint32(got), test.want, uint32(test.want))
			}
		})
	}
}

func TestActivatePoisonScaleProtection4EE7E0(t *testing.T) {
	tests := []struct {
		protection float64
		wantBits   uint32
	}{
		{0, 0x00000000},
		{0.125, 0x41480000},
		{0.9, 0x42b40000},
		{-0.125, 0xc1480000},
	}
	for _, test := range tests {
		if got := math.Float32bits(activatePoisonScaleProtection4EE7E0(test.protection)); got != test.wantBits {
			t.Errorf("scale(%g) bits = %#08x, want %#08x", test.protection, got, test.wantBits)
		}
	}
}

func TestActivatePoisonRound4EE7E0(t *testing.T) {
	tests := []struct {
		name  string
		value float32
		want  int32
	}{
		{"positive half even down", 0.5, 0},
		{"positive half even up", 1.5, 2},
		{"positive half even down two", 2.5, 2},
		{"negative half even", -1.5, -2},
		{"minimum valid", -2147483648, math.MinInt32},
		{"positive overflow", 2147483648, math.MinInt32},
		{"negative overflow", -2147483904, math.MinInt32},
		{"positive infinity", float32(math.Inf(1)), math.MinInt32},
		{"negative infinity", float32(math.Inf(-1)), math.MinInt32},
		{"nan", float32(math.NaN()), math.MinInt32},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if got := activatePoisonRound4EE7E0(test.value); got != test.want {
				t.Fatalf("round(%v) = %d, want %d", test.value, got, test.want)
			}
		})
	}
}
