package legacy

/*
#include <stdint.h>

// A bounded C-string scan avoids reading past short C.CString allocations.
static int nox_mapgen_name_length_503230(const char* p) {
	if (!p) {
		return -1;
	}
	for (int i = 0; i < 0x800; ++i) {
		if (!p[i]) {
			return i;
		}
	}
	return -1;
}
*/
import "C"

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"io"
	"log"
	"math"
	"os"
	"path/filepath"

	"github.com/opennox/libs/ifs"
)

// mapgenRenameRecords503230 preserves unmatched records byte-for-byte. A
// matching name changes only the record length and the name field; the flags,
// coordinates, and trailing payload are copied unchanged.
func mapgenRenameRecords503230(input io.Reader, output io.Writer, oldName, newName string) error {
	if len(newName) > mapgenMaxNameLength502ED0 {
		return fmt.Errorf("AreaMap replacement name exceeds %d bytes", mapgenMaxNameLength502ED0)
	}
	var word [4]byte
	if _, err := io.ReadFull(input, word[:]); err != nil {
		return err
	}
	if binary.LittleEndian.Uint32(word[:]) != mapgenMagic502ED0 {
		return fmt.Errorf("invalid AreaMap magic: %#x", binary.LittleEndian.Uint32(word[:]))
	}
	if err := mapgenWriteExact502ED0(output, word[:]); err != nil {
		return err
	}

	for count := 0; ; count++ {
		if _, err := io.ReadFull(input, word[:]); err != nil {
			return err
		}
		length := int32(binary.LittleEndian.Uint32(word[:]))
		if length == 0 {
			break
		}
		if count == mapgenMaxRecords502ED0 || length < 11 {
			return fmt.Errorf("invalid AreaMap record count or length: %d, %d", count, length)
		}
		var nameLength [1]byte
		if _, err := io.ReadFull(input, nameLength[:]); err != nil {
			return err
		}
		if nameLength[0] > mapgenMaxNameLength502ED0 || length < 11+int32(nameLength[0]) {
			return fmt.Errorf("invalid AreaMap name length: %d in record %d", nameLength[0], length)
		}
		originalNameLength := int64(nameLength[0])
		name := make([]byte, nameLength[0])
		if _, err := io.ReadFull(input, name); err != nil {
			return err
		}
		cName := name
		if i := bytes.IndexByte(name, 0); i >= 0 {
			cName = name[:i]
		}
		if string(cName) == oldName {
			newLength := int64(length) - int64(nameLength[0]) + int64(len(newName))
			if newLength > math.MaxInt32 {
				return fmt.Errorf("AreaMap renamed record length exceeds int32: %d", newLength)
			}
			binary.LittleEndian.PutUint32(word[:], uint32(newLength))
			nameLength[0] = byte(len(newName))
			name = []byte(newName)
		}
		if err := mapgenWriteExact502ED0(output, word[:]); err != nil {
			return err
		}
		if err := mapgenWriteExact502ED0(output, nameLength[:]); err != nil {
			return err
		}
		if err := mapgenWriteExact502ED0(output, name); err != nil {
			return err
		}
		// The payload starts after the declared original name, even when
		// the C-string comparison stopped at an embedded NUL.
		remaining := int64(length) - 1 - originalNameLength
		if _, err := io.CopyN(output, input, remaining); err != nil {
			return err
		}
	}
	binary.LittleEndian.PutUint32(word[:], 0)
	return mapgenWriteExact502ED0(output, word[:])
}

func mapgenRenameAreaMapFile503230(oldName, newName, source, directory string) error {
	backup := directory + mapgenBackupSuffix503140
	input, err := ifs.Open(backup)
	if err != nil {
		return err
	}
	defer input.Close()

	temp, err := os.CreateTemp(filepath.Dir(ifs.Normalize(source)), ".AreaMap-*.tmp")
	if err != nil {
		return err
	}
	defer os.Remove(temp.Name())
	defer temp.Close()
	if err := mapgenRenameRecords503230(input, temp, oldName, newName); err != nil {
		return err
	}
	if err := temp.Close(); err != nil {
		return err
	}
	return ifs.Rename(temp.Name(), source)
}

//export nox_mapgenRenameAreaMap_503230
func nox_mapgenRenameAreaMap_503230(oldName, newName, source, directory *C.char) C.int {
	src, ok := mapgenBoundedCPath503140(source)
	if !ok {
		return 0
	}
	dir, ok := mapgenBoundedCPath503140(directory)
	if !ok || len(dir)+len(mapgenBackupSuffix503140) >= mapgenPathBufferSize503140 {
		return 0
	}
	rollback := func(reason error) C.int {
		backup := dir + mapgenBackupSuffix503140
		if restoreErr := ifs.Rename(backup, src); restoreErr != nil {
			log.Printf("AreaMap rename failed (%v) and backup %q could not be restored to %q: %v", reason, backup, src, restoreErr)
		}
		return 0
	}
	oldLength := int(C.nox_mapgen_name_length_503230(oldName))
	newLength := int(C.nox_mapgen_name_length_503230(newName))
	if oldLength < 0 || newLength < 0 {
		return rollback(fmt.Errorf("unterminated AreaMap rename argument"))
	}
	old, new := C.GoStringN(oldName, C.int(oldLength)), C.GoStringN(newName, C.int(newLength))
	if err := mapgenRenameAreaMapFile503230(old, new, src, dir); err != nil {
		return rollback(err)
	}
	return 1
}
