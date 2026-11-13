package legacy

import (
	"bytes"
	"encoding/binary"
	"github.com/opennox/opennox/v1/legacy/common/alloc"
	"math"
	"runtime"
	"runtime/debug"
	"testing"
	"unsafe"
)

func TestGolemDeathFixtureMappedMassDomain549E90(t *testing.T) {
	address, free := alloc.MallocHigh(80)
	defer free()
	data := unsafe.Slice((*byte)(address), 80)
	for i := range data {
		data[i] = 0x69
	}
	copy(data[16:64], golemOriginalMassBytes549E90())
	mass := unsafe.Add(address, 32)
	before := bytes.Clone(data)
	guard, release := golemOriginalGuardPage549E90()
	if guard == nil {
		t.Fatal("actual protected page")
	}
	defer release()
	t.Run("mapped-neighbourhood", func(t *testing.T) {
		for index := int32(-16); index < 32; index++ {
			old, corrected := golemOriginalByteAddress549E90(mass, index), golemOriginalMappedMassAddress549E90(mass, index, guard)
			if old != corrected || *(*byte)(old) != *(*byte)(corrected) {
				t.Fatal("original mapped byte address/value changed", index)
			}
		}
	})
	t.Run("shared-divisor-byte12", func(t *testing.T) {
		binary.LittleEndian.PutUint32(data[44:], 0x12345680)
		for i, v := range []byte{0x80, 0x56, 0x34, 0x12} {
			if *(*byte)(golemOriginalMappedMassAddress549E90(mass, int32(12+i), guard)) != v {
				t.Fatal("packed divisor byte alias", i)
			}
		}
		copy(data, before)
	})
	for _, tc := range []struct {
		name  string
		index int32
	}{{"unmapped-positive", 0x10000000}, {"unmapped-negative", math.MinInt32}} {
		t.Run(tc.name, func(t *testing.T) {
			old := debug.SetPanicOnFault(true)
			defer debug.SetPanicOnFault(old)
			projected := golemOriginalMappedMassAddress549E90(mass, tc.index, guard)
			if projected != guard {
				t.Fatal("original unmapped domain")
			}
			var recovered any
			func() {
				defer func() { recovered = recover() }()
				value := *(*byte)(projected)
				runtime.KeepAlive(value)
			}()
			if _, ok := recovered.(runtime.Error); !ok {
				t.Fatalf("actual original-domain memory fault required: %T %v", recovered, recovered)
			}
		})
	}

	t.Run("mapped-DWORD-slot-addresses", func(t *testing.T) {
		for _, index := range []int32{-1, 0, 1, 11, 12, 0x7fffffff, math.MinInt32} {
			old, corrected := golemOriginalSlotAddress549E90(mass, index), golemOriginalMappedSlotAddress549E90(mass, index, 0x5cd5f0, guard)
			if old != corrected {
				t.Fatal("original mapped slot identity", index)
			}
		}
	})
	t.Run("unmapped-rock-table", func(t *testing.T) {
		old := debug.SetPanicOnFault(true)
		defer debug.SetPanicOnFault(old)
		projected := golemOriginalMappedSlotAddress549E90(mass, 0x10000000, 0x5cd5f0, guard)
		if projected != guard {
			t.Fatal("original unmapped rock-table domain")
		}
		var recovered any
		func() {
			defer func() { recovered = recover() }()
			value := *(*unsafe.Pointer)(projected)
			runtime.KeepAlive(value)
		}()
		if _, ok := recovered.(runtime.Error); !ok {
			t.Fatalf("actual original table-domain memory fault required: %T %v", recovered, recovered)
		}
	})
	if !bytes.Equal(data, before) {
		t.Fatal("whole mapped byte image or native guards")
	}
}
