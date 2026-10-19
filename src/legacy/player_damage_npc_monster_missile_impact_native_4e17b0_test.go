package legacy

import (
	"fmt"
	"math"
	"testing"
	"unsafe"

	"github.com/opennox/libs/object"
	"github.com/opennox/libs/things"
	"github.com/opennox/libs/types"

	"github.com/opennox/opennox/v1/legacy/common/alloc"
	"github.com/opennox/opennox/v1/server"
)

// C-owned native-width records traverse the real C damage dispatcher and
// registered PlayerDamage -> DefaultDamage -> UnitSetHP. No replacement HP,
// damage, attribution or armor callback supplies the result.
func TestPlayerDamageNPCMonsterSimpleMissileImpactNative4E17B0CRecords(t *testing.T) {
	srv := npcReflectLegacyServer4E17B0(t)
	srv.Audio.Init(srv)
	t.Cleanup(srv.Audio.Free)
	srv.FreeObjectTypes()
	t.Cleanup(srv.FreeObjectTypes)
	if err := srv.Types.ReadObjectType(&things.Thing{Name: "NativeStoneImpactNPC", OnDamage: &things.ProcFunc{Name: "PlayerDamage"}}); err != nil {
		t.Fatal(err)
	}
	if !srv.Objs.Init(100) {
		t.Fatal("native stone allocator")
	}
	t.Cleanup(srv.Objs.FreeObjects)
	for _, absorption := range []float32{0, .25, 1} {
		for _, raw := range []int32{1, 3, 19} {
			t.Run(fmt.Sprintf("armor-%g/raw-%d", absorption, raw), func(t *testing.T) {
				v := srv.Objs.NewObject(&server.ObjectType{Damage: srv.Types.ByID("NativeStoneImpactNPC").Damage})
				a, w := srv.Objs.NewObject(&server.ObjectType{}), srv.Objs.NewObject(&server.ObjectType{})
				u, freeU := alloc.New(server.MonsterUpdateData{})
				au, freeAU := alloc.New(server.MonsterUpdateData{})
				hp, freeHP := alloc.New(server.HealthData{})
				t.Cleanup(freeU)
				t.Cleanup(freeAU)
				t.Cleanup(freeHP)
				*u = server.MonsterUpdateData{Field518: math.Float32bits(absorption), Field1: math.Float32bits(.125), Field547: 99, Field546: 77, Field523_2: 88}
				*au = server.MonsterUpdateData{Field523_2: 66, Field1: math.Float32bits(.375)}
				*hp = server.HealthData{Cur: 2000, Field2: 2000, Max: 2000}
				v.TypeInd, v.ObjClass, v.ObjSubClass, v.Material = 71, object.ClassMonster|object.ClassComplex, 0x11012, 0x4000
				v.UpdateData, v.HealthData, v.Frame134 = unsafe.Pointer(u), hp, 77
				a.TypeInd, a.ObjClass, a.ObjSubClass, a.UpdateData = 1339, object.ClassMonster|object.ClassComplex, 0x70001, unsafe.Pointer(au)
				a.PrevPos = types.Ptf(101, 103)
				w.TypeInd, w.ObjClass, w.ObjSubClass, w.ObjOwner = 525, object.ClassSimple|object.ClassMissile, 0, a
				w.PrevPos = types.Ptf(17, 19)
				for _, p := range []unsafe.Pointer{unsafe.Pointer(v), unsafe.Pointer(a), unsafe.Pointer(w), unsafe.Pointer(u), unsafe.Pointer(au), unsafe.Pointer(hp)} {
					if unsafe.Sizeof(uintptr(0)) == 8 && uintptr(p) <= math.MaxUint32 {
						t.Fatalf("native record=%p, want >4GiB", p)
					}
				}
				beforeA, beforeAU, beforeW := *a, *au, *w
				accumulated := float32((1-float64(absorption))*float64(raw)) + float32(.125)
				rounded := int32(math.RoundToEven(float64(accumulated)))
				effective := max(rounded, 1)
				if !objectDamageDispatchCallNative(v, a, w, raw, object.DamageImpact) || hp.Cur != uint16(2000-effective) || hp.Field2 != 2000 || hp.Max != 2000 ||
					u.Field1 != math.Float32bits(accumulated-float32(rounded)) || u.Field518 != math.Float32bits(absorption) || u.Field547 != 1 || u.Field546 != 525 || u.Field523_2 != 88 ||
					!u.StatusFlags.Has(object.MonStatusInjured) || v.Obj130 != w || v.Pos132 != w.PrevPos || v.Field131 != 11 || v.Frame134 != 1400 ||
					*a != beforeA || *au != beforeAU || *w != beforeW {
					t.Fatalf("C->NPC stone raw=%d armor=%g HP=%d effective=%d carry=%g marker=%d/%d source/projectile unchanged=%t/%t/%t", raw, absorption, hp.Cur, effective, math.Float32frombits(u.Field1), u.Field547, u.Field546, *a == beforeA, *au == beforeAU, *w == beforeW)
				}
			})
		}
	}
}
