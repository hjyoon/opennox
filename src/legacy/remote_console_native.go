package legacy

/*
#include "GAME3_2.h"
*/
import "C"

import (
	"slices"
	"unsafe"

	"github.com/opennox/opennox/v1/common/memmap"
	"github.com/opennox/opennox/v1/common/ntype"
	"github.com/opennox/opennox/v1/legacy/common/alloc"
	"github.com/opennox/opennox/v1/server"
)

type remoteConsoleGrant struct {
	player *server.Player
	token  *C.int
}

var remoteConsoleGrants []remoteConsoleGrant

func remoteConsoleGrantIndex(playerIndex int) int {
	for i, grant := range remoteConsoleGrants {
		if grant.player.Index() == playerIndex {
			return i
		}
	}
	return -1
}

//export sub_4D11A0
func sub_4D11A0() {
	*memmap.PtrUint32(0x5D4594, 1548504) = 1
}

//export sub_4D11D0
func sub_4D11D0() {
	for _, grant := range remoteConsoleGrants {
		alloc.Free(grant.token)
	}
	remoteConsoleGrants = nil
}

//export sub_4D1210
func sub_4D1210(playerIndex C.int) {
	if remoteConsoleGrantIndex(int(playerIndex)) >= 0 {
		return
	}
	player := GetServer().S().Players.ByInd(ntype.PlayerInd(playerIndex))
	if player == nil {
		return
	}
	token, free := alloc.Calloc(1, unsafe.Sizeof(C.int(0)))
	if token == nil {
		free()
		return
	}
	remoteConsoleGrants = append(remoteConsoleGrants, remoteConsoleGrant{player, (*C.int)(token)})
}

//export sub_4D1250
func sub_4D1250(playerIndex C.int) *C.int {
	i := remoteConsoleGrantIndex(int(playerIndex))
	if i < 0 {
		return nil
	}
	token := remoteConsoleGrants[i].token
	remoteConsoleGrants = slices.Delete(remoteConsoleGrants, i, i+1)
	alloc.Free(token)
	// The original ABI returns the released entry's address as a status token.
	return token
}

//export sub_4D12A0
func sub_4D12A0(playerIndex C.int) C.int {
	return C.int(bool2int(remoteConsoleGrantIndex(int(playerIndex)) >= 0))
}
