package server

import (
	"math"
	"testing"
	"unsafe"

	"github.com/opennox/libs/types"
)

func TestAnkhTradableDrop4EE370NativeLayout(t *testing.T) {
	if got := unsafe.Sizeof(types.Pointf{}); got != 8 {
		t.Fatalf("Pointf size = %d, want 8", got)
	}
	if got := unsafe.Offsetof(types.Pointf{}.X); got != 0 {
		t.Fatalf("Pointf.X offset = %d, want 0", got)
	}
	if got := unsafe.Offsetof(types.Pointf{}.Y); got != 4 {
		t.Fatalf("Pointf.Y offset = %d, want 4", got)
	}
}

func TestAnkhTradableDropNative4EE370BindsPointersAndExactResult(t *testing.T) {
	owner := &Object{}
	item := &Object{}
	point := &types.Pointf{X: 3.5, Y: -9.25}
	runtime := AnkhTradableDropRuntime4EE370{
		DefaultDrop: func(gotOwner, gotItem *Object, gotPoint *types.Pointf) int32 {
			if gotOwner != owner || gotItem != item || gotPoint != point {
				t.Fatalf("DefaultDrop args = %p/%p/%p, want %p/%p/%p", gotOwner, gotItem, gotPoint, owner, item, point)
			}
			return math.MinInt32
		},
	}
	if got := ankhTradableDropNative4EE370(owner, item, point, runtime); got != math.MinInt32 {
		t.Fatalf("result = %d, want %d", got, int32(math.MinInt32))
	}
}

func TestAnkhTradableDrop4EE370ServerBindingForwardsNil(t *testing.T) {
	s := &Server{}
	runtime := AnkhTradableDropRuntime4EE370{
		DefaultDrop: func(owner, item *Object, point *types.Pointf) int32 {
			if owner != nil || item != nil || point != nil {
				t.Fatalf("DefaultDrop args = %p/%p/%p, want nil/nil/nil", owner, item, point)
			}
			return -77
		},
	}
	if got := s.AnkhTradableDrop4EE370(nil, nil, nil, runtime); got != -77 {
		t.Fatalf("result = %d, want -77", got)
	}
}
