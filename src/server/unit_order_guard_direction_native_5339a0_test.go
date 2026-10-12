package server

import (
	"bytes"
	"fmt"
	"math"
	"testing"
	"unsafe"

	"github.com/opennox/libs/object"
	"github.com/opennox/libs/types"

	"github.com/opennox/opennox/v1/common/unit/ai"
	"github.com/opennox/opennox/v1/legacy/common/alloc"
)

// GAME.EXE's GUARD branch reads the live facing at object+124 as a signed
// WORD and stores its sign-extended DWORD in the action argument. These
// literal expected words are independent of the native Dir16 conversion.
func TestUnitOrderGuardDirectionNative5339A0(t *testing.T) {
	s := unitOrderTestServer533900(t)
	source, freeSource := alloc.New(Object{ObjClass: object.ClassSimple})
	t.Cleanup(freeSource)
	unit, freeUnit := alloc.New(Object{})
	t.Cleanup(freeUnit)
	update, freeUpdate := alloc.New(MonsterUpdateData{})
	t.Cleanup(freeUpdate)
	definition, freeDefinition := alloc.New(MonsterDef{})
	t.Cleanup(freeDefinition)
	for _, pointer := range []unsafe.Pointer{unsafe.Pointer(source), unsafe.Pointer(unit), unsafe.Pointer(update), unsafe.Pointer(definition)} {
		if unsafe.Sizeof(uintptr(0)) == 8 && uintptr(pointer) <= math.MaxUint32 {
			t.Fatalf("native record not above 4 GiB: %p", pointer)
		}
	}
	records := func(pointer unsafe.Pointer, size uintptr) []byte {
		return bytes.Clone(unsafe.Slice((*byte)(pointer), int(size)))
	}
	sourceBefore := records(unsafe.Pointer(source), unsafe.Sizeof(*source))
	definitionBefore := records(unsafe.Pointer(definition), unsafe.Sizeof(*definition))
	directions := []struct {
		input uint16
		word  uint32
	}{
		{0x0000, 0x00000000}, {0x0001, 0x00000001},
		{0x00ff, 0x000000ff}, {0x0100, 0x00000100},
		{0x7fff, 0x00007fff}, {0x8000, 0xffff8000},
		{0xfffe, 0xfffffffe}, {0xffff, 0xffffffff},
	}
	for _, subclass := range []object.SubClass{0, object.SubClass(object.MonsterNPC)} {
		for _, direction := range directions {
			t.Run(fmt.Sprintf("subclass-%x/facing-%04x", uint32(subclass), direction.input), func(t *testing.T) {
				*unit = Object{
					TypeInd: 0x1234, ObjClass: object.ClassMonster,
					ObjSubClass: subclass, ObjFlags: object.FlagEnabled,
					PosVec: types.Ptf(41.25, -82.5), Direction1: Dir16(direction.input),
					Direction2: Dir16(0x1357), SpeedBase: math.Float32frombits(0x3c23d70a),
					UpdateData: unsafe.Pointer(update), serverHandle: s.handle,
				}
				*update = MonsterUpdateData{
					MonsterDef: definition, AIStackInd: 0,
					StatusFlags: object.MonsterStatus(0x80000085),
					Aggression:  0.875, Aggression2: 9.25, SightRange: 321.5,
				}
				update.AIStack[0] = AIStackItem{
					Action: uint32(ai.ACTION_IDLE),
					Args:   [4]uintptr{0x11223344, 0x55667788, 0x9abcdef0, 0x76543210},
				}
				before := records(unsafe.Pointer(unit), unsafe.Sizeof(*unit))
				s.OrderUnit533900(source, unit, 3, UnitOrderRuntime533900{})
				head := update.AIStackHead()
				if head == nil || head.Type() != ai.ACTION_GUARD || update.AIStackInd != 0 {
					t.Fatalf("GUARD action missing: head=%#v index=%d", head, update.AIStackInd)
				}
				if got := head.ArgU32(2); got != direction.word {
					t.Errorf("native facing word=%08x, want original signed DWORD %08x", got, direction.word)
				}
				if head.ArgPos(0) != unit.PosVec || head.Args[3] != 0x76543210 ||
					math.Float32bits(update.Aggression) != 0x3f000000 ||
					math.Float32bits(update.SightRange) != 0x437a0000 ||
					update.Aggression2 != 9.25 || uint32(update.StatusFlags) != 0x80000085 {
					t.Fatal("guard changed original position/unused argument/aggression/sight/status contract")
				}
				if !bytes.Equal(before, records(unsafe.Pointer(unit), unsafe.Sizeof(*unit))) ||
					!bytes.Equal(sourceBefore, records(unsafe.Pointer(source), unsafe.Sizeof(*source))) ||
					!bytes.Equal(definitionBefore, records(unsafe.Pointer(definition), unsafe.Sizeof(*definition))) {
					t.Fatal("guard argument update changed object/source/definition bytes")
				}
			})
		}
	}
}
