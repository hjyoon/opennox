package gui

import (
	"image"
	"image/color"
	"log/slog"
	"os"
	"path/filepath"
	"testing"

	"github.com/opennox/libs/datapath"
	"github.com/opennox/libs/noximage"
	"golang.org/x/image/font/gofont/goregular"

	"github.com/opennox/opennox/v1/client/noxrender"
	"github.com/opennox/opennox/v1/legacy/common/alloc/handles"
)

func listBoxTestRender(t *testing.T) *noxrender.NoxRender {
	t.Helper()
	path := t.TempDir()
	for _, name := range []string{"default", "large", "number", "small"} {
		if err := os.WriteFile(filepath.Join(path, name+".ttf"), goregular.TTF, 0600); err != nil {
			t.Fatal(err)
		}
	}
	oldData := datapath.Data()
	datapath.SetData(path)
	t.Cleanup(func() { datapath.SetData(oldData) })
	handles.Init()
	t.Cleanup(handles.Release)
	r := noxrender.NewRender(slog.Default(), nil)
	if err := r.Fonts.Load(0); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(r.Fonts.Free)
	data, free := noxrender.NewRenderData()
	t.Cleanup(free)
	data.SetRect3(image.Rect(0, 0, 320, 160))
	data.SetClip(true)
	data.SetClipRect(image.Rect(1, 1, 319, 159))
	data.SetClipRect2(image.Rect(1, 1, 318, 158))
	r.SetData(data)
	return r
}

func TestScrollListBoxDrawClipsRows(t *testing.T) {
	r := listBoxTestRender(t)
	g := New(r)
	defer g.alloc.Free()
	for _, tc := range []struct {
		name      string
		image     bool
		title     string
		scrollbar bool
		partial   bool
		multi     bool
		disabled  bool
	}{
		{name: "plain bottom row"},
		{name: "image bottom row", image: true},
		{name: "title and partial rows", title: "Inventory", partial: true},
		{name: "image title and partial rows", image: true, title: "Inventory", partial: true},
		{name: "scrollbar and partial rows", scrollbar: true, partial: true},
		{name: "stock identify style", image: true, scrollbar: true},
		{name: "multi selection", title: "Inventory", scrollbar: true, partial: true, multi: true},
		{name: "disabled list", image: true, disabled: true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			flags := StatusEnabled
			if tc.image {
				flags |= StatusImage
			}
			draw := WindowData{Style: StyleScrollListBox}
			draw.SetBackgroundColor(color.Transparent)
			draw.SetEnabledColor(color.Transparent)
			draw.SetDisabledColor(color.Transparent)
			draw.SetSelectedColor(color.RGBA{R: 128, A: 255})
			draw.SetTextColor(color.White)
			draw.SetText(tc.title)
			draw.SetFont(r.Fonts.FontPtrByName("default"))
			opts := ScrollListBoxData{Count: 4, Line_height: uint16(r.FontHeight(r.Fonts.DefaultFont()))}
			if tc.scrollbar {
				opts.Field_3 = 1
			}
			if tc.multi {
				opts.Field_4 = 1
			}
			win := NewScrollListBoxRaw(g, nil, flags, 35, 20, 148, 46, &draw, &opts)
			if win == nil {
				t.Fatal("cannot create listbox")
			}
			defer func() { win.Destroy(); g.FreeDestroyed() }()
			if tc.disabled {
				win.Flags &^= StatusEnabled
			}
			for _, line := range []string{
				"Durability: 50/50",
				"An item description with enough words to wrap.\nA second line of detail.\nA third line below the viewport.",
				"Last line",
			} {
				if !scrollListBoxAddLine(win, line, -1) {
					t.Fatal("cannot add description line")
				}
			}
			d := scrollListBoxData(win)
			if tc.partial {
				d.Field_13_1 = 5
			}
			scrollListBoxSetSelection(win, 0)
			if tc.multi {
				scrollListBoxSetSelection(win, 1)
			}
			clip := image.Rectangle{Min: win.GlobalPos(), Max: win.GlobalPos().Add(win.Size())}
			if tc.title != "" {
				clip.Min.Y += r.FontHeight(win.DrawData().Font()) + 1
			}
			if tc.scrollbar {
				clip.Max.X -= 10
			}
			clip = clip.Intersect(r.Data().Rect3())
			beforeClip, beforeClip2, enabled := r.Data().ClipRect(), r.Data().ClipRect2(), r.Data().Clip()
			// The title/background belong outside the row content clip. Use
			// the same window without populated rows as their pixel baseline.
			baseline := noximage.NewImage16(r.Data().Rect3())
			r.SetPixBuffer(baseline)
			count := d.Field_11_0
			d.Field_11_0 = 0
			win.Draw()
			d.Field_11_0 = count
			pix := noximage.NewImage16(r.Data().Rect3())
			r.SetPixBuffer(pix)
			win.Draw()
			if r.Data().Clip() != enabled || r.Data().ClipRect() != beforeClip || r.Data().ClipRect2() != beforeClip2 {
				t.Fatal("listbox did not restore caller clip state")
			}
			changed := 0
			for y := pix.Rect.Min.Y; y < pix.Rect.Max.Y; y++ {
				for x := pix.Rect.Min.X; x < pix.Rect.Max.X; x++ {
					ind := pix.PixOffset(x, y)
					if pix.Pix[ind] == baseline.Pix[ind] {
						continue
					}
					changed++
					if !image.Pt(x, y).In(clip) {
						t.Fatalf("row pixel (%d,%d) escaped content bounds %v", x, y, clip)
					}
				}
			}
			if changed == 0 {
				t.Fatal("listbox rows disappeared")
			}
		})
	}
}
