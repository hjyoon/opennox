package server

import (
	"unsafe"

	"github.com/opennox/libs/strman"

	"github.com/opennox/opennox/v1/common/ntype"
)

type warpReadUseNativeDeps53F830 struct {
	loadFPS            func() uint32
	loadFrame          func() uint32
	mapCheck           func(*Object, *Object) int32
	warpEnabled        func() int32
	currentQuestStage  func() int32
	nextStageThreshold func(int32) int32
	informText         func(uint8, uint8, int32)
	priorityMessage    func(*Object, string, uint8)
}

// WarpReadUseRuntime53F830 supplies the Quest state whose storage remains in
// the legacy runtime. All object, update-data, use-data and player pointers
// stay native-width in the server implementation.
type WarpReadUseRuntime53F830 struct {
	WarpEnabled        func() int32
	CurrentQuestStage  func() int32
	NextStageThreshold func(int32) int32
}

func warpReadUseNative53F830(
	owner, readable *Object,
	deps warpReadUseNativeDeps53F830,
) int32 {
	return warpReadUse53F830(warpReadUseHooks53F830[
		*Object,
		*PlayerUpdateData,
		*ReadableUseData,
		*Player,
	]{
		loadOwnerArg: func() *Object {
			return owner
		},
		loadClassLow: func(owner *Object) uint8 {
			return uint8(owner.ObjClass)
		},
		loadReadableArg: func() *Object {
			return readable
		},
		loadFPS: deps.loadFPS,
		loadUpdateData: func(owner *Object) *PlayerUpdateData {
			return (*PlayerUpdateData)(owner.UpdateData)
		},
		loadUseData: func(readable *Object) *ReadableUseData {
			return readable.UseData.AsReadable()
		},
		loadFrame: deps.loadFrame,
		loadReadState: func(data *ReadableUseData) uint32 {
			return data.TransientReadState
		},
		mapCheck:           deps.mapCheck,
		warpEnabled:        deps.warpEnabled,
		currentQuestStage:  deps.currentQuestStage,
		nextStageThreshold: deps.nextStageThreshold,
		loadPlayer: func(update *PlayerUpdateData) *Player {
			return update.Player
		},
		loadPlayerIndex: func(player *Player) uint8 {
			return player.PlayerInd
		},
		informText:      deps.informText,
		priorityMessage: deps.priorityMessage,
		storeReadState: func(data *ReadableUseData, frame uint32) {
			data.TransientReadState = frame
		},
	})
}

func warpReadUseServerDeps53F830(
	s *Server,
	runtime WarpReadUseRuntime53F830,
) warpReadUseNativeDeps53F830 {
	return warpReadUseNativeDeps53F830{
		loadFPS:   s.TickRate,
		loadFrame: s.Frame,
		mapCheck: func(owner, readable *Object) int32 {
			if s.MapTraceVision(owner, readable) {
				return 1
			}
			return 0
		},
		warpEnabled:        runtime.WarpEnabled,
		currentQuestStage:  runtime.CurrentQuestStage,
		nextStageThreshold: runtime.NextStageThreshold,
		informText: func(index, code uint8, value int32) {
			s.NetInformTextMsg(ntype.PlayerInd(index), code, int(value))
		},
		priorityMessage: func(owner *Object, key string, value uint8) {
			s.NetPriMsgToPlayer(owner, strman.ID(key), value)
		},
	}
}

// WarpReadUse53F830 binds GAME.EXE 0053F830 to native-width Object,
// PlayerUpdateData, ReadableUseData and Player pointers. It is the registered
// WarpReadUse implementation reached through SignCollide and other native use
// dispatchers.
func (s *Server) WarpReadUse53F830(
	owner, readable *Object,
	runtime WarpReadUseRuntime53F830,
) int32 {
	return warpReadUseNative53F830(
		owner,
		readable,
		warpReadUseServerDeps53F830(s, runtime),
	)
}

var (
	_ = [1]struct{}{}[unsafe.Sizeof(uintptr(0))-unsafe.Sizeof(Object{}.UseData)]
	_ = [1]struct{}{}[260-unsafe.Sizeof(ReadableUseData{})]
	_ = [1]struct{}{}[256-unsafe.Offsetof(ReadableUseData{}.TransientReadState)]
)
