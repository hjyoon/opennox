package legacy

import (
	"bytes"
	"encoding/json"
	"math"
	"os"
	"testing"
	"unsafe"

	"github.com/opennox/libs/object"
	"github.com/opennox/libs/types"
	"github.com/opennox/opennox/v1/legacy/common/alloc"
	"github.com/opennox/opennox/v1/legacy/common/alloc/handles"
	"github.com/opennox/opennox/v1/server"
)

// These call the actual production C symbols. The finite object cases use the
// actual collision queue and impulse callee; no C service is replaced.
func TestCircleCollisionCEntryOriginal(t *testing.T) {
	handles.Init()
	t.Cleanup(handles.Release)
	t.Cleanup(circleCollisionCABIReleaseFixturePool)
	blob, err := os.ReadFile("../server/testdata/circle_collision_startup_original.json")
	if err != nil {
		t.Fatal(err)
	}
	var fixture struct {
		Contexts int `json:"contexts"`
		Rows     []struct {
			Mode  string `json:"mode"`
			Input struct {
				Kind              string    `json:"kind"`
				Name              string    `json:"name"`
				FirstNew          [2]uint32 `json:"first_new"`
				SecondNew         [2]uint32 `json:"second_new"`
				FirstPosition     [2]uint32 `json:"first_position"`
				SecondPosition    [2]uint32 `json:"second_position"`
				FirstVelocity     [2]uint32 `json:"first_velocity"`
				SecondVelocity    [2]uint32 `json:"second_velocity"`
				FirstAccumulator  [2]uint32 `json:"first_accumulator"`
				SecondAccumulator [2]uint32 `json:"second_accumulator"`
				FirstMass         uint32    `json:"first_mass"`
				SecondMass        uint32    `json:"second_mass"`
				FirstRadius       uint32    `json:"first_radius"`
				SecondRadius      uint32    `json:"second_radius"`
				FirstClass        uint32    `json:"first_class"`
				SecondClass       uint32    `json:"second_class"`
				FirstFlags        uint32    `json:"first_flags"`
				SecondFlags       uint32    `json:"second_flags"`
				FirstType         uint16    `json:"first_type"`
				BoxSize           [2]uint32 `json:"box_size"`
				BoxOffsets        [8]uint32 `json:"box_offsets"`
				Move              int32     `json:"move"`
				Origin            [2]uint32 `json:"origin"`
				Point             [2]uint32 `json:"point"`
				Lo                uint32    `json:"lo"`
				Hi                uint32    `json:"hi"`
				AllowLo           int32     `json:"allow_lo"`
				AllowHi           int32     `json:"allow_hi"`
				CABI              bool      `json:"cabi"`
			} `json:"input"`
			First struct {
				New         [2]uint32 `json:"new"`
				Velocity    [2]uint32 `json:"velocity"`
				Accumulator [2]uint32 `json:"accumulator"`
			} `json:"first"`
			Second struct {
				New         [2]uint32 `json:"new"`
				Velocity    [2]uint32 `json:"velocity"`
				Accumulator [2]uint32 `json:"accumulator"`
			} `json:"second"`
			Events []struct {
				Service string `json:"service"`
			} `json:"events"`
			Result uint64    `json:"result_bits64"`
			Normal [2]uint32 `json:"normal"`
			Return int       `json:"return_int"`
		} `json:"rows"`
	}
	if err := json.Unmarshal(blob, &fixture); err != nil {
		t.Fatal(err)
	}
	if fixture.Contexts != 966 {
		t.Fatal("incomplete original circle fixture")
	}
	selected := 0
	for _, row := range fixture.Rows {
		i := row.Input
		if row.Mode != "actual-startup" {
			continue
		}
		if i.Kind != "box-geometry" && i.Kind != "first-nearest" && i.Kind != "second-nearest" && !i.CABI {
			continue
		}
		random := false
		for _, e := range row.Events {
			if e.Service == "random" {
				random = true
			}
		}
		if random {
			continue
		}
		selected++
		t.Run(i.Name, func(t *testing.T) {
			a, fa := alloc.New(server.Object{})
			b, fb := alloc.New(server.Object{})
			t.Cleanup(fa)
			t.Cleanup(fb)
			rigidNativeAddress(t, unsafe.Pointer(a))
			rigidNativeAddress(t, unsafe.Pointer(b))
			*a = server.Object{TypeInd: i.FirstType, ObjClass: object.Class(i.FirstClass), ObjFlags: object.Flags(i.FirstFlags), PosVec: rigidNativePoint(i.FirstPosition), NewPos: rigidNativePoint(i.FirstNew), VelVec: rigidNativePoint(i.FirstVelocity), Pos24: rigidNativePoint(i.FirstAccumulator), Mass: math.Float32frombits(i.FirstMass), Field29: 0x13572468}
			*b = server.Object{TypeInd: 102, ObjClass: object.Class(i.SecondClass), ObjFlags: object.Flags(i.SecondFlags), PosVec: rigidNativePoint(i.SecondPosition), NewPos: rigidNativePoint(i.SecondNew), VelVec: rigidNativePoint(i.SecondVelocity), Pos24: rigidNativePoint(i.SecondAccumulator), Mass: math.Float32frombits(i.SecondMass), Field29: 0x87654321}
			a.Shape.Kind, b.Shape.Kind = server.ShapeKindCircle, server.ShapeKindBox
			a.Shape.Circle.R, b.Shape.Circle.R = math.Float32frombits(i.FirstRadius), math.Float32frombits(i.SecondRadius)
			b.Shape.Box.W, b.Shape.Box.H = math.Float32frombits(i.BoxSize[0]), math.Float32frombits(i.BoxSize[1])
			fields := []*float32{&b.Shape.Box.LeftTop, &b.Shape.Box.LeftBottom, &b.Shape.Box.LeftBottom2, &b.Shape.Box.LeftTop2, &b.Shape.Box.RightTop, &b.Shape.Box.RightBottom, &b.Shape.Box.RightBottom2, &b.Shape.Box.RightTop2}
			for index, p := range fields {
				*p = math.Float32frombits(i.BoxOffsets[index])
			}
			wantA := rigidNativeExpected(a, row.First.New, row.First.Velocity, row.First.Accumulator)
			wantB := rigidNativeExpected(b, row.Second.New, row.Second.Velocity, row.Second.Accumulator)
			normal := types.Ptf(.5, -.5)
			result := 0
			switch i.Kind {
			case "box-geometry":
				got := circleGeometryCABI54A990(rigidNativePoint(i.FirstNew), math.Float32frombits(i.FirstRadius), b, &normal)
				if math.Float64bits(got) != row.Result {
					t.Errorf("C ABI result=%016x original=%016x", math.Float64bits(got), row.Result)
				}
			case "first-nearest", "second-nearest":
				if circleNearestCABI550280(i.Kind == "second-nearest", rigidNativePoint(i.Origin), math.Float32frombits(i.Lo), math.Float32frombits(i.Hi), i.AllowLo, i.AllowHi, rigidNativePoint(i.Point), &normal) {
					result = 1
				}
			default:
				Nox_xxx_allocHitArray_5486D0()
				t.Cleanup(Nox_xxx_allocHitArray_5486D0)
				result = circleObjectCABI550D00(i.Kind, a, b, i.Move, rigidNativePoint(i.Point))
			}
			if i.Kind == "box-geometry" || i.Kind == "first-nearest" || i.Kind == "second-nearest" {
				if rigidNativeBits(normal) != row.Normal {
					t.Errorf("C ABI normal=%08x original=%08x", rigidNativeBits(normal), row.Normal)
				}
			}
			if result != row.Return || !bytes.Equal(rigidNativeObjectBytes(a), wantA) || !bytes.Equal(rigidNativeObjectBytes(b), wantB) {
				t.Errorf("C ABI return=%d original=%d; object bytes=%t/%t accumulator=%08x original=%08x", result, row.Return, bytes.Equal(rigidNativeObjectBytes(a), wantA), bytes.Equal(rigidNativeObjectBytes(b), wantB), rigidNativeBits(a.Pos24), row.First.Accumulator)
			}
		})
	}
	if selected != 197 {
		t.Fatalf("C ABI original conditions=%d", selected)
	}
}
