package legacy

/*
#include "game2_audio_native.h"
extern void* dword_587000_81128;
*/
import "C"

import "unsafe"

// Game2AudioBridgeState is a read-only test/audit snapshot. All addresses refer
// to the real C-owned service, not parallel fixture owner arrays. Callers compare
// addresses only; the snapshot never extends the lifetime of a freed C owner.
type Game2AudioBridgeState struct {
	Initialized                              bool
	Context, ContextMaster, Master, Bank     unsafe.Pointer
	OrdinaryOwner                            unsafe.Pointer
	Driver                                   uintptr
	Events, Voices, Banks, Contexts, Samples int
	DefinitionPresent                        bool
	Definition                               [200]byte
}

func Game2AudioBridgeSnapshot(id int) Game2AudioBridgeState {
	game2AudioMu.Lock()
	defer game2AudioMu.Unlock()
	s := &game2Audio
	v := Game2AudioBridgeState{OrdinaryOwner: C.dword_587000_81128, Initialized: s.Initialized, Master: unsafe.Pointer(s.Master), Events: len(s.Live), Banks: len(game2AudioBanks), Contexts: len(game2AudioContexts), Samples: len(game2AudioSamples)}
	if s.Context != nil {
		v.Context = unsafe.Pointer(s.Context)
		v.ContextMaster = unsafe.Pointer(s.Context.EventMaster)
		v.Driver = s.Context.Driver
	}
	if s.Bank != nil {
		v.Bank = s.Bank.Token
		v.Voices = len(s.Bank.Voices)
	}
	if s.Initialized && id >= 0 && id < len(s.Definitions) {
		d := &s.Definitions[id]
		if d.Pointer != nil {
			copy(v.Definition[:], unsafe.Slice((*byte)(d.Pointer), 200))
			v.DefinitionPresent = true
		}
	}
	return v
}

// The ordinary Go API has no wrapper for this existing public C allocation
// entry. This bridge invokes that actual C→Go path, without synthesizing an
// event, mutating its algorithm or using private engine state to create one.
func Game2AudioBridgeCreateEvent(id int) unsafe.Pointer {
	p := C.nox_xxx_draw_452270(C.int(id))
	return C.nox_xxx_draw_452300(p)
}
