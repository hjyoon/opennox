package server

import (
	"unsafe"

	"github.com/opennox/libs/object"
)

// TriggerUpdateData is GAME.EXE's fixed-width 60-byte TriggerUpdate record.
// Field4 is a transient PE32 object pointer in the original executable. Native
// 64-bit builds keep that pointer in ObjectExt.triggerCollideTarget instead,
// leaving this wire/template record identical on every architecture.
type TriggerUpdateData struct {
	Flags            uint32         // 0
	Field4           uint32         // 4; PE32-only transient object pointer
	State            uint8          // 8
	Field9           uint8          // 9
	_                [2]byte        // 10
	ScriptCollide    ScriptCallback // 12
	ScriptActivate   ScriptCallback // 20
	ScriptDeactivate ScriptCallback // 28
	SoundActivate    uint32         // 36
	SoundDeactivate  uint32         // 40
	ClassInclude     uint32         // 44
	ClassExclude     uint32         // 48
	TeamInclude      uint8          // 52
	TeamExclude      uint8          // 53
	Colors           [6]uint8       // 54
}

// ToggleUpdateData has the same fixed-width record as TriggerUpdateData.
// GAME.EXE uses a different state machine and event pair for ToggleUpdate.
type ToggleUpdateData = TriggerUpdateData

var (
	_ = [1]struct{}{}[60-unsafe.Sizeof(TriggerUpdateData{})]
	_ = [1]struct{}{}[0-unsafe.Offsetof(TriggerUpdateData{}.Flags)]
	_ = [1]struct{}{}[4-unsafe.Offsetof(TriggerUpdateData{}.Field4)]
	_ = [1]struct{}{}[8-unsafe.Offsetof(TriggerUpdateData{}.State)]
	_ = [1]struct{}{}[12-unsafe.Offsetof(TriggerUpdateData{}.ScriptCollide)]
	_ = [1]struct{}{}[20-unsafe.Offsetof(TriggerUpdateData{}.ScriptActivate)]
	_ = [1]struct{}{}[28-unsafe.Offsetof(TriggerUpdateData{}.ScriptDeactivate)]
	_ = [1]struct{}{}[36-unsafe.Offsetof(TriggerUpdateData{}.SoundActivate)]
	_ = [1]struct{}{}[40-unsafe.Offsetof(TriggerUpdateData{}.SoundDeactivate)]
	_ = [1]struct{}{}[44-unsafe.Offsetof(TriggerUpdateData{}.ClassInclude)]
	_ = [1]struct{}{}[48-unsafe.Offsetof(TriggerUpdateData{}.ClassExclude)]
	_ = [1]struct{}{}[52-unsafe.Offsetof(TriggerUpdateData{}.TeamInclude)]
	_ = [1]struct{}{}[53-unsafe.Offsetof(TriggerUpdateData{}.TeamExclude)]
	_ = [1]struct{}{}[54-unsafe.Offsetof(TriggerUpdateData{}.Colors)]
)

func (t *ObjectType) UpdateDataTrigger() *TriggerUpdateData {
	if !t.Class().Has(object.ClassTrigger) {
		panic(t.Class().String())
	}
	return (*TriggerUpdateData)(t.UpdateData)
}

func (obj *Object) UpdateDataToggle() *ToggleUpdateData {
	return updateDataAs[ToggleUpdateData](obj)
}

func (obj *Object) TriggerCollideTarget() *Object {
	if obj == nil {
		return nil
	}
	return obj.GetExt().triggerCollideTarget
}

func (obj *Object) SetTriggerCollideTarget(target *Object) {
	if obj == nil {
		return
	}
	obj.SetExt().triggerCollideTarget = target
}
