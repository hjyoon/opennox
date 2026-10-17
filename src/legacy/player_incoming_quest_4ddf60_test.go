package legacy

import (
	"bytes"
	"encoding/binary"
	"math"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
	"unsafe"

	"github.com/opennox/opennox/v1/legacy/common/alloc"
	"github.com/opennox/opennox/v1/server"
)

func TestIncomingQuest4DDF60CEntryPreservesNativeUpdate(t *testing.T) {
	unit, freeUnit := alloc.New(server.Object{})
	t.Cleanup(freeUnit)
	update, freeUpdate := alloc.New(server.PlayerUpdateData{})
	t.Cleanup(freeUpdate)
	player, freePlayer := alloc.New(server.Player{})
	t.Cleanup(freePlayer)
	unit.UpdateData = unsafe.Pointer(update)
	update.Player = player
	update.EquippedWeapon = unit
	update.ExtraLives = 0x11223344
	update.Field137 = 0x55667788
	for i := range update.RespawnMarkers {
		update.RespawnMarkers[i] = byte(0x80 + i)
		update.QuestPlayerFlagsA[i] = byte(0x40 + i)
		update.QuestPlayerFlagsB[i] = byte(0x20 + i)
		update.QuestPlayerState[i] = 0x12340000 + uint32(i)
	}
	if unsafe.Sizeof(uintptr(0)) == 8 {
		if uintptr(unsafe.Pointer(unit)) <= math.MaxUint32 || uintptr(unsafe.Pointer(update)) <= math.MaxUint32 {
			t.Fatalf("fixture needs native addresses: unit=%p update=%p", unit, update)
		}
		if unsafe.Offsetof(update.Field138) != 648 || unsafe.Offsetof(update.RespawnMarkers)+4 != 552 {
			t.Fatal("native layout no longer exposes the original pending-load overwrite")
		}
	}
	for _, initial := range []uint32{0, 1, 0x87654321, math.MaxUint32} {
		update.Field138 = initial
		actual := unsafe.Slice((*byte)(unsafe.Pointer(update)), int(unsafe.Sizeof(*update)))
		want := append([]byte(nil), actual...)
		binary.LittleEndian.PutUint32(want[int(unsafe.Offsetof(update.Field138)):], 1)
		unitBytes := unsafe.Slice((*byte)(unsafe.Pointer(unit)), int(unsafe.Sizeof(*unit)))
		playerBytes := unsafe.Slice((*byte)(unsafe.Pointer(player)), int(unsafe.Sizeof(*player)))
		unitBefore := append([]byte(nil), unitBytes...)
		playerBefore := append([]byte(nil), playerBytes...)
		playerIncomingQuestCEntry4DDF60(unit)
		if !bytes.Equal(actual, want) || !bytes.Equal(unitBytes, unitBefore) || !bytes.Equal(playerBytes, playerBefore) {
			t.Fatalf("pending-load CGo store changed unrelated state: initial=%#x pending=%#x", initial, update.Field138)
		}
	}
	runtime.KeepAlive(unit)
	runtime.KeepAlive(update)
	runtime.KeepAlive(player)
}

// Compile the complete active C producer with its real headers. Guard controls
// cover Quest/non-Quest, host/non-host and a missing unit at the original store;
// the helper's actual native CGo behavior is checked separately above.
func TestIncomingQuest4DDF60CompleteProducer(t *testing.T) {
	compiler, err := exec.LookPath("cc")
	if err != nil {
		t.Fatal(err)
	}
	fixture, err := os.ReadFile("testdata/player_incoming_quest_4ddf60.c")
	if err != nil {
		t.Fatal(err)
	}
	const marker = "/* COMPLETE_INCOMING_BODY */"
	if strings.Count(string(fixture), marker) != 1 {
		t.Fatal("incoming Quest fixture body marker is ambiguous")
	}
	source := strings.Replace(string(fixture), marker,
		teamReportBody210(t, "GAME3_2.c", "int nox_xxx_netPlayerIncomingServ_4DDF60("), 1)
	dir, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	temp := t.TempDir()
	file, binary := filepath.Join(temp, "incoming-quest.c"), filepath.Join(temp, "incoming-quest")
	if err := os.WriteFile(file, []byte(source), 0600); err != nil {
		t.Fatal(err)
	}
	args := []string{"-std=gnu11", "-O2", "-g", "-fshort-wchar", "-fsigned-char", "-fno-strict-aliasing", "-fno-strict-overflow", "-Wno-incompatible-pointer-types", "-Wno-int-conversion", "-I", dir, file, "-o", binary}
	if runtime.GOARCH == "386" {
		args = append([]string{"-m32"}, args...)
	}
	if output, err := exec.Command(compiler, args...).CombinedOutput(); err != nil {
		t.Fatalf("compile complete incoming Quest producer: %v\n%s", err, output)
	}
	if output, err := exec.Command(binary).CombinedOutput(); err != nil {
		t.Fatalf("complete incoming Quest producer: %v\n%s", err, output)
	} else {
		t.Log(strings.TrimSpace(string(output)))
	}
}
