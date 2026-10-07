package legacy

import (
	"github.com/opennox/libs/spell"

	"github.com/opennox/opennox/v1/common/memmap"
	"github.com/opennox/opennox/v1/server"
)

func castEarthquakeNative52DE40(id spell.ID, second, owner, caster *server.Object, arg *server.SpellAcceptArg, level int) int {
	return int(GetServer().S().CastEarthquake52DE40(int32(id), second, owner, caster, arg, int32(level), server.EarthquakeCastRuntime52DE40{
		LevelCache: memmap.PtrUint32(0x5D4594, 2487700),
	}))
}

var earthquakeCastCall52DE40 = castEarthquakeNative52DE40
