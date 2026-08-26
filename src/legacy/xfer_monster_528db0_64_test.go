//go:build amd64 || arm64

package legacy

import (
	"testing"

	"github.com/opennox/libs/spell"
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
