package legacy

import (
	"math"
	"testing"
	"unsafe"

	"github.com/opennox/libs/types"
	"github.com/opennox/opennox/v1/legacy/common/alloc"
)

func TestQuestNecroSpawn51A7A0OriginalCEntryFullPositionPointer(t *testing.T) {
	position, free := alloc.New(types.Pointf{})
	defer free()
	*position = types.Pointf{X: math.Float32frombits(0x7fc12345), Y: math.Float32frombits(0x80000000)}
	if unsafe.Sizeof(uintptr(0)) > 4 && uintptr(unsafe.Pointer(position)) <= math.MaxUint32 {
		t.Fatalf("C position=%p, want above 4 GiB", position)
	}
	old := questNecroSpawnCall51A7A0
	defer func() { questNecroSpawnCall51A7A0 = old }()
	calls := 0
	for _, want := range []*types.Pointf{position, nil} {
		questNecroSpawnCall51A7A0 = func(got *types.Pointf) {
			calls++
			if got != want {
				t.Fatalf("C position=%p, want %p", got, want)
			}
			if got != nil && (math.Float32bits(got.X) != 0x7fc12345 || math.Float32bits(got.Y) != 0x80000000) {
				t.Fatal("C position bits changed")
			}
		}
		questNecroSpawnCEntry51A7A0(want)
	}
	if calls != 2 {
		t.Fatalf("calls=%d", calls)
	}
}
