package legacy

import (
	"testing"
	"unsafe"

	"github.com/opennox/opennox/v1/common/memmap"
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
	oldRecords := *memmap.PtrT[[32]importantRateControl](0x5D4594, 1565124)
	oldCounters := *memmap.PtrT[[32]uint16](0x5D4594, 1565524)
	oldCapacity := *memmap.PtrUint32(0x5D4594, 1565520)
	oldBefore := *memmap.PtrUint8(0x5D4594, 1565123)
	oldAfter := *memmap.PtrUint8(0x5D4594, 1565588)
	t.Cleanup(func() {
		gameFPSHook = oldFPS
		Set_dword_5d4594_2650652(oldMode)
		*memmap.PtrUint32(0x587000, 4728) = oldRate
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

			pool := alloc.NewClass("important-init-contract", 416, 1)
			handle := pool.UPtr()
			setImportantAllocClassC(handle)
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
