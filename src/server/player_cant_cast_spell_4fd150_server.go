package server

import (
	"unsafe"

	"github.com/opennox/libs/spell"
	"github.com/opennox/libs/things"

	noxflags "github.com/opennox/opennox/v1/common/flags"
	"github.com/opennox/opennox/v1/common/memmap"
)

const (
	playerCantCastGlobalsBase4FD150 = uintptr(0x5D4594)

	playerCantCastPixieTypeOffset4FD150       = uintptr(1569676)
	playerCantCastMagicMissileOffset4FD150    = uintptr(1569680)
	playerCantCastSmallFistOffset4FD150       = uintptr(1569684)
	playerCantCastMediumFistOffset4FD150      = uintptr(1569688)
	playerCantCastLargeFistOffset4FD150       = uintptr(1569692)
	playerCantCastDeathBallOffset4FD150       = uintptr(1569696)
	playerCantCastMeteorOffset4FD150          = uintptr(1569700)
	playerCantCastCrownTypeOffset4FD150       = uintptr(1569704)
	playerCantCastGameBallTypeOffset4FD150    = uintptr(1569708)
	playerCantCastImaginaryCasterOffset4FE7B0 = uintptr(1569720)
	playerSpellPowerFullStrengthModes4FE7B0   = spellPowerFullStrengthModes4FE7B0
	playerCantCastQuestMode4FD150             = uint32(4096)
)

type playerSpellPowerNativeDeps4FE7B0 struct {
	loadImaginaryCasterType  func() uint32
	storeImaginaryCasterType func(uint32)
	lookupObjectType         func(string) uint32
	hasGameFlag              func(uint32) int32
}

// playerSpellPowerNative4FE7B0 binds the exact GAME.EXE 004FE7B0 operation to
// native-width Object, update-data, and Player pointers. In particular, the
// caster TypeInd load intentionally precedes both the game-mode and nil gates.
func playerSpellPowerNative4FE7B0(
	spellID int32,
	obj *Object,
	deps playerSpellPowerNativeDeps4FE7B0,
) int32 {
	return spellPower4FE7B0(spellPowerHooks4FE7B0[*Object, unsafe.Pointer, *Player]{
		loadImaginaryCasterType:  deps.loadImaginaryCasterType,
		lookupObjectType:         deps.lookupObjectType,
		storeImaginaryCasterType: deps.storeImaginaryCasterType,
		loadCasterArg: func() *Object {
			return obj
		},
		loadCasterType: func(obj *Object) uint16 {
			return obj.TypeInd
		},
		hasGameFlag: deps.hasGameFlag,
		loadClass: func(obj *Object) uint32 {
			return uint32(obj.ObjClass)
		},
		loadUpdate: func(obj *Object) unsafe.Pointer {
			return obj.UpdateData
		},
		loadSpellArg: func() int32 {
			return spellID
		},
		loadPlayer: func(update unsafe.Pointer) *Player {
			return (*PlayerUpdateData)(update).Player
		},
		loadPlayerPower: func(player *Player, spellID int32) int32 {
			return int32(player.SpellLvl[spellID])
		},
		loadMonsterPower: func(update unsafe.Pointer) int32 {
			return int32((*MonsterUpdateData)(update).Field510)
		},
	})
}

type playerCantCastSpellNativeDeps4FD150 struct {
	hasGameFlag      func(uint32) int32
	loadGlobal       func(uintptr) uint32
	storeGlobal      func(uintptr, uint32)
	lookupObjectType func(string) uint32
	spellHasFlags    func(int32, uint32) int32
	spellPower       func(int32, *Object) int32
	balanceFloat     func(string, int32) float64
}

func playerCantCastSummonOffset4FD150(kind playerCantCastSummonType4FD150) uintptr {
	switch kind {
	case playerCantCastPixie4FD150:
		return playerCantCastPixieTypeOffset4FD150
	case playerCantCastMagicMissile4FD150:
		return playerCantCastMagicMissileOffset4FD150
	case playerCantCastSmallFist4FD150:
		return playerCantCastSmallFistOffset4FD150
	case playerCantCastMediumFist4FD150:
		return playerCantCastMediumFistOffset4FD150
	case playerCantCastLargeFist4FD150:
		return playerCantCastLargeFistOffset4FD150
	case playerCantCastDeathBall4FD150:
		return playerCantCastDeathBallOffset4FD150
	case playerCantCastMeteor4FD150:
		return playerCantCastMeteorOffset4FD150
	default:
		panic("unknown 4FD150 summon type")
	}
}

func playerCantCastSpellNative4FD150(
	unit *Object,
	spellID int32,
	bypassModeRules int32,
	deps playerCantCastSpellNativeDeps4FD150,
) int32 {
	return playerCantCastSpell4FD150(unit, spellID, bypassModeRules, playerCantCastSpellHooks4FD150[
		*Object, *Object,
	]{
		findParent:  (*Object).FindOwnerChainPlayer,
		hasGameFlag: deps.hasGameFlag,
		loadCrownTypeCache: func() uint32 {
			return deps.loadGlobal(playerCantCastCrownTypeOffset4FD150)
		},
		storeCrownTypeCache: func(value uint32) {
			deps.storeGlobal(playerCantCastCrownTypeOffset4FD150, value)
		},
		loadGameBallTypeCache: func() uint32 {
			return deps.loadGlobal(playerCantCastGameBallTypeOffset4FD150)
		},
		storeGameBallTypeCache: func(value uint32) {
			deps.storeGlobal(playerCantCastGameBallTypeOffset4FD150, value)
		},
		lookupObjectType: deps.lookupObjectType,
		spellHasFlags:    deps.spellHasFlags,
		loadFirstOwned: func(obj *Object) *Object {
			return obj.Field129
		},
		loadOwnedType: func(obj *Object) uint16 {
			return obj.TypeInd
		},
		loadNextOwned: func(obj *Object) *Object {
			return obj.Field128
		},
		hasTeam: func(obj *Object) int32 {
			if obj.HasTeam() {
				return 1
			}
			return 0
		},
		loadFirstInventory: func(obj *Object) *Object {
			return obj.InvFirstItem
		},
		loadInventoryFlags: func(obj *Object) uint32 {
			return uint32(obj.ObjFlags)
		},
		loadNextInventory: func(obj *Object) *Object {
			return obj.InvNextItem
		},
		hasEnchant: func(obj *Object, enchant uint8) int32 {
			if obj.HasEnchant(EnchantID(enchant)) {
				return 1
			}
			return 0
		},
		spellAllowed: func(spellID int32) int32 {
			if deps.hasGameFlag(playerCantCastQuestMode4FD150) != 0 ||
				(spellID != 111 && spellID != 112 && spellID != 114 && spellID != 113) {
				return 1
			}
			return 0
		},
		loadSummonType: func(kind playerCantCastSummonType4FD150) uint32 {
			return deps.loadGlobal(playerCantCastSummonOffset4FD150(kind))
		},
		countOwnedType: func(obj *Object, typeInd uint32) int32 {
			return obj.CountSubOfType(int32(typeInd))
		},
		spellPower:   deps.spellPower,
		balanceFloat: deps.balanceFloat,
	})
}

func playerSpellPowerServerDeps4FE7B0(s *Server) playerSpellPowerNativeDeps4FE7B0 {
	return playerSpellPowerNativeDeps4FE7B0{
		loadImaginaryCasterType: func() uint32 {
			return memmap.Uint32(playerCantCastGlobalsBase4FD150, playerCantCastImaginaryCasterOffset4FE7B0)
		},
		storeImaginaryCasterType: func(value uint32) {
			*memmap.PtrUint32(playerCantCastGlobalsBase4FD150, playerCantCastImaginaryCasterOffset4FE7B0) = value
		},
		lookupObjectType: func(name string) uint32 {
			return uint32(s.Types.IndByID(name))
		},
		hasGameFlag: func(mask uint32) int32 {
			if noxflags.HasGame(noxflags.GameFlag(mask)) {
				return 1
			}
			return 0
		},
	}
}

// SpellPower4FE7B0 returns the exact spell-power dword selected by
// GAME.EXE 004FE7B0 without routing native pointers through legacy C.
//
//go:noinline
func (s *Server) SpellPower4FE7B0(spellID spell.ID, obj *Object) int32 {
	return playerSpellPowerNative4FE7B0(int32(spellID), obj, playerSpellPowerServerDeps4FE7B0(s))
}

func playerCantCastSpellServerDeps4FD150(s *Server) playerCantCastSpellNativeDeps4FD150 {
	powerDeps := playerSpellPowerServerDeps4FE7B0(s)
	return playerCantCastSpellNativeDeps4FD150{
		hasGameFlag: powerDeps.hasGameFlag,
		loadGlobal: func(offset uintptr) uint32 {
			return memmap.Uint32(playerCantCastGlobalsBase4FD150, offset)
		},
		storeGlobal: func(offset uintptr, value uint32) {
			*memmap.PtrUint32(playerCantCastGlobalsBase4FD150, offset) = value
		},
		lookupObjectType: powerDeps.lookupObjectType,
		spellHasFlags: func(spellID int32, flags uint32) int32 {
			if s.Spells.HasFlags(spell.ID(spellID), things.SpellFlags(flags)) {
				return 1
			}
			return 0
		},
		spellPower: func(spellID int32, obj *Object) int32 {
			return playerSpellPowerNative4FE7B0(spellID, obj, powerDeps)
		},
		balanceFloat: func(key string, index int32) float64 {
			return s.Balance.FloatInd(key, int(index))
		},
	}
}

// CheckPlayerCantCastSpell4FD150 returns the original numeric cast rejection
// code while retaining every Object and update-data pointer at native width.
//
//go:noinline
func (s *Server) CheckPlayerCantCastSpell4FD150(unit *Object, spellID spell.ID, bypassModeRules int) int32 {
	return playerCantCastSpellNative4FD150(unit, int32(spellID), int32(bypassModeRules), playerCantCastSpellServerDeps4FD150(s))
}
