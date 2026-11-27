package legacy

/*
#include "client__gui__guibook.h"

*/
import "C"

import (
	"unsafe"

	"github.com/opennox/opennox/v1/common/memmap"
)

func bookRewardRandom(min, max, line int32) C.int {
	return C.int(nox_common_randomIntMinMax_415FF0(min, max, internCStr(bookStringsFile), line))
}

func completeBookReward() {
	kind := C.uintptr_t(memmap.Uint32(0x5D4594, 1046676))
	nox_xxx_book_45DBE0(kind, C.int(bookState.reward_id), C.int(bookState.reward_slot))
	sub_45D810()
}

func burstBookReward(color, x, y C.int) {
	for range 50 {
		interval := bookRewardRandom(3, 6, 1287)
		radius := bookRewardRandom(2, 5, 1286)
		vy := bookRewardRandom(-10, -1, 1284)
		vx := bookRewardRandom(-10, 10, 1283)
		dy := bookRewardRandom(0, 30, 1282)
		dx := bookRewardRandom(0, 30, 1281)
		nox_client_newScreenParticle_431540(color, x+dx, y+dy, vx, vy, 1, C.char(radius), C.char(interval), 2, 1)
	}
	nox_xxx_clientPlaySoundSpecial_452D80(795, 100)
	bookState.reward_start_frame = 0
}

func advanceBookReward(x, y *C.float) {
	velocity := &bookState.reward_velocity
	*x += velocity.field_0
	*y += velocity.field_4
	if float64(*x) >= float64(memmap.Int32(0x5D4594, 1046668)) &&
		float64(*y) >= float64(memmap.Int32(0x5D4594, 1046672)) {
		completeBookReward()
		return
	}
	segment := memmap.PtrUint32(0x5D4594, 1046628)
	offset := 8 * uintptr(*segment)
	if !(*x > C.float(memmap.Float32(0x5D4594, 1046692+offset))) {
		return
	}
	*segment++
	if int32(*segment) >= memmap.Int32(0x5D4594, 1046680) {
		completeBookReward()
		return
	}
	offset = 8 * uintptr(*segment)
	velocity.field_0 = C.float(memmap.Float32(0x5D4594, 1046692+offset) - memmap.Float32(0x5D4594, 1046684+offset))
	velocity.field_4 = C.float(memmap.Float32(0x5D4594, 1046696+offset) - memmap.Float32(0x5D4594, 1046688+offset))
	nox_xxx_utilNormalizeVector_509F20(velocity)
	speed := C.float(10)
	if clientDisplayStorage.width < 1000 {
		speed = 8
		if clientDisplayStorage.width < 750 {
			speed = 6
		}
	}
	velocity.field_0 *= speed
	velocity.field_4 *= speed
}

//export nox_xxx_bookDrawFn_45C7D0
func nox_xxx_bookDrawFn_45C7D0(data *C.uint32_t) C.int {
	color := C.int(0)
	if bookState.reward_ability != 0 {
		color = 3
	}
	if bookState.reward_active == 0 {
		return 1
	}
	if mark := uint32(bookState.reward_start_frame); mark != 0 &&
		uint32(nox_xxx_bookGet_430B40_get_mouse_prev_seq())-mark < 2*gameFPS() {
		return 1
	}
	win := (*nox_window)(unsafe.Pointer(data))
	var x, y C.uint
	nox_client_wndGetPosition_46AA60(win, &x, &y)
	if bookState.reward_start_frame != 0 {
		burstBookReward(color, C.int(x), C.int(y))
	}
	for range 2 {
		interval := bookRewardRandom(2, 4, 1331)
		radius := bookRewardRandom(1, 2, 1330)
		dy := bookRewardRandom(0, 30, 1326)
		dx := bookRewardRandom(0, 30, 1325)
		nox_client_newScreenParticle_431540(color, C.int(x)+dx, C.int(y)+dy, 0, 0, 0, C.char(radius), C.char(interval), 1, 1)
	}
	spell := int32(bookState.reward_id)
	var icon *nox_video_bag_image_t
	if bookState.reward_ability == 1 {
		icon = nox_xxx_spellGetAbilityIcon_425310(spell, 0)
	} else {
		icon = (*nox_video_bag_image_t)(nox_xxx_spellIcon_424A90(spell))
	}
	nox_client_drawImageAt_47D2C0(icon, int32(x), int32(y))
	px := (*C.float)(unsafe.Pointer(&bookState.reward_x_bits))
	py := (*C.float)(unsafe.Pointer(&bookState.reward_y_bits))
	advanceBookReward(px, py)
	posY := int32(nox_float2int(*py))
	posX := int32(nox_float2int(*px))
	nox_window_setPos_46A9B0(win, posX, posY)
	return 1
}

//export sub_45D870
func sub_45D870() {
	if bookState.reward_active == 0 {
		return
	}
	color := C.int(0)
	if bookState.reward_ability == 1 {
		color = 3
	}
	px := (*C.float)(unsafe.Pointer(&bookState.reward_x_bits))
	py := (*C.float)(unsafe.Pointer(&bookState.reward_y_bits))
	dx := (C.int(memmap.Uint32(0x5D4594, 1046668)) - nox_float2int(*px)) / 50
	dy := (C.int(memmap.Uint32(0x5D4594, 1046672)) - nox_float2int(*py)) / 50
	x, y := nox_float2int(*px), nox_float2int(*py)
	for range 50 {
		radius := bookRewardRandom(3, 4, 2483)
		py := y - bookRewardRandom(0, 30, 2479)
		px := x + bookRewardRandom(0, 30, 2478)
		nox_client_newScreenParticle_431540(color, px, py, 0, 0, 1, C.char(radius), 0, 0, 1)
		x += dx
		y += dy
	}
	completeBookReward()
}
