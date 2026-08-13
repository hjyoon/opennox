package server

import "testing"

func TestModifierDescriptionPointersForItemName(t *testing.T) {
	definitionDesc := uint16(0x1111)
	modifierDesc := uint16(0x2222)
	modifierIdentity := uint16(0x3333)
	definition := &Modifier{Desc8: &definitionDesc}
	modifier := &ModifierEff{desc8: &modifierDesc, identdesc16: &modifierIdentity}

	if got := definition.Description16(); got != &definitionDesc {
		t.Fatalf("definition description = %p, want %p", got, &definitionDesc)
	}
	if got := modifier.Description16(); got != &modifierDesc {
		t.Fatalf("modifier description = %p, want %p", got, &modifierDesc)
	}
	if got := modifier.IdentificationDescription16(); got != &modifierIdentity {
		t.Fatalf("modifier identity description = %p, want %p", got, &modifierIdentity)
	}
}
