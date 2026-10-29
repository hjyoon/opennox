package legacy

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"testing"
	"unsafe"

	"github.com/opennox/libs/object"
	"github.com/opennox/libs/strman"
	"github.com/opennox/opennox/v1/common/unit/ai"
	"github.com/opennox/opennox/v1/legacy/common/alloc"
	"github.com/opennox/opennox/v1/legacy/common/alloc/handles"
	"github.com/opennox/opennox/v1/legacy/common/ccall"
	"github.com/opennox/opennox/v1/server"
)

type terminalCompositionLegacyServer5320E0 struct {
	Server
	srv *server.Server
}

func (s *terminalCompositionLegacyServer5320E0) S() *server.Server { return s.srv }

func TestMonsterTerminalPublicCNativeComposition5320E0(t *testing.T) {
	const child = "NOX_MONSTER_TERMINAL_NATIVE_COMPOSE_CHILD"
	mode := os.Getenv(child)
	if mode == "" {
		binary, err := os.Executable()
		if err != nil {
			t.Fatal(err)
		}
		for _, mode := range []string{"shield-finish", "weapon-block", "dying"} {
			t.Run(mode, func(t *testing.T) {
				cmd := exec.Command(binary, "-test.run=^TestMonsterTerminalPublicCNativeComposition5320E0$", "-test.v")
				cmd.Env = append(os.Environ(), child+"="+mode)
				out, err := cmd.CombinedOutput()
				t.Logf("%s", out)
				if err != nil {
					t.Fatalf("actual terminal %s public C composition: %v", mode, err)
				}
			})
		}
		return
	}
	path := "../server/testdata/monster_terminal_native_original_5320e0.json"
	if override := os.Getenv("NOX_MONSTER_TERMINAL_ORIGINAL_FIXTURE"); override != "" {
		path = override
	}
	blob, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	var fixture struct {
		OriginalSHA string `json:"original_sha256"`
		ControlWord uint16 `json:"control_word"`
		Rows        []struct {
			Input struct {
				Function  string `json:"function"`
				Ready     uint8  `json:"ready"`
				UnitLow   uint8  `json:"unit_low_byte"`
				PopReturn uint32 `json:"pop_return"`
				Mutation  string `json:"mutation"`
			} `json:"input"`
			Result struct {
				ReturnByte uint8 `json:"return_byte"`
			} `json:"result"`
		} `json:"rows"`
	}
	if err := json.Unmarshal(blob, &fixture); err != nil {
		t.Fatal(err)
	}
	if fixture.OriginalSHA != "0040e2c0683b4d73a5fb976e400d5087dca680df2b195c9e27f8edbda2d4974a" || fixture.ControlWord != 0xe3f || len(fixture.Rows) != 1350 {
		t.Fatal("frozen original terminal fixture identity/profile/count")
	}
	key := func(fn string, ready, low uint8, ret uint32) string {
		return fmt.Sprintf("%s/%d/%d/%d", fn, ready, low, ret)
	}
	wantBytes := make(map[string]uint8)
	for _, row := range fixture.Rows {
		if row.Input.Mutation == "none" {
			k := key(row.Input.Function, row.Input.Ready, row.Input.UnitLow, row.Input.PopReturn)
			if _, exists := wantBytes[k]; exists {
				t.Fatal("duplicate frozen original context")
			}
			wantBytes[k] = row.Result.ReturnByte
		}
	}
	if len(wantBytes) != 450 {
		t.Fatal("incomplete original return contexts")
	}
	handles.Init()
	t.Cleanup(handles.Release)
	s := server.New(nil, nil, strman.New())
	t.Cleanup(s.Close)
	s.SetFrame(150)
	oldGet := GetServer
	GetServer = func() Server { return &terminalCompositionLegacyServer5320E0{srv: s} }
	t.Cleanup(func() { GetServer = oldGet })
	if !s.Objs.Init(64) {
		t.Fatal("native object pool")
	}
	t.Cleanup(s.Objs.FreeObjects)
	units := make(map[uint8]*server.Object)
	for i := 0; i < 64 && len(units) < 3; i++ {
		unit := s.Objs.NewObject(&server.ObjectType{})
		if unit == nil {
			t.Fatal("native unit allocation")
		}
		low := uint8(uintptr(unit.CObj()))
		if low == 0 || low == 16 || low == 128 {
			units[low] = unit
		}
	}
	if len(units) != 3 {
		t.Fatal("native units for frozen original low-byte identities")
	}
	type guarded struct {
		Before [32]byte
		Update server.MonsterUpdateData
		After  [32]byte
	}
	f, free := alloc.New(guarded{})
	t.Cleanup(free)
	typ := ai.ACTION_BLOCK_FINISH
	switch mode {
	case "shield-finish":
	case "weapon-block":
		typ = ai.ACTION_WEAPON_BLOCK
	case "dying":
		typ = ai.ACTION_DYING
	default:
		t.Fatal("unknown terminal mode")
	}
	a, ok := server.GetAIAction(typ).(cgoAIAction)
	if !ok {
		t.Fatal("ordinary native action dispatcher")
	}
	public := monsterTerminalPublicCEntry5320E0(mode)
	if public == nil {
		t.Fatal("public C terminal address")
	}
	rawUnit := func(p *server.Object) []byte { return unsafe.Slice((*byte)(unsafe.Pointer(p)), int(unsafe.Sizeof(*p))) }
	rawFrame := func(p *guarded) []byte { return unsafe.Slice((*byte)(unsafe.Pointer(p)), int(unsafe.Sizeof(*p))) }
	calls := 0
	for _, low := range []uint8{0, 16, 128} {
		unit := units[low]
		unit.ObjClass, unit.ObjFlags, unit.UpdateData = object.ClassMonster, 0, unsafe.Pointer(&f.Update)
		for _, ready := range []uint8{0, 1, 0x7f, 0x80, 0xff} {
			for _, index := range []int8{0, 1, 2} {
				*f = guarded{}
				for i := range f.Before {
					f.Before[i], f.After[i] = byte(0x31+i), byte(0xdb-i)
				}
				f.Update.Field120_3, f.Update.AIStackInd = ready, index
				for i := int8(0); i <= index; i++ {
					f.Update.AIStack[i].Action = uint32(ai.ACTION_IDLE)
				}
				f.Update.AIStack[index].Action = uint32(typ)
				if unsafe.Sizeof(uintptr(0)) == 8 && (uintptr(unit.CObj()) <= 0xffffffff || uintptr(unit.UpdateData) <= 0xffffffff) {
					t.Fatal("public composition links must exceed4GiB")
				}
				beforeUnit, beforeFrame := *unit, *f
				s.AI.StackChanged = false
				a.Update(unit)
				expectedUnit, expectedFrame, changed := *unit, *f, s.AI.StackChanged
				*unit, *f = beforeUnit, beforeFrame
				// PopAction's body is separately exercised here, without started
				// End callbacks. Its real return selects an already frozen original
				// root context; no declared Pop stub runs on the native C path.
				popReturn := int(index) - 1
				if ready != 0 {
					if got := unit.MonsterPopAction(); got != popReturn {
						t.Fatalf("stock stack pop return=%d want=%d", got, popReturn)
					}
					if !bytes.Equal(rawUnit(unit), rawUnit(&expectedUnit)) || !bytes.Equal(rawFrame(f), rawFrame(&expectedFrame)) {
						t.Fatal("ordinary terminal differs from real PopAction")
					}
				}
				want, exists := wantBytes[key(mode, ready, low, uint32(int32(popReturn)))]
				if !exists {
					t.Fatal("missing frozen original char result")
				}
				*unit, *f = beforeUnit, beforeFrame
				s.AI.StackChanged = false
				t.Logf("TERMINAL ORDINARY COMPLETED: mode=%s unit=%p update=%p low=%d ready=%d index=%d", mode, unit, unit.UpdateData, low, ready, index)
				got := uint8(ccall.CallIntPtr(public, unit.CObj()))
				if got != want || s.AI.StackChanged != changed || !bytes.Equal(rawUnit(unit), rawUnit(&expectedUnit)) || !bytes.Equal(rawFrame(f), rawFrame(&expectedFrame)) {
					t.Fatalf("public C terminal mismatch: mode=%s low=%d ready=%d index=%d char=%d/%d changed=%t/%t", mode, low, ready, index, got, want, s.AI.StackChanged, changed)
				}
				calls++
				t.Logf("TERMINAL NATIVE C PASS: mode=%s low=%d ready=%d index=%d char=%d original=true actual-PopAction=true full-canaries=true", mode, low, ready, index, got)
			}
		}
	}
	if calls != 45 {
		t.Fatal("incomplete native terminal compositions", calls)
	}
}
