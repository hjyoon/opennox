package legacy

/*
#include "GAME1_1.h"
*/
import "C"

import (
	"encoding/binary"
	"fmt"
	"io"
	"unsafe"

	crypt "github.com/opennox/noxcrypt"

	"github.com/opennox/opennox/v1/common/flags"
	"github.com/opennox/opennox/v1/common/memmap"
	"github.com/opennox/opennox/v1/internal/binfile"
	"github.com/opennox/opennox/v1/internal/cryptfile"
)

var serverPlayerSaveReaders = map[uint32]func(unsafe.Pointer, unsafe.Pointer) C.int{
	2:    sub_41A590,
	3:    sub_41AA30,
	4:    sub_41AC30,
	5:    nox_xxx_guiSpellbook_41B660,
	6:    nox_xxx_guiEnchantment_41B9C0,
	8:    nox_xxx_guiFieldbook_41B420,
	9:    sub_41BEC0,
	0x0a: sub_41C080,
	0x0b: func(_, _ unsafe.Pointer) C.int {
		return C.int(cryptfile.Global().ReadWriteAlign())
	},
}

func readServerPlayerSave(cf *cryptfile.CryptFile, unit *nox_object_t, info unsafe.Pointer) error {
	for {
		id, size, err := readClientSaveHeader(cf)
		if err != nil || id == 0 {
			return err
		}
		reader, known := serverPlayerSaveReaders[id]
		if !known {
			if err := cf.Seek(size, io.SeekCurrent); err != nil {
				return err
			}
			continue
		}
		if reader(unsafe.Pointer(unit), info) == 0 {
			return fmt.Errorf("read player save section %d", id)
		}
	}
}

func loadServerPlayerSave(path string, index int32) error {
	player := nox_common_playerInfoFromNum_417090(index)
	if player == nil || player.playerUnit == nil {
		return fmt.Errorf("no player unit for save index %d", index)
	}
	unit := player.playerUnit
	if err := cryptfile.OpenGlobal(path, cryptfile.ReadOnly, crypt.SaveKey); err != nil {
		return err
	}
	cf := cryptfile.Global()
	if noxflags.HasGame(noxflags.GameModeCoop) {
		nox_xxx_set_god_4EF500(0)
	}
	hp := memmap.PtrUint32(0x5D4594, 527696)
	mp := memmap.PtrUint32(0x5D4594, 527700)
	*hp = uint32(uint16(nox_xxx_unitGetHP_4EE780(unit)))
	*mp = uint32(uint16(nox_xxx_unitGetOldMana_4EEC80(unit)))
	sub_4EFF10(unit)
	sub_419E10(unit, 1)

	if err := readServerPlayerSave(cf, unit, unsafe.Pointer(asPlayerS(player).Info())); err != nil {
		for item := asObjectS(unit).InvFirstItem; item != nil; {
			next := item.InvNextItem
			nox_xxx_delayedDeleteObject_4E5CC0(asObjectC(item))
			item = next
		}
		sub_419E10(unit, 0)
		cryptfile.Close()
		return err
	}
	cryptfile.Close()
	sub_4EF140(unit)
	maximumHP := uint32(uint16(nox_xxx_unitGetMaxHP_4EE7A0(unit)))
	nox_xxx_unitDamageClear_4EE5E0(unit, C.int(maximumHP-*hp))
	maximumMana := uint32(uint16(nox_xxx_playerGetMaxMana_4EECB0(unit)))
	nox_xxx_playerManaSub_4EEBF0(unit, C.int32_t(maximumMana-*mp))
	nox_xxx_playerHP_4EE730(unit)
	sub_419E10(unit, 0)
	return nil
}

//export nox_xxx_cliPlrInfoLoadFromFile_41A2E0
func nox_xxx_cliPlrInfoLoadFromFile_41A2E0(path *C.char, index C.int) C.int {
	return C.int(bool2int(loadServerPlayerSave(GoString(path), int32(index)) == nil))
}

func serverPlayerSaveWire(path string) ([]byte, error) {
	f, err := binfile.BinfileOpen(path, binfile.ReadOnly)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	if err := f.SetKey(crypt.SaveKey); err != nil {
		return nil, err
	}
	stat, err := f.File.Stat()
	if err != nil {
		return nil, err
	}
	var word [4]byte
	var wire []byte
	for {
		if _, err := io.ReadFull(f, word[:]); err != nil {
			return nil, err
		}
		id := binary.LittleEndian.Uint32(word[:])
		if id == 0 {
			return binary.LittleEndian.AppendUint32(wire, 0), nil
		}
		start, err := f.Seek(0, io.SeekCurrent)
		if err != nil {
			return nil, err
		}
		if n, err := f.ReadAligned(word[:]); err != nil {
			return nil, err
		} else if n != len(word) {
			return nil, io.ErrUnexpectedEOF
		}
		end, err := f.Seek(0, io.SeekCurrent)
		if err != nil {
			return nil, err
		}
		size := int64(int32(binary.LittleEndian.Uint32(word[:])))
		if size < 0 || size > stat.Size()-end || end-start < 4 {
			return nil, fmt.Errorf("invalid player save section %d length %d", id, size)
		}
		if _, keep := serverPlayerSaveReaders[id]; !keep {
			if err := f.FileSeek(size, io.SeekCurrent); err != nil {
				return nil, err
			}
			continue
		}
		header, offset := int(end-start), len(wire)
		wire = append(wire, make([]byte, header+4+int(size))...)
		binary.LittleEndian.PutUint32(wire[offset:], id)
		binary.LittleEndian.PutUint32(wire[offset+header-4:], uint32(size))
		if _, err := io.ReadFull(f, wire[offset+header+4:]); err != nil {
			return nil, err
		}
	}
}

//export sub_41CAC0
func sub_41CAC0(path *C.char, output unsafe.Pointer) {
	data, err := serverPlayerSaveWire(GoString(path))
	if err == nil && output != nil {
		copy(unsafe.Slice((*byte)(output), len(data)), data)
	}
}
