package legacy

import (
	"bytes"
	"fmt"
	"math"
	"os"
	"os/exec"
	"strconv"
	"testing"
	"unsafe"

	"github.com/opennox/libs/object"
	"github.com/opennox/libs/types"
	"github.com/opennox/opennox/v1/common/ntype"
	"github.com/opennox/opennox/v1/internal/netlist"
	"github.com/opennox/opennox/v1/legacy/common/alloc"
	"github.com/opennox/opennox/v1/server"
)

func TestMonsterDeathExplosionNativeLayout549D80(t *testing.T) {
	obj := server.Object{}
	prefix := unsafe.Offsetof(obj.ScriptPickup) + unsafe.Sizeof(obj.ScriptPickup)
	want := [4]uintptr{unsafe.Offsetof(obj.PosVec) + unsafe.Offsetof(obj.PosVec.X), unsafe.Offsetof(obj.PosVec) + unsafe.Offsetof(obj.PosVec.Y), prefix, prefix}
	if got := monsterDeathOriginalLayout549D80(); got != want {
		t.Fatal("complete native C prefix and independently declared Go point layout", got, want)
	}
}

func monsterDeathNativeBeforeFixture549D80(t *testing.T) (*server.Server, *server.Object, *server.Player, []byte) {
	t.Helper()
	s, unit, update, missile := monsterMainBlockLegacyFixture547210(t)
	s.NetList.Init()
	t.Cleanup(s.NetList.Free)
	pl := s.Players.NewRaw(1001)
	if pl == nil {
		t.Fatal("cannot allocate native receiver")
	}
	pu, freePU := alloc.NewHigh(server.PlayerUpdateData{})
	t.Cleanup(freePU)
	pu.Player = pl
	receiver, freeReceiver := alloc.NewHigh(server.Object{})
	t.Cleanup(freeReceiver)
	*receiver = server.Object{ObjClass: object.ClassPlayer, PosVec: unit.PosVec, UpdateData: unsafe.Pointer(pu)}
	pl.PlayerUnit, pl.Field10, pl.Field12 = receiver, 100, 100
	if got := monsterDeathOriginalControls549D80(unit); got != [3]uintptr{uintptr(unsafe.Pointer(unit)), uintptr(unsafe.Pointer(&unit.PosVec.X)), uintptr(unsafe.Pointer(&unit.PosVec.Y))} {
		t.Fatal("typed C full native unit/point identity", got)
	}
	records := []struct {
		ptr  unsafe.Pointer
		size uintptr
	}{{unsafe.Pointer(unit), unsafe.Sizeof(*unit)}, {unsafe.Pointer(update), unsafe.Sizeof(*update)}, {unsafe.Pointer(unit.HealthData), unsafe.Sizeof(*unit.HealthData)}, {unsafe.Pointer(missile), unsafe.Sizeof(*missile)}, {unsafe.Pointer(pl), unsafe.Sizeof(*pl)}, {unsafe.Pointer(pu), unsafe.Sizeof(*pu)}, {unsafe.Pointer(receiver), unsafe.Sizeof(*receiver)}}
	var before [][]byte
	for _, record := range records {
		if unsafe.Sizeof(uintptr(0)) == 8 && uintptr(record.ptr) <= math.MaxUint32 {
			t.Fatal("all native layout and world-control records must exceed4GiB")
		}
		before = append(before, bytes.Clone(unsafe.Slice((*byte)(record.ptr), int(record.size))))
	}
	// The real default C point/spark encoders must traverse native receivers
	// and emit their actual packets before a legacy-int crash can earn credit.
	if monsterDeathOriginalPointControl549D80(unit) != 0 {
		t.Fatal("real C point-FX return")
	}
	pointPacket := []byte{0x81, 0x2c, 0x01, 0x2c, 0x01}
	if got := s.NetList.CopyPacketsA(ntype.PlayerInd(pl.PlayerInd), netlist.Kind1); !bytes.Equal(got, pointPacket) {
		t.Fatalf("real C native point-FX packet: %x", got)
	}
	if monsterDeathOriginalSparkControl549D80(unit) != 0 {
		t.Fatal("real C spark return")
	}
	if got := s.NetList.CopyPacketsA(ntype.PlayerInd(pl.PlayerInd), netlist.Kind1); !bytes.Equal(got, []byte{0x93, 0x2c, 0x01, 0x2c, 0x01, 0x80}) {
		t.Fatalf("real C native spark packet: %x", got)
	}
	monsterDeathOriginalPushControl549D80(unit)
	for i, record := range records {
		if !bytes.Equal(before[i], unsafe.Slice((*byte)(record.ptr), int(record.size))) {
			t.Fatal("whole valid native record changed during empty-map typed C controls", i)
		}
	}
	if unit.PosVec != types.Ptf(300, 300) {
		t.Fatal("independent source position")
	}
	return s, unit, pl, pointPacket
}

func TestMonsterDeathExplosionActualCBefore549D80(t *testing.T) {
	if os.Getenv("OPENNOX_MONSTER_DEATH_C_BEFORE") != "1" {
		return
	}
	if child := os.Getenv("OPENNOX_MONSTER_DEATH_C_CHILD"); child != "" {
		entry, err := strconv.ParseUint(child, 16, 32)
		if err != nil || (entry != 0x549d80 && entry != 0x549e00 && entry != 0x549e70 && entry != 0x54a250) {
			t.Fatal("unknown actual old C child", child)
		}
		TestMonsterDeathExplosionNativeLayout549D80(t)
		_, unit, _, _ := monsterDeathNativeBeforeFixture549D80(t)
		if unsafe.Sizeof(uintptr(0)) != 8 || monsterDeathOriginalABI549D80(uint32(entry)) != 32 || monsterDeathOriginalRuntimeOverride549D80 != nil {
			t.Fatal("unchanged old public C and native64 controls required")
		}
		t.Logf("MONSTER_DEATH_REAL_NATIVE_CONTROLS entry=%08x unit=%p native-layout-whole-records-real-C-point-spark-push=true", entry, unit)
		monsterDeathOriginalC549D80(uint32(entry), unit)
		t.Fatal("old legacy-int C entry unexpectedly returned")
	}
	for _, entry := range []uint32{0x549d80, 0x549e00, 0x549e70, 0x54a250} {
		t.Run(fmt.Sprintf("%08x", entry), func(t *testing.T) {
			cmd := exec.Command(os.Args[0], "-test.run=^TestMonsterDeathExplosionActualCBefore549D80$", "-test.v")
			cmd.Env = append(os.Environ(), "OPENNOX_MONSTER_DEATH_C_CHILD="+fmt.Sprintf("%08x", entry))
			out, err := cmd.CombinedOutput()
			t.Logf("actual old C child %08x: %v\n%s", entry, err, out)
			if err == nil || !bytes.Contains(out, []byte("MONSTER_DEATH_REAL_NATIVE_CONTROLS")) || bytes.Contains(out, []byte("unexpectedly returned")) {
				t.Fatal("actual crash requires complete native positive controls before entry", err)
			}
			if !bytes.Contains(out, []byte("SIGSEGV")) && !bytes.Contains(out, []byte("SIGBUS")) && !bytes.Contains(out, []byte("unexpected fault address")) {
				t.Fatal("child failure is not a reproduced native pointer fault")
			}
		})
	}
}
