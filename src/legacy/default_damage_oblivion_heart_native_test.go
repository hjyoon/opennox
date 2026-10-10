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

// Exact reported stock classes/subclasses/flags with fresh C-owned records.
// C dispatch, registered DefaultDamage, protection, effects and UnitSetHP are
// production services, not test replacements. This is not natural GUI combat.
func TestDefaultDamageOblivionHeartNative4E0B30CRecords(t *testing.T) {
	srv := npcReflectLegacyServer4E17B0(t)
	srv.Audio.Init(srv)
	t.Cleanup(srv.Audio.Free)
	srv.FreeObjectTypes()
	t.Cleanup(srv.FreeObjectTypes)
	if err := srv.Types.ReadObjectType(&things.Thing{Name: "NativeOblivionHeartDefault", OnDamage: &things.ProcFunc{Name: "DefaultDamage"}}); err != nil {
		t.Fatal(err)
	}
	if !srv.Objs.Init(64) {
		t.Fatal("C-owned object allocator")
	}
	t.Cleanup(srv.Objs.FreeObjects)
	for _, tc := range []struct {
		name            string
		typeID          uint16
		subclass, flags uint32
	}{
		{"Troll", 2273, 66050, 151519748},
		{"Urchin", 1339, 458753, 17302164},
		{"SmallAlbinoSpider", 1356, 131073, 17302164},
	} {
		for _, defense := range []string{"none", "invulnerable", "electric-immune"} {
			for _, typ := range []object.DamageType{object.DamageElectric, object.DamageAirborneElectric} {
				t.Run(fmt.Sprintf("%s/%s/type-%d", tc.name, defense, typ), func(t *testing.T) {
					v, a, w := srv.Objs.NewObject(&server.ObjectType{}), srv.Objs.NewObject(&server.ObjectType{}), srv.Objs.NewObject(&server.ObjectType{})
					u, freeU := alloc.New(server.MonsterUpdateData{})
					au, freeAU := alloc.New(server.PlayerUpdateData{})
					info, freeInfo := alloc.New(server.Player{})
					hp, freeHP := alloc.New(server.HealthData{})
					init, freeInit := alloc.New(server.ModifierInitData{})
					for _, free := range []func(){freeU, freeAU, freeInfo, freeHP, freeInit} {
						t.Cleanup(free)
					}
					*u = server.MonsterUpdateData{Field547: 99, Field546: 77, Field523_2: 88, Field1: math.Float32bits(0.125)}
					*hp = server.HealthData{Cur: 2000, Max: 2000, Field2: 2000}
					*info = server.Player{PlayerInd: 9}
					*au = server.PlayerUpdateData{Player: info, State: server.PlayerState13}
					v.TypeInd, v.ObjClass, v.ObjSubClass, v.ObjFlags, v.Material = tc.typeID, object.Class(2097154), object.SubClass(tc.subclass), object.Flags(tc.flags), 0x4000
					v.UpdateData, v.HealthData, v.Damage = unsafe.Pointer(u), hp, srv.Types.ByID("NativeOblivionHeartDefault").Damage
					a.TypeInd, a.ObjClass, a.ObjSubClass, a.ObjFlags, a.UpdateData = 713, object.Class(2621444), 0, object.Flags(16777732), unsafe.Pointer(au)
					a.PrevPos = types.Ptf(32.5, -18.25)
					w.TypeInd, w.ObjClass, w.ObjSubClass, w.ObjFlags, w.InitData = 223, object.Class(85458944), object.SubClass(16777216), object.Flags(16777488), unsafe.Pointer(init)
					w.PrevPos = types.Ptf(44, 7)
					// Unbuffed reported hits use the real damage/HP callbacks.
					// Shock's outer player-state boundary is tested independently.
					if defense == "invulnerable" {
						v.Buffs |= 1 << server.ENCHANT_INVULNERABLE
					}
					if defense == "electric-immune" {
						v.ObjSubClass |= 0x800
					}
					for _, p := range []unsafe.Pointer{unsafe.Pointer(v), unsafe.Pointer(a), unsafe.Pointer(w), unsafe.Pointer(u), unsafe.Pointer(au), unsafe.Pointer(info), unsafe.Pointer(hp), unsafe.Pointer(init)} {
						if unsafe.Sizeof(uintptr(0)) == 8 && uintptr(p) <= math.MaxUint32 {
							t.Fatalf("C-owned pointer=%p, want >4 GiB", p)
						}
					}
					beforeA, beforeW, beforeAU := *a, *w, *au
					var steps [3]uint16
					for hit := range 3 {
						want := uint16(2000)
						marker := uint32(0)
						if defense == "none" {
							want, marker = uint16(2000-12*(hit+1)), 1
						}
						if defense == "invulnerable" {
							marker = 99
						}
						if !objectDamageDispatchCallNative(v, a, w, 12, typ) || hp.Cur != want || hp.Max != 2000 || hp.Field2 != 2000 || u.Field547 != marker || u.Field1 != math.Float32bits(0.125) || *a != beforeA || *w != beforeW || *au != beforeAU || v.HasEnchant(server.ENCHANT_SHOCK) {
							t.Fatalf("actual C electric hit=%d HP=%d/%d marker=%d/%d type=%d", hit, hp.Cur, want, u.Field547, marker, typ)
						}
						if defense == "none" && (v.Obj130 != w || v.Pos132 != a.PrevPos || v.Field131 != uint32(typ) || v.Frame134 != 1400 || v.Field38 != math.MaxUint32 || u.Field546 != 223 || u.Field523_2 != 88 || !u.StatusFlags.Has(object.MonStatusInjured)) {
							t.Fatal("actual C ranged weapon attribution/position/state lost")
						}
						if defense != "none" && (v.Obj130 != nil || v.Frame134 != 0 || u.Field546 != 77 || u.Field523_2 != 88 || u.StatusFlags != 0) {
							t.Fatal("immune target acquired injury metadata")
						}
						steps[hit] = hp.Cur
					}
					t.Logf("C-owned >4 GiB C->registered DefaultDamage->UnitSetHP: target=%p source=%p weapon=%p name=%s class=2097154 subclass=%d flags=%d type=%d defense=%s HP=2000->%d->%d->%d", v, a, w, tc.name, tc.subclass, tc.flags, typ, defense, steps[0], steps[1], steps[2])
				})
			}
		}
	}
}
