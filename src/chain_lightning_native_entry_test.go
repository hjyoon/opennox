package opennox

import (
	"bytes"
	"os"
	"os/exec"
	"testing"
	"unsafe"

	"github.com/opennox/libs/strman"
	"github.com/opennox/libs/types"
	"github.com/opennox/opennox/v1/legacy"
	"github.com/opennox/opennox/v1/legacy/common/alloc"
	"github.com/opennox/opennox/v1/legacy/common/ccall"
	"github.com/opennox/opennox/v1/server"
)

// Each entry runs in its own process: a leftover PE32 entry can corrupt the
// native duration record or fault before Go can report an ordinary test error.
func TestChainLightningPublicEntryNativeRecord(t *testing.T) {
	const childEnv = "NOX_CHAIN_LIGHTNING_ENTRY_CHILD"
	mode := os.Getenv(childEnv)
	if mode == "" {
		binary, err := os.Executable()
		if err != nil {
			t.Fatal(err)
		}
		for _, mode := range []string{"create", "update", "destroy"} {
			t.Run(mode, func(t *testing.T) {
				cmd := exec.Command(binary, "-test.run=^TestChainLightningPublicEntryNativeRecord$", "-test.v")
				cmd.Env = append(os.Environ(), childEnv+"="+mode)
				out, err := cmd.CombinedOutput()
				if err != nil {
					t.Fatalf("native public %s entry: %v\n%s", mode, err, out)
				}
			})
		}
		return
	}
	s := NewServer(nil, nil, strman.New())
	t.Cleanup(s.Close)
	s.spells.duration.Init(s)
	oldServer, oldGet := noxServer, legacy.GetServer
	noxServer, legacy.GetServer = s, func() legacy.Server { return s }
	t.Cleanup(func() { noxServer, legacy.GetServer = oldServer, oldGet })
	type nativeFrame struct {
		Before [32]byte
		Record server.DurSpell
		After  [32]byte
	}
	frame, free := alloc.New(nativeFrame{})
	t.Cleanup(free)
	*frame = nativeFrame{}
	for i := range frame.Before {
		frame.Before[i], frame.After[i] = byte(0x51+i), byte(0xe7-i)
	}
	frame.Record = server.DurSpell{
		ID: 0xabcd, Spell: 43, Level: 3, Flag20: 1,
		Pos: types.Ptf(11.5, -22.25), Field36: 0x13579bdf,
		Frame68: 0xf1234567, Field76: 0x3fe00000, Flags88: 0x77556601,
	}
	if unsafe.Sizeof(uintptr(0)) == 8 && uintptr(frame.Record.C()) <= 0xffffffff {
		t.Fatal("C-owned duration record must be above4GiB")
	}
	t.Logf("native %s duration=%p size=%d", mode, &frame.Record, unsafe.Sizeof(frame.Record))
	expected := *frame
	switch mode {
	case "create":
		frame.Record.Field72, expected.Record.Field72 = 0x12345678, 0
		expected.Record.Field76 = 0
		if got := ccall.CallIntPtr(legacy.Get_nox_xxx_onStartLightning_52F820(), frame.Record.C()); got != 0 {
			t.Fatalf("create return = %d, want0", got)
		}
	case "update":
		frame.Record.Flag20, expected.Record.Flag20 = 0, 0
		if got := ccall.CallIntPtr(legacy.Get_nox_xxx_onFrameLightning_52F8A0(), frame.Record.C()); got != 1 {
			t.Fatalf("update without caster return = %d, want1", got)
		}
	case "destroy":
		ccall.CallVoidPtr(legacy.Get_sub_530100(), frame.Record.C())
	default:
		t.Fatal("unknown child mode", mode)
	}
	gotBytes := unsafe.Slice((*byte)(unsafe.Pointer(frame)), int(unsafe.Sizeof(*frame)))
	wantBytes := unsafe.Slice((*byte)(unsafe.Pointer(&expected)), int(unsafe.Sizeof(expected)))
	if !bytes.Equal(gotBytes, wantBytes) {
		t.Fatal("public entry changed unexpected duration or adjacent canary bytes")
	}
}
