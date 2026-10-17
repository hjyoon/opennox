package noximage_test

import (
	"image"
	"slices"
	"testing"

	"github.com/opennox/libs/noximage"
)

// Expected rows are literal visible pixels, independent of Row and PixOffset.
func TestImage16RowVisibleBounds(t *testing.T) {
	type row struct {
		y    int
		want []uint16
	}
	parent := &noximage.Image16{
		Rect: image.Rect(0, 0, 4, 3), Stride: 4,
		Pix: []uint16{10, 11, 12, 13, 20, 21, 22, 23, 30, 31, 32, 33},
	}
	roi := parent.SubImage(image.Rect(2, 1, 4, 3))
	cases := []struct {
		name string
		img  *noximage.Image16
		rows []row
	}{
		{"portrait", &noximage.Image16{Rect: image.Rect(0, 0, 3, 5), Stride: 3,
			Pix: []uint16{10, 11, 12, 20, 21, 22, 30, 31, 32, 40, 41, 42, 50, 51, 52}},
			[]row{{3, []uint16{40, 41, 42}}, {4, []uint16{50, 51, 52}}, {-1, nil}, {5, nil}}},
		{"landscape", &noximage.Image16{Rect: image.Rect(0, 0, 5, 3), Stride: 5,
			Pix: []uint16{10, 11, 12, 13, 14, 20, 21, 22, 23, 24, 30, 31, 32, 33, 34}},
			[]row{{2, []uint16{30, 31, 32, 33, 34}}, {-1, nil}, {3, nil}, {4, nil}}},
		{"nonzero_origin", &noximage.Image16{Rect: image.Rect(-2, 7, 1, 9), Stride: 3,
			Pix: []uint16{1, 2, 3, 4, 5, 6}},
			[]row{{7, []uint16{1, 2, 3}}, {8, []uint16{4, 5, 6}}, {6, nil}, {9, nil}}},
		{"padded_minimal_backing", &noximage.Image16{Rect: image.Rect(10, 20, 13, 23), Stride: 5,
			Pix: []uint16{11, 12, 13, 0xaaaa, 0xbbbb, 21, 22, 23, 0xcccc, 0xdddd, 31, 32, 33}},
			[]row{{20, []uint16{11, 12, 13}}, {21, []uint16{21, 22, 23}}, {22, []uint16{31, 32, 33}}, {19, nil}, {23, nil}}},
		{"bottom_right_subimage", roi,
			[]row{{1, []uint16{22, 23}}, {2, []uint16{32, 33}}, {0, nil}, {3, nil}}},
		{"nested_subimage", roi.SubImage(image.Rect(3, 2, 4, 3)),
			[]row{{2, []uint16{33}}, {1, nil}, {3, nil}}},
		{"empty_intersection", roi.SubImage(image.Rect(20, 20, 21, 21)), []row{{0, nil}}},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			for _, rc := range tc.rows {
				func() {
					defer func() {
						if v := recover(); v != nil {
							t.Errorf("Row(%d) panicked: %v", rc.y, v)
						}
					}()
					got := tc.img.Row(rc.y)
					if rc.want == nil {
						if got != nil {
							t.Errorf("Row(%d) = %v; want nil", rc.y, got)
						}
						return
					}
					if !slices.Equal(got, rc.want) {
						t.Errorf("Row(%d) = %v; want %v", rc.y, got, rc.want)
					}
					if len(got) != len(rc.want) || cap(got) != len(rc.want) {
						t.Errorf("Row(%d) len/cap = %d/%d; want %d/%d", rc.y, len(got), cap(got), len(rc.want), len(rc.want))
					}
				}()
			}
		})
	}
}

func TestImage16RowAliasesVisiblePixelsOnly(t *testing.T) {
	backing := []uint16{11, 12, 13, 0xaaaa, 0xbbbb, 21, 22, 23}
	img := &noximage.Image16{Rect: image.Rect(10, 20, 13, 22), Stride: 5, Pix: backing}
	got := img.Row(20)
	if len(got) != 3 || cap(got) != 3 {
		t.Fatalf("visible row len/cap = %d/%d; want 3/3", len(got), cap(got))
	}
	got[1] = 99
	if want := []uint16{11, 99, 13, 0xaaaa, 0xbbbb, 21, 22, 23}; !slices.Equal(backing, want) {
		t.Fatalf("row must alias its visible backing only: %v", backing)
	}
	appended := append(got, 0xffff)
	appended[0] = 77
	if want := []uint16{11, 99, 13, 0xaaaa, 0xbbbb, 21, 22, 23}; !slices.Equal(backing, want) {
		t.Fatalf("append exposed adjacent padding/backing: %v", backing)
	}
}
