package legacy

/*
#include "GAME3_3.h"
#include "GAME4_3.h"
#include "GAME5.h"

static int nox_call_objectType_parseCollide_go(int (*fnc)(char*, void*), char* arg1, void* arg2) { return fnc(arg1, arg2); }
*/
import "C"
import (
	"fmt"
	"strings"
	"unsafe"

	"github.com/opennox/libs/spell"
	"github.com/opennox/libs/types"

	"github.com/opennox/opennox/v1/server"
)

var (
	Nox_xxx_collideDeathBall_4E9E90 func(a1, a2 *server.Object, pos *types.Pointf)
	Nox_xxx_castCounterSpell_52BBB0 func(cspl spell.ID, a2, a3, a4 *server.Object, sa *server.SpellAcceptArg, lvl int) int
	Nox_xxx_changeOwner_52BE40      func(a1, a2 *server.Object)
)

func init() {
	server.RegisterObjectCollideGo("DefaultCollide", C.nox_xxx_collideDefault_4E87A0, collideNoop, 0)
	server.RegisterObjectCollideGo(
		"MonsterCollide",
		C.nox_xxx_collideMonsterEventProc_4E83B0,
		func(monster, other *server.Object, collision unsafe.Pointer) {
			monsterCollideCall4E83B0(monster, other, collision)
		},
		0,
	)
	server.RegisterObjectCollideGo(
		"PlayerCollide",
		C.nox_xxx_collidePlayer_4E8460,
		func(player, other *server.Object, collision unsafe.Pointer) {
			playerCollideCall4E8460(player, other, collision)
		},
		0,
	)
	server.RegisterObjectCollideGo(
		"ProjectileCollide",
		C.nox_xxx_collideProjectileGeneric_4E87B0,
		func(projectile, other *server.Object, collision unsafe.Pointer) {
			projectileCollideCall4E87B0(projectile, other, collision)
		},
		unsafe.Sizeof(server.ProjectileCollideData{}),
	)
	server.RegisterObjectCollideGo(
		"ProjectileSparkCollide",
		C.nox_xxx_collideProjectileSpark_4E8880,
		func(projectile, other *server.Object, collision unsafe.Pointer) {
			projectileSparkCollideCall4E8880(projectile, other, collision)
		},
		unsafe.Sizeof(server.ProjectileCollideData{}),
	)
	server.RegisterObjectCollideGo(
		"DoorCollide",
		C.nox_xxx_collideDoor_4E8AC0,
		func(door, unit *server.Object, collision unsafe.Pointer) {
			doorCollideCall4E8AC0(door, unit, collision)
		},
		0,
	)
	server.RegisterObjectCollideGo(
		"PickupCollide",
		C.nox_xxx_collidePickup_4E8DF0,
		func(item, unit *server.Object, collision unsafe.Pointer) {
			pickupCollideCall4E8DF0(item, unit, collision)
		},
		0,
	)
	server.RegisterObjectCollide("ExitCollide", C.nox_xxx_collideExit_4E9090, unsafe.Sizeof(server.ExitCollideData{}))
	server.RegisterObjectCollideGo(
		"DamageCollide",
		C.nox_xxx_collideDamage_4E9430,
		func(source, target *server.Object, collision unsafe.Pointer) {
			damageCollideCall4E9430(source, target, collision)
		},
		unsafe.Sizeof(server.DamageCollideData{}),
	)
	server.RegisterObjectCollideGo(
		"ManaDrainCollide",
		C.nox_xxx_collideManadrain_4E9490,
		func(source, target *server.Object, collision unsafe.Pointer) {
			manaDrainCollideCall4E9490(source, target, collision)
		},
		unsafe.Sizeof(server.ManaDrainCollideData{}),
	)
	server.RegisterObjectCollideGo(
		"BombCollide",
		C.nox_xxx_collideBomb_4E96F0,
		func(source, target *server.Object, collision unsafe.Pointer) {
			bombCollideCall4E96F0(source, target, collision)
		},
		unsafe.Sizeof(server.BombCollideData{}),
	)
	server.RegisterObjectCollideGo(
		"SparkExplosionCollide",
		C.nox_xxx_fireballCollide_4E9AC0,
		func(source, target *server.Object, collision unsafe.Pointer) {
			sparkExplosionCollideCall4E9AC0(source, target, collision)
		},
		unsafe.Sizeof(server.SparkExplosionCollideData{}),
	)
	server.RegisterObjectCollideGo(
		"ChestCollide",
		C.nox_xxx_collideChest_4E9C40,
		func(source, target *server.Object, collision unsafe.Pointer) {
			chestCollideCall4E9C40(source, target, collision)
		},
		0,
	)
	server.RegisterObjectCollideGo(
		"WallReflectCollide",
		C.nox_xxx_collideSulphurShot2_4E9D80,
		func(source, target *server.Object, collision unsafe.Pointer) {
			wallReflectCollideCall4E9D80(source, target, collision)
		},
		unsafe.Sizeof(server.ProjectileCollideData{}),
	)
	server.RegisterObjectCollideGo(
		"WallReflectSparkCollide",
		C.nox_xxx_collideWallReflectSpark_4EA200,
		func(source, target *server.Object, collision unsafe.Pointer) {
			wallReflectSparkCollideCall4EA200(source, target, collision)
		},
		unsafe.Sizeof(server.ProjectileCollideData{}),
	)
	server.RegisterObjectCollideGo(
		"PixieCollide",
		C.nox_xxx_collidePixie_4EA080,
		func(source, target *server.Object, collision unsafe.Pointer) {
			pixieCollideCall4EA080(source, target, collision)
		},
		unsafe.Sizeof(server.ProjectileCollideData{}),
	)
	server.RegisterObjectCollide("OwnCollide", C.sub_4EA2C0, 0)
	server.RegisterObjectCollide("SparkCollide", C.nox_xxx_collideSpark_4EA300, 8)
	server.RegisterObjectCollide("BarrelCollide", C.sub_4EAAA0, 0)
	server.RegisterObjectCollide(
		"AudioEventCollide",
		C.sub_4EAAD0,
		unsafe.Sizeof(server.AudioEventCollideData{}),
	)
	server.RegisterObjectCollide("TriggerCollide", C.nox_xxx_collideTrigger_54FCD0, 0)
	server.RegisterObjectCollide(
		"TeleportCollide",
		C.sub_4EACA0,
		unsafe.Sizeof(server.TeleportCollideData{}),
	)
	server.RegisterObjectCollideGo("ElevatorCollide", C.nox_xxx_collideDefault_4E87A0, collideNoop, 8)
	server.RegisterObjectCollideGo(
		"AwardSpellCollide",
		C.nox_xxx_collideSpellPedestal_4EAD20,
		func(source, target *server.Object, collision unsafe.Pointer) {
			awardSpellCollideCall4EAD20(source, target, collision)
		},
		unsafe.Sizeof(server.AwardSpellCollideData{}),
	)
	server.RegisterObjectCollide("DieCollide", C.nox_xxx_collideDie_4E99B0, 0)
	server.RegisterObjectCollideGo(
		"GlyphCollide",
		C.nox_xxx_collideGlyph_4E9A00,
		func(source, target *server.Object, collision unsafe.Pointer) {
			glyphCollideCall4E9A00(source, target, collision)
		},
		0,
	)
	server.RegisterObjectCollideGo(
		"SpellProjectileCollide",
		C.nox_xxx_spellFlyCollide_4E9500,
		func(source, target *server.Object, collision unsafe.Pointer) {
			spellProjectileCollideCall4E9500(source, target, collision)
		},
		0,
	)
	server.RegisterObjectCollideGo(
		"BoomCollide",
		C.nox_xxx_collideBoom_4E9770,
		func(source, target *server.Object, collision unsafe.Pointer) {
			boomCollideCall4E9770(source, target, collision)
		},
		0,
	)
	server.RegisterObjectCollide("SignCollide", C.nox_xxx_collideSign_4EAB40, 0)
	server.RegisterObjectCollide("PentagramCollide", C.nox_xxx_collidePentagram_4EAB20, 0)
	server.RegisterObjectCollideGo(
		"SpiderSpitCollide",
		C.nox_xxx_collideWebbing_4EA380,
		func(source, target *server.Object, collision unsafe.Pointer) {
			webbingCollideCall4EA380(source, target, collision)
		},
		0,
	)
	server.RegisterObjectCollideGo(
		"DeathBallCollide",
		C.nox_xxx_collideDeathBall_4E9E90,
		func(source, target *server.Object, collision unsafe.Pointer) {
			deathBallCollideCall4E9E90(source, target, collision)
		},
		0,
	)
	server.RegisterObjectCollideGo(
		"DeathBallFragmentCollide",
		C.nox_xxx_collideDeathBallFragment_4E9FE0,
		func(source, target *server.Object, collision unsafe.Pointer) {
			deathBallFragmentCollideCall4E9FE0(source, target, collision)
		},
		0,
	)
	server.RegisterObjectCollideGo("TelekinesisCollide", C.nox_xxx_collideTelekinesis_4EADE0, collideNoop, 0)
	server.RegisterObjectCollideGo(
		"FistCollide",
		C.nox_xxx_collideFist_4EADF0,
		func(source, target *server.Object, collision unsafe.Pointer) {
			fistCollideCall4EADF0(source, target, collision)
		},
		0,
	)
	server.RegisterObjectCollideGo(
		"TeleportWakeCollide",
		C.nox_xxx_collideTeleportWake_4EAE30,
		func(source, target *server.Object, collision unsafe.Pointer) {
			teleportWakeCollideCall4EAE30(source, target, collision)
		},
		unsafe.Sizeof(server.TeleportWakeCollideData{}),
	)
	server.RegisterObjectCollide("FlagCollide", C.sub_4EA400, 0)
	server.RegisterObjectCollideGo(
		"ChakramInMotionCollide",
		C.nox_xxx_collideChakram_4EAF00,
		func(source, target *server.Object, collision unsafe.Pointer) {
			chakramCollideCall4EAF00(source, target, collision)
		},
		0,
	)
	server.RegisterObjectCollideGo(
		"ArrowCollide",
		C.nox_xxx_collideArrow_4EB490,
		func(source, target *server.Object, collision unsafe.Pointer) {
			arrowCollideCall4EB490(source, target, collision)
		},
		unsafe.Sizeof(server.ArrowCollideData{}),
	)
	server.RegisterObjectCollideGo(
		"MonsterArrowCollide",
		C.nox_xxx_collideMonsterArrow_4EB800,
		func(source, target *server.Object, collision unsafe.Pointer) {
			monsterArrowCollideCall4EB800(source, target, collision)
		},
		unsafe.Sizeof(server.MonsterArrowCollideData{}),
	)
	server.RegisterObjectCollideGo(
		"BearTrapCollide",
		C.nox_xxx_collideBearTrap_4EB890,
		func(source, target *server.Object, collision unsafe.Pointer) {
			bearTrapCollideCall4EB890(source, target, collision)
		},
		0,
	)
	server.RegisterObjectCollideGo(
		"PoisonGasTrapCollide",
		C.nox_xxx_collidePoisonGasTrap_4EB910,
		func(source, target *server.Object, collision unsafe.Pointer) {
			poisonGasTrapCollideCall4EB910(source, target, collision)
		},
		0,
	)
	server.RegisterObjectCollideGo(
		"TrapDoorCollide",
		C.nox_xxx_collideTrapDoor_4EAB60,
		func(source, target *server.Object, collision unsafe.Pointer) {
			trapDoorCollideCall4EAB60(source, target, collision)
		},
		unsafe.Sizeof(server.TrapDoorCollideData{}),
	)
	server.RegisterObjectCollideGo(
		"BallCollide",
		C.nox_xxx_collideBall_4EBA00,
		func(source, target *server.Object, collision unsafe.Pointer) {
			ballCollideCall4EBA00(source, target, collision)
		},
		0,
	)
	server.RegisterObjectCollideGo(
		"HomeBaseCollide",
		C.nox_xxx_collideHomeBase_4EBB80,
		func(source, target *server.Object, collision unsafe.Pointer) {
			homeBaseCollideCall4EBB80(source, target, collision)
		},
		0,
	)
	server.RegisterObjectCollide("CrownCollide", C.sub_4EBB50, 0)
	server.RegisterObjectCollideGo(
		"UndeadKillerCollide",
		C.nox_xxx_collideUndeadKiller_4EBD40,
		func(source, target *server.Object, collision unsafe.Pointer) {
			undeadKillerCollideCall4EBD40(source, target, collision)
		},
		unsafe.Sizeof(server.UndeadKillerCollideData{}),
	)
	server.RegisterObjectCollideGo(
		"YellowStarShotCollide",
		C.nox_xxx_collideSulphurShot_4E9E50,
		func(source, target *server.Object, collision unsafe.Pointer) {
			yellowStarShotCollideCall4E9E50(source, target, collision)
		},
		unsafe.Sizeof(server.ProjectileCollideData{}),
	)
	server.RegisterObjectCollideGo(
		"MimicCollide",
		C.nox_xxx_collideMimic_4E83D0,
		func(mimic, other *server.Object, collision unsafe.Pointer) {
			mimicCollideCall4E83D0(mimic, other, collision)
		},
		0,
	)
	server.RegisterObjectCollideGo(
		"HarpoonCollide",
		C.nox_xxx_collideHarpoon_4EB6A0,
		func(source, target *server.Object, collision unsafe.Pointer) {
			harpoonCollideCall4EB6A0(source, target, collision)
		},
		unsafe.Sizeof(server.HarpoonCollideData{}),
	)
	server.RegisterObjectCollide("MonsterGeneratorCollide", C.nox_xxx_collideMonsterGen_4EBE10, 0)
	server.RegisterObjectCollide("SoulGateCollide", C.sub_4EBE40, unsafe.Sizeof(server.SoulGateCollideData{}))
	server.RegisterObjectCollideGo(
		"AnkhCollide",
		C.nox_xxx_collideAnkhQuest_4EBF40,
		func(source, target *server.Object, collision unsafe.Pointer) {
			ankhCollideCall4EBF40(source, target, collision)
		},
		0,
	)

	server.RegisterObjectCollideParse("ProjectileCollide", wrapObjectCollideParseC(C.sub_536D80))
	server.RegisterObjectCollideParse("ProjectileSparkCollide", wrapObjectCollideParseC(C.sub_536D80))
	server.RegisterObjectCollideParse("DamageCollide", wrapObjectCollideParseC(C.nox_xxx_collideDamageLoad_536E10))
	server.RegisterObjectCollideParse("ManaDrainCollide", wrapObjectCollideParseC(C.nox_xxx_collideManaDrainLoad_536E50))
	server.RegisterObjectCollideParse(
		"SparkExplosionCollide",
		wrapObjectCollideParseC(C.nox_xxx_collideSparkExplosionLoad_536DE0),
	)
	server.RegisterObjectCollideParse("WallReflectCollide", wrapObjectCollideParseC(C.sub_536D80))
	server.RegisterObjectCollideParse("WallReflectSparkCollide", wrapObjectCollideParseC(C.sub_536D80))
	server.RegisterObjectCollideParse("PixieCollide", wrapObjectCollideParseC(C.sub_536D80))
	server.RegisterObjectCollideParse("AudioEventCollide", wrapObjectCollideParseC(C.sub_536DA0))
	server.RegisterObjectCollideParse("MonsterArrowCollide", wrapObjectCollideParseC(C.sub_536E80))
	server.RegisterObjectCollideParse("YellowStarShotCollide", wrapObjectCollideParseC(C.sub_536D80))
}

func collideNoop(*server.Object, *server.Object, unsafe.Pointer) {}

func wrapObjectCollideParseC(ptr unsafe.Pointer) server.ObjectParseFunc {
	return func(objt *server.ObjectType, args []string) error {
		if Nox_call_objectType_parseCollide_go(ptr, strings.Join(args, " "), objt.CollideData) == 0 {
			return fmt.Errorf("cannot parse collide data for %q", objt.ID())
		}
		return nil
	}
}
func Nox_call_objectType_parseCollide_go(a1 unsafe.Pointer, a2 string, a3 unsafe.Pointer) int {
	cstr := CString(a2)
	defer StrFree(cstr)
	return int(C.nox_call_objectType_parseCollide_go((*[0]byte)(a1), cstr, a3))
}

//export nox_xxx_castCounterSpell_52BBB0
func nox_xxx_castCounterSpell_52BBB0(a1 int32, a2, a3, a4 *nox_object_t) {
	Nox_xxx_castCounterSpell_52BBB0(spell.ID(a1), asObjectS(a2), asObjectS(a3), asObjectS(a4), nil, 0)
}

//export nox_xxx_changeOwner_52BE40
func nox_xxx_changeOwner_52BE40(a1, a2 *nox_object_t) {
	Nox_xxx_changeOwner_52BE40(asObjectS(a1), asObjectS(a2))
}
