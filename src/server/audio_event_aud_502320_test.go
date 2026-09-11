package server

import (
	"errors"
	"fmt"
	"math"
	"reflect"
	"testing"
)

func TestReadAudioEventAUDList502320SignedCountAndCompletion(t *testing.T) {
	tests := []struct {
		name        string
		count       int32
		records     []bool
		want        bool
		wantRecords int
	}{
		{name: "minimum-signed-count", count: math.MinInt32, want: true},
		{name: "negative-count", count: -1, want: true},
		{name: "zero-count", count: 0, want: true},
		{name: "positive-completion", count: 3, records: []bool{true, true, true}, want: true, wantRecords: 3},
		{name: "record-failure", count: 3, records: []bool{true, false, true}, want: false, wantRecords: 2},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			countReads := 0
			recordReads := 0
			got := readAudioEventAUDList502320(audioEventAUDListHooks502320{
				readCount: func() int32 {
					countReads++
					return tc.count
				},
				readRecord: func() bool {
					result := tc.records[recordReads]
					recordReads++
					return result
				},
			})
			if got != tc.want || countReads != 1 || recordReads != tc.wantRecords {
				t.Fatalf("result/count reads/record reads = %t/%d/%d, want %t/1/%d", got, countReads, recordReads, tc.want, tc.wantRecords)
			}
		})
	}
}

const audioEventAUDDescriptorBase502370 = uint64(0x1_0000_0000)

type audioEventAUDFields502370 struct {
	maxDistance int32
	flags       uint32
	field8      uint32
	field12     uint32
	field16     uint32
	field20     uint32
}

type audioEventAUDWorld502370 struct {
	name        string
	nameErr     error
	sound       uint32
	initialized bool
	i16         []int16
	u8          []uint8
	i8          []int8
	fields      map[uint64]*audioEventAUDFields502370
	events      []string
	onResolve   func()
}

func newAudioEventAUDWorld502370() *audioEventAUDWorld502370 {
	return &audioEventAUDWorld502370{
		name:        "KnownSound",
		sound:       17,
		initialized: true,
		fields:      make(map[uint64]*audioEventAUDFields502370),
	}
}

func audioEventAUDDescriptor502370(sound uint32) uint64 {
	return audioEventAUDDescriptorBase502370 + uint64(sound)
}

func (w *audioEventAUDWorld502370) observe(event string) {
	w.events = append(w.events, event)
}

func (w *audioEventAUDWorld502370) nextI16() int16 {
	if len(w.i16) == 0 {
		panic("unexpected AUD int16 read")
	}
	value := w.i16[0]
	w.i16 = w.i16[1:]
	w.observe(fmt.Sprintf("i16:%d", value))
	return value
}

func (w *audioEventAUDWorld502370) nextU8() uint8 {
	if len(w.u8) == 0 {
		panic("unexpected AUD uint8 read")
	}
	value := w.u8[0]
	w.u8 = w.u8[1:]
	w.observe(fmt.Sprintf("u8:%02x", value))
	return value
}

func (w *audioEventAUDWorld502370) nextI8() int8 {
	if len(w.i8) == 0 {
		panic("unexpected AUD int8 read")
	}
	value := w.i8[0]
	w.i8 = w.i8[1:]
	w.observe(fmt.Sprintf("i8:%d", value))
	return value
}

func (w *audioEventAUDWorld502370) field(descriptor uint64) *audioEventAUDFields502370 {
	field := w.fields[descriptor]
	if field == nil {
		field = &audioEventAUDFields502370{}
		w.fields[descriptor] = field
	}
	return field
}

func (w *audioEventAUDWorld502370) hooks() audioEventAUDRecordHooks502370[uint32, uint64] {
	return audioEventAUDRecordHooks502370[uint32, uint64]{
		readName: func() (string, error) {
			w.observe("name:" + w.name)
			return w.name, w.nameErr
		},
		resolveSound: func(name string) uint32 {
			value := w.sound
			w.observe(fmt.Sprintf("resolve:%s=%d", name, value))
			if w.onResolve != nil {
				w.onResolve()
			}
			return value
		},
		zeroSound: 0,
		loadInitialized: func() bool {
			value := w.initialized
			w.observe(fmt.Sprintf("initialized:%t", value))
			return value
		},
		readI16: w.nextI16,
		readU8:  w.nextU8,
		readI8:  w.nextI8,
		skip: func(size int) {
			w.observe(fmt.Sprintf("skip:%d", size))
		},
		descriptor: func(sound uint32) uint64 {
			descriptor := audioEventAUDDescriptor502370(sound)
			w.observe(fmt.Sprintf("descriptor:%d=%#x", sound, descriptor))
			return descriptor
		},
		storeMaxDistance: func(descriptor uint64, value int32) {
			w.observe(fmt.Sprintf("max:%#x=%d", descriptor, value))
			w.field(descriptor).maxDistance = value
		},
		storeFlags: func(descriptor uint64, value uint32) {
			w.observe(fmt.Sprintf("flags:%#x=%08x", descriptor, value))
			w.field(descriptor).flags = value
		},
		storeField8: func(descriptor uint64, value uint32) {
			w.observe(fmt.Sprintf("field8:%#x=%d", descriptor, value))
			w.field(descriptor).field8 = value
		},
		incrementField12: func(descriptor uint64) {
			w.observe(fmt.Sprintf("inc12:%#x", descriptor))
			w.field(descriptor).field12++
		},
		storeField16: func(descriptor uint64, value uint32) {
			w.observe(fmt.Sprintf("field16:%#x=%d", descriptor, value))
			w.field(descriptor).field16 = value
		},
		storeField20: func(descriptor uint64, value uint32) {
			w.observe(fmt.Sprintf("field20:%#x=%d", descriptor, value))
			w.field(descriptor).field20 = value
		},
	}
}

func TestReadAudioEventAUDRecord502370ActiveOrderSignedFieldsAndNativeDescriptor(t *testing.T) {
	w := newAudioEventAUDWorld502370()
	w.i16 = []int16{math.MinInt16, 7}
	w.u8 = []uint8{0xfe, 0xfd}
	w.i8 = []int8{-1, 2, 0}
	descriptor := audioEventAUDDescriptor502370(w.sound)
	w.fields[descriptor] = &audioEventAUDFields502370{
		maxDistance: 9,
		field12:     math.MaxUint32,
		field16:     11,
	}

	if !readAudioEventAUDRecord502370(w.hooks()) {
		t.Fatal("AUD record result = false")
	}
	wantEvents := []string{
		"name:KnownSound",
		"resolve:KnownSound=17",
		"initialized:true",
		"i16:-32768",
		"u8:fe",
		"i16:7",
		"u8:fd",
		fmt.Sprintf("descriptor:17=%#x", descriptor),
		fmt.Sprintf("max:%#x=105", descriptor),
		fmt.Sprintf("flags:%#x=ffff8000", descriptor),
		fmt.Sprintf("field8:%#x=254", descriptor),
		fmt.Sprintf("field20:%#x=253", descriptor),
		"skip:3",
		"i8:-1",
		"skip:-1",
		fmt.Sprintf("inc12:%#x", descriptor),
		"i8:2",
		"skip:2",
		fmt.Sprintf("inc12:%#x", descriptor),
		"i8:0",
		fmt.Sprintf("field16:%#x=2", descriptor),
	}
	if !reflect.DeepEqual(w.events, wantEvents) {
		t.Fatalf("events =\n%v\nwant =\n%v", w.events, wantEvents)
	}
	if descriptor <= math.MaxUint32 {
		t.Fatalf("descriptor = %#x, want native-width token above 4 GiB", descriptor)
	}
	wantFields := &audioEventAUDFields502370{
		maxDistance: 105,
		flags:       0xffff8000,
		field8:      254,
		field12:     1,
		field16:     2,
		field20:     253,
	}
	if !reflect.DeepEqual(w.fields[descriptor], wantFields) {
		t.Fatalf("fields = %+v, want %+v", w.fields[descriptor], wantFields)
	}
}

func TestReadAudioEventAUDRecord502370NonpositiveDistancePreservesMaximum(t *testing.T) {
	for _, distance := range []int16{0, -1, math.MinInt16} {
		t.Run(fmt.Sprintf("distance-%d", distance), func(t *testing.T) {
			w := newAudioEventAUDWorld502370()
			w.i16 = []int16{-1, distance}
			w.u8 = []uint8{1, 2}
			w.i8 = []int8{0}
			descriptor := audioEventAUDDescriptor502370(w.sound)
			w.fields[descriptor] = &audioEventAUDFields502370{maxDistance: 777}

			if !readAudioEventAUDRecord502370(w.hooks()) {
				t.Fatal("AUD record result = false")
			}
			if got := w.fields[descriptor]; got.maxDistance != 777 || got.flags != math.MaxUint32 {
				t.Fatalf("fields = %+v, want preserved maximum and sign-extended flags", got)
			}
			for _, event := range w.events {
				if len(event) >= 4 && event[:4] == "max:" {
					t.Fatalf("unexpected maximum-distance store for %d: %v", distance, w.events)
				}
			}
		})
	}
}

func TestReadAudioEventAUDRecord502370SamplesInitializedAfterResolve(t *testing.T) {
	t.Run("resolve-enables-record", func(t *testing.T) {
		w := newAudioEventAUDWorld502370()
		w.initialized = false
		w.onResolve = func() { w.initialized = true }
		w.i16 = []int16{3, 0}
		w.u8 = []uint8{4, 5}
		w.i8 = []int8{0}
		if !readAudioEventAUDRecord502370(w.hooks()) {
			t.Fatal("AUD record result = false")
		}
		descriptor := audioEventAUDDescriptor502370(w.sound)
		if w.fields[descriptor] == nil || w.fields[descriptor].flags != 3 {
			t.Fatalf("descriptor = %+v, want active record", w.fields[descriptor])
		}
		wantPrefix := []string{"name:KnownSound", "resolve:KnownSound=17", "initialized:true"}
		if !reflect.DeepEqual(w.events[:len(wantPrefix)], wantPrefix) {
			t.Fatalf("events prefix = %v, want %v", w.events[:len(wantPrefix)], wantPrefix)
		}
	})

	t.Run("resolve-disables-record", func(t *testing.T) {
		w := newAudioEventAUDWorld502370()
		w.onResolve = func() { w.initialized = false }
		w.i8 = []int8{1, 0}
		if !readAudioEventAUDRecord502370(w.hooks()) {
			t.Fatal("AUD record result = false")
		}
		want := []string{
			"name:KnownSound", "resolve:KnownSound=17", "initialized:false",
			"skip:9", "i8:1", "skip:1", "i8:0",
		}
		if !reflect.DeepEqual(w.events, want) || len(w.fields) != 0 {
			t.Fatalf("events/fields = %v/%v, want inactive skip with no descriptor", w.events, w.fields)
		}
	})
}

func TestReadAudioEventAUDRecord502370ZeroSoundSkipsWithoutInitializationOrDescriptor(t *testing.T) {
	w := newAudioEventAUDWorld502370()
	w.sound = 0
	w.i8 = []int8{math.MinInt8, math.MaxInt8, 0}
	hooks := w.hooks()
	hooks.loadInitialized = func() bool {
		panic("zero sound must short-circuit the initialized load")
	}
	hooks.descriptor = func(uint32) uint64 {
		panic("inactive record must not select a descriptor")
	}

	if !readAudioEventAUDRecord502370(hooks) {
		t.Fatal("AUD record result = false")
	}
	want := []string{
		"name:KnownSound", "resolve:KnownSound=0", "skip:9",
		"i8:-128", "skip:-128", "i8:127", "skip:127", "i8:0",
	}
	if !reflect.DeepEqual(w.events, want) || len(w.fields) != 0 {
		t.Fatalf("events/fields = %v/%v, want exact inactive skip", w.events, w.fields)
	}
}

func TestReadAudioEventAUDRecord502370NameErrorStopsBeforeLookup(t *testing.T) {
	w := newAudioEventAUDWorld502370()
	w.nameErr = errors.New("truncated sound name")
	hooks := w.hooks()
	hooks.resolveSound = func(string) uint32 {
		panic("name failure must stop before sound lookup")
	}

	if readAudioEventAUDRecord502370(hooks) {
		t.Fatal("AUD record result = true, want name failure")
	}
	if want := []string{"name:KnownSound"}; !reflect.DeepEqual(w.events, want) {
		t.Fatalf("events = %v, want %v", w.events, want)
	}
}
