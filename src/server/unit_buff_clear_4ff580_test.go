package server

import (
	"fmt"
	"reflect"
	"testing"
)

const unitBuffClearObject4FF580 = uint64(0x1234567889abcdef)

type unitBuffClearWorld4FF580 struct {
	events    []string
	after     map[string]func()
	faultAt   int
	unit      uint64
	buffs     uint32
	durations [32]uint16
	powers    [32]uint8
}

func newUnitBuffClearWorld4FF580() *unitBuffClearWorld4FF580 {
	w := &unitBuffClearWorld4FF580{
		after: make(map[string]func()),
		unit:  unitBuffClearObject4FF580,
		buffs: 0xffffffff,
	}
	for i := range w.durations {
		w.durations[i] = uint16(0x8000 + i)
		w.powers[i] = uint8(0x80 + i)
	}
	return w
}

func (w *unitBuffClearWorld4FF580) observe(event string) {
	w.events = append(w.events, event)
	if w.faultAt != 0 && len(w.events) == w.faultAt {
		panic(event)
	}
	if after := w.after[event]; after != nil {
		after()
	}
}

func (w *unitBuffClearWorld4FF580) hooks() UnitBuffClearHooks4FF580[uint64] {
	return UnitBuffClearHooks4FF580[uint64]{
		LoadUnitArg: func() uint64 {
			unit := w.unit
			w.observe(fmt.Sprintf("unit=%016x", unit))
			return unit
		},
		SetBuffFlags: func(unit uint64, flags uint32) {
			w.observe(fmt.Sprintf("set:%016x:%08x", unit, flags))
			w.buffs = flags
		},
		StoreDuration: func(unit uint64, buff int32, value uint16) {
			w.observe(fmt.Sprintf("duration:%016x:%02x<=%04x", unit, uint32(buff), value))
			w.durations[buff] = value
		},
		StorePower: func(unit uint64, buff int32, value uint8) {
			w.observe(fmt.Sprintf("power:%016x:%02x<=%02x", unit, uint32(buff), value))
			w.powers[buff] = value
		},
	}
}

func unitBuffClearTrace4FF580(unit uint64) []string {
	want := []string{
		fmt.Sprintf("unit=%016x", unit),
		fmt.Sprintf("set:%016x:00000000", unit),
	}
	for buff := int32(0); buff < unitBuffClearSlots4FF580; buff++ {
		want = append(want,
			fmt.Sprintf("duration:%016x:%02x<=0000", unit, uint32(buff)),
			fmt.Sprintf("power:%016x:%02x<=00", unit, uint32(buff)),
		)
	}
	return want
}

func TestUnitBuffClear4FF580ExactTraceAndWidths(t *testing.T) {
	w := newUnitBuffClearWorld4FF580()
	UnitBuffClear4FF580(w.hooks())

	if want := unitBuffClearTrace4FF580(unitBuffClearObject4FF580); !reflect.DeepEqual(w.events, want) {
		t.Fatalf("events = %q, want exact flag/duration/power trace %q", w.events, want)
	}
	if w.buffs != 0 {
		t.Fatalf("buff flags = %#08x, want zero", w.buffs)
	}
	if w.durations != [32]uint16{} || w.powers != [32]uint8{} {
		t.Fatalf("buff storage was not cleared exactly: durations=%#v powers=%#v", w.durations, w.powers)
	}
}

func TestUnitBuffClear4FF580FlagCallbackPrecedesAndCannotSurviveStores(t *testing.T) {
	w := newUnitBuffClearWorld4FF580()
	h := w.hooks()
	originalSet := h.SetBuffFlags
	h.SetBuffFlags = func(unit uint64, flags uint32) {
		originalSet(unit, flags)
		for i := range w.durations {
			w.durations[i] = 0xffff
			w.powers[i] = 0xff
		}
	}
	UnitBuffClear4FF580(h)
	if w.durations != [32]uint16{} || w.powers != [32]uint8{} {
		t.Fatal("stores did not run after the flag callback")
	}
	if got := w.events[:3]; !reflect.DeepEqual(got, []string{
		"unit=1234567889abcdef",
		"set:1234567889abcdef:00000000",
		"duration:1234567889abcdef:00<=0000",
	}) {
		t.Fatalf("trace prefix = %q, want callback before first duration store", got)
	}
}

func TestUnitBuffClear4FF580CachesUnitAcrossMutation(t *testing.T) {
	w := newUnitBuffClearWorld4FF580()
	w.after["unit=1234567889abcdef"] = func() { w.unit = 0 }
	UnitBuffClear4FF580(w.hooks())
	if want := unitBuffClearTrace4FF580(unitBuffClearObject4FF580); !reflect.DeepEqual(w.events, want) {
		t.Fatalf("events = %q, want cached native-width unit trace %q", w.events, want)
	}
}

func TestUnitBuffClear4FF580DoesNotShortCircuitNullToken(t *testing.T) {
	w := newUnitBuffClearWorld4FF580()
	w.unit = 0
	UnitBuffClear4FF580(w.hooks())
	if want := unitBuffClearTrace4FF580(0); !reflect.DeepEqual(w.events, want) {
		t.Fatalf("events = %q, want unguarded null-token trace %q", w.events, want)
	}
}

func TestUnitBuffClear4FF580AllFaultPrefixes(t *testing.T) {
	baseline := newUnitBuffClearWorld4FF580()
	UnitBuffClear4FF580(baseline.hooks())
	want := append([]string(nil), baseline.events...)
	if len(want) != 66 {
		t.Fatalf("observable steps = %d, want 66", len(want))
	}

	for faultAt := 1; faultAt <= len(want); faultAt++ {
		t.Run(fmt.Sprintf("fault-%d", faultAt), func(t *testing.T) {
			w := newUnitBuffClearWorld4FF580()
			w.faultAt = faultAt
			var recovered any
			func() {
				defer func() { recovered = recover() }()
				UnitBuffClear4FF580(w.hooks())
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
