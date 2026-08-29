package legacy

/*
#include "GAME4.h"
*/
import "C"

import (
	"unsafe"

	noxflags "github.com/opennox/opennox/v1/common/flags"
	"github.com/opennox/opennox/v1/server"
)

type scriptCallbackInitNativeDeps4F5540 struct {
	readOnly      func() int32
	mapgenFile    func() unsafe.Pointer
	makeScript    func(unsafe.Pointer, *server.ScriptCallback) int32
	gameFlagCheck func(uint32) int32
	storeFunc     func(*server.ScriptCallback, int32)
}

func scriptCallbackInitRuntimeDeps4F5540() scriptCallbackInitNativeDeps4F5540 {
	return scriptCallbackInitNativeDeps4F5540{
		readOnly: nox_crypt_IsReadOnly,
		mapgenFile: func() unsafe.Pointer {
			return unsafe.Pointer(nox_xxx_mapgenGetSomeFile_426A60())
		},
		makeScript: func(file unsafe.Pointer, handler *server.ScriptCallback) int32 {
			return int32(C.nox_xxx_mapgenMakeScript_502790(
				(*C.FILE)(file),
				(*C.char)(unsafe.Pointer(handler)),
			))
		},
		gameFlagCheck: func(flag uint32) int32 {
			if noxflags.HasGame(noxflags.GameFlag(flag)) {
				return 1
			}
			return 0
		},
		storeFunc: func(handler *server.ScriptCallback, value int32) {
			handler.Func = value
		},
	}
}

func scriptCallbackInitNative4F5540(
	handler *server.ScriptCallback,
	deps scriptCallbackInitNativeDeps4F5540,
) int32 {
	return scriptCallbackInit4F5540(
		handler,
		scriptCallbackInitDeps4F5540[*server.ScriptCallback, unsafe.Pointer]{
			readOnly:      deps.readOnly,
			mapgenFile:    deps.mapgenFile,
			makeScript:    deps.makeScript,
			gameFlagCheck: deps.gameFlagCheck,
			storeFunc:     deps.storeFunc,
		},
	)
}

func scriptCallbackInitRuntime4F5540(handler *server.ScriptCallback) int32 {
	return scriptCallbackInitNative4F5540(handler, scriptCallbackInitRuntimeDeps4F5540())
}
