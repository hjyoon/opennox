package legacy

/*
#include "GAME1_1.h"
*/
import "C"

import (
	"fmt"
	"io"
	"unsafe"

	crypt "github.com/opennox/noxcrypt"

	"github.com/opennox/opennox/v1/common/memmap"
	"github.com/opennox/opennox/v1/internal/cryptfile"
)

const (
	clientSaveHeaderCount   = 3
	clientSaveHeaderOffset  = 55936
	clientSaveHeaderStride  = 12
	clientSaveInfoSectionID = 1
	clientSaveInfoSize      = int(C.sizeof_nox_savegame_xxx)
)

func clientSaveHeaderEntry(index int) (id uint32, present bool) {
	offset := clientSaveHeaderOffset + uintptr(index)*clientSaveHeaderStride
	return memmap.Uint32(0x587000, offset+4), memmap.Uint32(0x587000, offset) != 0
}

func clientSaveHeaderIO(cf *cryptfile.CryptFile, index int) error {
	switch index {
	case 0:
		return saveClientGUISection(cf, nil)
	case 1:
		return saveFileInfoSection(cf, clientSaveFileInfo())
	case 2:
		return saveClientMusicSection(cf)
	default:
		return fmt.Errorf("unknown client save section %d", index)
	}
}

func readClientSaveHeader(cf *cryptfile.CryptFile) (id uint32, size int64, err error) {
	id, err = cf.ReadU32()
	if err != nil || id == 0 {
		return id, 0, err
	}
	raw, err := cf.ReadAlignedU32()
	return id, int64(int32(raw)), err
}

func copyClientSavePath(buf []byte, path string) error {
	if len(path) >= len(buf) {
		return fmt.Errorf("player save path requires %d bytes", len(path)+1)
	}
	buf[copy(buf, path)] = 0
	return nil
}

func readPlayerSavePreview(path string, output unsafe.Pointer) error {
	record := (*[clientSaveInfoSize]byte)(output)
	const unknownFieldOffset = 1028
	record[unknownFieldOffset] = memmap.Uint8(0x5D4594, 527728)
	if err := cryptfile.OpenGlobal(path, cryptfile.ReadOnly, crypt.SaveKey); err != nil {
		return err
	}
	defer cryptfile.Close()
	cf := cryptfile.Global()
	for {
		id, size, err := readClientSaveHeader(cf)
		if err != nil {
			return err
		}
		if id == 0 {
			break
		}
		if id != clientSaveInfoSectionID {
			if err := cf.Seek(size, io.SeekCurrent); err != nil {
				return err
			}
			continue
		}

		info := *clientSaveFileInfo()
		_, first := clientSaveHeaderEntry(0)
		_, second := clientSaveHeaderEntry(1)
		if first && second {
			if err := saveFileInfoSection(cf, &info); err != nil {
				return err
			}
		} else if err := cf.Seek(size, io.SeekCurrent); err != nil {
			return err
		}
		*record = *(*[clientSaveInfoSize]byte)(unsafe.Pointer(&info))
	}
	return copyClientSavePath(record[4:1028], path)
}

//export sub_41A000
func sub_41A000(path *C.char, output *C.nox_savegame_xxx) C.int {
	return C.int(bool2int(readPlayerSavePreview(GoString(path), unsafe.Pointer(output)) == nil))
}

func loadClientSaveHeader(path string) error {
	if err := cryptfile.OpenGlobal(path, cryptfile.ReadOnly, crypt.SaveKey); err != nil {
		return err
	}
	defer cryptfile.Close()
	cf := cryptfile.Global()
	sub_4602F0()

	for {
		id, size, err := readClientSaveHeader(cf)
		if err != nil {
			return err
		}
		if id == 0 {
			break
		}
		found := false
		for index := range clientSaveHeaderCount {
			section, present := clientSaveHeaderEntry(index)
			if !present {
				break
			}
			if id == section {
				if err := clientSaveHeaderIO(cf, index); err != nil {
					return err
				}
				found = true
				break
			}
		}
		if !found {
			if err := cf.Seek(size, io.SeekCurrent); err != nil {
				return err
			}
		}
	}
	return copyClientSavePath(clientSaveFileInfo().PathBuf[:], path)
}

//export nox_xxx_plrLoad_41A480
func nox_xxx_plrLoad_41A480(path *C.char) C.int {
	return C.int(bool2int(loadClientSaveHeader(GoString(path)) == nil))
}

func writeClientSaveHeader(input unsafe.Pointer, all bool) error {
	info := clientSaveFileInfo()
	*(*[clientSaveInfoSize]byte)(unsafe.Pointer(info)) = *(*[clientSaveInfoSize]byte)(input)
	if err := cryptfile.OpenGlobal(info.Path(), cryptfile.WriteOnly, crypt.SaveKey); err != nil {
		return err
	}
	defer cryptfile.Close()
	cf := cryptfile.Global()
	for index := range clientSaveHeaderCount {
		id, present := clientSaveHeaderEntry(index)
		if !present {
			break
		}
		if !all && id != clientSaveInfoSectionID {
			continue
		}
		if err := cf.WriteU32(id); err != nil {
			return err
		}
		cf.SectionStart()
		err := clientSaveHeaderIO(cf, index)
		cf.SectionEnd()
		if err != nil {
			return err
		}
	}
	return cf.WriteU32(0)
}

//export sub_41CEE0
func sub_41CEE0(data unsafe.Pointer, all C.int) C.int {
	return C.int(bool2int(writeClientSaveHeader(data, all != 0) == nil))
}
