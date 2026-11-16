//go:build !server

package opennox

import (
	"bytes"
	"compress/gzip"
	"crypto/sha256"
	_ "embed"
	"encoding/binary"
	"encoding/hex"
	"encoding/json"
	"io"
	"math"
	"testing"
	"unsafe"

	"github.com/opennox/libs/object"
	"github.com/opennox/opennox/v1/client"
	"github.com/opennox/opennox/v1/legacy"
	"github.com/opennox/opennox/v1/legacy/common/alloc"
)

//go:embed testdata/drawable_predicate_original_v1.json.gz
var drawableQueueOriginalNative []byte

func TestDrawableQueueNativeOriginal(t *testing.T) {
	sum := sha256.Sum256(drawableQueueOriginalNative)
	if hex.EncodeToString(sum[:]) != "ac7e72f17a75dbd38c1e49fcc8504a1a8f0181e3e776ab25137dc320569d204b" {
		t.Fatal("sealed original Drawable predicate archive changed")
	}
	z, err := gzip.NewReader(bytes.NewReader(drawableQueueOriginalNative))
	if err != nil {
		t.Fatal(err)
	}
	plain, err := io.ReadAll(z)
	if err != nil {
		t.Fatal(err)
	}
	if err = z.Close(); err != nil {
		t.Fatal(err)
	}
	sum = sha256.Sum256(plain)
	if hex.EncodeToString(sum[:]) != "42804bc165c42671e4743f0f3ac57191e333909ab6082178bf486d638dd87fb0" {
		t.Fatal("sealed original Drawable rows changed")
	}
	var cases []struct {
		Input struct {
			Name         string
			Callback     int
			CallbackName string `json:"callback_name"`
			Class, Flags uint32
			Decoys       [3]uint32
		}
		Returns [4]int
		Final   struct {
			Class, Flags uint32
			DrawFunc     uint32 `json:"draw_func"`
		}
		Writes [][3]uint32 `json:"ordered_writes"`
	}
	if err = json.Unmarshal(plain, &cases); err != nil {
		t.Fatal(err)
	}
	if len(cases) != 13460 {
		t.Fatal("incomplete original Drawable contexts")
	}
	callbacks := [5]unsafe.Pointer{nil, legacy.Get_nox_thing_static_draw(), legacy.Get_nox_thing_static_random_draw(), legacy.Get_nox_thing_animate_draw(), legacy.Get_nox_thing_animate_state_draw()}
	original := [5]uint32{0, 0x4bcc20, 0x4bcc60, 0x4bbd60, 0x4bbf10}
	names := [5]string{"nil", "static", "random", "animate", "animate_state"}
	functions := [4]func(*client.Drawable) bool{spriteBackWallQueueNative4756E0, spriteFrontQueueNative475740, spriteOverlayQueueNative4757A0, spriteWorldQueueNative4757D0}
	for index, fn := range functions {
		if fn(nil) {
			t.Fatal("existing Go nil guard", index)
		}
	}
	for role, pointer := range callbacks {
		if unsafe.Sizeof(uintptr(0)) == 8 && role > 0 && uintptr(pointer) <= math.MaxUint32 {
			t.Fatal("actual high native callback required", role)
		}
	}
	checks := 0
	for _, placement := range []int{0, 64, 512} {
		raw, free := alloc.MakeHigh([]byte{}, 65536)
		t.Cleanup(free)
		if unsafe.Sizeof(uintptr(0)) == 8 && uintptr(unsafe.Pointer(&raw[0])) <= math.MaxUint32 {
			t.Fatal("foreign native renderer Drawable owner above4GiB required")
		}
		dr := (*client.Drawable)(unsafe.Pointer(&raw[0x1000+placement]))
		before := make([]byte, len(raw))
		for _, c := range cases {
			if c.Input.Callback < 0 || c.Input.Callback >= len(callbacks) || c.Input.CallbackName != names[c.Input.Callback] || c.Final.Class != c.Input.Class || c.Final.Flags != c.Input.Flags || c.Final.DrawFunc != original[c.Input.Callback] || len(c.Writes) != 0 {
				t.Fatal("original readonly field/function role", c.Input.Name)
			}
			for j := range raw {
				raw[j] = byte(j*37 + 29)
			}
			for j, offset := range []int{112, 120, 300} {
				binary.LittleEndian.PutUint32(raw[placement+0x1000+offset:], c.Input.Decoys[j])
			}
			dr.ObjClass = object.Class(c.Input.Class)
			dr.ObjFlags = object.Flags(c.Input.Flags)
			dr.DrawFuncPtr = callbacks[c.Input.Callback]
			copy(before, raw)
			for index, fn := range functions {
				value := 0
				if fn(dr) {
					value = 1
				}
				checks++
				if value != c.Returns[index] {
					t.Fatalf("actual renderer helper %s place%d entry%d got%d want%d", c.Input.Name, placement, index, value, c.Returns[index])
				}
				if !bytes.Equal(raw, before) {
					t.Fatal("renderer helper changed whole64KiB readonly owner", c.Input.Name, index)
				}
			}
		}
	}
	if checks != 161520 {
		t.Fatal("incomplete renderer helper original checks", checks)
	}
	t.Log("Actual renderer helper adapters:13460 original contexts x3 placements x4 entries;161520 scalar results and whole64KiB readonly owners")
}
