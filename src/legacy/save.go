package legacy

/*
#include "compat.h"
#include "GAME1.h"
#include "GAME1_1.h"
#include "GAME1_3.h"
#include "GAME2.h"
#include "GAME2_1.h"
#include "GAME3_2.h"
#include "GAME3_3.h"
#include "GAME4_2.h"
extern uint32_t dword_5d4594_1049844;
extern uint32_t dword_5d4594_1563096;
void nox_xxx_unitsNewAddToList_4DAC00();
int sub_41C280(void* a1);
int nox_xxx_parseFileInfoData_41C3B0(int a1);
int sub_41C780(int a1);
*/
import "C"
import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"unsafe"

	"github.com/opennox/libs/common"
	"github.com/opennox/libs/datapath"
	"github.com/opennox/libs/ifs"
	"github.com/opennox/libs/object"

	noxflags "github.com/opennox/opennox/v1/common/flags"
	"github.com/opennox/opennox/v1/common/memmap"
	"github.com/opennox/opennox/v1/internal/cryptfile"
	"github.com/opennox/opennox/v1/server"
)

var (
	Nox_savegame_rm                      func(name string, rmDir bool) error
	Nox_client_countPlayerFiles04_4DC7D0 func() int
	Nox_xxx_gameGet_4DB1B0               func() bool
	Sub_4DCC90                           func() int
	Sub_4DB1C0                           func() unsafe.Pointer
	Sub_4DCBF0                           func(a1 int)
	Nox_xxx_serverIsClosing_446180       func() int
	Sub_4DCC10                           func(a1p *server.Object) int
	Sub_4DCFB0                           func(a1p *server.Object)
	Sub_4DD0B0                           func(a1p *server.Object)
	Nox_setSaveFileName_4DB130           func(s string)
)

//export nox_setSaveFileName_4DB130
func nox_setSaveFileName_4DB130(s *C.char) {
	Nox_setSaveFileName_4DB130(GoString(s))
}

//export nox_savegame_rm_4DBE10
func nox_savegame_rm_4DBE10(cname *C.char, rmDir_cgo int32) {
	rmDir := int(rmDir_cgo)
	if cname == nil {
		return
	}
	saveName := GoString(cname)
	Nox_savegame_rm(saveName, rmDir != 0)
}

//export nox_client_countPlayerFiles04_4DC7D0
func nox_client_countPlayerFiles04_4DC7D0() int32 {
	return int32(Nox_client_countPlayerFiles04_4DC7D0())
}

func Nox_client_countSaveFiles_4DC550() int {
	return int(nox_client_countSaveFiles_4DC550())
}

//export nox_client_countSaveFiles_4DC550
func nox_client_countSaveFiles_4DC550() C.int {
	dir := datapath.Data("Save")
	_ = ifs.Mkdir(dir)
	hasPlayerFile := func(slot string) bool {
		_, err := ifs.Stat(filepath.Join(dir, slot, "Player.plr"))
		return !os.IsNotExist(err)
	}
	count := 0
	if hasPlayerFile(common.SaveAuto) {
		count++
	}
	for attempt := 0; attempt < 13; attempt++ {
		if hasPlayerFile(fmt.Sprintf(common.SaveFormat, count)) {
			count++
		}
	}
	return C.int(count)
}

func Nox_client_countPlayerFiles02_4DC630() int {
	return int(nox_client_countPlayerFiles02_4DC630())
}

//export nox_client_countPlayerFiles02_4DC630
func nox_client_countPlayerFiles02_4DC630() C.int {
	root := datapath.Data()
	dir := filepath.Join(root, "Save")
	_ = ifs.Mkdir(dir)
	_ = ifs.Chdir(dir)
	defer ifs.Chdir(root)

	var entry C.WIN32_FIND_DATAA
	search := C.FindFirstFileA(internCStr("*.plr"), &entry)
	if uintptr(unsafe.Pointer(search)) == ^uintptr(0) {
		return 0
	}
	defer C.FindClose(search)

	var info server.SaveGameInfo
	var count C.int
	for {
		if entry.dwFileAttributes&C.FILE_ATTRIBUTE_DIRECTORY == 0 {
			name := C.GoString(&entry.cFileName[0])
			Sub_41A000(root+"\\Save\\"+name, &info)
			if info.Flags&2 != 0 {
				count++
			}
		}
		if C.FindNextFileA(search, &entry) == 0 {
			return count
		}
	}
}

//export nox_xxx_gameGet_4DB1B0
func nox_xxx_gameGet_4DB1B0() int32 { return int32(bool2int(Nox_xxx_gameGet_4DB1B0())) }

//export sub_4DCC90
func sub_4DCC90() int32 { return int32(Sub_4DCC90()) }

//export sub_4DB1C0
func sub_4DB1C0() unsafe.Pointer { return Sub_4DB1C0() }

//export sub_4DCBF0
func sub_4DCBF0(a1_cgo int32) { a1 := int(a1_cgo); Sub_4DCBF0(a1) }

//export nox_xxx_serverIsClosing_446180
func nox_xxx_serverIsClosing_446180() int32 { return int32(Nox_xxx_serverIsClosing_446180()) }

//export sub_4DCC10
func sub_4DCC10(a1p *nox_object_t) int32 { return int32(Sub_4DCC10(asObjectS(a1p))) }

func Nox_xxx_mapLoadOrSaveMB_4DCC70(value int) {
	nox_xxx_mapLoadOrSaveMB_4DCC70(C.int(value))
}

//export nox_xxx_mapLoadOrSaveMB_4DCC70
func nox_xxx_mapLoadOrSaveMB_4DCC70(value C.int) C.int {
	*memmap.PtrUint32(0x5D4594, 1563072) = uint32(value)
	return value
}

func Nox_xxx_mapLoadRequired_4DCC80() int {
	return int(nox_xxx_mapLoadRequired_4DCC80())
}

//export nox_xxx_mapLoadRequired_4DCC80
func nox_xxx_mapLoadRequired_4DCC80() C.int {
	return C.int(*memmap.PtrUint32(0x5D4594, 1563072))
}

func Nox_xxx_game_4DCCB0() int {
	return int(nox_xxx_game_4DCCB0())
}

//export nox_xxx_game_4DCCB0
func nox_xxx_game_4DCCB0() C.int {
	if !noxflags.HasGame(noxflags.GameModeCoop) {
		return 1
	}
	host := GetServer().S().Players.ByInd(31)
	if host == nil || host.PlayerUnit == nil {
		return 0
	}
	unit := host.PlayerUnit
	if sub_4DCC90() != 0 || sub4139B0Native() != 0 ||
		gameFrame()-*memmap.PtrUint32(0x5D4594, 1563068) < 30 ||
		nox_xxx_guiCursor_477600() != 0 || unit.ObjFlags.Has(object.FlagAirborne) {
		return 0
	}
	return C.int(bool2int(int32(Sub_4DCC10(unit)) != 0))
}

//export sub_4DCFB0
func sub_4DCFB0(a1p *nox_object_t) { Sub_4DCFB0(asObjectS(a1p)) }

//export sub_4DD0B0
func sub_4DD0B0(a1p *nox_object_t) { Sub_4DD0B0(asObjectS(a1p)) }
func Nox_xxx_destroyEveryChatMB_528D60() {
	GetServer().S().DestroyEveryChat528D60()
}
func Nox_xxx_quickBarClose_4606B0() {
	C.nox_xxx_quickBarClose_4606B0()
}
func Nox_xxx_mapSaveMap_51E010(a1 string, a2 int) bool {
	return C.nox_xxx_mapSaveMap_51E010(internCStr(a1), C.int(a2)) != 0
}

func Sub_41A590(cf *cryptfile.CryptFile, u *server.Object, pinfo *server.PlayerInfo) error {
	old := cryptfile.Global()
	cryptfile.SetGlobal(cf)
	defer cryptfile.SetGlobal(old)
	if cf != nil && !cf.ReadOnly() {
		return playerAttribWriteNative41A590(cf, u, pinfo)
	}
	return playerAttribReadRuntime41A590(cf, u, pinfo)
}

//export sub_41A590
func sub_41A590(unit, info unsafe.Pointer) C.int {
	err := playerAttribReadRuntime41A590(
		cryptfile.Global(),
		asObjectS((*nox_object_t)(unit)),
		(*server.PlayerInfo)(info),
	)
	return C.int(bool2int(err == nil))
}

func Sub_41AA30(cf *cryptfile.CryptFile, u *server.Object, pinfo *server.PlayerInfo) error {
	old := cryptfile.Global()
	cryptfile.SetGlobal(cf)
	defer cryptfile.SetGlobal(old)
	if cf != nil && !cf.ReadOnly() {
		return playerStatusWriteNative41AA30(cf, u, pinfo)
	}
	return playerStatusReadRuntime41AA30(cf, u)
}

//export nox_xxx_playerStatusRead_native_41AA30
func nox_xxx_playerStatusRead_native_41AA30(unit unsafe.Pointer) C.int {
	err := playerStatusReadRuntime41AA30(
		cryptfile.Global(),
		asObjectS((*nox_object_t)(unit)),
	)
	return C.int(bool2int(err == nil))
}

func Sub_41AC30(cf *cryptfile.CryptFile, u *server.Object, pinfo *server.PlayerInfo) error {
	old := cryptfile.Global()
	cryptfile.SetGlobal(cf)
	defer cryptfile.SetGlobal(old)
	if cf != nil && !cf.ReadOnly() {
		return playerInventoryWriteNative41AC30(cf, u, pinfo)
	}
	return playerInventoryReadRuntime41AC30(cf, u)
}

//export nox_xxx_playerInventoryRead_native_41AC30
func nox_xxx_playerInventoryRead_native_41AC30(unit unsafe.Pointer) C.int {
	err := playerInventoryReadRuntime41AC30(
		cryptfile.Global(),
		asObjectS((*nox_object_t)(unit)),
	)
	return C.int(bool2int(err == nil))
}

func Nox_xxx_guiFieldbook_41B420(cf *cryptfile.CryptFile, u *server.Object, pinfo *server.PlayerInfo) error {
	old := cryptfile.Global()
	cryptfile.SetGlobal(cf)
	defer cryptfile.SetGlobal(old)
	if cf != nil && !cf.ReadOnly() {
		return playerFieldbookWriteNative41B420(cf, u, pinfo)
	}
	return playerFieldbookReadRuntime41B420(cf, u)
}

//export nox_xxx_playerFieldbookRead_native_41B420
func nox_xxx_playerFieldbookRead_native_41B420(unit unsafe.Pointer) C.int {
	err := playerFieldbookReadRuntime41B420(
		cryptfile.Global(),
		asObjectS((*nox_object_t)(unit)),
	)
	return C.int(bool2int(err == nil))
}

func Nox_xxx_guiSpellbook_41B660(cf *cryptfile.CryptFile, u *server.Object, pinfo *server.PlayerInfo) error {
	old := cryptfile.Global()
	cryptfile.SetGlobal(cf)
	defer cryptfile.SetGlobal(old)
	if cf != nil && !cf.ReadOnly() {
		return playerSpellbookWriteNative41B660(cf, u, pinfo)
	}
	return playerSpellbookReadRuntime41B660(cf, u)
}

//export nox_xxx_playerSpellbookRead_native_41B660
func nox_xxx_playerSpellbookRead_native_41B660(unit unsafe.Pointer) C.int {
	err := playerSpellbookReadRuntime41B660(
		cryptfile.Global(),
		asObjectS((*nox_object_t)(unit)),
	)
	return C.int(bool2int(err == nil))
}

func Nox_xxx_guiEnchantment_41B9C0(cf *cryptfile.CryptFile, u *server.Object, pinfo *server.PlayerInfo) error {
	old := cryptfile.Global()
	cryptfile.SetGlobal(cf)
	defer cryptfile.SetGlobal(old)
	if cf != nil && !cf.ReadOnly() {
		return playerEnchantmentWriteNative41B9C0(cf, u, pinfo)
	}
	return playerEnchantmentReadRuntime41B9C0(cf, u)
}

//export nox_xxx_playerEnchantmentRead_native_41B9C0
func nox_xxx_playerEnchantmentRead_native_41B9C0(unit unsafe.Pointer) C.int {
	err := playerEnchantmentReadRuntime41B9C0(
		cryptfile.Global(),
		asObjectS((*nox_object_t)(unit)),
	)
	return C.int(bool2int(err == nil))
}

func Sub_41BEC0(cf *cryptfile.CryptFile, u *server.Object, pinfo *server.PlayerInfo) error {
	old := cryptfile.Global()
	cryptfile.SetGlobal(cf)
	defer cryptfile.SetGlobal(old)
	if cf != nil && !cf.ReadOnly() {
		return playerJournalWriteNative41BEC0(cf, u, pinfo)
	}
	return playerJournalReadRuntime41BEC0(cf, u)
}

//export nox_xxx_playerJournalRead_native_41BEC0
func nox_xxx_playerJournalRead_native_41BEC0(unit unsafe.Pointer) C.int {
	err := playerJournalReadRuntime41BEC0(
		cryptfile.Global(),
		asObjectS((*nox_object_t)(unit)),
	)
	return C.int(bool2int(err == nil))
}

func Sub_41C080(cf *cryptfile.CryptFile, u *server.Object, pinfo *server.PlayerInfo) error {
	old := cryptfile.Global()
	cryptfile.SetGlobal(cf)
	defer cryptfile.SetGlobal(old)
	if cf != nil && !cf.ReadOnly() {
		return playerGameWriteNative41C080(cf, u, pinfo)
	}
	return playerGameReadRuntime41C080(cf, u)
}

//export nox_xxx_playerGameRead_native_41C080
func nox_xxx_playerGameRead_native_41C080(unit unsafe.Pointer) C.int {
	err := playerGameReadRuntime41C080(
		cryptfile.Global(),
		asObjectS((*nox_object_t)(unit)),
	)
	return C.int(bool2int(err == nil))
}

func Sub_41C280(cf *cryptfile.CryptFile) error {
	old := cryptfile.Global()
	cryptfile.SetGlobal(cf)
	defer cryptfile.SetGlobal(old)
	if C.sub_41C280(nil) == 0 {
		return errors.New("failed")
	}
	return nil
}

func Nox_xxx_parseFileInfoData_41C3B0(cf *cryptfile.CryptFile) error {
	old := cryptfile.Global()
	cryptfile.SetGlobal(cf)
	defer cryptfile.SetGlobal(old)
	if C.nox_xxx_parseFileInfoData_41C3B0(0) == 0 {
		return errors.New("failed")
	}
	return nil
}

func Sub_41C780(cf *cryptfile.CryptFile) error {
	old := cryptfile.Global()
	cryptfile.SetGlobal(cf)
	defer cryptfile.SetGlobal(old)
	if C.sub_41C780(0) == 0 {
		return errors.New("failed")
	}
	return nil
}
