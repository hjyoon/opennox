package legacy

import (
	"bytes"
	"encoding/binary"
	"math"
	"runtime"
	"testing"
	"unsafe"

	playerlib "github.com/opennox/libs/player"

	noxflags "github.com/opennox/opennox/v1/common/flags"
	"github.com/opennox/opennox/v1/common/memmap"
	"github.com/opennox/opennox/v1/internal/netstr"
	"github.com/opennox/opennox/v1/legacy/common/alloc"
	"github.com/opennox/opennox/v1/server"
)

func TestPlayerKillStats425CA0NativePointers(t *testing.T) {
	oldFlags := noxflags.GetGame()
	noxflags.ResetGame()
	noxflags.SetGame(noxflags.GameOnline)
	t.Cleanup(func() {
		noxflags.ResetGame()
		noxflags.SetGame(oldFlags)
	})

	oldPlayers := Get_dword_5d4594_608316()
	oldPairs := playerKillStatsPairCount425CA0()
	Set_dword_5d4594_608316(20)
	playerKillStatsSetPairCount425CA0(3)
	pair := memmap.PtrT[[2]byte](0x5D4594, 608320+2*3)
	oldPair := *pair
	*pair = [2]byte{0xa5, 0x5a}
	t.Cleanup(func() {
		Set_dword_5d4594_608316(oldPlayers)
		playerKillStatsSetPairCount425CA0(oldPairs)
		*pair = oldPair
	})

	first, freeFirst := alloc.New(server.Player{})
	defer freeFirst()
	second, freeSecond := alloc.New(server.Player{})
	defer freeSecond()
	first.Field4648, second.Field4648 = 17, 19
	for _, player := range []*server.Player{first, second} {
		if unsafe.Sizeof(uintptr(0)) == 8 && uintptr(unsafe.Pointer(player)) <= math.MaxUint32 {
			t.Fatalf("player address %p must be above 4 GiB", player)
		}
	}

	flagPickupObserverUpdate425CA0(first, second)
	if *pair != [2]byte{17, 19} || playerKillStatsPairCount425CA0() != 4 {
		t.Fatalf("pair=%v count=%d, want [17 19]/4", *pair, playerKillStatsPairCount425CA0())
	}
	if Get_dword_5d4594_608316() != 20 || first.Field4648 != 17 || second.Field4648 != 19 {
		t.Fatal("registered players or player count were changed")
	}
	runtime.KeepAlive(first)
	runtime.KeepAlive(second)
}

func TestPlayerKillStats425CA0NativeRegistration(t *testing.T) {
	oldFlags := noxflags.GetGame()
	noxflags.ResetGame()
	noxflags.SetGame(noxflags.GameOnline)
	oldGetServer := GetServer
	// No socket is opened. The real IP service returns zero for this empty
	// host stream; the pure original matrix separately tests nonzero IPs.
	srv := &server.Server{NetStr: netstr.NewStreams(func() uint32 { return 0 })}
	bridge := &playerTrackingLegacyServer425F10{srv: srv}
	GetServer = func() Server { return bridge }
	t.Cleanup(func() {
		GetServer = oldGetServer
		noxflags.ResetGame()
		noxflags.SetGame(oldFlags)
	})

	oldCount := Get_dword_5d4594_608316()
	oldPairs := playerKillStatsPairCount425CA0()
	records := memmap.PtrT[[96]byte](0x5D4594, 600124)
	pair := memmap.PtrT[[2]byte](0x5D4594, 608320)
	oldRecords, oldPair := *records, *pair
	for i := range records {
		records[i] = 0xa5
	}
	*pair = [2]byte{0xcc, 0xcc}
	Set_dword_5d4594_608316(0)
	playerKillStatsSetPairCount425CA0(0)
	t.Cleanup(func() {
		*records, *pair = oldRecords, oldPair
		Set_dword_5d4594_608316(oldCount)
		playerKillStatsSetPairCount425CA0(oldPairs)
	})

	first, freeFirst := alloc.New(server.Player{})
	defer freeFirst()
	second, freeSecond := alloc.New(server.Player{})
	defer freeSecond()
	first.Field4648, second.Field4648 = -1, -1
	first.PlayerInd, second.PlayerInd = 2, 31
	first.Field2068, second.Field2068 = 0x12345678, 0xaabbccdd
	first.SetField2096("First")
	second.SetField2096("Second")
	first.Info().SetPlayerClass(playerlib.Wizard)
	second.Info().SetPlayerClass(playerlib.Conjurer)
	for _, player := range []*server.Player{first, second} {
		if unsafe.Sizeof(uintptr(0)) == 8 && uintptr(unsafe.Pointer(player)) <= math.MaxUint32 {
			t.Fatalf("registration player %p must be above 4 GiB", player)
		}
	}
	flagPickupObserverUpdate425CA0(first, second)
	if first.Field4648 != 0 || second.Field4648 != 1 || Get_dword_5d4594_608316() != 2 || playerKillStatsPairCount425CA0() != 1 || *pair != [2]byte{0, 1} {
		t.Fatalf("indices/count/pairs/pair=%d/%d/%d/%d/%v", first.Field4648, second.Field4648, Get_dword_5d4594_608316(), playerKillStatsPairCount425CA0(), *pair)
	}
	if string(records[:7]) != "Second\x00" || binary.LittleEndian.Uint32(records[12:16]) != 0 || binary.LittleEndian.Uint32(records[16:20]) != first.Field2068 || records[20] != byte(playerlib.Wizard) {
		t.Fatalf("first record = %x, want original second-name/host-IP destinations", records[:32])
	}
	if binary.LittleEndian.Uint32(records[48:52]) != second.Field2068 || records[52] != byte(playerlib.Conjurer) {
		t.Fatal("second player's native team/class were not recorded")
	}
	if !bytes.Equal(records[32:48], bytes.Repeat([]byte{0xa5}, 16)) || !bytes.Equal(records[64:], bytes.Repeat([]byte{0xa5}, 32)) {
		t.Fatal("second name/IP or unrelated third record were cleared")
	}
	runtime.KeepAlive(first)
	runtime.KeepAlive(second)
}
