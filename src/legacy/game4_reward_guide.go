package legacy

/*
#include <string.h>
#include "GAME4.h"
#include "GAME3_3.h"
*/
import "C"

import (
	"encoding/binary"
	"unsafe"

	"github.com/opennox/libs/object"
	"github.com/opennox/opennox/v1/server"
)

type game4RewardRuntime struct {
	send     func(int, []byte) int
	name     func(*server.Object) string
	guide    func(string) int
	bonus    func() float64
	floatInt func(float32) int32
}

var game4RewardServices = func() game4RewardRuntime {
	return game4RewardRuntime{
		send:     func(ind int, p []byte) int { return GetServer().S().NetSendPacketXxx0(ind, p, nil, 1) },
		name:     func(o *server.Object) string { return C.GoString(C.nox_xxx_getUnitName_4E39D0(asObjectC(o))) },
		guide:    playerGuideByName41B420,
		bonus:    func() float64 { return float64(nox_xxx_gamedataGetFloat_419D40(internCStr("FieldGuideDamageBonus"))) },
		floatInt: game3MapFloatInt,
	}
}

//export nox_xxx_netSendRewardNotify_4FAD50
func nox_xxx_netSendRewardNotify_4FAD50(unit *C.nox_object_t, kind C.int, target *C.nox_object_t, flags C.char) C.int {
	o := asObjectS(unit)
	// This legacy EAX value is an opaque status word, never an address carrier.
	result := C.int(int32(uintptr(unsafe.Pointer(unit))))
	if o == nil || o.ObjClass&object.ClassPlayer == 0 {
		return result
	}
	if kind < 0 || kind > 2 {
		return kind - 2
	}
	u := (*server.PlayerUpdateData)(o.UpdateData)
	if u == nil || u.Player == nil || target == nil {
		return 0
	}
	packet := [5]byte{0xf0, byte(30 + kind), byte(flags)}
	binary.LittleEndian.PutUint16(packet[3:], uint16(asObjectS(target).NetCode))
	return C.int(game4RewardServices().send(int(u.Player.PlayerInd), packet[:]))
}

//export nox_game4_reward_string_go
func nox_game4_reward_string_go(unit *C.nox_object_t, name *C.char, flags C.char) {
	o := asObjectS(unit)
	if o == nil || uint8(o.ObjClass)&uint8(object.ClassPlayer) == 0 || name == nil || o.UpdateData == nil {
		return
	}
	u := (*server.PlayerUpdateData)(o.UpdateData)
	if u.Player == nil {
		return
	}
	// Bound actual bytes. The old uint8(strlen+1) gate accepted overflowing names.
	n := int(C.strnlen(name, 49))
	if n == 0 || n > 48 {
		return
	}
	packet := [52]byte{0xf0, 33}
	copy(packet[2:], unsafe.Slice((*byte)(unsafe.Pointer(name)), n))
	packet[51] = byte(flags)
	game4RewardServices().send(int(u.Player.PlayerInd), packet[:])
}

func game4GuideLevel(unit, creature *server.Object, r game4RewardRuntime) int32 {
	if unit == nil || creature == nil || uint8(unit.ObjClass)&uint8(object.ClassPlayer) == 0 || uint8(creature.ObjClass)&uint8(object.ClassMonster) == 0 || unit.UpdateData == nil {
		return 0
	}
	u := (*server.PlayerUpdateData)(unit.UpdateData)
	guide := r.guide(r.name(creature))
	if u.Player == nil || guide <= 0 || guide >= len(u.Player.BeastScrollLvl) {
		return 0
	}
	return int32(u.Player.BeastScrollLvl[guide])
}

//export sub_4FB000
func sub_4FB000(unit, creature *C.nox_object_t) C.int {
	return C.int(game4GuideLevel(asObjectS(unit), asObjectS(creature), game4RewardServices()))
}

//export sub_4FB050
func sub_4FB050(unit, creature *C.nox_object_t, damage *C.int) C.int {
	r := game4RewardServices()
	level := game4GuideLevel(asObjectS(unit), asObjectS(creature), r)
	if level == 0 {
		return 0
	}
	// The original stores the double expression in float before conversion.
	value := float32(r.bonus()*float64(*damage) + 0.5)
	result := r.floatInt(value)
	*damage = C.int(result)
	return C.int(result)
}
