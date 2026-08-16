package server

import (
	"reflect"
	"testing"
)

func TestItemDropRules53EBF0ExactOriginalRows(t *testing.T) {
	wantNames := []string{
		"BookOfOblivion", "OblivionHalberd", "OblivionHeart", "OblivionWierdling", "OblivionOrb",
		"AmuletOfClarity", "AmuletOfTeleportation", "BridgeGuardsBoots", "Spectacles", "MatildasCloak",
		"ThaviusStaff", "BlueOrbKeyOfTheLich", "RedOrbKeyOfTheLich", "SilverKey", "GoldKey",
		"SapphireKey", "RubyKey", "SponsorshipLetter", "MayorsScepter",
	}
	gotNames := make([]string, 0, len(itemDropRuleNames53EBF0))
	for i, row := range itemDropRuleNames53EBF0 {
		gotNames = append(gotNames, row.name)
		wantMask := uint32(1)
		if i == 11 || i == 12 {
			wantMask = 0
		}
		if row.mask != wantMask {
			t.Errorf("row %d %q mask = %d, want %d", i, row.name, row.mask, wantMask)
		}
	}
	if !reflect.DeepEqual(gotNames, wantNames) {
		t.Fatalf("names = %q, want %q", gotNames, wantNames)
	}
}

func TestItemDropRules53EC40InitializesOnceInRowOrder(t *testing.T) {
	var rules itemDropRules53EBF0
	var names []string
	rules.initialize53EC40(func(name string) uint32 {
		names = append(names, name)
		return uint32(len(names) + 100)
	})
	if !rules.initialized {
		t.Fatal("rules were not marked initialized")
	}
	for i := range rules.typeInd {
		if want := uint32(i + 101); rules.typeInd[i] != want {
			t.Errorf("type[%d] = %d, want %d", i, rules.typeInd[i], want)
		}
	}
	wantNames := make([]string, len(itemDropRuleNames53EBF0))
	for i, row := range itemDropRuleNames53EBF0 {
		wantNames[i] = row.name
	}
	if !reflect.DeepEqual(names, wantNames) {
		t.Fatalf("lookup order = %q, want %q", names, wantNames)
	}
	rules.initialize53EC40(func(string) uint32 {
		t.Fatal("initialized rules performed another lookup")
		return 0
	})
}

func TestItemDropRules53EBF0MembershipAndMaskAreSeparate(t *testing.T) {
	var rules itemDropRules53EBF0
	for i := range rules.typeInd {
		rules.typeInd[i] = uint32(i + 1)
	}
	rules.initialized = true

	for _, index := range []int{11, 12} {
		typeInd := uint16(index + 1)
		if got := rules.itemIsDroppable53EBF0(typeInd); got != 1 {
			t.Errorf("row %d membership = %d, want 1", index, got)
		}
		if got := rules.itemDropMask53EC80(typeInd, 1); got != 0 {
			t.Errorf("row %d mask = %d, want 0", index, got)
		}
	}
	if got := rules.itemDropMask53EC80(1, 1); got != 1 {
		t.Fatalf("first row mask = %d, want 1", got)
	}
	if got := rules.itemDropMask53EC80(1, 2); got != 0 {
		t.Fatalf("first row mask 2 = %d, want 0", got)
	}
	if got := rules.itemIsDroppable53EBF0(99); got != 0 {
		t.Fatalf("missing membership = %d, want 0", got)
	}
}

func TestItemDropRules53EBF0UsesFullLookupWordAndFirstMatch(t *testing.T) {
	var rules itemDropRules53EBF0
	rules.typeInd[0] = 0x10001
	rules.typeInd[1] = 1
	rules.typeInd[11] = 1
	rules.initialized = true
	if got := rules.itemIsDroppable53EBF0(1); got != 1 {
		t.Fatalf("membership = %d, want 1", got)
	}
	if got := rules.itemDropMask53EC80(1, 1); got != 1 {
		t.Fatalf("first matching mask = %d, want 1", got)
	}
}

func TestItemDropRules53EBF0NilObjectReturnsBeforeInitialization(t *testing.T) {
	s := &Server{}
	if got := s.ItemIsDroppable53EBF0(nil); got != 0 {
		t.Fatalf("membership = %d, want 0", got)
	}
	if got := s.ItemDropMask53EC80(nil, 1); got != 0 {
		t.Fatalf("mask = %d, want 0", got)
	}
	if s.itemDropRules.initialized {
		t.Fatal("nil object initialized the table")
	}
}
