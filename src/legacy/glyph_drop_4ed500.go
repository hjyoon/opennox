package legacy

import (
	"github.com/opennox/libs/types"

	"github.com/opennox/opennox/v1/server"
)

func glyphDropCall4ED500(
	owner, glyph *server.Object,
	point *types.Pointf,
) int32 {
	outer := GetServer()
	return outer.S().GlyphDrop4ED500(owner, glyph, point, server.GlyphDropRuntime4ED500{
		DropTrap: trapDropCall4ED580,
	})
}
