package server

import (
	"encoding/json"
	"math"
	"os"
	"reflect"
	"strings"
	"testing"
	"unsafe"

	"github.com/opennox/libs/types"
)

type fxState struct {
	Objects     [3][8]uint32 `json:"objects"`
	Direction   uint16       `json:"direction"`
	Update      int          `json:"update_ref"`
	Data        [2][4]uint32 `json:"data"`
	RandomIndex int          `json:"random_index"`
	Created     int          `json:"created"`
}
type fxInput struct {
	Name           string     `json:"name"`
	Kind           int        `json:"kind"`
	DirectionWords [2]uint32  `json:"direction_words"`
	Random         [32]uint64 `json:"random"`
	Returns        [8]int     `json:"returns"`
	Mutation       string     `json:"mutation"`
	At             int        `json:"at"`
	fxState
}
type fxEvent struct {
	Service string  `json:"service"`
	Args    []int64 `json:"args"`
	fxState
}
type fxRow struct {
	Input     fxInput   `json:"input"`
	Output    fxState   `json:"output"`
	Events    []fxEvent `json:"events"`
	Result    int       `json:"result_ref"`
	Source    int       `json:"source_ref"`
	Fault     bool      `json:"fault"`
	Preserved bool      `json:"nonoutput_bytes_preserved"`
}
type fxHooks struct {
	Delete    func(*Object)
	Random    func(float32, float32) float64
	Create    func(types.Pointf, int, int, types.Pointf, float32, *Object) *Object
	Push      func(*types.Pointf, float32, float32, float32)
	Direction func(int32) float32
}

func fxRun(t *testing.T, in fxInput, sourceRef int, want fxState) (fxState, []fxEvent, int, any) {
	t.Helper()
	var objs [3]*Object
	var data [2]*[4]uint32
	for n := range objs {
		objs[n] = new(Object)
	}
	for n := range data {
		data[n] = new([4]uint32)
	}
	if unsafe.Sizeof(uintptr(0)) == 8 && uintptr(unsafe.Pointer(objs[0])) <= math.MaxUint32 {
		t.Fatal("not high native objects")
	}

	ref := func(n int) *Object {
		if n < 0 {
			return nil
		}
		return objs[n]
	}
	dataref := func(n int) unsafe.Pointer {
		if n < 0 {
			return nil
		}
		return unsafe.Pointer(data[n])
	}
	id := func(o *Object) int {
		if o == nil {
			return -1
		}
		for n, p := range objs {
			if o == p {
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
		t.Fatal("unknown data")
		return -2
	}
	for n, v := range in.Objects {
		o := objs[n]
		*o = Object{PosVec: types.Ptf(math.Float32frombits(v[0]), math.Float32frombits(v[1])), PrevPos: types.Ptf(math.Float32frombits(v[2]), math.Float32frombits(v[3])), ForceVec: types.Ptf(math.Float32frombits(v[4]), math.Float32frombits(v[5])), Float28: math.Float32frombits(v[6]), SpeedCur: math.Float32frombits(v[7])}
	}
	objs[0].Direction1 = Dir16(in.Direction)
	objs[0].UpdateData = dataref(in.Update)
	for n := range data {
		*data[n] = in.Data[n]
	}

	randomIndex, created := 0, 0
	snapshot := func() fxState {
		s := fxState{Direction: uint16(objs[0].Direction1), Update: dataID(objs[0].UpdateData), RandomIndex: randomIndex, Created: created}
		for n, o := range objs {
			s.Objects[n] = [8]uint32{math.Float32bits(o.PosVec.X), math.Float32bits(o.PosVec.Y), math.Float32bits(o.PrevPos.X), math.Float32bits(o.PrevPos.Y), math.Float32bits(o.ForceVec.X), math.Float32bits(o.ForceVec.Y), math.Float32bits(o.Float28), math.Float32bits(o.SpeedCur)}
		}
		for n, d := range data {
			s.Data[n] = *d
		}
		return s
	}
	events := make([]fxEvent, 0)
	event := func(name string, args ...int64) {
		events = append(events, fxEvent{Service: name, Args: args, fxState: snapshot()})
	}
	applied := false
	mutate := func(service string, number int, p *types.Pointf) {
		prefix := service + ":"
		if applied || !strings.HasPrefix(in.Mutation, prefix) || in.At != number {
			return
		}
		applied = true
		o := objs[0]
		switch strings.TrimPrefix(in.Mutation, prefix) {
		case "position":
			o.PosVec = types.Ptf(100.1, -200.3)
		case "previous":
			o.PrevPos = types.Ptf(31.7, -19.9)
		case "force":
			o.ForceVec = types.Ptf(3.1, -7.3)
		case "speed":
			o.SpeedCur = 99.1
		case "direction":
			o.Direction1 = 127
		case "update":
			o.UpdateData = dataref(1)
		case "life":
			data[0][1] = 0xffffffff
		case "kind":
			data[0][3] = 4
		case "radius":
			data[0][0] = math.Float32bits(99.1)
		case "power":
			data[0][2] = math.Float32bits(-33.3)
		case "point":
			*p = types.Ptf(17.3, -29.7)
		default:
			t.Fatal("unknown mutation")
		}
	}
	h := fxHooks{
		Delete: func(o *Object) {
			if o != objs[0] {
				t.Fatal("delete identity")
			}
			event("delete", 0)
			mutate("delete", 0, nil)
		},
		Random: func(lo, hi float32) float64 {
			n := randomIndex
			event("random", int64(math.Float32bits(lo)), int64(math.Float32bits(hi)), int64(n))
			mutate("random", n, nil)
			randomIndex++
			return math.Float64frombits(in.Random[n])
		},
		Create: func(p types.Pointf, k, l int, v types.Pointf, z float32, owner *Object) *Object {
			if owner != objs[0] {
				t.Fatal("owner identity")
			}
			n := created
			event("create", int64(math.Float32bits(p.X)), int64(math.Float32bits(p.Y)), int64(k), int64(l), int64(math.Float32bits(v.X)), int64(math.Float32bits(v.Y)), int64(math.Float32bits(z)), 0)
			mutate("create", n, nil)
			created++
			return ref(in.Returns[n])
		},
		Push: func(p *types.Pointf, outer, inner, force float32) {
			alias := int64(0)
			if p != &objs[0].PosVec {
				alias = -1
			}
			event("push", alias, int64(math.Float32bits(p.X)), int64(math.Float32bits(p.Y)), int64(math.Float32bits(outer)), int64(math.Float32bits(inner)), int64(math.Float32bits(force)), 0, 0, 0)
			mutate("push", 0, p)
		},
		Direction: func(index int32) float32 {
			base := int32(int16(in.Direction)) * 2
			if index != base && index != base+1 {
				t.Fatal("signed direction index", index, base)
			}
			return math.Float32frombits(in.DirectionWords[index-base])
		},
	}
	var recovered any
	result := -1
	func() {
		defer func() { recovered = recover() }()
		switch in.Kind {
		case 0:
			SparkUpdateRaw53ADC0(ref(sourceRef), h.Delete)
		case 1:
			result = id(ProjectileTrailUpdateRaw53AEC0(ref(sourceRef), ProjectileTrailRawRuntime53AEC0{DirectionFloat: h.Direction, RandomFloat: h.Random, CreateSpark: h.Create}))
		case 2:
			PushUpdateRaw53B030(ref(sourceRef), h.Push)
		}
	}()

	return snapshot(), events, result, recovered
}
func TestSparkTrailPushOriginalGo(t *testing.T) {
	blob, err := os.ReadFile("testdata/spark_trail_push_original.json")
	if err != nil {
		t.Fatal(err)
	}
	var f struct {
		SHA      string  `json:"original_sha256"`
		CW       uint32  `json:"original_startup_FPCW"`
		Contexts int     `json:"contexts"`
		Rows     []fxRow `json:"rows"`
	}
	if err = json.Unmarshal(blob, &f); err != nil {
		t.Fatal(err)
	}
	if f.SHA != "0040e2c0683b4d73a5fb976e400d5087dca680df2b195c9e27f8edbda2d4974a" || f.CW != 0xe3f || f.Contexts != 487 || len(f.Rows) != 487 {
		t.Fatal("original provenance")
	}
	var report *os.File
	if path := os.Getenv("OPENNOX_FX_REPORT"); path != "" {
		report, err = os.OpenFile(path, os.O_CREATE|os.O_EXCL|os.O_WRONLY, 0600)
		if err != nil {
			t.Fatal(err)
		}
		defer report.Close()
	}
	for _, row := range f.Rows {
		t.Run(row.Input.Name, func(t *testing.T) {
			out, events, result, fault := fxRun(t, row.Input, 0, row.Output)
			if report != nil {
				if err := json.NewEncoder(report).Encode(map[string]any{"name": row.Input.Name, "output": out, "events": events, "result_ref": result, "state_matches": reflect.DeepEqual(out, row.Output), "events_match": reflect.DeepEqual(events, row.Events), "fault": fault != nil}); err != nil {
					t.Fatal(err)
				}
			}
			if !row.Preserved || fault != nil {
				t.Fatal("original/native fault/canary", fault)
			}
			if !reflect.DeepEqual(out, row.Output) || result != row.Result {
				t.Errorf("state/result differs: got %+v/%d want %+v/%d", out, result, row.Output, row.Result)
			}
			if !reflect.DeepEqual(events, row.Events) {
				t.Errorf("service arguments/order/state differ: got %d want %d", len(events), len(row.Events))
			}
		})
	}
}
func TestSparkTrailPushOriginalFaultsGo(t *testing.T) {
	blob, err := os.ReadFile("testdata/spark_trail_push_faults_original.json")
	if err != nil {
		t.Fatal(err)
	}
	var f struct {
		Rows []fxRow `json:"rows"`
	}
	if err = json.Unmarshal(blob, &f); err != nil {
		t.Fatal(err)
	}
	if len(f.Rows) != 6 {
		t.Fatal("fault provenance")
	}
	for _, row := range f.Rows {
		t.Run(row.Input.Name, func(t *testing.T) {
			out, events, result, fault := fxRun(t, row.Input, row.Source, row.Output)
			if (fault != nil) != row.Fault || !reflect.DeepEqual(out, row.Output) || !reflect.DeepEqual(events, row.Events) {
				t.Fatal("fault/gate/prefix/state differs", fault)
			}
			if !row.Fault && result != row.Result {
				t.Fatal("safe result identity")
			}
		})
	}
}
