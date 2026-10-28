package opennox

import (
	"bytes"
	"go/ast"
	"go/parser"
	"go/token"
	"math"
	"testing"
	"unsafe"

	"github.com/opennox/libs/object"
	"github.com/opennox/libs/spell"
	"github.com/opennox/libs/types"
	"github.com/opennox/opennox/v1/legacy/common/alloc"
	"github.com/opennox/opennox/v1/server"
)

func TestMonsterQuickbarSwapTargetCopiesActualNativeRecord(t *testing.T) {
	type frame struct {
		Before                [32]byte
		Caster, Other, Target server.Object
		Health                [2]server.HealthData
		Records               [3]server.DurSpell
		After                 [32]byte
	}
	f, free := alloc.New(frame{})
	t.Cleanup(free)
	*f = frame{}
	for i := range f.Before {
		f.Before[i], f.After[i] = byte(0x51+i), byte(0xb7-i)
	}
	f.Health[0], f.Health[1] = server.HealthData{Cur: 20, Max: 20}, server.HealthData{Cur: 1, Max: 1}
	f.Caster = server.Object{ObjClass: object.ClassPlayer, ScriptIDVal: 1000000001, HealthData: &f.Health[0], PosVec: types.Ptf(2822, 2758)}
	f.Target = server.Object{ObjClass: object.ClassMonster, ScriptIDVal: 1000138581, NetCode: 13, TypeInd: 1300,
		HealthData: &f.Health[1], PosVec: types.Ptf(2776, 2510)}
	f.Records[2] = server.DurSpell{ID: 7, Spell: uint32(spell.SPELL_SWAP), Caster16: &f.Caster, Target48: &f.Target, Frame60: 599, Frame68: 628}
	f.Records[1] = server.DurSpell{Spell: uint32(spell.SPELL_SWAP), Caster16: &f.Other, Target48: &f.Target, Frame60: 599, Next: &f.Records[2]}
	f.Records[0] = server.DurSpell{Spell: uint32(spell.SPELL_BLINK), Caster16: &f.Caster, Target48: &f.Target, Frame60: 599, Next: &f.Records[1]}
	for _, pointer := range []unsafe.Pointer{unsafe.Pointer(&f.Caster), unsafe.Pointer(&f.Target), unsafe.Pointer(&f.Records[2])} {
		if unsafe.Sizeof(uintptr(0)) == 8 && uintptr(pointer) <= math.MaxUint32 {
			t.Fatalf("actual observer must exercise native C-owned pointers above4GiB: %p", pointer)
		}
	}
	raw := func() []byte { return unsafe.Slice((*byte)(unsafe.Pointer(f)), int(unsafe.Sizeof(*f))) }
	before := append([]byte(nil), raw()...)
	got, ok := e2eReadMonsterQuickbarSwapTarget(600, 585, &f.Caster, &f.Records[0], &server.Server{})
	if !ok || got.id != 7 || got.created != 599 || got.trigger != 627 || got.wire != 13 ||
		got.player.ScriptID != 1000000001 || got.target.ScriptID != 1000138581 || got.target.Health != 1 ||
		got.target.Position != types.Ptf(2776, 2510) {
		t.Fatalf("observe actual live Swap target rather than another caster/spell or the nominated input target: %+v found=%t", got, ok)
	}
	if !bytes.Equal(before, raw()) {
		t.Fatal("native duration/object/health bytes and surrounding canaries must remain unchanged")
	}
	f.Target.PosVec, f.Health[1].Cur, f.Records[2].Target48 = types.Pointf{}, 0, nil
	if got.target.Position != types.Ptf(2776, 2510) || got.target.Health != 1 || got.target.ScriptID != 1000138581 {
		t.Fatal("subsequent native mutation must not change the copied observation")
	}
}

func TestMonsterQuickbarSwapTargetRejectsUnrelatedOrStaleRecords(t *testing.T) {
	for _, mode := range []string{"spell", "caster", "stale", "future", "dead-target", "missing-target", "invalid-wire", "expired-window"} {
		t.Run(mode, func(t *testing.T) {
			hp := server.HealthData{Cur: 1}
			unit := server.Object{ObjClass: object.ClassPlayer, HealthData: &hp}
			target := server.Object{ObjClass: object.ClassMonster, HealthData: &hp, NetCode: 13}
			record := server.DurSpell{Spell: uint32(spell.SPELL_SWAP), Caster16: &unit, Target48: &target, Frame60: 20}
			frame := uint32(25)
			switch mode {
			case "spell":
				record.Spell = uint32(spell.SPELL_BLINK)
			case "caster":
				record.Caster16 = &server.Object{}
			case "stale":
				record.Frame60 = 9
			case "future":
				record.Frame60 = 26
			case "dead-target":
				target.ObjFlags = object.FlagDead
			case "missing-target":
				record.Target48 = nil
			case "invalid-wire":
				target.NetCode = 0x8000
			case "expired-window":
				frame = 131
			}
			if got, ok := e2eReadMonsterQuickbarSwapTarget(frame, 10, &unit, &record, &server.Server{}); ok || got != (e2eMonsterQuickbarSwapTarget{}) {
				t.Fatal("an unrelated, stale or retired native record cannot supply actual-target evidence")
			}
		})
	}
}

func TestMonsterQuickbarSwapTargetRuntimeCallsOnlyReaders(t *testing.T) {
	file, err := parser.ParseFile(token.NewFileSet(), "e2e_monster_quickbar_swap_observe.go", nil, 0)
	if err != nil {
		t.Fatal(err)
	}
	allowed := map[string]bool{}
	for _, name := range []string{"Class", "Has", "Flags", "HasAny", "uint32", "uint16", "GetUnitNetCode", "e2eReadTargetObject",
		"HostUnit", "S", "e2eReadMonsterQuickbarSwapTarget", "Printf"} {
		allowed[name] = true
	}
	ast.Inspect(file, func(node ast.Node) bool {
		if call, ok := node.(*ast.CallExpr); ok {
			name := ""
			switch function := call.Fun.(type) {
			case *ast.Ident:
				name = function.Name
			case *ast.SelectorExpr:
				name = function.Sel.Name
			}
			if !allowed[name] {
				t.Errorf("actual Swap observation invokes unreviewed operation %q", name)
			}
		}
		return true
	})
}
