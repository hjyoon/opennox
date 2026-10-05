package legacy

import (
	"fmt"
	"testing"
	"unsafe"

	"github.com/opennox/libs/object"
	"github.com/opennox/libs/types"
	noxflags "github.com/opennox/opennox/v1/common/flags"
	"github.com/opennox/opennox/v1/common/unit/ai"
	"github.com/opennox/opennox/v1/legacy/common/alloc"
	"github.com/opennox/opennox/v1/server"
)

func TestMonsterMainBlink547210ActualLegacyBinding(t *testing.T) {
	for _, bot := range []bool{false, true} {
		for _, accept := range []int32{0, 1, -1} {
			t.Run(fmt.Sprintf("bot=%t/accept=%d", bot, accept), func(t *testing.T) {
				srv, unit, update := monsterLookAtFixture5125A0(t)
				oldFlags := noxflags.GetGame()
				noxflags.ResetGame()
				t.Cleanup(func() { noxflags.ResetGame(); noxflags.SetGame(oldFlags) })
				oldServer, oldCast := GetServer, Nox_xxx_castSpellByUser_4FDD20
				GetServer = func() Server { return &monsterMainLegacyServer547210{srv: srv} }
				t.Cleanup(func() { GetServer = oldServer; Nox_xxx_castSpellByUser_4FDD20 = oldCast })
				srv.Map.Init()
				srv.SetTickRate(30)
				srv.SetFrame(200)
				unit.ObjFlags, unit.SpeedBase = object.FlagEnabled, 2
				unit.PosVec = types.Ptf(300, 300)
				unit.Shape.Kind, unit.Shape.Circle.R = server.ShapeKindCircle, 60
				enemy, freeEnemy := alloc.New(server.Object{})
				t.Cleanup(freeEnemy)
				enemy.PosVec = types.Ptf(390, 300)
				// Center distance 90 exceeds FleeRange; surface distance 30 is
				// strictly inside its Blink half-range, exercising real geometry.
				update.CurrentEnemy, update.Aggression, update.FleeRange = enemy, 0.5, 65
				update.StatusFlags, update.Field376 = object.MonStatusCanCastSpells, 0x80000000
				update.Field127, update.Field371 = 0, 200
				update.Field370_0, update.Field370_2 = 7, 7
				var player *server.PlayerUpdateData
				if bot {
					var freePlayer func()
					player, freePlayer = alloc.New(server.PlayerUpdateData{})
					t.Cleanup(freePlayer)
					player.Field73 = update
					update.StatusFlags |= object.MonStatusBot
					update.Field545 = player
					if unsafe.Sizeof(uintptr(0)) == 8 && uintptr(unsafe.Pointer(player)) <= 0xffffffff {
						t.Fatal("bot player record is not a native high pointer")
					}
				}
				update.AIStack[0].Action = uint32(ai.ACTION_FLEE) // Blink precedes this gate.
				stack := update.AIStack
				calls := 0
				Nox_xxx_castSpellByUser_4FDD20 = func(id int32, got *server.Object, arg *server.SpellAcceptArg) int32 {
					calls++
					if id != 4 || got != unit || arg.Obj != unit || arg.Pos != unit.PosVec ||
						unit.Direction2 != server.DirFromVec(types.Pointf{}) {
						t.Fatal("actual legacy binding lost immediate cast arguments/direction")
					}
					if bot && (!unit.ObjClass.Has(object.ClassPlayer) || unit.ObjClass.Has(object.ClassMonster) ||
						unit.UpdateData != unsafe.Pointer(player) || player.Field73 != update) {
						t.Fatal("bot did not use its native player record during casting")
					}
					if !bot && (!unit.ObjClass.Has(object.ClassMonster) || unit.UpdateData != unsafe.Pointer(update)) {
						t.Fatal("ordinary monster unexpectedly morphed")
					}
					return accept
				}
				Nox_xxx_monsterMainAIFn_547210(unit)
				if calls != 1 || update.Field371 != 207 || update.AIStack != stack || update.AIStackInd != 0 ||
					!unit.ObjClass.Has(object.ClassMonster) || unit.UpdateData != unsafe.Pointer(update) ||
					(bot && uint32(unit.ObjSubClass) != 16) {
					t.Fatalf("native MainAI -> 541300 -> 4FDD20 missing: calls=%d deadline=%d class=%x data=%p frame=%d tickRate=%d speed=%g aggression=%g range=%g distance=%g buffs=%x recent=%t RNG=%p",
						calls, update.Field371, uint32(unit.ObjClass), unit.UpdateData, srv.Frame(), srv.TickRate(), unit.SpeedBase, update.Aggression,
						update.FleeRange, objectDistance_4E6C00(unit, enemy), unit.Buffs, srv.MonsterMoveAttemptRecent534810(unit), srv.Rand.Logic)
				}
			})
		}
	}
}
