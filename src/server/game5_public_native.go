package server

import (
	"math"

	"github.com/opennox/libs/types"
	noxflags "github.com/opennox/opennox/v1/common/flags"
)

// PlayerCantTalkNative57A160 reads the native player record's chat restriction.
// The original public name says CanTalk, but bit 3 means chat is restricted.
func PlayerCantTalkNative57A160(pl *Player) bool {
	return pl != nil && noxflags.HasGame(noxflags.GameClient) && (pl.Field3680>>3)&1 != 0
}

// ReflectVelocityNative57B810 shares the original wall-normal reflection with
// the public C entry. The compare retains the x87 product without a binary32
// spill. On the swap path only X passes through FLD/FSTP; Y is copied by MOV.
func ReflectVelocityNative57B810(normal, velocity *types.Pointf) *types.Pointf {
	x, y := math.Float32bits(velocity.X), math.Float32bits(velocity.Y)
	if !(float64(normal.Y)*float64(normal.X) > 0) {
		velocity.X = math.Float32frombits(y)
		velocity.Y = math.Float32frombits(reflectVelocityLoad32Native57B810(x))
	} else {
		velocity.X = math.Float32frombits(reflectVelocityLoad32Native57B810(y) ^ 0x80000000)
		velocity.Y = math.Float32frombits(reflectVelocityLoad32Native57B810(x) ^ 0x80000000)
	}
	return velocity
}

func reflectVelocityLoad32Native57B810(bits uint32) uint32 {
	if bits&0x7f800000 == 0x7f800000 && bits&0x007fffff != 0 {
		bits |= 0x00400000 // FLD/FSTP quiets a signaling NaN, retaining its payload.
	}
	return bits
}
