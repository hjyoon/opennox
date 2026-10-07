package server

import (
	"github.com/opennox/libs/spell"
	"github.com/opennox/libs/types"

	"github.com/opennox/opennox/v1/common/sound"
)

type triggerGlyphCastHooks52CCD0 struct {
	first     func() *Object
	next      func(*Object) *Object
	hasParent func(*Object, *Object) bool
	typeName  func(*Object) string
	castSound func(int32) sound.ID
	audio     func(sound.ID, *Object, int, uint32)
	dieGlyph  func(*Object)
}

// 0052CD1D compares six bytes, including the name's terminating NUL.
// Preserve the C string boundary if a type identifier contains an embedded NUL.
func triggerGlyphTypeName52CCD0(name string) bool {
	return name == "Glyph" || len(name) >= 6 && name[:6] == "Glyph\x00"
}

// 0052CD23..0052CD48 retains precision-53/chop differences and squares,
// adds Y-square then X-square without contraction, compares the register
// against the previous binary32 winner, and spills only a new winner.
func triggerGlyphDistance52CCD0(caster, glyph types.Pointf) float64 {
	dx := monsterMoveToRunAddChop53_544434(float64(caster.X), -float64(glyph.X))
	dy := monsterMoveToRunAddChop53_544434(float64(caster.Y), -float64(glyph.Y))
	ySquare := monsterMoveToRunSquareChop53_544434(dy)
	xSquare := monsterMoveToRunSquareChop53_544434(dx)
	return monsterMoveToRunAddChop53_544434(ySquare, xSquare)
}

// triggerGlyphCast52CCD0 restores GAME.EXE 0052CCD0..0052CDA0. Traverse
// the active world in its original order, including destroyed objects and
// the caster itself. Ownership precedes type and live position loads. Do
// not substitute an owned-list, spatial, visibility, class or flag filter.
func triggerGlyphCast52CCD0(id int32, caster *Object, h triggerGlyphCastHooks52CCD0) int32 {
	nearest := float32(100000000)
	var selected *Object
	for obj := h.first(); obj != nil; obj = h.next(obj) {
		if !h.hasParent(obj, caster) || !triggerGlyphTypeName52CCD0(h.typeName(obj)) {
			continue
		}
		distance := triggerGlyphDistance52CCD0(caster.PosVec, obj.PosVec)
		// FCOM tests C0 only. Less and unordered both replace the winner;
		// an ordered tie retains the earlier world object.
		if !(distance >= float64(nearest)) {
			nearest = float32(monsterMoveToRunSpill544440(distance))
			selected = obj
		}
	}
	if selected == nil {
		return 0
	}
	audio := h.castSound(id)
	h.audio(audio, caster, 0, 0)
	h.dieGlyph(selected)
	return 1
}

type TriggerGlyphCastRuntime52CCD0 struct {
	DieGlyph func(*Object)
}

// CastTriggerGlyph52CCD0 retains all six selector arguments at native
// pointer width. Only the spell ID and third object are used originally.
func (s *Server) CastTriggerGlyph52CCD0(id int32, _ *Object, caster *Object, _ *Object, _ *SpellAcceptArg, _ int32, runtime TriggerGlyphCastRuntime52CCD0) int32 {
	return triggerGlyphCast52CCD0(id, caster, triggerGlyphCastHooks52CCD0{
		first:     s.Objs.First,
		next:      (*Object).Next,
		hasParent: (*Object).HasOwner,
		typeName:  func(obj *Object) string { return s.Types.ByInd(int(obj.TypeInd)).ID() },
		castSound: func(id int32) sound.ID { return s.Spells.DefByInd(spell.ID(id)).GetCastSound() },
		audio:     func(id sound.ID, obj *Object, kind int, code uint32) { s.Audio.EventObj(id, obj, kind, code) },
		dieGlyph:  runtime.DieGlyph,
	})
}
