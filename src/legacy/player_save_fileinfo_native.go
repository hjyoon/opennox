package legacy

/*
#include "GAME1_1.h"
*/
import "C"

import (
	"errors"
	"fmt"
	"io"
	"unsafe"

	"github.com/opennox/opennox/v1/common/flags"
	"github.com/opennox/opennox/v1/common/memmap"
	"github.com/opennox/opennox/v1/internal/cryptfile"
	"github.com/opennox/opennox/v1/legacy/common/alloc"
	"github.com/opennox/opennox/v1/server"
)

func clientSaveFileInfo() *server.SaveGameInfo {
	return memmap.PtrT[server.SaveGameInfo](0x85B3FC, 10980)
}

func saveFileInfoBytes(cf *cryptfile.CryptFile, buf []byte) error {
	count, err := cf.ReadWrite(buf)
	if cf.ReadOnly() && len(buf) == 0 {
		return nil
	}
	if err != nil {
		return err
	}
	if len(buf) != count {
		if cf.ReadOnly() {
			return io.ErrUnexpectedEOF
		}
		return io.ErrShortWrite
	}
	return nil
}

func saveFileInfoString(cf *cryptfile.CryptFile, buf []byte, path bool) error {
	size := alloc.StrLenS(buf)
	if path {
		value, err := cf.ReadWriteU16(uint16(size))
		if err != nil {
			return err
		}
		size = int(int16(value))
	} else {
		value, err := cf.ReadWriteU8(byte(size))
		if err != nil {
			return err
		}
		size = int(value)
	}
	if size < 0 || size >= len(buf) {
		return fmt.Errorf("save text length %d exceeds %d bytes", size, len(buf)-1)
	}
	if err := saveFileInfoBytes(cf, buf[:size]); err != nil {
		return err
	}
	if cf.ReadOnly() {
		buf[size] = 0
	}
	return nil
}

func saveFileInfoName(cf *cryptfile.CryptFile, buf []uint16) error {
	size, err := cf.ReadWriteU8(byte(alloc.StrLenS(buf)))
	if err != nil {
		return err
	}
	if int(size) >= len(buf) {
		return fmt.Errorf("save player name length %d exceeds %d code units", size, len(buf)-1)
	}
	raw := unsafe.Slice((*byte)(unsafe.Pointer(&buf[0])), 2*int(size))
	if err := saveFileInfoBytes(cf, raw); err != nil {
		return err
	}
	if cf.ReadOnly() {
		buf[size] = 0
	}
	return nil
}

func saveFileInfoSection(cf *cryptfile.CryptFile, info *server.SaveGameInfo) error {
	if cf == nil {
		return errors.New("player file info requires an open crypt file")
	}
	version, err := cf.ReadWriteU16(12)
	if err != nil {
		return err
	}
	if int16(version) > 12 {
		return fmt.Errorf("unsupported player file info version %d", version)
	}

	if noxflags.HasGame(noxflags.GameOnline) {
		info.Flags &^= 1
		if noxflags.HasGame(noxflags.GameModeQuest) || nox_xxx_isQuest_4D6F50() != 0 || sub_4D6F70() != 0 {
			info.Flags |= 4
		} else {
			info.Flags |= 2
		}
	} else {
		info.Flags = info.Flags&^6 | 1
	}
	info.Flags, err = cf.ReadWriteU32(info.Flags)
	if err != nil {
		return err
	}
	if err := saveFileInfoString(cf, info.PathBuf[:], true); err != nil {
		return err
	}
	if err := saveFileInfoString(cf, info.Field1028[:], false); err != nil {
		return err
	}

	noxGetLocalTime((*C.noxSYSTEMTIME)(unsafe.Pointer(&info.Timestamp)))
	timestamp := (*[8]uint16)(unsafe.Pointer(&info.Timestamp))
	for i := range timestamp {
		timestamp[i], err = cf.ReadWriteU16(timestamp[i])
		if err != nil {
			return err
		}
	}
	appearance := (*[20]byte)(unsafe.Pointer(&info.Player))
	for _, offset := range [...]int{3, 0, 6, 9, 12} {
		if err := saveFileInfoBytes(cf, appearance[offset:offset+3]); err != nil {
			return err
		}
	}
	for i := 15; i < len(appearance); i++ {
		appearance[i], err = cf.ReadWriteU8(appearance[i])
		if err != nil {
			return err
		}
	}
	if err := saveFileInfoName(cf, info.Player.NameBuf[:]); err != nil {
		return err
	}
	info.Player.PlayerClassVal, err = cf.ReadWriteU8(info.Player.PlayerClassVal)
	if err != nil {
		return err
	}
	info.Player.IsFemaleVal, err = cf.ReadWriteU8(0)
	if err != nil {
		return err
	}
	if !cf.ReadOnly() {
		info.Field1276 = byte(sub_467590())
	}
	info.Field1276, err = cf.ReadWriteU8(info.Field1276)
	if err != nil {
		return err
	}
	if int16(version) >= 11 {
		length := copy(info.MapNameBuf[:len(info.MapNameBuf)-1], GoString(nox_xxx_mapGetMapName_409B40()))
		info.MapNameBuf[length] = 0
		if err := saveFileInfoString(cf, info.MapNameBuf[:], false); err != nil {
			return err
		}
	}
	if int16(version) < 12 {
		info.Stage = 0
		return nil
	}
	info.Stage, err = cf.ReadWriteU8(info.Stage)
	return err
}

//export nox_xxx_parseFileInfoData_41C3B0
func nox_xxx_parseFileInfoData_41C3B0(_ C.int) C.int {
	return C.int(bool2int(saveFileInfoSection(cryptfile.Global(), clientSaveFileInfo()) == nil))
}
