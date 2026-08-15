package opennox

import (
	"github.com/opennox/libs/object"

	"github.com/opennox/opennox/v1/server"
)

const playerObserverRejectedSlaveFlags_4E6280 = object.Flags(0x8020)

// playerObserverFindGoodSlave0_4E6280 preserves the original two-pass search.
// The camera is read once, while PlayerUnit is reloaded before the fallback
// pass so mutations made by the first-pass callbacks remain observable.
func playerObserverFindGoodSlave0_4E6280(
	pl *server.Player,
	firstGoodSlave func(*server.Object) *server.Object,
	nextGoodSlave func(*server.Object) *server.Object,
) *server.Object {
	camera := pl.CameraFollowObj
	var candidate *server.Object
	if camera != nil {
		candidate = nextGoodSlave(camera)
	} else {
		candidate = firstGoodSlave(pl.PlayerUnit)
	}
	if candidate = playerObserverUsableSlave_4E6280(candidate, nextGoodSlave); candidate != nil {
		return candidate
	}
	return playerObserverUsableSlave_4E6280(firstGoodSlave(pl.PlayerUnit), nextGoodSlave)
}

func playerObserverUsableSlave_4E6280(
	candidate *server.Object,
	nextGoodSlave func(*server.Object) *server.Object,
) *server.Object {
	for candidate != nil && candidate.ObjFlags&playerObserverRejectedSlaveFlags_4E6280 != 0 {
		candidate = nextGoodSlave(candidate)
	}
	return candidate
}

// playerObserverFindGoodSlave2_4EC3E0 and
// playerObserverFindGoodSlave_4EC420 are the two original traversal
// dependencies of 004E6280. They use native object links so 004E6280 does not
// reintroduce an object-pointer-to-int CGo boundary.
func playerObserverFindGoodSlave2_4EC3E0(owner *server.Object) *server.Object {
	return playerObserverFindGoodSlave2Contract4EC3E0(owner, playerObserverFindGoodSlave2Hooks4EC3E0[
		*server.Object,
		*server.MonsterUpdateData,
	]{
		loadFirstOwned: func(owner *server.Object) *server.Object {
			return owner.Field129
		},
		loadClassByte: func(candidate *server.Object) uint8 {
			return uint8(candidate.ObjClass)
		},
		loadUpdateData: func(candidate *server.Object) *server.MonsterUpdateData {
			return (*server.MonsterUpdateData)(candidate.UpdateData)
		},
		loadStatusByte: func(data *server.MonsterUpdateData) uint8 {
			return uint8(data.StatusFlags)
		},
		loadNextOwned: func(candidate *server.Object) *server.Object {
			return candidate.Field128
		},
	})
}

func playerObserverFindGoodSlave_4EC420(current *server.Object) *server.Object {
	return playerObserverFindGoodSlaveContract4EC420(current, playerObserverFindGoodSlaveHooks4EC420[
		*server.Object,
		*server.MonsterUpdateData,
	]{
		loadOwner: func(current *server.Object) *server.Object {
			return current.ObjOwner
		},
		loadNextOwned: func(obj *server.Object) *server.Object {
			return obj.Field128
		},
		loadClassByte: func(candidate *server.Object) uint8 {
			return uint8(candidate.ObjClass)
		},
		loadUpdateData: func(candidate *server.Object) *server.MonsterUpdateData {
			return (*server.MonsterUpdateData)(candidate.UpdateData)
		},
		loadStatusByte: func(data *server.MonsterUpdateData) uint8 {
			return uint8(data.StatusFlags)
		},
	})
}
