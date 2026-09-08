package legacy

/*
#include <stdlib.h>
#include "quest_journal_500540.h"
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

func questJournalEqualFold5005E0(left, right string) bool {
	if len(left) != len(right) {
		return false
	}
	for i := range len(left) {
		lc, rc := left[i], right[i]
		if 'A' <= lc && lc <= 'Z' {
			lc += 'a' - 'A'
		}
		if 'A' <= rc && rc <= 'Z' {
			rc += 'a' - 'A'
		}
		if lc != rc {
			return false
		}
	}
	return true
}

func questJournalFindQualified5005E0(qualified string) *C.nox_quest_journal_native {
	for entry := questJournalHead500540; entry != nil; entry = entry.next {
		if questJournalEqualFold5005E0(questJournalEntryName500540(entry), qualified) {
			return entry
		}
	}
	return nil
}

func questJournalFind5005E0(name string) *C.nox_quest_journal_native {
	return questJournalFindQualified5005E0(questJournalQualifiedName5005E0(name))
}

// QuestJournalGetInt500750 returns the exact dword stored in a matching entry,
// interpreted as the signed value used by the script stack.
func QuestJournalGetInt500750(name string) int32 {
	entry := questJournalFind5005E0(name)
	if entry == nil {
		return 0
	}
	return int32(uint32(entry.value))
}

// QuestJournalGetFloat500770 returns the stored dword as the float32 value
// produced by GAME.EXE's x87 load/store path. Loading a signaling NaN sets its
// quiet bit while preserving its sign and payload; every other bit pattern is
// returned unchanged.
func QuestJournalGetFloat500770(name string) float32 {
	entry := questJournalFind5005E0(name)
	if entry == nil {
		return 0
	}
	bits := uint32(entry.value)
	if bits&0x7f800000 == 0x7f800000 && bits&0x007fffff != 0 {
		bits |= 0x00400000
	}
	return math.Float32frombits(bits)
}

func questJournalSetResult500540(name string, kind, value uint32) (entry, result *C.nox_quest_journal_native) {
	qualified := questJournalQualifiedName5005E0(name)
	if entry := questJournalFindQualified5005E0(qualified); entry != nil {
		// GAME.EXE fixes the entry kind at creation. Calling the other setter
		// later changes only the value.
		entry.value = C.uint32_t(value)
		return entry, entry
	}
	// GAME.EXE copies into a 132-byte scratch buffer and node name without a
	// bound. Keep exact behavior for valid C strings while rejecting the
	// overflow domain instead of corrupting the adjacent journal head.
	if len(qualified) >= 132 {
		return nil, nil
	}
	entry = (*C.nox_quest_journal_native)(C.calloc(1, C.size_t(C.sizeof_nox_quest_journal_native)))
	if entry == nil {
		return nil, nil
	}
	nameBuf := unsafe.Slice((*byte)(unsafe.Pointer(&entry.name[0])), 132)
	copy(nameBuf, qualified)
	entry.kind = C.uint32_t(kind)
	entry.value = C.uint32_t(value)
	oldHead := questJournalHead500540
	entry.next = oldHead
	if oldHead != nil {
		oldHead.prev = entry
	}
	questJournalHead500540 = entry
	// The original leaves the pre-insertion head in EAX. Consequently the
	// first successful insertion returns NULL and later insertions return the
	// previous head, even though the new entry was allocated and linked.
	return entry, oldHead
}

func questJournalSet500540(name string, kind, value uint32) *C.nox_quest_journal_native {
	entry, _ := questJournalSetResult500540(name, kind, value)
	return entry
}

func questJournalSetExportCall500540(name string, value int32) *C.nox_quest_journal_native {
	cname := C.CString(name)
	defer C.free(unsafe.Pointer(cname))
	return C.nox_xxx_journalQuestSet_500540(cname, C.int32_t(value))
}

func questJournalGetIntExportCall500750(name string) int32 {
	cname := C.CString(name)
	defer C.free(unsafe.Pointer(cname))
	return int32(C.sub_500750(cname))
}

func questJournalGetFloatExportCall500770(name string) float64 {
	cname := C.CString(name)
	defer C.free(unsafe.Pointer(cname))
	return float64(C.sub_500770(cname))
}

func questJournalDeleteEntryExportCall500790(entry *C.nox_quest_journal_native) {
	C.sub_500790(entry)
}

func questJournalDeleteEntry500790(entry *C.nox_quest_journal_native) {
	questJournalDeleteEntryContract500790(entry, questJournalDeleteEntryHooks500790[*C.nox_quest_journal_native]{
		loadPrev: func(entry *C.nox_quest_journal_native) *C.nox_quest_journal_native {
			return entry.prev
		},
		loadNext: func(entry *C.nox_quest_journal_native) *C.nox_quest_journal_native {
			return entry.next
		},
		storeNext: func(entry, next *C.nox_quest_journal_native) {
			entry.next = next
		},
		storePrev: func(entry, prev *C.nox_quest_journal_native) {
			entry.prev = prev
		},
		loadHead: func() *C.nox_quest_journal_native {
			return questJournalHead500540
		},
		storeHead: func(entry *C.nox_quest_journal_native) {
			questJournalHead500540 = entry
		},
		freeEntry: func(entry *C.nox_quest_journal_native) {
			C.free(unsafe.Pointer(entry))
		},
	})
}

// QuestJournalDelete5007E0 removes exact or wildcard-matching entries using
// GAME.EXE's mixed ASCII-insensitive prefix and case-sensitive strstr rules.
func QuestJournalDelete5007E0(pattern string) {
	questJournalDeletePatternContract5007E0(pattern, questJournalDeletePatternHooks5007E0[*C.nox_quest_journal_native]{
		qualify: func(pattern string) string {
			return questJournalQualifiedName5005E0(pattern)
		},
		findExact: func(pattern string) *C.nox_quest_journal_native {
			return questJournalFind5005E0(pattern)
		},
		loadHead: func() *C.nox_quest_journal_native {
			return questJournalHead500540
		},
		loadNext: func(entry *C.nox_quest_journal_native) *C.nox_quest_journal_native {
			return entry.next
		},
		equalFoldPrefix: func(entry *C.nox_quest_journal_native, pattern string, count int) bool {
			return questJournalASCIIPrefixEqualFold5007E0(questJournalEntryName500540(entry), pattern, count)
		},
		firstSubstringRemaining: func(entry *C.nox_quest_journal_native, start int, needle string) (int, bool) {
			name := questJournalEntryName500540(entry)
			if start >= len(name) {
				name = ""
			} else {
				name = name[start:]
			}
			return questJournalFirstSubstringRemaining5007E0(name, needle)
		},
		deleteEntry: func(entry *C.nox_quest_journal_native) {
			questJournalDeleteEntry500790(entry)
		},
	})
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
	QuestJournalDelete5007E0("*:*")
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
func nox_xxx_journalQuestSet_500540(name *C.char, value C.int32_t) *C.nox_quest_journal_native {
	_, result := questJournalSetResult500540(GoString(name), 0, uint32(value))
	return result
}

//export nox_xxx_scriptGetJournal_5005E0
func nox_xxx_scriptGetJournal_5005E0(name *C.char) *C.char {
	return (*C.char)(unsafe.Pointer(questJournalFind5005E0(GoString(name))))
}

//export nox_xxx_journalQuestSetBool_5006B0
func nox_xxx_journalQuestSetBool_5006B0(name *C.char, value C.int32_t) *C.nox_quest_journal_native {
	_, result := questJournalSetResult500540(GoString(name), 1, uint32(value))
	return result
}

//export sub_500750
func sub_500750(name *C.char) C.int32_t {
	return C.int32_t(QuestJournalGetInt500750(GoString(name)))
}

//export sub_500770
func sub_500770(name *C.char) C.double {
	return C.double(QuestJournalGetFloat500770(GoString(name)))
}

//export sub_500790
func sub_500790(entry *C.nox_quest_journal_native) {
	questJournalDeleteEntry500790(entry)
}

//export sub_5007E0
func sub_5007E0(pattern *C.char) *C.char {
	QuestJournalDelete5007E0(GoString(pattern))
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
