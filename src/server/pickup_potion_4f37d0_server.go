package server

import (
	"math"

	"github.com/opennox/libs/player"
	"github.com/opennox/libs/spell"
	"github.com/opennox/libs/strman"

	noxflags "github.com/opennox/opennox/v1/common/flags"
	"github.com/opennox/opennox/v1/common/sound"
)

// PickupPotionRuntime4F37D0 contains operations still owned by the root or
// legacy boundary. All object pointers remain native-width; callback scalars
// use the original fixed-width values.
type PickupPotionRuntime4F37D0 struct {
	DefaultPickup     PickupDefaultRuntime4F31E0
	PlayerClassCanUse func(*Object, uint8) int32
	AdjustHealth      func(*Object, int32)
	AddMana           func(*Object, int32)
	DelayedDelete     func(*Object)
}

type pickupPotionNativeDeps4F37D0 struct {
	gameFlag          func(uint32) int32
	playerClassCanUse func(*Object, uint8) int32
	classFailure      func(*Object, string, uint8)
	audio             func(uint32, *Object, int32, uint32)
	playerState       func(*Object) int32
	healthMultiplier  func(uint8) float32
	adjustHealth      func(*Object, int32)
	manaMultiplier    func(uint8) float32
	addMana           func(*Object, int32)
	removePoison      func(*Object)
	spellAudio        func(int32, int32) uint32
	delayedDelete     func(*Object)
	decay             func(*Object)
	defaultPickup     func(*Object, *Object, int32, int32) int32
}

// pickupPotionFloatToInt4F37D0 models GAME.EXE 00419A70's default x87
// FISTP conversion of the already-rounded float32 argument. Out-of-range and
// non-finite inputs produce x87's signed 32-bit integer-indefinite value.
func pickupPotionFloatToInt4F37D0(value float32) int32 {
	rounded := math.RoundToEven(float64(value))
	if math.IsNaN(rounded) || rounded < math.MinInt32 || rounded > math.MaxInt32 {
		return math.MinInt32
	}
	return int32(rounded)
}

// pickupPotionScale4F37D0 preserves the x87 FILD int32, FMUL m32real,
// FSTP m32real, and FISTP int32 pipeline. Converting the exact operands to
// float64 before multiplication matches the 53-bit precision used by the
// original Windows x87 environment; the explicit float32 conversion models
// its stack spill before round-to-nearest-even integer conversion.
func pickupPotionScale4F37D0(base int32, multiplier float32) int32 {
	product := float32(float64(base) * float64(multiplier))
	return pickupPotionFloatToInt4F37D0(product)
}

func pickupPotionNative4F37D0(
	owner, potion *Object,
	arg3, arg4 int32,
	deps pickupPotionNativeDeps4F37D0,
) int32 {
	return pickupPotion4F37D0(
		owner,
		potion,
		pickupPotionHooks4F37D0[
			*Object,
			*PotionUseData,
			*HealthData,
			*PlayerUpdateData,
			*Player,
		]{
			loadPotionUseData: func(potion *Object) *PotionUseData {
				return potion.UseData.AsPotion()
			},
			loadPotionValue: func(use *PotionUseData) int32 {
				return use.Value
			},
			gameFlag: deps.gameFlag,
			loadOwnerClassLow: func(owner *Object) uint8 {
				return uint8(owner.ObjClass)
			},
			loadOwnerUpdate: func(owner *Object) *PlayerUpdateData {
				return (*PlayerUpdateData)(owner.UpdateData)
			},
			loadUpdatePlayer: func(update *PlayerUpdateData) *Player {
				return update.Player
			},
			loadPlayerClass: func(player *Player) uint8 {
				// Direct indexing deliberately faults on a nil Player, unlike the
				// convenience PlayerClass method's nil-to-Warrior behavior.
				return player.info[66]
			},
			playerClassCanUse:   deps.playerClassCanUse,
			classFailureMessage: deps.classFailure,
			loadOwnerNetCode: func(owner *Object) uint32 {
				return owner.NetCode
			},
			audio:           deps.audio,
			loadPlayerState: deps.playerState,
			loadPotionSubClassLow: func(potion *Object) uint8 {
				return uint8(potion.ObjSubClass)
			},
			loadOwnerHealth: func(owner *Object) *HealthData {
				return owner.HealthData
			},
			loadHealthCur: func(health *HealthData) uint16 {
				return health.Cur
			},
			loadHealthMax: func(health *HealthData) uint16 {
				return health.Max
			},
			scaleHealth: func(base int32, class uint8) int32 {
				return pickupPotionScale4F37D0(base, deps.healthMultiplier(class))
			},
			adjustHealth: deps.adjustHealth,
			scaleMana: func(base int32, class uint8) int32 {
				return pickupPotionScale4F37D0(base, deps.manaMultiplier(class))
			},
			loadManaCur: func(update *PlayerUpdateData) uint16 {
				return update.ManaCur
			},
			loadManaMax: func(update *PlayerUpdateData) uint16 {
				return update.ManaMax
			},
			addMana: deps.addMana,
			loadOwnerPoison: func(owner *Object) uint8 {
				return owner.Poison540
			},
			removePoison:  deps.removePoison,
			spellAudio:    deps.spellAudio,
			delayedDelete: deps.delayedDelete,
			decay:         deps.decay,
			loadArg4: func() int32 {
				return arg4
			},
			loadArg3: func() int32 {
				return arg3
			},
			defaultPickup: deps.defaultPickup,
		},
	)
}

func pickupPotionServerDeps4F37D0(
	s *Server,
	runtime PickupPotionRuntime4F37D0,
) pickupPotionNativeDeps4F37D0 {
	return pickupPotionNativeDeps4F37D0{
		gameFlag: func(flag uint32) int32 {
			if noxflags.HasGame(noxflags.GameFlag(flag)) {
				return 1
			}
			return 0
		},
		playerClassCanUse: runtime.PlayerClassCanUse,
		classFailure: func(owner *Object, message string, value uint8) {
			s.NetPriMsgToPlayer(owner, strman.ID(message), value)
		},
		audio: func(id uint32, owner *Object, kind int32, code uint32) {
			s.Audio.EventObj(sound.ID(id), owner, int(kind), code)
		},
		playerState: func(owner *Object) int32 {
			if s.Players.CheckXxx(owner) {
				return 1
			}
			return 0
		},
		healthMultiplier: func(class uint8) float32 {
			return s.Players.ClassStatsMult(player.Class(class)).Health
		},
		adjustHealth: runtime.AdjustHealth,
		manaMultiplier: func(class uint8) float32 {
			return s.Players.ClassStatsMult(player.Class(class)).Mana
		},
		addMana: runtime.AddMana,
		removePoison: func(owner *Object) {
			s.RemovePoison4EE9D0(owner)
		},
		spellAudio: func(spellID, field int32) uint32 {
			return uint32(s.Spells.DefByInd(spell.ID(spellID)).GetAudio(int(field)))
		},
		delayedDelete: runtime.DelayedDelete,
		decay: func(potion *Object) {
			s.DecayRemove5116F0(potion)
		},
		defaultPickup: func(owner, potion *Object, arg3, arg4 int32) int32 {
			return s.PickupDefault4F31E0(owner, potion, arg3, arg4, runtime.DefaultPickup)
		},
	}
}

// PickupPotion4F37D0 binds GAME.EXE's registered four-argument PotionPickup
// callback to native-width Object, UpdateData, Player, HealthData, and
// PotionUseData pointers while preserving its exact signed int32 result.
func (s *Server) PickupPotion4F37D0(
	owner, potion *Object,
	arg3, arg4 int32,
	runtime PickupPotionRuntime4F37D0,
) int32 {
	return pickupPotionNative4F37D0(
		owner,
		potion,
		arg3,
		arg4,
		pickupPotionServerDeps4F37D0(s, runtime),
	)
}
