//go:build 386 || arm

package server

import "unsafe"

//go:nocheckptr
func moverWaypointFromPE32(value uint32) *Waypoint {
	return (*Waypoint)(unsafe.Pointer(uintptr(value)))
}

//go:nocheckptr
func moverObjectFromPE32(value uint32) *Object {
	return (*Object)(unsafe.Pointer(uintptr(value)))
}
