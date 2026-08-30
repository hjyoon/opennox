package legacy

/*
#include "defs.h"
#include "GAME1.h"
#include "GAME1_1.h"
#include "GAME3_2.h"
#include "GAME3_3.h"
#include "GAME4.h"
#include "GAME4_1.h"
#include "GAME4_3.h"
#include "GAME5.h"
#include "server__magic__plyrspel.h"

extern uint32_t nox_tile_def_cnt;
extern unsigned int dword_5d4594_2650652;

void nox_xxx_updateProjectile_53AC10(nox_object_t* a1);
void nox_xxx_updateDeathBall_53D080(nox_object_t* a1);
void nox_xxx_updateOneSecondDie_53CB60_go(nox_object_t* source);
void nox_xxx_updateExpire_53DB00_go(nox_object_t* source);
static int nox_call_objectType_parseUpdate_go(int (*fnc)(char*, void*), char* arg1, void* arg2) { return fnc(arg1, arg2); }
void nox_xxx___mkgmtime_538280(nox_object_t* a1);
*/
import "C"
import (
	"fmt"
	"strings"
	"unsafe"

	"github.com/opennox/libs/player"

	"github.com/opennox/opennox/v1/server"
)

var (
	Nox_xxx_updatePlayer_4F8100         func(up *server.Object)
	Nox_xxx_updatePixie_53CD20          func(cobj *server.Object)
	Nox_xxx_updatePlayerObserver_4E62F0 func(a1p *server.Object)
	Nox_xxx_updateProjectile_53AC10     func(a1 *server.Object)
	Nox_xxx_updateDeathBall_53D080      func(a1 *server.Object)
	Nox_xxx___mkgmtime_538280           func(a1 *server.Object)
)

func Get_nox_xxx_updateTrigger_53B1B0() unsafe.Pointer {
	return C.nox_xxx_updateTrigger_53B1B0
}

func Get_nox_xxx_updateElevator_53B5D0() unsafe.Pointer {
	return C.nox_xxx_updateElevator_53B5D0
}

func Get_nox_xxx_updateElevatorShaft_53B380() unsafe.Pointer {
	return C.nox_xxx_updateElevatorShaft_53B380
}

func Get_nox_xxx_updateTrapDoor_53DE80() unsafe.Pointer {
	return C.nox_xxx_updateTrapDoor_53DE80
}

func Get_nox_xxx_updateSwitch_53B320() unsafe.Pointer {
	return C.nox_xxx_updateSwitch_53B320
}

func Get_nox_xxx_updateToggle_53B060() unsafe.Pointer {
	return C.nox_xxx_updateToggle_53B060
}

// 189 native pointer slots plus one pointer-alignment unit on 64-bit targets.
// This includes the sixteen path-waypoint pointers at PE32 fields 75..90.
// The 32-bit layout remains the original 2200-byte GAME.EXE record.
const monsterUpdateDataNativeSize = 2200 + 190*(cgoABIPointerSize-4)

var _ = [1]struct{}{}[monsterUpdateDataNativeSize-unsafe.Sizeof(server.MonsterUpdateData{})]

// Twenty-one pointer/native-word slots widen in place. On 64-bit targets they
// also introduce three four-byte internal alignment gaps and four bytes of
// trailing alignment. The 32-bit record remains the original 556 bytes.
const playerUpdateDataNativeSize = 556 + 25*(cgoABIPointerSize-4)

var _ = [1]struct{}{}[playerUpdateDataNativeSize-unsafe.Sizeof(server.PlayerUpdateData{})]

func init() {
	_ = nox_xxx_updatePlayer_4F8100
	server.RegisterObjectUpdate("PlayerUpdate", C.nox_xxx_updatePlayer_4F8100, unsafe.Sizeof(server.PlayerUpdateData{}))
	_ = nox_xxx_updateProjectile_53AC10
	server.RegisterObjectUpdate("ProjectileUpdate", C.nox_xxx_updateProjectile_53AC10, 0)
	server.RegisterObjectUpdate("SpellProjectileUpdate", C.nox_xxx_spellFlyUpdate_53B940, unsafe.Sizeof(server.SpellProjectileUpdateData{}))
	server.RegisterObjectUpdate("AntiSpellProjectileUpdate", C.nox_xxx_updateAntiSpellProj_53BB00, 28)
	server.RegisterObjectUpdate("DoorUpdate", C.nox_xxx_updateDoor_53AC50, unsafe.Sizeof(server.DoorUpdateData{}))
	server.RegisterObjectUpdate("SparkUpdate", C.nox_xxx_updateSpark_53ADC0, 16)
	server.RegisterObjectUpdate("ProjectileTrailUpdate", C.nox_xxx_updateProjTrail_53AEC0, 0)
	server.RegisterObjectUpdate("PushUpdate", C.nox_xxx_updatePush_53B030, 12)
	server.RegisterObjectUpdate("TriggerUpdate", C.nox_xxx_updateTrigger_53B1B0, unsafe.Sizeof(server.TriggerUpdateData{}))
	server.RegisterObjectUpdate("ToggleUpdate", C.nox_xxx_updateToggle_53B060, 60)
	server.RegisterObjectUpdate("MonsterUpdate", C.nox_xxx_unitUpdateMonster_50A5C0, unsafe.Sizeof(server.MonsterUpdateData{}))
	server.RegisterObjectUpdate("LoopAndDamageUpdate", C.sub_53B300, 16)
	server.RegisterObjectUpdate("ElevatorUpdate", C.nox_xxx_updateElevator_53B5D0, unsafe.Sizeof(server.ElevatorUpdateData{}))
	server.RegisterObjectUpdate("ElevatorShaftUpdate", C.nox_xxx_updateElevatorShaft_53B380, unsafe.Sizeof(server.ElevatorShaftUpdateData{}))
	server.RegisterObjectUpdate("PhantomPlayerUpdate", C.nox_xxx_updatePhantomPlayer_53B860, 0)
	server.RegisterObjectUpdate("ObeliskUpdate", C.nox_xxx_updateObelisk_53C580, unsafe.Sizeof(server.ObeliskUpdateData{}))
	server.RegisterObjectUpdate("LifetimeUpdate", C.nox_xxx_updateLifetime_53B8F0, 4)
	server.RegisterObjectUpdate("MagicMissileUpdate", C.nox_xxx_updateMagicMissile_53BDA0, 28)
	server.RegisterObjectUpdate("PixieUpdate", C.nox_xxx_updatePixie_53CD20, unsafe.Sizeof(server.PixieUpdateData{}))
	server.RegisterObjectUpdate("SkullUpdate", C.nox_xxx_updateShootingTrap_54F9A0, 52)
	server.RegisterObjectUpdate("PentagramUpdate", C.nox_xxx_updateTeleportPentagram_53BEF0, unsafe.Sizeof(server.PentagramUpdateData{}))
	server.RegisterObjectUpdate("InvisiblePentagramUpdate", C.nox_xxx_updateInvisiblePentagram_53C0C0, unsafe.Sizeof(server.PentagramUpdateData{}))
	server.RegisterObjectUpdate("SwitchUpdate", C.nox_xxx_updateSwitch_53B320, 0)
	server.RegisterObjectUpdate("BlowUpdate", C.nox_xxx_updateBlow_53C160, 0)
	server.RegisterObjectUpdate("MoverUpdate", C.nox_xxx_unitUpdateMover_54F740, unsafe.Sizeof(server.MoverUpdateData{}))
	server.RegisterObjectUpdate("BlackPowderBarrelUpdate", C.nox_xxx_updateBlackPowderBarrel_53C9A0, 0)
	server.RegisterObjectUpdate("OneSecondDieUpdate", C.nox_xxx_updateOneSecondDie_53CB60_go, 0)
	server.RegisterObjectUpdate("WaterBarrelUpdate", C.nox_xxx_updateWaterBarrel_53CB90, 0)
	server.RegisterObjectUpdate("SelfDestructUpdate", C.nox_xxx_updateSelfDestruct_53CC90, 0)
	server.RegisterObjectUpdate("BlackPowderBurnUpdate", C.nox_xxx_updateBlackPowderBurn_53CCB0, 0)
	server.RegisterObjectUpdate("DeathBallUpdate", C.nox_xxx_updateDeathBall_53D080, 0)
	server.RegisterObjectUpdate("DeathBallFragmentUpdate", C.nox_xxx_updateDeathBallFragment_53D220, 0)
	server.RegisterObjectUpdate("MoonglowUpdate", C.nox_xxx_updateMoonglow_53D270, 0)
	server.RegisterObjectUpdate("SentryGlobeUpdate", C.nox_xxx_updateSentryGlobe_510E60, unsafe.Sizeof(server.SentryUpdateData{}))
	server.RegisterObjectUpdate("TelekinesisUpdate", C.nox_xxx_updateTelekinesis_53D330, 0)
	server.RegisterObjectUpdate("FistUpdate", C.nox_xxx_updateFist_53D400, unsafe.Sizeof(server.FistUpdateData{}))
	server.RegisterObjectUpdate("MeteorShowerUpdate", C.nox_xxx_updateMeteorShower_53D5A0, 4)
	server.RegisterObjectUpdate("MeteorUpdate", C.nox_xxx_meteorExplode_53D6E0, 4)
	server.RegisterObjectUpdate(
		"ToxicCloudUpdate",
		C.nox_xxx_updateToxicCloud_53D850,
		unsafe.Sizeof(server.ToxicCloudUpdateData{}),
	)
	server.RegisterObjectUpdate(
		"SmallToxicCloudUpdate",
		C.nox_xxx_updateSmallToxicCloud_53D960,
		unsafe.Sizeof(server.ToxicCloudUpdateData{}),
	)
	server.RegisterObjectUpdate("ArachnaphobiaUpdate", C.nox_xxx_updateArachnaphobia_53DA60, 0)
	server.RegisterObjectUpdate("ExpireUpdate", C.nox_xxx_updateExpire_53DB00_go, 0)
	server.RegisterObjectUpdate("BreakUpdate", C.nox_xxx_updateBreak_53DB30, 0)
	server.RegisterObjectUpdate("OpenUpdate", C.nox_xxx_updateOpen_53DBB0, 0)
	server.RegisterObjectUpdate("BreakAndRemoveUpdate", C.nox_xxx_updateBreakAndRemove_53DC30, 0)
	server.RegisterObjectUpdate(
		"ChakramInMotionUpdate",
		C.nox_xxx_updateChakramInMotion_53DCC0,
		unsafe.Sizeof(server.ChakramUpdateData{}),
	)
	server.RegisterObjectUpdate("FlagUpdate", C.nox_xxx_updateFlag_53DDF0, 12)
	server.RegisterObjectUpdate("TrapDoorUpdate", C.nox_xxx_updateTrapDoor_53DE80, 0)
	// Allocate the native-width record used by restored Go boundaries. The
	// remaining C updater at GAME.EXE 0053DF40 is still ABI32-only and must be
	// lifted before BallUpdate can execute in a complete 64-bit product.
	server.RegisterObjectUpdate(
		"BallUpdate",
		C.nox_xxx_updateGameBall_53DF40,
		unsafe.Sizeof(server.GameBallUpdateData4EA800{}),
	)
	server.RegisterObjectUpdate(
		"CrownUpdate",
		C.nox_xxx_updateCrown_53E1D0,
		unsafe.Sizeof(server.CrownUpdateData{}),
	)
	server.RegisterObjectUpdate("UndeadKillerUpdate", C.nox_xxx_updateUndeadKiller_53E190, 0)
	server.RegisterObjectUpdate("HarpoonUpdate", C.nox_xxx_updateHarpoon_54F380, 4)
	server.RegisterObjectUpdate("MonsterGeneratorUpdate", C.nox_xxx_updateMonsterGenerator_54E930, 164)

	server.RegisterObjectUpdateParse("PushUpdate", wrapObjectUpdateParseC(C.sub_536550))
	server.RegisterObjectUpdateParse("TriggerUpdate", wrapObjectUpdateParseC(C.sub_5365B0))
	server.RegisterObjectUpdateParse("ToggleUpdate", wrapObjectUpdateParseC(C.sub_5365B0))
	server.RegisterObjectUpdateParse("LoopAndDamageUpdate", wrapObjectUpdateParseC(C.sub_536580))
	server.RegisterObjectUpdateParse("LifetimeUpdate", wrapObjectUpdateParseC(C.sub_536600))
	server.RegisterObjectUpdateParse("SkullUpdate", wrapObjectUpdateParseC(C.sub_5364E0))
}

//export nox_xxx_updatePlayer_4F8100
func nox_xxx_updatePlayer_4F8100(up *nox_object_t) { Nox_xxx_updatePlayer_4F8100(asObjectS(up)) }

//export nox_xxx_objectApplyForce_52DF80
func nox_xxx_objectApplyForce_52DF80(vec *C.float, obj *nox_object_t, force C.float) {
	GetServer().ApplyForce(asObjectS(obj), AsPointf(unsafe.Pointer(vec)), float64(force))
}

//export nox_xxx_updatePixie_53CD20
func nox_xxx_updatePixie_53CD20(cobj *nox_object_t) { Nox_xxx_updatePixie_53CD20(asObjectS(cobj)) }

//export nox_xxx_enemyAggro_5335D0
func nox_xxx_enemyAggro_5335D0(cobj *nox_object_t, r float32) *nox_object_t {
	return asObjectC(GetServer().S().EnemyAggroYyy(asObjectS(cobj), r))
}

//export sub_5336D0
func sub_5336D0(cobj *nox_object_t) C.double {
	return C.double(GetServer().S().Sub5336D0(asObjectS(cobj)))
}

//export nox_xxx_updatePlayerObserver_4E62F0
func nox_xxx_updatePlayerObserver_4E62F0(a1p *nox_object_t) {
	Nox_xxx_updatePlayerObserver_4E62F0(asObjectS(a1p))
}

//export nox_xxx_updateProjectile_53AC10
func nox_xxx_updateProjectile_53AC10(a1 *nox_object_t) {
	Nox_xxx_updateProjectile_53AC10(asObjectS(a1))
}

//export nox_xxx_updateDeathBall_53D080
func nox_xxx_updateDeathBall_53D080(a1 *nox_object_t) {
	Nox_xxx_updateDeathBall_53D080(asObjectS(a1))
}

var oneSecondDieUpdateCall53CB60 = func(source *server.Object) {
	outer := GetServer()
	outer.S().OneSecondDieUpdate53CB60(source, server.OneSecondDieUpdateRuntime53CB60{
		DelayedDelete: outer.DelayedDelete,
	})
}

func oneSecondDieUpdateExportCall53CB60(source *server.Object) {
	C.nox_xxx_updateOneSecondDie_53CB60_go(asObjectC(source))
}

//export nox_xxx_updateOneSecondDie_53CB60_go
func nox_xxx_updateOneSecondDie_53CB60_go(source *nox_object_t) {
	oneSecondDieUpdateCall53CB60(asObjectS(source))
}

var expireUpdateCall53DB00 = func(source *server.Object) {
	outer := GetServer()
	outer.S().ExpireUpdate53DB00(source, server.ExpireUpdateRuntime53DB00{
		DelayedDelete: outer.DelayedDelete,
	})
}

func expireUpdateExportCall53DB00(source *server.Object) {
	C.nox_xxx_updateExpire_53DB00_go(asObjectC(source))
}

//export nox_xxx_updateExpire_53DB00_go
func nox_xxx_updateExpire_53DB00_go(source *nox_object_t) {
	expireUpdateCall53DB00(asObjectS(source))
}

//export nox_xxx___mkgmtime_538280
func nox_xxx___mkgmtime_538280(a1 *nox_object_t) {
	Nox_xxx___mkgmtime_538280(asObjectS(a1))
}

func Get_nox_xxx___mkgmtime_538280() unsafe.Pointer {
	return C.nox_xxx___mkgmtime_538280
}

func wrapObjectUpdateParseC(ptr unsafe.Pointer) server.ObjectParseFunc {
	return func(objt *server.ObjectType, args []string) error {

		if Nox_call_objectType_parseUpdate_go(ptr, strings.Join(args, " "), objt.UpdateData) == 0 {
			return fmt.Errorf("cannot parse update data for %q", objt.ID())
		}
		return nil
	}
}

func Nox_server_doPlayersAutoRespawn_40A5F0() int {
	return int(C.nox_server_doPlayersAutoRespawn_40A5F0())
}
func Sub_4E4100() uint32 {
	return uint32(C.sub_4E4100())
}
func Nox_call_objectType_parseUpdate_go(a1 unsafe.Pointer, a2 string, a3 unsafe.Pointer) int {
	cstr := CString(a2)
	defer StrFree(cstr)
	return int(C.nox_call_objectType_parseUpdate_go((*[0]byte)(a1), cstr, a3))
}
func Nox_xxx_questCheckSecretArea_421C70(a1 *server.Object) {
	C.nox_xxx_questCheckSecretArea_421C70(asObjectC(a1))
}

//export nox_server_playerCanMove_4F9BC0
func nox_server_playerCanMove_4F9BC0(unit *nox_object_t) C.int {
	return C.int(server.PlayerCanMove4F9BC0(asObjectS(unit)))
}

func Nox_xxx_playerCanMove_4F9BC0(a1 *server.Object) int {
	return int(server.PlayerCanMove4F9BC0(a1))
}
func Sub_4F9AB0(a1 *server.Object) int {
	return int(C.sub_4F9AB0(asObjectC(a1)))
}
func Nox_xxx_playerConfusedGetDirection_4F7A40(a1 *server.Object) server.Dir16 {
	return GetServer().S().PlayerConfusedDirection4F7A40(a1)
}

//export nox_server_playerCanAttack_4F9C40
func nox_server_playerCanAttack_4F9C40(unit *nox_object_t) C.int {
	return C.int(server.PlayerCanAttack4F9C40(asObjectS(unit)))
}

func Nox_xxx_playerAttack_538960(a1 *server.Object) int {
	return int(C.nox_xxx_playerAttackNative_538960(asObjectC(a1)))
}
func Nox_xxx_playerRespawn_4F7EF0(a1 *server.Object) {
	C.nox_xxx_playerRespawn_4F7EF0(asObjectC(a1))
}
func Sub_4F9E10(a1 *server.Object) int {
	return int(C.sub_4F9E10(asObjectC(a1)))
}
func Sub_4F9A80(a1 *server.Object) int {
	return int(C.sub_4F9A80(asObjectC(a1)))
}
func Nox_xxx_monsterTestBlockShield_533E70(a1 *server.Object) int {
	return int(C.nox_xxx_monsterTestBlockShield_533E70(asObjectC(a1)))
}
func Nox_common_mapPlrActionToStateId_4FA2B0(a1 *server.Object) int {
	return int(C.nox_common_mapPlrActionToStateId_4FA2B0(asObjectC(a1)))
}
func Nox_xxx_playerCanAttack_4F9C40(a1 *server.Object) int {
	return int(server.PlayerCanAttack4F9C40(a1))
}
func Nox_xxx_checkWinkFlags_4F7DF0(a1 *server.Object) int {
	return int(winkGameBallReleaseCall4F7DF0(a1))
}
func Nox_xxx_playerInputAttack_4F9C70(a1 *server.Object) {
	C.nox_xxx_playerInputAttack_4F9C70(asObjectC(a1))
}
func Nox_xxx_playerSubStamina_4F7D30(a1 *server.Object, a2 int) int {
	return int(GetServer().S().PlayerSubStamina4F7D30(a1, int32(a2)))
}
func Nox_xxx_playerDoSchedSpell_4FB0E0(a1 *server.Object, a2 *server.Object) {
	C.nox_xxx_playerDoSchedSpell_4FB0E0(asObjectC(a1), asObjectC(a2))
}
func Nox_xxx_playerDoSchedSpellQueue_4FB1D0(a1 *server.Object, a2 *server.Object) {
	C.nox_xxx_playerDoSchedSpellQueue_4FB1D0(asObjectC(a1), asObjectC(a2))
}
func Sub_4E7540(a1 *server.Object, a2 *server.Object) {
	recordPlayerAttributionRuntime4E7540(a1, a2)
}
func Nox_xxx_playerCheckStrength_4F3180(a1 *server.Object, a2 *server.Object) bool {
	return playerCheckStrengthCall4F3180(a1, a2) != 0
}
func Nox_xxx_unitDamageClear_4EE5E0(a1 *server.Object, a2 int) {
	unitDamageClearCall4EE5E0(a1, int32(a2))
}
func Nox_xxx_unitAdjustHP_4EE460(a1 *server.Object, a2 int) {
	unitAdjustHPCall4EE460(a1, int32(a2))
}
func Sub_509CF0(a1 *byte, a2 player.Class, a3 uint32) int {
	return int(C.sub_509CF0((*C.char)(unsafe.Pointer(a1)), C.char(a2), C.int(a3)))
}
func Sub_4D79C0(a1 *server.Object) {
	C.sub_4D79C0(asObjectC(a1))
}
func Sub_4D7480(a1 *server.Object) {
	C.sub_4D7480(asObjectC(a1))
}
