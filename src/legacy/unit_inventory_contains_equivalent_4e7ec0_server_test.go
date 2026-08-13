package legacy

import (
	"testing"
	"unsafe"

	"github.com/opennox/libs/object"
	"github.com/opennox/opennox/v1/server"
)

func TestInventoryContainsEquivalentNative4E7EC0(t *testing.T) {
	item := &server.Object{TypeInd: 17}
	if got := Sub_4E7EC0(nil, item); got != 0 {
		t.Fatalf("nil owner result = %d, want 0", got)
	}
	if got := Sub_4E7EC0(&server.Object{}, nil); got != 0 {
		t.Fatalf("nil item result = %d, want 0", got)
	}

	mods := [5]server.ModifierEff{}
	candidateAttrs := server.ModifierInitData{
		Modifiers: [4]*server.ModifierEff{&mods[0], &mods[1], &mods[2], &mods[3]},
	}
	itemAttrs := candidateAttrs
	itemAttrs.Modifiers[2] = &mods[4]
	item.InitData = unsafe.Pointer(&itemAttrs)

	match := &server.Object{TypeInd: 17}
	modifierMismatch := &server.Object{
		TypeInd: 17, ObjClass: object.ClassWeapon,
		InitData: unsafe.Pointer(&candidateAttrs), InvNextItem: match,
	}
	typeMismatch := &server.Object{TypeInd: 18, InvNextItem: modifierMismatch}
	owner := &server.Object{InvFirstItem: typeMismatch}

	if got := Sub_4E7EC0(owner, item); got != 1 {
		t.Fatalf("later ordinary match result = %d, want 1", got)
	}
	if owner.InvFirstItem != typeMismatch || typeMismatch.InvNextItem != modifierMismatch || modifierMismatch.InvNextItem != match {
		t.Fatal("native adapter mutated the inventory list")
	}

	match.TypeInd = 19
	if got := Sub_4E7EC0(owner, item); got != 0 {
		t.Fatalf("nonmatching inventory result = %d, want 0", got)
	}

	guideCandidateUse := [...]byte{'G', 'u', 'i', 'd', 'e', 0}
	guideItemUse := guideCandidateUse
	guide := &server.Object{
		TypeInd: 17, ObjClass: object.ClassInfoBook,
		ObjSubClass: object.SubClass(object.BookFieldGuide),
	}
	guide.UseData.Ptr = unsafe.Pointer(&guideCandidateUse[0])
	item.UseData.Ptr = unsafe.Pointer(&guideItemUse[0])
	owner.InvFirstItem = guide
	if got := Sub_4E7EC0(owner, item); got != 1 {
		t.Fatalf("FieldGuide match result = %d, want 1", got)
	}
}
