package legacy

/*
#include <stdint.h>
*/
import "C"

import (
	"math"
	"sync"
)

type mapgenPointerRegistry struct {
	mu        sync.Mutex
	nextToken uint32
	byAddress map[uintptr]uint32
	byToken   map[uint32]uintptr
}

var mapgenPointers = mapgenPointerRegistry{nextToken: 0xf0000000}

func (r *mapgenPointerRegistry) register(address uintptr) uint32 {
	if address == 0 {
		return 0
	}
	r.mu.Lock()
	defer r.mu.Unlock()
	if token := r.byAddress[address]; token != 0 {
		return token
	}
	token := uint32(address)
	if token == 0 || r.byToken[token] != 0 {
		for {
			token = r.nextToken
			r.nextToken++
			if token != 0 && r.byToken[token] == 0 {
				break
			}
		}
	}
	if r.byAddress == nil {
		r.byAddress = make(map[uintptr]uint32)
		r.byToken = make(map[uint32]uintptr)
	}
	r.byAddress[address] = token
	r.byToken[token] = address
	return token
}

func (r *mapgenPointerRegistry) resolve(value uintptr) uintptr {
	if value > math.MaxUint32 {
		return value
	}
	r.mu.Lock()
	address := r.byToken[uint32(value)]
	r.mu.Unlock()
	if address != 0 {
		return address
	}
	return value
}

func (r *mapgenPointerRegistry) forget(address uintptr) {
	r.mu.Lock()
	defer r.mu.Unlock()
	if token := r.byAddress[address]; token != 0 {
		delete(r.byAddress, address)
		delete(r.byToken, token)
	}
}

//export nox_mapgen_register_address_go
func nox_mapgen_register_address_go(address C.uintptr_t) C.uint32_t {
	return C.uint32_t(mapgenPointers.register(uintptr(address)))
}

//export nox_mapgen_resolve_address_go
func nox_mapgen_resolve_address_go(value C.uintptr_t) C.uintptr_t {
	return C.uintptr_t(mapgenPointers.resolve(uintptr(value)))
}

//export nox_mapgen_forget_address_go
func nox_mapgen_forget_address_go(address C.uintptr_t) {
	mapgenPointers.forget(uintptr(address))
}
