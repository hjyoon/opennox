package legacy

import (
	"testing"

	"github.com/opennox/libs/object"

	"github.com/opennox/opennox/v1/common/sound"
	"github.com/opennox/opennox/v1/server"
)

func TestMonsterDieDropSound54A390(t *testing.T) {
	tests := []struct {
		name      string
		item      *server.Object
		wantSound sound.ID
	}{
		{name: "nil"},
		{name: "metal armor", item: &server.Object{ObjClass: object.ClassArmor, Material: uint16(object.MaterialMetal)}, wantSound: 805},
		{name: "wood armor", item: &server.Object{ObjClass: object.ClassArmor, Material: uint16(object.MaterialWood)}, wantSound: 811},
		{name: "hide armor", item: &server.Object{ObjClass: object.ClassArmor, Material: uint16(object.MaterialAnimalHide)}, wantSound: 808},
		{name: "cloth helmet", item: &server.Object{ObjClass: object.ClassArmor, ObjSubClass: 0x20, Material: uint16(object.MaterialCloth)}, wantSound: 817},
		{name: "cloth armor", item: &server.Object{ObjClass: object.ClassArmor, Material: uint16(object.MaterialCloth)}, wantSound: 814},
		{name: "armor material precedence", item: &server.Object{ObjClass: object.ClassArmor, Material: uint16(object.MaterialMetal | object.MaterialWood)}, wantSound: 805},
		{name: "wand class precedence", item: &server.Object{ObjClass: object.ClassWand | object.ClassWeapon, Material: uint16(object.MaterialMetal)}, wantSound: 831},
		{name: "metal weapon", item: &server.Object{ObjClass: object.ClassWeapon, Material: uint16(object.MaterialMetal)}, wantSound: 843},
		{name: "wood weapon", item: &server.Object{ObjClass: object.ClassWeapon, Material: uint16(object.MaterialWood)}, wantSound: 845},
		{name: "unsupported class", item: &server.Object{Material: uint16(object.MaterialMetal)}},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if got := monsterDieDropSound54A390(test.item); got != test.wantSound {
				t.Fatalf("drop sound = %d, want %d", got, test.wantSound)
			}
		})
	}
}

func TestMonsterDieCallbackKind54A7D0CoversMonsterBinTable(t *testing.T) {
	tests := []struct {
		name string
		kind server.MonsterDieCallbackKind54A7D0
	}{
		{name: "SWORDSMANDIE", kind: server.MonsterDieCallbackSwordsman54A7D0},
		{name: "URCHINSHAMANDIE", kind: server.MonsterDieCallbackUrchinShaman54A850},
		{name: "ARCHERDIE", kind: server.MonsterDieCallbackArcher54A890},
		{name: "OGREDIE", kind: server.MonsterDieCallbackOgre54A900},
		{name: "OGREWARLORDDIE", kind: server.MonsterDieCallbackOgreWarlord54A950},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			fnc := monsterDieFunctions[test.name]
			if fnc == nil {
				t.Fatal("monster.bin DIE_FUNCTION has no C function pointer")
			}
			got, ok := monsterDieCallbackKind54A7D0(fnc)
			if !ok || got != test.kind {
				t.Fatalf("callback kind = %d/%t, want %d/true", got, ok, test.kind)
			}
		})
	}
	if _, ok := monsterDieCallbackKind54A7D0(nil); ok {
		t.Fatal("nil DIE_FUNCTION was accepted")
	}
}
