package opennox

import (
	"encoding/json"
	"os"
	"sync"
	"testing"
	"unsafe"

	"github.com/opennox/libs/object"
	"github.com/opennox/libs/strman"
	"github.com/opennox/libs/types"
	"github.com/opennox/opennox/v1/legacy"
	"github.com/opennox/opennox/v1/server"
)

var radialDamageCaptureRegister4E25B0 sync.Once

func radialDamageCaptureSetup4E25B0() {
	radialDamageCaptureRegister4E25B0.Do(func() {
		server.RegisterObjectDamageGo("RadialDamageCaptureFixture4E25B0", legacy.RadialDamageCapturePointer4E25B0(false), legacy.RadialDamageCaptureCall4E25B0)
	})
}

func TestMapRadialDamageNative4E25B0RealMapTraceAndCallback(t *testing.T) {
	radialDamageCaptureSetup4E25B0()
	// The actual map index, ray test, damage dispatch, walls and C bridge run.
	// Only the target damage effect is captured; one mode moves the center
	// through that real C callback to distinguish pointers from value copies.
	b, err := os.ReadFile("server/testdata/map_radial_damage_original_4e25b0.json")
	if err != nil {
		t.Fatal(err)
	}
	var f struct {
		Rows []struct {
			Input  struct{ Name string }
			Events []struct {
				Service string
				Args    []int64
			}
		}
	}
	if err := json.Unmarshal(b, &f); err != nil {
		t.Fatal(err)
	}
	expectations := make(map[string]int32)
	for _, r := range f.Rows {
		for _, e := range r.Events {
			if e.Service == "damage" {
				expectations[r.Input.Name] = int32(e.Args[3])
			}
		}
	}
	if expectations["root/trace-result/1"] != 78 || expectations["root/distance/10.000001/0.0"] != 95 {
		t.Fatal("original full-root expectation")
	}
	for _, scenario := range []struct {
		name, original string
		center, target types.Pointf
	}{
		{"ordinary", "root/trace-result/1", types.Ptf(100.1, 50.7), types.Ptf(125.3, 52.7)},
		// Integer translation preserves the original exact 10.000000953674316
		// difference while keeping the actual trace away from the map boundary.
		{"inner-boundary", "root/distance/10.000001/0.0", types.Ptf(1, 1), types.Ptf(11.000001, 1)},
	} {
		want := expectations[scenario.original]
		for _, mode := range []string{"value", "native", "C"} {
			for _, shift := range []bool{false, true} {
				t.Run(scenario.name+"/"+mode+map[bool]string{false: "/fixed", true: "/live-center"}[shift], func(t *testing.T) {
					s := NewServer(nil, nil, strman.New())
					t.Cleanup(s.Close)
					if !s.Objs.Init(3) || s.Walls.Init() == 0 {
						t.Fatal("native object/wall allocation")
					}
					t.Cleanup(s.Objs.FreeObjects)
					t.Cleanup(s.Walls.Free)
					s.Map.Init()
					t.Cleanup(s.Map.Free)
					old, oldGet, oldFlag := noxServer, legacy.GetServer, doDamageWalls
					noxServer = s
					legacy.GetServer = func() legacy.Server { return s }
					doDamageWalls = false
					t.Cleanup(func() { noxServer = old; legacy.GetServer = oldGet; doDamageWalls = oldFlag })
					source := s.Objs.NewObject(&server.ObjectType{})
					source.PosVec = scenario.center
					callback := legacy.RadialDamageCapturePointer4E25B0(shift)
					var targets []*server.Object
					for range 2 {
						target := s.Objs.NewObject(&server.ObjectType{})
						target.PosVec = scenario.target
						target.NewPos = target.PosVec
						target.ObjClass = object.ClassSimple
						target.ObjFlags = object.FlagActive
						target.Damage = callback
						target.Shape.Kind = server.ShapeKindCircle
						target.Shape.Circle.R, target.Shape.Circle.R2 = 1, 1
						s.Map.AddObjectToIndex(target)
						targets = append(targets, target)
						if unsafe.Sizeof(uintptr(0)) == 8 && uintptr(unsafe.Pointer(target)) <= 0xffffffff {
							t.Fatal("target below4GiB")
						}
					}
					indexed := 0
					s.Map.EachObjInRect(types.Rectf{Min: source.PosVec.Sub(types.Ptf(96, 96)), Max: source.PosVec.Add(types.Ptf(96, 96))}, func(*server.Object) bool { indexed++; return true })
					if indexed != 2 || !s.MapTraceRay(source.PosVec, scenario.target, server.MapTraceFlag1) {
						t.Fatal("real fixture index/trace precondition", indexed)
					}
					switch mode {
					case "value":
						s.Nox_xxx_mapDamageUnitsAround(source.PosVec, 96, 10, 96, object.DamageExplosion, source, nil, false)
					case "native":
						s.MapDamageUnitsAroundNative4E25B0(&source.PosVec, 96, 10, 96, object.DamageExplosion, source, nil)
					case "C":
						legacy.MapRadialDamageCProbe4E25B0(&source.PosVec, 96, 10, 96, object.DamageExplosion, source, nil)
					}
					count, damage, typ, target, gotSource, weapon := legacy.RadialDamageCaptureResult4E25B0()
					wantCount := int32(2)
					if shift && mode != "value" {
						wantCount = 1
					}
					if count != wantCount || damage != want || typ != 7 || gotSource != source || weapon != nil || (target != targets[0] && target != targets[1]) || !doDamageWalls {
						t.Fatalf("actual map/trace/C callback = count%d damage%d type%d target%p source%p weapon%p flag%t", count, damage, typ, target, gotSource, weapon, doDamageWalls)
					}
					if shift && source.PosVec != (types.Ptf(500, 500)) {
						t.Fatal("actual C callback did not mutate center")
					}
				})
			}
		}
	}
}
