package legacy

import (
	"math"
	"testing"
	"unsafe"

	"github.com/opennox/libs/object"
	"github.com/opennox/libs/spell"
	"github.com/opennox/libs/strman"
	"github.com/opennox/libs/types"

	"github.com/opennox/opennox/v1/legacy/common/alloc"
	"github.com/opennox/opennox/v1/server"
)

type castConfuseLegacyServer52C1E0 struct {
	Server
	srv *server.Server
}

func TestCastConfuseCEntry52C1E0PreservesCasterAndLiveAttribution(t *testing.T) {
	srv := server.New(nil, nil, strman.New())
	t.Cleanup(srv.Close)
	oldServer, oldApply, oldFrame := GetServer, buffApplyExportImpl4FF380, gameFrameHook
	GetServer = func() Server { return &castConfuseLegacyServer52C1E0{srv: srv} }
	gameFrameHook = func() uint32 { return 0xfedcba98 }
	t.Cleanup(func() { GetServer, buffApplyExportImpl4FF380, gameFrameHook = oldServer, oldApply, oldFrame })
	source, freeSource := alloc.New(server.Object{})
	target, freeTarget := alloc.New(server.Object{})
	reloaded, freeReloaded := alloc.New(server.Object{})
	sourceUpdate, freeSourceUpdate := alloc.New(server.PlayerUpdateData{})
	targetUpdate, freeTargetUpdate := alloc.New(server.PlayerUpdateData{})
	reloadedUpdate, freeReloadedUpdate := alloc.New(server.PlayerUpdateData{})
	sourcePlayer, freeSourcePlayer := alloc.New(server.Player{})
	targetPlayer, freeTargetPlayer := alloc.New(server.Player{})
	reloadedPlayer, freeReloadedPlayer := alloc.New(server.Player{})
	arg, freeArg := alloc.New(server.SpellAcceptArg{})
	for _, free := range []func(){freeSource, freeTarget, freeReloaded, freeSourceUpdate, freeTargetUpdate, freeReloadedUpdate, freeSourcePlayer, freeTargetPlayer, freeReloadedPlayer, freeArg} {
		t.Cleanup(free)
	}
	*source = server.Object{ObjClass: object.ClassPlayer, UpdateData: unsafe.Pointer(sourceUpdate)}
	*target = server.Object{ObjClass: object.ClassPlayer, UpdateData: unsafe.Pointer(targetUpdate)}
	*reloaded = server.Object{ObjClass: object.ClassPlayer, UpdateData: unsafe.Pointer(reloadedUpdate)}
	sourceUpdate.Player, targetUpdate.Player, reloadedUpdate.Player = sourcePlayer, targetPlayer, reloadedPlayer
	sourcePlayer.PlayerInd = 0xfe
	*arg = server.SpellAcceptArg{Obj: target, Pos: types.Pointf{X: -33.5, Y: 12.25}}
	if unsafe.Sizeof(uintptr(0)) > 4 {
		for _, ptr := range []unsafe.Pointer{unsafe.Pointer(source), unsafe.Pointer(target), unsafe.Pointer(reloaded), unsafe.Pointer(sourceUpdate), unsafe.Pointer(targetUpdate), unsafe.Pointer(reloadedUpdate), unsafe.Pointer(sourcePlayer), unsafe.Pointer(targetPlayer), unsafe.Pointer(reloadedPlayer), unsafe.Pointer(arg)} {
			if uintptr(ptr) <= math.MaxUint32 {
				t.Fatalf("native pointer = %p, want above 4 GiB", ptr)
			}
		}
	}
	calls := 0
	buffApplyExportImpl4FF380 = func(unit *server.Object, buff int32, duration int16, power int8) {
		calls++
		if unit != target || buff != 3 || duration != 0 || power != 127 {
			t.Fatalf("buff call = %p/%d/%d/%d", unit, buff, duration, power)
		}
		arg.Obj = reloaded
	}
	if got := Nox_xxx_castConfuse_52C1E0(spell.SPELL_CONFUSE, reloaded, source, target, arg, -129); got != 1 || calls != 1 {
		t.Fatalf("C entry = %d, buff calls = %d", got, calls)
	}
	if pending, index, frame := reloadedPlayer.LastAggressorState(); pending != 1 || index != 0xfe || frame != 0xfedcba98 {
		t.Fatalf("live attribution = %d/%x/%x", pending, index, frame)
	}
	if pending, index, frame := targetPlayer.LastAggressorState(); pending != 0 || index != 0 || frame != 0 {
		t.Fatalf("stale target attribution = %d/%x/%x", pending, index, frame)
	}
	if arg.Obj != reloaded || arg.Pos != (types.Pointf{X: -33.5, Y: 12.25}) {
		t.Fatal("unexpected spell argument change")
	}
}

func TestCastConfuseCEntry52C1E0DelegatesSignedDwords(t *testing.T) {
	old := castConfuseCall52C1E0
	t.Cleanup(func() { castConfuseCall52C1E0 = old })
	source, freeSource := alloc.New(server.Object{})
	arg, freeArg := alloc.New(server.SpellAcceptArg{})
	t.Cleanup(freeSource)
	t.Cleanup(freeArg)
	calls := 0
	castConfuseCall52C1E0 = func(gotSource *server.Object, gotArg *server.SpellAcceptArg, power int32) int32 {
		calls++
		if gotSource != source || gotArg != arg || power != math.MinInt32 {
			t.Fatalf("export = %p/%p/%d", gotSource, gotArg, power)
		}
		return math.MinInt32
	}
	if got := Nox_xxx_castConfuse_52C1E0(spell.ID(math.MinInt32), nil, source, nil, arg, math.MinInt32); got != math.MinInt32 || calls != 1 {
		t.Fatalf("export result/calls = %d/%d", got, calls)
	}
}

func (s *castConfuseLegacyServer52C1E0) S() *server.Server { return s.srv }

func TestCastConfuseCEntry52C1E0PreservesNativeTargetAndPower(t *testing.T) {
	srv := server.New(nil, nil, strman.New())
	t.Cleanup(srv.Close)
	oldServer, oldApply := GetServer, buffApplyExportImpl4FF380
	GetServer = func() Server { return &castConfuseLegacyServer52C1E0{srv: srv} }
	t.Cleanup(func() { GetServer, buffApplyExportImpl4FF380 = oldServer, oldApply })
	target, freeTarget := alloc.New(server.Object{})
	arg, freeArg := alloc.New(server.SpellAcceptArg{})
	t.Cleanup(freeTarget)
	t.Cleanup(freeArg)
	if unsafe.Sizeof(uintptr(0)) > 4 {
		for _, ptr := range []unsafe.Pointer{unsafe.Pointer(target), unsafe.Pointer(arg)} {
			if uintptr(ptr) <= math.MaxUint32 {
				t.Fatalf("native pointer = %p, want above 4 GiB", ptr)
			}
		}
	}
	*arg = server.SpellAcceptArg{Obj: target, Pos: types.Pointf{X: -12.5, Y: 33.25}}
	wantArg := *arg
	calls := 0
	buffApplyExportImpl4FF380 = func(unit *server.Object, buff int32, duration int16, power int8) {
		calls++
		if unit != target || buff != 3 || duration != 0 || power != -128 {
			t.Errorf("buff call = %p/%d/%d/%d, want %p/3/0/-128", unit, buff, duration, power, target)
		}
	}
	// A nil caster keeps attribution inert while the actual C entry point and
	// buff C-to-Go export expose any narrowing of SpellAcceptArg.Obj safely.
	if got := Nox_xxx_castConfuse_52C1E0(spell.SPELL_CONFUSE, nil, nil, nil, arg, 0x180); got != 1 || calls != 1 {
		t.Fatalf("C entry = %d, buff calls = %d, want 1/1", got, calls)
	}
	if *arg != wantArg {
		t.Fatalf("C entry changed spell argument: %+v, want %+v", *arg, wantArg)
	}
	arg.Obj = nil
	if got := Nox_xxx_castConfuse_52C1E0(spell.SPELL_CONFUSE, nil, nil, nil, arg, 1); got != 0 || calls != 1 {
		t.Fatalf("nil target C entry = %d, buff calls = %d, want 0/1", got, calls)
	}
}
