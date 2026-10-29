package opennox

import (
	"bytes"
	"os"
	"os/exec"
	"testing"
	"unsafe"

	"github.com/opennox/libs/object"
	"github.com/opennox/libs/strman"
	"github.com/opennox/libs/types"
	noxflags "github.com/opennox/opennox/v1/common/flags"
	"github.com/opennox/opennox/v1/legacy"
	"github.com/opennox/opennox/v1/legacy/common/alloc"
	"github.com/opennox/opennox/v1/legacy/common/alloc/handles"
	"github.com/opennox/opennox/v1/legacy/common/ccall"
	"github.com/opennox/opennox/v1/server"
)

// Read-only external overlay; runtime source stays frozen at the CI revision.
func TestFirewalkNextPublicEntryNativeRecordProbe52ED40(t *testing.T) {
	const childEnv = "NOX_FIREWALK_NEXT_ENTRY_PROBE_CHILD"
	mode := os.Getenv(childEnv)
	if mode == "" {
		binary, err := os.Executable()
		if err != nil {
			t.Fatal(err)
		}
		for _, mode := range []string{"update-nil-target"} {
			t.Run(mode, func(t *testing.T) {
				cmd := exec.Command(binary, "-test.run=^TestFirewalkNextPublicEntryNativeRecordProbe52ED40$", "-test.v")
				cmd.Env = append(os.Environ(), childEnv+"="+mode)
				out, err := cmd.CombinedOutput()
				if err != nil {
					t.Fatalf("actual Firewalk %s C entry: %v\n%s", mode, err, out)
				}
			})
		}
		return
	}
	handles.Init()
	t.Cleanup(handles.Release)
	s := NewServer(nil, nil, strman.New())
	t.Cleanup(s.Close)
	oldServer, oldGet := noxServer, legacy.GetServer
	noxServer, legacy.GetServer = s, func() legacy.Server { return s }
	t.Cleanup(func() { noxServer, legacy.GetServer = oldServer, oldGet })
	oldFlags := noxflags.GetGame()
	noxflags.UnsetGame(oldFlags)
	t.Cleanup(func() { noxflags.UnsetGame(noxflags.GetGame()); noxflags.SetGame(oldFlags) })
	s.spells.duration.Init(s)
	s.SetFrame(142)
	s.SetTickRate(30)
	type frame struct {
		Before [32]byte
		Record server.DurSpell
		After  [32]byte
	}
	f, free := alloc.New(frame{})
	t.Cleanup(free)
	*f = frame{}
	source, freeSource := alloc.New(server.Object{})
	t.Cleanup(freeSource)
	*source = server.Object{ObjClass: object.ClassMonster, PosVec: types.Ptf(11.5, -22.25)}
	for i := range f.Before {
		f.Before[i], f.After[i] = byte(0x51+i), byte(0xe7-i)
	}
	f.Record = server.DurSpell{ID: 0xabcd, Spell: 28, Level: 3, Obj12: source, Obj24: source, Target48: nil, Pos: types.Ptf(100, 200), Frame68: 0x87654321, Field76: 0x55aa, Flags88: 0x77556601}
	if unsafe.Sizeof(uintptr(0)) == 8 && (uintptr(f.Record.C()) <= 0xffffffff || uintptr(source.CObj()) <= 0xffffffff) {
		t.Fatal("native C-owned Firewalk links must exceed 4 GiB")
	}
	expected := *f
	raw := func(p *frame) []byte { return unsafe.Slice((*byte)(unsafe.Pointer(p)), int(unsafe.Sizeof(*p))) }
	switch mode {
	case "update-nil-target":
		callback := legacy.Get_nox_xxx_firewalkTick_52ED40()
		want := s.spells.duration.callUpdate4FEEF0(callback, &expected.Record)
		if want != 1 {
			t.Fatal("ordinary native Firewalk update control")
		}
		t.Logf("native Firewalk update record=%p target=nil expected-return=%d", &f.Record, want)
		if got := ccall.CallIntPtr(callback, f.Record.C()); got != int(want) {
			t.Fatalf("C=%d ordinary=%d", got, want)
		}
	default:
		t.Fatal("unknown mode")
	}
	if !bytes.Equal(raw(f), raw(&expected)) {
		t.Fatal("public C record/canary differs from ordinary native dispatch")
	}
}
