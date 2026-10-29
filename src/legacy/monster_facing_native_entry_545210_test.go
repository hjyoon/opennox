package legacy

import (
	"bytes"
	"os"
	"os/exec"
	"testing"
	"unsafe"

	"github.com/opennox/libs/object"
	"github.com/opennox/libs/strman"
	"github.com/opennox/libs/types"
	"github.com/opennox/opennox/v1/common/unit/ai"
	"github.com/opennox/opennox/v1/legacy/common/alloc"
	"github.com/opennox/opennox/v1/legacy/common/alloc/handles"
	"github.com/opennox/opennox/v1/legacy/common/ccall"
	"github.com/opennox/opennox/v1/server"
)

type facingProbeLegacyServer545210 struct {
	Server
	srv *server.Server
}

func (s *facingProbeLegacyServer545210) S() *server.Server { return s.srv }

// Read-only baseline: registered ordinary native dispatch completes before
// the same C-owned unit/update/stack reaches its retained public C pointer.
// The AI dispatcher discards the original char return register.
func TestMonsterFacingNextPublicCEntriesNative545210(t *testing.T) {
	const child = "NOX_MONSTER_FACING_NEXT_ENTRY_CHILD"
	mode := os.Getenv(child)
	if mode == "" {
		binary, err := os.Executable()
		if err != nil {
			t.Fatal(err)
		}
		for _, mode := range []string{"location", "object", "angle", "set-angle", "pop"} {
			t.Run(mode, func(t *testing.T) {
				cmd := exec.Command(binary, "-test.run=^TestMonsterFacingNextPublicCEntriesNative545210$", "-test.v")
				cmd.Env = append(os.Environ(), child+"="+mode)
				out, err := cmd.CombinedOutput()
				t.Logf("%s", out)
				if err != nil {
					t.Fatalf("actual facing %s C entry: %v", mode, err)
				}
			})
		}
		return
	}
	handles.Init()
	t.Cleanup(handles.Release)
	s := server.New(nil, nil, strman.New())
	t.Cleanup(s.Close)
	s.SetFrame(150)
	oldGet := GetServer
	GetServer = func() Server { return &facingProbeLegacyServer545210{srv: s} }
	t.Cleanup(func() { GetServer = oldGet })
	if !s.Objs.Init(4) {
		t.Fatal("native object pool")
	}
	t.Cleanup(s.Objs.FreeObjects)
	unit, target := s.Objs.NewObject(&server.ObjectType{}), s.Objs.NewObject(&server.ObjectType{})
	unit.ObjClass, unit.ObjFlags = object.ClassMonster, 0
	unit.PosVec, unit.Direction1, unit.Direction2 = types.Ptf(10, 20), 0, 0
	target.PosVec = types.Ptf(10, 120)
	type guarded struct {
		Before [32]byte
		Update server.MonsterUpdateData
		After  [32]byte
	}
	f, free := alloc.New(guarded{})
	t.Cleanup(free)
	*f = guarded{}
	for i := range f.Before {
		f.Before[i], f.After[i] = byte(0x51+i), byte(0xe7-i)
	}
	unit.UpdateData = unsafe.Pointer(&f.Update)
	typ := ai.ACTION_FACE_LOCATION
	switch mode {
	case "object":
		typ = ai.ACTION_FACE_OBJECT
	case "angle", "pop":
		typ = ai.ACTION_FACE_ANGLE
	case "set-angle":
		typ = ai.ACTION_SET_ANGLE
	}
	f.Update.AIStackInd = 1
	f.Update.AIStack[0].Action = uint32(ai.ACTION_IDLE)
	f.Update.AIStack[1].Action = uint32(typ)
	switch mode {
	case "location":
		f.Update.AIStack[1].SetArgs(types.Ptf(10, 120))
	case "object":
		f.Update.AIStack[1].SetArgs(target)
	case "angle", "pop":
		unit.Direction2 = 252
		f.Update.AIStack[1].SetArgs(uint32(64))
	case "set-angle":
		f.Update.AIStack[1].SetArgs(int32(-1))
	}
	if unsafe.Sizeof(uintptr(0)) == 8 && (uintptr(unit.CObj()) <= 0xffffffff || uintptr(target.CObj()) <= 0xffffffff || uintptr(unit.UpdateData) <= 0xffffffff) {
		t.Fatal("native C-owned links must exceed4GiB")
	}
	a, ok := server.GetAIAction(typ).(cgoAIAction)
	if !ok || a.update == nil || a.cancel == nil {
		t.Fatal("retained public C action pointers")
	}
	beforeUnit, beforeFrame := *unit, *f
	if mode == "pop" {
		a.Cancel(unit)
	} else {
		a.Update(unit)
	}
	expectedUnit, expectedFrame := *unit, *f
	wantDirection, wantStack := server.Dir16(8), int8(1)
	if mode == "angle" {
		wantDirection = 4
	}
	if mode == "set-angle" {
		wantDirection, wantStack = 255, 0
	}
	if mode == "pop" {
		wantDirection, wantStack = 252, 0
	}
	if unit.Direction2 != wantDirection || f.Update.AIStackInd != wantStack {
		t.Fatalf("ordinary control direction=%d stack=%d", unit.Direction2, f.Update.AIStackInd)
	}
	t.Logf("ordinary native facing mode=%s unit=%p target=%p update=%p direction=%d stack=%d completed", mode, unit, target, unit.UpdateData, unit.Direction2, f.Update.AIStackInd)
	*unit, *f = beforeUnit, beforeFrame
	if mode == "pop" {
		ccall.CallVoidPtr(a.cancel, unit.CObj())
	} else {
		ccall.CallVoidPtr(a.update, unit.CObj())
	}
	rawUnit := func(p *server.Object) []byte { return unsafe.Slice((*byte)(unsafe.Pointer(p)), int(unsafe.Sizeof(*p))) }
	rawFrame := func(p *guarded) []byte { return unsafe.Slice((*byte)(unsafe.Pointer(p)), int(unsafe.Sizeof(*p))) }
	if !bytes.Equal(rawUnit(unit), rawUnit(&expectedUnit)) || !bytes.Equal(rawFrame(f), rawFrame(&expectedFrame)) {
		t.Fatal("actual public C differs from ordinary native object/update/stack/canary")
	}
}
