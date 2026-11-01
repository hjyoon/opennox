package server

import (
	"unsafe"

	"github.com/opennox/libs/types"
)

// PlayerBotRuntime4FAB20 supplies native world services for bot updates and
// respawn while sharing the independently compared constructor services.
type PlayerBotRuntime4FAB20 struct {
	Create              PlayerBotCreateRuntime4FA700
	Frame               func() uint32
	LoadSettings        func() *Settings
	NetworkMode         func() uint32
	DefaultItems        func(*Object, int32, int32)
	RespawnCorpse       func(*types.Pointf, int16)
	FindStart           func(*types.Pointf, *Object)
	Move                func(*Object, *types.Pointf)
	Audio               func(int32, *Object, int32, int32)
	GameFlag            func(uint32) bool
	Buff                func(*Object, int32, int16, int8)
	UpdateMonster       func(*Object)
	RegularPlayerUpdate unsafe.Pointer
}
