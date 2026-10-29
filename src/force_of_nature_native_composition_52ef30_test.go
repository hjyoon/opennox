package opennox

import (
	"bytes"
	_ "embed"
	"encoding/binary"
	"encoding/json"
	"math"
	"os"
	"os/exec"
	"testing"
	"unsafe"

	"github.com/opennox/libs/object"
	"github.com/opennox/libs/spell"
	"github.com/opennox/libs/strman"
	"github.com/opennox/libs/things"
	"github.com/opennox/libs/types"
	noxflags "github.com/opennox/opennox/v1/common/flags"
	"github.com/opennox/opennox/v1/common/sound"
	"github.com/opennox/opennox/v1/legacy"
	"github.com/opennox/opennox/v1/legacy/common/alloc/handles"
	"github.com/opennox/opennox/v1/legacy/common/ccall"
	"github.com/opennox/opennox/v1/server"
)

//go:embed server/testdata/spell_force_of_nature_native_original_52ef30.json
var forceCOriginal52EF30 []byte

// Real C entries compose the production factory, CreateObjectAt, charge
// sidecar, ray, deletion list and parsed AUD event. Arithmetic/positions come
// from the unchanged original fixture, with the actual empty-world ray's
// declared boolean selected before C update. Synthetic types/actors provide
// no original whole-world or earned campaign equivalence.
func TestForceOfNaturePublicCNativeComposition52EF30(t *testing.T) {
	const child = "NOX_FORCE_OF_NATURE_NATIVE_COMPOSITION_CHILD"
	mode := os.Getenv(child)
	if mode == "" {
		binary, err := os.Executable()
		if err != nil {
			t.Fatal(err)
		}
		for _, mode := range []string{"launch-circle", "glyph-one", "glyph-raw", "cancel-charge"} {
			t.Run(mode, func(t *testing.T) {
				cmd := exec.Command(binary, "-test.run=^TestForceOfNaturePublicCNativeComposition52EF30$", "-test.v")
				cmd.Env = append(os.Environ(), child+"="+mode)
				out, err := cmd.CombinedOutput()
				t.Logf("%s", out)
				if err != nil {
					t.Fatalf("actual ForceOfNature C composition: %v", err)
				}
			})
		}
		return
	}
	var fixture struct {
		Cases []struct {
			Input  map[string]any `json:"input"`
			Result map[string]any `json:"result"`
		} `json:"cases"`
	}
	if err := json.Unmarshal(forceCOriginal52EF30, &fixture); err != nil {
		t.Fatal(err)
	}
	if len(fixture.Cases) != 925 {
		t.Fatal("sealed original corpus")
	}
	u := func(m map[string]any, k string) uint32 { return uint32(m[k].(float64)) }
	point := func(v any) types.Pointf {
		p := v.([]any)
		return types.Ptf(math.Float32frombits(uint32(p[0].(float64))), math.Float32frombits(uint32(p[1].(float64))))
	}
	bits := func(p types.Pointf) [2]uint32 { return [2]uint32{math.Float32bits(p.X), math.Float32bits(p.Y)} }
	recordMode := uint32(0)
	if mode == "glyph-one" {
		recordMode = 1
	}
	if mode == "glyph-raw" {
		recordMode = math.MaxUint32
	}
	createIndex, launchIndex := -1, -1
	for i, c := range fixture.Cases {
		in := c.Input
		if in["mutation"].(map[string]any)["stage"] != "" || !in["allocate"].(bool) || u(in, "mode") != recordMode {
			continue
		}
		if in["function"] == "create" && u(in, "class") == 2 && !in["wand"].(bool) {
			createIndex = i
		}
		if in["function"] == "update" && u(in, "frame") == 100 && u(in, "expiry") == 101 && u(in, "direction") == 33 && u(in, "class") == 2 && u(in, "shape") == 2 && in["trace"].(bool) && recordMode == 0 {
			launchIndex = i
		}
		if in["function"] == "update" && recordMode != 0 && u(in, "field72")&0xffff == 192 && in["caster_nil"].(bool) {
			launchIndex = i
		}
	}
	if createIndex < 0 || launchIndex < 0 && mode != "cancel-charge" {
		t.Fatal("original create/launch contexts")
	}
	createCase := fixture.Cases[createIndex]
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
	t.Cleanup(s.spells.duration.Free)
	s.SetTickRate(30)
	s.SetFrame(100)
	s.Map.Init()
	t.Cleanup(s.Map.Free)
	if s.Walls.Init() == 0 {
		t.Fatal("native walls")
	}
	t.Cleanup(s.Walls.Free)
	s.FreeObjectTypes()
	t.Cleanup(s.FreeObjectTypes)
	for _, name := range []string{"NativeForceActor", "ForceOfNatureCharge", "DeathBall"} {
		if err := s.Types.ReadObjectType(&things.Thing{Name: name}); err != nil {
			t.Fatal(err)
		}
	}
	s.Types.ByID("DeathBall").Speed = 5.125
	if !s.Objs.Init(16) {
		t.Fatal("native objects")
	}
	t.Cleanup(s.Objs.FreeObjects)
	s.Audio.Init(s.Server)
	t.Cleanup(s.Audio.Free)
	name := sound.ID(38).String()
	wire := binary.LittleEndian.AppendUint32(nil, 1)
	wire = append(wire, byte(len(name)))
	wire = append(wire, name...)
	wire = binary.LittleEndian.AppendUint16(wire, 8)
	wire = append(wire, 0)
	wire = binary.LittleEndian.AppendUint16(wire, 40)
	wire = append(wire, 0, 0, 0, 0, 1, 's', 0)
	file := luaCompatAudioFile{bytes.NewReader(wire)}
	if !s.Audio.ReadAUD(file) || file.Len() != 0 || s.Audio.Field12(38) != 1 {
		t.Fatal("real AUD descriptor")
	}
	s.Audio.Reset()
	caster := s.Objs.NewObject(s.Types.ByID("NativeForceActor"))
	anchor := s.Objs.NewObject(s.Types.ByID("NativeForceActor"))
	caster.ObjClass, caster.ObjFlags = object.ClassMonster, 0
	caster.PosVec, caster.NewPos, caster.Direction1 = types.Ptf(123.125, -71.25), types.Ptf(123.125, -71.25), 33
	caster.Shape.Kind, caster.Shape.Circle.R = server.ShapeKindCircle, 3.25
	anchor.ObjClass, anchor.ObjFlags = object.ClassMonster, 0
	anchor.PosVec, anchor.NewPos, anchor.Direction1 = types.Ptf(71.125, 93.5), types.Ptf(71.125, 93.5), 192
	type guarded struct {
		Before [32]byte
		Record server.DurSpell
		After  [32]byte
	}
	f := plasmaNativeAllocation531580[guarded](t)
	for i := range f.Before {
		f.Before[i], f.After[i] = byte(0x51+i), byte(0xe7-i)
	}
	r := &f.Record
	*r = server.DurSpell{ID: 0xabcd, Spell: uint32(spell.SPELL_FORCE_OF_NATURE), Level: 3, Caster16: caster, Obj24: anchor, Flag20: recordMode, Pos: types.Ptf(100, 200), Frame68: 120, Field72: 0x6a5a0080, Field76: 0x3fdccccc, Flags88: 0x12340001}
	if unsafe.Sizeof(uintptr(0)) == 8 && (uintptr(r.C()) <= math.MaxUint32 || uintptr(caster.CObj()) <= math.MaxUint32) {
		t.Fatal("native C-owned links above4GiB")
	}
	expected := *f
	expected.Record.Field72 = int32(u(createCase.Result["state"].(map[string]any), "field72"))
	expected.Record.Flags88 = u(createCase.Result["state"].(map[string]any), "flags88")
	raw := func(p *guarded) []byte { return unsafe.Slice((*byte)(unsafe.Pointer(p)), int(unsafe.Sizeof(*p))) }
	if got := ccall.CallIntPtr(legacy.Get_sub_52EF30(), r.C()); got != 0 {
		t.Fatalf("C create=%d", got)
	}
	charge := s.spells.duration.forceOfNatureCharges[r]
	if charge == nil || int(charge.TypeInd) != s.Types.IndByID("ForceOfNatureCharge") || charge.Flags().Has(object.FlagDestroyed) {
		t.Fatal("real native charge allocation/store")
	}
	if bits(charge.PosVec) != bits(point(createCase.Result["state"].(map[string]any)["charge_pos"])) || !bytes.Equal(raw(f), raw(&expected)) {
		t.Fatal("original charge position or duration/canary")
	}
	if unsafe.Sizeof(uintptr(0)) == 8 && uintptr(charge.CObj()) <= math.MaxUint32 {
		t.Fatal("native charge above4GiB")
	}
	if mode == "cancel-charge" {
		ccall.CallVoidPtr(legacy.Get_sub_52F1D0(), r.C())
		if s.spells.duration.forceOfNatureCharges[r] != nil || !charge.Flags().Has(object.FlagDestroyed) || s.Objs.DeletedList != charge || !bytes.Equal(raw(f), raw(&expected)) {
			t.Fatal("real C cancellation/deletion/sidecar cleanup")
		}
		t.Logf("C cancel record=%p charge=%p real deletion and native canaries verified", r, charge)
		return
	}
	s.SetFrame(113)
	if got := ccall.CallIntPtr(legacy.Get_sub_52EFD0(), r.C()); got != 0 || s.spells.duration.forceOfNatureCharges[r] != nil || !charge.Flags().Has(object.FlagDestroyed) || s.Objs.DeletedList != charge {
		t.Fatal("real seven-frame charge retirement")
	}
	launch := fixture.Cases[launchIndex]
	var lineFrom, lineTo types.Pointf
	for _, e := range launch.Result["events"].([]any) {
		ev := e.(map[string]any)
		if ev["kind"] == "trace" {
			a := ev["args"].([]any)
			lineFrom, lineTo = point(a[0]), point(a[1])
		}
	}
	actualRay := s.MapTraceRay(lineFrom, lineTo, server.MapTraceFlag1|server.MapTraceFlag3)
	if recordMode == 0 && !actualRay {
		in := launch.Input
		for _, c := range fixture.Cases {
			other := c.Input
			if other["function"] == "update" && other["mutation"].(map[string]any)["stage"] == "" && !other["trace"].(bool) && u(other, "mode") == 0 && u(other, "class") == u(in, "class") && u(other, "shape") == u(in, "shape") && u(other, "direction") == u(in, "direction") && u(other, "frame") == 100 && u(other, "expiry") == 101 {
				launch = c
				break
			}
		}
		if launch.Input["trace"].(bool) {
			t.Fatal("original declared false-ray context")
		}
	}
	var ball *server.Object
	audioCalls := 0
	s.Audio.OnSound(func(id sound.ID, kind int, obj *server.Object, pos types.Pointf) {
		if id != 38 || kind != 0 || obj != caster || bits(pos) != bits(caster.PosVec) {
			t.Fatal("real object audio arguments")
		}
		audioCalls++
		ball = s.Objs.Pending
	})
	s.SetFrame(119)
	if got := ccall.CallIntPtr(legacy.Get_sub_52EFD0(), r.C()); got != 1 {
		t.Fatalf("C launch=%d", got)
	}
	want := launch.Result["state"].(map[string]any)
	wantDir := want["ball_direction"].([]any)
	if ball == nil || int(ball.TypeInd) != s.Types.IndByID("DeathBall") || ball.ObjOwner != caster || math.Float32bits(ball.SpeedCur) != math.Float32bits(5.125) || audioCalls != 1 || s.spells.duration.forceOfNatureLaunches != 1 {
		t.Fatal("real DeathBall create/owner/speed/audio")
	}
	if bits(ball.PosVec) != bits(point(want["ball_pos"])) || bits(ball.VelVec) != bits(point(want["ball_velocity"])) || uint16(ball.Direction1) != uint16(wantDir[0].(float64)) || uint16(ball.Direction2) != uint16(wantDir[1].(float64)) {
		t.Fatalf("real ball position/velocity/direction: pos=%#v velocity=%#v dirs=%d/%d original=%v", bits(ball.PosVec), bits(ball.VelVec), ball.Direction1, ball.Direction2, want)
	}
	if !bytes.Equal(raw(f), raw(&expected)) {
		t.Fatal("launch changed native record/canary")
	}
	ccall.CallVoidPtr(legacy.Get_sub_52F1D0(), r.C())
	if s.spells.duration.forceOfNatureCharges[r] != nil || !bytes.Equal(raw(f), raw(&expected)) {
		t.Fatal("C destroy stale sidecar/record")
	}
	t.Logf("C lifecycle mode=%s record=%p charge=%p ball=%p ray=%t pos=%#v velocity=%#v; real factory/creation/deletion/audio/native canaries verified", mode, r, charge, ball, actualRay, bits(ball.PosVec), bits(ball.VelVec))
}
