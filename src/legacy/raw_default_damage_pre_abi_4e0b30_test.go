package legacy

import (
	"bytes"
	"context"
	"fmt"
	"math"
	"os"
	"os/exec"
	"strings"
	"testing"
	"time"
	"unsafe"

	"github.com/opennox/opennox/v1/server"
)

func rawParentControlGoLayout4E0B30() [16]uintptr {
	o := server.Object{}
	return [16]uintptr{
		unsafe.Offsetof(o.ObjClass), unsafe.Offsetof(o.ObjSubClass), unsafe.Offsetof(o.ObjFlags), unsafe.Offsetof(o.Buffs),
		unsafe.Offsetof(o.Obj130), unsafe.Offsetof(o.Field131), unsafe.Offsetof(o.Pos132), unsafe.Offsetof(o.Frame134),
		unsafe.Offsetof(o.ObjOwner), unsafe.Offsetof(o.Damage), unsafe.Offsetof(o.DamageSound), unsafe.Offsetof(o.UpdateData),
		unsafe.Offsetof(o.ScriptPickup) + unsafe.Sizeof(o.ScriptPickup), unsafe.Sizeof(uintptr(0)), unsafe.Sizeof(o.Damage), unsafe.Sizeof(o.UpdateData),
	}
}

func TestRawDefaultDamageABIInitialBefore4E0B30(t *testing.T) {
	if os.Getenv("OPENNOX_RAW_DEFAULT_DAMAGE_BEFORE") != "1" {
		return
	}
	if rawParentABI4E0B30() != 1 {
		t.Fatal("actual original five-int raw parent ABI required")
	}
	for _, name := range []string{"raw_default_damage_native_4e0b30.go", "raw_default_damage_export_4e0b30.go"} {
		if _, err := os.Stat(name); !os.IsNotExist(err) {
			t.Fatal("parent engine absent before reference/consumer freeze", name)
		}
	}
}

func rawParentABINativeControls4E0B30(t *testing.T) (target, source, weapon unsafe.Pointer) {
	t.Helper()
	if got, want := rawParentControlLayout4E0B30(), rawParentControlGoLayout4E0B30(); got != want {
		t.Fatal("actual C offsetof and Go named field/native pointer widths", got, want)
	}
	heap, guard, free := regenerationRawHeap4E01D0()
	t.Cleanup(free)
	if heap == nil || guard < rawParentControlGoLayout4E0B30()[12] || unsafe.Sizeof(uintptr(0)) != 8 || uintptr(heap) <= math.MaxUint32 {
		t.Fatal("guarded C-owned heap above4GiB required", heap, guard)
	}
	target, source, weapon = unsafe.Add(heap, 4096), unsafe.Add(heap, 8192), unsafe.Add(heap, 12288)
	for _, p := range []unsafe.Pointer{target, source, weapon} {
		if uintptr(p) <= math.MaxUint32 || uintptr(p)%unsafe.Alignof(server.Object{}) != 0 {
			t.Fatal("aligned high object", p)
		}
		clear(unsafe.Slice((*byte)(p), int(rawParentControlGoLayout4E0B30()[12])))
	}
	// The typed public child is a healthy real-C endpoint for the very first
	// raw-parent service. Keep its native field read separate from the actual
	// old parent's narrowed call. No custom damage engine or helper stub runs.
	buffs := (*uint32)(unsafe.Add(target, unsafe.Offsetof(server.Object{}.Buffs)))
	whole := append([]byte(nil), unsafe.Slice((*byte)(heap), 65536)...)
	if rawParentDirectBuff4E0B30(target, 23) != 0 {
		t.Fatal("real C buff query, clear")
	}
	if !bytes.Equal(whole, unsafe.Slice((*byte)(heap), 65536)) {
		t.Fatal("complete C heap unchanged by clear-buff real child query")
	}
	*buffs = 1 << 23
	set := append([]byte(nil), unsafe.Slice((*byte)(heap), 65536)...)
	if rawParentDirectBuff4E0B30(target, 23) != 1 {
		t.Fatal("real C buff query, set")
	}
	if !bytes.Equal(set, unsafe.Slice((*byte)(heap), 65536)) {
		t.Fatal("complete C heap unchanged by set-buff real child query")
	}
	*buffs = 0
	if !bytes.Equal(whole, unsafe.Slice((*byte)(heap), 65536)) {
		t.Fatal("only the explicit Buffs control word changed and was restored")
	}
	if got := rawParentPublic4E0B30(nil, source, weapon, math.MinInt32, -1); got != 1 {
		t.Fatal("actual C nil-target early return before any source/weapon read", got)
	}
	for _, native := range []bool{false, true} {
		got := rawParentEcho4E0B30(native, target, source, weapon, math.MinInt32, -1)
		want := [5]uintptr{uintptr(target), uintptr(source), uintptr(weapon), uintptr(uint32(0x80000000)), uintptr(uint32(0xffffffff))}
		if !native {
			for i := 0; i < 3; i++ {
				want[i] = uintptr(uint32(want[i]))
			}
		}
		if got != want {
			t.Fatal("five-word exact C ABI transport echo", native, got, want)
		}
		t.Logf("RAW_DEFAULT_DAMAGE_ABI_ECHO native=%t target=%#x source=%#x weapon=%#x damage=%#x type=%#x transport_only=true", native, got[0], got[1], got[2], got[3], got[4])
	}
	if !bytes.Equal(whole, unsafe.Slice((*byte)(heap), 65536)) {
		t.Fatal("complete heap unchanged through real child, nil target and echoes")
	}
	t.Logf("RAW_DEFAULT_DAMAGE_REAL_C_INITIAL_OK target=%p source=%p weapon=%p root=%#x layout_facts=16 direct_buff_clear_set=true nil_target=1 whole_heap_unchanged=true", target, source, weapon, rawParentEntry4E0B30())
	return
}

func TestRawDefaultDamageABIPositiveControls4E0B30(t *testing.T) {
	if unsafe.Sizeof(uintptr(0)) != 8 {
		return
	}
	rawParentABINativeControls4E0B30(t)
}

func TestRawDefaultDamageActualCBefore4E0B30(t *testing.T) {
	if os.Getenv("OPENNOX_RAW_DEFAULT_DAMAGE_BEFORE") != "1" || unsafe.Sizeof(uintptr(0)) != 8 {
		return
	}
	if child := os.Getenv("OPENNOX_RAW_DEFAULT_DAMAGE_OLD_CHILD"); child != "" {
		if child != "target-first-service" {
			t.Fatal("unknown original raw parent fault child", child)
		}
		TestRawDefaultDamageABIInitialBefore4E0B30(t)
		target, source, weapon := rawParentABINativeControls4E0B30(t)
		got := rawParentPublic4E0B30(target, source, weapon, 17, 0)
		t.Fatal("unchanged old raw parent unexpectedly returned", got)
	}
	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()
	cmd := exec.CommandContext(ctx, os.Args[0], "-test.run=^TestRawDefaultDamageActualCBefore4E0B30$", "-test.v")
	cmd.Env = append(os.Environ(), "OPENNOX_RAW_DEFAULT_DAMAGE_OLD_CHILD=target-first-service")
	output, err := cmd.CombinedOutput()
	t.Logf("actual unchanged raw parent first-service child: %v\n%s", err, output)
	if ctx.Err() != nil || err == nil || !bytes.Contains(output, []byte("RAW_DEFAULT_DAMAGE_REAL_C_INITIAL_OK")) || bytes.Contains(output, []byte("unexpectedly returned")) {
		t.Fatal("independent real-C/native high-object controls must precede unforced actual old-C fault", err)
	}
	if !strings.Contains(string(output), "SIGSEGV") && !strings.Contains(string(output), "SIGBUS") && !strings.Contains(string(output), "unexpected fault address") {
		t.Fatal("actual old-C memory fault required", fmt.Sprint(err))
	}
	// The output retains the physical fault stack. Its classification must
	// identify the earliest callee/frame actually reached; this test earns no
	// additional source-only or weapon-only whole-parent crash credit.
}
