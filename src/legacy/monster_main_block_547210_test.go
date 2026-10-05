package legacy

import (
	"math"
	"testing"
	"unsafe"

	"github.com/opennox/libs/object"
	"github.com/opennox/libs/types"
	noxflags "github.com/opennox/opennox/v1/common/flags"
	"github.com/opennox/opennox/v1/common/memmap"
	"github.com/opennox/opennox/v1/common/unit/ai"
	"github.com/opennox/opennox/v1/legacy/common/alloc"
	"github.com/opennox/opennox/v1/server"
)

func monsterMainBlockLegacyFixture547210(t *testing.T) (*server.Server, *server.Object, *server.MonsterUpdateData, *server.Object) {
	t.Helper()
	srv, unit, update := monsterLookAtFixture5125A0(t)
	oldServer, oldFlags := GetServer, noxflags.GetGame()
	GetServer = func() Server { return &monsterMainLegacyServer547210{srv: srv} }
	noxflags.ResetGame()
	t.Cleanup(func() {
		GetServer = oldServer
		noxflags.ResetGame()
		noxflags.SetGame(oldFlags)
	})
	srv.Map.Init()
	if srv.Walls.Init() == 0 {
		t.Fatal("cannot initialize native empty wall map")
	}
	t.Cleanup(srv.Walls.Free)
	srv.SetFrame(1400)
	srv.SetTickRate(31)
	unit.ObjFlags, unit.SpeedBase = object.FlagActive|object.FlagEnabled, 2
	unit.PosVec, unit.NewPos = types.Ptf(300, 300), types.Ptf(300, 300)
	unit.Direction1, unit.Direction2 = 0, 0
	// The isolated legacy test binary does not load GAME.EXE's data blob.
	// Initialize only the facing-zero table entry to its original 1/0 vector
	// and restore it afterwards; no probe result or AI action is injected.
	directionX, directionY := memmap.PtrFloat32(0x587000, 194136), memmap.PtrFloat32(0x587000, 194140)
	oldX, oldY := *directionX, *directionY
	t.Logf("isolated direction table before initialization: %g/%g", oldX, oldY)
	*directionX, *directionY = server.SinCosDir(0)
	t.Cleanup(func() { *directionX, *directionY = oldX, oldY })
	health, freeHealth := alloc.New(server.HealthData{})
	t.Cleanup(freeHealth)
	*health = server.HealthData{Cur: 100, Max: 100}
	unit.HealthData = health
	*update = server.MonsterUpdateData{Aggression: 0.5, RetreatLevel: 0.25, AIStackInd: 0}
	update.AIStack[0].Action = uint32(ai.ACTION_FIGHT)
	missile, freeMissile := alloc.New(server.Object{})
	t.Cleanup(freeMissile)
	*missile = server.Object{
		ObjClass: object.ClassMissile, ObjFlags: object.FlagActive,
		PosVec: types.Ptf(340, 300), NewPos: types.Ptf(340, 300),
		PrevPos: types.Ptf(344, 300), VelVec: types.Ptf(-4, 0),
	}
	callback, _, ok := server.ObjectCollideHandler("DefaultCollide")
	if !ok || callback == nil {
		t.Fatal("stock collision eligibility handler is unavailable")
	}
	unit.Collide, missile.Collide = callback, callback
	t.Cleanup(srv.Map.Free)
	if unsafe.Sizeof(uintptr(0)) == 8 {
		for name, pointer := range map[string]unsafe.Pointer{
			"unit": unsafe.Pointer(unit), "update": unsafe.Pointer(update),
			"health": unsafe.Pointer(health), "missile": unsafe.Pointer(missile),
		} {
			if uintptr(pointer) <= math.MaxUint32 {
				t.Fatalf("%s pointer=%p, want actual native address above 4 GiB", name, pointer)
			}
		}
	}
	return srv, unit, update, missile
}

// Do not replace 533E70 with a successful test callback: the actual C service
// must enumerate the native missile index and apply its geometry, ownership,
// collision eligibility and vision predicates before MainAI can use it.
func TestMonsterMainBlock547210ActualMissileProbe(t *testing.T) {
	for _, mode := range []string{"incoming", "outgoing", "behind", "perpendicular-boundary", "range-boundary", "outside-range", "owner", "no-collision", "blind", "not-indexed"} {
		t.Run(mode, func(t *testing.T) {
			srv, unit, _, missile := monsterMainBlockLegacyFixture547210(t)
			want := 0
			switch mode {
			case "incoming":
				want = 1
			case "outgoing":
				missile.VelVec.X = 4
			case "behind":
				missile.PosVec.X, missile.PrevPos.X, missile.VelVec.X = 260, 256, 4
			case "perpendicular-boundary":
				missile.PosVec.Y, missile.PrevPos.Y = 320, 320
			case "range-boundary":
				// 00533E8D initializes the best distance to 1e9, not 100.
				// The circle enumerator admits an exact radius-100 missile.
				missile.PosVec.X, missile.PrevPos.X = 400, 404
				want = 1
			case "outside-range":
				missile.PosVec.X, missile.PrevPos.X = 400.25, 404.25
			case "owner":
				missile.ObjOwner = unit
			case "no-collision":
				missile.Collide = nil
			case "blind":
				unit.Buffs = 1 << server.ENCHANT_BLINDED
			}
			missile.NewPos = missile.PosVec
			if mode != "not-indexed" {
				srv.Map.AddObjectToIndex(missile)
			}
			if got := Nox_xxx_monsterTestBlockShield_533E70(unit); got != want {
				t.Fatalf("actual C missile probe=%d, want %d, unit=%p missile=%p", got, want, unit, missile)
			}
		})
	}
}

func TestMonsterMainBlock547210ActualLegacyBinding(t *testing.T) {
	for _, mode := range []string{"monster-shield", "npc-shield", "npc-weapon", "npc-weapon-and-shield"} {
		t.Run(mode, func(t *testing.T) {
			srv, unit, update, missile := monsterMainBlockLegacyFixture547210(t)
			want, deadline := ai.ACTION_BLOCK_ATTACK, uint32(1415)
			switch mode {
			case "monster-shield":
				update.StatusFlags = object.MonStatusCanBlock
			case "npc-shield":
				unit.ObjSubClass = object.SubClass(object.MonsterNPC)
				update.ArmorEquipFlags = 0x1000000
			case "npc-weapon", "npc-weapon-and-shield":
				unit.ObjSubClass = object.SubClass(object.MonsterNPC)
				update.WeaponEquipFlags = 0x400
				want, deadline = ai.ACTION_WAIT, 1431
				if mode == "npc-weapon-and-shield" {
					update.ArmorEquipFlags = 0x2000000
				}
			}
			srv.Map.AddObjectToIndex(missile)
			if Nox_xxx_monsterTestBlockShield_533E70(unit) != 1 {
				t.Fatal("actual native missile did not qualify for defense")
			}
			t.Logf("actual MainAI/C probe: unit=%p update=%p missile=%p", unit, update, missile)
			Nox_xxx_monsterMainAIFn_547210(unit)
			if update.AIStackInd != 1 || update.AIStack[0].Type() != ai.ACTION_FIGHT ||
				update.AIStackHead().Type() != want || update.AIStackHead().Args != [4]uintptr{uintptr(deadline)} ||
				!srv.AI.StackChanged || unit.HealthData.Cur != 100 {
				t.Fatalf("actual MainAI defense missing: index=%d stack=%+v changed=%t HP=%d, want %v/%d",
					update.AIStackInd, update.GetAIStack(), srv.AI.StackChanged, unit.HealthData.Cur, want, deadline)
			}
		})
	}
}
