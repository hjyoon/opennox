package opennox

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"math"
	"testing"
	"unsafe"

	"github.com/opennox/libs/object"
	"github.com/opennox/libs/strman"
	"github.com/opennox/libs/types"
	"github.com/opennox/noxscript/ns/asm"

	noxflags "github.com/opennox/opennox/v1/common/flags"
	"github.com/opennox/opennox/v1/common/unit/ai"
	"github.com/opennox/opennox/v1/legacy/common/alloc"
	"github.com/opennox/opennox/v1/server"
)

// A public synthetic SCRIPT03 exercises the real map-script reader and
// builtin 13. Neither the script resolver nor the C/Go builtin is substituted.
func monsterLookAtScript5125A0(t *testing.T, handle, direction int32) []byte {
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
	for i, name := range []string{"GLOBAL", "GLOBAL", "test_look_at"} {
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
			code = []uint32{
				uint32(asm.OpPushInt), uint32(handle), uint32(asm.OpPushInt), uint32(direction),
				uint32(asm.OpCallBuiltin), uint32(asm.BuiltinLookAtDirection), uint32(asm.OpReturn0),
			}
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

func TestMonsterLookAtScript5125A0NativeCallerAndTrigger(t *testing.T) {
	s := NewServer(nil, nil, strman.New())
	t.Cleanup(s.Close)
	if !s.Objs.Init(2) {
		t.Fatal("cannot initialize native object allocator")
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
			t.Fatalf("script native pointer below 4 GiB: unit=%p update=%p", unit, update)
		}
	}
	angles := [...]byte{160, 192, 224, 128, 0, 0, 96, 64, 32}
	for _, handle := range []int32{-2, -1} {
		for direction, angle := range angles {
			t.Run(fmt.Sprintf("%d/%d", handle, direction), func(t *testing.T) {
				for _, unit := range []*server.Object{caller, trigger} {
					unit.PosVec = types.Ptf(8640, -3120)
					unit.Direction1, unit.Direction2 = 13, 29
					update := unit.UpdateDataMonster()
					*update = server.MonsterUpdateData{AIStackInd: 0}
					update.AIStack[0] = server.AIStackItem{Action: uint32(ai.ACTION_WAIT), Args: [4]uintptr{1, 2, 3, 4}}
				}
				selected, other := trigger, caller
				if handle == -1 {
					selected, other = caller, trigger
				}
				before := other.UpdateDataMonster().AIStack[0]
				cosine, sine := server.SinCosDir(angle)
				want := [4]uintptr{
					uintptr(math.Float32bits(float32(float64(cosine)*10 + float64(selected.PosVec.X)))),
					uintptr(math.Float32bits(float32(float64(sine)*10 + float64(selected.PosVec.Y)))), 0, 0,
				}
				if err := s.Server.NoxScriptVM.ReadScript(bytes.NewReader(monsterLookAtScript5125A0(t, handle, int32(direction)))); err != nil {
					t.Fatal(err)
				}
				// PlayerDialogClose uses this same real CallByIndex path.
				if err := s.Server.NoxScriptVM.CallByIndex(2, caller, trigger); err != nil {
					t.Fatal(err)
				}
				update := selected.UpdateDataMonster()
				if update.AIStackInd != 1 || update.AIStack[0] != before || update.AIStack[1].Type() != ai.ACTION_FACE_LOCATION || update.AIStack[1].Args != want {
					t.Fatalf("script LookAt stack index=%d head=%+v want %08x", update.AIStackInd, update.AIStackHead(), want)
				}
				if update.Field124 != 702 || update.Field137 != 702 || selected.Direction1 != 13 || selected.Direction2 != 29 {
					t.Fatal("script LookAt skipped reset services or assigned facing immediately")
				}
				if other.UpdateDataMonster().AIStackInd != 0 || other.UpdateDataMonster().AIStack[0] != before {
					t.Fatal("script LookAt resolved the wrong caller/trigger")
				}
			})
		}
	}
}
