package legacy

/*
#include <stdint.h>
*/
import "C"

import (
	"errors"
	"syscall"
	"unsafe"

	"github.com/opennox/libs/ifs"
)

const mapgenPathBufferSize503140 = 0x800
const mapgenBackupSuffix503140 = `\AreaMap.bak`

func mapgenPrepareAreaMapWithFS503140(source, directory string, remove func(string) error, rename func(string, string) error) bool {
	if source == "" || directory == "" || len(directory)+len(mapgenBackupSuffix503140) >= mapgenPathBufferSize503140 {
		return false
	}
	backup := directory + mapgenBackupSuffix503140
	if err := remove(backup); errors.Is(err, syscall.EACCES) {
		return false
	}
	return rename(source, backup) == nil
}

func mapgenBoundedCString(p unsafe.Pointer) (string, bool) {
	if p == nil {
		return "", false
	}
	for offset := range mapgenPathBufferSize503140 {
		if *(*byte)(unsafe.Add(p, offset)) == 0 {
			return string(unsafe.Slice((*byte)(p), offset)), true
		}
	}
	return "", false
}

//export nox_mapgenPrepareAreaMap_503140
func nox_mapgenPrepareAreaMap_503140(source, directory *C.char) C.int {
	src, ok := mapgenBoundedCString(unsafe.Pointer(source))
	if !ok {
		return 0
	}
	dir, ok := mapgenBoundedCString(unsafe.Pointer(directory))
	if !ok {
		return 0
	}
	if mapgenPrepareAreaMapWithFS503140(src, dir, ifs.Remove, ifs.Rename) {
		return 1
	}
	return 0
}
