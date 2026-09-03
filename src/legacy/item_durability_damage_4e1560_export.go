package legacy

/*
#include "GAME3_2.h"

void nullsub_40(void);
void nullsub_41(void);
void nullsub_42(void);
*/
import "C"

import (
	"log/slog"
	"unsafe"

	"github.com/opennox/libs/object"

	"github.com/opennox/opennox/v1/legacy/common/ccall"
	"github.com/opennox/opennox/v1/server"
)

func itemDurabilityApplyDefendNative4E1560(
	effect *server.ModifierEff,
	item, owner, effective, source *server.Object,
	value *float32,
) bool {
	if effect == nil || effect.Defend76.Fnc == nil || value == nil {
		return true
	}
	fnc := effect.Defend76.Fnc
	if fnc == C.sub_4E0370 {
		*value *= effect.Defend76.Valf
		return true
	}
	if fnc == C.sub_4E0380 {
		*value *= 2.0 - effect.Defend76.Valf
		return true
	}
	if fnc == C.nullsub_40 || fnc == C.nullsub_41 || fnc == C.nullsub_42 {
		return true
	}
	if unsafe.Sizeof(uintptr(0)) != 4 {
		return false
	}
	ccall.CallVoidPtr6(
		fnc,
		effect.C(),
		item.CObj(),
		owner.CObj(),
		effective.CObj(),
		source.CObj(),
		unsafe.Pointer(value),
	)
	return true
}

func itemDurabilityRuntime4E1560(s *server.Server) server.ItemDurabilityDamageRuntime4E1560 {
	return server.ItemDurabilityDamageRuntime4E1560{
		ApplyDefend: itemDurabilityApplyDefendNative4E1560,
		Damage: func(item, source, effective *server.Object, damage int32, typ object.DamageType) bool {
			return item.CallDamage(source, effective, int(damage), typ)
		},
		ReportHealth: func(owner, item *server.Object, before, after uint16) {
			reportItemHealthNative4E1650(s, owner, item, before, after)
		},
		Unsupported: func(
			reason string,
			item, owner, source, effective *server.Object,
			amount float32,
			typ object.DamageType,
		) {
			if s == nil || s.Log == nil {
				return
			}
			ptr := func(obj *server.Object) uint64 {
				return uint64(uintptr(unsafe.Pointer(obj)))
			}
			s.Log.Error("item durability native branch is not ported",
				slog.String("reason", reason),
				slog.Uint64("item_ptr", ptr(item)),
				slog.Uint64("owner_ptr", ptr(owner)),
				slog.Uint64("source_ptr", ptr(source)),
				slog.Uint64("effective_ptr", ptr(effective)),
				slog.Float64("amount", float64(amount)),
				slog.Int64("damage_type", int64(typ)),
			)
		},
	}
}

func playerDamageWeaponNative4E1560(
	item, owner, source, effective *server.Object,
	amount float32,
	typ object.DamageType,
) bool {
	s := GetServer().S()
	return server.PlayerDamageWeaponNative4E1560(
		item, owner, source, effective, amount, typ, itemDurabilityRuntime4E1560(s),
	)
}

func equipDamageNative4E16D0(
	item, owner, source, effective *server.Object,
	amount float32,
	typ object.DamageType,
) bool {
	s := GetServer().S()
	return server.EquipDamageNative4E16D0(
		item, owner, source, effective, amount, typ, itemDurabilityRuntime4E1560(s),
	)
}

//export nox_xxx_playerDamageWeapon_4E1560_go
func nox_xxx_playerDamageWeapon_4E1560_go(
	itemp, ownerp, sourcep, effectivep *nox_object_t,
	amount C.float,
	typ C.int,
) {
	playerDamageWeaponNative4E1560(
		asObjectS(itemp), asObjectS(ownerp), asObjectS(sourcep), asObjectS(effectivep),
		float32(amount), object.DamageType(typ),
	)
}

//export nox_xxx_equipDamage_4E16D0_go
func nox_xxx_equipDamage_4E16D0_go(
	itemp, ownerp, sourcep, effectivep *nox_object_t,
	amount C.float,
	typ C.int,
) {
	equipDamageNative4E16D0(
		asObjectS(itemp), asObjectS(ownerp), asObjectS(sourcep), asObjectS(effectivep),
		float32(amount), object.DamageType(typ),
	)
}

func playerDamageWeaponEntry4E1560(
	item, owner, source, effective *server.Object,
	amount float32,
	typ object.DamageType,
) {
	C.nox_xxx_playerDamageWeapon_4E1560(
		asObjectC(item), asObjectC(owner), asObjectC(source), asObjectC(effective),
		C.float(amount), C.int(typ),
	)
}
