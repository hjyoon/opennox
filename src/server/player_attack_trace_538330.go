package server

import (
	"math"

	"github.com/opennox/libs/object"
	"github.com/opennox/libs/types"
)

// PlayerAttackData538330 matches the native C attack record. Modifiers can
// update this same record before damage, wall checks and weapon wear.
type PlayerAttackData538330 struct {
	Damage        float32
	DamageType    uint8
	_             [3]byte
	Radius        float32
	Owner         *Object
	Position      types.Pointf
	AllowDisabled uint32
	Weapon        *Object
	DirectionMask uint8
}

type PlayerAttackTraceState538330 struct {
	Distance float32
	Damaged  uint32
	Closest  *Object
}

type PlayerAttackTraceRuntime538330 struct {
	Scan           func(*types.Pointf, float32, func(*Object))
	Selection      PlayerAttackTargetRuntime5386A0
	Direction      func(types.Pointf, int32, types.Pointf) uint8
	Trace          func(types.Pointf, types.Pointf) bool
	PreHit         func(target, owner, weapon *Object, attack *PlayerAttackData538330)
	Damage         func(target, owner, weapon *Object, damage int32, damageType uint8)
	ReportPlayerFX func() bool
	PointFX        func(*types.Pointf)
	Force          func(origin *types.Pointf, target *Object, force float32)
	Walls          func(bounds [4]int32, center *types.Pointf, radius float32, damage int32, damageType uint8, source *Object)
	ItemDamage     func() float64
	WeaponDamage   func(weapon, owner, source, effective *Object, damage float32, damageType uint8)
}

func (state *PlayerAttackTraceState538330) Select(candidate, attacker *Object, r PlayerAttackTargetRuntime5386A0) {
	PlayerAttackSelectCandidate5386A0(candidate, attacker, &state.Distance, &state.Closest, r)
}

func (state *PlayerAttackTraceState538330) Hit(candidate *Object, attack *PlayerAttackData538330, r PlayerAttackTraceRuntime538330) {
	if candidate == nil || attack == nil || attack.Owner == nil || candidate == attack.Owner {
		return
	}
	owner := attack.Owner
	if r.Direction(owner.PosVec, int32(int16(owner.Direction1)), candidate.PosVec)&attack.DirectionMask == 0 {
		return
	}
	flags := candidate.ObjFlags
	if flags&0x8040 != 0 || attack.AllowDisabled == 0 && flags&8 != 0 {
		return
	}
	if candidate.Material != 0x4000 {
		state.Damaged = 1
	}
	if !r.Trace(attack.Owner.PosVec, candidate.PosVec) {
		return
	}
	r.PreHit(candidate, attack.Owner, attack.Weapon, attack)
	r.Damage(candidate, attack.Owner, attack.Weapon, playerAttackDamageInt538330(attack.Damage), attack.DamageType)
	if r.ReportPlayerFX() && attack.Owner.ObjClass.Has(object.ClassPlayer) &&
		!candidate.ObjClass.Has(object.ClassMonster) && candidate.HealthData != nil &&
		candidate.HealthData.Max != 0 && candidate.ObjFlags&0x8020 == 0 {
		r.PointFX(&candidate.PosVec)
	}
	if attack.Weapon == nil && playerAttackAnimation538510(attack.Owner) == 25 {
		r.Force(&attack.Owner.PosVec, candidate, 20)
	}
}

func playerAttackAnimation538510(owner *Object) uint8 {
	if owner == nil || owner.UpdateData == nil {
		return 0
	}
	if owner.ObjClass.Has(object.ClassPlayer) {
		if player := owner.UpdateDataPlayer().Player; player != nil {
			return uint8(player.Field8)
		}
		return 0
	}
	if owner.ObjClass.Has(object.ClassMonster) {
		return uint8(owner.UpdateDataMonster().Field517)
	}
	return 0
}

func (state *PlayerAttackTraceState538330) Trace(attacker *Object, attack *PlayerAttackData538330, r PlayerAttackTraceRuntime538330) int {
	if attacker == nil || attack == nil {
		return 0
	}
	state.Damaged, state.Closest = 0, nil
	var padding float32
	if attack.Weapon != nil && attack.Weapon.ObjSubClass&0x4000 != 0 {
		r.Scan(&attack.Position, attack.Radius, func(candidate *Object) {
			state.Hit(candidate, attack, r)
		})
		padding = 25
	} else {
		state.Distance = attack.Radius
		r.Scan(&attacker.PosVec, attack.Radius, func(candidate *Object) {
			state.Select(candidate, attacker, r.Selection)
		})
		state.Hit(state.Closest, attack, r)
	}
	bounds := playerAttackWallBounds538330(attack.Position, attack.Radius, padding)
	source := attack.Weapon
	if source == nil {
		source = attacker
	}
	radius := pointOnLineSpill57C8A0(pointOnLineAddChop53_57C8A0(float64(padding), float64(attack.Radius)))
	r.Walls(bounds, &attack.Position, radius, playerAttackDamageInt538330(attack.Damage), attack.DamageType, source)
	if attack.Weapon != nil && state.Damaged != 0 {
		damageType := attack.DamageType
		factor := r.ItemDamage()
		amount := pointOnLineSpill57C8A0(pointOnLineMulChop53_57C8A0(factor, float64(attack.Damage)))
		r.WeaponDamage(attack.Weapon, attack.Owner, state.Closest, state.Closest, amount, damageType)
	}
	return int(state.Damaged)
}

func playerAttackWallBounds538330(pos types.Pointf, radius, pad float32) [4]int32 {
	bound := func(p float32, r, padding float64) int32 {
		value := float64(pointOnLineSpill57C8A0(pointOnLineAddChop53_57C8A0(pointOnLineAddChop53_57C8A0(float64(p), r), padding)))
		if !(value >= -0x1p31 && value < 0x1p31) {
			return math.MinInt32 / 23
		}
		return int32(value) / 23
	}
	r, padding := float64(radius), float64(pad)
	return [4]int32{
		bound(pos.X, -r, -padding), bound(pos.Y, -r, -padding),
		bound(pos.X, r, padding), bound(pos.Y, r, padding),
	}
}

func playerAttackDamageInt538330(value float32) int32 {
	amount := pointOnLineAddChop53_57C8A0(float64(value), 0.5)
	if !(amount >= -0x1p63 && amount < 0x1p63) {
		return 0
	}
	return int32(uint32(int64(amount)))
}
