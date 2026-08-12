package legacy

import (
	"fmt"
	"testing"
	"unsafe"

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
		wantSize += 2 * unsafe.Sizeof(uintptr(0))
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
