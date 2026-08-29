package server

import "unsafe"

type ElevatorUpdateData struct {
	Field_0 uint32  // 0, 0
	Field_1 uint32  // 1, 4
	Field_2 uint32  // 2, 8
	Field_3 byte    // 3, 12
	_       [3]byte // 3, 13
	Field_4 uint32  // 4, 16
}

type ElevatorShaftUpdateData struct {
	Field_0 uint32  // 0, 0
	Field_1 uint32  // 1, 4; legacy pointer slot, runtime link lives in ObjectExt
	Field_2 uint32  // 2, 8
	Field_3 byte    // 3, 12
	_       [3]byte // 3, 13
}

var (
	_ = [1]struct{}{}[20-unsafe.Sizeof(ElevatorUpdateData{})]
	_ = [1]struct{}{}[0-unsafe.Offsetof(ElevatorUpdateData{}.Field_0)]
	_ = [1]struct{}{}[4-unsafe.Offsetof(ElevatorUpdateData{}.Field_1)]
	_ = [1]struct{}{}[8-unsafe.Offsetof(ElevatorUpdateData{}.Field_2)]
	_ = [1]struct{}{}[12-unsafe.Offsetof(ElevatorUpdateData{}.Field_3)]
	_ = [1]struct{}{}[16-unsafe.Offsetof(ElevatorUpdateData{}.Field_4)]

	_ = [1]struct{}{}[16-unsafe.Sizeof(ElevatorShaftUpdateData{})]
	_ = [1]struct{}{}[0-unsafe.Offsetof(ElevatorShaftUpdateData{}.Field_0)]
	_ = [1]struct{}{}[4-unsafe.Offsetof(ElevatorShaftUpdateData{}.Field_1)]
	_ = [1]struct{}{}[8-unsafe.Offsetof(ElevatorShaftUpdateData{}.Field_2)]
	_ = [1]struct{}{}[12-unsafe.Offsetof(ElevatorShaftUpdateData{}.Field_3)]
)

func (obj *Object) UpdateDataElevator() *ElevatorUpdateData {
	return (*ElevatorUpdateData)(obj.UpdateData)
}

func (obj *Object) UpdateDataElevatorShaft() *ElevatorShaftUpdateData {
	return (*ElevatorShaftUpdateData)(obj.UpdateData)
}

// ElevatorLinkFor returns the native link associated with the exact update
// data pointer cached by the caller. Both ElevatorUpdate and
// ElevatorShaftUpdate use the PE32 pointer slot at byte offset 4.
func (obj *Object) ElevatorLinkFor(data unsafe.Pointer) *Object {
	if obj == nil || data == nil || obj.serverHandle == 0 {
		return nil
	}
	ext := obj.GetExt()
	if ext.elevatorUpdateData != data {
		return nil
	}
	return ext.elevatorLink
}

// SetElevatorLinkFor associates a native link with an exact update-data
// pointer and keeps the original PE32 pointer slot zero on every architecture.
func (obj *Object) SetElevatorLinkFor(data unsafe.Pointer, link *Object) {
	if data != nil {
		*(*uint32)(unsafe.Add(data, unsafe.Offsetof(ElevatorUpdateData{}.Field_1))) = 0
	}
	if obj == nil || obj.serverHandle == 0 {
		return
	}
	ext := obj.SetExt()
	ext.elevatorUpdateData = data
	ext.elevatorLink = link
}

func (obj *Object) ElevatorLink() *Object {
	if obj == nil {
		return nil
	}
	return obj.ElevatorLinkFor(obj.UpdateData)
}

func (obj *Object) SetElevatorLink(link *Object) {
	if obj == nil {
		return
	}
	obj.SetElevatorLinkFor(obj.UpdateData, link)
}
