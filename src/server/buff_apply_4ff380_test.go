package server

import (
	"fmt"
	"math"
	"reflect"
	"testing"
)

const buffApplyObject4FF380 = uint64(0x1234567889abcdef)

type buffApplyWorld4FF380 struct {
	events       []string
	after        map[string]func()
	faultAt      int
	hecubah      uint32
	necromancer  uint32
	lookup       map[string]uint32
	buff         int32
	unit         uint64
	duration     int16
	power        int8
	typeIndex    uint16
	classLow     uint8
	subclass     uint32
	objectFlags  uint32
	gameFlags    map[uint32]int32
	active       int32
	timer        int32
	buffs        uint32
	spell        int32
	onAudio      int32
	durations    map[int32]uint16
	powers       map[int32]uint8
	setBuffFlags []uint32
}

func newBuffApplyWorld4FF380() *buffApplyWorld4FF380 {
	return &buffApplyWorld4FF380{
		after:       make(map[string]func()),
		hecubah:     0x1234,
		necromancer: 0x5678,
		lookup: map[string]uint32{
			"Hecubah":     0x1234,
			"Necromancer": 0x5678,
		},
		buff:      7,
		unit:      buffApplyObject4FF380,
		duration:  -2,
		power:     -3,
		typeIndex: 0x9abc,
		gameFlags: make(map[uint32]int32),
		buffs:     0x40000002,
		spell:     math.MinInt32,
		onAudio:   math.MaxInt32,
		durations: make(map[int32]uint16),
		powers:    make(map[int32]uint8),
	}
}

func (w *buffApplyWorld4FF380) observe(event string) {
	w.events = append(w.events, event)
	if w.faultAt != 0 && len(w.events) == w.faultAt {
		panic(event)
	}
	if after := w.after[event]; after != nil {
		after()
	}
}

func (w *buffApplyWorld4FF380) hooks() BuffApplyHooks4FF380[uint64] {
	return BuffApplyHooks4FF380[uint64]{
		LoadHecubahTypeID: func() uint32 {
			value := w.hecubah
			w.observe(fmt.Sprintf("hecubah=%08x", value))
			return value
		},
		StoreHecubahTypeID: func(value uint32) {
			w.observe(fmt.Sprintf("hecubah<=%08x", value))
			w.hecubah = value
		},
		LoadNecromancerTypeID: func() uint32 {
			value := w.necromancer
			w.observe(fmt.Sprintf("necromancer=%08x", value))
			return value
		},
		StoreNecromancerTypeID: func(value uint32) {
			w.observe(fmt.Sprintf("necromancer<=%08x", value))
			w.necromancer = value
		},
		LookupTypeID: func(name string) uint32 {
			value := w.lookup[name]
			w.observe(fmt.Sprintf("lookup:%s=%08x", name, value))
			return value
		},
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
		LoadDurationArg: func() int16 {
			value := w.duration
			w.observe(fmt.Sprintf("duration=%04x", uint16(value)))
			return value
		},
		LoadPowerArg: func() int8 {
			value := w.power
			w.observe(fmt.Sprintf("power=%02x", uint8(value)))
			return value
		},
		LoadTypeIndex: func(unit uint64) uint16 {
			value := w.typeIndex
			w.observe(fmt.Sprintf("type:%016x=%04x", unit, value))
			return value
		},
		LoadClassLow: func(unit uint64) uint8 {
			value := w.classLow
			w.observe(fmt.Sprintf("class-low:%016x=%02x", unit, value))
			return value
		},
		LoadSubclass: func(unit uint64) uint32 {
			value := w.subclass
			w.observe(fmt.Sprintf("subclass:%016x=%08x", unit, value))
			return value
		},
		LoadObjectFlags: func(unit uint64) uint32 {
			value := w.objectFlags
			w.observe(fmt.Sprintf("flags:%016x=%08x", unit, value))
			return value
		},
		GameFlag: func(flag uint32) int32 {
			value := w.gameFlags[flag]
			w.observe(fmt.Sprintf("game:%04x=%08x", flag, uint32(value)))
			return value
		},
		Audio: func(id int32, unit uint64, first, second int32) {
			w.observe(fmt.Sprintf("audio:%08x:%016x:%d:%d", uint32(id), unit, first, second))
		},
		TestBuff: func(unit uint64, buff int32) int32 {
			value := w.active
			w.observe(fmt.Sprintf("test:%016x:%08x=%08x", unit, uint32(buff), uint32(value)))
			return value
		},
		LoadBuffTimer: func(unit uint64, buff int32) int32 {
			value := w.timer
			w.observe(fmt.Sprintf("timer:%016x:%08x=%08x", unit, uint32(buff), uint32(value)))
			return value
		},
		BuffOff: func(unit uint64, buff int32) int32 {
			w.observe(fmt.Sprintf("off:%016x:%08x", unit, uint32(buff)))
			return math.MinInt32
		},
		StoreDuration: func(unit uint64, buff int32, value uint16) {
			w.observe(fmt.Sprintf("duration:%016x:%08x<=%04x", unit, uint32(buff), value))
			w.durations[buff] = value
		},
		StorePower: func(unit uint64, buff int32, value uint8) {
			w.observe(fmt.Sprintf("power:%016x:%08x<=%02x", unit, uint32(buff), value))
			w.powers[buff] = value
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
		EnchantSpell: func(buff int32) int32 {
			value := w.spell
			w.observe(fmt.Sprintf("spell:%08x=%08x", uint32(buff), uint32(value)))
			return value
		},
		SpellAudio: func(spell, selector int32) int32 {
			value := w.onAudio
			w.observe(fmt.Sprintf("spell-audio:%08x:%d=%08x", uint32(spell), selector, uint32(value)))
			return value
		},
	}
}

func TestBuffApply4FF380ExactSuccessTraceAndWidths(t *testing.T) {
	w := newBuffApplyWorld4FF380()
	BuffApply4FF380(w.hooks())
	want := []string{
		"hecubah=00001234", "buff=00000007", "unit=1234567889abcdef",
		"hecubah=00001234", "type:1234567889abcdef=9abc",
		"game:1000=00000000", "game:1000=00000000",
		"class-low:1234567889abcdef=00", "flags:1234567889abcdef=00000000",
		"test:1234567889abcdef:00000007=00000000", "off:1234567889abcdef:00000000",
		"duration=fffe", "power=fd",
		"duration:1234567889abcdef:00000007<=fffe", "power:1234567889abcdef:00000007<=fd",
		"buffs:1234567889abcdef=40000002", "set:1234567889abcdef:40000082",
		"spell:00000007=80000000", "spell-audio:80000000:1=7fffffff",
		"audio:7fffffff:1234567889abcdef:0:0",
	}
	if !reflect.DeepEqual(w.events, want) {
		t.Fatalf("events = %q, want exact trace %q", w.events, want)
	}
	if w.durations[7] != math.MaxUint16-1 || w.powers[7] != math.MaxUint8-2 {
		t.Fatalf("stored duration/power = %#x/%#x, want exact low word/byte", w.durations[7], w.powers[7])
	}
}

func TestBuffApply4FF380InitializesTypeCacheBeforeNilGate(t *testing.T) {
	w := newBuffApplyWorld4FF380()
	w.hecubah, w.necromancer, w.unit = 0, 0xffffffff, 0
	BuffApply4FF380(w.hooks())
	want := []string{
		"hecubah=00000000", "buff=00000007",
		"lookup:Hecubah=00001234", "hecubah<=00001234",
		"lookup:Necromancer=00005678", "necromancer<=00005678",
		"unit=0000000000000000",
	}
	if !reflect.DeepEqual(w.events, want) {
		t.Fatalf("events = %q, want cache-before-null trace %q", w.events, want)
	}
}

func TestBuffApply4FF380NonzeroHecubahCacheSkipsBothLookups(t *testing.T) {
	w := newBuffApplyWorld4FF380()
	w.necromancer, w.unit = 0, 0
	BuffApply4FF380(w.hooks())
	want := []string{"hecubah=00001234", "buff=00000007", "unit=0000000000000000"}
	if !reflect.DeepEqual(w.events, want) {
		t.Fatalf("events = %q, want initialized-cache trace %q", w.events, want)
	}
}

func TestBuffApply4FF380NamedTypeEarlyReturns(t *testing.T) {
	tests := []struct {
		name      string
		typeIndex uint16
		buff      int32
		quest     int32
		wantTail  []string
	}{
		{"hecubah-anti-magic", 0x1234, buffApplyAntiMagic4FF380, 0, nil},
		{"quest-hecubah-confused", 0x1234, buffApplyConfused4FF380, 1, []string{"audio:00000246:1234567889abcdef:0:0"}},
		{"quest-necromancer-confused", 0x5678, buffApplyConfused4FF380, -1, []string{"audio:00000253:1234567889abcdef:0:0"}},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			w := newBuffApplyWorld4FF380()
			w.typeIndex, w.buff = tc.typeIndex, tc.buff
			w.gameFlags[buffApplyQuestFlag4FF380] = tc.quest
			BuffApply4FF380(w.hooks())
			for _, event := range w.events {
				if event[:min(len(event), 5)] == "flags" || event[:min(len(event), 4)] == "test" {
					t.Fatalf("early-return trace reached normal application: %q", w.events)
				}
			}
			if len(tc.wantTail) == 0 {
				if got := w.events[len(w.events)-1]; got != "type:1234567889abcdef=1234" {
					t.Fatalf("anti-magic tail = %q, want type comparison; trace = %q", got, w.events)
				}
			} else if got := w.events[len(w.events)-1:]; !reflect.DeepEqual(got, tc.wantTail) {
				t.Fatalf("tail = %q, want %q; trace = %q", got, tc.wantTail, w.events)
			}
		})
	}
}

func TestBuffApply4FF380QuestFlagIsQueriedTwice(t *testing.T) {
	w := newBuffApplyWorld4FF380()
	w.buff = buffApplyConfused4FF380
	calls := 0
	h := w.hooks()
	h.GameFlag = func(flag uint32) int32 {
		calls++
		value := int32(0)
		if calls == 2 {
			value = 1
		}
		w.observe(fmt.Sprintf("game:%04x=%08x", flag, uint32(value)))
		return value
	}
	w.typeIndex = uint16(w.necromancer)
	BuffApply4FF380(h)
	if calls != 2 || w.events[len(w.events)-1] != "audio:00000253:1234567889abcdef:0:0" {
		t.Fatalf("calls/trace = %d/%q, want second-query Necromancer return", calls, w.events)
	}
}

func TestBuffApply4FF380NonCoopFearSubclassAlwaysRejects(t *testing.T) {
	for _, tc := range []struct {
		name      string
		typeIndex uint16
		wantAudio string
	}{
		{"other", 0x9abc, ""},
		{"hecubah", 0x1234, "audio:00000246:1234567889abcdef:0:0"},
		{"necromancer", 0x5678, "audio:00000253:1234567889abcdef:0:0"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			w := newBuffApplyWorld4FF380()
			w.buff = buffApplyAfraid4FF380
			w.classLow = buffApplyMonsterClass4FF380
			w.subclass = buffApplyFearSubclass4FF380
			w.typeIndex = tc.typeIndex
			BuffApply4FF380(w.hooks())
			if len(w.setBuffFlags) != 0 {
				t.Fatalf("fear-special path applied buff: trace = %q", w.events)
			}
			gotAudio := ""
			for _, event := range w.events {
				if len(event) >= 6 && event[:6] == "audio:" {
					gotAudio = event
				}
				if len(event) >= 6 && event[:6] == "flags:" {
					t.Fatalf("fear-special path loaded normal flags: %q", w.events)
				}
			}
			if gotAudio != tc.wantAudio {
				t.Fatalf("audio = %q, want %q; trace = %q", gotAudio, tc.wantAudio, w.events)
			}
		})
	}
}

func TestBuffApply4FF380CoopBypassesFearSpecialReturn(t *testing.T) {
	w := newBuffApplyWorld4FF380()
	w.buff = buffApplyAfraid4FF380
	w.classLow = buffApplyMonsterClass4FF380
	w.subclass = buffApplyFearSubclass4FF380
	w.gameFlags[buffApplyCoopFlag4FF380] = 1
	BuffApply4FF380(w.hooks())
	if len(w.setBuffFlags) != 1 || w.setBuffFlags[0] != w.buffs {
		t.Fatalf("set flags = %#v, want successful Afraid application; trace = %q", w.setBuffFlags, w.events)
	}
}

func TestBuffApply4FF380NormalRejectionGates(t *testing.T) {
	t.Run("object-flags", func(t *testing.T) {
		w := newBuffApplyWorld4FF380()
		w.objectFlags = 0x8000
		BuffApply4FF380(w.hooks())
		if len(w.setBuffFlags) != 0 || w.events[len(w.events)-1] != "flags:1234567889abcdef=00008000" {
			t.Fatalf("trace = %q, want flags rejection", w.events)
		}
	})
	t.Run("active-zero-timer", func(t *testing.T) {
		w := newBuffApplyWorld4FF380()
		w.active, w.timer = -1, 0
		BuffApply4FF380(w.hooks())
		if len(w.setBuffFlags) != 0 || w.events[len(w.events)-1] != "timer:1234567889abcdef:00000007=00000000" {
			t.Fatalf("trace = %q, want zero-timer rejection", w.events)
		}
	})
	t.Run("inactive-skips-timer", func(t *testing.T) {
		w := newBuffApplyWorld4FF380()
		w.active, w.timer = 0, math.MinInt32
		BuffApply4FF380(w.hooks())
		for _, event := range w.events {
			if len(event) >= 6 && event[:6] == "timer:" {
				t.Fatalf("inactive path loaded timer: %q", w.events)
			}
		}
	})
	t.Run("active-nonzero-timer-refreshes", func(t *testing.T) {
		w := newBuffApplyWorld4FF380()
		w.active, w.timer = 1, math.MinInt32
		BuffApply4FF380(w.hooks())
		if len(w.setBuffFlags) != 1 {
			t.Fatalf("trace = %q, want active nonzero-timer refresh", w.events)
		}
	})
}

func TestBuffApply4FF380X86ShiftAliasesAndKeepsFullBuffDword(t *testing.T) {
	for _, tc := range []struct {
		buff int32
		bit  uint32
	}{
		{0, 0}, {31, 31}, {32, 0}, {-1, 31}, {math.MinInt32, 0}, {math.MaxInt32, 31},
	} {
		t.Run(fmt.Sprintf("%08x", uint32(tc.buff)), func(t *testing.T) {
			w := newBuffApplyWorld4FF380()
			w.buff, w.buffs = tc.buff, 0
			BuffApply4FF380(w.hooks())
			if got, want := w.setBuffFlags[0], uint32(1)<<tc.bit; got != want {
				t.Fatalf("flags = %#08x, want %#08x", got, want)
			}
			if _, ok := w.durations[tc.buff]; !ok {
				t.Fatalf("duration store did not receive full buff dword %#08x", uint32(tc.buff))
			}
		})
	}
}

func TestBuffApply4FF380ReloadsLiveBuffsAfterStores(t *testing.T) {
	w := newBuffApplyWorld4FF380()
	w.buffs = 0
	w.after["power:1234567889abcdef:00000007<=fd"] = func() {
		w.buffs = 0x80000001
	}
	BuffApply4FF380(w.hooks())
	if got, want := w.setBuffFlags[0], uint32(0x80000081); got != want {
		t.Fatalf("flags = %#08x, want post-store live word %#08x; trace = %q", got, want, w.events)
	}
}

func TestBuffApply4FF380FaultPrefixes(t *testing.T) {
	baseline := newBuffApplyWorld4FF380()
	BuffApply4FF380(baseline.hooks())
	want := append([]string(nil), baseline.events...)
	for faultAt := 1; faultAt <= len(want); faultAt++ {
		t.Run(fmt.Sprintf("fault-%d", faultAt), func(t *testing.T) {
			w := newBuffApplyWorld4FF380()
			w.faultAt = faultAt
			var recovered any
			func() {
				defer func() { recovered = recover() }()
				BuffApply4FF380(w.hooks())
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
