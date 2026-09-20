package legacy

import (
	"testing"
	"unsafe"

	"github.com/opennox/opennox/v1/server"
)

func TestItemPreDamageCanApplyNative4E13B0KnownEffects(t *testing.T) {
	fns := itemPreDamageFunctionsNative4E13B0()
	tests := []struct {
		name string
		fnc  unsafe.Pointer
	}{
		{name: "drain mana", fnc: fns.drainMana},
		{name: "vampirism", fnc: fns.vampirism},
		{name: "poison", fnc: fns.poison},
		{name: "panic", fnc: fns.panic},
		{name: "sympathy", fnc: fns.sympathy},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			effect := &server.ModifierEff{AttackPreDmg64: server.ModifierEffFnc{Fnc: test.fnc}}
			if !itemPreDamageCanApplyNative4E13B0(effect) {
				t.Fatal("known callback was rejected")
			}
		})
	}
	if !itemPreDamageCanApplyNative4E13B0(nil) ||
		!itemPreDamageCanApplyNative4E13B0(&server.ModifierEff{}) {
		t.Fatal("nil or inert modifier was rejected")
	}
}

func TestItemPreDamageCanApplyNative4E13B0RejectsUnknownOnWideHost(t *testing.T) {
	if unsafe.Sizeof(uintptr(0)) == 4 {
		t.Skip("PE32 can invoke arbitrary legacy modifier callbacks")
	}
	effect := &server.ModifierEff{
		AttackPreDmg64: server.ModifierEffFnc{Fnc: unsafe.Pointer(new(byte))},
	}
	if itemPreDamageCanApplyNative4E13B0(effect) {
		t.Fatal("unknown callback accepted on wide host")
	}
}
