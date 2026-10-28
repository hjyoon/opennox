package legacy

import (
	"bytes"
	"os"
	"os/exec"
	"testing"
	"unsafe"

	"github.com/opennox/libs/object"
	"github.com/opennox/libs/spell"
	"github.com/opennox/libs/types"
	"github.com/opennox/opennox/v1/legacy/common/alloc"
	"github.com/opennox/opennox/v1/legacy/common/ccall"
	"github.com/opennox/opennox/v1/server"
)

type tagNativeEntryServer530160 struct {
	Server
	base *server.Server
}

func (s *tagNativeEntryServer530160) S() *server.Server { return s.base }

// Actual public C calls run in subprocesses so an old PE32 entry cannot
// corrupt the surrounding test process before its native bytes are checked.
func TestSpellTagPublicEntryNativeRecord530160(t *testing.T) {
	const childEnv = "NOX_TAG_NATIVE_ENTRY_CHILD"
	mode := os.Getenv(childEnv)
	if mode == "" {
		binary, err := os.Executable()
		if err != nil {
			t.Fatal(err)
		}
		for _, mode := range []string{"create-nonplayer", "update-living", "update-dead", "update-nil", "destroy-nonplayer"} {
			t.Run(mode, func(t *testing.T) {
				cmd := exec.Command(binary, "-test.run=^TestSpellTagPublicEntryNativeRecord530160$", "-test.v")
				cmd.Env = append(os.Environ(), childEnv+"="+mode)
				out, err := cmd.CombinedOutput()
				if err != nil {
					t.Fatalf("native Tag %s entry: %v\n%s", mode, err, out)
				}
			})
		}
		return
	}
	fake := &tagNativeEntryServer530160{base: &server.Server{}}
	old := GetServer
	GetServer = func() Server { return fake }
	t.Cleanup(func() { GetServer = old })
	type frame struct {
		Before [32]byte
		Record server.DurSpell
		After  [32]byte
	}
	f, free := alloc.New(frame{})
	t.Cleanup(free)
	caster, freeCaster := alloc.New(server.Object{})
	t.Cleanup(freeCaster)
	target, freeTarget := alloc.New(server.Object{})
	t.Cleanup(freeTarget)
	*f = frame{}
	*caster = server.Object{ObjClass: object.ClassMonster}
	*target = server.Object{ObjClass: object.ClassMonster}
	for i := range f.Before {
		f.Before[i], f.After[i] = byte(0x51+i), byte(0xe7-i)
	}
	f.Record = server.DurSpell{
		ID: 0xabcd, Spell: uint32(spell.SPELL_TAG), Level: 3, Obj12: target, Caster16: caster,
		Target48: target, Pos: types.Ptf(11.5, -22.25), Field36: 0x13579bdf,
		Frame68: 0xf1234567, Field76: 0x3fe00000, Flags88: 0x77556601,
	}
	if unsafe.Sizeof(uintptr(0)) == 8 {
		for _, p := range []unsafe.Pointer{f.Record.C(), caster.CObj(), target.CObj()} {
			if uintptr(p) <= 0xffffffff {
				t.Fatal("C-owned native Tag frame must be above4GiB")
			}
		}
	}
	if mode == "update-dead" {
		target.ObjFlags = object.Flags(0x20)
	} else if mode == "update-nil" {
		f.Record.Target48 = nil
	}
	before := append([]byte(nil), unsafe.Slice((*byte)(unsafe.Pointer(f)), int(unsafe.Sizeof(*f)))...)
	casterBefore, targetBefore := *caster, *target
	t.Logf("native Tag %s record=%p caster=%p target=%p size=%d", mode, &f.Record, caster, target, unsafe.Sizeof(f.Record))
	switch mode {
	case "create-nonplayer":
		if got := ccall.CallIntPtr(Get_nox_xxx_spellTagCreature_530160(), f.Record.C()); got != 1 {
			t.Fatalf("nonplayer create = %d, want1", got)
		}
	case "update-living", "update-dead", "update-nil":
		want := 1
		if mode == "update-living" {
			want = 0
		}
		if got := ccall.CallIntPtr(Get_sub_530250(), f.Record.C()); got != want {
			t.Fatalf("update = %d, want%d", got, want)
		}
	case "destroy-nonplayer":
		// The duration engine ignores the original destroy return register.
		ccall.CallVoidPtr(Get_sub_530270(), f.Record.C())
	default:
		t.Fatal("unknown child mode", mode)
	}
	if !bytes.Equal(before, unsafe.Slice((*byte)(unsafe.Pointer(f)), int(unsafe.Sizeof(*f)))) || *caster != casterBefore || *target != targetBefore {
		t.Fatal("Tag entry changed native input or adjacent canary bytes")
	}
}
