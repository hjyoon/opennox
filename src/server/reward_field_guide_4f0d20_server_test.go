package server

import (
	"math"
	"reflect"
	"sync/atomic"
	"testing"
	"unsafe"

	"github.com/opennox/libs/object"
	"github.com/opennox/libs/prand"
)

func TestRewardFieldGuideNativeLayout4F0D20(t *testing.T) {
	wantObjectSize := uintptr(780)
	wantInitData := uintptr(692)
	wantUseData := uintptr(736)
	if unsafe.Sizeof(uintptr(0)) == 8 {
		wantObjectSize = 928
		wantInitData = 760
		wantUseData = 848
	}
	checks := []struct {
		name string
		got  uintptr
		want uintptr
	}{
		{"Object size", unsafe.Sizeof(Object{}), wantObjectSize},
		{"Object.InitData", unsafe.Offsetof(Object{}.InitData), wantInitData},
		{"Object.UseData", unsafe.Offsetof(Object{}.UseData), wantUseData},
		{"RewardMarkerInitData size", unsafe.Sizeof(RewardMarkerInitData{}), 220},
		{"RewardMarkerInitData.RewardFlags", unsafe.Offsetof(RewardMarkerInitData{}.RewardFlags), 4},
		{"RewardMarkerInitData.Guides", unsafe.Offsetof(RewardMarkerInitData{}.Guides), 151},
		{"RewardMarkerInitData.ChanceMode", unsafe.Offsetof(RewardMarkerInitData{}.ChanceMode), 212},
		{"FieldGuideUseData size", unsafe.Sizeof(FieldGuideUseData{}), 64},
		{"FieldGuideUseData.CreatureBuf", unsafe.Offsetof(FieldGuideUseData{}.CreatureBuf), 0},
	}
	for _, check := range checks {
		if check.got != check.want {
			t.Errorf("%s = %d, want %d", check.name, check.got, check.want)
		}
	}
	if len(RewardMarkerInitData{}.Guides) != rewardFieldGuideCount4F0D20 {
		t.Fatalf("explicit guide count = %d, want %d", len(RewardMarkerInitData{}.Guides), rewardFieldGuideCount4F0D20)
	}
}

func TestRewardFieldGuideNativeUsesHighPointersCachedDataAndLivePass4F0D20(t *testing.T) {
	entry := &RewardMarkerInitData{RewardFlags: rewardFieldGuideExplicitFlag4F0D20}
	entry.Guides[1], entry.Guides[40] = 1, 1
	replacement := &RewardMarkerInitData{}
	marker := &Object{InitData: unsafe.Pointer(entry)}
	useData := &FieldGuideUseData{}
	created := &Object{}
	created.UseData.SetPtr(unsafe.Pointer(useData))
	if unsafe.Sizeof(uintptr(0)) == 8 && (uintptr(unsafe.Pointer(marker)) <= math.MaxUint32 || uintptr(unsafe.Pointer(created)) <= math.MaxUint32) {
		t.Fatalf("test pointers do not exercise native high addresses: marker=%p created=%p", marker, created)
	}

	var events []string
	got := rewardFieldGuideNative4F0D20(marker, 0xfeedbeef, rewardFieldGuideNativeDeps4F0D20{
		randomInt: func(minimum, maximum int32) int32 {
			events = append(events, "rng")
			if minimum != 0 || maximum != 1 {
				t.Fatalf("explicit RNG bounds = %d..%d, want 0..1", minimum, maximum)
			}
			marker.InitData = unsafe.Pointer(replacement)
			entry.Guides[1] = 0
			entry.Guides[35] = 1
			return 0
		},
		createObjectByType: func(typeName string) *Object {
			events = append(events, "create:"+typeName)
			return created
		},
	})
	if got != created || useData.Creature() != "Zombie" {
		t.Fatalf("result/creature = %p/%q, want %p/Zombie", got, useData.Creature(), created)
	}
	if !reflect.DeepEqual(events, []string{"rng", "create:FieldGuide"}) || marker.InitData != unsafe.Pointer(replacement) {
		t.Fatalf("events/live marker = %v/%p, want [rng create:FieldGuide]/replacement", events, marker.InitData)
	}
}

func TestRewardFieldGuideNativeAutomaticUsesLiveRows4F0D20(t *testing.T) {
	rows := []rewardFieldGuideDefinition4F0D20{
		{Weight: 2, GuideID: 1, Slots: 1},
		{Weight: 3, GuideID: 2, Slots: 1},
		{Slots: 0x1f},
	}
	marker := &Object{InitData: unsafe.Pointer(&RewardMarkerInitData{})}
	useData := &FieldGuideUseData{}
	created := &Object{}
	created.UseData.SetPtr(unsafe.Pointer(useData))
	got := rewardFieldGuideNative4F0D20(marker, math.MaxUint32, rewardFieldGuideNativeDeps4F0D20{
		rows: rows,
		pickSlots: func(stage uint32) uint32 {
			if stage != math.MaxUint32 {
				t.Fatalf("stage = %#x, want MaxUint32", stage)
			}
			return 1
		},
		randomInt: func(minimum, maximum int32) int32 {
			if minimum != 0 || maximum != 4 {
				t.Fatalf("weighted RNG bounds = %d..%d, want 0..4", minimum, maximum)
			}
			rows[0].Weight = 5
			rows[0].GuideID = 40
			return 3
		},
		createObjectByType: func(typeName string) *Object {
			if typeName != rewardFieldGuideType4F0D20 {
				t.Fatalf("created type = %q, want FieldGuide", typeName)
			}
			return created
		},
	})
	if got != created || useData.Creature() != "UrchinShaman" {
		t.Fatalf("result/creature = %p/%q, want %p/UrchinShaman", got, useData.Creature(), created)
	}
}

func TestRewardFieldGuideServerUsesLogicRNGAndObjectFactory4F0D20(t *testing.T) {
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

	templateUse := &FieldGuideUseData{}
	guideType := &ObjectType{
		s:           &srv.Types,
		ind:         1,
		id:          rewardFieldGuideType4F0D20,
		flags:       object.FlagNoCollide,
		UseData:     UseDataPtr{Ptr: unsafe.Pointer(templateUse)},
		UseDataSize: unsafe.Sizeof(*templateUse),
	}
	srv.Types.byID = map[string]*ObjectType{"fieldguide": guideType}
	srv.Types.byInd = []*ObjectType{nil, guideType}
	data := &RewardMarkerInitData{RewardFlags: rewardFieldGuideExplicitFlag4F0D20}
	data.Guides[40] = 1
	marker := &Object{InitData: unsafe.Pointer(data)}

	beforeRNG := srv.Rand.Logic.Index()
	got := srv.RewardFieldGuide4F0D20(marker, math.MaxUint32)
	if got == nil {
		t.Fatal("server object factory returned nil")
	}
	if got.TypeInd != 1 || got.UseDataFieldGuide().Creature() != "UrchinShaman" {
		t.Fatalf("server result type/creature = %d/%q, want 1/UrchinShaman", got.TypeInd, got.UseDataFieldGuide().Creature())
	}
	if index, want := srv.Rand.Logic.Index(), (beforeRNG+1)%4096; index != want {
		t.Fatalf("logic RNG index = %d, want %d after one explicit-selection draw", index, want)
	}
	got.UseData.Free()
}

func TestRewardFieldGuideNativeFaultBoundaries4F0D20(t *testing.T) {
	t.Run("nil marker faults before dependencies", func(t *testing.T) {
		calls := 0
		defer func() {
			fault := recover()
			if fault == nil || calls != 0 {
				t.Fatalf("nil marker recover/calls = %v/%d, want panic/0", fault, calls)
			}
		}()
		rewardFieldGuideNative4F0D20(nil, 0, rewardFieldGuideNativeDeps4F0D20{
			pickSlots: func(uint32) uint32 { calls++; return 1 },
		})
	})

	t.Run("nil InitData faults before dependencies", func(t *testing.T) {
		calls := 0
		defer func() {
			fault := recover()
			if fault == nil || calls != 0 {
				t.Fatalf("nil InitData recover/calls = %v/%d, want panic/0", fault, calls)
			}
		}()
		rewardFieldGuideNative4F0D20(&Object{}, 0, rewardFieldGuideNativeDeps4F0D20{
			pickSlots: func(uint32) uint32 { calls++; return 1 },
		})
	})

	t.Run("nil UseData faults after create and name lookup", func(t *testing.T) {
		data := &RewardMarkerInitData{RewardFlags: rewardFieldGuideExplicitFlag4F0D20}
		data.Guides[1] = 1
		created := &Object{}
		createdCalls := 0
		defer func() {
			fault := recover()
			if fault == nil || createdCalls != 1 {
				t.Fatalf("nil UseData recover/create calls = %v/%d, want panic/1", fault, createdCalls)
			}
		}()
		rewardFieldGuideNative4F0D20(&Object{InitData: unsafe.Pointer(data)}, 0, rewardFieldGuideNativeDeps4F0D20{
			randomInt: func(int32, int32) int32 { return 0 },
			createObjectByType: func(string) *Object {
				createdCalls++
				return created
			},
		})
	})
}
