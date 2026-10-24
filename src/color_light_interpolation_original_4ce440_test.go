package opennox

import (
	"encoding/binary"
	"encoding/hex"
	"encoding/json"
	"math"
	"os"
	"testing"
	"unsafe"

	"github.com/opennox/opennox/v1/client"
	"github.com/opennox/opennox/v1/client/noxrender"
)

func TestColorLightInterpolationOriginal4CE440(t *testing.T) {
	raw, err := os.ReadFile("testdata/color_light_interpolation_4ce440.json")
	if err != nil {
		t.Fatal(err)
	}
	var fixture struct {
		SHA       string `json:"original_sha256"`
		Unchanged bool   `json:"original_unchanged"`
		FPCW      uint32 `json:"fpcw"`
		Rows      []struct {
			Input struct {
				Name    string `json:"name"`
				Kind    string `json:"kind"`
				Count   byte   `json:"count"`
				Frame   uint32 `json:"frame"`
				Field42 uint32 `json:"field42"`
			} `json:"input"`
			Data   string `json:"data_hex"`
			Events []struct {
				Service string    `json:"service"`
				Values  [3]uint32 `json:"values"`
				Bits    uint32    `json:"value_bits"`
				Value   uint32    `json:"value"`
			} `json:"events"`
			Frames []uint32 `json:"frame_loads"`
			EAX    uint32   `json:"eax"`
		} `json:"rows"`
	}
	if err := json.Unmarshal(raw, &fixture); err != nil {
		t.Fatal(err)
	}
	if fixture.SHA != "0040e2c0683b4d73a5fb976e400d5087dca680df2b195c9e27f8edbda2d4974a" || !fixture.Unchanged || fixture.FPCW != 0xe7f || len(fixture.Rows) != 1548 {
		t.Fatal("unchanged original gameplay53/chop fixture")
	}
	for _, row := range fixture.Rows {
		t.Run(row.Input.Name, func(t *testing.T) {
			dr := &client.Drawable{Field_42: row.Input.Field42, LightFlags: 0x1234, LightColor: noxrender.RGB{R: 11, G: 22, B: 33}, LightIntensity: -3.25, LightPenumbra: 0x1234}
			if unsafe.Sizeof(uintptr(0)) == 8 && uintptr(unsafe.Pointer(dr)) <= math.MaxUint32 {
				t.Fatal("native drawable address must exceed4GiB")
			}
			data := dr.ColorLightData()
			decoded, err := hex.DecodeString(row.Data)
			if err != nil || len(decoded) != len(data) {
				t.Fatal("packed original animation data", err)
			}
			copy(data[:], decoded)
			before := *data
			frameCalls := 0
			frame := func() uint32 { frameCalls++; return row.Input.Frame }
			var result int
			switch row.Input.Kind {
			case "color":
				alterColorLightColor4CE440(dr, data, row.Input.Count, frame)
			case "intensity":
				alterColorLightIntensity4CE610(dr, data, row.Input.Count, frame)
			case "radius":
				result = alterColorLightRadius4CE760(dr, data, row.Input.Count, frame)
			default:
				t.Fatal("original helper kind", row.Input.Kind)
			}
			if *data != before || dr.Field_42 != row.Input.Field42 || frameCalls != len(row.Frames) {
				t.Fatalf("input/canary/frame drift: data=%t field42=%x/%x frames=%d/%d", *data != before, dr.Field_42, row.Input.Field42, frameCalls, len(row.Frames))
			}
			wantRGB, wantFlags := (noxrender.RGB{R: 11, G: 22, B: 33}), uint32(0x1234)
			wantIntensity, wantPenumbra := float32(-3.25), uint16(0x1234)
			if len(row.Events) > 1 {
				t.Fatal("single original setter boundary")
			}
			for _, event := range row.Events {
				// Compare raw original setter arguments before the unchanged
				// production setters clamp intensity or quantize cone size.
				delayOffset, keyOffset, stride, loopMask := 82, 2, 3, byte(1)
				if row.Input.Kind == "intensity" {
					delayOffset, keyOffset, stride, loopMask = 84, 50, 1, 4
				}
				if row.Input.Kind == "radius" {
					delayOffset, keyOffset, stride, loopMask = 86, 66, 1, 0x10
				}
				delay := binary.LittleEndian.Uint16(data[delayOffset : delayOffset+2])
				cycle := colorLightDivideChop53_4CE440(float64(row.Input.Frame), float64(uint32(row.Input.Count)*uint32(delay)))
				from, to, subframe := colorLightFrameSelection4CE440(cycle, row.Input.Count, delay, data[0]&loopMask != 0)
				value := func(channel int) float64 {
					return colorLightLinear4CE440(int(data[keyOffset+stride*from+channel]), int(data[keyOffset+stride*to+channel]), subframe, delay)
				}
				switch event.Service {
				case "color":
					for channel := 0; channel < 3; channel++ {
						if uint32(int64(value(channel))) != event.Values[channel] {
							t.Fatalf("raw original color channel%d=%v/%d", channel, value(channel), event.Values[channel])
						}
					}
				case "intensity":
					if got := math.Float32bits(fireballSpill32_52C790(value(0))); got != event.Bits {
						t.Fatalf("raw original intensity=%08x/%08x", got, event.Bits)
					}
				case "radius":
					if got := uint32(int64(value(0))); got != event.Value {
						t.Fatalf("raw original radius=%d/%d", got, event.Value)
					}
				}
				switch event.Service {
				case "color":
					wantFlags = 2
					wantRGB = noxrender.RGB{R: int(event.Values[0]), G: int(event.Values[1]), B: int(event.Values[2])}
				case "intensity":
					wantIntensity = math.Float32frombits(event.Bits)
					if wantIntensity > 63 {
						wantIntensity = 63
					}
				case "radius":
					wantPenumbra = uint16(setColorLightSize4CE760(&client.Drawable{}, int(event.Value)))
				default:
					t.Fatal("original setter", event.Service)
				}
			}
			if uint32(dr.LightFlags) != wantFlags || dr.LightColor != wantRGB || math.Float32bits(dr.LightIntensity) != math.Float32bits(wantIntensity) || dr.LightPenumbra != wantPenumbra {
				t.Fatalf("original boundary projection: flags=%x/%x RGB=%+v/%+v intensity=%08x/%08x penumbra=%d/%d", dr.LightFlags, wantFlags, dr.LightColor, wantRGB, math.Float32bits(dr.LightIntensity), math.Float32bits(wantIntensity), dr.LightPenumbra, wantPenumbra)
			}
			if row.Input.Kind == "radius" && len(row.Events) == 0 && uint32(result) != row.EAX {
				t.Fatalf("radius guard EAX=%x/%x", uint32(result), row.EAX)
			}
		})
	}
}
