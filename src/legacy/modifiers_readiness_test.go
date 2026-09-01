package legacy

import (
	"testing"
	"unsafe"

	"github.com/opennox/libs/object"

	"github.com/opennox/opennox/v1/server"
)

func TestItemCheckReadinessEffect4E0960UsesNativeModifierPointers(t *testing.T) {
	if got := Nox_xxx_itemCheckReadinessEffect_4E0960(nil); got != 0 {
		t.Fatalf("nil item result = %d, want 0", got)
	}

	wrongFunction := new(byte)
	ignored := &server.ModifierEff{Attack40: server.ModifierEffFnc{
		Fnc: ReadinessEffectPointer4E0960(),
		Val: 11,
	}}
	notReadiness := &server.ModifierEff{Attack40: server.ModifierEffFnc{
		Fnc: unsafe.Pointer(wrongFunction),
		Val: 22,
	}}
	readiness := &server.ModifierEff{Attack40: server.ModifierEffFnc{
		Fnc: ReadinessEffectPointer4E0960(),
		Val: 37,
	}}
	attrs := server.ModifierInitData{Modifiers: [4]*server.ModifierEff{
		ignored,
		nil,
		notReadiness,
		readiness,
	}}
	item := &server.Object{
		ObjClass: object.ClassWand,
		InitData: unsafe.Pointer(&attrs),
	}
	if got := Nox_xxx_itemCheckReadinessEffect_4E0960(item); got != 37 {
		t.Fatalf("readiness result = %d, want 37", got)
	}

	item.ObjClass = object.ClassPlayer
	if got := Nox_xxx_itemCheckReadinessEffect_4E0960(item); got != 0 {
		t.Fatalf("unsupported class result = %d, want 0", got)
	}
}
