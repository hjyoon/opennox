//go:build !server

package legacy

import (
	"bytes"
	"crypto/sha256"
	_ "embed"
	"encoding/binary"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"testing"
	"unsafe"

	"github.com/opennox/libs/object"
	"github.com/opennox/opennox/v1/client"
	"github.com/opennox/opennox/v1/legacy/common/alloc"
)

//go:embed testdata/drawable_predicates_original_4756e0.json
var spriteDrawOriginal4756E0 []byte

func TestSpriteDrawPredicatesOriginal4756E0(t *testing.T) {
	sum := sha256.Sum256(spriteDrawOriginal4756E0)
	if hex.EncodeToString(sum[:]) != "99326d1d61fbe4d28789143ce59c2c9f78165d79117fb1cc11cdfccce6f73525" {
		t.Fatal("independent original predicate reference changed")
	}
	var rows []struct {
		Callback string `json:"callback"`
		Class    uint32 `json:"cls"`
		Flags    uint32 `json:"flags"`
		Returns  [2]int `json:"returns"`
	}
	if err := json.Unmarshal(spriteDrawOriginal4756E0, &rows); err != nil {
		t.Fatal(err)
	}
	if len(rows) != 1280 {
		t.Fatal("incomplete original contexts")
	}
	callbacks := map[string]unsafe.Pointer{
		"nil": nil, "static": Get_nox_thing_static_draw(), "random": Get_nox_thing_static_random_draw(),
		"other1": Get_nox_thing_animate_draw(), "other2": Get_nox_thing_animate_state_draw(),
	}
	guard, free := alloc.NewHigh(struct {
		Before   [4]uint64
		Drawable client.Drawable
		After    [4]uint64
	}{})
	defer free()
	guard.Before, guard.After = [4]uint64{0xA5A5A5A5A5A5A5A5, 1, 2, 3}, [4]uint64{4, 5, 6, 0x5A5A5A5A5A5A5A5A}
	if unsafe.Sizeof(uintptr(0)) == 8 {
		if uintptr(unsafe.Pointer(guard))>>32 == 0 {
			t.Fatal("fixture must retain its native high address")
		}
		for name, ptr := range callbacks {
			if ptr != nil && uintptr(ptr)>>32 == 0 {
				t.Fatalf("low callback %s", name)
			}
		}
	}
	for _, row := range rows {
		t.Run(fmt.Sprintf("%s/flags-%08x/class-%08x", row.Callback, row.Flags, row.Class), func(t *testing.T) {
			guard.Drawable = client.Drawable{ObjClass: object.Class(row.Class), ObjFlags: object.Flags(row.Flags), DrawFuncPtr: callbacks[row.Callback]}
			if unsafe.Sizeof(uintptr(0)) == 8 {
				legacy := unsafe.Slice((*byte)(unsafe.Pointer(&guard.Drawable)), int(unsafe.Sizeof(guard.Drawable)))
				for _, offset := range []int{112, 120, 300} {
					binary.LittleEndian.PutUint32(legacy[offset:], 0x5A5AA5A5)
				}
			}
			memory := unsafe.Slice((*byte)(unsafe.Pointer(guard)), int(unsafe.Sizeof(*guard)))
			before := append([]byte(nil), memory...)
			got := [2]int{Nox_xxx_sprite_4756E0_drawable(&guard.Drawable), Nox_xxx_sprite_475740_drawable(&guard.Drawable)}
			if got != row.Returns {
				t.Fatalf("actual native C returns %v; unchanged original %v", got, row.Returns)
			}
			if !bytes.Equal(memory, before) {
				t.Fatal("predicate modified drawable or allocation guards")
			}
		})
	}
}
