package server

import (
	"unsafe"

	"github.com/opennox/libs/types"
)

// PentagramUpdateDataPrefix is the pointer-independent prefix observed by
// PentagramCollide. The collision callback only accesses Triggered at offset
// four.
type PentagramUpdateDataPrefix struct {
	Reserved0 [4]byte
	Triggered uint32
}

// PentagramUpdateData is GAME.EXE's fixed-width 24-byte PentagramUpdate
// record. DestinationPE32 is the original transient object-pointer slot and
// DestinationExtent is the adjacent value serialized by TransporterXfer.
// Native builds keep the actual destination in ObjectExt.transporterTarget so
// the two PE32 dwords can never overlap on 64-bit systems.
type PentagramUpdateData struct {
	State             uint8
	_                 [3]byte
	Triggered         uint32
	AnimationFrame    uint8
	AnimationTick     uint8
	_                 [2]byte
	DestinationPE32   uint32
	DestinationExtent uint32
	AnimationStep     uint8
	_                 [3]byte
}

func (obj *Object) UpdateDataPentagram() *PentagramUpdateData {
	return updateDataAs[PentagramUpdateData](obj)
}

// PentagramDestinationFor resolves the native destination associated with the
// exact update-data identity. PentagramUpdate's first twenty bytes deliberately
// share TransporterUpdateData's fixed PE32 layout and sidecar contract.
func (obj *Object) PentagramDestinationFor(data *PentagramUpdateData) *Object {
	return obj.TransporterTargetFor((*TransporterUpdateData)(unsafe.Pointer(data)))
}

func (obj *Object) SetPentagramDestinationFor(data *PentagramUpdateData, target *Object) {
	obj.SetTransporterTargetFor((*TransporterUpdateData)(unsafe.Pointer(data)), target)
}

func (obj *Object) PentagramDestination() *Object {
	if obj == nil {
		return nil
	}
	return obj.PentagramDestinationFor(obj.UpdateDataPentagram())
}

func (obj *Object) SetPentagramDestination(target *Object) {
	if obj == nil {
		return
	}
	obj.SetPentagramDestinationFor(obj.UpdateDataPentagram(), target)
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
	_ = [1]struct{}{}[24-unsafe.Sizeof(PentagramUpdateData{})]
	_ = [1]struct{}{}[4-unsafe.Offsetof(PentagramUpdateData{}.Triggered)]
	_ = [1]struct{}{}[8-unsafe.Offsetof(PentagramUpdateData{}.AnimationFrame)]
	_ = [1]struct{}{}[12-unsafe.Offsetof(PentagramUpdateData{}.DestinationPE32)]
	_ = [1]struct{}{}[16-unsafe.Offsetof(PentagramUpdateData{}.DestinationExtent)]
	_ = [1]struct{}{}[20-unsafe.Offsetof(PentagramUpdateData{}.AnimationStep)]
)
