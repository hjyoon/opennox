package render

import (
	"image"
	"slices"
	"testing"

	"github.com/opennox/libs/noximage"

	"github.com/opennox/opennox/v1/client/seat/headless"
)

func TestCopyBufferPreservesVisibleRows(t *testing.T) {
	parent := &noximage.Image16{
		Rect: image.Rect(0, 0, 8, 4), Stride: 8,
		Pix: []uint16{
			100, 101, 102, 103, 104, 105, 106, 107,
			200, 201, 202, 203, 204, 205, 206, 207,
			300, 301, 302, 303, 304, 305, 306, 307,
			400, 401, 402, 403, 404, 405, 406, 407,
		},
	}
	cases := []struct {
		name    string
		src     *noximage.Image16
		backing []uint16
		screen  image.Point
		want    []uint16
	}{
		// This landscape ROI isolates flat-copy corruption from Row's
		// independent portrait-height bug. Padding is not visible content.
		{"parent_subimage", parent.SubImage(image.Rect(1, 1, 4, 3)), parent.Pix,
			image.Pt(6, 4), []uint16{201, 202, 203, 301, 302, 303}},
		{"padded_nonzero_origin", &noximage.Image16{
			Rect: image.Rect(-2, 7, 1, 9), Stride: 5,
			Pix: []uint16{11, 12, 13, 0xaaaa, 0xbbbb, 21, 22, 23}},
			nil, image.Pt(6, 4), []uint16{11, 12, 13, 21, 22, 23}},
		{"portrait", &noximage.Image16{
			Rect: image.Rect(0, 0, 2, 4), Stride: 2,
			Pix: []uint16{11, 12, 21, 22, 31, 32, 41, 42}},
			nil, image.Pt(4, 8), []uint16{11, 12, 21, 22, 31, 32, 41, 42}},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			sc := headless.New(tc.screen)
			defer sc.Close()
			r, err := New(sc)
			if err != nil {
				t.Fatal(err)
			}
			defer r.Close()
			backing := tc.backing
			if backing == nil {
				backing = tc.src.Pix
			}
			before := slices.Clone(backing)
			// Exercise first allocation, second buffer allocation, and reuse.
			for frame := 1; frame <= 3; frame++ {
				r.CopyBuffer(tc.src)
				got, view := sc.Snapshot()
				if got == nil {
					t.Fatal("missing uploaded frame")
				}
				if got.Rect != tc.src.Rect || got.Stride != tc.src.Rect.Dx() || !slices.Equal(got.Pix, tc.want) {
					t.Errorf("frame %d: layout/pixels = %v stride %d %v; want %v stride %d %v", frame, got.Rect, got.Stride, got.Pix, tc.src.Rect, tc.src.Rect.Dx(), tc.want)
				}
				if wantView := (image.Rectangle{Max: tc.screen}); view != wantView {
					t.Errorf("frame %d: viewport = %v; want %v", frame, view, wantView)
				}
				if r.Ticks() != uint(frame) || sc.PresentCount() != uint64(frame) {
					t.Errorf("frame %d: ticks/presents = %d/%d", frame, r.Ticks(), sc.PresentCount())
				}
				if !slices.Equal(backing, before) {
					t.Fatal("upload changed source or parent padding")
				}
				// Both the source and returned Snapshot remain independent of
				// the stored screen. Restore only our source before next upload.
				old := tc.src.Pix[0]
				tc.src.Pix[0] ^= 0xffff
				again, _ := sc.Snapshot()
				tc.src.Pix[0] = old
				if again == nil || !slices.Equal(again.Pix, tc.want) {
					t.Errorf("frame %d: stored screen aliases the source", frame)
				}
				got.Pix[0] ^= 0xffff
				again, _ = sc.Snapshot()
				if again == nil || !slices.Equal(again.Pix, tc.want) {
					t.Errorf("frame %d: stored screen aliases the returned snapshot", frame)
				}
			}
			// Changing the source rectangle after buffer reuse must replace
			// both the upload surface and its owned pixel layout.
			next := &noximage.Image16{Rect: image.Rect(8, 9, 10, 11), Stride: 3,
				Pix: []uint16{51, 52, 0xeeee, 61, 62}}
			r.CopyBuffer(next)
			got, view := sc.Snapshot()
			if got == nil || got.Rect != next.Rect || got.Stride != 2 || !slices.Equal(got.Pix, []uint16{51, 52, 61, 62}) {
				t.Fatal("changed rectangle did not replace the uploaded visible rows")
			}
			if want := image.Rect(1, 0, 5, 4); tc.screen == image.Pt(6, 4) && view != want {
				t.Errorf("changed-rectangle viewport = %v; want %v", view, want)
			}
			if want := image.Rect(0, 2, 4, 6); tc.screen == image.Pt(4, 8) && view != want {
				t.Errorf("changed-rectangle viewport = %v; want %v", view, want)
			}
			if r.Ticks() != 4 || sc.PresentCount() != 4 {
				t.Fatal("changed rectangle presented more than once")
			}
		})
	}
}
