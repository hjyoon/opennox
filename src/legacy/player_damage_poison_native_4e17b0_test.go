package legacy

import (
	"fmt"
	"math"
	"testing"
	"unsafe"

	"github.com/opennox/libs/object"
	"github.com/opennox/libs/strman"
	"github.com/opennox/libs/types"

	noxflags "github.com/opennox/opennox/v1/common/flags"
	"github.com/opennox/opennox/v1/legacy/common/alloc"
	"github.com/opennox/opennox/v1/server"
)

func TestPlayerDamagePoisonNativeCallback4E17B0MonsterHP(t *testing.T) {
	srv := server.New(nil, nil, strman.New())
	t.Cleanup(srv.Close)
	srv.SetFrame(1400)
	oldGetServer := GetServer
	GetServer = func() Server { return &itemDurabilityLegacyServer4E1560{srv: srv} }
	t.Cleanup(func() { GetServer = oldGetServer })
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
	for _, immune := range []bool{false, true} {
		t.Run(fmt.Sprintf("immune-%t", immune), func(t *testing.T) {
			target, freeTarget := alloc.New(server.Object{})
			defer freeTarget()
			update, freeUpdate := alloc.New(server.MonsterUpdateData{})
			defer freeUpdate()
			*update = server.MonsterUpdateData{Field1: math.Float32bits(0.25), Field518: math.Float32bits(0.5), Field547: 99, Field546: 91}
			health, freeHealth := alloc.New(server.HealthData{})
			defer freeHealth()
			*health = server.HealthData{Cur: 20, Max: 20, Field2: 20}
			*target = server.Object{ObjClass: object.ClassMonster, ObjSubClass: 0x11012, Material: 0x4000,
				UpdateData: unsafe.Pointer(update), HealthData: health, Damage: playerDamageMeleeCallbackNative4E17B0(),
				Buffs:  1<<server.ENCHANT_SHIELD | 1<<server.ENCHANT_REFLECTIVE_SHIELD,
				Pos132: types.Ptf(31, 47), Frame134: 77}
			if immune {
				target.ObjSubClass |= 0x200
			}
			for _, ptr := range []unsafe.Pointer{unsafe.Pointer(target), unsafe.Pointer(update), unsafe.Pointer(health)} {
				if unsafe.Sizeof(uintptr(0)) == 8 && uintptr(ptr) <= math.MaxUint32 {
					t.Fatalf("poison pointer=%p, want >4 GiB", ptr)
				}
			}
			// Real C dispatcher, registered PlayerDamage and DefaultDamage adapters,
			// and UnitSetHP: neither damage nor HP callbacks are substituted.
			if !objectDamageDispatchCallNative(target, nil, nil, 1, object.DamagePoison) {
				t.Fatal("registered NPC poison callback rejected damage")
			}
			if immune {
				if health.Cur != 20 || update.Field547 != 0 || update.Field546 != 5 || target.Frame134 != 77 || target.Pos132 != types.Ptf(31, 47) {
					t.Fatal("immune poison tail changed")
				}
			} else if health.Cur != 19 || target.Field38 != math.MaxUint32 || update.Field547 != 2 || update.Field546 != 5 ||
				target.Obj130 != nil || target.Field131 != 5 || target.Frame134 != 1400 || target.Pos132 != (types.Pointf{}) || !update.StatusFlags.Has(object.MonStatusInjured) {
				t.Fatalf("NPC poison HP=%d marker=%d/%d frame=%d", health.Cur, update.Field546, update.Field547, target.Frame134)
			}
			if update.Field1 != math.Float32bits(0.25) || update.Field518 != math.Float32bits(0.5) {
				t.Fatal("poison changed NPC fractional carry/armor")
			}
			t.Logf("C dispatcher -> PlayerDamage -> DefaultDamage -> UnitSetHP: target=%p update=%p immune=%t HP=20->%d", target, update, immune, health.Cur)
		})
	}
}
