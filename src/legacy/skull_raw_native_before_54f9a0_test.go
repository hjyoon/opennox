package legacy

import (
	"bytes"
	"context"
	"encoding/binary"
	"encoding/json"
	"math"
	"os"
	"os/exec"
	"strconv"
	"testing"
	"time"
	"unsafe"

	"github.com/opennox/libs/things"
	"github.com/opennox/libs/types"
	"github.com/opennox/opennox/v1/common/sound"
	"github.com/opennox/opennox/v1/server"
)

func TestSkullRawNativeLayout54F9A0(t *testing.T) {
	o, d, c := server.Object{}, server.SkullUpdateData{}, server.ProjectileCollideData{}
	nativePrefix := unsafe.Offsetof(o.ScriptPickup) + unsafe.Sizeof(o.ScriptPickup)
	if unsafe.Sizeof(o) != nativePrefix+2*unsafe.Sizeof(uintptr(0)) {
		t.Fatal("Go object extends the complete C prefix by exactly its two native handle words")
	}
	want := [19]uintptr{
		nativePrefix, unsafe.Offsetof(o.TypeInd), unsafe.Offsetof(o.ObjClass), unsafe.Offsetof(o.ObjFlags),
		unsafe.Offsetof(o.PosVec), unsafe.Offsetof(o.PosVec) + 4,
		unsafe.Offsetof(o.Direction1), unsafe.Offsetof(o.Direction2),
		unsafe.Offsetof(o.Shape) + unsafe.Offsetof(o.Shape.Circle), unsafe.Offsetof(o.SpeedCur),
		unsafe.Offsetof(o.UpdateData), unsafe.Offsetof(o.CollideData), unsafe.Sizeof(d),
		unsafe.Offsetof(d.Enabled), unsafe.Offsetof(d.ProjectileType), unsafe.Offsetof(d.TargetReady),
		unsafe.Sizeof(c), unsafe.Offsetof(c.Damage), unsafe.Offsetof(d.ProjectileName),
	}
	if got := skullOriginalLayout54F9A0(); got != want {
		t.Fatal("complete independent native C/Go object, Skull and projectile layouts", got, want)
	}
}

func TestSkullRawFrozenReferences54F9A0(t *testing.T) {
	var inputs []struct {
		Name, Entry    string
		Unit, Target   int
		Flags          uint32
		SourceType     uint16 `json:"source_type"`
		Direction      uint16
		ProjectileType uint32 `json:"projectile_type"`
	}
	trollDeathFixture54A270(t, "skull_raw_independent_inputs_54f9a0.json.gz", "2fa912b92071a20903b0596adce24a80942874efeb6e80fe5b8d8f6def6dcc36", "8a004ff539f042e37bfa144916780ab9042d5a6ed48889229bb6d6e18bce8825", &inputs)
	if len(inputs) != 753 || inputs[0].Entry != "54f9a0" || inputs[0].Flags != 0x01000000 || inputs[0].SourceType != 85 || inputs[0].Direction != 1 || inputs[0].ProjectileType != 73 {
		t.Fatal("complete independent inputs and scalar transport")
	}
	for _, profile := range []struct{ name, encoded, decoded string }{
		{"tozero", "c5033d70e53a39874dd667354eb9ead31d970e1ca63ab73b3603b6334dedde93", "f9269f877bc70d69e4defb25eecb5c200ad742df9d468d469fa152c4e06c288b"},
		{"nearest", "08d9557293ef26581f612b6177b1df60f0e579addb9db47cf0c5b0e755d658de", "a7515ca2c9beb823f28871bf1c3256db6cd9a625c8a7078f7b4276110dccf76b"},
	} {
		var out []struct {
			Name   string
			Result *uint32
			Events []json.RawMessage
		}
		trollDeathFixture54A270(t, "skull_raw_native_original_54f9a0_pc53_"+profile.name+".json.gz", profile.encoded, profile.decoded, &out)
		if len(out) != len(inputs) || out[0].Result == nil || *out[0].Result != 29 {
			t.Fatal("complete original output transport", profile.name)
		}
		for i := range inputs {
			if out[i].Name != inputs[i].Name {
				t.Fatal("frozen independent input/output identity", i)
			}
		}
	}
}

func skullRawActualFixture54F9A0(t *testing.T) (source, target *server.Object, projectileID uint32) {
	t.Helper()
	TestSkullRawNativeLayout54F9A0(t)
	s := npcReflectLegacyServer4E17B0(t)
	s.Audio.Init(s)
	t.Cleanup(s.Audio.Free)
	s.FreeObjectTypes()
	if !s.Objs.Init(8) {
		t.Fatal("actual native object pool")
	}
	t.Cleanup(s.Objs.FreeObjects)
	t.Cleanup(s.FreeObjectTypes)
	for _, def := range []*things.Thing{
		{Name: "ArrowTrap1", OnUpdate: &things.ProcFunc{Name: "SkullUpdate"}},
		{Name: "MercArcherArrow", OnCollide: &things.ProcFunc{Name: "ProjectileCollide"}},
	} {
		if err := s.Types.ReadObjectType(def); err != nil {
			t.Fatal(err)
		}
	}
	sourceID := skullOriginalType54F9A0("ArrowTrap1")
	projectileID = skullOriginalType54F9A0("MercArcherArrow")
	if sourceID == 0 || projectileID == 0 || sourceID == projectileID || s.Types.ByID("ArrowTrap1").Ind() != int(sourceID) || s.Types.ByID("MercArcherArrow").Ind() != int(projectileID) {
		t.Fatal("real typed C numeric name resolution")
	}
	source, target = skullOriginalNew54F9A0(sourceID), skullOriginalNew54F9A0(sourceID)
	projectile := skullOriginalNew54F9A0(projectileID)
	if source == nil || target == nil || projectile == nil || source.UpdateData == nil || target.UpdateData == nil || source.Update == nil || projectile.CollideData == nil || projectile.Collide == nil {
		t.Fatal("actual C constructors and registered native 52/8-byte data records")
	}
	for _, unit := range []*server.Object{source, target, projectile} {
		t.Cleanup(func() { s.Objs.FreeObject(unit) })
		if unsafe.Sizeof(uintptr(0)) == 8 {
			for _, address := range []unsafe.Pointer{unsafe.Pointer(unit), unit.UpdateData, unit.CollideData} {
				if address != nil && uintptr(address) <= math.MaxUint32 {
					t.Fatal("actual C constructor object/data address must exceed4GiB", address)
				}
			}
		}
	}
	source.PosVec, target.PosVec = types.Ptf(300, 300), types.Ptf(320, 300)
	source.ObjFlags, source.Direction1, source.Direction2 = 0x11223300, 0, 77
	source.Shape.Kind, source.Shape.Circle.R, source.SpeedCur = server.ShapeKindCircle, 14, 1.25
	data := (*server.SkullUpdateData)(source.UpdateData)
	*data = server.SkullUpdateData{ScanDelay: 17, FireDelay: 19, TargetReady: 1, ProjectileType: projectileID, Enabled: 1}
	copy(data.ProjectileName[:], "MercArcherArrow")
	collide := (*server.ProjectileCollideData)(projectile.CollideData)
	collide.Damage, collide.Field4 = 1234, -321
	return skullRawActualControls54F9A0(t, s, source, target, projectile, projectileID)
}

func skullRawActualControls54F9A0(t *testing.T, s *server.Server, source, target, projectile *server.Object, id uint32) (*server.Object, *server.Object, uint32) {
	t.Helper()
	frames := [][]byte{
		unsafe.Slice((*byte)(unsafe.Pointer(source)), int(unsafe.Sizeof(*source))),
		unsafe.Slice((*byte)(unsafe.Pointer(target)), int(unsafe.Sizeof(*target))),
		unsafe.Slice((*byte)(unsafe.Pointer(projectile)), int(unsafe.Sizeof(*projectile))),
		unsafe.Slice((*byte)(source.UpdateData), 52), unsafe.Slice((*byte)(target.UpdateData), 52),
		unsafe.Slice((*byte)(projectile.CollideData), 8),
	}
	before := make([][]byte, len(frames))
	for i := range frames {
		before[i] = bytes.Clone(frames[i])
	}
	for _, unit := range []*server.Object{source, projectile} {
		wantPointers := [7]uintptr{uintptr(unsafe.Pointer(unit)), uintptr(unit.UpdateData), uintptr(unit.CollideData), uintptr(unsafe.Pointer(&unit.PosVec.X)), uintptr(unsafe.Pointer(&unit.PosVec.Y)), uintptr(unsafe.Pointer(&unit.Shape.Circle.R)), uintptr(unsafe.Pointer(&unit.SpeedCur))}
		wantWords := [16]uint32{uint32(unit.TypeInd), uint32(unit.ObjClass), uint32(unit.ObjFlags), math.Float32bits(unit.PosVec.X), math.Float32bits(unit.PosVec.Y), uint32(unit.Direction1), uint32(unit.Direction2), math.Float32bits(unit.Shape.Circle.R), math.Float32bits(unit.SpeedCur)}
		if unit.UpdateData != nil {
			d := unit.UpdateDataSkull()
			wantWords[9], wantWords[10], wantWords[11], wantWords[12], wantWords[13] = d.ScanDelay, d.FireDelay, uint32(d.TargetReady), d.ProjectileType, uint32(d.Enabled)
		}
		if unit.CollideData != nil {
			d := (*server.ProjectileCollideData)(unit.CollideData)
			wantWords[14], wantWords[15] = uint32(d.Damage), uint32(d.Field4)
		}
		pointers, words := skullOriginalControl54F9A0(unit)
		if pointers != wantPointers || words != wantWords {
			t.Fatal("actual C constructor/control full native object/data pointers and scalars", pointers, words, wantPointers, wantWords)
		}
	}
	if got := skullOriginalDirectionControl54F9A0(source, target); got != 1 {
		t.Fatal("actual typed C direction service and live native point pointers", got)
	}
	name := sound.ID(889).String()
	wire := binary.LittleEndian.AppendUint32(nil, 1)
	wire = append(wire, byte(len(name)))
	wire = append(wire, name...)
	wire = append(wire, 0, 0, 0, 40, 0, 1, 0, 0, 0, 1, 'x', 0)
	file := simplePolypAUDFile54CAE0{bytes.NewReader(wire)}
	if !s.Audio.ReadAUD(file) || file.Len() != 0 || s.Audio.Field12(889) != 1 {
		t.Fatal("actual stock sound descriptor reader")
	}
	s.Audio.Reset()
	skullOriginalAudioControl54F9A0(source)
	count := 0
	s.Audio.EachEvent(func(event *server.AudioEvent) {
		count++
		if event.Sound != 889 || event.Obj != source || event.Pos != source.PosVec || event.Kind != 0 || event.Code != 0 {
			t.Fatal("actual typed C audio and complete native event identity", event)
		}
	})
	if count != 1 {
		t.Fatal("actual typed C audio event count", count)
	}
	for i := range frames {
		if !bytes.Equal(before[i], frames[i]) {
			t.Fatal("complete native object/update/collide guard changed by controls", i)
		}
	}
	return source, target, id
}

func TestSkullRawActualNativeControls54F9A0(t *testing.T) {
	skullRawActualFixture54F9A0(t)
}

func TestSkullRawActualCBefore54F9A0(t *testing.T) {
	if os.Getenv("OPENNOX_SKULL_RAW_C_BEFORE") != "1" {
		return
	}
	if entry := os.Getenv("OPENNOX_SKULL_RAW_C_CHILD"); entry != "" {
		source, target, id := skullRawActualFixture54F9A0(t)
		index, err := strconv.Atoi(entry)
		if err != nil || index < 0 || index > 4 || unsafe.Sizeof(uintptr(0)) != 8 || skullOriginalABI54F9A0(index) != 32 {
			t.Fatal("actual unchanged signed32 C entry required")
		}
		t.Logf("SKULL_REAL_C_CONTROLS entry=%d source=%p target=%p full-native-layout-real-constructors-update52-collide8-direction-audio-whole-records=true", index, source, target)
		skullOriginalC54F9A0(index, source, target, id)
		t.Fatal("old signed32 C entry unexpectedly returned")
	}
	if unsafe.Sizeof(uintptr(0)) != 8 {
		t.Fatal("old pointer truncation reproduction requires a64-bit host")
	}
	t.Run("real-constructor-direction-audio-controls", func(t *testing.T) { skullRawActualFixture54F9A0(t) })
	for index, name := range []string{"update", "projectile", "wrapper", "scan", "candidate"} {
		t.Run(name, func(t *testing.T) {
			ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
			defer cancel()
			command := exec.CommandContext(ctx, os.Args[0], "-test.run=^TestSkullRawActualCBefore54F9A0$", "-test.v")
			command.Env = append(os.Environ(), "OPENNOX_SKULL_RAW_C_CHILD="+strconv.Itoa(index))
			out, err := command.CombinedOutput()
			t.Logf("actual old C %s: %v\n%s", name, err, out)
			if ctx.Err() != nil || err == nil || !bytes.Contains(out, []byte("SKULL_REAL_C_CONTROLS")) || bytes.Contains(out, []byte("unexpectedly returned")) {
				t.Fatal("actual native positive controls must precede the old C fault", err)
			}
			if !bytes.Contains(out, []byte("SIGSEGV")) && !bytes.Contains(out, []byte("SIGBUS")) && !bytes.Contains(out, []byte("unexpected fault address")) {
				t.Fatal("actual native C pointer fault required")
			}
		})
	}
}
