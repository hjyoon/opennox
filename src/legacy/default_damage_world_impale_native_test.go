package legacy

import (
	"testing"

	"github.com/opennox/libs/object"
	"github.com/opennox/libs/things"
)

// Real registered DefaultDamage callback and UnitSetHP, all records C-owned.
func TestDefaultDamagePlayerWorldImpaleCRecords4E0B30(t *testing.T) {
	srv := npcReflectLegacyServer4E17B0(t)
	srv.FreeObjectTypes()
	t.Cleanup(srv.FreeObjectTypes)
	if err := srv.Types.ReadObjectType(&things.Thing{Name: "NativeWorldImpaleDefault", OnDamage: &things.ProcFunc{Name: "DefaultDamage"}}); err != nil {
		t.Fatal(err)
	}
	if !srv.Objs.Init(16) {
		t.Fatal("native object allocator")
	}
	t.Cleanup(srv.Objs.FreeObjects)
	for _, owner := range []string{"self", "player", "NPC"} {
		t.Run(owner, func(t *testing.T) {
			v, a, w := spellMissileImpactCRecords4E17B0(t, srv, owner)
			w.ObjClass, w.ObjSubClass = object.ClassDangerous|object.ClassImmobile|object.ClassVisibleEnable, 0
			v.Damage = srv.Types.ByID("NativeWorldImpaleDefault").Damage
			for _, hp := range []uint16{57, 54, 51} {
				if !objectDamageDispatchCallNative(v, a, w, 3, object.DamageImpale) || v.HealthData.Cur != hp ||
					v.Obj130 != w || v.Field131 != 3 || v.Pos132 != w.PrevPos || v.Frame134 != 1400 {
					t.Fatalf("world IMPALE HP=%d want=%d source=%p hazard=%p", v.HealthData.Cur, hp, a, w)
				}
			}
		})
	}
}
