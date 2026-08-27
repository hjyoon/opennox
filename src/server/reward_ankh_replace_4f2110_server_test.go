package server

import (
	"math"
	"reflect"
	"testing"
	"unsafe"

	"github.com/opennox/libs/prand"
	"github.com/opennox/libs/types"
)

func TestRewardAnkhNativeLayout4F2110(t *testing.T) {
	wantObjectSize := uintptr(780)
	wantTypeInd := uintptr(4)
	wantPosition := uintptr(56)
	wantNext := uintptr(444)
	wantInitData := uintptr(692)
	if unsafe.Sizeof(uintptr(0)) == 8 {
		wantObjectSize = 928
		wantTypeInd = 8
		wantPosition = 60
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
		{"Object.PosVec", unsafe.Offsetof(Object{}.PosVec), wantPosition},
		{"Object.ObjNext", unsafe.Offsetof(Object{}.ObjNext), wantNext},
		{"Object.InitData", unsafe.Offsetof(Object{}.InitData), wantInitData},
		{"RewardMarkerInitData size", unsafe.Sizeof(RewardMarkerInitData{}), 220},
		{"RewardMarkerInitData.CategoryMask", unsafe.Offsetof(RewardMarkerInitData{}.CategoryMask), 0},
	}
	for _, check := range checks {
		if check.got != check.want {
			t.Errorf("%s = %d, want %d", check.name, check.got, check.want)
		}
	}
}

func TestRewardAnkhNativeKeepsHighPointersAndFirstDataByte4F2110(t *testing.T) {
	data := &RewardMarkerInitData{CategoryMask: 0x12345680}
	marker := &Object{
		TypeInd:  7,
		InitData: unsafe.Pointer(data),
		PosVec:   types.Pointf{X: 51, Y: 52},
	}
	ankh := new(Object)
	if unsafe.Sizeof(uintptr(0)) == 8 {
		if uintptr(unsafe.Pointer(marker)) <= math.MaxUint32 || uintptr(marker.InitData) <= math.MaxUint32 ||
			uintptr(unsafe.Pointer(ankh)) <= math.MaxUint32 {
			t.Fatalf("object/data pointers do not exercise native width: %#x/%#x/%#x",
				uintptr(unsafe.Pointer(marker)), uintptr(marker.InitData), uintptr(unsafe.Pointer(ankh)))
		}
	}

	markerCache := uint32(7)
	plusCache := uint32(8)
	firstCalls := 0
	var randomArgs []int32
	var deleted []*Object
	rewardAnkhReplaceNative4F2110(rewardAnkhReplaceNativeDeps4F2110{
		loadMarkerCache:  func() uint32 { return markerCache },
		storeMarkerCache: func(value uint32) { markerCache = value },
		loadPlusCache:    func() uint32 { return plusCache },
		storePlusCache:   func(value uint32) { plusCache = value },
		lookupType:       func(string) uint32 { t.Fatal("cached types were looked up"); return 0 },
		firstObject: func() *Object {
			firstCalls++
			return marker
		},
		randomInt: func(minimum, maximum int32, path string, line int32) int32 {
			randomArgs = []int32{minimum, maximum, line}
			if path != rewardAnkhRandomPath4F2110 {
				t.Fatalf("random path = %q", path)
			}
			return 0
		},
		newObject: func(name string) *Object {
			if name != rewardAnkhObjectTypeName4F2110 {
				t.Fatalf("object name = %q", name)
			}
			return ankh
		},
		runtime: RewardAnkhReplaceRuntime4F2110{
			CreateAt: func(object, owner *Object, point types.Pointf) {
				if object != ankh || owner != nil {
					t.Fatalf("create object/owner = %p/%p", object, owner)
				}
				object.PosVec = point
			},
			DelayedDelete: func(object *Object) {
				deleted = append(deleted, object)
			},
		},
	})
	if firstCalls != 2 || !reflect.DeepEqual(randomArgs, []int32{0, 0, 2460}) ||
		!reflect.DeepEqual(deleted, []*Object{marker}) {
		t.Fatalf("first/random/deleted = %d/%v/%v", firstCalls, randomArgs, deleted)
	}
	if ankh.PosVec != marker.PosVec || markerCache != 7 || plusCache != 8 {
		t.Fatalf("position/caches = %+v/%d/%d, want %+v/7/8", ankh.PosVec, markerCache, plusCache, marker.PosVec)
	}
}

func TestRewardAnkhServerUsesDedicatedCachesAndEmptyRangeDoesNotStepRNG4F2110(t *testing.T) {
	s := new(Server)
	s.Rand.Logic = prand.New(37)
	s.Types.fast.rewardAnkhMarker = 17
	s.Types.fast.rewardAnkhMarkerPlus = 19
	s.Types.fast.rewardMarkerPlus = 23
	s.Types.fast.rewardContainerMarker = 29
	s.Types.fast.rewardContainerMarkerPlus = 31
	before := s.Rand.Logic.Index()
	s.RewardAnkhReplace4F2110(RewardAnkhReplaceRuntime4F2110{
		CreateAt:      func(*Object, *Object, types.Pointf) { t.Fatal("empty world created an object") },
		DelayedDelete: func(*Object) { t.Fatal("empty world deleted an object") },
	})
	if got := s.Rand.Logic.Index(); got != before {
		t.Fatalf("empty 0..-1 RNG advanced index %d -> %d", before, got)
	}
	if s.Types.fast.rewardAnkhMarker != 17 || s.Types.fast.rewardAnkhMarkerPlus != 19 ||
		s.Types.fast.rewardMarkerPlus != 23 || s.Types.fast.rewardContainerMarker != 29 ||
		s.Types.fast.rewardContainerMarkerPlus != 31 {
		t.Fatalf("dedicated caches changed: ankh %d/%d activation %d container %d/%d",
			s.Types.fast.rewardAnkhMarker, s.Types.fast.rewardAnkhMarkerPlus,
			s.Types.fast.rewardMarkerPlus, s.Types.fast.rewardContainerMarker,
			s.Types.fast.rewardContainerMarkerPlus)
	}
}
