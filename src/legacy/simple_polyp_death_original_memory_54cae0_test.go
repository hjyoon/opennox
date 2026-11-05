package legacy

import (
	"bytes"
	"fmt"
	"math"
	"runtime/debug"
	"strings"
	"testing"
	"unsafe"

	"github.com/opennox/libs/types"
	"github.com/opennox/opennox/v1/server"
)

func TestSimplePolypFixtureMemoryDomain54CAE0(t *testing.T) {
	TestSimplePolypDeathNativeLayout54CAE0(t)
	heap, page, free := simplePolypOriginalHeap54CAE0()
	if heap == nil {
		t.Fatal("C-owned mapped heap and protected predecessor page")
	}
	defer free()
	if page <= unsafe.Sizeof(server.Object{}) {
		t.Fatal("entire native object offsets must fit inside one guard page")
	}
	if unsafe.Sizeof(uintptr(0)) == 8 && uintptr(heap) <= math.MaxUint32 {
		t.Fatal("native memory-domain fixture must exceed4GiB")
	}
	ptr := func(role int) *server.Object {
		return (*server.Object)(simplePolypOriginalBad54CAE0(heap, page, role))
	}
	if ptr(-1) != nil || ptr(-2) == nil || ptr(-3) == nil || ptr(-4) == nil || ptr(-2) == ptr(-3) || ptr(-3) == ptr(-4) {
		t.Fatal("independent nil/unmapped/partial pointer identities")
	}
	partial := ptr(-4)
	partial.ObjFlags = 0x11223300
	partial.PosVec = types.Ptf(7, 8)
	p, words := simplePolypOriginalControl54CAE0(partial)
	if p != [4]uintptr{uintptr(unsafe.Pointer(partial)), 0, uintptr(unsafe.Pointer(&partial.PosVec.X)), uintptr(unsafe.Pointer(&partial.PosVec.Y))} || words != [4]uint32{0x11223300, math.Float32bits(7), math.Float32bits(8), 0} {
		t.Fatal("actual typed C partial-record control", p, words)
	}
	if got := math.Float32bits(ptr(-3).PosVec.Y); got != 0x11223300 {
		t.Fatalf("partial source mapped Y = %08x", got)
	}
	data := unsafe.Slice((*byte)(heap), 0x10000)
	before := bytes.Clone(data)
	for _, test := range []struct {
		name string
		read func() uint32
	}{
		{"nil-flags", func() uint32 { return uint32(ptr(-1).ObjFlags) }},
		{"unmapped-flags", func() uint32 { return uint32(ptr(-2).ObjFlags) }},
		{"unmapped-x", func() uint32 { return math.Float32bits(ptr(-2).PosVec.X) }},
		{"unmapped-y", func() uint32 { return math.Float32bits(ptr(-2).PosVec.Y) }},
		{"unmapped-update", func() uint32 { return uint32(uintptr(ptr(-2).UpdateData)) }},
		{"unmapped-duration", func() uint32 { return uint32((*server.ToxicCloudUpdateData)(unsafe.Pointer(ptr(-2))).Duration) }},
		{"partial-source-unmapped-x", func() uint32 { return math.Float32bits(ptr(-3).PosVec.X) }},
	} {
		t.Run(test.name, func(t *testing.T) {
			old := debug.SetPanicOnFault(true)
			defer debug.SetPanicOnFault(old)
			var failure any
			func() {
				defer func() { failure = recover() }()
				_ = test.read()
			}()
			if failure == nil {
				t.Fatal("actual native protected/nil memory read unexpectedly returned")
			}
			text := fmt.Sprint(failure)
			if !strings.Contains(text, "memory") && !strings.Contains(text, "fault") {
				t.Fatal("not an actual native memory fault", failure)
			}
		})
	}
	if !bytes.Equal(data, before) {
		t.Fatal("complete fixture heap changed during independent read controls")
	}
}
