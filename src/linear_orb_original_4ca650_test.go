package opennox

import (
	"encoding/json"
	"image"
	"math"
	"os"
	"reflect"
	"testing"
	"unsafe"

	"github.com/opennox/opennox/v1/client"
)

func TestLinearOrbOriginal4CA650(t *testing.T) {
	raw, err := os.ReadFile("testdata/linear_orb_4ca650.json")
	if err != nil {
		t.Fatal(err)
	}
	var fixture struct {
		SHA       string `json:"original_sha256"`
		Unchanged bool   `json:"original_unchanged"`
		Rows      []struct {
			Input struct {
				Name  string    `json:"name"`
				Pos   [2]int    `json:"pos"`
				Dest  [2]uint16 `json:"dest"`
				Speed byte      `json:"speed"`
			} `json:"input"`
			Events []struct {
				Service string `json:"service"`
				Pos     [2]int `json:"pos"`
			} `json:"events"`
			Return int `json:"return"`
		} `json:"rows"`
	}
	if err := json.Unmarshal(raw, &fixture); err != nil {
		t.Fatal(err)
	}
	if fixture.SHA != "0040e2c0683b4d73a5fb976e400d5087dca680df2b195c9e27f8edbda2d4974a" || !fixture.Unchanged || len(fixture.Rows) != 1171 {
		t.Fatal("unchanged original linear-orb fixture")
	}
	for _, row := range fixture.Rows {
		t.Run(row.Input.Name, func(t *testing.T) {
			in := row.Input
			dr := &client.Drawable{PosVec: image.Pt(in.Pos[0], in.Pos[1]), LightDir: 0x1234, Field_42: 0x13579bdf}
			if unsafe.Sizeof(uintptr(0)) == 8 && uintptr(unsafe.Pointer(dr)) <= math.MaxUint32 {
				t.Fatal("native drawable address must exceed4GiB")
			}
			dr.UnionEffect().Field_108 = uint32(in.Dest[0]) | uint32(in.Dest[1])<<16
			dr.UnionEffect().Field_110 = uint32(in.Speed)<<24 | 0x123456
			before := *dr
			next, done := linearOrbStep4CA650(dr)
			if len(row.Events) != 1 {
				t.Fatal("one original position/delete service")
			}
			event := row.Events[0]
			wantDone := event.Service == "delete"
			if event.Service != "position" && !wantDone {
				t.Fatal(event.Service)
			}
			if done != wantDone || row.Return != 1-boolIntLinearOrb4CA650(wantDone) {
				t.Fatalf("done=%v want=%v originalReturn=%d", done, wantDone, row.Return)
			}
			// The caller discards the computed next point on deletion. Only
			// original position-service arguments define an observable next point.
			if !done && next != image.Pt(event.Pos[0], event.Pos[1]) {
				t.Fatalf("next=%v want=%v", next, event.Pos)
			}
			if !reflect.DeepEqual(*dr, before) {
				t.Fatal("drawable/input canary changed")
			}
		})
	}
}

func boolIntLinearOrb4CA650(v bool) int {
	if v {
		return 1
	}
	return 0
}
