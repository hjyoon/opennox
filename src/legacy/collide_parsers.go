package legacy

/*
#include "GAME3_2.h"
#include "wall_reflect_collide_4e9d80.h"
#include "audio_event_collide_4eaad0.h"
#include "spark_explosion_collide_4e9ac0.h"
#include "damage_collide_4e9430.h"
#include "mana_drain_collide_4e9490.h"
#include "monster_arrow_collide_4eb800.h"
#include "cgo_const.h"

typedef const char damage_type_const_name_t;
*/
import "C"

import (
	"slices"
	"strings"
	"unsafe"

	"github.com/opennox/libs/object"
	"github.com/opennox/opennox/v1/common/sound"
	"github.com/opennox/opennox/v1/server"
)

const unknownLegacyDamageType = int32(object.DamageAirborneElectric) + 1

func damageTypeNamed4E0A00(name string) int32 {
	if name, ok := strings.CutPrefix(name, "DAMAGE_"); ok {
		index := slices.Index(object.DamageTypeNames[:unknownLegacyDamageType], name)
		if index >= 0 {
			return int32(index)
		}
	}
	return unknownLegacyDamageType
}

//export nox_xxx_parseDamageTypeByName_4E0A00
func nox_xxx_parseDamageTypeByName_4E0A00(name *C.damage_type_const_name_t) C.int {
	return C.int(damageTypeNamed4E0A00(GoString((*C.char)(unsafe.Pointer(name)))))
}

// nextCollideSpaceToken terminates the token in the caller buffer and
// returns the following cursor, as the legacy strtok calls did.
func nextCollideSpaceToken(args *C.char) (string, *C.char) {
	raw := GoString(args)
	token := strings.TrimLeft(raw, " ")
	end := strings.IndexByte(token, ' ')
	if end < 0 {
		return token, nil
	}
	offset := len(raw) - len(token) + end
	*(*C.char)(unsafe.Add(unsafe.Pointer(args), offset)) = 0
	return token[:end], (*C.char)(unsafe.Add(unsafe.Pointer(args), offset+1))
}

//export sub_536D80
func sub_536D80(args *C.char, data unsafe.Pointer) C.int {
	if value, ok := server.ParseCRTDecimal32(GoString(args)); ok {
		*(*int32)(data) = value
	}
	return 1
}

//export sub_536DA0
func sub_536DA0(args *C.char, data *C.nox_audio_event_collide_data_t) C.int {
	const whitespace = " \t\r\n\v\f"
	name := strings.TrimLeft(GoString(args), whitespace)
	if stop := strings.IndexAny(name, whitespace); stop >= 0 {
		name = name[:stop]
	}
	id := sound.ByName(name)
	(*server.AudioEventCollideData)(unsafe.Pointer(data)).Sound = uint32(id)
	return C.int(bool2int(id != 0))
}

//export nox_xxx_collideSparkExplosionLoad_536DE0
func nox_xxx_collideSparkExplosionLoad_536DE0(args *C.char, data *C.nox_spark_explosion_collide_data_t) C.int {
	value, ok := server.ParseCRTDecimal32(GoString(args))
	if !ok {
		// The original scan leaves the input address in its stack slot.
		value = int32(uint8(uintptr(unsafe.Pointer(args))))
	}
	(*server.SparkExplosionCollideData)(unsafe.Pointer(data)).Power = uint8(value)
	return 1
}

//export nox_xxx_collideDamageLoad_536E10
func nox_xxx_collideDamageLoad_536E10(args *C.char, data *C.nox_damage_collide_data_t) C.int {
	cfg := (*server.DamageCollideData)(unsafe.Pointer(data))
	value, next := nextCollideSpaceToken(args)
	damage, _ := server.ParseCRTDecimal32(value)
	cfg.Damage = uint8(damage)
	name, _ := nextCollideSpaceToken(next)
	cfg.DamageType = damageTypeNamed4E0A00(name)
	return C.int(bool2int(cfg.DamageType != unknownLegacyDamageType))
}

//export nox_xxx_collideManaDrainLoad_536E50
func nox_xxx_collideManaDrainLoad_536E50(args *C.char, data *C.nox_mana_drain_collide_data_t) C.int {
	value, _ := nextCollideSpaceToken(args)
	amount, _ := server.ParseCRTDecimal32(value)
	(*server.ManaDrainCollideData)(unsafe.Pointer(data)).Amount = uint8(amount)
	return 1
}

//export sub_536E80
func sub_536E80(args *C.char, data *C.nox_monster_arrow_collide_data_t) C.int {
	value, _ := nextCollideSpaceToken(args)
	damage, _ := server.ParseCRTDecimal32(value)
	cfg := (*server.MonsterArrowCollideData)(unsafe.Pointer(data))
	// GAME.EXE restarts tokenization at the original address after the first
	// token has been terminated, so both fields use its value.
	cfg.CoopDamage, cfg.OtherDamage = damage, damage
	return 1
}
