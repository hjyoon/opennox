package server

import (
	"fmt"
	"math"
	"reflect"
	"testing"
)

const (
	spellDurationDuplicateRecordA4FEE50 = uint64(0x100000101)
	spellDurationDuplicateRecordB4FEE50 = uint64(0x200000202)
	spellDurationDuplicateRecordC4FEE50 = uint64(0x300000303)
	spellDurationDuplicateRecordD4FEE50 = uint64(0x400000404)
	spellDurationDuplicateRecordE4FEE50 = uint64(0x500000505)
	spellDurationDuplicateCaster4FEE50  = uint64(0x1234567889abcdef)
	spellDurationDuplicateAlias4FEE50   = uint64(0x0000000089abcdef)
)

type spellDurationDuplicateRecordState4FEE50 struct {
	flag20  uint32
	spell   uint32
	caster  uint64
	flags88 uint32
	next    uint64
}

type spellDurationDuplicateWorld4FEE50 struct {
	events    []string
	after     map[string]func()
	faultAt   int
	head      uint64
	casterArg uint64
	spellArg  uint32
	records   map[uint64]*spellDurationDuplicateRecordState4FEE50
}

func spellDurationDuplicateRecordName4FEE50(record uint64) string {
	switch record {
	case 0:
		return "nil"
	case spellDurationDuplicateRecordA4FEE50:
		return "A"
	case spellDurationDuplicateRecordB4FEE50:
		return "B"
	case spellDurationDuplicateRecordC4FEE50:
		return "C"
	case spellDurationDuplicateRecordD4FEE50:
		return "D"
	case spellDurationDuplicateRecordE4FEE50:
		return "E"
	default:
		return fmt.Sprintf("%x", record)
	}
}

func (w *spellDurationDuplicateWorld4FEE50) observe(event string) {
	w.events = append(w.events, event)
	if w.faultAt != 0 && len(w.events) == w.faultAt {
		panic(event)
	}
	if after := w.after[event]; after != nil {
		after()
	}
}

func (w *spellDurationDuplicateWorld4FEE50) hooks() spellDurationDuplicateHooks4FEE50[uint64, uint64] {
	return spellDurationDuplicateHooks4FEE50[uint64, uint64]{
		loadHead: func() uint64 {
			value := w.head
			w.observe("head=" + spellDurationDuplicateRecordName4FEE50(value))
			return value
		},
		loadCasterArg: func() uint64 {
			value := w.casterArg
			w.observe(fmt.Sprintf("caster-arg=%016x", value))
			return value
		},
		loadSpellArg: func() uint32 {
			value := w.spellArg
			w.observe(fmt.Sprintf("spell-arg=%08x", value))
			return value
		},
		loadFlag20: func(record uint64) uint32 {
			value := w.records[record].flag20
			w.observe(fmt.Sprintf("flag:%s=%08x", spellDurationDuplicateRecordName4FEE50(record), value))
			return value
		},
		loadSpell: func(record uint64) uint32 {
			value := w.records[record].spell
			w.observe(fmt.Sprintf("spell:%s=%08x", spellDurationDuplicateRecordName4FEE50(record), value))
			return value
		},
		loadCaster: func(record uint64) uint64 {
			value := w.records[record].caster
			w.observe(fmt.Sprintf("caster:%s=%016x", spellDurationDuplicateRecordName4FEE50(record), value))
			return value
		},
		loadFlagsLowByte: func(record uint64) byte {
			value := byte(w.records[record].flags88)
			w.observe(fmt.Sprintf("flags:%s=%02x", spellDurationDuplicateRecordName4FEE50(record), value))
			return value
		},
		loadNext: func(record uint64) uint64 {
			value := w.records[record].next
			w.observe(fmt.Sprintf("next:%s=%s", spellDurationDuplicateRecordName4FEE50(record), spellDurationDuplicateRecordName4FEE50(value)))
			return value
		},
	}
}

func newSpellDurationDuplicateWorld4FEE50() *spellDurationDuplicateWorld4FEE50 {
	return &spellDurationDuplicateWorld4FEE50{
		after:     make(map[string]func()),
		head:      spellDurationDuplicateRecordA4FEE50,
		casterArg: spellDurationDuplicateCaster4FEE50,
		spellArg:  math.MaxUint32,
		records: map[uint64]*spellDurationDuplicateRecordState4FEE50{
			spellDurationDuplicateRecordA4FEE50: {
				flag20: 0x00000100,
				next:   spellDurationDuplicateRecordB4FEE50,
			},
			spellDurationDuplicateRecordB4FEE50: {
				spell: 0x7fffffff,
				next:  spellDurationDuplicateRecordC4FEE50,
			},
			spellDurationDuplicateRecordC4FEE50: {
				spell:  math.MaxUint32,
				caster: spellDurationDuplicateAlias4FEE50,
				next:   spellDurationDuplicateRecordD4FEE50,
			},
			spellDurationDuplicateRecordD4FEE50: {
				spell:   math.MaxUint32,
				caster:  spellDurationDuplicateCaster4FEE50,
				flags88: 0xabcdef01,
				next:    spellDurationDuplicateRecordE4FEE50,
			},
			spellDurationDuplicateRecordE4FEE50: {
				spell:   math.MaxUint32,
				caster:  spellDurationDuplicateCaster4FEE50,
				flags88: 0xffffff80,
			},
		},
	}
}

func TestSpellDurationDuplicate4FEE50ExactShortCircuitTraceAndNativeTokens(t *testing.T) {
	w := newSpellDurationDuplicateWorld4FEE50()

	if got := spellDurationDuplicate4FEE50(w.hooks()); got != 1 {
		t.Fatalf("result = %d, want canonical one", got)
	}
	want := []string{
		"head=A", "caster-arg=1234567889abcdef", "spell-arg=ffffffff",
		"flag:A=00000100", "next:A=B",
		"flag:B=00000000", "spell:B=7fffffff", "next:B=C",
		"flag:C=00000000", "spell:C=ffffffff", "caster:C=0000000089abcdef", "next:C=D",
		"flag:D=00000000", "spell:D=ffffffff", "caster:D=1234567889abcdef", "flags:D=01", "next:D=E",
		"flag:E=00000000", "spell:E=ffffffff", "caster:E=1234567889abcdef", "flags:E=80",
	}
	if !reflect.DeepEqual(w.events, want) {
		t.Fatalf("events = %q, want %q", w.events, want)
	}
}

func TestSpellDurationDuplicate4FEE50EmptyAndExhaustedReturnCanonicalZero(t *testing.T) {
	t.Run("empty avoids arguments", func(t *testing.T) {
		w := newSpellDurationDuplicateWorld4FEE50()
		w.head = 0
		if got := spellDurationDuplicate4FEE50(w.hooks()); got != 0 {
			t.Fatalf("result = %d, want canonical zero", got)
		}
		if !reflect.DeepEqual(w.events, []string{"head=nil"}) {
			t.Fatalf("events = %q, want only head load", w.events)
		}
	})

	t.Run("exhausted", func(t *testing.T) {
		w := newSpellDurationDuplicateWorld4FEE50()
		w.records[spellDurationDuplicateRecordD4FEE50].next = 0
		if got := spellDurationDuplicate4FEE50(w.hooks()); got != 0 {
			t.Fatalf("result = %d, want canonical zero", got)
		}
		if got := w.events[len(w.events)-1]; got != "next:D=nil" {
			t.Fatalf("last event = %q, want exhausted next load", got)
		}
	})
}

func TestSpellDurationDuplicate4FEE50CachesArgumentsAndLoadsFieldsLive(t *testing.T) {
	w := newSpellDurationDuplicateWorld4FEE50()
	w.head = spellDurationDuplicateRecordA4FEE50
	w.spellArg = 31
	w.records = map[uint64]*spellDurationDuplicateRecordState4FEE50{
		spellDurationDuplicateRecordA4FEE50: {flags88: 1},
	}
	w.after["caster-arg=1234567889abcdef"] = func() {
		w.casterArg = spellDurationDuplicateAlias4FEE50
	}
	w.after["spell-arg=0000001f"] = func() {
		w.spellArg = 32
	}
	w.after["flag:A=00000000"] = func() {
		w.records[spellDurationDuplicateRecordA4FEE50].spell = 31
	}
	w.after["spell:A=0000001f"] = func() {
		w.records[spellDurationDuplicateRecordA4FEE50].caster = spellDurationDuplicateCaster4FEE50
	}
	w.after["caster:A=1234567889abcdef"] = func() {
		w.records[spellDurationDuplicateRecordA4FEE50].flags88 = 0xffffff80
	}
	w.after["flags:A=80"] = func() {
		w.records[spellDurationDuplicateRecordA4FEE50].next = spellDurationDuplicateRecordB4FEE50
	}

	if got := spellDurationDuplicate4FEE50(w.hooks()); got != 1 {
		t.Fatalf("result = %d, want cached-argument/live-field match", got)
	}
	want := []string{
		"head=A", "caster-arg=1234567889abcdef", "spell-arg=0000001f",
		"flag:A=00000000", "spell:A=0000001f", "caster:A=1234567889abcdef", "flags:A=80",
	}
	if !reflect.DeepEqual(w.events, want) {
		t.Fatalf("events = %q, want %q", w.events, want)
	}
}

func TestSpellDurationDuplicate4FEE50RejectedFlagsThenLoadsLiveNext(t *testing.T) {
	w := newSpellDurationDuplicateWorld4FEE50()
	w.spellArg = 31
	w.records = map[uint64]*spellDurationDuplicateRecordState4FEE50{
		spellDurationDuplicateRecordA4FEE50: {
			spell: 31, caster: spellDurationDuplicateCaster4FEE50, flags88: 1,
			next: spellDurationDuplicateRecordB4FEE50,
		},
		spellDurationDuplicateRecordC4FEE50: {
			spell: 31, caster: spellDurationDuplicateCaster4FEE50, flags88: 0x80,
		},
	}
	w.after["flags:A=01"] = func() {
		w.records[spellDurationDuplicateRecordA4FEE50].flags88 = 0
		w.records[spellDurationDuplicateRecordA4FEE50].next = spellDurationDuplicateRecordC4FEE50
	}

	if got := spellDurationDuplicate4FEE50(w.hooks()); got != 1 {
		t.Fatalf("result = %d, want match through live successor", got)
	}
	want := []string{
		"head=A", "caster-arg=1234567889abcdef", "spell-arg=0000001f",
		"flag:A=00000000", "spell:A=0000001f", "caster:A=1234567889abcdef", "flags:A=01", "next:A=C",
		"flag:C=00000000", "spell:C=0000001f", "caster:C=1234567889abcdef", "flags:C=80",
	}
	if !reflect.DeepEqual(w.events, want) {
		t.Fatalf("events = %q, want %q", w.events, want)
	}
}

func TestSpellDurationDuplicate4FEE50NilCasterIsValidIdentity(t *testing.T) {
	w := newSpellDurationDuplicateWorld4FEE50()
	w.casterArg = 0
	w.spellArg = 43
	w.records = map[uint64]*spellDurationDuplicateRecordState4FEE50{
		spellDurationDuplicateRecordA4FEE50: {spell: 43, caster: 0, flags88: 0xfffffffe},
	}
	if got := spellDurationDuplicate4FEE50(w.hooks()); got != 1 {
		t.Fatalf("result = %d, want nil-caster identity match", got)
	}
}

func TestSpellDurationDuplicate4FEE50DoesNotAddCycleGuard(t *testing.T) {
	w := newSpellDurationDuplicateWorld4FEE50()
	w.records = map[uint64]*spellDurationDuplicateRecordState4FEE50{
		spellDurationDuplicateRecordA4FEE50: {flag20: 1, next: spellDurationDuplicateRecordA4FEE50},
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
		spellDurationDuplicate4FEE50(w.hooks())
	}()
	if recovered != sentinel {
		t.Fatalf("recovered = %#v, want third-next sentinel", recovered)
	}
	want := []string{
		"head=A", "caster-arg=1234567889abcdef", "spell-arg=ffffffff",
		"flag:A=00000001", "next:A=A",
		"flag:A=00000001", "next:A=A",
		"flag:A=00000001", "next:A=A",
	}
	if !reflect.DeepEqual(w.events, want) {
		t.Fatalf("events = %q, want repeated cycle prefix %q", w.events, want)
	}
}

func TestSpellDurationDuplicate4FEE50FaultPrefixes(t *testing.T) {
	baseline := newSpellDurationDuplicateWorld4FEE50()
	spellDurationDuplicate4FEE50(baseline.hooks())
	want := append([]string(nil), baseline.events...)

	for faultAt := 1; faultAt <= len(want); faultAt++ {
		t.Run(fmt.Sprintf("fault-%d", faultAt), func(t *testing.T) {
			w := newSpellDurationDuplicateWorld4FEE50()
			w.faultAt = faultAt
			var recovered any
			func() {
				defer func() { recovered = recover() }()
				spellDurationDuplicate4FEE50(w.hooks())
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
