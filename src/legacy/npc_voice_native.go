package legacy

import (
	"strings"
	"unsafe"

	"github.com/opennox/libs/object"
	"github.com/opennox/opennox/v1/server"
)

func setNPCVoiceSetNative424320(obj *server.Object, data unsafe.Pointer) int32 {
	if obj == nil || uint8(obj.ObjClass)&uint8(object.ClassMonster) == 0 {
		return 0
	}
	(*server.MonsterUpdateData)(obj.UpdateData).SoundSet122 = data
	return 1
}

func npcVoiceWireNative52AD10(obj *server.Object, readOnly bool, readWrite func([]byte) uintptr) uintptr {
	var length [1]byte
	if readOnly {
		readWrite(length[:])
		var name [256]byte
		readWrite(name[:length[0]])
		return uintptr(setNPCVoiceSetNative424320(obj, getDefaultSoundSet(cStringBytes528DB0(name[:length[0]]))))
	}
	var data unsafe.Pointer
	if obj != nil && uint8(obj.ObjClass)&uint8(object.ClassMonster) != 0 {
		data = (*server.MonsterUpdateData)(obj.UpdateData).SoundSet122
	}
	if data == nil {
		return readWrite(length[:])
	}
	name := soundSetName(data)
	if i := strings.IndexByte(name, 0); i >= 0 {
		name = name[:i]
	}
	// The original casts strlen to a byte, including wrap at 256. A present
	// empty name still performs the zero-byte payload call and its CRC effect.
	length[0] = byte(len(name))
	readWrite(length[:])
	return readWrite([]byte(name[:length[0]]))
}
