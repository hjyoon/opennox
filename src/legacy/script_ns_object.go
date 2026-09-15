package legacy

/*
#include "defs.h"

unsigned int sub_516D00(nox_object_t* a1);
int nox_xxx_destroyEveryChatMB_528D60();
nox_object_t* nox_xxx_getObjectByScrName_4DA4F0(char* a1);
int nox_xxx_playDialogFile_44D900(unsigned char* a1, int a2);
int sub_467B00(int type_id, int amount);
extern int dword_5d4594_2386848;
extern unsigned int dword_5d4594_2386852;
*/
import "C"
import (
	"github.com/opennox/libs/types"

	"github.com/opennox/opennox/v1/common/memmap"
	"github.com/opennox/opennox/v1/server"
)

var (
	Nox_xxx_inventoryServPlace_4F36F0    func(obj, it *server.Object, a3, a4 int) bool
	Nox_xxx_inventoryServPlaceRaw_4F36F0 func(obj, it *server.Object, a3, a4 int32) int32
)

func Nox_xxx_getObjectByScrName_4DA4F0(name string) *server.Object {
	cstr := CString(name)
	defer StrFree(cstr)
	return asObjectS(C.nox_xxx_getObjectByScrName_4DA4F0(cstr))
}
func Nox_server_scriptMoveTo_5123C0(a1 *server.Object, a2 *server.Waypoint) {
	GetServer().S().ScriptMoveTo5123C0(a1, a2, server.ScriptMoveRuntime5123C0{
		MoverTypeID: func() uint32 {
			return uint32(Get_dword_5d4594_2386836())
		},
		SetOn: Nox_xxx_objectSetOn_4E75B0,
	})
}
func Nox_xxx_playerCanCarryItem_513B00(a1 *server.Object, a2 *server.Object) {
	s := GetServer().S()
	glyph := memmap.PtrUint32(0x5D4594, 2386856)
	if *glyph == 0 {
		*glyph = uint32(s.Types.IndByID("Glyph"))
	}
	s.PlayerCanCarryItem513B00(a1, a2, *glyph, server.PlayerCanCarryItemRuntime513B00{
		InventoryCapacity: func(typeID uint16, amount int32) int32 {
			return int32(C.sub_467B00(C.int(typeID), C.int(amount)))
		},
		PickupCount: func() int32 { return int32(C.dword_5d4594_2386848) },
		ItemCost:    s.ShopItemCostNoMerchant50E3D0,
		RandomPoint: func(radius float32, center, output *types.Pointf) {
			s.RandomReachablePointAroundInto4ED970(radius, center, output)
		},
		Drop:          objectDropDispatchCall4ED790,
		AlreadyWarned: func() bool { return C.dword_5d4594_2386852 != 0 },
		Warn: func(owner *server.Object) {
			s.NetPriMsgToPlayer(owner, "pickup.c:CarryingTooMuch", 0)
			C.dword_5d4594_2386852 = 1
		},
	})
}

func Sub_516D00(a1 *server.Object) {
	GetServer().S().ScriptRaiseZombie516D00(a1, monsterRaiseZombieRuntime534AB0())
}
func Nox_xxx_netSendChat_528AC0(a1 *server.Object, a2 string, a3 uint16) {
	GetServer().S().NetSendChat528AC0(a1, a2, a3)
}
