package legacy

import (
	"math"
	"testing"
	"unsafe"

	"github.com/opennox/libs/prand"

	"github.com/opennox/opennox/v1/server"
)

func TestRewardMarkerActivateCall4F0720KeepsNativePointers(t *testing.T) {
	s := new(server.Server)
	s.Rand.Logic = prand.New(2011)
	data := &server.RewardMarkerInitData{CategoryMask: 1}
	marker := &server.Object{InitData: unsafe.Pointer(data)}
	created := &server.Object{NetCode: 0xa5a5a5a5}
	if unsafe.Sizeof(uintptr(0)) == 8 && uintptr(unsafe.Pointer(marker)) <= math.MaxUint32 {
		t.Fatalf("marker address = %#x, want a high native pointer", uintptr(unsafe.Pointer(marker)))
	}

	wrong := func(*server.Object, uint32) *server.Object {
		t.Fatal("wrong reward creator called")
		return nil
	}
	runtime := server.RewardMarkerActivateRuntime4F0720{
		AbilityBook: wrong, FieldGuide: wrong, Weapon: wrong, Armor: wrong,
		Gem: wrong, Potion: wrong, Gem2: wrong,
		SpellBook: func(got *server.Object, stage uint32) *server.Object {
			if got != marker || stage != 9 {
				t.Fatalf("marker/stage = %p/%d, want %p/9", got, stage, marker)
			}
			return created
		},
	}

	if got := rewardMarkerActivateCall4F0720(s, marker, 7, runtime); got != created {
		t.Fatalf("result = %p, want %p", got, created)
	}
}
