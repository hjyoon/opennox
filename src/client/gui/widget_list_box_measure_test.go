package gui

import (
	"strings"
	"testing"
)

func TestScrollListBoxMeasuresScrollbarWidth(t *testing.T) {
	r := listBoxTestRender(t)
	g := New(r)
	defer g.alloc.Free()
	font := r.Fonts.DefaultFont()
	// Choose real glyphs whose wrapping distinguishes the stock identify
	// text widths (141 without a scrollbar, 131 with one).
	text := ""
	for n := 1; n < 60; n++ {
		s := strings.TrimSpace(strings.Repeat("item ", n))
		if r.GetStringSizeWrapped(font, s, 141).Y != r.GetStringSizeWrapped(font, s, 131).Y {
			text = s
			break
		}
	}
	if text == "" {
		t.Fatal("cannot distinguish the two text widths")
	}
	for _, scrollbar := range []bool{false, true} {
		for _, oneLine := range []bool{false, true} {
			flags := StatusEnabled
			if oneLine {
				flags |= StatusOneLine
			}
			draw := WindowData{Style: StyleScrollListBox}
			draw.SetFont(r.Fonts.FontPtrByName("default"))
			opts := ScrollListBoxData{Count: 3, Line_height: uint16(r.FontHeight(font))}
			wrapWidth := 141
			if scrollbar {
				opts.Field_3 = 1
				wrapWidth -= 10
			}
			win := NewScrollListBoxRaw(g, nil, flags, 0, 0, 148, 140, &draw, &opts)
			if win == nil {
				t.Fatal("cannot create listbox")
			}
			want := r.FontHeight(font)
			if !oneLine {
				want = max(want, r.GetStringSizeWrapped(font, text, wrapWidth).Y)
			}
			for i := 0; i < 3; i++ {
				if !scrollListBoxAddLine(win, text, -1) {
					t.Fatal("cannot add line")
				}
				d := scrollListBoxData(win)
				item := scrollListBoxItems(d)[i]
				if item.Field_130 != uint32(want) || item.Field_0 != uint32((i+1)*(want+1)) || d.Field_10 != item.Field_0 {
					t.Errorf("scrollbar=%t oneLine=%t row=%d height=%d bottom=%d total=%d; want height=%d bottom=%d for width=%d", scrollbar, oneLine, i, item.Field_130, item.Field_0, d.Field_10, want, (i+1)*(want+1), wrapWidth)
				}
			}
			win.Destroy()
			g.FreeDestroyed()
		}
	}
}
