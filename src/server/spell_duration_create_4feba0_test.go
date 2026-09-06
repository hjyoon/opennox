package server

import (
	"fmt"
	"reflect"
	"testing"
)

const (
	spellDurationCreateRecord4FEBA0  = uint64(0x100000111)
	spellDurationCreateSecond4FEBA0  = uint64(0x200000222)
	spellDurationCreateThird4FEBA0   = uint64(0x300000333)
	spellDurationCreateFourth4FEBA0  = uint64(0x400000444)
	spellDurationCreateTarget4FEBA0  = uint64(0x500000555)
	spellDurationCreateArg4FEBA0     = uint64(0x600000666)
	spellDurationCreateCreate4FEBA0  = uint64(0x700000777)
	spellDurationCreateUpdate4FEBA0  = uint64(0x800000888)
	spellDurationCreateDestroy4FEBA0 = uint64(0x900000999)
)

type spellDurationCreateObjectState4FEBA0 struct {
	flags  uint32
	typeID uint16
	x, y   float32
}

type spellDurationCreateAcceptState4FEBA0 struct {
	target uint64
	x, y   float32
}

type spellDurationCreateRecordState4FEBA0 struct {
	spell, level     uint32
	second, third    uint64
	sub108, sub104   uint64
	mode             uint32
	anchor           uint64
	x, y             float32
	field36          uint32
	target           uint64
	acceptX, acceptY float32
	create, update   uint64
	destroy          uint64
	frame60, frame64 uint32
	frame68, flags   uint32
}

type spellDurationCreateWorld4FEBA0 struct {
	glyphCache, glyphLookup uint32
	spell                   int32
	second, third, fourth   uint64
	arg                     uint64
	level                   int32
	create, update, destroy uint64
	duration                uint32
	newRecord               uint64
	duplicateResult         int32
	hasFlagsResult          int32
	audioID                 int32
	createResult            int32
	frames                  []uint32
	frameIndex              int

	objects map[uint64]*spellDurationCreateObjectState4FEBA0
	args    map[uint64]*spellDurationCreateAcceptState4FEBA0
	records map[uint64]*spellDurationCreateRecordState4FEBA0
	events  []string
	after   map[string]func()
	faultAt int
}

func newSpellDurationCreateWorld4FEBA0() *spellDurationCreateWorld4FEBA0 {
	return &spellDurationCreateWorld4FEBA0{
		glyphLookup:    7,
		spell:          31,
		second:         spellDurationCreateSecond4FEBA0,
		third:          spellDurationCreateThird4FEBA0,
		fourth:         spellDurationCreateFourth4FEBA0,
		arg:            spellDurationCreateArg4FEBA0,
		level:          -2,
		create:         spellDurationCreateCreate4FEBA0,
		update:         spellDurationCreateUpdate4FEBA0,
		destroy:        spellDurationCreateDestroy4FEBA0,
		duration:       0xfffffff0,
		newRecord:      spellDurationCreateRecord4FEBA0,
		hasFlagsResult: -7,
		audioID:        -123,
		frames:         []uint32{100, 101, 102},
		objects: map[uint64]*spellDurationCreateObjectState4FEBA0{
			spellDurationCreateSecond4FEBA0: {x: -1, y: -2},
			spellDurationCreateThird4FEBA0:  {x: 10.5, y: 20.25},
			spellDurationCreateFourth4FEBA0: {typeID: 7, x: 30.75, y: -40.5},
		},
		args: map[uint64]*spellDurationCreateAcceptState4FEBA0{
			spellDurationCreateArg4FEBA0: {target: spellDurationCreateTarget4FEBA0, x: -50.25, y: 60.5},
		},
		records: map[uint64]*spellDurationCreateRecordState4FEBA0{
			spellDurationCreateRecord4FEBA0: {flags: 0xa1b2c3d4},
		},
		after: make(map[string]func()),
	}
}

func (w *spellDurationCreateWorld4FEBA0) observe(event string) {
	w.events = append(w.events, event)
	if w.faultAt != 0 && len(w.events) == w.faultAt {
		panic(event)
	}
	if after := w.after[event]; after != nil {
		after()
	}
}

func (w *spellDurationCreateWorld4FEBA0) object(token uint64) *spellDurationCreateObjectState4FEBA0 {
	state := w.objects[token]
	if state == nil {
		panic(fmt.Sprintf("object:%x", token))
	}
	return state
}

func (w *spellDurationCreateWorld4FEBA0) accept(token uint64) *spellDurationCreateAcceptState4FEBA0 {
	state := w.args[token]
	if state == nil {
		panic(fmt.Sprintf("arg:%x", token))
	}
	return state
}

func (w *spellDurationCreateWorld4FEBA0) record(token uint64) *spellDurationCreateRecordState4FEBA0 {
	state := w.records[token]
	if state == nil {
		panic(fmt.Sprintf("record:%x", token))
	}
	return state
}

func (w *spellDurationCreateWorld4FEBA0) hooks() SpellDurationCreateHooks4FEBA0[uint64, uint64, uint64, uint64] {
	return SpellDurationCreateHooks4FEBA0[uint64, uint64, uint64, uint64]{
		LoadGlyphCache: func() uint32 {
			value := w.glyphCache
			w.observe(fmt.Sprintf("glyph-cache:%d", value))
			return value
		},
		LookupGlyph: func() uint32 {
			value := w.glyphLookup
			w.observe(fmt.Sprintf("glyph-lookup:%d", value))
			return value
		},
		StoreGlyphCache: func(value uint32) {
			w.observe(fmt.Sprintf("glyph-store:%d", value))
			w.glyphCache = value
		},
		LoadFourthArg: func() uint64 {
			value := w.fourth
			w.observe(fmt.Sprintf("fourth:%x", value))
			return value
		},
		LoadThirdArg: func() uint64 {
			value := w.third
			w.observe(fmt.Sprintf("third:%x", value))
			return value
		},
		LoadObjectFlags: func(object uint64) uint32 {
			value := w.object(object).flags
			w.observe(fmt.Sprintf("flags:%x:%x", object, value))
			return value
		},
		LoadObjectType: func(object uint64) uint16 {
			value := w.object(object).typeID
			w.observe(fmt.Sprintf("type:%x:%d", object, value))
			return value
		},
		LoadSpellArg: func() int32 {
			value := w.spell
			w.observe(fmt.Sprintf("spell:%d", value))
			return value
		},
		FindDuplicate: func(spellID int32, object uint64) int32 {
			w.observe(fmt.Sprintf("duplicate:%d:%x", spellID, object))
			return w.duplicateResult
		},
		CancelFor: func(spellID int32, object uint64) {
			w.observe(fmt.Sprintf("cancel-for:%d:%x", spellID, object))
		},
		BeforeCreate: func() {
			w.observe("before-create")
		},
		NewRecord: func() uint64 {
			value := w.newRecord
			w.observe(fmt.Sprintf("new-record:%x", value))
			return value
		},
		LoadLevelArg: func() int32 {
			value := w.level
			w.observe(fmt.Sprintf("level:%d", value))
			return value
		},
		LoadSecondArg: func() uint64 {
			value := w.second
			w.observe(fmt.Sprintf("second:%x", value))
			return value
		},
		StoreSpell: func(record uint64, value uint32) {
			w.observe(fmt.Sprintf("store-spell:%x:%x", record, value))
			w.record(record).spell = value
		},
		StoreLevel: func(record uint64, value uint32) {
			w.observe(fmt.Sprintf("store-level:%x:%x", record, value))
			w.record(record).level = value
		},
		StoreCaster: func(record, value uint64) {
			w.observe(fmt.Sprintf("store-caster:%x:%x", record, value))
			w.record(record).third = value
		},
		StoreSource: func(record, value uint64) {
			w.observe(fmt.Sprintf("store-source:%x:%x", record, value))
			w.record(record).second = value
		},
		StoreSub108: func(record, value uint64) {
			w.observe(fmt.Sprintf("store-sub108:%x:%x", record, value))
			w.record(record).sub108 = value
		},
		StoreSub104: func(record, value uint64) {
			w.observe(fmt.Sprintf("store-sub104:%x:%x", record, value))
			w.record(record).sub104 = value
		},
		StoreMode: func(record uint64, value uint32) {
			w.observe(fmt.Sprintf("store-mode:%x:%d", record, value))
			w.record(record).mode = value
		},
		StoreAnchor: func(record, value uint64) {
			w.observe(fmt.Sprintf("store-anchor:%x:%x", record, value))
			w.record(record).anchor = value
		},
		LoadPositionX: func(object uint64) float32 {
			value := w.object(object).x
			w.observe(fmt.Sprintf("position-x:%x:%g", object, value))
			return value
		},
		StorePositionX: func(record uint64, value float32) {
			w.observe(fmt.Sprintf("store-position-x:%x:%g", record, value))
			w.record(record).x = value
		},
		LoadPositionY: func(object uint64) float32 {
			value := w.object(object).y
			w.observe(fmt.Sprintf("position-y:%x:%g", object, value))
			return value
		},
		LoadAcceptArg: func() uint64 {
			value := w.arg
			w.observe(fmt.Sprintf("accept-arg:%x", value))
			return value
		},
		StoreField36: func(record uint64, value uint32) {
			w.observe(fmt.Sprintf("store-field36:%x:%d", record, value))
			w.record(record).field36 = value
		},
		StorePositionY: func(record uint64, value float32) {
			w.observe(fmt.Sprintf("store-position-y:%x:%g", record, value))
			w.record(record).y = value
		},
		LoadTarget: func(arg uint64) uint64 {
			value := w.accept(arg).target
			w.observe(fmt.Sprintf("accept-target:%x:%x", arg, value))
			return value
		},
		LoadCreateArg: func() uint64 {
			value := w.create
			w.observe(fmt.Sprintf("create-arg:%x", value))
			return value
		},
		StoreTarget: func(record, value uint64) {
			w.observe(fmt.Sprintf("store-target:%x:%x", record, value))
			w.record(record).target = value
		},
		LoadAcceptX: func(arg uint64) float32 {
			value := w.accept(arg).x
			w.observe(fmt.Sprintf("accept-x:%x:%g", arg, value))
			return value
		},
		StoreAcceptX: func(record uint64, value float32) {
			w.observe(fmt.Sprintf("store-accept-x:%x:%g", record, value))
			w.record(record).acceptX = value
		},
		LoadUpdateArg: func() uint64 {
			value := w.update
			w.observe(fmt.Sprintf("update-arg:%x", value))
			return value
		},
		LoadAcceptY: func(arg uint64) float32 {
			value := w.accept(arg).y
			w.observe(fmt.Sprintf("accept-y:%x:%g", arg, value))
			return value
		},
		StoreCreate: func(record, value uint64) {
			w.observe(fmt.Sprintf("store-create:%x:%x", record, value))
			w.record(record).create = value
		},
		StoreAcceptY: func(record uint64, value float32) {
			w.observe(fmt.Sprintf("store-accept-y:%x:%g", record, value))
			w.record(record).acceptY = value
		},
		LoadDestroyArg: func() uint64 {
			value := w.destroy
			w.observe(fmt.Sprintf("destroy-arg:%x", value))
			return value
		},
		StoreUpdate: func(record, value uint64) {
			w.observe(fmt.Sprintf("store-update:%x:%x", record, value))
			w.record(record).update = value
		},
		StoreDestroy: func(record, value uint64) {
			w.observe(fmt.Sprintf("store-destroy:%x:%x", record, value))
			w.record(record).destroy = value
		},
		LoadFrame: func() uint32 {
			value := w.frames[w.frameIndex]
			w.frameIndex++
			w.observe(fmt.Sprintf("frame:%d", value))
			return value
		},
		LoadDurationArg: func() uint32 {
			value := w.duration
			w.observe(fmt.Sprintf("duration:%x", value))
			return value
		},
		StoreFrame60: func(record uint64, value uint32) {
			w.observe(fmt.Sprintf("store-frame60:%x:%d", record, value))
			w.record(record).frame60 = value
		},
		StoreFrame64: func(record uint64, value uint32) {
			w.observe(fmt.Sprintf("store-frame64:%x:%d", record, value))
			w.record(record).frame64 = value
		},
		StoreFlagsLowByte: func(record uint64, value byte) {
			w.observe(fmt.Sprintf("store-flags-low:%x:%d", record, value))
			state := w.record(record)
			state.flags = state.flags&^0xff | uint32(value)
		},
		StoreFrame68: func(record uint64, value uint32) {
			w.observe(fmt.Sprintf("store-frame68:%x:%d", record, value))
			w.record(record).frame68 = value
		},
		AddRecord: func(record uint64) {
			w.observe(fmt.Sprintf("add:%x", record))
		},
		SpellHasFlags: func(spellID int32, mask uint32) int32 {
			w.observe(fmt.Sprintf("spell-flags:%d:%x", spellID, mask))
			return w.hasFlagsResult
		},
		SpellAudio: func(spellID, selector int32) int32 {
			w.observe(fmt.Sprintf("spell-audio:%d:%d", spellID, selector))
			return w.audioID
		},
		AudioEvent: func(audio int32, object uint64, kind, code int32) {
			w.observe(fmt.Sprintf("audio:%d:%x:%d:%d", audio, object, kind, code))
		},
		CallCreate: func(callback, record uint64) int32 {
			w.observe(fmt.Sprintf("call-create:%x:%x", callback, record))
			return w.createResult
		},
		CancelSpell: func(record uint64) {
			w.observe(fmt.Sprintf("cancel-spell:%x", record))
		},
	}
}

func TestSpellDurationCreate4FEBA0ExactSuccessOrderAndWrites(t *testing.T) {
	w := newSpellDurationCreateWorld4FEBA0()
	if got := SpellDurationCreate4FEBA0(w.hooks()); got != 1 {
		t.Fatalf("result = %d, want 1", got)
	}

	wantEvents := []string{
		"glyph-cache:0", "glyph-lookup:7", "glyph-store:7",
		"fourth:400000444", "third:300000333", "flags:300000333:0", "spell:31",
		"cancel-for:31:300000333", "before-create", "new-record:100000111",
		"level:-2", "second:200000222",
		"store-spell:100000111:1f", "store-level:100000111:fffffffe",
		"store-caster:100000111:300000333", "store-source:100000111:200000222",
		"store-sub108:100000111:0", "store-sub104:100000111:0",
		"glyph-cache:7", "type:400000444:7", "store-mode:100000111:1",
		"store-anchor:100000111:400000444", "position-x:400000444:30.75",
		"store-position-x:100000111:30.75", "position-y:400000444:-40.5",
		"accept-arg:600000666", "store-field36:100000111:0",
		"store-position-y:100000111:-40.5", "accept-target:600000666:500000555",
		"create-arg:700000777", "store-target:100000111:500000555",
		"accept-x:600000666:-50.25", "store-accept-x:100000111:-50.25",
		"update-arg:800000888", "accept-y:600000666:60.5",
		"store-create:100000111:700000777", "store-accept-y:100000111:60.5",
		"destroy-arg:900000999", "store-update:100000111:800000888",
		"store-destroy:100000111:900000999", "frame:100", "duration:fffffff0",
		"store-frame60:100000111:100", "frame:101", "store-frame64:100000111:101",
		"frame:102", "store-flags-low:100000111:0", "store-frame68:100000111:86",
		"add:100000111", "spell-flags:31:4", "spell-audio:31:1",
		"audio:-123:300000333:0:0", "call-create:700000777:100000111",
	}
	if !reflect.DeepEqual(w.events, wantEvents) {
		t.Fatalf("events =\n%q\nwant\n%q", w.events, wantEvents)
	}
	wantRecord := &spellDurationCreateRecordState4FEBA0{
		spell: 31, level: 0xfffffffe,
		second: spellDurationCreateSecond4FEBA0, third: spellDurationCreateThird4FEBA0,
		mode: 1, anchor: spellDurationCreateFourth4FEBA0,
		x: 30.75, y: -40.5,
		target: spellDurationCreateTarget4FEBA0, acceptX: -50.25, acceptY: 60.5,
		create: spellDurationCreateCreate4FEBA0, update: spellDurationCreateUpdate4FEBA0,
		destroy: spellDurationCreateDestroy4FEBA0,
		frame60: 100, frame64: 101, frame68: 86, flags: 0xa1b2c300,
	}
	if got := w.records[spellDurationCreateRecord4FEBA0]; !reflect.DeepEqual(got, wantRecord) {
		t.Fatalf("record = %#v, want %#v", got, wantRecord)
	}
}

func TestSpellDurationCreate4FEBA0AdmissionGate(t *testing.T) {
	tests := []struct {
		name      string
		third     uint64
		fourth    uint64
		flags     uint32
		fourthTyp uint16
		want      int32
		wantTail  string
	}{
		{name: "active-caster-bypasses-fourth-type", flags: 0, fourthTyp: 99, want: 1, wantTail: "call-create:700000777:100000111"},
		{name: "dead-caster-with-glyph", flags: 0x20, fourthTyp: 7, want: 1, wantTail: "call-create:700000777:100000111"},
		{name: "destroyed-caster-with-glyph", flags: 0x8000, fourthTyp: 7, want: 1, wantTail: "call-create:700000777:100000111"},
		{name: "dead-caster-with-other", flags: 0x20, fourthTyp: 8, want: 0, wantTail: "type:400000444:8"},
		{name: "nil-caster-with-glyph", third: 0, fourthTyp: 7, want: 1, wantTail: "call-create:700000777:100000111"},
		{name: "nil-caster-with-other", third: 0, fourthTyp: 8, want: 0, wantTail: "type:400000444:8"},
		{name: "dead-caster-without-fourth", fourth: 0, flags: 0x20, want: 1, wantTail: "call-create:700000777:100000111"},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			w := newSpellDurationCreateWorld4FEBA0()
			if tc.third == 0 && tc.name != "nil-caster-with-glyph" && tc.name != "nil-caster-with-other" {
				w.third = spellDurationCreateThird4FEBA0
			} else if tc.name == "nil-caster-with-glyph" || tc.name == "nil-caster-with-other" {
				w.third = 0
			}
			if tc.fourth == 0 && tc.name == "dead-caster-without-fourth" {
				w.fourth = 0
			}
			w.objects[spellDurationCreateThird4FEBA0].flags = tc.flags
			w.objects[spellDurationCreateFourth4FEBA0].typeID = tc.fourthTyp
			if got := SpellDurationCreate4FEBA0(w.hooks()); got != tc.want {
				t.Fatalf("result = %d, want %d", got, tc.want)
			}
			if got := w.events[len(w.events)-1]; got != tc.wantTail {
				t.Fatalf("event tail = %q, want %q (events %q)", got, tc.wantTail, w.events)
			}
			if w.third == 0 && tc.want == 1 {
				if got := w.records[w.newRecord].third; got != 0 {
					t.Fatalf("stored caster = %#x, want nil", got)
				}
			}
		})
	}
}

func TestSpellDurationCreate4FEBA0NilCasterWithoutGlyphFaultsAfterAllocation(t *testing.T) {
	w := newSpellDurationCreateWorld4FEBA0()
	w.third = 0
	w.fourth = 0
	defer func() {
		if got := recover(); got == nil {
			t.Fatal("nil caster did not fault at fallback position load")
		}
		want := []string{
			"glyph-cache:0", "glyph-lookup:7", "glyph-store:7", "fourth:0", "third:0",
			"spell:31", "before-create", "new-record:100000111", "level:-2", "second:200000222",
			"store-spell:100000111:1f", "store-level:100000111:fffffffe",
			"store-caster:100000111:0", "store-source:100000111:200000222",
			"store-sub108:100000111:0", "store-sub104:100000111:0",
			"store-mode:100000111:0", "store-anchor:100000111:0",
		}
		if !reflect.DeepEqual(w.events, want) {
			t.Fatalf("fault prefix = %q, want %q", w.events, want)
		}
	}()
	SpellDurationCreate4FEBA0(w.hooks())
}

func TestSpellDurationCreate4FEBA0DuplicateRequiresExactOne(t *testing.T) {
	for _, tc := range []struct {
		result int32
		want   int32
		last   string
	}{
		{result: 1, want: 1, last: "duplicate:59:300000333"},
		{result: 2, want: 1, last: "call-create:700000777:100000111"},
		{result: -1, want: 1, last: "call-create:700000777:100000111"},
	} {
		w := newSpellDurationCreateWorld4FEBA0()
		w.glyphCache = 7
		w.spell = spellDurationCreatePlasma4FEBA0
		w.duplicateResult = tc.result
		if got := SpellDurationCreate4FEBA0(w.hooks()); got != tc.want {
			t.Fatalf("duplicate result %d: result = %d, want %d", tc.result, got, tc.want)
		}
		if got := w.events[len(w.events)-1]; got != tc.last {
			t.Fatalf("duplicate result %d: tail = %q, want %q", tc.result, got, tc.last)
		}
		if tc.result != 1 && !containsSpellDurationCreateEvent4FEBA0(w.events, "cancel-for:59:300000333") {
			t.Fatalf("duplicate result %d did not continue to cancellation: %q", tc.result, w.events)
		}
	}
}

func TestSpellDurationCreate4FEBA0ReloadsGlyphCacheAndType(t *testing.T) {
	w := newSpellDurationCreateWorld4FEBA0()
	w.glyphCache = 7
	w.objects[w.third].flags = 0x20
	w.after["type:400000444:7"] = func() {
		// This mutation occurs after admission. The original caches the entry
		// value for that gate, but reloads both values after allocation.
		w.glyphCache = 8
		w.objects[w.fourth].typeID = 8
	}
	w.after["cancel-for:31:300000333"] = func() {
		w.objects[w.fourth].typeID = 9
	}
	w.after["before-create"] = func() {
		w.objects[w.fourth].typeID = 8
	}

	if got := SpellDurationCreate4FEBA0(w.hooks()); got != 1 {
		t.Fatalf("result = %d, want 1", got)
	}
	if !containsSpellDurationCreateEvent4FEBA0(w.events, "glyph-cache:8") ||
		!containsSpellDurationCreateEvent4FEBA0(w.events, "type:400000444:8") {
		t.Fatalf("missing live cache/type reload: %q", w.events)
	}
	if got := w.records[w.newRecord].anchor; got != w.fourth {
		t.Fatalf("anchor = %#x, want cached fourth %#x", got, w.fourth)
	}
}

func TestSpellDurationCreate4FEBA0AllocationAndCreateResults(t *testing.T) {
	w := newSpellDurationCreateWorld4FEBA0()
	w.newRecord = 0
	if got := SpellDurationCreate4FEBA0(w.hooks()); got != 0 {
		t.Fatalf("allocation result = %d, want 0", got)
	}
	if got := w.events[len(w.events)-1]; got != "new-record:0" {
		t.Fatalf("allocation tail = %q, want new-record:0", got)
	}

	for _, result := range []int32{1, -1, 2} {
		w = newSpellDurationCreateWorld4FEBA0()
		w.createResult = result
		if got := SpellDurationCreate4FEBA0(w.hooks()); got != 0 {
			t.Fatalf("create result %d: result = %d, want 0", result, got)
		}
		if got := w.events[len(w.events)-1]; got != "cancel-spell:100000111" {
			t.Fatalf("create result %d: tail = %q", result, got)
		}
	}

	w = newSpellDurationCreateWorld4FEBA0()
	w.create = 0
	w.createResult = 1
	w.hasFlagsResult = 0
	if got := SpellDurationCreate4FEBA0(w.hooks()); got != 1 {
		t.Fatalf("nil create result = %d, want 1", got)
	}
	if containsSpellDurationCreateEventPrefix4FEBA0(w.events, "call-create:") {
		t.Fatalf("nil create callback was called: %q", w.events)
	}
	if !containsSpellDurationCreateEvent4FEBA0(w.events, "spell-audio:31:0") {
		t.Fatalf("zero flags did not select cast audio: %q", w.events)
	}
}

func TestSpellDurationCreate4FEBA0FaultPrefixes(t *testing.T) {
	baseline := newSpellDurationCreateWorld4FEBA0()
	baseline.createResult = 1
	if got := SpellDurationCreate4FEBA0(baseline.hooks()); got != 0 {
		t.Fatalf("baseline result = %d, want 0", got)
	}
	want := append([]string(nil), baseline.events...)
	for faultAt := 1; faultAt <= len(want); faultAt++ {
		w := newSpellDurationCreateWorld4FEBA0()
		w.createResult = 1
		w.faultAt = faultAt
		func() {
			defer func() {
				if recover() == nil {
					t.Fatalf("fault %d did not propagate", faultAt)
				}
			}()
			SpellDurationCreate4FEBA0(w.hooks())
		}()
		if got := w.events; !reflect.DeepEqual(got, want[:faultAt]) {
			t.Fatalf("fault %d prefix = %q, want %q", faultAt, got, want[:faultAt])
		}
	}
}

func containsSpellDurationCreateEvent4FEBA0(events []string, want string) bool {
	for _, event := range events {
		if event == want {
			return true
		}
	}
	return false
}

func containsSpellDurationCreateEventPrefix4FEBA0(events []string, prefix string) bool {
	for _, event := range events {
		if len(event) >= len(prefix) && event[:len(prefix)] == prefix {
			return true
		}
	}
	return false
}
