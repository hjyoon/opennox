package legacy

import (
	"math"
	"testing"
	"unsafe"

	"github.com/opennox/opennox/v1/server"
)

func TestRewardAbilityBookCall4F0C70KeepsNativeMarkerPointer(t *testing.T) {
	s := new(server.Server)
	data := &server.RewardMarkerInitData{RewardFlags: 2}
	marker := &server.Object{InitData: unsafe.Pointer(data)}
	if unsafe.Sizeof(uintptr(0)) == 8 && uintptr(unsafe.Pointer(marker)) <= math.MaxUint32 {
		t.Fatalf("marker address = %#x, want a high native pointer", uintptr(unsafe.Pointer(marker)))
	}
	if got := rewardAbilityBookCall4F0C70(s, marker); got != nil {
		t.Fatalf("empty explicit marker result = %p, want nil", got)
	}
}
