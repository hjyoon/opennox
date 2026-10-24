package opennox

import (
	"encoding/binary"
	"encoding/json"
	"image"
	"math"
	"os"
	"reflect"
	"testing"
	"unsafe"

	"github.com/opennox/opennox/v1/client"
)

func TestManaBombCancelOriginal492B09(t *testing.T) {
	raw, err := os.ReadFile("testdata/mana_bomb_cancel_original_492b09.json")
	if err != nil {
		t.Fatal(err)
	}
	type input struct {
		Name      string  `json:"name"`
		Radius    float32 `json:"radius"`
		Center    [2]int  `json:"center"`
		Fill      byte    `json:"fill"`
		Failure   string  `json:"failure"`
		Frame     uint32  `json:"frame"`
		FrameStep uint32  `json:"frame_step"`
		Connected bool    `json:"connected"`
	}
	type activation struct {
		ID     int       `json:"id"`
		Pos    [2]int    `json:"pos"`
		Effect [5]uint32 `json:"effect"`
		Z      int16     `json:"z"`
		VelZ   int8      `json:"vel_z"`
		Aux    byte      `json:"aux"`
	}
	var fixture struct {
		SHA       string `json:"original_sha256"`
		Unchanged bool   `json:"original_unchanged"`
		Rows      []struct {
			Input       input            `json:"input"`
			FPCW        uint32           `json:"fpcw"`
			Trace       []map[string]any `json:"trace"`
			Activations []activation     `json:"activations"`
			Consumed    int              `json:"consumed_bytes"`
		} `json:"rows"`
	}
	if err := json.Unmarshal(raw, &fixture); err != nil {
		t.Fatal(err)
	}
	if fixture.SHA != "0040e2c0683b4d73a5fb976e400d5087dca680df2b195c9e27f8edbda2d4974a" || !fixture.Unchanged || len(fixture.Rows) != 32 {
		t.Fatal("unchanged original cancellation-arm observations")
	}
	count := 0
	for _, row := range fixture.Rows {
		// Preserve the current nearest-even helper contract. The live network
		// caller's rounding mode remains outside this service-order checkpoint.
		if row.FPCW != 0x27f {
			continue
		}
		count++
		t.Run(row.Input.Name, func(t *testing.T) {
			in := row.Input
			var trace []map[string]any
			var states []activation
			ids := map[*client.Drawable]int{}
			initial := map[*client.Drawable]client.Drawable{}
			randomIndex, spawnIndex, frameIndex := 0, 0, uint32(0)
			pattern := uint32(in.Fill) * 0x01010101
			hooks := manaBombCancelHooks48EA70{
				connected: func() bool {
					trace = append(trace, map[string]any{"service": "connected", "value": in.Connected})
					return in.Connected
				},
				radius: func() float32 {
					trace = append(trace, map[string]any{"service": "radius"})
					return in.Radius
				},
				typeID: func() int { return 7 }, // same preseeded original type identity
				random: func(lo, hi int) int {
					randomIndex++
					v := 0
					switch {
					case lo == 0 && hi == 255:
						v = (randomIndex*37 + 11) % 256
					case lo == 0:
						if hi >= 0 {
							v = []int{0, hi, hi / 2}[randomIndex%3]
						}
					case lo == 30 && hi == 40:
						v = 35
					case lo == 4 && hi == 10:
						v = 7
					default:
						t.Fatalf("unexpected RNG range %d..%d", lo, hi)
					}
					trace = append(trace, map[string]any{"service": "random", "lo": lo, "hi": hi, "value": v})
					return v
				},
				frame: func() uint32 {
					v := in.Frame + frameIndex*in.FrameStep
					frameIndex++
					trace = append(trace, map[string]any{"service": "frame", "value": v})
					return v
				},
				spawn: func(typ int, pos image.Point) *client.Drawable {
					spawnIndex++
					fail := in.Failure == "all" || in.Failure == "every5" && spawnIndex%5 == 0
					trace = append(trace, map[string]any{"service": "spawn", "type": typ, "pos": [2]int{pos.X, pos.Y}, "success": !fail, "id": spawnIndex})
					if fail {
						return nil
					}
					d := &client.Drawable{PosVec: pos, ZVal: uint16(pattern), VelZ: int8(in.Fill), Field_74_4: in.Fill, Field_127: pattern}
					e := d.UnionEffect()
					e.Field_108, e.Field_109, e.Field_110, e.Field_111, e.Field_112 = pattern, pattern, pattern, pattern, pattern
					if unsafe.Sizeof(uintptr(0)) == 8 && uintptr(unsafe.Pointer(d)) <= math.MaxUint32 {
						t.Fatal("native pointer must exceed4GiB")
					}
					ids[d], initial[d] = spawnIndex, *d
					return d
				},
				activate: func(d *client.Drawable) {
					e := d.UnionEffect()
					states = append(states, activation{ids[d], [2]int{d.PosVec.X, d.PosVec.Y}, [5]uint32{e.Field_108, e.Field_109, e.Field_110, e.Field_111, e.Field_112}, int16(d.ZVal), d.VelZ, d.Field_74_4})
					trace = append(trace, map[string]any{"service": "activate", "id": ids[d]})
					before, scrubbed := initial[d], *d
					scrubbed.ZVal, scrubbed.VelZ, scrubbed.Field_74_4 = before.ZVal, before.VelZ, before.Field_74_4
					*scrubbed.UnionEffect() = *before.UnionEffect()
					if !reflect.DeepEqual(scrubbed, before) {
						t.Fatal("write outside original declared stores")
					}
				},
			}
			packet := []byte{0xa3, 0, 0, 0, 0}
			binary.LittleEndian.PutUint16(packet[1:3], uint16(in.Center[0]))
			binary.LittleEndian.PutUint16(packet[3:5], uint16(in.Center[1]))
			if n := handleManaBombCancelNative48EA70(packet, hooks); n != row.Consumed {
				t.Fatalf("consumed=%d want%d", n, row.Consumed)
			}
			if len(states) != len(row.Activations) || len(states) != 0 && !reflect.DeepEqual(states, row.Activations) {
				t.Fatal("original activation state differs")
			}
			normalize := func(v any) any {
				b, err := json.Marshal(v)
				if err != nil {
					t.Fatal(err)
				}
				var result any
				if err := json.Unmarshal(b, &result); err != nil {
					t.Fatal(err)
				}
				return result
			}
			if !reflect.DeepEqual(normalize(trace), normalize(row.Trace)) {
				for i, got := range trace {
					if i >= len(row.Trace) || !reflect.DeepEqual(normalize(got), normalize(row.Trace[i])) {
						t.Fatalf("original service trace differs at%d: got%v want%v", i, got, row.Trace[i])
					}
				}
				t.Fatal("original service trace length differs")
			}
		})
	}
	if count != 16 {
		t.Fatalf("selected original nearest-mode rows=%d", count)
	}
}
