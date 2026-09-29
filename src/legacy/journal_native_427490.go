package legacy

/*
#include "GAME1_1.h"
#include "client__gui__guijourn.h"
*/
import "C"

import (
	"unsafe"

	"github.com/opennox/opennox/v1/server"
)

func Nox_xxx_cliBuildJournalString_469BC0() {
	C.nox_xxx_cliBuildJournalString_469BC0()
}

func journalEntryAddLegacy427500(unit *server.Object, message string, entryType uint16) *server.PlayerJournal {
	entry := GetServer().S().JournalEntryAdd427500(unit, message, entryType)
	if entry != nil && unit.ControllingPlayer().PlayerInd == server.HostPlayerIndex {
		Nox_xxx_cliBuildJournalString_469BC0()
	}
	return entry
}

//export nox_xxx_journalEntryAdd_427490
func nox_xxx_journalEntryAdd_427490(player *C.nox_playerInfo, message *C.char, entryType C.short) *C.nox_playerInfo_journal {
	entry := server.JournalEntryAdd427490(
		asPlayerS((*nox_playerInfo)(unsafe.Pointer(player))),
		GoString(message),
		uint16(entryType),
	)
	return (*C.nox_playerInfo_journal)(unsafe.Pointer(entry))
}

//export nox_xxx_comJournalEntryAdd_427500
func nox_xxx_comJournalEntryAdd_427500(unit *C.nox_object_t, message *C.char, entryType C.short) {
	journalEntryAddLegacy427500(
		asObjectS((*nox_object_t)(unsafe.Pointer(unit))),
		GoString(message),
		uint16(entryType),
	)
}

//export nox_xxx_journalEntryRemove_427590
func nox_xxx_journalEntryRemove_427590(player *C.nox_playerInfo, message *C.char) C.int {
	_, ok := server.JournalEntryRemove427590(
		asPlayerS((*nox_playerInfo)(unsafe.Pointer(player))),
		GoString(message),
	)
	return C.int(bool2int(ok))
}

//export nox_xxx_journalUpdateEntry_4276B0
func nox_xxx_journalUpdateEntry_4276B0(player *C.nox_playerInfo, message *C.char, entryType C.short) *C.nox_playerInfo_journal {
	entry := server.JournalEntryUpdate4276B0(
		asPlayerS((*nox_playerInfo)(unsafe.Pointer(player))),
		GoString(message),
		uint16(entryType),
	)
	return (*C.nox_playerInfo_journal)(unsafe.Pointer(entry))
}

//export sub_4277B0
func sub_4277B0(unit *C.nox_object_t, mask C.ushort) C.int {
	return C.int(server.JournalEntriesRemoveByMask4277B0(
		asObjectS((*nox_object_t)(unsafe.Pointer(unit))),
		uint16(mask),
	))
}
