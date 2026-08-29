package legacy

import (
	"unsafe"

	noxflags "github.com/opennox/opennox/v1/common/flags"
	"github.com/opennox/opennox/v1/internal/cryptfile"
	"github.com/opennox/opennox/v1/legacy/common/alloc"
	"github.com/opennox/opennox/v1/server"
)

type scriptHandlerXferNativeDeps4F5580 struct {
	rwVersion       func(uint16) uint16
	readOnly        func() int32
	rwNameLength    func(uint32) uint32
	rwNameBytes     func([]byte)
	gameFlagCheck   func(uint32) int32
	storeContext    func(unsafe.Pointer, []byte)
	indexByName     func([]byte) int32
	storeFunc       func(*server.ScriptCallback, int32)
	loadFunc        func(*server.ScriptCallback) int32
	loadContextName func(unsafe.Pointer) []byte
	callbackName    func(int32) []byte
	rwFlags         func(*server.ScriptCallback)
}

func scriptHandlerXferRuntimeDeps4F5580(
	cf *cryptfile.CryptFile,
) scriptHandlerXferNativeDeps4F5580 {
	return scriptHandlerXferNativeDeps4F5580{
		rwVersion: func(value uint16) uint16 {
			value, _ = cf.ReadWriteU16(value)
			return value
		},
		readOnly: func() int32 {
			if cf.ReadOnly() {
				return 1
			}
			return 0
		},
		rwNameLength: func(value uint32) uint32 {
			value, _ = cf.ReadWriteU32(value)
			return value
		},
		rwNameBytes: func(value []byte) {
			_, _ = cf.ReadWrite(value)
		},
		gameFlagCheck: func(flag uint32) int32 {
			if noxflags.HasGame(noxflags.GameFlag(flag)) {
				return 1
			}
			return 0
		},
		storeContext: func(context unsafe.Pointer, name []byte) {
			destination := unsafe.Slice((*byte)(context), len(name)+1)
			copy(destination, name)
			destination[len(name)] = 0
		},
		indexByName: func(name []byte) int32 {
			return int32(GetServer().S().NoxScriptVM.ScriptIndexByName(string(name)))
		},
		storeFunc: func(handler *server.ScriptCallback, value int32) {
			handler.Func = value
		},
		loadFunc: func(handler *server.ScriptCallback) int32 {
			return handler.Func
		},
		loadContextName: func(context unsafe.Pointer) []byte {
			return []byte(alloc.GoString((*byte)(context)))
		},
		callbackName: func(function int32) []byte {
			return []byte(GetServer().S().NoxScriptVM.ScriptNameByIndex(int(function)))
		},
		rwFlags: func(handler *server.ScriptCallback) {
			_, _ = cf.ReadWrite(unsafe.Slice((*byte)(unsafe.Pointer(handler)), 4))
		},
	}
}

func scriptHandlerXferNative4F5580(
	handler *server.ScriptCallback,
	context unsafe.Pointer,
	deps scriptHandlerXferNativeDeps4F5580,
) int32 {
	return scriptHandlerXfer4F5580(
		handler,
		context,
		scriptHandlerXferDeps4F5580[*server.ScriptCallback, unsafe.Pointer]{
			rwVersion:       deps.rwVersion,
			readOnly:        deps.readOnly,
			rwNameLength:    deps.rwNameLength,
			rwNameBytes:     deps.rwNameBytes,
			gameFlagCheck:   deps.gameFlagCheck,
			storeContext:    deps.storeContext,
			indexByName:     deps.indexByName,
			storeFunc:       deps.storeFunc,
			loadFunc:        deps.loadFunc,
			loadContextName: deps.loadContextName,
			callbackName:    deps.callbackName,
			rwFlags:         deps.rwFlags,
		},
	)
}

func scriptHandlerXferRuntime4F5580(
	handler *server.ScriptCallback,
	context unsafe.Pointer,
) int32 {
	return scriptHandlerXferNative4F5580(
		handler,
		context,
		scriptHandlerXferRuntimeDeps4F5580(cryptfile.Global()),
	)
}
