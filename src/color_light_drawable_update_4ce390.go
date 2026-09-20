package opennox

import (
	"encoding/binary"
	"math"

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
	whole := int64(cycle)
	phase := cycle - float64(int(whole))
	scaled := phase * float64(count)
	index := int(int64(scaled))
	if loop {
		from = index
		to = index + 1
		if to >= int(count) {
			to = 0
		}
	} else if whole&1 != 0 {
		from = int(count) - index - 1
		to = from - 1
		if to < 0 {
			to = 0
		}
	} else {
		from = index
		to = index + 1
		if to >= int(count) {
			to = int(count) - 1
		}
	}
	subframe := uint8(int64((scaled - float64(int(int64(scaled)))) * float64(delay)))
	return from, to, float64(subframe) / float64(delay)
}

func alterColorLightColor4CE440(dr *client.Drawable, data *[100]byte, count byte, frame func() uint32) {
	delay := binary.LittleEndian.Uint16(data[82:84])
	if count <= 1 || delay == 0 {
		return
	}
	cycle := float64(frame()) / float64(delay)
	from, to, fraction := colorLightKeyFrames4CE440(cycle/float64(count), count, delay, data[0]&1 != 0)
	interpolate := func(offset int) byte {
		start := float64(data[offset+3*from])
		end := float64(data[offset+3*to])
		return byte(int(start + (end-start)*fraction))
	}
	dr.SetLightColor(interpolate(2), interpolate(3), interpolate(4))
}

func alterColorLightIntensity4CE610(dr *client.Drawable, data *[100]byte, count byte, frame func() uint32) {
	delay := binary.LittleEndian.Uint16(data[84:86])
	if count <= 1 || delay == 0 {
		return
	}
	cycle := float64(frame()) / (float64(delay) * float64(count))
	from, to, fraction := colorLightKeyFrames4CE440(cycle, count, delay, data[0]&4 != 0)
	start := float64(data[50+from])
	end := float64(data[50+to])
	dr.SetLightIntensity(float32(start + (end-start)*fraction))
}

func alterColorLightRadius4CE760(dr *client.Drawable, data *[100]byte, count byte, frame func() uint32) int {
	result := int(dr.Field_42)
	delay := binary.LittleEndian.Uint16(data[86:88])
	if result != 0 || count <= 1 || delay == 0 {
		return result
	}
	cycle := float64(frame()) / (float64(delay) * float64(count))
	from, to, fraction := colorLightKeyFrames4CE440(cycle, count, delay, data[0]&0x10 != 0)
	start := float64(data[66+from])
	end := float64(data[66+to])
	size := int(start + (end-start)*fraction)
	return int(setColorLightSize4CE760(dr, size))
}

func setColorLightDirection4CE8C0(dr *client.Drawable, direction int) int64 {
	result := int64(float64(direction)*0.0027777778*65536.0 + 0.5)
	dr.LightDir = uint16(result)
	dr.Field_42 = 0
	return result
}

func setColorLightSize4CE760(dr *client.Drawable, size int) int64 {
	result := int64(float64(size)*0.0027777778*65536.0 + 0.5)
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
	dx := target.PosVec.X - dr.PosVec.X
	dy := target.PosVec.Y - dr.PosVec.Y
	distance := math.Sqrt(float64(dx)*float64(dx) + float64(dy)*float64(dy))
	if distance == 0 {
		return
	}
	direction := float64(float32(math.Acos(float64(dx)/distance))) * 57.295776
	if dy < 0 {
		direction = 360.0 - direction
	}
	setColorLightDirection4CE8C0(dr, int(direction))
}

func rotateColorLight4CE960(dr *client.Drawable, data *[100]byte, hooks colorLightDrawableHooks4CE390) {
	flags := binary.LittleEndian.Uint16(data[0:2])
	step := binary.LittleEndian.Uint16(data[94:96])
	if dr.Field_42 == 0 || flags&0x80 == 0 || step == 0 {
		return
	}
	start := binary.LittleEndian.Uint16(data[92:94])
	span := 360.0
	if flags&0x100 != 0 {
		span = float64(int(binary.LittleEndian.Uint16(data[96:98])) - int(start))
	}
	fps := float64(hooks.fps())
	stepD := float64(step)
	period := span / stepD * fps
	cycle := float64(hooks.frame()) / period
	whole := int64(float64(hooks.frame()) / period)
	direction := int((cycle-float64(int(whole)))*period*(stepD/fps) + float64(start))
	if direction >= 360 {
		direction -= 360
	} else if direction < 0 {
		direction += 360
	}
	setColorLightDirection4CE8C0(dr, direction)
}
