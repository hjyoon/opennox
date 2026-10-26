package opennox

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"math"
	"runtime"
	"testing"
	"unsafe"

	"github.com/opennox/libs/object"
	"github.com/opennox/libs/strman"
	"github.com/opennox/noxscript/ns/asm"
	ns4 "github.com/opennox/noxscript/ns/v4"

	noxflags "github.com/opennox/opennox/v1/common/flags"
	"github.com/opennox/opennox/v1/common/unit/ai"
	"github.com/opennox/opennox/v1/legacy/common/alloc"
	"github.com/opennox/opennox/v1/server"
)

// These are C-owned records, not integer-address aliases or Go pointers stored
// in C memory. The public script wrapper must preserve every byte except the
// original low-byte store at MonsterUpdateData.Field333.
func scriptNSRoamFlagFixture515C80(t *testing.T) (*server.Object, *server.MonsterUpdateData) {
	t.Helper()
	unit, freeUnit := alloc.New(server.Object{})
	update, freeUpdate := alloc.New(server.MonsterUpdateData{})
	t.Cleanup(freeUpdate)
	t.Cleanup(freeUnit)
	*unit = server.Object{ObjClass: object.ClassMonster, ObjFlags: object.FlagEnabled, Extent: 0xf1234567, NetCode: 123, UpdateData: unsafe.Pointer(update)}
	*update = server.MonsterUpdateData{Field333: 0xdeadbeef, Field124: 51, Field137: 93, AIStackInd: 0, Aggression: .01, RetreatLevel: .98, ResumeLevel: 1}
	update.AIStack[0] = server.AIStackItem{Action: uint32(ai.ACTION_WAIT), Args: [4]uintptr{101, 102, 103, 104}}
	if unsafe.Sizeof(uintptr(0)) == 8 && (uintptr(unsafe.Pointer(unit)) <= math.MaxUint32 || uintptr(unsafe.Pointer(update)) <= math.MaxUint32) {
		t.Fatalf("native script records are not above 4 GiB: unit=%p update=%p", unit, update)
	}
	return unit, update
}

func scriptNSRoamFlagBytes515C80(ptr unsafe.Pointer, size uintptr) []byte {
	return bytes.Clone(unsafe.Slice((*byte)(ptr), int(size)))
}

func TestScriptNSObjectRoamFlag515C80LowByteAndState(t *testing.T) {
	for _, class := range []object.Class{0, object.ClassPlayer, object.ClassMissile, object.ClassMonster, object.ClassMonster | object.ClassPlayer} {
		for _, flag := range []int{0, 0x42, 0x80, 0xff, 0x128, -1, math.MinInt32, math.MaxInt32} {
			t.Run(fmt.Sprintf("class-%x/flag-%d", class, flag), func(t *testing.T) {
				unit, update := scriptNSRoamFlagFixture515C80(t)
				unit.ObjClass = class
				unitBefore := scriptNSRoamFlagBytes515C80(unsafe.Pointer(unit), unsafe.Sizeof(*unit))
				want := scriptNSRoamFlagBytes515C80(unsafe.Pointer(update), unsafe.Sizeof(*update))
				if class.Has(object.ClassMonster) {
					want[unsafe.Offsetof(update.Field333)] = uint8(flag)
				}
				var api ns4.Obj = nsObj{Object: asObjectS(unit)}
				api.SetRoamFlag(flag)
				if !bytes.Equal(unitBefore, scriptNSRoamFlagBytes515C80(unsafe.Pointer(unit), unsafe.Sizeof(*unit))) || !bytes.Equal(want, scriptNSRoamFlagBytes515C80(unsafe.Pointer(update), unsafe.Sizeof(*update))) {
					t.Fatalf("public SetRoamFlag changed more than the monster low byte: class=%x flag=%d Field333=%08x", class, flag, update.Field333)
				}
				runtime.KeepAlive(api)
			})
		}
	}
}

func TestScriptNSObjectRoamFlag515C80EmptyAndWideArguments(t *testing.T) {
	(nsObj{}).SetRoamFlag(-1)
	unit, update := scriptNSRoamFlagFixture515C80(t)
	unit.UpdateData = nil
	unitBefore := scriptNSRoamFlagBytes515C80(unsafe.Pointer(unit), unsafe.Sizeof(*unit))
	(nsObj{Object: asObjectS(unit)}).SetRoamFlag(0x180)
	if !bytes.Equal(unitBefore, scriptNSRoamFlagBytes515C80(unsafe.Pointer(unit), unsafe.Sizeof(*unit))) || update.Field333 != 0xdeadbeef {
		t.Fatal("missing-update target was changed")
	}
	if unsafe.Sizeof(uintptr(0)) == 8 {
		unit.UpdateData = unsafe.Pointer(update)
		wide := int64(1)<<40 | 0x1a5
		(nsObj{Object: asObjectS(unit)}).SetRoamFlag(int(wide))
		if update.Field333 != 0xdeadbea5 {
			t.Fatalf("native int did not narrow to one byte: %x", update.Field333)
		}
	}
}

// A synthetic SCRIPT03 is only an explicit input. This invokes the existing
// map-script parser/resolver and frozen classic builtin, not a supplied result.
func scriptNSRoamFlagProgram515C80(t *testing.T, builtin asm.Builtin, handle, flag int32) []byte {
	t.Helper()
	var data bytes.Buffer
	word := func(value uint32) {
		if err := binary.Write(&data, binary.LittleEndian, value); err != nil {
			t.Fatal(err)
		}
	}
	data.WriteString("SCRIPT03STRG")
	word(0)
	data.WriteString("CODE")
	word(3)
	for i, name := range []string{"GLOBAL", "GLOBAL", "test_roam_flag"} {
		data.WriteString("FUNC")
		word(uint32(len(name)))
		data.WriteString(name)
		word(0)
		word(0)
		data.WriteString("SYMB")
		word(0)
		word(0)
		code := []uint32{uint32(asm.OpReturn0)}
		if i == 2 {
			code = []uint32{uint32(asm.OpPushInt), uint32(handle), uint32(asm.OpPushInt), uint32(flag), uint32(asm.OpCallBuiltin), uint32(builtin), uint32(asm.OpReturn0)}
		}
		data.WriteString("DATA")
		word(uint32(4 * len(code)))
		for _, value := range code {
			word(value)
		}
	}
	data.WriteString("DONE")
	return data.Bytes()
}

func scriptNSRoamFlagServer515C80(t *testing.T) *Server {
	t.Helper()
	s := NewServer(nil, nil, strman.New())
	t.Cleanup(s.Close)
	oldServer := noxServer
	noxServer = s
	t.Cleanup(func() { noxServer = oldServer })
	oldGame := noxflags.GetGame()
	noxflags.UnsetGame(noxflags.GameFlag22 | noxflags.GameFlag23)
	t.Cleanup(func() { noxflags.ResetGame(); noxflags.SetGame(oldGame) })
	return s
}

func TestScriptNSObjectRoamFlag515C80MatchesClassicBuiltin(t *testing.T) {
	s := scriptNSRoamFlagServer515C80(t)
	unit, update := scriptNSRoamFlagFixture515C80(t)
	for _, flag := range []int32{0, 0x80, 0x128, -1, math.MinInt32, math.MaxInt32} {
		t.Run(fmt.Sprint(flag), func(t *testing.T) {
			before := scriptNSRoamFlagBytes515C80(unsafe.Pointer(update), unsafe.Sizeof(*update))
			var api ns4.Obj = s.noxScriptP().toObj(unit)
			api.SetRoamFlag(int(flag))
			want := scriptNSRoamFlagBytes515C80(unsafe.Pointer(update), unsafe.Sizeof(*update))
			copy(unsafe.Slice((*byte)(unsafe.Pointer(update)), int(unsafe.Sizeof(*update))), before)
			if err := s.Server.NoxScriptVM.ReadScript(bytes.NewReader(scriptNSRoamFlagProgram515C80(t, asm.BuiltinSetRoamFlag, -1, flag))); err != nil {
				t.Fatal(err)
			}
			if err := s.Server.NoxScriptVM.CallByIndex(2, unit, nil); err != nil {
				t.Fatal(err)
			}
			if !bytes.Equal(want, scriptNSRoamFlagBytes515C80(unsafe.Pointer(update), unsafe.Sizeof(*update))) || update.Field333 != 0xdeadbe00|uint32(uint8(flag)) {
				t.Fatalf("modern/classic script routes differ: flag=%d Field333=%x", flag, update.Field333)
			}
		})
	}
}
