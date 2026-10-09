package server

import (
	"image"
	"math"
	"unsafe"

	"github.com/opennox/libs/object"
	"github.com/opennox/libs/types"
)

const (
	defaultDamageInvulnerableEnchant4E0B30 = EnchantID(23)
	defaultDamageShockEnchant4E0B30        = EnchantID(22)
	defaultDamageShieldEnchant4E0B30       = EnchantID(26)
	defaultDamageInvisibleEnchant4E0B30    = EnchantID(0)
	defaultDamageInvulnerableSound4E0B30   = 71
	defaultDamageShockSound4E0B30          = 135
	defaultDamageShockBalance4E0B30        = "ShockDamage"
	defaultDamageShockBalanceIndex4E0B30   = 4
)

// DefaultDamageWorldRuntime4E0B30 isolates the services used by the
// native-width world-object slice of GAME.EXE 004E0B30. Unsupported reports
// a branch that must be ported before it can be executed safely on a 64-bit
// host; the caller must not fall back to the ABI32 body after that point.
type DefaultDamageWorldRuntime4E0B30 struct {
	Frame               func() uint32
	GameplayFlag1       func() bool
	QuestMode           func() bool
	IsZombie            func(*Object) bool
	IsEnemy             func(*Object, *Object) bool
	Audio               func(int, *Object)
	BuffOff             func(*Object, EnchantID)
	FireProtection      func(*Object) float64
	ElectricProtection  func(*Object) float64
	MonsterHasHitSound  func(*Object) bool
	DefaultDamageSound  func(*Object, *Object)
	PlayerDamageSound   func(*Object, *Object)
	GameBallType        uint16
	GameBallOnDamage    func(*Object, *Object, int32)
	AdjustFieldGuide    func(*Object, *Object, int32) int32
	BalanceFloatInd     func(string, int) float64
	CallDamage          func(*Object, *Object, *Object, int32, object.DamageType) bool
	PlayerSetState      func(*Object, PlayerState) bool
	AdjustHP            func(*Object, int32)
	VampirismFX         func(int, image.Point, image.Point, uint16)
	ShieldReduce        func(*Object, *int32, object.DamageType, *Object)
	CanApplyLateDefend  func(*ModifierEff) bool
	ApplyLateDefend     func(*ModifierEff, *Object, *Object, *Object, *Object, int32, object.DamageType) int32
	CanApplyPreDamage   func(*ModifierEff) bool
	ApplyPreDamage      func(*ModifierEff, *Object, *Object, *Object, *int32)
	DamageClear         func(*Object, int32)
	DefaultDamageSoundC unsafe.Pointer
	PlayerDamageSoundC  unsafe.Pointer
	Unsupported         func(reason string, target, source, weapon *Object, damage int32, typ object.DamageType)
}

func defaultDamageUnsupported4E0B30(
	runtime DefaultDamageWorldRuntime4E0B30,
	reason string,
	target, source, weapon *Object,
	damage int32,
	typ object.DamageType,
) bool {
	if runtime.Unsupported != nil {
		runtime.Unsupported(reason, target, source, weapon, damage, typ)
	}
	// The original callback reports success for nearly every early-exit path.
	// Returning success is safer than invoking its pointer-truncating ABI32
	// body, while Unsupported keeps the missing branch visible in test logs.
	return true
}

func defaultDamageWeaponPreDamageModifiers4E0B30(weapon *Object) [4]*ModifierEff {
	if weapon == nil || weapon.InitData == nil ||
		!weapon.Class().HasAny(object.ClassWeapon|object.ClassWand) {
		return [4]*ModifierEff{}
	}
	return (*ModifierInitData)(weapon.InitData).Modifiers
}

// defaultDamageAttackQualifies4E1400 restores GAME.EXE 004E1400 without
// reading Object or WandUseData through their PE32 offsets. The helper is used
// by DefaultDamage's friendly-hit and Shock predicates in the original. The
// Shock call site always supplies a non-nil weapon, but retaining the no-weapon
// branch here documents the complete predicate and makes later callers safe to
// migrate without reintroducing an ABI32 call.
func defaultDamageAttackQualifies4E1400(source, weapon *Object) bool {
	if weapon == nil {
		if source.Class().Has(object.ClassPlayer) {
			return true
		}
		return source.Class().Has(object.ClassMonster) && uint32(source.SubClass())&0x10 != 0
	}

	class := weapon.Class()
	subclass := uint32(weapon.SubClass())
	if class.Has(object.ClassWand) {
		if subclass&0x047f0000 == 0 {
			return true
		}
		// GAME.EXE reads the low byte of WandUseData.Flags at offset 96.
		// AsWand deliberately preserves the original nil-fault boundary.
		if weapon.UseData.AsWand().Flags&2 != 0 {
			return true
		}
	} else if class.Has(object.ClassWeapon) && subclass&0x047f00fe == 0 {
		return true
	}
	return uint8(class)&uint8(object.ClassMonster) != 0
}

// DefaultDamageFieldGuide4E0B30 restores sub_4FB000/sub_4FB050 for the
// cooperative and Quest call site in GAME.EXE 004E0B30. Guide zero is the
// original invalid sentinel, and the per-player array retains its fixed
// 41-entry semantic layout on native-width hosts.
func (s *Server) DefaultDamageFieldGuide4E0B30(source, target *Object, damage int32) int32 {
	if source == nil || target == nil || !source.Class().Has(object.ClassPlayer) ||
		!target.Class().Has(object.ClassMonster) {
		return damage
	}
	update := source.UpdateDataPlayer()
	if update == nil || update.Player == nil {
		return damage
	}
	typ := s.Types.ByInd(int(target.TypeInd))
	if typ == nil {
		return damage
	}
	guide := 0
	for i := 1; i < len(rewardFieldGuideNames4F0D20); i++ {
		if rewardFieldGuideNames4F0D20[i] == typ.ID() {
			guide = i
			break
		}
	}
	if guide == 0 || update.Player.BeastScrollLvl[guide] == 0 {
		return damage
	}
	value := float32(s.Balance.Float("FieldGuideDamageBonus")*float64(damage) + 0.5)
	return int32(value)
}

// DefaultDamageWorld4E0B30 restores the unmodified world-object damage branch,
// player melee and unarmed electric spells against ordinary monsters, monster
// and source-less scripted BLADE/electric damage against ordinary monsters,
// missile IMPACT and Magic Missile EXPLOSION against ordinary monsters, the
// monster-on-monster self-weapon BITE and ordinary melee-weapon BLADE branches,
// Berserker Charge's player-self-weapon CRUSH and PlayerDamage's scripted-NPC
// weapon CRUSH tail, monster-fired missile PIERCE against players/monsters,
// ordinary player/NPC weapon BLADE/CRUSH and unarmed CLAW/CRUSH tails,
// unit-sourced SIMPLE CRUSH (including all three stock Fists),
// weapon-less player/monster electric damage against players, and unit-self-weapon
// ELECTRIC/AIRBORNE_ELECTRIC tails used by Shock Glyphs, plus SIMPLE|IMMOBILE
// ZAP_RAY (including stock SentryGlobe) against monsters/NPCs
// and weapon-less unit/source-less EXPLOSION (including Meteor) against units,
// and weapon-less raw MANA_BOMB/ZAP_RAY (including Mana Bomb/Death Ray),
// plus SIMPLE|DANGEROUS world POISON (including ToxicCloud) against monsters,
// and non-weapon caster IMPACT (including Earthquake) against units/world objects
// from GAME.EXE 004E0B30
// without narrowing Object pointers.
// Player targets use their dedicated damage callback in normal data; other
// protection, modifier, and equipment branches remain visible through
// Unsupported instead of entering the unsafe raw body.
func DefaultDamageWorld4E0B30(
	target, source, weapon *Object,
	damage int32,
	typ object.DamageType,
	runtime DefaultDamageWorldRuntime4E0B30,
) bool {
	if target == nil {
		return true
	}
	// GAME.EXE reads 84EA04 at each executed use, never on entry. In
	// particular protection, BuffOff and Defend can precede a later read.
	currentFrame := func() uint32 {
		if runtime.Frame != nil {
			return runtime.Frame()
		}
		return 0
	}
	if target.HasEnchant(defaultDamageInvulnerableEnchant4E0B30) {
		if byte(currentFrame())&3 == 0 && runtime.Audio != nil {
			runtime.Audio(defaultDamageInvulnerableSound4E0B30, target)
		}
		return true
	}

	var monsterUpdate *MonsterUpdateData
	if target.Class().Has(object.ClassMonster) {
		monsterUpdate = target.UpdateDataMonster()
		// GAME.EXE clears this latch before checking whether the monster is
		// already dead or whether the incoming damage will be admitted.
		monsterUpdate.Field547 = 0
		if typ == object.DamageFlame || typ == object.DamageLava || typ == object.DamageExplosion ||
			typ == object.DamagePlasma || typ == object.DamageDispelUndead {
			monsterUpdate.StatusFlags |= object.MonStatusOnFire
		}
	}
	if target.ObjFlags.Has(object.FlagDead) {
		zombie := false
		if runtime.IsZombie != nil {
			zombie = runtime.IsZombie(target)
		}
		if !zombie {
			return true
		}
		if weapon != nil {
			target.Obj130 = weapon
		} else {
			target.Obj130 = source
		}
		target.Field131 = uint32(typ)
		// 004E0BEA reads after IsZombie and the attribution/type stores.
		target.Frame134 = currentFrame()
		return true
	}
	// Shock Glyph supplies the caster as BOTH source and weapon. Keep that
	// identity through late defense, attribution, sound and Shield reduction.
	// Weapon/wand/missile casters have separate modifier branches.
	unitSelfWeaponElectric := source != nil && source == weapon && source.UpdateData != nil &&
		source.Class().HasAny(object.ClassPlayer|object.ClassMonster) &&
		!source.Class().HasAny(object.ClassWeapon|object.ClassWand|object.ClassMissile) &&
		(typ == object.DamageElectric || typ == object.DamageAirborneElectric)
	playerElectric := target.Class().Has(object.ClassPlayer) &&
		(unitSelfWeaponElectric || (source != nil && source.Class().HasAny(object.ClassPlayer|object.ClassMonster) &&
			source.UpdateData != nil && weapon == nil &&
			(typ == object.DamageElectric || typ == object.DamageAirborneElectric)))
	// Player/monster-fired ranged missiles share the original damage tail.
	// PIERCE (type 3, DamageImpale in libs) skips both protection branches.
	// Stock GolemArrow is MISSILE|WEAPON, subclass
	// 0x10: 004E1400 rejects this ranged weapon, not every WEAPON-class
	// missile. Keep melee/unit/wand shapes outside this no-Shock slice.
	missilePierce := typ == object.DamageImpale && source != nil && source != weapon &&
		source.Class().HasAny(object.ClassPlayer|object.ClassMonster) && source.UpdateData != nil &&
		weapon != nil && weapon.Class().Has(object.ClassMissile) &&
		!weapon.Class().HasAny(object.MaskUnits|object.ClassWand) &&
		!defaultDamageAttackQualifies4E1400(source, weapon)
	ordinaryMelee := playerDamageMeleeShape4E17B0(source, weapon, typ)
	// Physical self-strikes and Ghost DRAIN share IMPACT's qualifying
	// monster weapon/hostility/Defend/sound/HP tail (004E0C55 onward).
	monsterImpact := playerDamageMonsterImpactShape4E17B0(source, weapon, typ) ||
		playerDamageMonsterSelfStrikeShape4E17B0(source, weapon, typ)
	// The terminal owner and Earthquake caster need not be the same object.
	// Keep the already restored monster self-strike on its existing path;
	// PLAYER and class-zero self-casters do not qualify the melee predicate.
	casterImpact := !monsterImpact && defaultDamageCasterImpactShape4E0B30(weapon, typ)
	simpleCrush := playerDamageSimpleCrushShape4E17B0(source, weapon, typ)
	missileFlame := playerDamageMissileFlameShape4E17B0(source, weapon, typ)
	worldProjectile := playerDamageWorldProjectileShape4E17B0(source, weapon, typ)
	// FIRE|SIMPLE|DANGEROUS colliders (Flame/FlameCleanse) also reach
	// case 1 with signed raw damage, without the missile/melee qualifier.
	worldFlame := playerDamageWorldFlameShape4E17B0(weapon, typ)
	spellMissileExplosion := playerDamageMissileExplosionShape4E17B0(source, weapon, typ)
	spellMissileImpact := playerDamageSpellMissileImpactShape4E17B0(source, weapon, typ)
	worldImpale := (monsterUpdate != nil || target.Class().Has(object.ClassPlayer)) && playerDamageWorldImpaleShape4E17B0(source, weapon, typ)
	// Stock SentryGlobe is SIMPLE|IMMOBILE, not an electric spell or melee
	// weapon. Its terminal owner can be the globe itself or a unit. Keep this
	// admission in DefaultDamage; PlayerDamage retains its separate prefix.
	zapRay := (monsterUpdate != nil || target.Class().Has(object.ClassPlayer)) && typ == object.DamageZapRay && weapon != nil &&
		weapon.Class().Has(object.ClassSimple) && weapon.Class().Has(object.ClassImmobile) &&
		!weapon.Class().HasAny(object.MaskUnits|object.ClassWeapon|object.ClassWand|object.ClassMissile) &&
		(source == weapon || (source != nil && source.UpdateData != nil &&
			source.Class().HasAny(object.ClassPlayer|object.ClassMonster) &&
			!source.Class().HasAny(object.ClassWeapon|object.ClassWand|object.ClassMissile)))
	// Meteor's radial callback supplies the terminal owner (or nil), not a
	// missile weapon. This player tail still performs fire protection, live
	// Defend/sound/hurt/Shield and HP; a nil weapon skips 004E0C61's second
	// enemy query and Shock, but not 004E0C03's campaign owner gate.
	worldExplosion := playerDamageWorldExplosionShape4E17B0(source, weapon, typ)
	playerWeaponlessExplosion := target.Class().Has(object.ClassPlayer) && weapon == nil && typ == object.DamageExplosion &&
		(worldExplosion || source == nil || (source.UpdateData != nil && source.Class().HasAny(object.MaskUnits) &&
			!source.Class().HasAny(object.ClassMissile|object.ClassWeapon|object.ClassWand)))
	// Death Ray (16) and Mana Bomb (15) enter the ordinary signed raw tail,
	// not fire/electric protection. Nil weapons skip the second melee enemy
	// query and Shock, but retain the campaign owner gate and late defense.
	weaponlessRawSpell := playerDamageWeaponlessRawSpellShape4E17B0(source, weapon, typ) &&
		(source == nil || source.Class() == 0 || source.UpdateData != nil)
	// 0053D8C0 supplies the terminal owner, including a class-zero
	// ImaginaryCaster or the cloud itself. Do not require a unit source or a
	// particular type ID. Stock ToxicCloud is OBSTACLE|LIGHT|DANGEROUS|
	// SIMPLE (0x190008), not MISSILE; unit/weapon/wand/missile shapes keep
	// separate admission. Light/obstacle are rendering/collision properties.
	monsterCloudPoison := monsterUpdate != nil && typ == object.DamagePoison && weapon != nil &&
		weapon.Class().Has(object.ClassSimple) && weapon.Class().Has(object.ClassDangerous) &&
		!weapon.Class().HasAny(object.MaskUnits|object.ClassWeapon|object.ClassWand|object.ClassMissile)
	playerTail := playerElectric || playerWeaponlessExplosion || ((weaponlessRawSpell || worldImpale || missilePierce || missileFlame || worldProjectile || worldFlame || spellMissileExplosion || spellMissileImpact || ordinaryMelee || monsterImpact || casterImpact || simpleCrush || zapRay) && target.Class().Has(object.ClassPlayer))
	if playerTail && !casterImpact {
		if target.UpdateData == nil || target.HealthData == nil {
			return defaultDamageUnsupported4E0B30(runtime, "player without update/health", target, source, weapon, damage, typ)
		}
	}
	if target.Class().HasAny(object.MaskUnits) && monsterUpdate == nil && !playerTail {
		return defaultDamageUnsupported4E0B30(runtime, "non-monster unit target", target, source, weapon, damage, typ)
	}

	gameplay := runtime.GameplayFlag1 != nil && runtime.GameplayFlag1()
	if !gameplay && source != nil {
		owner := source.FindOwnerChainPlayer()
		if owner != nil && owner.Class().HasAny(object.MaskUnits) {
			enemy := runtime.IsEnemy != nil && runtime.IsEnemy(target, owner)
			// 004E0C36 skips Quest after an enemy result; 004E0C3A also
			// skips it for a different friendly target. Preserve the earlier
			// slices' service contract while restoring those reads for casters.
			quest := false
			if !casterImpact || (!enemy && target == owner) {
				quest = runtime.QuestMode != nil && runtime.QuestMode()
			}
			if !enemy && (target != owner || quest) {
				return true
			}
		}
	}
	zapRayTarget := "monster"
	if playerTail {
		zapRayTarget = "player"
	}
	if worldProjectile && weapon != nil {
		if runtime.IsEnemy == nil {
			return defaultDamageUnsupported4E0B30(runtime, "missing world projectile enemy service", target, source, weapon, damage, typ)
		}
		// 004E0C61 queries before NoUpdate, including non-unit owners.
		// A false result only excludes a LIVE qualifying melee attack;
		// the ordinary arrow/fireball/DeathBall classes do not qualify.
		if !runtime.IsEnemy(target, source) && target.Class().HasAny(object.MaskUnits) &&
			defaultDamageAttackQualifies4E1400(source, weapon) && !defaultDamageFriendlyException4E1470(weapon) {
			return true
		}
	}
	if zapRay {
		if runtime.IsEnemy == nil {
			return defaultDamageUnsupported4E0B30(runtime, "missing "+zapRayTarget+" ZAP_RAY enemy service", target, source, weapon, damage, typ)
		}
		// 004E0C61 queries even for a non-melee ray before 004E0C94's
		// NoUpdate check. 004E1400 rejects its class, so a false result does
		// not impose a second friendly gate after the campaign owner check.
		_ = runtime.IsEnemy(target, source)
	}
	if monsterImpact && target.Class().HasAny(object.MaskUnits) {
		// A MONSTER self-weapon qualifies 004E1400, unlike missiles and
		// PLAYER charge weapons. 004E0C55 queries hostility before NoUpdate
		// and excludes friendly hits even when gameplay flag 1 is enabled.
		if runtime.IsEnemy == nil {
			return defaultDamageUnsupported4E0B30(runtime, "missing monster self-weapon IMPACT enemy service", target, source, weapon, damage, typ)
		}
		if !runtime.IsEnemy(target, source) {
			return true
		}
	}
	if casterImpact && source != nil {
		if runtime.IsEnemy == nil {
			return defaultDamageUnsupported4E0B30(runtime, "missing caster IMPACT enemy service", target, source, weapon, damage, typ)
		}
		// 004E0C63 runs even for a PLAYER/class-zero caster or a world
		// target, before NoUpdate. Only a live qualifying attack on a unit
		// adds the friendly-hit gate; an owned MONSTER caster does qualify.
		if !runtime.IsEnemy(target, source) && target.Class().HasAny(object.MaskUnits) &&
			defaultDamageAttackQualifies4E1400(source, weapon) && !defaultDamageFriendlyException4E1470(weapon) {
			return true
		}
	}
	if worldImpale {
		if runtime.IsEnemy == nil {
			return defaultDamageUnsupported4E0B30(runtime, "missing player world IMPALE enemy service", target, source, weapon, damage, typ)
		}
		// 004E0C61 queries world hazards before NoUpdate, then reloads the
		// qualifier. A non-weapon spike does not qualify 004E1400, so the
		// false enemy result alone cannot make it harmless.
		if !runtime.IsEnemy(target, source) && target.Class().HasAny(object.MaskUnits) &&
			defaultDamageAttackQualifies4E1400(source, weapon) && !defaultDamageFriendlyException4E1470(weapon) {
			return true
		}
	}
	if spellMissileImpact && playerTail {
		if runtime.IsEnemy == nil {
			return defaultDamageUnsupported4E0B30(runtime, "missing player spell-missile IMPACT enemy service", target, source, weapon, damage, typ)
		}
		// 004E0C61 queries even a pure/self-sourced Pixie before NoUpdate.
		// 004E1400 normally rejects the missile, not the hit. Reload the
		// live class if an enemy callback changes that predicate's inputs.
		if !runtime.IsEnemy(target, source) && target.Class().HasAny(object.MaskUnits) &&
			defaultDamageAttackQualifies4E1400(source, weapon) && !defaultDamageFriendlyException4E1470(weapon) {
			return true
		}
	}
	if monsterCloudPoison && source != nil {
		if runtime.IsEnemy == nil {
			return defaultDamageUnsupported4E0B30(runtime, "missing monster world POISON enemy service", target, source, weapon, damage, typ)
		}
		// 004E0C61 queries before NoUpdate even for a non-melee cloud.
		// A false result does not reject the cloud. Re-read the live class
		// for 004E1400/004E1470 if the callback changes its inputs.
		if !runtime.IsEnemy(target, source) && target.Class().HasAny(object.MaskUnits) &&
			defaultDamageAttackQualifies4E1400(source, weapon) && !defaultDamageFriendlyException4E1470(weapon) {
			return true
		}
	}
	if worldFlame && source != nil {
		if runtime.IsEnemy == nil {
			return defaultDamageUnsupported4E0B30(runtime, "missing world FLAME enemy service", target, source, weapon, damage, typ)
		}
		// 004E0C61 queries BEFORE NoUpdate, even for a friendly fire.
		// Reload the qualifier after that callback, just as 004E0C75 does.
		if !runtime.IsEnemy(target, source) && target.Class().HasAny(object.MaskUnits) &&
			defaultDamageAttackQualifies4E1400(source, weapon) && !defaultDamageFriendlyException4E1470(weapon) {
			return true
		}
	}
	if target.ObjFlags.Has(object.FlagNoUpdate) {
		return true
	}
	// 004E0C90 requires both source and weapon. In particular, an unarmed
	// hit has no melee friendly-fire gate, and 004E1470 exempts War Hammer.
	// Neither exception bypasses the earlier campaign owner check.
	if ordinaryMelee && weapon != nil && target.Class().HasAny(object.MaskUnits) &&
		!defaultDamageFriendlyException4E1470(weapon) &&
		(runtime.IsEnemy == nil || !runtime.IsEnemy(target, source)) {
		return true
	}
	// 004E0C55 first queries IsEnemy, then 004E1400. A PLAYER-class
	// self-weapon does not qualify; a MONSTER-class self-weapon does.
	if unitSelfWeaponElectric && target.Class().HasAny(object.MaskUnits) &&
		(runtime.IsEnemy == nil || !runtime.IsEnemy(target, source)) &&
		defaultDamageAttackQualifies4E1400(source, weapon) && !defaultDamageFriendlyException4E1470(weapon) {
		return true
	}
	if ordinaryMelee && source.Class().Has(object.ClassMonster) && runtime.MonsterHasHitSound == nil {
		return defaultDamageUnsupported4E0B30(runtime, "missing monster hit-sound lookup", target, source, weapon, damage, typ)
	}
	if ordinaryMelee && playerTail && runtime.PlayerSetState == nil {
		return defaultDamageUnsupported4E0B30(runtime, "missing player melee hurt-state service", target, source, weapon, damage, typ)
	}
	if monsterImpact && (runtime.MonsterHasHitSound == nil || runtime.BuffOff == nil || runtime.DamageClear == nil ||
		(playerTail && runtime.PlayerSetState == nil)) {
		return defaultDamageUnsupported4E0B30(runtime, "missing monster self-weapon IMPACT tail service", target, source, weapon, damage, typ)
	}
	if playerElectric && ((source.Class().Has(object.ClassMonster) && runtime.MonsterHasHitSound == nil) || runtime.PlayerSetState == nil) {
		return defaultDamageUnsupported4E0B30(runtime, "missing player electric tail service", target, source, weapon, damage, typ)
	}
	if zapRay && (runtime.BuffOff == nil || runtime.DamageClear == nil ||
		(source.Class().Has(object.ClassMonster) && runtime.MonsterHasHitSound == nil)) {
		return defaultDamageUnsupported4E0B30(runtime, "missing "+zapRayTarget+" ZAP_RAY tail service", target, source, weapon, damage, typ)
	}
	if missilePierce && ((source.Class().Has(object.ClassMonster) && runtime.MonsterHasHitSound == nil) || runtime.BuffOff == nil ||
		runtime.IsEnemy == nil || runtime.DamageClear == nil || (playerTail && runtime.PlayerSetState == nil)) {
		return defaultDamageUnsupported4E0B30(runtime, "missing missile PIERCE tail service", target, source, weapon, damage, typ)
	}
	if simpleCrush && (runtime.MonsterHasHitSound == nil || runtime.BuffOff == nil ||
		runtime.IsEnemy == nil || runtime.DamageClear == nil || (playerTail && runtime.PlayerSetState == nil)) {
		return defaultDamageUnsupported4E0B30(runtime, "missing SIMPLE CRUSH tail service", target, source, weapon, damage, typ)
	}
	if missileFlame && (runtime.MonsterHasHitSound == nil || runtime.BuffOff == nil ||
		runtime.IsEnemy == nil || runtime.DamageClear == nil || (playerTail && runtime.PlayerSetState == nil)) {
		return defaultDamageUnsupported4E0B30(runtime, "missing missile FLAME tail service", target, source, weapon, damage, typ)
	}
	// The already ported monster explosion tail permits optional sound/AI
	// services. Require the complete service set only for the new player tail.
	if spellMissileExplosion && playerTail && (runtime.MonsterHasHitSound == nil || runtime.BuffOff == nil ||
		runtime.IsEnemy == nil || runtime.DamageClear == nil || runtime.PlayerSetState == nil) {
		return defaultDamageUnsupported4E0B30(runtime, "missing missile EXPLOSION tail service", target, source, weapon, damage, typ)
	}
	if spellMissileImpact && playerTail && (runtime.BuffOff == nil || runtime.DamageClear == nil ||
		(source.Class().Has(object.ClassMonster) && runtime.MonsterHasHitSound == nil)) {
		return defaultDamageUnsupported4E0B30(runtime, "missing player spell-missile IMPACT tail service", target, source, weapon, damage, typ)
	}
	if worldImpale && (runtime.BuffOff == nil || runtime.DamageClear == nil ||
		(source.Class().Has(object.ClassMonster) && runtime.MonsterHasHitSound == nil)) {
		return defaultDamageUnsupported4E0B30(runtime, "missing player world IMPALE tail service", target, source, weapon, damage, typ)
	}
	if playerWeaponlessExplosion && (runtime.DamageClear == nil || (source != nil &&
		(runtime.BuffOff == nil || (source.Class().Has(object.ClassMonster) && runtime.MonsterHasHitSound == nil)))) {
		return defaultDamageUnsupported4E0B30(runtime, "missing weapon-less player EXPLOSION tail service", target, source, weapon, damage, typ)
	}
	if weaponlessRawSpell && target.Class().HasAny(object.MaskUnits) && (runtime.DamageClear == nil ||
		(playerTail && runtime.PlayerSetState == nil) || (source != nil && (runtime.BuffOff == nil ||
		(source.Class().Has(object.ClassMonster) && runtime.MonsterHasHitSound == nil)))) {
		return defaultDamageUnsupported4E0B30(runtime, "missing weapon-less raw spell tail service", target, source, weapon, damage, typ)
	}
	monsterElectric := monsterUpdate != nil && (unitSelfWeaponElectric ||
		(weapon == nil && (source == nil || source.Class().HasAny(object.ClassPlayer|object.ClassMonster)) &&
			(typ == object.DamageElectric || typ == object.DamageAirborneElectric)))
	// Meteor passes its terminal player owner (or nil) with no weapon;
	// BlackPowderBarrelUpdate passes the SIMPLE|LIGHT breaking object.
	// 004E0C55's melee gate requires both source and weapon; 004E0D55
	// still applies EXPLOSION fire immunity/protection and the ordinary tail.
	monsterWeaponlessExplosion := monsterUpdate != nil && weapon == nil && typ == object.DamageExplosion &&
		(worldExplosion || source == nil || source.Class().HasAny(object.ClassPlayer|object.ClassMonster))
	if monsterWeaponlessExplosion && (runtime.DamageClear == nil || (source != nil &&
		(runtime.BuffOff == nil || (source.Class().Has(object.ClassMonster) && runtime.MonsterHasHitSound == nil)))) {
		return defaultDamageUnsupported4E0B30(runtime, "missing weapon-less monster EXPLOSION tail service", target, source, weapon, damage, typ)
	}
	// Campaign scripts use source-less BLADE damage for set-piece kills. The
	// original enters its no-source branch and still reaches DamageClear.
	sourceLessMonsterBlade := monsterUpdate != nil && source == nil && weapon == nil && typ == object.DamageBlade
	// The poison timer damages units without a source or weapon. Stock type 5
	// bypasses elemental/Shield protection but still runs NPC late defense.
	sourceLessMonsterPoison := monsterUpdate != nil && source == nil && weapon == nil && typ == object.DamagePoison
	// World collision token 6 supplies neither attacker nor weapon. Lava
	// still enters the ordinary monster fire-protection and injured/HP tail.
	monsterWorldLava := monsterUpdate != nil && source == nil && weapon == nil && typ == object.DamageLava
	monsterWeaponCrush := monsterUpdate != nil && uint32(target.SubClass())&0x10 != 0 &&
		source != nil && source.Class().Has(object.ClassMonster) && source.UpdateData != nil &&
		weapon != nil && weapon.Class().Has(object.ClassWeapon) && typ == object.DamageCrush
	// PlayerCollide supplies the charging Warrior as source AND weapon. CRUSH
	// skips fire/electric protection; 004E0ED0 skips distinct-weapon attribution
	// and 004E0FC9 records the damage type before the normal damage tail.
	playerCharge := monsterUpdate != nil && source != nil && source.Class().Has(object.ClassPlayer) &&
		weapon == source && typ == object.DamageCrush
	selfSourcedMissileImpact := monsterUpdate != nil && source != nil && source == weapon &&
		source.Class().Has(object.ClassMissile) && !source.Class().HasAny(object.MaskUnits) && typ == object.DamageImpact
	playerFiredMissileImpact := monsterUpdate != nil && source != nil && source.Class().Has(object.ClassPlayer) &&
		weapon != nil && weapon.Class().Has(object.ClassMissile) && typ == object.DamageImpact
	monsterFiredMissileImpact := monsterUpdate != nil && source != nil && source.Class().Has(object.ClassMonster) &&
		source.UpdateData != nil && weapon != nil && weapon.Class().Has(object.ClassMissile) && typ == object.DamageImpact
	missileImpact := selfSourcedMissileImpact || playerFiredMissileImpact || monsterFiredMissileImpact
	playerFiredMissileExplosion := monsterUpdate != nil && source != nil && source.Class().Has(object.ClassPlayer) &&
		weapon != nil && weapon.Class().Has(object.ClassMissile) && typ == object.DamageExplosion
	missileSourcedExplosion := monsterUpdate != nil && source != nil && source.Class().Has(object.ClassMissile) &&
		!source.Class().HasAny(object.MaskUnits) && weapon == nil && typ == object.DamageExplosion
	missileExplosion := playerFiredMissileExplosion || missileSourcedExplosion || spellMissileExplosion
	missileDamage := missileImpact || missileExplosion || missilePierce || missileFlame || worldProjectile
	if monsterUpdate != nil {
		// GAME.EXE rejects type 5 for poison-immune subclass 0x200 before
		// health, defense callbacks or attribution are touched.
		if sourceLessMonsterPoison && uint32(target.SubClass())&0x200 != 0 {
			return true
		}
		if target.HealthData == nil && !monsterCloudPoison && !worldFlame && !casterImpact && !monsterWorldLava && !(worldProjectile && typ == object.DamageFlame) {
			return defaultDamageUnsupported4E0B30(runtime, "monster without health", target, source, weapon, damage, typ)
		}
		// GAME.EXE 004E0EA1 handles WEAPON|WAND (0x1001000) alike,
		// while 004E0EF5 identifies a no-weapon hit by damage type 10.
		playerMelee := source != nil && source.Class().Has(object.ClassPlayer) &&
			((weapon != nil && weapon.Class().HasAny(object.ClassWeapon|object.ClassWand) && typ == object.DamageBlade) ||
				(weapon == nil && typ == object.DamageClaw))
		monsterBite := source != nil && source.Class().Has(object.ClassMonster) && source.UpdateData != nil &&
			weapon == source && typ == object.DamageBite
		// Armed NPCs and ordinary monster/player targets share the restored
		// ordinaryMelee path, including WAND melee and the Hammer exception.
		if !ordinaryMelee && !monsterImpact && !casterImpact && !simpleCrush && !playerMelee && !monsterBite && !missileDamage && !worldFlame && !worldImpale && !monsterWorldLava && !monsterCloudPoison && !monsterElectric && !monsterWeaponlessExplosion && !weaponlessRawSpell && !sourceLessMonsterBlade && !sourceLessMonsterPoison && !monsterWeaponCrush && !playerCharge && !zapRay {
			return defaultDamageUnsupported4E0B30(runtime, "unsupported monster damage shape", target, source, weapon, damage, typ)
		}
		// This monster subclass ignores both electric damage types.
		if monsterElectric && !unitSelfWeaponElectric && uint32(target.SubClass())&0x800 != 0 {
			return true
		}
		if monsterBite && runtime.MonsterHasHitSound == nil {
			return defaultDamageUnsupported4E0B30(runtime, "missing monster hit-sound lookup", target, source, weapon, damage, typ)
		}
		// The original's melee friendly-hit gate does not apply to a missile
		// SIMPLE CRUSH, SIMPLE|IMMOBILE ray or PLAYER-class charge weapon
		// (004E1400 is false).
		// The earlier campaign owner gate still applies to a friendly charge.
		// A weapon-less explosion never enters that second melee gate.
		if source != nil && !ordinaryMelee && !monsterImpact && !casterImpact && !simpleCrush && !missileDamage && !worldFlame && !worldImpale && !monsterCloudPoison && !monsterWeaponlessExplosion && !weaponlessRawSpell && !playerCharge && !unitSelfWeaponElectric && !zapRay && (runtime.IsEnemy == nil || !runtime.IsEnemy(target, source)) {
			return true
		}
	}
	vampirism := source != nil && target.Class().HasAny(object.MaskUnits) &&
		source.HasEnchant(damageVampirismEnchant4E0B30)
	if vampirism && !monsterCloudPoison && (runtime.Audio == nil || runtime.BalanceFloatInd == nil ||
		runtime.AdjustHP == nil || runtime.VampirismFX == nil) {
		return defaultDamageUnsupported4E0B30(runtime, "missing Vampirism service", target, source, weapon, damage, typ)
	}

	shockRetaliates := source != nil && weapon != nil &&
		target.HasEnchant(defaultDamageShockEnchant4E0B30) &&
		source.Class().HasAny(object.ClassPlayer|object.ClassMonster) &&
		defaultDamageAttackQualifies4E1400(source, weapon)
	if shockRetaliates {
		if runtime.Audio == nil || runtime.BuffOff == nil || runtime.BalanceFloatInd == nil ||
			runtime.CallDamage == nil ||
			(source.Class().Has(object.ClassPlayer) && runtime.PlayerSetState == nil) {
			return defaultDamageUnsupported4E0B30(runtime, "missing Shock retaliation service", target, source, weapon, damage, typ)
		}
		runtime.Audio(defaultDamageShockSound4E0B30, source)
		runtime.BuffOff(target, defaultDamageShockEnchant4E0B30)
		shockDamage := playerCollideRound4E8460(float32(runtime.BalanceFloatInd(
			defaultDamageShockBalance4E0B30, defaultDamageShockBalanceIndex4E0B30,
		)))
		_ = runtime.CallDamage(source, target, nil, shockDamage, object.DamageElectric)
		if source.Class().Has(object.ClassPlayer) {
			_ = runtime.PlayerSetState(source, PlayerState23)
		}
	}
	// 004E0C9E's Shock retaliation precedes 004E0D69's electric immunity.
	// This ordering becomes reachable when a MONSTER self-weapon is retained.
	if unitSelfWeaponElectric && monsterUpdate != nil && uint32(target.SubClass())&0x800 != 0 {
		return true
	}
	if casterImpact && (runtime.DamageClear == nil || (source != nil && (runtime.BuffOff == nil ||
		(source.Class().Has(object.ClassMonster) && source.UpdateData != nil && runtime.MonsterHasHitSound == nil)))) {
		return defaultDamageUnsupported4E0B30(runtime, "missing caster IMPACT tail service", target, source, weapon, damage, typ)
	}
	// IMPACT has no elemental immunity/protection branch. Nil health is
	// guarded by 004EE5E0 after attribution/sound, not by this attack prefix.
	if monsterCloudPoison {
		// 004E0D20 reloads class/subclass AFTER Shock. Normally a pure
		// cloud cannot retaliate, but an enemy callback can change its class.
		// Poison immunity exits before health, invisibility, Defend or HP.
		if target.Class().Has(object.ClassMonster) && uint32(target.SubClass())&0x200 != 0 {
			return true
		}
		if target.HealthData == nil {
			return defaultDamageUnsupported4E0B30(runtime, "monster without health", target, source, weapon, damage, typ)
		}
		if runtime.DamageClear == nil || (source != nil && (runtime.BuffOff == nil ||
			(source.Class().Has(object.ClassMonster) && source.UpdateData != nil && runtime.MonsterHasHitSound == nil))) {
			return defaultDamageUnsupported4E0B30(runtime, "missing monster world POISON tail service", target, source, weapon, damage, typ)
		}
		if vampirism && (runtime.Audio == nil || runtime.BalanceFloatInd == nil || runtime.AdjustHP == nil || runtime.VampirismFX == nil) {
			return defaultDamageUnsupported4E0B30(runtime, "missing Vampirism service", target, source, weapon, damage, typ)
		}
	}
	// 004E0D20 reloads class/subclass after Shock; 004E0D3E..004E0D52
	// rejects FLAME for fire-immune monsters before protection or attribution.
	if (missileFlame || worldFlame || monsterWorldLava || (worldProjectile && typ == object.DamageFlame)) && target.Class().Has(object.ClassMonster) && uint32(target.SubClass())&0x400 != 0 {
		return true
	}
	if worldProjectile && (target.HealthData == nil || runtime.BuffOff == nil || runtime.DamageClear == nil ||
		(playerTail && runtime.PlayerSetState == nil) ||
		(source.Class().Has(object.ClassMonster) && source.UpdateData != nil && runtime.MonsterHasHitSound == nil)) {
		return defaultDamageUnsupported4E0B30(runtime, "missing world projectile tail service", target, source, weapon, damage, typ)
	}
	if monsterWorldLava && (target.HealthData == nil || runtime.BuffOff == nil || runtime.DamageClear == nil) {
		return defaultDamageUnsupported4E0B30(runtime, "missing monster world LAVA tail service", target, source, weapon, damage, typ)
	}
	if worldFlame {
		// Immunity at 004E0D43 precedes health and tail-service admission.
		if target.HealthData == nil {
			return defaultDamageUnsupported4E0B30(runtime, "world FLAME without health", target, source, weapon, damage, typ)
		}
		if runtime.DamageClear == nil || (source != nil && (runtime.BuffOff == nil ||
			(source.Class().Has(object.ClassMonster) && source.UpdateData != nil && runtime.MonsterHasHitSound == nil))) {
			return defaultDamageUnsupported4E0B30(runtime, "missing world FLAME tail service", target, source, weapon, damage, typ)
		}
	}
	if (missileExplosion || monsterWeaponlessExplosion) && target.Class().Has(object.ClassMonster) && uint32(target.SubClass())&0x400 != 0 {
		// 004E0D5A..004E0D63: signed division truncates toward zero,
		// before fire protection's separate binary32 rounding/minimum.
		damage /= 2
	}

	// Admission only. The 004E0F77 call must capture the live inventory
	// after protection/position/BuffOff, not an entry-time effect plan.
	if target.Class().Has(object.ClassPlayer) ||
		(target.Class().Has(object.ClassMonster) && uint32(target.SubClass())&0x10 != 0) {
		_, ok := playerDamageLateDefendReady4E17B0(target, PlayerDamageRuntime4E17B0{
			CanApplyLateDefend: runtime.CanApplyLateDefend,
			ApplyLateDefend:    runtime.ApplyLateDefend,
		})
		if !ok {
			return defaultDamageUnsupported4E0B30(runtime, "monster defense callbacks", target, source, weapon, damage, typ)
		}
	}

	nonUnit := !target.Class().HasAny(object.MaskUnits)
	sourceLessLava := typ == object.DamageLava && source == nil && weapon == nil
	if typ != object.DamageBlade && typ != object.DamageClaw && typ != object.DamageBite &&
		!ordinaryMelee && !monsterImpact && !casterImpact && !simpleCrush && !missileDamage && !worldFlame && !worldImpale && !monsterWorldLava && !monsterCloudPoison && !nonUnit && !monsterElectric && !monsterWeaponlessExplosion && !weaponlessRawSpell && !sourceLessMonsterPoison && !playerTail && !monsterWeaponCrush && !playerCharge && !zapRay {
		return defaultDamageUnsupported4E0B30(runtime, "unsupported protection branch", target, source, weapon, damage, typ)
	}
	fireProtected := typ == object.DamageFlame || typ == object.DamageLava || typ == object.DamageExplosion
	electricProtected := typ == object.DamageElectric || typ == object.DamageAirborneElectric
	if fireProtected && runtime.FireProtection == nil {
		return defaultDamageUnsupported4E0B30(runtime, "missing fire-protection service", target, source, weapon, damage, typ)
	}
	if electricProtected && runtime.ElectricProtection == nil {
		return defaultDamageUnsupported4E0B30(runtime, "missing electric-protection service", target, source, weapon, damage, typ)
	}
	// Availability admission only. 004E11BF queries the live buff after
	// protection, Defend, pre-Damage, sound and every later hit callback.
	shieldNeedsService := target.HasEnchant(defaultDamageShieldEnchant4E0B30) && typ != object.DamagePoison &&
		(typ != object.DamageManaBomb || source != target)
	if shieldNeedsService && runtime.ShieldReduce == nil {
		return defaultDamageUnsupported4E0B30(runtime, "missing Shield reduction service", target, source, weapon, damage, typ)
	}
	// Read-only availability admission, not an execution plan. The original
	// 004E0FEF call rechecks the weapon class and captures its init base later.
	for _, modifier := range defaultDamageWeaponPreDamageModifiers4E0B30(weapon) {
		if modifier == nil || modifier.AttackPreDmg64.Fnc == nil {
			continue
		}
		if runtime.CanApplyPreDamage == nil || runtime.ApplyPreDamage == nil ||
			!runtime.CanApplyPreDamage(modifier) {
			return defaultDamageUnsupported4E0B30(runtime, "weapon pre-damage modifiers", target, source, weapon, damage, typ)
		}
	}
	playerSound := playerTail && target.DamageSound != nil && target.DamageSound == runtime.PlayerDamageSoundC
	if playerSound && runtime.PlayerDamageSound == nil {
		return defaultDamageUnsupported4E0B30(runtime, "missing player damage sound", target, source, weapon, damage, typ)
	}
	if target.DamageSound != nil && target.DamageSound != runtime.DefaultDamageSoundC && !playerSound {
		return defaultDamageUnsupported4E0B30(runtime, "custom damage sound", target, source, weapon, damage, typ)
	}
	// Electric resistance or late defend can raise even a small input above
	// the ball-release threshold. Validate the carrier tail before any stores.
	if playerTail && target.Field129 != nil {
		if runtime.GameBallType == 0 {
			return defaultDamageUnsupported4E0B30(runtime, "missing GameBall type", target, source, weapon, damage, typ)
		}
		if playerDamageOwnsType4E17B0(target, runtime.GameBallType) && runtime.GameBallOnDamage == nil {
			return defaultDamageUnsupported4E0B30(runtime, "GameBall drop", target, source, weapon, damage, typ)
		}
	}
	if fireProtected {
		protectionValue := runtime.FireProtection(target)
		if protectionValue != 0 && byte(currentFrame())&3 == 0 && runtime.Audio != nil {
			runtime.Audio(104, target)
		}
		// GAME.EXE compares the binary64 return before spilling it to v46,
		// then evaluates the damage expression in binary64 and spills v42 to
		// binary32 immediately before FISTP.
		protection := float32(protectionValue)
		scaled := float32((1.0 - float64(protection)) * float64(damage))
		damage = int32(math.RoundToEven(float64(scaled)))
		if damage == 0 {
			damage = 1
		}
	}
	if electricProtected {
		protectionValue := runtime.ElectricProtection(target)
		if protectionValue != 0 && byte(currentFrame())&3 == 0 && runtime.Audio != nil {
			runtime.Audio(108, target)
		}
		protection := float32(protectionValue)
		scaled := float32((1.0 - float64(protection)) * float64(damage))
		damage = int32(math.RoundToEven(float64(scaled)))
		if damage == 0 {
			damage = 1
		}
		// 004E0E60 reloads the class after protection/audio. PLAYER wins
		// over MONSTER, and both branches reload their native update address.
		class := target.Class()
		if class.Has(object.ClassPlayer) {
			// 004E0E6D writes only the low word; the adjacent word is live.
			target.UpdateDataPlayer().Field40_0 = 2
		} else if class.Has(object.ClassMonster) && uint32(target.SubClass())&0x10 != 0 {
			target.UpdateDataMonster().Field523_2 = 2
		}
	}
	if source == nil {
		target.Pos132 = types.Pointf{}
	} else if weapon != nil {
		if weapon.Class().HasAny(object.ClassWeapon | object.ClassWand) {
			target.Pos132 = source.PrevPos
		} else {
			target.Pos132 = weapon.PrevPos
		}
	} else {
		target.Pos132 = source.PrevPos
	}
	// 004E0ED4 tests only the MONSTER class, not the scripted-NPC subclass.
	// A distinct weapon records its type before BuffOff and the injured latch.
	if source != nil && target.Class().Has(object.ClassMonster) {
		if weapon != nil && weapon != source {
			update := target.UpdateDataMonster()
			update.Field547 = 1
			update.Field546 = uint32(weapon.TypeInd)
		} else if weapon == nil && (typ == object.DamageClaw || typ == object.DamageCrush) {
			update := target.UpdateDataMonster()
			update.Field547 = 1
			update.Field546 = uint32(source.TypeInd)
		}
	}
	if (source != nil || sourceLessLava) && runtime.BuffOff != nil {
		// GAME.EXE calls BuffOff even when INVSIBILITY is not currently set.
		runtime.BuffOff(target, defaultDamageInvisibleEnchant4E0B30)
	}
	// 004E0F5D reloads class/subclass after the callbacks above. Ordinary
	// monsters do not enter 004E1320; players and NPC-subclass monsters do.
	if target.Class().Has(object.ClassPlayer) ||
		(target.Class().Has(object.ClassMonster) && uint32(target.SubClass())&0x10 != 0) {
		ItemDefendEffects4E1320(target, source, weapon, &damage, int32(typ), ItemDefendEffectsRuntime4E1320{
			ApplyDefend: func(modifier *ModifierEff, item, owner, weapon, attacker *Object, context *[2]int32) {
				// A callback may introduce an unported live function after
				// admission. Do not enter ABI32 or alter its damage word; report
				// it at use and continue visiting supported later slots/items.
				if runtime.CanApplyLateDefend == nil || runtime.ApplyLateDefend == nil || !runtime.CanApplyLateDefend(modifier) {
					if runtime.Unsupported != nil {
						runtime.Unsupported("unsupported live late equipped-item defend effect", owner, attacker, weapon, context[0], object.DamageType(context[1]))
					}
					return
				}
				context[0] = runtime.ApplyLateDefend(modifier, item, owner, weapon, attacker, context[0], object.DamageType(context[1]))
			},
		})
	}
	if weapon != nil {
		target.Obj130 = weapon
	} else {
		target.Obj130 = source
	}
	// 004E0F91/004E0FAA reload class/update after late Defend; the entry
	// latch reset must not retain a stale Monster update for these stores.
	injuredMonster := target.Class().Has(object.ClassMonster)
	target.Field131 = uint32(typ)
	// 004E0F9A is after late Defend and before the injured/update stores;
	// pre-Damage, sound and Shield must not replay this timestamp.
	target.Frame134 = currentFrame()
	if injuredMonster {
		update := target.UpdateDataMonster()
		update.StatusFlags |= object.MonStatusInjured
		if update.Field547 == 0 {
			update.Field547 = 2
			update.Field546 = uint32(typ)
		}
	}
	// 004E0FD9/004E0FDD use the live weapon class after the attribution and
	// injured-latch stores. The helper caches only its native init base and
	// reloads each later slot/function after the preceding callback returns.
	if weapon != nil && weapon.Class().HasAny(object.ClassWeapon|object.ClassWand) {
		ItemPreDamage4E13B0(target, source, weapon, &damage, ItemPreDamageRuntime4E13B0{
			ApplyPreDamage: func(modifier *ModifierEff, w, attacker, owner *Object, d *int32) {
				if runtime.CanApplyPreDamage == nil || runtime.ApplyPreDamage == nil || !runtime.CanApplyPreDamage(modifier) {
					if runtime.Unsupported != nil {
						runtime.Unsupported("unsupported live weapon pre-damage effect", owner, attacker, w, *d, typ)
					}
					return
				}
				runtime.ApplyPreDamage(modifier, w, attacker, owner, d)
			},
		})
	}

	suppressDamageSound := target == weapon && target.Class().HasAny(object.ClassWeapon|object.ClassWand)
	if !suppressDamageSound && source != nil && source.Class().Has(object.ClassMonster) &&
		source.UpdateData != nil && runtime.MonsterHasHitSound != nil {
		suppressDamageSound = runtime.MonsterHasHitSound(source)
	}
	if !suppressDamageSound {
		soundSource := source
		if weapon != nil {
			soundSource = weapon
		}
		if playerSound {
			runtime.PlayerDamageSound(target, soundSource)
		} else if runtime.DefaultDamageSound != nil {
			runtime.DefaultDamageSound(target, soundSource)
		}
	}
	if vampirism {
		damageApplyVampirism4E0B30(
			source, target, weapon, damage,
			runtime.Audio, runtime.BalanceFloatInd, runtime.AdjustHP, runtime.VampirismFX,
		)
	}
	if playerTail {
		// Newly installed live effects may raise damage and attach a ball
		// after the carrier preflight. Preserve the executed hit prefix but
		// stop an unsupported required drop instead of silently omitting it.
		if damage >= 30 && target.Field129 != nil {
			if runtime.GameBallType == 0 {
				return defaultDamageUnsupported4E0B30(runtime, "unsupported live GameBall type", target, source, weapon, damage, typ)
			}
			if playerDamageOwnsType4E17B0(target, runtime.GameBallType) && runtime.GameBallOnDamage == nil {
				return defaultDamageUnsupported4E0B30(runtime, "unsupported live GameBall drop", target, source, weapon, damage, typ)
			}
		}
		if runtime.GameBallOnDamage != nil {
			runtime.GameBallOnDamage(source, target, damage)
		}
		// 004E1136/004E1147 reload class/update after damage sound,
		// Vampirism and GameBall. Neither entry class nor the earlier
		// electric-state address owns the hurt-state decision.
		if target.Class().Has(object.ClassPlayer) && damage >= 20 {
			if target.UpdateData == nil && !casterImpact {
				return defaultDamageUnsupported4E0B30(runtime, "unsupported live player hurt update", target, source, weapon, damage, typ)
			}
			state := target.UpdateDataPlayer().State
			if state != PlayerState1 && state != PlayerState15 {
				if runtime.PlayerSetState == nil {
					return defaultDamageUnsupported4E0B30(runtime, "unsupported live player hurt-state service", target, source, weapon, damage, typ)
				}
				_ = runtime.PlayerSetState(target, PlayerState30)
			}
		}
	}
	if monsterUpdate != nil && runtime.AdjustFieldGuide != nil {
		damage = runtime.AdjustFieldGuide(source, target, damage)
	}

	if source != nil {
		monster := source
		if !monster.Class().Has(object.ClassMonster) {
			monster = source.ObjOwner
		}
		if monster != nil && monster.Class().Has(object.ClassMonster) &&
			runtime.IsEnemy != nil && runtime.IsEnemy(target, monster) {
			update := monster.UpdateDataMonster()
			if update.Field130 == 0 {
				// The inlined 00532880 tail reads only for an empty latch,
				// after IsEnemy and every preceding hit callback has returned.
				update.Field130 = currentFrame()
			}
		}
	}
	if target.HasEnchant(defaultDamageShieldEnchant4E0B30) && typ != object.DamagePoison {
		if typ != object.DamageManaBomb || source != target {
			if runtime.ShieldReduce == nil {
				return defaultDamageUnsupported4E0B30(runtime, "unsupported live Shield reduction service", target, source, weapon, damage, typ)
			}
			shieldSource := weapon
			if shieldSource == nil {
				shieldSource = source
			}
			runtime.ShieldReduce(target, &damage, typ, shieldSource)
		}
		// 004E11FA also tests raw zero for a self-ManaBomb, even though
		// that gate skips reduction. Do not re-query after Shield depletion.
		if damage == 0 {
			return false
		}
	}
	if runtime.DamageClear != nil {
		runtime.DamageClear(target, damage)
	}
	return true
}
