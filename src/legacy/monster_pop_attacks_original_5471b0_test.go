package legacy

import (
	"bytes"
	"compress/gzip"
	"crypto/sha256"
	"encoding/binary"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"math"
	"os"
	"os/exec"
	"reflect"
	"strings"
	"testing"
	"unsafe"

	"github.com/opennox/libs/object"
	"github.com/opennox/opennox/v1/legacy/common/alloc"
	"github.com/opennox/opennox/v1/server"
)

type popData5471B0 struct {
	Index   int8
	Prefix  uint32
	Actions [24]uint32
}
type popState5471B0 struct {
	Update int `json:"unit_update"`
	Data   [2]popData5471B0
}
type popStep5471B0 struct {
	Data   int
	Index  int8
	Slot   *int
	Action uint32
	Prefix *uint32
	Update *int `json:"unit_update"`
}
type popInput5471B0 struct {
	Name   string
	Update int  `json:"unit_update"`
	Nil    bool `json:"nil_unit"`
	Data   [2]popData5471B0
	Steps  []popStep5471B0 `json:"pop_steps"`
}
type popEvent5471B0 struct {
	Service string
	Unit    bool
	Before  popState5471B0
}
type popWrite5471B0 struct {
	Offset int
	Bytes  string
}
type popRow5471B0 struct {
	Input  popInput5471B0
	Output struct {
		Events []popEvent5471B0
		Writes []popWrite5471B0
		State  popState5471B0
		Fault  []json.RawMessage
		Error  any
	}
}

func popOriginalRows5471B0(t *testing.T) []popRow5471B0 {
	t.Helper()
	raw, e := os.ReadFile("../server/testdata/monster_pop_attacks_original_5471b0.json.gz")
	if e != nil {
		t.Fatal(e)
	}
	if fmt.Sprintf("%x", sha256.Sum256(raw)) != "f745b2204becc35cb9b8657e68c808478b91b5073d095d1cf76ff1cda81ab153" {
		t.Fatal("sealed original gzip changed")
	}
	z, e := gzip.NewReader(bytes.NewReader(raw))
	if e != nil {
		t.Fatal(e)
	}
	decoded, e := io.ReadAll(z)
	if e != nil {
		t.Fatal(e)
	}
	if e = z.Close(); e != nil {
		t.Fatal(e)
	}
	if fmt.Sprintf("%x", sha256.Sum256(decoded)) != "39192bd5c6fae7ff67914fd4adbe3ed60249f5b90afc7860e7ef951408ac9506" {
		t.Fatal("original decoded digest")
	}
	var d struct{ Rows []popRow5471B0 }
	if e = json.Unmarshal(decoded, &d); e != nil {
		t.Fatal(e)
	}
	if len(d.Rows) != 3270 {
		t.Fatal("original row count")
	}
	return d.Rows
}

type popWorld5471B0 struct {
	Unit   *server.Object
	Data   [2]*server.MonsterUpdateData
	Heap   []byte
	Events []popEvent5471B0
	Writes []popWrite5471B0
	Calls  int
}

func newPopWorld5471B0(t *testing.T, row popRow5471B0) *popWorld5471B0 {
	t.Helper()
	ptr, free := alloc.MallocHigh(32768)
	t.Cleanup(free)
	heap := unsafe.Slice((*byte)(ptr), 32768)
	for i := range heap {
		heap[i] = byte(i*37 + 0xa5)
	}
	w := &popWorld5471B0{Unit: (*server.Object)(unsafe.Add(ptr, 1024)), Data: [2]*server.MonsterUpdateData{(*server.MonsterUpdateData)(unsafe.Add(ptr, 4096)), (*server.MonsterUpdateData)(unsafe.Add(ptr, 12288))}, Heap: heap, Events: make([]popEvent5471B0, 0), Writes: make([]popWrite5471B0, 0)}
	w.Unit.ObjClass = object.ClassMonster
	if unsafe.Sizeof(uintptr(0)) == 4 {
		// The pending pointer store replaces arbitrary canary bytes. Clear
		// only that cell through C before Go's barrier can shade the old
		// bytes as a heap pointer; the completed native frame is unchanged.
		alloc.Memset(unsafe.Pointer(&w.Unit.UpdateData), 0, unsafe.Sizeof(w.Unit.UpdateData))
	}
	w.Unit.UpdateData = w.pointer(row.Input.Update)
	for n, d := range w.Data {
		d.AIStackInd = row.Input.Data[n].Index
		d.Field132 = row.Input.Data[n].Prefix
		for i := range d.AIStack {
			d.AIStack[i].Action = row.Input.Data[n].Actions[i]
			for j := range d.AIStack[i].Args {
				v := uintptr(0xa5a5a5a5)
				if unsafe.Sizeof(v) == 8 {
					high := uint64(0x12345678)
					v |= uintptr(high << 32)
				}
				d.AIStack[i].Args[j] = v
			}
		}
	}
	if unsafe.Sizeof(uintptr(0)) == 8 {
		for _, v := range []unsafe.Pointer{unsafe.Pointer(w.Unit), unsafe.Pointer(w.Data[0]), unsafe.Pointer(w.Data[1])} {
			if uintptr(v) <= math.MaxUint32 {
				t.Fatal("native C-owned address must exceed4GiB")
			}
		}
	}
	return w
}
func (w *popWorld5471B0) pointer(role int) unsafe.Pointer {
	if role < 0 {
		return nil
	}
	return unsafe.Pointer(w.Data[role])
}
func (w *popWorld5471B0) snapshot() popState5471B0 {
	s := popState5471B0{Update: -1}
	if w.Unit.UpdateData != nil {
		switch w.Unit.UpdateData {
		case unsafe.Pointer(w.Data[0]):
			s.Update = 0
		case unsafe.Pointer(w.Data[1]):
			s.Update = 1
		default:
			panic("invalid native pointer projection")
		}
	}
	for n, d := range w.Data {
		s.Data[n].Index = d.AIStackInd
		s.Data[n].Prefix = d.Field132
		for i := range d.AIStack {
			s.Data[n].Actions[i] = d.AIStack[i].Action
		}
	}
	return s
}
func (w *popWorld5471B0) service(t *testing.T, row popRow5471B0, obj *server.Object) int {
	t.Helper()
	if obj != w.Unit || w.Calls >= len(row.Input.Steps) {
		t.Fatal("unexpected Pop service identity/order")
	}
	w.Events = append(w.Events, popEvent5471B0{"Pop50A160", true, w.snapshot()})
	step := row.Input.Steps[w.Calls]
	w.Calls++
	d := w.Data[step.Data]
	base := 0x3000 + step.Data*0x2000
	write := func(offset int, value uint32, size int) {
		buf := make([]byte, 4)
		binary.LittleEndian.PutUint32(buf, value)
		w.Writes = append(w.Writes, popWrite5471B0{offset, hex.EncodeToString(buf[:size])})
	}
	d.AIStackInd = step.Index
	write(base+544, uint32(uint8(step.Index)), 1)
	if step.Slot != nil {
		d.AIStack[*step.Slot].Action = step.Action
		write(base+552+24**step.Slot, step.Action, 4)
	}
	if step.Prefix != nil {
		d.Field132 = *step.Prefix
		write(base+528, *step.Prefix, 4)
	}
	if step.Update != nil {
		w.Unit.UpdateData = w.pointer(*step.Update)
		v := uint32(0)
		if *step.Update >= 0 {
			v = 0x21003000 + uint32(*step.Update)*0x2000
		}
		write(0x1000+748, v, 4)
	}
	value := uint32(0xa5a55a5a)
	return int(value)
}

// Project original writes independently of the service-input adapter. Compare
// the whole native allocation, including padding, neighbors and high Args.
func (w *popWorld5471B0) expected(t *testing.T, row popRow5471B0) []byte {
	t.Helper()
	out := bytes.Clone(w.Heap)
	for _, wr := range row.Output.Writes {
		b, e := hex.DecodeString(wr.Bytes)
		if e != nil {
			t.Fatal(e)
		}
		var address unsafe.Pointer
		var native []byte
		if wr.Offset == 0x1000+748 {
			v := binary.LittleEndian.Uint32(b)
			role := -1
			if v != 0 {
				role = int((v - 0x21003000) / 0x2000)
				if role < 0 || role > 1 {
					t.Fatal("original pointer role")
				}
			}
			value := uintptr(w.pointer(role))
			native = make([]byte, unsafe.Sizeof(value))
			if len(native) == 8 {
				binary.LittleEndian.PutUint64(native, uint64(value))
			} else {
				binary.LittleEndian.PutUint32(native, uint32(value))
			}
			address = unsafe.Pointer(&w.Unit.UpdateData)
		} else {
			n := (wr.Offset - 0x3000) / 0x2000
			if n < 0 || n > 1 {
				t.Fatal("original data-write role")
			}
			off := wr.Offset - (0x3000 + n*0x2000)
			d := w.Data[n]
			switch {
			case off == 544:
				address = unsafe.Pointer(&d.AIStackInd)
			case off == 528:
				address = unsafe.Pointer(&d.Field132)
			case off >= 552 && (off-552)%24 == 0 && (off-552)/24 < 24:
				address = unsafe.Pointer(&d.AIStack[(off-552)/24].Action)
			default:
				t.Fatal("unclassified original write", wr)
			}
			native = b
		}
		off := uintptr(address) - uintptr(unsafe.Pointer(&w.Heap[0]))
		copy(out[int(off):int(off)+len(native)], native)
	}
	return out
}
func TestMonsterPopAttacksNativeLayout5471B0(t *testing.T) {
	// C's shared prefix ends before the two Go-only EXT uintptr handles.
	end := unsafe.Offsetof(server.Object{}.ScriptPickup) + unsafe.Sizeof(server.Object{}.ScriptPickup)
	first, _ := reflect.TypeOf(server.Object{}).FieldByName("serverHandle")
	second, _ := reflect.TypeOf(server.Object{}).FieldByName("objectHandle")
	if first.Offset != end || second.Offset != end+unsafe.Sizeof(uintptr(0)) || unsafe.Sizeof(server.Object{}) != end+2*unsafe.Sizeof(uintptr(0)) {
		t.Fatal("C prefix / two Go EXT handles")
	}
	got := popOriginalLayout5471B0()
	want := [8]uintptr{unsafe.Offsetof(server.Object{}.ScriptPickup) + unsafe.Sizeof(server.Object{}.ScriptPickup), unsafe.Offsetof(server.Object{}.UpdateData), unsafe.Offsetof(server.MonsterUpdateData{}.Field132), unsafe.Offsetof(server.MonsterUpdateData{}.AIStackInd), unsafe.Offsetof(server.MonsterUpdateData{}.AIStack), unsafe.Sizeof(server.AIStackItem{}), unsafe.Offsetof(server.AIStackItem{}.Args), unsafe.Offsetof(server.AIStackItem{}.Field5)}
	if got != want {
		t.Fatal("independent C/Go native layout", got, want)
	}
}

type popEntry5471B0 interface {
	MonsterPopAttackActionsNative5471B0(func(*server.Object) int)
}

func TestMonsterPopAttacksOriginal5471B0(t *testing.T) {
	rows := popOriginalRows5471B0(t)
	for _, path := range []string{"C", "Go"} {
		t.Run(path, func(t *testing.T) {
			for _, row := range rows {
				t.Run(row.Input.Name, func(t *testing.T) {
					w := newPopWorld5471B0(t, row)
					want := w.expected(t, row)
					unit := w.Unit
					if row.Input.Nil {
						unit = nil
					}
					service := func(obj *server.Object) int { return w.service(t, row, obj) }
					var fault any
					func() {
						defer func() { fault = recover() }()
						if path == "C" {
							old := monsterPopAttacksOverride5471B0
							monsterPopAttacksOverride5471B0 = service
							defer func() { monsterPopAttacksOverride5471B0 = old }()
							popOriginalC5471B0(unit)
						} else {
							entry, ok := any(unit).(popEntry5471B0)
							if !ok {
								t.Fatal("native raw entry not implemented")
							}
							entry.MonsterPopAttackActionsNative5471B0(service)
						}
					}()
					if (fault != nil) != (len(row.Output.Fault) > 0) {
						t.Fatal("raw original fault prefix", fault, row.Output.Fault)
					}
					if !reflect.DeepEqual(w.Events, row.Output.Events) || !reflect.DeepEqual(w.Writes, row.Output.Writes) || w.snapshot() != row.Output.State {
						t.Fatal("original ordered services/writes/state", w.Events, row.Output.Events, w.Writes, row.Output.Writes, w.snapshot(), row.Output.State)
					}
					if !bytes.Equal(w.Heap, want) {
						for i := range want {
							if want[i] != w.Heap[i] {
								t.Fatalf("whole native allocation differs byte%d got%02x want%02x", i, w.Heap[i], want[i])
							}
						}
					}
				})
			}
		})
	}
}
func TestMonsterPopAttacksNativeCBefore5471B0(t *testing.T) {
	if os.Getenv("OPENNOX_POP_ATTACKS_C_BEFORE") != "1" {
		return
	}
	if os.Getenv("OPENNOX_POP_ATTACKS_C_CHILD") != "1" {
		cmd := exec.Command(os.Args[0], "-test.run=^TestMonsterPopAttacksNativeCBefore5471B0$", "-test.v")
		cmd.Env = append(os.Environ(), "OPENNOX_POP_ATTACKS_C_CHILD=1")
		out, err := cmd.CombinedOutput()
		t.Logf("%s", out)
		if err == nil || !strings.Contains(string(out), "POP_C_BEFORE_VALID_NATIVE_CONTROLS") || !strings.Contains(string(out), "SIGSEGV") || !strings.Contains(string(out), "_Cfunc_pop_original_public_5471b0") {
			t.Fatal("actual legacy32 public C failure after valid native controls", err)
		}
		return
	}
	TestMonsterPopAttacksNativeLayout5471B0(t)
	row := popOriginalRows5471B0(t)[95]
	w := newPopWorld5471B0(t, row)
	w.Data[0].AIStackInd = 0
	w.Data[0].AIStack[0].Action = 0
	w.Data[0].SoundSet122 = unsafe.Pointer(&w.Data[1].Field128)
	before := bytes.Clone(w.Heap)
	if popOriginalControl5471B0(w.Unit) != uintptr(w.Data[0].SoundSet122) || popOriginalAction5471B0(w.Unit) != 0 || !bytes.Equal(before, w.Heap) {
		t.Fatal("actual typed getter/full record controls")
	}
	if popOriginalABI5471B0() != 32 {
		t.Fatal("requires unchanged legacy int ABI")
	}
	t.Logf("POP_C_BEFORE_VALID_NATIVE_CONTROLS unit=%p update=%p publicABI=legacy32-explicit-narrowing typed-getters-native=true", w.Unit, w.Data[0])
	popOriginalC5471B0(w.Unit)
	t.Fatal("unchanged C unexpectedly returned")
}
