package opennox

import (
	"encoding/binary"

	"github.com/opennox/opennox/v1/client"
	"github.com/opennox/opennox/v1/client/noxrender"
	noxflags "github.com/opennox/opennox/v1/common/flags"
)

type colorLightDrawableHooks4CE390 struct {
	gameFlag22   func() bool
	frame        func() uint32
	fps          func() uint32
	delete       func(*client.Drawable)
	byStaticCode func(int) *client.Drawable
}

func (c *Client) colorLightDrawableHooks4CE390() colorLightDrawableHooks4CE390 {
	return colorLightDrawableHooks4CE390{
		gameFlag22: func() bool { return noxflags.HasGame(noxflags.GameFlag22) },
		frame:      c.srv.Frame,
		fps:        c.srv.TickRate,
		delete:     c.Objs.List5Delete,
		byStaticCode: func(code int) *client.Drawable {
			return c.Objs.ByNetCodeStatic(code)
		},
	}
}

// updateColorLightDrawable4CE390 keeps the ColorLight update entirely on
// native-width Go pointers. The legacy callback and its helpers originated as
// PE32 routines and cannot safely receive a 64-bit Drawable address.
func updateColorLightDrawable4CE390(vp *noxrender.Viewport, dr *client.Drawable, hooks colorLightDrawableHooks4CE390) int {
	if vp == nil || dr == nil {
		return 1
	}
	counts := dr.UnionEffect().Field_108
	if byte(counts) == 0 && byte(counts>>8) == 0 && byte(counts>>16) == 0 {
		hooks.delete(dr)
		return 1
	}
	if dr.PosVec.X < vp.World.Min.X-100 || dr.PosVec.X > vp.World.Min.X+vp.Size.X+100 ||
		dr.PosVec.Y < vp.World.Min.Y-100 || dr.PosVec.Y > vp.World.Min.Y+vp.Size.Y+100 {
		return 1
	}
	if hooks.gameFlag22() {
		return 1
	}

	data := dr.ColorLightData()
	alterColorLightColor4CE440(dr, data, byte(counts), hooks.frame)
	alterColorLightIntensity4CE610(dr, data, byte(counts>>8), hooks.frame)
	alterColorLightRadius4CE760(dr, data, byte(counts>>16), hooks.frame)
	rotateColorLight4CE960(dr, data, hooks)
	trackColorLight4CE8C0(dr, data, hooks.byStaticCode)
	return 1
}

func (c *Client) updateColorLightDrawable4CE390(vp *noxrender.Viewport, dr *client.Drawable) int {
	return updateColorLightDrawable4CE390(vp, dr, c.colorLightDrawableHooks4CE390())
}

func colorLightKeyFrames4CE440(cycle float64, count byte, delay uint16, loop bool) (from, to int, fraction float64) {
	from, to, subframe := colorLightFrameSelection4CE440(cycle, count, delay, loop)
	return from, to, colorLightDivideChop53_4CE440(float64(subframe), float64(delay))
}

func alterColorLightColor4CE440(dr *client.Drawable, data *[100]byte, count byte, frame func() uint32) {
	delay := binary.LittleEndian.Uint16(data[82:84])
	if int8(count) <= 1 || delay == 0 {
		return
	}
	cycle := colorLightDivideChop53_4CE440(float64(frame()), float64(uint32(delay)*uint32(count)))
	from, to, subframe := colorLightFrameSelection4CE440(cycle, count, delay, data[0]&1 != 0)
	interpolate := func(offset int) byte {
		return byte(int64(colorLightLinear4CE440(int(data[offset+3*from]), int(data[offset+3*to]), subframe, delay)))
	}
	dr.SetLightColor(interpolate(2), interpolate(3), interpolate(4))
}

func alterColorLightIntensity4CE610(dr *client.Drawable, data *[100]byte, count byte, frame func() uint32) {
	delay := binary.LittleEndian.Uint16(data[84:86])
	if int8(count) <= 1 || delay == 0 {
		return
	}
	cycle := colorLightDivideChop53_4CE440(float64(frame()), float64(uint32(delay)*uint32(count)))
	from, to, subframe := colorLightFrameSelection4CE440(cycle, count, delay, data[0]&4 != 0)
	value := colorLightLinear4CE440(int(data[50+from]), int(data[50+to]), subframe, delay)
	dr.SetLightIntensity(fireballSpill32_52C790(value))
}

func alterColorLightRadius4CE760(dr *client.Drawable, data *[100]byte, count byte, frame func() uint32) int {
	result := int(dr.Field_42)
	delay := binary.LittleEndian.Uint16(data[86:88])
	if result != 0 || int8(count) <= 1 || delay == 0 {
		return result
	}
	cycle := colorLightDivideChop53_4CE440(float64(frame()), float64(uint32(delay)*uint32(count)))
	from, to, subframe := colorLightFrameSelection4CE440(cycle, count, delay, data[0]&0x10 != 0)
	value := colorLightLinear4CE440(int(data[66+from]), int(data[66+to]), subframe, delay)
	return int(setColorLightSize4CE760(dr, int(int64(value))))
}

func setColorLightDirection4CE8C0(dr *client.Drawable, direction int) int64 {
	result := colorLightAngle4CE960(direction)
	dr.LightDir = uint16(result)
	dr.Field_42 = 0
	return result
}

func setColorLightSize4CE760(dr *client.Drawable, size int) int64 {
	result := colorLightAngle4CE960(size)
	dr.LightPenumbra = uint16(result)
	return result
}

func trackColorLight4CE8C0(dr *client.Drawable, data *[100]byte, byStaticCode func(int) *client.Drawable) {
	if data[0]&0x40 == 0 {
		return
	}
	target := byStaticCode(int(binary.LittleEndian.Uint32(data[88:92])))
	if target == nil {
		return
	}
	dx := int32(uint32(target.PosVec.X) - uint32(dr.PosVec.X))
	dy := int32(uint32(target.PosVec.Y) - uint32(dr.PosVec.Y))
	setColorLightDirection4CE8C0(dr, int(colorLightTrackedAngle4CE8C0(dx, dy)))
}

func rotateColorLight4CE960(dr *client.Drawable, data *[100]byte, hooks colorLightDrawableHooks4CE390) {
	if dr.Field_42 != 0 {
		return
	}
	flags := binary.LittleEndian.Uint16(data[0:2])
	step := int16(binary.LittleEndian.Uint16(data[94:96]))
	if flags&0x80 == 0 || step == 0 {
		return
	}
	fps := float64(hooks.fps())
	span := 360.0
	if flags&0x100 != 0 {
		span = float64(int(binary.LittleEndian.Uint16(data[96:98])) - int(binary.LittleEndian.Uint16(data[92:94])))
	}
	stepD := float64(step)
	basePeriod := colorLightDivide4CE960(span, stepD)
	frame := hooks.frame()
	period := fireballMulChop53_52C790(basePeriod, fps)
	cycle := colorLightDivide4CE960(float64(frame), period)
	whole := float64(int32(colorLightTruncate4CE960(cycle)))
	start := float64(binary.LittleEndian.Uint16(data[92:94]))
	phase := fireballAddChop53_52C790(cycle, -whole)
	phase = fireballMulChop53_52C790(phase, period)
	speed := colorLightDivide4CE960(stepD, fps)
	value := fireballAddChop53_52C790(fireballMulChop53_52C790(phase, speed), start)
	direction := int32(colorLightTruncate4CE960(value))
	if direction < 0 {
		direction += 360
	} else if direction >= 360 {
		direction -= 360
	}
	setColorLightDirection4CE8C0(dr, int(direction))
}
