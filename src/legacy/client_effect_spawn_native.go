package legacy

/*
#include "GAME1_2.h"
#include "GAME3_1.h"
#include "client__draw__fx.h"
*/
import "C"

import (
	"encoding/binary"
	"math"
	"unsafe"

	"github.com/opennox/opennox/v1/client"
)

const drawFXSource = `C:\NoxPost\src\client\Draw\Fx.c`

func drawFXRandom(min, max, line int32) int32 {
	return nox_common_randomIntMinMax_415FF0(min, max, internCStr(drawFXSource), line)
}

func drawableEffectBytes(dr *client.Drawable) *[20]byte {
	return (*[20]byte)(unsafe.Pointer(&dr.Union))
}

func setOrbPayload(dr *client.Drawable, destination *[2]C.uint16_t, radius, fade, mode byte) {
	payload := drawableEffectBytes(dr)
	binary.LittleEndian.PutUint16(payload[:2], uint16(destination[0]))
	binary.LittleEndian.PutUint16(payload[2:4], uint16(destination[1]))
	payload[11], payload[12], payload[13], payload[14] = radius, fade, mode, mode
}

//export noxClientSetOrbPayload
func noxClientSetOrbPayload(dr *nox_drawable, destination *C.uint16_t, radius, fade, mode C.uint8_t) {
	if dr != nil && destination != nil {
		setOrbPayload(asDrawable(dr), (*[2]C.uint16_t)(unsafe.Pointer(destination)), byte(radius), byte(fade), byte(mode))
	}
}

//export sub_499490
func sub_499490(typeID C.int, route *C.uint16_t, dx, dy C.int, radius, mode C.char) {
	points := (*[4]C.uint16_t)(unsafe.Pointer(route))
	dr := nox_xxx_spriteLoadAdd_45A360_drawable(int32(typeID), int32(dx)+int32(points[2]), int32(dy)+int32(points[3]))
	if dr != nil {
		setOrbPayload(asDrawable(dr), (*[2]C.uint16_t)(unsafe.Pointer(route)), byte(radius), byte(drawFXRandom(3, 10, 138)), byte(mode))
		nox_xxx_sprite_45A110_drawable(dr)
	}
}

func setManaBombOrbPayload(dr *client.Drawable, path *[4]C.short, angle uint16, reverse, fade, mode byte) {
	dx := int32(uint16(path[2])) - int32(uint16(path[0]))
	dy := int32(uint16(path[3])) - int32(uint16(path[1]))
	var distance uint16
	// GAME.EXE loads the wrapped sum as a signed DWORD before FSQRT.
	if square := dx*dx + dy*dy; square >= 0 {
		distance = uint16(math.Sqrt(float64(square)))
	}
	payload := drawableEffectBytes(dr)
	for i, coordinate := range path {
		binary.LittleEndian.PutUint16(payload[2*i:], uint16(coordinate))
	}
	binary.LittleEndian.PutUint16(payload[8:10], distance)
	payload[10], payload[11], payload[12] = byte(angle), reverse, fade
	payload[13], payload[14] = mode, mode
	dr.ClientUpdateFuncPtr = C.sub_4CA720
	dr.Field_127 = dr.Field_127&0xffff0000 | uint32(angle)
}

//export noxClientSetManaBombOrbPayload
func noxClientSetManaBombOrbPayload(dr *nox_drawable, path *C.short, angle C.uint16_t, reverse, fade, mode C.uint8_t) {
	if dr != nil && path != nil {
		setManaBombOrbPayload(asDrawable(dr), (*[4]C.short)(unsafe.Pointer(path)), uint16(angle), byte(reverse), byte(fade), byte(mode))
	}
}

//export sub_499520
func sub_499520(typeID C.int, path *C.short, angle C.short, reverse, mode C.char) {
	if path == nil {
		return
	}
	points := (*[4]C.short)(unsafe.Pointer(path))
	dr := nox_xxx_spriteLoadAdd_45A360_drawable(int32(typeID), int32(uint16(points[2])), int32(uint16(points[3])))
	if dr != nil {
		setManaBombOrbPayload(asDrawable(dr), points, uint16(angle), byte(reverse), byte(drawFXRandom(3, 10, 182)), byte(mode))
		nox_xxx_sprite_45A110_drawable(dr)
	}
}

func setBallisticFX(dr *client.Drawable, speed, start, end uint32, angle byte) {
	state := dr.UnionEffect()
	pos := dr.Pos()
	state.Field_108, state.Field_109 = uint32(pos.X)<<12, uint32(pos.Y)<<12
	dr.Field_74_4 = angle
	state.Field_110, state.Field_111, state.Field_112 = speed, start, end
}

//export nox_client_ballistic_fx_set_state_499610
func nox_client_ballistic_fx_set_state_499610(dr *nox_drawable, speed, start, end C.uint32_t, angle C.uint8_t) {
	if dr != nil {
		setBallisticFX(asDrawable(dr), uint32(speed), uint32(start), uint32(end), byte(angle))
	}
}

//export nox_xxx_makePointFxCli_499610
func nox_xxx_makePointFxCli_499610(typeID, count, maxSpeed, minLifetime, x, y C.int) C.int {
	for ; count > 0; count-- {
		dr := nox_xxx_spriteLoadAdd_45A360_drawable(int32(typeID), int32(x), int32(y))
		if dr == nil {
			continue
		}
		frame := gameFrame()
		angle := byte(drawFXRandom(0, 255, 227))
		speed := uint32(drawFXRandom(1, int32(maxSpeed), 230))
		end := frame + uint32(drawFXRandom(int32(minLifetime), 64, 233))
		d := asDrawable(dr)
		setBallisticFX(d, speed, frame, end, angle)
		d.ZVal = 0
		d.VelZ = int8(drawFXRandom(2, 10, 239))
		nox_xxx_sprite_45A110_drawable(dr)
	}
	return count
}

func setElectricSparkTrajectory(dr *client.Drawable, x, y int32, angleLine int32) {
	state := dr.UnionEffect()
	state.Field_108, state.Field_109 = uint32(x)<<12, uint32(y)<<12
	dr.Field_74_4 = byte(drawFXRandom(0, 255, angleLine))
	state.Field_110 = uint32(drawFXRandom(1, 3000, angleLine+3))
	state.Field_112 = gameFrame() + uint32(drawFXRandom(5, 20, angleLine+6))
	state.Field_111 = gameFrame()
}

//export nox_xxx_drawEnergyBolt_499710
func nox_xxx_drawEnergyBolt_499710(x, y C.int, z C.short, typeID C.int) C.int {
	for range 2 {
		dr := nox_xxx_spriteLoadAdd_45A360_drawable(int32(typeID), int32(x), int32(y))
		if dr == nil {
			continue
		}
		d := asDrawable(dr)
		setElectricSparkTrajectory(d, int32(x), int32(y), 268)
		d.ZVal = uint16(int32(z) + drawFXRandom(0, 20, 279))
		d.VelZ = int8(drawFXRandom(0, 4, 280))
		nox_xxx_sprite_45A110_drawable(dr)
	}
	return 0
}

func setFallingSparkPayload(dr *client.Drawable, origin *C.int2, z uint16, velocity int8, fade byte) {
	dr.ZVal, dr.ZVal2, dr.VelZ = z, 0, velocity
	payload := drawableEffectBytes(dr)
	binary.LittleEndian.PutUint32(payload[:4], uint32(origin.field_0))
	binary.LittleEndian.PutUint32(payload[4:8], uint32(origin.field_4))
	binary.LittleEndian.PutUint16(payload[8:10], z)
	payload[10] = fade
}

//export noxClientSetFallingSparkPayload
func noxClientSetFallingSparkPayload(dr *nox_drawable, origin *C.int2, z C.uint16_t, velocity C.int8_t, fade C.uint8_t) {
	if dr != nil && origin != nil {
		setFallingSparkPayload(asDrawable(dr), origin, uint16(z), int8(velocity), byte(fade))
	}
}

//export sub_499950
func sub_499950(typeID C.int, spawn, origin *C.int2, z C.ushort, velocity C.char) *nox_drawable {
	if spawn == nil || origin == nil {
		return nil
	}
	dr := nox_xxx_spriteLoadAdd_45A360_drawable(int32(typeID), int32(spawn.field_0), int32(spawn.field_4))
	if dr != nil {
		setFallingSparkPayload(asDrawable(dr), origin, uint16(z), int8(velocity), byte(drawFXRandom(3, 10, 406)))
		nox_xxx_sprite_45A110_drawable(dr)
	}
	return dr
}

//export nox_xxx_makeLightningParticles_4999D0
func nox_xxx_makeLightningParticles_4999D0(typeID C.int, from, to *C.int2) C.int {
	dx, dy := int32(to.field_0-from.field_0), int32(to.field_4-from.field_4)
	squared := dx*dx + dy*dy
	if squared <= 0 {
		return 0
	}
	length := int32(math.Sqrt(float64(squared)))
	at := drawFXRandom(0, length, 437)
	step := at
	for at <= length {
		x := int32(from.field_0) + dx*at/length
		y := int32(from.field_4) + dy*at/length
		dr := nox_xxx_spriteLoadAdd_45A360_drawable(int32(typeID), x, y)
		if dr != nil {
			d := asDrawable(dr)
			setElectricSparkTrajectory(d, x, y, 458)
			d.ZVal = uint16(drawFXRandom(15, 30, 471))
			d.VelZ = int8(drawFXRandom(-4, 4, 472))
			nox_xxx_sprite_45A110_drawable(dr)
		}
		step = drawFXRandom(8, 100, 439)
		at += step
	}
	return C.int(step)
}

//export nox_xxx_draw_499E70
func nox_xxx_draw_499E70(color, x, y, width, height, edge, motion C.int) C.int {
	for range 100 {
		px, py := x, y
		if edge == 2 {
			px += C.int(drawFXRandom(0, int32(width), 555))
		} else {
			py += C.int(drawFXRandom(0, int32(height), 563))
		}
		size := C.char(drawFXRandom(2, 5, 568))
		vy := C.int(drawFXRandom(-40, -20, 571))
		var vx C.int
		if motion == 1 {
			vx = C.int(drawFXRandom(-20, 0, 573))
		} else {
			vx = C.int(drawFXRandom(0, 20, 575))
		}
		nox_client_newScreenParticle_431540(color, px, py, vx, vy, 1, size, 0, 0, 1)
	}
	return 0
}

//export sub_49A150
func sub_49A150(pos *C.int2, typeID C.int, intensity C.uchar) C.int {
	value := int32(intensity)
	maxSpeed := 2400*value/255 + 200
	minLifetime := 10*value/255 + 5
	for range 180*value/255 + 10 {
		dr := nox_xxx_spriteLoadAdd_45A360_drawable(int32(typeID), int32(pos.field_0), int32(pos.field_4))
		if dr == nil {
			continue
		}
		frame := gameFrame()
		angle := byte(drawFXRandom(0, 255, 822))
		speed := uint32(drawFXRandom(1, maxSpeed, 825))
		end := frame + uint32(drawFXRandom(minLifetime, 96, 828))
		d := asDrawable(dr)
		setBallisticFX(d, speed, frame, end, angle)
		d.ZVal = uint16(drawFXRandom(5, 15, 834))
		d.ZVal2 = 0
		d.VelZ = int8(drawFXRandom(0, 8, 836))
		nox_xxx_sprite_45A110_drawable(dr)
	}
	return 0
}
