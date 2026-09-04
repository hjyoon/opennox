package opennox

import "github.com/opennox/opennox/v1/server"

type spellCancellationTraversalNativeDeps4FCEB0 struct {
	firstSpell  func() *server.DurSpell
	cancelSpell func(*server.DurSpell)
}

func spellCancellationTraversalNative4FCEB0(
	mode int32,
	deps spellCancellationTraversalNativeDeps4FCEB0,
) int32 {
	return spellCancellationTraversal4FCEB0(
		mode,
		spellCancellationTraversalHooks4FCEB0[*server.DurSpell, *server.Object]{
			firstSpell: deps.firstSpell,
			loadNext: func(current *server.DurSpell) *server.DurSpell {
				return current.Next
			},
			loadTarget: func(current *server.DurSpell) *server.Object {
				return current.Target48
			},
			loadTargetClass: func(target *server.Object) uint32 {
				return uint32(target.Class())
			},
			cancelSpell: deps.cancelSpell,
		},
	)
}

// SpellCancellationTraversal4FCEB0 binds GAME.EXE 004FCEB0 to the native
// duration-spell list. The original function accepts a full 32-bit scalar:
// only exact mode 1 preserves spells targeting Player-class objects.
//
//go:noinline
func (s *Server) SpellCancellationTraversal4FCEB0(mode int32) int32 {
	durations := &s.Server.Spells.Dur
	return spellCancellationTraversalNative4FCEB0(
		mode,
		spellCancellationTraversalNativeDeps4FCEB0{
			firstSpell: func() *server.DurSpell {
				return durations.List
			},
			cancelSpell: durations.CancelSpell,
		},
	)
}
