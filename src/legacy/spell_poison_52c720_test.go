package legacy

import (
	"math"
	"testing"
	"unsafe"

	"github.com/opennox/libs/object"
	"github.com/opennox/libs/prand"
	"github.com/opennox/libs/spell"
	"github.com/opennox/libs/types"

	"github.com/opennox/opennox/v1/legacy/common/alloc"
	"github.com/opennox/opennox/v1/server"
)

type poisonCastLegacyServer52C720 struct {
	Server
	srv *server.Server
}

func (s *poisonCastLegacyServer52C720) S() *server.Server { return s.srv }

// The old entry loaded arg.Obj's low DWORD as a C int before calling the
// already-native poison service. An immune monster requires no RNG or
// reporting, but still exposes the pointer truncation at that first load.
func TestPoisonCastGoEntryPreservesTarget52C720(t *testing.T) {
	target, freeTarget := alloc.New(server.Object{})
	defer freeTarget()
	arg, freeArg := alloc.New(server.SpellAcceptArg{})
	defer freeArg()
	*target = server.Object{ObjClass: object.ClassMonster, ObjSubClass: object.SubClass(0x200), Poison540: 7}
	*arg = server.SpellAcceptArg{Obj: target, Pos: types.Ptf(-1000, 2000)}
	before := *arg
	if unsafe.Sizeof(uintptr(0)) == 8 {
		for _, pointer := range []unsafe.Pointer{unsafe.Pointer(target), unsafe.Pointer(arg)} {
			if uintptr(pointer) <= math.MaxUint32 {
				t.Fatalf("Poison pointer=%p, want >4 GiB", pointer)
			}
		}
	}
	previous := GetServer
	t.Cleanup(func() { GetServer = previous })
	GetServer = func() Server { return &poisonCastLegacyServer52C720{srv: new(server.Server)} }
	if got := Nox_xxx_castPoison_52C720(spell.SPELL_POISON, nil, nil, nil, arg, 3); got != 1 || *arg != before || target.Poison540 != 7 {
		t.Fatalf("Poison result/arg/state=%d/%+v/%d", got, *arg, target.Poison540)
	}
}

func TestPoisonCastGoEntryUsesNativePoisonState52C720(t *testing.T) {
	target, freeTarget := alloc.New(server.Object{})
	defer freeTarget()
	health, freeHealth := alloc.New(server.HealthData{})
	defer freeHealth()
	arg, freeArg := alloc.New(server.SpellAcceptArg{})
	defer freeArg()
	*target = server.Object{HealthData: health}
	*arg = server.SpellAcceptArg{Obj: target, Pos: types.Ptf(-1, 2)}
	s := new(server.Server)
	s.Rand.Logic = prand.New(0)
	s.SetFrame(91)
	previous := GetServer
	t.Cleanup(func() { GetServer = previous })
	GetServer = func() Server { return &poisonCastLegacyServer52C720{srv: s} }
	if got := Nox_xxx_castPoison_52C720(spell.SPELL_POISON, nil, nil, nil, arg, 3); got != 1 ||
		target.Poison540 != 3 || target.Field542 != 1000 || health.Field16 != 91 || s.Rand.Logic.Index() != 1 {
		t.Fatalf("Poison result/state/timer/frame/RNG=%d/%d/%d/%d/%d", got, target.Poison540, target.Field542, health.Field16, s.Rand.Logic.Index())
	}
}

func TestPoisonCastGoEntryRecordsPlayerAttributionAfterRejection52C720(t *testing.T) {
	source, freeSource := alloc.New(server.Object{})
	defer freeSource()
	target, freeTarget := alloc.New(server.Object{})
	defer freeTarget()
	sourceUpdate, freeSourceUpdate := alloc.New(server.PlayerUpdateData{})
	defer freeSourceUpdate()
	targetUpdate, freeTargetUpdate := alloc.New(server.PlayerUpdateData{})
	defer freeTargetUpdate()
	sourcePlayer, freeSourcePlayer := alloc.New(server.Player{})
	defer freeSourcePlayer()
	targetPlayer, freeTargetPlayer := alloc.New(server.Player{})
	defer freeTargetPlayer()
	arg, freeArg := alloc.New(server.SpellAcceptArg{})
	defer freeArg()
	sourcePlayer.PlayerInd = 0xfe
	targetPlayer.Field3680 = 1 // the original observer gate rejects activation
	sourceUpdate.Player, targetUpdate.Player = sourcePlayer, targetPlayer
	*source = server.Object{ObjClass: object.ClassPlayer, UpdateData: unsafe.Pointer(sourceUpdate)}
	*target = server.Object{ObjClass: object.ClassPlayer, UpdateData: unsafe.Pointer(targetUpdate), Poison540: 7}
	arg.Obj = target
	previousServer, previousFrame := GetServer, gameFrameHook
	t.Cleanup(func() { GetServer, gameFrameHook = previousServer, previousFrame })
	GetServer = func() Server { return &poisonCastLegacyServer52C720{srv: new(server.Server)} }
	gameFrameHook = func() uint32 { return 0xfedcba98 }
	if got := Nox_xxx_castPoison_52C720(spell.SPELL_POISON, nil, source, nil, arg, -3); got != 1 || target.Poison540 != 7 {
		t.Fatalf("Poison rejection result/state=%d/%d", got, target.Poison540)
	}
	pending, index, frame := targetPlayer.LastAggressorState()
	if pending != 1 || index != 0xfe || frame != 0xfedcba98 {
		t.Fatalf("Poison rejected-target attribution=%d/%#x/%#x", pending, index, frame)
	}
}

func TestPoisonCastGoDispatchPreservesNativeArguments52C720(t *testing.T) {
	second, freeSecond := alloc.New(server.Object{})
	defer freeSecond()
	owner, freeOwner := alloc.New(server.Object{})
	defer freeOwner()
	caster, freeCaster := alloc.New(server.Object{})
	defer freeCaster()
	arg, freeArg := alloc.New(server.SpellAcceptArg{})
	defer freeArg()
	*arg = server.SpellAcceptArg{Obj: second, Pos: types.Ptf(-1, 2)}
	before := *arg
	previous := poisonCastCall52C720
	t.Cleanup(func() { poisonCastCall52C720 = previous })
	calls := 0
	poisonCastCall52C720 = func(id spell.ID, gotSecond, gotOwner, gotCaster *server.Object, gotArg *server.SpellAcceptArg, level int) int {
		calls++
		if id != spell.SPELL_POISON || gotSecond != second || gotOwner != owner || gotCaster != caster || gotArg != arg || level != -3 {
			t.Fatalf("Poison dispatch=%s/%p/%p/%p/%p/%d", id, gotSecond, gotOwner, gotCaster, gotArg, level)
		}
		return -17
	}
	if got := Nox_xxx_castPoison_52C720(spell.SPELL_POISON, second, owner, caster, arg, -3); got != -17 || calls != 1 || *arg != before {
		t.Fatalf("Poison result/calls/arg=%d/%d/%+v", got, calls, *arg)
	}
}
