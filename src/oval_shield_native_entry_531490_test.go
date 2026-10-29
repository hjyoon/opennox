package opennox

import (
	"bytes"
	"math"
	"os"
	"os/exec"
	"testing"
	"unsafe"

	"github.com/opennox/libs/object"
	"github.com/opennox/libs/types"
	"github.com/opennox/opennox/v1/legacy"
	"github.com/opennox/opennox/v1/legacy/common/alloc"
	"github.com/opennox/opennox/v1/legacy/common/ccall"
	"github.com/opennox/opennox/v1/server"
)

// Regression for the three originally faulting public C entries, compared
// with ordinary dispatch on the same C-owned native record and target.
func TestOvalShieldNextPublicEntryNativeRecordProbe531490(t *testing.T) {
	const childEnv = "NOX_OVAL_SHIELD_NEXT_ENTRY_PROBE_CHILD"
	mode := os.Getenv(childEnv)
	if mode == "" {
		binary, err := os.Executable()
		if err != nil {
			t.Fatal(err)
		}
		for _, mode := range []string{"create-nonplayer", "update-nil-target", "destroy-nil-target"} {
			t.Run(mode, func(t *testing.T) {
				cmd := exec.Command(binary, "-test.run=^TestOvalShieldNextPublicEntryNativeRecordProbe531490$", "-test.v")
				cmd.Env = append(os.Environ(), childEnv+"="+mode)
				out, err := cmd.CombinedOutput()
				if err != nil {
					t.Fatalf("actual OvalShield %s C entry: %v\n%s", mode, err, out)
				}
			})
		}
		return
	}
	sp := &spellsDuration{s: &Server{Server: &server.Server{}}}
	oldServer, oldGet := noxServer, legacy.GetServer
	noxServer, legacy.GetServer = sp.s, func() legacy.Server { return sp.s }
	t.Cleanup(func() { noxServer, legacy.GetServer = oldServer, oldGet })
	type frame struct {
		Before [32]byte
		Record server.DurSpell
		After  [32]byte
	}
	f, free := alloc.New(frame{})
	t.Cleanup(free)
	*f = frame{}
	target, freeTarget := alloc.New(server.Object{})
	t.Cleanup(freeTarget)
	*target = server.Object{ObjClass: object.ClassMonster}
	for i := range f.Before {
		f.Before[i], f.After[i] = byte(0x51+i), byte(0xe7-i)
	}
	f.Record = server.DurSpell{ID: 0xabcd, Spell: 67, Level: 3, Pos: types.Ptf(math.Float32frombits(0x3fdccccc), 200), Frame68: 0x87654321, Flags88: 0x77556601}
	if mode == "create-nonplayer" {
		f.Record.Target48 = target
	}
	if unsafe.Sizeof(uintptr(0)) == 8 && (uintptr(f.Record.C()) <= 0xffffffff || uintptr(target.CObj()) <= 0xffffffff) {
		t.Fatal("native C-owned OvalShield links must exceed 4 GiB")
	}
	expected := *f
	raw := func(v *frame) []byte { return unsafe.Slice((*byte)(unsafe.Pointer(v)), int(unsafe.Sizeof(*v))) }
	switch mode {
	case "create-nonplayer":
		callback := legacy.Get_sub_531490()
		want := sp.callCreate4FEBA0(callback, &expected.Record)
		if want != 1 {
			t.Fatalf("ordinary native create=%d", want)
		}
		t.Logf("native OvalShield create record=%p target=%p class=%d expected-return=%d", &f.Record, target, target.ObjClass, want)
		if got := ccall.CallIntPtr(callback, f.Record.C()); got != int(want) {
			t.Fatalf("C=%d ordinary=%d", got, want)
		}
	case "update-nil-target":
		callback := legacy.Get_sub_5314F0()
		want := sp.callUpdate4FEEF0(callback, &expected.Record)
		if want != 1 {
			t.Fatalf("ordinary native update=%d", want)
		}
		t.Logf("native OvalShield update record=%p target=nil expected-return=%d", &f.Record, want)
		if got := ccall.CallIntPtr(callback, f.Record.C()); got != int(want) {
			t.Fatalf("C=%d ordinary=%d", got, want)
		}
	case "destroy-nil-target":
		callback := legacy.Get_sub_531560()
		sp.callDestroy4FEDA0(callback, &expected.Record)
		t.Logf("native OvalShield destroy record=%p target=nil ordinary native dispatch completed", &f.Record)
		ccall.CallVoidPtr(callback, f.Record.C())
	default:
		t.Fatal("unknown mode")
	}
	if !bytes.Equal(raw(f), raw(&expected)) {
		t.Fatal("public C record/canary differs from ordinary native dispatch")
	}
}
