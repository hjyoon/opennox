package opennox

import (
	"fmt"
	"image"

	"github.com/opennox/libs/noximage"
	"github.com/opennox/libs/object"
	"github.com/opennox/libs/types"
	"github.com/opennox/opennox/v1/client"
	"github.com/opennox/opennox/v1/common/memmap"
	"github.com/opennox/opennox/v1/legacy"
	"github.com/opennox/opennox/v1/server"
)

// CheckMinimapClipping creates stock server objects, marks them through the
// normal minimap/network path, and checks pixels from the live C minimap pass.
// No client drawable, clip rectangle, or marker pixel is injected.
func (sc *e2eScenario) CheckMinimapClipping(name string) {
	var fixtures []*server.Object
	var codes []uint16
	var rect image.Rectangle
	var origin image.Point
	const zoom = 500
	points := []image.Point{}
	var renderPoints []image.Point
	sc.addWhen(0, name+" fixtures", 1200, func() bool {
		return noxServer.Players.HostUnit() != nil && sub_473670() != 0
	}, func() {
		legacy.SetMinimapZoom(zoom)
		pix := noxClient.r.PixBuffer()
		width := pix.Rect.Dx() / 6
		top := (pix.Rect.Dy() - width) / 2
		rect = image.Rect(0, top, width, top+width)
		center := noxClient.Viewport().World.Max
		span := width * zoom / 100
		origin = center.Sub(image.Pt(span/2, span/2))
		mid := width / 2
		points = []image.Point{
			image.Pt(mid+15, top+mid+15),
			image.Pt(0, top+mid), image.Pt(width-1, top+mid),
			image.Pt(mid, top), image.Pt(mid, top+width-1),
			image.Pt(-12, top+mid), image.Pt(width+12, top+mid),
			image.Pt(mid, top-12), image.Pt(mid, top+width+12),
		}
		player := noxServer.Players.HostUnit()
		for _, p := range points {
			item := noxServer.NewObjectByTypeID("RedApple")
			if item == nil {
				e2eError(fmt.Errorf("cannot create stock minimap fixture"))
				return
			}
			world := origin.Add(p.Sub(image.Pt(0, top)).Mul(zoom / 100))
			noxServer.CreateObjectAt(item, nil, types.Ptf(float32(world.X), float32(world.Y)))
			// Keep raster-boundary fixtures fixed even when an exact test
			// coordinate lies inside a map wall. Only server collision is
			// disabled; positions still arrive through normal object packets.
			item.ObjFlags |= object.FlagNoCollide
			noxServer.ObjectsAddPending()
			code := noxServer.GetUnitNetCode(item)
			if code <= 0 || code > 0xFFFF || noxServer.Players.Nox_xxx_netMarkMinimapObject_417190(player.ControllingPlayer().PlayerIndex(), item, 1) == 0 {
				e2eError(fmt.Errorf("cannot mark minimap fixture %p with code %#x", item, code))
				return
			}
			fixtures = append(fixtures, item)
			codes = append(codes, uint16(code))
			noxServer.NetSendInterestingIDOn(item)
			e2eLog.Printf("MINIMAP FIXTURE: object=%p wire=%#x world=%v projected=%v inside=%t", item, code, world, p, p.In(rect))
		}
	})
	checks := 0
	sc.addWhen(1, name+" pixels", 1200, func() bool {
		checks++
		if len(codes) != len(points) || len(codes) == 0 {
			return false
		}
		for _, code := range codes {
			dr := noxClient.Objs.ByNetCode(code)
			if dr == nil || !e2eMinimapContains(dr) {
				if checks == 1 || checks%300 == 0 {
					e2eLog.Printf("MINIMAP WAIT: wire=%#x drawable=%p tracked=%t checks=%d", code, dr, dr != nil && e2eMinimapContains(dr), checks)
				}
				return false
			}
		}
		renderPoints = renderPoints[:0]
		for i, code := range codes {
			dr := noxClient.Objs.ByNetCode(code)
			p := image.Pt(100*(dr.PosVec.X-origin.X)/zoom, rect.Min.Y+100*(dr.PosVec.Y-origin.Y)/zoom)
			// Normal item physics and integer position packets can shift a
			// marker by one pixel. Require the intended inside/outside side
			// and a one-pixel tolerance; radius-four edge markers still cross
			// their respective edge and exercise actual raster clipping.
			if abs(p.X-points[i].X) > 1 || abs(p.Y-points[i].Y) > 1 || p.In(rect) != points[i].In(rect) {
				if checks == 1 || checks%300 == 0 {
					e2eLog.Printf("MINIMAP POSITION WAIT: wire=%#x projected=%v want=%v checks=%d", code, p, points[i], checks)
				}
				return false
			}
			renderPoints = append(renderPoints, p)
		}
		return true
	}, func() {
		data := noxClient.r.Data()
		beforeData := *data
		livePix := noxClient.r.PixBuffer()
		// Render the live scene to an empty target, since redrawing an already
		// visible minimap can leave every marker pixel unchanged. This is the
		// real C pass and real networked drawables, not injected marker pixels.
		pix := noximage.NewImage16(livePix.Rect)
		noxClient.r.SetPixBuffer(pix)
		defer noxClient.r.SetPixBuffer(livePix)
		player := noxClient.Objs.ByNetCode(uint16(noxServer.GetUnitNetCode(noxServer.Players.HostUnit())))
		if !e2eClientDrawMinimap(player) {
			return
		}
		// The three black border strokes deliberately occupy a three-pixel
		// margin outside the interior; object markers must not escape it.
		border := rect.Inset(-3)
		inside, outside := 0, 0
		firstOutside := image.Point{}
		for y := pix.Rect.Min.Y; y < pix.Rect.Max.Y; y++ {
			for x := pix.Rect.Min.X; x < pix.Rect.Max.X; x++ {
				ind := pix.PixOffset(x, y)
				if pix.Pix[ind] == 0 {
					continue
				}
				p := image.Pt(x, y)
				if p.In(rect) {
					inside++
				} else if !p.In(border) {
					if outside == 0 {
						firstOutside = p
					}
					outside++
				}
			}
		}
		e2eLog.Printf("MINIMAP PIXELS: rect=%v zoom=%d tracked=%d inside_changes=%d outside_changes=%d first_outside=%v", rect, zoom, len(codes), inside, outside, firstOutside)
		if outside != 0 || inside == 0 {
			e2eError(fmt.Errorf("minimap pixel clipping failed: inside=%d outside=%d first=%v", inside, outside, firstOutside))
			return
		}
		if data.Clip() != beforeData.Clip() || data.ClipRect() != beforeData.ClipRect() || data.ClipRect2() != beforeData.ClipRect2() {
			e2eError(fmt.Errorf("minimap pass did not restore the previous clip state"))
			return
		}
		// In-bound markers must remain visible, including partial markers
		// overlapping each of the four edges. Off-screen fixtures
		// remain tracked: the fix must clip, not remove them from the list.
		color := uint16(memmap.Uint32(0x85B3FC, 940))
		for i, code := range codes {
			dr := noxClient.Objs.ByNetCode(code)
			p := image.Pt(100*(dr.PosVec.X-origin.X)/zoom, rect.Min.Y+100*(dr.PosVec.Y-origin.Y)/zoom)
			if p != renderPoints[i] || !e2eMinimapContains(dr) {
				e2eError(fmt.Errorf("minimap fixture moved or was untracked: code=%#x projected=%v want=%v", code, p, renderPoints[i]))
				return
			}
			if i < 5 && pix.Pix[pix.PixOffset(p.X, p.Y)] != color {
				e2eError(fmt.Errorf("in-bounds minimap marker missing: code=%#x projected=%v color=%#x want=%#x", code, p, pix.Pix[pix.PixOffset(p.X, p.Y)], color))
				return
			}
		}
		e2eLog.Printf("MINIMAP CLIPPING VERIFIED: markers=9 inside=1 edge=4 outside=4 clipped_pixels=0 restore=true")
		for _, fixture := range fixtures {
			noxServer.DelayedDelete(fixture)
		}
	})
}

func e2eMinimapContains(want *client.Drawable) bool {
	for dr := noxClient.Objs.FirstMinimapList(); dr != nil; dr = dr.Nox_xxx_cliNextMinimapObj_459EC0(dr) {
		if dr == want {
			return true
		}
	}
	return false
}
