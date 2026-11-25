package legacy

/*
#include "GAME4_3.h"
*/
import "C"

import (
	"strings"
	"unsafe"

	"github.com/opennox/opennox/v1/common/memmap"
	"github.com/opennox/opennox/v1/common/sound"
	"github.com/opennox/opennox/v1/internal/legacycrt"
	"github.com/opennox/opennox/v1/legacy/common/alloc"
	"github.com/opennox/opennox/v1/server"
)

func parseWandCharges(args *C.char, data *server.WandUseData) (next *C.char, ok bool) {
	token, next := nextCollideSpaceToken(args)
	if token == "" {
		return nil, false
	}
	count, _ := server.ParseCRTDecimal32(token)
	data.Charge, data.MaxCharge = uint8(count), uint8(count)
	data.Progress = 100
	return next, true
}

func wandCooldown5361B0(token string) uint32 {
	return server.WandCooldownFrames5361B0(gameFPS(), legacycrt.Atof(token))
}

//export sub_5361B0
func sub_5361B0(args *C.char, output unsafe.Pointer) C.int {
	data := (*server.WandUseData)(output)
	data.Kind = 1
	next, ok := parseWandCharges(args, data)
	if !ok {
		return 0
	}
	rate, next := nextCollideSpaceToken(next)
	if rate == "" {
		return 0
	}
	data.Cooldown = wandCooldown5361B0(rate)
	name, _ := nextCollideSpaceToken(next)
	if name == "" {
		return 0
	}
	data.Spell = uint32(nox_xxx_spellNameToN_4243F0(internCStr(name)))
	return 1
}

//export sub_536260
func sub_536260(args *C.char, output unsafe.Pointer) C.int {
	data := (*server.WandUseData)(output)
	data.Kind = 0
	next, ok := parseWandCharges(args, data)
	if !ok {
		return 0
	}
	name, next := nextCollideSpaceToken(next)
	if name == "" {
		return 0
	}
	alloc.StrCopyZero(data.ProjectileName[:], name)
	data.ProjectileType = 0
	rate, next := nextCollideSpaceToken(next)
	if rate == "" {
		return 0
	}
	data.Cooldown = wandCooldown5361B0(rate)
	flag, next := nextCollideSpaceToken(next)
	if flag == "MULTI_SHOT" {
		data.Flags |= 1
	}
	if name, _ := nextCollideSpaceToken(next); name != "" {
		data.Sound = uint32(sound.ByName(name))
	}
	return 1
}

func scanThingWord(args string) (word, rest string) {
	const whitespace = " \t\n\r\f\v"
	args = strings.TrimLeft(args, whitespace)
	if end := strings.IndexAny(args, whitespace); end >= 0 {
		return args[:end], args[end:]
	}
	return args, ""
}

//export sub_5364E0
func sub_5364E0(args *C.char, output unsafe.Pointer) C.int {
	data := (*server.SkullUpdateData)(output)
	if name, _ := scanThingWord(GoString(args)); name != "" {
		alloc.StrCopyZero(data.ProjectileName[:], name)
	}
	data.ProjectileType = 0
	return 1
}

//export sub_536550
func sub_536550(args *C.char, output *C.uint32_t) C.int {
	data := (*server.PushUpdateData53B030)(unsafe.Pointer(output))
	if radius, rest, ok := legacycrt.ScanFloat32(GoString(args)); ok {
		data.Radius = radius
		if force, _, ok := legacycrt.ScanFloat32(rest); ok {
			data.Force = force
		}
	}
	data.RadiusCopy = data.Radius
	return 1
}

func scanThingInt(args *C.char) (int32, *C.char) {
	value, consumed := scanLegacyLongDecimal((*byte)(unsafe.Pointer(args)))
	if consumed == 0 {
		return 0, nil
	}
	return int32(value), (*C.char)(unsafe.Add(unsafe.Pointer(args), consumed))
}

//export sub_536580
func sub_536580(args *C.char, output *C.int32_t) C.int {
	data := (*[3]int32)(unsafe.Pointer(output))
	for i := range data {
		value, next := scanThingInt(args)
		if next == nil {
			break
		}
		data[i], args = value, next
	}
	return 1
}

//export sub_5365B0
func sub_5365B0(args *C.char, output unsafe.Pointer) C.int {
	data := (*server.TriggerUpdateData)(output)
	if name, next := nextCollideSpaceToken(args); name != "" {
		data.SoundActivate = uint32(sound.ByName(name))
		if other, _ := nextCollideSpaceToken(next); other != "" {
			data.SoundDeactivate = uint32(sound.ByName(other))
		}
	}
	return 1
}

//export sub_536600
func sub_536600(args *C.char, output *C.int32_t) C.int {
	if value, next := scanThingInt(args); next != nil {
		*output = C.int32_t(value)
	}
	return 1
}

//export sub_536B40
func sub_536B40(args *C.char, output unsafe.Pointer) C.int {
	data := (*server.CreateSpawnObjectDeathData54E010)(output)
	name, rest := scanThingWord(GoString(args))
	if name != "" {
		alloc.StrCopyZero(data.TypeID[:], name)
	}
	effect, _ := scanThingWord(rest)
	data.Sound = uint32(sound.ByName(effect))
	return 1
}

//export nox_xxx_spellTurnUndeadUpdate_531410
func nox_xxx_spellTurnUndeadUpdate_531410() C.int {
	return 0
}

//export nox_xxx_unitIsMimic_534840
func nox_xxx_unitIsMimic_534840(unit *nox_object_t) C.int {
	cache := memmap.PtrUint32(0x5D4594, 2488524)
	if *cache == 0 {
		*cache = uint32(GetServer().S().Types.IndByID("Mimic"))
	}
	return C.int(bool2int(uint32(unit.typ_ind) == *cache))
}
