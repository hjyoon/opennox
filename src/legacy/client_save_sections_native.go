package legacy

/*
#include "GAME1_1.h"
#include "GAME3_3.h"
*/
import "C"

import (
	"encoding/binary"
	"errors"
	"fmt"
	"unsafe"

	"github.com/opennox/libs/noxnet/netmsg"

	"github.com/opennox/opennox/v1/common/flags"
	"github.com/opennox/opennox/v1/common/memmap"
	"github.com/opennox/opennox/v1/internal/cryptfile"
)

func saveClientGUISection(cf *cryptfile.CryptFile, arg unsafe.Pointer) error {
	if cf == nil {
		return errors.New("GUI save data requires an open crypt file")
	}
	version, err := cf.ReadWriteU16(3)
	if err != nil {
		return err
	}
	if int16(version) > 3 {
		return fmt.Errorf("unsupported GUI save version %d", version)
	}
	previous := cryptfile.Global()
	cryptfile.SetGlobal(cf)
	defer cryptfile.SetGlobal(previous)
	if sub_460940(arg) == 0 {
		return errors.New("read or write quickbar data")
	}
	if int16(version) >= 2 {
		row, err := cf.ReadWriteU8(byte(nox_xxx_buttonsGetSelectedRow_45E180()))
		if err != nil {
			return err
		}
		if cf.ReadOnly() {
			nox_xxx_clientUpdateButtonRow_45E110(C.int(row))
		}
		trap, err := cf.ReadWriteU8(byte(sub_4604E0()))
		if err != nil {
			return err
		}
		if cf.ReadOnly() {
			nox_client_trapSetSelect_4604B0(C.int(trap))
		}
	}
	if int16(version) >= 3 {
		player := nox_common_playerInfoGetByID_417040(int32(clientPlayerNetCode))
		order := byte(4)
		if player != nil {
			order = byte(player.field_3648)
		}
		order, err = cf.ReadWriteU8(order)
		if err != nil {
			return err
		}
		if cf.ReadOnly() && noxflags.HasGame(noxflags.GameModeCoop) && player != nil {
			nox_xxx_orderUnitLocal_500C70(C.int32_t(player.playerInd), C.int32_t(order))
		}
	}
	return nil
}

//export sub_41C280
func sub_41C280(arg unsafe.Pointer) C.int {
	return C.int(bool2int(saveClientGUISection(cryptfile.Global(), arg) == nil))
}

func saveMusicStateIO(cf *cryptfile.CryptFile, state *[4]uint32) error {
	for _, i := range [...]int{3, 2, 0, 1} {
		value, err := cf.ReadWriteU32(state[i])
		if err != nil {
			return err
		}
		state[i] = value
	}
	return nil
}

func saveClientMusicSection(cf *cryptfile.CryptFile) error {
	if cf == nil {
		return errors.New("music save data requires an open crypt file")
	}
	version, err := cf.ReadWriteU16(11)
	if err != nil {
		return err
	}
	if int16(version) > 11 {
		return fmt.Errorf("unsupported music save version %d", version)
	}
	online := noxflags.HasGame(noxflags.GameOnline)
	if int16(version) == 11 {
		enabled, err := cf.ReadWriteU8(byte(bool2int(!online)))
		if err != nil || enabled == 0 {
			return err
		}
	}
	var current [4]uint32
	var count uint32
	if !cf.ReadOnly() {
		sub_43DD10(unsafe.Pointer(&current))
		count = uint32(sub_43DB20())
	}
	if err := saveMusicStateIO(cf, &current); err != nil {
		return err
	}
	count, err = cf.ReadWriteU32(count)
	if err != nil {
		return err
	}
	const maxMusicFrames = 6
	if count > maxMusicFrames {
		return fmt.Errorf("invalid saved music count %d", count)
	}
	if cf.ReadOnly() {
		var frames [maxMusicFrames][4]uint32
		for i := range count {
			if err := saveMusicStateIO(cf, &frames[i]); err != nil {
				return err
			}
		}
		for i := range count {
			*(*[4]uint32)(game1MusicFrame(i)) = frames[i]
		}
		if !online {
			sub_43D9E0(unsafe.Pointer(&current))
			sub_43DB30(C.int(count))
		}
	} else {
		for i := range count {
			if err := saveMusicStateIO(cf, (*[4]uint32)(game1MusicFrame(i))); err != nil {
				return err
			}
		}
	}
	return nil
}

//export sub_41C780
func sub_41C780(_ C.int) C.int {
	return C.int(bool2int(saveClientMusicSection(cryptfile.Global()) == nil))
}

//export sub_41B3B0
func sub_41B3B0() C.int {
	var count C.int
	for row := range inventoryRows - 1 {
		for column := range inventoryColumns {
			count += C.int(inventoryGridCell(row, column).field_140)
		}
	}
	return count
}

//export nox_xxx_netSavePlayer_41CE00
func nox_xxx_netSavePlayer_41CE00() C.int {
	packet := [3]byte{byte(netmsg.MSG_REQUEST_SAVE_PLAYER)}
	binary.LittleEndian.PutUint16(packet[1:], uint16(clientPlayerNetCode))
	nox_xxx_netClientSend2_4E53C0(31, C.nox_important_payload_t(unsafe.Pointer(&packet[0])), C.int(len(packet)), nil, 1)
	return 1
}

//export sub_41D1A0
func sub_41D1A0(value C.int) C.int {
	*memmap.PtrInt32(0x5D4594, 527720) = int32(value)
	return value
}

//export sub_41D1B0
func sub_41D1B0() C.int {
	return C.int(memmap.Int32(0x5D4594, 527720))
}
