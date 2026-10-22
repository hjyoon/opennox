package server

import (
	"math"

	"github.com/opennox/libs/object"
)

func playerDamageSourceLessPhysicalShape4E17B0(source, weapon *Object, typ object.DamageType) bool {
	return source == nil && weapon == nil &&
		(typ == object.DamageBlade || typ == object.DamageCrush || typ == object.DamageBite)
}

// Source-less script hits skip 004E19CE's entire facing/weapon-defense
// prefix. Cases 0/8 use full armor at 004E1F84; case 2 uses half armor at
// 004E1EE8. Absorption and the marker base are cached before ObserveClear,
// while 004E20F0 and 004E2180 reload live carry and inventory independently.
func playerDamageSourceLessPhysical4E17B0(
	target *Object, damage int32, typ object.DamageType, r PlayerDamageRuntime4E17B0,
) (handled, result bool) {
	reject := func(reason string) (bool, bool) {
		return playerDamageUnsupported4E17B0(r, reason, target, nil, nil, damage, typ)
	}
	if target.Flags().HasAny(object.FlagNoUpdate | object.FlagDead) {
		return true, false
	}
	if target.HasEnchant(playerDamageInvulnerableEnchant4E17B0) {
		if r.Frame == nil {
			return reject("missing source-less physical invulnerability frame")
		}
		if byte(r.Frame())&3 == 0 {
			if r.Audio == nil {
				return reject("missing source-less physical invulnerability audio")
			}
			r.Audio(playerDamageInvulnerableSound4E17B0, target)
		}
		return true, true
	}
	if r.CoopMode != nil {
		// 004E1808 still queries Coop with a nil source. Its nil terminal
		// owner cannot be this player, so it never suppresses script damage.
		_ = r.CoopMode()
	}
	if target.UpdateData == nil {
		return reject("missing source-less physical player update")
	}
	cached := target.UpdateDataPlayer()
	if cached.Player == nil {
		return reject("missing source-less physical player info")
	}
	absorption := math.Float32frombits(cached.Field57)
	if cached.Player.Field3680&1 != 0 {
		return true, false
	}
	if r.DefaultDamage == nil {
		return reject("missing source-less physical default damage")
	}
	cached.Field76 = 0
	if cached.Player.ObserveTarget() != nil {
		if r.ObserveClear == nil {
			return reject("missing source-less physical observe service")
		}
		r.ObserveClear(target)
	}
	livePlayer := func() bool {
		return target.Class().Has(object.ClassPlayer) && target.UpdateData != nil
	}
	if !livePlayer() || !playerDamageArmorReady4E17B0(target, r) || !livePlayer() {
		return reject("unsupported live source-less physical armor record")
	}
	factor := float64(absorption)
	if typ == object.DamageCrush {
		factor *= .5
	}
	carry := &target.UpdateDataPlayer().Field21
	scaled := float32((1 - factor) * float64(damage))
	accumulated := scaled + math.Float32frombits(*carry)
	effective := playerDamageRound4E17B0(accumulated)
	*carry = math.Float32bits(accumulated - float32(effective))
	playerDamageApplyArmor4E17B0(target, nil, nil, damage-effective, typ, r)
	if !livePlayer() {
		return reject("unsupported live source-less physical marker record")
	}
	if cached.Field76 == 0 {
		cached.Field76, cached.Field75 = 2, uint32(typ)
	}
	if damage > 0 && effective == 0 {
		effective = 1
	}
	if r.GodMode != nil && r.GodMode() && target.Class().Has(object.ClassPlayer) {
		return true, true
	}
	if r.QuestMode != nil && r.QuestMode() {
		if r.QuestDamageScale == nil {
			return reject("missing source-less physical Quest scale")
		}
		before := effective
		effective = playerDamageRound4E17B0(float32(float64(r.QuestDamageScale()) * float64(effective)))
		if before > 0 && effective < 1 {
			effective = 1
		}
	}
	return true, r.DefaultDamage(target, nil, nil, effective, typ)
}
