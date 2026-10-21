package opennox

import (
	"fmt"
	"math"
	"testing"
	"unsafe"

	"github.com/opennox/libs/object"
	"github.com/opennox/libs/strman"
	"github.com/opennox/libs/things"
	"github.com/opennox/libs/types"

	noxflags "github.com/opennox/opennox/v1/common/flags"
	"github.com/opennox/opennox/v1/legacy"
	"github.com/opennox/opennox/v1/legacy/common/alloc"
	"github.com/opennox/opennox/v1/server"
)

// The reported class/type/flags and source == weapon cross the registered C
// dispatcher with C-owned records above 4 GiB. Damage, HP and hurt-state
// services are production bindings; this is not stock-map campaign evidence.
func TestPlayerDamageSelfMissileFlameNative4E17B0(t *testing.T) {
	base := server.New(nil, nil, strman.New())
	t.Cleanup(base.Close)
	s := &Server{Server: base}
	oldServer, oldGetServer := noxServer, legacy.GetServer
	noxServer, legacy.GetServer = s, func() legacy.Server { return s }
	t.Cleanup(func() { noxServer, legacy.GetServer = oldServer, oldGetServer })
	oldGame, oldEngine, oldGameplay := noxflags.GetGame(), noxflags.GetEngine(), noxflags.GetGamePlay()
	noxflags.UnsetGame(oldGame)
	noxflags.UnsetEngine(oldEngine)
	noxflags.UnsetGamePlay(oldGameplay)
	noxflags.SetGamePlay(noxflags.GameplayFlag1)
	t.Cleanup(func() {
		noxflags.UnsetGame(noxflags.GetGame())
		noxflags.UnsetEngine(noxflags.GetEngine())
		noxflags.UnsetGamePlay(noxflags.GetGamePlay())
		noxflags.SetGame(oldGame)
		noxflags.SetEngine(oldEngine)
		noxflags.SetGamePlay(oldGameplay)
	})
	s.SetFrame(1400)
	s.Audio.Init(s.Server)
	t.Cleanup(s.Audio.Free)
	s.FreeObjectTypes()
	t.Cleanup(s.FreeObjectTypes)
	if err := s.Types.ReadObjectType(&things.Thing{Name: "NativeSelfMissileFlame", OnDamage: &things.ProcFunc{Name: "PlayerDamage"}}); err != nil {
		t.Fatal(err)
	}
	if !s.Objs.Init(32) {
		t.Fatal("native object allocator")
	}
	t.Cleanup(s.Objs.FreeObjects)
	for _, raw := range []int{9, 96} {
		for _, god := range []bool{false, true} {
			t.Run(fmt.Sprintf("raw-%d/god-%t", raw, god), func(t *testing.T) {
				noxflags.UnsetEngine(noxflags.EngineGodMode)
				if god {
					noxflags.SetEngine(noxflags.EngineGodMode)
				}
				v := s.Objs.NewObject(&server.ObjectType{Damage: s.Types.ByID("NativeSelfMissileFlame").Damage})
				w := s.Objs.NewObject(&server.ObjectType{})
				u, freeU := alloc.New(server.PlayerUpdateData{})
				p, freeP := alloc.New(server.Player{})
				h, freeH := alloc.New(server.HealthData{})
				for _, free := range []func(){freeU, freeP, freeH} {
					t.Cleanup(free)
				}
				*p = server.Player{PlayerInd: 7}
				*u = server.PlayerUpdateData{Player: p, State: server.PlayerState13, Field76: 99, Field75: 77, Field21: math.Float32bits(.4)}
				*h = server.HealthData{Cur: 400, Max: 400, Field2: 400}
				v.TypeInd, v.ObjClass, v.ObjSubClass, v.ObjFlags = 713, 2621444, 0, 16777732
				v.UpdateData, v.HealthData, v.Material = unsafe.Pointer(u), h, 0x4000
				w.TypeInd, w.ObjClass, w.ObjSubClass, w.ObjFlags = 694, 2621441, 1, 553665028
				w.PosVec, w.PrevPos = types.Ptf(-20, 0), types.Ptf(-24, 7)
				for _, ptr := range []unsafe.Pointer{unsafe.Pointer(v), unsafe.Pointer(w), unsafe.Pointer(u), unsafe.Pointer(p), unsafe.Pointer(h)} {
					if unsafe.Sizeof(uintptr(0)) == 8 && uintptr(ptr) <= math.MaxUint32 {
						t.Fatalf("native C record %p is not above 4 GiB", ptr)
					}
				}
				before := *w
				wantHP := uint16(400 - raw)
				if god {
					wantHP = 400
				}
				wantState := server.PlayerState13
				if !god && raw >= 20 {
					wantState = server.PlayerState30
				}
				if !v.CallDamage(w, w, raw, object.DamageFlame) || h.Cur != wantHP || u.Field76 != 2 || u.Field75 != 1 ||
					u.Field21 != math.Float32bits(.4) || u.State != wantState || *w != before {
					t.Fatalf("native self-FLAME: HP=%d/%d marker=%d/%d carry=%g state=%d", h.Cur, wantHP, u.Field76, u.Field75, math.Float32frombits(u.Field21), u.State)
				}
				if !god && (v.Obj130 != w || v.Pos132 != w.PrevPos || v.Field131 != 1 || v.Frame134 != 1400 || v.Field38 != math.MaxUint32) {
					t.Fatal("native self-FLAME attribution/HP synchronization")
				}
				t.Logf("C-owned >4 GiB PlayerDamage -> DefaultDamage -> UnitSetHP: player=%p StrongFireball source=weapon=%p raw=%d HP=400->%d GodMode=%t", v, w, raw, h.Cur, god)
			})
		}
	}
}
