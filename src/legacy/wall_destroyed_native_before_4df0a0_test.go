package legacy

import (
	"bytes"
	"github.com/opennox/opennox/v1/server"
	"os"
	"os/exec"
	"strconv"
	"testing"
	"unsafe"
)

func TestWallDestroyedActualCBefore4DF0A0(t *testing.T) {
	if os.Getenv("OPENNOX_WALL_DESTROYED_C_BEFORE") != "1" {
		return
	}
	if raw := os.Getenv("OPENNOX_WALL_DESTROYED_C_CHILD"); raw != "" {
		recipient, err := strconv.Atoi(raw)
		if err != nil || (recipient != 0 && recipient != 32) {
			t.Fatal("invalid child recipient", raw)
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
		if wallDestroyedOriginalTyped4DF0A0() {
			t.Fatal("unchanged old public C implementation is required")
		}
		t.Logf("WALL_DESTROYED_REAL_NATIVE_CONTROLS recipient=%d wall=%p first=%p last=%p full-records=true real-ordinary-and-C-traversal=true", recipient, wall, units[0], units[1])
		wallDestroyedOriginalC4DF0A0(wall, int32(recipient))
		t.Fatal("unported raw C unexpectedly returned")
	}
	for _, recipient := range []int{0, 32} {
		t.Run(strconv.Itoa(recipient), func(t *testing.T) {
			cmd := exec.Command(os.Args[0], "-test.v", "-test.run=^TestWallDestroyedActualCBefore4DF0A0$")
			cmd.Env = append(os.Environ(), "OPENNOX_WALL_DESTROYED_C_CHILD="+strconv.Itoa(recipient))
			out, err := cmd.CombinedOutput()
			t.Logf("%s", out)
			if err == nil || !bytes.Contains(out, []byte("WALL_DESTROYED_REAL_NATIVE_CONTROLS")) || !bytes.Contains(out, []byte("SIGSEGV")) || !bytes.Contains(out, []byte("_Cfunc_wall_destroyed_original_public_4df0a0")) {
				t.Fatal("actual C must fault after independent valid native controls", err)
			}
		})
	}
}
