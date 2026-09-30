//go:build !server

package opennox

import (
	"testing"
	"unsafe"

	"github.com/opennox/opennox/v1/client"
	"github.com/opennox/opennox/v1/legacy"
)

func TestMonsterGeneratorDrawDispatch4BC750(t *testing.T) {
	fn := legacy.Get_nox_thing_monster_gen_draw()
	if fn == nil {
		t.Fatal("monster-generator callback pointer is nil")
	}
	// Unknown kind 1 is the original no-draw return, after the native data
	// reads. No Client instance, timer update, frame lookup or draw is needed.
	var data struct {
		size   uint32
		images [5]unsafe.Pointer
		count  [5]uint8
		delay  [5]uint8
		kind   [5]uint32
	}
	data.kind[0] = 1
	dr := &client.Drawable{DrawFuncPtr: fn, DrawData: unsafe.Pointer(&data)}
	if got, ok := (*Client)(nil).callMonsterGeneratorDraw4BC750(dr, nil); !ok || got != 0 {
		t.Fatalf("generator dispatch = (%d, %t), want (0, true)", got, ok)
	}
	dr.DrawFuncPtr = nil
	if got, ok := (*Client)(nil).callMonsterGeneratorDraw4BC750(dr, nil); ok || got != 0 {
		t.Fatalf("non-generator dispatch = (%d, %t), want (0, false)", got, ok)
	}
}
