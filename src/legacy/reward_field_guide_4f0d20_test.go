package legacy

import (
	"math"
	"testing"
	"unsafe"

	"github.com/opennox/opennox/v1/server"
)

func TestRewardFieldGuideCall4F0D20KeepsNativeMarkerPointerAndStage(t *testing.T) {
	s := new(server.Server)
	data := &server.RewardMarkerInitData{RewardFlags: 4}
	marker := &server.Object{InitData: unsafe.Pointer(data)}
	if unsafe.Sizeof(uintptr(0)) == 8 && uintptr(unsafe.Pointer(marker)) <= math.MaxUint32 {
		t.Fatalf("marker address = %#x, want a high native pointer", uintptr(unsafe.Pointer(marker)))
	}
	if got := rewardFieldGuideCall4F0D20(s, marker, math.MaxUint32); got != nil {
		t.Fatalf("empty explicit marker result = %p, want nil", got)
	}
}
