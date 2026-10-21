package gui

import (
	"encoding/json"
	"fmt"
	"image"
	"image/color"
	"os"
	"testing"
	"unsafe"

	noxcolor "github.com/opennox/libs/color"
	"github.com/opennox/libs/noximage"

	"github.com/opennox/opennox/v1/legacy/common/alloc"
)

func TestScrollListBoxColorsMatchOriginal4A3AC0(t *testing.T) {
	blob, err := os.ReadFile("testdata/list_box_colors_4a3ac0.json")
	if err != nil {
		t.Fatal(err)
	}
	var oracle struct {
		Original string `json:"original_sha256"`
		Palette  []uint32
		Rows     []struct {
			Index          int
			Default, Color uint32
		}
		Faults    []json.RawMessage `json:"memory_faults"`
		Unchanged bool              `json:"executable_and_code_unchanged"`
	}
	if err := json.Unmarshal(blob, &oracle); err != nil {
		t.Fatal(err)
	}
	if oracle.Original != "0040e2c0683b4d73a5fb976e400d5087dca680df2b195c9e27f8edbda2d4974a" ||
		len(oracle.Palette) != 17 || len(oracle.Rows) != 44 || len(oracle.Faults) != 0 || !oracle.Unchanged {
		t.Fatal("wrong original list-box color transcript")
	}
	g := New(nil)
	t.Cleanup(g.DestroyAll)
	for _, row := range oracle.Rows {
		t.Run(fmt.Sprintf("index=%d/default=%08x", row.Index, row.Default), func(t *testing.T) {
			draw := WindowData{Style: StyleScrollListBox, TextColorVal: row.Default}
			win := NewScrollListBoxRaw(g, nil, StatusEnabled, 0, 0, 200, 60, &draw,
				&ScrollListBoxData{Count: 1, Line_height: 10})
			text := alloc.InternCString16("Auto Save")
			if EventRespInt(win.Func94(AsWindowEvent(0x400D, uintptr(unsafe.Pointer(text)), uintptr(uint32(int32(row.Index)))))) != 1 {
				t.Fatal("native add-line event failed")
			}
			item := scrollListBoxItems(scrollListBoxData(win))[0]
			if item.Field_129 != row.Color {
				t.Fatalf("row color=%08x, original=%08x", item.Field_129, row.Color)
			}
			if win.DrawData().TextColorVal != row.Default {
				t.Fatal("row insertion changed the window's default text color")
			}
			win.Destroy()
			g.FreeDestroyed()
		})
	}
}

func TestScrollListBoxDrawKeepsPerRowColors(t *testing.T) {
	r := listBoxTestRender(t)
	g := New(r)
	t.Cleanup(g.DestroyAll)
	for _, imageList := range []bool{false, true} {
		for _, selected := range []int{-1, 0, 1} {
			t.Run(fmt.Sprintf("image=%t/selected=%d", imageList, selected), func(t *testing.T) {
				flags := StatusEnabled
				if imageList {
					flags |= StatusImage
				}
				draw := WindowData{Style: StyleScrollListBox}
				draw.SetFont(r.Fonts.FontPtrByName("default"))
				draw.SetBackgroundColor(color.Transparent)
				draw.SetEnabledColor(color.Transparent)
				draw.SetSelectedColor(noxcolor.RGB5551Color(24, 24, 24))
				draw.SetTextColor(color.White)
				win := NewScrollListBoxRaw(g, nil, flags, 10, 20, 280, 90, &draw,
					&ScrollListBoxData{Count: 3, Line_height: 24})
				defer func() { win.Destroy(); g.FreeDestroyed() }()
				for _, row := range []struct {
					text  string
					index int
				}{{"AUTO SAVE", 13}, {"MANUAL SAVE", 14}, {"DEFAULT", -1}} {
					if !scrollListBoxAddLine(win, row.text, row.index) {
						t.Fatal("cannot add row")
					}
				}
				// A later style update must not recolor previously inserted rows.
				win.DrawData().SetTextColor(noxcolor.RGB5551Color(255, 0, 255))
				scrollListBoxSetSelection(win, selected)
				pix := noximage.NewImage16(r.Data().Rect3())
				r.SetPixBuffer(pix)
				win.Draw()
				// TTF edge coverage blends glyphs with the selected background.
				// Check each row's color family; the oracle test above separately
				// checks every exact packed palette value.
				for row := range 3 {
					count := 0
					bounds := image.Rect(15, 22+row*25, 285, 44+row*25)
					for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
						for x := bounds.Min.X; x < bounds.Max.X; x++ {
							cl := noxcolor.RGBA5551(pix.Pix[pix.PixOffset(x, y)]).ColorNRGBA()
							if row == 0 && cl.B > cl.G && cl.G > cl.R ||
								row == 1 && cl.R > cl.G && cl.G > cl.B ||
								row == 2 && cl.R == cl.G && cl.G == cl.B && cl.R > 32 {
								count++
							}
						}
					}
					if count == 0 {
						t.Errorf("row %d has no glyph pixels in its original color family", row)
					}
				}
			})
		}
	}
}
