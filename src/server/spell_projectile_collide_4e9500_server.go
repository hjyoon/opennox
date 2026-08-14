package server

import (
	"math"
	"unsafe"

	"github.com/opennox/libs/spell"
	"github.com/opennox/libs/types"

	"github.com/opennox/opennox/v1/common/sound"
)

// SpellProjectileCollideRuntime4E9500 supplies high-level effects owned by
// the main game runtime. All object and update-data access remains native-width
// inside the server package.
type SpellProjectileCollideRuntime4E9500 struct {
	CheckDirection  func(types.Pointf, int16, types.Pointf) int32
	ChangeOwner     func(*Object, *Object)
	SetPlayerState  func(*Object, PlayerState) bool
	SpellAccept     func(spell.ID, *Object, *Object, *Object, *SpellAcceptArg, int) bool
	DelayedDelete   func(*Object)
	InversionEffect unsafe.Pointer
}

type spellProjectileNativeDeps4E9500 struct {
	runtime          SpellProjectileCollideRuntime4E9500
	randomInt        func(int32, int32) int32
	playerAnimFrames func(int32) (int32, int32)
	audio            func(uint32, *Object)
}

// spellProjectileWallReflect57B810 reproduces the collision-normal reflection
// used by the nil-other branch. The product is evaluated without a binary32
// spill, matching the x87 compare in GAME.EXE.
func spellProjectileWallReflect57B810(collision *types.Pointf, projectile *Object) {
	x, y := projectile.VelVec.X, projectile.VelVec.Y
	// FCOMP followed by TEST AH,0x41 takes the swap path for a negative
	// product, either signed zero, or an unordered (NaN) comparison.
	if !(float64(collision.Y)*float64(collision.X) > 0) {
		projectile.VelVec.X = y
		projectile.VelVec.Y = x
		return
	}
	projectile.VelVec.X = -y
	projectile.VelVec.Y = -x
}

// spellProjectileReflect4E0A70 is the native-pointer implementation of the
// reflect helper called by GAME.EXE 004E9500. Explicit binary32 spills mirror
// FSTP sites in the original x87 instruction stream.
func spellProjectileReflect4E0A70(projectile, other *Object) {
	dx := float64(projectile.PosVec.X) - float64(other.PosVec.X)
	dy := float64(projectile.PosVec.Y) - float64(other.PosVec.Y)
	direction := projectile.Direction1 + 128
	projectile.Direction1 = direction

	negativeY := float32(-dy)
	denominator := float32(math.Sqrt(dy*dy+dx*dx) + float64(float32(0.1)))
	denominator64 := float64(denominator)
	projection := (dy*float64(projectile.VelVec.Y) + dx*float64(projectile.VelVec.X)) / denominator64
	parallelX := float32(projection * dx / denominator64)
	parallelY := float32(projection * dy / denominator64)
	orthogonal := (float64(negativeY)*float64(projectile.VelVec.X) + dx*float64(projectile.VelVec.Y)) / denominator64
	orthogonalX := float32(orthogonal * float64(negativeY) / denominator64)
	orthogonalY := orthogonal * dx / denominator64
	projectile.VelVec.X = float32(float64(orthogonalX) - float64(parallelX))
	projectile.VelVec.Y = float32(orthogonalY - float64(parallelY))

	if int16(direction) >= 256 {
		projectile.Direction1 = direction - 256
	}
	projectile.NewPos = projectile.PrevPos
}

func spellProjectileInversionNative4FA4F0(
	target, projectile *Object,
	inversionEffect unsafe.Pointer,
) int32 {
	return spellProjectileInversion4FA4F0(target, projectile, spellProjectileInversionHooks4FA4F0[
		*Object,
		*Object,
		*ModifierInitData,
		*ModifierEff,
		unsafe.Pointer,
	]{
		firstItem: func(obj *Object) *Object {
			return obj.InvFirstItem
		},
		loadFlags: func(item *Object) uint32 {
			return uint32(item.ObjFlags)
		},
		loadClass: func(item *Object) uint32 {
			return uint32(item.ObjClass)
		},
		loadInitData: func(item *Object) *ModifierInitData {
			return (*ModifierInitData)(item.InitData)
		},
		loadModifier: func(data *ModifierInitData, slot int) *ModifierEff {
			return data.Modifiers[slot]
		},
		loadDefendCollide: func(modifier *ModifierEff) unsafe.Pointer {
			return modifier.DefendCollide88.Fnc
		},
		inversionEffect: inversionEffect,
		findParent:      (*Object).FindOwnerChainPlayer,
		loadInversionStrength: func(modifier *ModifierEff) int32 {
			return modifier.DefendCollide88.Val
		},
		nextItem: func(item *Object) *Object {
			return item.InvNextItem
		},
	})
}

// spellProjectilePlayerActionNative4E9500 covers every fixed state mapping in
// GAME.EXE 004FA2B0 plus state 13's great-sword result. The collision path can
// reach state 13 only when SetPlayerState rejects the requested 18..20 state;
// otherwise it reaches the three fixed mappings below.
func spellProjectilePlayerActionNative4E9500(obj *Object) int32 {
	update := (*PlayerUpdateData)(obj.UpdateData)
	switch update.State {
	case PlayerState0:
		return 4
	case PlayerState2, PlayerState10:
		return 21
	case PlayerState3:
		return 1
	case PlayerState4:
		return 2
	case PlayerState5:
		return 6
	case PlayerState12:
		return 3
	case PlayerState13:
		if update.Player.WeaponEquip&spellProjectileGreatSwordMask4E9500 != 0 {
			return 38
		}
		return 0
	case PlayerState15, PlayerState16, PlayerState17:
		return 40
	case PlayerState18:
		return 48
	case PlayerState19:
		return 49
	case PlayerState20:
		return 47
	case PlayerState21:
		return 30
	case PlayerState23:
		return 50
	case PlayerState24:
		return 19
	case PlayerStateShakeFist:
		return 20
	case PlayerStateLaugh:
		return 15
	case PlayerState27, PlayerStatePoint, PlayerState29:
		return 16
	case PlayerState30:
		return 52
	case PlayerState32:
		return 54
	default:
		return 0
	}
}

func spellProjectileCollideNative4E9500(
	projectile, other *Object,
	collision *types.Pointf,
	deps spellProjectileNativeDeps4E9500,
) {
	spellProjectileCollide4E9500(projectile, other, collision, spellProjectileCollideHooks4E9500[
		*Object,
		*SpellProjectileUpdateData,
		*PlayerUpdateData,
		*Player,
		*types.Pointf,
	]{
		loadProjectileUpdate: func(obj *Object) *SpellProjectileUpdateData {
			return (*SpellProjectileUpdateData)(obj.UpdateData)
		},
		wallReflect: spellProjectileWallReflect57B810,
		loadFlags: func(obj *Object) uint32 {
			return uint32(obj.ObjFlags)
		},
		loadClassLow: func(obj *Object) uint8 {
			return uint8(obj.ObjClass)
		},
		loadPlayerUpdate: func(obj *Object) *PlayerUpdateData {
			return (*PlayerUpdateData)(obj.UpdateData)
		},
		loadPlayerState: func(update *PlayerUpdateData) uint8 {
			return uint8(update.State)
		},
		loadDirection: func(obj *Object) int16 {
			return int16(obj.Direction1)
		},
		checkPrevious: func(target *Object, direction int32, projectile *Object) int32 {
			return deps.runtime.CheckDirection(target.PosVec, int16(direction), projectile.PrevPos)
		},
		audio:             deps.audio,
		projectileReflect: spellProjectileReflect4E0A70,
		changeOwner:       deps.runtime.ChangeOwner,
		loadPlayer: func(update *PlayerUpdateData) *Player {
			return update.Player
		},
		loadWeaponEquip: func(player *Player) uint32 {
			return player.WeaponEquip
		},
		randomInt: deps.randomInt,
		setPlayerState: func(obj *Object, state int32) {
			_ = deps.runtime.SetPlayerState(obj, PlayerState(state))
		},
		mapPlayerAction:  spellProjectilePlayerActionNative4E9500,
		playerAnimFrames: deps.playerAnimFrames,
		storeAnimFrame: func(update *PlayerUpdateData, frame uint8) {
			update.Field59_0 = frame
		},
		checkInversion: func(target, projectile *Object) int32 {
			return spellProjectileInversionNative4FA4F0(target, projectile, deps.runtime.InversionEffect)
		},
		hasEnchant: func(obj *Object, enchant uint32) int32 {
			if obj.HasEnchant(EnchantID(enchant)) {
				return 1
			}
			return 0
		},
		checkCurrent: func(target *Object, direction int32, projectile *Object) int32 {
			return deps.runtime.CheckDirection(target.PosVec, int16(direction), projectile.PosVec)
		},
		loadTarget: func(update *SpellProjectileUpdateData) *Object {
			return update.Target
		},
		loadLevel: func(update *SpellProjectileUpdateData) int32 {
			return int32(update.Level16)
		},
		loadOwner: func(update *SpellProjectileUpdateData) *Object {
			return update.Field0
		},
		loadSource: func(update *SpellProjectileUpdateData) *Object {
			return update.Field8
		},
		loadSpell: func(update *SpellProjectileUpdateData) int32 {
			return int32(update.Spell12)
		},
		spellAccept: func(spellID int32, source, owner, projectile, target *Object, level int32) int32 {
			arg := SpellAcceptArg{Obj: target}
			if deps.runtime.SpellAccept(spell.ID(spellID), source, owner, projectile, &arg, int(level)) {
				return 1
			}
			return 0
		},
		delayedDelete: deps.runtime.DelayedDelete,
	})
}

// SpellProjectileCollide4E9500 binds the original collision callback to
// native-width Object, PlayerUpdateData, Player, modifier and spell records.
func (s *Server) SpellProjectileCollide4E9500(
	projectile, other *Object,
	collision *types.Pointf,
	runtime SpellProjectileCollideRuntime4E9500,
) {
	spellProjectileCollideNative4E9500(projectile, other, collision, spellProjectileNativeDeps4E9500{
		runtime: runtime,
		randomInt: func(minimum, maximum int32) int32 {
			return int32(s.Rand.Logic.IntClamp(int(minimum), int(maximum)))
		},
		playerAnimFrames: func(action int32) (int32, int32) {
			first, last := s.PlayerAnimFrames(int(action))
			return int32(first), int32(last)
		},
		audio: func(id uint32, obj *Object) {
			s.Audio.EventObj(sound.ID(id), obj, 0, 0)
		},
	})
}
