package legacy

const (
	scriptHandlerXferVersion4F5580  = uint16(1)
	scriptHandlerXferMaxName4F5580  = uint32(1024)
	scriptHandlerXferGameFlag4F5580 = uint32(0x600000)
)

type scriptHandlerXferDeps4F5580[H any, C comparable] struct {
	rwVersion       func(uint16) uint16
	readOnly        func() int32
	rwNameLength    func(uint32) uint32
	rwNameBytes     func([]byte)
	gameFlagCheck   func(uint32) int32
	storeContext    func(C, []byte)
	indexByName     func([]byte) int32
	storeFunc       func(H, int32)
	loadFunc        func(H) int32
	loadContextName func(C) []byte
	callbackName    func(int32) []byte
	rwFlags         func(H)
}

// scriptHandlerXfer4F5580 preserves GAME.EXE 004F5580, including its signed
// version gate, exact-one read-mode branch, zero-length transfers, and fault
// prefixes. The original does not validate handler/context pointers, does not
// validate callback indices, and does not roll back a context/Func mutation if
// the final flags transfer faults.
func scriptHandlerXfer4F5580[H any, C comparable](
	handler H,
	context C,
	deps scriptHandlerXferDeps4F5580[H, C],
) int32 {
	version := deps.rwVersion(scriptHandlerXferVersion4F5580)
	if int16(version) > int16(scriptHandlerXferVersion4F5580) {
		return 0
	}

	if deps.readOnly() == 1 {
		nameLength := deps.rwNameLength(0)
		if nameLength >= scriptHandlerXferMaxName4F5580 {
			return 0
		}
		name := make([]byte, int(nameLength))
		deps.rwNameBytes(name)
		if nameLength != 0 {
			if deps.gameFlagCheck(scriptHandlerXferGameFlag4F5580) != 0 {
				deps.storeContext(context, name)
			} else {
				deps.storeFunc(handler, deps.indexByName(name))
			}
		}
		deps.rwFlags(handler)
		return 1
	}

	if deps.gameFlagCheck(scriptHandlerXferGameFlag4F5580) != 0 {
		var nilContext C
		if context != nilContext {
			name := deps.loadContextName(context)
			deps.rwNameLength(uint32(len(name)))
			deps.rwNameBytes(name)
			deps.rwFlags(handler)
			return 1
		}
	} else {
		function := deps.loadFunc(handler)
		if function != -1 {
			name := deps.callbackName(function)
			deps.rwNameLength(uint32(len(name)))
			deps.rwNameBytes(name)
			deps.rwFlags(handler)
			return 1
		}
	}

	deps.rwNameLength(0)
	deps.rwFlags(handler)
	return 1
}
