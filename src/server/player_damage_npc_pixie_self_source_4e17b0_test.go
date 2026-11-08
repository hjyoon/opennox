package server

import (
	"testing"

	"github.com/opennox/libs/object"
)

// This is the reported runtime shape, with the existing case-11 tail exposed
// at DefaultDamage. It needs no source unit/update record or new original data.
// Run unchanged on the old admission first; the failure is retained by root.
func TestPlayerDamageNPCPixieSelfSourceImpact4E17B0Admission(t *testing.T) {
	target := damageMeleeUnitFixture4E17B0(t, false)
	target.ObjClass = object.Class(2097154)
	target.ObjSubClass = object.SubClass(69778)
	target.ObjFlags = object.Flags(16777732)
	pixie := &Object{TypeInd: 77, ObjClass: object.Class(1572865), ObjSubClass: object.SubClass(3), ObjFlags: object.Flags(16794116)}
	before := *pixie
	update := target.UpdateDataMonster()
	update.Field547, update.Field546 = 99, 88
	runtime := damageMeleeRuntimeFixture4E17B0(t)
	runtime.CoopMode = func() bool { return false }
	reason := ""
	runtime.Unsupported = func(why string, v, a, w *Object, d int32, typ object.DamageType) {
		reason = why
		if v != target || a != pixie || w != pixie || d != 8 || typ != object.DamageImpact {
			t.Fatal("changed unsupported report arguments")
		}
	}
	calls := 0
	runtime.DefaultDamage = func(v, a, w *Object, d int32, typ object.DamageType) bool {
		calls++
		if v != target || a != pixie || w != pixie || d != 8 || typ != object.DamageImpact || update.Field547 != 1 || update.Field546 != uint32(pixie.TypeInd) {
			t.Fatalf("shared case-11 tail args/marker: damage=%d marker=%d/%d", d, update.Field547, update.Field546)
		}
		return true
	}
	handled, result := PlayerDamageNative4E17B0(target, pixie, pixie, 8, object.DamageImpact, runtime)
	if !handled || !result || reason != "" || calls != 1 {
		t.Fatalf("reported self-sourced Pixie admission: handled=%t result=%t unsupported=%q DefaultDamage-calls=%d", handled, result, reason, calls)
	}
	if *pixie != before || pixie.UpdateData != nil || target.HealthData.Cur != 200 {
		t.Fatal("source/update or stubbed-HP boundary changed")
	}
}
