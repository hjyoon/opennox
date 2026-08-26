package legacy

/*
#include <stdint.h>
#include <stdlib.h>

typedef struct nox_quest_journal_native {
	char name[132];
	uint32_t kind;
	uint32_t value;
	struct nox_quest_journal_native* next;
	struct nox_quest_journal_native* prev;
} nox_quest_journal_native;
*/
import "C"

import (
	"fmt"
	"math"
	"strings"
	"unsafe"

	noxflags "github.com/opennox/opennox/v1/common/flags"
	"github.com/opennox/opennox/v1/common/memmap"
	"github.com/opennox/opennox/v1/internal/cryptfile"
)

var questJournalHead500540 *C.nox_quest_journal_native

func questJournalQualifiedName5005E0(name string) string {
	if strings.ContainsRune(name, ':') {
		return name
	}
	return memmap.String(0x5D4594, 1570008) + ":" + name
}

func questJournalEntryName500540(entry *C.nox_quest_journal_native) string {
	if entry == nil {
		return ""
	}
	return C.GoString((*C.char)(unsafe.Pointer(&entry.name[0])))
}

func questJournalFind5005E0(name string) *C.nox_quest_journal_native {
	qualified := questJournalQualifiedName5005E0(name)
	for entry := questJournalHead500540; entry != nil; entry = entry.next {
		if strings.EqualFold(questJournalEntryName500540(entry), qualified) {
			return entry
		}
	}
	return nil
}

func questJournalSet500540(name string, kind, value uint32) *C.nox_quest_journal_native {
	if entry := questJournalFind5005E0(name); entry != nil {
		// GAME.EXE fixes the entry kind at creation. Calling the other setter
		// later changes only the value.
		entry.value = C.uint32_t(value)
		return entry
	}
	qualified := questJournalQualifiedName5005E0(name)
	if len(qualified) >= 132 {
		return nil
	}
	entry := (*C.nox_quest_journal_native)(C.calloc(1, C.size_t(C.sizeof_nox_quest_journal_native)))
	if entry == nil {
		return nil
	}
	nameBuf := unsafe.Slice((*byte)(unsafe.Pointer(&entry.name[0])), 132)
	copy(nameBuf, qualified)
	entry.kind = C.uint32_t(kind)
	entry.value = C.uint32_t(value)
	entry.next = questJournalHead500540
	if questJournalHead500540 != nil {
		questJournalHead500540.prev = entry
	}
	questJournalHead500540 = entry
	return entry
}

func questJournalDeleteEntry500790(entry *C.nox_quest_journal_native) {
	if entry == nil {
		return
	}
	if entry.prev != nil {
		entry.prev.next = entry.next
	} else {
		questJournalHead500540 = entry.next
	}
	if entry.next != nil {
		entry.next.prev = entry.prev
	}
	C.free(unsafe.Pointer(entry))
}

func questJournalMatch5007E0(name, pattern string) bool {
	star := strings.IndexByte(pattern, '*')
	if star < 0 {
		return strings.EqualFold(name, pattern)
	}
	// These branches intentionally preserve GAME.EXE's mixed comparison
	// rules: map/prefix matching is case-insensitive, while wildcard suffixes
	// are located with the case-sensitive strstr routine.
	if star == len(pattern)-1 {
		return len(name) >= star && strings.EqualFold(name[:star], pattern[:star])
	}
	if star == 0 {
		return strings.HasSuffix(name, pattern[1:])
	}
	colon := strings.IndexByte(pattern, ':')
	if colon < 0 || colon+1 != star || len(name) < colon+2 {
		return false
	}
	suffix := pattern[star+1:]
	suffixAt := len(name) - len(suffix)
	return suffixAt >= colon+2 &&
		strings.EqualFold(name[:colon+1], pattern[:colon+1]) &&
		strings.HasSuffix(name, suffix)
}

func questJournalDelete5007E0(pattern string) {
	qualified := questJournalQualifiedName5005E0(pattern)
	if pattern == "*:*" {
		for entry := questJournalHead500540; entry != nil; {
			next := entry.next
			questJournalDeleteEntry500790(entry)
			entry = next
		}
		return
	}
	for entry := questJournalHead500540; entry != nil; {
		next := entry.next
		if questJournalMatch5007E0(questJournalEntryName500540(entry), qualified) {
			questJournalDeleteEntry500790(entry)
		}
		entry = next
	}
}

func questJournalWriteNative500A60(cf *cryptfile.CryptFile) error {
	if cf == nil {
		return fmt.Errorf("missing quest-journal crypt file")
	}
	if err := cf.WriteU16(1); err != nil {
		return err
	}
	var count uint32
	if noxflags.HasGame(noxflags.GameModeCoop) {
		for entry := questJournalHead500540; entry != nil; entry = entry.next {
			count++
		}
	}
	if err := cf.WriteU32(count); err != nil {
		return err
	}
	if count == 0 {
		return nil
	}
	for entry := questJournalHead500540; entry != nil; entry = entry.next {
		if err := cf.WriteString8(questJournalEntryName500540(entry)); err != nil {
			return err
		}
		kind := uint32(entry.kind)
		if err := cf.WriteU32(kind); err != nil {
			return err
		}
		if kind == 0 || kind == 1 {
			if err := cf.WriteU32(uint32(entry.value)); err != nil {
				return err
			}
		}
	}
	return nil
}

func questJournalReadNative500B70(cf *cryptfile.CryptFile) error {
	if cf == nil {
		return fmt.Errorf("missing quest-journal crypt file")
	}
	questJournalDelete5007E0("*:*")
	version, err := cf.ReadU16()
	if err != nil {
		return err
	}
	if int16(version) > 1 {
		return fmt.Errorf("unsupported quest-journal version %d", version)
	}
	count, err := cf.ReadU32()
	if err != nil {
		return err
	}
	for i := uint32(0); i < count; i++ {
		name, err := cf.ReadString8()
		if err != nil {
			return err
		}
		kind, err := cf.ReadU32()
		if err != nil {
			return err
		}
		if kind != 0 && kind != 1 {
			continue
		}
		value, err := cf.ReadU32()
		if err != nil {
			return err
		}
		if questJournalSet500540(name, kind, value) == nil {
			return fmt.Errorf("cannot restore quest-journal entry %q", name)
		}
	}
	return nil
}

//export nox_xxx_journalQuestSet_500540
func nox_xxx_journalQuestSet_500540(name *C.char, value C.int) *C.char {
	entry := questJournalSet500540(GoString(name), 0, uint32(value))
	return (*C.char)(unsafe.Pointer(entry))
}

//export nox_xxx_scriptGetJournal_5005E0
func nox_xxx_scriptGetJournal_5005E0(name *C.char) *C.char {
	return (*C.char)(unsafe.Pointer(questJournalFind5005E0(GoString(name))))
}

//export nox_xxx_journalQuestSetBool_5006B0
func nox_xxx_journalQuestSetBool_5006B0(name *C.char, value C.int) *C.char {
	entry := questJournalSet500540(GoString(name), 1, uint32(value))
	return (*C.char)(unsafe.Pointer(entry))
}

//export sub_500750
func sub_500750(name *C.char) C.int {
	entry := questJournalFind5005E0(GoString(name))
	if entry == nil {
		return 0
	}
	return C.int(int32(entry.value))
}

//export sub_500770
func sub_500770(name *C.char) C.double {
	entry := questJournalFind5005E0(GoString(name))
	if entry == nil {
		return 0
	}
	return C.double(math.Float32frombits(uint32(entry.value)))
}

//export sub_500790
func sub_500790(entry unsafe.Pointer) {
	questJournalDeleteEntry500790((*C.nox_quest_journal_native)(entry))
}

//export sub_5007E0
func sub_5007E0(pattern *C.char) *C.char {
	questJournalDelete5007E0(GoString(pattern))
	return nil
}

//export sub_5009B0
func sub_5009B0(name *C.char) C.uint {
	value := GoString(name)
	if strings.ContainsRune(value, ':') {
		return C.uint(len(value) + 1)
	}
	return 0
}

//export sub_500A60
func sub_500A60() C.int {
	if err := questJournalWriteNative500A60(cryptfile.Global()); err != nil {
		return 0
	}
	return 1
}

//export sub_500B70
func sub_500B70() C.int {
	if err := questJournalReadNative500B70(cryptfile.Global()); err != nil {
		return 0
	}
	return 1
}
