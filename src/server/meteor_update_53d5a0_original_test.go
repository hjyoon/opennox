package server

import (
	"encoding/json"
	"image"
	"math"
	"os"
	"reflect"
	"testing"
	"unsafe"

	"github.com/opennox/libs/object"
	"github.com/opennox/libs/types"
)

type meteorOriginalState53D5A0 struct {
	Objects [4][9]uint32 `json:"objects"`
	Owners  [4]int       `json:"owners"`
	Updates [4]int       `json:"update_refs"`
	Data    [3]uint32    `json:"data"`
	Frame   uint32       `json:"frame"`
	FPS     uint32       `json:"fps"`
}
type meteorOriginalInput53D5A0 struct {
	Name        string `json:"name"`
	Kind        int    `json:"kind"`
	Direction   int32  `json:"direction"`
	Radius      uint64 `json:"radius_bits64"`
	Delay       int32  `json:"delay"`
	TraceResult int32  `json:"trace_result"`
	Allocate    bool   `json:"allocate"`
	Mutation    string `json:"mutation"`
	meteorOriginalState53D5A0
}
type meteorOriginalEvent53D5A0 struct {
	Service string  `json:"service"`
	Args    []int64 `json:"args"`
	meteorOriginalState53D5A0
}
type meteorOriginalFixture53D5A0 struct {
	SHA        string `json:"original_sha256"`
	CW         uint32 `json:"original_startup_FPCW"`
	Contexts   int    `json:"contexts"`
	Executions int    `json:"executions"`
	Table      struct {
		Words [512]uint32 `json:"words"`
	} `json:"direction_table"`
	Rows []struct {
		Input     meteorOriginalInput53D5A0   `json:"input"`
		Output    meteorOriginalState53D5A0   `json:"output"`
		Events    []meteorOriginalEvent53D5A0 `json:"events"`
		Preserved bool                        `json:"nonoutput_bytes_preserved"`
	} `json:"rows"`
}

func TestMeteorUpdatesOriginal53D5A0(t *testing.T) {
	blob, err := os.ReadFile("testdata/meteor_update_53d5a0_original.json")
	if err != nil {
		t.Fatal(err)
	}
	var fixture meteorOriginalFixture53D5A0
	if err = json.Unmarshal(blob, &fixture); err != nil {
		t.Fatal(err)
	}
	if fixture.SHA != "0040e2c0683b4d73a5fb976e400d5087dca680df2b195c9e27f8edbda2d4974a" || fixture.CW != 0xe3f || fixture.Contexts != 884 || fixture.Executions != 1768 || len(fixture.Rows) != 884 {
		t.Fatal("original provenance")
	}
	var report *os.File
	if path := os.Getenv("OPENNOX_METEOR_STATE_REPORT"); path != "" {
		report, err = os.OpenFile(path, os.O_CREATE|os.O_EXCL|os.O_WRONLY, 0600)
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
			input := row.Input
			frame, fps := input.Frame, input.FPS
			var objs [4]*Object
			var data [3]*MeteorUpdateData
			for n := range objs {
				objs[n] = new(Object)
			}
			for n := range data {
				data[n] = &MeteorUpdateData{Damage: int32(input.Data[n])}
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
					if p == o {
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
					if unsafe.Pointer(d) == p {
						return n
					}
				}
				t.Fatal("unknown update identity")
				return -2
			}
			for n, v := range input.Objects {
				o := objs[n]
				o.ObjClass = object.Class(v[0])
				o.ObjFlags = object.Flags(v[1])
				o.Field5 = v[2]
				o.PosVec = types.Ptf(math.Float32frombits(v[3]), math.Float32frombits(v[4]))
				o.ZVal = math.Float32frombits(v[5])
				o.Field27 = math.Float32frombits(v[6])
				o.Field32 = v[7]
				o.Field34 = v[8]
				o.ObjOwner = ref(input.Owners[n])
				if k := input.Updates[n]; k >= 0 {
					o.UpdateData = unsafe.Pointer(data[k])
				}
			}
			snapshot := func() meteorOriginalState53D5A0 {
				out := meteorOriginalState53D5A0{Frame: frame, FPS: fps}
				for n, o := range objs {
					out.Objects[n] = [9]uint32{uint32(o.ObjClass), uint32(o.ObjFlags), o.Field5, math.Float32bits(o.PosVec.X), math.Float32bits(o.PosVec.Y), math.Float32bits(o.ZVal), math.Float32bits(o.Field27), o.Field32, o.Field34}
					out.Owners[n] = objID(o.ObjOwner)
					out.Updates[n] = dataID(o.UpdateData)
				}
				for n, d := range data {
					out.Data[n] = uint32(d.Damage)
				}
				return out
			}
			events := make([]meteorOriginalEvent53D5A0, 0)
			event := func(name string, args ...int64) {
				events = append(events, meteorOriginalEvent53D5A0{Service: name, Args: args, meteorOriginalState53D5A0: snapshot()})
			}
			pointBits := func(p *types.Pointf) []int64 {
				return []int64{int64(math.Float32bits(p.X)), int64(math.Float32bits(p.Y))}
			}
			mutate := func(service string, p *types.Pointf, r *types.Rectf) {
				prefix := service + ":"
				if len(input.Mutation) < len(prefix) || input.Mutation[:len(prefix)] != prefix {
					return
				}
				switch action := input.Mutation[len(prefix):]; action {
				case "position":
					objs[0].PosVec = types.Ptf(500.1, -600.3)
				case "owner":
					objs[0].ObjOwner = objs[3]
				case "owner-nil":
					objs[0].ObjOwner = nil
				case "update":
					objs[0].UpdateData = unsafe.Pointer(data[2])
				case "spawn-update":
					objs[1].UpdateData = unsafe.Pointer(data[2])
				case "damage":
					data[0].Damage = 73
				case "negative-damage":
					data[0].Damage = -2147483647
				case "clock":
					frame = 0xfffffffc
					fps = 0
				case "created":
					objs[0].Field32 = frame
				case "flags":
					objs[1].Field5 = 0x400
				case "velocity":
					objs[1].Field27 = 101
				case "point":
					*p = types.Ptf(611.3, -817.7)
				case "trace-point":
					r.Max = types.Ptf(17.3, -91.7)
				default:
					t.Fatal("unknown mutation", action)
				}
			}
			newObject := func(name string) *Object {
				want := "Meteor"
				if input.Kind == 1 {
					want = "MeteorExplode"
				}
				if name != want {
					t.Fatal("allocation type", name)
				}
				event("new", int64(input.Kind))
				mutate("new", nil, nil)
				if input.Allocate {
					return objs[1]
				}
				return nil
			}
			createAt := func(o, owner *Object, p types.Pointf) {
				if o != objs[1] {
					t.Fatal("spawn identity")
				}
				event("create", append([]int64{1, int64(objID(owner))}, append(pointBits(&p), 0)...)...)
				mutate("create", nil, nil)
				o.ObjOwner = owner
				o.PosVec = p
			}
			deleteObject := func(o *Object) {
				if o != objs[0] {
					t.Fatal("delete identity")
				}
				event("delete", 0)
				mutate("delete", nil, nil)
			}
			shower := MeteorShowerRawRuntime53D5A0{
				Frame: func() uint32 { return frame }, FPS: func() uint32 { return fps },
				RandomInt: func(a, b int) int {
					name := "delay"
					result := input.Delay
					if a == 0 && b == 255 {
						name = "direction"
						result = input.Direction
					} else if a != 4 || b != 8 {
						t.Fatal("random bounds")
					}
					event(name, int64(a), int64(b))
					mutate(name, nil, nil)
					return int(result)
				},
				RandomFloat: func(a, b float64) float64 {
					if a != 4 || b != 12 {
						t.Fatal("radius bounds")
					}
					event("radius", int64(math.Float32bits(float32(a))), int64(math.Float32bits(float32(b))))
					mutate("radius", nil, nil)
					return math.Float64frombits(input.Radius)
				},
				DirectionFloat: func(index int32) float32 {
					if index < 0 || index >= 512 {
						t.Fatal("vector index")
					}
					return math.Float32frombits(fixture.Table.Words[index])
				},
				TraceRay: func(r *types.Rectf, o *Object) int32 {
					if o != objs[0] {
						t.Fatal("trace context")
					}
					event("trace", append(append(pointBits(&r.Min), pointBits(&r.Max)...), 0)...)
					mutate("trace", nil, r)
					return input.TraceResult
				},
				NewObjectByTypeID: newObject, CreateAt: createAt, DelayedDelete: deleteObject,
				Raise: func(o *Object, z float32) {
					if o != objs[1] {
						t.Fatal("raise identity")
					}
					event("raise", 1, int64(math.Float32bits(z)))
					mutate("raise", nil, nil)
					o.ZVal = z
				},
			}
			meteor := MeteorRawRuntime53D6E0{
				AudioEvent: func(id uint32, o *Object) {
					if o != objs[0] {
						t.Fatal("audio identity")
					}
					event("audio", int64(id), 0, 0, 0)
					mutate("audio", nil, nil)
				},
				MakeScorch: func(p *types.Pointf, k int) {
					event("scorch", append([]int64{int64(k)}, pointBits(p)...)...)
					mutate("scorch", p, nil)
				},
				Earthquake: func(p *types.Pointf, k int) {
					event("quake", append([]int64{int64(k)}, pointBits(p)...)...)
					mutate("quake", p, nil)
				},
				NewObjectByTypeID: newObject, CreateAt: createAt, DelayedDelete: deleteObject,
				FindParentChainPlayer: func(o *Object) *Object {
					if o != objs[0] {
						t.Fatal("parent identity")
					}
					owner := o.ObjOwner
					event("parent", 0)
					mutate("parent", nil, nil)
					return owner
				},
				DamageUnitsAround: func(p *types.Pointf, a, b float32, d int, typ object.DamageType, owner *Object, x Obj) {
					if x != nil {
						t.Fatal("exclusion")
					}
					event("units", append(pointBits(p), int64(math.Float32bits(a)), int64(math.Float32bits(b)), int64(d), int64(typ), int64(objID(owner)), -1)...)
					mutate("units", p, nil)
				},
				DamageWalls: func(r image.Rectangle, p *types.Pointf, a float32, d int, typ object.DamageType, o *Object) {
					if o != objs[0] {
						t.Fatal("wall source")
					}
					values := []int64{int64(r.Min.X), int64(r.Min.Y), int64(r.Max.X), int64(r.Max.Y)}
					event("walls", append(append(values, pointBits(p)...), int64(math.Float32bits(a)), int64(d), int64(typ), 0)...)
					mutate("walls", p, nil)
				},
			}
			if input.Kind == 0 {
				MeteorShowerUpdateRaw53D5A0(objs[0], shower)
			} else {
				MeteorUpdateRaw53D6E0(objs[0], meteor)
			}
			output := snapshot()
			if report != nil {
				if err := json.NewEncoder(report).Encode(map[string]any{"name": input.Name, "output": output, "events": events, "output_matches": reflect.DeepEqual(output, row.Output), "events_match": reflect.DeepEqual(events, row.Events)}); err != nil {
					t.Fatal(err)
				}
			}
			if !reflect.DeepEqual(output, row.Output) {
				t.Errorf("output differs from unchanged original")
			}
			if !reflect.DeepEqual(events, row.Events) {
				t.Errorf("service arguments/order/state differ: got %d events want %d", len(events), len(row.Events))
			}
		})
	}
}
