package server

import (
	"fmt"
	"math"
	"reflect"
	"testing"
)

const spellBuffOffObject4FF5B0 = uint64(0x1234567889abcdef)

type spellBuffOffWorld4FF5B0 struct {
	events       []string
	after        map[string]func()
	faultAt      int
	buff         int32
	unit         uint64
	buffs        uint32
	spell        int32
	offAudio     int32
	durations    map[int32]uint16
	powers       map[int32]uint8
	setBuffFlags []uint32
}

func newSpellBuffOffWorld4FF5B0() *spellBuffOffWorld4FF5B0 {
	return &spellBuffOffWorld4FF5B0{
		after:     make(map[string]func()),
		buff:      7,
		unit:      spellBuffOffObject4FF5B0,
		buffs:     0xa0000084,
		spell:     math.MinInt32,
		offAudio:  math.MaxInt32,
		durations: map[int32]uint16{7: 0xabcd},
		powers:    map[int32]uint8{7: 0xef},
	}
}

func (w *spellBuffOffWorld4FF5B0) observe(event string) {
	w.events = append(w.events, event)
	if w.faultAt != 0 && len(w.events) == w.faultAt {
		panic(event)
	}
	if after := w.after[event]; after != nil {
		after()
	}
}

func (w *spellBuffOffWorld4FF5B0) hooks() SpellBuffOffHooks4FF5B0[uint64] {
	return SpellBuffOffHooks4FF5B0[uint64]{
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
		LoadBuffs: func(unit uint64) uint32 {
			value := w.buffs
			w.observe(fmt.Sprintf("buffs:%016x=%08x", unit, value))
			return value
		},
		SetBuffFlags: func(unit uint64, flags uint32) {
			w.observe(fmt.Sprintf("set:%016x:%08x", unit, flags))
			w.setBuffFlags = append(w.setBuffFlags, flags)
			w.buffs = flags
		},
		StoreDuration: func(unit uint64, buff int32, value uint16) {
			w.observe(fmt.Sprintf("duration:%016x:%08x<=%04x", unit, uint32(buff), value))
			w.durations[buff] = value
		},
		StorePower: func(unit uint64, buff int32, value uint8) {
			w.observe(fmt.Sprintf("power:%016x:%08x<=%02x", unit, uint32(buff), value))
			w.powers[buff] = value
		},
		EnchantSpell: func(buff int32) int32 {
			value := w.spell
			w.observe(fmt.Sprintf("spell:%08x=%08x", uint32(buff), uint32(value)))
			return value
		},
		SpellAudio: func(spellID, field int32) int32 {
			value := w.offAudio
			w.observe(fmt.Sprintf("spell-audio:%08x:%d=%08x", uint32(spellID), field, uint32(value)))
			return value
		},
		Audio: func(audioID int32, unit uint64, kind, code int32) {
			w.observe(fmt.Sprintf("audio:%08x:%016x:%d:%d", uint32(audioID), unit, kind, code))
		},
	}
}

func spellBuffOffSuccessTrace4FF5B0(buff int32, unit uint64, buffs uint32) []string {
	mask := uint32(1) << (uint32(buff) & 31)
	return []string{
		fmt.Sprintf("buff=%08x", uint32(buff)),
		fmt.Sprintf("unit=%016x", unit),
		fmt.Sprintf("buffs:%016x=%08x", unit, buffs),
		fmt.Sprintf("set:%016x:%08x", unit, buffs&^mask),
		fmt.Sprintf("duration:%016x:%08x<=0000", unit, uint32(buff)),
		fmt.Sprintf("power:%016x:%08x<=00", unit, uint32(buff)),
		fmt.Sprintf("spell:%08x=80000000", uint32(buff)),
		"spell-audio:80000000:2=7fffffff",
		fmt.Sprintf("audio:7fffffff:%016x:0:0", unit),
	}
}

func TestSpellBuffOff4FF5B0ExactSuccessTraceAndWidths(t *testing.T) {
	w := newSpellBuffOffWorld4FF5B0()
	if got := SpellBuffOff4FF5B0(w.hooks()); got != 0 {
		t.Fatalf("result = %#08x, want zero", uint32(got))
	}
	want := spellBuffOffSuccessTrace4FF5B0(w.buff, spellBuffOffObject4FF5B0, 0xa0000084)
	if !reflect.DeepEqual(w.events, want) {
		t.Fatalf("events = %q, want exact trace %q", w.events, want)
	}
	if w.buffs != 0xa0000004 || w.durations[7] != 0 || w.powers[7] != 0 {
		t.Fatalf("state = flags %#08x duration %#04x power %#02x, want bit seven and slot storage cleared", w.buffs, w.durations[7], w.powers[7])
	}
}

func TestSpellBuffOff4FF5B0InactiveUsesX86MaskedShift(t *testing.T) {
	for _, buff := range []int32{0, 1, 31, 32, 63, -1, math.MinInt32, math.MaxInt32} {
		t.Run(fmt.Sprintf("buff-%08x", uint32(buff)), func(t *testing.T) {
			w := newSpellBuffOffWorld4FF5B0()
			w.buff, w.buffs = buff, 0
			got := SpellBuffOff4FF5B0(w.hooks())
			want := int32(uint32(1) << (uint32(buff) & 31))
			if got != want {
				t.Fatalf("result = %#08x, want x86 masked bit %#08x", uint32(got), uint32(want))
			}
			wantEvents := []string{
				fmt.Sprintf("buff=%08x", uint32(buff)),
				"unit=1234567889abcdef",
				"buffs:1234567889abcdef=00000000",
			}
			if !reflect.DeepEqual(w.events, wantEvents) {
				t.Fatalf("events = %q, want inactive trace %q", w.events, wantEvents)
			}
			if len(w.setBuffFlags) != 0 {
				t.Fatalf("inactive enchant mutated flags: %#v", w.setBuffFlags)
			}
		})
	}
}

func TestSpellBuffOff4FF5B0ActiveMaskedBitKeepsFullStoreIndex(t *testing.T) {
	w := newSpellBuffOffWorld4FF5B0()
	w.buff = 48
	w.buffs = uint32(1) << 16
	w.durations[48], w.powers[48] = 0xffff, 0xff
	if got := SpellBuffOff4FF5B0(w.hooks()); got != 0 {
		t.Fatalf("result = %#08x, want zero", uint32(got))
	}
	want := spellBuffOffSuccessTrace4FF5B0(48, spellBuffOffObject4FF5B0, uint32(1)<<16)
	if !reflect.DeepEqual(w.events, want) {
		t.Fatalf("events = %q, want masked-bit/full-index trace %q", w.events, want)
	}
	if w.durations[48] != 0 || w.powers[48] != 0 {
		t.Fatal("full signed dword store index was not preserved")
	}
}

func TestSpellBuffOff4FF5B0NoAudioEnchantmentsStoreFirst(t *testing.T) {
	for _, buff := range []int32{spellBuffOffNoAudio16_4FF5B0, spellBuffOffNoAudio30_4FF5B0} {
		t.Run(fmt.Sprintf("buff-%d", buff), func(t *testing.T) {
			w := newSpellBuffOffWorld4FF5B0()
			w.buff = buff
			w.buffs = uint32(1) << uint32(buff)
			w.durations[buff], w.powers[buff] = 0xffff, 0xff
			if got := SpellBuffOff4FF5B0(w.hooks()); got != 0 {
				t.Fatalf("result = %#08x, want zero", uint32(got))
			}
			want := []string{
				fmt.Sprintf("buff=%08x", uint32(buff)),
				"unit=1234567889abcdef",
				fmt.Sprintf("buffs:1234567889abcdef=%08x", uint32(1)<<uint32(buff)),
				"set:1234567889abcdef:00000000",
				fmt.Sprintf("duration:1234567889abcdef:%08x<=0000", uint32(buff)),
				fmt.Sprintf("power:1234567889abcdef:%08x<=00", uint32(buff)),
			}
			if !reflect.DeepEqual(w.events, want) {
				t.Fatalf("events = %q, want store-before-return trace %q", w.events, want)
			}
			if w.durations[buff] != 0 || w.powers[buff] != 0 {
				t.Fatal("special enchant storage was not cleared")
			}
		})
	}
}

func TestSpellBuffOff4FF5B0CachesUnitAndBuffDword(t *testing.T) {
	w := newSpellBuffOffWorld4FF5B0()
	w.after["unit=1234567889abcdef"] = func() { w.unit = 0 }
	w.after["buffs:1234567889abcdef=a0000084"] = func() { w.buffs = math.MaxUint32 }
	if got := SpellBuffOff4FF5B0(w.hooks()); got != 0 {
		t.Fatalf("result = %#08x, want zero", uint32(got))
	}
	want := spellBuffOffSuccessTrace4FF5B0(7, spellBuffOffObject4FF5B0, 0xa0000084)
	if !reflect.DeepEqual(w.events, want) {
		t.Fatalf("events = %q, want cached unit/buffs trace %q", w.events, want)
	}
	if len(w.setBuffFlags) != 1 || w.setBuffFlags[0] != 0xa0000004 {
		t.Fatalf("set flags = %#v, want cached flags with bit seven cleared", w.setBuffFlags)
	}
}

func TestSpellBuffOff4FF5B0DoesNotShortCircuitNullToken(t *testing.T) {
	w := newSpellBuffOffWorld4FF5B0()
	w.unit, w.buffs = 0, 0
	got := SpellBuffOff4FF5B0(w.hooks())
	if got != 0x80 {
		t.Fatalf("result = %#08x, want bit mask 0x80", uint32(got))
	}
	want := []string{"buff=00000007", "unit=0000000000000000", "buffs:0000000000000000=00000000"}
	if !reflect.DeepEqual(w.events, want) {
		t.Fatalf("events = %q, want unguarded null-token trace %q", w.events, want)
	}
}

func TestSpellBuffOff4FF5B0AllFaultPrefixes(t *testing.T) {
	baseline := newSpellBuffOffWorld4FF5B0()
	if got := SpellBuffOff4FF5B0(baseline.hooks()); got != 0 {
		t.Fatalf("baseline result = %#08x, want zero", uint32(got))
	}
	want := append([]string(nil), baseline.events...)
	if len(want) != 9 {
		t.Fatalf("observable steps = %d, want 9", len(want))
	}

	for faultAt := 1; faultAt <= len(want); faultAt++ {
		t.Run(fmt.Sprintf("fault-%d", faultAt), func(t *testing.T) {
			w := newSpellBuffOffWorld4FF5B0()
			w.faultAt = faultAt
			var recovered any
			func() {
				defer func() { recovered = recover() }()
				_ = SpellBuffOff4FF5B0(w.hooks())
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
