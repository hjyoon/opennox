//go:build amd64 || arm64

package legacy

import (
	"testing"

	"github.com/opennox/libs/spell"

	"github.com/opennox/opennox/v1/server"
)

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
