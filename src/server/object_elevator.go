package server

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

func (obj *Object) UpdateDataElevator() *ElevatorUpdateData {
	return (*ElevatorUpdateData)(obj.UpdateData)
}

func (obj *Object) UpdateDataElevatorShaft() *ElevatorShaftUpdateData {
	return (*ElevatorShaftUpdateData)(obj.UpdateData)
}

func (obj *Object) ElevatorLink() *Object {
	if obj == nil {
		return nil
	}
	return obj.GetExt().elevatorLink
}

func (obj *Object) SetElevatorLink(link *Object) {
	if obj != nil {
		obj.SetExt().elevatorLink = link
	}
}
