package server

import (
	"bytes"
	"encoding/binary"
	"encoding/json"
	"fmt"
	"image"
	"math"
	"os"
	"reflect"
	"testing"
	"unsafe"

	"github.com/opennox/libs/object"
	"github.com/opennox/libs/types"

	"github.com/opennox/opennox/v1/legacy/common/alloc"
)

type projectileOriginalEvent537850 struct {
	Service  string   `json:"service"`
	Current  []uint32 `json:"current_bits32"`
	Previous []uint32 `json:"previous_bits32"`
	From     []uint32 `json:"from_bits32"`
	To       []uint32 `json:"to_bits32"`
	Flags    uint32   `json:"flags"`
	Grid     []int32  `json:"grid"`
	Ray      []uint32 `json:"ray_bits32"`
}

// Keep the real spatial index, candidate filter, circle intersection and wall
// ray trace. The callback token is only an identity for the real filter;
// trace selection does not invoke collision callbacks or replace a world query.
func TestProjectileTrace537850NativeIndexedSmallCollider(t *testing.T) {
	for _, delta := range []types.Pointf{{X: 9}, {X: -9}, {Y: 9}, {Y: -9}, {X: 15}, {X: -15}, {Y: 15}, {Y: -15}} {
		t.Run(fmt.Sprintf("%g,%g", delta.X, delta.Y), func(t *testing.T) {
			s := newDirectedVisionMap(t)
			source, freeSource := alloc.New(Object{})
			target, freeTarget := alloc.New(Object{})
			token, freeToken := alloc.New(uint32(0x537850))
			t.Cleanup(freeSource)
			t.Cleanup(freeTarget)
			t.Cleanup(freeToken)
			steps := float32(2)
			if math.Abs(float64(delta.X))+math.Abs(float64(delta.Y)) == 15 {
				steps = 3
			}
			origin := types.Ptf(300, 300)
			endpoint := origin.Add(delta)
			position := origin.Add(delta.Mul(1 / steps))
			*source = Object{ObjClass: object.ClassMissile, ObjFlags: object.FlagActive, PosVec: origin, NewPos: endpoint, Collide: unsafe.Pointer(token), Field29: 0x13572468}
			*target = Object{ObjClass: object.ClassObstacle | object.ClassImmobile, ObjFlags: object.FlagActive | object.FlagMissileHit, PosVec: position, NewPos: position, Collide: unsafe.Pointer(token), Field29: 0x87654321}
			target.Shape.Kind = ShapeKindCircle
			target.Shape.Circle.R = .2
			target.Shape.Circle.R2 = .04
			for _, p := range []unsafe.Pointer{unsafe.Pointer(source), unsafe.Pointer(target), unsafe.Pointer(token)} {
				if unsafe.Sizeof(uintptr(0)) > 4 && uintptr(p) <= math.MaxUint32 {
					t.Fatal("native indexed collider input below 4 GiB")
				}
			}
			s.Map.AddObjectToIndex(target)
			raw := func(o *Object) []byte { return unsafe.Slice((*byte)(unsafe.Pointer(o)), int(unsafe.Sizeof(*o))) }
			beforeSource, beforeTarget := bytes.Clone(raw(source)), bytes.Clone(raw(target))
			hit, normal, ok := s.projectileTraceHitNative537850(source, ProjectileCollisionRuntime537770{
				SetTraceReady:  func(uint32) { t.Fatal("clear native wall map changed trace-ready") },
				StoreTraceGrid: func(image.Point) { t.Fatal("clear native wall map wrote a grid") },
				WallNormal: func(image.Point, types.Rectf) (types.Pointf, bool) {
					t.Fatal("clear native wall map requested a wall normal")
					return types.Pointf{}, false
				},
			})
			want := origin.Sub(position)
			if !ok || hit != target || normal != want || !bytes.Equal(raw(source), beforeSource) || !bytes.Equal(raw(target), beforeTarget) {
				t.Fatalf("actual indexed trace ok=%t hit=%p target=%p normal=%v want=%v source preserved=%t target preserved=%t", ok, hit, target, normal, want, bytes.Equal(raw(source), beforeSource), bytes.Equal(raw(target), beforeTarget))
			}
			t.Log("actual native spatial index -> original candidate filters -> circle intersection -> startup trace selection -> clear wall trace")
		})
	}
}

func TestProjectileTrace537850StartupOriginal(t *testing.T) {
	blob, err := os.ReadFile("testdata/projectile_trace_startup_original.json")
	if err != nil {
		t.Fatal(err)
	}
	var fixture struct {
		SHA       string            `json:"original_sha256"`
		Unchanged bool              `json:"original_unchanged"`
		CW        uint16            `json:"actual_startup_control_word"`
		Contexts  int               `json:"contexts"`
		Faults    []json.RawMessage `json:"memory_faults"`
		Rows      []struct {
			Mode  string `json:"mode"`
			Input struct {
				Name           string    `json:"name"`
				Position       [2]uint32 `json:"position_bits32"`
				NewPosition    [2]uint32 `json:"new_position_bits32"`
				TargetPosition [2]uint32 `json:"target_position_bits32"`
				HitMode        int       `json:"hit_mode"`
				HitIndex       int       `json:"hit_index"`
				Blocked        bool      `json:"blocked"`
				WallValid      bool      `json:"wall_valid"`
				WallPoint      [2]uint32 `json:"wall_point_bits32"`
				WallNormal     [2]uint32 `json:"wall_normal_bits32"`
				SampleMutation int       `json:"sample_mutation"`
				MapMutation    int       `json:"map_mutation"`
				WallMutation   int       `json:"wall_mutation"`
			} `json:"input"`
			OK             bool                            `json:"return_bool"`
			Hit            bool                            `json:"hit"`
			Normal         [2]uint32                       `json:"normal_bits32"`
			Position       [2]uint32                       `json:"final_position_bits32"`
			NewPosition    [2]uint32                       `json:"final_new_position_bits32"`
			TargetPosition [2]uint32                       `json:"final_target_position_bits32"`
			Grid           [2]int                          `json:"trace_grid"`
			Ready          uint32                          `json:"trace_ready"`
			Events         []projectileOriginalEvent537850 `json:"events"`
			Preserved      bool                            `json:"nonoutput_bytes_preserved"`
		} `json:"rows"`
	}
	if err := json.Unmarshal(blob, &fixture); err != nil {
		t.Fatal(err)
	}
	if fixture.SHA != "0040e2c0683b4d73a5fb976e400d5087dca680df2b195c9e27f8edbda2d4974a" || !fixture.Unchanged || fixture.CW != 0xe3f || fixture.Contexts != 166 || len(fixture.Rows) != 332 || len(fixture.Faults) != 0 {
		t.Fatal("incomplete unchanged original trace observations")
	}
	point := func(words [2]uint32) types.Pointf {
		return types.Ptf(math.Float32frombits(words[0]), math.Float32frombits(words[1]))
	}
	bits := func(p types.Pointf) []uint32 { return []uint32{math.Float32bits(p.X), math.Float32bits(p.Y)} }
	selected := 0
	for _, row := range fixture.Rows {
		if row.Mode != "actual-startup" {
			continue
		}
		selected++
		t.Run(row.Input.Name, func(t *testing.T) {
			i := row.Input
			source, freeSource := alloc.New(Object{})
			target, freeTarget := alloc.New(Object{})
			t.Cleanup(freeSource)
			t.Cleanup(freeTarget)
			*source = Object{PosVec: point(i.Position), NewPos: point(i.NewPosition), Field29: 0x12345678}
			*target = Object{PosVec: point(i.TargetPosition), Field29: 0x87654321}
			for _, p := range []*Object{source, target} {
				if unsafe.Sizeof(uintptr(0)) > 4 && uintptr(unsafe.Pointer(p)) <= math.MaxUint32 {
					t.Fatal("native trace object below 4 GiB")
				}
			}
			raw := func(obj *Object) []byte { return unsafe.Slice((*byte)(unsafe.Pointer(obj)), int(unsafe.Sizeof(*obj))) }
			expectedSource, expectedTarget := bytes.Clone(raw(source)), bytes.Clone(raw(target))
			writePoint := func(buf []byte, offset uintptr, words [2]uint32) {
				binary.LittleEndian.PutUint32(buf[int(offset):], words[0])
				binary.LittleEndian.PutUint32(buf[int(offset)+4:], words[1])
			}
			writePoint(expectedSource, unsafe.Offsetof(source.PosVec), row.Position)
			writePoint(expectedSource, unsafe.Offsetof(source.NewPos), row.NewPosition)
			writePoint(expectedTarget, unsafe.Offsetof(target.PosVec), row.TargetPosition)
			var events []projectileOriginalEvent537850
			samples := 0
			grid := image.Pt(12345, -67890)
			ready := uint32(7)
			h := projectileTraceHitHooks537850[*Object]{
				loadPosX: func(o *Object) float32 { return o.PosVec.X }, loadPosY: func(o *Object) float32 { return o.PosVec.Y }, loadNewPosX: func(o *Object) float32 { return o.NewPos.X }, loadNewPosY: func(o *Object) float32 { return o.NewPos.Y },
				storeNewPosX: func(o *Object, v float32) { o.NewPos.X = v }, storeNewPosY: func(o *Object, v float32) { o.NewPos.Y = v }, loadObjectPosX: func(o *Object) float32 { return o.PosVec.X }, loadObjectPosY: func(o *Object) float32 { return o.PosVec.Y },
				sampleObject: func(o *Object, current, previous *types.Pointf) *Object {
					if o != source {
						t.Fatal("sample source changed")
					}
					samples++
					events = append(events, projectileOriginalEvent537850{Service: "sample", Current: bits(*current), Previous: bits(*previous)})
					if samples == 1 && i.SampleMutation == 1 {
						source.PosVec = types.Ptf(17.5, -4.25)
						source.NewPos = types.Ptf(8.25, 2.5)
						target.PosVec = types.Ptf(5.5, 5.75)
					}
					if samples == 1 && i.SampleMutation == 2 {
						*current = types.Ptf(11.25, 2.75)
						*previous = types.Ptf(7.5, -4.5)
					}
					hit := i.HitMode == 1 && samples == i.HitIndex
					if i.HitMode == 2 {
						dx := float64(current.X) - float64(target.PosVec.X)
						dy := float64(current.Y) - float64(target.PosVec.Y)
						hit = dx*dx+dy*dy <= .2*.2
					}
					if hit {
						return target
					}
					return nil
				},
				mapTrace: func(from, to types.Pointf, flags MapTraceFlags) projectileMapTraceResult537850 {
					events = append(events, projectileOriginalEvent537850{Service: "map", From: bits(from), To: bits(to), Flags: uint32(flags)})
					if i.MapMutation != 0 {
						source.PosVec = types.Ptf(77.5, 88.25)
					}
					return projectileMapTraceResult537850{clear: !i.Blocked, point: point(i.WallPoint), grid: image.Pt(7, -3)}
				},
				storeTraceGrid: func(v image.Point) { grid = v }, setTraceReady: func(v uint32) { ready = v },
				wallNormal: func(g image.Point, ray types.Rectf) (types.Pointf, bool) {
					events = append(events, projectileOriginalEvent537850{Service: "wall", Grid: []int32{int32(g.X), int32(g.Y)}, Ray: append(bits(ray.Min), bits(ray.Max)...)})
					if i.WallMutation != 0 {
						source.PosVec = types.Ptf(17.75, -2.25)
						target.PosVec = types.Ptf(16.5, 1.25)
					}
					return point(i.WallNormal), i.WallValid
				},
			}
			hit, normal, ok := projectileTraceHit537850(source, h)
			normalBits := [2]uint32{math.Float32bits(normal.X), math.Float32bits(normal.Y)}
			if ok != row.OK || (hit == target) != row.Hit || (ok && normalBits != row.Normal) || grid != image.Pt(row.Grid[0], row.Grid[1]) || ready != row.Ready || !reflect.DeepEqual(events, row.Events) || !row.Preserved || !bytes.Equal(raw(source), expectedSource) || !bytes.Equal(raw(target), expectedTarget) {
				t.Fatalf("native trace ok/hit/normal=%t/%t/%08x,%08x original=%t/%t/%08x,%08x; events=%+v original events=%+v; source preserved=%t target preserved=%t", ok, hit == target, normalBits[0], normalBits[1], row.OK, row.Hit, row.Normal[0], row.Normal[1], events, row.Events, bytes.Equal(raw(source), expectedSource), bytes.Equal(raw(target), expectedTarget))
			}
		})
	}
	if selected != 166 {
		t.Fatalf("startup trace contexts=%d", selected)
	}
}
