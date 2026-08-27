package server

import (
	"math"
	"reflect"
	"testing"
	"unsafe"

	"github.com/opennox/libs/prand"
)

func TestRewardReplenishNativeLayout4F2210(t *testing.T) {
	wantObjectSize := uintptr(780)
	wantTypeInd := uintptr(4)
	wantNext := uintptr(444)
	wantInitData := uintptr(692)
	if unsafe.Sizeof(uintptr(0)) == 8 {
		wantObjectSize = 928
		wantTypeInd = 8
		wantNext = 448
		wantInitData = 760
	}
	checks := []struct {
		name string
		got  uintptr
		want uintptr
	}{
		{"Object size", unsafe.Sizeof(Object{}), wantObjectSize},
		{"Object.TypeInd", unsafe.Offsetof(Object{}.TypeInd), wantTypeInd},
		{"Object.ObjNext", unsafe.Offsetof(Object{}.ObjNext), wantNext},
		{"Object.InitData", unsafe.Offsetof(Object{}.InitData), wantInitData},
		{"RewardMarkerInitData size", unsafe.Sizeof(RewardMarkerInitData{}), 220},
		{"RewardMarkerInitData.Field216", unsafe.Offsetof(RewardMarkerInitData{}.Field216), 216},
	}
	for _, check := range checks {
		if check.got != check.want {
			t.Errorf("%s = %d, want %d", check.name, check.got, check.want)
		}
	}
}

func TestRewardReplenishNativeKeepsHighPointersAndLowByteWrites4F2210(t *testing.T) {
	inactiveData := &RewardMarkerInitData{Field216: 0x12345620}
	fixedData := &RewardMarkerInitData{Field216: 0xaabbcc01}
	plusData := &RewardMarkerInitData{Field216: 0xddeeff02}
	inactive := &Object{TypeInd: 10, InitData: unsafe.Pointer(inactiveData)}
	fixed := &Object{TypeInd: 10, InitData: unsafe.Pointer(fixedData)}
	plus := &Object{TypeInd: 20, InitData: unsafe.Pointer(plusData)}
	p1 := &Object{TypeInd: 30}
	p2 := &Object{TypeInd: 30}
	inactive.ObjNext = fixed
	fixed.ObjNext = plus
	plus.ObjNext = p1
	p1.ObjNext = p2
	if unsafe.Sizeof(uintptr(0)) == 8 {
		for name, pointer := range map[string]uintptr{
			"inactive": uintptr(unsafe.Pointer(inactive)),
			"fixed":    uintptr(unsafe.Pointer(fixed)),
			"plus":     uintptr(unsafe.Pointer(plus)),
			"p1":       uintptr(unsafe.Pointer(p1)),
			"p2":       uintptr(unsafe.Pointer(p2)),
			"data":     uintptr(unsafe.Pointer(inactiveData)),
		} {
			if pointer <= math.MaxUint32 {
				t.Fatalf("%s pointer does not exercise native width: %#x", name, pointer)
			}
		}
	}

	markerCache, plusCache, potionCache := uint32(10), uint32(20), uint32(30)
	firstCalls := 0
	var randomLines []int32
	var deleted []*Object
	rewardReplenishNative4F2210(rewardReplenishNativeDeps4F2210{
		loadMarkerCache:  func() uint32 { return markerCache },
		storeMarkerCache: func(value uint32) { markerCache = value },
		loadPlusCache:    func() uint32 { return plusCache },
		storePlusCache:   func(value uint32) { plusCache = value },
		loadPotionCache:  func() uint32 { return potionCache },
		storePotionCache: func(value uint32) { potionCache = value },
		lookupType:       func(string) uint32 { t.Fatal("cached type was looked up"); return 0 },
		firstObject: func() *Object {
			firstCalls++
			return inactive
		},
		randomInt: func(minimum, maximum int32, path string, line int32) int32 {
			if minimum != 0 || maximum != 1 || path != rewardReplenishRandomPath4F2210 {
				t.Fatalf("random args = %d/%d/%q/%d", minimum, maximum, path, line)
			}
			randomLines = append(randomLines, line)
			return 0
		},
		runtime: RewardReplenishRuntime4F2210{
			QuestStage:  func() uint32 { return 1 },
			PlayerCount: func() int32 { return 99 },
			DelayedDelete: func(object *Object) {
				deleted = append(deleted, object)
			},
		},
	})
	if firstCalls != 2 || !reflect.DeepEqual(randomLines, []int32{2660}) ||
		!reflect.DeepEqual(deleted, []*Object{p1}) {
		t.Fatalf("first/random/deleted = %d/%v/%v, want 2/[2660]/p1", firstCalls, randomLines, deleted)
	}
	if inactiveData.Field216 != 0x123456a0 || fixedData.Field216 != 0xaabbcc81 ||
		plusData.Field216 != 0xddeeff82 {
		t.Fatalf("Field216 = %#08x/%#08x/%#08x, want upper bytes preserved",
			inactiveData.Field216, fixedData.Field216, plusData.Field216)
	}
}

func TestRewardReplenishServerUsesDedicatedCachesAndLogicRNG4F2210(t *testing.T) {
	s := new(Server)
	s.Rand.Logic = prand.New(2011)
	p1 := &Object{TypeInd: 30}
	p2 := &Object{TypeInd: 30}
	p1.ObjNext = p2
	s.Objs.List = p1

	s.Types.fast.rewardReplenishMarker = 10
	s.Types.fast.rewardReplenishMarkerPlus = 20
	s.Types.fast.rewardReplenishRedPotion = 30
	s.Types.fast.rewardAnkhMarker = 101
	s.Types.fast.rewardAnkhMarkerPlus = 102
	s.Types.fast.rewardContainerMarker = 103
	s.Types.fast.rewardContainerMarkerPlus = 104
	s.Types.fast.rewardMarkerPlus = 105

	wantRNG := prand.New(2011)
	selected := wantRNG.IntClamp(0, 1)
	wantDeleted := p1
	if selected == 1 {
		wantDeleted = p2
	}
	var deleted []*Object
	s.RewardReplenish4F2210(RewardReplenishRuntime4F2210{
		QuestStage:  func() uint32 { return 0 },
		PlayerCount: func() int32 { return 1 },
		DelayedDelete: func(object *Object) {
			deleted = append(deleted, object)
		},
	})
	if !reflect.DeepEqual(deleted, []*Object{wantDeleted}) || s.Rand.Logic.Index() != wantRNG.Index() {
		t.Fatalf("deleted/RNG index = %v/%d, want %p/%d",
			deleted, s.Rand.Logic.Index(), wantDeleted, wantRNG.Index())
	}
	if s.Types.fast.rewardReplenishMarker != 10 ||
		s.Types.fast.rewardReplenishMarkerPlus != 20 ||
		s.Types.fast.rewardReplenishRedPotion != 30 ||
		s.Types.fast.rewardAnkhMarker != 101 ||
		s.Types.fast.rewardAnkhMarkerPlus != 102 ||
		s.Types.fast.rewardContainerMarker != 103 ||
		s.Types.fast.rewardContainerMarkerPlus != 104 ||
		s.Types.fast.rewardMarkerPlus != 105 {
		t.Fatalf("dedicated caches changed: %+v", s.Types.fast)
	}
}
