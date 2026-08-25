package legacy

import (
	"math"
	"testing"

	"github.com/opennox/libs/types"
	"github.com/opennox/opennox/v1/server"
)

func TestCircleBoxCollisionNative54A990GAMEEXERegions(t *testing.T) {
	box := &server.Object{NewPos: types.Ptf(100, 100)}
	box.Shape.Kind = server.ShapeKindBox
	box.Shape.Box.W = 20
	box.Shape.Box.H = 20
	box.Shape.Box.Calc()

	tests := []struct {
		name       string
		center     types.Pointf
		radius     float32
		wantResult float64
		wantNormal types.Pointf
	}{
		{name: "inside", center: types.Ptf(101, 100), radius: 10, wantResult: 1, wantNormal: types.Ptf(1, 0)},
		{name: "top right edge", center: types.Ptf(110, 90), radius: 10, wantResult: 5.8579, wantNormal: types.Ptf(0.70709997, -0.70709997)},
		{name: "bottom left edge", center: types.Ptf(90, 110), radius: 10, wantResult: 5.8579, wantNormal: types.Ptf(-0.70709997, 0.70709997)},
		{name: "top left edge", center: types.Ptf(90, 90), radius: 10, wantResult: 5.8579, wantNormal: types.Ptf(-0.70709997, -0.70709997)},
		{name: "bottom right edge", center: types.Ptf(110, 110), radius: 10, wantResult: 5.8579, wantNormal: types.Ptf(0.70709997, 0.70709997)},
		{name: "top corner", center: types.Ptf(100, 80), radius: 10, wantResult: 4.1419983, wantNormal: types.Ptf(0, -1)},
		{name: "left corner", center: types.Ptf(80, 100), radius: 10, wantResult: 4.1419983, wantNormal: types.Ptf(-1, 0)},
		{name: "right corner", center: types.Ptf(120, 100), radius: 10, wantResult: 4.1419983, wantNormal: types.Ptf(1, 0)},
		{name: "bottom corner", center: types.Ptf(100, 120), radius: 10, wantResult: 4.1419983, wantNormal: types.Ptf(0, 1)},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			gotResult, gotNormal := circleBoxCollisionNative54A990(tc.center, tc.radius, box)
			if math.Abs(gotResult-tc.wantResult) > 0.001 {
				t.Fatalf("result = %.9f, want %.9f", gotResult, tc.wantResult)
			}
			if math.Abs(float64(gotNormal.X-tc.wantNormal.X)) > 0.0001 ||
				math.Abs(float64(gotNormal.Y-tc.wantNormal.Y)) > 0.0001 {
				t.Fatalf("normal = (%g, %g), want (%g, %g)", gotNormal.X, gotNormal.Y, tc.wantNormal.X, tc.wantNormal.Y)
			}
		})
	}
}

func TestCircleBoxCollisionNative54A990RejectedCornerKeepsNormal(t *testing.T) {
	box := &server.Object{NewPos: types.Ptf(100, 100)}
	box.Shape.Kind = server.ShapeKindBox
	box.Shape.Box.W = 20
	box.Shape.Box.H = 20
	box.Shape.Box.Calc()

	gotResult, gotNormal := circleBoxCollisionNative54A990(types.Ptf(100, 60), 10, box)
	if gotResult >= 0 {
		t.Fatalf("result = %g, want a rejected negative distance", gotResult)
	}
	if gotNormal != (types.Pointf{}) {
		t.Fatalf("normal = %+v, want untouched zero value", gotNormal)
	}
}
