package legacy

import (
	"math"
	"testing"
	"unsafe"

	"github.com/opennox/opennox/v1/server"
)

func TestRewardSpellBookCall4F09F0KeepsNativeMarkerPointer(t *testing.T) {
	s := new(server.Server)
	data := &server.RewardMarkerInitData{RewardFlags: 1}
	marker := &server.Object{InitData: unsafe.Pointer(data)}
	if unsafe.Sizeof(uintptr(0)) == 8 && uintptr(unsafe.Pointer(marker)) <= math.MaxUint32 {
		t.Fatalf("marker address = %#x, want a high native pointer", uintptr(unsafe.Pointer(marker)))
	}
	if got := rewardSpellBookCall4F09F0(s, marker, math.MaxUint32); got != nil {
		t.Fatalf("empty explicit marker result = %p, want nil", got)
	}
}

func TestRewardRandomSlotsCall4F0B60FixedStagesNeedNoRNG(t *testing.T) {
	s := new(server.Server)
	if got := rewardRandomSlotsCall4F0B60(s, 0); got != 1 {
		t.Fatalf("stage zero = %#x, want 1", got)
	}
	if got := rewardRandomSlotsCall4F0B60(s, math.MaxUint32); got != 16 {
		t.Fatalf("large unsigned stage = %#x, want 16", got)
	}
}
