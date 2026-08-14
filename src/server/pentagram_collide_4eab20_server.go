package server

import (
	"unsafe"

	"github.com/opennox/libs/types"
)

// PentagramUpdateDataPrefix is the pointer-independent prefix observed by
// PentagramCollide. The complete 24-byte original update record contains a
// legacy object pointer and will be widened when the Pentagram update path is
// restored; the collision callback only accesses Triggered at offset four.
type PentagramUpdateDataPrefix struct {
	Reserved0 [4]byte
	Triggered uint32
}

func pentagramCollideNative4EAB20(
	source, target *Object,
	collision *types.Pointf,
) *Object {
	return pentagramCollide4EAB20(
		source,
		target,
		collision,
		pentagramCollideHooks4EAB20[*Object, *PentagramUpdateDataPrefix]{
			loadUpdateData: func(obj *Object) *PentagramUpdateDataPrefix {
				return (*PentagramUpdateDataPrefix)(obj.UpdateData)
			},
			storeTriggered: func(data *PentagramUpdateDataPrefix, value uint32) {
				data.Triggered = value
			},
		},
	)
}

// PentagramCollide4EAB20 binds the original registered callback to the native
// Object update-data pointer. The callback dispatcher ignores the original EAX
// residue, so the public callback-shaped method has no return value.
func (s *Server) PentagramCollide4EAB20(
	source, target *Object,
	collision *types.Pointf,
) {
	_ = pentagramCollideNative4EAB20(source, target, collision)
}

var (
	_ = [1]struct{}{}[8-unsafe.Sizeof(PentagramUpdateDataPrefix{})]
	_ = [1]struct{}{}[4-unsafe.Offsetof(PentagramUpdateDataPrefix{}.Triggered)]
)
