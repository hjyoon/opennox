package server

import (
	"fmt"
	"math"
	"reflect"
	"runtime"
	"testing"
	"unsafe"
)

const (
	audioInsertDescriptor501EA0 = uint64(0x100000001)
	audioInsertEvent501EA0      = uint64(0x200000002)
	audioListDescriptor501F30   = uint64(0x300000003)
	audioListHead501F30         = uint64(0x400000004)
	audioListMiddle501F30       = uint64(0x500000005)
	audioListNew501F30          = uint64(0x600000006)
)

func TestSoundBitmap501EF0And501F10UseRawDwordBits(t *testing.T) {
	var bitmap soundBitmap501E80
	tests := []struct {
		id   int32
		word int
		mask uint32
	}{
		{0, 0, 0x00000001},
		{31, 0, 0x80000000},
		{32, 1, 0x00000001},
		{1023, 31, 0x80000000},
	}
	for _, test := range tests {
		if got := testSoundBitmap501EF0(&bitmap, test.id); got != 0 {
			t.Fatalf("test before set for %d = %#x, want 0", test.id, got)
		}
		setSoundBitmap501F10(&bitmap, test.id)
		if got := testSoundBitmap501EF0(&bitmap, test.id); got != test.mask {
			t.Fatalf("test after set for %d = %#x, want raw mask %#x", test.id, got, test.mask)
		}
		if bitmap[test.word]&test.mask == 0 {
			t.Fatalf("bitmap word %d does not contain %#x", test.word, test.mask)
		}
	}
	for word, value := range bitmap {
		want := uint32(0)
		switch word {
		case 0:
			want = 0x80000001
		case 1:
			want = 0x00000001
		case 31:
			want = 0x80000000
		}
		if value != want {
			t.Fatalf("bitmap[%d] = %#x, want %#x", word, value, want)
		}
	}
}

func TestAudioEventInsert501EA0FirstUseOrderAndCachedSound(t *testing.T) {
	var events []string
	soundID := int32(77)
	var storedHead uint64 = math.MaxUint64
	var storedPercentage int32
	audioEventInsert501EA0(audioInsertEvent501EA0, -123, audioEventInsertHooks501EA0[uint64, uint64]{
		loadEventSound: func(event uint64) int32 {
			events = append(events, fmt.Sprintf("sound:%#x", event))
			return soundID
		},
		descriptor: func(id int32) uint64 {
			events = append(events, fmt.Sprintf("descriptor:%d", id))
			return audioInsertDescriptor501EA0
		},
		testBitmap: func(id int32) uint32 {
			events = append(events, fmt.Sprintf("test:%d", id))
			soundID = 99
			return 0
		},
		setBitmap: func(id int32) {
			events = append(events, fmt.Sprintf("set:%d", id))
		},
		storeListHead: func(descriptor, head uint64) {
			events = append(events, fmt.Sprintf("head:%#x:%#x", descriptor, head))
			storedHead = head
		},
		storeEventPercent: func(event uint64, percentage int32) {
			events = append(events, fmt.Sprintf("percent:%#x:%d", event, percentage))
			storedPercentage = percentage
		},
		insert: func(descriptor, event uint64) {
			events = append(events, fmt.Sprintf("insert:%#x:%#x", descriptor, event))
		},
	})
	want := []string{
		"sound:0x200000002", "descriptor:77", "test:77", "set:77",
		"head:0x100000001:0x0", "percent:0x200000002:-123",
		"insert:0x100000001:0x200000002",
	}
	if !reflect.DeepEqual(events, want) {
		t.Fatalf("events = %v, want %v", events, want)
	}
	if storedHead != 0 || storedPercentage != -123 {
		t.Fatalf("head/percentage = %#x/%d, want 0/-123", storedHead, storedPercentage)
	}
}

func TestAudioEventInsert501EA0ExistingBitmapSkipsInitialization(t *testing.T) {
	var events []string
	audioEventInsert501EA0(audioInsertEvent501EA0, 41, audioEventInsertHooks501EA0[uint64, uint64]{
		loadEventSound: func(uint64) int32 {
			events = append(events, "sound")
			return 31
		},
		descriptor: func(int32) uint64 {
			events = append(events, "descriptor")
			return audioInsertDescriptor501EA0
		},
		testBitmap: func(int32) uint32 {
			events = append(events, "test")
			return 0x80000000
		},
		setBitmap: func(int32) {
			t.Fatal("setBitmap called for an existing bit")
		},
		storeListHead: func(uint64, uint64) {
			t.Fatal("list head cleared for an existing bit")
		},
		storeEventPercent: func(uint64, int32) {
			events = append(events, "percent")
		},
		insert: func(uint64, uint64) {
			events = append(events, "insert")
		},
	})
	want := []string{"sound", "descriptor", "test", "percent", "insert"}
	if !reflect.DeepEqual(events, want) {
		t.Fatalf("events = %v, want %v", events, want)
	}
}

type audioEventListWorld501F30 struct {
	events           []string
	head             map[uint64]uint64
	percent          map[uint64]int32
	sound            map[uint64]int32
	soundDescriptors map[int32]uint64
	flags            map[uint64]uint8
	limit            map[uint64]int32
	next             map[uint64]uint64
	afterPercentLoad func(uint64)
}

func newAudioEventListWorld501F30() *audioEventListWorld501F30 {
	return &audioEventListWorld501F30{
		head:             make(map[uint64]uint64),
		percent:          make(map[uint64]int32),
		sound:            make(map[uint64]int32),
		soundDescriptors: make(map[int32]uint64),
		flags:            make(map[uint64]uint8),
		limit:            make(map[uint64]int32),
		next:             make(map[uint64]uint64),
	}
}

func (w *audioEventListWorld501F30) hooks() audioEventListHooks501F30[uint64, uint64] {
	return audioEventListHooks501F30[uint64, uint64]{
		loadHead: func(descriptor uint64) uint64 {
			w.events = append(w.events, fmt.Sprintf("head:%#x", descriptor))
			return w.head[descriptor]
		},
		storeHead: func(descriptor, head uint64) {
			w.events = append(w.events, fmt.Sprintf("store-head:%#x:%#x", descriptor, head))
			w.head[descriptor] = head
		},
		loadPercent: func(event uint64) int32 {
			w.events = append(w.events, fmt.Sprintf("percent:%#x", event))
			value := w.percent[event]
			if w.afterPercentLoad != nil {
				w.afterPercentLoad(event)
			}
			return value
		},
		loadSound: func(event uint64) int32 {
			w.events = append(w.events, fmt.Sprintf("sound:%#x", event))
			return w.sound[event]
		},
		descriptor: func(soundID int32) uint64 {
			w.events = append(w.events, fmt.Sprintf("descriptor:%d", soundID))
			return w.soundDescriptors[soundID]
		},
		loadFlagsLow: func(descriptor uint64) uint8 {
			w.events = append(w.events, fmt.Sprintf("flags:%#x", descriptor))
			return w.flags[descriptor]
		},
		loadLimit: func(descriptor uint64) int32 {
			w.events = append(w.events, fmt.Sprintf("limit:%#x", descriptor))
			return w.limit[descriptor]
		},
		loadNext: func(event uint64) uint64 {
			w.events = append(w.events, fmt.Sprintf("next:%#x", event))
			return w.next[event]
		},
		storeNext: func(event, next uint64) {
			w.events = append(w.events, fmt.Sprintf("store-next:%#x:%#x", event, next))
			w.next[event] = next
		},
	}
}

func (w *audioEventListWorld501F30) insert() {
	insertAudioEventList501F30(audioListDescriptor501F30, audioListNew501F30, w.hooks())
}

func TestAudioEventList501F30EmptyListStoreOrder(t *testing.T) {
	w := newAudioEventListWorld501F30()
	w.next[audioListNew501F30] = math.MaxUint64
	w.insert()
	want := []string{
		"head:0x300000003",
		"store-next:0x600000006:0x0",
		"store-head:0x300000003:0x600000006",
	}
	if !reflect.DeepEqual(w.events, want) {
		t.Fatalf("events = %v, want %v", w.events, want)
	}
	if w.head[audioListDescriptor501F30] != audioListNew501F30 || w.next[audioListNew501F30] != 0 {
		t.Fatalf("head/next = %#x/%#x", w.head[audioListDescriptor501F30], w.next[audioListNew501F30])
	}
}

func TestAudioEventList501F30OrderingAndRewiring(t *testing.T) {
	t.Run("higher percentage inserts at head", func(t *testing.T) {
		w := newAudioEventListWorld501F30()
		w.head[audioListDescriptor501F30] = audioListHead501F30
		w.percent[audioListHead501F30] = 10
		w.percent[audioListNew501F30] = 20
		w.insert()
		if w.head[audioListDescriptor501F30] != audioListNew501F30 || w.next[audioListNew501F30] != audioListHead501F30 {
			t.Fatalf("head/new next = %#x/%#x", w.head[audioListDescriptor501F30], w.next[audioListNew501F30])
		}
		want := []string{
			"head:0x300000003", "percent:0x600000006", "percent:0x400000004",
			"store-next:0x600000006:0x400000004", "store-head:0x300000003:0x600000006",
		}
		if !reflect.DeepEqual(w.events, want) {
			t.Fatalf("events = %v, want %v", w.events, want)
		}
	})

	t.Run("middle insertion rewires previous node", func(t *testing.T) {
		w := newAudioEventListWorld501F30()
		w.head[audioListDescriptor501F30] = audioListHead501F30
		w.next[audioListHead501F30] = audioListMiddle501F30
		w.percent[audioListHead501F30] = 20
		w.percent[audioListMiddle501F30] = 10
		w.percent[audioListNew501F30] = 15
		w.limit[audioListDescriptor501F30] = 3
		w.insert()
		if w.head[audioListDescriptor501F30] != audioListHead501F30 ||
			w.next[audioListHead501F30] != audioListNew501F30 ||
			w.next[audioListNew501F30] != audioListMiddle501F30 {
			t.Fatalf("links = head %#x, old-next %#x, new-next %#x",
				w.head[audioListDescriptor501F30], w.next[audioListHead501F30], w.next[audioListNew501F30])
		}
	})

	t.Run("lower percentage appends at tail", func(t *testing.T) {
		w := newAudioEventListWorld501F30()
		w.head[audioListDescriptor501F30] = audioListHead501F30
		w.percent[audioListHead501F30] = 20
		w.percent[audioListNew501F30] = 10
		w.limit[audioListDescriptor501F30] = 2
		w.insert()
		if w.head[audioListDescriptor501F30] != audioListHead501F30 ||
			w.next[audioListHead501F30] != audioListNew501F30 || w.next[audioListNew501F30] != 0 {
			t.Fatalf("tail links = head %#x, old-next %#x, new-next %#x",
				w.head[audioListDescriptor501F30], w.next[audioListHead501F30], w.next[audioListNew501F30])
		}
	})
}

func TestAudioEventList501F30NearFlagAndLiveSound(t *testing.T) {
	w := newAudioEventListWorld501F30()
	w.head[audioListDescriptor501F30] = audioListHead501F30
	w.percent[audioListHead501F30] = 10
	w.percent[audioListNew501F30] = 12
	w.sound[audioListNew501F30] = 81
	w.soundDescriptors[81] = audioInsertDescriptor501EA0
	w.flags[audioInsertDescriptor501EA0] = 0x10
	w.insert()
	if w.head[audioListDescriptor501F30] != audioListNew501F30 || w.next[audioListNew501F30] != audioListHead501F30 {
		t.Fatalf("near-priority links = head %#x, new-next %#x",
			w.head[audioListDescriptor501F30], w.next[audioListNew501F30])
	}
	wantMiddle := []string{
		"sound:0x600000006", "descriptor:81", "flags:0x100000001",
	}
	if !reflect.DeepEqual(w.events[3:6], wantMiddle) {
		t.Fatalf("near-priority access = %v, want %v", w.events[3:6], wantMiddle)
	}
}

func TestAudioEventList501F30CapacityDropLeavesNewLinkUntouched(t *testing.T) {
	w := newAudioEventListWorld501F30()
	w.head[audioListDescriptor501F30] = audioListHead501F30
	w.percent[audioListHead501F30] = 20
	w.percent[audioListNew501F30] = 10
	w.limit[audioListDescriptor501F30] = 1
	w.next[audioListNew501F30] = audioListMiddle501F30
	w.insert()
	if w.head[audioListDescriptor501F30] != audioListHead501F30 ||
		w.next[audioListHead501F30] != 0 || w.next[audioListNew501F30] != audioListMiddle501F30 {
		t.Fatalf("drop mutated links = head %#x, head-next %#x, new-next %#x",
			w.head[audioListDescriptor501F30], w.next[audioListHead501F30], w.next[audioListNew501F30])
	}
	wantSuffix := []string{"percent:0x400000004", "limit:0x300000003"}
	if !reflect.DeepEqual(w.events[len(w.events)-2:], wantSuffix) {
		t.Fatalf("drop suffix = %v, want %v", w.events[len(w.events)-2:], wantSuffix)
	}
}

func TestAudioEventList501F30CachesPercentAndUsesInt32Overflow(t *testing.T) {
	t.Run("new percentage is loaded once", func(t *testing.T) {
		w := newAudioEventListWorld501F30()
		w.head[audioListDescriptor501F30] = audioListHead501F30
		w.percent[audioListHead501F30] = 10
		w.percent[audioListNew501F30] = 20
		w.afterPercentLoad = func(event uint64) {
			if event == audioListNew501F30 {
				w.percent[event] = -100
			}
		}
		w.insert()
		if w.head[audioListDescriptor501F30] != audioListNew501F30 {
			t.Fatalf("head = %#x, want cached-percentage event %#x", w.head[audioListDescriptor501F30], audioListNew501F30)
		}
		loads := 0
		for _, event := range w.events {
			if event == "percent:0x600000006" {
				loads++
			}
		}
		if loads != 1 {
			t.Fatalf("new percentage loads = %d, want 1", loads)
		}
	})

	t.Run("subtraction and negation wrap as signed dwords", func(t *testing.T) {
		w := newAudioEventListWorld501F30()
		w.head[audioListDescriptor501F30] = audioListHead501F30
		w.percent[audioListHead501F30] = -1
		w.percent[audioListNew501F30] = math.MaxInt32
		w.sound[audioListNew501F30] = 9
		w.soundDescriptors[9] = audioInsertDescriptor501EA0
		w.limit[audioListDescriptor501F30] = 2
		w.insert()
		if w.head[audioListDescriptor501F30] != audioListHead501F30 ||
			w.next[audioListHead501F30] != audioListNew501F30 {
			t.Fatalf("overflow ordering = head %#x, head-next %#x; want original head then new event",
				w.head[audioListDescriptor501F30], w.next[audioListHead501F30])
		}
	})
}

func TestAudioEventInsertNativeLayout501EA0(t *testing.T) {
	wantEventSize := uintptr(36)
	wantEventSound := uintptr(4)
	wantEventNext := uintptr(28)
	wantEventPercent := uintptr(32)
	wantDescriptorSize := uintptr(28)
	wantDescriptorFlags := uintptr(4)
	wantDescriptorField16 := uintptr(16)
	wantDescriptorLimit := uintptr(20)
	wantDescriptorHead := uintptr(24)
	if unsafe.Sizeof(uintptr(0)) == 8 {
		wantEventSize = 64
		wantEventSound = 8
		wantEventNext = 48
		wantEventPercent = 56
		wantDescriptorSize = 40
		wantDescriptorFlags = 8
		wantDescriptorField16 = 20
		wantDescriptorLimit = 24
		wantDescriptorHead = 32
	}
	checks := []struct {
		name string
		got  uintptr
		want uintptr
	}{
		{"AudioEvent size", unsafe.Sizeof(AudioEvent{}), wantEventSize},
		{"AudioEvent.Sound", unsafe.Offsetof(AudioEvent{}.Sound), wantEventSound},
		{"AudioEvent.list28", unsafe.Offsetof(AudioEvent{}.list28), wantEventNext},
		{"AudioEvent.Perc", unsafe.Offsetof(AudioEvent{}.Perc), wantEventPercent},
		{"audioEvent2 size", unsafe.Sizeof(audioEvent2{}), wantDescriptorSize},
		{"audioEvent2.Flags", unsafe.Offsetof(audioEvent2{}.Flags), wantDescriptorFlags},
		{"audioEvent2.Field16", unsafe.Offsetof(audioEvent2{}.Field16), wantDescriptorField16},
		{"audioEvent2.Field20", unsafe.Offsetof(audioEvent2{}.Field20), wantDescriptorLimit},
		{"audioEvent2.Field24", unsafe.Offsetof(audioEvent2{}.Field24), wantDescriptorHead},
	}
	for _, check := range checks {
		if check.got != check.want {
			t.Errorf("%s = %d, want %d", check.name, check.got, check.want)
		}
	}
}

func TestAudioEventInsertNative501EA0PreservesPointersAndNarrowsDwords(t *testing.T) {
	const soundID = 31
	audio := &serverAudio{}
	descriptor := &audio.bySound[soundID]
	descriptor.Field20 = 2
	stale := &AudioEvent{}
	first := &AudioEvent{Sound: soundID, list28: stale}
	percentage := 7
	if unsafe.Sizeof(uintptr(0)) == 8 {
		widePercentage := uint64(1)<<32 | 7
		percentage = int(widePercentage)
		for name, pointer := range map[string]unsafe.Pointer{
			"descriptor": unsafe.Pointer(descriptor),
			"stale":      unsafe.Pointer(stale),
			"first":      unsafe.Pointer(first),
		} {
			if uintptr(pointer) <= math.MaxUint32 {
				t.Fatalf("%s pointer = %p, want native address above 4 GiB", name, pointer)
			}
		}
	}

	audio.AddAudio(first, percentage)
	if got := testSoundBitmap501EF0(&audio.bitmap, soundID); got != 0x80000000 {
		t.Fatalf("bitmap result = %#x, want raw bit 0x80000000", got)
	}
	if first.Perc != 7 || first.list28 != nil || descriptor.Field24 != first {
		t.Fatalf("first event = percent %d, next %p, head %p", first.Perc, first.list28, descriptor.Field24)
	}

	second := &AudioEvent{Sound: soundID}
	audio.AddAudio(second, 20)
	if descriptor.Field24 != second || second.list28 != first || first.list28 != nil {
		t.Fatalf("native links = head %p, second-next %p, first-next %p", descriptor.Field24, second.list28, first.list28)
	}
	runtime.KeepAlive(stale)
	runtime.KeepAlive(first)
	runtime.KeepAlive(second)
	runtime.KeepAlive(audio)
}
