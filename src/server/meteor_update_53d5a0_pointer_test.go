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

func TestMeteorRawPointerContracts53D5A0(t *testing.T) {
	blob, err := os.ReadFile("testdata/meteor_update_53d5a0_faults_original.json")
	if err != nil {
		t.Fatal(err)
	}
	var fixture struct {
		SHA    string `json:"original_sha256"`
		Faults int    `json:"faults"`
		Safe   int    `json:"safe_controls"`
		Rows   []struct {
			Input  meteorOriginalInput53D5A0   `json:"input"`
			Source int                         `json:"source_ref"`
			Fault  bool                        `json:"fault"`
			Events []meteorOriginalEvent53D5A0 `json:"events"`
		} `json:"rows"`
	}
	if err = json.Unmarshal(blob, &fixture); err != nil {
		t.Fatal(err)
	}
	if fixture.SHA != "0040e2c0683b4d73a5fb976e400d5087dca680df2b195c9e27f8edbda2d4974a" || len(fixture.Rows) != 10 || fixture.Faults != 6 || fixture.Safe != 4 {
		t.Fatal("original pointer provenance")
	}
	for _, row := range fixture.Rows {
		t.Run(fmt.Sprintf("%d/%s", row.Input.Kind, row.Input.Name), func(t *testing.T) {
			input := row.Input
			var objs [4]*Object
			var data [3]*MeteorUpdateData
			for n := range objs {
				objs[n] = new(Object)
			}
			for n := range data {
				data[n] = &MeteorUpdateData{Damage: int32(input.Data[n])}
			}
			for n, v := range input.Objects {
				o := objs[n]
				o.PosVec = types.Ptf(math.Float32frombits(v[3]), math.Float32frombits(v[4]))
				o.ZVal = math.Float32frombits(v[5])
				o.Field32 = v[7]
				o.Field34 = v[8]
				if k := input.Updates[n]; k >= 0 {
					o.UpdateData = unsafe.Pointer(data[k])
				}
				if k := input.Owners[n]; k >= 0 {
					o.ObjOwner = objs[k]
				}
			}
			var events []string
			event := func(name string) { events = append(events, name) }
			newObject := func(string) *Object {
				event("new")
				if input.Allocate {
					return objs[1]
				}
				return nil
			}
			create := func(*Object, *Object, types.Pointf) { event("create") }
			deleted := func(*Object) { event("delete") }
			shower := MeteorShowerRawRuntime53D5A0{
				Frame: func() uint32 { return input.Frame }, FPS: func() uint32 { return input.FPS },
				RandomInt: func(a, b int) int {
					if a == 0 && b == 255 {
						event("direction")
						return int(input.Direction)
					}
					event("delay")
					return int(input.Delay)
				},
				RandomFloat: func(float64, float64) float64 { event("radius"); return math.Float64frombits(input.Radius) },
				DirectionFloat: func(index int32) float32 {
					x, y := SinCosDir(byte(index >> 1))
					if index&1 == 0 {
						return x
					}
					return y
				},
				TraceRay:          func(*types.Rectf, *Object) int32 { event("trace"); return input.TraceResult },
				NewObjectByTypeID: newObject, CreateAt: create, Raise: func(*Object, float32) { event("raise") }, DelayedDelete: deleted,
			}
			meteor := MeteorRawRuntime53D6E0{
				AudioEvent: func(uint32, *Object) { event("audio") }, MakeScorch: func(*types.Pointf, int) { event("scorch") }, Earthquake: func(*types.Pointf, int) { event("quake") },
				NewObjectByTypeID: newObject, CreateAt: create, FindParentChainPlayer: func(o *Object) *Object { event("parent"); return o.ObjOwner },
				DamageUnitsAround: func(*types.Pointf, float32, float32, int, object.DamageType, *Object, Obj) { event("units") }, DelayedDelete: deleted,
			}
			var source *Object
			if row.Source >= 0 {
				source = objs[row.Source]
			}
			var recovered any
			func() {
				defer func() { recovered = recover() }()
				if input.Kind == 0 {
					MeteorShowerUpdateRaw53D5A0(source, shower)
				} else {
					MeteorUpdateRaw53D6E0(source, meteor)
				}
			}()
			if (recovered != nil) != row.Fault {
				t.Fatalf("fault %v, want %v", recovered, row.Fault)
			}
			if recovered != nil {
				if _, ok := recovered.(interface{ RuntimeError() }); !ok {
					t.Fatalf("non-memory panic %v", recovered)
				}
			}
			var want []string
			for _, v := range row.Events {
				want = append(want, v.Service)
			}
			if !reflect.DeepEqual(events, want) {
				t.Fatalf("service fault prefix %v, want %v", events, want)
			}
		})
	}
}
