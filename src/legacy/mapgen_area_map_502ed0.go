package legacy

/*
#include <stdint.h>
*/
import "C"

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"

	"github.com/opennox/libs/ifs"
)

const mapgenMagic502ED0 = 0xCAFEDEAD
const mapgenMaxNameLength502ED0 = 63
const mapgenMaxRecords502ED0 = 2048

func mapgenWriteExact502ED0(w io.Writer, data []byte) error {
	n, err := w.Write(data)
	if err != nil {
		return err
	}
	if n != len(data) {
		return io.ErrShortWrite
	}
	return nil
}

// mapgenRewriteRecords502ED0 keeps the original byte representation of every
// record whose C-string name does not match target. The original returns true
// when at least one record survives, even if target was not present.
func mapgenRewriteRecords502ED0(input io.Reader, output io.Writer, target string) (bool, error) {
	var word [4]byte
	if _, err := io.ReadFull(input, word[:]); err != nil {
		return false, err
	}
	if binary.LittleEndian.Uint32(word[:]) != mapgenMagic502ED0 {
		return false, fmt.Errorf("invalid AreaMap magic: %#x", binary.LittleEndian.Uint32(word[:]))
	}
	if err := mapgenWriteExact502ED0(output, word[:]); err != nil {
		return false, err
	}

	var copiedAny bool
	var kept int
	for {
		if _, err := io.ReadFull(input, word[:]); err != nil {
			return false, err
		}
		length := int32(binary.LittleEndian.Uint32(word[:]))
		if length == 0 {
			break
		}
		if length < 11 {
			return false, fmt.Errorf("invalid AreaMap record length: %d", length)
		}
		var nameLength [1]byte
		if _, err := io.ReadFull(input, nameLength[:]); err != nil {
			return false, err
		}
		if nameLength[0] > mapgenMaxNameLength502ED0 || length < 11+int32(nameLength[0]) {
			return false, fmt.Errorf("invalid AreaMap name length: %d in record %d", nameLength[0], length)
		}
		name := make([]byte, nameLength[0])
		if _, err := io.ReadFull(input, name); err != nil {
			return false, err
		}
		cName := name
		if i := bytes.IndexByte(cName, 0); i >= 0 {
			cName = cName[:i]
		}
		remaining := int64(length) - 1 - int64(nameLength[0])
		if string(cName) == target {
			if _, err := io.CopyN(io.Discard, input, remaining); err != nil {
				return false, err
			}
			continue
		}
		if kept == mapgenMaxRecords502ED0 {
			return false, fmt.Errorf("AreaMap has more than %d retained records", mapgenMaxRecords502ED0)
		}
		if err := mapgenWriteExact502ED0(output, word[:]); err != nil {
			return false, err
		}
		if err := mapgenWriteExact502ED0(output, nameLength[:]); err != nil {
			return false, err
		}
		if err := mapgenWriteExact502ED0(output, name); err != nil {
			return false, err
		}
		if _, err := io.CopyN(output, input, remaining); err != nil {
			return false, err
		}
		copiedAny = true
		kept++
	}
	binary.LittleEndian.PutUint32(word[:], 0)
	if err := mapgenWriteExact502ED0(output, word[:]); err != nil {
		return false, err
	}
	return copiedAny, nil
}

func mapgenRewriteAreaMapFile502ED0(target, source, directory string) (bool, error) {
	backup := directory + mapgenBackupSuffix503140
	input, err := ifs.Open(backup)
	if err != nil {
		return false, err
	}
	defer input.Close()

	// The original opens source with "wb" immediately. Stage beside source so
	// malformed input or a short write cannot leave a partial AreaMap there.
	temp, err := os.CreateTemp(filepath.Dir(ifs.Normalize(source)), ".AreaMap-*.tmp")
	if err != nil {
		return false, err
	}
	defer os.Remove(temp.Name())
	defer temp.Close()
	copiedAny, err := mapgenRewriteRecords502ED0(input, temp, target)
	if err != nil {
		return false, err
	}
	if err := temp.Close(); err != nil {
		return false, err
	}
	if err := ifs.Rename(temp.Name(), source); err != nil {
		return false, err
	}
	return copiedAny, nil
}

//export nox_mapgenRewriteAreaMap_502ED0
func nox_mapgenRewriteAreaMap_502ED0(name, source, directory *C.char) C.int {
	src, ok := mapgenBoundedCPath503140(source)
	if !ok {
		return -1
	}
	dir, ok := mapgenBoundedCPath503140(directory)
	if !ok || len(dir)+len(mapgenBackupSuffix503140) >= mapgenPathBufferSize503140 {
		return -1
	}
	target := C.GoString(name)
	copiedAny, err := mapgenRewriteAreaMapFile502ED0(target, src, dir)
	if err != nil {
		backup := dir + mapgenBackupSuffix503140
		if restoreErr := ifs.Rename(backup, src); restoreErr != nil {
			log.Printf("AreaMap rewrite failed (%v) and backup %q could not be restored to %q: %v", err, backup, src, restoreErr)
		}
		return -1
	}
	if copiedAny {
		return 1
	}
	return 0
}
