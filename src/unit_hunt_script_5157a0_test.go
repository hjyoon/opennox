package opennox

import (
	"bytes"
	"encoding/binary"
	"math"
	"testing"
	"unsafe"

	"github.com/opennox/libs/object"
	"github.com/opennox/libs/strman"
	"github.com/opennox/noxscript/ns/asm"

	noxflags "github.com/opennox/opennox/v1/common/flags"
	"github.com/opennox/opennox/v1/common/unit/ai"
	"github.com/opennox/opennox/v1/legacy/common/alloc"
	"github.com/opennox/opennox/v1/server"
)

// Encode a public, synthetic SCRIPT03 function, then load it with the real
// map-script reader. No stock script bytes or replacement builtin are used.
func unitHuntScript5157A0(t *testing.T, handle int32) []byte {
	t.Helper()
	var data bytes.Buffer
	word := func(v uint32) {
		if err := binary.Write(&data, binary.LittleEndian, v); err != nil {
			t.Fatal(err)
		}
	}
	data.WriteString("SCRIPT03STRG")
	word(0)
	data.WriteString("CODE")
	word(3)
	for i, name := range []string{"GLOBAL", "GLOBAL", "test_hunt"} {
		data.WriteString("FUNC")
		word(uint32(len(name)))
		data.WriteString(name)
		word(0) // return values
		word(0) // arguments
		data.WriteString("SYMB")
		word(0) // locals
		word(0) // unused
		code := []uint32{uint32(asm.OpReturn0)}
		if i == 2 {
			code = []uint32{uint32(asm.OpPushInt), uint32(handle), uint32(asm.OpCallBuiltin), uint32(asm.BuiltinCreatureHunt), uint32(asm.OpReturn0)}
		}
		data.WriteString("DATA")
		word(uint32(len(code) * 4))
		for _, v := range code {
			word(v)
		}
	}
	data.WriteString("DONE")
	return data.Bytes()
}

func TestUnitHuntScriptCallback5157A0NativeCallerAndTrigger(t *testing.T) {
	s := NewServer(nil, nil, strman.New())
	t.Cleanup(s.Close)
	if !s.Objs.Init(2) {
		t.Fatal("cannot initialize server object allocator")
	}
	t.Cleanup(s.Objs.FreeObjects)
	oldServer := noxServer
	noxServer = s
	t.Cleanup(func() { noxServer = oldServer })
	oldGame := noxflags.GetGame() & (noxflags.GameFlag22 | noxflags.GameFlag23)
	noxflags.UnsetGame(noxflags.GameFlag22 | noxflags.GameFlag23)
	t.Cleanup(func() { noxflags.SetGame(oldGame) })
	oldEngine := noxflags.GetEngine() & noxflags.EngineShowAI
	noxflags.UnsetEngine(noxflags.EngineShowAI)
	t.Cleanup(func() { noxflags.SetEngine(oldEngine) })
	s.SetFrame(702)
	caller := s.Objs.NewObject(&server.ObjectType{})
	trigger := s.Objs.NewObject(&server.ObjectType{})
	for _, unit := range []*server.Object{caller, trigger} {
		update, free := alloc.New(server.MonsterUpdateData{})
		t.Cleanup(free)
		unit.ObjClass, unit.UpdateData = object.ClassMonster, unsafe.Pointer(update)
		if unsafe.Sizeof(uintptr(0)) == 8 &&
			(uintptr(unsafe.Pointer(unit)) <= math.MaxUint32 || uintptr(unit.UpdateData) <= math.MaxUint32) {
			t.Fatalf("callback native pointer below 4 GiB: unit=%p update=%p", unit, update)
		}
	}
	for _, tc := range []struct {
		name   string
		handle int32
		class  object.Class
		flags  object.Flags
		want   ai.ActionType
	}{
		{"trigger", -2, object.ClassMonster, 0, ai.ACTION_HUNT},
		{"caller", -1, object.ClassMonster, 0, ai.ACTION_HUNT},
		{"no-update is not dead", -2, object.ClassMonster, object.FlagNoUpdate, ai.ACTION_HUNT},
		{"dead", -2, object.ClassMonster, object.FlagDead, ai.ACTION_WAIT},
		{"player", -2, object.ClassPlayer, 0, ai.ACTION_WAIT},
		{"destroyed is rejected by script resolver", -2, object.ClassMonster, object.FlagDestroyed, ai.ACTION_WAIT},
	} {
		t.Run(tc.name, func(t *testing.T) {
			for _, unit := range []*server.Object{caller, trigger} {
				unit.ObjClass, unit.ObjFlags = object.ClassMonster, 0
				update := unit.UpdateDataMonster()
				*update = server.MonsterUpdateData{AIStackInd: 0}
				update.AIStack[0] = server.AIStackItem{Action: uint32(ai.ACTION_WAIT), Args: [4]uintptr{1, 2, 3, 4}}
			}
			selected, other := trigger, caller
			if tc.handle == -1 {
				selected, other = caller, trigger
			}
			update := selected.UpdateDataMonster()
			triggerUpdate := trigger.UpdateDataMonster()
			selected.ObjClass, selected.ObjFlags = tc.class, tc.flags
			if err := s.Server.NoxScriptVM.ReadScript(bytes.NewReader(unitHuntScript5157A0(t, tc.handle))); err != nil {
				t.Fatal(err)
			}
			block := &triggerUpdate.ScriptEnemySighted
			*block = server.ScriptCallback{Func: 2}
			out := uint32(0xfeedbeef)
			if got := s.Server.NoxScriptVM.ScriptCallbackRaw(block, caller, trigger, &out); got != &out || out != 0 || block.Flags != 0 {
				t.Fatalf("callback return=%p/%#x flags=%#x", got, out, block.Flags)
			}
			if update.AIStackInd != 0 || update.AIStack[0].Type() != tc.want ||
				other.UpdateDataMonster().AIStack[0].Type() != ai.ACTION_WAIT ||
				other.UpdateDataMonster().AIStack[0].Args != ([4]uintptr{1, 2, 3, 4}) {
				t.Fatalf("Hunt VM selected=%+v other=%+v want action=%v", update.AIStack[0], other.UpdateDataMonster().AIStack[0], tc.want)
			}
			if tc.want == ai.ACTION_HUNT && (update.AIStack[0].Args != ([4]uintptr{}) || update.Field124 != 702 || update.Field137 != 702) {
				t.Fatal("script Hunt skipped the native stack-reset services")
			}
			if tc.want == ai.ACTION_WAIT && update.AIStack[0].Args != ([4]uintptr{1, 2, 3, 4}) {
				t.Fatal("rejected script Hunt modified the existing action payload")
			}
		})
	}
}
