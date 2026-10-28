package server

import (
	"encoding/json"
	"fmt"
	"math"
	"os"
	"reflect"
	"testing"
	"unsafe"

	"github.com/opennox/libs/object"
	"github.com/opennox/libs/types"
)

type cloudOriginalState53D850 struct {
	Objects   [7][5]uint32 `json:"objects"`
	Owners    [7]int       `json:"owners"`
	Updates   [7]int       `json:"update_refs"`
	Callbacks [7]int       `json:"callbacks"`
	Data      [2]uint32    `json:"data"`
	Frame     uint32       `json:"frame"`
}
type cloudOriginalInput53D850 struct {
	Name     string `json:"name"`
	Kind     int    `json:"kind"`
	Visits   []int  `json:"visits"`
	Trace    int32  `json:"trace_result"`
	Enemy    int32  `json:"enemy_result"`
	Damage   int32  `json:"damage"`
	Delay    int32  `json:"delay"`
	Mutation string `json:"mutation"`
	cloudOriginalState53D850
}
type cloudOriginalEvent53D850 struct {
	Service string  `json:"service"`
	Args    []int64 `json:"args"`
	cloudOriginalState53D850
}
type cloudOriginalRow53D850 struct {
	Input     cloudOriginalInput53D850   `json:"input"`
	Output    cloudOriginalState53D850   `json:"output"`
	Events    []cloudOriginalEvent53D850 `json:"events"`
	Preserved bool                       `json:"nonoutput_bytes_preserved"`
}
type cloudOriginalFixture53D850 struct {
	SHA        string                   `json:"original_sha256"`
	CW         uint32                   `json:"original_startup_FPCW"`
	Contexts   int                      `json:"contexts"`
	Executions int                      `json:"executions"`
	Rows       []cloudOriginalRow53D850 `json:"rows"`
}

func cloudOriginalRun53D850(t *testing.T, input cloudOriginalInput53D850, sourceRef, targetRef int, want cloudOriginalState53D850) (cloudOriginalState53D850, []cloudOriginalEvent53D850, any) {
	t.Helper()
	var objs [7]*Object
	var data [2]*ToxicCloudUpdateData
	var callbacks [2]unsafe.Pointer
	for n := range objs {
		objs[n] = new(Object)
	}
	for n := range data {
		data[n] = &ToxicCloudUpdateData{Duration: int32(input.Data[n])}
	}
	for n := range callbacks {
		callbacks[n] = unsafe.Pointer(new(byte))
	}
	ref := func(n int) *Object {
		if n < 0 {
			return nil
		}
		return objs[n]
	}
	objID := func(o *Object) int {
		if o == nil {
			return -1
		}
		for n, p := range objs {
			if o == p {
				return n
			}
		}
		t.Fatal("unknown object identity")
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
		t.Fatal("unknown update identity")
		return -2
	}
	fnID := func(p unsafe.Pointer) int {
		if p == nil {
			return -1
		}
		for n, fn := range callbacks {
			if fn == p {
				return n
			}
		}
		t.Fatal("unknown callback identity")
		return -2
	}
	for n, v := range input.Objects {
		o := objs[n]
		o.ObjClass = object.Class(v[0])
		o.ObjFlags = object.Flags(v[1])
		o.PosVec = types.Ptf(math.Float32frombits(v[2]), math.Float32frombits(v[3]))
		o.Field34 = v[4]
		o.ObjOwner = ref(input.Owners[n])
		if k := input.Updates[n]; k >= 0 {
			o.UpdateData = unsafe.Pointer(data[k])
		}
		if k := input.Callbacks[n]; k >= 0 {
			o.Damage = callbacks[k]
		}
	}
	snapshot := func() cloudOriginalState53D850 {
		out := cloudOriginalState53D850{Frame: input.Frame}
		for n, o := range objs {
			out.Objects[n] = [5]uint32{uint32(o.ObjClass), uint32(o.ObjFlags), math.Float32bits(o.PosVec.X), math.Float32bits(o.PosVec.Y), o.Field34}
			out.Owners[n] = objID(o.ObjOwner)
			out.Updates[n] = dataID(o.UpdateData)
			out.Callbacks[n] = fnID(o.Damage)
		}
		for n, d := range data {
			out.Data[n] = uint32(d.Duration)
		}
		return out
	}
	events := make([]cloudOriginalEvent53D850, 0)
	event := func(name string, args ...int64) {
		events = append(events, cloudOriginalEvent53D850{Service: name, Args: args, cloudOriginalState53D850: snapshot()})
	}
	pointBits := func(p *types.Pointf) []int64 {
		return []int64{int64(math.Float32bits(p.X)), int64(math.Float32bits(p.Y))}
	}
	applied := false
	mutate := func(service string, p *types.Pointf, r *types.Rectf) {
		prefix := service + ":"
		if applied || len(input.Mutation) < len(prefix) || input.Mutation[:len(prefix)] != prefix {
			return
		}
		applied = true
		switch action := input.Mutation[len(prefix):]; action {
		case "position":
			objs[0].PosVec = types.Ptf(500.1, -600.3)
		case "target-position":
			objs[1].PosVec = types.Ptf(611.3, -817.7)
		case "owner":
			objs[0].ObjOwner = objs[6]
		case "owner-nil":
			objs[0].ObjOwner = nil
		case "update":
			objs[0].UpdateData = unsafe.Pointer(data[1])
		case "duration":
			data[0].Duration = 9
		case "expire":
			data[0].Duration = 0
		case "clock":
			input.Frame = 0xfffffffc
		case "deadline":
			objs[0].Field34 = 0
		case "callback":
			objs[1].Damage = callbacks[1]
		case "target-class":
			objs[1].ObjClass = 0
		case "point":
			*p = types.Ptf(711.3, -917.7)
		case "trace-point":
			r.Max = types.Ptf(17.3, -91.7)
		default:
			t.Fatal("unknown mutation", action)
		}
	}
	h := ToxicCloudRawRuntime53D850{
		Frame: func() uint32 { return input.Frame },
		EachInCircle: func(p *types.Pointf, r float32, visit func(*Object)) {
			callback := int64(1)
			if input.Kind == 2 {
				callback = 3
			}
			event("circle", append(pointBits(p), int64(math.Float32bits(r)), callback, 0)...)
			mutate("circle", p, nil)
			for _, n := range input.Visits {
				visit(objs[n])
			}
		},
		RandomInt: func(a, b int) int {
			name, result := "delay", input.Delay
			if a == 3 && b == 10 {
				name = "damage-random"
				result = input.Damage
			} else if a != 5 || b != 10 {
				t.Fatal("random bounds")
			}
			event(name, int64(a), int64(b))
			mutate(name, nil, nil)
			return int(result)
		},
		TraceRay: func(r *types.Rectf, flags MapTraceFlags) int32 {
			event("trace", append(append(pointBits(&r.Min), pointBits(&r.Max)...), 0, 0, int64(flags))...)
			mutate("trace", nil, r)
			return input.Trace
		},
		FindOwnerPlayer: func(o *Object) *Object {
			if o != objs[0] {
				t.Fatal("owner source")
			}
			owner := o.ObjOwner
			event("owner", 0)
			mutate("owner", nil, nil)
			return owner
		},
		Damage: func(fn unsafe.Pointer, target, owner, weapon *Object, d int32, typ object.DamageType) {
			if fn == nil {
				var required DamageFunc
				required(target, owner, weapon, d, typ)
			}
			if weapon != objs[0] {
				t.Fatal("damage weapon")
			}
			event("damage", int64(fnID(fn)), int64(objID(target)), int64(objID(owner)), 0, int64(d), int64(typ))
			mutate("damage", nil, nil)
		},
		IsEnemy: func(owner, target *Object) int32 {
			event("enemy", int64(objID(owner)), int64(objID(target)))
			mutate("enemy", nil, nil)
			return input.Enemy
		},
		ActivatePoison: func(target *Object, a, b int32) int32 {
			event("poison", int64(objID(target)), int64(a), int64(b))
			mutate("poison", nil, nil)
			return 1
		},
		DelayedDelete: func(o *Object) {
			if o != objs[0] {
				t.Fatal("delete identity")
			}
			event("delete", 0)
			mutate("delete", nil, nil)
		},
	}
	var recovered any
	func() {
		defer func() { recovered = recover() }()
		if input.Kind == 0 || input.Kind == 2 {
			ToxicCloudUpdateRaw53D850(ref(sourceRef), input.Kind == 2, h)
		} else {
			ToxicCloudAffectTargetRaw53D8C0(ref(targetRef), ref(sourceRef), input.Kind == 3, h)
		}
	}()
	return snapshot(), events, recovered
}

func TestToxicCloudOriginal53D850(t *testing.T) {
	blob, err := os.ReadFile("testdata/toxic_cloud_53d850_original.json")
	if err != nil {
		t.Fatal(err)
	}
	var fixture cloudOriginalFixture53D850
	if err = json.Unmarshal(blob, &fixture); err != nil {
		t.Fatal(err)
	}
	if fixture.SHA != "0040e2c0683b4d73a5fb976e400d5087dca680df2b195c9e27f8edbda2d4974a" || fixture.CW != 0xe3f || fixture.Contexts != 952 || fixture.Executions != 952 || len(fixture.Rows) != 952 {
		t.Fatal("original provenance")
	}
	var report *os.File
	if p := os.Getenv("OPENNOX_CLOUD_STATE_REPORT"); p != "" {
		report, err = os.OpenFile(p, os.O_CREATE|os.O_EXCL|os.O_WRONLY, 0600)
		if err != nil {
			t.Fatal(err)
		}
		defer report.Close()
	}
	for _, row := range fixture.Rows {
		t.Run(row.Input.Name, func(t *testing.T) {
			if !row.Preserved {
				t.Fatal("original canary")
			}
			output, events, recovered := cloudOriginalRun53D850(t, row.Input, 0, 1, row.Output)
			if recovered != nil {
				t.Fatal("unexpected fault", recovered)
			}
			if report != nil {
				if err := json.NewEncoder(report).Encode(map[string]any{"name": row.Input.Name, "output": output, "events": events, "output_matches": reflect.DeepEqual(output, row.Output), "events_match": reflect.DeepEqual(events, row.Events)}); err != nil {
					t.Fatal(err)
				}
			}
			if !reflect.DeepEqual(output, row.Output) {
				t.Error("output differs from unchanged original")
			}
			if !reflect.DeepEqual(events, row.Events) {
				t.Errorf("service arguments/order/state differ: got %d want %d", len(events), len(row.Events))
			}
		})
	}
}

func TestToxicCloudOriginalPointerContracts53D850(t *testing.T) {
	blob, err := os.ReadFile("testdata/toxic_cloud_53d850_faults_original.json")
	if err != nil {
		t.Fatal(err)
	}
	var fixture struct {
		SHA    string `json:"original_sha256"`
		Faults int    `json:"faults"`
		Safe   int    `json:"safe_controls"`
		Rows   []struct {
			cloudOriginalRow53D850
			Source int  `json:"source_ref"`
			Target int  `json:"target_ref"`
			Fault  bool `json:"fault"`
		} `json:"rows"`
	}
	if err = json.Unmarshal(blob, &fixture); err != nil {
		t.Fatal(err)
	}
	if fixture.SHA != "0040e2c0683b4d73a5fb976e400d5087dca680df2b195c9e27f8edbda2d4974a" || fixture.Faults != 12 || fixture.Safe != 4 || len(fixture.Rows) != 16 {
		t.Fatal("original pointer provenance")
	}
	for _, row := range fixture.Rows {
		t.Run(fmt.Sprintf("%d/%s", row.Input.Kind, row.Input.Name), func(t *testing.T) {
			output, events, recovered := cloudOriginalRun53D850(t, row.Input, row.Source, row.Target, row.Output)
			if (recovered != nil) != row.Fault {
				t.Fatalf("fault %v, want %v", recovered, row.Fault)
			}
			if recovered != nil {
				if _, ok := recovered.(interface{ RuntimeError() }); !ok {
					t.Fatalf("non-memory panic: %v", recovered)
				}
			}
			if !reflect.DeepEqual(output, row.Output) || !reflect.DeepEqual(events, row.Events) {
				t.Fatal("original fault-visible state/service prefix differs")
			}
		})
	}
}
