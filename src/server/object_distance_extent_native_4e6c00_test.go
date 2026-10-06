package server

import (
	"bytes"
	"fmt"
	"math"
	"testing"
	"unsafe"

	"github.com/opennox/opennox/v1/legacy/common/alloc"
)

// FST DWORD at 004E6C54/004E6CA1 uses ToZero. For binary32 height/2,
// shifting the integer significand models the subnormal spill independently
// of the port's floating arithmetic and Nextafter helpers.
func objectDistanceHeightReference4E6C54(word uint32) float64 {
	sign, magnitude := word&0x80000000, word&0x7fffffff
	switch {
	case magnitude < 0x01000000:
		magnitude >>= 1
	case magnitude < 0x7f800000:
		magnitude -= 0x00800000
	}
	return float64(math.Float32frombits(sign | magnitude))
}

func TestObjectDistanceExtentNative4E6C00HeightChop(t *testing.T) {
	shape, free := alloc.New(Shape{})
	defer free()
	if unsafe.Sizeof(uintptr(0)) == 8 && uintptr(unsafe.Pointer(shape)) <= math.MaxUint32 {
		t.Fatalf("native shape below 4 GiB: %p", shape)
	}
	for _, word := range []uint32{0, 1, 3, 7, 0x007fffff, 0x00800001, 0x00ffffff, 0x01000001, 0x3f800001, 0x7f7fffff, 0x7f800000, 0x7fc12345,
		0x80000000, 0x80000001, 0x80000003, 0x80000007, 0x807fffff, 0x80800001, 0x80ffffff, 0x81000001, 0xbf800001, 0xff7fffff, 0xff800000, 0xffc12345} {
		t.Run(fmt.Sprintf("height-%08x", word), func(t *testing.T) {
			*shape = Shape{Kind: ShapeKindBox}
			shape.Box.W, shape.Box.H = float32(math.Inf(-1)), math.Float32frombits(word)
			before := bytes.Clone(unsafe.Slice((*byte)(unsafe.Pointer(shape)), int(unsafe.Sizeof(*shape))))
			got, want := objectDistanceShapeExtent4E6C00(shape), objectDistanceHeightReference4E6C54(word)
			if !(math.IsNaN(got) && math.IsNaN(want)) && math.Float64bits(got) != math.Float64bits(want) {
				t.Errorf("original height spill=%016x want=%016x", math.Float64bits(got), math.Float64bits(want))
			}
			if !bytes.Equal(before, unsafe.Slice((*byte)(unsafe.Pointer(shape)), len(before))) {
				t.Fatal("distance extent changed native shape record")
			}
		})
	}
	t.Run("ordered-width-versus-chopped-height", func(t *testing.T) {
		// 3 subnormal units / 2 spills as 1 unit. A 3-unit width retains
		// 1.5 units in the register and therefore wins the strict comparison.
		*shape = Shape{Kind: ShapeKindBox}
		shape.Box.W, shape.Box.H = math.Float32frombits(3), math.Float32frombits(3)
		want := float64(math.Float32frombits(3)) * 0.5
		if got := objectDistanceShapeExtent4E6C00(shape); got != want {
			t.Fatalf("width register was displaced by rounded-up height: %016x want=%016x", math.Float64bits(got), math.Float64bits(want))
		}
	})
	t.Run("height-reference-calibration", func(t *testing.T) {
		for _, tc := range [][2]uint32{{3, 1}, {7, 3}, {0x007fffff, 0x003fffff}, {0x00800001, 0x00400000}, {0x00ffffff, 0x007fffff},
			{0x01000001, 0x00800001}, {0x80000003, 0x80000001}, {0x80800001, 0x80400000}} {
			if got := float32(objectDistanceHeightReference4E6C54(tc[0])); math.Float32bits(got) != tc[1] {
				t.Fatalf("independent integer calibration=%08x want=%08x", math.Float32bits(got), tc[1])
			}
		}
	})
}
