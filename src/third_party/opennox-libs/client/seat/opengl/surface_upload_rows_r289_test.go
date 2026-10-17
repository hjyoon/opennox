//go:build darwin && arm64 && cgo && opengltest

package opengl

import (
	"fmt"
	"image"
	"io"
	"log/slog"
	"runtime"
	"slices"
	"testing"

	"github.com/go-gl/gl/v3.3-core/gl"
	"github.com/opennox/libs/noximage"
)

// These are literal endpoint colors, not a production color decoder or an
// assumption that the legal unsized GL_RGBA texture has eight-bit storage.
var uploadColorsR289 = [...]struct {
	word uint16
	rgba [4]byte
}{
	{0xfc00, [4]byte{255, 0, 0, 255}},
	{0x83e0, [4]byte{0, 255, 0, 255}},
	{0x801f, [4]byte{0, 0, 255, 255}},
	{0xffff, [4]byte{255, 255, 255, 255}},
	{0xffe0, [4]byte{255, 255, 0, 255}},
	{0x83ff, [4]byte{0, 255, 255, 255}},
	{0xfc1f, [4]byte{255, 0, 255, 255}},
	{0x8000, [4]byte{0, 0, 0, 255}},
	{0x7c00, [4]byte{255, 0, 0, 0}},
	{0x03e0, [4]byte{0, 255, 0, 0}},
	{0x001f, [4]byte{0, 0, 255, 0}},
	{0x7fff, [4]byte{255, 255, 255, 0}},
	{0x7fe0, [4]byte{255, 255, 0, 0}},
	{0x03ff, [4]byte{0, 255, 255, 0}},
	{0x7c1f, [4]byte{255, 0, 255, 0}},
	{0x0000, [4]byte{0, 0, 0, 0}},
}

func TestSurfaceUpdatePacked16RowsAndUnpackStateR289(t *testing.T) {
	// No t.Run: its separate goroutine would not inherit this thread's context.
	runtime.LockOSThread()
	defer runtime.UnlockOSThread()
	context, err := newUploadContextR289()
	if err != nil {
		t.Fatal(err)
	}
	defer func() {
		if err := context.close(); err != nil {
			t.Errorf("owned CGL cleanup: %v", err)
		}
	}()
	if context.initialUnpack != 4 {
		t.Fatalf("fresh context initial UNPACK_ALIGNMENT=%d, want observed default4", context.initialUnpack)
	}
	var backend Window
	if err := backend.Init(slog.New(slog.NewTextHandler(io.Discard, nil))); err != nil {
		t.Fatal(err)
	}
	defer func() {
		backend.Close()
		if code := gl.GetError(); code != gl.NO_ERROR {
			t.Errorf("actual Window.Close: GLError=%#x", code)
		}
	}()
	integer := func(key uint32) int32 {
		var value int32
		gl.GetIntegerv(key, &value)
		return value
	}
	checkGL := func(where string) {
		t.Helper()
		if code := gl.GetError(); code != gl.NO_ERROR {
			t.Fatalf("%s: GLError=%#x", where, code)
		}
	}
	checkGL("actual Window.Init")
	initialPack, initialUnpack := integer(gl.PACK_ALIGNMENT), integer(gl.UNPACK_ALIGNMENT)
	cases := 0
	for _, width := range []int{1, 2, 3, 4, 5, 6, 7, 8, 1024} {
		for _, incoming := range []int32{1, 2, 4, 8} {
			// Inline function scopes cleanup to each case, in this same goroutine.
			func() {
				label := fmt.Sprintf("width=%d incomingUNPACK=%d", width, incoming)
				const height = 2
				pixels := width * height
				backing := make([]uint16, pixels+8)
				for y := 0; y < height; y++ {
					for x := 0; x < width; x++ {
						backing[y*width+x] = uploadColorsR289[(x+3*y)%len(uploadColorsR289)].word
					}
				}
				copy(backing[pixels:], []uint16{0x1357, 0x2468, 0x369c, 0x4ace, 0x5bdf, 0x6abc, 0x7cde, 0x0123})
				before := slices.Clone(backing)
				// Eight extra words keep even the baseline's maximum padded row
				// read inside the allocation, without padding the logical image.
				maximumReadBytes := ((2*width+7)&^7)*(height-1) + 2*width
				if maximumReadBytes > 2*len(backing) {
					t.Fatalf("%s physical allocation does not contain alignment8 row reads", label)
				}
				img := &noximage.Image16{Rect: image.Rect(0, 0, width, height), Stride: width, Pix: backing[:pixels:pixels]}
				surface, ok := backend.NewSurface(image.Pt(width, height), false).(*Surface)
				if !ok {
					t.Fatal("actual Window.NewSurface did not return *opengl.Surface")
				}
				defer func() {
					surface.Destroy()
					if code := gl.GetError(); code != gl.NO_ERROR {
						t.Errorf("%s actual Surface.Destroy: GLError=%#x", label, code)
					}
				}()
				checkGL(label + " actual Window.NewSurface")
				packingKeys := []uint32{gl.UNPACK_ROW_LENGTH, gl.UNPACK_SKIP_ROWS, gl.UNPACK_SKIP_PIXELS, gl.UNPACK_SWAP_BYTES, gl.PIXEL_UNPACK_BUFFER_BINDING, gl.PACK_ROW_LENGTH, gl.PACK_SKIP_ROWS, gl.PACK_SKIP_PIXELS, gl.PACK_SWAP_BYTES, gl.PIXEL_PACK_BUFFER_BINDING}
				for _, key := range packingKeys {
					if got := integer(key); got != 0 {
						t.Fatalf("%s unexpected packing/PBO state %#x=%d", label, key, got)
					}
				}
				previousUnpack := integer(gl.UNPACK_ALIGNMENT)
				previousPack := integer(gl.PACK_ALIGNMENT)
				defer gl.PixelStorei(gl.UNPACK_ALIGNMENT, previousUnpack)
				gl.PixelStorei(gl.UNPACK_ALIGNMENT, incoming)
				surface.Update(img) // The sole production uploader, exactly once.
				checkGL(label + " actual Surface.Update")
				if got := integer(gl.UNPACK_ALIGNMENT); got != incoming {
					t.Errorf("%s uploader changed shared UNPACK_ALIGNMENT to %d", label, got)
				}
				if got := integer(gl.PACK_ALIGNMENT); got != previousPack {
					t.Errorf("%s uploader changed unrelated PACK_ALIGNMENT to %d", label, got)
				}
				for _, key := range packingKeys {
					if got := integer(key); got != 0 {
						t.Errorf("%s uploader changed unrelated packing/PBO state %#x=%d", label, key, got)
					}
				}
				if got := integer(gl.TEXTURE_BINDING_2D); got != int32(surface.tex) {
					t.Fatalf("%s actual upload texture=%d, want owned%d", label, got, surface.tex)
				}
				defer gl.PixelStorei(gl.PACK_ALIGNMENT, previousPack)
				// Readback hygiene only: this does not alter upload UNPACK state.
				gl.PixelStorei(gl.PACK_ALIGNMENT, 1)
				got := make([]byte, 4*pixels)
				gl.GetTexImage(gl.TEXTURE_2D, 0, gl.RGBA, gl.UNSIGNED_BYTE, gl.Ptr(got))
				checkGL(label + " actual texture RGBA8 readback")
				bad, first := 0, ""
				for y := 0; y < height; y++ {
					for x := 0; x < width; x++ {
						at := 4 * (y*width + x)
						actual := [4]byte{got[at], got[at+1], got[at+2], got[at+3]}
						want := uploadColorsR289[(x+3*y)%len(uploadColorsR289)].rgba
						if actual != want {
							bad++
							if first == "" {
								first = fmt.Sprintf("(%d,%d)=%v want%v", x, y, actual, want)
							}
						}
					}
				}
				if bad != 0 {
					t.Errorf("%s: %d/%d pixels differ; first %s", label, bad, pixels, first)
				}
				if !slices.Equal(backing, before) {
					t.Errorf("%s uploader/readback changed source pixels or eight canary words", label)
				}
				runtime.KeepAlive(img)
				runtime.KeepAlive(backing)
				cases++
			}()
		}
	}
	if cases != 36 || integer(gl.PACK_ALIGNMENT) != initialPack || integer(gl.UNPACK_ALIGNMENT) != initialUnpack {
		t.Fatalf("matrix/fixture state restoration: cases=%d PACK=%d/%d UNPACK=%d/%d", cases, integer(gl.PACK_ALIGNMENT), initialPack, integer(gl.UNPACK_ALIGNMENT), initialUnpack)
	}
	checkGL("matrix complete")
}
