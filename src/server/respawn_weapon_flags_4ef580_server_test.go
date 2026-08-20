package server

import "testing"

func newRespawnWeaponFlagsServer4EF580(allowed uint8) *Server {
	s := &Server{}
	s.Types.byInd = make([]*ObjectType, 9)
	for i := 1; i <= 8; i++ {
		s.Types.byInd[i] = &ObjectType{ind: uint16(i), allowed: allowed&(1<<uint(i-1)) != 0}
	}

	s.Armor.table[0] = armorRecord{Bit: 0x400, TypeInd: 1}
	s.Armor.table[1] = armorRecord{Bit: 0x4, TypeInd: 2}
	s.Armor.table[2] = armorRecord{Bit: 0x1, TypeInd: 3}
	s.Weapons.table[0] = weaponRecord{Bit: 0x8000, TypeInd: 4}
	s.Armor.table[3] = armorRecord{Bit: 0x4000, TypeInd: 5}
	s.Weapons.table[1] = weaponRecord{Bit: 0x100, TypeInd: 6}
	s.Weapons.table[2] = weaponRecord{Bit: 0x200, TypeInd: 7}
	s.Armor.table[4] = armorRecord{Bit: 0x1000000, TypeInd: 8}
	return s
}

func TestRespawnWeaponFlags4EF580NativeTablesExhaustive(t *testing.T) {
	for want := 0; want <= 0xff; want++ {
		s := newRespawnWeaponFlagsServer4EF580(uint8(want))
		if got := s.RespawnWeaponFlags4EF580(); got != uint8(want) {
			t.Fatalf("allowed %#02x: flags = %#02x", want, got)
		}
	}
}

func TestRespawnWeaponFlags4EF580NativeReadsLiveTypesAndIndices(t *testing.T) {
	s := newRespawnWeaponFlagsServer4EF580(0)
	s.Types.byInd[1].allowed = true
	s.Types.byInd[8].allowed = true
	if got := s.RespawnWeaponFlags4EF580(); got != 0x81 {
		t.Fatalf("initial flags = %#02x, want 0x81", got)
	}

	s.Types.byInd[1].allowed = false
	s.Types.byInd[2].allowed = true
	s.Armor.table[4].TypeInd = 3
	s.Types.byInd[3].allowed = true
	if got := s.RespawnWeaponFlags4EF580(); got != 0x86 {
		t.Fatalf("mutated flags = %#02x, want 0x86", got)
	}
}

func TestRespawnWeaponFlags4EF580NativeMissingTypePreservesFault(t *testing.T) {
	s := newRespawnWeaponFlagsServer4EF580(0xff)
	s.Weapons.table[1].TypeInd = 0
	defer func() {
		if recover() == nil {
			t.Fatal("missing type did not fault")
		}
	}()
	s.RespawnWeaponFlags4EF580()
}
