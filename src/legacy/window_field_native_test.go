package legacy

import (
	"bytes"
	_ "embed"
	"encoding/binary"
	"fmt"
	"math"
	"testing"
	"unsafe"

	"github.com/opennox/opennox/v1/client/gui"
	"github.com/opennox/opennox/v1/legacy/common/alloc"
)

//go:embed testdata/window_field_original_v1.json.gz
var windowFieldNativeOriginal []byte

type windowFieldNativeCase struct {
	Input struct {
		Index             int
		Entry, Name       string
		Nonnull           bool
		X, Y, Image, Font uint32
		InitialX          uint32 `json:"initial_x"`
		InitialY          uint32 `json:"initial_y"`
		Field0            uint32
		InitialBg         uint32 `json:"initial_bg"`
		InitialHl         uint32 `json:"initial_hl"`
		InitialSel        uint32 `json:"initial_sel"`
	}
	Return uint32  `json:"return_u32"`
	Role   *string `json:"return_role"`
	Final  map[string]uint32
	Writes [][2]uint32 `json:"ordered_writes"`
}

func TestWindowFieldNativeLayout(t *testing.T) {
	var win gui.Window
	draw := win.DrawData()
	pointer := unsafe.Sizeof(uintptr(0))
	want := [20]uintptr{
		pointer, unsafe.Sizeof(win), unsafe.Sizeof(*draw), uintptr(unsafe.Pointer(draw)) - uintptr(unsafe.Pointer(&win)),
		unsafe.Offsetof(draw.Field0), unsafe.Offsetof(draw.BgImageHnd), unsafe.Offsetof(draw.HlImageHnd), unsafe.Offsetof(draw.SelImageHnd),
		unsafe.Offsetof(draw.ImgPtVal) + unsafe.Offsetof(draw.ImgPtVal.X), unsafe.Offsetof(draw.ImgPtVal) + unsafe.Offsetof(draw.ImgPtVal.Y), unsafe.Offsetof(draw.FontPtr),
		unsafe.Sizeof(draw.Field0), unsafe.Sizeof(draw.BgImageHnd), unsafe.Sizeof(draw.HlImageHnd), unsafe.Sizeof(draw.SelImageHnd),
		unsafe.Sizeof(draw.ImgPtVal.X), unsafe.Sizeof(draw.ImgPtVal.Y), unsafe.Sizeof(draw.FontPtr), 4, pointer,
	}
	if got := windowFieldNativeLayoutC(); got != want {
		t.Fatal("native Window/draw data and pointer-return ABI", got, want)
	}
	if unsafe.Sizeof(win) > 0x1000 || unsafe.Sizeof(draw.ImgPtVal.X) != pointer {
		t.Fatal("native owner extent or signed coordinate width")
	}
}

func TestWindowFieldNativeOriginal(t *testing.T) {
	var cases []windowFieldNativeCase
	npcVoiceOriginalDecode(t, windowFieldNativeOriginal, "57216256283080464ad1126327c3972b9f2f177bf8312acfa570d91368e3cf1c", "8f3bcdf2b8668423d7078d7c18102cdb1449001cecaa43511ba41cf5aec2655d", 2418, &cases)
	for route := 0; route < 3; route++ {
		for _, placement := range []int{0, 64, 512} {
			raw, free := alloc.MakeHigh([]byte{}, 65536)
			t.Cleanup(free)
			if unsafe.Sizeof(uintptr(0)) == 8 && uintptr(unsafe.Pointer(&raw[0])) <= math.MaxUint32 {
				t.Fatal("actual foreign Window graph above4GiB required")
			}
			win := (*gui.Window)(unsafe.Pointer(&raw[0x1000+placement]))
			draw := win.DrawData()
			images := []unsafe.Pointer{nil, unsafe.Pointer(&raw[0x3000+placement]), unsafe.Pointer(&raw[0x4000+placement])}
			fonts := []unsafe.Pointer{nil, unsafe.Pointer(&raw[0x6000+placement]), unsafe.Pointer(&raw[0x7000+placement])}
			project := func(original uint32, font bool) unsafe.Pointer {
				values, pointers := []uint32{0, 0x21003000, 0x21004000}, images
				if font {
					values, pointers = []uint32{0, 0x21006000, 0x21007000}, fonts
				}
				for i, value := range values {
					if value == original {
						return pointers[i]
					}
				}
				t.Fatal("unknown original opaque-pointer role")
				return nil
			}
			expected := make([]byte, len(raw))
			for _, c := range cases {
				t.Run(fmt.Sprintf("route%d/place%d/%s-%s", route, placement, c.Input.Entry, c.Input.Name), func(t *testing.T) {
					i := c.Input.Index
					if i < 0 || i > 4 || len(c.Final) != 7 || c.Input.Image > 2 || c.Input.Font > 2 || c.Input.InitialBg > 2 || c.Input.InitialHl > 2 || c.Input.InitialSel > 2 || (c.Role != nil) != (i == 0 || i == 4) {
						t.Fatal("invalid sealed original Window/return domain")
					}
					for j := range raw {
						raw[j] = byte(j*43 + 19)
					}
					input := [7]uint32{c.Input.InitialX, c.Input.InitialY, c.Input.Field0, c.Input.InitialBg, c.Input.InitialHl, c.Input.InitialSel, c.Input.Font}
					windowFieldNativeSeedC(unsafe.Pointer(&raw[0]), placement, &input)
					if draw.Field0 != c.Input.Field0 || draw.ImgPtVal.X != int(int32(c.Input.InitialX)) || draw.ImgPtVal.Y != int(int32(c.Input.InitialY)) || unsafe.Pointer(draw.BgImageHnd) != images[c.Input.InitialBg] || unsafe.Pointer(draw.HlImageHnd) != images[c.Input.InitialHl] || unsafe.Pointer(draw.SelImageHnd) != images[c.Input.InitialSel] || draw.FontPtr != fonts[c.Input.Font] {
						t.Fatal("actual native C seed differs from named Go fields")
					}
					copy(expected, raw)
					var owner unsafe.Pointer
					if c.Input.Nonnull {
						owner = unsafe.Pointer(win)
					}
					word, pointer := windowFieldNativeCall(route, i, owner, images[c.Input.Image], int32(c.Input.X), int32(c.Input.Y))
					if c.Role != nil {
						var want unsafe.Pointer
						switch *c.Role {
						case "nil":
						case "window":
							want = unsafe.Pointer(win)
						case "font0":
							want = fonts[1]
						case "font1":
							want = fonts[2]
						default:
							t.Fatal("unknown original return identity")
						}
						if pointer != want || word != 0 {
							t.Fatalf("original return %s: got %#x/%p want %p", *c.Role, word, pointer, want)
						}
					} else if word != c.Return || pointer != nil {
						t.Fatalf("original scalar return: got %#x/%p want %#x", word, pointer, c.Return)
					}
					if draw.Field0 != c.Final["field0"] || draw.ImgPtVal.X != int(int32(c.Final["x"])) || draw.ImgPtVal.Y != int(int32(c.Final["y"])) || unsafe.Pointer(draw.BgImageHnd) != project(c.Final["bg"], false) || unsafe.Pointer(draw.HlImageHnd) != project(c.Final["hl"], false) || unsafe.Pointer(draw.SelImageHnd) != project(c.Final["sel"], false) || draw.FontPtr != project(c.Final["font"], true) {
						t.Fatal("original seven final field words/identities differ")
					}
					for _, write := range c.Writes {
						var slot unsafe.Pointer
						var value uintptr
						switch write[0] {
						case 0x1000 + 96:
							slot, value = unsafe.Pointer(&draw.ImgPtVal.X), uintptr(int(int32(write[1])))
						case 0x1000 + 100:
							slot, value = unsafe.Pointer(&draw.ImgPtVal.Y), uintptr(int(int32(write[1])))
						case 0x1000 + 60:
							slot, value = unsafe.Pointer(&draw.BgImageHnd), uintptr(project(write[1], false))
						case 0x1000 + 76:
							slot, value = unsafe.Pointer(&draw.HlImageHnd), uintptr(project(write[1], false))
						case 0x1000 + 92:
							slot, value = unsafe.Pointer(&draw.SelImageHnd), uintptr(project(write[1], false))
						default:
							t.Fatal("undeclared original write offset")
						}
						at := uintptr(slot) - uintptr(unsafe.Pointer(&raw[0]))
						if unsafe.Sizeof(uintptr(0)) == 8 {
							binary.LittleEndian.PutUint64(expected[at:at+8], uint64(value))
						} else {
							binary.LittleEndian.PutUint32(expected[at:at+4], uint32(value))
						}
					}
					if !bytes.Equal(raw, expected) {
						t.Fatal("whole64KiB Window graph changed outside exact original native write projection")
					}
				})
			}
		}
	}
}
