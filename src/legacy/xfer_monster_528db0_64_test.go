//go:build amd64 || arm64

package legacy

import (
	"encoding/binary"
	"path/filepath"
	"testing"
	"unsafe"

	"github.com/opennox/libs/object"
	"github.com/opennox/libs/spell"

	noxflags "github.com/opennox/opennox/v1/common/flags"
	"github.com/opennox/opennox/v1/internal/cryptfile"
	"github.com/opennox/opennox/v1/server"
)

func TestMonsterXferTail528DB0DormantHealth(t *testing.T) {
	oldFlags := noxflags.GetGame()
	t.Cleanup(func() {
		noxflags.ResetGame()
		noxflags.SetGame(oldFlags)
	})
	for _, tc := range []struct {
		name    string
		host    bool
		dormant byte
		wantCur uint16
		wantMax uint16
	}{
		{name: "host dormant NPC", host: true, dormant: 1},
		{name: "host nonzero dormant flag", host: true, dormant: 2},
		{name: "host live monster", host: true, wantCur: 11, wantMax: 75},
		{name: "client dormant NPC", dormant: 1, wantCur: 11, wantMax: 75},
	} {
		t.Run(tc.name, func(t *testing.T) {
			noxflags.ResetGame()
			if tc.host {
				noxflags.SetGame(noxflags.GameHost)
			}

			// Synthetic version-64 tail: dormant byte, field0, subclass,
			// current health, four definition bytes, empty buffs and poison.
			// Max and previous health are deliberately absent from the wire.
			data := []byte{tc.dormant}
			data = binary.LittleEndian.AppendUint32(data, 0x12345678)
			data = binary.LittleEndian.AppendUint32(data, 0)
			data = binary.LittleEndian.AppendUint16(data, 11)
			data = append(data, 2, 3, 4, 5)
			data = binary.LittleEndian.AppendUint16(data, 2)
			data = append(data, 0, 0, 0x5a)
			path := filepath.Join(t.TempDir(), "monster-tail.bin")
			cf, err := cryptfile.OpenFile(path, cryptfile.WriteOnly, -1)
			if err != nil {
				t.Fatal(err)
			}
			if _, err := cf.Write(data); err != nil {
				_ = cf.Close()
				t.Fatal(err)
			}
			if err := cf.Close(); err != nil {
				t.Fatal(err)
			}
			cf, err = cryptfile.OpenFile(path, cryptfile.ReadOnly, -1)
			if err != nil {
				t.Fatal(err)
			}
			defer cf.Close()

			update := new(server.MonsterUpdateData)
			health := &server.HealthData{Cur: 99, Field2: 22, Max: 75}
			unit := &server.Object{ObjClass: object.ClassMonster, UpdateData: unsafe.Pointer(update), HealthData: health}
			if err := monsterXferTail528DB0(cf, unit, 64); err != nil {
				t.Fatal(err)
			}
			if *health != (server.HealthData{Cur: tc.wantCur, Field2: 22, Max: tc.wantMax}) {
				t.Fatalf("health = %v, want Cur=%d Field2=22 Max=%d", health, tc.wantCur, tc.wantMax)
			}
			if update.Field361 != uint32(tc.dormant)<<8|4 || update.Field0 != 0x12345678 {
				t.Fatalf("definition fields = %#x/%#x", update.Field361, update.Field0)
			}
			if sentinel, err := cf.ReadU8(); err != nil || sentinel != 0x5a {
				t.Fatalf("tail alignment sentinel = %#x, %v", sentinel, err)
			}
		})
	}
}

func TestMonsterParseSpellID528DB0(t *testing.T) {
	tests := []struct {
		name       string
		value      string
		allowEmpty bool
		want       spell.ID
		wantErr    bool
	}{
		{name: "known", value: "SPELL_BLINK", want: spell.SPELL_BLINK},
		{name: "serialized invalid sentinel", value: "SPELL_INVALID", want: spell.SPELL_INVALID},
		{name: "optional empty", allowEmpty: true, want: spell.SPELL_INVALID},
		{name: "required empty", wantErr: true},
		{name: "unknown", value: "SPELL_NOT_IN_NOX", wantErr: true},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got, err := monsterParseSpellID528DB0(tc.value, tc.allowEmpty)
			if (err != nil) != tc.wantErr {
				t.Fatalf("monsterParseSpellID528DB0(%q, %v) error = %v, wantErr %v", tc.value, tc.allowEmpty, err, tc.wantErr)
			}
			if got != tc.want {
				t.Fatalf("monsterParseSpellID528DB0(%q, %v) = %v, want %v", tc.value, tc.allowEmpty, got, tc.want)
			}
		})
	}
}

func TestMonsterParseAbilityID528DB0(t *testing.T) {
	for id, name := range server.AbilityNames {
		got, err := monsterParseAbilityID528DB0(name)
		if err != nil || got != server.Ability(id) {
			t.Fatalf("monsterParseAbilityID528DB0(%q) = %v, %v; want %v, nil", name, got, err, server.Ability(id))
		}
	}
	if _, err := monsterParseAbilityID528DB0("ABILITY_UNKNOWN"); err == nil {
		t.Fatal("unknown ability was accepted")
	}
}

func TestMonsterShopParam528DB0UsesItemXferKind(t *testing.T) {
	fieldGuide := &server.ObjectType{Xfer: Get_nox_xxx_XFerFieldGuide_4F6390()}
	abilityReward := &server.ObjectType{Xfer: Get_nox_xxx_XFerAbilityReward_4F6240()}
	spellReward := &server.ObjectType{}

	objectIndex := func(name string) (uint32, bool) {
		if name == "Wasp" {
			return 1331, true
		}
		return 0, false
	}
	objectName := func(index uint32) (string, bool) {
		if index == 1331 {
			return "Wasp", true
		}
		return "", false
	}

	tests := []struct {
		name      string
		typ       *server.ObjectType
		wire      string
		param     uint32
		wantParam uint32
		wantWire  string
	}{
		{name: "field guide creature", typ: fieldGuide, wire: "Wasp", param: 1331, wantParam: 1331, wantWire: "Wasp"},
		{name: "ability reward", typ: abilityReward, wire: server.AbilityHarpoon.String(), param: uint32(server.AbilityHarpoon), wantParam: uint32(server.AbilityHarpoon), wantWire: server.AbilityHarpoon.String()},
		{name: "spell reward", typ: spellReward, wire: spell.SPELL_BLINK.String(), param: uint32(spell.SPELL_BLINK), wantParam: uint32(spell.SPELL_BLINK), wantWire: spell.SPELL_BLINK.String()},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			gotParam, err := monsterParseShopParam528DB0(tc.typ, tc.wire, objectIndex)
			if err != nil || gotParam != tc.wantParam {
				t.Fatalf("parse = %d, %v; want %d, nil", gotParam, err, tc.wantParam)
			}
			gotWire, err := monsterShopParamName528DB0(tc.typ, tc.param, objectName)
			if err != nil || gotWire != tc.wantWire {
				t.Fatalf("write = %q, %v; want %q, nil", gotWire, err, tc.wantWire)
			}
		})
	}

	if got, err := monsterParseShopParam528DB0(fieldGuide, "", objectIndex); err != nil || got != 0 {
		t.Fatalf("empty optional parameter = %d, %v; want 0, nil", got, err)
	}
	if got, err := monsterShopParamName528DB0(spellReward, 0, objectName); err != nil || got != "" {
		t.Fatalf("zero optional parameter = %q, %v; want empty, nil", got, err)
	}
	if _, err := monsterParseShopParam528DB0(fieldGuide, "UnknownCreature", objectIndex); err == nil {
		t.Fatal("unknown field-guide creature was accepted")
	}
}

func TestMonsterShopModifierSlot528DB0RoundTrip(t *testing.T) {
	ids := map[string]int{"Material1": 0, "Replenishment4": 7}
	names := map[int]string{0: "Material1", 7: "Replenishment4"}
	modifierID := func(name string) int {
		if id, ok := ids[name]; ok {
			return id
		}
		return 0xff
	}
	modifierName := func(id int) (string, bool) {
		name, ok := names[id]
		return name, ok
	}

	for _, name := range []string{"Material1", "Replenishment4"} {
		slot := monsterParseShopModifierSlot528DB0(name, modifierID)
		if slot == 0 {
			t.Fatalf("modifier %q encoded as absent", name)
		}
		got, err := monsterShopModifierName528DB0(slot, modifierName)
		if err != nil || got != name {
			t.Fatalf("modifier %q round trip = %q, %v", name, got, err)
		}
	}
	if got := monsterParseShopModifierSlot528DB0("", modifierID); got != 0 {
		t.Fatalf("empty modifier encoded as %#x", got)
	}
	if got := monsterParseShopModifierSlot528DB0("Unknown", modifierID); got != 0 {
		t.Fatalf("unknown modifier encoded as %#x", got)
	}
	if got, err := monsterShopModifierName528DB0(0, modifierName); err != nil || got != "" {
		t.Fatalf("absent modifier decoded as %q, %v", got, err)
	}
	if _, err := monsterShopModifierName528DB0(0x100, modifierName); err == nil {
		t.Fatal("out-of-range modifier token was accepted")
	}
	if _, err := monsterShopModifierName528DB0(8, func(int) (string, bool) { return "", false }); err == nil {
		t.Fatal("unresolvable modifier ID was accepted")
	}
}
