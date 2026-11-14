package legacy

import (
	"unsafe"

	"github.com/opennox/opennox/v1/legacy/common/alloc"
)

func soundSetFieldNative4240F0(data, field unsafe.Pointer, value uint32) int32 {
	if setSoundSetField(data, alloc.GoString((*byte)(field)), value) {
		return 1
	}
	return 0
}

// The public legacy entry retains its original token/service and error-return
// contract. The existing production Go loader separately defers file cleanup.
// Numeric sound IDs remain packed in foreign84-byte storage; names and links
// belong to Go metadata rather than truncated DWORD pointer fields.
func soundSetParseNative424170(path unsafe.Pointer) int32 {
	file := soundSetParserOpen424170(path)
	if file == nil || soundSetParserKey424170(file) == 0 {
		return 0
	}
	var name, soundName [256]byte
	for soundSetParserToken424170(file, unsafe.Pointer(&name[0])) != 0 {
		data := soundSetParserAllocate424170()
		ownedName := alloc.GoString(&name[0])
		addSoundSetRecord(ownedName, data)
		soundSetParserNameStored424170(ownedName)
		for soundSetParserToken424170(file, unsafe.Pointer(&name[0])) != 0 && alloc.GoString(&name[0]) != "END" && soundSetParserToken424170(file, unsafe.Pointer(&soundName[0])) != 0 {
			value := soundSetParserSound424170(unsafe.Pointer(&soundName[0]))
			if soundSetFieldNative4240F0(data, unsafe.Pointer(&name[0]), value) == 0 {
				return 0
			}
		}
	}
	soundSetParserClose424170(file)
	return 1
}
