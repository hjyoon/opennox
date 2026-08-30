package server

import (
	"reflect"
	"testing"
	"unsafe"

	"github.com/opennox/libs/object"
)

func TestPlayerCantCastSpellNative4FD150Layouts(t *testing.T) {
	wantType := uintptr(4)
	wantClass := uintptr(8)
	wantFlags := uintptr(16)
	wantInventoryNext := uintptr(496)
	wantInventoryFirst := uintptr(504)
	wantOwnedNext := uintptr(512)
	wantOwnedFirst := uintptr(516)
	wantUpdate := uintptr(748)
	wantPlayer := uintptr(276)
	wantSpellLevels := uintptr(3696)
	if unsafe.Sizeof(uintptr(0)) == 8 {
		wantType = 8
		wantClass = 12
		wantFlags = 20
		wantInventoryNext = 528
		wantInventoryFirst = 544
		wantOwnedNext = 560
		wantOwnedFirst = 568
		wantUpdate = 872
		wantPlayer = 336
		wantSpellLevels = 4992
	}

	for _, check := range []struct {
		name string
		got  uintptr
		want uintptr
	}{
		{"Object.TypeInd", unsafe.Offsetof(Object{}.TypeInd), wantType},
		{"Object.ObjClass", unsafe.Offsetof(Object{}.ObjClass), wantClass},
		{"Object.ObjFlags", unsafe.Offsetof(Object{}.ObjFlags), wantFlags},
		{"Object.InvNextItem", unsafe.Offsetof(Object{}.InvNextItem), wantInventoryNext},
		{"Object.InvFirstItem", unsafe.Offsetof(Object{}.InvFirstItem), wantInventoryFirst},
		{"Object.Field128", unsafe.Offsetof(Object{}.Field128), wantOwnedNext},
		{"Object.Field129", unsafe.Offsetof(Object{}.Field129), wantOwnedFirst},
		{"Object.UpdateData", unsafe.Offsetof(Object{}.UpdateData), wantUpdate},
		{"PlayerUpdateData.Player", unsafe.Offsetof(PlayerUpdateData{}.Player), wantPlayer},
		{"Player.SpellLvl", unsafe.Offsetof(Player{}.SpellLvl), wantSpellLevels},
		{"Player.SpellLvl element", unsafe.Sizeof(Player{}.SpellLvl[0]), 4},
	} {
		if check.got != check.want {
			t.Errorf("%s = %d, want %d", check.name, check.got, check.want)
		}
	}
}

func TestPlayerSpellPowerNative4FE7B0PreservesPointerState(t *testing.T) {
	cache := uint32(0)
	var events []string
	gameFlags := uint32(0)
	deps := playerSpellPowerNativeDeps4FE7B0{
		loadImaginaryCasterType: func() uint32 {
			events = append(events, "load-cache")
			return cache
		},
		storeImaginaryCasterType: func(value uint32) {
			events = append(events, "store-cache")
			cache = value
		},
		lookupObjectType: func(name string) uint32 {
			events = append(events, "lookup:"+name)
			return 77
		},
		hasGameFlag: func(mask uint32) int32 {
			events = append(events, "game")
			if gameFlags&mask != 0 {
				return 1
			}
			return 0
		},
	}

	imaginary := &Object{TypeInd: 77}
	if got := playerSpellPowerNative4FE7B0(50, imaginary, deps); got != 1 {
		t.Fatalf("imaginary caster power = %d, want 1", got)
	}
	if want := []string{"load-cache", "lookup:ImaginaryCaster", "store-cache"}; !reflect.DeepEqual(events, want) {
		t.Fatalf("imaginary events = %v, want %v", events, want)
	}

	events = nil
	gameFlags = playerSpellPowerFullStrengthModes4FE7B0
	if got := playerSpellPowerNative4FE7B0(50, nil, deps); got != 3 {
		t.Fatalf("forced game power = %d, want 3", got)
	}
	if want := []string{"load-cache", "game"}; !reflect.DeepEqual(events, want) {
		t.Fatalf("forced game events = %v, want %v", events, want)
	}

	gameFlags = 0
	player := &Player{}
	player.SpellLvl[50] = 0x76543210
	playerUpdate := &PlayerUpdateData{Player: player}
	playerObject := &Object{ObjClass: object.ClassPlayer, UpdateData: unsafe.Pointer(playerUpdate)}
	if got := playerSpellPowerNative4FE7B0(50, playerObject, deps); got != 0x76543210 {
		t.Fatalf("player power = %#x, want 0x76543210", got)
	}

	monsterUpdate := &MonsterUpdateData{Field510: 0x12345678}
	monsterObject := &Object{ObjClass: object.ClassMonster, UpdateData: unsafe.Pointer(monsterUpdate)}
	if got := playerSpellPowerNative4FE7B0(50, monsterObject, deps); got != 0x12345678 {
		t.Fatalf("monster power = %#x, want 0x12345678", got)
	}
	if got := playerSpellPowerNative4FE7B0(50, &Object{}, deps); got != 3 {
		t.Fatalf("non-player power = %d, want 3", got)
	}
	if got := playerSpellPowerNative4FE7B0(50, nil, deps); got != 2 {
		t.Fatalf("nil power = %d, want 2", got)
	}
}

func TestPlayerCantCastSpellNative4FD150TraversesNativeLinks(t *testing.T) {
	globals := make(map[uintptr]uint32)
	gameFlags := playerCantCastModeKOTR4FD150
	var events []string
	deps := playerCantCastSpellNativeDeps4FD150{
		hasGameFlag: func(mask uint32) int32 {
			events = append(events, "game")
			if gameFlags&mask != 0 {
				return 1
			}
			return 0
		},
		loadGlobal: func(offset uintptr) uint32 {
			events = append(events, "load")
			return globals[offset]
		},
		storeGlobal: func(offset uintptr, value uint32) {
			events = append(events, "store")
			globals[offset] = value
		},
		lookupObjectType: func(name string) uint32 {
			events = append(events, "lookup:"+name)
			return 0x1234
		},
		spellHasFlags: func(int32, uint32) int32 {
			events = append(events, "spell-flags")
			return 1
		},
		spellPower: func(int32, *Object) int32 {
			t.Fatal("KOTR path reached spell power")
			return 0
		},
		balanceFloat: func(string, int32) float64 {
			t.Fatal("KOTR path reached balance")
			return 0
		},
	}

	owned := &Object{TypeInd: 0x1234}
	unit := &Object{Field129: owned, TeamVal: ObjectTeam{ID: 9}}
	if got := playerCantCastSpellNative4FD150(unit, 22, 0, deps); got != 17 {
		t.Fatalf("KOTR result = %d, want 17", got)
	}
	if globals[playerCantCastCrownTypeOffset4FD150] != 0x1234 {
		t.Fatalf("crown cache = %#x, want 0x1234", globals[playerCantCastCrownTypeOffset4FD150])
	}
	if unit.Field129 != owned || owned.Field128 != nil {
		t.Fatal("owned native links were mutated")
	}

	gameFlags = playerCantCastModeCTF4FD150
	events = nil
	first := &Object{}
	flag := &Object{ObjFlags: object.Flags(playerCantCastInventoryFlag4FD150)}
	first.InvNextItem = flag
	unit = &Object{InvFirstItem: first}
	if got := playerCantCastSpellNative4FD150(unit, 23, 0, deps); got != 13 {
		t.Fatalf("CTF result = %d, want 13", got)
	}
	if unit.InvFirstItem != first || first.InvNextItem != flag {
		t.Fatal("inventory native links were mutated")
	}
}

func TestPlayerCantCastSpellNative4FD150CountsOwnedSummons(t *testing.T) {
	globals := map[uintptr]uint32{playerCantCastMagicMissileOffset4FD150: 0x4321}
	owned := &Object{TypeInd: 0x4321}
	unit := &Object{Field129: owned}
	deps := playerCantCastSpellNativeDeps4FD150{
		hasGameFlag:      func(uint32) int32 { return 0 },
		loadGlobal:       func(offset uintptr) uint32 { return globals[offset] },
		storeGlobal:      func(offset uintptr, value uint32) { globals[offset] = value },
		lookupObjectType: func(string) uint32 { return 0 },
		spellHasFlags:    func(int32, uint32) int32 { return 0 },
		spellPower: func(spellID int32, got *Object) int32 {
			if spellID != 50 || got != unit {
				t.Fatalf("spell power args = %d/%p, want 50/%p", spellID, got, unit)
			}
			return 2
		},
		balanceFloat: func(key string, index int32) float64 {
			if key != "MagicMissileCount" || index != 1 {
				t.Fatalf("balance args = %q/%d", key, index)
			}
			return 1
		},
	}

	if got := playerCantCastSpellNative4FD150(unit, 50, 1, deps); got != 3 {
		t.Fatalf("missile result = %d, want 3", got)
	}
	owned.ObjFlags |= object.FlagDestroyed
	if got := playerCantCastSpellNative4FD150(unit, 50, 1, deps); got != 0 {
		t.Fatalf("destroyed missile result = %d, want 0", got)
	}
}
