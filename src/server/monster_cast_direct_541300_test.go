package server

import (
	"fmt"
	"math"
	"runtime"
	"testing"
	"unsafe"

	"github.com/opennox/libs/object"
	"github.com/opennox/libs/types"
)

func TestMonsterCastDirect541300NativeRecordsAndCastResult(t *testing.T) {
	for _, bot := range []bool{false, true} {
		for _, result := range []int32{0, 1, -1} {
			t.Run(fmt.Sprintf("bot-%t/result-%d", bot, result), func(t *testing.T) {
				unit := monsterActionTestObject50A910(t)
				update := unit.UpdateDataMonster()
				player := &PlayerUpdateData{Field73: update}
				update.Field545 = player
				if bot {
					update.StatusFlags = object.MonStatusBot
				}
				unit.ObjClass |= object.Class(0x81220000)
				unit.ObjSubClass, unit.Direction1, unit.Direction2 = 0x12345, 7, 8
				unit.PosVec = types.Ptf(10, 20)
				arg := &SpellAcceptArg{Obj: unit, Pos: types.Ptf(15, 25)}
				before := *unit
				calls := 0
				if !MonsterCastDirect541300(4, unit, arg, func(id int32, got *Object, args *SpellAcceptArg) int32 {
					calls++
					if id != 4 || got != unit || args != arg || args.Obj != unit || unit.Direction1 != 7 ||
						unit.Direction2 != DirFromVec(types.Ptf(5, 5)) {
						t.Fatal("direct cast lost arguments or preceded the original direction update")
					}
					wantClass, wantData, wantSub := before.ObjClass, unsafe.Pointer(update), before.ObjSubClass
					if bot {
						wantClass = wantClass&^object.ClassMonster | object.ClassPlayer
						wantData, wantSub = unsafe.Pointer(player), 0
					}
					if unit.ObjClass != wantClass || unit.UpdateData != wantData || unit.ObjSubClass != wantSub {
						t.Fatal("cast did not see the original bot-morph state")
					}
					for _, ptr := range []unsafe.Pointer{unsafe.Pointer(unit), unsafe.Pointer(update), unsafe.Pointer(player), unsafe.Pointer(args)} {
						if unsafe.Sizeof(uintptr(0)) == 8 && uintptr(ptr) <= math.MaxUint32 {
							t.Fatalf("native cast pointer=%p, want above 4 GiB", ptr)
						}
					}
					return result
				}) || calls != 1 {
					t.Fatal("direct cast used acceptance as a service-success gate")
				}
				before.Direction2 = DirFromVec(types.Ptf(5, 5))
				if bot {
					before.ObjSubClass = 16
				}
				if *unit != before {
					t.Fatal("direct cast lost class bits, update link, or unrelated unit state")
				}
				runtime.KeepAlive(player)
				runtime.KeepAlive(unit)
			})
		}
	}
}

func TestMonsterCastDirect541300CachedBotFlagAndLiveReverseLink(t *testing.T) {
	for _, mode := range []string{"replace player record", "clear cached bot", "set cached bot"} {
		t.Run(mode, func(t *testing.T) {
			unit := monsterActionTestObject50A910(t)
			cached := unit.UpdateDataMonster()
			originalPlayer := &PlayerUpdateData{Field73: cached}
			replacementMonster := new(MonsterUpdateData)
			replacementPlayer := &PlayerUpdateData{Field73: replacementMonster}
			cached.Field545, cached.StatusFlags = originalPlayer, object.MonStatusBot
			if mode == "set cached bot" {
				cached.StatusFlags = 0
			}
			arg := &SpellAcceptArg{Obj: unit, Pos: unit.PosVec}
			if !MonsterCastDirect541300(4, unit, arg, func(int32, *Object, *SpellAcceptArg) int32 {
				switch mode {
				case "replace player record":
					unit.UpdateData = unsafe.Pointer(replacementPlayer)
				case "clear cached bot":
					cached.StatusFlags &^= object.MonStatusBot
				case "set cached bot":
					cached.StatusFlags |= object.MonStatusBot
				}
				return 0
			}) {
				t.Fatal("valid direct service was not handled")
			}
			wantClass, wantData, wantSub := object.ClassMonster, unsafe.Pointer(replacementMonster), object.SubClass(16)
			if mode == "clear cached bot" {
				wantClass, wantData, wantSub = object.ClassPlayer, unsafe.Pointer(originalPlayer), 0
			} else if mode == "set cached bot" {
				wantData, wantSub = unsafe.Pointer(cached), 0
			}
			if unit.ObjClass != wantClass || unit.UpdateData != wantData || unit.ObjSubClass != wantSub {
				t.Fatal("direct service cached the wrong flag or restored an old reverse link")
			}
			runtime.KeepAlive(originalPlayer)
			runtime.KeepAlive(replacementPlayer)
			runtime.KeepAlive(unit)
		})
	}
}

func TestMonsterCastDirect541300ContainsInvalidMetadata(t *testing.T) {
	for _, mode := range []string{"nil unit", "nil data", "wrong class", "nil args", "nil service", "missing bot link"} {
		t.Run(mode, func(t *testing.T) {
			unit := monsterActionTestObject50A910(t)
			arg := &SpellAcceptArg{Obj: unit}
			cast := func(int32, *Object, *SpellAcceptArg) int32 { t.Fatal("invalid direct cast reached service"); return 0 }
			switch mode {
			case "nil unit":
				unit = nil
			case "nil data":
				unit.UpdateData = nil
			case "wrong class":
				unit.ObjClass = object.ClassPlayer
			case "nil args":
				arg = nil
			case "nil service":
				cast = nil
			case "missing bot link":
				unit.UpdateDataMonster().StatusFlags = object.MonStatusBot
			}
			var before Object
			if unit != nil {
				before = *unit
			}
			if MonsterCastDirect541300(4, unit, arg, cast) || (unit != nil && *unit != before) {
				t.Fatal("invalid direct metadata changed unit state")
			}
		})
	}
}
