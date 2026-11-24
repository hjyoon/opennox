package legacy

/*
#include "GAME1_2.h"
#include "GAME2_3.h"
#include "GAME3_1.h"
#include "GAME3_3.h"
#include "client__video__draw_common.h"
extern uint32_t nox_color_black_2650656;
*/
import "C"

import (
	"encoding/binary"
	"unsafe"

	"github.com/opennox/opennox/v1/client"
	"github.com/opennox/opennox/v1/common/memmap"
)

//export nox_xxx_clientReportSecondaryWeapon_4BF010
func nox_xxx_clientReportSecondaryWeapon_4BF010(drawable *C.nox_drawable) C.int {
	packet := [3]byte{0xe0}
	code := objectFieldDrawableCode578B00((*client.Drawable)(unsafe.Pointer(drawable)))
	binary.LittleEndian.PutUint16(packet[1:], uint16(code))
	return nox_xxx_netClientSend2_4E53C0(31, C.nox_important_payload_t(unsafe.Pointer(&packet[0])), 3, nil, 1)
}

//export nox_client_getPlayerDollState_4BF7E0
func nox_client_getPlayerDollState_4BF7E0(state *C.nox_player_doll_state_t) C.int {
	if state == nil {
		return 0
	}
	*state = C.nox_player_doll_state_t{}
	player := Get_dword_8531A0_2576()
	if player == nil {
		return 0
	}
	state.armor_mask, state.weapon_mask = C.uint32_t(player.ArmorEquip), C.uint32_t(player.WeaponEquip)
	state.color_skin = C.uint32_t(player.Colors.Skin)
	state.color_hair = C.uint32_t(player.Colors.Hair)
	state.color_mustache = C.uint32_t(player.Colors.Mustache)
	state.color_goatee = C.uint32_t(player.Colors.Goatee)
	state.color_beard = C.uint32_t(player.Colors.Beard)
	state.color_unknown = C.uint32_t(player.Colors.UnkColor)
	state.variant = C.uint8_t(player.Info().FemaleFlag())
	return 1
}

//export nox_client_getPlayerDollImage_4BF9F0
func nox_client_getPlayerDollImage_4BF9F0(offset C.uintptr_t, layer C.int) *C.nox_video_bag_image_t {
	return (*C.nox_video_bag_image_t)(*memmap.PtrPtr(0x973A20, uintptr(offset)+4*uintptr(layer)))
}

//export sub_4BF7E0
func sub_4BF7E0(origin *C.uint32_t) C.short {
	p := (*[2]C.uint32_t)(unsafe.Pointer(origin))
	x, y := C.int(p[0]+11), C.int(p[1]+15)
	C.nox_client_drawSetColor_434460(C.int(C.nox_color_black_2650656))
	C.nox_client_drawRectFilledOpaque_49CE30(x, y, 200, 200)
	if *memmap.PtrPtr(0x852978, 8) == nil {
		return 0
	}
	var doll C.nox_player_doll_state_t
	if nox_client_getPlayerDollState_4BF7E0(&doll) == 0 {
		return 0
	}
	nox_draw_setMaterial_4341D0(1, int32(doll.color_hair))
	nox_draw_setMaterial_4341D0(2, int32(doll.color_goatee))
	nox_draw_setMaterial_4341D0(3, int32(doll.color_unknown))
	nox_draw_setMaterial_4341D0(4, int32(doll.color_beard))
	nox_draw_setMaterial_4341D0(5, int32(doll.color_mustache))
	nox_draw_setMaterial_4341D0(6, int32(doll.color_skin))
	base := C.uintptr_t(16)
	if doll.color_skin == doll.color_hair {
		base = 24
	}
	nox_client_drawImageAt_47D2C0(nox_client_getPlayerDollImage_4BF9F0(base, C.int(doll.variant)), int32(x), int32(y))

	armor := C.uintptr_t(32 + 104*uintptr(doll.variant))
	for layer := 0; layer < 24; layer++ {
		bit := uint32(1) << layer
		if uint32(doll.armor_mask)&bit != 0 {
			sub_4BF9F0(C.int(bit), C.int(sub_415CD0(int32(bit))), x, y, armor, C.int(layer), 0)
		}
	}
	if doll.armor_mask&2 != 0 {
		sub_4BF9F0(2, C.int(sub_415CD0(2)), x, y, armor, 0, 1)
	}
	for layer := 24; layer < 26; layer++ {
		bit := uint32(1) << layer
		if uint32(doll.armor_mask)&bit != 0 {
			sub_4BF9F0(C.int(bit), C.int(sub_415CD0(int32(bit))), x, y, armor, C.int(layer), 0)
		}
	}
	weapons := C.uintptr_t(256 + 108*uintptr(doll.variant))
	result := C.short(doll.weapon_mask)
	for layer := 0; layer < 27; layer++ {
		result = C.short(doll.weapon_mask)
		bit := uint32(1) << layer
		if uint32(doll.weapon_mask)&bit != 0 {
			result = sub_4BF9F0(C.int(bit), C.int(sub_415840(int32(bit))), x, y, weapons, C.int(layer), 0)
		}
	}
	return result
}

//export sub_4BF9F0
func sub_4BF9F0(mask, thingType, x, y C.int, offset C.uintptr_t, layer, cloak C.int) C.short {
	drawable := sub_461600(thingType)
	if drawable == nil {
		return 0
	}
	var def unsafe.Pointer
	if drawable.flags28&0x2000000 != 0 {
		def = nox_xxx_equipClothFindDefByTT_413270(int32(drawable.field_27))
	} else {
		def = nox_xxx_getProjectileClassById_413250(int32(drawable.field_27))
	}
	applyEquipmentMaterials(def, drawable)
	var img *C.nox_video_bag_image_t
	if cloak != 0 {
		img = (*C.nox_video_bag_image_t)(*memmap.PtrPtr(0x5D4594, 1319052))
	} else {
		img = nox_client_getPlayerDollImage_4BF9F0(offset, layer)
	}
	nox_client_drawImageAt_47D2C0(img, int32(x), int32(y))
	return C.short(uintptr(unsafe.Pointer(drawable)))
}

func loadDollTexture(source, destination uintptr) {
	name := (*C.char)(*memmap.PtrPtr(0x587000, source))
	*memmap.PtrPtr(0x973A20, destination) = unsafe.Pointer(nox_xxx_gLoadImg_42F970(name))
}

//export sub_4BFAD0
func sub_4BFAD0() C.int {
	for variant := uintptr(0); variant < 2; variant++ {
		loadDollTexture(180960+4*variant, 16+4*variant)
		loadDollTexture(180968+4*variant, 24+4*variant)
		for layer := uintptr(0); layer < 26; layer++ {
			index := 26*variant + layer
			loadDollTexture(180976+4*index, 32+4*index)
		}
		for layer := uintptr(0); layer < 27; layer++ {
			index := 27*variant + layer
			loadDollTexture(181184+4*index, 256+4*index)
		}
	}
	*memmap.PtrPtr(0x5D4594, 1319052) = unsafe.Pointer(Nox_xxx_gLoadImg("MaleMedievalCloakTop").C())
	return 1
}
