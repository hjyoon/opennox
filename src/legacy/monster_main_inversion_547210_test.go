package legacy

import (
	"fmt"
	"reflect"
	"testing"
	"unsafe"

	"github.com/opennox/libs/object"
	"github.com/opennox/libs/spell"
	"github.com/opennox/libs/things"
	"github.com/opennox/libs/types"

	noxflags "github.com/opennox/opennox/v1/common/flags"
	"github.com/opennox/opennox/v1/common/memmap"
	"github.com/opennox/opennox/v1/common/unit/ai"
	"github.com/opennox/opennox/v1/legacy/common/alloc"
	"github.com/opennox/opennox/v1/server"
)

func TestMonsterMainInversion547210ActualLegacyWrapper(t *testing.T) {
	for _, confused := range []bool{false, true} {
		for _, duration := range []bool{false, true} {
			t.Run(fmt.Sprintf("confused=%t/duration=%t", confused, duration), func(t *testing.T) {
				srv, unit, update := monsterLookAtFixture5125A0(t)
				oldFlags := noxflags.GetGame()
				noxflags.ResetGame()
				t.Cleanup(func() { noxflags.ResetGame(); noxflags.SetGame(oldFlags) })
				oldServer := GetServer
				GetServer = func() Server { return &monsterMainLegacyServer547210{srv: srv} }
				t.Cleanup(func() { GetServer = oldServer })
				srv.Map.Init()
				srv.SetFrame(200)
				srv.SetTickRate(30)
				unit.ObjFlags = object.FlagEnabled
				unit.PosVec = types.Ptf(300, 300)
				if confused {
					unit.Buffs = 1 << server.ENCHANT_CONFUSED
				}
				update.StatusFlags = object.MonStatusCanCastSpells
				update.Field363, update.Field362_0, update.Field362_2 = 200, 7, 7
				update.Field124, update.Field137 = 0, 0
				flags := unsafe.Slice(&update.Field373, server.SpellsMax-1)
				flags[int(spell.SPELL_INVERSION)-1] = 0x08000000
				def := &server.SpellDef{ID: spell.SPELL_INVERSION, Def: things.Spell{Flags: things.SpellMobsCanCast}}
				if duration {
					def.Def.Flags |= things.SpellDuration
				}
				field := reflect.ValueOf(&srv.Spells).Elem().FieldByName("byID")
				reflect.NewAt(field.Type(), unsafe.Pointer(field.UnsafeAddr())).Elem().Set(reflect.ValueOf(map[spell.ID]*server.SpellDef{spell.SPELL_INVERSION: def}))
				missile, freeMissile := alloc.New(server.Object{})
				t.Cleanup(freeMissile)
				data, freeData := alloc.New(server.MissileUpdateData{})
				t.Cleanup(freeData)
				*data = server.MissileUpdateData{Target: unit}
				*missile = server.Object{ObjClass: object.ClassMissile, ObjSubClass: object.SubClass(object.MissileMagic), ObjFlags: object.FlagActive, PosVec: unit.PosVec, NewPos: unit.PosVec, UpdateData: unsafe.Pointer(data)}
				srv.Map.AddObjectToIndex(missile)
				flag := memmap.PtrUint32(0x5D4594, 2489156)
				previousFlag := *flag
				t.Cleanup(func() { *flag = previousFlag })
				*flag = 77
				base := update.AIStack[0]
				Nox_xxx_monsterMainAIFn_547210(unit)
				wantIndex, wantType := int8(2), ai.ACTION_CAST_SPELL_ON_OBJECT
				if confused {
					wantIndex += 2
				}
				if duration {
					wantIndex++
					wantType = ai.ACTION_CAST_DURATION_SPELL
				}
				head := update.AIStackHead()
				if update.AIStackInd != wantIndex || head.Type() != wantType || head.ArgU32(0) != uint32(spell.SPELL_INVERSION) || head.ArgObj(2) != unit ||
					update.AIStack[0] != base || update.Field363 != 207 || *flag != 1 || !srv.AI.StackChanged {
					t.Fatalf("real wrapper did not reserve native inversion: index=%d head=%+v deadline=%d flag=%d", update.AIStackInd, head, update.Field363, *flag)
				}
				if confused && (update.AIStack[1].Type() != ai.DEPENDENCY_IS_ENCHANTED || update.AIStack[1].ArgU32(0) != uint32(server.ENCHANT_CONFUSED) || update.AIStack[2].Type() != ai.ACTION_CONFUSED) {
					t.Fatal("actual wrapper lost original confused prefix")
				}
			})
		}
	}
}
