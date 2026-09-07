package server

import (
	"fmt"
	"math"
	"reflect"
	"testing"
)

const unitBuffPowerObject4FF570 = uint64(0x1234567889abcdef)

type unitBuffPowerWorld4FF570 struct {
	events  []string
	after   map[string]func()
	faultAt int
	buff    int32
	unit    uint64
	power   uint8
}

func (w *unitBuffPowerWorld4FF570) observe(event string) {
	w.events = append(w.events, event)
	if w.faultAt != 0 && len(w.events) == w.faultAt {
		panic(event)
	}
	if after := w.after[event]; after != nil {
		after()
	}
}

func (w *unitBuffPowerWorld4FF570) hooks() UnitBuffPowerHooks4FF570[uint64] {
	return UnitBuffPowerHooks4FF570[uint64]{
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
		LoadPower: func(unit uint64, buff int32) uint8 {
			value := w.power
			w.observe(fmt.Sprintf("power:%016x:%08x=%02x", unit, uint32(buff), value))
			return value
		},
	}
}

func TestUnitBuffPower4FF570ExactLoadOrderAndWidths(t *testing.T) {
	w := &unitBuffPowerWorld4FF570{
		buff:  -1,
		unit:  unitBuffPowerObject4FF570,
		power: 0xfe,
	}
	if got := UnitBuffPower4FF570(w.hooks()); got != 0xfe {
		t.Fatalf("result = %#02x, want exact byte 0xfe", got)
	}
	want := []string{
		"buff=ffffffff",
		"unit=1234567889abcdef",
		"power:1234567889abcdef:ffffffff=fe",
	}
	if !reflect.DeepEqual(w.events, want) {
		t.Fatalf("events = %q, want exact load order and widths %q", w.events, want)
	}
}

func TestUnitBuffPower4FF570CachesEarlierLoadsAcrossMutation(t *testing.T) {
	w := &unitBuffPowerWorld4FF570{
		after: make(map[string]func()),
		buff:  7,
		unit:  unitBuffPowerObject4FF570,
		power: math.MaxUint8,
	}
	w.after["buff=00000007"] = func() {
		w.buff = math.MinInt32
	}
	w.after["unit=1234567889abcdef"] = func() {
		w.unit = 0
	}
	if got := UnitBuffPower4FF570(w.hooks()); got != math.MaxUint8 {
		t.Fatalf("result = %#02x, want exact byte 0xff", got)
	}
	want := []string{
		"buff=00000007",
		"unit=1234567889abcdef",
		"power:1234567889abcdef:00000007=ff",
	}
	if !reflect.DeepEqual(w.events, want) {
		t.Fatalf("events = %q, want cached argument trace %q", w.events, want)
	}
}

func TestUnitBuffPower4FF570ForwardsEveryDwordIndexWithoutGuard(t *testing.T) {
	for _, buff := range []int32{
		math.MinInt32, -33, -1, 0, 31, 32, 33, math.MaxInt32,
	} {
		t.Run(fmt.Sprintf("buff-%08x", uint32(buff)), func(t *testing.T) {
			w := &unitBuffPowerWorld4FF570{
				buff:  buff,
				unit:  unitBuffPowerObject4FF570,
				power: uint8(uint32(buff)),
			}
			want := uint8(uint32(buff))
			if got := UnitBuffPower4FF570(w.hooks()); got != want {
				t.Fatalf("result = %#02x, want %#02x", got, want)
			}
			last := w.events[len(w.events)-1]
			wantLast := fmt.Sprintf("power:1234567889abcdef:%08x=%02x", uint32(buff), want)
			if last != wantLast {
				t.Fatalf("power load = %q, want unchecked full-dword index %q", last, wantLast)
			}
		})
	}
}

func TestUnitBuffPower4FF570DoesNotShortCircuitNullToken(t *testing.T) {
	w := &unitBuffPowerWorld4FF570{buff: 3, unit: 0, power: 0x80}
	if got := UnitBuffPower4FF570(w.hooks()); got != 0x80 {
		t.Fatalf("result = %#02x, want exact byte 0x80", got)
	}
	want := []string{
		"buff=00000003",
		"unit=0000000000000000",
		"power:0000000000000000:00000003=80",
	}
	if !reflect.DeepEqual(w.events, want) {
		t.Fatalf("events = %q, want unguarded null-token trace %q", w.events, want)
	}
}

func TestUnitBuffPower4FF570AllFaultPrefixes(t *testing.T) {
	baseline := &unitBuffPowerWorld4FF570{
		buff:  31,
		unit:  unitBuffPowerObject4FF570,
		power: math.MaxUint8,
	}
	UnitBuffPower4FF570(baseline.hooks())
	want := append([]string(nil), baseline.events...)

	for faultAt := 1; faultAt <= len(want); faultAt++ {
		t.Run(fmt.Sprintf("fault-%d", faultAt), func(t *testing.T) {
			w := &unitBuffPowerWorld4FF570{
				faultAt: faultAt,
				buff:    31,
				unit:    unitBuffPowerObject4FF570,
				power:   math.MaxUint8,
			}
			var recovered any
			func() {
				defer func() { recovered = recover() }()
				UnitBuffPower4FF570(w.hooks())
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
