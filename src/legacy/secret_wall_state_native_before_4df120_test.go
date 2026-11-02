package legacy

import (
	"bytes"
	"os"
	"os/exec"
	"strconv"
	"testing"
	"unsafe"

	"github.com/opennox/libs/object"
	"github.com/opennox/libs/strman"
	"github.com/opennox/opennox/v1/legacy/common/alloc"
	"github.com/opennox/opennox/v1/server"
)

func TestSecretWallStateNativeLayout4DF120(t *testing.T) {
	wall, unit, update, player := server.Wall{}, server.Object{}, server.PlayerUpdateData{}, server.Player{}
	prefix := unsafe.Offsetof(unit.ScriptPickup) + unsafe.Sizeof(unit.ScriptPickup)
	want := [8]uintptr{unsafe.Offsetof(wall.Field10), unsafe.Sizeof(wall), unsafe.Offsetof(unit.ObjClass), unsafe.Offsetof(unit.UpdateData), prefix, unsafe.Offsetof(update.Player), unsafe.Offsetof(player.PlayerInd), unsafe.Offsetof(update.QuestWarpGate) + unsafe.Sizeof(update.QuestWarpGate)}
	if got := secretWallOriginalLayout4DF120(); got != want {
		t.Fatal("independent C/native wall, object, update and player layouts", got, want)
	}
	if unsafe.Sizeof(unit) > 0x800 || unsafe.Sizeof(update) > 0x800 || unsafe.Sizeof(player) > 0x2000 {
		t.Fatal("guarded original-comparison record strides are too small")
	}
}

type secretWallLegacyServer4DF120 struct {
	Server
	native *server.Server
}

func (s *secretWallLegacyServer4DF120) S() *server.Server { return s.native }

func secretWallLivePlayersFixture4DF120(t *testing.T) (*server.Server, *server.Wall, [2]*server.Object) {
	t.Helper()
	s := server.New(nil, nil, strman.New())
	t.Cleanup(s.Close)
	old := GetServer
	GetServer = func() Server { return &secretWallLegacyServer4DF120{native: s} }
	t.Cleanup(func() { GetServer = old })
	wall, freeWall := alloc.New(server.Wall{})
	t.Cleanup(freeWall)
	wall.Field10 = 0x9234
	var units [2]*server.Object
	for i := 0; i < 3; i++ {
		pl := s.Players.NewRaw(500 + i)
		if pl == nil {
			t.Fatal("cannot create native player slots")
		}
		if i == 1 {
			continue // The real first/next services skip active slots without units.
		}
		unit, freeUnit := alloc.New(server.Object{})
		update, freeUpdate := alloc.New(server.PlayerUpdateData{})
		t.Cleanup(freeUnit)
		t.Cleanup(freeUpdate)
		unit.ObjClass, update.Player = object.ClassPlayer, pl
		unit.UpdateData, pl.PlayerUnit = unsafe.Pointer(update), unit
		units[i/2] = unit
	}
	return s, wall, units
}

func TestSecretWallStateActualCBefore4DF120(t *testing.T) {
	if os.Getenv("OPENNOX_SECRET_WALL_C_BEFORE") != "1" {
		return
	}
	if raw := os.Getenv("OPENNOX_SECRET_WALL_C_CHILD"); raw != "" {
		entry, err := strconv.Atoi(raw)
		if err != nil || entry < 0 || entry > 1 {
			t.Fatal("invalid child entry", raw)
		}
		TestSecretWallStateNativeLayout4DF120(t)
		s, wall, units := secretWallLivePlayersFixture4DF120(t)
		if s.Players.FirstUnit() != units[0] || s.Players.NextUnit(units[0]) != units[1] || s.Players.NextUnit(units[1]) != nil {
			t.Fatal("real ordinary native player traversal")
		}
		type record struct {
			ptr  unsafe.Pointer
			size uintptr
		}
		records := []record{{unsafe.Pointer(wall), unsafe.Sizeof(*wall)}}
		for _, unit := range units {
			update := (*server.PlayerUpdateData)(unit.UpdateData)
			records = append(records, record{unsafe.Pointer(unit), unsafe.Sizeof(*unit)}, record{unit.UpdateData, unsafe.Sizeof(*update)}, record{unsafe.Pointer(update.Player), unsafe.Sizeof(*update.Player)})
		}
		before := make([][]byte, len(records))
		for i, r := range records {
			if unsafe.Sizeof(uintptr(0)) != 8 || uintptr(r.ptr) <= 0xffffffff {
				t.Fatal("before qualification requires actual native records above4GiB")
			}
			before[i] = bytes.Clone(unsafe.Slice((*byte)(r.ptr), int(r.size)))
		}
		if secretWallOriginalFirst4DF120() != units[0] || secretWallOriginalNext4DF120(units[0]) != units[1] || secretWallOriginalNext4DF120(units[1]) != nil {
			t.Fatal("actual existing typed C first/next services lose native identity")
		}
		for _, unit := range units {
			id, recipient := secretWallOriginalControls4DF120(wall, unit)
			if id != wall.Field10 || recipient != (*server.PlayerUpdateData)(unit.UpdateData).Player.PlayerInd {
				t.Fatal("typed C wall and complete object/update/player pointer-chain controls")
			}
		}
		for i, r := range records {
			if !bytes.Equal(before[i], unsafe.Slice((*byte)(r.ptr), int(r.size))) {
				t.Fatal("whole valid native record changed during controls", i)
			}
		}
		if secretWallOriginalTyped4DF120(entry) {
			t.Fatal("unchanged old public C implementation is required")
		}
		t.Logf("SECRET_WALL_REAL_NATIVE_CONTROLS entry=%d wall=%p first=%p last=%p full-records=true real-ordinary-and-C-traversal=true", entry, wall, units[0], units[1])
		secretWallOriginalC4DF120(entry, wall)
		t.Fatal("unported raw C unexpectedly returned")
	}
	for entry := 0; entry < 2; entry++ {
		t.Run(strconv.Itoa(entry), func(t *testing.T) {
			cmd := exec.Command(os.Args[0], "-test.v", "-test.run=^TestSecretWallStateActualCBefore4DF120$")
			cmd.Env = append(os.Environ(), "OPENNOX_SECRET_WALL_C_CHILD="+strconv.Itoa(entry))
			out, err := cmd.CombinedOutput()
			t.Logf("%s", out)
			if err == nil || !bytes.Contains(out, []byte("SECRET_WALL_REAL_NATIVE_CONTROLS")) || !bytes.Contains(out, []byte("SIGSEGV")) || !bytes.Contains(out, []byte("_Cfunc_secret_wall_original_public_4df120")) {
				t.Fatal("actual C must fault after independent valid native controls", err)
			}
		})
	}
}
