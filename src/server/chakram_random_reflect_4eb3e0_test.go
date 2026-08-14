package server

import (
	"math"
	"testing"

	"github.com/opennox/libs/types"
)

func TestChakramRandomReflectNative4EB3E0RotationAndMoveOrder(t *testing.T) {
	obj := &Object{
		PosVec:  types.Pointf{X: 100, Y: 200},
		NewPos:  types.Pointf{X: 300, Y: 400},
		PrevPos: types.Pointf{X: -12.5, Y: 33.25},
		VelVec:  types.Pointf{X: 3, Y: 4},
	}
	var moveCalls int
	chakramRandomReflectNative4EB3E0(obj, chakramRandomReflectNativeDeps4EB3E0{
		randomInt: func(minimum, maximum int32) int32 {
			if minimum != -64 || maximum != 64 {
				t.Fatalf("random bounds = (%d, %d), want (-64, 64)", minimum, maximum)
			}
			return 0
		},
		moveUpdate: func(got *Object) {
			moveCalls++
			if got != obj {
				t.Fatalf("move object = %p, want %p", got, obj)
			}
			if got.PosVec != got.PrevPos || got.NewPos != got.PrevPos {
				t.Fatalf("move observed positions = (%+v, %+v), want previous %+v", got.PosVec, got.NewPos, got.PrevPos)
			}
		},
	})
	direction := uint8(directionFromVector509ED0(3, 4) + 128)
	cosine, sine := SinCosDir(direction)
	wantVelocity := types.Pointf{
		X: float32(float64(float32(5)) * float64(cosine)),
		Y: float32(float64(float32(5)) * float64(sine)),
	}
	if obj.VelVec != wantVelocity || moveCalls != 1 {
		t.Fatalf("result = (velocity %+v, moves %d), want (%+v, 1)", obj.VelVec, moveCalls, wantVelocity)
	}
}

func TestChakramRandomReflectNative4EB3E0WrapsDirectionModulo256(t *testing.T) {
	tests := []struct {
		name   string
		vel    types.Pointf
		random int32
	}{
		{"negative sum", types.Pointf{X: 1, Y: 0}, -64},
		{"positive overflow", types.Pointf{X: -1, Y: 0}, 64},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			obj := &Object{VelVec: tc.vel}
			chakramRandomReflectNative4EB3E0(obj, chakramRandomReflectNativeDeps4EB3E0{
				randomInt:  func(int32, int32) int32 { return tc.random },
				moveUpdate: func(*Object) {},
			})
			direction := uint8(tc.random + directionFromVector509ED0(tc.vel.X, tc.vel.Y) + 128)
			cosine, sine := SinCosDir(direction)
			speed := float32(math.Sqrt(float64(tc.vel.X)*float64(tc.vel.X) + float64(tc.vel.Y)*float64(tc.vel.Y)))
			want := types.Pointf{
				X: float32(float64(speed) * float64(cosine)),
				Y: float32(float64(speed) * float64(sine)),
			}
			if obj.VelVec != want {
				t.Fatalf("velocity = %+v, want %+v for direction %d", obj.VelVec, want, direction)
			}
		})
	}
}

func TestChakramRandomReflectNative4EB3E0PreservesPreviousPositionBits(t *testing.T) {
	previous := types.Pointf{
		X: math.Float32frombits(0x7fc01234),
		Y: math.Float32frombits(0x80000000),
	}
	obj := &Object{PrevPos: previous}
	chakramRandomReflectNative4EB3E0(obj, chakramRandomReflectNativeDeps4EB3E0{
		randomInt:  func(int32, int32) int32 { return 0 },
		moveUpdate: func(*Object) {},
	})
	for _, got := range []types.Pointf{obj.PosVec, obj.NewPos} {
		if math.Float32bits(got.X) != math.Float32bits(previous.X) ||
			math.Float32bits(got.Y) != math.Float32bits(previous.Y) {
			t.Fatalf("position bits = (%08x, %08x), want (%08x, %08x)",
				math.Float32bits(got.X), math.Float32bits(got.Y),
				math.Float32bits(previous.X), math.Float32bits(previous.Y))
		}
	}
}
