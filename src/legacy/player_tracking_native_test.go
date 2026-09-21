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

type playerTrackingLegacyServer425F10 struct {
	Server
	srv *server.Server
}

func (s *playerTrackingLegacyServer425F10) S() *server.Server {
	return s.srv
}

func TestPlayerTrackingNativeWidth425F10(t *testing.T) {
	oldGameFlags := noxflags.GetGame()
	oldEngineFlags := noxflags.GetEngine()
	noxflags.ResetGame()
	noxflags.SetGame(noxflags.GameOnline)
	noxflags.ResetEngine()
	t.Cleanup(func() {
		noxflags.ResetGame()
		noxflags.SetGame(oldGameFlags)
		noxflags.ResetEngine()
		noxflags.SetEngine(oldEngineFlags)
	})

	srv := &server.Server{NetStr: netstr.NewStreams(func() uint32 { return 0 })}
	bridge := &playerTrackingLegacyServer425F10{srv: srv}
	oldGetServer := GetServer
	GetServer = func() Server { return bridge }
	t.Cleanup(func() { GetServer = oldGetServer })

	record := memmap.PtrT[[32]byte](0x5D4594, 600124)
	oldRecord := *record
	oldCount := Get_dword_5d4594_608316()
	*record = [32]byte{}
	Set_dword_5d4594_608316(0)
	t.Cleanup(func() {
		*record = oldRecord
		Set_dword_5d4594_608316(oldCount)
	})

	pl, freePlayer := alloc.New(server.Player{})
	defer freePlayer()
	pl.PlayerInd = server.HostPlayerIndex
	pl.Field2068 = 0x12345678
	pl.Field4648 = -1
	pl.SetField2096("ArmHost")
	pl.Info().SetPlayerClass(playerlib.Wizard)
	if unsafe.Sizeof(uintptr(0)) == 8 && uintptr(unsafe.Pointer(pl)) <= math.MaxUint32 {
		t.Fatalf("player address = %p, want a native address above 4 GiB", pl)
	}

	Sub_425F10(pl)
	if got := Get_dword_5d4594_608316(); got != 1 {
		t.Fatalf("tracking record count = %d, want 1", got)
	}
	if got := pl.Field4648; got != 0 {
		t.Fatalf("player tracking index = %d, want 0", got)
	}
	if got := bytes.TrimRight(record[:12], "\x00"); !bytes.Equal(got, []byte("ArmHost")) {
		t.Fatalf("tracking name = %q, want %q", got, "ArmHost")
	}
	if got := binary.LittleEndian.Uint32(record[16:20]); got != pl.Field2068 {
		t.Fatalf("tracking team identifier = %#x, want %#x", got, pl.Field2068)
	}
	if got := record[20]; got != byte(playerlib.Wizard) {
		t.Fatalf("tracking class = %d, want %d", got, playerlib.Wizard)
	}
	if got := record[21]; got != 1 {
		t.Fatalf("tracking active flag = %d, want 1", got)
	}
	if got := binary.LittleEndian.Uint32(record[24:28]); got == 0 {
		t.Fatal("tracking timestamp was not recorded")
	}
	if got := record[28]; got != 1 {
		t.Fatalf("tracking player status = %d, want 1", got)
	}

	Sub_425E90(pl, 0)
	Sub_425ED0(pl, 0)
	if record[21] != 0 || record[28] != 0 {
		t.Fatalf("cleared tracking flags = active:%d status:%d, want 0/0", record[21], record[28])
	}
	Sub_425E90(pl, 1)
	Sub_425ED0(pl, 1)
	if record[21] != 1 || record[28] != 1 {
		t.Fatalf("restored tracking flags = active:%d status:%d, want 1/1", record[21], record[28])
	}

	Sub_425F10(pl)
	if got := Get_dword_5d4594_608316(); got != 1 {
		t.Fatalf("duplicate registration changed record count to %d, want 1", got)
	}
	runtime.KeepAlive(pl)
}
