package server

import (
	"math"
	"runtime"
	"slices"
	"sync/atomic"
	"testing"
	"unsafe"

	"github.com/opennox/libs/object"
	"github.com/opennox/libs/prand"
)

func TestRewardPotionNativeLayouts4F1C40(t *testing.T) {
	wantDefinitionSize := uintptr(24)
	wantName := uintptr(4)
	wantType := uintptr(12)
	wantKind := uintptr(16)
	wantSlots := uintptr(20)
	if unsafe.Sizeof(uintptr(0)) == 8 {
		wantDefinitionSize = 40
		wantName = 8
		wantType = 24
		wantKind = 28
		wantSlots = 32
	}
	checks := []struct {
		name string
		got  uintptr
		want uintptr
	}{
		{"reward definition size", unsafe.Sizeof(rewardObjectDefinition4F0640{}), wantDefinitionSize},
		{"reward definition Weight", unsafe.Offsetof(rewardObjectDefinition4F0640{}.Weight), 0},
		{"reward definition Name", unsafe.Offsetof(rewardObjectDefinition4F0640{}.Name), wantName},
		{"reward definition TypeInd", unsafe.Offsetof(rewardObjectDefinition4F0640{}.TypeInd), wantType},
		{"reward definition Kind", unsafe.Offsetof(rewardObjectDefinition4F0640{}.Kind), wantKind},
		{"reward definition Slots", unsafe.Offsetof(rewardObjectDefinition4F0640{}.Slots), wantSlots},
		{"reward definition weight width", unsafe.Sizeof(rewardObjectDefinition4F0640{}.Weight), 4},
		{"reward definition type width", unsafe.Sizeof(rewardObjectDefinition4F0640{}.TypeInd), 4},
		{"reward definition kind width", unsafe.Sizeof(rewardObjectDefinition4F0640{}.Kind), 4},
		{"reward definition slots width", unsafe.Sizeof(rewardObjectDefinition4F0640{}.Slots), 4},
	}
	for _, check := range checks {
		if check.got != check.want {
			t.Errorf("%s on %s/%s = %d, want %d", check.name, runtime.GOOS, runtime.GOARCH, check.got, check.want)
		}
	}
}

func TestRewardPotionNativePreservesObjectPointerAndServices4F1C40(t *testing.T) {
	created := &Object{NetCode: 0xfeedbeef}
	if unsafe.Sizeof(uintptr(0)) == 8 && uintptr(unsafe.Pointer(created)) <= math.MaxUint32 {
		t.Fatalf("object pointer does not exercise a native high address: %p", created)
	}
	var events []string
	got := rewardPotionNative4F1C40(8, rewardPotionNativeDeps4F1C40{
		objects: []rewardObjectDefinition4F0640{
			{Weight: 1, Name: "Potion", TypeInd: 7, Kind: 0x10000004, Slots: 16}, {},
		},
		pickSlots: func(stage uint32) uint32 {
			events = append(events, "slots")
			if stage != 8 {
				t.Fatalf("stage = %d, want 8", stage)
			}
			return 16
		},
		randomInt: func(minimum, maximum int32) int32 {
			events = append(events, "rng")
			if minimum != 0 || maximum != 0 {
				t.Fatalf("RNG = %d..%d, want 0..0", minimum, maximum)
			}
			return 0
		},
		objectTypeAllowed: func(typeInd uint32) bool {
			events = append(events, "allowed")
			return typeInd == 7
		},
		createObject: func(typeInd uint32) *Object {
			events = append(events, "create")
			if typeInd != 7 {
				t.Fatalf("create type = %d, want 7", typeInd)
			}
			return created
		},
	})
	wantEvents := []string{"slots", "allowed", "rng", "allowed", "create"}
	if got != created || !slices.Equal(events, wantEvents) {
		t.Fatalf("result/events = %p/%v, want %p/%v", got, events, created, wantEvents)
	}
}

func TestRewardPotionServerUsesLogicRNGRegistryAndFactory4F1C40(t *testing.T) {
	srv := new(Server)
	srv.handle = atomic.AddUintptr(&serverLast, 1)
	servers.Store(srv.handle, srv)
	t.Cleanup(func() { servers.Delete(srv.handle) })
	srv.Rand.Logic = prand.New(2011)
	srv.Objs.init(srv.handle)
	if !srv.Objs.Init(2) {
		t.Fatal("object allocator initialization failed")
	}
	t.Cleanup(srv.Objs.FreeObjects)

	potionType := &ObjectType{
		s:        &srv.Types,
		ind:      1,
		id:       "RedPotion",
		allowed:  true,
		class:    object.ClassFood,
		subclass: object.SubClass(object.FoodPotion | object.FoodHealthPotion),
		flags:    object.FlagNoCollide,
	}
	srv.Types.byID = map[string]*ObjectType{"redpotion": potionType}
	srv.Types.byInd = []*ObjectType{nil, potionType}
	srv.rewardDefinitions.Objects = [58]rewardObjectDefinition4F0640{
		{Weight: 1, Name: "RedPotion", TypeInd: 1, Kind: 4, Slots: 1},
		{},
	}

	marker := &Object{NetCode: 0xa5a5a5a5}
	if unsafe.Sizeof(uintptr(0)) == 8 && uintptr(unsafe.Pointer(marker)) <= math.MaxUint32 {
		t.Fatalf("marker pointer does not exercise a native high address: %p", marker)
	}
	beforeRNG := srv.Rand.Logic.Index()
	got := srv.RewardPotion4F1C40(marker, 0)
	if got == nil || got.TypeInd != 1 || !got.Class().Has(object.ClassFood) || !got.SubClass().AsFood().Has(object.FoodPotion) {
		t.Fatalf("server result = %#v, want native potion type 1", got)
	}
	if marker.NetCode != 0xa5a5a5a5 {
		t.Fatal("ignored marker was modified")
	}
	if index, want := srv.Rand.Logic.Index(), (beforeRNG+1)%4096; index != want {
		t.Fatalf("logic RNG index = %d, want %d after object draw", index, want)
	}
}
