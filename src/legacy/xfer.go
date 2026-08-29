package legacy

/*
#include "defs.h"
#include "GAME1.h"
#include "GAME3_3.h"
#include "GAME4.h"
#include "GAME4_2.h"
#include "server__script__script.h"
#include "xfer_door_4f4cb0.h"
#include "xfer_trigger_4f4e50.h"
#include "xfer_hole_4f51d0.h"
#include "xfer_transporter_4f5300.h"
#include "xfer_elevator_4f53d0.h"
*/
import "C"
import (
	"unsafe"

	"github.com/opennox/opennox/v1/internal/cryptfile"
	"github.com/opennox/opennox/v1/server"
)

var (
	Nox_xxx_xfer_saveObj51DF90        func(cf *cryptfile.CryptFile, a1p *server.Object) int
	Nox_xxx_XFer_ReadShopItem_52A840  func(a1 unsafe.Pointer, a2 int)
	Nox_xxx_XFer_WriteShopItem_52A5F0 func(a1 unsafe.Pointer)
)

func init() {
	server.RegisterObjectXfer("DefaultXfer", C.nox_xxx_XFerDefault_4F49A0)
	server.RegisterObjectXfer("SpellPagePedestalXfer", C.nox_xxx_XFerSpellPagePedistal_4F4A20)
	server.RegisterObjectXfer("SpellRewardXfer", C.nox_xxx_XFerSpellReward_4F5F30)
	server.RegisterObjectXfer("AbilityRewardXfer", C.nox_xxx_XFerAbilityReward_4F6240)
	server.RegisterObjectXfer("FieldGuideXfer", C.nox_xxx_XFerFieldGuide_4F6390)
	server.RegisterObjectXfer("ReadableXfer", C.nox_xxx_XFerReadable_4F4AB0)
	server.RegisterObjectXfer("ExitXfer", C.nox_xxx_XFerExit_4F4B90)
	server.RegisterObjectXfer("DoorXfer", C.nox_xxx_XFerDoor_4F4CB0)
	server.RegisterObjectXfer("TriggerXfer", C.nox_xxx_unitTriggerXfer_4F4E50)
	server.RegisterObjectXfer("MonsterXfer", C.nox_xxx_XFerMonster_528DB0)
	server.RegisterObjectXfer("HoleXfer", C.nox_xxx_XFerHole_4F51D0)
	server.RegisterObjectXfer("TransporterXfer", C.nox_xxx_XFerTransporter_4F5300)
	server.RegisterObjectXfer("ElevatorXfer", C.nox_xxx_XFerElevator_4F53D0)
	server.RegisterObjectXfer("ElevatorShaftXfer", C.nox_xxx_XFerElevatorShaft_4F54A0)
	server.RegisterObjectXfer("MoverXfer", C.nox_xxx_XFerMover_4F5730)
	server.RegisterObjectXfer("GlyphXfer", C.nox_xxx_XFerGlyph_4F5890)
	server.RegisterObjectXfer("InvisibleLightXfer", C.nox_xxx_XFerInvLight_4F5AA0)
	server.RegisterObjectXfer("SentryXfer", C.nox_xxx_XFerSentry_4F5E50)
	server.RegisterObjectXfer("WeaponXfer", C.nox_xxx_XFerWeapon_4F64A0)
	server.RegisterObjectXfer("ArmorXfer", C.nox_xxx_XFerArmor_4F6860)
	server.RegisterObjectXfer("TeamXfer", C.nox_xxx_XFerTeam_4F6D20)
	server.RegisterObjectXfer("GoldXfer", C.nox_xxx_XFerGold_4F6EC0)
	server.RegisterObjectXfer("AmmoXfer", C.nox_xxx_XFerAmmo_4F6B20)
	server.RegisterObjectXfer("NPCXfer", C.nox_xxx_XFerNPC_52ADE0)
	server.RegisterObjectXfer("ObeliskXfer", C.nox_xxx_XFerObelisk_4F6F60)
	server.RegisterObjectXfer("ToxicCloudXfer", C.nox_xxx_XFerToxicCloud_4F70A0)
	server.RegisterObjectXfer("MonsterGeneratorXfer", C.nox_xxx_XFerMonsterGen_4F7130)
	server.RegisterObjectXfer("RewardMarkerXfer", C.nox_xxx_XFerRewardMarker_4F74D0)
}

//export nox_xxx_xfer_saveObj_51DF90
func nox_xxx_xfer_saveObj_51DF90(a1 *nox_object_t) int32 {
	return int32(Nox_xxx_xfer_saveObj51DF90(cryptfile.Global(), asObjectS(a1)))
}

//export nox_xxx_XFer_ReadShopItem_52A840
func nox_xxx_XFer_ReadShopItem_52A840(a1 unsafe.Pointer, a2_cgo int32) {
	a2 := int(a2_cgo)
	Nox_xxx_XFer_ReadShopItem_52A840(a1, a2)
}

//export nox_xxx_XFer_WriteShopItem_52A5F0
func nox_xxx_XFer_WriteShopItem_52A5F0(a1 unsafe.Pointer) {
	Nox_xxx_XFer_WriteShopItem_52A5F0(a1)
}

func Get_nox_xxx_XFerFieldGuide_4F6390() unsafe.Pointer {
	return C.nox_xxx_XFerFieldGuide_4F6390
}

func Get_nox_xxx_XFerAbilityReward_4F6240() unsafe.Pointer {
	return C.nox_xxx_XFerAbilityReward_4F6240
}

func Get_nox_xxx_XFerDoor_4F4CB0() unsafe.Pointer {
	return C.nox_xxx_XFerDoor_4F4CB0
}

func Get_nox_xxx_UnitTriggerXfer_4F4E50() unsafe.Pointer {
	return C.nox_xxx_unitTriggerXfer_4F4E50
}

func Get_nox_xxx_XFerExit_4F4B90() unsafe.Pointer {
	return C.nox_xxx_XFerExit_4F4B90
}

func Get_nox_xxx_XFerHole_4F51D0() unsafe.Pointer {
	return C.nox_xxx_XFerHole_4F51D0
}

func Get_nox_xxx_XFerTransporter_4F5300() unsafe.Pointer {
	return C.nox_xxx_XFerTransporter_4F5300
}

func Get_nox_xxx_XFerElevator_4F53D0() unsafe.Pointer {
	return C.nox_xxx_XFerElevator_4F53D0
}

func Get_nox_xxx_XFerElevatorShaft_4F54A0() unsafe.Pointer {
	return C.nox_xxx_XFerElevatorShaft_4F54A0
}

func Get_nox_xxx_XFerMover_4F5730() unsafe.Pointer {
	return C.nox_xxx_XFerMover_4F5730
}

func Get_nox_xxx_XFerGlyph_4F5890() unsafe.Pointer {
	return C.nox_xxx_XFerGlyph_4F5890
}

func objectMapScriptHandlerNative4F4530(object *server.Object, context unsafe.Pointer) int32 {
	return int32(C.nox_xxx_xferReadScriptHandler_4F5580(
		unsafe.Pointer(&object.ScriptPickup),
		(*C.char)(context),
	))
}

func Nox_xxx_mapReadWriteObjData_4F4530(a1 *server.Object, a2 int) int {
	return int(objectMapReadWriteNative4F4530(cryptfile.Global(), a1, int32(a2)))
}
