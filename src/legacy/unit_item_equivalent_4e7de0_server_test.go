package legacy

import (
	"testing"
	"unsafe"

	"github.com/opennox/libs/object"
	"github.com/opennox/opennox/v1/server"
)

func TestItemEquivalentNative4E7DE0(t *testing.T) {
	candidate := &server.Object{TypeInd: 0xffff}
	item := &server.Object{TypeInd: 0xffff, ObjClass: object.ClassInfoBook, ObjSubClass: ^object.SubClass(0)}
	if got := Sub_4E7DE0(nil, item); got != 0 {
		t.Fatalf("nil candidate result = %d, want 0", got)
	}
	if got := Sub_4E7DE0(candidate, nil); got != 0 {
		t.Fatalf("nil item result = %d, want 0", got)
	}
	if got := Sub_4E7DE0(candidate, &server.Object{TypeInd: 0xfffe}); got != 0 {
		t.Fatalf("type mismatch result = %d, want 0", got)
	}
	if got := Sub_4E7DE0(candidate, item); got != 1 {
		t.Fatalf("ordinary result = %d, want 1", got)
	}

	mods := [4]server.ModifierEff{}
	candidateAttrs := server.ModifierInitData{
		Modifiers: [4]*server.ModifierEff{&mods[0], &mods[1], &mods[2], &mods[3]},
	}
	itemAttrs := candidateAttrs
	candidate.ObjClass = object.ClassWeapon
	candidate.InitData = unsafe.Pointer(&candidateAttrs)
	item.InitData = unsafe.Pointer(&itemAttrs)
	if got := Sub_4E7DE0(candidate, item); got != 1 {
		t.Fatalf("modifier match result = %d, want 1", got)
	}
	itemAttrs.Modifiers[3] = &mods[0]
	if got := Sub_4E7DE0(candidate, item); got != 0 {
		t.Fatalf("modifier mismatch result = %d, want 0", got)
	}

	candidateUse := [...]byte{'G', 'u', 'i', 'd', 'e', 0, 'x'}
	itemUse := [...]byte{'G', 'u', 'i', 'd', 'e', 0, 'y'}
	candidate.ObjClass = object.ClassInfoBook
	candidate.ObjSubClass = object.SubClass(object.BookFieldGuide)
	candidate.UseData.Ptr = unsafe.Pointer(&candidateUse[0])
	item.UseData.Ptr = unsafe.Pointer(&itemUse[0])
	if got := Sub_4E7DE0(candidate, item); got != 1 {
		t.Fatalf("FieldGuide match result = %d, want 1", got)
	}
	itemUse[4] = 'X'
	if got := Sub_4E7DE0(candidate, item); got != 0 {
		t.Fatalf("FieldGuide mismatch result = %d, want 0", got)
	}

	candidate.ObjSubClass = object.SubClass(1 | object.BookFieldGuide)
	candidateUse[0], itemUse[0] = 77, 77
	if got := Sub_4E7DE0(candidate, item); got != 1 {
		t.Fatalf("SpellBook precedence result = %d, want 1", got)
	}
}
