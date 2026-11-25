package legacy

/*
#include "GAME1_2.h"
#include "GAME2_1.h"
*/
import "C"

import (
	"fmt"
	"io"
	"unsafe"

	"github.com/opennox/opennox/v1/common/flags"
	"github.com/opennox/opennox/v1/common/memmap"
	"github.com/opennox/opennox/v1/internal/cryptfile"
	"github.com/opennox/opennox/v1/legacy/common/alloc"
)

func mapMetadataTransfer(cf *cryptfile.CryptFile, field []byte) error {
	if cf.ReadOnly() {
		_, err := io.ReadFull(cf, field)
		return err
	}
	n, err := cf.Write(field)
	if err == nil && n != len(field) {
		return io.ErrShortWrite
	}
	return err
}

func mapMetadataName(cf *cryptfile.CryptFile, buf []byte) error {
	size, err := cf.ReadWriteU8(byte(alloc.StrLenS(buf)))
	if err != nil {
		return err
	}
	if int(size) > len(buf) {
		return fmt.Errorf("map quest name has %d bytes, maximum %d", size, len(buf))
	}
	if err := mapMetadataTransfer(cf, buf[:size]); err != nil {
		return err
	}
	if int(size) < len(buf) {
		buf[size] = 0
	}
	return nil
}

func mapRWMetadata(cf *cryptfile.CryptFile) error {
	if cf == nil {
		return fmt.Errorf("map metadata requires an open crypt file")
	}
	version, err := cf.ReadWriteU16(3)
	if err != nil {
		return err
	}
	if version > 3 {
		return fmt.Errorf("unsupported map metadata version %d", version)
	}
	dst := memmap.PtrT[[1462]byte](0x973F18, 2408)
	buf := *dst
	if version >= 1 {
		offset := 0
		for _, size := range [...]int{64, 512, 16, 64, 64, 128, 128, 256, 128, 32, 4} {
			if err := mapMetadataTransfer(cf, buf[offset:offset+size]); err != nil {
				return err
			}
			offset += size
		}
		if version == 2 {
			for i := range 2 {
				if err := mapMetadataTransfer(cf, buf[1396+i:1397+i]); err != nil {
					return err
				}
			}
		} else if cf.ReadOnly() {
			buf[1396], buf[1397] = 2, 16
		}
	}
	if version < 3 {
		buf[1398] = memmap.Uint8(0x5D4594, 741376)
		buf[1430] = memmap.Uint8(0x5D4594, 741380)
	} else {
		for _, offset := range [...]int{1398, 1430} {
			if err := mapMetadataName(cf, buf[offset:offset+32]); err != nil {
				return err
			}
		}
	}
	*dst = buf
	return nil
}

//export nox_server_mapRWMapInfo_42A6E0
func nox_server_mapRWMapInfo_42A6E0() C.int {
	return C.int(bool2int(mapRWMetadata(cryptfile.Global()) == nil))
}

func mapRWAmbientColor(cf *cryptfile.CryptFile) error {
	if cf == nil {
		return fmt.Errorf("ambient map lighting requires an open crypt file")
	}
	version, err := cf.ReadWriteU16(1)
	if err != nil {
		return err
	}
	if int16(version) < 1 {
		return fmt.Errorf("unsupported ambient map lighting version %d", version)
	}
	var color [3]uint32
	if !cf.ReadOnly() {
		color = *(*[3]uint32)(unsafe.Pointer(C.nox_xxx_getAmbientColor_469BB0()))
	}
	for i, intensity := range color {
		value, err := cf.ReadWriteU32(intensity)
		if err != nil {
			return err
		}
		color[i] = value
	}
	if cf.ReadOnly() {
		C.sub_469B90((*C.int)(unsafe.Pointer(&color[0])))
		if noxflags.HasGame(noxflags.GameClient | noxflags.GameFlag22) {
			sub_4349C0((*C.uint)(unsafe.Pointer(&color[0])))
		}
	}
	return nil
}

//export nox_server_mapRWAmbientData_429200
func nox_server_mapRWAmbientData_429200() C.int {
	return C.int(bool2int(mapRWAmbientColor(cryptfile.Global()) == nil))
}
