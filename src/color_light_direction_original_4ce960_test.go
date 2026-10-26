package opennox

import (
	"encoding/hex"
	"encoding/json"
	"image"
	"math"
	"os"
	"reflect"
	"testing"
	"unsafe"

	"github.com/opennox/opennox/v1/client"
)

func TestColorLightDirectionOriginal4CE960(t *testing.T) {
	raw, err := os.ReadFile("testdata/color_light_direction_4ce960.json")
	if err != nil {
		t.Fatal(err)
	}
	var fixture struct {
		SHA       string `json:"original_sha256"`
		Unchanged bool   `json:"original_unchanged"`
		FPCW      uint32 `json:"fpcw"`
		Rows      []struct {
			Input struct {
				Name        string  `json:"name"`
				Kind        string  `json:"kind"`
				Value       int64   `json:"value"`
				Frame       uint32  `json:"frame"`
				FPS         uint32  `json:"fps"`
				Field42     *uint32 `json:"field42"`
				InitialDir  *uint16 `json:"initial_dir"`
				InitialSize *uint16 `json:"initial_size"`
				Code        uint32  `json:"code"`
				Target      bool    `json:"target"`
				Pos         []int   `json:"pos"`
				TargetPos   []int   `json:"target_pos"`
				LookupPos   []int   `json:"lookup_pos"`
			} `json:"input"`
			Data   string `json:"data_hex"`
			Events []struct {
				Service string `json:"service"`
				Code    uint32 `json:"code"`
			} `json:"events"`
			Loads []struct {
				Service string `json:"service"`
				Value   uint32 `json:"value"`
			} `json:"loads"`
			Direction uint16  `json:"direction"`
			Size      uint16  `json:"size"`
			Field42   uint32  `json:"field42"`
			Return    *uint64 `json:"return_qword"`
		} `json:"rows"`
	}
	if err := json.Unmarshal(raw, &fixture); err != nil {
		t.Fatal(err)
	}
	if fixture.SHA != "0040e2c0683b4d73a5fb976e400d5087dca680df2b195c9e27f8edbda2d4974a" || !fixture.Unchanged || fixture.FPCW != 0xe7f || len(fixture.Rows) != 4067 {
		t.Fatal("unchanged original direction fixture")
	}
	for _, row := range fixture.Rows {
		t.Run(row.Input.Name, func(t *testing.T) {
			in := row.Input
			dr := &client.Drawable{PosVec: image.Pt(150, -250), LightDir: 0x1234, LightPenumbra: 0xabcd, Field_42: 0x13579bdf, LightFlags: 0x2468, LightIntensity: -3.25}
			target := &client.Drawable{}
			if unsafe.Sizeof(uintptr(0)) == 8 && (uintptr(unsafe.Pointer(dr)) <= math.MaxUint32 || uintptr(unsafe.Pointer(target)) <= math.MaxUint32) {
				t.Fatal("native drawable addresses must exceed4GiB")
			}
			if in.Field42 != nil {
				dr.Field_42 = *in.Field42
			}
			if in.InitialDir != nil {
				dr.LightDir = *in.InitialDir
			}
			if in.InitialSize != nil {
				dr.LightPenumbra = *in.InitialSize
			}
			if len(in.Pos) == 2 {
				dr.PosVec = image.Pt(in.Pos[0], in.Pos[1])
			}
			if len(in.TargetPos) == 2 {
				target.PosVec = image.Pt(in.TargetPos[0], in.TargetPos[1])
			}
			data := dr.ColorLightData()
			decoded, err := hex.DecodeString(row.Data)
			if err != nil || len(decoded) != len(data) {
				t.Fatal("original data", err)
			}
			copy(data[:], decoded)
			before, targetBefore := *dr, *target
			var loads, lookups []string
			hooks := colorLightDrawableHooks4CE390{
				fps:   func() uint32 { loads = append(loads, "fps"); return in.FPS },
				frame: func() uint32 { loads = append(loads, "frame"); return in.Frame },
				byStaticCode: func(code int) *client.Drawable {
					if code != int(in.Code) {
						t.Fatalf("static code=%x want=%x", code, in.Code)
					}
					lookups = append(lookups, "lookup")
					if len(in.LookupPos) == 2 {
						dr.PosVec = image.Pt(in.LookupPos[0], in.LookupPos[1])
					}
					if in.Target {
						return target
					}
					return nil
				},
			}
			var result int64
			switch in.Kind {
			case "direction":
				result = setColorLightDirection4CE8C0(dr, int(in.Value))
			case "size":
				result = setColorLightSize4CE760(dr, int(in.Value))
			case "rotate":
				rotateColorLight4CE960(dr, data, hooks)
			case "track":
				trackColorLight4CE8C0(dr, data, hooks.byStaticCode)
			default:
				t.Fatal(in.Kind)
			}
			if dr.LightDir != row.Direction || dr.LightPenumbra != row.Size || dr.Field_42 != row.Field42 {
				t.Fatalf("direction/size/lock=%d/%d/%x want=%d/%d/%x", dr.LightDir, dr.LightPenumbra, dr.Field_42, row.Direction, row.Size, row.Field42)
			}
			if row.Return != nil && uint64(result) != *row.Return {
				t.Fatalf("return=%x want=%x", uint64(result), *row.Return)
			}
			var wantLoads, wantLookups []string
			for _, l := range row.Loads {
				wantLoads = append(wantLoads, l.Service)
			}
			for _, e := range row.Events {
				if e.Service == "lookup" {
					wantLookups = append(wantLookups, "lookup")
				}
			}
			if !reflect.DeepEqual(loads, wantLoads) || !reflect.DeepEqual(lookups, wantLookups) {
				t.Fatalf("loads=%v/%v lookup=%v/%v", loads, wantLoads, lookups, wantLookups)
			}
			after := *dr
			after.LightDir, after.LightPenumbra, after.Field_42 = before.LightDir, before.LightPenumbra, before.Field_42
			wantPos := before.PosVec
			if len(in.LookupPos) == 2 && len(lookups) > 0 {
				wantPos = image.Pt(in.LookupPos[0], in.LookupPos[1])
			}
			if dr.PosVec != wantPos {
				t.Fatalf("position=%v want=%v", dr.PosVec, wantPos)
			}
			after.PosVec = before.PosVec
			if !reflect.DeepEqual(after, before) || !reflect.DeepEqual(*target, targetBefore) {
				t.Fatal("drawable/target canary changed")
			}
		})
	}
}
