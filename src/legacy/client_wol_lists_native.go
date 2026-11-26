package legacy

/*
#include "GAME1_1.h"
#include <stdlib.h>

typedef const char wol_const_char_t;
*/
import "C"

import (
	"strings"
	"unsafe"

	noxflags "github.com/opennox/opennox/v1/common/flags"
	"github.com/opennox/opennox/v1/common/memmap"
	"github.com/opennox/opennox/v1/legacy/common/alloc"
)

// The original 32-bit lists use the same field order with four-byte pointers.
type wolChannel struct {
	_        [6]uint32
	State    uint32
	Messages *wolChannelMessage
	Count    uint32
	Next     *wolChannel
}

type wolChannelMessage struct {
	Text unsafe.Pointer
	_    [36]byte
	Next *wolChannelMessage
}

type wolGameListNode struct {
	Data unsafe.Pointer
	_    [16]byte
	Next *wolGameListNode
}

type wolServiceEntry struct {
	_     [20]byte
	Next  *wolServiceEntry
	Title [71]byte
	Kind  [5]byte
	Args  [128]byte
}

type wolEndpointPort = C.uint16_t

func wolChannelHead() **wolChannel {
	return &legacyPointers.wolChannelHead
}

func wolGameListHead() **wolGameListNode {
	return &legacyPointers.wolGameHead
}

func wolServiceListHead() **wolServiceEntry {
	return &legacyPointers.wolServiceHead
}

func wolFreeOwnedRecord(record unsafe.Pointer) {
	if record != nil && !alloc.TryFreePtr(record) {
		C.free(record)
	}
}

//export sub_41EC30
func sub_41EC30() C.int {
	head := *wolChannelHead()
	legacyWords.dword_5d4594_529332 = 0
	legacyWords.dword_5d4594_529336 = 0
	for channel := head; channel != nil; channel = channel.Next {
		for message := channel.Messages; message != nil; {
			next := message.Next
			wolFreeOwnedRecord(message.Text)
			wolFreeOwnedRecord(unsafe.Pointer(message))
			message = next
		}
		channel.Messages = nil
		channel.Count = 0
		channel.State = 0
	}
	result := sub_41E2F0()
	if result == 7 {
		result = sub_41DA70(sub_41E2F0(), 10)
	}
	legacyWords.dword_5d4594_529340 = 0
	return result
}

//export sub_41F4B0
func sub_41F4B0() C.int {
	for game := *wolGameListHead(); game != nil; {
		next := game.Next
		wolFreeOwnedRecord(game.Data)
		wolFreeOwnedRecord(unsafe.Pointer(game))
		game = next
	}
	legacyPointers.wolGameHead = nil
	legacyWords.dword_5d4594_531652 = 0
	legacyWords.dword_5d4594_531656 = 0
	result := sub_41E2F0()
	if result == 7 {
		result = sub_41DA70(sub_41E2F0(), 11)
	}
	return result
}

//export sub_41F790
func sub_41F790(name *C.wol_const_char_t) *C.uint32_t {
	if name == nil {
		return nil
	}
	want := alloc.GoString((*byte)(unsafe.Pointer(name)))
	for game := *wolGameListHead(); game != nil; game = game.Next {
		if alloc.GoString((*byte)(unsafe.Add(game.Data, 36))) == want {
			return (*C.uint32_t)(unsafe.Pointer(game))
		}
	}
	return nil
}

//export sub_41F800
func sub_41F800(name *C.wol_const_char_t) unsafe.Pointer {
	if game := sub_41F790(name); game != nil {
		return (*wolGameListNode)(unsafe.Pointer(game)).Data
	}
	return nil
}

//export sub_41D650
func sub_41D650() C.int {
	sub_41F800((*C.wol_const_char_t)(sub_41FA40()))
	*memmap.PtrUint32(0x5D4594, 371700) = 1
	return 0
}

//export sub_41D670
func sub_41D670(_ *C.char) C.int {
	var address [128]C.char
	var port C.uint16_t
	if sub_420230(&address[0], &port) != 0 {
		wolAbortLadderRequest()
	}
	return 0
}

//export sub_41D6C0
func sub_41D6C0() C.int {
	var address [128]C.char
	var port C.uint16_t
	if sub_420230(&address[0], &port) == 0 {
		return 0
	}
	players := &GetServer().S().Players
	split := players.Count() > 25
	for player, i := players.First(), 0; player != nil; player, i = players.Next(player), i+1 {
		// The stock second batch starts after the first 25 and the intervening player.
		if split && i == 25 {
			continue
		}
		if player.Field2096Buf[0] != 0 {
			wolAbortLadderRequest()
		}
	}
	return 1
}

func wolAbortLadderRequest() {
	sub_4200E0()
	C.abort()
}

//export sub_420230
func sub_420230(address *C.char, port *C.uint16_t) C.int {
	if address == nil || port == nil {
		return 0
	}
	for endpoint := *wolServiceListHead(); endpoint != nil; endpoint = endpoint.Next {
		if wolEqualFoldASCII(alloc.GoStringS(endpoint.Kind[:]), "LAD") && wolReadEndpoint(endpoint, address, port) {
			return 1
		}
	}
	return 0
}

//export sub_420360
func sub_420360(address *C.char, port *C.uint16_t) C.int {
	if address == nil || port == nil {
		return 0
	}
	*address, *port = 0, 0
	for endpoint := *wolServiceListHead(); endpoint != nil; endpoint = endpoint.Next {
		if !wolEqualFoldASCII(alloc.GoStringS(endpoint.Kind[:]), "GAM") {
			continue
		}
		title := "Gameres server"
		if noxflags.HasGame(noxflags.GameModeQuest) {
			title = "Quest gameres server"
		} else if !noxflags.HasGame(noxflags.GameModeCoop) {
			return 0
		}
		if wolEqualFoldASCII(alloc.GoStringS(endpoint.Title[:]), title) && wolReadEndpoint(endpoint, address, port) {
			return 1
		}
	}
	return 0
}

func wolReadEndpoint(endpoint *wolServiceEntry, address *C.char, port *C.uint16_t) bool {
	*address, *port = 0, 0
	var fields [3]string
	i := 0
	for part := range strings.SplitSeq(alloc.GoStringS(endpoint.Args[:]), ";") {
		if part == "" {
			continue
		}
		fields[i] = part
		i++
		if i == len(fields) {
			break
		}
	}
	if host := fields[1]; host != "" {
		alloc.StrCopy(unsafe.Slice((*byte)(unsafe.Pointer(address)), len(host)+1), host)
	}
	*port = C.uint16_t(wolParsePort(fields[2]))
	return *address != 0 && *port != 0
}

func wolEqualFoldASCII(first, second string) bool {
	if len(first) != len(second) {
		return false
	}
	for i := 0; i < len(first); i++ {
		a, b := first[i], second[i]
		if a >= 'A' && a <= 'Z' {
			a += 'a' - 'A'
		}
		if b >= 'A' && b <= 'Z' {
			b += 'a' - 'A'
		}
		if a != b {
			return false
		}
	}
	return true
}

func wolParsePort(text string) uint16 {
	text = strings.TrimLeft(text, " \t\n\r\v\f")
	negative := strings.HasPrefix(text, "-")
	if negative || strings.HasPrefix(text, "+") {
		text = text[1:]
	}
	var port uint16
	for i := 0; i < len(text); i++ {
		digit := text[i] - '0'
		if digit > 9 {
			break
		}
		port = port*10 + uint16(digit)
	}
	if negative {
		port = -port
	}
	return port
}
