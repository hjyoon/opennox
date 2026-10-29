package opennox

import (
	"bytes"
	"github.com/opennox/libs/object"
	"github.com/opennox/libs/spell"
	"github.com/opennox/libs/strman"
	"github.com/opennox/libs/types"
	noxflags "github.com/opennox/opennox/v1/common/flags"
	"github.com/opennox/opennox/v1/legacy"
	"github.com/opennox/opennox/v1/legacy/common/alloc"
	"github.com/opennox/opennox/v1/legacy/common/ccall"
	"github.com/opennox/opennox/v1/server"
	"os"
	"os/exec"
	"testing"
	"unsafe"
)

// Direct C callbacks and ordinary duration dispatch must agree for guarded
// native records. This read-only next-entry probe records the actual result
// without changing any engine source or campaign inputs.
func TestTeleportPopNextPublicEntryNativeRecordProbe530820(t *testing.T) {
	const childEnv = "NOX_TELEPORT_POP_NEXT_ENTRY_PROBE_CHILD"
	mode := os.Getenv(childEnv)
	if mode == "" {
		binary, err := os.Executable()
		if err != nil {
			t.Fatal(err)
		}
		for _, mode := range []string{"create", "update"} {
			t.Run(mode, func(t *testing.T) {
				cmd := exec.Command(binary, "-test.run=^TestTeleportPopNextPublicEntryNativeRecordProbe530820$", "-test.v")
				cmd.Env = append(os.Environ(), childEnv+"="+mode)
				out, err := cmd.CombinedOutput()
				if err != nil {
					t.Fatalf("actual TeleportPop %s C entry: %v\n%s", mode, err, out)
				}
			})
		}
		return
	}
	s := NewServer(nil, nil, strman.New())
	t.Cleanup(s.Close)
	s.spells.duration.Init(s)
	s.SetFrame(0xf0)
	oldServer, oldGet := noxServer, legacy.GetServer
	noxServer, legacy.GetServer = s, func() legacy.Server { return s }
	t.Cleanup(func() { noxServer, legacy.GetServer = oldServer, oldGet })
	oldFlags := noxflags.GetGame()
	noxflags.UnsetGame(oldFlags)
	t.Cleanup(func() { noxflags.UnsetGame(noxflags.GetGame()); noxflags.SetGame(oldFlags) })
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
	*target = server.Object{ObjClass: object.ClassMonster, ObjFlags: object.Flags(0x20)}
	for i := range f.Before {
		f.Before[i], f.After[i] = byte(0x51+i), byte(0xe7-i)
	}
	f.Record = server.DurSpell{ID: 0xabcd, Spell: uint32(spell.SPELL_TELEPORT_POP), Level: 3, Obj12: target, Caster16: target, Target48: target, Pos: types.Ptf(11.5, -22.25), Frame68: 0x87654321, Field36: 0x13579bdf, Flags88: 0x77556601}
	if unsafe.Sizeof(uintptr(0)) == 8 && (uintptr(f.Record.C()) <= 0xffffffff || uintptr(target.CObj()) <= 0xffffffff) {
		t.Fatal("native C-owned TeleportPop frames must be above4GiB")
	}
	expected := *f
	var callback unsafe.Pointer
	var want int32
	switch mode {
	case "create":
		callback = legacy.Get_nox_xxx_castTele_530820()
		want = s.spells.duration.callCreate4FEBA0(callback, &expected.Record)
	case "update":
		callback = legacy.Get_sub_530880()
		want = s.spells.duration.callUpdate4FEEF0(callback, &expected.Record)
	default:
		t.Fatal("unknown probe mode")
	}
	t.Logf("native TeleportPop %s record=%p target=%p expected-return=%d expected-timer=%#x", mode, &f.Record, target, want, expected.Record.Frame68)
	got := ccall.CallIntPtr(callback, f.Record.C())
	if got != int(want) {
		t.Fatalf("public C return=%d ordinary native Go dispatch=%d", got, want)
	}
	raw := func(v *frame) []byte { return unsafe.Slice((*byte)(unsafe.Pointer(v)), int(unsafe.Sizeof(*v))) }
	if !bytes.Equal(raw(f), raw(&expected)) {
		t.Fatalf("public C record/canary differs from ordinary native Go dispatch: timer=%#x want=%#x", f.Record.Frame68, expected.Record.Frame68)
	}
}
