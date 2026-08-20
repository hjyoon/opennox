package server

import (
	"github.com/opennox/libs/object"
	"github.com/opennox/libs/types"
)

// RespawnInventoryClassRuntime4EDD00 supplies the two services still owned by
// the outer server/legacy layers. Random reachable-point selection and every
// Object field access remain in native-width server code.
type RespawnInventoryClassRuntime4EDD00 struct {
	Detach   func(*Object, *Object)
	CreateAt func(*Object, *Object, types.Pointf)
}

func respawnInventoryClassServerDeps4EDD00(
	s *Server,
	runtime RespawnInventoryClassRuntime4EDD00,
) respawnInventoryClassNativeDeps4EDD00 {
	return respawnInventoryClassNativeDeps4EDD00{
		detachInventory: runtime.Detach,
		randomReachable: s.RandomReachablePointAroundInto4ED970,
		createAt:        runtime.CreateAt,
	}
}

// RespawnInventoryClass4EDD00 detaches every item whose full class dword
// overlaps classMask and recreates it at an exact radius-60 reachable point.
func (s *Server) RespawnInventoryClass4EDD00(
	owner *Object,
	classMask object.Class,
	runtime RespawnInventoryClassRuntime4EDD00,
) {
	respawnInventoryClassNative4EDD00(
		owner,
		uint32(classMask),
		respawnInventoryClassServerDeps4EDD00(s, runtime),
	)
}

// DropPlayerInventoryClassRuntime4EDD70 supplies the restored 004ED790
// dispatcher. Player iteration, inventory fields, and random-point selection
// are owned by Server.
type DropPlayerInventoryClassRuntime4EDD70 struct {
	Dispatch func(*Object, *Object, *types.Pointf) int32
}

func dropPlayerInventoryClassServerDeps4EDD70(
	s *Server,
	runtime DropPlayerInventoryClassRuntime4EDD70,
) dropPlayerInventoryClassNativeDeps4EDD70 {
	return dropPlayerInventoryClassNativeDeps4EDD70{
		firstPlayer: s.Players.FirstUnit,
		nextPlayer:  s.questNextPlayerUnit4DA7F0,
		randomReachable: func(radius float32, center, output *types.Pointf) *types.Pointf {
			return s.RandomReachablePointAroundInto4ED970(radius, center, output)
		},
		drop: runtime.Dispatch,
	}
}

// DropPlayerInventoryClass4EDD70 drops every direct player inventory item
// carrying class bit 0x10000000 at an exact radius-50 reachable point.
func (s *Server) DropPlayerInventoryClass4EDD70(
	runtime DropPlayerInventoryClassRuntime4EDD70,
) {
	dropPlayerInventoryClassNative4EDD70(
		dropPlayerInventoryClassServerDeps4EDD70(s, runtime),
	)
}
