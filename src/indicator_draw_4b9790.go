//go:build !server

package opennox

import (
	"image"
	"math"
	"unsafe"

	noxcolor "github.com/opennox/libs/color"

	"github.com/opennox/opennox/v1/client"
	"github.com/opennox/opennox/v1/client/noxrender"
	"github.com/opennox/opennox/v1/common/memmap"
	"github.com/opennox/opennox/v1/legacy"
)

type indicatorLine4B9790 struct {
	from image.Point
	to   image.Point
}

type indicatorDrawKind4B9790 uint8

const (
	indicatorDrawUnknown4B9790 indicatorDrawKind4B9790 = iota
	indicatorDrawPlayerWaypoint4B9790
	indicatorDrawPressurePlate4BBB30
)

type playerWaypointDrawState4B9790 struct {
	center image.Point
	circle [16]indicatorLine4B9790
	lines  [5]indicatorLine4B9790
}

func playerWaypointDrawStateFor4B9790(dr *client.Drawable, vp *noxrender.Viewport, frame uint32) playerWaypointDrawState4B9790 {
	// The original waypoint drawer uses the viewport's world origin directly
	// and, unlike most sprite drawers, does not add Screen.Min.
	state := playerWaypointDrawState4B9790{center: dr.PosVec.Sub(vp.World.Min)}
	for i := range state.circle {
		angle := uint8(i * 16)
		next := uint8(((i + 1) % len(state.circle)) * 16)
		dir1 := sincosTable16[angle]
		dir2 := sincosTable16[next]
		state.circle[i] = indicatorLine4B9790{
			from: state.center.Add(image.Pt(10*dir1.X/16, 10*dir1.Y/16)),
			to:   state.center.Add(image.Pt(10*dir2.X/16, 10*dir2.Y/16)),
		}
	}
	angle := uint8(2 * uint8(frame))
	for i := range state.lines {
		next := angle + 102
		dir1 := sincosTable16[angle]
		dir2 := sincosTable16[next]
		state.lines[i] = indicatorLine4B9790{
			from: state.center.Add(image.Pt(10*dir1.X/16, 10*dir1.Y/16)),
			to:   state.center.Add(image.Pt(10*dir2.X/16, 10*dir2.Y/16)),
		}
		angle = next
	}
	return state
}

// PlayerWaypointDraw declares its viewport argument as an int in the PE32
// source, so invoking it with a native-width pointer loses the upper address
// bits before either the viewport or drawable can be read.
func (c *Client) drawPlayerWaypoint4B9790(dr *client.Drawable, vp *noxrender.Viewport) int {
	state := playerWaypointDrawStateFor4B9790(dr, vp, c.srv.Frame())
	cl := noxcolor.RGBA5551(memmap.Uint32(0x85B3FC, 940))
	draw := c.r.Data()
	draw.SetAlphaEnabled(true)
	draw.SetColor2(cl)
	for _, line := range state.circle {
		c.r.DrawLine(line.from, line.to, cl)
	}
	for _, line := range state.lines {
		c.r.DrawLine(line.from, line.to, cl)
	}
	draw.SetAlphaEnabled(false)
	return 1
}

type pressurePlateDrawState4BBB30 struct {
	visible bool
	colors  [2]noxcolor.RGBA5551
	lines   [2][4]indicatorLine4B9790
}

func pressurePlateFloatToInt4BBB30(value float32) int {
	if math.IsNaN(float64(value)) || value >= 2147483648 || value < -2147483648 {
		return math.MinInt32
	}
	return int(math.RoundToEven(float64(value)))
}

func pressurePlateDrawStateFor4BBB30(dr *client.Drawable, vp *noxrender.Viewport) pressurePlateDrawState4BBB30 {
	effect := dr.UnionEffect()
	rgb := [6]byte{
		byte(effect.Field_108), byte(effect.Field_108 >> 8), byte(effect.Field_108 >> 16),
		byte(effect.Field_108 >> 24), byte(effect.Field_109), byte(effect.Field_109 >> 8),
	}
	state := pressurePlateDrawState4BBB30{
		visible: rgb != [6]byte{},
		colors: [2]noxcolor.RGBA5551{
			noxcolor.RGB5551Color(rgb[0], rgb[1], rgb[2]),
			noxcolor.RGB5551Color(rgb[3], rgb[4], rgb[5]),
		},
	}
	if !state.visible {
		return state
	}

	base := vp.ToScreenPos(dr.PosVec)
	box := &dr.Shape.Box
	a := base.Add(image.Pt(pressurePlateFloatToInt4BBB30(box.LeftTop), pressurePlateFloatToInt4BBB30(box.LeftBottom)))
	b := base.Add(image.Pt(pressurePlateFloatToInt4BBB30(box.RightBottom2), pressurePlateFloatToInt4BBB30(box.RightTop2)))
	c := base.Add(image.Pt(pressurePlateFloatToInt4BBB30(box.LeftBottom2), pressurePlateFloatToInt4BBB30(box.LeftTop2)))
	d := base.Add(image.Pt(pressurePlateFloatToInt4BBB30(box.RightTop), pressurePlateFloatToInt4BBB30(box.RightBottom)))
	for layer := range state.lines {
		off := image.Pt(layer, 0)
		state.lines[layer] = [4]indicatorLine4B9790{
			{from: a.Add(off), to: d.Add(off)},
			{from: b.Add(off), to: d.Add(off)},
			{from: a.Add(off), to: c.Add(off)},
			{from: b.Add(off), to: c.Add(off)},
		}
	}
	return state
}

// PressurePlateDraw copies the drawable pointer to a 32-bit int before
// reading its shape and packed RGB values. Decode those fields through the
// native Drawable layout instead.
func (c *Client) drawPressurePlate4BBB30(dr *client.Drawable, vp *noxrender.Viewport) int {
	state := pressurePlateDrawStateFor4BBB30(dr, vp)
	if !state.visible {
		return 1
	}
	for layer, lines := range state.lines {
		cl := state.colors[layer]
		c.r.Data().SetColor2(cl)
		for _, line := range lines {
			c.r.DrawLine(line.from, line.to, cl)
		}
	}
	return 1
}

func indicatorDrawKindFor4B9790(fn unsafe.Pointer) indicatorDrawKind4B9790 {
	switch fn {
	case legacy.Get_nox_thing_player_waypoint_draw():
		return indicatorDrawPlayerWaypoint4B9790
	case legacy.Get_nox_thing_pressure_plate_draw():
		return indicatorDrawPressurePlate4BBB30
	default:
		return indicatorDrawUnknown4B9790
	}
}

func (c *Client) callIndicatorDraw4B9790(dr *client.Drawable, vp *noxrender.Viewport) (int, bool) {
	switch indicatorDrawKindFor4B9790(dr.DrawFuncPtr) {
	case indicatorDrawPlayerWaypoint4B9790:
		return c.drawPlayerWaypoint4B9790(dr, vp), true
	case indicatorDrawPressurePlate4BBB30:
		return c.drawPressurePlate4BBB30(dr, vp), true
	default:
		return 0, false
	}
}
