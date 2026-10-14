//go:build !server

package opennox

import (
	"bytes"
	"encoding/binary"
	"encoding/json"
	"image"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"slices"
	"strconv"
	"testing"
	"unicode/utf16"
	"unsafe"

	noxcolor "github.com/opennox/libs/color"
	"github.com/opennox/libs/datapath"
	"github.com/opennox/libs/noximage"
	"github.com/opennox/opennox/v1/client/noxrender"
	"github.com/opennox/opennox/v1/client/render"
	"github.com/opennox/opennox/v1/client/seat/headless"
	"github.com/opennox/opennox/v1/common/memmap"
	"github.com/opennox/opennox/v1/legacy"
	"github.com/opennox/opennox/v1/legacy/common/alloc/handles"
	"golang.org/x/image/font/gofont/goregular"
)

func TestCenterMessagesActualHeadlessNativeResetAndHUD(t *testing.T) {
	const key = "OPENNOX_TEST_CENTER_MESSAGES_R202"
	if raw := os.Getenv(key); raw != "" {
		lang, err := strconv.Atoi(raw)
		if err != nil {
			t.Fatal(err)
		}
		centerMessagesActualChild202(t, lang)
		return
	}
	for _, lang := range []int{1, 8} {
		cmd := exec.Command(os.Args[0], "-test.run=^"+t.Name()+"$", "-test.count=1", "-test.v")
		cmd.Env = append(os.Environ(), key+"="+strconv.Itoa(lang))
		out, err := cmd.CombinedOutput()
		if err != nil {
			t.Fatalf("actual centered HUD/lang%d: %v\n%s", lang, err, out)
		}
		if !bytes.Contains(out, []byte("CENTER HUD ACTUAL PASS")) {
			t.Fatalf("missing actual completion: %s", out)
		}
		t.Log(string(out))
	}
}

func centerMessagesActualChild202(t *testing.T, lang int) {
	legacy.InitBlobData()
	t.Chdir(t.TempDir())
	for _, name := range []string{"default", "large", "number", "small"} {
		if err := os.WriteFile(name+".ttf", goregular.TTF, 0600); err != nil {
			t.Fatal(err)
		}
	}
	oldData := datapath.Data()
	datapath.SetData(".")
	t.Cleanup(func() { datapath.SetData(oldData) })
	handles.Init()
	t.Cleanup(handles.Release)
	c, _, d := actualVideoRender200(t)
	if err := c.r.GetFonts().Load(lang); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(c.r.GetFonts().Free)
	font := c.r.GetFonts().DefaultFont()
	if font == nil {
		t.Fatal("real default font missing")
	}
	pix := noximage.NewImage16(image.Rect(0, 0, 640, 480))
	c.r.SetPixBuffer(pix)
	seat := headless.New(pix.Size())
	t.Cleanup(func() { _ = seat.Close() })
	window, err := render.New(seat)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(window.Close)
	c.Win = window
	oldRender := *legacy.VideoModeRenderDataC
	legacy.Set_nox_draw_curDrawData_3799572(d)
	t.Cleanup(func() { *legacy.VideoModeRenderDataC = oldRender })
	legacy.Set_nox_win_width(640)
	legacy.Set_nox_win_height(480)
	legacy.Set_nox_color_black_2650656(0)
	*legacy.Get_nox_color_white_2523948_ptr() = 0x7fff
	*memmap.PtrUint32(0x5d4594, 2597996) = 0x03e0
	legacy.Set_nox_client_gui_flag_1556112(0)
	*memmap.PtrUint32(0x5d4594, 1096424) = 0
	c.Viewport().Screen = pix.Rect
	c.Viewport().Size = pix.Size()
	var original struct {
		Offsets [8]int32 `json:"shadow_offsets"`
	}
	data, err := os.ReadFile(filepathCenterOriginal202(t))
	if err != nil {
		t.Fatal(err)
	}
	if err := json.Unmarshal(data, &original); err != nil {
		t.Fatal(err)
	}
	for i, want := range original.Offsets {
		if memmap.Int32(0x587000, 107848+uintptr(4*i)) != want {
			t.Fatal("actual original shadow offsets differ")
		}
	}
	center := memmap.Slice(0x5d4594, 823804)[:1932]
	whole := memmap.Slice(0x5d4594, 823804-32)[:1932+64]
	if unsafe.Sizeof(uintptr(0)) > 4 && uintptr(unsafe.Pointer(&center[0])) <= 0xffffffff {
		t.Fatal("high-address real message buffer required")
	}
	texts := []string{"First notice", "Second notice", "Third notice"}
	if lang == 8 {
		texts = []string{"첫 알림 ⚔", "다음 알림 ★", "마지막 알림 ☀"}
	}
	writeText := func(dst []byte, text string) {
		for i, word := range append(utf16.Encode([]rune(text)), 0) {
			binary.LittleEndian.PutUint16(dst[2*i:], word)
		}
	}
	resetRaster := func() {
		*d = noxrender.RenderData{}
		d.SetRect3(pix.Rect)
		d.SetClip(true)
		d.SetClipRect(pix.Rect)
		d.SetClipRect2(image.Rect(0, 0, 639, 479))
		d.SetSelectColor(0x4210)
		c.r.ClearScreen(d.BgColor())
	}
	startPresentation := memmap.Uint32(0x973a20, 496)
	frames, resetCases := 0, 0
	for _, frame := range []uint32{0, 37, 0x80000000, 0xffffffff} {
		c.Server.SetFrame(frame)
		for index := uint32(0); index < 3; index++ {
			for mask := uint32(0); mask < 8; mask++ {
				for _, hud := range []bool{false, true} {
					for i := range whole {
						whole[i] = 0xa5
					}
					for row := 0; row < 3; row++ {
						slot := center[row*644 : (row+1)*644]
						writeText(slot, texts[row])
						expiry := frame - 1
						if mask&(1<<row) != 0 {
							expiry = frame
						}
						binary.LittleEndian.PutUint32(slot[636:], expiry)
					}
					legacy.CenterMessageSetIndexC445530(index)
					before := bytes.Clone(whole)
					resetRaster()
					y := 3*c.Viewport().Size.Y/4 + c.Viewport().Screen.Min.Y - 15
					slotIndex := index
					wantReturned := frame
					for order := 0; order < 3; order++ {
						slot := center[int(slotIndex)*644:]
						if binary.LittleEndian.Uint32(slot[636:]) < frame {
							wantReturned = frame
							break
						}
						text := texts[slotIndex]
						width := c.r.GetStringSizeWrapped(font, text, 0).X
						x := (640 - width) / 2
						d.SetTextColor(noxcolor.RGBA5551(0))
						for shadow := 0; shadow < 4; shadow++ {
							c.r.DrawString(font, text, image.Pt(x+int(original.Offsets[2*shadow]), y+int(original.Offsets[2*shadow+1])))
						}
						color := noxcolor.RGBA5551(0x7fff)
						if order != 0 {
							color = noxcolor.RGBA5551(0x03e0)
						}
						d.SetTextColor(color)
						c.r.DrawString(font, text, image.Pt(x, y))
						y -= 4 + c.r.FontHeight(font)
						wantReturned = slotIndex
						slotIndex = (slotIndex + 2) % 3
					}
					want := slices.Clone(pix.Pix)
					resetRaster()
					if hud {
						legacy.Nox_xxx_drawMinimapAndLines_4738E0()
					} else if got := uint32(legacy.CenterMessageDrawC445530()); got != wantReturned {
						t.Fatalf("original live frame/ring return=%x want=%x", got, wantReturned)
					}
					presentParticleRegistration4B01B0(t, pix, window, seat, want, frames)
					if !bytes.Equal(whole, before) || legacy.CenterMessageIndexC445530() != index || c.Server.Frame() != frame {
						t.Fatal("HUD changed ring/packed bytes/index/game clock")
					}
					frames++
				}
			}
		}
	}
	c.Server.SetFrame(37)
	for index := uint32(0); index < 3; index++ {
		for _, source := range []string{"", "A", "가🚀B"} {
			for i := range whole {
				whole[i] = 0xa5
			}
			writeText(whole[32+1936:], source)
			legacy.CenterMessageSetIndexC445530(index)
			want := bytes.Clone(whole)
			for row := 0; row < 3; row++ {
				slot := want[32+row*644:]
				writeText(slot, source)
				clear(slot[636:641])
			}
			if got := legacy.CenterMessageResetC445450(); got != unsafe.Pointer(&center[1288]) || legacy.CenterMessageIndexC445530() != 0 || !bytes.Equal(whole, want) {
				t.Fatal("real reset return/index/full UTF-16 and packed guards")
			}
			resetRaster()
			expected := slices.Clone(pix.Pix)
			legacy.Nox_xxx_drawMinimapAndLines_4738E0()
			presentParticleRegistration4B01B0(t, pix, window, seat, expected, frames)
			frames++
			resetCases++
		}
	}
	if memmap.Uint32(0x973a20, 496) != startPresentation || c.Server.Frame() != 37 || c.Server.Rand.Logic.Index() != 93 || c.Server.Rand.Other.Index() != 17 {
		t.Fatal("HUD changed presentation scalar/game clock/RNG")
	}
	t.Logf("CENTER HUD ACTUAL PASS: lang%d; %d native CGo/raster/headless frames, %d resets; direct and actual HUD caller, unsigned expiry, ring order, shadow pixels and packed bytes verified", lang, frames, resetCases)
}

// Capture the immutable original record's path before the isolated child changes cwd.
func filepathCenterOriginal202(t *testing.T) string {
	t.Helper()
	_, file, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("test source path missing")
	}
	return filepath.Join(filepath.Dir(file), "legacy/testdata/center_messages_445450_445530_original.json")
}
