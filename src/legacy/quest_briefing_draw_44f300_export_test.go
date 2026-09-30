package legacy

import (
	"bytes"
	"image"
	"math"
	"reflect"
	"strings"
	"testing"
	"unsafe"

	"github.com/opennox/opennox/v1/client"
	"github.com/opennox/opennox/v1/client/gui"
	"github.com/opennox/opennox/v1/client/noxrender"
	"github.com/opennox/opennox/v1/legacy/common/alloc"
)

func preserveQuestBriefingDraw44F300(t *testing.T) {
	t.Helper()
	oldCall, oldBlink := questBriefingDrawCall44F300, *questBriefingBlinkSlot44F300()
	t.Cleanup(func() { questBriefingDrawCall44F300, *questBriefingBlinkSlot44F300() = oldCall, oldBlink })
}

func TestQuestBriefingDraw44F300CEntryFullPointersAndReturn(t *testing.T) {
	preserveQuestBriefingDraw44F300(t)
	win, freeWin := alloc.New(gui.Window{})
	draw, freeDraw := alloc.New(gui.WindowData{})
	t.Cleanup(freeWin)
	t.Cleanup(freeDraw)
	for _, ptr := range []unsafe.Pointer{unsafe.Pointer(win), unsafe.Pointer(draw)} {
		if unsafe.Sizeof(uintptr(0)) > 4 && uintptr(ptr) <= math.MaxUint32 {
			t.Fatalf("pointer=%p, want above 4 GiB", ptr)
		}
	}
	var gotWin *gui.Window
	var gotDraw *gui.WindowData
	questBriefingDrawCall44F300 = func(w *gui.Window, d *gui.WindowData) int32 { gotWin, gotDraw = w, d; return math.MinInt32 + 1 }
	if got := questBriefingDrawCEntry44F300(win, draw); got != math.MinInt32+1 || gotWin != win || gotDraw != draw {
		t.Fatalf("return=%d win=%p/%p draw=%p/%p", got, gotWin, win, gotDraw, draw)
	}
	if got := questBriefingDrawCEntry44F300(nil, nil); got != math.MinInt32+1 || gotWin != nil || gotDraw != nil {
		t.Fatalf("nil delegation: return=%d win=%p draw=%p", got, gotWin, gotDraw)
	}
}

func TestQuestBriefingDraw44F300NativeCEntryFieldsAndCaches(t *testing.T) {
	slots := preserveQuestPreviewSlots44E110(t)
	preserveQuestBriefingDraw44F300(t)
	win, freeWin := alloc.New(gui.Window{})
	draw, freeDraw := alloc.New(gui.WindowData{})
	vp, freeVP := alloc.New(noxrender.Viewport{})
	normal, freeNormal := alloc.Malloc(8)
	title, freeTitle := alloc.Malloc(8)
	changedTitle, freeChanged := alloc.Malloc(8)
	for _, free := range []func(){freeWin, freeDraw, freeVP, freeNormal, freeTitle, freeChanged} {
		t.Cleanup(free)
	}
	draw.SetFont(title)
	*questPreviewFontSlot44E110() = normal
	vp.Screen.Min = image.Pt(11, -9)
	vp.World.Min = image.Pt(math.MaxInt32, math.MinInt32)
	var drs [12]*client.Drawable
	var before [12][]byte
	for i := range drs {
		dr, free := alloc.New(client.Drawable{})
		t.Cleanup(free)
		buf := unsafe.Slice((*byte)(unsafe.Pointer(dr)), int(unsafe.Sizeof(*dr)))
		for n := range buf {
			buf[n] = 0xa5
		}
		dr.PosVec = image.Pt(-20, 30)
		drs[i], *slots[i], before[i] = dr, dr, bytes.Clone(buf)
	}
	texts := make(map[string]*uint16)
	keys := make(map[*uint16]string)
	for _, key := range []string{"1", "2a", "2b", "3a", "3b", "4a", "4b", "7a", "7b", "5a", "5b", "6a", "6b", "8a", "8b", "9a", "9b", "10a", "10b", "11a", "11b", "12"} {
		p, free := alloc.New(uint16(0))
		t.Cleanup(free)
		texts[key], keys[p] = p, key
	}
	for _, p := range []unsafe.Pointer{unsafe.Pointer(win), unsafe.Pointer(draw), unsafe.Pointer(vp), normal, title, changedTitle,
		unsafe.Pointer(drs[0]), unsafe.Pointer(texts["12"])} {
		if unsafe.Sizeof(uintptr(0)) > 4 && uintptr(p) <= math.MaxUint32 {
			t.Fatalf("pointer=%p, want above 4 GiB", p)
		}
	}
	h := questBriefingDrawNativeHooks44F300(draw)
	h.viewport = func() *noxrender.Viewport { return vp }
	h.initPreviews = func() {}
	h.resetParticles = func() {}
	h.hideBook = func(v int32) int32 {
		if v != 1 {
			t.Fatalf("hide=%d", v)
		}
		return -1
	}
	h.prepare = func() {}
	h.dimensions = func() (int32, int32) { return 640, 480 }
	h.loadText = func(key, source string, line int32) *uint16 {
		key = strings.TrimPrefix(key, "GeneralPrint:QuestSplash")
		if source != questBriefingDrawSource44F300 || texts[key] == nil {
			t.Fatalf("key=%q source=%q line=%d", key, source, line)
		}
		return texts[key]
	}
	h.measure = func(font unsafe.Pointer, text *uint16) int32 {
		want := normal
		if keys[text] == "1" {
			want = title
		}
		if font != want {
			t.Fatalf("measure %q font=%p want=%p", keys[text], font, want)
		}
		if keys[text] == "1" {
			return 100
		}
		return 10
	}
	h.color = func(c questBriefingColor44F300) uint32 { return uint32(c) + 0x80000000 }
	var colors []uint32
	h.setColor = func(c uint32) { colors = append(colors, c) }
	titleCalls, normalCalls, promptCalls := 0, 0, 0
	h.drawText = func(font unsafe.Pointer, text *uint16, x, y int32) int32 {
		switch keys[text] {
		case "1":
			want := title
			if titleCalls > 0 {
				want = changedTitle
			}
			if font != want {
				t.Fatalf("title #%d font=%p want=%p", titleCalls, font, want)
			}
			titleCalls++
			draw.SetFont(changedTitle)
		case "12":
			promptCalls++
			if font != changedTitle || x != 315 || y != 450 {
				t.Fatalf("prompt font=%p coords=%d,%d", font, x, y)
			}
		default:
			normalCalls++
			if font != normal {
				t.Fatalf("normal %q font=%p want=%p", keys[text], font, normal)
			}
		}
		return math.MinInt32 + 1
	}
	wrapCalls := 0
	h.wrapText = func(font unsafe.Pointer, text *uint16, x, y, w, height int32) int32 {
		wrapCalls++
		if font != normal || height != 0 || keys[text] == "" {
			t.Fatalf("wrap font=%p text=%p height=%d", font, text, height)
		}
		return -1
	}
	var iconSlots []int
	h.drawDrawable = func(gotVP *noxrender.Viewport, dr *client.Drawable) {
		if gotVP != vp {
			t.Fatalf("viewport=%p want=%p", gotVP, vp)
		}
		for i, p := range drs {
			if dr == p {
				iconSlots = append(iconSlots, i)
				return
			}
		}
		t.Fatalf("unknown drawable=%p", dr)
	}
	h.frame = func() uint32 { return 1 }
	*questBriefingBlinkSlot44F300() = 1
	questBriefingDrawCall44F300 = func(gotWin *gui.Window, gotDraw *gui.WindowData) int32 {
		if gotWin != win || gotDraw != draw {
			t.Fatalf("entry win=%p draw=%p", gotWin, gotDraw)
		}
		return questBriefingDraw44F300(h)
	}
	if got := questBriefingDrawCEntry44F300(win, draw); got != math.MinInt32+1 {
		t.Fatalf("return=%d", got)
	}
	if want := []int{0, 1, 3, 2, 9, 11, 10, 6, 7, 5, 4, 8}; !reflect.DeepEqual(iconSlots, want) {
		t.Fatalf("icons=%v want=%v", iconSlots, want)
	}
	points := [12]image.Point{{73, 123}, {565, 117}, {525, 222}, {133, 192}, {219, 345}, {186, 333}, {484, 278}, {503, 303}, {220, 322}, {182, 262}, {185, 234}, {201, 251}}
	for i, dr := range drs {
		x, y := int32(points[i].X), int32(points[i].Y)
		want := image.Pt(int(x+int32(vp.World.Min.X)-int32(vp.Screen.Min.X)), int(y+int32(vp.World.Min.Y)-int32(vp.Screen.Min.Y)))
		if dr.PosVec != want || *slots[i] != dr {
			t.Fatalf("slot %d=%p pos=%v want=%v", i, *slots[i], dr.PosVec, want)
		}
		dr.PosVec = image.Pt(-20, 30)
		if !bytes.Equal(unsafe.Slice((*byte)(unsafe.Pointer(dr)), int(unsafe.Sizeof(*dr))), before[i]) {
			t.Fatalf("non-position bytes modified in drawable %d", i)
		}
	}
	if titleCalls != 5 || normalCalls != 16 || wrapCalls != 4 || promptCalls != 1 || len(colors) != 23 || *questBriefingBlinkSlot44F300() != 1 {
		t.Fatalf("title=%d normal=%d wrap=%d prompt=%d colors=%d blink=%d", titleCalls, normalCalls, wrapCalls, promptCalls, len(colors), *questBriefingBlinkSlot44F300())
	}
}

func TestQuestBriefingDraw44F300NativeNilHelperValuesAreNotSkipped(t *testing.T) {
	preserveQuestPreviewSlots44E110(t)
	preserveQuestBriefingDraw44F300(t)
	draw, free := alloc.New(gui.WindowData{})
	t.Cleanup(free)
	h := questBriefingDrawNativeHooks44F300(draw)
	h.viewport = func() *noxrender.Viewport { return nil }
	h.initPreviews, h.resetParticles, h.prepare = func() {}, func() {}, func() {}
	h.hideBook = func(int32) int32 { return -1 }
	h.dimensions = func() (int32, int32) { return 640, 480 }
	loads, measures, draws, wraps, positions, icons := 0, 0, 0, 0, 0, 0
	h.loadText = func(string, string, int32) *uint16 { loads++; return nil }
	h.measure = func(font unsafe.Pointer, text *uint16) int32 {
		measures++
		if font != nil || text != nil {
			t.Fatalf("measure font=%p text=%p", font, text)
		}
		return 10
	}
	h.setColor = func(uint32) {}
	h.drawText = func(font unsafe.Pointer, text *uint16, x, y int32) int32 {
		draws++
		if font != nil || text != nil {
			t.Fatalf("draw font=%p text=%p", font, text)
		}
		return math.MinInt32 + 1
	}
	h.wrapText = func(font unsafe.Pointer, text *uint16, x, y, w, height int32) int32 {
		wraps++
		if font != nil || text != nil || height != 0 {
			t.Fatalf("wrap font=%p text=%p height=%d", font, text, height)
		}
		return -1
	}
	h.position = func(vp *noxrender.Viewport, dr *client.Drawable, x, y int32) {
		positions++
		if vp != nil || dr != nil {
			t.Fatalf("map vp=%p dr=%p", vp, dr)
		}
	}
	h.drawDrawable = func(vp *noxrender.Viewport, dr *client.Drawable) {
		icons++
		if vp != nil || dr != nil {
			t.Fatalf("icon vp=%p dr=%p", vp, dr)
		}
	}
	h.frame = func() uint32 { return 30 }
	*questBriefingBlinkSlot44F300() = 0
	questBriefingDrawCall44F300 = func(_ *gui.Window, d *gui.WindowData) int32 {
		if d != draw {
			t.Fatalf("draw=%p want=%p", d, draw)
		}
		return questBriefingDraw44F300(h)
	}
	if got := questBriefingDrawCEntry44F300(nil, draw); got != math.MinInt32+1 {
		t.Fatalf("return=%d", got)
	}
	if loads != 22 || measures != 18 || draws != 22 || wraps != 4 || positions != 12 || icons != 12 || *questBriefingBlinkSlot44F300() != 1 {
		t.Fatalf("loads=%d measures=%d draws=%d wraps=%d positions=%d icons=%d blink=%d", loads, measures, draws, wraps, positions, icons, *questBriefingBlinkSlot44F300())
	}
}
