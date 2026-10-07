package legacy

import (
	"github.com/opennox/libs/spell"

	"github.com/opennox/opennox/v1/server"
)

func castTriggerGlyphNative52CCD0(id spell.ID, second, caster, fourth *server.Object, arg *server.SpellAcceptArg, level int) int {
	return int(GetServer().S().CastTriggerGlyph52CCD0(int32(id), second, caster, fourth, arg, int32(level), server.TriggerGlyphCastRuntime52CCD0{
		DieGlyph: Nox_xxx_dieGlyph_54DF30,
	}))
}

var triggerGlyphCastCall52CCD0 = castTriggerGlyphNative52CCD0
