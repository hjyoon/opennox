package server

import "unsafe"

// TransporterUpdateData is GAME.EXE's fixed-width 20-byte TransporterUpdate
// record. TargetPE32 is a transient object pointer in the original executable.
// Native builds keep that pointer in ObjectExt.transporterTarget so it cannot
// overlap TargetExtent on 64-bit systems.
type TransporterUpdateData struct {
	_            [12]byte
	TargetPE32   uint32
	TargetExtent uint32
}

var (
	_ = [1]struct{}{}[20-unsafe.Sizeof(TransporterUpdateData{})]
	_ = [1]struct{}{}[12-unsafe.Offsetof(TransporterUpdateData{}.TargetPE32)]
	_ = [1]struct{}{}[16-unsafe.Offsetof(TransporterUpdateData{}.TargetExtent)]
)

func (obj *Object) UpdateDataTransporter() *TransporterUpdateData {
	return updateDataAs[TransporterUpdateData](obj)
}

// TransporterTargetFor returns the native target associated with the exact
// update-data pointer cached by the caller. Objects without a Server handle
// can occur at the public C ABI boundary and have no native extension state.
func (obj *Object) TransporterTargetFor(data *TransporterUpdateData) *Object {
	if obj == nil || data == nil || obj.serverHandle == 0 {
		return nil
	}
	ext := obj.GetExt()
	if ext.transporterUpdateData != data {
		return nil
	}
	return ext.transporterTarget
}

// SetTransporterTargetFor associates a native target with an exact update-data
// pointer and keeps the original PE32 pointer slot zero on every architecture.
func (obj *Object) SetTransporterTargetFor(data *TransporterUpdateData, target *Object) {
	if data != nil {
		data.TargetPE32 = 0
	}
	if obj == nil || obj.serverHandle == 0 {
		return
	}
	ext := obj.SetExt()
	ext.transporterUpdateData = data
	ext.transporterTarget = target
}

func (obj *Object) TransporterTarget() *Object {
	if obj == nil {
		return nil
	}
	return obj.TransporterTargetFor(obj.UpdateDataTransporter())
}

func (obj *Object) SetTransporterTarget(target *Object) {
	if obj == nil {
		return
	}
	data := obj.UpdateDataTransporter()
	obj.SetTransporterTargetFor(data, target)
}
