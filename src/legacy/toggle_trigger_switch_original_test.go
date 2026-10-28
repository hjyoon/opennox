package legacy

import (
	"encoding/json"
	"os"
	"reflect"
	"strings"
	"testing"
	"unsafe"

	"bytes"
	"encoding/binary"
	"github.com/opennox/libs/object"
	"github.com/opennox/opennox/v1/server"
)

type ttsState struct {
	Type    uint32        `json:"type"`
	Flags   uint32        `json:"flags"`
	Anim    uint32        `json:"anim"`
	Next    uint32        `json:"next"`
	Sync    uint32        `json:"sync"`
	Update  int           `json:"update_ref"`
	Collide int           `json:"collide_ref"`
	Records [2][15]uint32 `json:"records"`
	Callers [2]int        `json:"callers"`
	Frame   uint32        `json:"frame"`
	FPS     uint32        `json:"fps"`
	Caches  [2]uint32     `json:"caches"`
}
type ttsInput struct {
	Name     string    `json:"name"`
	Kind     int       `json:"kind"`
	Lookup   [2]uint32 `json:"lookup"`
	Queue    uint8     `json:"queue"`
	Mutation string    `json:"mutation"`
	ttsState
}
type ttsEvent struct {
	Service string  `json:"service"`
	Args    []int64 `json:"args"`
	ttsState
}
type ttsRow struct {
	Input     ttsInput   `json:"input"`
	Output    ttsState   `json:"output"`
	Events    []ttsEvent `json:"events"`
	Result    *uint8     `json:"result"`
	Source    int        `json:"source_ref"`
	Fault     bool       `json:"fault"`
	Preserved bool       `json:"nonoutput_bytes_preserved"`
}

func ttsRun(t *testing.T, in ttsInput, sourceRef int, want ttsState) (ttsState, []ttsEvent, uint8, any) {
	t.Helper()
	var objs [3]*server.Object
	var data [2]*server.TriggerUpdateData
	var callbacks [2]unsafe.Pointer
	var frames, before [7][]byte
	for n := range objs {
		frames[n] = radialCFrameOriginal(t, unsafe.Sizeof(server.Object{}))
		objs[n] = (*server.Object)(unsafe.Pointer(&frames[n][32]))
		*objs[n] = server.Object{}
	}
	for n := range data {
		frames[3+n] = radialCFrameOriginal(t, 60)
		data[n] = (*server.TriggerUpdateData)(unsafe.Pointer(&frames[3+n][32]))
		frames[5+n] = radialCFrameOriginal(t, 1)
		callbacks[n] = unsafe.Pointer(&frames[5+n][32])
	}

	ref := func(n int) *server.Object {
		if n < 0 {
			return nil
		}
		return objs[n]
	}
	record := func(n int) unsafe.Pointer {
		if n < 0 {
			return nil
		}
		return unsafe.Pointer(data[n])
	}
	callback := func(n int) unsafe.Pointer {
		if n < 0 {
			return nil
		}
		return callbacks[n]
	}
	id := func(o *server.Object) int {
		if o == nil {
			return -1
		}
		for n, p := range objs {
			if p == o {
				return n
			}
		}
		t.Fatal("unknown object")
		return -2
	}
	dataID := func(p unsafe.Pointer) int {
		if p == nil {
			return -1
		}
		for n, d := range data {
			if p == unsafe.Pointer(d) {
				return n
			}
		}
		t.Fatal("unknown record")
		return -2
	}
	fnID := func(p unsafe.Pointer) int {
		if p == nil {
			return -1
		}
		for n, d := range callbacks {
			if p == d {
				return n
			}
		}
		t.Fatal("unknown callback")
		return -2
	}
	o := objs[0]
	*o = server.Object{TypeInd: uint16(in.Type), ObjFlags: object.Flags(in.Flags), Field33: in.Anim, Field34: in.Next, Field38: in.Sync, UpdateData: record(in.Update), Collide: callback(in.Collide)}
	for n := range data {
		*(*[15]uint32)(unsafe.Pointer(data[n])) = in.Records[n]
		data[n].Field4 = 0xabbaabba
	}
	for n := range frames {
		before[n] = append([]byte(nil), frames[n]...)
	}
	snapshot := func() ttsState {
		s := ttsState{Type: uint32(o.TypeInd), Flags: uint32(o.ObjFlags), Anim: o.Field33, Next: o.Field34, Sync: o.Field38, Update: dataID(o.UpdateData), Collide: fnID(o.Collide), Callers: in.Callers, Frame: in.Frame, FPS: in.FPS, Caches: in.Caches}
		for n := range data {
			s.Records[n] = *(*[15]uint32)(unsafe.Pointer(data[n]))
			s.Records[n][1] = 0
		}
		return s
	}
	events := make([]ttsEvent, 0)
	event := func(name string, args ...int64) {
		events = append(events, ttsEvent{Service: name, Args: args, ttsState: snapshot()})
	}
	applied := false
	mutate := func(service string) {
		prefix := service + ":"
		if applied || !strings.HasPrefix(in.Mutation, prefix) {
			return
		}
		applied = true
		switch strings.TrimPrefix(in.Mutation, prefix) {
		case "clock":
			in.Frame = 0xfffffffc
		case "fps":
			in.FPS = 0x80000001
		case "flags":
			o.ObjFlags = 0x1000041
		case "anim":
			o.Field33 = 99
		case "deadline":
			o.Field34 = 0xfffffffe
		case "update":
			o.UpdateData = record(1)
		case "caller":
			in.Callers[0] = 2
		case "caller-nil":
			in.Callers[0] = -1
		case "request":
			data[0].Flags ^= 1
		case "one-shot":
			data[0].Flags |= 2
		case "mode":
			data[0].State = 1
		case "sound":
			data[0].SoundActivate = 99
			data[0].SoundDeactivate = 100
		case "script":
			*(*uint32)(unsafe.Pointer(&data[0].ScriptActivate)) = 0x77777777
			*(*uint32)(unsafe.Pointer(&data[0].ScriptDeactivate)) = 0x88888888
		case "cache":
			in.Caches = [2]uint32{99, 100}
		case "type":
			o.TypeInd = 18
		case "collide":
			o.Collide = nil
		default:
			t.Fatal("unknown mutation")
		}
	}
	h := server.ToggleTriggerRawRuntime53B060{
		Frame: func() uint32 { return in.Frame }, FPS: func() uint32 { return in.FPS },
		MarkAnimFrame: func(unit *server.Object, v uint32) {
			if unit != o {
				t.Fatal("animation object")
			}
			event("anim", 0, int64(v))
			o.Field33 = v
			o.Field38 = 0xffffffff
			mutate("anim")
		},
		AudioEvent: func(v uint32, unit *server.Object) {
			if unit != o {
				t.Fatal("audio object")
			}
			event("audio", int64(v), 0, 0, 0)
			mutate("audio")
		},
		CollideTarget: func(unit *server.Object, d *server.TriggerUpdateData) *server.Object {
			if unit != o {
				t.Fatal("caller source")
			}
			n := dataID(unsafe.Pointer(d))
			return ref(in.Callers[n])
		},
		ScriptCallback: func(b *server.ScriptCallback, caller, trigger *server.Object, ev server.ScriptEventType) {
			if trigger != o {
				t.Fatal("script source")
			}
			recordID, off := -1, uintptr(0)
			for n, d := range data {
				for _, delta := range []uintptr{20, 28} {
					if unsafe.Pointer(b) == unsafe.Add(unsafe.Pointer(d), delta) {
						recordID = n
						off = delta
					}
				}
			}
			if recordID < 0 {
				t.Fatal("script record identity")
			}
			wantEvent := int32(8)
			if in.Kind == 0 {
				wantEvent = 10
			}
			if off == 28 {
				wantEvent++
			}
			if int32(ev) != wantEvent {
				t.Fatal("port event metadata", ev, wantEvent)
			}
			event("script", int64(recordID), int64(off), int64(id(caller)), 0)
			mutate("script")
		},
	}
	tc := server.TriggerTypeCacheRuntime53B1B0{
		Load: func(n int) uint32 { return in.Caches[n] }, Store: func(n int, v uint32) { in.Caches[n] = v },
		Lookup: func(name string) uint32 {
			n := 0
			if name == "PressurePlate" {
				n = 1
			} else if name != "Trigger" {
				t.Fatal(name)
			}
			event("lookup", int64(n))
			mutate("lookup")
			return in.Lookup[n]
		},
	}
	immediate := func(unit *server.Object) bool { return server.TriggerImmediateTypeRaw53B1B0(unit, tc) }
	s := server.SwitchUpdateRawRuntime53B320{
		NeedSync: func(unit *server.Object) {
			if unit != o {
				t.Fatal("sync object")
			}
			event("sync", 0)
			o.Field38 = 0xffffffff
			mutate("sync")
		},
		QueueCollision: func(unit *server.Object) uint8 {
			if unit != o {
				t.Fatal("queue object")
			}
			event("queue", 0)
			mutate("queue")
			return in.Queue
		},
	}
	var result uint8
	var recovered any
	func() {
		defer func() { recovered = recover() }()
		oldH, oldI, oldS := toggleTriggerRawRuntime53B060, triggerImmediateType53B1B0, switchRawRuntime53B320
		defer func() {
			toggleTriggerRawRuntime53B060 = oldH
			triggerImmediateType53B1B0 = oldI
			switchRawRuntime53B320 = oldS
		}()
		toggleTriggerRawRuntime53B060 = func() server.ToggleTriggerRawRuntime53B060 { return h }
		triggerImmediateType53B1B0 = immediate
		switchRawRuntime53B320 = func() server.SwitchUpdateRawRuntime53B320 { return s }
		result = toggleTriggerSwitchProbe(in.Kind, ref(sourceRef))
	}()
	putptr := func(b []byte, off uintptr, p unsafe.Pointer) {
		if unsafe.Sizeof(uintptr(0)) == 8 {
			binary.LittleEndian.PutUint64(b[32+off:], uint64(uintptr(p)))
		} else {
			binary.LittleEndian.PutUint32(b[32+off:], uint32(uintptr(p)))
		}
	}
	expected := [7][]byte{}
	for n := range expected {
		expected[n] = append([]byte(nil), before[n]...)
	}
	binary.LittleEndian.PutUint16(expected[0][32+unsafe.Offsetof(o.TypeInd):], uint16(want.Type))
	for n, off := range []uintptr{unsafe.Offsetof(o.ObjFlags), unsafe.Offsetof(o.Field33), unsafe.Offsetof(o.Field34), unsafe.Offsetof(o.Field38)} {
		binary.LittleEndian.PutUint32(expected[0][32+off:], []uint32{want.Flags, want.Anim, want.Next, want.Sync}[n])
	}
	putptr(expected[0], unsafe.Offsetof(o.UpdateData), record(want.Update))
	putptr(expected[0], unsafe.Offsetof(o.Collide), callback(want.Collide))
	for n := range data {
		for k, v := range want.Records[n] {
			if k != 1 {
				binary.LittleEndian.PutUint32(expected[3+n][32+4*k:], v)
			}
		}
	}
	for n := range frames {
		if !bytes.Equal(expected[n], frames[n]) {
			t.Fatalf("native object/record/callback frame/canary %d", n)
		}
	}

	return snapshot(), events, result, recovered
}
func TestToggleTriggerSwitchOriginalNativeC(t *testing.T) {
	blob, err := os.ReadFile("../server/testdata/toggle_trigger_switch_original.json")
	if err != nil {
		t.Fatal(err)
	}
	var f struct {
		SHA      string   `json:"original_sha256"`
		CW       uint32   `json:"original_startup_FPCW"`
		Contexts int      `json:"contexts"`
		Rows     []ttsRow `json:"rows"`
	}
	if err = json.Unmarshal(blob, &f); err != nil {
		t.Fatal(err)
	}
	if f.SHA != "0040e2c0683b4d73a5fb976e400d5087dca680df2b195c9e27f8edbda2d4974a" || f.CW != 0xe3f || f.Contexts != 1152 || len(f.Rows) != 1152 {
		t.Fatal("original provenance")
	}
	for _, row := range f.Rows {
		t.Run(row.Input.Name, func(t *testing.T) {
			out, events, result, fault := ttsRun(t, row.Input, 0, row.Output)
			if !row.Preserved || fault != nil || row.Result == nil {
				t.Fatal("unexpected fault/provenance", fault)
			}
			if result != *row.Result || !reflect.DeepEqual(out, row.Output) {
				t.Errorf("return/state differs: result %02x want %02x state %+v want %+v", result, *row.Result, out, row.Output)
			}
			if !reflect.DeepEqual(events, row.Events) {
				t.Errorf("original service inputs/order/state differ: got %+v want %+v", events, row.Events)
			}
		})
	}
}
func TestToggleTriggerSwitchOriginalFaultsNativeC(t *testing.T) {
	blob, err := os.ReadFile("../server/testdata/toggle_trigger_switch_faults_original.json")
	if err != nil {
		t.Fatal(err)
	}
	var f struct {
		Rows []ttsRow `json:"rows"`
	}
	if err = json.Unmarshal(blob, &f); err != nil {
		t.Fatal(err)
	}
	if len(f.Rows) != 12 {
		t.Fatal("fault provenance")
	}
	for _, row := range f.Rows {
		t.Run(row.Input.Name, func(t *testing.T) {
			out, events, result, fault := ttsRun(t, row.Input, row.Source, row.Output)
			if (fault != nil) != row.Fault || !reflect.DeepEqual(out, row.Output) || !reflect.DeepEqual(events, row.Events) {
				t.Fatal("original fault/gate/prefix/state differs", fault)
			}
			if !row.Fault && (row.Result == nil || result != *row.Result) {
				t.Fatal("safe gate return byte")
			}
		})
	}
}
