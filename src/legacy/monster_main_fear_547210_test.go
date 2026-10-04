package legacy

import (
	"fmt"
	"math"
	"reflect"
	"testing"
	"unsafe"

	"github.com/opennox/libs/object"

	noxflags "github.com/opennox/opennox/v1/common/flags"
	"github.com/opennox/opennox/v1/common/unit/ai"
	"github.com/opennox/opennox/v1/legacy/common/alloc"
	"github.com/opennox/opennox/v1/server"
)

// The real wrapper also calls C -> Go audio. Keep every shared record C-owned
// so strict cgo checks exercise the live bridge, not a replacement test hook.
func TestMonsterMainFear547210ActualLegacyWrapper(t *testing.T) {
	for _, mode := range []string{"wait", "idle", "confused", "anti magic", "disabled", "full", "one slot", "dead head"} {
		t.Run(mode, func(t *testing.T) {
			srv, unit, update := monsterLookAtFixture5125A0(t)
			oldFlags := noxflags.GetGame()
			noxflags.ResetGame()
			t.Cleanup(func() { noxflags.ResetGame(); noxflags.SetGame(oldFlags) })
			oldServer := GetServer
			GetServer = func() Server { return &monsterMainLegacyServer547210{srv: srv} }
			t.Cleanup(func() { GetServer = oldServer })
			srv.SetFrame(200)
			unit.ObjFlags = object.FlagEnabled
			unit.Buffs = 1 << server.ENCHANT_AFRAID
			unit.SpeedBase = 1.95
			sounds, freeSounds := alloc.New([13]uint32{})
			t.Cleanup(freeSounds)
			*sounds = [13]uint32{12: 600}
			update.SoundSet122 = unsafe.Pointer(sounds)
			switch mode {
			case "idle":
				update.AIStack[0].Action = uint32(ai.ACTION_IDLE)
				update.Field137 = 200 + unit.NetCode
			case "confused":
				unit.Buffs |= 1 << server.ENCHANT_CONFUSED
			case "anti magic":
				unit.Buffs |= 1 << server.ENCHANT_ANTI_MAGIC
			case "disabled":
				unit.ObjFlags &^= object.FlagEnabled
			case "full":
				update.AIStackInd = int8(len(update.AIStack) - 1)
				update.AIStackHead().Action = uint32(ai.ACTION_WAIT)
			case "one slot":
				update.AIStackInd = int8(len(update.AIStack) - 2)
				update.AIStackHead().Action = uint32(ai.ACTION_WAIT)
			case "dead head":
				update.AIStack[0].Action = uint32(ai.ACTION_DEAD)
			}
			before := *update
			t.Logf("actual fear wrapper: unit=%p update=%p sound=%p mode=%s", unit, update, sounds, mode)
			Nox_xxx_monsterMainAIFn_547210(unit)
			queued := reflect.ValueOf(&srv.Audio).Elem().FieldByName("delayedObj")
			if queued.Len() != 1 {
				t.Fatalf("real C -> Go fear audio queued %d events, want 1", queued.Len())
			}
			event := queued.Index(0)
			if event.FieldByName("ID").Int() != 600 || event.FieldByName("Obj").Pointer() != uintptr(unsafe.Pointer(unit)) ||
				event.FieldByName("Kind").Int() != 0 || event.FieldByName("Code").Uint() != 0 {
				t.Fatalf("real fear audio id=%d unit=%x kind=%d code=%d, want 600/%x/0/0", event.FieldByName("ID").Int(), event.FieldByName("Obj").Pointer(), event.FieldByName("Kind").Int(), event.FieldByName("Code").Uint(), uintptr(unsafe.Pointer(unit)))
			}
			if unsafe.Sizeof(uintptr(0)) == 8 && uintptr(unsafe.Pointer(sounds)) <= math.MaxUint32 {
				t.Fatal("sound fixture was not above 4 GiB")
			}
			if mode == "full" || mode == "dead head" {
				if *update != before || srv.AI.StackChanged {
					t.Fatal("real wrapper mutated a rejected stack")
				}
				return
			}
			if mode == "one slot" {
				if update.AIStackInd != int8(len(update.AIStack)-1) || update.AIStackHead().Type() != ai.DEPENDENCY_IS_ENCHANTED || update.AIStackHead().ArgU32(0) != 11 {
					t.Fatal("real wrapper lost the partial fear dependency")
				}
				return
			}
			index := 1
			if mode == "idle" {
				index = 0
			}
			if mode == "confused" {
				if update.AIStack[1].Type() != ai.DEPENDENCY_IS_ENCHANTED || update.AIStack[1].ArgU32(0) != 3 || update.AIStack[2].Type() != ai.ACTION_CONFUSED {
					t.Fatal("real wrapper lost confusion before fear")
				}
				index += 2
			}
			wantArgs := [4]uintptr{uintptr(math.Float32bits(unit.PosVec.X)), uintptr(math.Float32bits(unit.PosVec.Y)), 0, 0}
			if update.AIStackInd != int8(index+1) || update.AIStack[index].Type() != ai.DEPENDENCY_IS_ENCHANTED || update.AIStack[index].ArgU32(0) != 11 ||
				update.AIStackHead().Type() != ai.ACTION_FLEE || update.AIStackHead().Args != wantArgs || !srv.AI.StackChanged {
				t.Fatalf("real native fear stack=%s", fmt.Sprint(update.GetAIStack()))
			}
		})
	}
}
