package server

import (
	"github.com/opennox/libs/strman"

	"github.com/opennox/opennox/v1/common/sound"
)

// PickupGoldRuntime4F3A60 contains the scalar protection operation and the
// object-bearing callbacks still owned by the root or legacy boundary. Object
// and data pointers never cross an int-typed ABI.
type PickupGoldRuntime4F3A60 struct {
	DefaultPickup   PickupDefaultRuntime4F31E0
	ProtectGold     func(uint32, int32)
	DelayedDelete   func(*Object)
	SendLineMessage func(*Object, string, uint32)
}

type pickupGoldNativeDeps4F3A60 struct {
	protectGold     func(uint32, int32)
	delayedDelete   func(*Object)
	loadString      func(string, string, int) string
	sendLineMessage func(*Object, string, uint32)
	audio           func(uint32, *Object, int32, uint32)
	defaultPickup   func(*Object, *Object, int32, int32) int32
}

func playerAddGoldNative4FA590(
	owner *Object,
	amount uint32,
	protect func(uint32, int32),
) {
	playerAddGold4FA590(owner, amount, playerAddGoldHooks4FA590[
		*Object,
		*PlayerUpdateData,
		*Player,
	]{
		loadUpdate: func(owner *Object) *PlayerUpdateData {
			return (*PlayerUpdateData)(owner.UpdateData)
		},
		loadPlayer: func(update *PlayerUpdateData) *Player {
			return update.Player
		},
		loadGold: func(player *Player) uint32 {
			return player.GoldVal
		},
		storeGold: func(player *Player, value uint32) {
			player.GoldVal = value
		},
		loadToken: func(player *Player) uint32 {
			return player.ProtPlayerGold
		},
		protect: protect,
	})
}

func pickupGoldNative4F3A60(
	owner, item *Object,
	arg3, arg4 int32,
	deps pickupGoldNativeDeps4F3A60,
) int32 {
	return pickupGold4F3A60(owner, item, pickupGoldHooks4F3A60[
		*Object,
		*GoldInitData,
		string,
	]{
		loadOwnerClassLow: func(owner *Object) uint8 {
			return uint8(owner.ObjClass)
		},
		loadGoldInitData: func(item *Object) *GoldInitData {
			return (*GoldInitData)(item.InitData)
		},
		loadGoldAmount: func(data *GoldInitData) uint32 {
			return data.Amount
		},
		addGold: func(owner *Object, amount uint32) {
			playerAddGoldNative4FA590(owner, amount, deps.protectGold)
		},
		delayedDelete:   deps.delayedDelete,
		loadString:      deps.loadString,
		sendLineMessage: deps.sendLineMessage,
		audio:           deps.audio,
		loadArg4: func() int32 {
			return arg4
		},
		loadArg3: func() int32 {
			return arg3
		},
		defaultPickup: deps.defaultPickup,
	})
}

func pickupGoldServerDeps4F3A60(
	s *Server,
	runtime PickupGoldRuntime4F3A60,
) pickupGoldNativeDeps4F3A60 {
	return pickupGoldNativeDeps4F3A60{
		protectGold:   runtime.ProtectGold,
		delayedDelete: runtime.DelayedDelete,
		loadString: func(key, path string, line int) string {
			_ = line // retained by the generic provenance contract
			return s.Strings().GetStringInFile(strman.ID(key), path)
		},
		sendLineMessage: runtime.SendLineMessage,
		audio: func(id uint32, owner *Object, kind int32, code uint32) {
			s.Audio.EventObj(sound.ID(id), owner, int(kind), code)
		},
		defaultPickup: func(owner, item *Object, arg3, arg4 int32) int32 {
			return s.PickupDefault4F31E0(owner, item, arg3, arg4, runtime.DefaultPickup)
		},
	}
}

// PickupGold4F3A60 binds GAME.EXE's registered four-argument GoldPickup
// callback and its gold-add dependency to native-width Object,
// PlayerUpdateData, Player, and GoldInitData pointers.
func (s *Server) PickupGold4F3A60(
	owner, item *Object,
	arg3, arg4 int32,
	runtime PickupGoldRuntime4F3A60,
) int32 {
	return pickupGoldNative4F3A60(
		owner,
		item,
		arg3,
		arg4,
		pickupGoldServerDeps4F3A60(s, runtime),
	)
}
