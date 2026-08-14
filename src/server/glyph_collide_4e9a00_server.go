package server

import (
	"unsafe"

	noxflags "github.com/opennox/opennox/v1/common/flags"
)

type GlyphCollideRuntime4E9A00 struct {
	Trigger func(*Object, *Object)
}

type glyphCollideGateNativeDeps4E9A30 struct {
	gameFlag        func(uint32) int32
	firstPlayerUnit func() *Object
	unitsOnSameTeam func(*Object, *Object) int32
	findParent      func(*Object) *Object
	abilityActive   func(*Object, int32) int32
}

func glyphCollideGateNative4E9A30(
	source, target *Object,
	deps glyphCollideGateNativeDeps4E9A30,
) int32 {
	return glyphCollideGate4E9A30(source, target, glyphCollideGateHooks4E9A30[*Object]{
		gameFlag:        deps.gameFlag,
		firstPlayerUnit: deps.firstPlayerUnit,
		loadFlags: func(obj *Object) uint32 {
			return uint32(obj.ObjFlags)
		},
		unitsOnSameTeam: deps.unitsOnSameTeam,
		findParent:      deps.findParent,
		classLow: func(obj *Object) uint8 {
			return uint8(obj.ObjClass)
		},
		abilityActive: deps.abilityActive,
	})
}

// GlyphCollideAllowed4E9A30 binds the shared trap eligibility helper to the
// server's native-width object, player-list, team, owner-chain, and ability
// representations.
func (s *Server) GlyphCollideAllowed4E9A30(source, target *Object) int32 {
	return glyphCollideGateNative4E9A30(source, target, glyphCollideGateNativeDeps4E9A30{
		gameFlag: func(flag uint32) int32 {
			if noxflags.HasGame(noxflags.GameFlag(flag)) {
				return 1
			}
			return 0
		},
		firstPlayerUnit: s.Players.FirstUnit,
		unitsOnSameTeam: func(first, second *Object) int32 {
			if UnitsHaveSameTeam4EC520(first, second) {
				return 1
			}
			return 0
		},
		findParent: (*Object).FindOwnerChainPlayer,
		abilityActive: func(obj *Object, ability int32) int32 {
			if s.Abils.IsActive(obj, Ability(ability)) {
				return 1
			}
			return 0
		},
	})
}

// GlyphCollide4E9A00 binds the zero-byte registered callback to native-width
// object pointers while leaving the trap effect at its existing runtime
// boundary.
func (s *Server) GlyphCollide4E9A00(
	source, target *Object,
	collision unsafe.Pointer,
	runtime GlyphCollideRuntime4E9A00,
) {
	glyphCollide4E9A00(source, target, collision, glyphCollideHooks4E9A00[*Object]{
		allowed: s.GlyphCollideAllowed4E9A30,
		trigger: runtime.Trigger,
	})
}
