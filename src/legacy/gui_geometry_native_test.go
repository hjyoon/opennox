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
	"github.com/opennox/opennox/v1/client/noxrender"
	"github.com/opennox/opennox/v1/common/ntype"
	"github.com/opennox/opennox/v1/legacy/common/alloc"
)

//go:embed testdata/gui_geometry_original_v1.json.gz
var guiGeometryNativeOriginal []byte

type guiGeometryNativeCase struct {
	Input struct {
		Index              int
		Entry, Name, Alias string
		Present            bool
		Enabled, Flags     uint32
		Point              [2]uint32
		Bounds             [4]uint32
		Viewport           [13]uint32
		Out                [2]uint32
	}
	Offsets struct{ Point, Bounds, Out, Viewport, Window int } `json:"argument_offsets"`
	Return  uint32                                             `json:"return_u32"`
	Role    *string                                            `json:"return_role"`
	Compat  *int32                                             `json:"compat_boolean"`
	Final   struct {
		Flags    uint32 `json:"window_flags"`
		Point    [2]uint32
		Bounds   [4]uint32
		Viewport [13]uint32
		Out      [2]uint32
	}
	Writes [][3]uint32 `json:"ordered_writes"`
}

func guiGeometryNativeCases(t *testing.T) []guiGeometryNativeCase {
	t.Helper()
	var cases []guiGeometryNativeCase
	npcVoiceOriginalDecode(t, guiGeometryNativeOriginal, "98e77764cc95e8e73b33340c8dda769c6305d8cf11f102d0ca8b07c8f266cf53", "3b7631e099974f0d304c29c117cb5015f1f5399bb0cc5e2e8364962021861e79", 4874, &cases)
	return cases
}

func TestGuiGeometryNativeLayout(t *testing.T) {
	var win gui.Window
	var vp noxrender.Viewport
	var point ntype.Point32
	word := unsafe.Sizeof(uintptr(0))
	want := [25]uintptr{word, unsafe.Sizeof(win), uintptr(unsafe.Pointer(&win.DrawData().Field0)) - uintptr(unsafe.Pointer(&win)), unsafe.Offsetof(win.Flags), unsafe.Sizeof(vp)}
	for j := 0; j < 13; j++ {
		want[5+j] = uintptr(j) * word
	}
	want[18] = unsafe.Sizeof(point)
	want[19] = unsafe.Offsetof(point.Y)
	want[20] = unsafe.Sizeof([4]int32{})
	want[21] = unsafe.Sizeof([2]float32{})
	want[22] = unsafe.Sizeof([4]float32{})
	want[23] = unsafe.Sizeof(win.DrawData().Field0)
	want[24] = unsafe.Sizeof(uint32(0))
	if got := guiGeometryNativeLayoutC(); got != want {
		t.Fatal("native Window/viewport and fixed32 geometry layouts", got, want)
	}
	if unsafe.Offsetof(vp.World) != 4*word || unsafe.Offsetof(vp.Size) != 8*word || unsafe.Offsetof(vp.Field10) != 10*word || unsafe.Offsetof(vp.Field11) != 11*word || unsafe.Offsetof(vp.Jiggle12) != 12*word {
		t.Fatal("native viewport field grouping")
	}
}

func guiGeometryNativeRun(t *testing.T, facade bool) {
	t.Helper()
	cases := guiGeometryNativeCases(t)
	entries := [5]string{"sub_46AE10", "nox_xxx_pointInRect_4281F0", "sub_428220", "sub_4739E0", "sub_473A10"}
	first, last := 0, 3
	if facade {
		first, last = 3, 4
	}
	for route := first; route < last; route++ {
		for _, placement := range []int{0, 64, 512} {
			raw, free := alloc.MakeHigh([]byte{}, 65536)
			t.Cleanup(free)
			if unsafe.Sizeof(uintptr(0)) == 8 && uintptr(unsafe.Pointer(&raw[0])) <= math.MaxUint32 {
				t.Fatal("foreign native owners above4GiB required")
			}
			expected := make([]byte, len(raw))
			for _, c := range cases {
				if facade && c.Input.Index != 0 {
					continue
				}
				t.Run(fmt.Sprintf("route%d/place%d/%s-%s", route, placement, c.Input.Entry, c.Input.Name), func(t *testing.T) {
					i := c.Input.Index
					if i < 0 || i >= len(entries) || c.Input.Entry != entries[i] || (c.Role != nil) != (i == 0) || (c.Compat != nil) != (i == 0) {
						t.Fatal("invalid sealed entry/return domain")
					}
					point, bounds, out := 0x2000, 0x3000, 0x4000
					if (i == 1 || i == 2) && c.Input.Alias != "distinct" {
						switch c.Input.Alias {
						case "bounds0":
							point = bounds
						case "bounds1":
							point = bounds + 4
						case "bounds2":
							point = bounds + 8
						default:
							t.Fatal("invalid fixed32 bounds alias")
						}
					}
					if i >= 3 {
						switch c.Input.Alias {
						case "distinct":
						case "point":
							out = point
						case "point_y":
							out = point + 4
						case "before_point":
							out = point - 4
						default:
							t.Fatal("invalid fixed32 output alias")
						}
					}
					if c.Offsets.Point != point || c.Offsets.Bounds != bounds || c.Offsets.Out != out || c.Offsets.Viewport != 0x1000 || c.Offsets.Window != 0x5000 {
						t.Fatal("original owner roles/alias offsets")
					}
					for j := range raw {
						raw[j] = byte(j*37 + 29)
					}
					input := [25]uint32{uint32(i), 0, c.Input.Enabled, c.Input.Flags}
					if c.Input.Present {
						input[1] = 1
					}
					copy(input[4:6], c.Input.Point[:])
					copy(input[6:10], c.Input.Bounds[:])
					copy(input[10:23], c.Input.Viewport[:])
					copy(input[23:25], c.Input.Out[:])
					guiGeometryNativeSeedC(unsafe.Pointer(&raw[0]), placement, &input, point, bounds, out)
					win := (*gui.Window)(unsafe.Pointer(&raw[placement+0x5000]))
					vp := (*noxrender.Viewport)(unsafe.Pointer(&raw[placement+0x1000]))
					vpWords := (*[13]uintptr)(unsafe.Pointer(vp))
					for j, v := range c.Input.Viewport {
						want := uintptr(int(int32(v)))
						if j == 10 || j == 11 {
							want = uintptr(v)
						}
						if vpWords[j] != want {
							t.Fatal("C-seeded native viewport projection", j)
						}
					}
					if win.DrawData().Field0 != c.Input.Flags {
						t.Fatal("C-seeded native Window Field0")
					}
					copy(expected, raw)
					var owner unsafe.Pointer
					if c.Input.Present {
						owner = unsafe.Pointer(win)
					}
					if facade {
						if got := guiGeometryNativeStyleFacade(owner, int32(c.Input.Enabled)); got != *c.Compat {
							t.Fatal("Go facade normalized0/1", got, *c.Compat)
						}
					} else {
						word, pointer := guiGeometryNativeCall(route, i, owner, int32(c.Input.Enabled), unsafe.Pointer(vp), unsafe.Pointer(&raw[placement+point]), unsafe.Pointer(&raw[placement+bounds]), unsafe.Pointer(&raw[placement+out]))
						if i == 0 {
							want := owner
							role := "nil"
							if c.Input.Present {
								role = "window"
							}
							if pointer != want || word != 0 || *c.Role != role || c.Return != map[bool]uint32{false: 0, true: 0x21005000}[c.Input.Present] {
								t.Fatal("exact nullable full native Window pointer", pointer, want)
							}
						} else if word != c.Return || pointer != nil {
							t.Fatalf("original declared return got%#x/%p want%#x", word, pointer, c.Return)
						}
					}
					for j, write := range c.Writes {
						if write[1] != 4 {
							t.Fatal("non-word original write")
						}
						at := 0
						if i == 0 {
							if !c.Input.Present || j != 0 || write[0] != 0x21005024 {
								t.Fatal("write outside original Window Field0")
							}
							at = placement + 0x5000 + int(uintptr(unsafe.Pointer(&win.DrawData().Field0))-uintptr(unsafe.Pointer(win)))
						} else if i == 3 || i == 4 {
							if j > 1 || write[0] != 0x21000000+uint32(out+4*j) {
								t.Fatal("original ordered fixed32 output write")
							}
							at = placement + out + 4*j
						} else {
							t.Fatal("predicate wrote memory")
						}
						binary.LittleEndian.PutUint32(expected[at:at+4], write[2])
					}
					wantWrites := 0
					if i == 0 && c.Input.Present {
						wantWrites = 1
					}
					if i >= 3 {
						wantWrites = 2
					}
					if len(c.Writes) != wantWrites {
						t.Fatal("exact original write count")
					}
					if win.DrawData().Field0 != c.Final.Flags || *(*[2]uint32)(unsafe.Pointer(&raw[placement+point])) != c.Final.Point || *(*[4]uint32)(unsafe.Pointer(&raw[placement+bounds])) != c.Final.Bounds || *(*[2]uint32)(unsafe.Pointer(&raw[placement+out])) != c.Final.Out {
						t.Fatal("original final Window/geometry words")
					}
					for j, v := range c.Final.Viewport {
						if uint32(vpWords[j]) != v {
							t.Fatal("original readonly viewport", j)
						}
					}
					if !bytes.Equal(raw, expected) {
						t.Fatal("whole64KiB graph differs outside ordered original store projection")
					}
				})
			}
		}
	}
}

func TestGuiGeometryNativeOriginal(t *testing.T) { guiGeometryNativeRun(t, false) }

func TestGuiGeometryNativeStyleFacade(t *testing.T) {
	saved := game2QuickbarTestHooks
	game2QuickbarTestHooks = nil
	defer func() { game2QuickbarTestHooks = saved }()
	guiGeometryNativeRun(t, true)
}

func TestGuiGeometryNativeStyleHook(t *testing.T) {
	saved := game2QuickbarTestHooks
	defer func() { game2QuickbarTestHooks = saved }()
	for _, value := range []int32{0, 1, -1, math.MaxInt32, math.MinInt32} {
		calls := 0
		game2QuickbarTestHooks = &game2QuickbarHooks{Style: func(win unsafe.Pointer, on int32) int32 {
			calls++
			if win != nil || on != 123 {
				t.Fatal("hook arguments")
			}
			return value
		}}
		if got := guiGeometryNativeStyleFacade(nil, 123); got != value || calls != 1 {
			t.Fatal("existing hook bypass return preserved", got, value, calls)
		}
	}
}

func TestGuiGeometryNativeDoorDraw(t *testing.T) {
	for route := 0; route < 2; route++ {
		for _, placement := range []int{0, 64, 512} {
			raw, free := alloc.MakeHigh([]byte{}, 65536)
			t.Cleanup(free)
			if unsafe.Sizeof(uintptr(0)) == 8 && uintptr(unsafe.Pointer(&raw[0])) <= math.MaxUint32 {
				t.Fatal("native DoorDraw owners above4GiB")
			}
			for _, frame := range []int{0, 8, 16, 24} {
				for metal := 1; metal <= 2; metal++ {
					t.Run(fmt.Sprintf("route%d/place%d/frame%d/metal%d", route, placement, frame, metal), func(t *testing.T) {
						for j := range raw {
							raw[j] = byte(j*37 + 29)
						}
						input := [25]uint32{}
						viewport := [13]uint32{12, 24, 652, 504, 100, 200, 740, 680, 640, 480, 0, 0, 0}
						copy(input[10:23], viewport[:])
						guiGeometryNativeSeedC(unsafe.Pointer(&raw[0]), placement, &input, 0x2000, 0x3000, 0x4000)
						guiGeometryNativeDoorSeed(unsafe.Pointer(&raw[0]), placement, frame, metal)
						before := append([]byte(nil), raw...)
						image, x, y, calls, result := guiGeometryNativeDoorCall(route, unsafe.Pointer(&raw[0]), placement)
						dx, dy := int32(0), int32(0)
						switch frame {
						case 0:
							dx, dy = -15, -20
						case 8:
							dx, dy = 15, -20
						case 16:
							dx, dy = 8, 2
						case 24:
							dx, dy = -8, 2
						}
						imageAt := 0x7000
						if metal == 2 {
							imageAt = 0x7200
						}
						if frame == 8 || frame == 24 {
							imageAt += 0x100
						}
						if image != unsafe.Pointer(&raw[placement+imageAt]) || x != 143+12-100-64+dx || y != 279+24-200-79+dy || calls != 1 || result != 1 || !bytes.Equal(raw, before) {
							t.Fatal("bounded actual DoorDraw native viewport/image/XY/readonly graph", image, x, y, calls, result)
						}
					})
				}
			}
		}
	}
}
