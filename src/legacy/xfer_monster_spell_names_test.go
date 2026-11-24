package legacy

import (
	"encoding/binary"
	"os"
	"path/filepath"
	"testing"
	"unsafe"

	"github.com/opennox/libs/spell"
	"github.com/opennox/opennox/v1/internal/cryptfile"
	"github.com/opennox/opennox/v1/server"
)

func TestMonsterXferNamedSpellsKeepOriginalInvalidSlot528DB0(t *testing.T) {
	wire := binary.LittleEndian.AppendUint32(nil, 3)
	for _, item := range []struct {
		name string
		word uint32
	}{
		{"SPELL_SUMMON_Lich", 0x65432100}, // Name from G_Oasis.map.
		{"SPELL_SUMMON_LICH", 0x65432101},
		{"SPELL_UNKNOWN", 0x65432102}, // The last unknown name updates slot zero.
	} {
		wire = append(wire, byte(len(item.name)))
		wire = append(wire, item.name...)
		wire = binary.LittleEndian.AppendUint32(wire, item.word)
	}
	wire = append(wire, make([]byte, 8)...)
	for _, name := range []string{"SPELL_SUMMON_Lich", "SPELL_SUMMON_LICH", ""} {
		wire = append(wire, byte(len(name)))
		wire = append(wire, name...)
	}
	wire = binary.LittleEndian.AppendUint32(wire, 0xdeadbeef)
	path := filepath.Join(t.TempDir(), "named-spells.bin")
	if err := os.WriteFile(path, wire, 0o600); err != nil {
		t.Fatal(err)
	}
	cf, err := cryptfile.OpenFile(path, cryptfile.ReadOnly, -1)
	if err != nil {
		t.Fatal(err)
	}
	defer cf.Close()

	var data server.MonsterUpdateData
	if err := monsterXferDefinitionSpells528DB0(cf, &data, 64); err != nil {
		t.Fatal(err)
	}
	slots := unsafe.Slice(&data.Field372, server.SpellsMax)
	if slots[0] != 0x65432102 || slots[spell.SPELL_SUMMON_LICH] != 0x65432101 {
		t.Fatalf("spell slots zero=%#x lich=%#x", slots[0], slots[spell.SPELL_SUMMON_LICH])
	}
	if err := monsterXferDefinitionTimings528DB0(cf, &data, 64, false); err != nil {
		t.Fatal(err)
	}
	if data.Field511 != 0 || data.Field512 != uint32(spell.SPELL_SUMMON_LICH) || data.Field513 != 0 {
		t.Fatalf("automatic spell IDs = %d, %d, %d", data.Field511, data.Field512, data.Field513)
	}
	if marker, err := monsterRWU32(cf, 0); err != nil || marker != 0xdeadbeef {
		t.Fatalf("trailing DWORD = %#x, err = %v", marker, err)
	}
}
