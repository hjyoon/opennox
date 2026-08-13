package opennox

import (
	"github.com/opennox/libs/object"
	"github.com/opennox/libs/types"
)

// targetSearchArg4E6EA0 is the native counterpart of the original ten-word
// search record. Source36 replaces the 32-bit object address stored at offset
// +36; this structure is runtime-only and is never serialized as the Win32
// layout.
type targetSearchArg4E6EA0[T comparable] struct {
	Field0             uint32
	Field4             uint32
	Field8             uint32
	ClassAllow12       object.Class
	ClassDisallow16    object.Class
	SubClassAllow20    object.SubClass
	SubClassDisallow24 object.SubClass
	FlagsAllow28       object.Flags
	FlagsDisallow32    object.Flags
	Source36           T
}

type targetSearch4E6EA0Hooks[T comparable] struct {
	eachInCircle func(types.Pointf, float32, func(T) bool)
	class        func(T) object.Class
	subClass     func(T) object.SubClass
	flags        func(T) object.Flags
	position     func(T) types.Pointf
	directionInd func(T) int16
	sameTeam     func(T, T) bool
	playerStatus func(T) uint32
	isEnemy      func(T, T) bool
	direction    func(types.Pointf, int16, types.Pointf) uint32
	canInteract  func(T, T, int) bool
}

// targetSearch4E6EA0 reproduces the setup at 004E6EA0. GAME.EXE rounds the
// squared radius through a float32 global before iteration and stores the
// source in the search record. The native implementation keeps both pieces of
// scratch state local while preserving those observable value semantics.
func targetSearch4E6EA0[T comparable](source T, radius float32, arg *targetSearchArg4E6EA0[T], h targetSearch4E6EA0Hooks[T]) T {
	if arg == nil {
		var zero T
		return zero
	}
	var found T
	best := radius * radius
	arg.Source36 = source
	h.eachInCircle(h.position(source), radius, func(candidate T) bool {
		targetSearchCandidate4E6EF0(candidate, arg, &best, &found, h)
		return true
	})
	return found
}

// targetSearchCandidate4E6EF0 preserves the original callback's short-circuit
// order and its repeated loads of Source36 around calls. Those reloads matter
// when a called boundary mutates the search record. Distance arithmetic is not
// prematurely rounded to float32, while each accepted best value is rounded
// back through the original float32 scratch slot.
func targetSearchCandidate4E6EF0[T comparable](candidate T, arg *targetSearchArg4E6EA0[T], best *float32, found *T, h targetSearch4E6EA0Hooks[T]) {
	if uint8(h.flags(candidate))&uint8(object.FlagDestroyed) != 0 {
		return
	}
	if h.sameTeam(candidate, arg.Source36) {
		return
	}
	if h.class(candidate)&object.ClassPlayer != 0 && h.playerStatus(candidate)&1 != 0 {
		return
	}
	if arg.Field8 != 0 && !h.isEnemy(arg.Source36, candidate) {
		return
	}

	source := arg.Source36
	direction := h.direction(h.position(source), h.directionInd(source), h.position(candidate))
	if direction&arg.Field0 == 0 {
		return
	}

	source = arg.Source36
	if candidate == source {
		return
	}
	if arg.Field4 != 0 && !h.canInteract(source, candidate, 0) {
		return
	}

	class := h.class(candidate)
	if class&arg.ClassAllow12 == 0 || class&arg.ClassDisallow16 != 0 {
		return
	}
	flags := h.flags(candidate)
	if flags&arg.FlagsAllow28 == 0 || flags&arg.FlagsDisallow32 != 0 {
		return
	}
	subClass := h.subClass(candidate)
	if subClass != 0 && (subClass&arg.SubClassAllow20 == 0 || subClass&arg.SubClassDisallow24 != 0) {
		return
	}

	// Both positions originate as float32. Convert before subtraction so the
	// differences and products are not prematurely rounded back to float32.
	source = arg.Source36
	candidatePos := h.position(candidate)
	sourcePos := h.position(source)
	dx := float64(candidatePos.X) - float64(sourcePos.X)
	dy := float64(candidatePos.Y) - float64(sourcePos.Y)
	distance := dx*dx + dy*dy
	// GAME.EXE tests x87 C0 only. C0 is set for less-than and unordered, so
	// NaN in either operand selects this candidate. FSTP then rounds to float32.
	if !(distance >= float64(*best)) {
		*best = float32(distance)
		*found = candidate
	}
}
