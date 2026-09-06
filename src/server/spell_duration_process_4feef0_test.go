package server

import (
	"fmt"
	"math"
	"reflect"
	"testing"
)

const (
	spellDurationProcessRecordA4FEEF0 = uint64(0x100000101)
	spellDurationProcessRecordB4FEEF0 = uint64(0x200000202)
	spellDurationProcessRecordC4FEEF0 = uint64(0x300000303)
	spellDurationProcessRecordD4FEEF0 = uint64(0x400000404)
	spellDurationProcessRecordE4FEEF0 = uint64(0x500000505)

	spellDurationProcessDeadCaster4FEEF0 = uint64(0x1111111100000101)
	spellDurationProcessLiveCaster4FEEF0 = uint64(0x2222222200000202)
	spellDurationProcessDeadObj124FEEF0  = uint64(0x3333333300000303)
	spellDurationProcessLiveObj124FEEF0  = uint64(0x4444444400000404)
	spellDurationProcessDeadObj244FEEF0  = uint64(0x5555555500000505)
	spellDurationProcessLiveObj244FEEF0  = uint64(0x6666666600000606)

	spellDurationProcessUpdateD4FEEF0 = uint64(0x7777777700000707)
	spellDurationProcessUpdateE4FEEF0 = uint64(0x8888888800000808)
)

type spellDurationProcessRecordState4FEEF0 struct {
	flags   uint32
	next    uint64
	caster  uint64
	obj12   uint64
	flag20  uint32
	obj24   uint64
	frame60 uint32
	frame68 uint32
	update  uint64
}

type spellDurationProcessWorld4FEEF0 struct {
	events       []string
	after        map[string]func()
	faultAt      int
	head         uint64
	currentFrame uint32
	records      map[uint64]*spellDurationProcessRecordState4FEEF0
	objects      map[uint64]uint32
	updates      map[uint64]int32
}

func spellDurationProcessRecordName4FEEF0(record uint64) string {
	switch record {
	case 0:
		return "nil"
	case spellDurationProcessRecordA4FEEF0:
		return "A"
	case spellDurationProcessRecordB4FEEF0:
		return "B"
	case spellDurationProcessRecordC4FEEF0:
		return "C"
	case spellDurationProcessRecordD4FEEF0:
		return "D"
	case spellDurationProcessRecordE4FEEF0:
		return "E"
	default:
		return fmt.Sprintf("%x", record)
	}
}

func spellDurationProcessObjectName4FEEF0(object uint64) string {
	switch object {
	case 0:
		return "nil"
	case spellDurationProcessDeadCaster4FEEF0:
		return "dead-caster"
	case spellDurationProcessLiveCaster4FEEF0:
		return "live-caster"
	case spellDurationProcessDeadObj124FEEF0:
		return "dead-obj12"
	case spellDurationProcessLiveObj124FEEF0:
		return "live-obj12"
	case spellDurationProcessDeadObj244FEEF0:
		return "dead-obj24"
	case spellDurationProcessLiveObj244FEEF0:
		return "live-obj24"
	default:
		return fmt.Sprintf("%x", object)
	}
}

func spellDurationProcessUpdateName4FEEF0(update uint64) string {
	switch update {
	case 0:
		return "nil"
	case spellDurationProcessUpdateD4FEEF0:
		return "update-D"
	case spellDurationProcessUpdateE4FEEF0:
		return "update-E"
	default:
		return fmt.Sprintf("%x", update)
	}
}

func newSpellDurationProcessWorld4FEEF0() *spellDurationProcessWorld4FEEF0 {
	return &spellDurationProcessWorld4FEEF0{
		after:        make(map[string]func()),
		head:         spellDurationProcessRecordA4FEEF0,
		currentFrame: 10,
		records: map[uint64]*spellDurationProcessRecordState4FEEF0{
			spellDurationProcessRecordA4FEEF0: {
				flags: 0xabcdef01, next: spellDurationProcessRecordB4FEEF0,
			},
			spellDurationProcessRecordB4FEEF0: {
				flags: 0x12345680, next: spellDurationProcessRecordC4FEEF0,
				caster: spellDurationProcessDeadCaster4FEEF0, obj12: spellDurationProcessDeadObj124FEEF0,
				obj24: spellDurationProcessLiveObj244FEEF0, frame60: 99, frame68: 99, update: spellDurationProcessUpdateE4FEEF0,
			},
			spellDurationProcessRecordC4FEEF0: {
				next:   spellDurationProcessRecordD4FEEF0,
				caster: spellDurationProcessLiveCaster4FEEF0, obj12: spellDurationProcessLiveObj124FEEF0,
				obj24: spellDurationProcessDeadObj244FEEF0, frame60: 1, frame68: 10, update: spellDurationProcessUpdateE4FEEF0,
			},
			spellDurationProcessRecordD4FEEF0: {
				flags: 0xffffff02, next: spellDurationProcessRecordE4FEEF0, flag20: 1,
				obj24: spellDurationProcessLiveObj244FEEF0, frame60: 7, frame68: 7, update: spellDurationProcessUpdateD4FEEF0,
			},
			spellDurationProcessRecordE4FEEF0: {
				caster:  spellDurationProcessLiveCaster4FEEF0,
				frame60: 7, frame68: 11,
			},
		},
		objects: map[uint64]uint32{
			spellDurationProcessDeadCaster4FEEF0: 0x00008000,
			spellDurationProcessLiveCaster4FEEF0: 0x80000000,
			spellDurationProcessDeadObj124FEEF0:  0x12345620,
			spellDurationProcessLiveObj124FEEF0:  0xabcdef10,
			spellDurationProcessDeadObj244FEEF0:  0x76543220,
			spellDurationProcessLiveObj244FEEF0:  0x00008000,
		},
		updates: map[uint64]int32{
			spellDurationProcessUpdateD4FEEF0: 1,
			spellDurationProcessUpdateE4FEEF0: 0,
		},
	}
}

func (w *spellDurationProcessWorld4FEEF0) observe(event string) {
	w.events = append(w.events, event)
	if w.faultAt != 0 && len(w.events) == w.faultAt {
		panic(event)
	}
	if after := w.after[event]; after != nil {
		after()
	}
}

func (w *spellDurationProcessWorld4FEEF0) hooks() SpellDurationProcessHooks4FEEF0[uint64, uint64, uint64] {
	return SpellDurationProcessHooks4FEEF0[uint64, uint64, uint64]{
		LoadFirst: func() uint64 {
			value := w.head
			w.observe("head=" + spellDurationProcessRecordName4FEEF0(value))
			return value
		},
		LoadFlagsLowByte: func(record uint64) byte {
			value := byte(w.records[record].flags)
			w.observe(fmt.Sprintf("flags:%s=%02x", spellDurationProcessRecordName4FEEF0(record), value))
			return value
		},
		LoadNext: func(record uint64) uint64 {
			value := w.records[record].next
			w.observe(fmt.Sprintf("next:%s=%s", spellDurationProcessRecordName4FEEF0(record), spellDurationProcessRecordName4FEEF0(value)))
			return value
		},
		Destroy: func(record uint64) {
			w.observe("destroy:" + spellDurationProcessRecordName4FEEF0(record))
		},
		LoadCaster: func(record uint64) uint64 {
			value := w.records[record].caster
			w.observe(fmt.Sprintf("caster:%s=%s", spellDurationProcessRecordName4FEEF0(record), spellDurationProcessObjectName4FEEF0(value)))
			return value
		},
		LoadObjectFlags: func(object uint64) uint32 {
			value := w.objects[object]
			w.observe(fmt.Sprintf("object-flags:%s=%08x", spellDurationProcessObjectName4FEEF0(object), value))
			return value
		},
		StoreCaster: func(record, object uint64) {
			w.observe(fmt.Sprintf("store-caster:%s=%s", spellDurationProcessRecordName4FEEF0(record), spellDurationProcessObjectName4FEEF0(object)))
			w.records[record].caster = object
		},
		LoadObj12: func(record uint64) uint64 {
			value := w.records[record].obj12
			w.observe(fmt.Sprintf("obj12:%s=%s", spellDurationProcessRecordName4FEEF0(record), spellDurationProcessObjectName4FEEF0(value)))
			return value
		},
		LoadObjectFlagsLowByte: func(object uint64) byte {
			value := byte(w.objects[object])
			w.observe(fmt.Sprintf("object-flags-low:%s=%02x", spellDurationProcessObjectName4FEEF0(object), value))
			return value
		},
		StoreObj12: func(record, object uint64) {
			w.observe(fmt.Sprintf("store-obj12:%s=%s", spellDurationProcessRecordName4FEEF0(record), spellDurationProcessObjectName4FEEF0(object)))
			w.records[record].obj12 = object
		},
		LoadFlag20: func(record uint64) uint32 {
			value := w.records[record].flag20
			w.observe(fmt.Sprintf("flag20:%s=%08x", spellDurationProcessRecordName4FEEF0(record), value))
			return value
		},
		LoadObj24: func(record uint64) uint64 {
			value := w.records[record].obj24
			w.observe(fmt.Sprintf("obj24:%s=%s", spellDurationProcessRecordName4FEEF0(record), spellDurationProcessObjectName4FEEF0(value)))
			return value
		},
		StoreObj24: func(record, object uint64) {
			w.observe(fmt.Sprintf("store-obj24:%s=%s", spellDurationProcessRecordName4FEEF0(record), spellDurationProcessObjectName4FEEF0(object)))
			w.records[record].obj24 = object
		},
		LoadFrame68: func(record uint64) uint32 {
			value := w.records[record].frame68
			w.observe(fmt.Sprintf("frame68:%s=%08x", spellDurationProcessRecordName4FEEF0(record), value))
			return value
		},
		LoadFrame60: func(record uint64) uint32 {
			value := w.records[record].frame60
			w.observe(fmt.Sprintf("frame60:%s=%08x", spellDurationProcessRecordName4FEEF0(record), value))
			return value
		},
		LoadCurrentFrame: func() uint32 {
			value := w.currentFrame
			w.observe(fmt.Sprintf("current-frame=%08x", value))
			return value
		},
		LoadUpdate: func(record uint64) uint64 {
			value := w.records[record].update
			w.observe(fmt.Sprintf("update:%s=%s", spellDurationProcessRecordName4FEEF0(record), spellDurationProcessUpdateName4FEEF0(value)))
			return value
		},
		CallUpdate: func(update, record uint64) int32 {
			value := w.updates[update]
			w.observe(fmt.Sprintf("call-update:%s:%s=%08x", spellDurationProcessUpdateName4FEEF0(update), spellDurationProcessRecordName4FEEF0(record), uint32(value)))
			return value
		},
		Cancel: func(record uint64) {
			w.observe("cancel:" + spellDurationProcessRecordName4FEEF0(record))
		},
	}
}

func TestSpellDurationProcess4FEEF0ExactTrace(t *testing.T) {
	w := newSpellDurationProcessWorld4FEEF0()
	SpellDurationProcess4FEEF0(w.hooks())

	want := []string{
		"head=A",
		"flags:A=01", "next:A=B", "destroy:A",
		"flags:B=80", "next:B=C",
		"caster:B=dead-caster", "object-flags:dead-caster=00008000", "store-caster:B=nil",
		"obj12:B=dead-obj12", "object-flags-low:dead-obj12=20", "store-obj12:B=nil",
		"caster:B=nil", "flag20:B=00000000", "cancel:B",
		"flags:C=00", "next:C=D",
		"caster:C=live-caster", "object-flags:live-caster=80000000",
		"obj12:C=live-obj12", "object-flags-low:live-obj12=10", "caster:C=live-caster",
		"obj24:C=dead-obj24", "object-flags-low:dead-obj24=20", "store-obj24:C=nil",
		"frame68:C=0000000a", "frame60:C=00000001", "current-frame=0000000a", "cancel:C",
		"flags:D=02", "next:D=E",
		"caster:D=nil", "obj12:D=nil", "caster:D=nil", "flag20:D=00000001",
		"obj24:D=live-obj24", "object-flags-low:live-obj24=00",
		"frame68:D=00000007", "frame60:D=00000007",
		"update:D=update-D", "call-update:update-D:D=00000001", "cancel:D",
		"flags:E=00", "next:E=nil",
		"caster:E=live-caster", "object-flags:live-caster=80000000",
		"obj12:E=nil", "caster:E=live-caster", "obj24:E=nil",
		"frame68:E=0000000b", "frame60:E=00000007", "current-frame=0000000a", "update:E=nil",
	}
	if !reflect.DeepEqual(w.events, want) {
		t.Fatalf("trace mismatch\n got: %q\nwant: %q", w.events, want)
	}
}

func TestSpellDurationProcess4FEEF0EmptyStopsAfterHead(t *testing.T) {
	w := newSpellDurationProcessWorld4FEEF0()
	w.head = 0
	SpellDurationProcess4FEEF0(w.hooks())
	if !reflect.DeepEqual(w.events, []string{"head=nil"}) {
		t.Fatalf("events = %q, want only head load", w.events)
	}
}

func TestSpellDurationProcess4FEEF0CachesSuccessorFrameAndUpdate(t *testing.T) {
	w := newSpellDurationProcessWorld4FEEF0()
	w.head = spellDurationProcessRecordE4FEEF0
	record := w.records[spellDurationProcessRecordE4FEEF0]
	record.next = spellDurationProcessRecordD4FEEF0
	record.frame68 = math.MaxUint32
	record.frame60 = 0
	record.update = spellDurationProcessUpdateE4FEEF0
	w.currentFrame = math.MaxUint32 - 1
	w.after["next:E=D"] = func() {
		record.next = spellDurationProcessRecordC4FEEF0
	}
	w.after["frame68:E=ffffffff"] = func() {
		record.frame68 = 0
	}
	w.after["update:E=update-E"] = func() {
		record.update = spellDurationProcessUpdateD4FEEF0
		w.updates[spellDurationProcessUpdateE4FEEF0] = -1
	}
	w.records[spellDurationProcessRecordD4FEEF0] = &spellDurationProcessRecordState4FEEF0{flags: 1}

	SpellDurationProcess4FEEF0(w.hooks())
	wantSuffix := []string{
		"frame68:E=ffffffff", "frame60:E=00000000", "current-frame=fffffffe",
		"update:E=update-E", "call-update:update-E:E=ffffffff", "cancel:E",
		"flags:D=01", "next:D=nil", "destroy:D",
	}
	if len(w.events) < len(wantSuffix) || !reflect.DeepEqual(w.events[len(w.events)-len(wantSuffix):], wantSuffix) {
		t.Fatalf("events = %q, want cached-value/successor suffix %q", w.events, wantSuffix)
	}
}

func TestSpellDurationProcess4FEEF0FrameBranchesAreUnsigned(t *testing.T) {
	tests := []struct {
		name         string
		frame60      uint32
		frame68      uint32
		current      uint32
		wantCurrent  bool
		wantUpdate   bool
		wantCanceled bool
	}{
		{name: "equal-always-updates", frame60: 7, frame68: 7, current: math.MaxUint32, wantUpdate: true},
		{name: "expired-equal-current", frame60: 1, frame68: 10, current: 10, wantCurrent: true, wantCanceled: true},
		{name: "future", frame60: 1, frame68: math.MaxUint32, current: math.MaxUint32 - 1, wantCurrent: true, wantUpdate: true},
		{name: "wrapped-small-is-expired", frame60: math.MaxUint32, frame68: 0, current: math.MaxUint32 - 1, wantCurrent: true, wantCanceled: true},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			w := newSpellDurationProcessWorld4FEEF0()
			w.head = spellDurationProcessRecordE4FEEF0
			record := w.records[spellDurationProcessRecordE4FEEF0]
			record.frame60 = tc.frame60
			record.frame68 = tc.frame68
			record.update = spellDurationProcessUpdateE4FEEF0
			w.currentFrame = tc.current
			SpellDurationProcess4FEEF0(w.hooks())
			gotCurrent, gotUpdate, gotCanceled := false, false, false
			for _, event := range w.events {
				gotCurrent = gotCurrent || len(event) >= len("current-frame=") && event[:len("current-frame=")] == "current-frame="
				gotUpdate = gotUpdate || event == "update:E=update-E"
				gotCanceled = gotCanceled || event == "cancel:E"
			}
			if gotCurrent != tc.wantCurrent || gotUpdate != tc.wantUpdate || gotCanceled != tc.wantCanceled {
				t.Fatalf("events = %q; current/update/canceled = %t/%t/%t, want %t/%t/%t",
					w.events, gotCurrent, gotUpdate, gotCanceled, tc.wantCurrent, tc.wantUpdate, tc.wantCanceled)
			}
		})
	}
}

func TestSpellDurationProcess4FEEF0ObjectMasksAndShortCircuits(t *testing.T) {
	w := newSpellDurationProcessWorld4FEEF0()
	w.head = spellDurationProcessRecordE4FEEF0
	record := w.records[spellDurationProcessRecordE4FEEF0]
	record.caster = spellDurationProcessLiveCaster4FEEF0
	record.obj12 = spellDurationProcessLiveObj124FEEF0
	record.obj24 = spellDurationProcessLiveObj244FEEF0
	record.frame60 = 1
	record.frame68 = 2
	w.currentFrame = 2
	w.objects[spellDurationProcessLiveCaster4FEEF0] = 0x80000000
	w.objects[spellDurationProcessLiveObj124FEEF0] = 0x00008000
	w.objects[spellDurationProcessLiveObj244FEEF0] = 0x00008000

	SpellDurationProcess4FEEF0(w.hooks())
	if record.caster == 0 || record.obj12 == 0 || record.obj24 == 0 {
		t.Fatalf("nonmatching masks cleared pointers: caster/obj12/obj24 = %#x/%#x/%#x", record.caster, record.obj12, record.obj24)
	}
	if got := w.events[len(w.events)-1]; got != "cancel:E" {
		t.Fatalf("last event = %q, want expiry cancellation", got)
	}

	w = newSpellDurationProcessWorld4FEEF0()
	w.head = spellDurationProcessRecordB4FEEF0
	w.records[spellDurationProcessRecordB4FEEF0].next = 0
	SpellDurationProcess4FEEF0(w.hooks())
	for _, event := range w.events {
		if event == "obj24:B=live-obj24" || event == "update:B=update-E" {
			t.Fatalf("nil-caster/zero-Flag20 path observed forbidden event %q in %q", event, w.events)
		}
	}
}

func TestSpellDurationProcess4FEEF0UsesSavedSuccessorAfterCancel(t *testing.T) {
	w := newSpellDurationProcessWorld4FEEF0()
	w.head = spellDurationProcessRecordC4FEEF0
	w.records[spellDurationProcessRecordC4FEEF0].next = spellDurationProcessRecordD4FEEF0
	w.records[spellDurationProcessRecordD4FEEF0] = &spellDurationProcessRecordState4FEEF0{flags: 1}
	w.after["cancel:C"] = func() {
		w.head = spellDurationProcessRecordE4FEEF0
		w.records[spellDurationProcessRecordC4FEEF0].next = spellDurationProcessRecordE4FEEF0
	}
	SpellDurationProcess4FEEF0(w.hooks())
	if !reflect.DeepEqual(w.events[len(w.events)-3:], []string{"flags:D=01", "next:D=nil", "destroy:D"}) {
		t.Fatalf("events = %q, want saved successor D after cancellation", w.events)
	}
}

func TestSpellDurationProcess4FEEF0DoesNotAddCycleGuard(t *testing.T) {
	w := newSpellDurationProcessWorld4FEEF0()
	w.head = spellDurationProcessRecordA4FEEF0
	w.records[spellDurationProcessRecordA4FEEF0].next = spellDurationProcessRecordA4FEEF0
	destroys := 0
	sentinel := &struct{}{}
	w.after["destroy:A"] = func() {
		destroys++
		if destroys == 3 {
			panic(sentinel)
		}
	}
	var recovered any
	func() {
		defer func() { recovered = recover() }()
		SpellDurationProcess4FEEF0(w.hooks())
	}()
	if recovered != sentinel {
		t.Fatalf("recovered = %#v, want third-destroy sentinel", recovered)
	}
}

func TestSpellDurationProcess4FEEF0FaultPrefixes(t *testing.T) {
	baseline := newSpellDurationProcessWorld4FEEF0()
	SpellDurationProcess4FEEF0(baseline.hooks())
	want := append([]string(nil), baseline.events...)

	for faultAt := 1; faultAt <= len(want); faultAt++ {
		t.Run(fmt.Sprintf("fault-%d", faultAt), func(t *testing.T) {
			w := newSpellDurationProcessWorld4FEEF0()
			w.faultAt = faultAt
			var recovered any
			func() {
				defer func() { recovered = recover() }()
				SpellDurationProcess4FEEF0(w.hooks())
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
