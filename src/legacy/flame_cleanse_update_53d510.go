package legacy

/*
#include "defs.h"
#include "GAME4_3.h"
*/
import "C"

import (
	"unsafe"

	"github.com/opennox/opennox/v1/server"
)

func Get_nox_xxx_updateFlameCleanse_53D510() unsafe.Pointer {
	return C.nox_xxx_updateFlameCleanse_53D510
}

func flameCleanseUpdateNative53D510(obj *server.Object) {
	outer := GetServer()
	outer.S().FlameCleanseUpdate53D510(obj, server.FlameCleanseUpdateRuntime53D510{
		DelayedDelete: outer.DelayedDelete,
	})
}

var flameCleanseUpdateCall53D510 = flameCleanseUpdateNative53D510

func init() {
	// The cast assigns this callback dynamically; retain its legacy identity
	// while dispatching Object.CallUpdate through the native implementation.
	server.RegisterObjectUpdateGo("FlameCleanseUpdate", C.nox_xxx_updateFlameCleanse_53D510, func(obj *server.Object) {
		flameCleanseUpdateCall53D510(obj)
	}, 0)
}
