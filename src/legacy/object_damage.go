package legacy

/*
#include "GAME3_2.h"
#include "GAME3_3.h"
#include "GAME4_3.h"
int nox_xxx_soundPlayerDamageSound_5328B0(nox_object_t* a1, nox_object_t* a2);
int nox_xxx_damageDefaultProc_4E0B30_go(nox_object_t* target, nox_object_t* source,
	nox_object_t* weapon, int damage, int damage_type);
*/
import "C"
import (
	"log/slog"

	"github.com/opennox/libs/object"

	noxflags "github.com/opennox/opennox/v1/common/flags"
	"github.com/opennox/opennox/v1/common/sound"
	"github.com/opennox/opennox/v1/server"
)

func init() {
	server.RegisterObjectDamage("DefaultDamage", C.nox_xxx_damageDefaultProc_4E0B30_go)
	server.RegisterObjectDamage("SkeletonDamage", C.sub_4E23C0)
	server.RegisterObjectDamage("PlayerDamage", C.nox_server_handler_PlayerDamage_4E17B0)
	server.RegisterObjectDamage("StoneDamage", C.sub_4E24B0)
	server.RegisterObjectDamage("MechGolemDamage", C.sub_4E24E0)
	server.RegisterObjectDamage("FlammableDamage", C.nox_xxx_damageFlammable_4E2520)
	server.RegisterObjectDamage("BlackPowderDamage", C.nox_xxx_damageBlackPowder_4E2560)
	server.RegisterObjectDamage("ArmorDamage", C.nox_xxx_damageArmor_4E1500)
	server.RegisterObjectDamage("WeaponDamage", C.sub_4E14B0)
	server.RegisterObjectDamage("BallDamage", C.sub_4E14A0)
	server.RegisterObjectDamage("MonsterGeneratorDamage", C.nox_xxx_damageMonsterGen_4E27D0)

	server.RegisterObjectDamageSound("DefaultDamageSound", C.nox_xxx_soundDefaultDamageSound_532E20)
	server.RegisterObjectDamageSound("PlayerDamageSound", C.nox_xxx_soundPlayerDamageSound_5328B0)
}

//export nox_xxx_damageDefaultProc_4E0B30_go
func nox_xxx_damageDefaultProc_4E0B30_go(
	targetp, sourcep, weaponp *nox_object_t,
	damage C.int,
	damageType C.int,
) C.int {
	target := asObjectS(targetp)
	source := asObjectS(sourcep)
	weapon := asObjectS(weaponp)
	s := GetServer().S()
	runtime := server.DefaultDamageWorldRuntime4E0B30{
		Frame:         s.Frame,
		GameplayFlag1: func() bool { return noxflags.HasGamePlay(noxflags.GameplayFlag1) },
		QuestMode:     func() bool { return noxflags.HasGame(noxflags.GameModeQuest) },
		IsZombie:      s.IsZombie,
		IsEnemy:       s.IsEnemyTo,
		Audio: func(id int, obj *server.Object) {
			s.Audio.EventObj(sound.ID(id), obj, 0, 0)
		},
		BuffOff:             Nox_xxx_spellBuffOff_4FF5B0,
		DefaultDamageSound:  func(target, source *server.Object) { server.Nox_xxx_soundDefaultDamageSound_532E20(target, source) },
		DamageClear:         unitDamageClearCall4EE5E0,
		DefaultDamageSoundC: C.nox_xxx_soundDefaultDamageSound_532E20,
		Unsupported: func(reason string, target, source, weapon *server.Object, damage int32, typ object.DamageType) {
			if s.Log != nil {
				s.Log.Error("DefaultDamage native branch is not ported",
					slog.String("reason", reason),
					slog.Uint64("target_type", uint64(target.TypeInd)),
					slog.Uint64("target_class", uint64(target.ObjClass)),
					slog.Int64("damage", int64(damage)),
					slog.Int64("damage_type", int64(typ)),
					slog.Uint64("source_ptr", uint64(uintptr(source.CObj()))),
					slog.Uint64("weapon_ptr", uint64(uintptr(weapon.CObj()))),
				)
			}
		},
	}
	return C.int(bool2int(server.DefaultDamageWorld4E0B30(
		target, source, weapon, int32(damage), object.DamageType(damageType), runtime,
	)))
}
