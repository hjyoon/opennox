package legacy

import (
	"bytes"
	"fmt"
	"reflect"
	"testing"
	"unsafe"

	"github.com/opennox/opennox/v1/common/memmap"
	"github.com/opennox/opennox/v1/legacy/client/audio/ail"
	"github.com/opennox/opennox/v1/legacy/timer"
)

func asnOwnedFixture(t *testing.T) (*asnArena, map[unsafe.Pointer]bool) {
	t.Helper()
	a := newASNArena(t)
	oldAlloc, oldFree, oldGlobal := audioNativeAllocationService, audioNativeReleaseService, audioGlobalOwner
	oldSvc, oldCb := audioSourceTestServices, audioNativeFixtureCallbackService
	oldTicks, oldTimer := PlatformTicks, timer.PlatformTicks
	oldRecord, oldLookup := AudioBankRecordNative, AudioBankLookupNative
	oldSelect, oldRead := AudioBankSelectNative, AudioBankReadNative
	oldClose, oldRemaining := AudioBankCloseReadNative, AudioBankRemainingNative
	oldAvailability := memmap.Uint32(0x5D4594, 1193332)
	owners := make(map[unsafe.Pointer]bool)
	audioNativeAllocationService = func(n uintptr) unsafe.Pointer {
		p := a.take(n)
		owners[p] = true
		return p
	}
	audioNativeReleaseService = func(p unsafe.Pointer) {
		if !owners[p] {
			t.Fatalf("double free or cross-owner release: %p", p)
		}
		delete(owners, p)
	}
	PlatformTicks = func() uint64 { return 0x100000047 }
	timer.PlatformTicks = PlatformTicks
	audioGlobalOwner = (*audioGlobalNative)(a.take(unsafe.Sizeof(audioGlobalNative{})))
	audioGlobalOwner.Devices.init()
	audioGlobalOwner.Contexts.init()
	audioGlobalOwner.Group.Init()
	audioGlobalOwner.GroupInitialized = 1
	t.Cleanup(func() {
		audioNativeAllocationService, audioNativeReleaseService, audioGlobalOwner = oldAlloc, oldFree, oldGlobal
		audioSourceTestServices, audioNativeFixtureCallbackService = oldSvc, oldCb
		PlatformTicks, timer.PlatformTicks = oldTicks, oldTimer
		AudioBankRecordNative, AudioBankLookupNative = oldRecord, oldLookup
		AudioBankSelectNative, AudioBankReadNative = oldSelect, oldRead
		AudioBankCloseReadNative, AudioBankRemainingNative = oldClose, oldRemaining
		*memmap.PtrUint32(0x5D4594, 1193332) = oldAvailability
		a.guards()
		if len(owners) != 0 {
			t.Errorf("retained native owners: %d", len(owners))
		}
	})
	return a, owners
}

func TestAudioSourceNativeGlobalAllocationBeforeClock(t *testing.T) {
	_, _ = asnOwnedFixture(t)
	previous, previousClock, previousLegacyClock := audioGlobalOwner, timer.PlatformTicks, PlatformTicks
	audioGlobalOwner, timer.PlatformTicks, PlatformTicks = nil, nil, nil
	var owner *audioGlobalNative
	defer func() {
		audioGlobalOwner, timer.PlatformTicks, PlatformTicks = previous, previousClock, previousLegacyClock
		if owner != nil {
			audioNativeFree(unsafe.Pointer(owner))
		}
	}()
	owner = audioNativeGlobal()
	if owner.Devices.first() != nil || owner.Contexts.first() != nil || owner.GroupInitialized != 0 || owner.Group != (timer.TimerGroup{}) {
		t.Fatal("allocation/data binding must not initialize the audio clock")
	}
	calls := 0
	timer.PlatformTicks = func() uint64 { calls++; return 0x100000047 }
	if audioNativeMasterGroup() != &owner.Group || owner.GroupInitialized != 1 || calls == 0 {
		t.Fatal("first audio use must initialize its native master group")
	}
	owner.Group.Timers[0].SetRaw(0x1000)
	owner.Group.Timers[0].Update()
	before := calls
	if audioNativeMasterGroup() != &owner.Group || owner.Group.Timers[0].Current != 0x10000000 || calls != before {
		t.Fatal("later voices must retain the configured master gain and clock state")
	}
}

func TestAudioSourceNativePoolCABI(t *testing.T) {
	_, _ = asnOwnedFixture(t)
	for _, size := range []int32{0, 1, 3, 4, 7, 16, 37, 84} {
		t.Run(fmt.Sprintf("payload-%d", size), func(t *testing.T) {
			pool := asnPointer(9, nil, nil, 3, size, 0)
			if pool == nil {
				t.Fatal("pool allocation")
			}
			items := make([]unsafe.Pointer, 3)
			for i := range items {
				items[i] = asnPointer(10, pool, nil, 0, 0, 0)
				if items[i] == nil || uintptr(items[i])%unsafe.Sizeof(uintptr(0)) != 0 {
					t.Fatal("unaligned/truncated pool item")
				}
				for j := range unsafe.Slice((*byte)(items[i]), int(size)) {
					unsafe.Slice((*byte)(items[i]), int(size))[j] = byte(i*31 + j)
				}
			}
			if asnPointer(10, pool, nil, 0, 0, 0) != nil {
				t.Fatal("exhausted pool")
			}
			for _, i := range []int{1, 0, 2} {
				if got := asnPointer(11, pool, items[i], 0, 0, 0); got != unsafe.Add(items[i], -int(unsafe.Sizeof(uintptr(0)))) {
					t.Fatal("native free header")
				}
			}
			for _, i := range []int{2, 0, 1} {
				if got := asnPointer(10, pool, nil, 0, 0, 0); got != items[i] {
					t.Fatal("free-list LIFO order/pointer")
				}
				for j, b := range unsafe.Slice((*byte)(items[i]), int(size)) {
					if b != byte(i*31+j) {
						t.Fatal("free-list header overwrote payload")
					}
				}
			}
			asnPointer(12, pool, nil, 0, 0, 0)
		})
	}
	if asnPointer(9, nil, nil, 0, 8, 0) != nil || asnPointer(9, nil, nil, -1, 8, 0) != nil || asnPointer(9, nil, nil, 1, -1, 0) != nil {
		t.Fatal("invalid native allocation sizes must be rejected")
	}
}

func TestAudioSourceNativeCacheProducerPumpCABI(t *testing.T) {
	a, _ := asnOwnedFixture(t)
	bank := a.take(16)
	rows := (*[2][9]uint32)(a.take(2 * 36))
	rows[0][6], rows[0][7], rows[0][8] = 22050, 9, 256
	rows[1][6], rows[1][7], rows[1][8] = 11025, 4, 0
	payloads := [][]byte{[]byte("abcdefghijkl"), []byte("XYZ")}
	selected, offset, closed := -1, 0, 0
	AudioBankRecordNative = func(p unsafe.Pointer, id int32) unsafe.Pointer {
		if p != bank || id < 0 || id >= 2 {
			t.Fatal("bank/id pointer changed")
		}
		return unsafe.Pointer(&rows[id][0])
	}
	AudioBankSelectNative = func(p unsafe.Pointer, id int32) int32 {
		if p != bank {
			t.Fatal("select bank pointer")
		}
		selected, offset = int(id), 0
		return 1
	}
	AudioBankRemainingNative = func(p unsafe.Pointer) uint32 {
		if p != bank {
			t.Fatal("remaining bank pointer")
		}
		return uint32(len(payloads[selected]) - offset)
	}
	AudioBankReadNative = func(p unsafe.Pointer, data *byte, n int32) int32 {
		if p != bank {
			t.Fatal("read bank pointer")
		}
		count := copy(unsafe.Slice(data, int(n)), payloads[selected][offset:])
		offset += count
		return int32(count)
	}
	AudioBankCloseReadNative = func(p unsafe.Pointer) unsafe.Pointer {
		if p != bank {
			t.Fatal("close bank pointer")
		}
		closed++
		return nil
	}
	cache := asnPointer(13, bank, nil, 3*(5+24), 2, 5)
	if cache == nil {
		t.Fatal("native cache")
	}
	first := asnPointer(14, cache, nil, 0, 0, 0)
	if first == nil {
		t.Fatal("three-chunk sample")
	}
	buffer := (*AudioBufferNative)(asnPointer(17, first, nil, 0, 0, 0))
	var payload []byte
	var sizes []uint32
	for n := buffer.Chunks.first(); n != nil; n = n.next() {
		c := (*AudioChunkNative)(unsafe.Pointer(n))
		if c.Buffer != buffer {
			t.Fatal("native chunk owner")
		}
		payload = append(payload, unsafe.Slice(c.Data, int(c.Size))...)
		sizes = append(sizes, c.Size)
	}
	if !bytes.Equal(payload, payloads[0]) || !reflect.DeepEqual(sizes, []uint32{5, 5, 2}) || buffer.Size != 12 || *buffer.Format != [7]uint32{4, 2, 22050, 2, 2, 0, 256} {
		t.Fatal("sample bytes, native chunk sizes, total or format")
	}
	if asnPointer(15, cache, nil, 0, 0, 0) != first || asnPointer(14, cache, nil, 0, 0, 0) != first || closed != 1 {
		t.Fatal("cache hit must preserve owner without reopening")
	}
	asnPointer(16, first, nil, 0, 0, 0)
	if asnScalar(11, first, nil, 0) != 1 || asnPointer(14, cache, nil, 1, 0, 0) != nil || closed != 1 {
		t.Fatal("retained sample must prevent chunk eviction")
	}
	if asnScalar(10, first, nil, 0) != 0 {
		t.Fatal("release reference")
	}
	second := asnPointer(14, cache, nil, 1, 0, 0)
	if second == nil || closed != 2 || asnPointer(15, cache, nil, 0, 0, 0) != nil {
		t.Fatal("LRU eviction after release")
	}
	buffer = (*AudioBufferNative)(asnPointer(17, second, nil, 0, 0, 0))
	if buffer.Size != 3 || *buffer.Format != [7]uint32{4, 0, 11025, 1, 2, 0, 0} {
		t.Fatal("replacement sample format")
	}
	ctx := (*AudioContextNative)(a.take(unsafe.Sizeof(AudioContextNative{})))
	methods := (*AudioMethodsNative)(a.take(unsafe.Sizeof(AudioMethodsNative{})))
	methods.Create, methods.Destroy = asnCallback(14, false), asnCallback(15, true)
	ctx.Methods = methods
	createCalls, destroyCalls := 0, 0
	audioNativeFixtureCallbackService = func(kind int, p unsafe.Pointer) int32 {
		v := (*AudioVoiceNative)(p)
		if v.Context != ctx || v.Methods != methods {
			t.Fatal("native source constructor pointer")
		}
		switch kind {
		case 14:
			createCalls++
		case 15:
			destroyCalls++
		default:
			t.Fatal("unexpected method callback")
		}
		return 0
	}
	voice := (*AudioVoiceNative)(asnPointer(5, unsafe.Pointer(ctx), nil, 0, 0, 0))
	if voice == nil || createCalls != 1 {
		t.Fatal("source constructor")
	}
	asnPointer(3, unsafe.Pointer(voice), unsafe.Pointer(buffer), 0, 0, 0)
	sample := (*AudioSample)(a.take(unsafe.Sizeof(AudioSample{})))
	sample.Field1, sample.Smp = unsafe.Pointer(voice), ail.Sample(uintptr(a.take(16)))
	sample.Data1, sample.Data2 = (*byte)(a.take(0x4000)), (*byte)(a.take(0x4000))
	ready, loads := 0, 0
	audioSourceTestServices = &audioSourceServices{
		UserData: func(h ail.Sample) *AudioSample {
			if h != sample.Smp {
				t.Fatal("EOS native handle")
			}
			return sample
		},
		Ready: func(h ail.Sample) int {
			if h != sample.Smp {
				t.Fatal("ready native handle")
			}
			ready++
			if ready <= 2 {
				return ready - 1
			}
			return -1
		},
		Load: func(h ail.Sample, index uint32, p unsafe.Pointer, n uint32) {
			if h != sample.Smp || index != uint32(loads) {
				t.Fatal("Miles boundary handle/index")
			}
			if loads == 0 && (!bytes.Equal(unsafe.Slice((*byte)(p), int(n)), payloads[1]) || p != unsafe.Pointer((*AudioChunkNative)(unsafe.Pointer(buffer.Chunks.first())).Data)) {
				t.Fatal("real cache→producer→pump zero-copy payload")
			}
			if loads == 1 && (n != 0 || p != nil) {
				t.Fatal("EOF buffer")
			}
			loads++
		},
	}
	if asnScalar(0, unsafe.Pointer(sample), nil, 0) != -1 || loads != 2 || sample.Field3 != 1 {
		t.Fatal("connected pump completion")
	}
	asnScalar(1, unsafe.Pointer(uintptr(sample.Smp)), nil, 0)
	if sample.Flag7 != 1 || voice.Buffer != nil || voice.Flags&5 != 0 {
		t.Fatal("connected EOS producer reset")
	}
	// This voice was constructed outside the context list, so release it
	// through its matching native destroy owner rather than unlinking it.
	sub_4BD7A0(unsafe.Pointer(voice))
	if destroyCalls != 1 {
		t.Fatal("source destructor")
	}
	asnPointer(18, cache, nil, 0, 0, 0)
}
