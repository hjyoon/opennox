package server

const (
	warpReadUsePlayerClass53F830 = uint8(0x04)
	warpReadUseInformCode53F830  = uint8(21)
	warpReadUseClosedKey53F830   = "GeneralPrint:WarpClosed"
)

type warpReadUseHooks53F830[O, U, D, P any] struct {
	loadOwnerArg       func() O
	loadClassLow       func(O) uint8
	loadReadableArg    func() O
	loadFPS            func() uint32
	loadUpdateData     func(O) U
	loadUseData        func(O) D
	loadFrame          func() uint32
	loadReadState      func(D) uint32
	mapCheck           func(O, O) int32
	warpEnabled        func() int32
	currentQuestStage  func() int32
	nextStageThreshold func(int32) int32
	loadPlayer         func(U) P
	loadPlayerIndex    func(P) uint8
	informText         func(uint8, uint8, int32)
	priorityMessage    func(O, string, uint8)
	storeReadState     func(D, uint32)
}

// warpReadUse53F830 preserves GAME.EXE 0053F830. The Player class byte is
// the first dereference. On the Player path, the readable argument, FPS,
// owner UpdateData, readable UseData, frame and trailing read-state dword are
// observed in original instruction order. UpdateData and UseData remain
// cached across all callbacks.
//
// The cooldown subtraction and three-times-FPS product are uint32 operations;
// a zero state bypasses the cooldown. Visibility must return exactly one.
// When Quest warps are open, the current stage is rounded to its next
// threshold, then the Player pointer and index are loaded from the cached
// UpdateData and informed with code 21. Otherwise, the owner receives the
// exact GeneralPrint:WarpClosed priority message. After either message, the
// live frame is loaded again and stored through the cached UseData pointer.
// Every path returns canonical one.
func warpReadUse53F830[O, U, D, P any](
	hooks warpReadUseHooks53F830[O, U, D, P],
) int32 {
	owner := hooks.loadOwnerArg()
	if hooks.loadClassLow(owner)&warpReadUsePlayerClass53F830 == 0 {
		return 1
	}

	readable := hooks.loadReadableArg()
	fps := hooks.loadFPS()
	update := hooks.loadUpdateData(owner)
	data := hooks.loadUseData(readable)
	frame := hooks.loadFrame()
	state := hooks.loadReadState(data)
	if state != 0 && frame-state <= 3*fps {
		return 1
	}
	if hooks.mapCheck(owner, readable) != 1 {
		return 1
	}

	if hooks.warpEnabled() != 0 {
		stage := hooks.currentQuestStage()
		threshold := hooks.nextStageThreshold(stage)
		player := hooks.loadPlayer(update)
		index := hooks.loadPlayerIndex(player)
		hooks.informText(index, warpReadUseInformCode53F830, threshold)
	} else {
		hooks.priorityMessage(owner, warpReadUseClosedKey53F830, 1)
	}
	frame = hooks.loadFrame()
	hooks.storeReadState(data, frame)
	return 1
}
