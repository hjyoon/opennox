package legacy

import (
	"math"
	"os"
	"os/exec"
	"runtime"
	"testing"
	"unsafe"

	"github.com/opennox/libs/object"

	"github.com/opennox/opennox/v1/legacy/common/alloc"
	"github.com/opennox/opennox/v1/server"
)

const summonManaCostOracleChild500CA0 = "OPENNOX_TEST_SUMMON_MANA_COST_500CA0"

var summonManaCostLegacyOracleTable500CA0 = [...]int32{
	15, 60, 60, 60, 30, 30, 30, 15, 60, 60,
	15, 30, 30, 30, 85, 85, 30, 60, 85, 60,
	30, 30, 60, 30, 15, 30, 85, 30, 30, 15,
	60, 30, 30, 30, 30, 60, 60, 60, 60, 60,
}

func TestSummonManaCostExport500CA0PreservesNativePointerAndInt32(t *testing.T) {
	unit := new(server.Object)
	if unsafe.Sizeof(uintptr(0)) == 8 && uintptr(unsafe.Pointer(unit)) <= math.MaxUint32 {
		t.Fatalf("unit pointer = %p, want native address above 4 GiB", unit)
	}

	oldCall := summonManaCostCall500CA0
	t.Cleanup(func() { summonManaCostCall500CA0 = oldCall })
	var gotID int32
	var gotUnit *server.Object
	summonManaCostCall500CA0 = func(spellID int32, unit *server.Object) int32 {
		gotID = spellID
		gotUnit = unit
		return math.MinInt32 + 0x500ca0
	}

	var pin runtime.Pinner
	pin.Pin(unit)
	defer pin.Unpin()
	if got := summonManaCostExportCall500CA0(math.MinInt32, unit); got != math.MinInt32+0x500ca0 {
		t.Fatalf("result = %d, want %d", got, math.MinInt32+0x500ca0)
	}
	if gotID != math.MinInt32 || gotUnit != unit {
		t.Fatalf("call = (%d, %p), want (%d, %p)", gotID, gotUnit, math.MinInt32, unit)
	}
	if got := summonManaCostExportCall500CA0(math.MaxInt32, nil); got != math.MinInt32+0x500ca0 {
		t.Fatalf("null-unit result = %d, want %d", got, math.MinInt32+0x500ca0)
	}
	if gotID != math.MaxInt32 || gotUnit != nil {
		t.Fatalf("null-unit call = (%d, %p), want (%d, nil)", gotID, gotUnit, math.MaxInt32)
	}
	runtime.KeepAlive(unit)
}

func TestSummonManaCostExport500CA0MatchesOracleBlob(t *testing.T) {
	if os.Getenv(summonManaCostOracleChild500CA0) != "1" {
		cmd := exec.Command(os.Args[0], "-test.run=^TestSummonManaCostExport500CA0MatchesOracleBlob$", "-test.count=1")
		cmd.Env = append(os.Environ(), summonManaCostOracleChild500CA0+"=1")
		if out, err := cmd.CombinedOutput(); err != nil {
			t.Fatalf("isolated summon-cost oracle test failed: %v\n%s", err, out)
		}
		return
	}

	InitBlobData()
	unit, freeUnit := alloc.New(server.Object{})
	defer freeUnit()
	unit.ObjClass = object.ClassPlayer
	for i, want := range summonManaCostLegacyOracleTable500CA0 {
		spellID := int32(75 + i)
		if got := summonManaCostExportCall500CA0(spellID, unit); got != want {
			t.Fatalf("spell %d cost = %d, want oracle %d", spellID, got, want)
		}
	}
	unit.ObjClass = object.ClassMonster
	if got := summonManaCostExportCall500CA0(75, unit); got != 0 {
		t.Fatalf("Monster cost = %d, want 0", got)
	}
	if got := summonManaCostExportCall500CA0(75, nil); got != 0 {
		t.Fatalf("null-unit cost = %d, want 0", got)
	}
}
