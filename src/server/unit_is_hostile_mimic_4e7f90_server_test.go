package server

import (
	"testing"

	"github.com/opennox/libs/object"
)

func TestUnitIsHostileMimicNative4E7F90UsesFixedWidthFields(t *testing.T) {
	player := &Object{ObjClass: object.ClassPlayer, Field35: 0x11223344}
	mimic := &Object{TypeInd: 1, Field35: 0x55667788}
	owner := &Object{}
	cache := uint32(0x00010001)
	deps := unitIsHostileMimicNativeDeps4E7F90{
		loadMimicCache: func() uint32 { return cache },
		lookupType: func(string) uint32 {
			t.Fatal("nonzero cache performed lookup")
			return 0
		},
		storeMimicCache: func(uint32) { t.Fatal("nonzero cache performed store") },
		isEnemy:         func(*Object, *Object) int32 { return 0 },
		isQuest:         func() int32 { return 1 },
	}

	if got := unitIsHostileMimicNative4E7F90(player, mimic, deps); got != 1 {
		t.Fatalf("high cache versus uint16 type result = %d, want 1", got)
	}
	cache = 1
	if got := unitIsHostileMimicNative4E7F90(player, mimic, deps); got != 0 {
		t.Fatalf("unowned Quest Mimic result = %d, want 0", got)
	}
	mimic.ObjOwner = owner
	if got := unitIsHostileMimicNative4E7F90(player, mimic, deps); got != 1 {
		t.Fatalf("owned Quest Mimic result = %d, want 1", got)
	}
	mimic.ObjOwner = nil
	if player.ObjClass != object.ClassPlayer || player.Field35 != 0x11223344 ||
		mimic.TypeInd != 1 || mimic.ObjOwner != nil || mimic.Field35 != 0x55667788 {
		t.Fatal("native predicate mutated object fields")
	}
}

func TestServerIsHostileMimic4E7F90InitializesCacheBeforeNil(t *testing.T) {
	s := &Server{}
	s.Types.byID = map[string]*ObjectType{
		"mimic": {ind: 0x2468},
	}
	if s.IsHostileMimicXxx(nil, nil) {
		t.Fatal("nil objects reported hostile")
	}
	if s.Types.fast.mimic != 0x2468 {
		t.Fatalf("Mimic cache = %#x, want %#x", s.Types.fast.mimic, 0x2468)
	}
}
