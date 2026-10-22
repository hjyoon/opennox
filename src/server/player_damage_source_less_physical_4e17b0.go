package server

import (
	"math"

	"github.com/opennox/libs/object"
)

func playerDamageSourceLessPrefixShape4E17B0(source, weapon *Object, damage int32, typ object.DamageType) bool {
	if source != nil || weapon != nil {
		return false
	}
	switch typ {
	case object.DamagePoison, object.DamageLava:
		// A legacy partial runtime can still inline these positive timer/
		// collision tails using HP/protection services without DefaultDamage.
		return damage <= 0
	case object.DamageExplosion, object.DamagePlasma, object.DamageManaBomb, object.DamageZapRay:
		return false // These signed source-less prefixes are already native.
	default:
		return true
	}
}

// Source-less player hits skip 004E19CE's facing/weapon-defense prefix.
// This is the complete original type switch: full/half armor, raw armor
// wear, electric scaling, raw damage and the default case have distinct
// carry/wear/marker behavior. Complete production runtimes use this prefix;
// callers lacking DefaultDamage retain the earlier supported slices.
// Cached absorption/markers and live carry/inventory stay separate.
func playerDamageSourceLessPlayer4E17B0(
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
	if !target.Class().Has(object.ClassPlayer) || target.UpdateData == nil {
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
	if !livePlayer() {
		return reject("unsupported live source-less physical player record")
	}
	physical := typ == object.DamageBlade || typ == object.DamageCrush || typ == object.DamageImpale ||
		typ == object.DamageExplosion || typ == object.DamageBite || typ == object.DamageClaw || typ == object.DamageImpact
	electric := typ == object.DamageElectric || typ == object.DamageAirborneElectric
	wearArmor := physical || electric || typ == object.DamageFlame || typ == object.DamageLava
	if wearArmor && (!playerDamageArmorReady4E17B0(target, r) || !livePlayer()) {
		return reject("unsupported live source-less physical armor record")
	}
	effective, wear := damage, damage
	if physical || electric {
		var scaled float32
		if electric {
			if r.ElectricArmorScale == nil {
				return reject("missing source-less electric armor scale")
			}
			scaled = float32(float64(r.ElectricArmorScale(target)) * float64(damage))
		} else {
			factor := float64(absorption)
			if typ == object.DamageCrush {
				factor *= .5
			}
			scaled = float32((1 - factor) * float64(damage))
		}
		// 004E20F0 reloads the update AFTER electric scaling returns.
		if !livePlayer() {
			return reject("unsupported live source-less carry record")
		}
		carry := &target.UpdateDataPlayer().Field21
		accumulated := scaled + math.Float32frombits(*carry)
		effective = playerDamageRound4E17B0(accumulated)
		*carry = math.Float32bits(accumulated - float32(effective))
		if !electric {
			wear = damage - effective
		} // Electric cases wear armor with the original signed amount.
	}
	if wearArmor {
		playerDamageApplyArmor4E17B0(target, nil, nil, wear, typ, r)
	}
	if !livePlayer() {
		return reject("unsupported live source-less physical marker record")
	}
	// Case13 and values outside0..17 take the original default branch,
	// which forwards raw damage without assigning a type marker here.
	if typ >= object.DamageBlade && typ <= object.DamageAirborneElectric && typ != object.DamageDeathMagic && cached.Field76 == 0 {
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
