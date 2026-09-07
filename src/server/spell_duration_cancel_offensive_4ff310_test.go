package server

import (
	"fmt"
	"math"
	"reflect"
	"testing"
)

const (
	spellDurationCancelOffensiveRecordA4FF310 = uint64(0x100000101)
	spellDurationCancelOffensiveRecordB4FF310 = uint64(0x200000202)
	spellDurationCancelOffensiveRecordC4FF310 = uint64(0x300000303)
	spellDurationCancelOffensiveRecordD4FF310 = uint64(0x400000404)
	spellDurationCancelOffensiveCaster4FF310  = uint64(0x1234567889abcdef)
	spellDurationCancelOffensiveAlias4FF310   = uint64(0x0000000089abcdef)
)

type spellDurationCancelOffensiveRecordState4FF310 struct {
	caster uint64
	next   uint64
	spell  int32
}

type spellDurationCancelOffensiveWorld4FF310 struct {
	events    []string
	after     map[string]func()
	faultAt   int
	head      uint64
	casterArg uint64
	records   map[uint64]*spellDurationCancelOffensiveRecordState4FF310
	flags     map[int32]uint32
}

func spellDurationCancelOffensiveRecordName4FF310(record uint64) string {
	switch record {
	case 0:
		return "nil"
	case spellDurationCancelOffensiveRecordA4FF310:
		return "A"
	case spellDurationCancelOffensiveRecordB4FF310:
		return "B"
	case spellDurationCancelOffensiveRecordC4FF310:
		return "C"
	case spellDurationCancelOffensiveRecordD4FF310:
		return "D"
	default:
		return fmt.Sprintf("%x", record)
	}
}

func (w *spellDurationCancelOffensiveWorld4FF310) observe(event string) {
	w.events = append(w.events, event)
	if w.faultAt != 0 && len(w.events) == w.faultAt {
		panic(event)
	}
	if after := w.after[event]; after != nil {
		after()
	}
}

func (w *spellDurationCancelOffensiveWorld4FF310) hooks() SpellDurationCancelOffensiveHooks4FF310[uint64, uint64] {
	return SpellDurationCancelOffensiveHooks4FF310[uint64, uint64]{
		LoadFirst: func() uint64 {
			value := w.head
			w.observe("head=" + spellDurationCancelOffensiveRecordName4FF310(value))
			return value
		},
		LoadCasterArg: func() uint64 {
			value := w.casterArg
			w.observe(fmt.Sprintf("caster-arg=%016x", value))
			return value
		},
		LoadCaster: func(record uint64) uint64 {
			value := w.records[record].caster
			w.observe(fmt.Sprintf("caster:%s=%016x", spellDurationCancelOffensiveRecordName4FF310(record), value))
			return value
		},
		LoadNext: func(record uint64) uint64 {
			value := w.records[record].next
			w.observe(fmt.Sprintf("next:%s=%s", spellDurationCancelOffensiveRecordName4FF310(record), spellDurationCancelOffensiveRecordName4FF310(value)))
			return value
		},
		LoadSpell: func(record uint64) int32 {
			value := w.records[record].spell
			w.observe(fmt.Sprintf("spell:%s=%08x", spellDurationCancelOffensiveRecordName4FF310(record), uint32(value)))
			return value
		},
		LoadSpellFlags: func(spellID int32) uint32 {
			value := w.flags[spellID]
			w.observe(fmt.Sprintf("flags:%08x=%08x", uint32(spellID), value))
			return value
		},
		Cancel: func(record uint64) {
			w.observe("cancel:" + spellDurationCancelOffensiveRecordName4FF310(record))
		},
	}
}

func newSpellDurationCancelOffensiveWorld4FF310() *spellDurationCancelOffensiveWorld4FF310 {
	return &spellDurationCancelOffensiveWorld4FF310{
		after:     make(map[string]func()),
		head:      spellDurationCancelOffensiveRecordA4FF310,
		casterArg: spellDurationCancelOffensiveCaster4FF310,
		records: map[uint64]*spellDurationCancelOffensiveRecordState4FF310{
			spellDurationCancelOffensiveRecordA4FF310: {
				caster: spellDurationCancelOffensiveAlias4FF310,
				next:   spellDurationCancelOffensiveRecordB4FF310,
				spell:  math.MinInt32,
			},
			spellDurationCancelOffensiveRecordB4FF310: {
				caster: spellDurationCancelOffensiveCaster4FF310,
				next:   spellDurationCancelOffensiveRecordC4FF310,
				spell:  math.MinInt32,
			},
			spellDurationCancelOffensiveRecordC4FF310: {
				caster: spellDurationCancelOffensiveCaster4FF310,
				next:   spellDurationCancelOffensiveRecordD4FF310,
				spell:  math.MaxInt32,
			},
			spellDurationCancelOffensiveRecordD4FF310: {
				caster: spellDurationCancelOffensiveCaster4FF310,
				spell:  -1,
			},
		},
		flags: map[int32]uint32{
			math.MinInt32: 0xffffff00,
			math.MaxInt32: 0x12340020,
			-1:            0x80000020,
		},
	}
}

func TestSpellDurationCancelOffensive4FF310ExactTraceAndNativeTokens(t *testing.T) {
	w := newSpellDurationCancelOffensiveWorld4FF310()
	SpellDurationCancelOffensive4FF310(w.hooks())

	want := []string{
		"head=A", "caster-arg=1234567889abcdef",
		"caster:A=0000000089abcdef", "next:A=B",
		"caster:B=1234567889abcdef", "next:B=C", "spell:B=80000000", "flags:80000000=ffffff00",
		"caster:C=1234567889abcdef", "next:C=D", "spell:C=7fffffff", "flags:7fffffff=12340020", "cancel:C",
		"caster:D=1234567889abcdef", "next:D=nil", "spell:D=ffffffff", "flags:ffffffff=80000020", "cancel:D",
	}
	if !reflect.DeepEqual(w.events, want) {
		t.Fatalf("events = %q, want %q", w.events, want)
	}
}

func TestSpellDurationCancelOffensive4FF310EmptyAvoidsCasterArgument(t *testing.T) {
	w := newSpellDurationCancelOffensiveWorld4FF310()
	w.head = 0
	SpellDurationCancelOffensive4FF310(w.hooks())
	if !reflect.DeepEqual(w.events, []string{"head=nil"}) {
		t.Fatalf("events = %q, want only the head load", w.events)
	}
}

func TestSpellDurationCancelOffensive4FF310UsesFlagsLowByteAndFullSignedSpellDword(t *testing.T) {
	tests := []struct {
		name  string
		flags uint32
		want  bool
	}{
		{"low-offensive", 0x00000020, true},
		{"low-offensive-with-upper-bytes", 0xffffff20, true},
		{"offensive-bit-in-second-byte-only", 0x00002000, false},
		{"upper-bytes-only", 0xffffff00, false},
		{"adjacent-low-bits", 0x0000001f, false},
		{"next-low-bit", 0x00000040, false},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			w := newSpellDurationCancelOffensiveWorld4FF310()
			w.records = map[uint64]*spellDurationCancelOffensiveRecordState4FF310{
				spellDurationCancelOffensiveRecordA4FF310: {
					caster: spellDurationCancelOffensiveCaster4FF310,
					spell:  math.MinInt32,
				},
			}
			w.flags = map[int32]uint32{math.MinInt32: tc.flags}
			SpellDurationCancelOffensive4FF310(w.hooks())
			got := w.events[len(w.events)-1] == "cancel:A"
			if got != tc.want {
				t.Fatalf("flags %#08x canceled = %t, want %t; events = %q", tc.flags, got, tc.want, w.events)
			}
			if !containsString4FF310(w.events, "flags:80000000="+fmt.Sprintf("%08x", tc.flags)) {
				t.Fatalf("events = %q, want exact signed spell dword at flags callback", w.events)
			}
		})
	}
}

func containsString4FF310(values []string, want string) bool {
	for _, value := range values {
		if value == want {
			return true
		}
	}
	return false
}

func TestSpellDurationCancelOffensive4FF310CachesFieldsAndSavedSuccessor(t *testing.T) {
	w := newSpellDurationCancelOffensiveWorld4FF310()
	w.records = map[uint64]*spellDurationCancelOffensiveRecordState4FF310{
		spellDurationCancelOffensiveRecordA4FF310: {
			caster: spellDurationCancelOffensiveCaster4FF310,
			next:   spellDurationCancelOffensiveRecordB4FF310,
			spell:  1,
		},
		spellDurationCancelOffensiveRecordB4FF310: {
			caster: spellDurationCancelOffensiveCaster4FF310,
			spell:  2,
		},
		spellDurationCancelOffensiveRecordC4FF310: {
			caster: spellDurationCancelOffensiveCaster4FF310,
			spell:  3,
		},
	}
	w.flags = map[int32]uint32{math.MinInt32: 0x20, 3: 0}
	w.after["head=A"] = func() {
		w.head = spellDurationCancelOffensiveRecordB4FF310
	}
	w.after["caster-arg=1234567889abcdef"] = func() {
		w.casterArg = spellDurationCancelOffensiveAlias4FF310
	}
	w.after["caster:A=1234567889abcdef"] = func() {
		w.records[spellDurationCancelOffensiveRecordA4FF310].next = spellDurationCancelOffensiveRecordC4FF310
	}
	w.after["next:A=C"] = func() {
		w.records[spellDurationCancelOffensiveRecordA4FF310].caster = spellDurationCancelOffensiveAlias4FF310
		w.records[spellDurationCancelOffensiveRecordA4FF310].spell = math.MinInt32
	}
	w.after["flags:80000000=00000020"] = func() {
		w.records[spellDurationCancelOffensiveRecordA4FF310].next = spellDurationCancelOffensiveRecordB4FF310
	}
	w.after["cancel:A"] = func() {
		delete(w.records, spellDurationCancelOffensiveRecordA4FF310)
	}

	SpellDurationCancelOffensive4FF310(w.hooks())
	want := []string{
		"head=A", "caster-arg=1234567889abcdef",
		"caster:A=1234567889abcdef", "next:A=C", "spell:A=80000000", "flags:80000000=00000020", "cancel:A",
		"caster:C=1234567889abcdef", "next:C=nil", "spell:C=00000003", "flags:00000003=00000000",
	}
	if !reflect.DeepEqual(w.events, want) {
		t.Fatalf("events = %q, want cached-field/saved-successor trace %q", w.events, want)
	}
}

func TestSpellDurationCancelOffensive4FF310NilCasterIsValidIdentity(t *testing.T) {
	w := newSpellDurationCancelOffensiveWorld4FF310()
	w.casterArg = 0
	w.records = map[uint64]*spellDurationCancelOffensiveRecordState4FF310{
		spellDurationCancelOffensiveRecordA4FF310: {spell: 7},
	}
	w.flags = map[int32]uint32{7: 0x20}
	SpellDurationCancelOffensive4FF310(w.hooks())
	if got := w.events[len(w.events)-1]; got != "cancel:A" {
		t.Fatalf("last event = %q, want nil-caster identity cancellation", got)
	}
}

func TestSpellDurationCancelOffensive4FF310DoesNotAddCycleGuard(t *testing.T) {
	w := newSpellDurationCancelOffensiveWorld4FF310()
	w.records = map[uint64]*spellDurationCancelOffensiveRecordState4FF310{
		spellDurationCancelOffensiveRecordA4FF310: {
			caster: spellDurationCancelOffensiveAlias4FF310,
			next:   spellDurationCancelOffensiveRecordA4FF310,
		},
	}
	nextLoads := 0
	sentinel := &struct{}{}
	w.after["next:A=A"] = func() {
		nextLoads++
		if nextLoads == 3 {
			panic(sentinel)
		}
	}

	var recovered any
	func() {
		defer func() { recovered = recover() }()
		SpellDurationCancelOffensive4FF310(w.hooks())
	}()
	if recovered != sentinel {
		t.Fatalf("recovered = %#v, want third-next sentinel", recovered)
	}
	want := []string{
		"head=A", "caster-arg=1234567889abcdef",
		"caster:A=0000000089abcdef", "next:A=A",
		"caster:A=0000000089abcdef", "next:A=A",
		"caster:A=0000000089abcdef", "next:A=A",
	}
	if !reflect.DeepEqual(w.events, want) {
		t.Fatalf("events = %q, want repeated cycle prefix %q", w.events, want)
	}
}

func TestSpellDurationCancelOffensive4FF310FaultPrefixes(t *testing.T) {
	baseline := newSpellDurationCancelOffensiveWorld4FF310()
	SpellDurationCancelOffensive4FF310(baseline.hooks())
	want := append([]string(nil), baseline.events...)

	for faultAt := 1; faultAt <= len(want); faultAt++ {
		t.Run(fmt.Sprintf("fault-%d", faultAt), func(t *testing.T) {
			w := newSpellDurationCancelOffensiveWorld4FF310()
			w.faultAt = faultAt
			var recovered any
			func() {
				defer func() { recovered = recover() }()
				SpellDurationCancelOffensive4FF310(w.hooks())
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
