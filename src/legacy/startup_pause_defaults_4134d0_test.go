//go:build darwin && arm64

package legacy

import (
	"bytes"
	"encoding/binary"
	"runtime"
	"testing"
	"unsafe"

	flags "github.com/opennox/opennox/v1/common/flags"
	"github.com/opennox/opennox/v1/common/memmap"
)

func spSaveCells4134D0(t *testing.T) {
	t.Helper()
	var prior [6]uint32
	for i := range prior {
		prior[i] = *(*uint32)(startupPauseDefaultCell4134D0(i))
	}
	t.Cleanup(func() {
		for i, v := range prior {
			*(*uint32)(startupPauseDefaultCell4134D0(i)) = v
		}
	})
}

func TestStartupPauseNativeDefaults4134D0(t *testing.T) {
	t.Run("startup_C_globals", func(t *testing.T) {
		spSaveCells4134D0(t)
		offsets := []uintptr{251708, 251712, 251716, 251720, 251744}
		var prior [5]uint32
		for i, off := range offsets {
			p := (*uint32)(memmap.PtrOff(0x5D4594, off))
			prior[i] = *p
		}
		defer func() {
			for i, off := range offsets {
				*(*uint32)(memmap.PtrOff(0x5D4594, off)) = prior[i]
			}
		}()
		for i, off := range offsets {
			p := (*uint32)(memmap.PtrOff(0x5D4594, off))
			*p = 0x11223300 + uint32(i)
			field := i
			if i == 4 {
				field = 5
			}
			if unsafe.Pointer(p) == startupPauseDefaultCell4134D0(field) {
				t.Fatal("C global aliased memmap shadow")
			}
		}
		for i := 0; i < 6; i++ {
			*(*uint32)(startupPauseDefaultCell4134D0(i)) = 0x80000000 + uint32(i)
		}
		if startupPauseFixtureCall4134D0(0x4134d0, 0, nil, nil) != 0 {
			t.Fatal("current startup policy return")
		}
		for i := 0; i < 6; i++ {
			want := uint32(0)
			if i >= 4 {
				want = 0x80000000 + uint32(i)
			}
			if *(*uint32)(startupPauseDefaultCell4134D0(i)) != want {
				t.Fatal("actual C owner reset extent")
			}
		}
		for i, off := range offsets {
			if *(*uint32)(memmap.PtrOff(0x5D4594, off)) != 0x11223300+uint32(i) {
				t.Fatal("memmap shadow changed")
			}
		}
	})
	t.Run("disabled_CD_buffer", func(t *testing.T) {
		p := startupPauseDefaultBuffer4134D0()
		b := unsafe.Slice((*byte)(p), 12)
		prior := append([]byte(nil), b...)
		defer copy(b, prior)
		copy(b, spPattern4134D0(0x6a)[:12])
		want := append([]byte(nil), b...)
		clear(want[:2])
		if startupPauseFixtureDrive4134D0() != nil || !bytes.Equal(b, want) {
			t.Fatal("two BYTE stores/current disabled fallback")
		}
	})
	t.Run("actual_guard_query", func(t *testing.T) {
		spSaveCells4134D0(t)
		for _, v := range []uint32{0, 1, 0x80000000, 0xffffffff} {
			*(*uint32)(startupPauseDefaultCell4134D0(5)) = v
			want := int32(0)
			if v != 0 {
				want = 1
			}
			if startupPauseFixtureCall4134D0(0x4139b0, 0, nil, nil) != want {
				t.Fatal("canonical int32 C global query")
			}
		}
	})
	t.Run("default_pause_query_service", func(t *testing.T) {
		prior := flags.GetGame()
		defer func() { flags.UnsetGame(flags.GetGame()); flags.SetGame(prior) }()
		for _, on := range []bool{false, true} {
			flags.UnsetGame(0x40000)
			want := uint32(0)
			if on {
				flags.SetGame(0x40000)
				want = 1
			}
			if startupPauseDefaultService4134D0(0x4138e0, nil) != want {
				t.Fatal("actual C int check of exported C bool service")
			}
		}
	})
	t.Run("typed_focus_registration", func(t *testing.T) {
		spSaveCells4134D0(t)
		startupPauseFixtureFocusBegin4134D0()
		defer startupPauseFixtureFocusEnd4134D0()
		var pauses []uint32
		o := &startupPauseObserver4134D0{Service: func(site uint32, args []startupPauseArg4134D0) (uint32, bool) {
			switch site {
			case 0x41392e, 0x413941:
				return 0, false
			case 0x4138e0:
				return 0x80000000, true
			case 0x413900:
				return 0, true
			case 0x4138ec, 0x413914:
				pauses = append(pauses, args[0].Word)
				return 0, true
			default:
				t.Fatal("unexpected service")
			}
			return 0, true
		}}
		pop := pushStartupPauseObserver4134D0(o)
		defer pop()
		if startupPauseFixtureCall4134D0(0x413920, 0, nil, nil) != 1 {
			t.Fatal("registration return")
		}
		for _, kind := range []int{1, 2} {
			if startupPauseFixtureFocusCount4134D0(kind) != 1 || !startupPauseFixtureFocusRecord4134D0(kind, 0) || uintptr(startupPauseFixtureFocusOwner4134D0(kind)) <= 0xffffffff || uintptr(startupPauseCallback4134D0(uint32(kind))) <= 0xffffffff {
				t.Fatal("typed native callback array/identity")
			}
		}
		*(*uint32)(startupPauseDefaultCell4134D0(4)) = 0
		startupPauseFixtureFocusInvoke4134D0(1, 0, -1)
		startupPauseFixtureFocusInvoke4134D0(2, 0, -2147483648)
		if len(pauses) != 2 || pauses[0] != 0 || pauses[1] != 1 || *(*uint32)(startupPauseDefaultCell4134D0(4)) != 0x80000000 {
			t.Fatal("stored callbacks use full reply/order")
		}
	})
	t.Run("actual_const_qsort", func(t *testing.T) {
		b := make([]byte, 4*84)
		keys := []uint32{4, 1, 3, 2}
		records := make([][]byte, 4)
		for i, key := range keys {
			for n := 0; n < 84; n++ {
				b[i*84+n] = byte(0xb0 + i)
			}
			binary.LittleEndian.PutUint32(b[i*84+80:], key)
			records[i] = append([]byte(nil), b[i*84:(i+1)*84]...)
		}
		var pin runtime.Pinner
		pin.Pin(&b[0])
		defer pin.Unpin()
		startupPauseFixtureQsort4134D0(unsafe.Pointer(&b[0]), 4)
		seen := map[uint32]bool{}
		for i := range keys {
			r := b[i*84 : (i+1)*84]
			key := binary.LittleEndian.Uint32(r[80:])
			if key != uint32(i+1) || seen[key] {
				t.Fatal("native qsort raw comparator")
			}
			seen[key] = true
			found := false
			for _, old := range records {
				found = found || bytes.Equal(r, old)
			}
			if !found {
				t.Fatal("whole84-byte record changed")
			}
		}
	})
}

func TestStartupPauseNativePinnedAndReentry4134D0(t *testing.T) {
	t.Run("pinned_unaligned_records", func(t *testing.T) {
		for _, backend := range []string{"Go", "C"} {
			var a, b [85]byte
			binary.LittleEndian.PutUint32(a[81:], 0x80000000)
			binary.LittleEndian.PutUint32(b[81:], 0xffffffff)
			var pin runtime.Pinner
			pin.Pin(&a[0])
			pin.Pin(&b[0])
			func() {
				defer pin.Unpin()
				reads := 0
				o := &startupPauseObserver4134D0{Operand: func(e startupPauseOperand4134D0) {
					if e.Complete && !e.Write {
						reads++
						if reads == 1 {
							runtime.GC()
							binary.LittleEndian.PutUint32(b[81:], 0)
						}
					}
				}}
				pop := pushStartupPauseObserver4134D0(o)
				defer pop()
				ap, bp := unsafe.Pointer(&a[1]), unsafe.Pointer(&b[1])
				if uintptr(ap) <= 0xffffffff || uintptr(bp) <= 0xffffffff {
					t.Fatal("actual pinned high Go heap")
				}
				var result int32
				if backend == "C" {
					result = startupPauseFixtureCall4134D0(0x413f60, 0, ap, bp)
				} else {
					result = sub413F60Native(ap, bp)
				}
				if uint32(result) != 0x80000000 || reads != 2 {
					t.Fatal("unaligned fullptr/ordered cached read/GC/live second operand")
				}
				runtime.KeepAlive(a)
				runtime.KeepAlive(b)
			}()
		}
	})
	t.Run("nested_observer_cached_pending", func(t *testing.T) {
		var outer, inner [6]uint32
		outer[3] = 1
		inner[5] = 0x80000000
		var pin runtime.Pinner
		pin.Pin(&outer[0])
		pin.Pin(&inner[0])
		defer pin.Unpin()
		o, i := &startupPauseObserver4134D0{}, &startupPauseObserver4134D0{}
		for n := range outer {
			o.Cells[n] = unsafe.Pointer(&outer[n])
			i.Cells[n] = unsafe.Pointer(&inner[n])
		}
		nested := int32(0)
		calls := 0
		o.Operand = func(e startupPauseOperand4134D0) {
			if e.Site == 0x4137e0 && e.Complete {
				func() {
					pop := pushStartupPauseObserver4134D0(i)
					defer pop()
					runtime.GC()
					nested = startupPauseFixtureCall4134D0(0x4139b0, 0, nil, nil)
				}()
				outer[3] = 0xffffffff
			}
		}
		o.Service = func(site uint32, _ []startupPauseArg4134D0) (uint32, bool) {
			if site != 0x4137f3 || outer[3] != 0 {
				t.Fatal("cached branch then clear before music")
			}
			calls++
			outer[3] = 0xcafebabe
			return 0, true
		}
		pop := pushStartupPauseObserver4134D0(o)
		defer pop()
		startupPauseFixtureCall4134D0(0x4137e0, 0, nil, nil)
		if nested != 1 || calls != 1 || outer[3] != 0xcafebabe {
			t.Fatal("nested C reentry/default owner isolation")
		}
	})
	t.Run("observer_LIFO", func(t *testing.T) {
		outer := pushStartupPauseObserver4134D0(&startupPauseObserver4134D0{})
		inner := pushStartupPauseObserver4134D0(&startupPauseObserver4134D0{})
		rejected := false
		func() { defer func() { rejected = recover() != nil }(); outer() }()
		inner()
		outer()
		if !rejected {
			t.Fatal("observer LIFO guard")
		}
	})
}
