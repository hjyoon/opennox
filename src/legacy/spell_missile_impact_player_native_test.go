package legacy

import (
	"math"
	"testing"
	"unsafe"

	"github.com/opennox/libs/object"
	"github.com/opennox/libs/things"
	"github.com/opennox/libs/types"

	"github.com/opennox/opennox/v1/legacy/common/alloc"
	"github.com/opennox/opennox/v1/server"
)

// All records crossing the actual C dispatcher are C-owned, including both
// native-width update records, player info and health. No HP/damage service
// or registered production callback is substituted by these tests.
func spellMissileImpactCRecords4E17B0(t *testing.T, srv *server.Server, owner string) (target, source, missile *server.Object) {
	t.Helper()
	target, missile = srv.Objs.NewObject(&server.ObjectType{}), srv.Objs.NewObject(&server.ObjectType{})
	update, freeUpdate := alloc.New(server.PlayerUpdateData{})
	player, freePlayer := alloc.New(server.Player{})
	health, freeHealth := alloc.New(server.HealthData{})
	*player = server.Player{PlayerInd: 7}
	*update = server.PlayerUpdateData{Player: player, State: server.PlayerState13}
	*health = server.HealthData{Cur: 60, Max: 60, Field2: 60}
	target.ObjClass, target.ObjFlags, target.Material = object.ClassPlayer, 0, 0x4000
	target.UpdateData, target.HealthData = unsafe.Pointer(update), health
	missile.TypeInd, missile.ObjClass, missile.ObjSubClass, missile.ObjFlags = 991, object.Class(0x180001), 3, 0
	missile.PrevPos, missile.PosVec = types.Ptf(17, 9), types.Ptf(19, 11)
	source = missile
	pointers := []unsafe.Pointer{unsafe.Pointer(target), unsafe.Pointer(missile), unsafe.Pointer(update), unsafe.Pointer(player), unsafe.Pointer(health)}
	for _, free := range []func(){freeUpdate, freePlayer, freeHealth} {
		t.Cleanup(free)
	}
	if owner != "self" {
		source = srv.Objs.NewObject(&server.ObjectType{})
		source.ObjFlags, source.Material = 0, 0x4000
		if owner == "player" {
			p, freeP := alloc.New(server.Player{})
			u, freeU := alloc.New(server.PlayerUpdateData{})
			*p = server.Player{PlayerInd: 8}
			*u = server.PlayerUpdateData{Player: p, State: server.PlayerState13}
			source.ObjClass, source.UpdateData = object.ClassPlayer, unsafe.Pointer(u)
			pointers = append(pointers, unsafe.Pointer(p), unsafe.Pointer(u))
			t.Cleanup(freeP)
			t.Cleanup(freeU)
		} else {
			u, freeU := alloc.New(server.MonsterUpdateData{})
			source.ObjClass, source.ObjSubClass, source.UpdateData = object.ClassMonster, 0x10002, unsafe.Pointer(u)
			pointers = append(pointers, unsafe.Pointer(u))
			t.Cleanup(freeU)
		}
		missile.ObjOwner = source
		pointers = append(pointers, unsafe.Pointer(source))
	}
	for _, ptr := range pointers {
		if unsafe.Sizeof(uintptr(0)) == 8 && uintptr(ptr) <= math.MaxUint32 {
			t.Fatalf("C-owned pointer=%p, want >4 GiB", ptr)
		}
	}
	return
}

func TestDefaultDamagePlayerSpellMissileImpactCRecords4E0B30(t *testing.T) {
	srv := npcReflectLegacyServer4E17B0(t)
	srv.FreeObjectTypes()
	t.Cleanup(srv.FreeObjectTypes)
	if err := srv.Types.ReadObjectType(&things.Thing{Name: "NativeSpellMissileImpactDefault", OnDamage: &things.ProcFunc{Name: "DefaultDamage"}}); err != nil {
		t.Fatal(err)
	}
	if !srv.Objs.Init(16) {
		t.Fatal("native object allocator")
	}
	t.Cleanup(srv.Objs.FreeObjects)
	for _, owner := range []string{"self", "player", "NPC"} {
		t.Run(owner, func(t *testing.T) {
			v, a, w := spellMissileImpactCRecords4E17B0(t, srv, owner)
			v.Damage = srv.Types.ByID("NativeSpellMissileImpactDefault").Damage
			beforeMissile := *w
			for hit, hp := range []uint16{52, 44, 36} {
				if !objectDamageDispatchCallNative(v, a, w, 8, object.DamageImpact) || v.HealthData.Cur != hp ||
					v.Obj130 != w || v.Field131 != 11 || v.Frame134 != srv.Frame() || v.Pos132 != w.PrevPos ||
					v.Field38 != math.MaxUint32 || v.UpdateDataPlayer().State != server.PlayerState13 || *w != beforeMissile {
					t.Fatalf("C entry hit=%d HP=%d want=%d source=%p weapon=%p", hit, v.HealthData.Cur, hp, a, w)
				}
			}
			t.Logf("C-owned >4 GiB C->DefaultDamage: player=%p terminal-parent=%p Pixie=%p owner=%s IMPACT=8 HP=60->52->44->36", v, a, w, owner)
		})
	}
}
