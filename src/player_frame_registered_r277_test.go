//go:build !server

package opennox

import (
	"image"
	"slices"
	"testing"
	"unsafe"

	"github.com/opennox/libs/noximage"
	"github.com/opennox/libs/strman"

	"github.com/opennox/opennox/v1/client"
	"github.com/opennox/opennox/v1/client/noxrender"
	"github.com/opennox/opennox/v1/client/render"
	"github.com/opennox/opennox/v1/client/seat/headless"
	"github.com/opennox/opennox/v1/common/memmap"
	"github.com/opennox/opennox/v1/legacy"
	"github.com/opennox/opennox/v1/legacy/common/alloc"
	"github.com/opennox/opennox/v1/legacy/common/alloc/handles"
	"github.com/opennox/opennox/v1/server"
)

// Exercise the registered vector callback once per case, using the original
// frame helper's signed quotient contract. No production selector predicts the
// callback. The bag is a legal independent colored fixture, not stock player
// artwork or a physical-GPU/Host Load opacity oracle.
func TestPlayerFrameRegisteredVectorR277(t *testing.T) {
	t.Chdir(t.TempDir())
	destroyOnceBag48EA70(t)
	handles.Init()
	t.Cleanup(handles.Release)
	s := server.New(nil, nil, strman.New())
	t.Cleanup(s.Close)
	cl := client.NewClient(nil, nil, s)
	t.Cleanup(cl.Close)
	t.Cleanup(func() { alloc.FreePtr(cl.Viewport().C()) })
	cl.Objs.Init(4)
	t.Cleanup(cl.Objs.Free)
	r := cl.Render()
	t.Cleanup(r.Part.Free)
	draw, freeDraw := noxrender.NewRenderData()
	t.Cleanup(freeDraw)
	r.SetData(draw)
	pix := noximage.NewImage16(image.Rect(0, 0, 64, 64))
	r.SetPixBuffer(pix)
	if err := r.GetBag().ReadVideoBag(); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(r.GetBag().Free)
	c := &Client{Client: cl, srv: &Server{Server: s}, r: NewNoxRender(r)}
	c.tiles.lightsOutBuf, _ = alloc.Make([]uint32{}, 6)
	t.Cleanup(func() { alloc.FreeSlice(c.tiles.lightsOutBuf) })
	for i := 0; i < 3; i++ {
		c.tiles.lightsOutBuf[i] = 255
	}
	for x := range c.tiles.nox_arr2_853BC0 {
		for y := range c.tiles.nox_arr2_853BC0[x] {
			c.tiles.nox_arr2_853BC0[x][y] = noxrender.RGB{R: 255 << 16, G: 255 << 16, B: 255 << 16}
		}
	}
	oldClient, oldServer := noxClient, noxServer
	noxClient, noxServer = c, c.srv
	t.Cleanup(func() { noxClient, noxServer = oldClient, oldServer })
	connected := memmap.PtrUint32(0x5D4594, 815764)
	previousConnected := *connected
	*connected = 0
	t.Cleanup(func() { *connected = previousConnected })
	previousFade := legacy.Get_nox_client_fadeObjects_80836()
	legacy.Set_nox_client_fadeObjects_80836(0)
	t.Cleanup(func() { legacy.Set_nox_client_fadeObjects_80836(previousFade) })
	vp := cl.Viewport()
	vp.Screen, vp.World, vp.Size = pix.Rect, pix.Rect, pix.Size()
	sc := headless.New(pix.Size())
	t.Cleanup(func() { _ = sc.Close() })
	output, err := render.New(sc)
	if err != nil {
		t.Fatal(err)
	}

	ani, freeAni := alloc.New(client.AnimationVector{})
	t.Cleanup(freeAni)
	frames, freeFrames := alloc.Make([]noxrender.ImageHandle{}, 3)
	t.Cleanup(freeFrames)
	for i := range frames {
		frames[i] = r.GetBag().ImageByIndex(i).C()
	}
	for _, p := range []unsafe.Pointer{unsafe.Pointer(ani), unsafe.Pointer(&frames[0]), vp.C(), draw.C()} {
		if unsafe.Sizeof(uintptr(0)) == 8 && uintptr(p) <= 0xffffffff {
			t.Fatal("registered frame fixture requires native records above 4 GiB")
		}
	}
	for _, handle := range frames {
		if unsafe.Sizeof(uintptr(0)) == 8 && uintptr(handle) <= 0xffffffff {
			t.Fatal("registered frame fixture requires native image handles above 4 GiB")
		}
	}

	// These literal frame/color expectations do not call getAnimFrameInd,
	// sub_4B8FA0, or any independent copy of the production selector.
	tests := []struct {
		name      string
		kind      client.AnimKind
		count     uint16
		frame     uint32
		netCode   uint32
		wantFrame int // -1 means no image/raster call is permitted.
		wantKeep  bool
	}{
		{"oneshot-red", client.AnimOneShot, 3, 0, 0, 0, true},
		{"oneshot-green", client.AnimOneShot, 3, 1, 0, 1, true},
		{"oneshot-blue", client.AnimOneShot, 3, 2, 0, 2, true},
		{"oneshot-clamped-blue", client.AnimOneShot, 3, 3, 0, 2, true},
		{"loop-red", client.AnimLoop, 3, 0, 0, 0, true},
		{"loop-green", client.AnimLoop, 3, 1, 0, 1, true},
		{"loop-blue", client.AnimLoop, 3, 2, 0, 2, true},
		{"loop-wrapped-red", client.AnimLoop, 3, 3, 0, 0, true},
		{"loop-DWORD-add-wrap", client.AnimLoop, 3, 0xffffffff, 1, 0, true},
		{"oneshot-negative-int32-min", client.AnimOneShot, 3, 0x80000000, 0, -1, true},
		{"oneshot-negative-minus-one", client.AnimOneShot, 3, 0xffffffff, 0, -1, true},
		{"loop-negative-int32-min", client.AnimLoop, 3, 0x80000000, 0, -1, true},
		{"loop-negative-minus-one", client.AnimLoop, 3, 0xffffffff, 0, -1, true},
		{"remove-negative-int32-min-retains-pool", client.AnimOneShotRemove, 3, 0x80000000, 0, -1, true},
		{"remove-negative-minus-one-retains-pool", client.AnimOneShotRemove, 3, 0xffffffff, 0, -1, true},
		{"remove-expiry-control", client.AnimOneShotRemove, 3, 3, 0, -1, false},
		{"oneshot-zero-count-negative-control", client.AnimOneShot, 0, 0, 0, -1, true},
		{"kind-three-default-zero-control", client.AnimLoopAndFade, 3, 0x80000000, 0, 0, true},
	}
	for i, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			*ani = client.AnimationVector{Size: uint32(unsafe.Sizeof(*ani)), Cnt40: tc.count, Kind: tc.kind}
			for _, dir := range []int{0, 1, 2, 3, 5, 6, 7, 8} {
				ani.Frames[dir] = &frames[0]
			}
			dr := cl.Objs.New()
			if dr == nil {
				t.Fatal("real drawable pool exhausted")
			}
			cl.Objs.Count, cl.Objs.List1 = 1, dr
			defer func() {
				// A failed baseline Remove comparison may already have freed dr.
				// Compare only the opaque pointer; never inspect a freed record.
				if cl.Objs.List1 == dr {
					c.Nox_xxx_spriteDeleteStatic_45A4E0_drawable(dr)
				}
			}()
			dr.ObjFlags = 0x01000004
			dr.NetCode32, dr.PosVec = tc.netCode, image.Pt(24, 24)
			dr.DrawData = unsafe.Pointer(ani)
			dr.DrawFuncPtr = legacy.Get_nox_thing_vector_animate_draw()
			if unsafe.Sizeof(uintptr(0)) == 8 && uintptr(dr.C()) <= 0xffffffff {
				t.Fatal("real drawable pool returned a non-native fixture address")
			}
			for j := range pix.Pix {
				pix.Pix[j] = 0x4210
			}
			*draw = noxrender.RenderData{}
			draw.SetClip(true)
			draw.SetClipRect(pix.Rect)
			draw.SetClipRect2(pix.Rect)
			s.SetFrame(tc.frame)
			beforeAni, beforeDR := *ani, *dr
			beforeFrames := slices.Clone(frames)
			beforeVP := *vp
			logic, other := s.Rand.Logic.Index(), s.Rand.Other.Index()

			// One real registered C bridge call, with no preflight selector or
			// second draw. Negative OneShot/Loop results must suppress rendering.
			got := dr.CallDraw(vp)
			wantCount := 0
			if tc.wantKeep {
				wantCount = 1
			}
			// FIRST post-callback check: baseline high-bit Remove incorrectly
			// deletes. Fatal here before any post-call dr dereference.
			if cl.Objs.Count != wantCount || (tc.wantKeep && cl.Objs.List1 != dr) || (!tc.wantKeep && cl.Objs.List1 != nil) {
				t.Fatalf("registered callback pool: count=%d list=%p; want keep=%t count=%d", cl.Objs.Count, cl.Objs.List1, tc.wantKeep, wantCount)
			}
			wantReturn := 0
			if tc.wantFrame >= 0 {
				wantReturn = 1
			}
			if got != wantReturn {
				t.Fatalf("registered callback return=%d, want %d without reselecting its frame", got, wantReturn)
			}
			if *ani != beforeAni || !slices.Equal(frames, beforeFrames) || *vp != beforeVP || s.Frame() != tc.frame || s.Rand.Logic.Index() != logic || s.Rand.Other.Index() != other {
				t.Fatal("registered callback changed shared vector/frame-array/viewport/frame/RNG state")
			}
			if tc.wantKeep && tc.wantFrame < 0 && *dr != beforeDR {
				t.Fatal("negative-frame early return changed its retained drawable")
			}
			want := make([]uint16, len(pix.Pix))
			for j := range want {
				want[j] = 0x4210
			}
			if tc.wantFrame >= 0 {
				color := [...]uint16{0x7c00, 0x03e0, 0x001f}[tc.wantFrame]
				for y := 24; y < 26; y++ {
					for x := 24; x < 26; x++ {
						want[y*64+x] = color
					}
				}
			}
			if !slices.Equal(pix.Pix, want) {
				for j, word := range want {
					if pix.Pix[j] != word {
						t.Fatalf("registered frame raster (%d,%d)=%04x, want=%04x selected=%d", j%64, j/64, pix.Pix[j], word, tc.wantFrame)
					}
				}
			}
			output.CopyBuffer(pix)
			shown, view := sc.Snapshot()
			if shown == nil || view != pix.Rect || !slices.Equal(shown.Pix, want) {
				t.Fatalf("headless copy did not preserve registered callback frame %d", i)
			}
		})
	}
}
