package server

import (
	"fmt"
	"reflect"
	"testing"
)

const (
	spellDurationCancelRoot4FE9D0      = uint64(0x100000101)
	spellDurationCancelChildA4FE9D0    = uint64(0x200000202)
	spellDurationCancelChildOld4FE9D0  = uint64(0x300000303)
	spellDurationCancelChildLive4FE9D0 = uint64(0x400000404)
	spellDurationCancelCaster4FE9D0    = uint64(0x500000505)
	spellDurationCancelTargetA4FE9D0   = uint64(0x600000606)
	spellDurationCancelTargetB4FE9D0   = uint64(0x700000707)
	spellDurationCancelUpdate4FE9D0    = uint64(0x800000808)
	spellDurationCancelPlayer4FE9D0    = uint64(0x900000909)
)

type spellDurationCancelRecordState4FE9D0 struct {
	spell  uint32
	caster uint64
	target uint64
	sub108 uint64
	next   uint64
	flags  byte
}

type spellDurationCancelObjectState4FE9D0 struct {
	class  byte
	update uint64
}

type spellDurationCancelWorld4FE9D0 struct {
	events      []string
	faultAt     int
	records     map[uint64]*spellDurationCancelRecordState4FE9D0
	objects     map[uint64]*spellDurationCancelObjectState4FE9D0
	players     map[uint64]uint64
	indices     map[uint64]byte
	afterReport func()
	afterStop   map[uint64]func()
}

func newSpellDurationCancelWorld4FE9D0() *spellDurationCancelWorld4FE9D0 {
	return &spellDurationCancelWorld4FE9D0{
		records: map[uint64]*spellDurationCancelRecordState4FE9D0{
			spellDurationCancelRoot4FE9D0: {
				spell:  8,
				caster: spellDurationCancelCaster4FE9D0,
				target: spellDurationCancelTargetB4FE9D0,
				flags:  0x40,
			},
			spellDurationCancelChildA4FE9D0: {
				target: spellDurationCancelTargetA4FE9D0,
				next:   spellDurationCancelChildOld4FE9D0,
			},
			spellDurationCancelChildOld4FE9D0:  {},
			spellDurationCancelChildLive4FE9D0: {},
		},
		objects: map[uint64]*spellDurationCancelObjectState4FE9D0{
			spellDurationCancelCaster4FE9D0: {
				class:  0x84,
				update: spellDurationCancelUpdate4FE9D0,
			},
		},
		players: map[uint64]uint64{
			spellDurationCancelUpdate4FE9D0: spellDurationCancelPlayer4FE9D0,
		},
		indices: map[uint64]byte{
			spellDurationCancelPlayer4FE9D0: 7,
		},
		afterStop: make(map[uint64]func()),
	}
}

func (w *spellDurationCancelWorld4FE9D0) record(id uint64) *spellDurationCancelRecordState4FE9D0 {
	if w.records[id] == nil {
		w.records[id] = &spellDurationCancelRecordState4FE9D0{}
	}
	return w.records[id]
}

func (w *spellDurationCancelWorld4FE9D0) object(id uint64) *spellDurationCancelObjectState4FE9D0 {
	if w.objects[id] == nil {
		w.objects[id] = &spellDurationCancelObjectState4FE9D0{}
	}
	return w.objects[id]
}

func (w *spellDurationCancelWorld4FE9D0) observe(event string) {
	w.events = append(w.events, event)
	if w.faultAt != 0 && len(w.events) == w.faultAt {
		panic(event)
	}
}

func (w *spellDurationCancelWorld4FE9D0) hooks() SpellDurationCancelHooks4FE9D0[uint64, uint64, uint64, uint64] {
	return SpellDurationCancelHooks4FE9D0[uint64, uint64, uint64, uint64]{
		LoadCaster: func(record uint64) uint64 {
			value := w.record(record).caster
			w.observe(fmt.Sprintf("caster:%x:%x", record, value))
			return value
		},
		LoadClassLowByte: func(object uint64) byte {
			value := w.object(object).class
			w.observe(fmt.Sprintf("class:%x:%02x", object, value))
			return value
		},
		LoadSpell: func(record uint64) uint32 {
			value := w.record(record).spell
			w.observe(fmt.Sprintf("spell:%x:%d", record, value))
			return value
		},
		LoadUpdate: func(object uint64) uint64 {
			value := w.object(object).update
			w.observe(fmt.Sprintf("update:%x:%x", object, value))
			return value
		},
		LoadPlayer: func(update uint64) uint64 {
			value := w.players[update]
			w.observe(fmt.Sprintf("player:%x:%x", update, value))
			return value
		},
		LoadPlayerIndex: func(player uint64) byte {
			value := w.indices[player]
			w.observe(fmt.Sprintf("index:%x:%02x", player, value))
			return value
		},
		ReportSpellStat: func(index byte, spell uint32, status byte) {
			w.observe(fmt.Sprintf("report:%02x:%d:%02x", index, spell, status))
			if w.afterReport != nil {
				w.afterReport()
			}
		},
		LoadSub108: func(record uint64) uint64 {
			value := w.record(record).sub108
			w.observe(fmt.Sprintf("sub108:%x:%x", record, value))
			return value
		},
		LoadTarget: func(record uint64) uint64 {
			value := w.record(record).target
			w.observe(fmt.Sprintf("target:%x:%x", record, value))
			return value
		},
		StopRay: func(record, target uint64) {
			w.observe(fmt.Sprintf("stop:%x:%x", record, target))
			if after := w.afterStop[record]; after != nil {
				after()
			}
		},
		LoadNext: func(record uint64) uint64 {
			value := w.record(record).next
			w.observe(fmt.Sprintf("next:%x:%x", record, value))
			return value
		},
		LoadFlagsLowByte: func(record uint64) byte {
			value := w.record(record).flags
			w.observe(fmt.Sprintf("flags:%x:%02x", record, value))
			return value
		},
		StoreFlagsLowByte: func(record uint64, flags byte) {
			w.observe(fmt.Sprintf("store-flags:%x:%02x", record, flags))
			w.record(record).flags = flags
		},
	}
}

func TestSpellDurationCancel4FE9D0ReloadsSpellNextAndFlags(t *testing.T) {
	w := newSpellDurationCancelWorld4FE9D0()
	root := w.record(spellDurationCancelRoot4FE9D0)
	w.afterReport = func() {
		root.spell = spellDurationCancelChainLightning4FE9D0
		root.sub108 = spellDurationCancelChildA4FE9D0
	}
	w.afterStop[spellDurationCancelChildA4FE9D0] = func() {
		w.record(spellDurationCancelChildA4FE9D0).next = spellDurationCancelChildLive4FE9D0
		root.flags = 0x86
	}

	got := SpellDurationCancel4FE9D0(spellDurationCancelRoot4FE9D0, w.hooks())

	wantEvents := []string{
		"caster:100000101:500000505",
		"class:500000505:84",
		"spell:100000101:8",
		"update:500000505:800000808",
		"player:800000808:900000909",
		"index:900000909:07",
		"report:07:8:0f",
		"spell:100000101:43",
		"sub108:100000101:200000202",
		"target:200000202:600000606",
		"stop:200000202:600000606",
		"next:200000202:400000404",
		"target:400000404:0",
		"next:400000404:0",
		"flags:100000101:86",
		"store-flags:100000101:87",
	}
	if !reflect.DeepEqual(w.events, wantEvents) {
		t.Fatalf("trace mismatch\n got: %q\nwant: %q", w.events, wantEvents)
	}
	if got != 0x87 || root.flags != 0x87 {
		t.Fatalf("result/flags = %02x/%02x, want 87/87", got, root.flags)
	}
	for _, event := range w.events {
		if event == "target:300000303:0" {
			t.Fatal("visited the stale Next record cached before StopRay")
		}
	}
}

func TestSpellDurationCancel4FE9D0InitialChainCanBecomeSingleRay(t *testing.T) {
	w := newSpellDurationCancelWorld4FE9D0()
	root := w.record(spellDurationCancelRoot4FE9D0)
	root.spell = spellDurationCancelChainLightning4FE9D0
	root.target = spellDurationCancelTargetA4FE9D0
	w.afterReport = func() {
		root.spell = 22
		root.target = spellDurationCancelTargetB4FE9D0
	}
	w.afterStop[spellDurationCancelRoot4FE9D0] = func() {
		root.flags = 0xfe
	}

	got := SpellDurationCancel4FE9D0(spellDurationCancelRoot4FE9D0, w.hooks())

	wantTail := []string{
		"report:07:43:00",
		"spell:100000101:22",
		"target:100000101:700000707",
		"stop:100000101:700000707",
		"flags:100000101:fe",
		"store-flags:100000101:ff",
	}
	if len(w.events) < len(wantTail) || !reflect.DeepEqual(w.events[len(w.events)-len(wantTail):], wantTail) {
		t.Fatalf("trace tail = %q, want %q", w.events, wantTail)
	}
	if got != 0xff || root.flags != 0xff {
		t.Fatalf("result/flags = %02x/%02x, want ff/ff", got, root.flags)
	}
}

func TestSpellDurationCancel4FE9D0SkipsPlayerLoadsForMissingOrNonPlayerCaster(t *testing.T) {
	for _, tc := range []struct {
		name   string
		caster uint64
		class  byte
		want   []string
	}{
		{
			name: "missing",
			want: []string{
				"caster:100000101:0",
				"spell:100000101:8",
				"target:100000101:0",
				"flags:100000101:40",
				"store-flags:100000101:41",
			},
		},
		{
			name:   "non-player",
			caster: spellDurationCancelCaster4FE9D0,
			class:  0x82,
			want: []string{
				"caster:100000101:500000505",
				"class:500000505:82",
				"spell:100000101:8",
				"target:100000101:0",
				"flags:100000101:40",
				"store-flags:100000101:41",
			},
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			w := newSpellDurationCancelWorld4FE9D0()
			root := w.record(spellDurationCancelRoot4FE9D0)
			root.caster = tc.caster
			root.target = 0
			if tc.caster != 0 {
				w.object(tc.caster).class = tc.class
			}

			if got := SpellDurationCancel4FE9D0(spellDurationCancelRoot4FE9D0, w.hooks()); got != 0x41 {
				t.Fatalf("result = %02x, want 41", got)
			}
			if !reflect.DeepEqual(w.events, tc.want) {
				t.Fatalf("events = %q, want %q", w.events, tc.want)
			}
		})
	}
}

func TestSpellDurationCancel4FE9D0DoesNotGuardNilRecord(t *testing.T) {
	stop := &struct{}{}
	var events []string
	var recovered any
	func() {
		defer func() { recovered = recover() }()
		SpellDurationCancel4FE9D0(uint64(0), SpellDurationCancelHooks4FE9D0[uint64, uint64, uint64, uint64]{
			LoadCaster: func(record uint64) uint64 {
				events = append(events, fmt.Sprintf("caster:%x", record))
				panic(stop)
			},
		})
	}()
	if recovered != stop {
		t.Fatalf("recovered = %#v, want sentinel", recovered)
	}
	if want := []string{"caster:0"}; !reflect.DeepEqual(events, want) {
		t.Fatalf("events = %q, want %q", events, want)
	}
}

func TestSpellDurationCancel4FE9D0FaultPrefixes(t *testing.T) {
	baseline := newSpellDurationCancelWorld4FE9D0()
	baseline.record(spellDurationCancelRoot4FE9D0).target = spellDurationCancelTargetA4FE9D0
	SpellDurationCancel4FE9D0(spellDurationCancelRoot4FE9D0, baseline.hooks())
	want := append([]string(nil), baseline.events...)

	for faultAt := 1; faultAt <= len(want); faultAt++ {
		t.Run(fmt.Sprintf("fault-%d", faultAt), func(t *testing.T) {
			w := newSpellDurationCancelWorld4FE9D0()
			w.record(spellDurationCancelRoot4FE9D0).target = spellDurationCancelTargetA4FE9D0
			w.faultAt = faultAt
			var recovered any
			func() {
				defer func() { recovered = recover() }()
				SpellDurationCancel4FE9D0(spellDurationCancelRoot4FE9D0, w.hooks())
			}()
			if recovered == nil {
				t.Fatal("fault sentinel was not recovered")
			}
			if prefix := want[:faultAt]; !reflect.DeepEqual(w.events, prefix) {
				t.Fatalf("events = %q, want fault prefix %q", w.events, prefix)
			}
		})
	}
}
