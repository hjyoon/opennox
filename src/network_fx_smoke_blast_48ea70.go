package opennox

import (
	"encoding/binary"
	"image"

	"github.com/opennox/opennox/v1/client"
)

type smokeBlastState48EA70 struct {
	Pos image.Point
}

type smokeBlastHooks48EA70 struct {
	connected func() bool
	types     func() (smoke, puff int)
	spawn     func(typ int, pos image.Point) *client.Drawable
	random    func(min, max int) int
	activate  func(dr *client.Drawable)
}

func decodeSmokeBlastState48EA70(data []byte) (smokeBlastState48EA70, bool) {
	if len(data) < 5 {
		return smokeBlastState48EA70{}, false
	}
	return smokeBlastState48EA70{
		Pos: image.Pt(
			int(int16(binary.LittleEndian.Uint16(data[1:3]))),
			int(int16(binary.LittleEndian.Uint16(data[3:5]))),
		),
	}, true
}

func resolveSmokeBlastTypesNative48EA70(smoke, puff *int, lookup func(string) int) (int, int) {
	if *smoke == 0 {
		*smoke = lookup("Smoke")
		*puff = lookup("Puff")
	}
	return *smoke, *puff
}

func handleSmokeBlastNative48EA70(data []byte, hooks smokeBlastHooks48EA70) int {
	state, ok := decodeSmokeBlastState48EA70(data)
	if !ok {
		return -1
	}
	if !hooks.connected() {
		return 5
	}

	smokeType, puffType := hooks.types()
	if dr := hooks.spawn(smokeType, state.Pos); dr != nil {
		dr.ZVal = 20
		hooks.activate(dr)
	}
	for i := 0; i < 6; i++ {
		y := state.Pos.Y + hooks.random(-15, 15)
		x := state.Pos.X + hooks.random(-15, 15)
		if dr := hooks.spawn(puffType, image.Pt(x, y)); dr != nil {
			dr.ZVal = uint16(hooks.random(5, 25))
			hooks.activate(dr)
		}
	}
	return 5
}

func (c *Client) handleSmokeBlastPacketNative48EA70(data []byte) int {
	return handleSmokeBlastNative48EA70(data, smokeBlastHooks48EA70{
		connected: nox_client_isConnected,
		types: func() (int, int) {
			return resolveSmokeBlastTypesNative48EA70(
				&c.fxSmokeBlastSmokeType,
				&c.fxSmokeBlastPuffType,
				c.Things.IndByID,
			)
		},
		spawn:    c.Nox_xxx_spriteLoadAdd_45A360_drawable,
		random:   c.srv.Rand.Other.Int,
		activate: c.Objs.List34Add,
	})
}
