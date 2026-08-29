package server

import "unsafe"

// MoverUpdateData is GAME.EXE's fixed-width 36-byte MoverUpdate record.
// Field_3, Field_5, and Field_7 are transient PE32 pointer slots in the
// original executable. Native builds keep those pointers in ObjectExt so a
// wider pointer cannot overlap the adjacent serialized fields.
type MoverUpdateData struct {
	Field_0 uint8   // 0, 0
	_       [3]byte // 0, 1
	Field_1 float32 // 1, 4
	Field_2 int32   // 2, 8
	Field_3 uint32  // 3, 12; waypoint pointer PE32
	Field_4 uint32  // 4, 16
	Field_5 uint32  // 5, 20; waypoint pointer PE32
	Field_6 uint32  // 6, 24
	Field_7 uint32  // 7, 28; target object pointer PE32
	Field_8 uint32  // 8, 32
}

var (
	_ = [1]struct{}{}[36-unsafe.Sizeof(MoverUpdateData{})]
	_ = [1]struct{}{}[0-unsafe.Offsetof(MoverUpdateData{}.Field_0)]
	_ = [1]struct{}{}[4-unsafe.Offsetof(MoverUpdateData{}.Field_1)]
	_ = [1]struct{}{}[8-unsafe.Offsetof(MoverUpdateData{}.Field_2)]
	_ = [1]struct{}{}[12-unsafe.Offsetof(MoverUpdateData{}.Field_3)]
	_ = [1]struct{}{}[16-unsafe.Offsetof(MoverUpdateData{}.Field_4)]
	_ = [1]struct{}{}[20-unsafe.Offsetof(MoverUpdateData{}.Field_5)]
	_ = [1]struct{}{}[24-unsafe.Offsetof(MoverUpdateData{}.Field_6)]
	_ = [1]struct{}{}[28-unsafe.Offsetof(MoverUpdateData{}.Field_7)]
	_ = [1]struct{}{}[32-unsafe.Offsetof(MoverUpdateData{}.Field_8)]
)

func (obj *Object) UpdateDataMover() *MoverUpdateData {
	return (*MoverUpdateData)(obj.UpdateData)
}

func moverWaypointPE32Slot(data *MoverUpdateData, slot int) *uint32 {
	switch slot {
	case 3:
		return &data.Field_3
	case 5:
		return &data.Field_5
	default:
		panic("Mover waypoint slot must be 3 or 5")
	}
}

// MoverWaypointFor returns the native waypoint associated with the exact
// update-data pointer cached by the caller. On original-width builds it also
// accepts a pointer left in the PE32 slot by an unported legacy caller.
func (obj *Object) MoverWaypointFor(data *MoverUpdateData, slot int) *Waypoint {
	if data == nil {
		return nil
	}
	pe32 := *moverWaypointPE32Slot(data, slot)
	if obj != nil && obj.serverHandle != 0 {
		ext := obj.GetExt()
		if ext.moverUpdateData == data {
			var waypoint *Waypoint
			switch slot {
			case 3:
				waypoint = ext.moverWaypoint3
			case 5:
				waypoint = ext.moverWaypoint5
			}
			if waypoint != nil || pe32 == 0 {
				return waypoint
			}
		}
	}
	return moverWaypointFromPE32(pe32)
}

// SetMoverWaypointFor associates a native waypoint with an exact update-data
// pointer and keeps the original PE32 pointer slot zero.
func (obj *Object) SetMoverWaypointFor(data *MoverUpdateData, slot int, waypoint *Waypoint) {
	if data == nil {
		return
	}
	*moverWaypointPE32Slot(data, slot) = 0
	if obj == nil || obj.serverHandle == 0 {
		return
	}
	ext := obj.SetExt()
	if ext.moverUpdateData != data {
		ext.moverUpdateData = data
		ext.moverWaypoint3 = nil
		ext.moverWaypoint5 = nil
		ext.moverTarget = nil
	}
	switch slot {
	case 3:
		ext.moverWaypoint3 = waypoint
	case 5:
		ext.moverWaypoint5 = waypoint
	}
}

// MoverTargetFor returns the native target associated with the exact update
// data pointer. Field_7 remains the fixed-width PE32 compatibility slot.
func (obj *Object) MoverTargetFor(data *MoverUpdateData) *Object {
	if data == nil {
		return nil
	}
	if obj != nil && obj.serverHandle != 0 {
		ext := obj.GetExt()
		if ext.moverUpdateData == data {
			if ext.moverTarget != nil || data.Field_7 == 0 {
				return ext.moverTarget
			}
		}
	}
	return moverObjectFromPE32(data.Field_7)
}

// SetMoverTargetFor associates a native target with an exact update-data
// pointer and keeps the original PE32 pointer slot zero.
func (obj *Object) SetMoverTargetFor(data *MoverUpdateData, target *Object) {
	if data != nil {
		data.Field_7 = 0
	}
	if obj == nil || data == nil || obj.serverHandle == 0 {
		return
	}
	ext := obj.SetExt()
	if ext.moverUpdateData != data {
		ext.moverUpdateData = data
		ext.moverWaypoint3 = nil
		ext.moverWaypoint5 = nil
	}
	ext.moverTarget = target
}

func (obj *Object) MoverTarget() *Object {
	if obj == nil {
		return nil
	}
	return obj.MoverTargetFor(obj.UpdateDataMover())
}

func (obj *Object) SetMoverTarget(target *Object) {
	if obj == nil {
		return
	}
	obj.SetMoverTargetFor(obj.UpdateDataMover(), target)
}
