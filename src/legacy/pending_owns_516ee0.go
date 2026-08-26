package legacy

/*
#include <stdint.h>
*/
import "C"

import (
	"sync"

	"github.com/opennox/opennox/v1/server"
)

type pendingOwn516F90 struct {
	ownerScriptID int32
	ownedScriptID int32
}

// pendingOwnStore516EE0 replaces GAME.EXE's allocator-backed 12-byte linked
// records. The serialized IDs remain fixed-width, while the slice metadata and
// the Object links produced during resolution use the host pointer width.
type pendingOwnStore516EE0 struct {
	sync.Mutex
	active  bool
	entries []pendingOwn516F90
}

func (p *pendingOwnStore516EE0) alloc() bool {
	p.Lock()
	defer p.Unlock()
	p.entries = nil
	p.active = true
	return true
}

func (p *pendingOwnStore516EE0) free() int32 {
	p.Lock()
	defer p.Unlock()
	p.entries = nil
	p.active = false
	return 0
}

func (p *pendingOwnStore516EE0) clear() {
	p.Lock()
	defer p.Unlock()
	p.entries = nil
}

func (p *pendingOwnStore516EE0) add(ownerScriptID, ownedScriptID int32) bool {
	p.Lock()
	defer p.Unlock()
	if !p.active {
		return false
	}
	p.entries = append(p.entries, pendingOwn516F90{
		ownerScriptID: ownerScriptID,
		ownedScriptID: ownedScriptID,
	})
	return true
}

func (p *pendingOwnStore516EE0) take() []pendingOwn516F90 {
	p.Lock()
	defer p.Unlock()
	entries := p.entries
	p.entries = nil
	return entries
}

func resolvePendingOwns516FC0(entries []pendingOwn516F90, objectByScriptID func(int32) *server.Object, setOwner func(*server.Object, *server.Object)) {
	// GAME.EXE prepends each record and walks the resulting list from its head,
	// so ownership is restored in reverse insertion order.
	for i := len(entries) - 1; i >= 0; i-- {
		entry := entries[i]
		owner := objectByScriptID(entry.ownerScriptID)
		owned := objectByScriptID(entry.ownedScriptID)
		if owner != nil && owned != nil {
			setOwner(owner, owned)
		}
	}
}

var pendingOwns516EE0 pendingOwnStore516EE0

//export nox_xxx_pendingOwnAlloc_native_516EE0
func nox_xxx_pendingOwnAlloc_native_516EE0() C.int {
	if pendingOwns516EE0.alloc() {
		return 1
	}
	return 0
}

//export nox_xxx_pendingOwnFree_native_516F10
func nox_xxx_pendingOwnFree_native_516F10() C.int {
	return C.int(pendingOwns516EE0.free())
}

//export nox_xxx_pendingOwnClear_native_516F30
func nox_xxx_pendingOwnClear_native_516F30() {
	pendingOwns516EE0.clear()
}

//export nox_xxx_pendingOwnAdd_native_516F90
func nox_xxx_pendingOwnAdd_native_516F90(ownerScriptID, ownedScriptID C.int) C.int {
	if pendingOwns516EE0.add(int32(ownerScriptID), int32(ownedScriptID)) {
		return 1
	}
	return 0
}

//export nox_xxx_pendingOwnResolve_native_516FC0
func nox_xxx_pendingOwnResolve_native_516FC0() {
	srv := GetServer().S()
	resolvePendingOwns516FC0(pendingOwns516EE0.take(), srv.ObjectByScriptID4ECF10, srv.ObjSetOwner)
}
