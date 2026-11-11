package legacy

import (
	"bytes"
	"fmt"
	"github.com/opennox/opennox/v1/common/memmap"
	"github.com/opennox/opennox/v1/legacy/common/alloc"
	"math"
	"runtime"
	"testing"
	"unsafe"
)

func g2bTestIsolate(t *testing.T) {
	t.Helper()
	var words [5]uint32
	for i := range words {
		words[i] = g2bTestWord(i, 0, false)
	}
	var ptrs [3]unsafe.Pointer
	for i := range ptrs {
		ptrs[i] = g2bTestPointer(i, nil, false)
	}
	bh, qh := game2BriefingTestHooks, game2QuickbarTestHooks
	byteFlag := memmap.Uint8(0x5D4594, 831252)
	stage := memmap.Uint32(0x5D4594, 832468)
	image, text := *memmap.PtrPtr(0x5D4594, 832460), *memmap.PtrPtr(0x5D4594, 832464)
	t.Cleanup(func() {
		for i, v := range words {
			g2bTestWord(i, v, true)
		}
		for i, v := range ptrs {
			g2bTestPointer(i, v, true)
		}
		game2BriefingTestHooks = bh
		game2QuickbarTestHooks = qh
		*memmap.PtrUint8(0x5D4594, 831252) = byteFlag
		*memmap.PtrUint32(0x5D4594, 832468) = stage
		*memmap.PtrPtr(0x5D4594, 832460) = image
		*memmap.PtrPtr(0x5D4594, 832464) = text
	})
}

func TestGame2BriefingNativeThingFlagsAndScalarABI(t *testing.T) {
	g2bTestIsolate(t)
	data, free := alloc.MakeHigh([]byte{}, int(g2bTestLayout(0))+32)
	defer free()
	for i := range data {
		data[i] = 0xa7
	}
	p := unsafe.Pointer(&data[16])
	game2BriefingTestHooks = &game2BriefingHooks{Thing: func(id int32) unsafe.Pointer {
		if id == 99 {
			return p
		}
		return nil
	}}
	if unsafe.Sizeof(uintptr(0)) == 8 && uintptr(p) <= math.MaxUint32 {
		t.Fatal("native thing pointer")
	}
	for _, pri := range []uint32{0, 0x400000, 0x20000000, 0x20400000, 0xffffffff} {
		for _, sub := range []uint32{0, 8, 16, 24, 0xffffffff} {
			*(*uint32)(unsafe.Add(p, g2bTestLayout(1))) = pri
			*(*uint32)(unsafe.Add(p, g2bTestLayout(2))) = sub
			before := append([]byte(nil), data...)
			runtime.GC()
			expected := []int32{int32(pri >> 22 & 1), 0, 0}
			if pri&0x400000 != 0 && sub&0x18 != 0 {
				expected[1] = 1
			}
			if pri&0x20400000 != 0 {
				expected[2] = 1
			}
			for op, w := range expected {
				if g2bTestCall(op, 99, nil) != w || g2bTestCall(op, 0, nil) != 0 {
					t.Fatal("typed Thing flags/public nil")
				}
			}
			if !bytes.Equal(data, before) {
				t.Fatal("thing/canary mutation")
			}
		}
	}
	for _, value := range []uint32{0, 1, 0x80000000, 0xffffffff} {
		g2bTestWord(0, 0xface, true)
		g2bTestWord(1, value, true)
		if g2bTestCall(4, 0, nil) != int32(value) {
			t.Fatal("copy flag DWORD return")
		}
		want := uint32(0xface)
		if value != 0 {
			want = 1
		}
		if g2bTestWord(0, 0, false) != want {
			t.Fatal("conditional flag write")
		}
		g2bTestCall(3, 0, nil)
		if g2bTestCall(5, 0, nil) != 0 {
			t.Fatal("flag clear")
		}
	}
	for _, v := range []int{0, 1, 127, 128, 254, 255, -1} {
		if g2bTestCall(6, v, nil) != int32(byte(v)) || g2bTestCall(7, 0, nil) != int32(byte(v)) {
			t.Fatal("char/BYTE ABI")
		}
	}
	for _, v := range []int{0, -1, math.MinInt32, math.MaxInt32} {
		g2bTestCall(8, v, nil)
		if g2bTestCall(9, 0, nil) != int32(v) {
			t.Fatal("quest stage DWORD")
		}
	}
	for _, v := range []uint32{0, 1, 254, 255, 256, 0xffffffff} {
		g2bTestWord(2, v, true)
		want := float64(0)
		if v == 255 {
			want = 1
		}
		if g2bTestScroll() != want {
			t.Fatal("scroll double ABI")
		}
	}
}

func TestGame2BriefingNativeScoreAndPointerSideSlots(t *testing.T) {
	g2bTestIsolate(t)
	size, off := g2bTestLayout(3), g2bTestLayout(4)
	data, free := alloc.MakeHigh([]byte{}, int(size*2+32))
	defer free()
	for i := range data {
		data[i] = 0xa7
	}
	a, b := unsafe.Pointer(&data[16]), unsafe.Pointer(&data[16+int(size)])
	for _, x := range []uint32{0, 1, 0x7fffffff, 0x80000000, 0xffffffff} {
		for _, y := range []uint32{0, 1, 0x7fffffff, 0x80000000, 0xffffffff} {
			*(*uint32)(unsafe.Add(a, off)) = x
			*(*uint32)(unsafe.Add(b, off)) = y
			before := append([]byte(nil), data...)
			want := int32(0)
			if x < y {
				want = 1
			} else if x > y {
				want = -1
			}
			if g2bTestCompare(a, b) != want || !bytes.Equal(data, before) {
				t.Fatal("unsigned score comparator/C const adapter")
			}
		}
	}
	image, fi := alloc.MakeHigh([]byte{}, 32)
	defer fi()
	text, ft := alloc.MakeHigh([]uint16{}, 8)
	defer ft()
	ip, tp := unsafe.Pointer(&image[0]), unsafe.Pointer(&text[0])
	calls := 0
	game2BriefingTestHooks = &game2BriefingHooks{Image: func(name string) unsafe.Pointer {
		calls++
		if name != "WarriorChapterBegin8" {
			t.Fatal("image fallback")
		}
		return ip
	}}
	sideBytes := append([]byte(nil), memmap.Slice(0x5D4594, 832460)[:12]...)
	if g2bTestImage(ip) != ip || calls != 0 || g2bTestImage(nil) != ip || calls != 1 || *memmap.PtrPtr(0x5D4594, 832460) != ip {
		t.Fatal("native image side slot")
	}
	if g2bTestText(tp) != tp || *memmap.PtrPtr(0x5D4594, 832464) != tp {
		t.Fatal("UTF16 side slot")
	}
	for i := 0; i < 4; i++ {
		runtime.GC()
	}
	if *memmap.PtrPtr(0x5D4594, 832460) != ip || *memmap.PtrPtr(0x5D4594, 832464) != tp {
		t.Fatal("native C owners after GC")
	}
	if unsafe.Sizeof(uintptr(0)) == 8 && !bytes.Equal(sideBytes, memmap.Slice(0x5D4594, 832460)[:12]) {
		t.Fatal("native pointers overwritten packed image/text/stage DWORDs")
	}
	// Real native callback ABI accepts a high pointer while its unused arguments
	// remain opaque; no GetClient/render context is required for these gates.
	if g2bTestCall(11, 0, a) != 1 || g2bTestCall(12, 23, a) != 1 || g2bTestCall(12, 22, a) != 0 {
		t.Fatal("native window/draw ABI")
	}
}

func TestGame2BriefingNativeCleanupActualCSlots(t *testing.T) {
	g2bTestIsolate(t)
	slots := questPreviewSlots44E110()
	var saved [12]unsafe.Pointer
	for i, s := range slots {
		saved[i] = *(*unsafe.Pointer)(unsafe.Pointer(s))
	}
	defer func() {
		for i, s := range slots {
			*(*unsafe.Pointer)(unsafe.Pointer(s)) = saved[i]
		}
	}()
	objects, free := alloc.MakeHigh([]byte{}, 16*12)
	defer free()
	for i, s := range slots {
		*(*unsafe.Pointer)(unsafe.Pointer(s)) = unsafe.Pointer(&objects[16*i])
	}
	window, fw := alloc.MakeHigh([]byte{}, 32)
	defer fw()
	wp := unsafe.Pointer(&window[0])
	g2bTestPointer(0, wp, true)
	g2bTestPointer(1, wp, true)
	g2bTestPointer(2, wp, true)
	g2bTestWord(3, 0xffffffff, true)
	var log []string
	game2QuickbarTestHooks = &game2QuickbarHooks{Destroy: func(p unsafe.Pointer) int32 {
		if p != wp {
			t.Fatal("actual outer window owner")
		}
		log = append(log, "window")
		return -8
	}}
	order := []int{3, 0, 2, 1, 4, 5, 6, 7, 8, 9, 10, 11}
	game2BriefingTestHooks = &game2BriefingHooks{DeleteSprite: func(p unsafe.Pointer) int32 {
		i := len(log) - 1
		if i < 0 || i >= len(order) || p != unsafe.Pointer(&objects[16*order[i]]) {
			t.Fatal("actual C-slot lazy cleanup order")
		}
		log = append(log, fmt.Sprint(order[i]))
		return int32(700 + order[i])
	}}
	if g2bTestCall(10, 0, nil) != 711 || len(log) != 13 {
		t.Fatalf("last sprite delete return %v", log)
	}
	for _, s := range slots {
		if *(*unsafe.Pointer)(unsafe.Pointer(s)) != nil {
			t.Fatal("actual C sprite slot not cleared")
		}
	}
	for i := 0; i < 3; i++ {
		if g2bTestPointer(i, nil, false) != nil {
			t.Fatal("actual C root/image slot not cleared")
		}
	}
	if g2bTestWord(3, 0, false) != 0 {
		t.Fatal("briefing active flag")
	}
	log = nil
	game2BriefingTestHooks.DeleteSprite = func(unsafe.Pointer) int32 { t.Fatal("empty cleanup must skip deletion"); return 0 }
	if g2bTestCall(10, 0, nil) != 0 || len(log) != 0 {
		t.Fatal("empty cleanup")
	}
}
