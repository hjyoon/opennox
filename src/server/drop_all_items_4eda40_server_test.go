package server

import (
	"math"
	"testing"

	"github.com/opennox/libs/prand"
	"github.com/opennox/libs/types"
)

func TestDropAllItemsServerDeps4EDA40UseLogicRNGAndRuntimeDispatcher(t *testing.T) {
	s := new(Server)
	s.Rand.Logic = prand.New(0)
	owner := new(Object)
	item := new(Object)
	point := &types.Pointf{X: 1, Y: 2}
	dispatchCalls := 0
	deps := dropAllItemsServerDeps4EDA40(s, DropAllItemsRuntime4EDA40{
		Dispatch: func(gotOwner, gotItem *Object, gotPoint *types.Pointf) int32 {
			dispatchCalls++
			if gotOwner != owner || gotItem != item || gotPoint != point {
				t.Fatalf("dispatch args = %p/%p/%p, want %p/%p/%p", gotOwner, gotItem, gotPoint, owner, item, point)
			}
			return math.MinInt32
		},
	})

	gotRandom := deps.randomFloat(-3, 3, dropAllItemsSource4EDA40, 823)
	if bits := math.Float64bits(gotRandom); bits != 0x400321d643000000 {
		t.Fatalf("first random result bits = %#016x, want 0x400321d643000000", bits)
	}
	if index := s.Rand.Logic.Index(); index != 1 {
		t.Fatalf("logic index = %d, want 1", index)
	}
	if got := deps.drop(owner, item, point); got != math.MinInt32 || dispatchCalls != 1 {
		t.Fatalf("dispatch result/calls = %d/%d, want %d/1", got, dispatchCalls, int32(math.MinInt32))
	}
}
