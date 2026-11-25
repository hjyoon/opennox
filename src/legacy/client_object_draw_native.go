package legacy

/*
#include "client__draw__doordraw.h"
#include "client__draw__maidendraw.h"
#include "client__draw__parse__parse.h"
#include "client__draw__triggerdraw.h"
#include "client__draw__weapondraw.h"
*/
import "C"

import (
	"image"
	"math"
	"unsafe"

	noxflags "github.com/opennox/opennox/v1/common/flags"
)

//export nox_thing_door_draw
func nox_thing_door_draw(view *C.nox_draw_viewport_t, drawable *C.nox_drawable) C.int {
	frame := drawable.field_74_4
	if data := (*C.nox_static_random_draw_data_t)(drawable.field_76); data != nil && data.images != nil && frame < data.count {
		img := unsafe.Slice(data.images, int(data.count))[frame]
		nox_xxx_drawObject_4C4770_draw(view, drawable, unsafe.Pointer(img))
	}
	door := asDrawable(drawable).UnionDoor()
	if !noxflags.HasGame(noxflags.GameModeQuest) || door.Field_108_0 != 1 {
		return 1
	}
	if door.Field_109 == nil {
		door.Field_110 = Nox_xxx_gLoadImg("DoorLockSilverSW").C()
		door.Field_109 = Nox_xxx_gLoadImg("DoorLockSilverSE").C()
		door.Field_111 = Nox_xxx_gLoadImg("DoorLockGoldSW").C()
		door.Field_112 = Nox_xxx_gLoadImg("DoorLockGoldSE").C()
	}
	pos := C.int2{field_0: C.int(drawable.pos.x), field_4: C.int(drawable.pos.y)}
	var screen C.int2
	sub_4739E0(view, &pos, &screen)
	x, y := int32(screen.field_0)-64, int32(screen.field_4)-79
	light := GetClient().Sub469920(image.Pt(int(pos.field_0), int(pos.field_4)))
	sub_4345F0(1)
	nox_draw_setColorMultAndIntensityRGB_433CD0(C.uchar(light[0]), C.uchar(light[1]), C.uchar(light[2]))

	// Read the direction and lock type after the object draw and light setup.
	var west bool
	switch drawable.field_74_4 {
	case 0:
		x, y, west = x-15, y-20, true
	case 8:
		x, y = x+15, y-20
	case 16:
		x, y, west = x+8, y+2, true
	default:
		x, y = x-8, y+2
	}
	img := door.Field_112
	if west {
		img = door.Field_111
	}
	if door.Field_108_1 == 1 {
		img = door.Field_109
		if west {
			img = door.Field_110
		}
	}
	nox_client_drawImageAt_47D2C0((*nox_video_bag_image_t)(img), x, y)
	return 1
}

//export nox_things_door_draw_parse
func nox_things_door_draw_parse(thing *C.nox_thing, file *C.nox_memfile, attrs *C.char) C.bool {
	return parseDoorDraw(thing, file, attrs, false)
}

func parseDoorDraw(thing *C.nox_thing, file *C.nox_memfile, attrs *C.char, high bool) C.bool {
	thing.draw_func = (*[0]byte)(C.nox_thing_door_draw)
	data := loadStaticRandomDrawData(attrs, file, high)
	thing.field_5c = unsafe.Pointer(data)
	return C.bool(data != nil)
}

//export nox_thing_maiden_draw
func nox_thing_maiden_draw(view *C.uint32_t, drawable *C.nox_drawable) C.int {
	viewport := (*C.nox_draw_viewport_t)(unsafe.Pointer(view))
	if !noxflags.HasGame(noxflags.GameFlag22) {
		record := nox_npc_by_id(int32(drawable.field_32))
		if record == nil {
			return 1
		}
		colors := (*[6]int32)(unsafe.Add(record, 8))
		for i := range colors {
			nox_draw_setMaterial_4341D0(int32(i+1), colors[i])
		}
		return C.int(nox_thing_monster_draw(viewport, drawable))
	}
	for obj := nox_server_getFirstObject_4DA790(); obj != nil; obj = nox_server_getNextObject_4DA7A0(obj) {
		if uint32(obj.extent) == uint32(drawable.field_32) {
			colors := (*[6]C.nox_color3_t)(unsafe.Pointer(nox_client_maidenColors_native_4BBF90(obj)))
			for i := range colors {
				c := colors[i]
				nox_draw_setMaterial_4340A0(int32(i+1), int32(c.r), int32(c.g), int32(c.b))
			}
			break
		}
	}
	return C.int(nox_thing_monster_draw(viewport, drawable))
}

func triggerDrawCoordinate(origin int32, offset C.float) int32 {
	rounded := math.RoundToEven(float64(offset))
	if math.IsNaN(rounded) || rounded >= 1<<31 || rounded < -1<<31 {
		return origin + math.MinInt32
	}
	return origin + int32(rounded)
}

//export nox_thing_trigger_draw
func nox_thing_trigger_draw(view *C.nox_draw_viewport_t, drawable *C.nox_drawable) C.int {
	x := int32(view.x1) - int32(view.field_4) + int32(drawable.pos.x)
	y := int32(view.y1) - int32(view.field_5) + int32(drawable.pos.y)
	shape := &drawable.shape
	left := game3LinePoint{triggerDrawCoordinate(x, shape.box_left_top), triggerDrawCoordinate(y, shape.box_left_bottom)}
	bottom := game3LinePoint{triggerDrawCoordinate(x, shape.box_right_bottom_2), triggerDrawCoordinate(y, shape.box_right_top_2)}
	top := game3LinePoint{triggerDrawCoordinate(x, shape.box_left_bottom_2), triggerDrawCoordinate(y, shape.box_left_top_2)}
	right := game3LinePoint{triggerDrawCoordinate(x, shape.box_right_top), triggerDrawCoordinate(y, shape.box_right_bottom)}

	nox_client_drawSetColor_434460(int32(Get_nox_color_black_2650656()))
	nox_client_drawEnableAlpha_434560(1)
	for _, line := range [...][2]game3LinePoint{{left, right}, {bottom, right}, {left, top}, {bottom, top}} {
		nox_client_drawAddPoint_49F500(line[0].x, line[0].y)
		nox_client_drawAddPoint_49F500(line[1].x, line[1].y)
		nox_client_drawLineFromPoints_49E4B0()
	}
	nox_client_drawEnableAlpha_434560(0)
	return 1
}

//export nox_thing_weapon_draw
func nox_thing_weapon_draw(view *C.int, drawable *C.nox_drawable) C.int {
	sub_4B95D0(drawable)
	return nox_thing_static_draw((*C.uint32_t)(unsafe.Pointer(view)), drawable)
}

//export nox_thing_weapon_animate_draw
func nox_thing_weapon_animate_draw(view *C.int, drawable *C.nox_drawable) C.int {
	sub_4B95D0(drawable)
	return nox_thing_animate_draw((*C.uint32_t)(unsafe.Pointer(view)), drawable)
}
