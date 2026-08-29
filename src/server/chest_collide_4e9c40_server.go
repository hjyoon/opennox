package server

import (
	"unsafe"

	"github.com/opennox/libs/strman"

	noxflags "github.com/opennox/opennox/v1/common/flags"
	"github.com/opennox/opennox/v1/common/sound"
)

// ChestCollideRuntime4E9C40 supplies shared legacy state and the two large
// item-drop effects reached after this callback. Those effects retain their
// own restoration scope; this type keeps their object pointers native-width.
type ChestCollideRuntime4E9C40 struct {
	Ticks              func() uint64
	LoadFeedbackTicks  func() uint64
	StoreFeedbackTicks func(uint64)
	DelayedDelete      func(*Object)
	ChestOpen          func(*Object, *Object)
	DropAllItems       func(*Object)
}

type chestCollideNativeDeps4E9C40 struct {
	gameFlagsCheck     func(uint32) int32
	loadTypeName       func(*Object) string
	delayedDelete      func(*Object)
	audio              func(uint32, *Object)
	ticks              func() uint64
	loadFeedbackTicks  func() uint64
	priorityMessage    func(*Object, string)
	storeFeedbackTicks func(uint64)
	callDeath          func(unsafe.Pointer, *Object)
	chestOpen          func(*Object, *Object)
	dropAllItems       func(*Object)
}

func chestCollideNative4E9C40(
	source, target *Object,
	collision unsafe.Pointer,
	deps chestCollideNativeDeps4E9C40,
) {
	chestCollide4E9C40(source, target, collision, chestCollideHooks4E9C40[*Object, unsafe.Pointer]{
		loadClassLow: func(obj *Object) uint8 {
			return uint8(obj.ObjClass)
		},
		loadFlags: func(obj *Object) uint32 {
			return uint32(obj.ObjFlags)
		},
		gameFlagsCheck: deps.gameFlagsCheck,
		loadSubclass: func(obj *Object) uint32 {
			return uint32(obj.ObjSubClass)
		},
		firstItem: func(obj *Object) *Object {
			return obj.InvFirstItem
		},
		loadTypeName: deps.loadTypeName,
		nextItem: func(obj *Object) *Object {
			return obj.InvNextItem
		},
		delayedDelete:      deps.delayedDelete,
		audio:              deps.audio,
		ticks:              deps.ticks,
		loadFeedbackTicks:  deps.loadFeedbackTicks,
		priorityMessage:    deps.priorityMessage,
		storeFeedbackTicks: deps.storeFeedbackTicks,
		loadDeath: func(obj *Object) unsafe.Pointer {
			return obj.Death
		},
		callDeath:    deps.callDeath,
		chestOpen:    deps.chestOpen,
		dropAllItems: deps.dropAllItems,
	})
}

// ChestCollide4E9C40 binds the zero-byte registered callback to native-width
// Object, inventory, Death callback, and shared feedback state.
func (s *Server) ChestCollide4E9C40(
	source, target *Object,
	collision unsafe.Pointer,
	runtime ChestCollideRuntime4E9C40,
) {
	chestCollideNative4E9C40(source, target, collision, chestCollideNativeDeps4E9C40{
		gameFlagsCheck: func(flag uint32) int32 {
			if noxflags.HasGame(noxflags.GameFlag(flag)) {
				return 1
			}
			return 0
		},
		loadTypeName: func(obj *Object) string {
			return s.Types.ByInd(int(uint16(obj.TypeInd))).ID()
		},
		delayedDelete: runtime.DelayedDelete,
		audio: func(id uint32, obj *Object) {
			s.Audio.EventObj(sound.ID(id), obj, 0, 0)
		},
		ticks:             runtime.Ticks,
		loadFeedbackTicks: runtime.LoadFeedbackTicks,
		priorityMessage: func(obj *Object, message string) {
			s.NetPriMsgToPlayer(obj, strman.ID(message), 0)
		},
		storeFeedbackTicks: runtime.StoreFeedbackTicks,
		callDeath: func(death unsafe.Pointer, obj *Object) {
			CallObjectDeath(death, obj)
		},
		chestOpen:    runtime.ChestOpen,
		dropAllItems: runtime.DropAllItems,
	})
}
