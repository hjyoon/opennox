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

func TestRewardAbilityBookNativeLayout4F0C70(t *testing.T) {
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
		{"RewardMarkerInitData.Abilities", unsafe.Offsetof(RewardMarkerInitData{}.Abilities), 145},
		{"RewardMarkerInitData.ChanceMode", unsafe.Offsetof(RewardMarkerInitData{}.ChanceMode), 212},
		{"AbilityRewardUseData size", unsafe.Sizeof(AbilityRewardUseData{}), 1},
		{"AbilityRewardUseData.Ability", unsafe.Offsetof(AbilityRewardUseData{}.Ability), 0},
	}
	for _, check := range checks {
		if check.got != check.want {
			t.Errorf("%s = %d, want %d", check.name, check.got, check.want)
		}
	}
	if len(RewardMarkerInitData{}.Abilities) != rewardAbilityBookCount4F0C70 {
		t.Fatalf("explicit ability count = %d, want %d", len(RewardMarkerInitData{}.Abilities), rewardAbilityBookCount4F0C70)
	}
}

func TestRewardAbilityBookNativeUsesHighPointersCachedDataAndLivePass4F0C70(t *testing.T) {
	entry := &RewardMarkerInitData{RewardFlags: rewardAbilityBookExplicitFlag4F0C70}
	entry.Abilities[1], entry.Abilities[5] = 1, 1
	replacement := &RewardMarkerInitData{}
	marker := &Object{InitData: unsafe.Pointer(entry)}
	useData := &AbilityRewardUseData{}
	created := &Object{}
	created.UseData.SetPtr(unsafe.Pointer(useData))
	if unsafe.Sizeof(uintptr(0)) == 8 && (uintptr(unsafe.Pointer(marker)) <= math.MaxUint32 || uintptr(unsafe.Pointer(created)) <= math.MaxUint32) {
		t.Fatalf("test pointers do not exercise native high addresses: marker=%p created=%p", marker, created)
	}

	var events []string
	got := rewardAbilityBookNative4F0C70(marker, rewardAbilityBookNativeDeps4F0C70{
		randomInt: func(minimum, maximum int32) int32 {
			events = append(events, "rng")
			if minimum != 0 || maximum != 1 {
				t.Fatalf("explicit RNG bounds = %d..%d, want 0..1", minimum, maximum)
			}
			marker.InitData = unsafe.Pointer(replacement)
			entry.Abilities[1] = 0
			entry.Abilities[4] = 1
			return 0
		},
		createObjectByType: func(typeName string) *Object {
			events = append(events, "create:"+typeName)
			return created
		},
	})
	if got != created || useData.Ability != 4 {
		t.Fatalf("result/ability = %p/%d, want %p/4", got, useData.Ability, created)
	}
	if !reflect.DeepEqual(events, []string{"rng", "create:AbilityBook"}) || marker.InitData != unsafe.Pointer(replacement) {
		t.Fatalf("events/live marker = %v/%p, want [rng create:AbilityBook]/replacement", events, marker.InitData)
	}
}

func TestRewardAbilityBookNativeAutomaticPreservesSignedResult4F0C70(t *testing.T) {
	marker := &Object{InitData: unsafe.Pointer(&RewardMarkerInitData{})}
	useData := &AbilityRewardUseData{}
	created := &Object{}
	created.UseData.SetPtr(unsafe.Pointer(useData))
	got := rewardAbilityBookNative4F0C70(marker, rewardAbilityBookNativeDeps4F0C70{
		randomInt: func(minimum, maximum int32) int32 {
			if minimum != 1 || maximum != 5 {
				t.Fatalf("automatic RNG bounds = %d..%d, want 1..5", minimum, maximum)
			}
			return -1
		},
		createObjectByType: func(typeName string) *Object {
			if typeName != rewardAbilityBookType4F0C70 {
				t.Fatalf("created type = %q, want AbilityBook", typeName)
			}
			return created
		},
	})
	if got != created || useData.Ability != 0xff {
		t.Fatalf("result/ability = %p/%#x, want %p/0xff", got, useData.Ability, created)
	}
}

func TestRewardAbilityBookServerUsesLogicRNGAndObjectFactory4F0C70(t *testing.T) {
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

	templateUse := &AbilityRewardUseData{}
	bookType := &ObjectType{
		s:           &srv.Types,
		ind:         1,
		id:          rewardAbilityBookType4F0C70,
		flags:       object.FlagNoCollide,
		UseData:     UseDataPtr{Ptr: unsafe.Pointer(templateUse)},
		UseDataSize: unsafe.Sizeof(*templateUse),
	}
	srv.Types.byID = map[string]*ObjectType{"abilitybook": bookType}
	srv.Types.byInd = []*ObjectType{nil, bookType}
	data := &RewardMarkerInitData{RewardFlags: rewardAbilityBookExplicitFlag4F0C70}
	data.Abilities[5] = 1
	marker := &Object{InitData: unsafe.Pointer(data)}

	beforeRNG := srv.Rand.Logic.Index()
	got := srv.RewardAbilityBook4F0C70(marker)
	if got == nil {
		t.Fatal("server object factory returned nil")
	}
	if got.TypeInd != 1 || got.UseDataAbilityReward().Ability != 5 {
		t.Fatalf("server result type/ability = %d/%d, want 1/5", got.TypeInd, got.UseDataAbilityReward().Ability)
	}
	if index, want := srv.Rand.Logic.Index(), (beforeRNG+1)%4096; index != want {
		t.Fatalf("logic RNG index = %d, want %d after one explicit-selection draw", index, want)
	}
	got.UseData.Free()
}

func TestRewardAbilityBookNativeFaultBoundaries4F0C70(t *testing.T) {
	t.Run("nil marker faults before dependencies", func(t *testing.T) {
		calls := 0
		defer func() {
			fault := recover()
			if fault == nil || calls != 0 {
				t.Fatalf("nil marker recover/calls = %v/%d, want panic/0", fault, calls)
			}
		}()
		rewardAbilityBookNative4F0C70(nil, rewardAbilityBookNativeDeps4F0C70{
			randomInt:          func(int32, int32) int32 { calls++; return 1 },
			createObjectByType: func(string) *Object { calls++; return nil },
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
		rewardAbilityBookNative4F0C70(&Object{}, rewardAbilityBookNativeDeps4F0C70{
			randomInt:          func(int32, int32) int32 { calls++; return 1 },
			createObjectByType: func(string) *Object { calls++; return nil },
		})
	})

	t.Run("nil UseData faults after create", func(t *testing.T) {
		data := &RewardMarkerInitData{RewardFlags: rewardAbilityBookExplicitFlag4F0C70}
		data.Abilities[5] = 1
		created := &Object{}
		createdCalls := 0
		defer func() {
			fault := recover()
			if fault == nil || createdCalls != 1 {
				t.Fatalf("nil UseData recover/create calls = %v/%d, want panic/1", fault, createdCalls)
			}
		}()
		rewardAbilityBookNative4F0C70(&Object{InitData: unsafe.Pointer(data)}, rewardAbilityBookNativeDeps4F0C70{
			randomInt: func(int32, int32) int32 { return 0 },
			createObjectByType: func(string) *Object {
				createdCalls++
				return created
			},
		})
	})
}
