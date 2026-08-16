package server

import (
	"unsafe"

	"github.com/opennox/libs/types"

	"github.com/opennox/opennox/v1/common/sound"
)

// GlyphDropRuntime4ED500 supplies the TrapDrop dependency that is still owned
// by the legacy drop cluster. Object, GlyphInitData, direction, and audio
// operations remain native-width in server.
type GlyphDropRuntime4ED500 struct {
	DropTrap func(*Object, *Object, *types.Pointf) int32
}

type glyphDropNativeDeps4ED500 struct {
	dropTrap func(*Object, *Object, *types.Pointf) int32
	audio    func(uint32, *Object, int32, uint32)
}

func glyphDropNative4ED500(
	owner, glyph *Object,
	point *types.Pointf,
	deps glyphDropNativeDeps4ED500,
) int32 {
	return glyphDrop4ED500(owner, glyph, point, glyphDropHooks4ED500[
		*Object,
		*GlyphInitData,
		*types.Pointf,
	]{
		dropTrap: deps.dropTrap,
		loadInitData: func(glyph *Object) *GlyphInitData {
			return (*GlyphInitData)(glyph.InitData)
		},
		loadPointX: func(point *types.Pointf) float32 {
			return point.X
		},
		storeGlyphX: func(data *GlyphInitData, value float32) {
			data.SpellArg.Pos.X = value
		},
		loadPointY: func(point *types.Pointf) float32 {
			return point.Y
		},
		storeGlyphY: func(data *GlyphInitData, value float32) {
			data.SpellArg.Pos.Y = value
		},
		loadObjectX: func(owner *Object) float32 {
			return owner.PosVec.X
		},
		loadObjectY: func(owner *Object) float32 {
			return owner.PosVec.Y
		},
		vectorDirection: directionFromVector509ED0,
		storeDirection2: func(glyph *Object, direction uint16) {
			glyph.Direction2 = Dir16(direction)
		},
		storeDirection1: func(glyph *Object, direction uint16) {
			glyph.Direction1 = Dir16(direction)
		},
		audio: deps.audio,
	})
}

// GlyphDrop4ED500 binds GAME.EXE 004ED500 to native Object, GlyphInitData, and
// Pointf pointers. TrapDrop remains an explicit dependency until 004ED580 is
// restored separately.
func (s *Server) GlyphDrop4ED500(
	owner, glyph *Object,
	point *types.Pointf,
	runtime GlyphDropRuntime4ED500,
) int32 {
	return glyphDropNative4ED500(owner, glyph, point, glyphDropNativeDeps4ED500{
		dropTrap: runtime.DropTrap,
		audio: func(id uint32, obj *Object, kind int32, code uint32) {
			s.Audio.EventObj(sound.ID(id), obj, int(kind), code)
		},
	})
}

var (
	_ = [1]struct{}{}[0-unsafe.Offsetof(GlyphInitData{}.Spells)]
	_ = [1]struct{}{}[20-unsafe.Offsetof(GlyphInitData{}.SpellsCnt)]
	_ = [1]struct{}{}[24-unsafe.Offsetof(GlyphInitData{}.SpellArg)]
)
