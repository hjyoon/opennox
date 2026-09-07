package server

import (
	"fmt"
	"math"
	"reflect"
	"testing"
)

const unitBuffTimerObject4FF550 = uint64(0x1234567889abcdef)

type unitBuffTimerWorld4FF550 struct {
	events   []string
	after    map[string]func()
	faultAt  int
	buff     int32
	unit     uint64
	duration uint16
}

func (w *unitBuffTimerWorld4FF550) observe(event string) {
	w.events = append(w.events, event)
	if w.faultAt != 0 && len(w.events) == w.faultAt {
		panic(event)
	}
	if after := w.after[event]; after != nil {
		after()
	}
}

func (w *unitBuffTimerWorld4FF550) hooks() UnitBuffTimerHooks4FF550[uint64] {
	return UnitBuffTimerHooks4FF550[uint64]{
		LoadBuffArg: func() int32 {
			value := w.buff
			w.observe(fmt.Sprintf("buff=%08x", uint32(value)))
			return value
		},
		LoadUnitArg: func() uint64 {
			value := w.unit
			w.observe(fmt.Sprintf("unit=%016x", value))
			return value
		},
		LoadDuration: func(unit uint64, buff int32) uint16 {
			value := w.duration
			w.observe(fmt.Sprintf("duration:%016x:%08x=%04x", unit, uint32(buff), value))
			return value
		},
	}
}

func TestUnitBuffTimer4FF550ExactLoadOrderAndWidths(t *testing.T) {
	w := &unitBuffTimerWorld4FF550{
		buff:     -1,
		unit:     unitBuffTimerObject4FF550,
		duration: 0xfedc,
	}
	if got := UnitBuffTimer4FF550(w.hooks()); got != 0xfedc {
		t.Fatalf("result = %#08x, want zero-extended word 0x0000fedc", got)
	}
	want := []string{
		"buff=ffffffff",
		"unit=1234567889abcdef",
		"duration:1234567889abcdef:ffffffff=fedc",
	}
	if !reflect.DeepEqual(w.events, want) {
		t.Fatalf("events = %q, want exact load order and widths %q", w.events, want)
	}
}

func TestUnitBuffTimer4FF550CachesEarlierLoadsAcrossMutation(t *testing.T) {
	w := &unitBuffTimerWorld4FF550{
		after:    make(map[string]func()),
		buff:     7,
		unit:     unitBuffTimerObject4FF550,
		duration: math.MaxUint16,
	}
	w.after["buff=00000007"] = func() {
		w.buff = math.MinInt32
	}
	w.after["unit=1234567889abcdef"] = func() {
		w.unit = 0
	}
	if got := UnitBuffTimer4FF550(w.hooks()); got != math.MaxUint16 {
		t.Fatalf("result = %#08x, want zero-extended 0x0000ffff", got)
	}
	want := []string{
		"buff=00000007",
		"unit=1234567889abcdef",
		"duration:1234567889abcdef:00000007=ffff",
	}
	if !reflect.DeepEqual(w.events, want) {
		t.Fatalf("events = %q, want cached argument trace %q", w.events, want)
	}
}

func TestUnitBuffTimer4FF550ForwardsEveryDwordIndexWithoutGuard(t *testing.T) {
	for _, buff := range []int32{
		math.MinInt32, -33, -1, 0, 31, 32, 33, math.MaxInt32,
	} {
		t.Run(fmt.Sprintf("buff-%08x", uint32(buff)), func(t *testing.T) {
			w := &unitBuffTimerWorld4FF550{
				buff:     buff,
				unit:     unitBuffTimerObject4FF550,
				duration: uint16(uint32(buff)),
			}
			want := uint32(uint16(uint32(buff)))
			if got := UnitBuffTimer4FF550(w.hooks()); got != want {
				t.Fatalf("result = %#08x, want %#08x", got, want)
			}
			last := w.events[len(w.events)-1]
			wantLast := fmt.Sprintf("duration:1234567889abcdef:%08x=%04x", uint32(buff), uint16(uint32(buff)))
			if last != wantLast {
				t.Fatalf("duration load = %q, want unchecked full-dword index %q", last, wantLast)
			}
		})
	}
}

func TestUnitBuffTimer4FF550DoesNotShortCircuitNullToken(t *testing.T) {
	w := &unitBuffTimerWorld4FF550{buff: 3, unit: 0, duration: 0x8000}
	if got := UnitBuffTimer4FF550(w.hooks()); got != 0x8000 {
		t.Fatalf("result = %#08x, want zero-extended word 0x00008000", got)
	}
	want := []string{
		"buff=00000003",
		"unit=0000000000000000",
		"duration:0000000000000000:00000003=8000",
	}
	if !reflect.DeepEqual(w.events, want) {
		t.Fatalf("events = %q, want unguarded null-token trace %q", w.events, want)
	}
}

func TestUnitBuffTimer4FF550AllFaultPrefixes(t *testing.T) {
	baseline := &unitBuffTimerWorld4FF550{
		buff:     31,
		unit:     unitBuffTimerObject4FF550,
		duration: math.MaxUint16,
	}
	UnitBuffTimer4FF550(baseline.hooks())
	want := append([]string(nil), baseline.events...)

	for faultAt := 1; faultAt <= len(want); faultAt++ {
		t.Run(fmt.Sprintf("fault-%d", faultAt), func(t *testing.T) {
			w := &unitBuffTimerWorld4FF550{
				faultAt:  faultAt,
				buff:     31,
				unit:     unitBuffTimerObject4FF550,
				duration: math.MaxUint16,
			}
			var recovered any
			func() {
				defer func() { recovered = recover() }()
				UnitBuffTimer4FF550(w.hooks())
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
