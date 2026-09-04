package legacy

/*
#include "defs.h"
#include "common__system__team.h"
#include "client__gui__window.h"
int sub_4BDFD0();
int getFlagValueFromFlagIndex(signed int a1);
int  modifyWndInputHandler(int a1, int a2, int a3, int a4);
int  nox_xxx_clientUpdateButtonRow_45E110(int a1);
nox_object_team_t* nox_xxx_objGetTeamByNetCode_418C80(int a1);
void  nox_xxx_printCentered_445490(wchar2_t* a1);
char  playerDropATrap(int playerObj);
char playerInfoStructParser_0(void* a1);
char playerInfoStructParser_1(void* a1, int* a3);
*/
import "C"
import (
	"unsafe"

	"github.com/opennox/opennox/v1/server"
)

func Sub_4BDFD0() {
	C.sub_4BDFD0()
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
func PlayerInfoStructParser_0(a1 unsafe.Pointer) int {
	return int(C.playerInfoStructParser_0(a1))
}
func PlayerInfoStructParser_1(a1 unsafe.Pointer, a2 *int32) int {
	return int(C.playerInfoStructParser_1(a1, (*C.int)(unsafe.Pointer(a2))))
}
func PlayerDropATrap(a1 *server.Object) {
	C.playerDropATrap(C.int(uintptr(a1.CObj())))
}
func GetFlagValueFromFlagIndex(a1 int) uint32 {
	return uint32(C.getFlagValueFromFlagIndex(C.int(a1)))
}
