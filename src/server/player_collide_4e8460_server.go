package server

import (
	"unsafe"

	"github.com/opennox/libs/object"
	"github.com/opennox/libs/types"

	"github.com/opennox/opennox/v1/common/sound"
)

type playerCollideNativeDeps4E8460 struct {
	abilityActive  func(*Object, Ability) bool
	setState       func(*Object, PlayerState)
	earthquake     func(types.Pointf, int)
	disableAbility func(*Object, Ability)
	balanceFloat   func(string) float64
	floatToInt     func(float32) int32
	findParent     func(*Object) *Object
	damage         func(*Object, *Object, *Object, int32, object.DamageType)
	applyEnchant   func(*Object, EnchantID, uint32, uint32)
	wallFlags      func(uint8) uint32
	audio          func(uint32, *Object, int32, uint32)
	damageMap      func(int32, int32, int32, object.DamageType, *Object)
	damageClear    func(*Object, int32)
	move           func(*Object, types.Pointf)
	frameRate      func() uint32
	disableEnchant func(*Object, EnchantID)
}

// PlayerCollideRuntime4E8460 supplies the remaining game-level side effects
// that intentionally live above package server. Object reads and all native
// pointer-width-sensitive dispatch stay inside PlayerCollide4E8460.
type PlayerCollideRuntime4E8460 struct {
	SetState       func(*Object, PlayerState)
	DisableAbility func(*Object, Ability)
	ApplyEnchant   func(*Object, EnchantID, uint32, uint32)
	DamageMap      func(int32, int32, int32, object.DamageType, *Object)
	DamageClear    func(*Object, int32)
	Move           func(*Object, types.Pointf)
	DisableEnchant func(*Object, EnchantID)
}

func playerCollideNative4E8460(
	player, other *Object,
	collision unsafe.Pointer,
	deps playerCollideNativeDeps4E8460,
) {
	playerCollide4E8460(player, other, collision, playerCollideHooks4E8460[
		*Object,
		*HealthData,
		*Wall,
		unsafe.Pointer,
	]{
		abilityActive: func(obj *Object, ability uint32) int32 {
			if deps.abilityActive(obj, Ability(ability)) {
				return 1
			}
			return 0
		},
		class:          func(obj *Object) uint32 { return uint32(obj.ObjClass) },
		flags:          func(obj *Object) uint32 { return uint32(obj.ObjFlags) },
		flagsLow:       func(obj *Object) uint8 { return uint8(obj.ObjFlags) },
		health:         func(obj *Object) *HealthData { return obj.HealthData },
		healthCurrent:  func(health *HealthData) uint16 { return health.Cur },
		healthMax:      func(health *HealthData) uint16 { return health.Max },
		mass:           func(obj *Object) float32 { return obj.Mass },
		doorState:      func(obj *Object) uint8 { return obj.UpdateDataDoor().LockCode },
		setState:       func(obj *Object, state uint32) { deps.setState(obj, PlayerState(state)) },
		earthquake:     func(obj *Object, jiggle int32) { deps.earthquake(obj.PosVec, int(jiggle)) },
		disableAbility: func(obj *Object, ability uint32) { deps.disableAbility(obj, Ability(ability)) },
		balanceFloat:   deps.balanceFloat,
		floatToInt:     deps.floatToInt,
		bounce:         func(player, other *Object) { playerBerserkBounceNative4E86E0(player, other) },
		findParent:     deps.findParent,
		damage: func(obj, source, attacker *Object, damage int32, damageType uint32) {
			deps.damage(obj, source, attacker, damage, object.DamageType(damageType))
		},
		applyEnchant: func(obj *Object, enchant, duration, power uint32) {
			deps.applyEnchant(obj, EnchantID(enchant), duration, power)
		},
		collisionWall: func(obj *Object) *Wall { return obj.UpdateDataPlayer().CollisionWall },
		wallTile:      func(wall *Wall) uint8 { return wall.Tile1 },
		wallFlags:     deps.wallFlags,
		audio:         deps.audio,
		newPosY:       func(obj *Object) float32 { return obj.NewPos.Y },
		newPosX:       func(obj *Object) float32 { return obj.NewPos.X },
		damageMap: func(x, y, damage int32, damageType uint32, source *Object) {
			deps.damageMap(x, y, damage, object.DamageType(damageType), source)
		},
		damageClear: deps.damageClear,
		move:        func(obj *Object) { deps.move(obj, obj.PrevPos) },
		hasEnchant: func(obj *Object, enchant uint32) int32 {
			if obj.HasEnchant(EnchantID(enchant)) {
				return 1
			}
			return 0
		},
		enchantTimer: func(obj *Object, enchant uint32) uint32 {
			return uint32(obj.EnchantDur(EnchantID(enchant)))
		},
		frameRate: deps.frameRate,
		enchantPower: func(obj *Object, enchant uint32) uint32 {
			return uint32(obj.EnchantPower(EnchantID(enchant)))
		},
		disableEnchant: func(obj *Object, enchant uint32) { deps.disableEnchant(obj, EnchantID(enchant)) },
	})
}

func playerBerserkBounceNative4E86E0(player, other *Object) {
	playerBerserkBounce4E86E0(player, other, playerBerserkBounceHooks4E86E0[*Object]{
		mass: func(obj *Object) float32 { return obj.Mass },
		velocity: func(obj *Object, axis int) float32 {
			if axis == 0 {
				return obj.VelVec.X
			}
			return obj.VelVec.Y
		},
		store: func(obj *Object, axis int, value float32) {
			if axis == 0 {
				obj.VelVec.X = value
			} else {
				obj.VelVec.Y = value
			}
		},
	})
}

// PlayerCollide4E8460 binds the original Player collision state transition to
// native-pointer Object, HealthData, PlayerUpdateData, Wall and callback APIs.
func (s *Server) PlayerCollide4E8460(
	player, other *Object,
	collision unsafe.Pointer,
	runtime PlayerCollideRuntime4E8460,
) {
	playerCollideNative4E8460(player, other, collision, playerCollideNativeDeps4E8460{
		abilityActive:  s.Abils.IsActive,
		setState:       runtime.SetState,
		earthquake:     s.Nox_xxx_earthquakeSend_4D9110,
		disableAbility: runtime.DisableAbility,
		balanceFloat:   s.Balance.Float,
		floatToInt:     playerCollideRound4E8460,
		findParent:     (*Object).FindOwnerChainPlayer,
		damage: func(target, source, attacker *Object, damage int32, damageType object.DamageType) {
			target.CallDamage(source, attacker, int(damage), damageType)
		},
		applyEnchant: runtime.ApplyEnchant,
		wallFlags:    func(tile uint8) uint32 { return s.Walls.DefByInd(int(tile)).Flags32 },
		audio: func(id uint32, obj *Object, kind int32, code uint32) {
			s.Audio.EventObj(sound.ID(id), obj, int(kind), code)
		},
		damageMap:      runtime.DamageMap,
		damageClear:    runtime.DamageClear,
		move:           runtime.Move,
		frameRate:      s.TickRate,
		disableEnchant: runtime.DisableEnchant,
	})
}
