package legacy

/*
#include "GAME2_3.h"

*/
import "C"

import "unsafe"

const (
	clientAllyHealthOffset = 1200916
	clientAllyHealthCount  = 32
)

type clientAllyHealth struct {
	code      uint32
	alternate byte
	_         byte
	current   uint16
	maximum   uint16
	_         uint16
	active    uint32
}

var _ = [1]struct{}{}[unsafe.Sizeof(clientAllyHealth{})-16]

func clientAllyHealthRecords() *[clientAllyHealthCount]clientAllyHealth {
	return (*[clientAllyHealthCount]clientAllyHealth)(g2ServerMem(clientAllyHealthOffset))
}

func clientAllyHealthFind(code uint32) *clientAllyHealth {
	tab := clientAllyHealthRecords()
	for i := range tab {
		if entry := &tab[i]; entry.active != 0 && entry.code == code {
			return entry
		}
	}
	return nil
}

func clientAllyHealthFree() *clientAllyHealth {
	tab := clientAllyHealthRecords()
	for i := range tab {
		if entry := &tab[i]; entry.active == 0 {
			return entry
		}
	}
	return nil
}

//export sub_494FF0
func sub_494FF0() *C.char {
	return (*C.char)(unsafe.Pointer(clientAllyHealthFree()))
}

//export sub_495020
func sub_495020(code C.int) *C.char {
	return (*C.char)(unsafe.Pointer(clientAllyHealthFind(uint32(code))))
}

//export sub_495060
func sub_495060(code C.int, current, maximum C.short) C.int {
	for _, record := range clientAllyHealthRecords() {
		if record.active == 1 && record.code == uint32(code) {
			return 1
		}
	}
	record := clientAllyHealthFree()
	if record == nil {
		return 0
	}
	record.code, record.alternate, record.active = uint32(code), 0, 1
	record.current, record.maximum = uint16(current), uint16(maximum)
	return 1
}

//export sub_4950C0
func sub_4950C0(code C.int) C.int {
	record := clientAllyHealthFind(uint32(code))
	if record == nil {
		return 0
	}
	record.active = 0
	return 1
}

//export sub_4950F0
func sub_4950F0(code C.int, alternate C.char) C.int {
	record := clientAllyHealthFind(uint32(code))
	if record == nil {
		return 0
	}
	record.alternate = byte(alternate)
	return 1
}

//export sub_495120
func sub_495120(code C.int, current, maximum C.short) C.int {
	record := clientAllyHealthFind(uint32(code))
	if record == nil {
		return 0
	}
	record.current, record.maximum = uint16(current), uint16(maximum)
	return 1
}

//export sub_495150
func sub_495150(code C.int, current C.short) C.int {
	record := clientAllyHealthFind(uint32(code))
	if record == nil {
		return 0
	}
	record.current = uint16(current)
	return 1
}

//export sub_495180
func sub_495180(code C.int, current, maximum *C.uint16_t, alternate *C.uint8_t) C.int {
	record := clientAllyHealthFind(uint32(code))
	if record == nil {
		return 0
	}
	*current, *maximum, *alternate = C.uint16_t(record.current), C.uint16_t(record.maximum), C.uint8_t(record.alternate)
	return 1
}

//export sub_4951C0
func sub_4951C0() *C.char {
	tab := clientAllyHealthRecords()
	for i := range tab {
		record := &tab[i]
		record.code, record.alternate, record.current, record.maximum, record.active = 0, 0, 0, 0, 0
	}
	return (*C.char)(g2ServerMem(clientAllyHealthOffset + unsafe.Sizeof(*tab) + unsafe.Offsetof(tab[0].current)))
}

//export nox_xxx_unitSpriteCheckAlly_4951F0
func nox_xxx_unitSpriteCheckAlly_4951F0(code C.int) C.int {
	return C.int(bool2int(clientAllyHealthFind(uint32(code)) != nil))
}

//export sub_494F00
func sub_494F00() C.int {
	names := [...]string{"Spark", "BlueSpark", "YellowSpark", "CyanSpark", "GreenSpark", "Puff"}
	things := GetClient().Cli().Things
	for index, name := range names {
		id := uint32(things.IndByID(name))
		if index == 1 {
			legacyWords.dword_5d4594_1200776 = C.uint32_t(id)
		} else {
			*g2ServerWord(1200772 + 4*uintptr(index)) = id
		}
		if id == 0 {
			return 0
		}
	}
	for index, name := range names[:5] {
		id := uint32(things.IndByID(name))
		*g2ServerWord(1200812 + 4*uintptr(index)) = id
		if id == 0 {
			return 0
		}
	}
	violet := things.IndByID("VioletSpark")
	legacyWords.dword_5d4594_1200796 = C.uint32_t(violet)
	return C.int(bool2int(violet != 0))
}
