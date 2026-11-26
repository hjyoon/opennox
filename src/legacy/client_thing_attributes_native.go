package legacy

/*
#include "GAME1.h"
#include "GAME1_1.h"
#include "GAME1_2.h"
#include "GAME2.h"
#include "GAME2_1.h"
#include "GAME2_2.h"
#include "GAME3_1.h"
#include "GAME3_2.h"
#include "GAME3_3.h"
#include "GAME4.h"
#include "GAME4_3.h"
#include "client__drawable__update__dball.h"
#include "client__drawable__update__drainup.h"
#include "client__drawable__update__healup.h"
#include "client__drawable__update__manabomb.h"
#include "client__drawable__update__mmislup.h"
#include "client__drawable__update__mtailup.h"
#include "client__drawable__update__sparklup.h"
#include "client__drawable__update__telwake.h"
#include "client__drawable__update__vortexup.h"

*/
import "C"

import (
	"math"
	"strconv"
	"strings"
	"unsafe"

	"github.com/opennox/opennox/v1/common/memmap"
)

func legacyThingAngle(value string, limit int64) (C.uint16_t, bool) {
	value = strings.TrimLeft(value, " \t\r\n\v\f")
	if value == "" {
		return 0, false
	}
	pos := 0
	if value[0] == '+' || value[0] == '-' {
		pos++
	}
	for pos < len(value) && value[pos] >= '0' && value[pos] <= '9' {
		pos++
	}
	degrees, err := strconv.ParseInt(value[:pos], 10, 64)
	if err != nil || degrees < 0 || degrees >= limit {
		return 0, false
	}
	turn := memmap.Float64(0x581450, 9560)
	scale := math.Float64frombits(uint64(legacyWords.qword_581450_9552))
	bias := math.Float64frombits(uint64(legacyWords.qword_581450_9544))
	return C.uint16_t(int64(float64(degrees)*turn*scale + bias)), true
}

//export nox_parse_thing_light_dir
func nox_parse_thing_light_dir(thing *C.nox_thing, _ *C.nox_memfile, attr *C.char) C.bool {
	angle, ok := legacyThingAngle(GoString(attr), 360)
	if !ok {
		return false
	}
	thing.light_dir = angle
	thing.field_10 = 0
	return true
}

//export nox_parse_thing_light_penumbra
func nox_parse_thing_light_penumbra(thing *C.nox_thing, _ *C.nox_memfile, attr *C.char) C.bool {
	angle, ok := legacyThingAngle(GoString(attr), 180)
	if !ok {
		return false
	}
	thing.light_penumbra = angle
	return true
}

var legacyThingClientUpdates = map[string]unsafe.Pointer{
	"ColorLightUpdateDraw":        unsafe.Pointer(C.nox_xxx_updDrawColorlight_4CE390),
	"CloudUpdateDraw":             unsafe.Pointer(C.nox_xxx_updDrawCloud_4CE1D0),
	"SmallCloudUpdateDraw":        unsafe.Pointer(C.sub_4CE360),
	"DeathBallUpdateDraw":         unsafe.Pointer(C.nox_xxx_updDrawDBall_4CDF80),
	"DeathBallFragmentUpdateDraw": unsafe.Pointer(C.sub_4CE0A0),
	"DeathBallChargeUpdateDraw":   unsafe.Pointer(C.nox_xxx_updDrawDBallCharge_4CE0C0),
	"MagicUpdateDraw":             unsafe.Pointer(C.nox_xxx_updDrawMagic_4CDD80),
	"SparkleTrailUpdateDraw":      unsafe.Pointer(C.nox_xxx_updDrawSparkleTrail_4CDBF0),
	"MagicMissileUpdateDraw":      unsafe.Pointer(C.nox_xxx_updDrawMagicMissile_4CD9E0),
	"TeleportWakeUpdateDraw":      unsafe.Pointer(C.nox_xxx_updDrawTeleportWake_4CD8D0),
	"DrainManaUpdateDraw":         unsafe.Pointer(C.sub_4CD690),
	"HealUpdateDraw":              unsafe.Pointer(C.sub_4CD450),
	"CharmUpdateDraw":             unsafe.Pointer(C.sub_4CD400),
	"TitanFireballUpdateDraw":     unsafe.Pointer(C.sub_4CCE70),
	"StrongFireballUpdateDraw":    unsafe.Pointer(C.sub_4CD090),
	"FireballUpdateDraw":          unsafe.Pointer(C.sub_4CD0C0),
	"WeakFireballUpdateDraw":      unsafe.Pointer(C.sub_4CD0F0),
	"PitifulFireballUpdateDraw":   unsafe.Pointer(C.sub_4CD120),
	"FistUpdateDraw":              unsafe.Pointer(C.nox_xxx_updDrawFist_4CCDB0),
	"MeteorUpdateDraw":            unsafe.Pointer(C.sub_4CCD00),
	"UndeadKillerClientUpdate":    unsafe.Pointer(C.nox_xxx_updDrawUndeadKiller_4CCCF0),
	"ManaBombChargeClientUpdate":  unsafe.Pointer(C.nox_xxx_updDrawManabombCharge_4CCAC0),
	"VortexSourceClientUpdate":    unsafe.Pointer(C.nox_xxx_updDrawVortexSource_4CC950),
	"LinearOrbUpdateDraw":         unsafe.Pointer(C.sub_4CA650),
	"MonsterGeneratorUpdateDraw":  unsafe.Pointer(C.nox_xxx_updDrawMonsterGen_4BC920),
}

//export nox_parse_thing_client_update
func nox_parse_thing_client_update(thing *C.nox_thing, _ *C.nox_memfile, attr *C.char) C.bool {
	const delimiters = " \t\n\r"
	value := GoString(attr)
	name := strings.TrimLeft(value, delimiters)
	if end := strings.IndexAny(name, delimiters); end >= 0 {
		*(*C.char)(unsafe.Add(unsafe.Pointer(attr), len(value)-len(name)+end)) = 0
		name = name[:end]
	}
	update := legacyThingClientUpdates[name]
	if update == nil {
		return false
	}
	thing.client_update = update
	return true
}

//export nox_parse_thing_pretty_image
func nox_parse_thing_pretty_image(thing *C.nox_thing, file *C.nox_memfile, _ *C.char) C.bool {
	id := C.int(nox_memfile_read_i32(file))
	var kind C.char
	var name [256]C.char
	if id == -1 {
		kind = C.char(nox_memfile_read_u8(file))
		length := nox_memfile_read_u8(file)
		if n := nox_memfile_read(unsafe.Pointer(&name[0]), 1, C.int(length), file); n != C.uint(length) {
			return false
		}
	}
	ref := nox_xxx_readImgMB_42FAA0(id, kind, &name[0])
	thing.pretty_image = C.uint32_t(uintptr(unsafe.Pointer(ref)))
	return true
}
