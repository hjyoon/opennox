package opennox

import (
	"encoding/json"
	"math"
	"os"
	"strings"
	"sync"
	"testing"
	"unsafe"

	"github.com/opennox/libs/object"
	"github.com/opennox/libs/strman"
	"github.com/opennox/libs/types"
	"github.com/opennox/opennox/v1/common/memmap"
	"github.com/opennox/opennox/v1/internal/porttest"
	"github.com/opennox/opennox/v1/legacy"
	"github.com/opennox/opennox/v1/server"
)

var velocityUpdateRegister5118A0 sync.Once

func TestObjectsVelocityNative5118A0RealQueuePreparationAndMap(t *testing.T) {
	velocityUpdateRegister5118A0.Do(func() {
		server.RegisterObjectUpdateGo("VelocityFixture5118A0", legacy.ObjectsVelocityUpdatePointer5118A0(), func(*server.Object) {}, 0)
	})
	data, err := os.ReadFile("server/testdata/objects_velocity_original_5118a0.json")
	if err != nil {
		t.Fatal(err)
	}
	var fixture struct {
		Rows []struct {
			Input  porttest.VelocityInput5118A0
			Result struct{ State porttest.VelocityState5118A0 }
		}
	}
	if err := json.Unmarshal(data, &fixture); err != nil {
		t.Fatal(err)
	}
	point := func(v [2]uint32) types.Pointf {
		return types.Ptf(math.Float32frombits(v[0]), math.Float32frombits(v[1]))
	}
	words := func(v types.Pointf) [2]uint32 { return [2]uint32{math.Float32bits(v.X), math.Float32bits(v.Y)} }
	selected := 0
	for _, row := range fixture.Rows {
		if !strings.HasPrefix(row.Input.Name, "native-real/") {
			continue
		}
		selected++
		for _, mode := range []string{"Go", "C"} {
			t.Run(row.Input.Name+"/"+mode, func(t *testing.T) {
				s := NewServer(nil, nil, strman.New())
				t.Cleanup(s.Close)
				if !s.Objs.Init(3) || s.Walls.Init() == 0 {
					t.Fatal("native object/wall allocation")
				}
				t.Cleanup(s.Objs.FreeObjects)
				t.Cleanup(s.Walls.Free)
				s.Map.Init()
				t.Cleanup(s.Map.Free)
				oldServer, oldGet := noxServer, legacy.GetServer
				noxServer, legacy.GetServer = s, func() legacy.Server { return s }
				t.Cleanup(func() { noxServer, legacy.GetServer = oldServer, oldGet })
				cache := (*[10]uint32)(unsafe.Pointer(memmap.PtrUint32(0x5D4594, 2386580)))
				oldCache := *cache
				*cache = row.Input.Cache
				t.Cleanup(func() { *cache = oldCache })
				i, want := row.Input.Objects[0], row.Result.State.Objects[0]
				o := s.Objs.NewObject(&server.ObjectType{})
				o.TypeInd, o.ObjClass, o.ObjFlags = uint16(i.Typ), object.Class(i.Cls), object.Flags(i.Flags)
				o.PosVec, o.NewPos, o.VelVec = point(i.Pos), point(i.New), point(i.Vel)
				o.ForceVec, o.Pos24, o.Float28 = point(i.Force), point(i.Acc), math.Float32frombits(i.Drag)
				o.Shape.Kind, o.Mass, o.Field38, o.Update = server.ShapeKindCenter, 1, 0, legacy.ObjectsVelocityUpdatePointer5118A0()
				if unsafe.Sizeof(uintptr(0)) == 8 && uintptr(unsafe.Pointer(o)) <= 0xffffffff {
					t.Fatal("native object below4GiB")
				}
				s.Map.AddObjectToIndex(o)
				if !s.MapTraceRay(o.NewPos, point(want.New), server.MapTraceFlag1) {
					t.Fatal("actual empty-map trace precondition")
				}
				restore := legacy.ObjectsVelocityQueueFixture5118A0(o)
				t.Cleanup(restore)
				if mode == "C" {
					if result := legacy.ObjectsVelocityCProbe5118A0(math.Float32frombits(row.Input.Step)); result != 0 {
						t.Fatal("public C return", result)
					}
				} else {
					legacy.Nox_xxx_updateObjectsVelocity_5118A0(math.Float32frombits(row.Input.Step))
				}
				if words(o.NewPos) != want.New || words(o.VelVec) != want.Vel || words(o.Pos24) != want.Acc || words(o.PosVec) != want.Pos || words(o.ForceVec) != want.Force {
					t.Fatalf("real services: new=%08x vel=%08x acc=%08x; original new=%08x vel=%08x acc=%08x", words(o.NewPos), words(o.VelVec), words(o.Pos24), want.New, want.Vel, want.Acc)
				}
				wantSync := uint32(0)
				if strings.HasSuffix(row.Input.Name, "/moving") {
					wantSync = math.MaxUint32
				}
				if o.Field38 != wantSync || !o.ObjFlags.Has(object.FlagPartitioned) || o.CollideP1 != o.NewPos || o.CollideP2 != o.NewPos {
					t.Fatal("real sync/collider/map index", o.Field38, uint32(o.ObjFlags), o.CollideP1, o.CollideP2, o.NewPos)
				}
				found := 0
				s.Map.EachObjInRect(types.Rectf{Min: o.NewPos.Sub(types.Ptf(1, 1)), Max: o.NewPos.Add(types.Ptf(1, 1))}, func(it *server.Object) bool {
					if it == o {
						found++
					}
					return true
				})
				if found != 1 || *cache != row.Result.State.Cache {
					t.Fatal("real object index/cache", found, *cache)
				}
			})
		}
	}
	if selected != 2 {
		t.Fatal("original real-service case count", selected)
	}
}
