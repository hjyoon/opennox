package legacy

import (
	"bytes"
	"encoding/json"
	"math"
	"os"
	"os/exec"
	"testing"
	"unsafe"

	"github.com/opennox/libs/object"
	"github.com/opennox/libs/strman"
	"github.com/opennox/libs/types"
	noxflags "github.com/opennox/opennox/v1/common/flags"
	"github.com/opennox/opennox/v1/common/unit/ai"
	"github.com/opennox/opennox/v1/legacy/common/alloc"
	"github.com/opennox/opennox/v1/legacy/common/alloc/handles"
	"github.com/opennox/opennox/v1/legacy/common/ccall"
	"github.com/opennox/opennox/v1/server"
)

// Frozen original direction stores/pop decisions reach the actual registered
// C entries and production PopAction/reset. The original PopAction body stays
// a declared boundary; this composition checks its native effects separately.
// No started-action End callbacks or earned campaign AI behavior are claimed.
func TestMonsterFacingPublicCNativeComposition545210(t *testing.T) {
	const child = "NOX_MONSTER_FACING_COMPOSITION_CHILD"
	name := os.Getenv(child)
	if name == "" {
		binary, err := os.Executable()
		if err != nil {
			t.Fatal(err)
		}
		for _, name := range []string{"location-basic-0", "location-basic-1", "object-basic-128", "object-nil", "angle-192", "angle-196", "angle-207", "set-angle-ffffffff", "set-angle-12345678", "location-edge-7f7fffff", "object-edge-7f7fffff", "location-edge-7f800000", "object-edge-ff800000", "location-edge-7fc12345", "object-edge-ffc12345"} {
			t.Run(name, func(t *testing.T) {
				cmd := exec.Command(binary, "-test.run=^TestMonsterFacingPublicCNativeComposition545210$", "-test.v")
				cmd.Env = append(os.Environ(), child+"="+name)
				out, err := cmd.CombinedOutput()
				t.Logf("%s", out)
				if err != nil {
					t.Fatalf("actual facing C composition: %v", err)
				}
			})
		}
		return
	}
	raw, err := os.ReadFile("../server/testdata/monster_facing_native_original_545210.json")
	if err != nil {
		t.Fatal(err)
	}
	var fixture struct {
		Cases []struct {
			Input  map[string]any `json:"input"`
			Result map[string]any `json:"result"`
		} `json:"cases"`
	}
	if err := json.Unmarshal(raw, &fixture); err != nil {
		t.Fatal(err)
	}
	if len(fixture.Cases) != 1352 {
		t.Fatal("sealed original corpus")
	}
	index := -1
	for i, row := range fixture.Cases {
		if row.Input["name"] == name {
			index = i
			break
		}
	}
	if index < 0 {
		t.Fatal("missing original context")
	}
	c, result := fixture.Cases[index].Input, fixture.Cases[index].Result
	u := func(m map[string]any, k string) uint32 { return uint32(m[k].(float64)) }
	pt := func(v any) types.Pointf {
		a := v.([]any)
		return types.Ptf(math.Float32frombits(uint32(a[0].(float64))), math.Float32frombits(uint32(a[1].(float64))))
	}
	handles.Init()
	t.Cleanup(handles.Release)
	s := server.New(nil, nil, strman.New())
	t.Cleanup(s.Close)
	s.SetFrame(150)
	oldGet := GetServer
	GetServer = func() Server { return &facingProbeLegacyServer545210{srv: s} }
	t.Cleanup(func() { GetServer = oldGet })
	oldEngine := noxflags.GetEngine()
	noxflags.UnsetEngine(oldEngine)
	t.Cleanup(func() { noxflags.UnsetEngine(noxflags.GetEngine()); noxflags.SetEngine(oldEngine) })
	if !s.Objs.Init(4) {
		t.Fatal("native object pool")
	}
	t.Cleanup(s.Objs.FreeObjects)
	unit, target := s.Objs.NewObject(&server.ObjectType{}), s.Objs.NewObject(&server.ObjectType{})
	unit.ObjClass, unit.ObjFlags = object.ClassMonster, 0
	unit.PosVec, unit.Direction1, unit.Direction2 = pt(c["pos"]), server.Dir16(u(c, "direction1")), server.Dir16(u(c, "direction2"))
	target.PosVec = pt(c["target_pos"])
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
	function := c["function"].(string)
	typesByName := map[string]ai.ActionType{"location": ai.ACTION_FACE_LOCATION, "object": ai.ACTION_FACE_OBJECT, "angle": ai.ACTION_FACE_ANGLE, "set-angle": ai.ACTION_SET_ANGLE}
	typ, ok := typesByName[function]
	if !ok {
		t.Fatal("context is not a public facing root")
	}
	f.Update.AIStackInd = int8(u(c, "head_index"))
	if f.Update.AIStackInd != 1 {
		t.Fatal("bounded pop-to-idle composition")
	}
	f.Update.AIStack[0].Action = uint32(ai.ACTION_IDLE)
	head := &f.Update.AIStack[1]
	head.Action = uint32(typ)
	switch function {
	case "location":
		head.SetArgs(pt(c["target_pos"]))
	case "object":
		if !c["target_nil"].(bool) {
			head.SetArgs(target)
		}
	default:
		head.SetArgs(u(c, "angle"))
	}
	f.Update.Field2, f.Update.Field67, f.Update.Field74, f.Update.Field91 = 7, 8, 9, 10
	f.Update.Field120_1, f.Update.Field120_2, f.Update.Field120_3 = 1, 2, 3
	f.Update.Field124, f.Update.Field137 = 91, 92
	if unsafe.Sizeof(uintptr(0)) == 8 && (uintptr(unit.CObj()) <= 0xffffffff || uintptr(target.CObj()) <= 0xffffffff || uintptr(unit.UpdateData) <= 0xffffffff) {
		t.Fatal("native C-owned pointers above4GiB")
	}
	expectedUnit, expectedFrame := *unit, *f
	state := result["state"].(map[string]any)
	expectedUnit.Direction1, expectedUnit.Direction2 = server.Dir16(u(state, "direction1")), server.Dir16(u(state, "direction2"))
	wantPops := len(result["events"].([]any))
	if wantPops > 1 {
		t.Fatal("original pop count")
	}
	if wantPops == 1 {
		expectedFrame.Update.AIStackInd = 0
		expectedFrame.Update.Field2, expectedFrame.Update.Field67, expectedFrame.Update.Field74, expectedFrame.Update.Field91 = 0, 0, 0, 0
		expectedFrame.Update.Field120_1, expectedFrame.Update.Field120_2, expectedFrame.Update.Field120_3 = 0, 0, 0
		expectedFrame.Update.Field124, expectedFrame.Update.Field137 = 150, 150
	}
	rawObject := func(p *server.Object) []byte { return unsafe.Slice((*byte)(unsafe.Pointer(p)), int(unsafe.Sizeof(*p))) }
	rawFrame := func(p *guarded) []byte { return unsafe.Slice((*byte)(unsafe.Pointer(p)), int(unsafe.Sizeof(*p))) }
	beforeTarget := append([]byte(nil), rawObject(target)...)
	a, ok := server.GetAIAction(typ).(cgoAIAction)
	if !ok || a.update == nil {
		t.Fatal("native public C registry pointer")
	}
	s.AI.StackChanged = false
	ccall.CallVoidPtr(a.update, unit.CObj())
	if !bytes.Equal(rawObject(unit), rawObject(&expectedUnit)) || !bytes.Equal(rawFrame(f), rawFrame(&expectedFrame)) || !bytes.Equal(rawObject(target), beforeTarget) || s.AI.StackChanged != (wantPops != 0) {
		t.Fatalf("C context=%s dirs=%d/%d original=%d/%d stack=%d want=%d changed=%t", name, unit.Direction1, unit.Direction2, expectedUnit.Direction1, expectedUnit.Direction2, f.Update.AIStackInd, expectedFrame.Update.AIStackInd, s.AI.StackChanged)
	}
	t.Logf("C facing context=%s unit=%p target=%p update=%p directions=%d/%d pops=%d; real stack/reset and all object/update/target canaries verified", name, unit, target, unit.UpdateData, unit.Direction1, unit.Direction2, wantPops)
}
