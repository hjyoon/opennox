package server

import (
	"github.com/opennox/libs/spell"
	"github.com/opennox/libs/types"

	"github.com/opennox/opennox/v1/common/sound"
)

// markSlotCast52CBD0 restores GAME.EXE 0052CBD0..0052CCCF. Unlike Mark's
// first-empty/oldest selection, Mark 1..4 always address the named slot. The
// selector supplies IDs 46..49; objects, update data and slot pointers retain
// native width, while the original shared cache and packed charges stay DWORDs.
func markSlotCast52CBD0(id int32, caster, aim *Object, h markCastHooks52CA80) int32 {
	glyph := h.loadCache()
	if glyph == 0 {
		glyph = h.lookupType("Glyph")
		h.storeCache(glyph)
	}
	positionObject := caster
	if aim != nil && uint32(aim.TypeInd) == glyph {
		positionObject = aim
	}
	if uint8(caster.ObjClass)&4 == 0 {
		return 1
	}
	position := &positionObject.PosVec
	data := (*PlayerUpdateData)(caster.UpdateData)
	index := int(id - int32(spell.SPELL_MARK_1))
	marker := data.Field29[index]
	if marker != nil {
		h.move(marker, position)
		// 0052CC31 reloads from the entry-cached player record before the
		// 0052CC35 frame read, even if movement replaced the live record.
		marker = data.Field29[index]
		frame := h.frame()
		marker.Field34 = frame
	} else {
		names := [4]string{"TeleportGlyph1", "TeleportGlyph2", "TeleportGlyph3", "TeleportGlyph4"}
		marker = h.newObject(names[index])
		data.Field29[index] = marker
		if marker == nil {
			h.audio(h.castSound(id), caster, 0, 0)
			return 1 // failed allocation sounds but does not recharge
		}
		// Preserve the entry-selected point address and the live Y/X loads
		// after allocation, rather than snapshotting before its callback.
		y, x := position.Y, position.X
		h.createAt(marker, caster, types.Ptf(x, y))
	}
	shift := uint(index * 8)
	data.Field39 = data.Field39&^(uint32(0xff)<<shift) | uint32(3)<<shift
	h.audio(h.castSound(id), caster, 0, 0)
	return 1
}

// The two original Mark routines use the same Glyph cache and movement/
// creation services; their runtime layout is intentionally shared.
func (s *Server) CastMarkSlot52CBD0(id int32, _ *Object, caster, aim *Object, _ *SpellAcceptArg, _ int32, runtime MarkCastRuntime52CA80) int32 {
	return markSlotCast52CBD0(id, caster, aim, markCastHooks52CA80{
		loadCache:  func() uint32 { return *runtime.GlyphTypeCache },
		storeCache: func(kind uint32) { *runtime.GlyphTypeCache = kind },
		lookupType: func(name string) uint32 { return uint32(s.Types.IndByID(name)) },
		frame:      s.Frame,
		move:       runtime.Move,
		newObject:  s.NewObjectByTypeID,
		createAt:   runtime.CreateAt,
		castSound:  func(id int32) sound.ID { return s.Spells.DefByInd(spell.ID(id)).GetCastSound() },
		audio:      func(id sound.ID, marker *Object, kind int, code uint32) { s.Audio.EventObj(id, marker, kind, code) },
	})
}
