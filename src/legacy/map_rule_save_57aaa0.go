package legacy

/*
#include "GAME1_1.h"
#include "GAME5_2.h"
#include "cgo_const.h"

extern uint32_t dword_5d4594_2650652;
*/
import "C"

import (
	"encoding/binary"
	"fmt"
	"unsafe"

	"github.com/opennox/libs/ifs"
	"github.com/opennox/libs/spell"
	"github.com/opennox/libs/things"
	"github.com/opennox/opennox/v1/internal/binfile"
	"github.com/opennox/opennox/v1/legacy/common/alloc"
	"github.com/opennox/opennox/v1/server"
)

//export sub_57AAA0
func sub_57AAA0(filename *C.nox_cgo_const_char, data *C.char, ruleList *C.int) (result C.char) {
	settings := (*server.Settings2)(unsafe.Pointer(data))
	mode := settings.Field52
	if status := int8(mode); status < 0 {
		return C.char(status)
	}
	mapName := alloc.GoStringS(settings.Field0[:8])
	file, err := ifs.Create("maps\\" + mapName + "\\" + GoString((*C.char)(filename)))
	if err != nil {
		return 0
	}
	out := binfile.NewTextFile(file)
	defer func() {
		if err := out.Close(); err != nil {
			result = -1
		}
	}()

	var online, mapDefaults server.Settings2
	compareDefaults := C.dword_5d4594_2650652 != 0
	if compareDefaults {
		online.Field0, mapDefaults.Field0 = settings.Field0, settings.Field0
		C.sub_57A1E0((*C.int)(unsafe.Pointer(&online)), nil, nil, 4, C.short(mode))
		C.sub_57A1E0((*C.int)(unsafe.Pointer(&mapDefaults)), nil, nil, 3, C.short(mode))
	}

	rules := (*C.nox_list_item_t)(unsafe.Pointer(ruleList))
	if rules != nil {
		for link := C.nox_common_list_getFirstSafe_425890(rules); link != nil; link = C.nox_common_list_getNextSafe_4258A0(link) {
			item := (*C.nox_rule_list_item_t)(unsafe.Pointer(link))
			// The legacy rule parser widens file bytes. Its formatter writes
			// the low byte of each code unit back into the file.
			var line [len(item.text) + 1]byte
			n := 0
			for _, code := range item.text[:] {
				if code == 0 {
					break
				}
				line[n] = byte(code)
				n++
			}
			line[n] = '\n'
			_, _ = out.WriteString(alloc.GoStringS(line[:n+1]))
		}
	}

	_, _ = out.WriteString(GoString(C.sub_57A1B0(C.short(mode))))
	_, _ = out.WriteString("\n")
	world := GetServer().S()
	const classes = things.SpellClassAny | things.SpellClassWizard | things.SpellClassConjurer
	for id := spell.ID(1); id < server.SpellsMax; id++ {
		if !world.Spells.DefByInd(id).IsValid() || mapRuleAllowsSpell57AAA0(&settings.Field24, id) ||
			world.Spells.Flags(id)&classes == 0 {
			continue
		}
		if compareDefaults && !mapRuleAllowsSpell57AAA0(&online.Field24, id) && mapRuleAllowsSpell57AAA0(&mapDefaults.Field24, id) {
			continue
		}
		_, _ = fmt.Fprintf(out, "set spell \"%s\" off\n", id)
	}
	for i := range 26 {
		bit := uint32(1) << i
		if settings.Field48&bit != 0 {
			continue
		}
		if name := world.Armor.Sub_415E40(bit); name != "" {
			_, _ = fmt.Fprintf(out, "set Armor \"%s\" off\n", name)
		}
	}
	weapons := binary.LittleEndian.Uint32(settings.Field44[:])
	for i := range 27 {
		bit := uint32(1) << i
		if weapons&bit != 0 {
			continue
		}
		if name := world.Weapons.Sub_4159B0(bit); name != "" {
			_, _ = fmt.Fprintf(out, "set weapon \"%s\" off\n", name)
		}
	}
	return 0
}

func mapRuleAllowsSpell57AAA0(settings *server.Settings3, id spell.ID) bool {
	return settings.Vals[id/32]&(uint32(1)<<uint(id%32)) != 0
}
