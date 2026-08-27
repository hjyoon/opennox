package server

import (
	"math"
	"reflect"
	"runtime"
	"strings"
	"sync/atomic"
	"testing"
	"unsafe"

	"github.com/opennox/libs/object"
	"github.com/opennox/libs/prand"
)

func TestRewardGemNativeLayouts4F1D30(t *testing.T) {
	wantObjectSize := uintptr(780)
	wantInitData := uintptr(692)
	if unsafe.Sizeof(uintptr(0)) == 8 {
		wantObjectSize = 928
		wantInitData = 760
	}
	checks := []struct {
		name string
		got  uintptr
		want uintptr
	}{
		{"Object size", unsafe.Sizeof(Object{}), wantObjectSize},
		{"Object.InitData", unsafe.Offsetof(Object{}.InitData), wantInitData},
		{"GoldInitData size", unsafe.Sizeof(GoldInitData{}), 4},
		{"GoldInitData.Amount", unsafe.Offsetof(GoldInitData{}.Amount), 0},
		{"gold range width", unsafe.Sizeof(rewardGemGoldBounds4F1D30{}), 12},
		{"gold range Minimum", unsafe.Offsetof(rewardGemGoldBounds4F1D30{}.Minimum), 0},
		{"gold range Maximum", unsafe.Offsetof(rewardGemGoldBounds4F1D30{}.Maximum), 4},
		{"gold range Line", unsafe.Offsetof(rewardGemGoldBounds4F1D30{}.Line), 8},
	}
	for _, check := range checks {
		if check.got != check.want {
			t.Errorf("%s on %s/%s = %d, want %d", check.name, runtime.GOOS, runtime.GOARCH, check.got, check.want)
		}
	}
}

func TestRewardGemNativePreservesPointersAndExactServices4F1D30(t *testing.T) {
	data := &GoldInitData{}
	created := &Object{NetCode: 0xfeedbeef, InitData: unsafe.Pointer(data)}
	if unsafe.Sizeof(uintptr(0)) == 8 && uintptr(unsafe.Pointer(created)) <= math.MaxUint32 {
		t.Fatalf("object pointer does not exercise native high address: %p", created)
	}
	var events []string
	draws := []int32{90, 2, -1}
	got := rewardGemNative4F1D30(4, rewardGemNativeDeps4F1D30{
		pickSlots: func(stage uint32) uint32 {
			events = append(events, "slots")
			if stage != 4 {
				t.Fatalf("stage = %d, want 4", stage)
			}
			return 4
		},
		randomInt: func(minimum, maximum int32, path string, line int32) int32 {
			events = append(events, "rng:"+strings.TrimPrefix(path, `C:\NoxPost\src\server\GameMech\`))
			if path != rewardGemRandomPath4F1D30 {
				t.Fatalf("RNG path = %q", path)
			}
			switch line {
			case 2181:
				if minimum != 1 || maximum != 100 {
					t.Fatalf("gate RNG = %d..%d", minimum, maximum)
				}
			case 2197:
				if minimum != 1 || maximum != 2 {
					t.Fatalf("gold type RNG = %d..%d", minimum, maximum)
				}
			case 2219:
				if minimum != 400 || maximum != 1000 {
					t.Fatalf("amount RNG = %d..%d", minimum, maximum)
				}
			default:
				t.Fatalf("unexpected RNG line %d", line)
			}
			draw := draws[0]
			draws = draws[1:]
			return draw
		},
		createObjectByType: func(name string) *Object {
			events = append(events, "create:"+name)
			if name != rewardGemGoldPileType4F1D30 {
				t.Fatalf("created type = %q, want pile", name)
			}
			return created
		},
	})
	if got != created || data.Amount != math.MaxUint32 || len(draws) != 0 {
		t.Fatalf("result/amount/draws = %p/%#x/%v, want %p/MaxUint32/empty", got, data.Amount, draws, created)
	}
	wantEvents := []string{
		"slots", "rng:Reward.c", "rng:Reward.c", "create:QuestGoldPile", "rng:Reward.c",
	}
	if !reflect.DeepEqual(events, wantEvents) {
		t.Fatalf("native events = %v", events)
	}
}

func TestRewardGemNativeBoundsMatchSealedGAMEEXE4F1D30(t *testing.T) {
	want := [...]rewardGemGoldBounds4F1D30{
		{Minimum: 100, Maximum: 250, Line: 2226},
		{Minimum: 200, Maximum: 500, Line: 2222},
		{Minimum: 400, Maximum: 1000, Line: 2219},
		{Minimum: 800, Maximum: 2000, Line: 2216},
		{Minimum: 1600, Maximum: 4000, Line: 2213},
	}
	if rewardGemGoldAmountBounds4F1D30 != want {
		t.Fatalf("native gold bounds = %v, want %v", rewardGemGoldAmountBounds4F1D30, want)
	}
}

func TestRewardGemServerUsesLogicRNGFactoryAndBothEntrypoints4F1D30(t *testing.T) {
	srv := new(Server)
	srv.handle = atomic.AddUintptr(&serverLast, 1)
	servers.Store(srv.handle, srv)
	t.Cleanup(func() { servers.Delete(srv.handle) })
	srv.Rand.Logic = prand.New(2011)
	srv.Objs.init(srv.handle)
	if !srv.Objs.Init(4) {
		t.Fatal("object allocator initialization failed")
	}
	t.Cleanup(srv.Objs.FreeObjects)

	template := &GoldInitData{}
	newGoldType := func(index uint16, id string) *ObjectType {
		return &ObjectType{
			s: &srv.Types, ind: index, id: id, allowed: true,
			class: object.ClassSimple, flags: object.FlagNoCollide,
			InitData: unsafe.Pointer(template), InitDataSize: unsafe.Sizeof(*template),
		}
	}
	chestType := newGoldType(1, rewardGemGoldChestType4F1D30)
	pileType := newGoldType(2, rewardGemGoldPileType4F1D30)
	srv.Types.byID = map[string]*ObjectType{
		strings.ToLower(chestType.id): chestType,
		strings.ToLower(pileType.id):  pileType,
	}
	srv.Types.byInd = []*ObjectType{nil, chestType, pileType}

	marker := &Object{NetCode: 0xa5a5a5a5}
	if unsafe.Sizeof(uintptr(0)) == 8 && uintptr(unsafe.Pointer(marker)) <= math.MaxUint32 {
		t.Fatalf("marker pointer does not exercise native high address: %p", marker)
	}
	beforeRNG := srv.Rand.Logic.Index()
	first := srv.RewardGem4F1D30(marker, 0)
	second := srv.RewardGem2_4F1F00(marker, 0)
	for index, got := range []*Object{first, second} {
		if got == nil || (got.TypeInd != 1 && got.TypeInd != 2) {
			t.Fatalf("entrypoint %d result = %#v", index, got)
		}
		amount := got.InitDataGold().Amount
		if amount < 100 || amount > 250 {
			t.Fatalf("entrypoint %d amount = %d, want 100..250", index, amount)
		}
	}
	if marker.NetCode != 0xa5a5a5a5 {
		t.Fatal("ignored marker was modified")
	}
	if index, want := srv.Rand.Logic.Index(), (beforeRNG+4)%4096; index != want {
		t.Fatalf("logic RNG index = %d, want %d after two fixed-stage gold rewards", index, want)
	}
}

func TestRewardGemNativeFaultBoundaries4F1D30(t *testing.T) {
	t.Run("nil gold InitData faults only at store", func(t *testing.T) {
		created := &Object{}
		var events []string
		defer func() {
			fault := recover()
			want := []string{"slots", "type", "create", "amount"}
			if fault == nil || !reflect.DeepEqual(events, want) {
				t.Fatalf("fault/events = %v/%v, want panic/%v", fault, events, want)
			}
		}()
		rewardGemNative4F1D30(0, rewardGemNativeDeps4F1D30{
			pickSlots: func(uint32) uint32 { events = append(events, "slots"); return 1 },
			randomInt: func(_, _ int32, _ string, line int32) int32 {
				if line == rewardGemGoldTypeLine4F1D30 {
					events = append(events, "type")
					return 1
				}
				events = append(events, "amount")
				return 100
			},
			createObjectByType: func(string) *Object { events = append(events, "create"); return created },
		})
	})

	t.Run("nil gem is returned without InitData load", func(t *testing.T) {
		draws := []int32{91, 90}
		got := rewardGemNative4F1D30(4, rewardGemNativeDeps4F1D30{
			pickSlots: func(uint32) uint32 { return 4 },
			randomInt: func(int32, int32, string, int32) int32 {
				draw := draws[0]
				draws = draws[1:]
				return draw
			},
			createObjectByType: func(string) *Object { return nil },
		})
		if got != nil || len(draws) != 0 {
			t.Fatalf("nil gem/draws = %#v/%v", got, draws)
		}
	})
}
