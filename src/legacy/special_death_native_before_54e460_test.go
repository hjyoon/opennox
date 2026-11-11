package legacy

import (
	"bytes"
	"context"
	"encoding/binary"
	"math"
	"os"
	"os/exec"
	"reflect"
	"strconv"
	"testing"
	"time"
	"unsafe"

	"github.com/opennox/libs/object"
	"github.com/opennox/libs/things"
	"github.com/opennox/libs/types"
	"github.com/opennox/opennox/v1/common/sound"
	"github.com/opennox/opennox/v1/legacy/common/alloc"
	"github.com/opennox/opennox/v1/server"
)

func TestSpecialDeathRawNativeLayout54E460(t *testing.T) {
	o, p, d, g := server.Object{}, server.Player{}, server.PlayerUpdateData{}, server.MonsterGenUpdateData{}
	counter, ok := reflect.TypeOf(p).FieldByName("field4668")
	if !ok || counter.Type.Kind() != reflect.Uint32 {
		t.Fatal("native player counter DWORD")
	}
	flags, ok := reflect.TypeOf(p).FieldByName("field4692")
	if !ok || flags.Type.Kind() != reflect.Uint32 {
		t.Fatal("native player flags DWORD")
	}
	prefix := unsafe.Offsetof(o.ScriptPickup) + unsafe.Sizeof(o.ScriptPickup)
	if unsafe.Sizeof(o) != prefix+2*unsafe.Sizeof(uintptr(0)) {
		t.Fatal("complete C object plus two Go native handle words")
	}
	want := [29]uintptr{
		prefix, unsafe.Offsetof(o.ObjClass), unsafe.Offsetof(o.ObjFlags),
		unsafe.Offsetof(o.PosVec), unsafe.Offsetof(o.PosVec) + 4,
		unsafe.Offsetof(o.Field27), unsafe.Offsetof(o.Field29),
		unsafe.Offsetof(o.ObjOwner), unsafe.Offsetof(o.Obj130), unsafe.Offsetof(o.UpdateData),
		unsafe.Sizeof(d), unsafe.Offsetof(d.Field29), unsafe.Offsetof(d.Player),
		unsafe.Sizeof(g), unsafe.Offsetof(g.Field56), unsafe.Offsetof(g.FuncInd60),
		unsafe.Sizeof(p), counter.Offset, flags.Offset, unsafe.Offsetof(d.ExtraLives), unsafe.Offsetof(d.Field137),
		unsafe.Offsetof(g.Field48), unsafe.Offsetof(g.FuncInd52), unsafe.Offsetof(g.Field64), unsafe.Offsetof(g.FuncInd68),
		unsafe.Offsetof(g.ScriptCollision), unsafe.Sizeof(g.Field0), unsafe.Sizeof(d.Field29), unsafe.Sizeof(uintptr(0)),
	}
	if got := specialDeathOriginalLayout54E460(); got != want {
		t.Fatal("independent complete C/Go object/player/update layouts", got, want)
	}
}

func TestSpecialDeathRawFrozenReferences54E460(t *testing.T) {
	for _, group := range []struct {
		input, encoded, decoded, output string
		count                           int
		profiles                        [2][2]string
	}{
		{"marker_boulder_generator_independent_inputs_54e460.json.gz", "33513f4d4be79a95526c408703bbe11ecb71bd7dfa69283c3d29f84d3dbf2cf8", "f9ea40aefdd2215f5b680df1ca90f6d5dcc170ffd84bd72350ad096e9444d0c3", "marker_boulder_generator_original_54e460", 200, [2][2]string{
			{"c575d3d1f06b4b963fc759609be895e71883b1473112780602c921d7865eeb5a", "6f790dea714ab3061aecdeed56eab8cea3d15c1c0ded1a90454795046cc6f297"},
			{"7095f87e5535ebbf97792ea37d9fd573c0274ed9955cba062653e8ab694b1557", "542f20f9b67ab5fc1ae66d856d2771d8a839c7194ee8adb04ae30b11c32cbeb7"},
		}},
		{"credit_independent_inputs_4d61b0.json.gz", "a086a617a8b90e951ac06183f77992dfb07f9c6b811757a36c668be3dff1e542", "4090b764c2022e8b8745aea57626e86eff624747d3f80518a366dff7aa4b1f0c", "generator_credit_original_4d61b0", 286, [2][2]string{
			{"0eb974de529c5318fb01df4a4319a88893d55b3dfc379e3d1ba10c7f91bca2b5", "f0c52f84e982bcf12e461aa1f86fa0712631434a17baceae2c96251ff5064759"},
			{"0eb974de529c5318fb01df4a4319a88893d55b3dfc379e3d1ba10c7f91bca2b5", "f0c52f84e982bcf12e461aa1f86fa0712631434a17baceae2c96251ff5064759"},
		}},
	} {
		var inputs []struct{ Name string }
		trollDeathFixture54A270(t, group.input, group.encoded, group.decoded, &inputs)
		if len(inputs) != group.count {
			t.Fatal("frozen independent input count", group.input)
		}
		for i, profile := range []string{"tozero", "nearest"} {
			var outputs []struct{ Name string }
			trollDeathFixture54A270(t, group.output+"_pc53_"+profile+".json.gz", group.profiles[i][0], group.profiles[i][1], &outputs)
			if len(outputs) != len(inputs) {
				t.Fatal("complete original output count", group.output)
			}
			for i := range inputs {
				if outputs[i].Name != inputs[i].Name {
					t.Fatal("independent input/original-output identity", group.output, i)
				}
			}
		}
	}
}

func specialDeathActualFixture54E460(t *testing.T) (out [4]*server.Object) {
	t.Helper()
	TestSpecialDeathRawNativeLayout54E460(t)
	s := npcReflectLegacyServer4E17B0(t)
	s.Audio.Init(s)
	t.Cleanup(s.Audio.Free)
	s.FreeObjectTypes()
	if !s.Objs.InitHigh(16) {
		t.Fatal("actual native object pool")
	}
	t.Cleanup(s.Objs.FreeObjects)
	t.Cleanup(s.FreeObjectTypes)
	var units []*server.Object
	for i, name := range []string{"Marker", "Boulder", "MonsterGenerator", "Player", "BigRock", "MediumRock", "SmallRock", "DestroyedGenerator"} {
		def := &things.Thing{Name: name}
		if i == 2 {
			def.OnUpdate = &things.ProcFunc{Name: "MonsterGeneratorUpdate"}
		} else if i == 3 {
			def.OnUpdate = &things.ProcFunc{Name: "PlayerUpdate"}
		}
		if err := s.Types.ReadObjectType(def); err != nil {
			t.Fatal(err)
		}
		id := specialDeathOriginalType54E460(name)
		if id == 0 || s.Types.ByID(name).Ind() != int(id) {
			t.Fatal("actual typed C numeric name resolver", name)
		}
		u := specialDeathOriginalNew54E460(id)
		if u == nil || uint32(u.TypeInd) != id {
			t.Fatal("actual C constructor", name)
		}
		t.Cleanup(func() { s.Objs.FreeObject(u) })
		u.ObjFlags, u.ObjClass, u.PosVec = 0x11223300, object.ClassMonster, types.Ptf(300+float32(i), 320+float32(i))
		u.Field27, u.Field29 = -1.25, 0x3fa00000+uint32(i)
		units = append(units, u)
		if i < 4 {
			out[i] = u
		}
	}
	parent := out[3]
	parent.ObjClass = object.ClassPlayer
	if parent.UpdateData == nil || parent.Update == nil || out[2].UpdateData == nil || out[2].Update == nil {
		t.Fatal("real registered C constructors with complete PlayerUpdate/MonsterGeneratorUpdate allocations")
	}
	p, free := alloc.NewHigh(server.Player{})
	t.Cleanup(free)
	d := (*server.PlayerUpdateData)(parent.UpdateData)
	d.Player, d.Field29, d.ExtraLives, d.Field137 = p, [4]*server.Object{out[0], units[4], units[5], units[6]}, 0x12345678, 0xabcdef01
	*playerDieQuestTestWord54D2B0(t, p, "field4668") = 0xaabbccdd
	*playerDieQuestTestWord54D2B0(t, p, "field4692") = 0x80000000
	g := (*server.MonsterGenUpdateData)(out[2].UpdateData)
	g.Field0[0], g.Field56, g.FuncInd60, g.Field64, g.FuncInd68 = out[0], 0x11223344, 0xffffffff, 0x55667788, 0xfffffffe
	for _, u := range out[:3] {
		u.ObjOwner, u.Obj130 = parent, parent
	}
	var records [][]byte
	for _, u := range units {
		records = append(records, unsafe.Slice((*byte)(unsafe.Pointer(u)), int(unsafe.Sizeof(*u))))
	}
	records = append(records, unsafe.Slice((*byte)(parent.UpdateData), int(unsafe.Sizeof(*d))), unsafe.Slice((*byte)(out[2].UpdateData), int(unsafe.Sizeof(*g))), unsafe.Slice((*byte)(unsafe.Pointer(p)), int(unsafe.Sizeof(*p))))
	before := make([][]byte, len(records))
	for i, record := range records {
		before[i] = bytes.Clone(record)
		if unsafe.Sizeof(uintptr(0)) == 8 && uintptr(unsafe.Pointer(&record[0])) <= math.MaxUint32 {
			t.Fatal("entire actual C object/update/player allocation must exceed4GiB")
		}
	}
	for _, u := range units {
		pointers, words := specialDeathOriginalObject54E460(u)
		wantPointers := [8]uintptr{uintptr(unsafe.Pointer(u)), uintptr(u.UpdateData), uintptr(unsafe.Pointer(u.ObjOwner)), uintptr(unsafe.Pointer(u.Obj130)), uintptr(unsafe.Pointer(&u.PosVec.X)), uintptr(unsafe.Pointer(&u.PosVec.Y)), uintptr(unsafe.Pointer(&u.Field27)), uintptr(unsafe.Pointer(&u.Field29))}
		wantWords := [7]uint32{uint32(u.TypeInd), uint32(u.ObjClass), uint32(u.ObjFlags), math.Float32bits(u.PosVec.X), math.Float32bits(u.PosVec.Y), math.Float32bits(u.Field27), u.Field29}
		if pointers != wantPointers || words != wantWords {
			t.Fatal("actual C constructor/typed C object reader", pointers, words)
		}
	}
	pp, pw := specialDeathOriginalPlayer54E460(parent)
	wantPP := [10]uintptr{uintptr(parent.UpdateData), uintptr(unsafe.Pointer(p)), uintptr(unsafe.Pointer(d.Field29[0])), uintptr(unsafe.Pointer(d.Field29[1])), uintptr(unsafe.Pointer(d.Field29[2])), uintptr(unsafe.Pointer(d.Field29[3])), uintptr(unsafe.Pointer(playerDieQuestTestWord54D2B0(t, p, "field4668"))), uintptr(unsafe.Pointer(playerDieQuestTestWord54D2B0(t, p, "field4692"))), uintptr(unsafe.Pointer(&d.ExtraLives)), uintptr(unsafe.Pointer(&d.Field137))}
	if pp != wantPP || pw != [4]uint32{0xaabbccdd, 0x80000000, 0x12345678, 0xabcdef01} {
		t.Fatal("complete actual native C player/update control", pp, pw)
	}
	gp, gw := specialDeathOriginalGenerator54E460(out[2])
	if gp != [4]uintptr{uintptr(out[2].UpdateData), uintptr(unsafe.Pointer(&g.Field56)), uintptr(unsafe.Pointer(&g.FuncInd60)), uintptr(unsafe.Pointer(out[0]))} || gw != [4]uint32{0x11223344, 0xffffffff, 0x55667788, 0xfffffffe} {
		t.Fatal("complete actual native C generator/update control", gp, gw)
	}
	for _, u := range out {
		if specialDeathOriginalParent54E460(u) != parent {
			t.Fatal("real typed C owner-chain service")
		}
	}
	ids := []sound.ID{757, 1000}
	wire := binary.LittleEndian.AppendUint32(nil, uint32(len(ids)))
	for _, id := range ids {
		name := id.String()
		wire = append(wire, byte(len(name)))
		wire = append(wire, name...)
		wire = append(wire, 0, 0, 0, 40, 0, 1, 0, 0, 0, 1, 'x', 0)
	}
	file := simplePolypAUDFile54CAE0{bytes.NewReader(wire)}
	if !s.Audio.ReadAUD(file) || file.Len() != 0 {
		t.Fatal("real sound descriptor reader")
	}
	for i, id := range ids {
		s.Audio.Reset()
		if s.Audio.Field12(id) != 1 {
			t.Fatal("real enabled audio descriptor", id)
		}
		u := out[i+1]
		specialDeathOriginalAudio54E460(uint32(id), u)
		count := 0
		s.Audio.EachEvent(func(event *server.AudioEvent) {
			count++
			if event.Sound != id || event.Obj != u || event.Pos != u.PosVec || event.Kind != 0 || event.Code != 0 {
				t.Fatal("real typed C audio object/event identity", event)
			}
		})
		if count != 1 {
			t.Fatal("one actual audio control per source", count)
		}
	}
	for i, record := range records {
		if !bytes.Equal(record, before[i]) {
			t.Fatal("whole native constructor/update/player allocation changed by read/owner/audio controls", i)
		}
	}
	return
}

func TestSpecialDeathRawActualNativeControls54E460(t *testing.T) {
	specialDeathActualFixture54E460(t)
}

func TestSpecialDeathRawActualCBefore54E460(t *testing.T) {
	if os.Getenv("OPENNOX_SPECIAL_DEATH_C_BEFORE") != "1" {
		return
	}
	if entry := os.Getenv("OPENNOX_SPECIAL_DEATH_C_CHILD"); entry != "" {
		units := specialDeathActualFixture54E460(t)
		i, err := strconv.Atoi(entry)
		if err != nil || i < 0 || i > 3 || unsafe.Sizeof(uintptr(0)) != 8 || specialDeathOriginalABI54E460(i) != 32 {
			t.Fatal("actual unchanged signed32 C entry required")
		}
		t.Logf("SPECIAL_DEATH_REAL_C_CONTROLS entry=%d source=%p full-native-layout-eight-C-constructors-complete-updates-player-owner-audio-whole-records=true", i, units[i])
		specialDeathOriginalC54E460(i, units[i])
		t.Fatal("old signed32 C entry unexpectedly returned")
	}
	if unsafe.Sizeof(uintptr(0)) != 8 {
		t.Fatal("old C pointer truncation reproduction requires a64-bit host")
	}
	t.Run("real-constructors-owner-audio-controls", func(t *testing.T) { specialDeathActualFixture54E460(t) })
	for i, name := range []string{"Marker", "Boulder", "MonsterGenerator", "GeneratorCredit"} {
		t.Run(name, func(t *testing.T) {
			ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
			defer cancel()
			cmd := exec.CommandContext(ctx, os.Args[0], "-test.run=^TestSpecialDeathRawActualCBefore54E460$", "-test.v")
			cmd.Env = append(os.Environ(), "OPENNOX_SPECIAL_DEATH_C_CHILD="+strconv.Itoa(i))
			out, err := cmd.CombinedOutput()
			t.Logf("actual unchanged C %s: %v\n%s", name, err, out)
			if ctx.Err() != nil || err == nil || !bytes.Contains(out, []byte("SPECIAL_DEATH_REAL_C_CONTROLS")) || bytes.Contains(out, []byte("unexpectedly returned")) {
				t.Fatal("complete native positive controls must precede actual old C fault", err)
			}
			if !bytes.Contains(out, []byte("SIGSEGV")) && !bytes.Contains(out, []byte("SIGBUS")) && !bytes.Contains(out, []byte("unexpected fault address")) {
				t.Fatal("actual native C pointer fault required")
			}
		})
	}
}
