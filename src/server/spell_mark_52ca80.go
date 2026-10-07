package server

import (
	"unsafe"

	"github.com/opennox/libs/spell"
	"github.com/opennox/libs/types"

	"github.com/opennox/opennox/v1/common/sound"
)

type markCastHooks52CA80 struct {
	loadCache  func() uint32
	storeCache func(uint32)
	lookupType func(string) uint32
	frame      func() uint32
	move       func(*Object, *types.Pointf)
	newObject  func(string) *Object
	createAt   func(*Object, *Object, types.Pointf)
	castSound  func(int32) sound.ID
	audio      func(sound.ID, *Object, int, uint32)
}

// markCast52CA80 restores GAME.EXE 0052CA80..0052CBC2. Object, player-data,
// and marker pointers remain native-width; the shared Glyph cache, timestamps,
// and four packed charge bytes retain their original widths and load order.
func markCast52CA80(id int32, caster, aim *Object, h markCastHooks52CA80) int32 {
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
	index := 0
	for index < 4 && data.Field29[index] != nil {
		index++
	}
	if index == 4 {
		oldest := h.frame()
		// 0052CAF4 retains the caster DWORD as the invalid-index fallback
		// when no timestamp is strictly older than the current frame. Do
		// not silently select slot zero or turn its required fault into
		// success. This value is only an array index, never a pointer.
		index = int(uint32(uintptr(unsafe.Pointer(caster))))
		for i := 0; i < 4; i++ {
			stamp := data.Field29[i].Field34
			if stamp < oldest {
				oldest, index = stamp, i
			}
		}
	}
	marker := data.Field29[index]
	if marker != nil {
		h.move(marker, position)
		// Reload the selected pointer from the cached update record,
		// before the frame read, after the movement callback returns.
		marker = data.Field29[index]
		frame := h.frame()
		marker.Field34 = frame
	} else {
		names := [4]string{"TeleportGlyph1", "TeleportGlyph2", "TeleportGlyph3", "TeleportGlyph4"}
		marker = h.newObject(names[index])
		data.Field29[index] = marker
		if marker == nil {
			h.audio(h.castSound(id), caster, 0, 0)
			return 1 // failed allocation still sounds, but does not recharge
		}
		// Keep the position address selected at entry, not a point value:
		// allocation may change its contents before the original Y/X loads.
		y, x := position.Y, position.X
		h.createAt(marker, caster, types.Ptf(x, y))
	}
	shift := uint(index * 8)
	data.Field39 = data.Field39&^(uint32(0xff)<<shift) | uint32(3)<<shift
	h.audio(h.castSound(id), caster, 0, 0)
	return 1
}

type MarkCastRuntime52CA80 struct {
	GlyphTypeCache *uint32
	Move           func(*Object, *types.Pointf)
	CreateAt       func(*Object, *Object, types.Pointf)
}

func (s *Server) CastMark52CA80(id int32, _ *Object, caster, aim *Object, _ *SpellAcceptArg, _ int32, runtime MarkCastRuntime52CA80) int32 {
	return markCast52CA80(id, caster, aim, markCastHooks52CA80{
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
