package legacy

import (
	"fmt"
	"testing"
	"unsafe"

	noxflags "github.com/opennox/opennox/v1/common/flags"
	"github.com/opennox/opennox/v1/common/memmap"
	"github.com/opennox/opennox/v1/common/ntype"
	"github.com/opennox/opennox/v1/legacy/common/alloc"
	"github.com/opennox/opennox/v1/legacy/common/alloc/handles"
)

type importantRateControl struct {
	ResendsPerUpdate byte
	ResendInterval   byte
	UpdateRate       byte
	Reserved3        byte
	Threshold        uint32
	LowerThreshold   uint32
}

type importantPacketLegacy struct {
	CreatedFrame         uint32
	LastSendFrame        [32]uint32
	RetryDelay           [32]byte
	SendCount            byte
	Reserved165          [3]byte
	AcknowledgedMask     uint32
	SentMask             uint32
	RecipientMask        uint32
	RemoveIfDisconnected uint32
	SequenceEnabled      byte
	Reserved185          byte
	Sequence             [32]uint16
	Recipient            int8
	Payload              [150]byte
	PayloadSize          byte
	Reserved402          [2]byte
	LegacyRelatedObject  uint32
	LegacyNext           uint32
	LegacyPrev           uint32
}

func preserveImportantPacketState(t *testing.T) {
	t.Helper()
	oldFrameHook := gameFrameHook
	oldAlloc := importantAllocClassC()
	oldNativeFirst, oldNativeLast := importantNativeListC()
	oldRawFirst, oldRawLast := importantListHeadsC()
	oldRecipientMask := importantRecipientMaskC()
	oldCounters := *memmap.PtrT[[32]uint16](0x5D4594, 1565524)
	oldCapacity := importantCapacityC()
	t.Cleanup(func() {
		currentAlloc := importantAllocClassC()
		if currentAlloc != nil && currentAlloc != oldAlloc {
			setImportantNativeListC(nil, nil)
			setImportantAllocClassC(nil)
			if pool := alloc.AsClass(currentAlloc); pool != nil {
				pool.Free()
			}
		}
		gameFrameHook = oldFrameHook
		setImportantNativeListC(oldNativeFirst, oldNativeLast)
		setImportantListHeadsC(oldRawFirst, oldRawLast)
		setImportantRecipientMaskC(oldRecipientMask)
		*memmap.PtrT[[32]uint16](0x5D4594, 1565524) = oldCounters
		setImportantCapacityC(oldCapacity)
		setImportantAllocClassC(oldAlloc)
	})
}

func TestImportantStateInitMatchesGAMEEXEContract(t *testing.T) {
	if got := unsafe.Sizeof(importantRateControl{}); got != 12 {
		t.Fatalf("rate-control record size: got %d, want 12", got)
	}
	if got := unsafe.Offsetof(importantRateControl{}.ResendsPerUpdate); got != 0 {
		t.Fatalf("resends-per-update offset: got %d, want 0", got)
	}
	if got := unsafe.Offsetof(importantRateControl{}.ResendInterval); got != 1 {
		t.Fatalf("resend-interval offset: got %d, want 1", got)
	}
	if got := unsafe.Offsetof(importantRateControl{}.UpdateRate); got != 2 {
		t.Fatalf("update-rate offset: got %d, want 2", got)
	}
	if got := unsafe.Offsetof(importantRateControl{}.Threshold); got != 4 {
		t.Fatalf("threshold offset: got %d, want 4", got)
	}
	if got := unsafe.Offsetof(importantRateControl{}.LowerThreshold); got != 8 {
		t.Fatalf("lower-threshold offset: got %d, want 8", got)
	}
	handles.Init()
	t.Cleanup(handles.Release)

	oldFPS := gameFPSHook
	oldMode := Get_dword_5d4594_2650652()
	oldRate := *memmap.PtrUint32(0x587000, 4728)
	oldFirst, oldLast := importantListHeadsC()
	oldNativeFirst, oldNativeLast := importantNativeListC()
	oldRecords := *memmap.PtrT[[32]importantRateControl](0x5D4594, 1565124)
	oldCounters := *memmap.PtrT[[32]uint16](0x5D4594, 1565524)
	oldCapacity := *memmap.PtrUint32(0x5D4594, 1565520)
	oldBefore := *memmap.PtrUint8(0x5D4594, 1565123)
	oldAfter := *memmap.PtrUint8(0x5D4594, 1565588)
	t.Cleanup(func() {
		gameFPSHook = oldFPS
		Set_dword_5d4594_2650652(oldMode)
		*memmap.PtrUint32(0x587000, 4728) = oldRate
		setImportantNativeListC(oldNativeFirst, oldNativeLast)
		setImportantListHeadsC(oldFirst, oldLast)
		*memmap.PtrT[[32]importantRateControl](0x5D4594, 1565124) = oldRecords
		*memmap.PtrT[[32]uint16](0x5D4594, 1565524) = oldCounters
		*memmap.PtrUint32(0x5D4594, 1565520) = oldCapacity
		*memmap.PtrUint8(0x5D4594, 1565123) = oldBefore
		*memmap.PtrUint8(0x5D4594, 1565588) = oldAfter
	})

	gameFPSHook = func() uint32 { return 30 }
	*memmap.PtrUint32(0x587000, 4728) = 3

	for _, tc := range []struct {
		name      string
		mode      int
		threshold uint32
	}{
		{name: "normal", mode: 0, threshold: 60},
		{name: "rate-limited", mode: 1, threshold: 20},
	} {
		t.Run(tc.name, func(t *testing.T) {
			records := memmap.PtrT[[32]importantRateControl](0x5D4594, 1565124)
			for i := range records {
				records[i] = importantRateControl{
					ResendsPerUpdate: 0xA1,
					ResendInterval:   0xA2,
					UpdateRate:       0xA3,
					Reserved3:        0xA4,
					Threshold:        0xA5A5A5A5,
					LowerThreshold:   0xA6A6A6A6,
				}
			}
			counters := memmap.PtrT[[32]uint16](0x5D4594, 1565524)
			for i := range counters {
				counters[i] = 0x5A5A
			}
			*memmap.PtrUint32(0x5D4594, 1565520) = 0x89ABCDEF
			*memmap.PtrUint8(0x5D4594, 1565123) = 0xC3
			*memmap.PtrUint8(0x5D4594, 1565588) = 0xD4
			setImportantListHeadsC(0x13579BDF, 0x2468ACE0)
			Set_dword_5d4594_2650652(tc.mode)

			pool := alloc.NewClass("important-init-contract", importantPacketSizeC(), 1)
			handle := pool.UPtr()
			setImportantAllocClassC(handle)
			packet := pool.NewObject()
			setImportantNativeListC(packet, packet)
			if got := importantAllocClassC(); got != handle {
				t.Fatalf("native allocation class: got %p, want %p", got, handle)
			}
			legacyHandle := *memmap.PtrUint32(0x5D4594, 1565508)
			if unsafe.Sizeof(uintptr(0)) == 4 {
				if legacyHandle != uint32(uintptr(handle)) {
					t.Fatalf("32-bit allocation mirror: got %#x, want %#x", legacyHandle, uint32(uintptr(handle)))
				}
			} else if legacyHandle != 0 {
				t.Fatalf("64-bit legacy allocation slot: got %#x, want 0", legacyHandle)
			}
			t.Cleanup(func() {
				if alloc.AsClass(handle) != nil {
					pool.Free()
				}
				setImportantAllocClassC(nil)
			})

			if got := Sub_4E4DE0(); got != int(tc.threshold) {
				t.Errorf("return value: got %d, want %d", got, tc.threshold)
			}
			if importantAllocClassC() != nil || alloc.AsClass(handle) != nil {
				t.Fatal("initialization did not free and clear the native allocation class")
			}
			if got := *memmap.PtrUint32(0x5D4594, 1565508); got != 0 {
				t.Fatalf("legacy allocation slot: got %#x, want 0", got)
			}
			if first, last := importantListHeadsC(); first != 0 || last != 0 {
				t.Errorf("list heads: got (%#x, %#x), want (0, 0)", first, last)
			}
			if first, last := importantNativeListC(); first != nil || last != nil {
				t.Errorf("native list heads: got (%p, %p), want (nil, nil)", first, last)
			}
			for i, got := range records {
				want := importantRateControl{
					ResendsPerUpdate: 1,
					ResendInterval:   2,
					UpdateRate:       3,
					Threshold:        tc.threshold,
				}
				if got != want {
					t.Errorf("record %d: got %+v, want %+v", i, got, want)
				}
			}
			for i, got := range counters {
				if got != 0 {
					t.Errorf("counter %d: got %#x, want 0", i, got)
				}
			}
			if got := *memmap.PtrUint32(0x5D4594, 1565520); got != 0x89ABCDEF {
				t.Errorf("capacity changed: got %#x, want %#x", got, uint32(0x89ABCDEF))
			}
			if got := *memmap.PtrUint8(0x5D4594, 1565123); got != 0xC3 {
				t.Errorf("byte before records changed: got %#x, want %#x", got, byte(0xC3))
			}
			if got := *memmap.PtrUint8(0x5D4594, 1565588); got != 0xD4 {
				t.Errorf("byte after counters changed: got %#x, want %#x", got, byte(0xD4))
			}
		})
	}
}

func TestImportantRateThresholdsMatchGAMEEXEContract(t *testing.T) {
	oldFPS := gameFPSHook
	oldMode := Get_dword_5d4594_2650652()
	oldRate := *memmap.PtrUint32(0x587000, 4728)
	oldRecords := *memmap.PtrT[[32]importantRateControl](0x5D4594, 1565124)
	t.Cleanup(func() {
		gameFPSHook = oldFPS
		Set_dword_5d4594_2650652(oldMode)
		*memmap.PtrUint32(0x587000, 4728) = oldRate
		*memmap.PtrT[[32]importantRateControl](0x5D4594, 1565124) = oldRecords
	})

	tests := []struct {
		name          string
		index         int
		mode          int
		updateRate    uint32
		fps           []uint32
		resends       byte
		interval      byte
		wantThreshold uint32
		wantLower     uint32
		wantReturn    int32
	}{
		{name: "interval-zero", index: 0, mode: 0, updateRate: 0, fps: []uint32{30}, resends: 7, interval: 0},
		{name: "interval-one", index: 31, mode: 0, updateRate: 0, fps: []uint32{30}, resends: 3, interval: 1, wantThreshold: 90, wantReturn: 90},
		{name: "interval-two", index: 0, mode: 2, updateRate: 0, fps: []uint32{30}, resends: 3, interval: 2, wantThreshold: 180, wantReturn: 180},
		{name: "lower-threshold", index: 31, mode: -1, updateRate: 0, fps: []uint32{60, 60}, resends: 3, interval: 3, wantThreshold: 540, wantLower: 360, wantReturn: 540},
		{name: "independent-fps-loads", index: 0, mode: 0, updateRate: 0, fps: []uint32{30, 60}, resends: 1, interval: 3, wantThreshold: 180, wantLower: 60, wantReturn: 180},
		{name: "divide-before-multiply", index: 31, mode: 1, updateRate: 3, fps: []uint32{31, 31}, resends: 2, interval: 3, wantThreshold: 60, wantLower: 40, wantReturn: 60},
		{name: "rate-limited", index: 0, mode: 1, updateRate: 3, fps: []uint32{60, 60}, resends: 3, interval: 5, wantThreshold: 300, wantLower: 240, wantReturn: 300},
		{name: "uint32-wrap", index: 31, mode: 0, updateRate: 0, fps: []uint32{^uint32(0), ^uint32(0)}, resends: 0xFF, interval: 0xFF, wantThreshold: 0xFFFF01FF, wantLower: 0xFFFF02FE, wantReturn: -65025},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			records := memmap.PtrT[[32]importantRateControl](0x5D4594, 1565124)
			for i := range records {
				records[i] = importantRateControl{
					ResendsPerUpdate: byte(0x40 + i),
					ResendInterval:   byte(0x60 + i),
					UpdateRate:       byte(0x80 + i),
					Reserved3:        byte(0xA0 + i),
					Threshold:        0xB0000000 | uint32(i),
					LowerThreshold:   0xC0000000 | uint32(i),
				}
			}
			records[tc.index].ResendsPerUpdate = tc.resends
			records[tc.index].ResendInterval = tc.interval
			before := *records
			Set_dword_5d4594_2650652(tc.mode)
			*memmap.PtrUint32(0x587000, 4728) = tc.updateRate
			fpsCalls := 0
			gameFPSHook = func() uint32 {
				v := tc.fps[fpsCalls]
				fpsCalls++
				return v
			}

			if got := updateImportantRateControlC(tc.index); got != tc.wantReturn {
				t.Errorf("return value: got %d (%#x), want %d (%#x)", got, uint32(got), tc.wantReturn, uint32(tc.wantReturn))
			}
			if fpsCalls != len(tc.fps) {
				t.Errorf("FPS loads: got %d, want %d", fpsCalls, len(tc.fps))
			}
			got := records[tc.index]
			if got.Threshold != tc.wantThreshold || got.LowerThreshold != tc.wantLower {
				t.Errorf("thresholds: got (%#x, %#x), want (%#x, %#x)", got.Threshold, got.LowerThreshold, tc.wantThreshold, tc.wantLower)
			}
			if got.ResendsPerUpdate != tc.resends || got.ResendInterval != tc.interval ||
				got.UpdateRate != before[tc.index].UpdateRate || got.Reserved3 != before[tc.index].Reserved3 {
				t.Errorf("calculation modified input bytes: got %+v, before %+v", got, before[tc.index])
			}
			for i := range records {
				if i != tc.index && records[i] != before[i] {
					t.Errorf("record %d changed: got %+v, want %+v", i, records[i], before[i])
				}
			}
		})
	}
}

func TestImportantPlayerCountersResetMatchesGAMEEXEContract(t *testing.T) {
	counters := memmap.PtrT[[32]uint16](0x5D4594, 1565524)
	before := memmap.PtrUint8(0x5D4594, 1565523)
	after := memmap.PtrUint8(0x5D4594, 1565588)
	oldCounters := *counters
	oldBefore := *before
	oldAfter := *after
	t.Cleanup(func() {
		*counters = oldCounters
		*before = oldBefore
		*after = oldAfter
	})

	for i := range counters {
		counters[i] = uint16(0x8001 + i*0x101)
	}
	*before = 0xA5
	*after = 0x5A

	if got := Sub_4E4ED0(); got != 0 {
		t.Errorf("return value: got %d, want 0", got)
	}
	for i, got := range counters {
		if got != 0 {
			t.Errorf("counter %d: got %#x, want 0", i, got)
		}
	}
	if got := *before; got != 0xA5 {
		t.Errorf("byte before counters changed: got %#x, want %#x", got, byte(0xA5))
	}
	if got := *after; got != 0x5A {
		t.Errorf("byte after counters changed: got %#x, want %#x", got, byte(0x5A))
	}
}

func TestImportantRateControlsResetMatchesGAMEEXEContract(t *testing.T) {
	oldFPS := gameFPSHook
	oldMode := Get_dword_5d4594_2650652()
	rateValue := memmap.PtrUint32(0x587000, 4728)
	oldRate := *rateValue
	records := memmap.PtrT[[32]importantRateControl](0x5D4594, 1565124)
	oldRecords := *records
	before := memmap.PtrUint8(0x5D4594, 1565123)
	after := memmap.PtrUint8(0x5D4594, 1565508)
	oldBefore := *before
	oldAfter := *after
	t.Cleanup(func() {
		gameFPSHook = oldFPS
		Set_dword_5d4594_2650652(oldMode)
		*rateValue = oldRate
		*records = oldRecords
		*before = oldBefore
		*after = oldAfter
	})

	for _, tc := range []struct {
		name string
		mode int
	}{
		{name: "normal", mode: 0},
		{name: "rate-limited", mode: 1},
	} {
		t.Run(tc.name, func(t *testing.T) {
			for i := range records {
				records[i] = importantRateControl{
					ResendsPerUpdate: 0xA1,
					ResendInterval:   0xA2,
					UpdateRate:       0xA3,
					Reserved3:        0xA4,
					Threshold:        0xA5A5A5A5,
					LowerThreshold:   0xA6A6A6A6,
				}
			}
			*before = 0xC3
			*after = 0xD4
			*rateValue = 1
			Set_dword_5d4594_2650652(tc.mode)
			fpsCalls := 0
			gameFPSHook = func() uint32 {
				fps := uint32(96 + fpsCalls)
				fpsCalls++
				*rateValue = uint32(fpsCalls + 1)
				return fps
			}

			wantDivisor := uint32(1)
			if tc.mode == 1 {
				wantDivisor = 32
			}
			wantReturn := int(2 * (uint32(127) / wantDivisor))
			if got := Sub_4E4EF0(); got != wantReturn {
				t.Errorf("return value: got %d, want %d", got, wantReturn)
			}
			if fpsCalls != 32 {
				t.Errorf("FPS loads: got %d, want 32", fpsCalls)
			}
			for i, got := range records {
				divisor := uint32(1)
				if tc.mode == 1 {
					divisor = uint32(i + 1)
				}
				want := importantRateControl{
					ResendsPerUpdate: 1,
					ResendInterval:   2,
					UpdateRate:       byte(i + 1),
					Threshold:        2 * (uint32(96+i) / divisor),
				}
				if got != want {
					t.Errorf("record %d: got %+v, want %+v", i, got, want)
				}
			}
			if got := *before; got != 0xC3 {
				t.Errorf("byte before records changed: got %#x, want %#x", got, byte(0xC3))
			}
			if got := *after; got != 0xD4 {
				t.Errorf("byte after records changed: got %#x, want %#x", got, byte(0xD4))
			}
		})
	}
}

func TestImportantPlayerCounterResetMatchesGAMEEXEContract(t *testing.T) {
	counters := memmap.PtrT[[32]uint16](0x5D4594, 1565524)
	before := memmap.PtrUint8(0x5D4594, 1565523)
	after := memmap.PtrUint8(0x5D4594, 1565588)
	oldCounters := *counters
	oldBefore := *before
	oldAfter := *after
	t.Cleanup(func() {
		*counters = oldCounters
		*before = oldBefore
		*after = oldAfter
	})

	for _, ind := range []int{0, 31} {
		t.Run(fmt.Sprintf("index-%d", ind), func(t *testing.T) {
			for i := range counters {
				counters[i] = uint16(0x4001 + i*0x101)
			}
			want := *counters
			want[ind] = 0
			*before = 0xA5
			*after = 0x5A

			if got := resetImportantPlayerCounterC(ind); got != ind {
				t.Errorf("return value: got %d, want %d", got, ind)
			}
			if got := *counters; got != want {
				t.Errorf("counters: got %#v, want %#v", got, want)
			}
			if got := *before; got != 0xA5 {
				t.Errorf("byte before counters changed: got %#x, want %#x", got, byte(0xA5))
			}
			if got := *after; got != 0x5A {
				t.Errorf("byte after counters changed: got %#x, want %#x", got, byte(0x5A))
			}
		})
	}
}

func TestImportantPlayerRateControlResetMatchesGAMEEXEContract(t *testing.T) {
	oldFPS := gameFPSHook
	oldMode := Get_dword_5d4594_2650652()
	rateValue := memmap.PtrUint32(0x587000, 4728)
	oldRate := *rateValue
	records := memmap.PtrT[[32]importantRateControl](0x5D4594, 1565124)
	oldRecords := *records
	before := memmap.PtrUint8(0x5D4594, 1565123)
	after := memmap.PtrUint8(0x5D4594, 1565508)
	oldBefore := *before
	oldAfter := *after
	t.Cleanup(func() {
		gameFPSHook = oldFPS
		Set_dword_5d4594_2650652(oldMode)
		*rateValue = oldRate
		*records = oldRecords
		*before = oldBefore
		*after = oldAfter
	})

	for _, mode := range []int{0, 1} {
		for _, ind := range []int{0, 31} {
			t.Run(fmt.Sprintf("mode-%d-index-%d", mode, ind), func(t *testing.T) {
				for i := range records {
					records[i] = importantRateControl{
						ResendsPerUpdate: byte(0x40 + i),
						ResendInterval:   byte(0x60 + i),
						UpdateRate:       byte(0x80 + i),
						Reserved3:        byte(0xA0 + i),
						Threshold:        0xB0000000 | uint32(i),
						LowerThreshold:   0xC0000000 | uint32(i),
					}
				}
				want := *records
				divisor := uint32(1)
				if mode == 1 {
					divisor = 3
				}
				want[ind] = importantRateControl{
					ResendsPerUpdate: 1,
					ResendInterval:   2,
					UpdateRate:       3,
					Reserved3:        byte(0xA0 + ind),
					Threshold:        2 * (uint32(31) / divisor),
				}
				*before = 0xC3
				*after = 0xD4
				*rateValue = 3
				Set_dword_5d4594_2650652(mode)
				fpsCalls := 0
				gameFPSHook = func() uint32 {
					fpsCalls++
					*rateValue = 7
					return 31
				}

				wantReturn := int(want[ind].Threshold)
				if got := Nox_xxx_playerResetImportantCtr_4E4F40(ntype.PlayerInd(ind)); got != wantReturn {
					t.Errorf("return value: got %d, want %d", got, wantReturn)
				}
				if fpsCalls != 1 {
					t.Errorf("FPS loads: got %d, want 1", fpsCalls)
				}
				if got := *records; got != want {
					t.Errorf("records: got %#v, want %#v", got, want)
				}
				if got := *before; got != 0xC3 {
					t.Errorf("byte before records changed: got %#x, want %#x", got, byte(0xC3))
				}
				if got := *after; got != 0xD4 {
					t.Errorf("byte after records changed: got %#x, want %#x", got, byte(0xD4))
				}
			})
		}
	}
}

func TestImportantPacketCleanupMatchesGAMEEXEContract(t *testing.T) {
	const legacySize = uintptr(416)
	wantSize := legacySize
	if unsafe.Sizeof(uintptr(0)) > 4 {
		wantSize += 3 * unsafe.Sizeof(uintptr(0))
	}
	if got := importantPacketSizeC(); got != wantSize {
		t.Fatalf("packet allocation size: got %d, want %d", got, wantSize)
	}

	handles.Init()
	t.Cleanup(handles.Release)
	oldAlloc := importantAllocClassC()
	oldNativeFirst, oldNativeLast := importantNativeListC()
	oldRawFirst, oldRawLast := importantListHeadsC()
	pool := alloc.NewClass("important-cleanup-contract", wantSize, 5)
	handle := pool.UPtr()
	setImportantAllocClassC(handle)
	setImportantNativeListC(nil, nil)
	t.Cleanup(func() {
		setImportantNativeListC(nil, nil)
		setImportantAllocClassC(nil)
		if alloc.AsClass(handle) != nil {
			pool.Free()
		}
		setImportantNativeListC(oldNativeFirst, oldNativeLast)
		setImportantListHeadsC(oldRawFirst, oldRawLast)
		setImportantAllocClassC(oldAlloc)
	})

	packets := make([]unsafe.Pointer, 5)
	before := make([][legacySize]byte, len(packets))
	types := [...]byte{0x31, 0x30, 0x32, 0x34, 0x33}
	for i := range packets {
		packets[i] = pool.NewObject()
		if packets[i] == nil {
			t.Fatalf("packet %d allocation failed", i)
		}
	}
	for i := range packets {
		data := unsafe.Slice((*byte)(packets[i]), legacySize)
		for j := range data {
			data[j] = byte((i*37 + j*13) % 251)
		}
		data[251] = types[i]
		if i > 0 {
			setImportantPacketPrevC(packets[i], packets[i-1])
		} else {
			setImportantPacketPrevC(packets[i], nil)
		}
		if i+1 < len(packets) {
			setImportantPacketNextC(packets[i], packets[i+1])
		} else {
			setImportantPacketNextC(packets[i], nil)
		}
		copy(before[i][:], data)
	}
	setImportantNativeListC(packets[0], packets[len(packets)-1])

	if got := cleanupImportantPacketsC(); got != 0 {
		t.Errorf("return value: got %d, want 0", got)
	}
	first, last := importantNativeListC()
	if first != packets[1] || last != packets[3] {
		t.Fatalf("surviving list ends: got (%p, %p), want (%p, %p)", first, last, packets[1], packets[3])
	}
	if got := importantPacketPrevC(packets[1]); got != nil {
		t.Errorf("surviving head prev: got %p, want nil", got)
	}
	if got := importantPacketNextC(packets[1]); got != packets[3] {
		t.Errorf("surviving head next: got %p, want %p", got, packets[3])
	}
	if got := importantPacketPrevC(packets[3]); got != packets[1] {
		t.Errorf("surviving tail prev: got %p, want %p", got, packets[1])
	}
	if got := importantPacketNextC(packets[3]); got != nil {
		t.Errorf("surviving tail next: got %p, want nil", got)
	}
	rawLink := func(packet unsafe.Pointer, offset uintptr) uint32 {
		return *(*uint32)(unsafe.Add(packet, offset))
	}
	if unsafe.Sizeof(uintptr(0)) == 4 {
		if got := rawLink(packets[1], 408); got != uint32(uintptr(packets[3])) {
			t.Errorf("32-bit surviving head raw next: got %#x, want %#x", got, uint32(uintptr(packets[3])))
		}
		if got := rawLink(packets[1], 412); got != 0 {
			t.Errorf("32-bit surviving head raw prev: got %#x, want 0", got)
		}
		if got := rawLink(packets[3], 408); got != 0 {
			t.Errorf("32-bit surviving tail raw next: got %#x, want 0", got)
		}
		if got := rawLink(packets[3], 412); got != uint32(uintptr(packets[1])) {
			t.Errorf("32-bit surviving tail raw prev: got %#x, want %#x", got, uint32(uintptr(packets[1])))
		}
	} else {
		for _, i := range []int{1, 3} {
			if next, prev := rawLink(packets[i], 408), rawLink(packets[i], 412); next != 0 || prev != 0 {
				t.Errorf("64-bit packet %d legacy links: got (%#x, %#x), want (0, 0)", i, next, prev)
			}
		}
	}

	for _, i := range []int{0, 2, 4} {
		data := unsafe.Slice((*byte)(packets[i]), wantSize)
		for off, got := range data {
			if got != alloc.DeadChar {
				t.Errorf("removed packet %d byte %d: got %#x, want %#x", i, off, got, byte(alloc.DeadChar))
				break
			}
		}
	}
	for _, i := range []int{1, 3} {
		data := unsafe.Slice((*byte)(packets[i]), legacySize)
		for off, got := range data {
			if off >= 408 {
				continue
			}
			if got != before[i][off] {
				t.Errorf("retained packet %d byte %d: got %#x, want %#x", i, off, got, before[i][off])
				break
			}
		}
	}

	rawFirst, rawLast := importantListHeadsC()
	if unsafe.Sizeof(uintptr(0)) == 4 {
		if rawFirst != uint32(uintptr(packets[1])) || rawLast != uint32(uintptr(packets[3])) {
			t.Errorf("32-bit list mirrors: got (%#x, %#x), want (%#x, %#x)", rawFirst, rawLast, uint32(uintptr(packets[1])), uint32(uintptr(packets[3])))
		}
	} else if rawFirst != 0 || rawLast != 0 {
		t.Errorf("64-bit legacy list slots: got (%#x, %#x), want (0, 0)", rawFirst, rawLast)
	}

	removeImportantPacketC(packets[1])
	first, last = importantNativeListC()
	if first != packets[3] || last != packets[3] || importantPacketPrevC(packets[3]) != nil || importantPacketNextC(packets[3]) != nil {
		t.Fatalf("single-node list after head removal: got ends (%p, %p)", first, last)
	}
	removeImportantPacketC(packets[3])
	if first, last = importantNativeListC(); first != nil || last != nil {
		t.Fatalf("empty list after tail removal: got (%p, %p)", first, last)
	}
	if rawFirst, rawLast = importantListHeadsC(); rawFirst != 0 || rawLast != 0 {
		t.Errorf("empty legacy list slots: got (%#x, %#x), want (0, 0)", rawFirst, rawLast)
	}
	if got := cleanupImportantPacketsC(); got != 0 {
		t.Errorf("empty-list return value: got %d, want 0", got)
	}
}

func TestImportantPacketCreationMatchesGAMEEXEContract(t *testing.T) {
	const legacySize = uintptr(416)
	if got := unsafe.Sizeof(importantPacketLegacy{}); got != legacySize {
		t.Fatalf("Go legacy packet size: got %d, want %d", got, legacySize)
	}
	offsets := []struct {
		name string
		got  uintptr
		want uintptr
	}{
		{name: "created frame", got: unsafe.Offsetof(importantPacketLegacy{}.CreatedFrame), want: 0},
		{name: "last-send frames", got: unsafe.Offsetof(importantPacketLegacy{}.LastSendFrame), want: 4},
		{name: "retry delays", got: unsafe.Offsetof(importantPacketLegacy{}.RetryDelay), want: 132},
		{name: "send count", got: unsafe.Offsetof(importantPacketLegacy{}.SendCount), want: 164},
		{name: "acknowledged mask", got: unsafe.Offsetof(importantPacketLegacy{}.AcknowledgedMask), want: 168},
		{name: "sent mask", got: unsafe.Offsetof(importantPacketLegacy{}.SentMask), want: 172},
		{name: "recipient mask", got: unsafe.Offsetof(importantPacketLegacy{}.RecipientMask), want: 176},
		{name: "disconnect policy", got: unsafe.Offsetof(importantPacketLegacy{}.RemoveIfDisconnected), want: 180},
		{name: "sequence switch", got: unsafe.Offsetof(importantPacketLegacy{}.SequenceEnabled), want: 184},
		{name: "sequences", got: unsafe.Offsetof(importantPacketLegacy{}.Sequence), want: 186},
		{name: "recipient", got: unsafe.Offsetof(importantPacketLegacy{}.Recipient), want: 250},
		{name: "payload", got: unsafe.Offsetof(importantPacketLegacy{}.Payload), want: 251},
		{name: "payload size", got: unsafe.Offsetof(importantPacketLegacy{}.PayloadSize), want: 401},
		{name: "related object", got: unsafe.Offsetof(importantPacketLegacy{}.LegacyRelatedObject), want: 404},
		{name: "next", got: unsafe.Offsetof(importantPacketLegacy{}.LegacyNext), want: 408},
		{name: "prev", got: unsafe.Offsetof(importantPacketLegacy{}.LegacyPrev), want: 412},
	}
	for _, off := range offsets {
		if off.got != off.want {
			t.Errorf("%s offset: got %d, want %d", off.name, off.got, off.want)
		}
	}
	wantNativeSize := legacySize
	if unsafe.Sizeof(uintptr(0)) > 4 {
		wantNativeSize += 3 * unsafe.Sizeof(uintptr(0))
	}
	if got := importantPacketSizeC(); got != wantNativeSize {
		t.Fatalf("native packet size: got %d, want %d", got, wantNativeSize)
	}

	handles.Init()
	t.Cleanup(handles.Release)
	preserveImportantPacketState(t)
	setImportantNativeListC(nil, nil)
	setImportantAllocClassC(nil)

	setImportantRecipientMaskC(uint32(1) << 2)
	if got := sendImportantPacketC(0x82, []byte{0x31}, nil, 0, true); got != 1 {
		t.Errorf("excluded-only return value: got %d, want 1", got)
	}
	if importantAllocClassC() != nil {
		t.Fatal("excluded-only send allocated a packet class")
	}
	if first, last := importantNativeListC(); first != nil || last != nil {
		t.Fatalf("excluded-only list: got (%p, %p), want (nil, nil)", first, last)
	}

	setImportantRecipientMaskC((uint32(1) << 0) | (uint32(1) << 2))
	if got := sendImportantPacketC(0, make([]byte, 151), nil, 0, false); got != 0 {
		t.Errorf("oversized return value: got %d, want 0", got)
	}
	if importantAllocClassC() != nil {
		t.Fatal("oversized send allocated a packet class")
	}

	const packetCount = 4
	pool := alloc.NewClass("important-create-contract", wantNativeSize, packetCount)
	setImportantAllocClassC(pool.UPtr())
	mask := (uint32(1) << 0) | (uint32(1) << 2) | (uint32(1) << 31)
	setImportantRecipientMaskC(mask)
	counters := memmap.PtrT[[32]uint16](0x5D4594, 1565524)
	for i := range counters {
		counters[i] = uint16(0x1000 + i*0x10)
	}
	initialCounters := *counters
	frame := uint32(0x12345670)
	gameFrameHook = func() uint32 { return frame }
	relatedObject, freeRelatedObject := alloc.Malloc(1)
	t.Cleanup(freeRelatedObject)

	type createdPacket struct {
		ptr       unsafe.Pointer
		frame     uint32
		recipient int8
		payload   []byte
		remove    uint32
		sequence  bool
		related   unsafe.Pointer
	}
	created := make([]createdPacket, 0, packetCount)
	send := func(recipient int, payload []byte, related unsafe.Pointer, remove int, sequence bool) {
		t.Helper()
		wantFrame := frame
		if got := sendImportantPacketC(recipient, payload, related, remove, sequence); got != 1 {
			t.Fatalf("send recipient %#x: got %d, want 1", recipient, got)
		}
		first, _ := importantNativeListC()
		if first == nil {
			t.Fatalf("send recipient %#x did not create a list head", recipient)
		}
		created = append(created, createdPacket{
			ptr: first, frame: wantFrame, recipient: int8(uint8(recipient)), payload: append([]byte(nil), payload...),
			remove: uint32(remove), sequence: sequence, related: related,
		})
		frame++
	}

	payloadAtLimit := make([]byte, 150)
	for i := range payloadAtLimit {
		payloadAtLimit[i] = byte((i*29 + 0x31) & 0xFF)
	}
	send(255, payloadAtLimit, relatedObject, 0x12345678, true)
	broadcast := created[len(created)-1].ptr
	send(2, []byte{0x40, 0x02}, nil, 1, true)
	single := created[len(created)-1].ptr
	send(0x82, []byte{0x41, 0x82, 0x7F, 0x00}, nil, 0, true)
	excluded := created[len(created)-1].ptr
	send(0, []byte{0x42}, nil, 0, false)
	withoutSequence := created[len(created)-1].ptr

	for _, item := range created {
		packet := (*importantPacketLegacy)(item.ptr)
		if packet.CreatedFrame != item.frame {
			t.Errorf("recipient %#x created frame: got %#x, want %#x", uint8(item.recipient), packet.CreatedFrame, item.frame)
		}
		if packet.Recipient != item.recipient {
			t.Errorf("recipient byte: got %#x, want %#x", uint8(packet.Recipient), uint8(item.recipient))
		}
		if packet.RecipientMask != mask {
			t.Errorf("recipient %#x mask: got %#x, want %#x", uint8(item.recipient), packet.RecipientMask, mask)
		}
		if packet.RemoveIfDisconnected != item.remove {
			t.Errorf("recipient %#x disconnect policy: got %#x, want %#x", uint8(item.recipient), packet.RemoveIfDisconnected, item.remove)
		}
		wantSequence := byte(0)
		if item.sequence {
			wantSequence = 1
		}
		if packet.SequenceEnabled != wantSequence {
			t.Errorf("recipient %#x sequence switch: got %d, want %d", uint8(item.recipient), packet.SequenceEnabled, wantSequence)
		}
		if packet.PayloadSize != byte(len(item.payload)) {
			t.Errorf("recipient %#x payload size: got %d, want %d", uint8(item.recipient), packet.PayloadSize, len(item.payload))
		}
		for i, want := range item.payload {
			if got := packet.Payload[i]; got != want {
				t.Errorf("recipient %#x payload byte %d: got %#x, want %#x", uint8(item.recipient), i, got, want)
			}
		}
		for i, got := range packet.Payload[len(item.payload):] {
			if got != 0 {
				t.Errorf("recipient %#x payload tail byte %d: got %#x, want 0", uint8(item.recipient), len(item.payload)+i, got)
				break
			}
		}
		if packet.SendCount != 0 || packet.AcknowledgedMask != 0 || packet.SentMask != 0 {
			t.Errorf("recipient %#x send state: count=%d acknowledged=%#x sent=%#x, want zero", uint8(item.recipient), packet.SendCount, packet.AcknowledgedMask, packet.SentMask)
		}
		for i, got := range packet.LastSendFrame {
			if got != 0 {
				t.Errorf("recipient %#x last-send frame %d: got %#x, want 0", uint8(item.recipient), i, got)
				break
			}
		}
		for i, got := range packet.RetryDelay {
			if got != 0 {
				t.Errorf("recipient %#x retry delay %d: got %#x, want 0", uint8(item.recipient), i, got)
				break
			}
		}
		if got := importantPacketRelatedObjectC(item.ptr); got != item.related {
			t.Errorf("recipient %#x related object: got %p, want %p", uint8(item.recipient), got, item.related)
		}
		if unsafe.Sizeof(uintptr(0)) == 4 {
			if packet.LegacyRelatedObject != uint32(uintptr(item.related)) {
				t.Errorf("recipient %#x 32-bit related slot: got %#x, want %#x", uint8(item.recipient), packet.LegacyRelatedObject, uint32(uintptr(item.related)))
			}
		} else if packet.LegacyRelatedObject != 0 {
			t.Errorf("recipient %#x 64-bit related slot: got %#x, want 0", uint8(item.recipient), packet.LegacyRelatedObject)
		}
	}

	wantBroadcastSequence := [32]uint16{}
	wantBroadcastSequence[0] = initialCounters[0]
	wantBroadcastSequence[2] = initialCounters[2]
	wantBroadcastSequence[31] = initialCounters[31]
	if got := (*importantPacketLegacy)(broadcast).Sequence; got != wantBroadcastSequence {
		t.Errorf("broadcast sequences: got %#v, want %#v", got, wantBroadcastSequence)
	}
	wantSingleSequence := [32]uint16{}
	wantSingleSequence[2] = initialCounters[2] + 1
	if got := (*importantPacketLegacy)(single).Sequence; got != wantSingleSequence {
		t.Errorf("single-recipient sequences: got %#v, want %#v", got, wantSingleSequence)
	}
	wantExcludedSequence := [32]uint16{}
	wantExcludedSequence[0] = initialCounters[0] + 1
	wantExcludedSequence[31] = initialCounters[31] + 1
	if got := (*importantPacketLegacy)(excluded).Sequence; got != wantExcludedSequence {
		t.Errorf("excluded-recipient sequences: got %#v, want %#v", got, wantExcludedSequence)
	}
	if got := (*importantPacketLegacy)(withoutSequence).Sequence; got != [32]uint16{} {
		t.Errorf("sequence-disabled values: got %#v, want zero", got)
	}
	wantCounters := initialCounters
	wantCounters[0] += 2
	wantCounters[2] += 2
	wantCounters[31] += 2
	if got := *counters; got != wantCounters {
		t.Errorf("player counters: got %#v, want %#v", got, wantCounters)
	}

	first, last := importantNativeListC()
	if first != withoutSequence || last != broadcast {
		t.Fatalf("list ends: got (%p, %p), want (%p, %p)", first, last, withoutSequence, broadcast)
	}
	wantOrder := []unsafe.Pointer{withoutSequence, excluded, single, broadcast}
	for i, packet := range wantOrder {
		var wantPrev, wantNext unsafe.Pointer
		if i > 0 {
			wantPrev = wantOrder[i-1]
		}
		if i+1 < len(wantOrder) {
			wantNext = wantOrder[i+1]
		}
		if got := importantPacketPrevC(packet); got != wantPrev {
			t.Errorf("list item %d prev: got %p, want %p", i, got, wantPrev)
		}
		if got := importantPacketNextC(packet); got != wantNext {
			t.Errorf("list item %d next: got %p, want %p", i, got, wantNext)
		}
	}
}

func TestImportantPacketAllocationRecoveryMatchesGAMEEXEContract(t *testing.T) {
	handles.Init()
	t.Cleanup(handles.Release)

	t.Run("empty-list-failure", func(t *testing.T) {
		preserveImportantPacketState(t)
		setImportantNativeListC(nil, nil)
		setImportantRecipientMaskC(1)
		pool := alloc.NewClass("important-empty-full-contract", importantPacketSizeC(), 1)
		setImportantAllocClassC(pool.UPtr())
		if consumed := pool.NewObject(); consumed == nil {
			t.Fatal("failed to consume the only allocation")
		}
		if got := sendImportantPacketC(0, []byte{0x31}, nil, 0, false); got != 0 {
			t.Errorf("return value: got %d, want 0", got)
		}
		if first, last := importantNativeListC(); first != nil || last != nil {
			t.Errorf("list after failed recovery: got (%p, %p), want (nil, nil)", first, last)
		}
	})

	t.Run("oldest-packet-reused", func(t *testing.T) {
		preserveImportantPacketState(t)
		setImportantRecipientMaskC(1)
		pool := alloc.NewClass("important-reuse-contract", importantPacketSizeC(), 1)
		setImportantAllocClassC(pool.UPtr())
		oldest := pool.NewObject()
		if oldest == nil {
			t.Fatal("failed to allocate the oldest packet")
		}
		oldRecord := (*importantPacketLegacy)(oldest)
		oldRecord.CreatedFrame = 7
		oldRecord.Recipient = 31
		setImportantPacketPrevC(oldest, nil)
		setImportantPacketNextC(oldest, nil)
		setImportantNativeListC(oldest, oldest)
		gameFrameHook = func() uint32 { return 11 }

		if got := sendImportantPacketC(0, []byte{0x44, 0x55}, nil, 0, false); got != 1 {
			t.Fatalf("return value: got %d, want 1", got)
		}
		first, last := importantNativeListC()
		if first != oldest || last != oldest {
			t.Fatalf("reused list ends: got (%p, %p), want (%p, %p)", first, last, oldest, oldest)
		}
		got := (*importantPacketLegacy)(oldest)
		if got.CreatedFrame != 11 || got.Recipient != 0 || got.PayloadSize != 2 || got.Payload[0] != 0x44 || got.Payload[1] != 0x55 {
			t.Errorf("reused packet contents: frame=%d recipient=%d size=%d payload=%#x, want frame=11 recipient=0 size=2 payload=4455", got.CreatedFrame, got.Recipient, got.PayloadSize, got.Payload[:2])
		}
		if importantPacketPrevC(oldest) != nil || importantPacketNextC(oldest) != nil {
			t.Error("reused single packet has non-nil links")
		}
	})
}

func TestImportantPacketLazyAllocationMatchesGAMEEXEContract(t *testing.T) {
	handles.Init()
	t.Cleanup(handles.Release)

	for _, tc := range []struct {
		name     string
		flags    noxflags.GameFlag
		capacity uint32
	}{
		{name: "normal-static", capacity: 256},
		{name: "host-static", flags: noxflags.GameHost, capacity: 3072},
		{name: "coop-dynamic-precedence", flags: noxflags.GameHost | noxflags.GameModeCoop, capacity: 512},
	} {
		t.Run(tc.name, func(t *testing.T) {
			oldFlags := noxflags.GetGame()
			t.Cleanup(func() {
				noxflags.ResetGame()
				noxflags.SetGame(oldFlags)
			})
			preserveImportantPacketState(t)
			noxflags.ResetGame()
			noxflags.SetGame(tc.flags)
			setImportantAllocClassC(nil)
			setImportantNativeListC(nil, nil)
			setImportantRecipientMaskC(1)
			gameFrameHook = func() uint32 { return 0x10203040 }

			if got := sendImportantPacketC(0, []byte{0x45}, nil, 0, false); got != 1 {
				t.Fatalf("return value: got %d, want 1", got)
			}
			if importantAllocClassC() == nil {
				t.Fatal("lazy allocation did not create a packet class")
			}
			if got := importantCapacityC(); got != tc.capacity {
				t.Errorf("capacity: got %d, want %d", got, tc.capacity)
			}
			first, last := importantNativeListC()
			if first == nil || first != last {
				t.Fatalf("lazy-created list ends: got (%p, %p), want one shared non-nil packet", first, last)
			}
			packet := (*importantPacketLegacy)(first)
			if packet.CreatedFrame != 0x10203040 || packet.PayloadSize != 1 || packet.Payload[0] != 0x45 {
				t.Errorf("lazy-created packet: frame=%#x size=%d payload=%#x", packet.CreatedFrame, packet.PayloadSize, packet.Payload[0])
			}
		})
	}
}
