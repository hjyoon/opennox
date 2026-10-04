//go:build !server

package legacy

import (
	"fmt"
	"image"
	"image/color"
	"math"
	"reflect"
	"strconv"
	"testing"
	"unsafe"

	noxcolor "github.com/opennox/libs/color"
	"golang.org/x/image/font"

	"github.com/opennox/opennox/v1/client"
	"github.com/opennox/opennox/v1/client/gui"
	"github.com/opennox/opennox/v1/client/noxrender"
)

type bottleSlotText471A80 struct {
	text      string
	pos       image.Point
	drawCalls int
}

type bottleSlotRender471A80 struct {
	Render2
	data        *noxrender.RenderData
	fonts       noxrender.RenderFonts
	height      int
	heightCalls int
	texts       []bottleSlotText471A80
}

func (r *bottleSlotRender471A80) Data() *noxrender.RenderData      { return r.data }
func (r *bottleSlotRender471A80) GetFonts() *noxrender.RenderFonts { return &r.fonts }
func (r *bottleSlotRender471A80) FontHeight(font.Face) int {
	r.heightCalls++
	return r.height
}
func (r *bottleSlotRender471A80) DrawString(_ font.Face, s string, pos image.Point) int {
	r.texts = append(r.texts, bottleSlotText471A80{s, pos, bottleSlotDrawCallCount471A80()})
	return -13
}

type bottleSlotClient471A80 struct {
	Client
	cli    *client.Client
	render *bottleSlotRender471A80
}

func (c *bottleSlotClient471A80) Cli() *client.Client { return c.cli }
func (c *bottleSlotClient471A80) R2() Render2         { return c.render }

func newBottleSlotFixture471A80(t *testing.T, slot int, size image.Point) (*gui.Window, *bottleSlotRender471A80, image.Point) {
	t.Helper()
	g := gui.New(nil)
	data, free := noxrender.NewRenderData()
	r := &bottleSlotRender471A80{data: data, height: 10}
	c := &bottleSlotClient471A80{cli: &client.Client{GUI: g}, render: r}
	oldClient := GetClient
	GetClient = func() Client { return c }
	t.Cleanup(func() {
		g.DestroyAll()
		GetClient = oldClient
		free()
	})
	// Stock 004714E0: cure/health/mana x offsets are 6/34/62 under
	// the 91x201 meter root. The C position service must include the parent.
	parent := g.NewWindowRaw(nil, gui.StatusEnabled, size.X-91, size.Y-201, 91, 201, nil)
	x := []int{34, 62, 6}[slot]
	win := g.NewWindowRaw(parent, gui.StatusEnabled, x, 166, 28, 30, nil)
	requireNativeWindowAddress(t, parent)
	requireNativeWindowAddress(t, win)
	return win, r, image.Pt(size.X-91+x, size.Y-35)
}

func checkBottleSlotRecords471A80(t *testing.T, out [30]int64) {
	t.Helper()
	for index, name := range map[int]string{
		21: "PE32 viewport and adjacent words", 22: "other native drawables",
		23: "selected drawable fields", 24: "selected type ID", 25: "binding",
		28: "other slot records", 29: "restored window widget data",
	} {
		if out[index] != 1 {
			t.Errorf("modified %s", name)
		}
	}
	if out[20] != 1 || out[26] != 0x40000000 {
		t.Errorf("slot result/GUI flags = %d/%#x, want 1/0x40000000", out[20], out[26])
	}
	if unsafe.Sizeof(uintptr(0)) > 4 && uint64(out[13]) <= math.MaxUint32 {
		t.Errorf("drawable address %#x did not exercise native pointer width", out[13])
	}
}

func checkBottleSlotViewport471A80(t *testing.T, out [30]int64, words [13]uint32) {
	t.Helper()
	for i, word := range words {
		want := int64(int32(word))
		if i == 10 || i == 11 {
			want = int64(word)
		}
		if out[i] != want {
			t.Errorf("native viewport field %d = %d, want %d", i, out[i], want)
		}
	}
}

func TestBottleSlotCViewport471A80(t *testing.T) {
	for slot, name := range []string{"health", "mana", "antidote"} {
		t.Run(name, func(t *testing.T) {
			win, r, pos := newBottleSlotFixture471A80(t, slot, image.Pt(1024, 768))
			for _, tc := range []struct {
				name  string
				words [13]uint32
			}{
				{"stock_1024x768", [13]uint32{0, 0, 1024, 768, 0, 0, 0, 0, 1024, 768}},
				{"stock_640x480", [13]uint32{0, 0, 640, 480, 0, 0, 0, 0, 640, 480}},
				{"all_PE32_words", [13]uint32{10, 20, 300, 400, 50, 60, 700, 800, 290, 380, 0xfedcba98, 0x89abcdef, 12}},
				{"signed_fields_unsigned_flags", [13]uint32{0x80000000, 0xffffffff, 0x7fffffff, 0xfffffffe, 0xfffffff0, 0x80000001, 0x80000002, 0x80000003, 0xffffffe0, 0xffffffd0, 0xffffffff, 0x80000000, 0xfffffff4}},
				{"refresh_after_nonzero_fields", [13]uint32{0, 0, 1280, 720, 0, 0, 0, 0, 1280, 720}},
			} {
				t.Run(tc.name, func(t *testing.T) {
					r.texts, r.heightCalls = nil, 0
					out := bottleSlotDrawC471A80(win, tc.words, slot, 3, true, false, -1, [4]uint16{'P', 'G', 'D'})
					checkBottleSlotRecords471A80(t, out)
					checkBottleSlotViewport471A80(t, out, tc.words)
					wantPos := pos.Add(image.Pt(14, 15))
					wantMappedX := int64(int32(tc.words[0])) + int64(wantPos.X) - int64(int32(tc.words[4]))
					wantMappedY := int64(int32(tc.words[1])) + int64(wantPos.Y) - int64(int32(tc.words[5]))
					if out[19] != 1 || out[27] != 3 || out[15] != int64(wantPos.X) || out[16] != int64(wantPos.Y) || out[17] != wantMappedX || out[18] != wantMappedY {
						t.Errorf("draw count/count/position/mapped = %d/%d/(%d,%d)/(%d,%d), want 1/3/%v/(%d,%d)", out[19], out[27], out[15], out[16], out[17], out[18], wantPos, wantMappedX, wantMappedY)
					}
					wantText := []bottleSlotText471A80{{"3", pos.Add(image.Pt(-2, 0)), 1}, {"PGD", pos.Add(image.Pt(-2, 23)), 1}}
					if !reflect.DeepEqual(r.texts, wantText) || r.heightCalls != 2 {
						t.Errorf("text/height calls = %+v/%d, want %+v/2", r.texts, r.heightCalls, wantText)
					}
				})
			}
		})
	}
}

func TestBottleSlotCCountAndBinding471A80(t *testing.T) {
	for slot, name := range []string{"health", "mana", "antidote"} {
		for _, tc := range []struct {
			name     string
			count    uint16
			callback bool
			newCount int
			height   int
			binding  [4]uint16
			text     string
		}{
			{"empty", 0, true, -1, 10, [4]uint16{'P', 'G', 'D'}, "PGD"},
			{"missing_draw_callback", 7, false, -1, 10, [4]uint16{'P', 'G', 'D'}, "PGD"},
			{"maximum_unsigned_count", 65535, true, -1, 7, [4]uint16{'P', 'G', 'D'}, "PGD"},
			{"live_count_after_callback", 3, true, 1, 13, [4]uint16{'해', '독'}, "해독"},
			{"zero_count_after_callback_empty_binding", 3, true, 0, 16, [4]uint16{}, ""},
		} {
			t.Run(fmt.Sprintf("%s/%s", name, tc.name), func(t *testing.T) {
				win, r, pos := newBottleSlotFixture471A80(t, slot, image.Pt(640, 480))
				r.height = tc.height
				r.data.SetTextColor(color.RGBA{R: 4, G: 8, B: 12, A: 255})
				oldColor := r.data.TextColor()
				words := [13]uint32{0, 0, 640, 480, 0, 0, 0, 0, 640, 480}
				out := bottleSlotDrawC471A80(win, words, slot, tc.count, tc.callback, false, tc.newCount, tc.binding)
				checkBottleSlotRecords471A80(t, out)
				wantCalls := 0
				if tc.count != 0 && tc.callback {
					wantCalls = 1
					checkBottleSlotViewport471A80(t, out, words)
				}
				wantCount := int(tc.count)
				if wantCalls != 0 && tc.newCount >= 0 {
					wantCount = tc.newCount
				}
				wantPos := image.Pt(-111, -222)
				if wantCalls != 0 {
					wantPos = pos.Add(image.Pt(14, 15))
				}
				if out[19] != int64(wantCalls) || out[27] != int64(wantCount) || out[15] != int64(wantPos.X) || out[16] != int64(wantPos.Y) {
					t.Errorf("draw/count/position = %d/%d/(%d,%d), want %d/%d/%v", out[19], out[27], out[15], out[16], wantCalls, wantCount, wantPos)
				}
				var wantText []bottleSlotText471A80
				if tc.count != 0 {
					wantText = append(wantText, bottleSlotText471A80{strconv.Itoa(wantCount), pos.Add(image.Pt(-2, 10-tc.height)), wantCalls})
				}
				wantText = append(wantText, bottleSlotText471A80{tc.text, pos.Add(image.Pt(-2, 33-tc.height)), wantCalls})
				if !reflect.DeepEqual(r.texts, wantText) || r.heightCalls != len(wantText) {
					t.Errorf("text/height calls = %+v/%d, want %+v/%d", r.texts, r.heightCalls, wantText, len(wantText))
				}
				wantColor := oldColor
				if tc.count != 0 {
					wantColor = noxcolor.RGBA5551(0xffff)
				}
				if got, want := r.data.TextColor(), wantColor; !reflect.DeepEqual(got, want) {
					gr, gg, gb, ga := got.RGBA()
					wr, wg, wb, wa := want.RGBA()
					if gr != wr || gg != wg || gb != wb || ga != wa {
						t.Errorf("text color = %v, want %v", got, want)
					}
				}
			})
		}
	}
}

func TestBottleSlotCBorrowedViewportIsolation471A80(t *testing.T) {
	for slot, name := range []string{"health", "mana", "antidote"} {
		t.Run(name, func(t *testing.T) {
			win, _, _ := newBottleSlotFixture471A80(t, slot, image.Pt(1024, 768))
			words := [13]uint32{0, 0, 1024, 768, 0, 0, 0, 0, 1024, 768}
			out := bottleSlotDrawC471A80(win, words, slot, 1, true, true, -1, [4]uint16{'P', 'G', 'D'})
			checkBottleSlotRecords471A80(t, out)
			checkBottleSlotViewport471A80(t, out, words)
			if out[19] != 1 {
				t.Errorf("draw callback count = %d, want 1", out[19])
			}
		})
	}
}
