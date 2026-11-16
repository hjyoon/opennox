package legacy

import (
	"bytes"
	_ "embed"
	"fmt"
	"math"
	"testing"
	"unsafe"

	"github.com/opennox/opennox/v1/client"
	"github.com/opennox/opennox/v1/legacy/common/alloc"
)

//go:embed testdata/drawable_predicate_original_v1.json.gz
var drawablePredicateNativeOriginal []byte

type drawablePredicateNativeCase struct {
	Input struct {
		Name         string
		Callback     int
		CallbackName string `json:"callback_name"`
		Class, Flags uint32
		Decoys       [3]uint32
	}
	Returns [4]int32
	Final   struct {
		Class, Flags uint32
		DrawFunc     uint32 `json:"draw_func"`
	}
	Writes [][3]uint32 `json:"ordered_writes"`
}

func drawablePredicateNativeCases(t *testing.T) []drawablePredicateNativeCase {
	t.Helper()
	var cases []drawablePredicateNativeCase
	npcVoiceOriginalDecode(t, drawablePredicateNativeOriginal, "ac7e72f17a75dbd38c1e49fcc8504a1a8f0181e3e776ab25137dc320569d204b", "42804bc165c42671e4743f0f3ac57191e333909ab6082178bf486d638dd87fb0", 13460, &cases)
	return cases
}

func TestDrawablePredicateNativeLayout(t *testing.T) {
	var dr client.Drawable
	word := unsafe.Sizeof(uintptr(0))
	align := unsafe.Alignof(dr)
	prefix := (unsafe.Offsetof(dr.Field_127) + unsafe.Sizeof(dr.Field_127) + align - 1) &^ (align - 1)
	want := [13]uintptr{word, prefix, unsafe.Offsetof(dr.ObjClass), unsafe.Offsetof(dr.ObjFlags), unsafe.Offsetof(dr.DrawFuncPtr), unsafe.Sizeof(dr.ObjClass), unsafe.Sizeof(dr.ObjFlags), unsafe.Sizeof(dr.DrawFuncPtr), unsafe.Offsetof(dr.Field_127), unsafe.Sizeof(dr.Field_127), 4, word, word}
	if got := drawablePredicateNativeLayoutC(); got != want {
		t.Fatal("native Drawable predicate fields and C function-pointer ABI", got, want)
	}
	if unsafe.Sizeof(dr) != prefix+word || unsafe.Sizeof(dr) > 0x1000 {
		t.Fatal("C prefix and existing Go-only clientHandle")
	}
	wantRoles := [5]unsafe.Pointer{nil, Get_nox_thing_static_draw(), Get_nox_thing_static_random_draw(), Get_nox_thing_animate_draw(), Get_nox_thing_animate_state_draw()}
	for role, want := range wantRoles {
		if got := drawablePredicateNativeRoleC(role); got != want {
			t.Fatal("C/Go canonical function identity", role, got, want)
		}
		if role > 0 && want == nil {
			t.Fatal("nil native draw-function identity")
		}
		if word == 8 && role > 0 && uintptr(want) <= math.MaxUint32 {
			t.Fatal("full high native callback identity required")
		}
		for prior := 0; prior < role; prior++ {
			if want == wantRoles[prior] {
				t.Fatal("distinct canonical callback identities")
			}
		}
	}
}

func drawablePredicateNativeRun(t *testing.T, routes []int, call func(int, int, unsafe.Pointer) int32, subtests bool) {
	t.Helper()
	cases := drawablePredicateNativeCases(t)
	roles := [5]unsafe.Pointer{}
	for role := range roles {
		roles[role] = drawablePredicateNativeRoleC(role)
	}
	originalRoles := [5]uint32{0, 0x4bcc20, 0x4bcc60, 0x4bbd60, 0x4bbf10}
	names := [5]string{"nil", "static", "random", "animate", "animate_state"}
	for _, route := range routes {
		for _, placement := range []int{0, 64, 512} {
			raw, free := alloc.MakeHigh([]byte{}, 65536)
			t.Cleanup(free)
			if unsafe.Sizeof(uintptr(0)) == 8 && uintptr(unsafe.Pointer(&raw[0])) <= math.MaxUint32 {
				t.Fatal("foreign Drawable owner above4GiB required")
			}
			dr := (*client.Drawable)(unsafe.Pointer(&raw[0x1000+placement]))
			before := make([]byte, len(raw))
			for _, c := range cases {
				check := func(t *testing.T) {
					if c.Input.Callback < 0 || c.Input.Callback >= len(roles) || c.Input.CallbackName != names[c.Input.Callback] || c.Final.Class != c.Input.Class || c.Final.Flags != c.Input.Flags || c.Final.DrawFunc != originalRoles[c.Input.Callback] || len(c.Writes) != 0 {
						t.Fatal("invalid original function/field/readonly domain", c.Input.Name)
					}
					for j := range raw {
						raw[j] = byte(j*37 + 29)
					}
					input := [6]uint32{c.Input.Class, c.Input.Flags, uint32(c.Input.Callback), c.Input.Decoys[0], c.Input.Decoys[1], c.Input.Decoys[2]}
					drawablePredicateNativeSeedC(unsafe.Pointer(&raw[0]), placement, &input)
					if uint32(dr.ObjClass) != c.Input.Class || uint32(dr.ObjFlags) != c.Input.Flags || dr.DrawFuncPtr != roles[c.Input.Callback] {
						t.Fatal("C-seeded shared native Drawable fields", c.Input.Name)
					}
					copy(before, raw)
					for index, want := range c.Returns {
						if want < 0 || want > 1 {
							t.Fatal("original normalized int return", c.Input.Name)
						}
						if got := call(route, index, unsafe.Pointer(dr)); got != want {
							t.Fatalf("%s route%d place%d entry%d: got%d want%d", c.Input.Name, route, placement, index, got, want)
						}
						if !bytes.Equal(raw, before) {
							t.Fatal("whole64KiB Drawable owner/decoys/Go-only extension changed", c.Input.Name, index)
						}
					}
				}
				if subtests {
					t.Run(fmt.Sprintf("route%d/place%d/%s", route, placement, c.Input.Name), check)
				} else {
					check(t)
				}
			}
		}
	}
}

func TestDrawablePredicateNativeOriginal(t *testing.T) {
	drawablePredicateNativeRun(t, []int{0, 1, 2}, drawablePredicateNativeCall, true)
}
