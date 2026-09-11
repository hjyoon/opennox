package server

import (
	"encoding/binary"
	"errors"
	"fmt"
	"math"
	"reflect"
	"testing"
)

const audioEventAVNTDescriptorBase502120 = uint64(0x100000000)

type audioEventAVNTFields502120 struct {
	maxDistance int32
	flags       uint32
	field8      uint32
	field12     uint32
	field16     uint32
	field20     uint32
}

type audioEventAVNTWorld502120 struct {
	initialized bool
	name        string
	nameErr     error
	sound       uint32
	data        []byte
	offset      int
	fields      map[uint64]*audioEventAVNTFields502120
	events      []string
	onReadName  func()
	onResolve   func()
}

func newAudioEventAVNTWorld502120() *audioEventAVNTWorld502120 {
	return &audioEventAVNTWorld502120{
		initialized: true,
		name:        "KnownSound",
		sound:       17,
		fields:      make(map[uint64]*audioEventAVNTFields502120),
	}
}

func audioEventAVNTDescriptor502120(sound uint32) uint64 {
	return audioEventAVNTDescriptorBase502120 + uint64(sound)
}

func (w *audioEventAVNTWorld502120) observe(event string) {
	w.events = append(w.events, event)
}

func (w *audioEventAVNTWorld502120) nextU8() uint8 {
	if w.offset >= len(w.data) {
		panic("unexpected AVNT byte read")
	}
	value := w.data[w.offset]
	w.offset++
	w.observe(fmt.Sprintf("u8:%02x", value))
	return value
}

func (w *audioEventAVNTWorld502120) nextI16() int16 {
	if w.offset+2 > len(w.data) {
		panic("unexpected AVNT word read")
	}
	value := int16(binary.LittleEndian.Uint16(w.data[w.offset : w.offset+2]))
	w.offset += 2
	w.observe(fmt.Sprintf("i16:%d", value))
	return value
}

func (w *audioEventAVNTWorld502120) skip(size int) {
	if w.offset+size > len(w.data) {
		panic("unexpected AVNT skip")
	}
	w.observe(fmt.Sprintf("skip:%d", size))
	w.offset += size
}

func (w *audioEventAVNTWorld502120) field(descriptor uint64) *audioEventAVNTFields502120 {
	field := w.fields[descriptor]
	if field == nil {
		field = &audioEventAVNTFields502120{}
		w.fields[descriptor] = field
	}
	return field
}

func (w *audioEventAVNTWorld502120) hooks() audioEventAVNTHooks502120[uint32, uint64] {
	return audioEventAVNTHooks502120[uint32, uint64]{
		loadInitialized: func() bool {
			value := w.initialized
			w.observe(fmt.Sprintf("initialized:%t", value))
			return value
		},
		readName: func() (string, error) {
			w.observe("name:" + w.name)
			if w.onReadName != nil {
				w.onReadName()
			}
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
		readU8:    w.nextU8,
		readI16:   w.nextI16,
		skip:      w.skip,
		descriptor: func(sound uint32) uint64 {
			descriptor := audioEventAVNTDescriptor502120(sound)
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

func checkAudioEventAVNTFields502120(t *testing.T, got, want *audioEventAVNTFields502120) {
	t.Helper()
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("fields = %+v, want %+v", got, want)
	}
}

func TestReadAudioEventAVNT502120AllOpcodesOrderSignedWordsAndNativeDescriptor(t *testing.T) {
	w := newAudioEventAVNTWorld502120()
	w.data = []byte{
		1, 0xa1,
		5, 0xa5,
		6, 0xb1, 0xb2,
		8, 0xc0, 0xc1, 0xc2, 0xc3, 0xc4, 0xc5, 0xc6, 0xc7,
		2, 0xfe,
		3, 0xfd,
		4, 0xfc,
		7, 3, 'a', 'b', 'c', 1, 'z', 0,
		9, 0xff, 0xff,
		10, 0x00, 0x80,
		0,
	}

	if !readAudioEventAVNT502120(w.hooks()) {
		t.Fatal("AVNT result = false, want terminator success")
	}
	descriptor := audioEventAVNTDescriptor502120(w.sound)
	wantEvents := []string{
		"initialized:true",
		"name:KnownSound",
		"resolve:KnownSound=17",
		"u8:01", "skip:1",
		"u8:05", "skip:1",
		"u8:06", "skip:2",
		"u8:08", "skip:8",
		"u8:02", "u8:fe",
		fmt.Sprintf("descriptor:17=%#x", descriptor),
		fmt.Sprintf("field16:%#x=254", descriptor),
		"u8:03", "u8:fd",
		fmt.Sprintf("descriptor:17=%#x", descriptor),
		fmt.Sprintf("field8:%#x=253", descriptor),
		"u8:04", "u8:fc",
		fmt.Sprintf("descriptor:17=%#x", descriptor),
		fmt.Sprintf("field20:%#x=252", descriptor),
		"u8:07", "u8:03", "skip:3",
		fmt.Sprintf("descriptor:17=%#x", descriptor),
		fmt.Sprintf("inc12:%#x", descriptor),
		"u8:01", "skip:1",
		fmt.Sprintf("descriptor:17=%#x", descriptor),
		fmt.Sprintf("inc12:%#x", descriptor),
		"u8:00",
		"u8:09", "i16:-1",
		fmt.Sprintf("descriptor:17=%#x", descriptor),
		fmt.Sprintf("max:%#x=-15", descriptor),
		"u8:0a", "i16:-32768",
		fmt.Sprintf("descriptor:17=%#x", descriptor),
		fmt.Sprintf("flags:%#x=ffff8000", descriptor),
		"u8:00",
	}
	if !reflect.DeepEqual(w.events, wantEvents) {
		t.Fatalf("events =\n%v\nwant =\n%v", w.events, wantEvents)
	}
	if w.offset != len(w.data) {
		t.Fatalf("offset = %d, want %d", w.offset, len(w.data))
	}
	if descriptor <= math.MaxUint32 {
		t.Fatalf("descriptor = %#x, want a native-width token above 4 GiB", descriptor)
	}
	checkAudioEventAVNTFields502120(t, w.fields[descriptor], &audioEventAVNTFields502120{
		maxDistance: -15,
		flags:       0xffff8000,
		field8:      253,
		field12:     2,
		field16:     254,
		field20:     252,
	})
}

func TestReadAudioEventAVNT502120CachesEntryInitializedGate(t *testing.T) {
	stream := []byte{
		2, 1,
		3, 2,
		4, 3,
		7, 1, 'x', 0,
		9, 0xfe, 0xff,
		10, 0xfd, 0xff,
		0,
	}

	t.Run("disabled-at-entry", func(t *testing.T) {
		w := newAudioEventAVNTWorld502120()
		w.initialized = false
		w.data = append([]byte(nil), stream...)
		w.onResolve = func() { w.initialized = true }
		descriptor := audioEventAVNTDescriptor502120(w.sound)
		initial := &audioEventAVNTFields502120{
			maxDistance: 77,
			flags:       88,
			field8:      99,
			field12:     100,
			field16:     101,
			field20:     102,
		}
		w.fields[descriptor] = initial

		if !readAudioEventAVNT502120(w.hooks()) {
			t.Fatal("AVNT result = false")
		}
		checkAudioEventAVNTFields502120(t, w.fields[descriptor], &audioEventAVNTFields502120{
			maxDistance: 77,
			flags:       88,
			field8:      99,
			field12:     101,
			field16:     101,
			field20:     102,
		})
	})

	t.Run("enabled-at-entry", func(t *testing.T) {
		w := newAudioEventAVNTWorld502120()
		w.data = append([]byte(nil), stream...)
		w.onReadName = func() { w.initialized = false }
		descriptor := audioEventAVNTDescriptor502120(w.sound)

		if !readAudioEventAVNT502120(w.hooks()) {
			t.Fatal("AVNT result = false")
		}
		checkAudioEventAVNTFields502120(t, w.fields[descriptor], &audioEventAVNTFields502120{
			maxDistance: -30,
			flags:       0xfffffffd,
			field8:      2,
			field12:     1,
			field16:     1,
			field20:     3,
		})
	})
}

func TestReadAudioEventAVNT502120ZeroSoundStillCountsOpcode7(t *testing.T) {
	w := newAudioEventAVNTWorld502120()
	w.sound = 0
	w.data = []byte{
		2, 7,
		3, 8,
		4, 9,
		7, 2, 'a', 'b', 0,
		9, 1, 0,
		10, 2, 0,
		0,
	}
	descriptor := audioEventAVNTDescriptor502120(0)
	w.fields[descriptor] = &audioEventAVNTFields502120{
		maxDistance: 10,
		flags:       11,
		field8:      12,
		field12:     math.MaxUint32,
		field16:     14,
		field20:     15,
	}

	if !readAudioEventAVNT502120(w.hooks()) {
		t.Fatal("AVNT result = false")
	}
	checkAudioEventAVNTFields502120(t, w.fields[descriptor], &audioEventAVNTFields502120{
		maxDistance: 10,
		flags:       11,
		field8:      12,
		field12:     0,
		field16:     14,
		field20:     15,
	})
}

func TestReadAudioEventAVNT502120Opcode7UsesUnsignedLength(t *testing.T) {
	w := newAudioEventAVNTWorld502120()
	w.data = append([]byte{7, 0xff}, make([]byte, 255)...)
	w.data = append(w.data, 0, 0)

	if !readAudioEventAVNT502120(w.hooks()) {
		t.Fatal("AVNT result = false")
	}
	descriptor := audioEventAVNTDescriptor502120(w.sound)
	if w.offset != len(w.data) || w.fields[descriptor].field12 != 1 {
		t.Fatalf("offset/count = %d/%d, want %d/1", w.offset, w.fields[descriptor].field12, len(w.data))
	}
}

func TestReadAudioEventAVNT502120NameErrorAndInvalidOpcode(t *testing.T) {
	t.Run("name-error", func(t *testing.T) {
		w := newAudioEventAVNTWorld502120()
		w.nameErr = errors.New("truncated name")
		if readAudioEventAVNT502120(w.hooks()) {
			t.Fatal("AVNT result = true, want name failure")
		}
		want := []string{"initialized:true", "name:KnownSound"}
		if !reflect.DeepEqual(w.events, want) {
			t.Fatalf("events = %v, want %v", w.events, want)
		}
	})

	t.Run("invalid-opcode", func(t *testing.T) {
		w := newAudioEventAVNTWorld502120()
		w.data = []byte{0xff, 0}
		if readAudioEventAVNT502120(w.hooks()) {
			t.Fatal("AVNT result = true, want invalid-opcode failure")
		}
		if w.offset != 1 {
			t.Fatalf("offset = %d, want one consumed opcode byte", w.offset)
		}
	})
}
