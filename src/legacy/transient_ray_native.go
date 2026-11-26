package legacy

/*
#include "client__draw__drawrays.h"
#include "client__draw__fx.h"
#include "cgo_const.h"
typedef const nox_drawable nox_transient_const_drawable;
extern uint32_t dword_5d4594_1304328;
*/
import "C"

import (
	"encoding/binary"
	"slices"
	"unsafe"

	"github.com/opennox/libs/noxnet/netmsg"
)

var transientRays struct {
	drawables [C.NOX_CLIENT_TRANSIENT_RAY_CAPACITY]*nox_drawable
	count     int
}

//export nox_client_transient_ray_add
func nox_client_transient_ray_add(dr *nox_drawable) C.bool {
	if dr == nil || transientRays.count >= len(transientRays.drawables) {
		return false
	}
	transientRays.drawables[transientRays.count] = dr
	transientRays.count++
	*g2ServerWord(1304308) = uint32(transientRays.count)
	return true
}

func clientTransientRayCount49BDD0() int {
	return transientRays.count
}

//export nox_client_transient_ray_at
func nox_client_transient_ray_at(index C.size_t) *nox_drawable {
	if index >= C.size_t(transientRays.count) {
		return nil
	}
	return transientRays.drawables[index]
}

//export nox_client_transient_ray_contains
func nox_client_transient_ray_contains(dr *C.nox_transient_const_drawable) C.int {
	return C.int(bool2int(slices.Contains(transientRays.drawables[:transientRays.count], dr)))
}

func clientTransientRayClear49BDD0() {
	clear(transientRays.drawables[:transientRays.count])
	transientRays.count = 0
	*g2ServerWord(1304308) = 0
}

//export nox_client_transient_ray_count
func nox_client_transient_ray_count() C.size_t {
	return C.size_t(clientTransientRayCount49BDD0())
}

//export nox_client_transient_ray_clear
func nox_client_transient_ray_clear() {
	clientTransientRayClear49BDD0()
}

func setTransientRayPayload(dr *nox_drawable, endpoints *[8]byte) {
	ray := drawableRayPayload(asDrawable(dr))
	ray[0] = 0
	copy(ray[5:], endpoints[:])
}

//export nox_client_transient_ray_set_payload
func nox_client_transient_ray_set_payload(dr *nox_drawable, endpoints *C.nox_cgo_const_uchar) {
	setTransientRayPayload(dr, (*[8]byte)(unsafe.Pointer(endpoints)))
}

func initTransientRayTypes() {
	if *g2ServerWord(1304316) != 0 {
		return
	}
	for _, entry := range [...]struct {
		name string
		slot uintptr
	}{
		{"DynamicLightning", 1304316},
		{"DynamicChainLightning", 1304320},
		{"DynamicEnergyBolt", 1304324},
		{"GreenZap", 1304348},
		{"OrbRay", 1304328},
		{"PlasmaRay", 1304332},
		{"DrainManaOrb", 1304336},
		{"HealOrb", 1304340},
		{"CharmOrb", 1304344},
	} {
		value := uint32(nox_xxx_getTTByNameSpriteMB_44CFC0(internCStr(entry.name)))
		if entry.slot == 1304328 {
			C.dword_5d4594_1304328 = C.uint32_t(value)
		} else {
			*g2ServerWord(entry.slot) = value
		}
	}
}

const transientRaySource = `C:\NoxPost\src\Client\Draw\drawrays.c`

func transientRayRand(min, max, line int32) int32 {
	return nox_common_randomIntMinMax_415FF0(min, max, internCStr(transientRaySource), line)
}

func drawTransientRayOrb(typeSlot uintptr, endpoints *[4]C.uint16_t, radius C.char, line int32) {
	if transientRayRand(0, 100, line) >= 50 {
		return
	}
	dx := transientRayRand(-20, 20, line+2)
	dy := transientRayRand(-20, 20, line+3)
	sub_499490(C.int(*g2ServerWord(typeSlot)), &endpoints[0], C.int(dx), C.int(dy), radius, 0)
}

//export nox_xxx_netDrawRays_49BDD0
func nox_xxx_netDrawRays_49BDD0(data *C.uchar) *nox_drawable {
	if data == nil {
		return nil
	}
	return drawTransientRay49BDD0((*[9]byte)(unsafe.Pointer(data)))
}

func drawTransientRay49BDD0(packet *[9]byte) *nox_drawable {
	if transientRays.count >= len(transientRays.drawables) {
		return nil
	}
	initTransientRayTypes()
	var endpoints [4]C.uint16_t
	for i := range endpoints {
		endpoints[i] = C.uint16_t(binary.LittleEndian.Uint16(packet[1+2*i:]))
	}
	x1, y1 := int32(endpoints[0]), int32(endpoints[1])
	x2, y2 := int32(endpoints[2]), int32(endpoints[3])
	x, y := x1+(x2-x1)/2, y1+(y2-y1)/2
	typeID := int32(C.dword_5d4594_1304328)
	switch netmsg.Op(packet[0]) {
	case netmsg.MSG_FX_PLASMA:
		typeID = int32(*g2ServerWord(1304332))
	case netmsg.MSG_FX_LIGHTNING:
		typeID = int32(*g2ServerWord(1304316))
	case netmsg.MSG_FX_ENERGY_BOLT:
		typeID = int32(*g2ServerWord(1304324))
	case netmsg.MSG_FX_CHAIN_LIGHTNING_BOLT:
		typeID = int32(*g2ServerWord(1304320))
	case netmsg.MSG_FX_DRAIN_MANA:
		radius := C.char(transientRayRand(6, 12, 137))
		drawTransientRayOrb(1304336, &endpoints, radius, 140)
	case netmsg.MSG_FX_CHARM:
		radius := C.char(transientRayRand(6, 12, 101))
		drawTransientRayOrb(1304344, &endpoints, radius, 104)
		reversed := [4]C.uint16_t{endpoints[2], endpoints[3], endpoints[0], endpoints[1]}
		drawTransientRayOrb(1304344, &reversed, radius, 112)
	case netmsg.MSG_FX_GREATER_HEAL:
		radius := C.char(transientRayRand(6, 12, 123))
		drawTransientRayOrb(1304340, &endpoints, radius, 126)
	default:
		return nil
	}

	dr := nox_xxx_spriteLoadAdd_45A360_drawable(typeID, x, y)
	if dr == nil {
		return nil
	}
	setTransientRayPayload(dr, (*[8]byte)(packet[1:]))
	if !nox_client_transient_ray_add(dr) {
		nox_xxx_spriteDeleteStatic_45A4E0_drawable(dr)
		return nil
	}
	return dr
}
