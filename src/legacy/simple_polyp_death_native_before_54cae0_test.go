package legacy

import (
	"bytes"
	"context"
	"encoding/binary"
	"encoding/json"
	"io"
	"math"
	"os"
	"os/exec"
	"strconv"
	"testing"
	"time"
	"unsafe"

	"github.com/opennox/libs/things"
	"github.com/opennox/libs/types"
	"github.com/opennox/opennox/v1/common/memmap"
	"github.com/opennox/opennox/v1/common/sound"
	"github.com/opennox/opennox/v1/legacy/common/alloc"
	"github.com/opennox/opennox/v1/server"
)

func TestSimplePolypDeathNativeLayout54CAE0(t *testing.T) {
	o := server.Object{}
	want := [5]uintptr{unsafe.Offsetof(o.ObjFlags), unsafe.Offsetof(o.PosVec), unsafe.Offsetof(o.PosVec) + 4, unsafe.Offsetof(o.UpdateData), unsafe.Offsetof(o.ScriptPickup) + unsafe.Sizeof(o.ScriptPickup)}
	if got := simplePolypOriginalLayout54CAE0(); got != want {
		t.Fatal("independent complete native C/Go layout", got, want)
	}
}

// The production AUD reader initializes real sound descriptors. The source
// then traverses the real typed C audio entry and native object event queue.
type simplePolypAUDFile54CAE0 struct{ *bytes.Reader }

func (f simplePolypAUDFile54CAE0) Skip(n int) { _, _ = f.Seek(int64(n), io.SeekCurrent) }
func (f simplePolypAUDFile54CAE0) ReadU8() uint8 {
	v, err := f.ReadByte()
	if err != nil {
		panic(err)
	}
	return v
}
func (f simplePolypAUDFile54CAE0) ReadI8() int8 { return int8(f.ReadU8()) }
func (f simplePolypAUDFile54CAE0) ReadU16() uint16 {
	return uint16(f.ReadU8()) | uint16(f.ReadU8())<<8
}
func (f simplePolypAUDFile54CAE0) ReadI16() int16 { return int16(f.ReadU16()) }
func (f simplePolypAUDFile54CAE0) ReadU32() uint32 {
	return uint32(f.ReadU16()) | uint32(f.ReadU16())<<16
}
func (f simplePolypAUDFile54CAE0) ReadI32() int32 { return int32(f.ReadU32()) }
func (f simplePolypAUDFile54CAE0) ReadString8() (string, error) {
	b := make([]byte, int(f.ReadU8()))
	_, err := io.ReadFull(f, b)
	return string(b), err
}

func simplePolypActualFixture54CAE0(t *testing.T) (*server.Server, *server.Object, uint32) {
	t.Helper()
	TestSimplePolypDeathNativeLayout54CAE0(t)
	s := npcReflectLegacyServer4E17B0(t)
	s.Audio.Init(s)
	t.Cleanup(s.Audio.Free)
	s.FreeObjectTypes()
	if !s.Objs.Init(8) {
		t.Fatal("native object pool")
	}
	t.Cleanup(s.Objs.FreeObjects)
	t.Cleanup(s.FreeObjectTypes)
	if err := s.Types.ReadObjectType(&things.Thing{Name: "ToxicCloud", OnUpdate: &things.ProcFunc{Name: "ToxicCloudUpdate"}}); err != nil {
		t.Fatal(err)
	}
	id := simplePolypOriginalType54CAE0()
	if id == 0 || s.Types.ByID("ToxicCloud").Ind() != int(id) {
		t.Fatal("real C name lookup")
	}
	cloud := simplePolypOriginalNew54CAE0(id)
	if cloud == nil || cloud.UpdateData == nil || cloud.Update == nil {
		t.Fatal("actual C constructor and registered native update record")
	}
	t.Cleanup(func() { s.Objs.FreeObject(cloud) })
	cloud.PosVec, cloud.ObjFlags = types.Ptf(7, 8), 0x10203040
	data := (*server.ToxicCloudUpdateData)(cloud.UpdateData)
	data.Duration = -123
	frame := bytes.Clone(unsafe.Slice((*byte)(unsafe.Pointer(cloud)), int(unsafe.Sizeof(*cloud))))
	update := bytes.Clone(unsafe.Slice((*byte)(cloud.UpdateData), int(unsafe.Sizeof(*data))))
	p, w := simplePolypOriginalControl54CAE0(cloud)
	if p != [4]uintptr{uintptr(unsafe.Pointer(cloud)), uintptr(cloud.UpdateData), uintptr(unsafe.Pointer(&cloud.PosVec.X)), uintptr(unsafe.Pointer(&cloud.PosVec.Y))} || w != [4]uint32{0x10203040, math.Float32bits(7), math.Float32bits(8), uint32(0xffffff85)} {
		t.Fatal("actual C constructor/control native pointers and scalars", p, w)
	}
	if !bytes.Equal(frame, unsafe.Slice((*byte)(unsafe.Pointer(cloud)), int(unsafe.Sizeof(*cloud)))) || !bytes.Equal(update, unsafe.Slice((*byte)(cloud.UpdateData), int(unsafe.Sizeof(*data)))) {
		t.Fatal("whole constructor/control records")
	}
	unit, free := alloc.NewHigh(server.Object{})
	t.Cleanup(free)
	unit.PosVec, unit.ObjFlags = types.Ptf(300, 300), 0x11223300
	if unsafe.Sizeof(uintptr(0)) == 8 {
		for _, p := range []unsafe.Pointer{unsafe.Pointer(unit), unsafe.Pointer(cloud), cloud.UpdateData} {
			if uintptr(p) <= math.MaxUint32 {
				t.Fatal("source, real C object and update must exceed4GiB", p)
			}
		}
	}
	ids := []sound.ID{764, 284, 753}
	wire := binary.LittleEndian.AppendUint32(nil, uint32(len(ids)))
	for _, id := range ids {
		name := id.String()
		wire = append(wire, byte(len(name)))
		wire = append(wire, name...)
		wire = append(wire, 0, 0, 0, 40, 0, 1, 0, 0, 0, 1, 'x', 0)
	}
	f := simplePolypAUDFile54CAE0{bytes.NewReader(wire)}
	if !s.Audio.ReadAUD(f) || f.Len() != 0 {
		t.Fatal("actual sound descriptor reader")
	}
	s.Audio.Reset()
	before := bytes.Clone(unsafe.Slice((*byte)(unsafe.Pointer(unit)), int(unsafe.Sizeof(*unit))))
	for i, id := range ids {
		if s.Audio.Field12(id) != 1 {
			t.Fatal("actual enabled sound descriptor", id)
		}
		simplePolypOriginalAudio54CAE0(uint32(id), unit)
		count := 0
		s.Audio.EachEvent(func(event *server.AudioEvent) {
			count++
			if event.Obj != unit || event.Pos != unit.PosVec || event.Kind != 0 || event.Code != 0 {
				t.Fatal("actual typed C audio/native event pointer and values", event)
			}
		})
		if count != i+1 {
			t.Fatal("actual native audio event count", count)
		}
	}
	if !bytes.Equal(before, unsafe.Slice((*byte)(unsafe.Pointer(unit)), int(unsafe.Sizeof(*unit)))) {
		t.Fatal("whole source record changed by typed C audio controls")
	}
	return s, unit, id
}

func TestSimplePolypDeathActualCBefore54CAE0(t *testing.T) {
	if os.Getenv("OPENNOX_SIMPLE_POLYP_C_BEFORE") != "1" {
		return
	}
	if entry := os.Getenv("OPENNOX_SIMPLE_POLYP_C_CHILD"); entry != "" {
		_, unit, id := simplePolypActualFixture54CAE0(t)
		index, err := strconv.Atoi(entry)
		if err != nil || index < 0 || index > 2 || unsafe.Sizeof(uintptr(0)) != 8 || simplePolypOriginalABI54CAE0(index) != 32 {
			t.Fatal("actual unchanged old signed32 C entry required")
		}
		cache := memmap.PtrUint32(0x5D4594, 2491672)
		old := *cache
		*cache = id
		defer func() { *cache = old }()
		t.Logf("SIMPLE_POLYP_REAL_C_CONTROLS entry=%s source=%p type=%d complete-layout-real-C-constructor-update-whole-records-and-three-real-C-audio-events=true", entry, unit, id)
		simplePolypOriginalC54CAE0(index, unit)
		t.Fatal("old signed32 C entry unexpectedly returned")
	}
	t.Run("real-constructor-and-audio-controls", func(t *testing.T) { simplePolypActualFixture54CAE0(t) })
	for index, name := range []string{"ImpEgg", "Polyp", "Potion"} {
		t.Run(name, func(t *testing.T) {
			ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
			defer cancel()
			cmd := exec.CommandContext(ctx, os.Args[0], "-test.run=^TestSimplePolypDeathActualCBefore54CAE0$", "-test.v")
			cmd.Env = append(os.Environ(), "OPENNOX_SIMPLE_POLYP_C_CHILD="+strconv.Itoa(index))
			out, err := cmd.CombinedOutput()
			t.Logf("actual old C %s result: %v\n%s", name, err, out)
			if ctx.Err() != nil || err == nil || !bytes.Contains(out, []byte("SIMPLE_POLYP_REAL_C_CONTROLS")) || bytes.Contains(out, []byte("unexpectedly returned")) {
				t.Fatal("valid native positive controls before actual old C fault", err)
			}
			if !bytes.Contains(out, []byte("SIGSEGV")) && !bytes.Contains(out, []byte("SIGBUS")) && !bytes.Contains(out, []byte("unexpected fault address")) {
				t.Fatal("actual native pointer fault required")
			}
		})
	}
}

func TestSimplePolypDeathFrozenReferences54CAE0(t *testing.T) {
	var inputs []struct{ Name string }
	trollDeathFixture54A270(t, "simple_polyp_death_independent_inputs_54cae0.json.gz", "54334f7371c06bc3f046293acc31e87004e286bb7293600b3d1819245d705d26", "b1875ee4b7a1e356ddfa8dd915919278c2e3597565b315a453c0c8713c19efc5", &inputs)
	if len(inputs) != 149 {
		t.Fatal("independent input count")
	}
	for _, profile := range []struct{ name, encoded, decoded string }{
		{"tozero", "0cc6f398fbbf2b9a3671eb2071ecf3066dd9dfbddd2000322999a3f772af46eb", "8f839d3732dd1ca957b97845f9841ce290742f495b7a8c39737a901c4db451e8"},
		{"nearest", "14427ed8cd5257fd981a0b3a191a036cd9eb50182e71422d986614e64855d8e0", "1ac2b826570efce639c37228620946ad6606cbf9fda4dea09fafc00e0b694f11"},
	} {
		t.Run(profile.name, func(t *testing.T) {
			var outputs []json.RawMessage
			trollDeathFixture54A270(t, "simple_polyp_death_native_original_54cae0_pc53_"+profile.name+".json.gz", profile.encoded, profile.decoded, &outputs)
			if len(outputs) != len(inputs) {
				t.Fatal("complete original output count")
			}
			for i, output := range outputs {
				var row struct{ Name string }
				if err := json.Unmarshal(output, &row); err != nil || row.Name != inputs[i].Name {
					t.Fatal("frozen original input/output identity", i, err)
				}
			}
		})
	}
}
