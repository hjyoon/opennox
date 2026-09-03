package legacy

/*
#include "GAME3_2.h"
#include "GAME3_3.h"
#include "GAME4_3.h"
#include "GAME5_2.h"
int nox_xxx_soundPlayerDamageSound_5328B0(nox_object_t* a1, nox_object_t* a2);
double nox_xxx_itemApplyDefendEffect_415C00(nox_object_t* item);
int nox_server_handler_PlayerDamage_4E17B0_go(nox_object_t* target, nox_object_t* source,
	nox_object_t* weapon, int damage, int damage_type);
int nox_xxx_damageDefaultProc_4E0B30_go(nox_object_t* target, nox_object_t* source,
	nox_object_t* weapon, int damage, int damage_type);
int sub_4E14A0_go(nox_object_t* target, nox_object_t* source,
	nox_object_t* weapon, int damage, int damage_type);
int sub_4E14B0_go(nox_object_t* target, nox_object_t* source,
	nox_object_t* weapon, int damage, int damage_type);
int nox_xxx_damageArmor_4E1500_go(nox_object_t* target, nox_object_t* source,
	nox_object_t* weapon, int damage, int damage_type);
*/
import "C"
import (
	"log/slog"
	"unsafe"

	"github.com/opennox/libs/object"
	"github.com/opennox/libs/types"

	noxflags "github.com/opennox/opennox/v1/common/flags"
	"github.com/opennox/opennox/v1/common/sound"
	"github.com/opennox/opennox/v1/server"
)

func defaultDamageCallNative4E0B30(
	target, source, weapon *server.Object,
	damage int32,
	typ object.DamageType,
) bool {
	return nox_xxx_damageDefaultProc_4E0B30_go(
		asObjectC(target), asObjectC(source), asObjectC(weapon), C.int(damage), C.int(typ),
	) != 0
}

func init() {
	server.RegisterObjectDamageGo("DefaultDamage", C.nox_xxx_damageDefaultProc_4E0B30_go, defaultDamageCallNative4E0B30)
	server.RegisterObjectDamageGo("SkeletonDamage", C.sub_4E23C0, func(target, source, weapon *server.Object, damage int32, typ object.DamageType) bool {
		s := GetServer().S()
		return server.SkeletonDamage4E23C0(target, source, weapon, damage, typ, server.SkeletonDamageRuntime4E23C0{
			Frame: s.Frame,
			Audio: func(id int, obj *server.Object) {
				s.Audio.EventObj(sound.ID(id), obj, 0, 0)
			},
			Direction: func(pos types.Pointf, direction int16, attackPos types.Pointf) int32 {
				return int32(Nox_server_testTwoPointsAndDirection_4E6E50(pos, direction, attackPos))
			},
			Default: defaultDamageCallNative4E0B30,
		})
	})
	server.RegisterObjectDamageGo("PlayerDamage", C.nox_server_handler_PlayerDamage_4E17B0_go, func(target, source, weapon *server.Object, damage int32, typ object.DamageType) bool {
		return nox_server_handler_PlayerDamage_4E17B0_go(
			asObjectC(target), asObjectC(source), asObjectC(weapon), C.int(damage), C.int(typ),
		) != 0
	})
	server.RegisterObjectDamageGo("StoneDamage", C.sub_4E24B0, func(target, source, weapon *server.Object, damage int32, typ object.DamageType) bool {
		return server.StoneDamage4E24B0(target, source, weapon, damage, typ, defaultDamageCallNative4E0B30)
	})
	server.RegisterObjectDamageGo("MechGolemDamage", C.sub_4E24E0, func(target, source, weapon *server.Object, damage int32, typ object.DamageType) bool {
		return server.MechGolemDamage4E24E0(target, source, weapon, damage, typ, defaultDamageCallNative4E0B30)
	})
	server.RegisterObjectDamageGo("FlammableDamage", C.nox_xxx_damageFlammable_4E2520, func(target, source, weapon *server.Object, damage int32, typ object.DamageType) bool {
		return server.FlammableDamage4E2520(target, source, weapon, damage, typ, defaultDamageCallNative4E0B30)
	})
	server.RegisterObjectDamageGo("BlackPowderDamage", C.nox_xxx_damageBlackPowder_4E2560, func(target, source, weapon *server.Object, damage int32, typ object.DamageType) bool {
		return server.BlackPowderDamage4E2560(target, source, weapon, damage, typ, defaultDamageCallNative4E0B30)
	})
	server.RegisterObjectDamageGo("ArmorDamage", C.nox_xxx_damageArmor_4E1500_go, func(target, source, weapon *server.Object, damage int32, typ object.DamageType) bool {
		return nox_xxx_damageArmor_4E1500_go(
			asObjectC(target), asObjectC(source), asObjectC(weapon), C.int(damage), C.int(typ),
		) != 0
	})
	server.RegisterObjectDamageGo("WeaponDamage", C.sub_4E14B0_go, func(target, source, weapon *server.Object, damage int32, typ object.DamageType) bool {
		return sub_4E14B0_go(
			asObjectC(target), asObjectC(source), asObjectC(weapon), C.int(damage), C.int(typ),
		) != 0
	})
	server.RegisterObjectDamageGo("BallDamage", C.sub_4E14A0_go, func(target, source, weapon *server.Object, damage int32, typ object.DamageType) bool {
		return sub_4E14A0_go(
			asObjectC(target), asObjectC(source), asObjectC(weapon), C.int(damage), C.int(typ),
		) != 0
	})
	server.RegisterObjectDamageGo("MonsterGeneratorDamage", C.nox_xxx_damageMonsterGen_4E27D0, func(target, source, weapon *server.Object, damage int32, typ object.DamageType) bool {
		legacyServer := GetServer()
		s := legacyServer.S()
		return server.MonsterGeneratorDamage4E27D0(target, source, weapon, damage, typ, server.MonsterGeneratorDamageRuntime4E27D0{
			Frame: s.Frame,
			PointFX: func(op, subtype byte, pos types.Pointf) {
				s.Nox_xxx_netSendPointFx2_523150(op, subtype, pos)
			},
			Audio: func(id int, object *server.Object) {
				s.Audio.EventObj(sound.ID(id), object, 0, 0)
			},
			Script: func(block *server.ScriptCallback, caller, trigger *server.Object, event server.ScriptEventType) {
				legacyServer.NoxScriptC().ScriptCallback(block, caller, trigger, event)
			},
			Default: defaultDamageCallNative4E0B30,
		})
	})

	server.RegisterObjectDamageSound("DefaultDamageSound", C.nox_xxx_soundDefaultDamageSound_532E20)
	server.RegisterObjectDamageSound("PlayerDamageSound", C.nox_xxx_soundPlayerDamageSound_5328B0)
}

//export nox_object_call_damage_native
func nox_object_call_damage_native(
	targetp, sourcep, weaponp *nox_object_t,
	damage C.int,
	damageType C.int,
) C.int {
	target := asObjectS(targetp)
	if target == nil {
		return 0
	}
	return C.int(bool2int(target.CallDamage(
		asObjectS(sourcep), asObjectS(weaponp), int(damage), object.DamageType(damageType),
	)))
}

func fireProtectionCall4DFE40(s *server.Server, target *server.Object) float64 {
	return s.FireProtection4DFE40(target, server.FireProtectionRuntime4DFE40{
		FireProtectEngage: C.sub_4DFD10,
	})
}

func defaultDamageWorldRuntime4E0B30(s *server.Server) server.DefaultDamageWorldRuntime4E0B30 {
	return server.DefaultDamageWorldRuntime4E0B30{
		Frame:         s.Frame,
		GameplayFlag1: func() bool { return noxflags.HasGamePlay(noxflags.GameplayFlag1) },
		QuestMode:     func() bool { return noxflags.HasGame(noxflags.GameModeQuest) },
		IsZombie:      s.IsZombie,
		IsEnemy:       s.IsEnemyTo,
		Audio: func(id int, obj *server.Object) {
			s.Audio.EventObj(sound.ID(id), obj, 0, 0)
		},
		BuffOff: Nox_xxx_spellBuffOff_4FF5B0,
		FireProtection: func(target *server.Object) float64 {
			return fireProtectionCall4DFE40(s, target)
		},
		MonsterHasHitSound: func(monster *server.Object) bool {
			if monster == nil || monster.UpdateData == nil || !monster.Class().Has(object.ClassMonster) {
				return false
			}
			soundSet := monster.UpdateDataMonster().SoundSet122
			if soundSet == nil {
				return false
			}
			id := sound.ID(*(*uint32)(unsafe.Add(soundSet, 8*4)))
			return id != 0 && s.Audio.Field12(id) > 0
		},
		DefaultDamageSound: func(target, source *server.Object) {
			server.Nox_xxx_soundDefaultDamageSound_532E20(target, source)
		},
		AdjustFieldGuide: func(source, target *server.Object, damage int32) int32 {
			if !noxflags.HasGame(noxflags.GameModeCoop | noxflags.GameModeQuest) {
				return damage
			}
			return s.DefaultDamageFieldGuide4E0B30(source, target, damage)
		},
		DamageClear:         unitDamageClearCall4EE5E0,
		DefaultDamageSoundC: C.nox_xxx_soundDefaultDamageSound_532E20,
		Unsupported: func(reason string, target, source, weapon *server.Object, damage int32, typ object.DamageType) {
			if s.Log == nil {
				return
			}
			objectFields := func(obj *server.Object) (uint64, uint64, uint64, uint64, string, uint64) {
				if obj == nil {
					return 0, 0, 0, 0, "", 0
				}
				name := ""
				if ot := s.Types.ByInd(int(obj.TypeInd)); ot != nil {
					name = ot.ID()
				}
				return uint64(obj.TypeInd), uint64(obj.ObjClass), uint64(obj.ObjSubClass), uint64(obj.ObjFlags), name, uint64(uintptr(obj.CObj()))
			}
			targetType, targetClass, targetSubClass, targetFlags, targetName, targetPtr := objectFields(target)
			sourceType, sourceClass, sourceSubClass, sourceFlags, sourceName, sourcePtr := objectFields(source)
			weaponType, weaponClass, weaponSubClass, weaponFlags, weaponName, weaponPtr := objectFields(weapon)
			s.Log.Error("DefaultDamage native branch is not ported",
				slog.String("reason", reason),
				slog.Uint64("target_ptr", targetPtr),
				slog.Uint64("target_type", targetType),
				slog.String("target_name", targetName),
				slog.Uint64("target_class", targetClass),
				slog.Uint64("target_subclass", targetSubClass),
				slog.Uint64("target_flags", targetFlags),
				slog.Int64("damage", int64(damage)),
				slog.Int64("damage_type", int64(typ)),
				slog.Uint64("source_ptr", sourcePtr),
				slog.Uint64("source_type", sourceType),
				slog.String("source_name", sourceName),
				slog.Uint64("source_class", sourceClass),
				slog.Uint64("source_subclass", sourceSubClass),
				slog.Uint64("source_flags", sourceFlags),
				slog.Uint64("weapon_ptr", weaponPtr),
				slog.Uint64("weapon_type", weaponType),
				slog.String("weapon_name", weaponName),
				slog.Uint64("weapon_class", weaponClass),
				slog.Uint64("weapon_subclass", weaponSubClass),
				slog.Uint64("weapon_flags", weaponFlags),
			)
		},
	}
}

func reportItemHealthNative4E1650(s *server.Server, owner, item *server.Object, before, after uint16) {
	if s == nil || owner == nil || item == nil || item.HealthData == nil ||
		!owner.ObjClass.Has(object.ClassPlayer) || owner.UpdateData == nil {
		return
	}
	player := owner.UpdateDataPlayer().Player
	if player == nil || s.NetSendPacketXxx == nil {
		return
	}
	if !noxflags.HasGame(noxflags.GameModeCoop) &&
		C.sub_57B190(C.ushort(before), C.ushort(item.HealthData.Max)) ==
			C.sub_57B190(C.ushort(after), C.ushort(item.HealthData.Max)) {
		return
	}
	packet := server.BuildShopItemHealthPacket4D87A0(item)
	s.NetSendPacketXxx1(player.Index(), packet[:], nil, 0)
}

//export nox_server_handler_PlayerDamage_4E17B0_go
func nox_server_handler_PlayerDamage_4E17B0_go(
	targetp, sourcep, weaponp *nox_object_t,
	damage C.int,
	damageType C.int,
) C.int {
	target := asObjectS(targetp)
	source := asObjectS(sourcep)
	weapon := asObjectS(weaponp)
	s := GetServer().S()
	handled, result := server.PlayerDamageNative4E17B0(target, source, weapon, int32(damage), object.DamageType(damageType), server.PlayerDamageRuntime4E17B0{
		Frame:            s.Frame,
		CoopMode:         func() bool { return noxflags.HasGame(noxflags.GameModeCoop) },
		QuestMode:        func() bool { return noxflags.HasGame(noxflags.GameModeQuest) },
		QuestDamageScale: func() float32 { return float32(C.sub_4E40B0()) },
		GodMode:          func() bool { return noxflags.HasEngine(noxflags.EngineGodMode) },
		IsEnemy:          s.IsEnemyTo,
		Audio: func(id int, obj *server.Object) {
			s.Audio.EventObj(sound.ID(id), obj, 0, 0)
		},
		BuffOff: Nox_xxx_spellBuffOff_4FF5B0,
		ObserveClear: func(player *server.Object) {
			if Nox_xxx_playerObserveClear_4DDEF0 != nil {
				Nox_xxx_playerObserveClear_4DDEF0(player)
			}
		},
		ItemArmorValue: func(item *server.Object) float32 {
			return float32(C.nox_xxx_itemApplyDefendEffect_415C00(asObjectC(item)))
		},
		CanDamageArmor: func(item *server.Object) bool {
			return item != nil && item.Damage == C.nox_xxx_damageArmor_4E1500_go
		},
		DamageArmor: func(item, source, weapon *server.Object, damage int32, typ object.DamageType) bool {
			return server.DefaultDamageWorld4E0B30(
				item, source, weapon, damage, typ, defaultDamageWorldRuntime4E0B30(s),
			)
		},
		ReportArmorHealth: func(owner, item *server.Object, before, after uint16) {
			reportItemHealthNative4E1650(s, owner, item, before, after)
		},
		FireProtection: func(target *server.Object) float64 {
			return fireProtectionCall4DFE40(s, target)
		},
		PlayerDamageSound: func(target, source *server.Object) {
			C.nox_xxx_soundPlayerDamageSound_5328B0(asObjectC(target), asObjectC(source))
		},
		PlayerDamageSoundC: C.nox_xxx_soundPlayerDamageSound_5328B0,
		DamageClear:        unitDamageClearCall4EE5E0,
		Unsupported: func(reason string, target, source, weapon *server.Object, damage int32, typ object.DamageType) {
			if s.Log != nil {
				s.Log.Error("PlayerDamage native branch is not ported",
					slog.String("reason", reason),
					slog.Int64("damage", int64(damage)),
					slog.Int64("damage_type", int64(typ)),
					slog.Uint64("target_ptr", uint64(uintptr(target.CObj()))),
					slog.Uint64("source_ptr", uint64(uintptr(source.CObj()))),
					slog.Uint64("weapon_ptr", uint64(uintptr(weapon.CObj()))),
				)
			}
		},
	})
	if handled {
		return C.int(bool2int(result))
	}
	if unsafe.Sizeof(uintptr(0)) == 4 {
		return C.int(C.nox_server_handler_PlayerDamage_4E17B0(
			C.int(uintptr(unsafe.Pointer(targetp))),
			C.int(uintptr(unsafe.Pointer(sourcep))),
			C.int(uintptr(unsafe.Pointer(weaponp))),
			damage,
			damageType,
		))
	}
	return 0
}

//export sub_4E14A0_go
func sub_4E14A0_go(
	targetp, sourcep, weaponp *nox_object_t,
	damage C.int,
	damageType C.int,
) C.int {
	return C.int(bool2int(server.BallDamage4E14A0(
		asObjectS(targetp), asObjectS(sourcep), asObjectS(weaponp),
		int32(damage), object.DamageType(damageType),
	)))
}

//export sub_4E14B0_go
func sub_4E14B0_go(
	targetp, sourcep, weaponp *nox_object_t,
	damage C.int,
	damageType C.int,
) C.int {
	s := GetServer().S()
	return C.int(bool2int(server.WeaponDamage4E14B0(
		asObjectS(targetp), asObjectS(sourcep), asObjectS(weaponp),
		int32(damage), object.DamageType(damageType),
		func(target, source, weapon *server.Object, damage int32, typ object.DamageType) bool {
			return server.DefaultDamageWorld4E0B30(
				target, source, weapon, damage, typ, defaultDamageWorldRuntime4E0B30(s),
			)
		},
	)))
}

//export nox_xxx_damageArmor_4E1500_go
func nox_xxx_damageArmor_4E1500_go(
	targetp, sourcep, weaponp *nox_object_t,
	damage C.int,
	damageType C.int,
) C.int {
	s := GetServer().S()
	return C.int(bool2int(server.ArmorDamage4E1500(
		asObjectS(targetp), asObjectS(sourcep), asObjectS(weaponp),
		int32(damage), object.DamageType(damageType),
		func(target, source, weapon *server.Object, damage int32, typ object.DamageType) bool {
			return server.DefaultDamageWorld4E0B30(
				target, source, weapon, damage, typ, defaultDamageWorldRuntime4E0B30(s),
			)
		},
	)))
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
	runtime := defaultDamageWorldRuntime4E0B30(s)
	return C.int(bool2int(server.DefaultDamageWorld4E0B30(
		target, source, weapon, int32(damage), object.DamageType(damageType), runtime,
	)))
}
