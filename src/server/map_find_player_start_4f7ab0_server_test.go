package server

import (
	"math"
	"runtime"
	"testing"
	"unsafe"

	"github.com/opennox/libs/object"
	"github.com/opennox/libs/prand"
	"github.com/opennox/libs/types"
)

func TestMapFindPlayerStartNative4F7AB0Layout(t *testing.T) {
	wantTypeIndex := uintptr(4)
	wantFlags := uintptr(16)
	wantTeam := uintptr(48)
	wantPosition := uintptr(56)
	if unsafe.Sizeof(uintptr(0)) == 8 {
		wantTypeIndex = 8
		wantFlags = 20
		wantTeam = 52
		wantPosition = 60
	}
	checks := []struct {
		name string
		got  uintptr
		want uintptr
	}{
		{"Object.TypeInd", unsafe.Offsetof(Object{}.TypeInd), wantTypeIndex},
		{"Object.ObjFlags", unsafe.Offsetof(Object{}.ObjFlags), wantFlags},
		{"Object.TeamVal", unsafe.Offsetof(Object{}.TeamVal), wantTeam},
		{"Object.PosVec", unsafe.Offsetof(Object{}.PosVec), wantPosition},
		{"ObjectTeam.ID", unsafe.Offsetof(ObjectTeam{}.ID), 4},
		{"ObjectTeam size", unsafe.Sizeof(ObjectTeam{}), 8},
		{"Pointf size", unsafe.Sizeof(types.Pointf{}), 8},
	}
	for _, check := range checks {
		if check.got != check.want {
			t.Errorf("%s = %d, want %d", check.name, check.got, check.want)
		}
	}
}

func TestMapFindPlayerStartNative4F7AB0PreservesHighPointers(t *testing.T) {
	player := new(Object)
	start := new(Object)
	output := new(types.Pointf)
	start.TypeInd = 9
	start.ObjFlags = object.FlagEnabled
	start.PosVec = types.Pointf{X: 123.5, Y: -77.25}

	var pin runtime.Pinner
	pin.Pin(player)
	pin.Pin(start)
	pin.Pin(output)
	defer pin.Unpin()
	if unsafe.Sizeof(uintptr(0)) == 8 {
		for name, pointer := range map[string]uintptr{
			"player": uintptr(unsafe.Pointer(player)),
			"start":  uintptr(unsafe.Pointer(start)),
			"output": uintptr(unsafe.Pointer(output)),
		} {
			if pointer <= math.MaxUint32 {
				t.Fatalf("%s pointer = %#x, want native address above 4 GiB", name, pointer)
			}
		}
	}

	cache := uint32(9)
	mapFindPlayerStartNative4F7AB0(player, output, mapFindPlayerStartNativeDeps4F7AB0{
		loadCachedType:  func() uint32 { return cache },
		lookupType:      func(string) uint32 { return 0 },
		storeCachedType: func(value uint32) { cache = value },
		touchTeam:       func(uint8) {},
		firstObject:     func() *Object { return start },
		firstPlayer:     func() *Object { return nil },
		nextPlayer:      func(*Object) *Object { return nil },
		isEnemyTo:       func(*Object, *Object) bool { return false },
		teamContains:    func(*Object, uint8) bool { return false },
		randomInt:       func(int32, int32, string, int32) int32 { return 0 },
	})
	if *output != start.PosVec {
		t.Fatalf("output = %+v, want %+v", *output, start.PosVec)
	}
	runtime.KeepAlive(player)
	runtime.KeepAlive(start)
	runtime.KeepAlive(output)
}

func TestServerMapFindPlayerStart4F7AB0Integration(t *testing.T) {
	s := new(Server)
	s.Rand.Logic = prand.New(0)
	s.Types.byID = map[string]*ObjectType{
		"playerstart": {ind: 9},
	}
	player := new(Object)
	start := &Object{
		TypeInd:  9,
		ObjFlags: object.FlagEnabled,
		PosVec:   types.Pointf{X: 43.25, Y: -91.5},
	}
	s.Objs.List = start

	got := s.MapFindPlayerStart4F7AB0(player)
	if got != start.PosVec || s.Types.fast.playerStart4F7AB0 != 9 {
		t.Fatalf("result = %+v cache=%d, want %+v cache=9", got, s.Types.fast.playerStart4F7AB0, start.PosVec)
	}
}
