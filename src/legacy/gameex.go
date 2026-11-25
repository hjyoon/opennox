package legacy

/*
#include "defs.h"
#include "common__system__team.h"
#include "client__gui__window.h"
int sub_4BDFD0(void);
int sub_4BE610(void);
int  modifyWndInputHandler(int a1, int a2, int a3, int a4);
int  nox_xxx_clientUpdateButtonRow_45E110(int a1);
nox_object_team_t* nox_xxx_objGetTeamByNetCode_418C80(int a1);
void  nox_xxx_printCentered_445490(wchar2_t* a1);
*/
import "C"
import (
	"unsafe"

	"github.com/opennox/libs/types"

	"github.com/opennox/opennox/v1/server"
)

func Sub_4BDFD0() {
	C.sub_4BDFD0()
}
func Sub_4BE610() {
	C.sub_4BE610()
}
func Mix_MouseKeyboardWeaponRoll(a1 *server.Object, a2 int8) int {
	return int(weaponRollNative10001EE0(a1, a2, weaponRollNativeDeps10001EE0{
		loadWeaponFlags: objectNPCWeaponEquipFlags,
		classCanUse:     Nox_xxx_playerClassCanUseItem_57B3D0,
		checkStrength:   Nox_xxx_playerCheckStrength_4F3180,
		tryDequip:       Nox_xxx_playerTryDequip_4F2FB0,
		tryEquip:        Nox_xxx_playerTryEquip_4F2F70,
	}))
}

// GameEx's optional name messages compare the low bytes of the approved UTF-16 name.
func gameExPlayerNameMatches(query *byte, name []C.wchar2_t) bool {
	for i, unit := range name {
		got := *(*byte)(unsafe.Add(unsafe.Pointer(query), i))
		if got != byte(unit) {
			return false
		}
		if got == 0 {
			return true
		}
	}
	return *(*byte)(unsafe.Add(unsafe.Pointer(query), len(name))) == 0
}

func PlayerInfoStructParser_0(record unsafe.Pointer) int {
	return PlayerInfoStructParser_1(record, nil)
}

func PlayerInfoStructParser_1(record unsafe.Pointer, teamOut *unsafe.Pointer) int {
	if record == nil || uintptr(record) == ^uintptr(1) {
		return 0
	}
	name := (*byte)(unsafe.Add(record, 2))
	for player := nox_common_playerInfoGetFirst_416EA0(); player != nil; player = nox_common_playerInfoGetNext_416EE0(player) {
		if !gameExPlayerNameMatches(name, player.name_final[:]) {
			continue
		}
		team := nox_xxx_objGetTeamByNetCode_418C80(C.int(player.netCode))
		if team == nil {
			return 0
		}
		if teamOut != nil {
			*teamOut = unsafe.Pointer(team)
		}
		result := unsafe.Slice((*byte)(record), 2)
		result[1] = byte(team.id)
		result[0] = byte(asPlayerS(player).PlayerClass())
		return 1
	}
	return 0
}
func PlayerDropATrap(a1 *server.Object) {
	playerDropATrapNative10002030(a1, func(owner, item *server.Object, point *types.Pointf) {
		objectDropBoundedCall4ED810(owner, item, point)
	})
}
func GetFlagValueFromFlagIndex(index int) uint32 {
	exponent := int32(index)
	if exponent < 0 {
		return 0
	}
	return uint32(1) << uint32(exponent)
}

var gameExFlags uint32 = 0x1e

func Get_gameex_flags() uint32 {
	return gameExFlags
}

func Get_gameex_flags_ptr() *uint32 {
	return &gameExFlags
}
