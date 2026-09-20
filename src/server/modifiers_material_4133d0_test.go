package server

import (
	"testing"
	"unsafe"

	"github.com/opennox/libs/object"
)

func TestItemHasMaterial7Modifier4133D0(t *testing.T) {
	material7Name := []byte("Material7\x00")
	otherName := []byte("Material6\x00")
	material7 := &ModifierEff{name0: &material7Name[0]}
	other := &ModifierEff{name0: &otherName[0]}
	data := &ModifierInitData{}
	item := &Object{ObjClass: object.ClassArmor, InitData: unsafe.Pointer(data)}

	data.Modifiers[1] = material7
	if !ItemHasMaterial7Modifier4133D0(item) {
		t.Fatal("Material7 in the material slot was not detected")
	}
	data.Modifiers[1] = other
	if ItemHasMaterial7Modifier4133D0(item) {
		t.Fatal("a different material was detected as Material7")
	}
	data.Modifiers[0], data.Modifiers[1] = material7, nil
	if ItemHasMaterial7Modifier4133D0(item) {
		t.Fatal("Material7 outside the material slot was detected")
	}
	item.ObjClass = object.ClassMonster
	data.Modifiers[1] = material7
	if ItemHasMaterial7Modifier4133D0(item) {
		t.Fatal("an ineligible object class was detected")
	}
	if ItemHasMaterial7Modifier4133D0(nil) {
		t.Fatal("nil item was detected")
	}
}
