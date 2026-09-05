package server

import (
	"github.com/opennox/libs/spell"
	"github.com/opennox/libs/things"

	"github.com/opennox/opennox/v1/common/sound"
)

// SpellAcceptRuntime4FD400 supplies the cast effects that remain owned by the
// outer game runtime. Object and SpellAcceptArg pointers stay native-width in
// the server package; spell IDs, levels, timeouts and return codes retain the
// original fixed-width contract.
type SpellAcceptRuntime4FD400 struct {
	CaptureMagic func(spell.ID, *Object) int32
	Audio        func(sound.ID, *Object)
	Instant      func(spell.ID, *Object, *Object, *Object, *SpellAcceptArg, int32) int32
	Duration     func(spell.ID, *Object, *Object, *Object, *SpellAcceptArg, int32, uint32) int32
	PlasmaTime   func() float64
}

type spellAcceptNativeDeps4FD400 struct {
	spellHasFlags func(int32, uint32) int32
	tickRate      func() uint32
	runtime       SpellAcceptRuntime4FD400
}

func spellAcceptNative4FD400(
	spellID int32,
	second, third, fourth *Object,
	arg *SpellAcceptArg,
	level int32,
	deps spellAcceptNativeDeps4FD400,
) int32 {
	return spellAccept4FD400(spellAcceptHooks4FD400[*Object, *SpellAcceptArg]{
		loadSpellArg:  func() int32 { return spellID },
		loadThirdArg:  func() *Object { return third },
		loadSecondArg: func() *Object { return second },
		loadAcceptArg: func() *SpellAcceptArg { return arg },
		spellHasFlags: deps.spellHasFlags,
		loadTarget: func(arg *SpellAcceptArg) *Object {
			return arg.Obj
		},
		loadClassLow: func(obj *Object) uint8 {
			return uint8(obj.ObjClass)
		},
		captureMagic: func(id int32, obj *Object) int32 {
			return deps.runtime.CaptureMagic(spell.ID(id), obj)
		},
		audio: func(id int32, obj *Object, _, _ int32) {
			deps.runtime.Audio(sound.ID(id), obj)
		},
		loadLevelArg:  func() int32 { return level },
		loadFourthArg: func() *Object { return fourth },
		tickRate:      deps.tickRate,
		plasmaTime:    deps.runtime.PlasmaTime,
		instant: func(id int32, second, third, fourth *Object, arg *SpellAcceptArg, level int32) int32 {
			return deps.runtime.Instant(spell.ID(id), second, third, fourth, arg, level)
		},
		duration: func(
			_ spellAcceptDispatch4FD400,
			id int32,
			second, third, fourth *Object,
			arg *SpellAcceptArg,
			level int32,
			timeout uint32,
		) int32 {
			return deps.runtime.Duration(spell.ID(id), second, third, fourth, arg, level, timeout)
		},
	})
}

func spellAcceptServerDeps4FD400(s *Server, runtime SpellAcceptRuntime4FD400) spellAcceptNativeDeps4FD400 {
	return spellAcceptNativeDeps4FD400{
		spellHasFlags: func(spellID int32, mask uint32) int32 {
			if s.Spells.HasFlags(spell.ID(spellID), things.SpellFlags(mask)) {
				return 1
			}
			return 0
		},
		tickRate: s.TickRate,
		runtime:  runtime,
	}
}

// SpellAccept4FD400 binds GAME.EXE 004FD400 to native-width Object and
// SpellAcceptArg pointers. The entry guards are part of the original function;
// no additional definition, target, level or result normalization is applied.
//
//go:noinline
func (s *Server) SpellAccept4FD400(
	spellID spell.ID,
	second, third, fourth *Object,
	arg *SpellAcceptArg,
	level int32,
	runtime SpellAcceptRuntime4FD400,
) int32 {
	return spellAcceptNative4FD400(
		int32(spellID), second, third, fourth, arg, level,
		spellAcceptServerDeps4FD400(s, runtime),
	)
}
